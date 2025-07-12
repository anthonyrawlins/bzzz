package github

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/deepblackcloud/bzzz/pubsub"
	"github.com/deepblackcloud/bzzz/reasoning"
	"github.com/libp2p/go-libp2p/core/peer"
)

// Conversation represents the history of a discussion for a task.
type Conversation struct {
	TaskID          int
	TaskTitle       string
	TaskDescription string
	History         []string
	LastUpdated     time.Time
}

// Integration handles the integration between GitHub tasks and Bzzz P2P coordination
type Integration struct {
	client *Client
	pubsub *pubsub.PubSub
	ctx    context.Context
	config *IntegrationConfig

	// activeDiscussions stores the conversation history for each task.
	activeDiscussions map[int]*Conversation
	discussionLock    sync.RWMutex
}

// IntegrationConfig holds configuration for GitHub-Bzzz integration
type IntegrationConfig struct {
	PollInterval time.Duration // How often to check for new tasks
	MaxTasks     int           // Maximum tasks to process simultaneously
	AgentID      string        // This agent's identifier
	Capabilities []string      // What types of tasks this agent can handle
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
		client:            client,
		pubsub:            ps,
		ctx:               ctx,
		config:            config,
		activeDiscussions: make(map[int]*Conversation),
	}
}

// Start begins the GitHub-Bzzz integration
func (i *Integration) Start() {
	fmt.Printf("🔗 Starting GitHub-Bzzz integration for agent: %s\n", i.config.AgentID)

	// Register the handler for incoming meta-discussion messages
	i.pubsub.SetAntennaeMessageHandler(i.handleMetaDiscussion)

	// Start task polling
	go i.pollForTasks()
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
	tasks, err := i.client.ListAvailableTasks()
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}
	if len(tasks) == 0 {
		return nil
	}
	fmt.Printf("📋 Found %d available tasks\n", len(tasks))

	suitableTasks := i.filterSuitableTasks(tasks)
	if len(suitableTasks) == 0 {
		return nil
	}

	task := suitableTasks[0]
	claimedTask, err := i.client.ClaimTask(task.Number, i.config.AgentID)
	if err != nil {
		return fmt.Errorf("failed to claim task %d: %w", task.Number, err)
	}
	fmt.Printf("✋ Claimed task #%d: %s\n", claimedTask.Number, claimedTask.Title)

	if err := i.announceTaskClaim(claimedTask); err != nil {
		fmt.Printf("⚠️ Failed to announce task claim: %v\n", err)
	}

	go i.executeTask(claimedTask)
	return nil
}

// filterSuitableTasks filters tasks based on agent capabilities
func (i *Integration) filterSuitableTasks(tasks []*Task) []*Task {
	// (Implementation is unchanged)
	return tasks
}

// canHandleTaskType checks if this agent can handle the given task type
func (i *Integration) canHandleTaskType(taskType string) bool {
	// (Implementation is unchanged)
	return true
}

// announceTaskClaim announces a task claim over the P2P network
func (i *Integration) announceTaskClaim(task *Task) error {
	// (Implementation is unchanged)
	return nil
}

// executeTask starts the task by generating and proposing a plan.
func (i *Integration) executeTask(task *Task) {
	fmt.Printf("🚀 Starting execution of task #%d: %s\n", task.Number, task.Title)

	// === REASONING STEP ===
	fmt.Printf("🧠 Reasoning about task #%d to form a plan...\n", task.Number)
	prompt := fmt.Sprintf("You are an expert AI developer. Based on the following GitHub issue, create a concise, step-by-step plan to resolve it. Issue Title: %s. Issue Body: %s.", task.Title, task.Description)
	model := "phi3"

	plan, err := reasoning.GenerateResponse(i.ctx, model, prompt)
	if err != nil {
		fmt.Printf("❌ Failed to generate execution plan for task #%d: %v\n", task.Number, err)
		return
	}
	fmt.Printf("📝 Generated Plan for task #%d:\n%s\n", task.Number, plan)

	// === META-DISCUSSION STEP ===
	// Store the initial state of the conversation
	i.discussionLock.Lock()
	i.activeDiscussions[task.Number] = &Conversation{
		TaskID:          task.Number,
		TaskTitle:       task.Title,
		TaskDescription: task.Description,
		History:         []string{fmt.Sprintf("Plan by %s: %s", i.config.AgentID, plan)},
		LastUpdated:     time.Now(),
	}
	i.discussionLock.Unlock()

	// Announce the plan on the Antennae channel
	metaMsg := map[string]interface{}{
		"issue_id":  task.Number,
		"message":   "Here is my proposed plan of action. What are your thoughts?",
		"plan":      plan,
	}
	if err := i.pubsub.PublishAntennaeMessage(pubsub.MetaDiscussion, metaMsg); err != nil {
		fmt.Printf("⚠️ Failed to publish plan to meta-discussion channel: %v\n", err)
	}
}

// handleMetaDiscussion is the core handler for incoming Antennae messages.
func (i *Integration) handleMetaDiscussion(msg pubsub.Message, from peer.ID) {
	issueID, ok := msg.Data["issue_id"].(float64)
	if !ok {
		fmt.Printf("⚠️ Received meta-discussion message with invalid issue_id\n")
		return
	}
	taskID := int(issueID)

	i.discussionLock.Lock()
	convo, exists := i.activeDiscussions[taskID]
	if !exists {
		i.discussionLock.Unlock()
		// We are not involved in this conversation, so we ignore it.
		return
	}

	// Append the new message to the history
	incomingMessage, _ := msg.Data["message"].(string)
	convo.History = append(convo.History, fmt.Sprintf("Response from %s: %s", from.ShortString(), incomingMessage))
	convo.LastUpdated = time.Now()
	i.discussionLock.Unlock()

	fmt.Printf("🎯 Received peer feedback for task #%d. Reasoning about a response...\n", taskID)

	// === REASONING STEP (RESPONSE) ===
	// Construct a prompt with the full conversation history
	historyStr := strings.Join(convo.History, "\n")
	prompt := fmt.Sprintf(
		"You are an AI developer collaborating on a task. "+
			"This is the original task: Title: %s, Body: %s. "+
			"This is the conversation so far:\n%s\n\n"+
			"Based on the last message, provide a concise and helpful response.",
		convo.TaskTitle, convo.TaskDescription, historyStr,
	)
	model := "phi3"

	response, err := reasoning.GenerateResponse(i.ctx, model, prompt)
	if err != nil {
		fmt.Printf("❌ Failed to generate response for task #%d: %v\n", taskID, err)
		return
	}

	fmt.Printf("💬 Sending response for task #%d...\n", taskID)

	// Publish the response
	responseMsg := map[string]interface{}{
		"issue_id": taskID,
		"message":  response,
	}
	if err := i.pubsub.PublishAntennaeMessage(pubsub.MetaDiscussion, responseMsg); err != nil {
		fmt.Printf("⚠️ Failed to publish response for task #%d: %v\n", taskID, err)
	}
}
