package sandbox

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

const (
	// DefaultDockerImage is the image used if a task does not specify one.
	DefaultDockerImage = "registry.home.deepblack.cloud/tony/bzzz-sandbox:latest"
)

// Sandbox represents a stateful, isolated execution environment for a single task.
type Sandbox struct {
	ID          string // The ID of the running container.
	HostPath    string // The path on the host machine mounted as the workspace.
	Workspace   string // The path inside the container that is the workspace.
	dockerCli   *client.Client
	ctx         context.Context
}

// CommandResult holds the output of a command executed in the sandbox.
type CommandResult struct {
	StdOut   string
	StdErr   string
	ExitCode int
}

// CreateSandbox provisions a new Docker container for a task.
func CreateSandbox(ctx context.Context, taskImage string) (*Sandbox, error) {
	if taskImage == "" {
		taskImage = DefaultDockerImage
	}

	// Create a new Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	// Create a temporary directory on the host
	hostPath, err := os.MkdirTemp("", "bzzz-sandbox-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir for sandbox: %w", err)
	}

	// Read GitHub token for authentication
	githubToken := os.Getenv("BZZZ_GITHUB_TOKEN")
	if githubToken == "" {
		// Try to read from file
		tokenBytes, err := os.ReadFile("/home/tony/AI/secrets/passwords_and_tokens/gh-token")
		if err == nil {
			githubToken = strings.TrimSpace(string(tokenBytes))
		}
	}

	// Define container configuration
	containerConfig := &container.Config{
		Image:        taskImage,
		Tty:          true, // Keep the container running
		OpenStdin:    true,
		WorkingDir:   "/home/agent/work",
		User:         "agent",
		Env: []string{
			"GITHUB_TOKEN=" + githubToken,
			"GH_TOKEN=" + githubToken,
		},
	}

	// Define host configuration (e.g., volume mounts, resource limits)
	hostConfig := &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/home/agent/work", hostPath)},
		Resources: container.Resources{
			NanoCPUs: 2 * 1000000000, // 2 CPUs
			Memory:   2 * 1024 * 1024 * 1024, // 2GB
		},
	}

	// Create the container
	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		os.RemoveAll(hostPath) // Clean up the directory if container creation fails
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Start the container
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		os.RemoveAll(hostPath) // Clean up
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	fmt.Printf("✅ Sandbox container %s created successfully.\n", resp.ID[:12])

	return &Sandbox{
		ID:          resp.ID,
		HostPath:    hostPath,
		Workspace:   "/home/agent/work",
		dockerCli:   cli,
		ctx:         ctx,
	}, nil
}

// DestroySandbox stops and removes the container and its associated host directory.
func (s *Sandbox) DestroySandbox() error {
	if s == nil || s.ID == "" {
		return nil
	}

	// Define a timeout for stopping the container
	timeout := 30 // seconds
	
	// Stop the container
	fmt.Printf("🛑 Stopping sandbox container %s...\n", s.ID[:12])
	err := s.dockerCli.ContainerStop(s.ctx, s.ID, container.StopOptions{Timeout: &timeout})
	if err != nil {
		// Log the error but continue to try and clean up
		fmt.Printf("⚠️  Error stopping container %s: %v. Proceeding with cleanup.\n", s.ID, err)
	}

	// Remove the container
	err = s.dockerCli.ContainerRemove(s.ctx, s.ID, container.RemoveOptions{Force: true})
	if err != nil {
		fmt.Printf("⚠️  Error removing container %s: %v. Proceeding with cleanup.\n", s.ID, err)
	}

	// Remove the host directory
	fmt.Printf("🗑️  Removing host directory %s...\n", s.HostPath)
	err = os.RemoveAll(s.HostPath)
	if err != nil {
		return fmt.Errorf("failed to remove host directory %s: %w", s.HostPath, err)
	}

	fmt.Printf("✅ Sandbox %s destroyed successfully.\n", s.ID[:12])
	return nil
}

// RunCommand executes a shell command inside the sandbox.
func (s *Sandbox) RunCommand(command string) (*CommandResult, error) {
	// Configuration for the exec process
	execConfig := container.ExecOptions{
		Cmd:          []string{"/bin/sh", "-c", command},
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
	}

	// Create the exec instance
	execID, err := s.dockerCli.ContainerExecCreate(s.ctx, s.ID, execConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec in container: %w", err)
	}

	// Start the exec process
	resp, err := s.dockerCli.ContainerExecAttach(s.ctx, execID.ID, container.ExecStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to exec in container: %w", err)
	}
	defer resp.Close()

	// Read the output
	var stdout, stderr bytes.Buffer
	_, err = stdcopy.StdCopy(&stdout, &stderr, resp.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read exec output: %w", err)
	}

	// Inspect the exec process to get the exit code
	inspect, err := s.dockerCli.ContainerExecInspect(s.ctx, execID.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect exec in container: %w", err)
	}

	return &CommandResult{
		StdOut:   stdout.String(),
		StdErr:   stderr.String(),
		ExitCode: inspect.ExitCode,
	}, nil
}

// WriteFile writes content to a file inside the sandbox's workspace.
func (s *Sandbox) WriteFile(path string, content []byte) error {
	// Create a temporary file on the host
	tmpfile, err := os.CreateTemp("", "bzzz-write-")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}
	tmpfile.Close()

	// Copy the file into the container
	dstPath := filepath.Join(s.Workspace, path)
	
	// Create tar archive of the file
	tarBuf := new(bytes.Buffer)
	tw := tar.NewWriter(tarBuf)
	
	fileInfo, err := os.Stat(tmpfile.Name())
	if err != nil {
		return fmt.Errorf("failed to stat temp file: %w", err)
	}
	
	header := &tar.Header{
		Name: filepath.Base(path),
		Size: fileInfo.Size(),
		Mode: 0644,
	}
	
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header: %w", err)
	}
	
	fileContent, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return fmt.Errorf("failed to read temp file: %w", err)
	}
	
	if _, err := tw.Write(fileContent); err != nil {
		return fmt.Errorf("failed to write to tar: %w", err)
	}
	
	if err := tw.Close(); err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}
	
	return s.dockerCli.CopyToContainer(s.ctx, s.ID, filepath.Dir(dstPath), tarBuf, container.CopyToContainerOptions{})
}

// ReadFile reads the content of a file from the sandbox's workspace.
func (s *Sandbox) ReadFile(path string) ([]byte, error) {
	srcPath := filepath.Join(s.Workspace, path)

	// Copy the file from the container
	reader, _, err := s.dockerCli.CopyFromContainer(s.ctx, s.ID, srcPath)
	if err != nil {
		return nil, fmt.Errorf("failed to copy from container: %w", err)
	}
	defer reader.Close()

	// The result is a tar archive, so we need to extract it
	tr := tar.NewReader(reader)
	if _, err := tr.Next(); err != nil {
		return nil, fmt.Errorf("failed to get tar header: %w", err)
	}

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, tr); err != nil {
		return nil, fmt.Errorf("failed to read file content from tar: %w", err)
	}

	return buf.Bytes(), nil
}
