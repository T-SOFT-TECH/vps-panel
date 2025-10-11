package deployment

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
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

	// Step 2: Pre-allocate deployment resources (ports and domain)
	// This must happen BEFORE creating .env file so environment variables can reference the deployment URL
	s.logBuild(deployment.ID, "Allocating deployment resources...", "info")

	// Allocate ports first
	if err := s.ensureAvailablePorts(project, deployment.ID); err != nil {
		return fmt.Errorf("failed to ensure available ports: %w", err)
	}

	// Pre-generate domain (if needed) so it's available during build
	if err := s.ensureProjectDomain(project, deployment.ID); err != nil {
		return fmt.Errorf("failed to ensure project domain: %w", err)
	}

	// Step 3: Detect framework and prepare for deployment
	s.logBuild(deployment.ID, "Detecting project structure...", "info")

	// For SvelteKit projects, ensure they have adapter-node
	if err := s.ensureSvelteKitAdapter(workDir, deployment.ID); err != nil {
		return fmt.Errorf("failed to prepare SvelteKit project: %w", err)
	}

	// Create .env file with environment variables from database
	// Now includes system variables like DEPLOYMENT_URL since domain is already allocated
	if err := s.createEnvFile(workDir, project, deployment.ID); err != nil {
		return fmt.Errorf("failed to create environment file: %w", err)
	}

	// Check if project uses PocketBase - if so, use docker-compose multi-container setup
	if project.BaaSType == models.BaaSPocketBase {
		s.logBuild(deployment.ID, "Detected PocketBase backend - using multi-container deployment", "info")

		// Ensure PocketBase directory structure
		if err := s.ensurePocketBaseStructure(workDir, deployment.ID); err != nil {
			return fmt.Errorf("failed to setup PocketBase structure: %w", err)
		}

		// Generate docker-compose.yml for multi-container deployment
		if err := s.generatePocketBaseDeploymentFiles(workDir, project, deployment.ID); err != nil {
			return fmt.Errorf("failed to generate PocketBase deployment files: %w", err)
		}

		// Deploy using docker-compose
		return s.deployWithDockerCompose(ctx, deployment, project, workDir)
	}

	// For non-PocketBase projects, use single container deployment
	// Generate Dockerfile if needed
	if err := s.ensureDockerfile(workDir, project); err != nil {
		return fmt.Errorf("failed to create Dockerfile: %w", err)
	}

	// Step 4: Build Docker image (includes install and build steps)
	s.logBuild(deployment.ID, "Building Docker image...", "info")
	imageName := fmt.Sprintf("vps-panel/project-%d:latest", project.ID)

	// Create a log callback that logs to the database
	logCallback := func(message string) {
		s.logBuild(deployment.ID, message, "info")
	}

	if err := s.dockerService.BuildImage(ctx, workDir, imageName, logCallback); err != nil {
		// Provide helpful error message
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "file does not exist") || strings.Contains(errorMsg, "no such file") {
			detectedFramework := s.detectFramework(workDir)
			s.logBuild(deployment.ID, fmt.Sprintf("Detected framework: %s", detectedFramework), "info")
			s.logBuild(deployment.ID, "Build failed - the output directory may not match your framework's build output", "error")
			s.logBuild(deployment.ID, fmt.Sprintf("Current output directory setting: %s", project.OutputDir), "error")
			s.logBuild(deployment.ID, "Check your project's build configuration and set the correct output directory in project settings", "error")
		}
		return fmt.Errorf("failed to build Docker image: %w", err)
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

	// Step 6: Update Caddy configuration (domain was already created in step 2)
	s.logBuild(deployment.ID, "Updating reverse proxy configuration...", "info")
	if err := s.caddyService.GenerateConfig(project); err != nil {
		return fmt.Errorf("failed to generate Caddy config: %w", err)
	}

	if err := s.caddyService.Reload(); err != nil {
		log.Printf("Warning: failed to reload Caddy: %v", err)
	}

	// Step 7: Wait for Caddy to provision SSL certificate
	// Caddy automatically provisions certificates after reload, but it happens asynchronously
	// We need to wait for this process to complete
	if len(project.Domains) > 0 {
		s.logBuild(deployment.ID, "Waiting for SSL certificate provisioning to complete...", "info")
		s.logBuild(deployment.ID, "Caddy is obtaining SSL certificate from Let's Encrypt...", "info")

		// Wait for Caddy's ACME process to complete (typically takes 5-15 seconds)
		// This gives Caddy enough time to:
		// 1. Start the ACME process
		// 2. Complete the TLS-ALPN-01 challenge
		// 3. Download and install the certificate
		time.Sleep(20 * time.Second)

		// Verify certificate was obtained
		if err := s.verifySSLCertificate(project, deployment.ID); err != nil {
			// Don't fail deployment, just warn
			log.Printf("Warning: SSL certificate verification incomplete: %v", err)
			s.logBuild(deployment.ID, "SSL certificate will be fully active on first user access", "warning")
		} else {
			s.logBuild(deployment.ID, "✓ SSL certificate provisioned successfully", "info")
			// Get the domain to show the user
			for _, d := range project.Domains {
				if d.IsActive {
					s.logBuild(deployment.ID, fmt.Sprintf("Your app is now live at: https://%s", d.Domain), "info")
					break
				}
			}
		}
	}

	return nil
}

