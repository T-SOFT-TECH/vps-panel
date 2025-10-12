package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/vps-panel/backend/internal/config"
	"github.com/vps-panel/backend/internal/models"
	"github.com/vps-panel/backend/internal/services/deployment"
	"github.com/vps-panel/backend/internal/services/webhook"
)

type WebhookHandler struct {
	db                *gorm.DB
	cfg               *config.Config
	deploymentService *deployment.DeploymentService
	webhookService    *webhook.Service
}

func NewWebhookHandler(db *gorm.DB, cfg *config.Config) (*WebhookHandler, error) {
	deploymentService, err := deployment.NewDeploymentService(db, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment service: %w", err)
	}

	// Get panel URL for webhook service
	panelURL := cfg.PanelURL
	if panelURL == "" {
		panelURL = fmt.Sprintf("http://localhost:%s", cfg.Port)
	}

	webhookService := webhook.NewService(panelURL)

	return &WebhookHandler{
		db:                db,
		cfg:               cfg,
		deploymentService: deploymentService,
		webhookService:    webhookService,
	}, nil
}

// GitHub webhook payload structures
type GitHubPushPayload struct {
	Ref        string `json:"ref"`
	Repository struct {
		CloneURL string `json:"clone_url"`
		HTMLURL  string `json:"html_url"`
		SSHURL   string `json:"ssh_url"`
	} `json:"repository"`
	HeadCommit struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		Author  struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
	} `json:"head_commit"`
}

// GitLab webhook payload structures
type GitLabPushPayload struct {
	Ref     string `json:"ref"`
	Project struct {
		HTTPUrl string `json:"http_url"`
		SSHUrl  string `json:"ssh_url"`
	} `json:"project"`
	Commits []struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		Author  struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
	} `json:"commits"`
}

// Gitea webhook payload structures (similar to GitHub)
type GiteaPushPayload struct {
	Ref        string `json:"ref"`
	Repository struct {
		CloneURL string `json:"clone_url"`
		HTMLURL  string `json:"html_url"`
		SSHURL   string `json:"ssh_url"`
	} `json:"repository"`
	HeadCommit struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		Author  struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
	} `json:"head_commit"`
}

// HandleGitHub processes GitHub webhook events
func (h *WebhookHandler) HandleGitHub(c *fiber.Ctx) error {
	// Get project ID from URL parameter
	projectIDStr := c.Params("project_id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	// Load project with webhook secret
	var project models.Project
	if err := h.db.First(&project, uint(projectID)).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	// Check if auto-deploy is enabled
	if !project.AutoDeploy {
		return c.JSON(fiber.Map{
			"message": "Auto-deploy is disabled for this project",
		})
	}

	// Verify GitHub signature using project's webhook secret
	signature := c.Get("X-Hub-Signature-256")
	if !h.verifyGitHubSignature(c.Body(), signature, project.WebhookSecret) {
		log.Printf("GitHub webhook: Invalid signature for project %d", project.ID)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid signature",
		})
	}

	// Parse payload
	var payload GitHubPushPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid payload",
		})
	}

	// Extract branch from ref (refs/heads/main -> main)
	branch := strings.TrimPrefix(payload.Ref, "refs/heads/")

	// Check if this is the branch we should auto-deploy
	targetBranch := project.AutoDeployBranch
	if targetBranch == "" {
		targetBranch = project.GitBranch // Default to project's main branch
	}

	if branch != targetBranch {
		return c.JSON(fiber.Map{
			"message": fmt.Sprintf("Push to %s ignored. Auto-deploy configured for %s", branch, targetBranch),
		})
	}

	// Create deployment
	now := time.Now()
	newDeployment := models.Deployment{
		ProjectID:     project.ID,
		CommitHash:    payload.HeadCommit.ID,
		CommitMessage: payload.HeadCommit.Message,
		CommitAuthor:  payload.HeadCommit.Author.Name,
		Branch:        branch,
		Status:        models.DeploymentPending,
		TriggeredBy:   "webhook-github",
		StartedAt:     &now,
	}

	if err := h.db.Create(&newDeployment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create deployment",
		})
	}

	// Trigger deployment asynchronously
	go func() {
		if err := h.deploymentService.Deploy(newDeployment.ID); err != nil {
			log.Printf("Webhook deployment failed for project %d: %v", project.ID, err)
		} else {
			log.Printf("Webhook deployment successful for project %d (deployment %d)", project.ID, newDeployment.ID)
		}
	}()

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":       "Deployment triggered successfully",
		"deployment_id": newDeployment.ID,
		"project_id":    project.ID,
		"branch":        branch,
		"commit":        payload.HeadCommit.ID[:7],
	})
}

