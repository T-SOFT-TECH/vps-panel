package routes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/vps-panel/backend/internal/api/handlers"
	"github.com/vps-panel/backend/internal/api/middleware"
	"github.com/vps-panel/backend/internal/config"
)

func Setup(app *fiber.App, db *gorm.DB, cfg *config.Config) {
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg)
	projectHandler := handlers.NewProjectHandler(db, cfg)
	deploymentHandler := handlers.NewDeploymentHandler(db, cfg)
	webhookHandler := handlers.NewWebhookHandler(db, cfg)
	gitProviderHandler := handlers.NewGitProviderHandler(db, cfg)

	// API v1 routes
	api := app.Group("/api/v1")

	// Public routes
	auth := api.Group("/auth")
	auth.Get("/registration-status", authHandler.CheckRegistrationStatus)
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.RefreshToken)

	// OAuth callbacks (public - OAuth providers redirect here)
	api.Get("/auth/oauth/callback/github", authHandler.GitHubOAuthCallback)
	api.Get("/auth/oauth/callback/gitea", authHandler.GiteaOAuthCallback)

	// Protected routes (require authentication)
	protected := api.Group("", middleware.AuthMiddleware(cfg.JWTSecret))

	// OAuth routes (protected - user must be logged in to connect accounts)
	oauth := protected.Group("/auth/oauth")
	oauth.Get("/github/init", authHandler.GitHubOAuthInit)
	oauth.Get("/gitea/init", authHandler.GiteaOAuthInit)

	// User routes
	users := protected.Group("/users")
	users.Get("/me", authHandler.GetCurrentUser)
	users.Put("/me", authHandler.UpdateProfile)

	// Git Providers routes
	providers := protected.Group("/git-providers")
	providers.Get("/", gitProviderHandler.GetAll)
	providers.Post("/", gitProviderHandler.Create)
	providers.Get("/:id", gitProviderHandler.GetByID)
	providers.Put("/:id", gitProviderHandler.Update)
	providers.Delete("/:id", gitProviderHandler.Delete)
	providers.Post("/:id/disconnect", gitProviderHandler.Disconnect)
	providers.Get("/:id/repositories", gitProviderHandler.ListRepositories)

	// Project routes
	projects := protected.Group("/projects")
	projects.Get("/", projectHandler.GetAll)
	projects.Post("/detect", projectHandler.DetectFramework)    // Auto-detect framework/BaaS
	projects.Post("/branches", projectHandler.ListBranches)     // List git branches
	projects.Post("/directories", projectHandler.ListDirectories) // List repo subdirectories (monorepo)
	projects.Get("/:id", projectHandler.GetByID)
	projects.Post("/", projectHandler.Create)
	projects.Put("/:id", projectHandler.Update)
	projects.Delete("/:id", projectHandler.Delete)

	// Deployment routes
	deployments := projects.Group("/:id/deployments")
	deployments.Get("/", deploymentHandler.GetAll)
	deployments.Get("/:deploymentId", deploymentHandler.GetByID)
	deployments.Post("/", deploymentHandler.Create)
	deployments.Post("/:deploymentId/cancel", deploymentHandler.Cancel)
	deployments.Get("/:deploymentId/logs", deploymentHandler.GetLogs)

	// Environment variables
	environments := projects.Group("/:id/environments")
	environments.Get("/", projectHandler.GetEnvironments)
	environments.Post("/", projectHandler.AddEnvironment)
	environments.Put("/:envId", projectHandler.UpdateEnvironment)
	environments.Delete("/:envId", projectHandler.DeleteEnvironment)

	// Domain management
	domains := projects.Group("/:id/domains")
	domains.Get("/", projectHandler.GetDomains)
	domains.Post("/", projectHandler.AddDomain)
	domains.Put("/:domainId", projectHandler.UpdateDomain)
	domains.Delete("/:domainId", projectHandler.DeleteDomain)

	// Webhooks (no auth - validated by secret)
	webhooks := api.Group("/webhooks")
	webhooks.Post("/github", webhookHandler.HandleGitHub)
	webhooks.Post("/gitlab", webhookHandler.HandleGitLab)
	webhooks.Post("/bitbucket", webhookHandler.HandleBitbucket)
}
