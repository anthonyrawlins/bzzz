package github

import (
	"context"
	"fmt"
	"time"

	"github.com/deepblackcloud/bzzz/pubsub"
	"github.com/deepblackcloud/bzzz/reasoning" // Import the new reasoning module
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

// executeTask is where the agent performs the work for a claimed task.
// It now includes a reasoning step to form a plan before execution.
func (i *Integration) executeTask(task *Task) {
	fmt.Printf("🚀 Starting execution of task #%d: %s\n", task.Number, task.Title)

	// Announce that the task has started
	progressData := map[string]interface{}{
		"task_id":   task.Number,
		"agent_id":  i.config.AgentID,
		"status":    "started",
		"timestamp": time.Now().Unix(),
	}
	if err := i.pubsub.PublishBzzzMessage(pubsub.TaskProgress, progressData); err != nil {
		fmt.Printf("⚠️ Failed to announce task start: %v\n", err)
	}

	// === REASONING STEP ===
	// Use Ollama to generate a plan based on the task's title and body.
	fmt.Printf("🧠 Reasoning about task #%d to form a plan...\n", task.Number)
	
	// Construct the prompt for the reasoning model
	prompt := fmt.Sprintf("You are an expert AI developer. Based on the following GitHub issue, create a concise, step-by-step plan to resolve it. Issue Title: %s. Issue Body: %s.", task.Title, task.Body)
	
	// Select a model (this could be made more dynamic later)
	model := "phi3" 

	// Generate the plan using the reasoning module
	plan, err := reasoning.GenerateResponse(i.ctx, model, prompt)
	if err != nil {
		fmt.Printf("❌ Failed to generate execution plan for task #%d: %v\n", task.Number, err)
		// Announce failure and return
		// (Error handling logic would be more robust in a real implementation)
		return
	}

	fmt.Printf("📝 Generated Plan for task #%d:\n%s\n", task.Number, plan)

	// === META-DISCUSSION STEP ===
	// Announce the plan on the Antennae channel for peer review.
	metaMsg := map[string]interface{}{
		"issue_id":  task.Number,
		"node_id":   i.config.AgentID,
		"message":   "Here is my proposed plan of action. Please review and provide feedback within the next 60 seconds.",
		"plan":      plan,
		"msg_id":    fmt.Sprintf("plan-%d-%d", task.Number, time.Now().UnixNano()),
		"parent_id": nil,
		"hop_count": 1,
		"timestamp": time.Now().Unix(),
	}

	if err := i.pubsub.PublishAntennaeMessage(pubsub.MetaDiscussion, metaMsg); err != nil {
		fmt.Printf("⚠️ Failed to publish plan to meta-discussion channel: %v\n", err)
	}

	// Wait for a short "objection period"
	fmt.Println("⏳ Waiting 60 seconds for peer feedback on the plan...")
	time.Sleep(60 * time.Second)
	// In a full implementation, this would listen for responses on the meta-discussion topic.

	// For now, we assume the plan is good and proceed.
	fmt.Printf("✅ Plan for task #%d approved. Proceeding with execution.\n", task.Number)

	// Complete the task on GitHub
	results := map[string]interface{}{
		"status":        "completed",
		"execution_plan": plan,
		"agent_id":      i.config.AgentID,
		"deliverables":  []string{"Plan generated and approved", "Execution logic would run here"},
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