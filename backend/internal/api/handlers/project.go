package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/vps-panel/backend/internal/config"
	"github.com/vps-panel/backend/internal/models"
	"github.com/vps-panel/backend/internal/services/caddy"
	"github.com/vps-panel/backend/internal/services/deployment"
	"github.com/vps-panel/backend/internal/services/detector"
	"github.com/vps-panel/backend/internal/services/docker"
	"github.com/vps-panel/backend/internal/services/git"
	"github.com/vps-panel/backend/internal/services/webhook"
)

type ProjectHandler struct {
	db             *gorm.DB
	cfg            *config.Config
	webhookService *webhook.Service
}

func NewProjectHandler(db *gorm.DB, cfg *config.Config) *ProjectHandler {
	return &ProjectHandler{
		db:             db,
		cfg:            cfg,
		webhookService: webhook.NewService(),
	}
}

// resolveGitCredentials resolves OAuth placeholder tokens to actual credentials
// Returns empty strings if OAuth provider is not connected
func (h *ProjectHandler) resolveGitCredentials(userID uint, username, token string) (string, string, error) {
	// Check if token is an OAuth placeholder (e.g., "github_oauth", "gitea_oauth")
	if strings.HasSuffix(token, "_oauth") {
		providerType := strings.TrimSuffix(token, "_oauth")

		// First try the new Git Providers system
		var provider models.GitProvider
		if err := h.db.Where("user_id = ? AND type = ? AND connected = ?", userID, providerType, true).
			First(&provider).Error; err == nil {
			if h.cfg.IsDevelopment() {
				println("Resolved OAuth credentials for provider:", providerType, "Username:", provider.Username)
			}
			return provider.Username, provider.Token, nil
		}

		// Fallback to legacy user fields for backward compatibility
		var user models.User
		if err := h.db.First(&user, userID).Error; err == nil {
			switch providerType {
			case "github":
				if user.GitHubConnected {
					return user.GitHubUsername, user.GitHubToken, nil
				}
			case "gitea":
				if user.GiteaConnected {
					return user.GiteaUsername, user.GiteaToken, nil
				}
			}
		}

		// OAuth provider not found or not connected
		return "", "", fiber.NewError(fiber.StatusUnauthorized, "Git provider not connected. Please reconnect your "+providerType+" account.")
	}

	// Return credentials as-is if not OAuth placeholder
	return username, token, nil
}

type CreateProjectRequest struct {
	Name           string                `json:"name" validate:"required"`
	Description    string                `json:"description"`
	GitURL         string                `json:"git_url" validate:"required"`
	GitBranch      string                `json:"git_branch"`
	GitUsername    string                `json:"git_username"`
	GitToken       string                `json:"git_token"`
	RootDirectory  string                `json:"root_directory"`
	Framework      models.FrameworkType  `json:"framework" validate:"required"`
	BaaSType       models.BaaSType       `json:"baas_type"`
	BuildCommand   string                `json:"build_command"`
	OutputDir      string                `json:"output_dir"`
	InstallCommand string                `json:"install_command"`
	NodeVersion    string                `json:"node_version"`
	FrontendPort   int                   `json:"frontend_port"`
	BackendPort    int                   `json:"backend_port"`
	AutoDeploy     bool                  `json:"auto_deploy"`
	CustomDomain   string                `json:"custom_domain"`
}

func (h *ProjectHandler) GetAll(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	var projects []models.Project
	if err := h.db.Where("user_id = ?", userID).
		Preload("Domains").
		Order("created_at DESC").
		Find(&projects).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch projects",
		})
	}

	return c.JSON(fiber.Map{
		"projects": projects,
		"total":    len(projects),
	})
}

func (h *ProjectHandler) GetByID(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).
		Preload("Domains").
		Preload("Deployments", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(10)
		}).
		First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Project not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch project",
		})
	}

	return c.JSON(project)
}

