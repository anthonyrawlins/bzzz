package github

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/anthonyrawlins/bzzz/executor"
	"github.com/anthonyrawlins/bzzz/logging"
	"github.com/anthonyrawlins/bzzz/pkg/hive"
	"github.com/anthonyrawlins/bzzz/pkg/types"
	"github.com/anthonyrawlins/bzzz/pubsub"
	"github.com/anthonyrawlins/bzzz/reasoning"
	"github.com/libp2p/go-libp2p/core/peer"
)

// HiveIntegration handles dynamic repository discovery via Hive API
type HiveIntegration struct {
	hiveClient *hive.HiveClient
	githubToken string
	pubsub *pubsub.PubSub
	hlog *logging.HypercoreLog
	ctx context.Context
	config *IntegrationConfig

	// Repository management
	repositories map[int]*RepositoryClient // projectID -> GitHub client
	repositoryLock sync.RWMutex

	// Conversation tracking
	activeDiscussions map[string]*Conversation // "projectID:taskID" -> conversation
	discussionLock sync.RWMutex
}

// RepositoryClient wraps a GitHub client for a specific repository
type RepositoryClient struct {
	Client     *Client
	Repository hive.Repository
	LastSync   time.Time
}

// NewHiveIntegration creates a new Hive-based GitHub integration
func NewHiveIntegration(ctx context.Context, hiveClient *hive.HiveClient, githubToken string, ps *pubsub.PubSub, hlog *logging.HypercoreLog, config *IntegrationConfig) *HiveIntegration {
	if config.PollInterval == 0 {
		config.PollInterval = 30 * time.Second
	}
	if config.MaxTasks == 0 {
		config.MaxTasks = 3
	}

	return &HiveIntegration{
		hiveClient:        hiveClient,
		githubToken:       githubToken,
		pubsub:            ps,
		hlog:              hlog,
		ctx:               ctx,
		config:            config,
		repositories:      make(map[int]*RepositoryClient),
		activeDiscussions: make(map[string]*Conversation),
	}
}

// Start begins the Hive-GitHub integration
func (hi *HiveIntegration) Start() {
	fmt.Printf("🔗 Starting Hive-GitHub integration for agent: %s\n", hi.config.AgentID)
	
	// Register the handler for incoming meta-discussion messages
	hi.pubsub.SetAntennaeMessageHandler(hi.handleMetaDiscussion)
	
	// Start repository discovery and task polling
	go hi.repositoryDiscoveryLoop()
	go hi.taskPollingLoop()
}

// repositoryDiscoveryLoop periodically discovers active repositories from Hive
func (hi *HiveIntegration) repositoryDiscoveryLoop() {
	ticker := time.NewTicker(5 * time.Minute) // Check for new repositories every 5 minutes
	defer ticker.Stop()
	
	// Initial discovery
	hi.syncRepositories()
	
	for {
		select {
		case <-hi.ctx.Done():
			return
		case <-ticker.C:
			hi.syncRepositories()
		}
	}
}

// syncRepositories synchronizes the list of active repositories from Hive
func (hi *HiveIntegration) syncRepositories() {
	repositories, err := hi.hiveClient.GetActiveRepositories(hi.ctx)
	if err != nil {
		fmt.Printf("❌ Failed to get active repositories: %v\n", err)
		return
	}
	
	hi.repositoryLock.Lock()
	defer hi.repositoryLock.Unlock()
	
	// Track which repositories we've seen
	currentRepos := make(map[int]bool)
	
	for _, repo := range repositories {
		currentRepos[repo.ProjectID] = true
		
		// Check if we already have a client for this repository
		if _, exists := hi.repositories[repo.ProjectID]; !exists {
			// Create new GitHub client for this repository
			githubConfig := &Config{
				AccessToken: hi.githubToken,
				Owner:       repo.Owner,
				Repository:  repo.Repository,
			}
			
			client, err := NewClient(hi.ctx, githubConfig)
			if err != nil {
				fmt.Printf("❌ Failed to create GitHub client for %s/%s: %v\n", repo.Owner, repo.Repository, err)
				continue
			}
			
			hi.repositories[repo.ProjectID] = &RepositoryClient{
				Client:     client,
				Repository: repo,
				LastSync:   time.Now(),
			}
			
			fmt.Printf("✅ Added repository: %s/%s (Project ID: %d)\n", repo.Owner, repo.Repository, repo.ProjectID)
		}
	}
	
	// Remove repositories that are no longer active
	for projectID := range hi.repositories {
		if !currentRepos[projectID] {
			delete(hi.repositories, projectID)
			fmt.Printf("🗑️ Removed inactive repository (Project ID: %d)\n", projectID)
		}
	}
	
	fmt.Printf("📊 Repository sync complete: %d active repositories\n", len(hi.repositories))
}

