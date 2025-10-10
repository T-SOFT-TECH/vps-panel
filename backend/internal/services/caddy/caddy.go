package caddy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/vps-panel/backend/internal/models"
)

type CaddyService struct {
	configPath string
	reloadCmd  string
}

func NewCaddyService(configPath, reloadCmd string) *CaddyService {
	return &CaddyService{
		configPath: configPath,
		reloadCmd:  reloadCmd,
	}
}

const caddyConfigTemplate = `# {{ .ProjectName }}
{{ range .Domains }}{{ .Domain }}{{ if ne .Domain (index $.Domains 0).Domain }}, {{ end }}{{ end }} {
    # Enable compression
    encode gzip zstd

    {{ if .HasCustomAPI }}
    # Custom API routes
    handle /api/user/* {
        reverse_proxy 127.0.0.1:{{ $.FrontendPort }}
    }

    handle /api/admin/* {
        reverse_proxy 127.0.0.1:{{ $.FrontendPort }}
    }
    {{ end }}

    {{ if .HasBackend }}
    # Backend API routes
    handle /api/* {
        reverse_proxy 127.0.0.1:{{ $.BackendPort }}
    }

    # Backend admin panel
    handle /_/* {
        reverse_proxy 127.0.0.1:{{ $.BackendPort }}
    }
    {{ end }}

    # Frontend
    reverse_proxy 127.0.0.1:{{ $.FrontendPort }}

    # Security headers
    header {
        Strict-Transport-Security "max-age=31536000; includeSubDomains; preload"
        X-Frame-Options "DENY"
        X-Content-Type-Options "nosniff"
        Referrer-Policy "strict-origin-when-cross-origin"
    }

    # Logging
    log {
        output file /var/log/caddy/{{ $.ProjectName }}.log {
            roll_size 100MB
            roll_keep 3
        }
        format json
    }
}
`

type CaddyConfig struct {
	ProjectName  string
	Domains      []DomainConfig
	FrontendPort int
	BackendPort  int
	HasBackend   bool
	HasCustomAPI bool
}

type DomainConfig struct {
	Domain string
}

func (s *CaddyService) GenerateConfig(project *models.Project) error {
	// Ensure config directory exists
	if err := os.MkdirAll(s.configPath, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Ensure Caddy log directory exists with proper permissions
	logDir := "/var/log/caddy"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Build config data
	config := CaddyConfig{
		ProjectName:  sanitizeProjectName(project.Name),
		FrontendPort: project.FrontendPort,
		BackendPort:  project.BackendPort,
		HasBackend:   project.BaaSType != "",
		HasCustomAPI: project.Framework == models.FrameworkSvelteKit,
	}

	// Add domains
	for _, domain := range project.Domains {
		if domain.IsActive {
			config.Domains = append(config.Domains, DomainConfig{
				Domain: domain.Domain,
			})
		}
	}

	// If no domains, skip
	if len(config.Domains) == 0 {
		return fmt.Errorf("no active domains configured for project")
	}

	// Parse template
	tmpl, err := template.New("caddy").Parse(caddyConfigTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Write config file
	configFile := filepath.Join(s.configPath, fmt.Sprintf("%s.caddy", config.ProjectName))
	file, err := os.Create(configFile)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, config); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

func (s *CaddyService) RemoveConfig(projectName string) error {
	configFile := filepath.Join(s.configPath, fmt.Sprintf("%s.caddy", sanitizeProjectName(projectName)))
	if err := os.Remove(configFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config file: %w", err)
	}
	return nil
}

func (s *CaddyService) Reload() error {
	// Parse reload command
	parts := strings.Fields(s.reloadCmd)
	if len(parts) == 0 {
		return fmt.Errorf("invalid reload command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to reload Caddy: %w (output: %s)", err, string(output))
	}

	return nil
}

func sanitizeProjectName(name string) string {
	// Replace spaces and special characters with hyphens
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	return name
}
