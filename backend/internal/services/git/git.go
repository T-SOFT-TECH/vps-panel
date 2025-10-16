package git

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type GitService struct {
	baseDir string
}

func NewGitService(baseDir string) *GitService {
	return &GitService{baseDir: baseDir}
}

type CloneOptions struct {
	URL      string
	Branch   string
	Depth    int
	Username string // For private repos
	Token    string // Access token for private repos
}

func (s *GitService) Clone(projectName string, opts CloneOptions) (string, error) {
	// Create project directory
	repoPath := filepath.Join(s.baseDir, projectName)

	// Check if directory already exists (redeployment scenario)
	if _, err := os.Stat(repoPath); err == nil {
		// Directory exists - check if it's a git repository
		if _, err := git.PlainOpen(repoPath); err == nil {
			// It's a valid git repo - pull latest changes instead of cloning
			return repoPath, s.Pull(repoPath, opts)
		}

		// Directory exists but not a git repo - remove it
		if err := os.RemoveAll(repoPath); err != nil {
			return "", fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(s.baseDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create base directory: %w", err)
	}

	// Clone repository
	cloneOpts := &git.CloneOptions{
		URL:      opts.URL,
		Progress: os.Stdout,
	}

	// Add authentication if provided (for private repos)
	if opts.Username != "" && opts.Token != "" {
		cloneOpts.Auth = &http.BasicAuth{
			Username: opts.Username,
			Password: opts.Token,
		}
	}

	if opts.Branch != "" {
		cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(opts.Branch)
		cloneOpts.SingleBranch = true
	}

	if opts.Depth > 0 {
		cloneOpts.Depth = opts.Depth
	}

	_, err := git.PlainClone(repoPath, false, cloneOpts)
	if err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	return repoPath, nil
}

func (s *GitService) Pull(repoPath string, opts CloneOptions) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Clean up OLD data directories that may have been committed to Git
	// before the .runtime/ structure was implemented.
	// These can cause permission issues during git pull.
	// NOTE: We do NOT remove .runtime/ - it's gitignored and contains persistent data!
	oldDirs := []string{
		"pb_data",           // Old PocketBase data location
		"node_modules",      // Node.js dependencies (sometimes accidentally committed)
		".next",             // Next.js build output
		"dist",              // Build output
		"build",             // Build output
		".svelte-kit",       // SvelteKit build output
		".nuxt",             // Nuxt build output
		".output",           // Nitro/Nuxt output
	}

	for _, dir := range oldDirs {
		dirPath := filepath.Join(repoPath, dir)
		if _, err := os.Stat(dirPath); err == nil {
			// Directory exists - try to remove it
			fmt.Printf("Cleaning up old directory from Git: %s\n", dirPath)
			if removeErr := forceRemoveAll(dirPath); removeErr != nil {
				// Log warning but don't fail the deployment
				fmt.Printf("Warning: Could not remove %s: %v\n", dirPath, removeErr)
			} else {
				fmt.Printf("✓ Cleaned up: %s\n", dirPath)
			}
		}
	}

	pullOpts := &git.PullOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
	}

	// Add authentication if provided (for private repos)
	if opts.Username != "" && opts.Token != "" {
		pullOpts.Auth = &http.BasicAuth{
			Username: opts.Username,
			Password: opts.Token,
		}
	}

	// Specify branch if provided
	if opts.Branch != "" {
		pullOpts.ReferenceName = plumbing.NewBranchReferenceName(opts.Branch)
	}

	err = worktree.Pull(pullOpts)

	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to pull: %w", err)
	}

	return nil
}

// forceRemoveAll forcefully removes a directory and its contents, handling permission issues
func forceRemoveAll(path string) error {
	// First try normal removal
	err := os.RemoveAll(path)
	if err == nil {
		return nil
	}

	// If that fails, try to chmod recursively then remove
	_ = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue even if we can't access some files
		}
		// Try to make everything writable
		_ = os.Chmod(filePath, 0777)
		return nil
	})

	// Try removing again after chmod
	return os.RemoveAll(path)
}

func (s *GitService) GetLatestCommit(repoPath string) (*CommitInfo, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	return &CommitInfo{
		Hash:    commit.Hash.String(),
		Message: commit.Message,
		Author:  commit.Author.Name,
		Email:   commit.Author.Email,
		Date:    commit.Author.When,
	}, nil
}

type CommitInfo struct {
	Hash    string
	Message string
	Author  string
	Email   string
	Date    interface{}
}

func (s *GitService) Cleanup(projectName string) error {
	repoPath := filepath.Join(s.baseDir, projectName)
	return os.RemoveAll(repoPath)
}

// ListBranches retrieves all remote branches for a repository
func ListBranches(repoURL, username, token string) ([]string, error) {
	remote := git.NewRemote(nil, &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repoURL},
	})

	listOpts := &git.ListOptions{}

	// Add authentication if provided (for private repos)
	if username != "" && token != "" {
		listOpts.Auth = &http.BasicAuth{
			Username: username,
			Password: token,
		}
	}

	refs, err := remote.List(listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list remote refs: %w", err)
	}

	var branches []string
	for _, ref := range refs {
		if ref.Name().IsBranch() {
			branchName := ref.Name().Short()
			branches = append(branches, branchName)
		}
	}

	return branches, nil
}
