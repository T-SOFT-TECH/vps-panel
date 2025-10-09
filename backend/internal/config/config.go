package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// Server
	Port        string
	Host        string
	Environment string
	CorsOrigins string

	// Database
	DBDriver   string
	DBPath     string
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string

	// Docker
	DockerHost string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Caddy
	CaddyConfigPath string
	CaddyReloadCmd  string

	// Deployment
	ProjectsDir         string
	BuildTimeout        int
	MaxConcurrentBuilds int

	// Security
	JWTSecret string

	// Admin
	AdminEmail    string
	AdminPassword string

	// Webhooks
	WebhookSecret string

	// OAuth
	OAuthCallbackURL string
}

func Load() *Config {
	return &Config{
		// Server
		Port:        getEnv("PORT", "8080"),
		Host:        getEnv("HOST", "0.0.0.0"),
		Environment: getEnv("ENV", "development"),
		CorsOrigins: getEnv("CORS_ORIGINS", "http://localhost:5173,http://localhost:4173"),

		// Database
		DBDriver:   getEnv("DB_DRIVER", "sqlite"),
		DBPath:     getEnv("DB_PATH", "./data/vps-panel.db"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "vps_panel"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),

		// Docker
		DockerHost: getEnv("DOCKER_HOST", "unix:///var/run/docker.sock"),

		// Redis
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		// Caddy
		CaddyConfigPath: getEnv("CADDY_CONFIG_PATH", "/etc/caddy/sites"),
		CaddyReloadCmd:  getEnv("CADDY_RELOAD_CMD", "systemctl reload caddy"),

		// Deployment
		ProjectsDir:         getEnv("PROJECTS_DIR", "./data/projects"),
		BuildTimeout:        getEnvAsInt("BUILD_TIMEOUT", 600),
		MaxConcurrentBuilds: getEnvAsInt("MAX_CONCURRENT_BUILDS", 3),

		// Security
		JWTSecret: getEnv("JWT_SECRET", "change-this-secret-key"),

		// Admin
		AdminEmail:    getEnv("ADMIN_EMAIL", "admin@example.com"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin"),

		// Webhooks
		WebhookSecret: getEnv("WEBHOOK_SECRET", "webhook-secret"),

		// OAuth
		OAuthCallbackURL: getEnv("OAUTH_CALLBACK_URL", "http://localhost:8080/api/v1/auth/oauth/callback"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.Environment) == "development"
}

func (c *Config) IsProduction() bool {
	return strings.ToLower(c.Environment) == "production"
}
