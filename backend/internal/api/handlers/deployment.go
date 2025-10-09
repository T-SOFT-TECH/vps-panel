package handlers

import (
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/vps-panel/backend/internal/config"
	"github.com/vps-panel/backend/internal/models"
	"github.com/vps-panel/backend/internal/services/deployment"
)

type DeploymentHandler struct {
	db                *gorm.DB
	cfg               *config.Config
	deploymentService *deployment.DeploymentService
}

func NewDeploymentHandler(db *gorm.DB, cfg *config.Config) *DeploymentHandler {
	deploymentService, err := deployment.NewDeploymentService(db, cfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize deployment service: %v", err)
		log.Println("Deployments will be queued but not executed")
	}

	return &DeploymentHandler{
		db:                db,
		cfg:               cfg,
		deploymentService: deploymentService,
	}
}

func (h *DeploymentHandler) GetAll(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, _ := strconv.ParseUint(c.Params("id"), 10, 32)

	// Verify project ownership
	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	var deployments []models.Deployment
	if err := h.db.Where("project_id = ?", projectID).
		Order("created_at DESC").
		Find(&deployments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch deployments",
		})
	}

	return c.JSON(fiber.Map{
		"deployments": deployments,
		"total":       len(deployments),
	})
}

func (h *DeploymentHandler) GetByID(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, _ := strconv.ParseUint(c.Params("id"), 10, 32)
	deploymentID, _ := strconv.ParseUint(c.Params("deploymentId"), 10, 32)

	// Verify project ownership
	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	var deployment models.Deployment
	if err := h.db.Where("id = ? AND project_id = ?", deploymentID, projectID).
		Preload("BuildLogs").
		First(&deployment).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Deployment not found",
		})
	}

	return c.JSON(deployment)
}

func (h *DeploymentHandler) Create(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, _ := strconv.ParseUint(c.Params("id"), 10, 32)

	// Verify project ownership
	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	// Create deployment record
	now := time.Now()
	deployment := models.Deployment{
		ProjectID:     uint(projectID),
		Branch:        project.GitBranch,
		Status:        models.DeploymentPending,
		TriggeredBy:   "manual",
		TriggeredByID: userID,
		StartedAt:     &now,
	}

	if err := h.db.Create(&deployment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create deployment",
		})
	}

	// Execute deployment asynchronously
	if h.deploymentService != nil {
		go func() {
			if err := h.deploymentService.Deploy(deployment.ID); err != nil {
				log.Printf("Deployment %d failed: %v", deployment.ID, err)
			}
		}()
	} else {
		log.Printf("Warning: Deployment %d created but deployment service not available", deployment.ID)
	}

	return c.Status(fiber.StatusCreated).JSON(deployment)
}

func (h *DeploymentHandler) Cancel(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, _ := strconv.ParseUint(c.Params("id"), 10, 32)
	deploymentID, _ := strconv.ParseUint(c.Params("deploymentId"), 10, 32)

	// Verify project ownership
	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	var deployment models.Deployment
	if err := h.db.Where("id = ? AND project_id = ?", deploymentID, projectID).First(&deployment).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Deployment not found",
		})
	}

	// Can only cancel pending or building deployments
	if deployment.Status != models.DeploymentPending && deployment.Status != models.DeploymentBuilding {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot cancel deployment in current state",
		})
	}

	deployment.Status = models.DeploymentCancelled
	now := time.Now()
	deployment.CompletedAt = &now

	if err := h.db.Save(&deployment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to cancel deployment",
		})
	}

	// TODO: Cancel the actual deployment process

	return c.JSON(deployment)
}

func (h *DeploymentHandler) GetLogs(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, _ := strconv.ParseUint(c.Params("id"), 10, 32)
	deploymentID, _ := strconv.ParseUint(c.Params("deploymentId"), 10, 32)

	// Verify project ownership
	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	var logs []models.BuildLog
	if err := h.db.Where("deployment_id = ?", deploymentID).
		Order("created_at ASC").
		Find(&logs).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch logs",
		})
	}

	return c.JSON(fiber.Map{
		"logs":  logs,
		"total": len(logs),
	})
}
