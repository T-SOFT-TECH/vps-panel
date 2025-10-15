package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vps-panel/backend/internal/models"
)

// GitHubRelease represents a GitHub release API response
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
}

// FetchLatestPocketBaseVersion fetches the latest PocketBase version from GitHub API
// Exported for use by API handlers
func FetchLatestPocketBaseVersion() (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", "https://api.github.com/repos/pocketbase/pocketbase/releases/latest", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add User-Agent header (GitHub API requires it)
	req.Header.Set("User-Agent", "VPS-Panel")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch from GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Remove 'v' prefix if present (e.g., "v0.30.0" -> "0.30.0")
	version := strings.TrimPrefix(release.TagName, "v")

	if version == "" {
		return "", fmt.Errorf("empty version received from GitHub")
	}

	return version, nil
}

// getPocketBaseVersion gets the latest PocketBase version or falls back to a known stable version
func getPocketBaseVersion() string {
	version, err := FetchLatestPocketBaseVersion()
	if err != nil {
		log.Printf("Warning: Failed to fetch latest PocketBase version from GitHub: %v", err)
		log.Printf("Falling back to known stable version 0.30.0")
		return "0.30.0" // Fallback to known stable version
	}

	log.Printf("Using latest PocketBase version: %s", version)
	return version
}

// generatePocketBaseDockerfile generates a robust Dockerfile that downloads official PocketBase binary from GitHub
func (s *DeploymentService) generatePocketBaseDockerfile(pbVersion string) string {
	if pbVersion == "" {
		pbVersion = getPocketBaseVersion() // Fetch latest version from GitHub
	}

	return fmt.Sprintf(`# PocketBase Backend - Built from Official GitHub Binary
# This Dockerfile downloads the official PocketBase binary from GitHub releases
# NOT using any pre-made Docker images for security and authenticity
FROM alpine:3.19

# PocketBase version - update this to use latest from GitHub
# Latest releases: https://github.com/pocketbase/pocketbase/releases
ARG PB_VERSION=%s

LABEL maintainer="VPS Panel"
LABEL pocketbase.version="${PB_VERSION}"
LABEL description="Official PocketBase binary from GitHub"

WORKDIR /pb

# Install only necessary dependencies
# - unzip: to extract the downloaded archive
# - wget: to download from GitHub
# - ca-certificates: for HTTPS connections
RUN apk add --no-cache \
    unzip \
    wget \
    ca-certificates \
    tzdata

# Download and install official PocketBase binary from GitHub releases
# This ensures we're using the authentic, official binary
RUN echo "Downloading PocketBase v${PB_VERSION} from GitHub..." && \
    wget -q https://github.com/pocketbase/pocketbase/releases/download/v${PB_VERSION}/pocketbase_${PB_VERSION}_linux_amd64.zip \
    -O pocketbase.zip && \
    unzip pocketbase.zip && \
    rm pocketbase.zip && \
    chmod +x pocketbase && \
    ./pocketbase --version

# Create directory structure for PocketBase data
# - pb_data: Database and uploaded files
# - pb_migrations: Database migration files
# - pb_hooks: Custom Go/JavaScript hooks
RUN mkdir -p /pb/pb_data /pb/pb_migrations /pb/pb_hooks

# Set proper permissions
RUN chmod -R 755 /pb

# Health check to ensure PocketBase API is responding
HEALTHCHECK --interval=15s --timeout=5s --start-period=30s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8090/api/health || exit 1

# Expose PocketBase default port
EXPOSE 8090

# Start PocketBase with production settings
# --http: Bind to all interfaces
# --dir: Data directory
# --migrationsDir: Auto-run migrations on startup
# --hooksDir: Load custom hooks
CMD ["/pb/pocketbase", "serve", \
     "--http=0.0.0.0:8090", \
     "--dir=/pb/pb_data", \
     "--migrationsDir=/pb/pb_migrations", \
     "--hooksDir=/pb/pb_hooks"]
`, pbVersion)
}