// taskPollingLoop periodically polls all repositories for available tasks
func (hi *HiveIntegration) taskPollingLoop() {
	ticker := time.NewTicker(hi.config.PollInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-hi.ctx.Done():
			return
		case <-ticker.C:
			hi.pollAllRepositories()
		}
	}
}

// pollAllRepositories checks all active repositories for available tasks
func (hi *HiveIntegration) pollAllRepositories() {
	hi.repositoryLock.RLock()
	repositories := make([]*RepositoryClient, 0, len(hi.repositories))
	for _, repo := range hi.repositories {
		repositories = append(repositories, repo)
	}
	hi.repositoryLock.RUnlock()
	
	if len(repositories) == 0 {
		return
	}
	
	fmt.Printf("🔍 Polling %d repositories for available tasks...\n", len(repositories))
	
	var allTasks []*types.EnhancedTask
	
	// Collect tasks from all repositories
	for _, repoClient := range repositories {
		tasks, err := hi.getRepositoryTasks(repoClient)
		if err != nil {
			fmt.Printf("❌ Failed to get tasks from %s/%s: %v\n", 
				repoClient.Repository.Owner, repoClient.Repository.Repository, err)
			continue
		}
		allTasks = append(allTasks, tasks...)
	}
	
	if len(allTasks) == 0 {
		return
	}
	
	fmt.Printf("📋 Found %d total available tasks across all repositories\n", len(allTasks))
	
	// Apply filtering and selection
	suitableTasks := hi.filterSuitableTasks(allTasks)
	if len(suitableTasks) == 0 {
		fmt.Printf("⚠️ No suitable tasks for agent capabilities: %v\n", hi.config.Capabilities)
		return
	}
	
	// Select and claim the highest priority task
	task := suitableTasks[0]
	hi.claimAndExecuteTask(task)
}

// getRepositoryTasks fetches available tasks from a specific repository
func (hi *HiveIntegration) getRepositoryTasks(repoClient *RepositoryClient) ([]*types.EnhancedTask, error) {
	// Get tasks from GitHub
	githubTasks, err := repoClient.Client.ListAvailableTasks()
	if err != nil {
		return nil, err
	}
	
	// Convert to enhanced tasks with project context
	var enhancedTasks []*types.EnhancedTask
	for _, task := range githubTasks {
		enhancedTask := &types.EnhancedTask{
			ID:          task.ID,
			Number:      task.Number,
			Title:       task.Title,
			Description: task.Description,
			State:       task.State,
			Labels:      task.Labels,
			Assignee:    task.Assignee,
			CreatedAt:   task.CreatedAt,
			UpdatedAt:   task.UpdatedAt,
			TaskType:    task.TaskType,
			Priority:    task.Priority,
			Requirements: task.Requirements,
			Deliverables: task.Deliverables,
			Context:     task.Context,
			ProjectID:  repoClient.Repository.ProjectID,
			GitURL:     repoClient.Repository.GitURL,
			Repository: repoClient.Repository,
		}
		enhancedTasks = append(enhancedTasks, enhancedTask)
	}
	
		return enhancedTasks, nil
}

// filterSuitableTasks filters tasks based on agent capabilities
func (hi *HiveIntegration) filterSuitableTasks(tasks []*types.EnhancedTask) []*types.EnhancedTask {
	var suitable []*types.EnhancedTask
	
	for _, task := range tasks {
		if hi.canHandleTaskType(task.TaskType) {
			suitable = append(suitable, task)
		}
	}
	
	return suitable
}

