package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/deepblackcloud/bzzz/pubsub"
)

// AntennaeMonitor tracks and logs antennae coordination activity
type AntennaeMonitor struct {
	ctx            context.Context
	pubsub         *pubsub.PubSub
	logFile        *os.File
	metricsFile    *os.File
	activeSessions map[string]*CoordinationSession
	metrics        *CoordinationMetrics
	mu             sync.RWMutex
	isRunning      bool
}

// CoordinationSession tracks an active coordination session
type CoordinationSession struct {
	SessionID       string                 `json:"session_id"`
	StartTime       time.Time              `json:"start_time"`
	LastActivity    time.Time              `json:"last_activity"`
	Repositories    []string               `json:"repositories"`
	Tasks           []string               `json:"tasks"`
	Participants    []string               `json:"participants"`
	Messages        []CoordinationMessage  `json:"messages"`
	Dependencies    []TaskDependency       `json:"dependencies"`
	Status          string                 `json:"status"` // active, completed, escalated, failed
	Outcome         map[string]interface{} `json:"outcome"`
}

// CoordinationMessage represents a message in the coordination session
type CoordinationMessage struct {
	Timestamp   time.Time              `json:"timestamp"`
	FromAgent   string                 `json:"from_agent"`
	MessageType string                 `json:"message_type"`
	Content     map[string]interface{} `json:"content"`
	Topic       string                 `json:"topic"`
}

// TaskDependency represents a detected task dependency
type TaskDependency struct {
	Repository     string `json:"repository"`
	TaskNumber     int    `json:"task_number"`
	DependsOn      string `json:"depends_on"`
	DependencyType string `json:"dependency_type"`
	DetectedAt     time.Time `json:"detected_at"`
}

// CoordinationMetrics tracks quantitative coordination data
type CoordinationMetrics struct {
	StartTime               time.Time `json:"start_time"`
	TotalSessions          int       `json:"total_sessions"`
	ActiveSessions         int       `json:"active_sessions"`
	CompletedSessions      int       `json:"completed_sessions"`
	EscalatedSessions      int       `json:"escalated_sessions"`
	FailedSessions         int       `json:"failed_sessions"`
	TotalMessages          int       `json:"total_messages"`
	TaskAnnouncements      int       `json:"task_announcements"`
	DependenciesDetected   int       `json:"dependencies_detected"`
	AgentParticipations    map[string]int `json:"agent_participations"`
	AverageSessionDuration time.Duration  `json:"average_session_duration"`
	LastUpdated            time.Time `json:"last_updated"`
}

// NewAntennaeMonitor creates a new antennae monitoring system
func NewAntennaeMonitor(ctx context.Context, ps *pubsub.PubSub, logDir string) (*AntennaeMonitor, error) {
	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log files
	timestamp := time.Now().Format("20060102_150405")
	logPath := filepath.Join(logDir, fmt.Sprintf("antennae_activity_%s.jsonl", timestamp))
	metricsPath := filepath.Join(logDir, fmt.Sprintf("antennae_metrics_%s.json", timestamp))

	logFile, err := os.Create(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create activity log file: %w", err)
	}

	metricsFile, err := os.Create(metricsPath)
	if err != nil {
		logFile.Close()
		return nil, fmt.Errorf("failed to create metrics file: %w", err)
	}

	monitor := &AntennaeMonitor{
		ctx:            ctx,
		pubsub:         ps,
		logFile:        logFile,
		metricsFile:    metricsFile,
		activeSessions: make(map[string]*CoordinationSession),
		metrics: &CoordinationMetrics{
			StartTime:            time.Now(),
			AgentParticipations: make(map[string]int),
		},
	}

	fmt.Printf("📊 Antennae Monitor initialized\n")
	fmt.Printf("   Activity Log: %s\n", logPath)
	fmt.Printf("   Metrics File: %s\n", metricsPath)

	return monitor, nil
}

// Start begins monitoring antennae coordination activity
func (am *AntennaeMonitor) Start() {
	if am.isRunning {
		return
	}
	am.isRunning = true

	fmt.Println("🔍 Starting Antennae coordination monitoring...")

	// Start monitoring routines
	go am.monitorCoordinationMessages()
	go am.monitorTaskAnnouncements()
	go am.periodicMetricsUpdate()
	go am.sessionCleanup()
}

// Stop stops the monitoring system
func (am *AntennaeMonitor) Stop() {
	if !am.isRunning {
		return
	}
	am.isRunning = false

	// Save final metrics
	am.saveMetrics()

	// Close files
	if am.logFile != nil {
		am.logFile.Close()
	}
	if am.metricsFile != nil {
		am.metricsFile.Close()
	}

	fmt.Println("🛑 Antennae monitoring stopped")
}

