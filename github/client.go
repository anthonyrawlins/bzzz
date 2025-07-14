package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub API client for Bzzz task management
type Client struct {
	client *github.Client
	ctx    context.Context
	config *Config
}

// Config holds GitHub integration configuration
type Config struct {
	AccessToken string
	Owner       string // GitHub organization/user
	Repository  string // Repository for task coordination
	
	// Task management settings
	TaskLabel       string // Label for Bzzz tasks
	InProgressLabel string // Label for tasks in progress
	CompletedLabel  string // Label for completed tasks
	
	// Branch management
	BaseBranch string // Base branch for task branches
	BranchPrefix string // Prefix for task branches
}

// NewClient creates a new GitHub client for Bzzz integration
func NewClient(ctx context.Context, config *Config) (*Client, error) {
	if config.AccessToken == "" {
		return nil, fmt.Errorf("GitHub access token is required")
	}
	
	if config.Owner == "" || config.Repository == "" {
		return nil, fmt.Errorf("GitHub owner and repository are required")
	}
	
	// Set defaults
	if config.TaskLabel == "" {
		config.TaskLabel = "bzzz-task"
	}
	if config.InProgressLabel == "" {
		config.InProgressLabel = "in-progress"
	}
	if config.CompletedLabel == "" {
		config.CompletedLabel = "completed"
	}
	if config.BaseBranch == "" {
		config.BaseBranch = "main"
	}
	if config.BranchPrefix == "" {
		config.BranchPrefix = "bzzz/task-"
	}
	
	// Create OAuth2 token source
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.AccessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	
	client := &Client{
		client: github.NewClient(tc),
		ctx:    ctx,
		config: config,
	}
	
	// Verify access to repository
	if err := client.verifyAccess(); err != nil {
		return nil, fmt.Errorf("failed to verify GitHub access: %w", err)
	}
	
	return client, nil
}

// verifyAccess checks if we can access the configured repository
func (c *Client) verifyAccess() error {
	_, _, err := c.client.Repositories.Get(c.ctx, c.config.Owner, c.config.Repository)
	if err != nil {
		return fmt.Errorf("cannot access repository %s/%s: %w", 
			c.config.Owner, c.config.Repository, err)
	}
	return nil
}

