// Package scaffold handles torch repository scaffolding.
package scaffold

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// TorchManifest represents the .hecate/torch.json file.
type TorchManifest struct {
	TorchID     string `json:"torch_id"`
	Name        string `json:"name"`
	Brief       string `json:"brief,omitempty"`
	Root        string `json:"root"`
	InitiatedAt int64  `json:"initiated_at"`
	InitiatedBy string `json:"initiated_by,omitempty"`
}

// TemplateData holds data for rendering templates.
type TemplateData struct {
	Name        string
	Brief       string
	RepoURL     string
	Date        string
	InitiatedBy string
}

// Result holds the result of scaffolding.
type Result struct {
	Success          bool
	HecateDir        string
	AgentsCloned     bool
	ReadmeCreated    bool
	ChangelogCreated bool
	VisionCreated    bool
	GitInitialized   bool
	GitCommitted     bool
	Warnings         []string
	Error            error
}

const (
	agentsRepoURL = "https://github.com/hecate-social/hecate-agents.git"
)

// Scaffold creates the full torch repository structure.
// It creates:
//   - .hecate/torch.json
//   - .hecate/agents/ (cloned from hecate-agents)
//   - README.md (from template)
//   - CHANGELOG.md (from template)
func Scaffold(root string, manifest TorchManifest) Result {
	result := Result{
		HecateDir: filepath.Join(root, ".hecate"),
	}

	// 1. Create .hecate directory
	if err := os.MkdirAll(result.HecateDir, 0755); err != nil {
		result.Error = fmt.Errorf("create .hecate directory: %w", err)
		return result
	}

	// 2. Write torch.json
	if err := writeTorchManifest(result.HecateDir, manifest); err != nil {
		result.Error = fmt.Errorf("write torch.json: %w", err)
		return result
	}

	// 3. Clone hecate-agents
	agentsDir := filepath.Join(result.HecateDir, "agents")
	if err := cloneAgents(agentsDir); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to clone agents: %v", err))
	} else {
		result.AgentsCloned = true
	}

	// 4. Prepare template data
	data := TemplateData{
		Name:        manifest.Name,
		Brief:       manifest.Brief,
		RepoURL:     inferRepoURL(root),
		Date:        time.Now().Format("2006-01-02"),
		InitiatedBy: manifest.InitiatedBy,
	}

	// 5. Generate README.md
	if err := generateFromTemplate(root, agentsDir, "README.md", data); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to create README.md: %v", err))
	} else {
		result.ReadmeCreated = true
	}

	// 6. Generate CHANGELOG.md
	if err := generateFromTemplate(root, agentsDir, "CHANGELOG.md", data); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to create CHANGELOG.md: %v", err))
	} else {
		result.ChangelogCreated = true
	}

	// 7. Generate VISION.md
	if err := generateFromTemplate(root, agentsDir, "VISION.md", data); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to create VISION.md: %v", err))
	} else {
		result.VisionCreated = true
	}

	// 8. Create .gitignore
	if err := createGitignore(root); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to create .gitignore: %v", err))
	}

	// 9. Initialize git repository
	if err := gitInit(root); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to git init: %v", err))
	} else {
		result.GitInitialized = true

		// 10. Git add and commit
		if err := gitCommit(root, manifest.Name); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to git commit: %v", err))
		} else {
			result.GitCommitted = true
		}
	}

	result.Success = true
	return result
}

func createGitignore(root string) error {
	gitignorePath := filepath.Join(root, ".gitignore")

	// Don't overwrite existing
	if _, err := os.Stat(gitignorePath); err == nil {
		return nil
	}

	content := `# OS
.DS_Store
Thumbs.db

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# Build
/target/
/_build/
/deps/
/node_modules/
/dist/
/build/

# Secrets (never commit these!)
.env
.env.local
*.pem
*.key
credentials.json

# Logs
*.log
logs/
`
	return os.WriteFile(gitignorePath, []byte(content), 0644)
}

func gitInit(root string) error {
	// Check if already a git repo
	gitDir := filepath.Join(root, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return nil // Already initialized
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = root
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}

func gitCommit(root, torchName string) error {
	// Stage all files
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = root
	if output, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add: %w: %s", err, string(output))
	}

	// Commit
	msg := fmt.Sprintf("Initialize torch: %s\n\nScaffolded by Hecate TUI", torchName)
	commitCmd := exec.Command("git", "commit", "-m", msg)
	commitCmd.Dir = root
	if output, err := commitCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit: %w: %s", err, string(output))
	}
	return nil
}