func (h *ProjectHandler) Create(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	var req CreateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Set defaults
	if req.GitBranch == "" {
		req.GitBranch = "main"
	}
	if req.FrontendPort == 0 {
		req.FrontendPort = 3000
	}
	if req.BackendPort == 0 {
		req.BackendPort = 8090
	}
	if req.InstallCommand == "" {
		req.InstallCommand = "npm install"
	}
	if req.BuildCommand == "" {
		req.BuildCommand = "npm run build"
	}

	// Resolve OAuth placeholder tokens to actual credentials
	gitUsername, gitToken, err := h.resolveGitCredentials(userID, req.GitUsername, req.GitToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	project := models.Project{
		UserID:         userID,
		Name:           req.Name,
		Description:    req.Description,
		GitURL:         req.GitURL,
		GitBranch:      req.GitBranch,
		GitUsername:    gitUsername,
		GitToken:       gitToken,
		RootDirectory:  req.RootDirectory,
		Framework:      req.Framework,
		BaaSType:       req.BaaSType,
		BuildCommand:   req.BuildCommand,
		OutputDir:      req.OutputDir,
		InstallCommand: req.InstallCommand,
		NodeVersion:    req.NodeVersion,
		FrontendPort:   req.FrontendPort,
		BackendPort:    req.BackendPort,
		AutoDeploy:     req.AutoDeploy,
		Status:         "pending",
	}

	// Generate webhook secret if auto-deploy is enabled
	if req.AutoDeploy {
		project.WebhookSecret = generateWebhookSecret()
		if project.AutoDeployBranch == "" {
			project.AutoDeployBranch = req.GitBranch
		}
	}

	if err := h.db.Create(&project).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create project",
		})
	}

	// Create custom domain if provided
	if req.CustomDomain != "" {
		domain := models.Domain{
			ProjectID:  project.ID,
			Domain:     req.CustomDomain,
			IsActive:   true,
			SSLEnabled: true,
		}
		if err := h.db.Create(&domain).Error; err != nil {
			// Don't fail project creation if domain creation fails, just log it
			// The user can add the domain later
			println("Warning: failed to create custom domain:", err.Error())
		}
	}

	// Automatically create webhook in Git provider if auto-deploy is enabled
	// This makes the experience seamless like Vercel - no manual setup needed!
	if req.AutoDeploy && project.WebhookSecret != "" {
		// Find the connected Git provider for this project
		var providers []models.GitProvider
		if err := h.db.Where("user_id = ?", userID).Find(&providers).Error; err == nil {
			var matchingProvider *models.GitProvider

			// Match provider by Git URL
			for i := range providers {
				provider := &providers[i]
				if (strings.Contains(project.GitURL, "github.com") && provider.Type == "github") ||
					(strings.Contains(project.GitURL, "gitlab.com") && provider.Type == "gitlab") ||
					(provider.Type == "gitea" && strings.Contains(project.GitURL, provider.URL)) {
					matchingProvider = provider
					break
				}
			}

			// Auto-create webhook if we found a matching provider
			if matchingProvider != nil {
				baseURL := getBaseURL(c, h.cfg)
				if err := h.webhookService.CreateWebhook(&project, matchingProvider, baseURL); err != nil {
					log.Printf("Auto-webhook creation failed for project %d: %v (project created successfully)", project.ID, err)
					// Don't fail project creation - user can enable webhook manually later
				} else {
					log.Printf("✓ Automatically created webhook for new project %d via %s", project.ID, matchingProvider.Type)
				}
			}
		}
	}

	return c.Status(fiber.StatusCreated).JSON(project)
}

func (h *ProjectHandler) Update(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	var req CreateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update fields
	project.Name = req.Name
	project.Description = req.Description
	project.GitURL = req.GitURL
	project.GitBranch = req.GitBranch
	project.Framework = req.Framework
	project.BaaSType = req.BaaSType
	project.BuildCommand = req.BuildCommand
	project.OutputDir = req.OutputDir
	project.InstallCommand = req.InstallCommand
	project.NodeVersion = req.NodeVersion
	project.FrontendPort = req.FrontendPort
	project.BackendPort = req.BackendPort
	project.AutoDeploy = req.AutoDeploy

	if err := h.db.Save(&project).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update project",
		})
	}

	return c.JSON(project)
}

