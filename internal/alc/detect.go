package alc

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// VentureConfig represents the .hecate/venture.json file format.
type VentureConfig struct {
	VentureID string `json:"venture_id"`
	Name      string `json:"name"`
	Brief     string `json:"brief,omitempty"`
}

// DetectResult holds the result of venture detection.
type DetectResult struct {
	Found  bool
	Source string // "git" or "config"
	Config *VentureConfig
}

// DetectVenture attempts to detect a venture from the current directory.
// It checks (in order):
// 1. Git remote URL - matches against known ventures (via daemon API)
// 2. .hecate/venture.json in CWD or parent directories
//
// Returns the detection result. Caller should use the daemon API to resolve
// the venture ID to full venture info.
func DetectVenture() DetectResult {
	// First, try to find .hecate/venture.json
	if config := findVentureConfig(); config != nil {
		return DetectResult{
			Found:  true,
			Source: "config",
			Config: config,
		}
	}

	// Next, try git remote URL
	if remoteURL := getGitRemoteURL(); remoteURL != "" {
		// Return the URL - caller will match against daemon's venture list
		return DetectResult{
			Found:  true,
			Source: "git",
			Config: &VentureConfig{
				// Store remote URL in Name for matching
				Name: remoteURL,
			},
		}
	}

	return DetectResult{Found: false}
}

// findVentureConfig searches for .hecate/venture.json in CWD and parent directories.
func findVentureConfig() *VentureConfig {
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}

	dir := cwd
	for {
		configPath := filepath.Join(dir, ".hecate", "venture.json")
		if data, err := os.ReadFile(configPath); err == nil {
			var config VentureConfig
			if json.Unmarshal(data, &config) == nil && config.VentureID != "" {
				return &config
			}
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	return nil
}

// getGitRemoteURL returns the git remote origin URL if in a git repository.
func getGitRemoteURL() string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	url := strings.TrimSpace(string(output))
	return normalizeGitURL(url)
}

// normalizeGitURL converts various git URL formats to a canonical form.
// Examples:
//   - git@github.com:org/repo.git -> github.com/org/repo
//   - https://github.com/org/repo.git -> github.com/org/repo
//   - https://github.com/org/repo -> github.com/org/repo
func normalizeGitURL(url string) string {
	if url == "" {
		return ""
	}

	// Handle SSH format: git@github.com:org/repo.git
	if strings.HasPrefix(url, "git@") {
		url = strings.TrimPrefix(url, "git@")
		url = strings.Replace(url, ":", "/", 1)
	}

	// Handle HTTPS format: https://github.com/org/repo.git
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")

	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")

	return url
}

// SaveVentureConfig saves venture configuration to .hecate/venture.json in CWD.
func SaveVentureConfig(config VentureConfig) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	hecateDir := filepath.Join(cwd, ".hecate")
	if err := os.MkdirAll(hecateDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(hecateDir, "venture.json"), data, 0644)
}
