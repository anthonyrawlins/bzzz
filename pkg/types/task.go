package types

import (
	"time"

	"github.com/anthonyrawlins/bzzz/pkg/hive"
)

// EnhancedTask extends a basic Task with project-specific context.
// It's the primary data structure passed between the github, executor,
// and reasoning components.
type EnhancedTask struct {
	// Core task details, originally from the GitHub issue.
	ID          int64
	Number      int
	Title       string
	Description string
	State       string
	Labels      []string
	Assignee    string
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// Bzzz-specific fields parsed from the issue body or labels.
	TaskType     string
	Priority     int
	Requirements []string
	Deliverables []string
	Context      map[string]interface{}

	// Hive-integration fields providing repository context.
	ProjectID  int
	GitURL     string
	Repository hive.Repository
}
