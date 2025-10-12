package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/vps-panel/backend/internal/models"
)

// Service handles automatic webhook creation/deletion via Git provider APIs
type Service struct{}

// NewService creates a new webhook service
func NewService() *Service {
	return &Service{}
}

// CreateWebhook automatically creates a webhook in the Git provider
// baseURL should be the panel's base URL (e.g., https://panel.example.com)
func (s *Service) CreateWebhook(project *models.Project, provider *models.GitProvider, baseURL string) error {
	switch provider.Type {
	case "github":
		return s.createGitHubWebhook(project, provider, baseURL)
	case "gitlab":
		return s.createGitLabWebhook(project, provider, baseURL)
	case "gitea":
		return s.createGiteaWebhook(project, provider, baseURL)
	default:
		return fmt.Errorf("unsupported provider type: %s", provider.Type)
	}
}

// DeleteWebhook automatically deletes a webhook from the Git provider
// baseURL should be the panel's base URL (e.g., https://panel.example.com)
func (s *Service) DeleteWebhook(project *models.Project, provider *models.GitProvider, baseURL string) error {
	switch provider.Type {
	case "github":
		return s.deleteGitHubWebhook(project, provider, baseURL)
	case "gitlab":
		return s.deleteGitLabWebhook(project, provider, baseURL)
	case "gitea":
		return s.deleteGiteaWebhook(project, provider, baseURL)
	default:
		return fmt.Errorf("unsupported provider type: %s", provider.Type)
	}
}

// GitHub Webhook Management

