package deployment

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"gorm.io/gorm"

	"github.com/vps-panel/backend/internal/config"
	"github.com/vps-panel/backend/internal/models"
	"github.com/vps-panel/backend/internal/services/caddy"
	"github.com/vps-panel/backend/internal/services/docker"
	"github.com/vps-panel/backend/internal/services/git"
)

type DeploymentService struct {
	db            *gorm.DB
	cfg           *config.Config
	gitService    *git.GitService
	dockerService *docker.DockerService
	caddyService  *caddy.CaddyService
}

func NewDeploymentService(db *gorm.DB, cfg *config.Config) (*DeploymentService, error) {
	dockerService, err := docker.NewDockerService()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker service: %w", err)
	}

	return &DeploymentService{
		db:            db,
		cfg:           cfg,
		gitService:    git.NewGitService(cfg.ProjectsDir),
		dockerService: dockerService,
		caddyService:  caddy.NewCaddyService(cfg.CaddyConfigPath, cfg.CaddyReloadCmd),
	}, nil
}

func (s *DeploymentService) Deploy(deploymentID uint) error {
	// Load deployment
	var deployment models.Deployment
	if err := s.db.Preload("Project").Preload("Project.Domains").Preload("Project.Environments").
		First(&deployment, deploymentID).Error; err != nil {
		return fmt.Errorf("failed to load deployment: %w", err)
	}

	project := deployment.Project

	// Update deployment status
	deployment.Status = models.DeploymentBuilding
	s.db.Save(&deployment)

	// Execute deployment steps
	ctx := context.Background()
	startTime := time.Now()

	if err := s.executeDeployment(ctx, &deployment, &project); err != nil {
		// Mark deployment as failed
		deployment.Status = models.DeploymentFailed
		deployment.ErrorMessage = err.Error()
		now := time.Now()
		deployment.CompletedAt = &now
		deployment.Duration = int(time.Since(startTime).Seconds())
		s.db.Save(&deployment)

		s.logBuild(deployment.ID, fmt.Sprintf("Deployment failed: %v", err), "error")
		return err
	}

	// Mark deployment as successful
	deployment.Status = models.DeploymentSuccess
	now := time.Now()
	deployment.CompletedAt = &now
	deployment.Duration = int(time.Since(startTime).Seconds())
	s.db.Save(&deployment)

	// Update project status
	project.Status = "active"
	project.LastDeployed = &now
	s.db.Save(&project)

	s.logBuild(deployment.ID, "Deployment completed successfully!", "info")
	return nil
}

func (s *DeploymentService) executeDeployment(ctx context.Context, deployment *models.Deployment, project *models.Project) error {
	// Step 1: Clone repository
	s.logBuild(deployment.ID, "Cloning repository...", "info")
	repoPath, err := s.gitService.Clone(fmt.Sprintf("project-%d", project.ID), git.CloneOptions{
		URL:      project.GitURL,
		Branch:   project.GitBranch,
		Depth:    1,
		Username: project.GitUsername,
		Token:    project.GitToken,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Get commit info
	commitInfo, err := s.gitService.GetLatestCommit(repoPath)
	if err != nil {
		log.Printf("Warning: failed to get commit info: %v", err)
	} else {
		deployment.CommitHash = commitInfo.Hash
		deployment.CommitMessage = commitInfo.Message
		deployment.CommitAuthor = commitInfo.Author
		s.db.Save(&deployment)
	}

	// Determine working directory (for monorepos with root_directory specified)
	workDir := repoPath
	if project.RootDirectory != "" {
		workDir = filepath.Join(repoPath, project.RootDirectory)
		s.logBuild(deployment.ID, fmt.Sprintf("Using subdirectory: %s", project.RootDirectory), "info")
	}

	// Step 2: Detect framework and generate Dockerfile if needed
	s.logBuild(deployment.ID, "Detecting project structure...", "info")
	if err := s.ensureDockerfile(workDir, project); err != nil {
		return fmt.Errorf("failed to create Dockerfile: %w", err)
	}

	// Step 3: Install dependencies and build
	s.logBuild(deployment.ID, "Installing dependencies...", "info")
	if err := s.runCommand(workDir, project.InstallCommand); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	s.logBuild(deployment.ID, "Building project...", "info")
	if err := s.runCommand(workDir, project.BuildCommand); err != nil {
		return fmt.Errorf("failed to build project: %w", err)
	}

	// Step 4: Build Docker image
	s.logBuild(deployment.ID, "Building Docker image...", "info")
	imageName := fmt.Sprintf("vps-panel/project-%d:latest", project.ID)
	if err := s.dockerService.BuildImage(ctx, workDir, imageName); err != nil {
		return fmt.Errorf("failed to build Docker image: %w", err)
	}

	// Step 5: Deploy container
	deployment.Status = models.DeploymentDeploying
	s.db.Save(&deployment)
	s.logBuild(deployment.ID, "Deploying container...", "info")

	containerID, err := s.dockerService.CreateContainer(ctx, project, imageName)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	if err := s.dockerService.StartContainer(ctx, containerID); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Step 6: Update Caddy configuration
	s.logBuild(deployment.ID, "Updating reverse proxy configuration...", "info")
	if err := s.caddyService.GenerateConfig(project); err != nil {
		return fmt.Errorf("failed to generate Caddy config: %w", err)
	}

	if err := s.caddyService.Reload(); err != nil {
		log.Printf("Warning: failed to reload Caddy: %v", err)
	}

	return nil
}

func (s *DeploymentService) ensureDockerfile(repoPath string, project *models.Project) error {
	dockerfilePath := filepath.Join(repoPath, "Dockerfile")

	// Check if Dockerfile already exists
	if _, err := os.Stat(dockerfilePath); err == nil {
		return nil // Dockerfile exists
	}

	// Generate Dockerfile based on framework
	dockerfile := s.generateDockerfile(project)
	return os.WriteFile(dockerfilePath, []byte(dockerfile), 0644)
}

func (s *DeploymentService) generateDockerfile(project *models.Project) string {
	nodeVersion := project.NodeVersion
	if nodeVersion == "" {
		nodeVersion = "20"
	}

	outputDir := project.OutputDir
	if outputDir == "" {
		outputDir = "build"
	}

	return fmt.Sprintf(`FROM node:%s-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM node:%s-alpine

WORKDIR /app

COPY --from=builder /app/%s ./%s
COPY --from=builder /app/package*.json ./

RUN npm ci --production

EXPOSE %d

CMD ["node", "%s/index.js"]
`, nodeVersion, nodeVersion, outputDir, outputDir, project.FrontendPort, outputDir)
}

func (s *DeploymentService) runCommand(workDir, command string) error {
	var cmd *exec.Cmd

	// Use appropriate shell based on OS
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (s *DeploymentService) logBuild(deploymentID uint, message, logType string) {
	log.Println(message)

	buildLog := models.BuildLog{
		DeploymentID: deploymentID,
		Log:          message,
		LogType:      logType,
	}
	s.db.Create(&buildLog)
}
