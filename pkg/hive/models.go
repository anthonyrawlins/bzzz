package hive

import "time"

// Project represents a project managed by the Hive system
type Project struct {
	ID                  int                    `json:"id"`
	Name                string                 `json:"name"`
	Description         string                 `json:"description"`
	Status              string                 `json:"status"`
	GitURL              string                 `json:"git_url"`
	Owner               string                 `json:"owner"`
	Repository          string                 `json:"repository"`
	Branch              string                 `json:"branch"`
	BzzzEnabled         bool                   `json:"bzzz_enabled"`
	ReadyToClaim        bool                   `json:"ready_to_claim"`
	PrivateRepo         bool                   `json:"private_repo"`
	GitHubTokenRequired bool                   `json:"github_token_required"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// Task represents a task (GitHub issue) from the Hive system
type Task struct {
	ID          int                    `json:"id"`
	ProjectID   int                    `json:"project_id"`
	ProjectName string                 `json:"project_name"`
	GitURL      string                 `json:"git_url"`
	Owner       string                 `json:"owner"`
	Repository  string                 `json:"repository"`
	Branch      string                 `json:"branch"`
	
	// GitHub issue fields
	IssueNumber int    `json:"issue_number"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`
	Assignee    string `json:"assignee,omitempty"`
	
	// Task metadata
	TaskType     string                 `json:"task_type"`
	Priority     int                    `json:"priority"`
	Labels       []string               `json:"labels"`
	Requirements []string               `json:"requirements,omitempty"`
	Deliverables []string               `json:"deliverables,omitempty"`
	Context      map[string]interface{} `json:"context,omitempty"`
	
	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TaskClaim represents a task claim in the Hive system
type TaskClaim struct {
	ID        int       `json:"id"`
	ProjectID int       `json:"project_id"`
	TaskID    int       `json:"task_id"`
	AgentID   string    `json:"agent_id"`
	Status    string    `json:"status"` // claimed, in_progress, completed, failed
	ClaimedAt time.Time `json:"claimed_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Results   map[string]interface{} `json:"results,omitempty"`
}

// ProjectActivationRequest represents a request to activate/deactivate a project
type ProjectActivationRequest struct {
	BzzzEnabled  bool `json:"bzzz_enabled"`
	ReadyToClaim bool `json:"ready_to_claim"`
}

// ProjectRegistrationRequest represents a request to register a new project
type ProjectRegistrationRequest struct {
	Name                string `json:"name"`
	Description         string `json:"description"`
	GitURL              string `json:"git_url"`
	PrivateRepo         bool   `json:"private_repo"`
	BzzzEnabled         bool   `json:"bzzz_enabled"`
	AutoActivate        bool   `json:"auto_activate"`
}

// AgentCapability represents an agent's capabilities for task matching
type AgentCapability struct {
	AgentID      string   `json:"agent_id"`
	NodeID       string   `json:"node_id"`
	Capabilities []string `json:"capabilities"`
	Models       []string `json:"models"`
	Status       string   `json:"status"`
	LastSeen     time.Time `json:"last_seen"`
}

// CoordinationEvent represents a P2P coordination event
type CoordinationEvent struct {
	EventID     string                 `json:"event_id"`
	ProjectID   int                    `json:"project_id"`
	TaskID      int                    `json:"task_id"`
	EventType   string                 `json:"event_type"` // task_claimed, plan_proposed, escalated, completed
	AgentID     string                 `json:"agent_id"`
	Message     string                 `json:"message"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// ErrorResponse represents an error response from the Hive API
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// HealthStatus represents the health status of the Hive system
type HealthStatus struct {
	Status   string    `json:"status"`
	Version  string    `json:"version"`
	Database string    `json:"database"`
	Uptime   string    `json:"uptime"`
	CheckedAt time.Time `json:"checked_at"`
}