// monitorCoordinationMessages listens for antennae meta-discussion messages
func (am *AntennaeMonitor) monitorCoordinationMessages() {
	// Subscribe to antennae topic
	msgChan := make(chan pubsub.Message, 100)
	
	// This would be implemented with actual pubsub subscription
	// For now, we'll simulate receiving messages
	
	for am.isRunning {
		select {
		case <-am.ctx.Done():
			return
		case msg := <-msgChan:
			am.processCoordinationMessage(msg)
		case <-time.After(1 * time.Second):
			// Continue monitoring
		}
	}
}

// monitorTaskAnnouncements listens for task announcements
func (am *AntennaeMonitor) monitorTaskAnnouncements() {
	// Subscribe to bzzz coordination topic
	msgChan := make(chan pubsub.Message, 100)
	
	for am.isRunning {
		select {
		case <-am.ctx.Done():
			return
		case msg := <-msgChan:
			am.processTaskAnnouncement(msg)
		case <-time.After(1 * time.Second):
			// Continue monitoring
		}
	}
}

// processCoordinationMessage processes an antennae coordination message
func (am *AntennaeMonitor) processCoordinationMessage(msg pubsub.Message) {
	am.mu.Lock()
	defer am.mu.Unlock()

	coordMsg := CoordinationMessage{
		Timestamp:   time.Now(),
		FromAgent:   msg.From,
		MessageType: msg.Type,
		Content:     msg.Data,
		Topic:       "antennae/meta-discussion",
	}

	// Log the message
	am.logActivity("coordination_message", coordMsg)

	// Update metrics
	am.metrics.TotalMessages++
	am.metrics.AgentParticipations[msg.From]++

	// Determine session ID (could be extracted from message content)
	sessionID := am.extractSessionID(msg.Data)
	
	// Get or create session
	session := am.getOrCreateSession(sessionID)
	session.LastActivity = time.Now()
	session.Messages = append(session.Messages, coordMsg)
	
	// Add participant if new
	if !contains(session.Participants, msg.From) {
		session.Participants = append(session.Participants, msg.From)
	}

	// Update session status based on message type
	am.updateSessionStatus(session, msg)

	fmt.Printf("🧠 Antennae message: %s from %s (Session: %s)\n", 
		msg.Type, msg.From, sessionID)
}

// processTaskAnnouncement processes a task announcement
func (am *AntennaeMonitor) processTaskAnnouncement(msg pubsub.Message) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Log the announcement
	am.logActivity("task_announcement", msg.Data)

	// Update metrics
	am.metrics.TaskAnnouncements++

	// Extract task information
	if repo, ok := msg.Data["repository"].(map[string]interface{}); ok {
		if repoName, ok := repo["name"].(string); ok {
			fmt.Printf("📢 Task announced: %s\n", repoName)
			
			// Check for dependencies and create coordination session if needed
			if task, ok := msg.Data["task"].(map[string]interface{}); ok {
				if deps, ok := task["dependencies"].([]interface{}); ok && len(deps) > 0 {
					sessionID := fmt.Sprintf("coord_%d", time.Now().Unix())
					session := am.getOrCreateSession(sessionID)
					session.Repositories = append(session.Repositories, repoName)
					
					fmt.Printf("🔗 Dependencies detected, creating coordination session: %s\n", sessionID)
				}
			}
		}
	}
}

// getOrCreateSession gets an existing session or creates a new one
func (am *AntennaeMonitor) getOrCreateSession(sessionID string) *CoordinationSession {
	if session, exists := am.activeSessions[sessionID]; exists {
		return session
	}

	session := &CoordinationSession{
		SessionID:    sessionID,
		StartTime:    time.Now(),
		LastActivity: time.Now(),
		Status:       "active",
		Messages:     make([]CoordinationMessage, 0),
		Repositories: make([]string, 0),
		Tasks:        make([]string, 0),
		Participants: make([]string, 0),
		Dependencies: make([]TaskDependency, 0),
	}

	am.activeSessions[sessionID] = session
	am.metrics.TotalSessions++
	am.metrics.ActiveSessions++

	fmt.Printf("🆕 New coordination session created: %s\n", sessionID)
	return session
}