// HandleGitLab processes GitLab webhook events
func (h *WebhookHandler) HandleGitLab(c *fiber.Ctx) error {
	// Get project ID from URL parameter
	projectIDStr := c.Params("project_id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	// Load project
	var project models.Project
	if err := h.db.First(&project, uint(projectID)).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	// Check if auto-deploy is enabled
	if !project.AutoDeploy {
		return c.JSON(fiber.Map{
			"message": "Auto-deploy is disabled for this project",
		})
	}

	// Verify GitLab token using project's webhook secret
	token := c.Get("X-Gitlab-Token")
	if token != project.WebhookSecret {
		log.Printf("GitLab webhook: Invalid token for project %d", project.ID)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token",
		})
	}

	// Parse payload
	var payload GitLabPushPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid payload",
		})
	}

	// Extract branch from ref (refs/heads/main -> main)
	branch := strings.TrimPrefix(payload.Ref, "refs/heads/")

	// Check if this is the branch we should auto-deploy
	targetBranch := project.AutoDeployBranch
	if targetBranch == "" {
		targetBranch = project.GitBranch
	}

	if branch != targetBranch {
		return c.JSON(fiber.Map{
			"message": fmt.Sprintf("Push to %s ignored. Auto-deploy configured for %s", branch, targetBranch),
		})
	}

	// Get latest commit info
	var commitID, commitMessage, commitAuthor string
	if len(payload.Commits) > 0 {
		lastCommit := payload.Commits[len(payload.Commits)-1]
		commitID = lastCommit.ID
		commitMessage = lastCommit.Message
		commitAuthor = lastCommit.Author.Name
	}

	// Create deployment
	now := time.Now()
	newDeployment := models.Deployment{
		ProjectID:     project.ID,
		CommitHash:    commitID,
		CommitMessage: commitMessage,
		CommitAuthor:  commitAuthor,
		Branch:        branch,
		Status:        models.DeploymentPending,
		TriggeredBy:   "webhook-gitlab",
		StartedAt:     &now,
	}

	if err := h.db.Create(&newDeployment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create deployment",
		})
	}

	// Trigger deployment asynchronously
	go func() {
		if err := h.deploymentService.Deploy(newDeployment.ID); err != nil {
			log.Printf("Webhook deployment failed for project %d: %v", project.ID, err)
		} else {
			log.Printf("Webhook deployment successful for project %d (deployment %d)", project.ID, newDeployment.ID)
		}
	}()

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":       "Deployment triggered successfully",
		"deployment_id": newDeployment.ID,
		"project_id":    project.ID,
		"branch":        branch,
		"commit":        commitID[:7],
	})
}

