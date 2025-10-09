package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/vps-panel/backend/internal/api/middleware"
	"github.com/vps-panel/backend/internal/config"
	"github.com/vps-panel/backend/internal/models"
	"github.com/vps-panel/backend/internal/services/oauth"
)

type AuthHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewAuthHandler(db *gorm.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{db: db, cfg: cfg}
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token"`
	User         *models.User `json:"user"`
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if user already exists
	var existingUser models.User
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "User already exists",
		})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash password",
		})
	}

	// Create user
	user := models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
		Role:         "user",
	}

	if err := h.db.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	// Generate tokens
	token, refreshToken, err := h.generateTokens(&user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate tokens",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         &user,
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Find user
	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	// Generate tokens
	token, refreshToken, err := h.generateTokens(&user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate tokens",
		})
	}

	return c.JSON(AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         &user,
	})
}

func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	// TODO: Implement refresh token logic
	return c.SendStatus(fiber.StatusNotImplemented)
}

func (h *AuthHandler) GetCurrentUser(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(user)
}

func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Update fields
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	if err := h.db.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user",
		})
	}

	return c.JSON(user)
}

func (h *AuthHandler) generateTokens(user *models.User) (string, string, error) {
	// Access token (short-lived)
	claims := &middleware.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		return "", "", err
	}

	// Refresh token (long-lived)
	refreshClaims := &middleware.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	refreshTokenJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := refreshTokenJWT.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// OAuth handlers

// GitHubOAuthInit initiates GitHub OAuth flow
func (h *AuthHandler) GitHubOAuthInit(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	// Get provider ID from query parameter
	providerIDStr := c.Query("provider_id")
	if providerIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "provider_id query parameter is required",
		})
	}

	providerID, err := strconv.ParseUint(providerIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid provider_id",
		})
	}

	// Get provider configuration
	var provider models.GitProvider
	if err := h.db.Where("id = ? AND user_id = ? AND type = ?", providerID, userID, models.ProviderGitHub).First(&provider).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Provider not found",
		})
	}

	// Generate random state for CSRF protection with user ID and provider ID embedded
	stateToken, err := generateRandomState()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate state",
		})
	}

	// Combine user ID, provider ID, and random state: userID:providerID:randomState
	state := fmt.Sprintf("%d:%d:%s", userID, providerID, stateToken)

	// Store state in session/cookie for verification
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		MaxAge:   600,
	})

	githubService := oauth.NewGitHubService(
		provider.ClientID,
		provider.ClientSecret,
		h.cfg.OAuthCallbackURL,
	)

	authURL := githubService.GetAuthURL(state)
	return c.JSON(fiber.Map{
		"url": authURL,
	})
}

// GitHubOAuthCallback handles the GitHub OAuth callback
func (h *AuthHandler) GitHubOAuthCallback(c *fiber.Ctx) error {
	// Verify state
	state := c.Query("state")
	storedState := c.Cookies("oauth_state")

	if state == "" || storedState == "" || state != storedState {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid state parameter",
		})
	}

	// Extract user ID and provider ID from state (format: userID:providerID:randomToken)
	parts := strings.Split(state, ":")
	if len(parts) != 3 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid state format",
		})
	}

	userID, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID in state",
		})
	}

	providerID, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid provider ID in state",
		})
	}

	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No code provided",
		})
	}

	// Get provider configuration
	var provider models.GitProvider
	if err := h.db.Where("id = ? AND user_id = ?", providerID, userID).First(&provider).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Provider not found",
		})
	}

	githubService := oauth.NewGitHubService(
		provider.ClientID,
		provider.ClientSecret,
		h.cfg.OAuthCallbackURL,
	)

	// Exchange code for token
	token, err := githubService.ExchangeCode(code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to exchange code for token",
		})
	}

	// Get GitHub user info
	githubUser, err := githubService.GetUser(token.AccessToken)
	if err != nil {
		if h.cfg.IsDevelopment() {
			println("GitHub GetUser error:", err.Error())
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to get user info from GitHub",
			"details": err.Error(),
		})
	}

	// Update provider with OAuth connection
	provider.Connected = true
	provider.Token = token.AccessToken
	provider.Username = githubUser.Login

	if err := h.db.Save(&provider).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save GitHub connection",
		})
	}

	// Clear state cookie
	c.ClearCookie("oauth_state")

	// Get first CORS origin for redirect
	frontendURL := strings.Split(h.cfg.CorsOrigins, ",")[0]

	// Redirect to frontend with success
	return c.Redirect(frontendURL + "/settings/git-providers?connected=true")
}