// canHandleTaskType checks if this agent can handle the given task type
func (hi *HiveIntegration) canHandleTaskType(taskType string) bool {
	for _, capability := range hi.config.Capabilities {
		if capability == taskType || capability == "general" || capability == "task-coordination" {
			return true
		}
	}
	return false
}

// claimAndExecuteTask claims a task and begins execution
func (hi *HiveIntegration) claimAndExecuteTask(task *types.EnhancedTask) {
	hi.repositoryLock.RLock()
	repoClient, exists := hi.repositories[task.ProjectID]
	hi.repositoryLock.RUnlock()
	
	if !exists {
		fmt.Printf("❌ Repository client not found for project %d\n", task.ProjectID)
		return
	}
	
	// Claim the task in GitHub
	_, err := repoClient.Client.ClaimTask(task.Number, hi.config.AgentID)
	if err != nil {
		fmt.Printf("❌ Failed to claim task %d in %s/%s: %v\n", 
			task.Number, task.Repository.Owner, task.Repository.Repository, err)
		return
	}
	
	fmt.Printf("✋ Claimed task #%d from %s/%s: %s\n", 
		task.Number, task.Repository.Owner, task.Repository.Repository, task.Title)
	
	// Log the claim
	hi.hlog.Append(logging.TaskClaimed, map[string]interface{}{
		"task_id":    task.Number,
		"repository": fmt.Sprintf("%s/%s", task.Repository.Owner, task.Repository.Repository),
		"title":      task.Title,
	})

	// Report claim to Hive
	if err := hi.hiveClient.ClaimTask(hi.ctx, task.ProjectID, task.Number, hi.config.AgentID); err != nil {
		fmt.Printf("⚠️ Failed to report task claim to Hive: %v\n", err)
	}
	
	// Start task execution
	go hi.executeTask(task, repoClient)
}

// executeTask executes a claimed task with reasoning and coordination
func (hi *HiveIntegration) executeTask(task *EnhancedTask, repoClient *RepositoryClient) {
	// Define the dynamic topic for this task
	taskTopic := fmt.Sprintf("bzzz/meta/issue/%d", task.Number)
	hi.pubsub.JoinDynamicTopic(taskTopic)
	defer hi.pubsub.LeaveDynamicTopic(taskTopic)

	fmt.Printf("🚀 Starting execution of task #%d in sandbox...\n", task.Number)

	// The executor now handles the entire iterative process.
	branchName, err := executor.ExecuteTask(hi.ctx, task, hi.hlog)
	if err != nil {
		fmt.Printf("❌ Failed to execute task #%d: %v\n", task.Number, err)
		hi.hlog.Append(logging.TaskFailed, map[string]interface{}{"task_id": task.Number, "reason": "task execution failed in sandbox"})
		return
	}

	// Create a pull request
	pr, err := repoClient.Client.CreatePullRequest(task.Number, branchName, hi.config.AgentID)
	if err != nil {
		fmt.Printf("❌ Failed to create pull request for task #%d: %v\n", task.Number, err)
		hi.hlog.Append(logging.TaskFailed, map[string]interface{}{"task_id": task.Number, "reason": "failed to create pull request"})
		return
	}

	fmt.Printf("✅ Successfully created pull request for task #%d: %s\n", task.Number, pr.GetHTMLURL())
	hi.hlog.Append(logging.TaskCompleted, map[string]interface{}{
		"task_id":   task.Number,
		"pr_url":    pr.GetHTMLURL(),
		"pr_number": pr.GetNumber(),
	})

	// Report completion to Hive
	if err := hi.hiveClient.UpdateTaskStatus(hi.ctx, task.ProjectID, task.Number, "completed", map[string]interface{}{
		"pull_request_url": pr.GetHTMLURL(),
	}); err != nil {
		fmt.Printf("⚠️ Failed to report task completion to Hive: %v\n", err)
	}
}