func (s *Service) createGitHubWebhook(project *models.Project, provider *models.GitProvider, baseURL string) error {
	// Extract owner/repo from Git URL
	owner, repo, err := parseGitHubURL(project.GitURL)
	if err != nil {
		return fmt.Errorf("failed to parse GitHub URL: %w", err)
	}

	webhookURL := fmt.Sprintf("%s/api/v1/webhooks/github/%d", strings.TrimSuffix(baseURL, "/"), project.ID)

	payload := map[string]interface{}{
		"name":   "web",
		"active": true,
		"events": []string{"push"},
		"config": map[string]interface{}{
			"url":          webhookURL,
			"content_type": "json",
			"secret":       project.WebhookSecret,
			"insecure_ssl": "0",
		},
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/hooks", owner, repo)
	return s.makeGitHubRequest("POST", apiURL, provider.Token, payload)
}

func (s *Service) deleteGitHubWebhook(project *models.Project, provider *models.GitProvider, baseURL string) error {
	// Extract owner/repo from Git URL
	owner, repo, err := parseGitHubURL(project.GitURL)
	if err != nil {
		return fmt.Errorf("failed to parse GitHub URL: %w", err)
	}

	// First, find the webhook ID
	webhookURL := fmt.Sprintf("%s/api/v1/webhooks/github/%d", strings.TrimSuffix(baseURL, "/"), project.ID)

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/hooks", owner, repo)

	// Get all webhooks
	var hooks []map[string]interface{}
	if err := s.makeGitHubRequestGet(apiURL, provider.Token, &hooks); err != nil {
		return err
	}

	// Find our webhook
	var webhookID int
	for _, hook := range hooks {
		if config, ok := hook["config"].(map[string]interface{}); ok {
			if url, ok := config["url"].(string); ok && url == webhookURL {
				if id, ok := hook["id"].(float64); ok {
					webhookID = int(id)
					break
				}
			}
		}
	}

	if webhookID == 0 {
		// Webhook not found, consider it deleted
		return nil
	}

	// Delete the webhook
	deleteURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/hooks/%d", owner, repo, webhookID)
	return s.makeGitHubRequest("DELETE", deleteURL, provider.Token, nil)
}

// GitLab Webhook Management

func (s *Service) createGitLabWebhook(project *models.Project, provider *models.GitProvider, baseURL string) error {
	projectPath, err := parseGitLabURL(project.GitURL)
	if err != nil {
		return fmt.Errorf("failed to parse GitLab URL: %w", err)
	}

	// URL encode the project path
	encodedPath := strings.ReplaceAll(projectPath, "/", "%2F")

	webhookURL := fmt.Sprintf("%s/api/v1/webhooks/gitlab/%d", strings.TrimSuffix(baseURL, "/"), project.ID)

	payload := map[string]interface{}{
		"url":                    webhookURL,
		"token":                  project.WebhookSecret,
		"push_events":            true,
		"push_events_branch_filter": project.AutoDeployBranch,
		"enable_ssl_verification": true,
	}

	apiURL := fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/hooks", encodedPath)
	return s.makeGitLabRequest("POST", apiURL, provider.Token, payload)
}

func (s *Service) deleteGitLabWebhook(project *models.Project, provider *models.GitProvider, baseURL string) error {
	projectPath, err := parseGitLabURL(project.GitURL)
	if err != nil {
		return fmt.Errorf("failed to parse GitLab URL: %w", err)
	}

	encodedPath := strings.ReplaceAll(projectPath, "/", "%2F")
	webhookURL := fmt.Sprintf("%s/api/v1/webhooks/gitlab/%d", strings.TrimSuffix(baseURL, "/"), project.ID)

	// Get all webhooks
	apiURL := fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/hooks", encodedPath)

	var hooks []map[string]interface{}
	if err := s.makeGitLabRequestGet(apiURL, provider.Token, &hooks); err != nil {
		return err
	}

	// Find our webhook
	var webhookID int
	for _, hook := range hooks {
		if url, ok := hook["url"].(string); ok && url == webhookURL {
			if id, ok := hook["id"].(float64); ok {
				webhookID = int(id)
				break
			}
		}
	}

	if webhookID == 0 {
		return nil
	}

	// Delete the webhook
	deleteURL := fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/hooks/%d", encodedPath, webhookID)
	return s.makeGitLabRequest("DELETE", deleteURL, provider.Token, nil)
}

// Gitea Webhook Management

func (s *Service) createGiteaWebhook(project *models.Project, provider *models.GitProvider, baseURL string) error {
	// Get Gitea host from provider configuration or project URL
	giteaHost, owner, repo, err := parseGiteaURL(project.GitURL)
	if err != nil {
		return fmt.Errorf("failed to parse Gitea URL: %w", err)
	}

	webhookURL := fmt.Sprintf("%s/api/v1/webhooks/gitea/%d", strings.TrimSuffix(baseURL, "/"), project.ID)

	payload := map[string]interface{}{
		"type":   "gitea",
		"active": true,
		"events": []string{"push"},
		"config": map[string]interface{}{
			"url":          webhookURL,
			"content_type": "json",
			"secret":       project.WebhookSecret,
		},
	}

	apiURL := fmt.Sprintf("%s/api/v1/repos/%s/%s/hooks", giteaHost, owner, repo)
	return s.makeGiteaRequest("POST", apiURL, provider.Token, payload)
}

func (s *Service) deleteGiteaWebhook(project *models.Project, provider *models.GitProvider, baseURL string) error {
	giteaHost, owner, repo, err := parseGiteaURL(project.GitURL)
	if err != nil {
		return fmt.Errorf("failed to parse Gitea URL: %w", err)
	}

	webhookURL := fmt.Sprintf("%s/api/v1/webhooks/gitea/%d", strings.TrimSuffix(baseURL, "/"), project.ID)

	// Get all webhooks
	apiURL := fmt.Sprintf("%s/api/v1/repos/%s/%s/hooks", giteaHost, owner, repo)

	var hooks []map[string]interface{}
	if err := s.makeGiteaRequestGet(apiURL, provider.Token, &hooks); err != nil {
		return err
	}

	// Find our webhook
	var webhookID int
	for _, hook := range hooks {
		if config, ok := hook["config"].(map[string]interface{}); ok {
			if url, ok := config["url"].(string); ok && url == webhookURL {
				if id, ok := hook["id"].(float64); ok {
					webhookID = int(id)
					break
				}
			}
		}
	}

	if webhookID == 0 {
		return nil
	}

	// Delete the webhook
	deleteURL := fmt.Sprintf("%s/api/v1/repos/%s/%s/hooks/%d", giteaHost, owner, repo, webhookID)
	return s.makeGiteaRequest("DELETE", deleteURL, provider.Token, nil)
}

// HTTP Request Helpers

func (s *Service) makeGitHubRequest(method, url, token string, payload interface{}) error {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (s *Service) makeGitHubRequestGet(url, token string, result interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func (s *Service) makeGitLabRequest(method, url, token string, payload interface{}) error {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	req.Header.Set("PRIVATE-TOKEN", token)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitLab API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (s *Service) makeGitLabRequestGet(url, token string, result interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("PRIVATE-TOKEN", token)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitLab API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func (s *Service) makeGiteaRequest(method, url, token string, payload interface{}) error {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "token "+token)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Gitea API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (s *Service) makeGiteaRequestGet(url, token string, result interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "token "+token)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Gitea API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

// URL Parsing Helpers

// parseGitHubURL extracts owner and repo from GitHub URL
// Examples:
//   - https://github.com/owner/repo.git -> owner, repo
//   - git@github.com:owner/repo.git -> owner, repo
func parseGitHubURL(gitURL string) (owner, repo string, err error) {
	// Remove .git suffix
	gitURL = strings.TrimSuffix(gitURL, ".git")

	// Handle HTTPS URLs
	if strings.HasPrefix(gitURL, "https://github.com/") {
		parts := strings.Split(strings.TrimPrefix(gitURL, "https://github.com/"), "/")
		if len(parts) >= 2 {
			return parts[0], parts[1], nil
		}
	}

	// Handle SSH URLs
	if strings.HasPrefix(gitURL, "git@github.com:") {
		parts := strings.Split(strings.TrimPrefix(gitURL, "git@github.com:"), "/")
		if len(parts) >= 2 {
			return parts[0], parts[1], nil
		}
	}

	return "", "", fmt.Errorf("invalid GitHub URL format: %s", gitURL)
}

// parseGitLabURL extracts project path from GitLab URL
// Example: https://gitlab.com/group/subgroup/project.git -> group/subgroup/project
func parseGitLabURL(gitURL string) (string, error) {
	gitURL = strings.TrimSuffix(gitURL, ".git")

	if strings.HasPrefix(gitURL, "https://gitlab.com/") {
		return strings.TrimPrefix(gitURL, "https://gitlab.com/"), nil
	}

	if strings.HasPrefix(gitURL, "git@gitlab.com:") {
		return strings.TrimPrefix(gitURL, "git@gitlab.com:"), nil
	}

	return "", fmt.Errorf("invalid GitLab URL format: %s", gitURL)
}

// parseGiteaURL extracts host, owner and repo from Gitea URL
// Example: https://gitea.example.com/owner/repo.git -> gitea.example.com, owner, repo
func parseGiteaURL(gitURL string) (host, owner, repo string, err error) {
	gitURL = strings.TrimSuffix(gitURL, ".git")

	// Handle HTTPS URLs
	if strings.HasPrefix(gitURL, "https://") {
		gitURL = strings.TrimPrefix(gitURL, "https://")
		parts := strings.SplitN(gitURL, "/", 3)
		if len(parts) >= 3 {
			return "https://" + parts[0], parts[1], parts[2], nil
		}
	}

	// Handle HTTP URLs
	if strings.HasPrefix(gitURL, "http://") {
		gitURL = strings.TrimPrefix(gitURL, "http://")
		parts := strings.SplitN(gitURL, "/", 3)
		if len(parts) >= 3 {
			return "http://" + parts[0], parts[1], parts[2], nil
		}
	}

	// Handle SSH URLs (git@host:owner/repo)
	if strings.HasPrefix(gitURL, "git@") {
		gitURL = strings.TrimPrefix(gitURL, "git@")
		// Split by : to separate host from path
		hostAndPath := strings.SplitN(gitURL, ":", 2)
		if len(hostAndPath) == 2 {
			host := hostAndPath[0]
			path := hostAndPath[1]
			parts := strings.Split(path, "/")
			if len(parts) >= 2 {
				return "https://" + host, parts[0], parts[1], nil
			}
		}
	}

	return "", "", "", fmt.Errorf("invalid Gitea URL format: %s", gitURL)
}