// GiteaOAuthInit initiates Gitea OAuth flow
func (h *AuthHandler) GiteaOAuthInit(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	// Get provider ID from query parameter
	providerIDStr := c.Query("provider_id")
	if providerIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "provider_id query parameter is required",
		})
	}

	providerID, err := strconv.ParseUint(providerIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid provider_id",
		})
	}

	// Get provider configuration
	var provider models.GitProvider
	if err := h.db.Where("id = ? AND user_id = ? AND type = ?", providerID, userID, models.ProviderGitea).First(&provider).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Provider not found",
		})
	}

	// Gitea requires a URL
	if provider.URL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Gitea provider must have a URL configured",
		})
	}

	// Generate random state for CSRF protection with user ID and provider ID embedded
	stateToken, err := generateRandomState()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate state",
		})
	}

	// Combine user ID, provider ID, and random state: userID:providerID:randomState
	state := fmt.Sprintf("%d:%d:%s", userID, providerID, stateToken)

	// Store state in session/cookie for verification
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state_gitea",
		Value:    state,
		Path:     "/",
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		MaxAge:   600,
	})

	giteaService := oauth.NewGiteaService(
		provider.URL,
		provider.ClientID,
		provider.ClientSecret,
		h.cfg.OAuthCallbackURL,
	)

	authURL := giteaService.GetAuthURL(state)
	return c.JSON(fiber.Map{
		"url": authURL,
	})
}

// GiteaOAuthCallback handles the Gitea OAuth callback
func (h *AuthHandler) GiteaOAuthCallback(c *fiber.Ctx) error {
	// Verify state
	state := c.Query("state")
	storedState := c.Cookies("oauth_state_gitea")

	if state == "" || storedState == "" || state != storedState {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid state parameter",
		})
	}

	// Extract user ID and provider ID from state (format: userID:providerID:randomToken)
	parts := strings.Split(state, ":")
	if len(parts) != 3 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid state format",
		})
	}

	userID, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID in state",
		})
	}

	providerID, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid provider ID in state",
		})
	}

	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No code provided",
		})
	}

	// Get provider configuration
	var provider models.GitProvider
	if err := h.db.Where("id = ? AND user_id = ?", providerID, userID).First(&provider).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Provider not found",
		})
	}

	giteaService := oauth.NewGiteaService(
		provider.URL,
		provider.ClientID,
		provider.ClientSecret,
		h.cfg.OAuthCallbackURL,
	)

	// Exchange code for token
	token, err := giteaService.ExchangeCode(code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to exchange code for token",
		})
	}

	// Get Gitea user info
	giteaUser, err := giteaService.GetUser(token.AccessToken)
	if err != nil {
		if h.cfg.IsDevelopment() {
			println("Gitea GetUser error:", err.Error())
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to get user info from Gitea",
			"details": err.Error(),
		})
	}

	// Update provider with OAuth connection
	provider.Connected = true
	provider.Token = token.AccessToken
	provider.Username = giteaUser.Login

	if err := h.db.Save(&provider).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save Gitea connection",
		})
	}

	// Clear state cookie
	c.ClearCookie("oauth_state_gitea")

	// Get first CORS origin for redirect
	frontendURL := strings.Split(h.cfg.CorsOrigins, ",")[0]

	// Redirect to frontend with success
	return c.Redirect(frontendURL + "/settings/git-providers?connected=true")
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