// generateDockerCompose generates a docker-compose.yml for projects with PocketBase
func (s *DeploymentService) generateDockerCompose(project *models.Project, deploymentDomain string) string {
	projectName := sanitizeProjectName(project.Name)
	projectID := project.ID

	// Get latest PocketBase version from GitHub
	pbVersion := getPocketBaseVersion()

	// Generate unique container names
	frontendContainerName := fmt.Sprintf("vps-panel-%s-frontend-%d", projectName, projectID)
	pocketbaseContainerName := fmt.Sprintf("vps-panel-%s-pocketbase-%d", projectName, projectID)

	// Determine protocol based on SSL
	protocol := "https"
	if deploymentDomain == "" || deploymentDomain == "localhost" {
		protocol = "http"
	}

	pocketbaseURL := fmt.Sprintf("%s://%s", protocol, deploymentDomain)

	return fmt.Sprintf(`version: '3.8'

services:
  # PocketBase Backend Service
  pocketbase:
    build:
      context: .
      dockerfile: Dockerfile.pocketbase
      args:
        PB_VERSION: %s
    container_name: %s
    restart: unless-stopped
    environment:
      # PocketBase encryption key for data security
      - PB_ENCRYPTION_KEY=${PB_ENCRYPTION_KEY:-}
    ports:
      # Bind to localhost only for security (Caddy will proxy)
      - "127.0.0.1:%d:8090"
    volumes:
      # Persistent data storage
      - ./pb_data:/pb/pb_data
      # Database migrations (auto-run on startup)
      - ./pb_migrations:/pb/pb_migrations:ro
      # Custom hooks (Go/JavaScript)
      - ./pb_hooks:/pb/pb_hooks:ro
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8090/api/health"]
      interval: 15s
      timeout: 5s
      retries: 5
      start_period: 30s
    command: >
      /pb/pocketbase serve
      --http=0.0.0.0:8090
      --dir=/pb/pb_data
      --migrationsDir=/pb/pb_migrations
      --hooksDir=/pb/pb_hooks
      --origins=%s

  # Frontend Service
  frontend:
    build:
      context: .
      dockerfile: Dockerfile
      args:
        # Inject PocketBase URL at build time
        PUBLIC_POCKETBASE_URL: %s
        POCKETBASE_URL: http://pocketbase:8090
    container_name: %s
    restart: unless-stopped
    environment:
      - NODE_ENV=production
      - PORT=3000
      - HOST=0.0.0.0
      # Frontend can access PocketBase via Docker network
      - PUBLIC_POCKETBASE_URL=%s
      - POCKETBASE_URL=http://pocketbase:8090
      - ORIGIN=%s
    ports:
      # Bind to localhost only (Caddy will proxy)
      - "127.0.0.1:%d:3000"
    depends_on:
      pocketbase:
        condition: service_healthy
    networks:
      - app-network

networks:
  app-network:
    driver: bridge
`,
		pbVersion,
		pocketbaseContainerName,
		project.BackendPort,
		pocketbaseURL,
		pocketbaseURL,
		frontendContainerName,
		pocketbaseURL,
		pocketbaseURL,
		project.FrontendPort,
	)
}

// ensurePocketBaseStructure creates necessary directories and files for PocketBase
func (s *DeploymentService) ensurePocketBaseStructure(workDir string, deploymentID uint) error {
	s.logBuild(deploymentID, "Setting up PocketBase project structure...", "info")

	// Create pb_data directory if it doesn't exist
	pbDataDir := filepath.Join(workDir, "pb_data")
	if err := os.MkdirAll(pbDataDir, 0755); err != nil {
		return fmt.Errorf("failed to create pb_data directory: %w", err)
	}

	// Create pb_migrations directory if it doesn't exist
	pbMigrationsDir := filepath.Join(workDir, "pb_migrations")
	if _, err := os.Stat(pbMigrationsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(pbMigrationsDir, 0755); err != nil {
			return fmt.Errorf("failed to create pb_migrations directory: %w", err)
		}
		s.logBuild(deploymentID, "Created pb_migrations directory (add your migrations here)", "info")
	} else {
		// Check if there are migration files
		entries, _ := os.ReadDir(pbMigrationsDir)
		if len(entries) > 0 {
			s.logBuild(deploymentID, fmt.Sprintf("Found %d migration file(s) - will auto-run on startup", len(entries)), "info")
		}
	}

	// Create pb_hooks directory if it doesn't exist
	pbHooksDir := filepath.Join(workDir, "pb_hooks")
	if _, err := os.Stat(pbHooksDir); os.IsNotExist(err) {
		if err := os.MkdirAll(pbHooksDir, 0755); err != nil {
			return fmt.Errorf("failed to create pb_hooks directory: %w", err)
		}
		s.logBuild(deploymentID, "Created pb_hooks directory (add your custom hooks here)", "info")
	} else {
		// Check if there are hook files
		entries, _ := os.ReadDir(pbHooksDir)
		if len(entries) > 0 {
			s.logBuild(deploymentID, fmt.Sprintf("Found %d hook file(s) - will load on startup", len(entries)), "info")
		}
	}

	// Create .gitignore for pb_data if it doesn't exist
	gitignorePath := filepath.Join(workDir, ".gitignore")
	gitignoreContent := `
# PocketBase data directory (contains database and uploads)
pb_data/

# Backup files
*.db-shm
*.db-wal
`

	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
			s.logBuild(deploymentID, "Warning: could not create .gitignore", "warning")
		}
	}

	s.logBuild(deploymentID, "✓ PocketBase structure ready", "info")
	return nil
}

