package projects

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Detector finds projects in directories
type Detector struct {
	maxDepth int
}

// NewDetector creates a project detector
func NewDetector() *Detector {
	return &Detector{
		maxDepth: 3, // Don't recurse too deep
	}
}

// DetectCurrent checks if current directory is a project
func (d *Detector) DetectCurrent() (*Project, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return d.DetectAt(cwd)
}

// DetectAt checks if path is a project
func (d *Detector) DetectAt(path string) (*Project, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, nil
	}

	hasGit := d.hasDir(path, ".git")
	hasHecate := d.hasFile(path, "HECATE.md")
	hasWorkspace := d.hasDir(path, ".hecate")

	if !hasGit && !hasHecate {
		return nil, nil
	}

	project := &Project{
		Name:         filepath.Base(path),
		Path:         path,
		HasWorkspace: hasWorkspace,
		DetectedAt:   time.Now(),
		CurrentPhase: PhaseAnD, // Default
	}

	// Determine type
	if hasGit && hasHecate {
		project.Type = ProjectTypeBoth
	} else if hasGit {
		project.Type = ProjectTypeGit
	} else {
		project.Type = ProjectTypeHecate
	}

	// Get git info
	if hasGit {
		project.GitBranch = d.readGitBranch(path)
		project.GitRemote = d.readGitRemote(path)
	}

	// Get hecate info
	if hasHecate {
		title, phase := d.readHecateMd(path)
		project.HecateTitle = title
		project.HecatePhase = phase
		if phase != "" {
			project.CurrentPhase = Phase(strings.ToLower(phase))
		}
	}

	return project, nil
}

// Scan finds all projects under a path
func (d *Detector) Scan(root string) ([]*Project, error) {
	var projects []*Project

	err := d.walkProjects(root, 0, func(p *Project) {
		projects = append(projects, p)
	})

	return projects, err
}

func (d *Detector) walkProjects(path string, depth int, fn func(*Project)) error {
	if depth > d.maxDepth {
		return nil
	}

	// Check if this directory is a project
	project, err := d.DetectAt(path)
	if err != nil {
		return err
	}

	if project != nil {
		fn(project)
		// Don't recurse into projects
		return nil
	}

	// Scan subdirectories
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil // Ignore permission errors
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Skip hidden dirs (except we want to find projects)
		if strings.HasPrefix(name, ".") {
			continue
		}
		// Skip common non-project dirs
		if name == "node_modules" || name == "vendor" || name == "__pycache__" {
			continue
		}

		subpath := filepath.Join(path, name)
		if err := d.walkProjects(subpath, depth+1, fn); err != nil {
			continue // Skip errors in subdirs
		}
	}

	return nil
}

func (d *Detector) hasDir(path, name string) bool {
	info, err := os.Stat(filepath.Join(path, name))
	return err == nil && info.IsDir()
}

func (d *Detector) hasFile(path, name string) bool {
	info, err := os.Stat(filepath.Join(path, name))
	return err == nil && !info.IsDir()
}

func (d *Detector) readGitBranch(path string) string {
	headPath := filepath.Join(path, ".git", "HEAD")
	data, err := os.ReadFile(headPath)
	if err != nil {
		return ""
	}
	content := strings.TrimSpace(string(data))
	// Format: ref: refs/heads/branch-name
	if strings.HasPrefix(content, "ref: refs/heads/") {
		return strings.TrimPrefix(content, "ref: refs/heads/")
	}
	// Detached HEAD - return short hash
	if len(content) >= 7 {
		return content[:7]
	}
	return content
}

func (d *Detector) readGitRemote(path string) string {
	configPath := filepath.Join(path, ".git", "config")
	file, err := os.Open(configPath)
	if err != nil {
		return ""
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	inRemoteOrigin := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "[remote \"origin\"]" {
			inRemoteOrigin = true
			continue
		}
		if inRemoteOrigin {
			if strings.HasPrefix(line, "[") {
				break
			}
			if strings.HasPrefix(line, "url = ") {
				return strings.TrimPrefix(line, "url = ")
			}
		}
	}
	return ""
}

func (d *Detector) readHecateMd(path string) (title, phase string) {
	filePath := filepath.Join(path, "HECATE.md")
	file, err := os.Open(filePath)
	if err != nil {
		return "", ""
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Look for title (# Title)
		if strings.HasPrefix(line, "# ") && title == "" {
			title = strings.TrimPrefix(line, "# ")
		}
		// Look for phase marker
		lower := strings.ToLower(line)
		if strings.Contains(lower, "phase:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				phase = strings.TrimSpace(parts[1])
			}
		}
	}
	return title, phase
}