// HandleGitea processes Gitea webhook events
func (h *WebhookHandler) HandleGitea(c *fiber.Ctx) error {
	// Get project ID from URL parameter
	projectIDStr := c.Params("project_id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	// Load project
	var project models.Project
	if err := h.db.First(&project, uint(projectID)).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	// Check if auto-deploy is enabled
	if !project.AutoDeploy {
		return c.JSON(fiber.Map{
			"message": "Auto-deploy is disabled for this project",
		})
	}

	// Verify Gitea signature using project's webhook secret
	signature := c.Get("X-Gitea-Signature")
	if !h.verifyGiteaSignature(c.Body(), signature, project.WebhookSecret) {
		log.Printf("Gitea webhook: Invalid signature for project %d", project.ID)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid signature",
		})
	}

	// Parse payload (Gitea uses same format as GitHub)
	var payload GiteaPushPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid payload",
		})
	}

	// Extract branch from ref (refs/heads/main -> main)
	branch := strings.TrimPrefix(payload.Ref, "refs/heads/")

	// Check if this is the branch we should auto-deploy
	targetBranch := project.AutoDeployBranch
	if targetBranch == "" {
		targetBranch = project.GitBranch
	}

	if branch != targetBranch {
		return c.JSON(fiber.Map{
			"message": fmt.Sprintf("Push to %s ignored. Auto-deploy configured for %s", branch, targetBranch),
		})
	}

	// Create deployment
	now := time.Now()
	newDeployment := models.Deployment{
		ProjectID:     project.ID,
		CommitHash:    payload.HeadCommit.ID,
		CommitMessage: payload.HeadCommit.Message,
		CommitAuthor:  payload.HeadCommit.Author.Name,
		Branch:        branch,
		Status:        models.DeploymentPending,
		TriggeredBy:   "webhook-gitea",
		StartedAt:     &now,
	}

	if err := h.db.Create(&newDeployment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create deployment",
		})
	}

	// Trigger deployment asynchronously
	go func() {
		if err := h.deploymentService.Deploy(newDeployment.ID); err != nil {
			log.Printf("Webhook deployment failed for project %d: %v", project.ID, err)
		} else {
			log.Printf("Webhook deployment successful for project %d (deployment %d)", project.ID, newDeployment.ID)
		}
	}()

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":       "Deployment triggered successfully",
		"deployment_id": newDeployment.ID,
		"project_id":    project.ID,
		"branch":        branch,
		"commit":        payload.HeadCommit.ID[:7],
	})
}

// verifyGitHubSignature verifies GitHub webhook signature (HMAC SHA-256)
func (h *WebhookHandler) verifyGitHubSignature(payload []byte, signature string, secret string) bool {
	if signature == "" || secret == "" {
		return false
	}

	// Remove "sha256=" prefix
	signature = strings.TrimPrefix(signature, "sha256=")

	// Calculate HMAC
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}

// verifyGiteaSignature verifies Gitea webhook signature (HMAC SHA-256)
func (h *WebhookHandler) verifyGiteaSignature(payload []byte, signature string, secret string) bool {
	if signature == "" || secret == "" {
		return false
	}

	// Calculate HMAC
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}

// EnableWebhook enables webhook auto-deploy for a project and generates a secret
func (h *WebhookHandler) EnableWebhook(c *fiber.Ctx) error {
	projectID := c.Params("id")

	// Parse user ID from JWT
	userID := c.Locals("userID").(uint)

	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	// Generate webhook secret if not exists
	if project.WebhookSecret == "" {
		project.WebhookSecret = generateWebhookSecret()
	}

	// Enable auto-deploy
	project.AutoDeploy = true

	// Set auto-deploy branch to current branch if not set
	if project.AutoDeployBranch == "" {
		project.AutoDeployBranch = project.GitBranch
	}

	if err := h.db.Save(&project).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to enable webhook",
		})
	}

	// Try to automatically create webhook in Git provider
	autoCreated := false
	var autoCreateError string
	var provider models.GitProvider

	// Find the Git provider for this project
	// Try to match by Git URL hostname
	if err := h.db.Where("user_id = ? AND connected = ?", userID, true).Find(&[]models.GitProvider{}).Error; err == nil {
		var providers []models.GitProvider
		h.db.Where("user_id = ? AND connected = ?", userID, true).Find(&providers)

		// Try to find matching provider by Git URL
		for _, p := range providers {
			// Simple matching logic - can be enhanced
			if (p.Type == "github" && strings.Contains(project.GitURL, "github.com")) ||
				(p.Type == "gitlab" && strings.Contains(project.GitURL, "gitlab.com")) ||
				(p.Type == "gitea" && !strings.Contains(project.GitURL, "github.com") && !strings.Contains(project.GitURL, "gitlab.com")) {
				provider = p
				break
			}
		}

		// If we found a provider, try to create the webhook automatically
		if provider.ID != 0 {
			if err := h.webhookService.CreateWebhook(&project, &provider); err != nil {
				log.Printf("Failed to auto-create webhook for project %d: %v", project.ID, err)
				autoCreateError = err.Error()
			} else {
				autoCreated = true
				log.Printf("Successfully auto-created webhook for project %d via %s", project.ID, provider.Type)
			}
		}
	}

	// Generate webhook URLs
	baseURL := h.cfg.PanelURL
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://localhost:%s", h.cfg.Port)
	}

	response := fiber.Map{
		"message":      "Webhook enabled successfully",
		"auto_created": autoCreated,
		"webhook": fiber.Map{
			"secret": project.WebhookSecret,
			"urls": fiber.Map{
				"github": fmt.Sprintf("%s/api/v1/webhooks/github/%d", baseURL, project.ID),
				"gitlab": fmt.Sprintf("%s/api/v1/webhooks/gitlab/%d", baseURL, project.ID),
				"gitea":  fmt.Sprintf("%s/api/v1/webhooks/gitea/%d", baseURL, project.ID),
			},
			"branch": project.AutoDeployBranch,
		},
	}

	if autoCreated {
		response["message"] = "Webhook enabled and automatically configured in your Git provider!"
	} else if autoCreateError != "" {
		response["manual_setup_required"] = true
		response["auto_create_error"] = autoCreateError
		response["message"] = "Webhook enabled. Please configure manually in your Git provider."
	}

	return c.JSON(response)
}