func (h *ProjectHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	// Step 1: Stop and remove Docker containers
	ctx := context.Background()
	dockerService, err := docker.NewDockerService()
	if err == nil {
		defer dockerService.Close()

		projectName := fmt.Sprintf("vps-panel-project-%d", project.ID)
		workDir := filepath.Join(h.cfg.ProjectsDir, fmt.Sprintf("project-%d", project.ID))

		// Stop docker-compose containers if they exist
		if err := dockerService.ComposeDown(ctx, workDir, projectName); err != nil {
			log.Printf("Warning: failed to stop docker containers for project %d: %v", project.ID, err)
		} else {
			log.Printf("✓ Stopped Docker containers for project %d", project.ID)
		}

		// Also try to stop individual container (for non-compose deployments)
		containerName := fmt.Sprintf("vps-panel-%s-%d", project.Name, project.ID)
		if err := dockerService.RemoveContainer(ctx, containerName); err != nil {
			log.Printf("Note: individual container cleanup for project %d: %v", project.ID, err)
		}
	}

	// Step 2: Delete project directory
	projectDir := filepath.Join(h.cfg.ProjectsDir, fmt.Sprintf("project-%d", project.ID))
	if err := os.RemoveAll(projectDir); err != nil {
		log.Printf("Warning: failed to delete project directory %s: %v", projectDir, err)
	} else {
		log.Printf("✓ Deleted project directory: %s", projectDir)
	}

	// Step 3: Remove Caddy configuration
	caddyService := caddy.NewCaddyService(h.cfg.CaddyConfigPath, h.cfg.CaddyReloadCmd)
	caddyConfigFile := filepath.Join(h.cfg.CaddyConfigPath, fmt.Sprintf("project-%d.caddy", project.ID))
	if err := os.Remove(caddyConfigFile); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Warning: failed to delete Caddy config: %v", err)
		}
	} else {
		log.Printf("✓ Deleted Caddy configuration")
		// Reload Caddy to apply changes
		if err := caddyService.Reload(); err != nil {
			log.Printf("Warning: failed to reload Caddy: %v", err)
		}
	}

	// Step 4: Delete webhook from Git provider if auto-deploy was enabled
	if project.AutoDeploy && project.WebhookSecret != "" {
		// Find the connected Git provider
		var providers []models.GitProvider
		if err := h.db.Where("user_id = ?", userID).Find(&providers).Error; err == nil {
			for i := range providers {
				provider := &providers[i]
				if (strings.Contains(project.GitURL, "github.com") && provider.Type == "github") ||
					(strings.Contains(project.GitURL, "gitlab.com") && provider.Type == "gitlab") ||
					(provider.Type == "gitea" && strings.Contains(project.GitURL, provider.URL)) {
					// Try to delete the webhook
					baseURL := getBaseURL(c, h.cfg)
					if err := h.webhookService.DeleteWebhook(&project, provider, baseURL); err != nil {
						log.Printf("Warning: failed to delete webhook for project %d: %v", project.ID, err)
					} else {
						log.Printf("✓ Deleted webhook from %s", provider.Type)
					}
					break
				}
			}
		}
	}

	// Step 5: Delete from database (cascades to deployments, domains, environments)
	if err := h.db.Delete(&project).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete project from database",
		})
	}

	log.Printf("✓ Successfully deleted project %d: %s", project.ID, project.Name)
	return c.SendStatus(fiber.StatusNoContent)
}

// Environment variable handlers
func (h *ProjectHandler) GetEnvironments(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, _ := strconv.ParseUint(c.Params("id"), 10, 32)

	// Verify project ownership
	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	var envs []models.Environment
	h.db.Where("project_id = ?", projectID).Find(&envs)

	return c.JSON(fiber.Map{
		"environments": envs,
	})
}