// createEnvFile creates a .env file with environment variables from the database
// It automatically injects system variables like DEPLOYMENT_URL that are available during build
func (s *DeploymentService) createEnvFile(workDir string, project *models.Project, deploymentID uint) error {
	envFilePath := filepath.Join(workDir, ".env")

	// Check if .env already exists in repository
	if _, err := os.Stat(envFilePath); err == nil {
		s.logBuild(deploymentID, "Using existing .env file from repository", "info")
		s.logBuild(deploymentID, "Note: System variables (DEPLOYMENT_URL, etc.) will still be injected", "info")

		// Read existing content
		existingContent, err := os.ReadFile(envFilePath)
		if err != nil {
			return fmt.Errorf("failed to read existing .env file: %w", err)
		}

		// Prepend system variables
		var envContent strings.Builder
		envContent.WriteString("# System variables (auto-injected by VPS Panel)\n")
		s.writeSystemVariables(&envContent, project)
		envContent.WriteString("\n# Variables from repository .env file\n")
		envContent.Write(existingContent)

		if err := os.WriteFile(envFilePath, []byte(envContent.String()), 0644); err != nil {
			return fmt.Errorf("failed to update .env file: %w", err)
		}

		return nil
	}

	// Build .env content from database
	var envContent strings.Builder
	envContent.WriteString("# Environment variables from VPS Panel\n")
	envContent.WriteString("# Generated at build time\n\n")

	// Add system variables first
	envContent.WriteString("# System variables (auto-injected)\n")
	s.writeSystemVariables(&envContent, project)
	envContent.WriteString("\n")

	// Add user-configured variables
	if len(project.Environments) > 0 {
		envContent.WriteString("# User-configured variables\n")
		for _, env := range project.Environments {
			envContent.WriteString(fmt.Sprintf("%s=%s\n", env.Key, env.Value))
		}
	} else {
		envContent.WriteString("# No user-configured environment variables\n")
	}

	if err := os.WriteFile(envFilePath, []byte(envContent.String()), 0644); err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}

	totalVars := len(project.Environments) + 3 // +3 for system variables
	s.logBuild(deploymentID, fmt.Sprintf("Created .env file with %d environment variables (%d user + 3 system)", totalVars, len(project.Environments)), "info")
	return nil
}

// writeSystemVariables writes auto-injected system variables to the .env file
func (s *DeploymentService) writeSystemVariables(builder *strings.Builder, project *models.Project) {
	// Get deployment URL from the first active domain
	deploymentURL := ""
	deploymentDomain := ""
	for _, domain := range project.Domains {
		if domain.IsActive {
			deploymentDomain = domain.Domain
			if domain.SSLEnabled {
				deploymentURL = fmt.Sprintf("https://%s", domain.Domain)
			} else {
				deploymentURL = fmt.Sprintf("http://%s", domain.Domain)
			}
			break
		}
	}

	// Write system variables
	builder.WriteString(fmt.Sprintf("DEPLOYMENT_URL=%s\n", deploymentURL))
	builder.WriteString(fmt.Sprintf("DEPLOYMENT_DOMAIN=%s\n", deploymentDomain))
	builder.WriteString(fmt.Sprintf("PUBLIC_DEPLOYMENT_URL=%s\n", deploymentURL))
}

