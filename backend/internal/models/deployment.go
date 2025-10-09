package models

import (
	"time"

	"gorm.io/gorm"
)

type DeploymentStatus string

const (
	DeploymentPending   DeploymentStatus = "pending"
	DeploymentBuilding  DeploymentStatus = "building"
	DeploymentDeploying DeploymentStatus = "deploying"
	DeploymentSuccess   DeploymentStatus = "success"
	DeploymentFailed    DeploymentStatus = "failed"
	DeploymentCancelled DeploymentStatus = "cancelled"
)

type Deployment struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	ProjectID uint `gorm:"not null;index" json:"project_id"`

	// Git info
	CommitHash    string `json:"commit_hash"`
	CommitMessage string `json:"commit_message"`
	CommitAuthor  string `json:"commit_author"`
	Branch        string `json:"branch"`

	// Status
	Status        DeploymentStatus `gorm:"type:varchar(20);default:pending" json:"status"`
	StartedAt     *time.Time       `json:"started_at,omitempty"`
	CompletedAt   *time.Time       `json:"completed_at,omitempty"`
	Duration      int              `json:"duration"` // seconds
	ErrorMessage  string           `gorm:"type:text" json:"error_message,omitempty"`

	// Trigger
	TriggeredBy   string `json:"triggered_by"`   // webhook, manual, api
	TriggeredByID uint   `json:"triggered_by_id"` // user ID if manual

	// Relationships
	Project  Project    `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	BuildLogs []BuildLog `gorm:"foreignKey:DeploymentID" json:"build_logs,omitempty"`
}

func (Deployment) TableName() string {
	return "deployments"
}