func (h *ProjectHandler) AddEnvironment(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, _ := strconv.ParseUint(c.Params("id"), 10, 32)

	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	var req struct {
		Key      string `json:"key" validate:"required"`
		Value    string `json:"value" validate:"required"`
		IsSecret bool   `json:"is_secret"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	env := models.Environment{
		ProjectID: uint(projectID),
		Key:       req.Key,
		Value:     req.Value,
		IsSecret:  req.IsSecret,
	}

	if err := h.db.Create(&env).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create environment variable",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(env)
}

func (h *ProjectHandler) UpdateEnvironment(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, _ := strconv.ParseUint(c.Params("id"), 10, 32)
	envID, _ := strconv.ParseUint(c.Params("envId"), 10, 32)

	// Verify ownership
	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	var env models.Environment
	if err := h.db.Where("id = ? AND project_id = ?", envID, projectID).First(&env).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Environment variable not found",
		})
	}

	var req struct {
		Value string `json:"value"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	env.Value = req.Value
	h.db.Save(&env)

	return c.JSON(env)
}

func (h *ProjectHandler) DeleteEnvironment(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, _ := strconv.ParseUint(c.Params("id"), 10, 32)
	envID, _ := strconv.ParseUint(c.Params("envId"), 10, 32)

	// Verify ownership
	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	if err := h.db.Where("id = ? AND project_id = ?", envID, projectID).Delete(&models.Environment{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete environment variable",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// Domain handlers
func (h *ProjectHandler) GetDomains(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, _ := strconv.ParseUint(c.Params("id"), 10, 32)

	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	var domains []models.Domain
	h.db.Where("project_id = ?", projectID).Find(&domains)

	return c.JSON(fiber.Map{
		"domains": domains,
	})
}

func (h *ProjectHandler) AddDomain(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, _ := strconv.ParseUint(c.Params("id"), 10, 32)

	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).
		Preload("Domains").
		First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	var req struct {
		Domain     string `json:"domain" validate:"required"`
		SSLEnabled bool   `json:"ssl_enabled"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate domain format
	if req.Domain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Domain cannot be empty",
		})
	}

	domain := models.Domain{
		ProjectID:  uint(projectID),
		Domain:     req.Domain,
		IsActive:   true,
		SSLEnabled: req.SSLEnabled,
	}

	if err := h.db.Create(&domain).Error; err != nil {
		// Check for duplicate domain error
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "UNIQUE") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Domain already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add domain",
		})
	}

	// Update Caddy configuration
	if err := h.updateCaddyForProject(&project); err != nil {
		// Log error but don't fail the request
		println("Warning: Failed to update Caddy configuration:", err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(domain)
}

func (h *ProjectHandler) UpdateDomain(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, _ := strconv.ParseUint(c.Params("id"), 10, 32)
	domainID, _ := strconv.ParseUint(c.Params("domainId"), 10, 32)

	// Verify ownership
	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).
		Preload("Domains").
		First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	var domain models.Domain
	if err := h.db.Where("id = ? AND project_id = ?", domainID, projectID).First(&domain).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Domain not found",
		})
	}

	var req struct {
		Domain     *string `json:"domain"`
		IsActive   *bool   `json:"is_active"`
		SSLEnabled *bool   `json:"ssl_enabled"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update fields if provided
	if req.Domain != nil {
		domain.Domain = *req.Domain
	}
	if req.IsActive != nil {
		domain.IsActive = *req.IsActive
	}
	if req.SSLEnabled != nil {
		domain.SSLEnabled = *req.SSLEnabled
	}

	if err := h.db.Save(&domain).Error; err != nil {
		// Check for duplicate domain error
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "UNIQUE") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Domain already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update domain",
		})
	}

	// Update Caddy configuration
	if err := h.updateCaddyForProject(&project); err != nil {
		println("Warning: Failed to update Caddy configuration:", err.Error())
	}

	return c.JSON(domain)
}

func (h *ProjectHandler) DeleteDomain(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, _ := strconv.ParseUint(c.Params("id"), 10, 32)
	domainID, _ := strconv.ParseUint(c.Params("domainId"), 10, 32)

	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).
		Preload("Domains").
		First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	// Check that we're not deleting the last domain
	var domainCount int64
	h.db.Model(&models.Domain{}).Where("project_id = ?", projectID).Count(&domainCount)
	if domainCount <= 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot delete the last domain. Projects must have at least one domain.",
		})
	}

	if err := h.db.Where("id = ? AND project_id = ?", domainID, projectID).Delete(&models.Domain{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete domain",
		})
	}

	// Update Caddy configuration
	if err := h.updateCaddyForProject(&project); err != nil {
		println("Warning: Failed to update Caddy configuration:", err.Error())
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// DetectFramework analyzes a git repository to detect framework and BaaS
func (h *ProjectHandler) DetectFramework(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	var req struct {
		GitURL        string `json:"git_url" validate:"required"`
		GitBranch     string `json:"git_branch"`
		GitUsername   string `json:"git_username"`
		GitToken      string `json:"git_token"`
		RootDirectory string `json:"root_directory"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.GitBranch == "" {
		req.GitBranch = "main"
	}

	// Resolve OAuth placeholder tokens to actual credentials
	resolvedUsername, resolvedToken, err := h.resolveGitCredentials(userID, req.GitUsername, req.GitToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	req.GitUsername = resolvedUsername
	req.GitToken = resolvedToken

	// Create temporary directory for cloning
	tempBaseDir := os.TempDir()
	projectName := "detect-" + randomString(8)

	// Clone the repository (shallow clone)
	gitService := git.NewGitService(tempBaseDir)
	repoPath, err := gitService.Clone(projectName, git.CloneOptions{
		URL:      req.GitURL,
		Branch:   req.GitBranch,
		Depth:    1, // Shallow clone
		Username: req.GitUsername,
		Token:    req.GitToken,
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to clone repository. Please check the URL and branch.",
		})
	}
	defer os.RemoveAll(repoPath)

	// If root directory specified, use that subdirectory for detection
	detectionPath := repoPath
	if req.RootDirectory != "" {
		detectionPath = filepath.Join(repoPath, req.RootDirectory)
	}

	// Detect framework and BaaS
	info, err := detector.DetectFromPath(detectionPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to analyze repository",
		})
	}

	return c.JSON(info)
}

// ListBranches lists all branches for a git repository
func (h *ProjectHandler) ListBranches(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	var req struct {
		GitURL      string `json:"git_url" validate:"required"`
		GitUsername string `json:"git_username"`
		GitToken    string `json:"git_token"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Resolve OAuth placeholder tokens to actual credentials
	resolvedUsername, resolvedToken, err := h.resolveGitCredentials(userID, req.GitUsername, req.GitToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	req.GitUsername = resolvedUsername
	req.GitToken = resolvedToken

	branches, err := git.ListBranches(req.GitURL, req.GitUsername, req.GitToken)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to list branches. Please check the repository URL and credentials.",
		})
	}

	return c.JSON(fiber.Map{
		"branches": branches,
	})
}

