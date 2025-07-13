package coordination

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/deepblackcloud/bzzz/pubsub"
	"github.com/deepblackcloud/bzzz/reasoning"
	"github.com/libp2p/go-libp2p/core/peer"
)

// MetaCoordinator manages advanced cross-repository coordination
type MetaCoordinator struct {
	pubsub               *pubsub.PubSub
	ctx                  context.Context
	dependencyDetector   *DependencyDetector
	
	// Active coordination sessions
	activeSessions       map[string]*CoordinationSession // sessionID -> session
	sessionLock          sync.RWMutex
	
	// Configuration
	maxSessionDuration   time.Duration
	maxParticipants      int
	escalationThreshold  int
}

// CoordinationSession represents an active multi-agent coordination
type CoordinationSession struct {
	SessionID           string                 `json:"session_id"`
	Type                string                 `json:"type"` // dependency, conflict, planning
	Participants        map[string]*Participant `json:"participants"`
	TasksInvolved       []*TaskContext         `json:"tasks_involved"`
	Messages            []CoordinationMessage  `json:"messages"`
	Status              string                 `json:"status"` // active, resolved, escalated
	CreatedAt           time.Time              `json:"created_at"`
	LastActivity        time.Time              `json:"last_activity"`
	Resolution          string                 `json:"resolution,omitempty"`
	EscalationReason    string                 `json:"escalation_reason,omitempty"`
}

// Participant represents an agent in a coordination session
type Participant struct {
	AgentID      string    `json:"agent_id"`
	PeerID       string    `json:"peer_id"`
	Repository   string    `json:"repository"`
	Capabilities []string  `json:"capabilities"`
	LastSeen     time.Time `json:"last_seen"`
	Active       bool      `json:"active"`
}