func (s *DeploymentService) ensureSvelteKitAdapter(repoPath string, deploymentID uint) error {
	packageJSONPath := filepath.Join(repoPath, "package.json")
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return nil // Not a Node.js project
	}

	content := string(data)

	// Check if it's a SvelteKit project
	if !strings.Contains(content, `"@sveltejs/kit"`) {
		return nil // Not SvelteKit, skip
	}

	s.logBuild(deploymentID, "Detected SvelteKit project", "info")

	// Check if adapter-node is already present
	if strings.Contains(content, `"@sveltejs/adapter-node"`) {
		s.logBuild(deploymentID, "Project already has adapter-node configured ✓", "info")
		return nil
	}

	// Check for other adapters
	hasCloudflare := strings.Contains(content, `"@sveltejs/adapter-cloudflare"`)
	hasVercel := strings.Contains(content, `"@sveltejs/adapter-vercel"`)
	hasAuto := strings.Contains(content, `"@sveltejs/adapter-auto"`)

	if hasCloudflare || hasVercel || hasAuto {
		s.logBuild(deploymentID, "⚠️  Project configured for serverless platforms (Cloudflare/Vercel)", "warning")
		s.logBuild(deploymentID, "Automatically configuring for VPS deployment...", "info")

		// Add adapter-node to package.json
		adapterVersion := `"^5.2.12"`

		// Insert after @sveltejs/kit
		newContent := strings.Replace(content,
			`"@sveltejs/kit"`,
			`"@sveltejs/adapter-node": `+adapterVersion+`,
		"@sveltejs/kit"`,
			1)

		if err := os.WriteFile(packageJSONPath, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("failed to update package.json: %w", err)
		}

		s.logBuild(deploymentID, "✓ Added @sveltejs/adapter-node to package.json", "info")

		// Remove package-lock.json since we modified package.json
		lockFilePath := filepath.Join(repoPath, "package-lock.json")
		if _, err := os.Stat(lockFilePath); err == nil {
			os.Remove(lockFilePath)
			s.logBuild(deploymentID, "Removed package-lock.json (will be regenerated)", "info")
		}
	}

	// Update or create svelte.config.js to use adapter-node
	svelteConfigPath := filepath.Join(repoPath, "svelte.config.js")

	// Check if svelte.config.js exists
	existingConfig, err := os.ReadFile(svelteConfigPath)
	if err == nil {
		// File exists, check if it's already using adapter-node
		if strings.Contains(string(existingConfig), "@sveltejs/adapter-node") {
			s.logBuild(deploymentID, "svelte.config.js already configured with adapter-node", "info")
			return nil
		}
	}

	svelteConfig := `import adapter from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter()
	}
};

export default config;
`

	if err := os.WriteFile(svelteConfigPath, []byte(svelteConfig), 0644); err != nil {
		return fmt.Errorf("failed to create svelte.config.js: %w", err)
	}

	s.logBuild(deploymentID, "✓ Updated svelte.config.js to use adapter-node", "info")

	return nil
}

func (s *DeploymentService) ensureDockerfile(repoPath string, project *models.Project) error {
	dockerfilePath := filepath.Join(repoPath, "Dockerfile")

	// Check if Dockerfile already exists
	if _, err := os.Stat(dockerfilePath); err == nil {
		return nil // Dockerfile exists
	}

	// Detect framework from package.json if outputDir not specified
	if project.OutputDir == "" {
		detectedDir := s.detectOutputDirectory(repoPath)
		if detectedDir != "" {
			project.OutputDir = detectedDir
			s.db.Save(project)
		}
	}

	// Generate Dockerfile based on framework
	dockerfile := s.generateDockerfile(project, repoPath)
	return os.WriteFile(dockerfilePath, []byte(dockerfile), 0644)
}

// detectFramework detects the framework from package.json
func (s *DeploymentService) detectFramework(repoPath string) string {
	packageJSONPath := filepath.Join(repoPath, "package.json")
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return "Unknown"
	}

	content := string(data)

	// Detect framework based on dependencies
	if strings.Contains(content, `"@angular/core"`) {
		return "Angular"
	}
	if strings.Contains(content, `"@sveltejs/kit"`) || strings.Contains(content, `"@sveltejs/adapter-node"`) {
		return "SvelteKit"
	}
	if strings.Contains(content, `"next"`) {
		return "Next.js"
	}
	if strings.Contains(content, `"nuxt"`) {
		return "Nuxt"
	}
	if strings.Contains(content, `"vite"`) && strings.Contains(content, `"react"`) {
		return "Vite + React"
	}
	if strings.Contains(content, `"vite"`) && strings.Contains(content, `"vue"`) {
		return "Vite + Vue"
	}
	if strings.Contains(content, `"vite"`) {
		return "Vite"
	}

	return "Node.js (Generic)"
}