// ListDirectories lists subdirectories in a repository (for monorepo support)
func (h *ProjectHandler) ListDirectories(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	var req struct {
		GitURL      string `json:"git_url" validate:"required"`
		GitBranch   string `json:"git_branch"`
		GitUsername string `json:"git_username"`
		GitToken    string `json:"git_token"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.GitBranch == "" {
		req.GitBranch = "main"
	}

	// Resolve OAuth placeholder tokens
	resolvedUsername, resolvedToken, err := h.resolveGitCredentials(userID, req.GitUsername, req.GitToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	req.GitUsername = resolvedUsername
	req.GitToken = resolvedToken

	if h.cfg.IsDevelopment() {
		println("ListDirectories - Cloning:", req.GitURL, "Branch:", req.GitBranch, "HasCredentials:", req.GitUsername != "" && req.GitToken != "")
	}

	// Create temporary directory for cloning
	tempBaseDir := os.TempDir()
	projectName := "dirs-" + randomString(8)

	// Clone the repository (shallow clone)
	gitService := git.NewGitService(tempBaseDir)
	repoPath, err := gitService.Clone(projectName, git.CloneOptions{
		URL:      req.GitURL,
		Branch:   req.GitBranch,
		Depth:    1,
		Username: req.GitUsername,
		Token:    req.GitToken,
	})
	if err != nil {
		if h.cfg.IsDevelopment() {
			println("Failed to clone repository:", err.Error())
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to clone repository: " + err.Error(),
		})
	}
	defer os.RemoveAll(repoPath)

	// List subdirectories
	directories, err := h.listSubdirectories(repoPath)
	if err != nil {
		if h.cfg.IsDevelopment() {
			println("Failed to list directories:", err.Error())
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list directories: " + err.Error(),
		})
	}

	if h.cfg.IsDevelopment() {
		println("Found", len(directories), "directories with package.json:", directories)
	}

	return c.JSON(fiber.Map{
		"directories": directories,
	})
}

// CheckPocketBaseUpdate checks if a PocketBase update is available for the project
func (h *ProjectHandler) CheckPocketBaseUpdate(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	// Check if project uses PocketBase
	if project.BaaSType != models.BaaSPocketBase {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "This project does not use PocketBase",
		})
	}

	// Fetch latest version from GitHub
	latestVersion, err := deployment.FetchLatestPocketBaseVersion()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":          "Failed to fetch latest PocketBase version",
			"current_version": project.PocketBaseVersion,
			"update_available": false,
		})
	}

	// Compare versions
	currentVersion := project.PocketBaseVersion
	if currentVersion == "" {
		currentVersion = "unknown"
	}

	updateAvailable := currentVersion != latestVersion && currentVersion != "unknown"

	return c.JSON(fiber.Map{
		"current_version":   currentVersion,
		"latest_version":    latestVersion,
		"update_available":  updateAvailable,
		"project_id":        project.ID,
		"project_name":      project.Name,
	})
}

// UpdatePocketBase triggers a PocketBase update by redeploying with the latest version
func (h *ProjectHandler) UpdatePocketBase(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	// Check if project uses PocketBase
	if project.BaaSType != models.BaaSPocketBase {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "This project does not use PocketBase",
		})
	}

	// Fetch latest version
	latestVersion, err := deployment.FetchLatestPocketBaseVersion()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch latest PocketBase version from GitHub",
		})
	}

	// Check if already on latest version
	if project.PocketBaseVersion == latestVersion {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error":           "Project is already running the latest PocketBase version",
			"current_version": project.PocketBaseVersion,
			"latest_version":  latestVersion,
		})
	}

	// Create a new deployment to update PocketBase
	deployment := models.Deployment{
		ProjectID:     uint(projectID),
		Status:        models.DeploymentPending,
		CommitHash:    "pocketbase-update-" + latestVersion,
		Branch:        project.GitBranch,
		CommitMessage: fmt.Sprintf("Update PocketBase from %s to %s", project.PocketBaseVersion, latestVersion),
	}

	if err := h.db.Create(&deployment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create deployment",
		})
	}

	// Log the update request
	log.Printf("PocketBase update requested for project %d: %s → %s",
		project.ID, project.PocketBaseVersion, latestVersion)

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message":          "PocketBase update initiated",
		"deployment_id":    deployment.ID,
		"current_version":  project.PocketBaseVersion,
		"target_version":   latestVersion,
		"deployment":       deployment,
	})
}

// CreatePocketBaseAdmin creates a new admin account in PocketBase via API
func (h *ProjectHandler) CreatePocketBaseAdmin(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).
		Preload("Domains").
		First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	// Check if project uses PocketBase
	if project.BaaSType != models.BaaSPocketBase {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "This project does not use PocketBase",
		})
	}

	var req struct {
		Email           string `json:"email" validate:"required,email"`
		Password        string `json:"password" validate:"required,min=8"`
		PasswordConfirm string `json:"password_confirm" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Password != req.PasswordConfirm {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Passwords do not match",
		})
	}

	// Get PocketBase URL from project domains
	var pocketbaseURL string
	for _, domain := range project.Domains {
		if domain.IsActive {
			protocol := "https"
			if !domain.SSLEnabled {
				protocol = "http"
			}
			pocketbaseURL = fmt.Sprintf("%s://%s", protocol, domain.Domain)
			break
		}
	}

	if pocketbaseURL == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "No active domain found for project",
		})
	}

	// Create admin via PocketBase API
	adminData := map[string]interface{}{
		"email":           req.Email,
		"password":        req.Password,
		"passwordConfirm": req.PasswordConfirm,
	}

	jsonData, err := json.Marshal(adminData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to prepare admin data",
		})
	}

	// Make request to PocketBase API
	apiURL := fmt.Sprintf("%s/api/admins", pocketbaseURL)
	resp, err := http.Post(apiURL, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to connect to PocketBase: " + err.Error(),
		})
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return c.Status(resp.StatusCode).JSON(fiber.Map{
			"error":   "Failed to create admin account",
			"details": string(body),
		})
	}

	log.Printf("✓ Created PocketBase admin for project %d: %s", project.ID, req.Email)

	return c.JSON(fiber.Map{
		"message": "Admin account created successfully",
		"email":   req.Email,
		"url":     fmt.Sprintf("%s/_", pocketbaseURL),
	})
}