// CoordinationMessage represents a message in a coordination session
type CoordinationMessage struct {
	MessageID   string                 `json:"message_id"`
	FromAgentID string                 `json:"from_agent_id"`
	FromPeerID  string                 `json:"from_peer_id"`
	Content     string                 `json:"content"`
	MessageType string                 `json:"message_type"` // proposal, question, agreement, concern
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewMetaCoordinator creates a new meta coordination system
func NewMetaCoordinator(ctx context.Context, ps *pubsub.PubSub) *MetaCoordinator {
	mc := &MetaCoordinator{
		pubsub:              ps,
		ctx:                 ctx,
		activeSessions:      make(map[string]*CoordinationSession),
		maxSessionDuration:  30 * time.Minute,
		maxParticipants:     5,
		escalationThreshold: 10, // Max messages before escalation consideration
	}
	
	// Initialize dependency detector
	mc.dependencyDetector = NewDependencyDetector(ctx, ps)
	
	// Set up message handler for meta-discussions
	ps.SetAntennaeMessageHandler(mc.handleMetaMessage)
	
	// Start session management
	go mc.sessionCleanupLoop()
	
	fmt.Printf("🎯 Advanced Meta Coordinator initialized\n")
	return mc
}

// handleMetaMessage processes incoming Antennae meta-discussion messages
func (mc *MetaCoordinator) handleMetaMessage(msg pubsub.Message, from peer.ID) {
	messageType, hasType := msg.Data[\"message_type\"].(string)
	if !hasType {
		return // Not a coordination message
	}
	
	switch messageType {
	case \"dependency_detected\":
		mc.handleDependencyDetection(msg, from)
	case \"coordination_request\":
		mc.handleCoordinationRequest(msg, from)
	case \"coordination_response\":
		mc.handleCoordinationResponse(msg, from)
	case \"session_message\":
		mc.handleSessionMessage(msg, from)
	case \"escalation_request\":
		mc.handleEscalationRequest(msg, from)
	default:
		// Handle as general meta-discussion
		mc.handleGeneralDiscussion(msg, from)
	}
}

// handleDependencyDetection creates a coordination session for detected dependencies
func (mc *MetaCoordinator) handleDependencyDetection(msg pubsub.Message, from peer.ID) {
	dependency, hasDep := msg.Data[\"dependency\"]
	if !hasDep {
		return
	}
	
	// Parse dependency information
	depBytes, _ := json.Marshal(dependency)
	var dep TaskDependency
	if err := json.Unmarshal(depBytes, &dep); err != nil {
		fmt.Printf(\"❌ Failed to parse dependency: %v\\n\", err)
		return
	}
	
	// Create coordination session
	sessionID := fmt.Sprintf(\"dep_%d_%d_%d\", dep.Task1.ProjectID, dep.Task1.TaskID, time.Now().Unix())
	
	session := &CoordinationSession{
		SessionID:     sessionID,
		Type:          \"dependency\",
		Participants:  make(map[string]*Participant),
		TasksInvolved: []*TaskContext{dep.Task1, dep.Task2},
		Messages:      []CoordinationMessage{},
		Status:        \"active\",
		CreatedAt:     time.Now(),
		LastActivity:  time.Now(),
	}
	
	// Add participants
	session.Participants[dep.Task1.AgentID] = &Participant{
		AgentID:    dep.Task1.AgentID,
		Repository: dep.Task1.Repository,
		LastSeen:   time.Now(),
		Active:     true,
	}
	session.Participants[dep.Task2.AgentID] = &Participant{
		AgentID:    dep.Task2.AgentID,
		Repository: dep.Task2.Repository,
		LastSeen:   time.Now(),
		Active:     true,
	}
	
	mc.sessionLock.Lock()
	mc.activeSessions[sessionID] = session
	mc.sessionLock.Unlock()
	
	fmt.Printf(\"🎯 Created coordination session %s for dependency: %s\\n\", sessionID, dep.Relationship)
	
	// Generate coordination plan
	mc.generateCoordinationPlan(session, &dep)
}

// generateCoordinationPlan creates an AI-generated plan for coordination
func (mc *MetaCoordinator) generateCoordinationPlan(session *CoordinationSession, dep *TaskDependency) {
	prompt := fmt.Sprintf(`
You are an expert AI project coordinator managing a distributed development team.

SITUATION:
- A dependency has been detected between two tasks in different repositories
- Task 1: %s/%s #%d (Agent: %s)
- Task 2: %s/%s #%d (Agent: %s) 
- Relationship: %s
- Reason: %s

COORDINATION REQUIRED:
Generate a concise coordination plan that addresses:
1. What specific coordination is needed between the agents
2. What order should tasks be completed in (if any)
3. What information/artifacts need to be shared
4. What potential conflicts to watch for
5. Success criteria for coordinated completion

Keep the plan practical and actionable. Focus on specific next steps.`,
		dep.Task1.Repository, dep.Task1.Title, dep.Task1.TaskID, dep.Task1.AgentID,
		dep.Task2.Repository, dep.Task2.Title, dep.Task2.TaskID, dep.Task2.AgentID,
		dep.Relationship, dep.Reason)
	
	plan, err := reasoning.GenerateResponse(mc.ctx, \"phi3\", prompt)
	if err != nil {
		fmt.Printf(\"❌ Failed to generate coordination plan: %v\\n\", err)
		return
	}
	
	// Create initial coordination message
	coordMessage := CoordinationMessage{
		MessageID:   fmt.Sprintf(\"plan_%d\", time.Now().Unix()),
		FromAgentID: \"meta_coordinator\",
		FromPeerID:  \"system\",
		Content:     plan,
		MessageType: \"proposal\",
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			\"session_id\": session.SessionID,
			\"plan_type\":  \"coordination\",
		},
	}
	
	session.Messages = append(session.Messages, coordMessage)
	
	// Broadcast coordination plan to participants
	mc.broadcastToSession(session, map[string]interface{}{
		\"message_type\":    \"coordination_plan\",
		\"session_id\":      session.SessionID,
		\"plan\":            plan,
		\"tasks_involved\":  session.TasksInvolved,
		\"participants\":    session.Participants,
		\"message\":         fmt.Sprintf(\"Coordination plan generated for dependency: %s\", dep.Relationship),
	})
	
	fmt.Printf(\"📋 Generated and broadcasted coordination plan for session %s\\n\", session.SessionID)
}

// broadcastToSession sends a message to all participants in a session
func (mc *MetaCoordinator) broadcastToSession(session *CoordinationSession, data map[string]interface{}) {
	if err := mc.pubsub.PublishAntennaeMessage(pubsub.MetaDiscussion, data); err != nil {
		fmt.Printf(\"❌ Failed to broadcast to session %s: %v\\n\", session.SessionID, err)
	}
}

// handleCoordinationResponse processes responses from agents in coordination
func (mc *MetaCoordinator) handleCoordinationResponse(msg pubsub.Message, from peer.ID) {
	sessionID, hasSession := msg.Data[\"session_id\"].(string)
	if !hasSession {
		return
	}
	
	mc.sessionLock.RLock()
	session, exists := mc.activeSessions[sessionID]
	mc.sessionLock.RUnlock()
	
	if !exists || session.Status != \"active\" {
		return
	}
	
	agentResponse, hasResponse := msg.Data[\"response\"].(string)
	agentID, hasAgent := msg.Data[\"agent_id\"].(string)
	
	if !hasResponse || !hasAgent {
		return
	}
	
	// Update participant activity
	if participant, exists := session.Participants[agentID]; exists {
		participant.LastSeen = time.Now()
		participant.PeerID = from.ShortString()
	}
	
	// Add message to session
	coordMessage := CoordinationMessage{
		MessageID:   fmt.Sprintf(\"resp_%s_%d\", agentID, time.Now().Unix()),
		FromAgentID: agentID,
		FromPeerID:  from.ShortString(),
		Content:     agentResponse,
		MessageType: \"response\",
		Timestamp:   time.Now(),
	}
	
	session.Messages = append(session.Messages, coordMessage)
	session.LastActivity = time.Now()
	
	fmt.Printf(\"💬 Coordination response from %s in session %s\\n\", agentID, sessionID)
	
	// Check if coordination is complete
	mc.evaluateSessionProgress(session)
}

// evaluateSessionProgress determines if a session needs escalation or can be resolved
func (mc *MetaCoordinator) evaluateSessionProgress(session *CoordinationSession) {
	// Check for escalation conditions
	if len(session.Messages) >= mc.escalationThreshold {
		mc.escalateSession(session, \"Message limit exceeded - human intervention needed\")
		return
	}
	
	if time.Since(session.CreatedAt) > mc.maxSessionDuration {
		mc.escalateSession(session, \"Session duration exceeded - human intervention needed\")
		return
	}
	
	// Check for agreement keywords in recent messages
	recentMessages := session.Messages
	if len(recentMessages) > 3 {
		recentMessages = session.Messages[len(session.Messages)-3:]
	}
	
	agreementCount := 0
	for _, msg := range recentMessages {
		content := strings.ToLower(msg.Content)
		if strings.Contains(content, \"agree\") || strings.Contains(content, \"sounds good\") ||
		   strings.Contains(content, \"approved\") || strings.Contains(content, \"looks good\") {
			agreementCount++
		}
	}
	
	// If majority agreement, consider resolved
	if agreementCount >= len(session.Participants)-1 {
		mc.resolveSession(session, \"Consensus reached among participants\")
	}
}

// escalateSession escalates a session to human intervention
func (mc *MetaCoordinator) escalateSession(session *CoordinationSession, reason string) {
	session.Status = \"escalated\"
	session.EscalationReason = reason
	
	fmt.Printf(\"🚨 Escalating coordination session %s: %s\\n\", session.SessionID, reason)
	
	// Create escalation message
	escalationData := map[string]interface{}{
		\"message_type\":       \"escalation\",
		\"session_id\":         session.SessionID,
		\"escalation_reason\":  reason,
		\"session_summary\":    mc.generateSessionSummary(session),
		\"participants\":       session.Participants,
		\"tasks_involved\":     session.TasksInvolved,
		\"requires_human\":     true,
	}
	
	mc.broadcastToSession(session, escalationData)
}

// resolveSession marks a session as successfully resolved
func (mc *MetaCoordinator) resolveSession(session *CoordinationSession, resolution string) {
	session.Status = \"resolved\"
	session.Resolution = resolution
	
	fmt.Printf(\"✅ Resolved coordination session %s: %s\\n\", session.SessionID, resolution)
	
	// Broadcast resolution
	resolutionData := map[string]interface{}{
		\"message_type\": \"resolution\",
		\"session_id\":   session.SessionID,
		\"resolution\":   resolution,
		\"summary\":      mc.generateSessionSummary(session),
	}
	
	mc.broadcastToSession(session, resolutionData)
}

// generateSessionSummary creates a summary of the coordination session
func (mc *MetaCoordinator) generateSessionSummary(session *CoordinationSession) string {
	return fmt.Sprintf(
		\"Session %s (%s): %d participants, %d messages, duration %v\",
		session.SessionID, session.Type, len(session.Participants),
		len(session.Messages), time.Since(session.CreatedAt).Round(time.Minute))
}

// sessionCleanupLoop removes old inactive sessions
func (mc *MetaCoordinator) sessionCleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-mc.ctx.Done():
			return
		case <-ticker.C:
			mc.cleanupInactiveSessions()
		}
	}
}

// cleanupInactiveSessions removes sessions that are old or resolved
func (mc *MetaCoordinator) cleanupInactiveSessions() {
	mc.sessionLock.Lock()
	defer mc.sessionLock.Unlock()
	
	for sessionID, session := range mc.activeSessions {
		// Remove sessions older than 2 hours or already resolved/escalated
		if time.Since(session.LastActivity) > 2*time.Hour || 
		   session.Status == \"resolved\" || session.Status == \"escalated\" {
			delete(mc.activeSessions, sessionID)
			fmt.Printf(\"🧹 Cleaned up session %s (status: %s)\\n\", sessionID, session.Status)
		}
	}
}

// handleGeneralDiscussion processes general meta-discussion messages
func (mc *MetaCoordinator) handleGeneralDiscussion(msg pubsub.Message, from peer.ID) {
	// Handle non-coordination meta discussions
	fmt.Printf(\"💭 General meta-discussion from %s: %v\\n\", from.ShortString(), msg.Data)
}

// GetActiveSessions returns current coordination sessions
func (mc *MetaCoordinator) GetActiveSessions() map[string]*CoordinationSession {
	mc.sessionLock.RLock()
	defer mc.sessionLock.RUnlock()
	
	sessions := make(map[string]*CoordinationSession)
	for k, v := range mc.activeSessions {
		sessions[k] = v
	}
	return sessions
}

// handleSessionMessage processes messages within coordination sessions
func (mc *MetaCoordinator) handleSessionMessage(msg pubsub.Message, from peer.ID) {
	sessionID, hasSession := msg.Data[\"session_id\"].(string)
	if !hasSession {
		return
	}
	
	mc.sessionLock.RLock()
	session, exists := mc.activeSessions[sessionID]
	mc.sessionLock.RUnlock()
	
	if !exists {
		return
	}
	
	session.LastActivity = time.Now()
	fmt.Printf(\"📨 Session message in %s from %s\\n\", sessionID, from.ShortString())
}

// handleCoordinationRequest processes requests to start coordination
func (mc *MetaCoordinator) handleCoordinationRequest(msg pubsub.Message, from peer.ID) {
	fmt.Printf(\"🎯 Coordination request from %s\\n\", from.ShortString())
	// Implementation for handling coordination requests
}

// handleEscalationRequest processes escalation requests
func (mc *MetaCoordinator) handleEscalationRequest(msg pubsub.Message, from peer.ID) {
	fmt.Printf(\"🚨 Escalation request from %s\\n\", from.ShortString())
	// Implementation for handling escalation requests
}