// detectOutputDirectory detects the build output directory from package.json
func (s *DeploymentService) detectOutputDirectory(repoPath string) string {
	packageJSONPath := filepath.Join(repoPath, "package.json")
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return "build" // default
	}

	content := string(data)

	// Detect framework based on dependencies
	if strings.Contains(content, `"@angular/core"`) {
		return "dist" // Angular builds to dist/<project-name>/browser
	}
	if strings.Contains(content, `"@sveltejs/kit"`) || strings.Contains(content, `"@sveltejs/adapter-node"`) {
		return "build"
	}
	if strings.Contains(content, `"next"`) {
		return ".next"
	}
	if strings.Contains(content, `"nuxt"`) {
		return ".output"
	}
	if strings.Contains(content, `"vite"`) {
		return "dist"
	}

	return "build" // default fallback
}

func (s *DeploymentService) generateDockerfile(project *models.Project, repoPath string) string {
	nodeVersion := project.NodeVersion
	if nodeVersion == "" {
		nodeVersion = "20"
	}

	outputDir := project.OutputDir
	if outputDir == "" {
		outputDir = s.detectOutputDirectory(repoPath)
	}

	// Detect framework for specialized Dockerfile generation
	framework := s.detectFramework(repoPath)

	// Angular needs special handling for nested dist directory
	if framework == "Angular" {
		return s.generateAngularDockerfile(nodeVersion)
	}

	// For SvelteKit, we need to handle the build output differently
	if outputDir == "build" {
		return s.generateSvelteKitDockerfile(nodeVersion, outputDir)
	}

	// Generic static site Dockerfile for frameworks that output static files
	if outputDir == "dist" || outputDir == ".output" {
		return s.generateStaticSiteDockerfile(nodeVersion, outputDir)
	}

	// For Next.js and other server-based frameworks
	return s.generateServerDockerfile(nodeVersion, outputDir)
}

// generateSvelteKitDockerfile generates a SvelteKit-specific Dockerfile
func (s *DeploymentService) generateSvelteKitDockerfile(nodeVersion, outputDir string) string {
	return fmt.Sprintf(`FROM node:%s-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies with --legacy-peer-deps to handle peer dependency conflicts
RUN if [ -f package-lock.json ]; then npm ci --legacy-peer-deps; else npm install --legacy-peer-deps; fi

COPY . .
RUN npm run build

# Show build output for debugging
RUN echo "=== Build Directory Contents ===" && ls -laR /app/%s || echo "Build directory not found at /app/%s"

FROM node:%s-alpine

WORKDIR /app

# Copy the entire build output directory
COPY --from=builder /app/%s ./
COPY --from=builder /app/package*.json ./

# Install only production dependencies with --legacy-peer-deps
RUN npm install --production --legacy-peer-deps

EXPOSE 3000

ENV PORT=3000
ENV HOST=0.0.0.0
ENV NODE_ENV=production

CMD ["node", "index.js"]
`, nodeVersion, outputDir, outputDir, nodeVersion, outputDir)
}

// generateAngularDockerfile generates an Angular-specific Dockerfile
// Angular builds to dist/<project-name>/browser/
func (s *DeploymentService) generateAngularDockerfile(nodeVersion string) string {
	return fmt.Sprintf(`FROM node:%s-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --legacy-peer-deps

# Copy source code
COPY . .

# Build the Angular app
RUN npm run build

# Show the dist directory structure for debugging
RUN echo "=== Dist Directory Contents ===" && ls -laR /app/dist

# Production stage - serve with express
FROM node:%s-alpine

WORKDIR /app

# Install a simple HTTP server for Angular
RUN npm install -g http-server

# Copy the built Angular app
# Angular outputs to dist/<project-name>/browser/ so we copy everything in dist
COPY --from=builder /app/dist ./dist

EXPOSE 3000

ENV PORT=3000

# Serve the Angular app from the dist directory
# The http-server will automatically find the browser folder
CMD ["sh", "-c", "cd /app/dist && http-server -p $PORT -a 0.0.0.0 --proxy http://localhost:$PORT? $(ls -d */ | head -1)browser"]
`, nodeVersion, nodeVersion)
}

// generateStaticSiteDockerfile generates a Dockerfile for static site frameworks (Vite, etc.)
func (s *DeploymentService) generateStaticSiteDockerfile(nodeVersion, outputDir string) string {
	return fmt.Sprintf(`FROM node:%s-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --legacy-peer-deps

# Copy source code
COPY . .

# Build the app
RUN npm run build

# Show build output for debugging
RUN echo "=== Build Output ===" && ls -laR /app/%s

# Production stage - serve with a simple HTTP server
FROM node:%s-alpine

WORKDIR /app

# Install http-server for serving static files
RUN npm install -g http-server

# Copy the built static files
COPY --from=builder /app/%s ./%s

EXPOSE 3000

ENV PORT=3000

# Serve the static files
CMD ["sh", "-c", "http-server ./%s -p $PORT -a 0.0.0.0 --proxy http://localhost:$PORT?"]
`, nodeVersion, outputDir, nodeVersion, outputDir, outputDir, outputDir)
}

