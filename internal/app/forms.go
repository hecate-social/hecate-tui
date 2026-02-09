package app

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/commands"
	"github.com/hecate-social/hecate-tui/internal/editor"
	"github.com/hecate-social/hecate-tui/internal/modes"
	"github.com/hecate-social/hecate-tui/internal/scaffold"
	"github.com/hecate-social/hecate-tui/internal/ui"
)

// showForm initializes and displays a form overlay.
func (a *App) showForm(formType string) tea.Cmd {
	switch formType {
	case "torch_init":
		cwd, _ := os.Getwd()
		a.formView = ui.NewTorchForm(a.theme, a.styles, cwd)
		formWidth := 60
		if a.width > 0 && a.width < 70 {
			formWidth = a.width - 4
		}
		a.formView.SetWidth(formWidth)
		a.formReady = true
		a.setMode(modes.Form)
		return a.formView.Init()
	default:
		a.chat.InjectSystemMessage("Unknown form type: " + formType)
		return nil
	}
}

// handleFormResult processes form submission or cancellation.
func (a *App) handleFormResult(result ui.FormResult) tea.Cmd {
	a.formReady = false
	a.setMode(modes.Normal)

	if !result.Submitted {
		a.chat.InjectSystemMessage("Cancelled.")
		return nil
	}

	// Handle based on form type
	switch result.FormID {
	case "torch_init":
		pathInput := result.Values["path"]
		name := result.Values["name"]
		brief := result.Values["brief"]

		// Path is required
		if strings.TrimSpace(pathInput) == "" {
			a.chat.InjectSystemMessage(a.styles.Error.Render("Path is required"))
			return nil
		}

		// Expand path
		cwd, _ := os.Getwd()
		path := ui.ExpandPath(pathInput, cwd)

		// Infer name from path if not provided
		if strings.TrimSpace(name) == "" {
			name = ui.InferName(path)
		}

		return a.createTorchFromForm(path, name, brief)

	default:
		a.chat.InjectSystemMessage("Unknown form: " + result.FormID)
		return nil
	}
}

// createTorchFromForm creates a torch after form submission.
func (a *App) createTorchFromForm(path, name, brief string) tea.Cmd {
	return func() tea.Msg {
		s := a.styles

		// Create directory if it doesn't exist
		if err := os.MkdirAll(path, 0755); err != nil {
			return commands.InjectSystemMsg{Content: s.Error.Render("Failed to create directory: " + err.Error())}
		}

		// Create torch via daemon
		torch, err := a.client.InitiateTorch(name, brief)
		if err != nil {
			return commands.InjectSystemMsg{Content: s.Error.Render("Failed to initiate torch: " + err.Error())}
		}

		// Scaffold the repository structure
		manifest := scaffold.TorchManifest{
			TorchID:     torch.TorchID,
			Name:        torch.Name,
			Brief:       torch.Brief,
			Root:        path,
			InitiatedAt: torch.InitiatedAt,
			InitiatedBy: torch.InitiatedBy,
		}

		result := scaffold.Scaffold(path, manifest)

		// Build output message
		var b strings.Builder
		b.WriteString(s.StatusOK.Render("Torch Initiated"))
		b.WriteString("\n\n")
		b.WriteString(s.CardTitle.Render("Torch: " + torch.Name))
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("    ID: "))
		b.WriteString(s.CardValue.Render(torch.TorchID))
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("  Root: "))
		b.WriteString(s.Subtle.Render(path))
		b.WriteString("\n\n")

		b.WriteString(s.CardTitle.Render("Scaffolded:"))
		b.WriteString("\n")

		if result.Success {
			b.WriteString(s.StatusOK.Render("  ✓ "))
			b.WriteString(s.Subtle.Render(".hecate/torch.json"))
			b.WriteString("\n")
		}
		if result.AgentsCloned {
			b.WriteString(s.StatusOK.Render("  ✓ "))
			b.WriteString(s.Subtle.Render(".hecate/agents/"))
			b.WriteString("\n")
		}
		if result.ReadmeCreated {
			b.WriteString(s.StatusOK.Render("  ✓ "))
			b.WriteString(s.Subtle.Render("README.md"))
			b.WriteString("\n")
		}
		if result.ChangelogCreated {
			b.WriteString(s.StatusOK.Render("  ✓ "))
			b.WriteString(s.Subtle.Render("CHANGELOG.md"))
			b.WriteString("\n")
		}

		if result.GitInitialized {
			b.WriteString(s.StatusOK.Render("  ✓ "))
			b.WriteString(s.Subtle.Render("git init"))
			b.WriteString("\n")
		}

		if result.GitCommitted {
			b.WriteString(s.StatusOK.Render("  ✓ "))
			b.WriteString(s.Subtle.Render("git commit"))
			b.WriteString("\n")
		}

		for _, warn := range result.Warnings {
			b.WriteString(s.StatusWarning.Render("  ⚠ " + warn))
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("Next: gh repo create --public --source=. --push"))

		return commands.TorchCreatedMsg{Path: path, Message: b.String()}
	}
}

func (a *App) openEditor(path string) tea.Cmd {
	if path != "" {
		ed, err := editor.NewWithFile(path)
		if err != nil {
			a.chat.InjectSystemMessage("Could not open file: " + err.Error())
			return nil
		}
		a.editorView = ed
	} else {
		a.editorView = editor.New()
	}

	a.editorView.SetSize(a.width, a.editorHeight())
	a.editorView.Focus()
	a.editorReady = true
	a.setMode(modes.Edit)
	return a.editorView.Init()
}

func (a *App) editorHeight() int {
	return a.height - 2 // header + status bar
}
