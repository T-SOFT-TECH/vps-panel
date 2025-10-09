package detector

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/vps-panel/backend/internal/models"
)

type ProjectInfo struct {
	Framework      models.FrameworkType `json:"framework"`
	BaaSType       models.BaaSType      `json:"baas_type"`
	Detected       bool                 `json:"detected"`
	BuildCommand   string               `json:"build_command"`
	InstallCommand string               `json:"install_command"`
	OutputDir      string               `json:"output_dir"`
	StartCommand   string               `json:"start_command"`
	DevCommand     string               `json:"dev_command"`
	FrontendPort   int                  `json:"frontend_port"`
	BackendPort    int                  `json:"backend_port"`
	NodeVersion    string               `json:"node_version"`
}

// DetectFromPath analyzes a cloned repository to detect framework and BaaS
func DetectFromPath(repoPath string) (*ProjectInfo, error) {
	info := &ProjectInfo{
		Detected: false,
	}

	// First, try detecting from root directory
	packageJSONPath := filepath.Join(repoPath, "package.json")
	if _, err := os.Stat(packageJSONPath); err == nil {
		detectFromPackageJSON(packageJSONPath, info)
	}

	// Check for other framework indicators in root
	detectFromFileStructure(repoPath, info)

	// If nothing detected in root, check common subdirectories (monorepo structure)
	if !info.Detected {
		commonDirs := []string{"frontend", "client", "web", "app", "packages/frontend", "packages/client"}
		for _, dir := range commonDirs {
			subPath := filepath.Join(repoPath, dir)
			if dirExists(subPath) {
				// Try package.json in subdirectory
				subPackageJSONPath := filepath.Join(subPath, "package.json")
				if fileExists(subPackageJSONPath) {
					detectFromPackageJSON(subPackageJSONPath, info)
					if info.Detected {
						break
					}
				}

				// Try file structure detection in subdirectory
				detectFromFileStructure(subPath, info)
				if info.Detected {
					break
				}
			}
		}
	}

	// Check for BaaS in backend directories if not found yet
	if info.BaaSType == "" {
		backendDirs := []string{"backend", "server", "api", "packages/backend"}
		for _, dir := range backendDirs {
			subPath := filepath.Join(repoPath, dir)
			if dirExists(subPath) {
				detectBaaSFromPath(subPath, info)
				if info.BaaSType != "" {
					break
				}
			}
		}
	}

	// Set framework-specific configurations
	if info.Detected {
		setFrameworkDefaults(info)
	}

	// Set BaaS-specific configurations
	if info.BaaSType != "" {
		setBaaSDefaults(info)
	}

	return info, nil
}

// setFrameworkDefaults sets build configurations based on detected framework
func setFrameworkDefaults(info *ProjectInfo) {
	switch info.Framework {
	case "sveltekit":
		info.BuildCommand = "npm run build"
		info.InstallCommand = "npm install"
		info.OutputDir = "build"
		info.StartCommand = "node build"
		info.DevCommand = "npm run dev"
		info.FrontendPort = 3000
		info.NodeVersion = "20"

	case "nextjs":
		info.BuildCommand = "npm run build"
		info.InstallCommand = "npm install"
		info.OutputDir = ".next"
		info.StartCommand = "npm start"
		info.DevCommand = "npm run dev"
		info.FrontendPort = 3000
		info.NodeVersion = "20"

	case "nuxt":
		info.BuildCommand = "npm run build"
		info.InstallCommand = "npm install"
		info.OutputDir = ".output"
		info.StartCommand = "node .output/server/index.mjs"
		info.DevCommand = "npm run dev"
		info.FrontendPort = 3000
		info.NodeVersion = "20"

	case "react":
		info.BuildCommand = "npm run build"
		info.InstallCommand = "npm install"
		info.OutputDir = "dist"
		info.StartCommand = "npm run preview"
		info.DevCommand = "npm run dev"
		info.FrontendPort = 5173 // Vite default
		info.NodeVersion = "20"

	case "vue":
		info.BuildCommand = "npm run build"
		info.InstallCommand = "npm install"
		info.OutputDir = "dist"
		info.StartCommand = "npm run preview"
		info.DevCommand = "npm run dev"
		info.FrontendPort = 5173 // Vite default
		info.NodeVersion = "20"

	case "angular":
		info.BuildCommand = "ng build"
		info.InstallCommand = "npm install"
		info.OutputDir = "dist"
		info.StartCommand = "npm start"
		info.DevCommand = "ng serve"
		info.FrontendPort = 4200
		info.NodeVersion = "20"

	default:
		// Generic defaults
		info.BuildCommand = "npm run build"
		info.InstallCommand = "npm install"
		info.OutputDir = "dist"
		info.StartCommand = "npm start"
		info.DevCommand = "npm run dev"
		info.FrontendPort = 3000
		info.NodeVersion = "20"
	}
}