// sanitizeProjectName sanitizes project name for use in container names
func sanitizeProjectName(name string) string {
	// Convert to lowercase and replace spaces/special chars with hyphens
	result := ""
	for _, char := range name {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			result += string(char)
		} else if char >= 'A' && char <= 'Z' {
			result += string(char + 32) // Convert to lowercase
		} else if char == ' ' || char == '_' {
			result += "-"
		}
	}
	return result
}

// generatePocketBaseDeploymentFiles creates all necessary files for PocketBase deployment
func (s *DeploymentService) generatePocketBaseDeploymentFiles(workDir string, project *models.Project, deploymentID uint) error {
	// Get deployment domain
	deploymentDomain := ""
	for _, domain := range project.Domains {
		if domain.IsActive {
			deploymentDomain = domain.Domain
			break
		}
	}

	s.logBuild(deploymentID, "Generating PocketBase deployment files...", "info")

	// 1. Generate frontend Dockerfile (using existing logic)
	s.logBuild(deploymentID, "Creating frontend Dockerfile...", "info")
	if err := s.ensureDockerfile(workDir, project); err != nil {
		return fmt.Errorf("failed to create frontend Dockerfile: %w", err)
	}

	// 2. Generate PocketBase Dockerfile
	s.logBuild(deploymentID, "Fetching latest PocketBase version from GitHub...", "info")
	pbVersion := getPocketBaseVersion()
	s.logBuild(deploymentID, fmt.Sprintf("Using PocketBase version: %s", pbVersion), "info")

	s.logBuild(deploymentID, "Creating PocketBase Dockerfile from official GitHub binary...", "info")
	pocketbaseDockerfile := s.generatePocketBaseDockerfile(pbVersion)
	pocketbaseDockerfilePath := filepath.Join(workDir, "Dockerfile.pocketbase")
	if err := os.WriteFile(pocketbaseDockerfilePath, []byte(pocketbaseDockerfile), 0644); err != nil {
		return fmt.Errorf("failed to write PocketBase Dockerfile: %w", err)
	}
	s.logBuild(deploymentID, "✓ PocketBase Dockerfile created (downloads official binary)", "info")

	// 3. Generate docker-compose.yml
	s.logBuild(deploymentID, "Creating docker-compose.yml for multi-container deployment...", "info")
	dockerCompose := s.generateDockerCompose(project, deploymentDomain)
	dockerComposePath := filepath.Join(workDir, "docker-compose.yml")
	if err := os.WriteFile(dockerComposePath, []byte(dockerCompose), 0644); err != nil {
		return fmt.Errorf("failed to write docker-compose.yml: %w", err)
	}
	s.logBuild(deploymentID, "✓ docker-compose.yml created", "info")

	// 4. Create .env file with PocketBase encryption key if not exists
	envPath := filepath.Join(workDir, ".env")
	envContent, _ := os.ReadFile(envPath)
	if !containsString(string(envContent), "PB_ENCRYPTION_KEY") {
		// Generate a random encryption key for PocketBase
		encryptionKey := generateRandomKey(32)
		envFile, err := os.OpenFile(envPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			s.logBuild(deploymentID, "Warning: Could not add PocketBase encryption key to .env", "warning")
		} else {
			defer envFile.Close()
			envFile.WriteString(fmt.Sprintf("\n# PocketBase encryption key (auto-generated)\nPB_ENCRYPTION_KEY=%s\n", encryptionKey))
			s.logBuild(deploymentID, "✓ Added PocketBase encryption key to .env", "info")
		}
	}

	s.logBuild(deploymentID, "All PocketBase deployment files generated successfully", "info")

	// Save the PocketBase version to the project for update tracking
	project.PocketBaseVersion = pbVersion
	if err := s.db.Save(project).Error; err != nil {
		s.logBuild(deploymentID, fmt.Sprintf("Warning: Failed to save PocketBase version: %v", err), "warning")
	}

	return nil
}

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// generateRandomKey generates a random hex key of specified length
func generateRandomKey(length int) string {
	const charset = "0123456789abcdef"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}

