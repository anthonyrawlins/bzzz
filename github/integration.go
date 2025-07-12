package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/deepblackcloud/bzzz/pubsub"
	"github.com/deepblackcloud/bzzz/reasoning"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	// humanEscalationWebhookURL is the N8N webhook for escalating tasks.
	humanEscalationWebhookURL = "https://n8n.home.deepblack.cloud/webhook-test/human-escalation"
	// conversationHistoryLimit is the number of messages before auto-escalation.
	conversationHistoryLimit = 10
)

var escalationKeywords = []string{"stuck", "help", "human", "escalate", "clarification needed", "manual intervention"}

// Conversation represents the history of a discussion for a task.
type Conversation struct {
	TaskID          int
	TaskTitle       string
	TaskDescription string
	History         []string
	LastUpdated     time.Time
	IsEscalated     bool
}

// Integration handles the integration between GitHub tasks and Bzzz P2P coordination
type Integration struct {
	client *Client
	pubsub *pubsub.PubSub
	ctx    context.Context
	config *IntegrationConfig

	activeDiscussions map[int]*Conversation
	discussionLock    sync.RWMutex
}

// IntegrationConfig holds configuration for GitHub-Bzzz integration
type IntegrationConfig struct {
	PollInterval time.Duration
	MaxTasks     int
	AgentID      string
	Capabilities []string
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
	i.pubsub.SetAntennaeMessageHandler(i.handleMetaDiscussion)
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

	go i.executeTask(claimedTask)
	return nil
}

// executeTask starts the task by generating and proposing a plan.
func (i *Integration) executeTask(task *Task) {
	fmt.Printf("🚀 Starting execution of task #%d: %s\n", task.Number, task.Title)

	prompt := fmt.Sprintf("You are an expert AI developer. Based on the following GitHub issue, create a concise, step-by-step plan to resolve it. Issue Title: %s. Issue Body: %s.", task.Title, task.Description)
	plan, err := reasoning.GenerateResponse(i.ctx, "phi3", prompt)
	if err != nil {
		fmt.Printf("❌ Failed to generate execution plan for task #%d: %v\n", task.Number, err)
		return
	}
	fmt.Printf("📝 Generated Plan for task #%d:\n%s\n", task.Number, plan)

	i.discussionLock.Lock()
	i.activeDiscussions[task.Number] = &Conversation{
		TaskID:          task.Number,
		TaskTitle:       task.Title,
		TaskDescription: task.Description,
		History:         []string{fmt.Sprintf("Plan by %s: %s", i.config.AgentID, plan)},
		LastUpdated:     time.Now(),
	}
	i.discussionLock.Unlock()

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
		return
	}
	taskID := int(issueID)

	i.discussionLock.Lock()
	convo, exists := i.activeDiscussions[taskID]
	if !exists || convo.IsEscalated {
		i.discussionLock.Unlock()
		return
	}

	incomingMessage, _ := msg.Data["message"].(string)
	convo.History = append(convo.History, fmt.Sprintf("Response from %s: %s", from.ShortString(), incomingMessage))
	convo.LastUpdated = time.Now()
	i.discussionLock.Unlock()

	fmt.Printf("🎯 Received peer feedback for task #%d. Reasoning about a response...\n", taskID)

	historyStr := strings.Join(convo.History, "\n")
	prompt := fmt.Sprintf(
		"You are an AI developer collaborating on a task. "+
			"This is the original task: Title: %s, Body: %s. "+
			"This is the conversation so far:\n%s\n\n"+
			"Based on the last message, provide a concise and helpful response.",
		convo.TaskTitle, convo.TaskDescription, historyStr,
	)

	response, err := reasoning.GenerateResponse(i.ctx, "phi3", prompt)
	if err != nil {
		fmt.Printf("❌ Failed to generate response for task #%d: %v\n", taskID, err)
		return
	}

	// Check if the situation requires human intervention
	if i.shouldEscalate(response, convo.History) {
		fmt.Printf("🚨 Escalating task #%d for human review.\n", taskID)
		convo.IsEscalated = true
		go i.triggerHumanEscalation(convo, response)
		return
	}

	fmt.Printf("💬 Sending response for task #%d...\n", taskID)
	responseMsg := map[string]interface{}{
		"issue_id": taskID,
		"message":  response,
	}
	if err := i.pubsub.PublishAntennaeMessage(pubsub.MetaDiscussion, responseMsg); err != nil {
		fmt.Printf("⚠️ Failed to publish response for task #%d: %v\n", taskID, err)
	}
}

// shouldEscalate determines if a task needs human intervention.
func (i *Integration) shouldEscalate(response string, history []string) bool {
	// Rule 1: Check for keywords in the latest response
	lowerResponse := strings.ToLower(response)
	for _, keyword := range escalationKeywords {
		if strings.Contains(lowerResponse, keyword) {
			return true
		}
	}

	// Rule 2: Check if the conversation is too long
	if len(history) >= conversationHistoryLimit {
		return true
	}

	return false
}

// triggerHumanEscalation sends the conversation details to the N8N webhook.
func (i *Integration) triggerHumanEscalation(convo *Conversation, reason string) {
	// 1. Announce the escalation to other agents
	escalationMsg := map[string]interface{}{
		"issue_id": convo.TaskID,
		"message":  "This task has been escalated for human review. No further automated action will be taken.",
		"reason":   reason,
	}
	if err := i.pubsub.PublishAntennaeMessage(pubsub.MetaDiscussion, escalationMsg); err != nil {
		fmt.Printf("⚠️ Failed to publish escalation message for task #%d: %v\n", convo.TaskID, err)
	}

	// 2. Send the payload to the N8N webhook
	payload := map[string]interface{}{
		"task_id":          convo.TaskID,
		"task_title":       convo.TaskTitle,
		"escalation_agent": i.config.AgentID,
		"reason":           reason,
		"history":          strings.Join(convo.History, "\n"),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("❌ Failed to marshal escalation payload for task #%d: %v\n", convo.TaskID, err)
		return
	}

	req, err := http.NewRequestWithContext(i.ctx, "POST", humanEscalationWebhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Printf("❌ Failed to create escalation request for task #%d: %v\n", convo.TaskID, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("❌ Failed to send escalation webhook for task #%d: %v\n", convo.TaskID, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		fmt.Printf("⚠️ Human escalation webhook for task #%d returned non-2xx status: %d\n", convo.TaskID, resp.StatusCode)
	} else {
		fmt.Printf("✅ Successfully escalated task #%d to human administrator.\n", convo.TaskID)
	}
}

// Unchanged functions
func (i *Integration) filterSuitableTasks(tasks []*Task) []*Task { return tasks }
func (i *Integration) canHandleTaskType(taskType string) bool   { return true }
func (i *Integration) announceTaskClaim(task *Task) error       { return nil }