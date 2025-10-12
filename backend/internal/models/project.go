package models

import (
	"time"

	"gorm.io/gorm"
)

type FrameworkType string
type BaaSType string

const (
	// Framework types
	FrameworkSvelteKit FrameworkType = "sveltekit"
	FrameworkReact     FrameworkType = "react"
	FrameworkVue       FrameworkType = "vue"
	FrameworkAngular   FrameworkType = "angular"
	FrameworkNext      FrameworkType = "nextjs"
	FrameworkNuxt      FrameworkType = "nuxt"

	// BaaS types
	BaaSPocketBase BaaSType = "pocketbase"
	BaaSSupabase   BaaSType = "supabase"
	BaaSFirebase   BaaSType = "firebase"
	BaaSAppwrite   BaaSType = "appwrite"
)

type Project struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Basic info
	Name        string `gorm:"not null" json:"name"`
	Description string `json:"description"`
	UserID      uint   `gorm:"not null;index" json:"user_id"`

	// Repository
	GitURL        string `gorm:"not null" json:"git_url"`
	GitBranch     string `gorm:"default:main" json:"git_branch"`
	GitUsername   string `json:"git_username,omitempty"`   // For private repos
	GitToken      string `json:"-"`                        // Access token (never sent to frontend)
	RootDirectory string `json:"root_directory,omitempty"` // Subdirectory for monorepos (e.g., "frontend")

	// Framework & Backend
	Framework FrameworkType `gorm:"type:varchar(50)" json:"framework"`
	BaaSType  BaaSType      `gorm:"type:varchar(50)" json:"baas_type"`

	// Build configuration
	BuildCommand   string `json:"build_command"`    // npm run build
	OutputDir      string `json:"output_dir"`       // build, dist, .next
	InstallCommand string `json:"install_command"`  // npm install
	NodeVersion    string `json:"node_version"`     // 20, 18, etc.

	// Ports
	FrontendPort int `gorm:"default:3000" json:"frontend_port"`
	BackendPort  int `gorm:"default:8090" json:"backend_port"`

	// Deployment settings
	AutoDeploy     bool   `gorm:"default:false" json:"auto_deploy"` // Webhook auto-deploy
	DeploymentPath string `json:"deployment_path"`                  // /home/user/apps/project-name
	WebhookSecret  string `json:"webhook_secret,omitempty"`         // Secret for webhook verification
	AutoDeployBranch string `json:"auto_deploy_branch,omitempty"`   // Branch to auto-deploy (defaults to GitBranch)

	// Status
	Status       string `gorm:"default:pending" json:"status"` // pending, deploying, active, failed
	LastDeployed *time.Time `json:"last_deployed,omitempty"`

	// Relationships
	User        User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Deployments []Deployment  `gorm:"foreignKey:ProjectID" json:"deployments,omitempty"`
	Environments []Environment `gorm:"foreignKey:ProjectID" json:"environments,omitempty"`
	Domains     []Domain      `gorm:"foreignKey:ProjectID" json:"domains,omitempty"`
}

func (Project) TableName() string {
	return "projects"
}