// requestAssistance publishes a help request to the task-specific topic.
func (hi *HiveIntegration) requestAssistance(task *EnhancedTask, reason, topic string) {
	fmt.Printf("🆘 Agent %s is requesting assistance for task #%d: %s\n", hi.config.AgentID, task.Number, reason)
	hi.hlog.Append(logging.TaskHelpRequested, map[string]interface{}{
		"task_id": task.Number,
		"reason":  reason,
	})

	helpRequest := map[string]interface{}{
		"issue_id":   task.Number,
		"repository": fmt.Sprintf("%s/%s", task.Repository.Owner, task.Repository.Repository),
		"reason":     reason,
	}

	hi.pubsub.PublishToDynamicTopic(topic, pubsub.TaskHelpRequest, helpRequest)
}

// handleMetaDiscussion handles all incoming messages from dynamic and static topics.
func (hi *HiveIntegration) handleMetaDiscussion(msg pubsub.Message, from peer.ID) {
	switch msg.Type {
	case pubsub.TaskHelpRequest:
		hi.handleHelpRequest(msg, from)
	case pubsub.TaskHelpResponse:
		hi.handleHelpResponse(msg, from)
	default:
		// Handle other meta-discussion messages (e.g., peer feedback)
	}
}

// handleHelpRequest is called when another agent requests assistance.
func (hi *HiveIntegration) handleHelpRequest(msg pubsub.Message, from peer.ID) {
	issueID, _ := msg.Data["issue_id"].(float64)
	reason, _ := msg.Data["reason"].(string)
	fmt.Printf("🙋 Received help request for task #%d from %s: %s\n", int(issueID), from.ShortString(), reason)

	// Simple logic: if we are not busy, we can help.
	// A more advanced agent would check its capabilities against the reason.
	canHelp := true // Placeholder for more complex logic

	if canHelp {
		fmt.Printf("✅ Agent %s can help with task #%d\n", hi.config.AgentID, int(issueID))
		hi.hlog.Append(logging.TaskHelpOffered, map[string]interface{}{
			"task_id":      int(issueID),
			"requester_id": from.ShortString(),
		})

		response := map[string]interface{}{
			"issue_id":     issueID,
			"can_help":     true,
			"capabilities": hi.config.Capabilities,
		}
		taskTopic := fmt.Sprintf("bzzz/meta/issue/%d", int(issueID))
		hi.pubsub.PublishToDynamicTopic(taskTopic, pubsub.TaskHelpResponse, response)
	}
}

// handleHelpResponse is called when an agent receives an offer for help.
func (hi *HiveIntegration) handleHelpResponse(msg pubsub.Message, from peer.ID) {
	issueID, _ := msg.Data["issue_id"].(float64)
	canHelp, _ := msg.Data["can_help"].(bool)

	if canHelp {
		fmt.Printf("🤝 Received help offer for task #%d from %s\n", int(issueID), from.ShortString())
		hi.hlog.Append(logging.TaskHelpReceived, map[string]interface{}{
			"task_id":   int(issueID),
			"helper_id": from.ShortString(),
		})
		// In a full implementation, the agent would now delegate a sub-task
		// or use the helper's capabilities. For now, we just log it.
	}
}

// shouldEscalate determines if a task needs human intervention
func (hi *HiveIntegration) shouldEscalate(response string, history []string) bool {
	// Check for escalation keywords
	lowerResponse := strings.ToLower(response)
	keywords := []string{"stuck", "help", "human", "escalate", "clarification needed", "manual intervention"}
	
	for _, keyword := range keywords {
		if strings.Contains(lowerResponse, keyword) {
			return true
		}
	}
	
	// Check conversation length
	if len(history) >= 10 {
		return true
	}
	
	return false
}

// triggerHumanEscalation sends escalation to Hive and N8N
func (hi *HiveIntegration) triggerHumanEscalation(projectID int, convo *Conversation, reason string) {
	hi.hlog.Append(logging.Escalation, map[string]interface{}{
		"task_id": convo.TaskID,
		"reason":  reason,
	})

	// Report to Hive system
	if err := hi.hiveClient.UpdateTaskStatus(hi.ctx, projectID, convo.TaskID, "escalated", map[string]interface{}{
		"escalation_reason": reason,
		"conversation_length": len(convo.History),
		"escalated_by": hi.config.AgentID,
	}); err != nil {
		fmt.Printf("⚠️ Failed to report escalation to Hive: %v\n", err)
	}
	
	fmt.Printf("✅ Task #%d in project %d escalated for human intervention\n", convo.TaskID, projectID)
}
