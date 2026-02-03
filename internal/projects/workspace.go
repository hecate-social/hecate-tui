package projects

import (
	"os"
	"path/filepath"
)

// Workspace manages a project's .hecate/ directory
type Workspace struct {
	project *Project
	root    string
}

// WorkspaceFiles defines the standard .hecate files
var WorkspaceFiles = []string{
	"QUEUE.md",
	"RESPONSES.md",
	"STATUS.md",
}

// OpenWorkspace opens or creates a workspace for a project
func OpenWorkspace(project *Project) (*Workspace, error) {
	root := filepath.Join(project.Path, ".hecate")

	w := &Workspace{
		project: project,
		root:    root,
	}

	return w, nil
}

// Exists checks if the workspace exists
func (w *Workspace) Exists() bool {
	info, err := os.Stat(w.root)
	return err == nil && info.IsDir()
}

// Init creates the workspace directory and files
func (w *Workspace) Init() error {
	if err := os.MkdirAll(w.root, 0755); err != nil {
		return err
	}

	// Create QUEUE.md if not exists
	queuePath := filepath.Join(w.root, "QUEUE.md")
	if _, err := os.Stat(queuePath); os.IsNotExist(err) {
		if err := w.writeQueue(); err != nil {
			return err
		}
	}

	// Create RESPONSES.md if not exists
	responsesPath := filepath.Join(w.root, "RESPONSES.md")
	if _, err := os.Stat(responsesPath); os.IsNotExist(err) {
		if err := w.writeResponses(); err != nil {
			return err
		}
	}

	// Create STATUS.md if not exists
	statusPath := filepath.Join(w.root, "STATUS.md")
	if _, err := os.Stat(statusPath); os.IsNotExist(err) {
		if err := w.writeStatus(); err != nil {
			return err
		}
	}

	w.project.HasWorkspace = true
	return nil
}

// Path returns the workspace root path
func (w *Workspace) Path() string {
	return w.root
}

// QueuePath returns path to QUEUE.md
func (w *Workspace) QueuePath() string {
	return filepath.Join(w.root, "QUEUE.md")
}

// ResponsesPath returns path to RESPONSES.md
func (w *Workspace) ResponsesPath() string {
	return filepath.Join(w.root, "RESPONSES.md")
}

// StatusPath returns path to STATUS.md
func (w *Workspace) StatusPath() string {
	return filepath.Join(w.root, "STATUS.md")
}

// ReadQueue reads the queue file content
func (w *Workspace) ReadQueue() (string, error) {
	data, err := os.ReadFile(w.QueuePath())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ReadResponses reads the responses file content
func (w *Workspace) ReadResponses() (string, error) {
	data, err := os.ReadFile(w.ResponsesPath())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ReadStatus reads the status file content
func (w *Workspace) ReadStatus() (string, error) {
	data, err := os.ReadFile(w.StatusPath())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (w *Workspace) writeQueue() error {
	content := `# Hecate's Queue

*Commands from the goddess. Read and obey.*

---

## How To Use

The Queue contains tasks for the apprentice (Claude).
Tasks are listed in priority order. Complete them one at a time.

---

## Tasks

*(Add tasks below)*

---
`
	return os.WriteFile(w.QueuePath(), []byte(content), 0644)
}

func (w *Workspace) writeResponses() error {
	content := `# Apprentice Responses

*Write here when you need Hecate's attention.*

---

## How To Use

When you:
- Complete a task → Report it here
- Have a question → Ask it here
- Hit a blocker → Describe it here
- Need a decision → Request it here

**Format:**
` + "```markdown" + `
## [DATE] [TYPE]: Brief Title

[Your message]

---
` + "```" + `

Types: ` + "`COMPLETE`, `QUESTION`, `BLOCKED`, `DECISION`, `UPDATE`" + `

---

## Messages

*(Write below this line)*

---
`
	return os.WriteFile(w.ResponsesPath(), []byte(content), 0644)
}

func (w *Workspace) writeStatus() error {
	content := `# Apprentice Status

*Current state of the apprentice's work.*

---

## Current Task

**None**

## Last Active

**Never**

---

## Session Log

*(Session updates appear here)*

---
`
	return os.WriteFile(w.StatusPath(), []byte(content), 0644)
}
