package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/vps-panel/backend/internal/config"
	"github.com/vps-panel/backend/internal/models"
	"github.com/vps-panel/backend/internal/services/oauth"
)

type GitProviderHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewGitProviderHandler(db *gorm.DB, cfg *config.Config) *GitProviderHandler {
	return &GitProviderHandler{db: db, cfg: cfg}
}

type CreateProviderRequest struct {
	Type         models.ProviderType `json:"type" validate:"required"`
	Name         string              `json:"name" validate:"required"`
	URL          string              `json:"url"`
	ClientID     string              `json:"client_id" validate:"required"`
	ClientSecret string              `json:"client_secret" validate:"required"`
	IsDefault    bool                `json:"is_default"`
}

type UpdateProviderRequest struct {
	Name         string `json:"name"`
	URL          string `json:"url"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	IsDefault    bool   `json:"is_default"`
}

// GetAll returns all Git providers for the current user
func (h *GitProviderHandler) GetAll(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	var providers []models.GitProvider
	if err := h.db.Where("user_id = ?", userID).Find(&providers).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch providers",
		})
	}

	// Convert to summaries (without sensitive data)
	summaries := make([]models.ProviderSummary, len(providers))
	for i, p := range providers {
		summaries[i] = p.ToSummary()
	}

	return c.JSON(fiber.Map{
		"providers": summaries,
	})
}

// Create adds a new Git provider
func (h *GitProviderHandler) Create(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	var req CreateProviderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate provider type
	if req.Type != models.ProviderGitHub && req.Type != models.ProviderGitLab && req.Type != models.ProviderGitea {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid provider type. Must be: github, gitlab, or gitea",
		})
	}

	// For self-hosted providers, URL is required
	if (req.Type == models.ProviderGitea || req.Type == models.ProviderGitLab) && req.URL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "URL is required for self-hosted providers",
		})
	}

	// If this is set as default, unset other defaults of the same type
	if req.IsDefault {
		h.db.Model(&models.GitProvider{}).
			Where("user_id = ? AND type = ?", userID, req.Type).
			Update("is_default", false)
	}

	provider := models.GitProvider{
		UserID:       userID,
		Type:         req.Type,
		Name:         req.Name,
		URL:          req.URL,
		ClientID:     req.ClientID,
		ClientSecret: req.ClientSecret,
		IsDefault:    req.IsDefault,
	}

	if err := h.db.Create(&provider).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create provider",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(provider.ToSummary())
}

// GetByID returns a specific provider
func (h *GitProviderHandler) GetByID(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	providerID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid provider ID",
		})
	}

	var provider models.GitProvider
	if err := h.db.Where("id = ? AND user_id = ?", providerID, userID).First(&provider).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Provider not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch provider",
		})
	}

	return c.JSON(provider.ToSummary())
}

// Update modifies a Git provider
func (h *GitProviderHandler) Update(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	providerID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid provider ID",
		})
	}

	var provider models.GitProvider
	if err := h.db.Where("id = ? AND user_id = ?", providerID, userID).First(&provider).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Provider not found",
		})
	}

	var req UpdateProviderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update fields if provided
	if req.Name != "" {
		provider.Name = req.Name
	}
	if req.URL != "" {
		provider.URL = req.URL
	}
	if req.ClientID != "" {
		provider.ClientID = req.ClientID
	}
	if req.ClientSecret != "" {
		provider.ClientSecret = req.ClientSecret
	}

	// Handle default flag
	if req.IsDefault && !provider.IsDefault {
		// Unset other defaults of the same type
		h.db.Model(&models.GitProvider{}).
			Where("user_id = ? AND type = ? AND id != ?", userID, provider.Type, providerID).
			Update("is_default", false)
	}
	provider.IsDefault = req.IsDefault

	if err := h.db.Save(&provider).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update provider",
		})
	}

	return c.JSON(provider.ToSummary())
}

// Delete removes a Git provider
func (h *GitProviderHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	providerID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid provider ID",
		})
	}

	var provider models.GitProvider
	if err := h.db.Where("id = ? AND user_id = ?", providerID, userID).First(&provider).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Provider not found",
		})
	}

	if err := h.db.Delete(&provider).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete provider",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// Disconnect disconnects OAuth for a provider
func (h *GitProviderHandler) Disconnect(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	providerID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid provider ID",
		})
	}

	var provider models.GitProvider
	if err := h.db.Where("id = ? AND user_id = ?", providerID, userID).First(&provider).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Provider not found",
		})
	}

	// Clear OAuth data
	provider.Connected = false
	provider.Token = ""
	provider.Username = ""

	if err := h.db.Save(&provider).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to disconnect provider",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Provider disconnected successfully",
	})
}

// ListRepositories lists repositories from a specific provider
func (h *GitProviderHandler) ListRepositories(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	providerID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid provider ID",
		})
	}

	var provider models.GitProvider
	if err := h.db.Where("id = ? AND user_id = ?", providerID, userID).First(&provider).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Provider not found",
		})
	}

	if !provider.Connected {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Provider not connected. Please connect first.",
		})
	}

	switch provider.Type {
	case models.ProviderGitHub:
		return h.listGitHubRepos(c, &provider)
	case models.ProviderGitea:
		return h.listGiteaRepos(c, &provider)
	case models.ProviderGitLab:
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
			"error": "GitLab support coming soon",
		})
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Unknown provider type",
		})
	}
}

func (h *GitProviderHandler) listGitHubRepos(c *fiber.Ctx, provider *models.GitProvider) error {
	githubService := oauth.NewGitHubService(
		provider.ClientID,
		provider.ClientSecret,
		h.cfg.OAuthCallbackURL,
	)

	repos, err := githubService.ListRepositories(provider.Token)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list repositories",
		})
	}

	return c.JSON(fiber.Map{
		"repositories": repos,
	})
}

func (h *GitProviderHandler) listGiteaRepos(c *fiber.Ctx, provider *models.GitProvider) error {
	giteaService := oauth.NewGiteaService(
		provider.URL,
		provider.ClientID,
		provider.ClientSecret,
		h.cfg.OAuthCallbackURL,
	)

	repos, err := giteaService.ListRepositories(provider.Token)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list repositories",
		})
	}

	return c.JSON(fiber.Map{
		"repositories": repos,
	})
}
