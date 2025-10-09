package deployment

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

	// Step 3: Build Docker image (includes install and build steps)
	s.logBuild(deployment.ID, "Building Docker image...", "info")
	imageName := fmt.Sprintf("vps-panel/project-%d:latest", project.ID)
	if err := s.dockerService.BuildImage(ctx, workDir, imageName); err != nil {
		return fmt.Errorf("failed to build Docker image: %w", err)
	}

	// Step 4: Check and assign available ports
	s.logBuild(deployment.ID, "Checking port availability...", "info")
	if err := s.ensureAvailablePorts(project, deployment.ID); err != nil {
		return fmt.Errorf("failed to ensure available ports: %w", err)
	}

	// Step 5: Deploy container
	deployment.Status = models.DeploymentDeploying
	s.db.Save(&deployment)
	s.logBuild(deployment.ID, "Deploying container...", "info")

	containerID, err := s.dockerService.CreateContainer(ctx, project, imageName)
	if err != nil {
		// Check for port conflict
		if strings.Contains(err.Error(), "address already in use") {
			portMsg := ""
			if project.FrontendPort > 0 && project.BackendPort > 0 {
				portMsg = fmt.Sprintf("frontend port %d or backend port %d", project.FrontendPort, project.BackendPort)
			} else if project.FrontendPort > 0 {
				portMsg = fmt.Sprintf("port %d", project.FrontendPort)
			} else if project.BackendPort > 0 {
				portMsg = fmt.Sprintf("port %d", project.BackendPort)
			}
			return fmt.Errorf("port conflict: %s is already in use. Please use a different port for your project", portMsg)
		}
		return fmt.Errorf("failed to create container: %w", err)
	}

	if err := s.dockerService.StartContainer(ctx, containerID); err != nil {
		// Check for port conflict on start
		if strings.Contains(err.Error(), "address already in use") {
			// Clean up the created container
			s.dockerService.RemoveContainer(ctx, fmt.Sprintf("vps-panel-%s-%d", project.Name, project.ID))

			portMsg := ""
			if project.FrontendPort > 0 && project.BackendPort > 0 {
				portMsg = fmt.Sprintf("frontend port %d or backend port %d", project.FrontendPort, project.BackendPort)
			} else if project.FrontendPort > 0 {
				portMsg = fmt.Sprintf("port %d", project.FrontendPort)
			} else if project.BackendPort > 0 {
				portMsg = fmt.Sprintf("port %d", project.BackendPort)
			}
			return fmt.Errorf("port conflict: %s is already in use. Please use a different port for your project", portMsg)
		}
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Step 6: Update Caddy configuration (optional - only if domains are configured)
	if len(project.Domains) > 0 && s.hasActiveDomains(project) {
		s.logBuild(deployment.ID, "Updating reverse proxy configuration...", "info")
		if err := s.caddyService.GenerateConfig(project); err != nil {
			return fmt.Errorf("failed to generate Caddy config: %w", err)
		}

		if err := s.caddyService.Reload(); err != nil {
			log.Printf("Warning: failed to reload Caddy: %v", err)
		}
	} else {
		s.logBuild(deployment.ID, fmt.Sprintf("No custom domains configured. Access your app at: http://<server-ip>:%d", project.FrontendPort), "info")
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

// isPortAvailable checks if a port is available for binding
func (s *DeploymentService) isPortAvailable(port int) bool {
	// Try to listen on the port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false // Port is not available
	}
	listener.Close()
	return true
}

// findAvailablePort finds an available port starting from the given port
func (s *DeploymentService) findAvailablePort(startPort int) (int, error) {
	// Try up to 100 ports
	for port := startPort; port < startPort+100; port++ {
		if s.isPortAvailable(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port found in range %d-%d", startPort, startPort+100)
}

// hasActiveDomains checks if the project has any active domains
func (s *DeploymentService) hasActiveDomains(project *models.Project) bool {
	for _, domain := range project.Domains {
		if domain.IsActive {
			return true
		}
	}
	return false
}

// ensureAvailablePorts checks and assigns available ports to the project
func (s *DeploymentService) ensureAvailablePorts(project *models.Project, deploymentID uint) error {
	portUpdated := false

	// Check and assign frontend port
	if project.FrontendPort > 0 {
		if !s.isPortAvailable(project.FrontendPort) {
			s.logBuild(deploymentID, fmt.Sprintf("Frontend port %d is in use, finding available port...", project.FrontendPort), "info")

			// Start searching from 3001 to avoid conflicts with VPS Panel (3000)
			newPort, err := s.findAvailablePort(3001)
			if err != nil {
				return fmt.Errorf("failed to find available frontend port: %w", err)
			}

			s.logBuild(deploymentID, fmt.Sprintf("Automatically assigned frontend port: %d", newPort), "info")
			project.FrontendPort = newPort
			portUpdated = true
		}
	}

	// Check and assign backend port
	if project.BackendPort > 0 {
		if !s.isPortAvailable(project.BackendPort) {
			s.logBuild(deploymentID, fmt.Sprintf("Backend port %d is in use, finding available port...", project.BackendPort), "info")

			// Start searching from 8080 for backend ports
			newPort, err := s.findAvailablePort(8080)
			if err != nil {
				return fmt.Errorf("failed to find available backend port: %w", err)
			}

			s.logBuild(deploymentID, fmt.Sprintf("Automatically assigned backend port: %d", newPort), "info")
			project.BackendPort = newPort
			portUpdated = true
		}
	}

	// Save project if ports were updated
	if portUpdated {
		if err := s.db.Save(project).Error; err != nil {
			return fmt.Errorf("failed to save updated ports: %w", err)
		}
	}

	return nil
}
