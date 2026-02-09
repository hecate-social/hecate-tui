package alc

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// TorchConfig represents the .hecate/torch.json file format.
type TorchConfig struct {
	TorchID string `json:"torch_id"`
	Name    string `json:"name"`
	Brief   string `json:"brief,omitempty"`
}

// DetectResult holds the result of torch detection.
type DetectResult struct {
	Found  bool
	Source string // "git" or "config"
	Config *TorchConfig
}

// DetectTorch attempts to detect a torch from the current directory.
// It checks (in order):
// 1. Git remote URL - matches against known torches (via daemon API)
// 2. .hecate/torch.json in CWD or parent directories
//
// Returns the detection result. Caller should use the daemon API to resolve
// the torch ID to full torch info.
func DetectTorch() DetectResult {
	// First, try to find .hecate/torch.json
	if config := findTorchConfig(); config != nil {
		return DetectResult{
			Found:  true,
			Source: "config",
			Config: config,
		}
	}

	// Next, try git remote URL
	if remoteURL := getGitRemoteURL(); remoteURL != "" {
		// Return the URL - caller will match against daemon's torch list
		return DetectResult{
			Found:  true,
			Source: "git",
			Config: &TorchConfig{
				// Store remote URL in Name for matching
				Name: remoteURL,
			},
		}
	}

	return DetectResult{Found: false}
}

// findTorchConfig searches for .hecate/torch.json in CWD and parent directories.
func findTorchConfig() *TorchConfig {
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}

	dir := cwd
	for {
		configPath := filepath.Join(dir, ".hecate", "torch.json")
		if data, err := os.ReadFile(configPath); err == nil {
			var config TorchConfig
			if json.Unmarshal(data, &config) == nil && config.TorchID != "" {
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

// SaveTorchConfig saves torch configuration to .hecate/torch.json in CWD.
func SaveTorchConfig(config TorchConfig) error {
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

	return os.WriteFile(filepath.Join(hecateDir, "torch.json"), data, 0644)
}
