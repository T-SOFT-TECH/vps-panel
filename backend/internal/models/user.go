package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Email        string `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string `gorm:"not null" json:"-"`
	Name         string `json:"name"`
	Role         string `gorm:"default:user" json:"role"` // admin, user

	// OAuth connections
	GitHubConnected bool   `json:"github_connected"`
	GitHubToken     string `json:"-"` // Never send to frontend
	GitHubUsername  string `json:"github_username,omitempty"`
	GitLabConnected bool   `json:"gitlab_connected"`
	GitLabToken     string `json:"-"` // Never send to frontend
	GitLabUsername  string `json:"gitlab_username,omitempty"`
	GiteaConnected  bool   `json:"gitea_connected"`
	GiteaToken      string `json:"-"` // Never send to frontend
	GiteaUsername   string `json:"gitea_username,omitempty"`
	GiteaURL        string `json:"gitea_url,omitempty"` // User's Gitea instance URL

	// Relationships
	Projects []Project `gorm:"foreignKey:UserID" json:"projects,omitempty"`
}

func (User) TableName() string {
	return "users"
}