// generateServerDockerfile generates a Dockerfile for server-based frameworks (Next.js, etc.)
func (s *DeploymentService) generateServerDockerfile(nodeVersion, outputDir string) string {
	return fmt.Sprintf(`FROM node:%s-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --legacy-peer-deps

# Copy source code
COPY . .

# Build the app
RUN npm run build

# Show build output for debugging
RUN echo "=== Build Output ===" && ls -laR /app

# Production stage
FROM node:%s-alpine

WORKDIR /app

# Copy build output and dependencies
COPY --from=builder /app/%s ./%s
COPY --from=builder /app/package*.json ./
COPY --from=builder /app/node_modules ./node_modules

EXPOSE 3000

ENV PORT=3000
ENV HOST=0.0.0.0
ENV NODE_ENV=production

# Start the server
CMD ["npm", "start"]
`, nodeVersion, nodeVersion, outputDir, outputDir)
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

// generateSubdomain generates a unique subdomain for a project
func (s *DeploymentService) generateSubdomain(project *models.Project) string {
	// Sanitize project name for use in subdomain
	sanitized := strings.ToLower(project.Name)
	sanitized = strings.ReplaceAll(sanitized, " ", "-")
	sanitized = strings.ReplaceAll(sanitized, "_", "-")
	// Remove any non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, char := range sanitized {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			result.WriteRune(char)
		}
	}
	sanitized = result.String()

	// Add project ID to ensure uniqueness
	return fmt.Sprintf("%s-%d", sanitized, project.ID)
}

// ensureProjectDomain ensures the project has at least one active domain
// If no domain exists, it auto-generates a subdomain like Vercel does
func (s *DeploymentService) ensureProjectDomain(project *models.Project, deploymentID uint) error {
	// Check if project already has active domains
	for _, domain := range project.Domains {
		if domain.IsActive {
			s.logBuild(deploymentID, fmt.Sprintf("Using configured domain: %s", domain.Domain), "info")
			return nil
		}
	}

	// No active domains, need to create one
	// Check if we have a base domain configured
	baseDomain := s.cfg.PanelDomain
	if baseDomain == "" {
		return fmt.Errorf("no panel domain configured. Set PANEL_DOMAIN environment variable or add a custom domain to the project")
	}

	// Strip protocol and port if present
	baseDomain = strings.TrimPrefix(baseDomain, "http://")
	baseDomain = strings.TrimPrefix(baseDomain, "https://")
	if idx := strings.Index(baseDomain, ":"); idx != -1 {
		baseDomain = baseDomain[:idx]
	}

	// Generate subdomain
	subdomain := s.generateSubdomain(project)
	fullDomain := fmt.Sprintf("%s.%s", subdomain, baseDomain)

	s.logBuild(deploymentID, fmt.Sprintf("No custom domain configured. Auto-generating subdomain: %s", fullDomain), "info")

	// Create domain entry
	domain := models.Domain{
		ProjectID:  project.ID,
		Domain:     fullDomain,
		IsActive:   true,
		SSLEnabled: true,
	}

	if err := s.db.Create(&domain).Error; err != nil {
		return fmt.Errorf("failed to create auto-generated domain: %w", err)
	}

	// Add to project's domains list so Caddy can use it
	project.Domains = append(project.Domains, domain)

	s.logBuild(deploymentID, fmt.Sprintf("✓ Your app will be available at: https://%s", fullDomain), "info")

	return nil
}

// verifySSLCertificate verifies that SSL certificate was obtained by Caddy
// After Caddy reload, certificates are obtained automatically but asynchronously
// This function simply verifies the certificate is working
func (s *DeploymentService) verifySSLCertificate(project *models.Project, deploymentID uint) error {
	// Get the first active domain
	var domain string
	for _, d := range project.Domains {
		if d.IsActive {
			domain = d.Domain
			break
		}
	}

	if domain == "" {
		return fmt.Errorf("no active domain found")
	}

	// Create HTTP client with reasonable timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false, // Validate the certificate
			},
		},
	}

	url := fmt.Sprintf("https://%s", domain)
	s.logBuild(deploymentID, "Verifying SSL certificate...", "info")

	// Make a simple HEAD request to verify SSL is working
	resp, err := client.Head(url)
	if err != nil {
		return fmt.Errorf("SSL verification failed: %v", err)
	}
	defer resp.Body.Close()

	return nil
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
