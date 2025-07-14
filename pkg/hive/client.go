package hive

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HiveClient provides integration with the Hive task coordination system
type HiveClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewHiveClient creates a new Hive API client
func NewHiveClient(baseURL, apiKey string) *HiveClient {
	return &HiveClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Repository represents a Git repository configuration from Hive
type Repository struct {
	ProjectID            int    `json:"project_id"`
	Name                 string `json:"name"`
	GitURL               string `json:"git_url"`
	Owner                string `json:"owner"`
	Repository           string `json:"repository"`
	Branch               string `json:"branch"`
	BzzzEnabled          bool   `json:"bzzz_enabled"`
	ReadyToClaim         bool   `json:"ready_to_claim"`
	PrivateRepo          bool   `json:"private_repo"`
	GitHubTokenRequired  bool   `json:"github_token_required"`
}

// ActiveRepositoriesResponse represents the response from /api/bzzz/active-repos
type ActiveRepositoriesResponse struct {
	Repositories []Repository `json:"repositories"`
}

// TaskClaimRequest represents a task claim request to Hive
type TaskClaimRequest struct {
	TaskNumber int    `json:"task_number"`
	AgentID    string `json:"agent_id"`
	ClaimedAt  int64  `json:"claimed_at"`
}

// TaskStatusUpdate represents a task status update to Hive
type TaskStatusUpdate struct {
	Status    string                 `json:"status"`
	UpdatedAt int64                  `json:"updated_at"`
	Results   map[string]interface{} `json:"results,omitempty"`
}

// GetActiveRepositories fetches all repositories marked for Bzzz consumption
func (c *HiveClient) GetActiveRepositories(ctx context.Context) ([]Repository, error) {
	url := fmt.Sprintf("%s/api/bzzz/active-repos", c.BaseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add authentication if API key is provided
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var response ActiveRepositoriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return response.Repositories, nil
}

// GetProjectTasks fetches bzzz-task labeled issues for a specific project
func (c *HiveClient) GetProjectTasks(ctx context.Context, projectID int) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/bzzz/projects/%d/tasks", c.BaseURL, projectID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var tasks []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return tasks, nil
}

// ClaimTask registers a task claim with the Hive system
func (c *HiveClient) ClaimTask(ctx context.Context, projectID, taskID int, agentID string) error {
	url := fmt.Sprintf("%s/api/bzzz/projects/%d/claim", c.BaseURL, projectID)
	
	claimRequest := TaskClaimRequest{
		TaskNumber: taskID,
		AgentID:    agentID,
		ClaimedAt:  time.Now().Unix(),
	}
	
	jsonData, err := json.Marshal(claimRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal claim request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("claim request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// UpdateTaskStatus updates the task status in the Hive system
func (c *HiveClient) UpdateTaskStatus(ctx context.Context, projectID, taskID int, status string, results map[string]interface{}) error {
	url := fmt.Sprintf("%s/api/bzzz/projects/%d/status", c.BaseURL, projectID)
	
	statusUpdate := TaskStatusUpdate{
		Status:    status,
		UpdatedAt: time.Now().Unix(),
		Results:   results,
	}
	
	jsonData, err := json.Marshal(statusUpdate)
	if err != nil {
		return fmt.Errorf("failed to marshal status update: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status update failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// HealthCheck verifies connectivity to the Hive API
func (c *HiveClient) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.BaseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Hive API health check failed with status: %d", resp.StatusCode)
	}
	
	return nil
}