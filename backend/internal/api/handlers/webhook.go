package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/vps-panel/backend/internal/config"
	"github.com/vps-panel/backend/internal/models"
)

type WebhookHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewWebhookHandler(db *gorm.DB, cfg *config.Config) *WebhookHandler {
	return &WebhookHandler{db: db, cfg: cfg}
}

type GitHubPushPayload struct {
	Ref        string `json:"ref"`
	Repository struct {
		CloneURL string `json:"clone_url"`
		HTMLURL  string `json:"html_url"`
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

func (h *WebhookHandler) HandleGitHub(c *fiber.Ctx) error {
	// Verify GitHub signature
	signature := c.Get("X-Hub-Signature-256")
	if !h.verifyGitHubSignature(c.Body(), signature) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid signature",
		})
	}

	var payload GitHubPushPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid payload",
		})
	}

	// Extract branch from ref (refs/heads/main -> main)
	branch := strings.TrimPrefix(payload.Ref, "refs/heads/")

	// Find project by git URL and branch
	var project models.Project
	if err := h.db.Where("git_url LIKE ? AND git_branch = ? AND auto_deploy = ?",
		"%"+payload.Repository.CloneURL+"%", branch, true).
		First(&project).Error; err != nil {
		return c.JSON(fiber.Map{
			"message": "No matching project found or auto-deploy disabled",
		})
	}

	// Create deployment
	now := time.Now()
	deployment := models.Deployment{
		ProjectID:     project.ID,
		CommitHash:    payload.HeadCommit.ID,
		CommitMessage: payload.HeadCommit.Message,
		CommitAuthor:  payload.HeadCommit.Author.Name,
		Branch:        branch,
		Status:        models.DeploymentPending,
		TriggeredBy:   "webhook",
		StartedAt:     &now,
	}

	if err := h.db.Create(&deployment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create deployment",
		})
	}

	// TODO: Queue deployment job

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":       "Deployment queued",
		"deployment_id": deployment.ID,
		"project_id":    project.ID,
	})
}

func (h *WebhookHandler) HandleGitLab(c *fiber.Ctx) error {
	// Verify GitLab token
	token := c.Get("X-Gitlab-Token")
	if token != h.cfg.WebhookSecret {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token",
		})
	}

	// TODO: Implement GitLab webhook handling
	return c.SendStatus(fiber.StatusNotImplemented)
}

func (h *WebhookHandler) HandleBitbucket(c *fiber.Ctx) error {
	// TODO: Implement Bitbucket webhook handling
	return c.SendStatus(fiber.StatusNotImplemented)
}

func (h *WebhookHandler) verifyGitHubSignature(payload []byte, signature string) bool {
	if signature == "" {
		return false
	}

	// Remove "sha256=" prefix
	signature = strings.TrimPrefix(signature, "sha256=")

	// Calculate HMAC
	mac := hmac.New(sha256.New, []byte(h.cfg.WebhookSecret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}
