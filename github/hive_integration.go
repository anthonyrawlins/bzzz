package github

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/deepblackcloud/bzzz/pkg/hive"
	"github.com/deepblackcloud/bzzz/pubsub"
	"github.com/deepblackcloud/bzzz/reasoning"
	"github.com/libp2p/go-libp2p/core/peer"
)

// HiveIntegration handles dynamic repository discovery via Hive API
type HiveIntegration struct {
	hiveClient        *hive.HiveClient
	githubToken       string
	pubsub            *pubsub.PubSub
	ctx               context.Context
	config            *IntegrationConfig
	
	// Repository management
	repositories      map[int]*RepositoryClient // projectID -> GitHub client
	repositoryLock    sync.RWMutex
	
	// Conversation tracking
	activeDiscussions map[string]*Conversation // "projectID:taskID" -> conversation
	discussionLock    sync.RWMutex
}

// RepositoryClient wraps a GitHub client for a specific repository
type RepositoryClient struct {
	Client     *Client
	Repository hive.Repository
	LastSync   time.Time
}

// NewHiveIntegration creates a new Hive-based GitHub integration
func NewHiveIntegration(ctx context.Context, hiveClient *hive.HiveClient, githubToken string, ps *pubsub.PubSub, config *IntegrationConfig) *HiveIntegration {
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
	
	var allTasks []*EnhancedTask
	
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
func (hi *HiveIntegration) getRepositoryTasks(repoClient *RepositoryClient) ([]*EnhancedTask, error) {
	// Get tasks from GitHub
	githubTasks, err := repoClient.Client.ListAvailableTasks()
	if err != nil {
		return nil, err
	}
	
	// Convert to enhanced tasks with project context
	var enhancedTasks []*EnhancedTask
	for _, task := range githubTasks {
		enhancedTask := &EnhancedTask{
			Task:       *task,
			ProjectID:  repoClient.Repository.ProjectID,
			GitURL:     repoClient.Repository.GitURL,
			Repository: repoClient.Repository,
		}
		enhancedTasks = append(enhancedTasks, enhancedTask)
	}
	
	return enhancedTasks, nil
}

// EnhancedTask extends Task with project context
type EnhancedTask struct {
	Task
	ProjectID  int
	GitURL     string
	Repository hive.Repository
}

// filterSuitableTasks filters tasks based on agent capabilities
func (hi *HiveIntegration) filterSuitableTasks(tasks []*EnhancedTask) []*EnhancedTask {
	var suitable []*EnhancedTask
	
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
func (hi *HiveIntegration) claimAndExecuteTask(task *EnhancedTask) {
	hi.repositoryLock.RLock()
	repoClient, exists := hi.repositories[task.ProjectID]
	hi.repositoryLock.RUnlock()
	
	if !exists {
		fmt.Printf("❌ Repository client not found for project %d\n", task.ProjectID)
		return
	}
	
	// Claim the task in GitHub
	claimedTask, err := repoClient.Client.ClaimTask(task.Number, hi.config.AgentID)
	if err != nil {
		fmt.Printf("❌ Failed to claim task %d in %s/%s: %v\n", 
			task.Number, task.Repository.Owner, task.Repository.Repository, err)
		return
	}
	
	fmt.Printf("✋ Claimed task #%d from %s/%s: %s\n", 
		claimedTask.Number, task.Repository.Owner, task.Repository.Repository, claimedTask.Title)
	
	// Report claim to Hive
	if err := hi.hiveClient.ClaimTask(hi.ctx, task.ProjectID, task.Number, hi.config.AgentID); err != nil {
		fmt.Printf("⚠️ Failed to report task claim to Hive: %v\n", err)
	}
	
	// Start task execution
	go hi.executeTask(task, repoClient)
}

// executeTask executes a claimed task with reasoning and coordination
func (hi *HiveIntegration) executeTask(task *EnhancedTask, repoClient *RepositoryClient) {
	fmt.Printf("🚀 Starting execution of task #%d from %s/%s: %s\n", 
		task.Number, task.Repository.Owner, task.Repository.Repository, task.Title)
	
	// Generate execution plan using reasoning
	prompt := fmt.Sprintf("You are an expert AI developer working on a distributed task from repository %s/%s. "+
		"Create a concise, step-by-step plan to resolve this GitHub issue. "+
		"Issue Title: %s. Issue Body: %s. Project Context: %s", 
		task.Repository.Owner, task.Repository.Repository, task.Title, task.Description, task.GitURL)
	
	plan, err := reasoning.GenerateResponse(hi.ctx, "phi3", prompt)
	if err != nil {
		fmt.Printf("❌ Failed to generate execution plan for task #%d: %v\n", task.Number, err)
		return
	}
	
	fmt.Printf("📝 Generated Plan for task #%d:\n%s\n", task.Number, plan)
	
	// Start meta-discussion
	conversationKey := fmt.Sprintf("%d:%d", task.ProjectID, task.Number)
	
	hi.discussionLock.Lock()
	hi.activeDiscussions[conversationKey] = &Conversation{
		TaskID:          task.Number,
		TaskTitle:       task.Title,
		TaskDescription: task.Description,
		History:         []string{fmt.Sprintf("Plan by %s (%s/%s): %s", hi.config.AgentID, task.Repository.Owner, task.Repository.Repository, plan)},
		LastUpdated:     time.Now(),
	}
	hi.discussionLock.Unlock()
	
	// Announce plan for peer review
	metaMsg := map[string]interface{}{
		"project_id":  task.ProjectID,
		"issue_id":    task.Number,
		"repository":  fmt.Sprintf("%s/%s", task.Repository.Owner, task.Repository.Repository),
		"message":     "Here is my proposed plan for this cross-repository task. What are your thoughts?",
		"plan":        plan,
		"git_url":     task.GitURL,
	}
	
	if err := hi.pubsub.PublishAntennaeMessage(pubsub.MetaDiscussion, metaMsg); err != nil {
		fmt.Printf("⚠️ Failed to publish plan to meta-discussion channel: %v\n", err)
	}
}

// handleMetaDiscussion handles incoming meta-discussion messages
func (hi *HiveIntegration) handleMetaDiscussion(msg pubsub.Message, from peer.ID) {
	projectID, hasProject := msg.Data["project_id"].(float64)
	issueID, hasIssue := msg.Data["issue_id"].(float64)
	
	if !hasProject || !hasIssue {
		return
	}
	
	conversationKey := fmt.Sprintf("%d:%d", int(projectID), int(issueID))
	
	hi.discussionLock.Lock()
	convo, exists := hi.activeDiscussions[conversationKey]
	if !exists || convo.IsEscalated {
		hi.discussionLock.Unlock()
		return
	}
	
	incomingMessage, _ := msg.Data["message"].(string)
	repository, _ := msg.Data["repository"].(string)
	
	convo.History = append(convo.History, fmt.Sprintf("Response from %s (%s): %s", from.ShortString(), repository, incomingMessage))
	convo.LastUpdated = time.Now()
	hi.discussionLock.Unlock()
	
	fmt.Printf("🎯 Received peer feedback for task #%d in project %d. Generating response...\n", int(issueID), int(projectID))
	
	// Generate intelligent response
	historyStr := strings.Join(convo.History, "\n")
	prompt := fmt.Sprintf(
		"You are an AI developer collaborating on a distributed task across multiple repositories. "+
		"Repository: %s. Task: %s. Description: %s. "+
		"Conversation history:\n%s\n\n"+
		"Based on the last message, provide a concise and helpful response for cross-repository coordination.",
		repository, convo.TaskTitle, convo.TaskDescription, historyStr,
	)
	
	response, err := reasoning.GenerateResponse(hi.ctx, "phi3", prompt)
	if err != nil {
		fmt.Printf("❌ Failed to generate response for task #%d: %v\n", int(issueID), err)
		return
	}
	
	// Check for escalation
	if hi.shouldEscalate(response, convo.History) {
		fmt.Printf("🚨 Escalating task #%d in project %d for human review.\n", int(issueID), int(projectID))
		convo.IsEscalated = true
		go hi.triggerHumanEscalation(int(projectID), convo, response)
		return
	}
	
	fmt.Printf("💬 Sending response for task #%d in project %d...\n", int(issueID), int(projectID))
	
	responseMsg := map[string]interface{}{
		"project_id": int(projectID),
		"issue_id":   int(issueID),
		"repository": repository,
		"message":    response,
	}
	
	if err := hi.pubsub.PublishAntennaeMessage(pubsub.MetaDiscussion, responseMsg); err != nil {
		fmt.Printf("⚠️ Failed to publish response: %v\n", err)
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