func writeTorchManifest(hecateDir string, manifest TorchManifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(hecateDir, "torch.json"), data, 0644)
}

func cloneAgents(agentsDir string) error {
	// Check if already exists
	if _, err := os.Stat(agentsDir); err == nil {
		return nil // Already exists, skip
	}

	// Clone with depth 1 (shallow clone)
	cmd := exec.Command("git", "clone", "--depth", "1", agentsRepoURL, agentsDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}

	// Remove .git directory to make it part of the repo
	gitDir := filepath.Join(agentsDir, ".git")
	if err := os.RemoveAll(gitDir); err != nil {
		return fmt.Errorf("remove .git: %w", err)
	}

	return nil
}

func generateFromTemplate(root, agentsDir, filename string, data TemplateData) error {
	outputPath := filepath.Join(root, filename)

	// Don't overwrite existing files
	if _, err := os.Stat(outputPath); err == nil {
		return nil // Already exists, skip
	}

	// Try to load template from agents
	tmplPath := filepath.Join(agentsDir, "templates", filename+".tmpl")
	tmplContent, err := os.ReadFile(tmplPath)
	if err != nil {
		// Fall back to default template
		tmplContent = []byte(defaultTemplate(filename, data))
		// Write directly without template processing
		return os.WriteFile(outputPath, tmplContent, 0644)
	}

	// Parse and execute template
	tmpl, err := template.New(filename).Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	return os.WriteFile(outputPath, buf.Bytes(), 0644)
}

func defaultTemplate(filename string, data TemplateData) string {
	switch filename {
	case "README.md":
		brief := ""
		if data.Brief != "" {
			brief = "\n" + data.Brief + "\n"
		}
		return fmt.Sprintf(`# %s
%s
## Getting Started

Start the Hecate TUI:

`+"```bash"+`
hecate-tui
`+"```"+`

---

*Managed with [Hecate](https://github.com/hecate-social/hecate)*
`, data.Name, brief)

	case "CHANGELOG.md":
		return fmt.Sprintf(`# Changelog

All notable changes to **%s** will be documented in this file.

## [Unreleased]

### Added
- Initialized torch

---

*Managed with [Hecate](https://github.com/hecate-social/hecate)*
`, data.Name)

	case "VISION.md":
		brief := ""
		if data.Brief != "" {
			brief = data.Brief
		}
		return fmt.Sprintf(`# Vision: %s

> %s

## Problem

What problem are we solving? Why does it matter?

## Vision

What does success look like?

## Scope

### In Scope

-

### Out of Scope

-

## Repositories

| Repository | Role |
|------------|------|
| | |

## Constraints

-

## Success Criteria

- [ ]

---

*Initiated %s*
`, data.Name, brief, data.Date)

	default:
		return ""
	}
}

func inferRepoURL(root string) string {
	// Try to get git remote URL
	cmd := exec.Command("git", "-C", root, "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "https://github.com/your-org/" + filepath.Base(root)
	}
	return strings.TrimSpace(string(output))
}

// ScaffoldVision creates a VISION.md file in the given root directory if it doesn't exist.
// Uses the template from .hecate/agents/templates/ if available, otherwise falls back to default.
func ScaffoldVision(root string, manifest TorchManifest) (bool, error) {
	visionPath := filepath.Join(root, "VISION.md")

	// Already exists — nothing to do
	if _, err := os.Stat(visionPath); err == nil {
		return false, nil
	}

	agentsDir := filepath.Join(root, ".hecate", "agents")
	data := TemplateData{
		Name:        manifest.Name,
		Brief:       manifest.Brief,
		Date:        time.Now().Format("2006-01-02"),
		InitiatedBy: manifest.InitiatedBy,
	}

	if err := generateFromTemplate(root, agentsDir, "VISION.md", data); err != nil {
		return false, err
	}
	return true, nil
}

// VisionPath returns the path to VISION.md in the given torch root.
func VisionPath(root string) string {
	return filepath.Join(root, "VISION.md")
}

// VisionExists checks if VISION.md exists in the given torch root.
func VisionExists(root string) bool {
	_, err := os.Stat(VisionPath(root))
	return err == nil
}

// RetryCloneAgents attempts to clone agents again after a failure.
func RetryCloneAgents(root string) error {
	agentsDir := filepath.Join(root, ".hecate", "agents")

	// Remove partial clone if exists
	os.RemoveAll(agentsDir)

	return cloneAgents(agentsDir)
}
