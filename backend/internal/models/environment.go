package models

import (
	"time"

	"gorm.io/gorm"
)

type Environment struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	ProjectID uint   `gorm:"not null;index" json:"project_id"`
	Key       string `gorm:"not null" json:"key"`
	Value     string `gorm:"type:text;not null" json:"value"`
	IsSecret  bool   `gorm:"default:false" json:"is_secret"`

	// Relationships
	Project Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

func (Environment) TableName() string {
	return "environments"
}

type Domain struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	ProjectID uint   `gorm:"not null;index" json:"project_id"`
	Domain    string `gorm:"uniqueIndex;not null" json:"domain"`
	IsActive  bool   `gorm:"default:true" json:"is_active"`
	SSLEnabled bool  `gorm:"default:true" json:"ssl_enabled"`

	// Relationships
	Project Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

func (Domain) TableName() string {
	return "domains"
}

type BuildLog struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	DeploymentID uint   `gorm:"not null;index" json:"deployment_id"`
	Log          string `gorm:"type:text" json:"log"`
	LogType      string `gorm:"default:info" json:"log_type"` // info, error, warning

	// Relationships
	Deployment Deployment `gorm:"foreignKey:DeploymentID" json:"deployment,omitempty"`
}

func (BuildLog) TableName() string {
	return "build_logs"
}