// ResetPocketBaseDatabase completely resets the PocketBase database
func (h *ProjectHandler) ResetPocketBaseDatabase(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	// Check if project uses PocketBase
	if project.BaaSType != models.BaaSPocketBase {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "This project does not use PocketBase",
		})
	}

	// Stop PocketBase container
	ctx := context.Background()
	dockerService, err := docker.NewDockerService()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to connect to Docker",
		})
	}
	defer dockerService.Close()

	projectName := fmt.Sprintf("vps-panel-project-%d", project.ID)
	workDir := filepath.Join(h.cfg.ProjectsDir, fmt.Sprintf("project-%d", project.ID))

	log.Printf("Stopping PocketBase container for project %d to reset database...", project.ID)

	// Stop containers
	if err := dockerService.ComposeDown(ctx, workDir, projectName); err != nil {
		log.Printf("Warning: failed to stop containers: %v", err)
	}

	// Delete pb_data directory
	pbDataDir := filepath.Join(workDir, "pb_data")
	if err := os.RemoveAll(pbDataDir); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete PocketBase data: " + err.Error(),
		})
	}

	log.Printf("✓ Deleted PocketBase data for project %d", project.ID)

	// Recreate empty pb_data directory
	if err := os.MkdirAll(pbDataDir, 0755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create fresh pb_data directory",
		})
	}

	// Restart containers
	if err := dockerService.ComposeUp(ctx, workDir, projectName); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to restart PocketBase: " + err.Error(),
		})
	}

	log.Printf("✓ Reset and restarted PocketBase for project %d", project.ID)

	return c.JSON(fiber.Map{
		"message": "PocketBase database has been reset successfully. You can now create a new admin account.",
	})
}

