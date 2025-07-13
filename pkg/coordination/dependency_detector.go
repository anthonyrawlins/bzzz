package coordination

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/deepblackcloud/bzzz/pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

// DependencyDetector analyzes tasks across repositories for relationships
type DependencyDetector struct {
	pubsub            *pubsub.PubSub
	ctx               context.Context
	knownTasks        map[string]*TaskContext // taskKey -> context
	dependencyRules   []DependencyRule
	coordinationHops  int
}

// TaskContext represents a task with its repository and project context
type TaskContext struct {
	TaskID      int    `json:"task_id"`
	ProjectID   int    `json:"project_id"`
	Repository  string `json:"repository"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Keywords    []string `json:"keywords"`
	AgentID     string `json:"agent_id"`
	ClaimedAt   time.Time `json:"claimed_at"`
}

// DependencyRule defines how to detect task relationships
type DependencyRule struct {
	Name        string
	Description string
	Keywords    []string
	Validator   func(task1, task2 *TaskContext) (bool, string)
}

// TaskDependency represents a detected relationship between tasks
type TaskDependency struct {
	Task1       *TaskContext `json:"task1"`
	Task2       *TaskContext `json:"task2"`
	Relationship string       `json:"relationship"`
	Confidence   float64      `json:"confidence"`
	Reason       string       `json:"reason"`
	DetectedAt   time.Time    `json:"detected_at"`
}

// NewDependencyDetector creates a new cross-repository dependency detector
func NewDependencyDetector(ctx context.Context, ps *pubsub.PubSub) *DependencyDetector {
	dd := &DependencyDetector{
		pubsub:           ps,
		ctx:              ctx,
		knownTasks:       make(map[string]*TaskContext),
		coordinationHops: 3, // Limit meta discussion depth
	}
	
	// Initialize common dependency detection rules
	dd.initializeDependencyRules()
	
	// Subscribe to task announcements for dependency detection
	go dd.listenForTaskAnnouncements()
	
	return dd
}

// initializeDependencyRules sets up common patterns for task relationships
func (dd *DependencyDetector) initializeDependencyRules() {
	dd.dependencyRules = []DependencyRule{
		{
			Name:        "API_Contract",
			Description: "Tasks involving API contracts and implementations",
			Keywords:    []string{"api", "endpoint", "contract", "interface", "schema"},
			Validator: func(task1, task2 *TaskContext) (bool, string) {
				// Check if one task defines API and another implements it
				text1 := strings.ToLower(task1.Title + " " + task1.Description)
				text2 := strings.ToLower(task2.Title + " " + task2.Description)
				
				if (strings.Contains(text1, "api") && strings.Contains(text2, "implement")) ||
				   (strings.Contains(text2, "api") && strings.Contains(text1, "implement")) {
					return true, "API definition and implementation dependency"
				}
				return false, ""
			},
		},
		{
			Name:        "Database_Schema",
			Description: "Database schema changes affecting multiple services",
			Keywords:    []string{"database", "schema", "migration", "table", "model"},
			Validator: func(task1, task2 *TaskContext) (bool, string) {
				text1 := strings.ToLower(task1.Title + " " + task1.Description)
				text2 := strings.ToLower(task2.Title + " " + task2.Description)
				
				dbKeywords := []string{"database", "schema", "migration", "table"}
				hasDB1 := false
				hasDB2 := false
				
				for _, keyword := range dbKeywords {
					if strings.Contains(text1, keyword) { hasDB1 = true }
					if strings.Contains(text2, keyword) { hasDB2 = true }
				}
				
				if hasDB1 && hasDB2 {
					return true, "Database schema dependency detected"
				}
				return false, ""
			},
		},
		{
			Name:        "Configuration_Dependency",
			Description: "Configuration changes affecting multiple components",
			Keywords:    []string{"config", "environment", "settings", "parameters"},
			Validator: func(task1, task2 *TaskContext) (bool, string) {
				text1 := strings.ToLower(task1.Title + " " + task1.Description)
				text2 := strings.ToLower(task2.Title + " " + task2.Description)
				
				if (strings.Contains(text1, "config") || strings.Contains(text1, "environment")) &&
				   (strings.Contains(text2, "config") || strings.Contains(text2, "environment")) {
					return true, "Configuration dependency - coordinated changes needed"
				}
				return false, ""
			},
		},
		{
			Name:        "Security_Compliance",
			Description: "Security changes requiring coordinated implementation",
			Keywords:    []string{"security", "auth", "permission", "token", "encrypt"},
			Validator: func(task1, task2 *TaskContext) (bool, string) {
				text1 := strings.ToLower(task1.Title + " " + task1.Description)
				text2 := strings.ToLower(task2.Title + " " + task2.Description)
				
				secKeywords := []string{"security", "auth", "permission", "token"}
				hasSecu1 := false
				hasSecu2 := false
				
				for _, keyword := range secKeywords {
					if strings.Contains(text1, keyword) { hasSecu1 = true }
					if strings.Contains(text2, keyword) { hasSecu2 = true }
				}
				
				if hasSecu1 && hasSecu2 {
					return true, "Security implementation requires coordination"
				}
				return false, ""
			},
		},
	}
}

// RegisterTask adds a task to the dependency tracking system
func (dd *DependencyDetector) RegisterTask(task *TaskContext) {
	taskKey := fmt.Sprintf("%d:%d", task.ProjectID, task.TaskID)
	dd.knownTasks[taskKey] = task
	
	fmt.Printf("🔍 Registered task for dependency detection: %s/%s #%d\n", 
		task.Repository, task.Title, task.TaskID)
	
	// Check for dependencies with existing tasks
	dd.detectDependencies(task)
}

// detectDependencies analyzes a new task against existing tasks for relationships
func (dd *DependencyDetector) detectDependencies(newTask *TaskContext) {
	for _, existingTask := range dd.knownTasks {
		// Skip self-comparison
		if existingTask.TaskID == newTask.TaskID && existingTask.ProjectID == newTask.ProjectID {
			continue
		}
		
		// Skip if same repository (handled by single-repo coordination)
		if existingTask.Repository == newTask.Repository {
			continue
		}
		
		// Apply dependency detection rules
		for _, rule := range dd.dependencyRules {
			if matches, reason := rule.Validator(newTask, existingTask); matches {
				dependency := &TaskDependency{
					Task1:        newTask,
					Task2:        existingTask,
					Relationship: rule.Name,
					Confidence:   0.8, // Could be improved with ML
					Reason:       reason,
					DetectedAt:   time.Now(),
				}
				
				dd.announceDependency(dependency)
			}
		}
	}
}

// announceDependency broadcasts a detected dependency for agent coordination
func (dd *DependencyDetector) announceDependency(dep *TaskDependency) {
	fmt.Printf("🔗 Dependency detected: %s/%s #%d ↔ %s/%s #%d (%s)\n",
		dep.Task1.Repository, dep.Task1.Title, dep.Task1.TaskID,
		dep.Task2.Repository, dep.Task2.Title, dep.Task2.TaskID,
		dep.Relationship)
	
	// Create coordination message for Antennae meta-discussion
	coordMsg := map[string]interface{}{
		"message_type":   "dependency_detected",
		"dependency":     dep,
		"coordination_request": fmt.Sprintf(
			"Cross-repository dependency detected between tasks. "+
			"Agent working on %s/%s #%d and agent working on %s/%s #%d should coordinate. "+
			"Relationship: %s. Reason: %s",
			dep.Task1.Repository, dep.Task1.Title, dep.Task1.TaskID,
			dep.Task2.Repository, dep.Task2.Title, dep.Task2.TaskID,
			dep.Relationship, dep.Reason,
		),
		"agents_involved": []string{dep.Task1.AgentID, dep.Task2.AgentID},
		"repositories":    []string{dep.Task1.Repository, dep.Task2.Repository},
		"hop_count":       0,
		"max_hops":        dd.coordinationHops,
		"detected_at":     dep.DetectedAt.Unix(),
	}
	
	// Publish to Antennae meta-discussion channel
	if err := dd.pubsub.PublishAntennaeMessage(pubsub.MetaDiscussion, coordMsg); err != nil {
		fmt.Printf("❌ Failed to announce dependency: %v\n", err)
	} else {
		fmt.Printf("📡 Dependency coordination request sent to Antennae channel\n")
	}
}

// listenForTaskAnnouncements monitors the P2P mesh for task claims
func (dd *DependencyDetector) listenForTaskAnnouncements() {
	// This would integrate with the existing pubsub system
	// to automatically detect when agents claim tasks
	fmt.Printf("👂 Dependency detector listening for task announcements...\n")
	
	// In a real implementation, this would subscribe to TaskClaim messages
	// and extract task context for dependency analysis
}

// GetKnownTasks returns all tasks currently being tracked
func (dd *DependencyDetector) GetKnownTasks() map[string]*TaskContext {
	return dd.knownTasks
}

// GetDependencyRules returns the configured dependency detection rules
func (dd *DependencyDetector) GetDependencyRules() []DependencyRule {
	return dd.dependencyRules
}

// AddCustomRule allows adding project-specific dependency detection
func (dd *DependencyDetector) AddCustomRule(rule DependencyRule) {
	dd.dependencyRules = append(dd.dependencyRules, rule)
	fmt.Printf("➕ Added custom dependency rule: %s\n", rule.Name)
}