package github

import (
	"context"
	"fmt"
	"time"

	"github.com/deepblackcloud/bzzz/pubsub"
)

// Integration handles the integration between GitHub tasks and Bzzz P2P coordination
type Integration struct {
	client *Client
	pubsub *pubsub.PubSub
	ctx    context.Context
	config *IntegrationConfig
}

// IntegrationConfig holds configuration for GitHub-Bzzz integration
type IntegrationConfig struct {
	PollInterval  time.Duration // How often to check for new tasks
	MaxTasks      int           // Maximum tasks to process simultaneously
	AgentID       string        // This agent's identifier
	Capabilities  []string      // What types of tasks this agent can handle
}

// NewIntegration creates a new GitHub-Bzzz integration
func NewIntegration(ctx context.Context, client *Client, ps *pubsub.PubSub, config *IntegrationConfig) *Integration {
	if config.PollInterval == 0 {
		config.PollInterval = 30 * time.Second
	}
	if config.MaxTasks == 0 {
		config.MaxTasks = 3
	}
	
	return &Integration{
		client: client,
		pubsub: ps,
		ctx:    ctx,
		config: config,
	}
}

// Start begins the GitHub-Bzzz integration
func (i *Integration) Start() {
	fmt.Printf("🔗 Starting GitHub-Bzzz integration for agent: %s\n", i.config.AgentID)
	
	// Start task polling
	go i.pollForTasks()
	
	// Start listening for P2P task announcements
	go i.listenForTaskAnnouncements()
}

// pollForTasks periodically checks GitHub for available tasks
func (i *Integration) pollForTasks() {
	ticker := time.NewTicker(i.config.PollInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-i.ctx.Done():
			return
		case <-ticker.C:
			if err := i.checkAndClaimTasks(); err != nil {
				fmt.Printf("❌ Error checking tasks: %v\n", err)
			}
		}
	}
}

// checkAndClaimTasks looks for available tasks and claims suitable ones
func (i *Integration) checkAndClaimTasks() error {
	// Get available tasks
	tasks, err := i.client.ListAvailableTasks()
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}
	
	if len(tasks) == 0 {
		return nil
	}
	
	fmt.Printf("📋 Found %d available tasks\n", len(tasks))
	
	// Filter tasks based on capabilities
	suitableTasks := i.filterSuitableTasks(tasks)
	
	if len(suitableTasks) == 0 {
		fmt.Printf("⚠️ No suitable tasks for agent capabilities: %v\n", i.config.Capabilities)
		return nil
	}
	
	// Claim the highest priority suitable task
	task := suitableTasks[0] // Assuming sorted by priority
	claimedTask, err := i.client.ClaimTask(task.Number, i.config.AgentID)
	if err != nil {
		return fmt.Errorf("failed to claim task %d: %w", task.Number, err)
	}
	
	fmt.Printf("✋ Claimed task #%d: %s\n", claimedTask.Number, claimedTask.Title)
	
	// Announce the claim over P2P
	if err := i.announceTaskClaim(claimedTask); err != nil {
		fmt.Printf("⚠️ Failed to announce task claim: %v\n", err)
	}
	
	// Start working on the task
	go i.executeTask(claimedTask)
	
	return nil
}

// filterSuitableTasks filters tasks based on agent capabilities
func (i *Integration) filterSuitableTasks(tasks []*Task) []*Task {
	suitable := make([]*Task, 0)
	
	for _, task := range tasks {
		// Check if this agent can handle this task type
		if i.canHandleTaskType(task.TaskType) {
			suitable = append(suitable, task)
		}
	}
	
	return suitable
}

// canHandleTaskType checks if this agent can handle the given task type
func (i *Integration) canHandleTaskType(taskType string) bool {
	for _, capability := range i.config.Capabilities {
		if capability == taskType || capability == "general" {
			return true
		}
	}
	return false
}

// announceTaskClaim announces a task claim over the P2P network
func (i *Integration) announceTaskClaim(task *Task) error {
	data := map[string]interface{}{
		"task_id":     task.Number,
		"task_title":  task.Title,
		"task_type":   task.TaskType,
		"agent_id":    i.config.AgentID,
		"claimed_at":  time.Now().Unix(),
		"github_url":  fmt.Sprintf("https://github.com/%s/%s/issues/%d", 
			i.client.config.Owner, i.client.config.Repository, task.Number),
	}
	
	return i.pubsub.PublishBzzzMessage(pubsub.TaskClaim, data)
}

// executeTask simulates task execution
func (i *Integration) executeTask(task *Task) {
	fmt.Printf("🚀 Starting execution of task #%d: %s\n", task.Number, task.Title)
	
	// Announce task progress
	progressData := map[string]interface{}{
		"task_id":   task.Number,
		"agent_id":  i.config.AgentID,
		"status":    "started",
		"timestamp": time.Now().Unix(),
	}
	
	if err := i.pubsub.PublishBzzzMessage(pubsub.TaskProgress, progressData); err != nil {
		fmt.Printf("⚠️ Failed to announce task progress: %v\n", err)
	}
	
	// Simulate work (in a real implementation, this would be actual task execution)
	workDuration := time.Duration(30+task.Priority*10) * time.Second
	fmt.Printf("⏳ Working on task for %v...\n", workDuration)
	time.Sleep(workDuration)
	
	// Complete the task
	results := map[string]interface{}{
		"status":        "completed",
		"execution_time": workDuration.String(),
		"agent_id":      i.config.AgentID,
		"deliverables":  []string{"Implementation completed", "Tests passed", "Documentation updated"},
	}
	
	if err := i.client.CompleteTask(task.Number, i.config.AgentID, results); err != nil {
		fmt.Printf("❌ Failed to complete task #%d: %v\n", task.Number, err)
		return
	}
	
	// Announce completion over P2P
	completionData := map[string]interface{}{
		"task_id":     task.Number,
		"agent_id":    i.config.AgentID,
		"completed_at": time.Now().Unix(),
		"results":     results,
	}
	
	if err := i.pubsub.PublishBzzzMessage(pubsub.TaskComplete, completionData); err != nil {
		fmt.Printf("⚠️ Failed to announce task completion: %v\n", err)
	}
	
	fmt.Printf("✅ Completed task #%d: %s\n", task.Number, task.Title)
}

// listenForTaskAnnouncements listens for task announcements from other agents
func (i *Integration) listenForTaskAnnouncements() {
	// This would integrate with the pubsub message handlers
	// For now, it's a placeholder that demonstrates the pattern
	fmt.Printf("👂 Listening for task announcements from other agents...\n")
}