// setBaaSDefaults sets backend port based on BaaS type
func setBaaSDefaults(info *ProjectInfo) {
	switch info.BaaSType {
	case "pocketbase":
		info.BackendPort = 8090

	case "supabase":
		info.BackendPort = 54321 // Supabase local default

	case "firebase":
		info.BackendPort = 9099 // Firebase emulator default

	case "appwrite":
		info.BackendPort = 80 // Appwrite default

	default:
		info.BackendPort = 8080
	}
}

type PackageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Scripts         map[string]string `json:"scripts"`
}

func detectFromPackageJSON(path string, info *ProjectInfo) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return
	}

	// Detect framework
	if hasPackage(pkg, "@sveltejs/kit") || hasPackage(pkg, "svelte") {
		info.Framework = "sveltekit"
		info.Detected = true
	} else if hasPackage(pkg, "next") {
		info.Framework = "nextjs"
		info.Detected = true
	} else if hasPackage(pkg, "nuxt") {
		info.Framework = "nuxt"
		info.Detected = true
	} else if hasPackage(pkg, "react") && !hasPackage(pkg, "next") {
		info.Framework = "react"
		info.Detected = true
	} else if hasPackage(pkg, "vue") && !hasPackage(pkg, "nuxt") {
		info.Framework = "vue"
		info.Detected = true
	} else if hasPackage(pkg, "@angular/core") {
		info.Framework = "angular"
		info.Detected = true
	}

	// Detect BaaS
	if hasPackage(pkg, "pocketbase") {
		info.BaaSType = "pocketbase"
	} else if hasPackage(pkg, "@supabase/supabase-js") {
		info.BaaSType = "supabase"
	} else if hasPackage(pkg, "firebase") {
		info.BaaSType = "firebase"
	} else if hasPackage(pkg, "appwrite") {
		info.BaaSType = "appwrite"
	}
}

func hasPackage(pkg PackageJSON, name string) bool {
	if _, ok := pkg.Dependencies[name]; ok {
		return true
	}
	if _, ok := pkg.DevDependencies[name]; ok {
		return true
	}
	return false
}

func detectFromFileStructure(repoPath string, info *ProjectInfo) {
	// Check for svelte.config.js (SvelteKit)
	if fileExists(filepath.Join(repoPath, "svelte.config.js")) {
		info.Framework = "sveltekit"
		info.Detected = true
		return
	}

	// Check for next.config.js (Next.js)
	if fileExists(filepath.Join(repoPath, "next.config.js")) ||
	   fileExists(filepath.Join(repoPath, "next.config.mjs")) {
		info.Framework = "nextjs"
		info.Detected = true
		return
	}

	// Check for nuxt.config.js (Nuxt)
	if fileExists(filepath.Join(repoPath, "nuxt.config.js")) ||
	   fileExists(filepath.Join(repoPath, "nuxt.config.ts")) {
		info.Framework = "nuxt"
		info.Detected = true
		return
	}

	// Check for angular.json (Angular)
	if fileExists(filepath.Join(repoPath, "angular.json")) {
		info.Framework = "angular"
		info.Detected = true
		return
	}

	// Check for BaaS in same directory
	detectBaaSFromPath(repoPath, info)
}

func detectBaaSFromPath(repoPath string, info *ProjectInfo) {
	// Check for BaaS config files
	if fileExists(filepath.Join(repoPath, "pocketbase")) {
		info.BaaSType = "pocketbase"
		return
	}

	// Check for firebase.json
	if fileExists(filepath.Join(repoPath, "firebase.json")) {
		info.BaaSType = "firebase"
		return
	}

	// Check for supabase directory
	if dirExists(filepath.Join(repoPath, "supabase")) {
		info.BaaSType = "supabase"
		return
	}

	// Check for package.json with BaaS dependencies
	packageJSONPath := filepath.Join(repoPath, "package.json")
	if fileExists(packageJSONPath) {
		data, err := os.ReadFile(packageJSONPath)
		if err == nil {
			var pkg PackageJSON
			if err := json.Unmarshal(data, &pkg); err == nil {
				if hasPackage(pkg, "pocketbase") {
					info.BaaSType = "pocketbase"
				} else if hasPackage(pkg, "@supabase/supabase-js") {
					info.BaaSType = "supabase"
				} else if hasPackage(pkg, "firebase") {
					info.BaaSType = "firebase"
				} else if hasPackage(pkg, "appwrite") {
					info.BaaSType = "appwrite"
				}
			}
		}
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// DetectFromGitURL clones a repo temporarily and detects framework/BaaS
func DetectFromGitURL(gitURL, branch string) (*ProjectInfo, error) {
	// Create temp directory
	tempDir := filepath.Join(os.TempDir(), "vps-panel-detect-"+randomString(8))
	defer os.RemoveAll(tempDir)

	// Clone the repository (shallow clone for speed)
	// We'll use the git service for this
	// For now, return a placeholder
	info := &ProjectInfo{
		Detected: false,
	}

	return info, nil
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}
