package handlers

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/vps-panel/backend/internal/config"
	"github.com/vps-panel/backend/internal/models"
	"github.com/vps-panel/backend/internal/services/detector"
	"github.com/vps-panel/backend/internal/services/git"
)

type ProjectHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewProjectHandler(db *gorm.DB, cfg *config.Config) *ProjectHandler {
	return &ProjectHandler{db: db, cfg: cfg}
}

// resolveGitCredentials resolves OAuth placeholder tokens to actual credentials
func (h *ProjectHandler) resolveGitCredentials(userID uint, username, token string) (string, string) {
	// Check if token is an OAuth placeholder (e.g., "github_oauth", "gitea_oauth")
	if strings.HasSuffix(token, "_oauth") {
		providerType := strings.TrimSuffix(token, "_oauth")

		// First try the new Git Providers system
		var provider models.GitProvider
		if err := h.db.Where("user_id = ? AND type = ? AND connected = ?", userID, providerType, true).
			First(&provider).Error; err == nil {
			return provider.Username, provider.Token
		}

		// Fallback to legacy user fields for backward compatibility
		var user models.User
		if err := h.db.First(&user, userID).Error; err == nil {
			switch providerType {
			case "github":
				if user.GitHubConnected {
					return user.GitHubUsername, user.GitHubToken
				}
			case "gitea":
				if user.GiteaConnected {
					return user.GiteaUsername, user.GiteaToken
				}
			}
		}
	}

	// Return credentials as-is if not OAuth placeholder
	return username, token
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
	gitUsername, gitToken := h.resolveGitCredentials(userID, req.GitUsername, req.GitToken)

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

	// TODO: Clean up deployments, containers, and Caddy config

	if err := h.db.Delete(&project).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete project",
		})
	}

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
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
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

	domain := models.Domain{
		ProjectID:  uint(projectID),
		Domain:     req.Domain,
		IsActive:   true,
		SSLEnabled: req.SSLEnabled,
	}

	if err := h.db.Create(&domain).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add domain",
		})
	}

	// TODO: Update Caddy configuration

	return c.Status(fiber.StatusCreated).JSON(domain)
}

func (h *ProjectHandler) DeleteDomain(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, _ := strconv.ParseUint(c.Params("id"), 10, 32)
	domainID, _ := strconv.ParseUint(c.Params("domainId"), 10, 32)

	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	if err := h.db.Where("id = ? AND project_id = ?", domainID, projectID).Delete(&models.Domain{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete domain",
		})
	}

	// TODO: Update Caddy configuration

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
	req.GitUsername, req.GitToken = h.resolveGitCredentials(userID, req.GitUsername, req.GitToken)

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
	req.GitUsername, req.GitToken = h.resolveGitCredentials(userID, req.GitUsername, req.GitToken)

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
	req.GitUsername, req.GitToken = h.resolveGitCredentials(userID, req.GitUsername, req.GitToken)

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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to clone repository",
		})
	}
	defer os.RemoveAll(repoPath)

	// List subdirectories
	directories, err := h.listSubdirectories(repoPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list directories",
		})
	}

	return c.JSON(fiber.Map{
		"directories": directories,
	})
}

func (h *ProjectHandler) listSubdirectories(rootPath string) ([]string, error) {
	var directories []string

	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
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
			}
		}
	}

	return directories, nil
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}