func (h *ProjectHandler) listSubdirectories(rootPath string) ([]string, error) {
	var directories []string

	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}

	if h.cfg.IsDevelopment() {
		println("Scanning directory:", rootPath)
		println("Total entries found:", len(entries))
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Skip hidden directories and common non-app directories
			name := entry.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == ".git" {
				continue
			}

			// Check if directory has package.json (indicating it might be an app directory)
			packageJsonPath := filepath.Join(rootPath, name, "package.json")
			if _, err := os.Stat(packageJsonPath); err == nil {
				directories = append(directories, name)
				if h.cfg.IsDevelopment() {
					println("Found app directory:", name)
				}
			}
		}
	}

	return directories, nil
}

// updateCaddyForProject regenerates Caddy configuration for a project
func (h *ProjectHandler) updateCaddyForProject(project *models.Project) error {
	// Import caddy service
	caddyService := caddy.NewCaddyService(h.cfg.CaddyConfigPath, h.cfg.CaddyReloadCmd)

	// Reload domains for the project
	var updatedProject models.Project
	if err := h.db.Where("id = ?", project.ID).
		Preload("Domains").
		First(&updatedProject).Error; err != nil {
		return err
	}

	// Check if project uses PocketBase
	if updatedProject.BaaSType == models.BaaSPocketBase {
		// Use PocketBase-specific config
		if err := caddyService.GenerateConfigWithPocketBase(&updatedProject); err != nil {
			return err
		}
	} else {
		// Generate standard Caddy config
		if err := caddyService.GenerateConfig(&updatedProject); err != nil {
			return err
		}
	}

	// Reload Caddy to apply changes
	return caddyService.Reload()
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}
