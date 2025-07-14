package test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/deepblackcloud/bzzz/pubsub"
)

// TaskSimulator generates realistic task scenarios for testing antennae coordination
type TaskSimulator struct {
	pubsub       *pubsub.PubSub
	ctx          context.Context
	isRunning    bool
	repositories []MockRepository
	scenarios    []CoordinationScenario
}

// MockRepository represents a simulated repository with tasks
type MockRepository struct {
	Owner       string   `json:"owner"`
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Tasks       []MockTask `json:"tasks"`
	Dependencies []string `json:"dependencies"` // Other repos this depends on
}

// MockTask represents a simulated GitHub issue/task
type MockTask struct {
	Number      int      `json:"number"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Labels      []string `json:"labels"`
	Difficulty  string   `json:"difficulty"`  // easy, medium, hard
	TaskType    string   `json:"task_type"`   // feature, bug, refactor, etc.
	Dependencies []TaskDependency `json:"dependencies"`
	EstimatedHours int   `json:"estimated_hours"`
	RequiredSkills []string `json:"required_skills"`
}

// TaskDependency represents a cross-repository task dependency
type TaskDependency struct {
	Repository string `json:"repository"`
	TaskNumber int    `json:"task_number"`
	DependencyType string `json:"dependency_type"` // api_contract, database_schema, config, security
}

// CoordinationScenario represents a test scenario for antennae coordination
type CoordinationScenario struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Repositories []string `json:"repositories"`
	Tasks       []ScenarioTask `json:"tasks"`
	ExpectedCoordination []string `json:"expected_coordination"`
}

// ScenarioTask links tasks across repositories for coordination testing
type ScenarioTask struct {
	Repository string `json:"repository"`
	TaskNumber int    `json:"task_number"`
	Priority   int    `json:"priority"`
	BlockedBy  []ScenarioTask `json:"blocked_by"`
}

// NewTaskSimulator creates a new task simulator
func NewTaskSimulator(ps *pubsub.PubSub, ctx context.Context) *TaskSimulator {
	sim := &TaskSimulator{
		pubsub: ps,
		ctx:    ctx,
		repositories: generateMockRepositories(),
		scenarios: generateCoordinationScenarios(),
	}
	return sim
}

// Start begins the task simulation
func (ts *TaskSimulator) Start() {
	if ts.isRunning {
		return
	}
	ts.isRunning = true
	
	fmt.Println("🎭 Starting Task Simulator for Antennae Testing")
	
	// Start different simulation routines
	go ts.simulateTaskAnnouncements()
	go ts.simulateCoordinationScenarios()
	go ts.simulateAgentResponses()
}

// Stop stops the task simulation
func (ts *TaskSimulator) Stop() {
	ts.isRunning = false
	fmt.Println("🛑 Task Simulator stopped")
}

// simulateTaskAnnouncements periodically announces available tasks
func (ts *TaskSimulator) simulateTaskAnnouncements() {
	ticker := time.NewTicker(45 * time.Second)
	defer ticker.Stop()
	
	for ts.isRunning {
		select {
		case <-ts.ctx.Done():
			return
		case <-ticker.C:
			ts.announceRandomTask()
		}
	}
}

// announceRandomTask announces a random task from the mock repositories
func (ts *TaskSimulator) announceRandomTask() {
	if len(ts.repositories) == 0 {
		return
	}
	
	repo := ts.repositories[rand.Intn(len(ts.repositories))]
	if len(repo.Tasks) == 0 {
		return
	}
	
	task := repo.Tasks[rand.Intn(len(repo.Tasks))]
	
	announcement := map[string]interface{}{
		"type": "task_available",
		"repository": map[string]interface{}{
			"owner": repo.Owner,
			"name":  repo.Name,
			"url":   repo.URL,
		},
		"task": task,
		"announced_at": time.Now().Unix(),
		"announced_by": "task_simulator",
	}
	
	fmt.Printf("📢 Announcing task: %s/#%d - %s\n", repo.Name, task.Number, task.Title)
	
	if err := ts.pubsub.PublishBzzzMessage(pubsub.TaskAnnouncement, announcement); err != nil {
		fmt.Printf("❌ Failed to announce task: %v\n", err)
	}
}

// simulateCoordinationScenarios runs coordination test scenarios
func (ts *TaskSimulator) simulateCoordinationScenarios() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()
	
	scenarioIndex := 0
	
	for ts.isRunning {
		select {
		case <-ts.ctx.Done():
			return
		case <-ticker.C:
			if len(ts.scenarios) > 0 {
				scenario := ts.scenarios[scenarioIndex%len(ts.scenarios)]
				ts.runCoordinationScenario(scenario)
				scenarioIndex++
			}
		}
	}
}

// runCoordinationScenario executes a specific coordination test scenario
func (ts *TaskSimulator) runCoordinationScenario(scenario CoordinationScenario) {
	fmt.Printf("🎯 Running coordination scenario: %s\n", scenario.Name)
	fmt.Printf("   Description: %s\n", scenario.Description)
	
	// Announce the scenario start
	scenarioStart := map[string]interface{}{
		"type": "coordination_scenario_start",
		"scenario": scenario,
		"started_at": time.Now().Unix(),
	}
	
	if err := ts.pubsub.PublishAntennaeMessage(pubsub.CoordinationRequest, scenarioStart); err != nil {
		fmt.Printf("❌ Failed to announce scenario start: %v\n", err)
		return
	}
	
	// Announce each task in the scenario with dependencies
	for _, task := range scenario.Tasks {
		taskAnnouncement := map[string]interface{}{
			"type": "scenario_task",
			"scenario_name": scenario.Name,
			"repository": task.Repository,
			"task_number": task.TaskNumber,
			"priority": task.Priority,
			"blocked_by": task.BlockedBy,
			"announced_at": time.Now().Unix(),
		}
		
		fmt.Printf("   📋 Task: %s/#%d (Priority: %d)\n", task.Repository, task.TaskNumber, task.Priority)
		
		if err := ts.pubsub.PublishBzzzMessage(pubsub.TaskAnnouncement, taskAnnouncement); err != nil {
			fmt.Printf("❌ Failed to announce scenario task: %v\n", err)
		}
		
		time.Sleep(2 * time.Second) // Stagger announcements
	}
}

// simulateAgentResponses simulates agent responses to create coordination activity
func (ts *TaskSimulator) simulateAgentResponses() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	responses := []string{
		"I can handle this frontend task",
		"This requires database schema changes first",
		"Need to coordinate with API team",
		"This conflicts with my current work",
		"I have the required Python skills",
		"This should be done after the security review",
		"I can start this immediately",
		"This needs clarification on requirements",
	}
	
	for ts.isRunning {
		select {
		case <-ts.ctx.Done():
			return
		case <-ticker.C:
			if rand.Float32() < 0.7 { // 70% chance of response
				response := responses[rand.Intn(len(responses))]
				ts.simulateAgentResponse(response)
			}
		}
	}
}

// simulateAgentResponse simulates an agent response for coordination
func (ts *TaskSimulator) simulateAgentResponse(response string) {
	agentResponse := map[string]interface{}{
		"type": "agent_response",
		"agent_id": fmt.Sprintf("sim-agent-%d", rand.Intn(3)+1),
		"message": response,
		"timestamp": time.Now().Unix(),
		"confidence": rand.Float32()*0.4 + 0.6, // 0.6-1.0 confidence
	}
	
	fmt.Printf("🤖 Simulated agent response: %s\n", response)
	
	if err := ts.pubsub.PublishAntennaeMessage(pubsub.MetaDiscussion, agentResponse); err != nil {
		fmt.Printf("❌ Failed to publish agent response: %v\n", err)
	}
}

// generateMockRepositories creates realistic mock repositories for testing
func generateMockRepositories() []MockRepository {
	return []MockRepository{
		{
			Owner: "deepblackcloud",
			Name:  "hive",
			URL:   "https://github.com/deepblackcloud/hive",
			Dependencies: []string{"bzzz", "distributed-ai-dev"},
			Tasks: []MockTask{
				{
					Number: 15,
					Title: "Add WebSocket support for real-time coordination",
					Description: "Implement WebSocket endpoints for real-time agent coordination",
					Labels: []string{"bzzz-task", "feature", "realtime"},
					Difficulty: "medium",
					TaskType: "feature",
					EstimatedHours: 8,
					RequiredSkills: []string{"websockets", "python", "fastapi"},
					Dependencies: []TaskDependency{
						{Repository: "bzzz", TaskNumber: 23, DependencyType: "api_contract"},
					},
				},
				{
					Number: 16,
					Title: "Implement agent authentication system",
					Description: "Add secure authentication for bzzz agents",
					Labels: []string{"bzzz-task", "security", "auth"},
					Difficulty: "hard",
					TaskType: "security",
					EstimatedHours: 12,
					RequiredSkills: []string{"security", "jwt", "python"},
				},
			},
		},
		{
			Owner: "deepblackcloud",
			Name:  "bzzz",
			URL:   "https://github.com/deepblackcloud/bzzz",
			Dependencies: []string{"hive"},
			Tasks: []MockTask{
				{
					Number: 23,
					Title: "Define coordination API contract",
					Description: "Standardize API contract for cross-repository coordination",
					Labels: []string{"bzzz-task", "api", "coordination"},
					Difficulty: "medium",
					TaskType: "api_design",
					EstimatedHours: 6,
					RequiredSkills: []string{"api_design", "go", "documentation"},
				},
				{
					Number: 24,
					Title: "Implement dependency detection algorithm",
					Description: "Auto-detect task dependencies across repositories",
					Labels: []string{"bzzz-task", "algorithm", "coordination"},
					Difficulty: "hard",
					TaskType: "feature",
					EstimatedHours: 16,
					RequiredSkills: []string{"algorithms", "go", "graph_theory"},
				},
			},
		},
		{
			Owner: "deepblackcloud",
			Name:  "distributed-ai-dev",
			URL:   "https://github.com/deepblackcloud/distributed-ai-dev",
			Dependencies: []string{},
			Tasks: []MockTask{
				{
					Number: 8,
					Title: "Add support for bzzz coordination",
					Description: "Integrate with bzzz P2P coordination system",
					Labels: []string{"bzzz-task", "integration", "p2p"},
					Difficulty: "medium",
					TaskType: "integration",
					EstimatedHours: 10,
					RequiredSkills: []string{"p2p", "python", "integration"},
					Dependencies: []TaskDependency{
						{Repository: "bzzz", TaskNumber: 23, DependencyType: "api_contract"},
						{Repository: "hive", TaskNumber: 16, DependencyType: "security"},
					},
				},
			},
		},
	}
}

// generateCoordinationScenarios creates test scenarios for coordination
func generateCoordinationScenarios() []CoordinationScenario {
	return []CoordinationScenario{
		{
			Name: "Cross-Repository API Integration",
			Description: "Testing coordination when multiple repos need to implement a shared API",
			Repositories: []string{"hive", "bzzz", "distributed-ai-dev"},
			Tasks: []ScenarioTask{
				{Repository: "bzzz", TaskNumber: 23, Priority: 1, BlockedBy: []ScenarioTask{}},
				{Repository: "hive", TaskNumber: 15, Priority: 2, BlockedBy: []ScenarioTask{{Repository: "bzzz", TaskNumber: 23}}},
				{Repository: "distributed-ai-dev", TaskNumber: 8, Priority: 3, BlockedBy: []ScenarioTask{{Repository: "bzzz", TaskNumber: 23}, {Repository: "hive", TaskNumber: 16}}},
			},
			ExpectedCoordination: []string{
				"API contract should be defined first",
				"Authentication system blocks integration work",
				"WebSocket implementation depends on API contract",
			},
		},
		{
			Name: "Security-First Development",
			Description: "Testing coordination when security requirements block other work",
			Repositories: []string{"hive", "distributed-ai-dev"},
			Tasks: []ScenarioTask{
				{Repository: "hive", TaskNumber: 16, Priority: 1, BlockedBy: []ScenarioTask{}},
				{Repository: "distributed-ai-dev", TaskNumber: 8, Priority: 2, BlockedBy: []ScenarioTask{{Repository: "hive", TaskNumber: 16}}},
			},
			ExpectedCoordination: []string{
				"Security authentication must be completed first",
				"Integration work blocked until auth system ready",
			},
		},
		{
			Name: "Parallel Development Conflict",
			Description: "Testing coordination when agents might work on conflicting tasks",
			Repositories: []string{"hive", "bzzz"},
			Tasks: []ScenarioTask{
				{Repository: "hive", TaskNumber: 15, Priority: 1, BlockedBy: []ScenarioTask{}},
				{Repository: "bzzz", TaskNumber: 24, Priority: 1, BlockedBy: []ScenarioTask{}},
			},
			ExpectedCoordination: []string{
				"Both tasks modify coordination logic",
				"Need to coordinate implementation approach",
			},
		},
	}
}

// GetMockRepositories returns the mock repositories for external use
func (ts *TaskSimulator) GetMockRepositories() []MockRepository {
	return ts.repositories
}

// GetScenarios returns the coordination scenarios for external use
func (ts *TaskSimulator) GetScenarios() []CoordinationScenario {
	return ts.scenarios
}

// PrintStatus prints the current simulation status
func (ts *TaskSimulator) PrintStatus() {
	fmt.Printf("🎭 Task Simulator Status:\n")
	fmt.Printf("   Running: %v\n", ts.isRunning)
	fmt.Printf("   Mock Repositories: %d\n", len(ts.repositories))
	fmt.Printf("   Coordination Scenarios: %d\n", len(ts.scenarios))
	
	totalTasks := 0
	for _, repo := range ts.repositories {
		totalTasks += len(repo.Tasks)
	}
	fmt.Printf("   Total Mock Tasks: %d\n", totalTasks)
}