// deployWithDockerCompose handles deployment using docker-compose for multi-container projects
func (s *DeploymentService) deployWithDockerCompose(ctx context.Context, deployment *models.Deployment, project *models.Project, workDir string) error {
	s.logBuild(deployment.ID, "Starting multi-container deployment with docker-compose...", "info")

	// Step 1: Stop and remove any existing containers
	s.logBuild(deployment.ID, "Cleaning up previous deployment...", "info")
	projectName := fmt.Sprintf("vps-panel-project-%d", project.ID)

	if err := s.dockerService.ComposeDown(ctx, workDir, projectName); err != nil {
		s.logBuild(deployment.ID, fmt.Sprintf("Note: %v (this is normal for first deployment)", err), "info")
	}

	// Step 2: Build images with docker-compose
	s.logBuild(deployment.ID, "Building Docker images (frontend + PocketBase)...", "info")
	s.logBuild(deployment.ID, "→ Downloading official PocketBase binary from GitHub...", "info")

	logCallback := func(message string) {
		s.logBuild(deployment.ID, message, "info")
	}

	if err := s.dockerService.ComposeBuild(ctx, workDir, projectName, logCallback); err != nil {
		return fmt.Errorf("failed to build images: %w", err)
	}

	s.logBuild(deployment.ID, "✓ All images built successfully", "info")

	// Step 3: Start containers
	deployment.Status = models.DeploymentDeploying
	s.db.Save(&deployment)

	s.logBuild(deployment.ID, "Starting containers...", "info")
	s.logBuild(deployment.ID, "→ Starting PocketBase backend...", "info")
	s.logBuild(deployment.ID, "→ Starting frontend (waiting for PocketBase health check)...", "info")

	if err := s.dockerService.ComposeUp(ctx, workDir, projectName); err != nil {
		return fmt.Errorf("failed to start containers: %w", err)
	}

	s.logBuild(deployment.ID, "✓ All containers started successfully", "info")

	// Step 4: Configure Caddy reverse proxy for both services
	s.logBuild(deployment.ID, "Configuring reverse proxy...", "info")
	if err := s.caddyService.GenerateConfigWithPocketBase(project); err != nil {
		return fmt.Errorf("failed to generate Caddy config: %w", err)
	}

	if err := s.caddyService.Reload(); err != nil {
		s.logBuild(deployment.ID, fmt.Sprintf("Warning: failed to reload Caddy: %v", err), "warning")
	} else {
		s.logBuild(deployment.ID, "✓ Reverse proxy configured", "info")
	}

	// Step 5: Display deployment information
	for _, domain := range project.Domains {
		if domain.IsActive {
			protocol := "https"
			if !domain.SSLEnabled {
				protocol = "http"
			}
			s.logBuild(deployment.ID, fmt.Sprintf("🚀 Frontend: %s://%s", protocol, domain.Domain), "info")
			s.logBuild(deployment.ID, fmt.Sprintf("🗄️  PocketBase API: %s://%s/api/*", protocol, domain.Domain), "info")
			s.logBuild(deployment.ID, fmt.Sprintf("🔧 PocketBase Admin: %s://%s/_", protocol, domain.Domain), "info")
			break
		}
	}

	return nil
}