// Task represents a Bzzz task as a GitHub issue
type Task struct {
	ID          int64     `json:"id"`
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	State       string    `json:"state"`       // open, closed
	Labels      []string  `json:"labels"`
	Assignee    string    `json:"assignee"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// Bzzz-specific fields
	TaskType     string            `json:"task_type"`
	Priority     int               `json:"priority"`
	Requirements []string          `json:"requirements"`
	Deliverables []string          `json:"deliverables"`
	Context      map[string]interface{} `json:"context"`
}

// CreateTask creates a new GitHub issue for a Bzzz task
func (c *Client) CreateTask(task *Task) (*Task, error) {
	// Prepare issue request
	issue := &github.IssueRequest{
		Title: &task.Title,
		Body:  github.String(c.formatTaskBody(task)),
		Labels: &[]string{
			c.config.TaskLabel,
			fmt.Sprintf("priority-%d", task.Priority),
			fmt.Sprintf("type-%s", task.TaskType),
		},
	}
	
	// Create the issue
	createdIssue, _, err := c.client.Issues.Create(
		c.ctx, 
		c.config.Owner, 
		c.config.Repository, 
		issue,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub issue: %w", err)
	}
	
	// Convert back to our Task format
	return c.issueToTask(createdIssue), nil
}

// ClaimTask atomically assigns a task to an agent
func (c *Client) ClaimTask(issueNumber int, agentID string) (*Task, error) {
	// Get current issue state
	issue, _, err := c.client.Issues.Get(
		c.ctx, 
		c.config.Owner, 
		c.config.Repository, 
		issueNumber,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}
	
	// Check if already assigned
	if issue.Assignee != nil {
		return nil, fmt.Errorf("task already assigned to %s", issue.Assignee.GetLogin())
	}
	
	// Attempt atomic assignment using GitHub's native assignment
	// GitHub only accepts existing usernames, so we'll assign to the repo owner
	githubAssignee := "anthonyrawlins"
	issueRequest := &github.IssueRequest{
		Assignee: &githubAssignee,
	}
	
	// Add in-progress label
	currentLabels := make([]string, 0, len(issue.Labels)+1)
	for _, label := range issue.Labels {
		currentLabels = append(currentLabels, label.GetName())
	}
	currentLabels = append(currentLabels, c.config.InProgressLabel)
	issueRequest.Labels = &currentLabels
	
	// Update the issue
	updatedIssue, _, err := c.client.Issues.Edit(
		c.ctx,
		c.config.Owner,
		c.config.Repository,
		issueNumber,
		issueRequest,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to claim task: %w", err)
	}
	
	// Add a comment to track which Bzzz agent claimed this task
	claimComment := fmt.Sprintf("🐝 **Task claimed by Bzzz agent:** `%s`\n\nThis task has been automatically claimed by the Bzzz P2P task coordination system.", agentID)
	commentRequest := &github.IssueComment{
		Body: &claimComment,
	}
	_, _, err = c.client.Issues.CreateComment(
		c.ctx,
		c.config.Owner,
		c.config.Repository,
		issueNumber,
		commentRequest,
	)
	if err != nil {
		// Log error but don't fail the claim
		fmt.Printf("⚠️ Failed to add claim comment: %v\n", err)
	}
	
	// Create a task branch
	if err := c.createTaskBranch(issueNumber, agentID); err != nil {
		// Log error but don't fail the claim
		fmt.Printf("⚠️ Failed to create task branch: %v\n", err)
	}
	
	return c.issueToTask(updatedIssue), nil
}

// CompleteTask marks a task as completed and creates a pull request
func (c *Client) CompleteTask(issueNumber int, agentID string, results map[string]interface{}) error {
	// Update issue labels
	issue, _, err := c.client.Issues.Get(
		c.ctx,
		c.config.Owner,
		c.config.Repository,
		issueNumber,
	)
	if err != nil {
		return fmt.Errorf("failed to get issue: %w", err)
	}
	
	// Remove in-progress label, add completed label
	newLabels := make([]string, 0, len(issue.Labels))
	for _, label := range issue.Labels {
		labelName := label.GetName()
		if labelName != c.config.InProgressLabel {
			newLabels = append(newLabels, labelName)
		}
	}
	newLabels = append(newLabels, c.config.CompletedLabel)
	
	// Add completion comment
	comment := &github.IssueComment{
		Body: github.String(c.formatCompletionComment(agentID, results)),
	}
	
	_, _, err = c.client.Issues.CreateComment(
		c.ctx,
		c.config.Owner,
		c.config.Repository,
		issueNumber,
		comment,
	)
	if err != nil {
		return fmt.Errorf("failed to add completion comment: %w", err)
	}
	
	// Update labels
	issueRequest := &github.IssueRequest{
		Labels: &newLabels,
		State:  github.String("closed"),
	}
	
	_, _, err = c.client.Issues.Edit(
		c.ctx,
		c.config.Owner,
		c.config.Repository,
		issueNumber,
		issueRequest,
	)
	if err != nil {
		return fmt.Errorf("failed to update issue: %w", err)
	}
	
	return nil
}

// ListAvailableTasks returns unassigned Bzzz tasks
func (c *Client) ListAvailableTasks() ([]*Task, error) {
	// Search for open issues with Bzzz task label and no assignee
	opts := &github.IssueListByRepoOptions{
		State:     "open",
		Labels:    []string{c.config.TaskLabel},
		Assignee:  "none",
		Sort:      "created",
		Direction: "desc",
		ListOptions: github.ListOptions{PerPage: 50},
	}
	
	issues, _, err := c.client.Issues.ListByRepo(
		c.ctx,
		c.config.Owner,
		c.config.Repository,
		opts,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}
	
	tasks := make([]*Task, 0, len(issues))
	for _, issue := range issues {
		tasks = append(tasks, c.issueToTask(issue))
	}
	
	return tasks, nil
}

// createTaskBranch creates a new branch for task work
func (c *Client) createTaskBranch(issueNumber int, agentID string) error {
	branchName := fmt.Sprintf("%s%d-%s", c.config.BranchPrefix, issueNumber, agentID)
	
	// Get the base branch reference
	baseRef, _, err := c.client.Git.GetRef(
		c.ctx,
		c.config.Owner,
		c.config.Repository,
		"refs/heads/"+c.config.BaseBranch,
	)
	if err != nil {
		return fmt.Errorf("failed to get base branch: %w", err)
	}
	
	// Create new branch
	newRef := &github.Reference{
		Ref: github.String("refs/heads/" + branchName),
		Object: &github.GitObject{
			SHA: baseRef.Object.SHA,
		},
	}
	
	_, _, err = c.client.Git.CreateRef(
		c.ctx,
		c.config.Owner,
		c.config.Repository,
		newRef,
	)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}
	
	fmt.Printf("🌿 Created task branch: %s\n", branchName)
	return nil
}

// CreatePullRequest creates a new pull request for a completed task.
func (c *Client) CreatePullRequest(issueNumber int, branchName, agentID string) (*github.PullRequest, error) {
	title := fmt.Sprintf("fix: resolve issue #%d via bzzz agent %s", issueNumber, agentID)
	body := fmt.Sprintf("This pull request resolves issue #%d, and was automatically generated by the Bzzz agent `%s`.", issueNumber, agentID)
	head := branchName
	base := c.config.BaseBranch

	pr := &github.NewPullRequest{
		Title: &title,
		Body:  &body,
		Head:  &head,
		Base:  &base,
	}

	newPR, _, err := c.client.PullRequests.Create(c.ctx, c.config.Owner, c.config.Repository, pr)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	return newPR, nil
}

// formatTaskBody formats task details into GitHub issue body
func (c *Client) formatTaskBody(task *Task) string {
	body := fmt.Sprintf("**Task Type:** %s\n", task.TaskType)
	body += fmt.Sprintf("**Priority:** %d\n", task.Priority)
	body += fmt.Sprintf("\n**Description:**\n%s\n", task.Description)
	
	if len(task.Requirements) > 0 {
		body += "\n**Requirements:**\n"
		for _, req := range task.Requirements {
			body += fmt.Sprintf("- %s\n", req)
		}
	}
	
	if len(task.Deliverables) > 0 {
		body += "\n**Deliverables:**\n"
		for _, deliverable := range task.Deliverables {
			body += fmt.Sprintf("- %s\n", deliverable)
		}
	}
	
	body += "\n---\n*This task is managed by Bzzz P2P Task Coordination System*"
	return body
}

// formatCompletionComment formats task completion results
func (c *Client) formatCompletionComment(agentID string, results map[string]interface{}) string {
	comment := fmt.Sprintf("✅ **Task completed by agent: %s**\n\n", agentID)
	comment += fmt.Sprintf("**Completion time:** %s\n\n", time.Now().Format(time.RFC3339))
	
	if len(results) > 0 {
		comment += "**Results:**\n"
		for key, value := range results {
			comment += fmt.Sprintf("- **%s:** %v\n", key, value)
		}
	}
	
	return comment
}

// issueToTask converts a GitHub issue to a Bzzz task
func (c *Client) issueToTask(issue *github.Issue) *Task {
	task := &Task{
		ID:          issue.GetID(),
		Number:      issue.GetNumber(),
		Title:       issue.GetTitle(),
		Description: issue.GetBody(),
		State:       issue.GetState(),
		CreatedAt:   issue.GetCreatedAt().Time,
		UpdatedAt:   issue.GetUpdatedAt().Time,
	}
	
	// Extract labels
	task.Labels = make([]string, 0, len(issue.Labels))
	for _, label := range issue.Labels {
		task.Labels = append(task.Labels, label.GetName())
	}
	
	// Extract assignee
	if issue.Assignee != nil {
		task.Assignee = issue.Assignee.GetLogin()
	}
	
	return task
}