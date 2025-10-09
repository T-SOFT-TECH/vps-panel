package models

import (
	"time"

	"gorm.io/gorm"
)

type ProviderType string

const (
	ProviderGitHub ProviderType = "github"
	ProviderGitLab ProviderType = "gitlab"
	ProviderGitea  ProviderType = "gitea"
)

// GitProvider stores user-configured Git OAuth providers
type GitProvider struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID uint `gorm:"not null;index" json:"user_id"`

	// Provider configuration
	Type         ProviderType `gorm:"type:varchar(50);not null" json:"type"` // github, gitlab, gitea
	Name         string       `gorm:"not null" json:"name"`                  // e.g., "My Gitea Server", "Company GitHub"
	URL          string       `json:"url,omitempty"`                         // For self-hosted (Gitea, GitLab)
	ClientID     string       `gorm:"not null" json:"-"`                     // Never send to frontend in GET requests
	ClientSecret string       `gorm:"not null" json:"-"`                     // Never send to frontend

	// OAuth state
	Connected bool   `gorm:"default:false" json:"connected"`
	Token     string `json:"-"` // OAuth access token
	Username  string `json:"username,omitempty"`

	// Settings
	IsDefault bool `gorm:"default:false" json:"is_default"` // Default provider for this type

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (GitProvider) TableName() string {
	return "git_providers"
}

// ProviderSummary is sent to frontend (without sensitive data)
type ProviderSummary struct {
	ID        uint         `json:"id"`
	Type      ProviderType `json:"type"`
	Name      string       `json:"name"`
	URL       string       `json:"url,omitempty"`
	Connected bool         `json:"connected"`
	Username  string       `json:"username,omitempty"`
	IsDefault bool         `json:"is_default"`
	CreatedAt time.Time    `json:"created_at"`
}

// ToSummary converts GitProvider to ProviderSummary (safe for frontend)
func (p *GitProvider) ToSummary() ProviderSummary {
	return ProviderSummary{
		ID:        p.ID,
		Type:      p.Type,
		Name:      p.Name,
		URL:       p.URL,
		Connected: p.Connected,
		Username:  p.Username,
		IsDefault: p.IsDefault,
		CreatedAt: p.CreatedAt,
	}
}