// DisableWebhook disables webhook auto-deploy for a project
func (h *WebhookHandler) DisableWebhook(c *fiber.Ctx) error {
	projectID := c.Params("id")

	// Parse user ID from JWT
	userID := c.Locals("userID").(uint)

	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	// Try to automatically delete webhook from Git provider
	autoDeleted := false
	var provider models.GitProvider

	// Find the Git provider for this project
	if err := h.db.Where("user_id = ? AND connected = ?", userID, true).Find(&[]models.GitProvider{}).Error; err == nil {
		var providers []models.GitProvider
		h.db.Where("user_id = ? AND connected = ?", userID, true).Find(&providers)

		for _, p := range providers {
			if (p.Type == "github" && strings.Contains(project.GitURL, "github.com")) ||
				(p.Type == "gitlab" && strings.Contains(project.GitURL, "gitlab.com")) ||
				(p.Type == "gitea" && !strings.Contains(project.GitURL, "github.com") && !strings.Contains(project.GitURL, "gitlab.com")) {
				provider = p
				break
			}
		}

		// If we found a provider, try to delete the webhook automatically
		if provider.ID != 0 {
			if err := h.webhookService.DeleteWebhook(&project, &provider); err != nil {
				log.Printf("Failed to auto-delete webhook for project %d: %v", project.ID, err)
			} else {
				autoDeleted = true
				log.Printf("Successfully auto-deleted webhook for project %d from %s", project.ID, provider.Type)
			}
		}
	}

	// Disable auto-deploy
	project.AutoDeploy = false

	if err := h.db.Save(&project).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to disable webhook",
		})
	}

	message := "Webhook disabled successfully"
	if autoDeleted {
		message = "Webhook disabled and automatically removed from your Git provider"
	} else {
		message = "Webhook disabled. You may want to manually remove it from your Git provider."
	}

	return c.JSON(fiber.Map{
		"message":      message,
		"auto_deleted": autoDeleted,
	})
}

// GetWebhookInfo returns webhook configuration for a project
func (h *WebhookHandler) GetWebhookInfo(c *fiber.Ctx) error {
	projectID := c.Params("id")

	// Parse user ID from JWT
	userID := c.Locals("userID").(uint)

	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	if !project.AutoDeploy {
		return c.JSON(fiber.Map{
			"enabled": false,
		})
	}

	// Generate webhook URLs
	baseURL := h.cfg.PanelURL
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://localhost:%d", h.cfg.Port)
	}

	return c.JSON(fiber.Map{
		"enabled": true,
		"webhook": fiber.Map{
			"secret": project.WebhookSecret,
			"urls": fiber.Map{
				"github": fmt.Sprintf("%s/api/webhooks/github/%d", baseURL, project.ID),
				"gitlab": fmt.Sprintf("%s/api/webhooks/gitlab/%d", baseURL, project.ID),
				"gitea":  fmt.Sprintf("%s/api/webhooks/gitea/%d", baseURL, project.ID),
			},
			"branch": project.AutoDeployBranch,
		},
	})
}

// generateWebhookSecret generates a random secret for webhook verification
func generateWebhookSecret() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	secret := make([]byte, 32)
	for i := range secret {
		secret[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(secret)
}
