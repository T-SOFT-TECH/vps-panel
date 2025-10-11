package docker

import (
	"archive/tar"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	"github.com/vps-panel/backend/internal/models"
)

type DockerService struct {
	client *client.Client
}

func NewDockerService() (*DockerService, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &DockerService{client: cli}, nil
}

// LogCallback is a function type for logging build output
type LogCallback func(message string)

func (s *DockerService) BuildImage(ctx context.Context, buildPath string, imageName string, logFn LogCallback) error {
	// Create tar archive from build path
	buildContext, err := s.createTarArchive(buildPath)
	if err != nil {
		return fmt.Errorf("failed to create tar archive: %w", err)
	}
	defer buildContext.Close()

	buildOptions := types.ImageBuildOptions{
		Tags:       []string{imageName},
		Dockerfile: "Dockerfile",
		Remove:     true,
	}

	response, err := s.client.ImageBuild(ctx, buildContext, buildOptions)
	if err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}
	defer response.Body.Close()

	// Parse build output and check for errors
	type buildOutput struct {
		Stream string `json:"stream"`
		Error  string `json:"error"`
	}

	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Parse JSON to get the actual output
		var output buildOutput
		if err := json.Unmarshal([]byte(line), &output); err == nil {
			// Log the stream output if present
			if output.Stream != "" && logFn != nil {
				// Remove trailing newline as logBuild adds one
				msg := output.Stream
				if len(msg) > 0 && msg[len(msg)-1] == '\n' {
					msg = msg[:len(msg)-1]
				}
				logFn(msg)
			}

			// Check for errors
			if output.Error != "" {
				return fmt.Errorf("docker build error: %s", output.Error)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading build output: %w", err)
	}

	return nil
}

// createTarArchive creates a tar archive from a directory
func (s *DockerService) createTarArchive(srcPath string) (io.ReadCloser, error) {
	pr, pw := io.Pipe()

	go func() {
		tw := tar.NewWriter(pw)
		defer tw.Close()
		defer pw.Close()

		// Walk through the directory and add files to tar
		filepath.Walk(srcPath, func(file string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Create tar header
			header, err := tar.FileInfoHeader(fi, fi.Name())
			if err != nil {
				return err
			}

			// Update the name to be relative to srcPath
			relPath, err := filepath.Rel(srcPath, file)
			if err != nil {
				return err
			}
			header.Name = filepath.ToSlash(relPath)

			// Skip the root directory itself
			if header.Name == "." {
				return nil
			}

			// Write header
			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			// If it's a file (not a directory), write the content
			if !fi.IsDir() {
				f, err := os.Open(file)
				if err != nil {
					return err
				}
				defer f.Close()

				if _, err := io.Copy(tw, f); err != nil {
					return err
				}
			}

			return nil
		})
	}()

	return pr, nil
}

func (s *DockerService) CreateContainer(ctx context.Context, project *models.Project, imageName string) (string, error) {
	containerName := fmt.Sprintf("vps-panel-%s-%d", project.Name, project.ID)

	// Port bindings
	// Container always uses port 3000 internally, map to assigned host port
	portBindings := nat.PortMap{}
	if project.FrontendPort > 0 {
		portBindings[nat.Port("3000/tcp")] = []nat.PortBinding{
			{HostIP: "127.0.0.1", HostPort: fmt.Sprintf("%d", project.FrontendPort)},
		}
	}
	if project.BackendPort > 0 {
		portBindings[nat.Port("8080/tcp")] = []nat.PortBinding{
			{HostIP: "127.0.0.1", HostPort: fmt.Sprintf("%d", project.BackendPort)},
		}
	}

	config := &container.Config{
		Image: imageName,
		Env:   s.buildEnvVars(project),
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		RestartPolicy: container.RestartPolicy{
			Name: container.RestartPolicyUnlessStopped,
		},
	}

	// Remove existing container if exists
	s.RemoveContainer(ctx, containerName)

	// Create container
	resp, err := s.client.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	return resp.ID, nil
}

func (s *DockerService) StartContainer(ctx context.Context, containerID string) error {
	return s.client.ContainerStart(ctx, containerID, container.StartOptions{})
}

func (s *DockerService) StopContainer(ctx context.Context, containerID string) error {
	timeout := 10
	return s.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout})
}

func (s *DockerService) RemoveContainer(ctx context.Context, containerName string) error {
	return s.client.ContainerRemove(ctx, containerName, container.RemoveOptions{
		Force: true,
	})
}

func (s *DockerService) GetContainerLogs(ctx context.Context, containerID string) (io.ReadCloser, error) {
	return s.client.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     false,
		Timestamps: true,
	})
}

func (s *DockerService) buildEnvVars(project *models.Project) []string {
	envVars := []string{
		"NODE_ENV=production",
		"PORT=3000", // Container always listens on port 3000 internally
	}

	// Add project-specific environment variables
	for _, env := range project.Environments {
		envVars = append(envVars, fmt.Sprintf("%s=%s", env.Key, env.Value))
	}

	return envVars
}

func (s *DockerService) Close() error {
	return s.client.Close()
}

// ComposeDown stops and removes containers created by docker compose
func (s *DockerService) ComposeDown(ctx context.Context, workDir string, projectName string) error {
	cmd := fmt.Sprintf("docker compose -f docker-compose.yml -p %s down --remove-orphans", projectName)
	return execCommand(ctx, workDir, cmd)
}

// ComposeBuild builds images defined in docker-compose.yml
func (s *DockerService) ComposeBuild(ctx context.Context, workDir string, projectName string, logFn LogCallback) error {
	cmd := fmt.Sprintf("docker compose -f docker-compose.yml -p %s build --no-cache", projectName)
	return execCommandWithOutput(ctx, workDir, cmd, logFn)
}

// ComposeUp starts containers defined in docker-compose.yml
func (s *DockerService) ComposeUp(ctx context.Context, workDir string, projectName string) error {
	cmd := fmt.Sprintf("docker compose -f docker-compose.yml -p %s up -d", projectName)
	return execCommand(ctx, workDir, cmd)
}

// execCommand executes a shell command in a given directory
func execCommand(ctx context.Context, workDir string, command string) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}

	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// execCommandWithOutput executes a command and streams output via callback
func execCommandWithOutput(ctx context.Context, workDir string, command string, logFn LogCallback) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}

	cmd.Dir = workDir

	// Capture stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Read stdout in goroutine
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			if logFn != nil {
				logFn(scanner.Text())
			}
		}
	}()

	// Read stderr in goroutine
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			if logFn != nil {
				logFn(scanner.Text())
			}
		}
	}()

	return cmd.Wait()
}
