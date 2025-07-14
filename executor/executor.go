package executor

import (
	"context"
	"fmt"
	"strings"

	"github.com/anthonyrawlins/bzzz/logging"
	"github.com/anthonyrawlins/bzzz/pkg/types"
	"github.com/anthonyrawlins/bzzz/reasoning"
	"github.com/anthonyrawlins/bzzz/sandbox"
)

const maxIterations = 10 // Prevents infinite loops

// ExecuteTaskResult contains the result of task execution
type ExecuteTaskResult struct {
	BranchName string
	Sandbox    *sandbox.Sandbox
}

// ExecuteTask manages the entire lifecycle of a task using a sandboxed environment.
// Returns sandbox reference so it can be destroyed after PR creation
func ExecuteTask(ctx context.Context, task *types.EnhancedTask, hlog *logging.HypercoreLog) (*ExecuteTaskResult, error) {
	// 1. Create the sandbox environment
	sb, err := sandbox.CreateSandbox(ctx, "") // Use default image for now
	if err != nil {
		return nil, fmt.Errorf("failed to create sandbox: %w", err)
	}
	// NOTE: Do NOT defer destroy here - let caller handle it

	// 2. Clone the repository inside the sandbox
	cloneCmd := fmt.Sprintf("git clone %s .", task.GitURL)
	if _, err := sb.RunCommand(cloneCmd); err != nil {
		sb.DestroySandbox() // Clean up on error
		return nil, fmt.Errorf("failed to clone repository in sandbox: %w", err)
	}
	hlog.Append(logging.TaskProgress, map[string]interface{}{"task_id": task.Number, "status": "cloned repo"})

	// 3. The main iterative development loop
	var lastCommandOutput string
	for i := 0; i < maxIterations; i++ {
		// a. Generate the next command based on the task and previous output
		nextCommand, err := generateNextCommand(ctx, task, lastCommandOutput)
		if err != nil {
			sb.DestroySandbox() // Clean up on error
			return nil, fmt.Errorf("failed to generate next command: %w", err)
		}

		hlog.Append(logging.TaskProgress, map[string]interface{}{
			"task_id":   task.Number,
			"iteration": i,
			"command":   nextCommand,
		})

		// b. Check for completion command
		if strings.HasPrefix(nextCommand, "TASK_COMPLETE") {
			fmt.Println("✅ Agent has determined the task is complete.")
			break // Exit loop to proceed with PR creation
		}

		// c. Execute the command in the sandbox
		result, err := sb.RunCommand(nextCommand)
		if err != nil {
			// Log the error and feed it back to the agent
			lastCommandOutput = fmt.Sprintf("Command failed: %v\nStdout: %s\nStderr: %s", err, result.StdOut, result.StdErr)
			continue
		}

		// d. Store the output for the next iteration
		lastCommandOutput = fmt.Sprintf("Stdout: %s\nStderr: %s", result.StdOut, result.StdErr)
	}

	// 4. Create a new branch and commit the changes
	branchName := fmt.Sprintf("bzzz-task-%d", task.Number)
	if _, err := sb.RunCommand(fmt.Sprintf("git checkout -b %s", branchName)); err != nil {
		sb.DestroySandbox() // Clean up on error
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}
	if _, err := sb.RunCommand("git add ."); err != nil {
		sb.DestroySandbox() // Clean up on error
		return nil, fmt.Errorf("failed to add files: %w", err)
	}
	commitCmd := fmt.Sprintf("git commit -m 'feat: resolve task #%d'", task.Number)
	if _, err := sb.RunCommand(commitCmd); err != nil {
		sb.DestroySandbox() // Clean up on error
		return nil, fmt.Errorf("failed to commit changes: %w", err)
	}

	// 5. Push the new branch
	if _, err := sb.RunCommand(fmt.Sprintf("git push origin %s", branchName)); err != nil {
		sb.DestroySandbox() // Clean up on error
		return nil, fmt.Errorf("failed to push branch: %w", err)
	}

	hlog.Append(logging.TaskProgress, map[string]interface{}{"task_id": task.Number, "status": "pushed changes"})
	return &ExecuteTaskResult{
		BranchName: branchName,
		Sandbox:    sb,
	}, nil
}

// generateNextCommand uses the LLM to decide the next command to execute.
func generateNextCommand(ctx context.Context, task *types.EnhancedTask, lastOutput string) (string, error) {
	prompt := fmt.Sprintf(
		"You are an AI developer agent in the Bzzz P2P distributed development network, working in a sandboxed shell environment.\n\n"+
			"TASK DETAILS:\n"+
			"Title: %s\nDescription: %s\n\n"+
			"CAPABILITIES & RESOURCES:\n"+
			"- You can issue shell commands to solve this GitHub issue\n"+
			"- You are part of a collaborative P2P mesh with other AI agents\n"+
			"- If stuck, you can ask for help by using keywords: 'stuck', 'help', 'clarification needed', 'manual intervention'\n"+
			"- Complex problems automatically escalate to human experts via N8N webhooks\n"+
			"- You have access to git, build tools, editors, and development utilities\n"+
			"- GitHub CLI (gh) is available for creating PRs: use 'gh pr create --title \"title\" --body \"description\"'\n"+
			"- GitHub authentication is configured automatically\n"+
			"- Work is preserved even if issues occur - your changes are committed and pushed\n\n"+
			"COLLABORATION GUIDELINES:\n"+
			"- Use clear, descriptive commit messages\n"+
			"- Break complex problems into smaller steps\n"+
			"- Ask for help early if you encounter unfamiliar technologies\n"+
			"- Document your reasoning in commands where helpful\n\n"+
			"PREVIOUS OUTPUT:\n---\n%s\n---\n\n"+
			"Based on this context, what is the single next shell command you should run?\n"+
			"If you believe the task is complete and ready for a pull request, respond with 'TASK_COMPLETE'.\n"+
			"If you need help, include relevant keywords in your response.",
		task.Title, task.Description, lastOutput,
	)

	// Using the main reasoning engine to generate the command
	command, err := reasoning.GenerateResponse(ctx, "phi3", prompt)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(command), nil
}