// updateSessionStatus updates session status based on message content
func (am *AntennaeMonitor) updateSessionStatus(session *CoordinationSession, msg pubsub.Message) {
	// Analyze message content to determine status changes
	if content, ok := msg.Data["type"].(string); ok {
		switch content {
		case "consensus_reached":
			session.Status = "completed"
			am.metrics.ActiveSessions--
			am.metrics.CompletedSessions++
		case "escalation_triggered":
			session.Status = "escalated"
			am.metrics.ActiveSessions--
			am.metrics.EscalatedSessions++
		case "coordination_failed":
			session.Status = "failed"
			am.metrics.ActiveSessions--
			am.metrics.FailedSessions++
		}
	}
}

// periodicMetricsUpdate saves metrics periodically
func (am *AntennaeMonitor) periodicMetricsUpdate() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for am.isRunning {
		select {
		case <-am.ctx.Done():
			return
		case <-ticker.C:
			am.saveMetrics()
			am.printStatus()
		}
	}
}

// sessionCleanup removes old inactive sessions
func (am *AntennaeMonitor) sessionCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for am.isRunning {
		select {
		case <-am.ctx.Done():
			return
		case <-ticker.C:
			am.cleanupOldSessions()
		}
	}
}

// cleanupOldSessions removes sessions inactive for more than 10 minutes
func (am *AntennaeMonitor) cleanupOldSessions() {
	am.mu.Lock()
	defer am.mu.Unlock()

	cutoff := time.Now().Add(-10 * time.Minute)
	cleaned := 0

	for sessionID, session := range am.activeSessions {
		if session.LastActivity.Before(cutoff) && session.Status == "active" {
			session.Status = "timeout"
			delete(am.activeSessions, sessionID)
			am.metrics.ActiveSessions--
			am.metrics.FailedSessions++
			cleaned++
		}
	}

	if cleaned > 0 {
		fmt.Printf("🧹 Cleaned up %d inactive sessions\n", cleaned)
	}
}

// logActivity logs an activity to the activity log file
func (am *AntennaeMonitor) logActivity(activityType string, data interface{}) {
	logEntry := map[string]interface{}{
		"timestamp":     time.Now().Unix(),
		"activity_type": activityType,
		"data":          data,
	}

	if jsonBytes, err := json.Marshal(logEntry); err == nil {
		am.logFile.WriteString(string(jsonBytes) + "\n")
		am.logFile.Sync()
	}
}

// saveMetrics saves current metrics to file
func (am *AntennaeMonitor) saveMetrics() {
	am.mu.RLock()
	defer am.mu.RUnlock()

	am.metrics.LastUpdated = time.Now()
	
	// Calculate average session duration
	if am.metrics.CompletedSessions > 0 {
		totalDuration := time.Duration(0)
		completed := 0
		
		for _, session := range am.activeSessions {
			if session.Status == "completed" {
				totalDuration += session.LastActivity.Sub(session.StartTime)
				completed++
			}
		}
		
		if completed > 0 {
			am.metrics.AverageSessionDuration = totalDuration / time.Duration(completed)
		}
	}

	if jsonBytes, err := json.MarshalIndent(am.metrics, "", "  "); err == nil {
		am.metricsFile.Seek(0, 0)
		am.metricsFile.Truncate(0)
		am.metricsFile.Write(jsonBytes)
		am.metricsFile.Sync()
	}
}

// printStatus prints current monitoring status
func (am *AntennaeMonitor) printStatus() {
	am.mu.RLock()
	defer am.mu.RUnlock()

	fmt.Printf("📊 Antennae Monitor Status:\n")
	fmt.Printf("   Total Sessions: %d (Active: %d, Completed: %d)\n", 
		am.metrics.TotalSessions, am.metrics.ActiveSessions, am.metrics.CompletedSessions)
	fmt.Printf("   Messages: %d, Announcements: %d\n", 
		am.metrics.TotalMessages, am.metrics.TaskAnnouncements)
	fmt.Printf("   Dependencies Detected: %d\n", am.metrics.DependenciesDetected)
	fmt.Printf("   Active Participants: %d\n", len(am.metrics.AgentParticipations))
}

// GetMetrics returns current metrics
func (am *AntennaeMonitor) GetMetrics() *CoordinationMetrics {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.metrics
}

// Helper functions
func (am *AntennaeMonitor) extractSessionID(data map[string]interface{}) string {
	if sessionID, ok := data["session_id"].(string); ok {
		return sessionID
	}
	if scenarioName, ok := data["scenario_name"].(string); ok {
		return fmt.Sprintf("scenario_%s", scenarioName)
	}
	return fmt.Sprintf("session_%d", time.Now().Unix())
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}