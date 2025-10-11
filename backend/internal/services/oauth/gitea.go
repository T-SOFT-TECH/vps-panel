package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
)

type GiteaService struct {
	config      *oauth2.Config
	instanceURL string
}

// NewGiteaService creates a new Gitea OAuth service
// instanceURL should be the base URL of the Gitea instance (e.g., "https://gitea.example.com")
func NewGiteaService(instanceURL, clientID, clientSecret, callbackURL string) *GiteaService {
	return &GiteaService{
		instanceURL: instanceURL,
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  callbackURL + "/gitea",
			Scopes:       []string{"read:user", "read:repository"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  instanceURL + "/login/oauth/authorize",
				TokenURL: instanceURL + "/login/oauth/access_token",
			},
		},
	}
}

func (s *GiteaService) GetAuthURL(state string) string {
	return s.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (s *GiteaService) ExchangeCode(code string) (*oauth2.Token, error) {
	return s.config.Exchange(context.Background(), code)
}

type GiteaUser struct {
	Login     string `json:"login"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	Username  string `json:"username"`
}

func (s *GiteaService) GetUser(token string) (*GiteaUser, error) {
	req, err := http.NewRequest("GET", s.instanceURL+"/api/v1/user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Use Bearer token for OAuth2
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gitea api error (status %d): %s", resp.StatusCode, string(body))
	}

	var user GiteaUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("failed to parse gitea user: %w (body: %s)", err, string(body))
	}

	return &user, nil
}

type GiteaRepo struct {
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Private       bool   `json:"private"`
	HTMLURL       string `json:"html_url"`
	CloneURL      string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
	Owner         struct {
		Username string `json:"username"`
	} `json:"owner"`
}

func (s *GiteaService) ListRepositories(token string) ([]GiteaRepo, error) {
	req, err := http.NewRequest("GET", s.instanceURL+"/api/v1/user/repos?limit=100", nil)
	if err != nil {
		return nil, err
	}

	// Gitea supports both "token" and "Bearer" authorization
	// Try Bearer first as it's more standard for OAuth2
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gitea api error (status %d): %s", resp.StatusCode, string(body))
	}

	var repos []GiteaRepo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}

	return repos, nil
}
