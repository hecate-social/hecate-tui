// Package ui provides shared UI components for the TUI.
package ui

import (
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

// FormResult is sent when a form completes (submit or cancel).
type FormResult struct {
	Submitted bool
	Values    map[string]string
	FormID    string // Identifies which form this result is from
}

// FormModel wraps a huh form for use in the TUI.
type FormModel struct {
	form   *huh.Form
	theme  *theme.Theme
	styles *theme.Styles
	width  int
	formID string
	title  string
}

// NewTorchForm creates a form for initiating a new torch.
func NewTorchForm(t *theme.Theme, s *theme.Styles, cwd string) *FormModel {
	var path, name, brief string
	var confirm bool

	// Use charm theme
	huhTheme := huh.ThemeCharm()

	// Default path placeholder uses shortened cwd
	cwdDisplay := shortenHome(cwd)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("path").
				Title("Path").
				Description("Directory to create (relative or absolute)").
				Placeholder(cwdDisplay + "/my-venture").
				Value(&path),
			huh.NewInput().
				Key("name").
				Title("Name").
				Description("Leave empty to use directory name").
				Placeholder("(auto from path)").
				Value(&name),
			huh.NewInput().
				Key("brief").
				Title("Brief").
				Description("Optional description").
				Placeholder("A revolutionary new product...").
				Value(&brief),
			huh.NewConfirm().
				Key("confirm").
				Title("").
				Affirmative("Create").
				Negative("Cancel").
				Value(&confirm),
		),
	).WithTheme(huhTheme).
		WithWidth(55).
		WithShowHelp(false)

	return &FormModel{
		form:   form,
		theme:  t,
		styles: s,
		formID: "torch_init",
		title:  "New Torch",
		width:  55,
	}
}

// shortenHome replaces home directory with ~
func shortenHome(path string) string {
	if home := os.Getenv("HOME"); home != "" && len(path) > 0 {
		if strings.HasPrefix(path, home) {
			return "~" + path[len(home):]
		}
	}
	return path
}

// ExpandPath expands ~ and makes path absolute relative to cwd.
func ExpandPath(path, cwd string) string {
	if path == "" {
		return cwd
	}

	// Expand ~
	if strings.HasPrefix(path, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(home, path[1:])
		}
	}

	// Make absolute if relative
	if !filepath.IsAbs(path) {
		path = filepath.Join(cwd, path)
	}

	return filepath.Clean(path)
}

// InferName extracts the project name from a path.
func InferName(path string) string {
	return filepath.Base(path)
}

// Init initializes the form.
func (m *FormModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update handles form input.
func (m *FormModel) Update(msg tea.Msg) (*FormModel, tea.Cmd) {
	// Check for escape key to cancel
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "esc" {
			return m, func() tea.Msg {
				return FormResult{
					Submitted: false,
					FormID:    m.formID,
				}
			}
		}
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	// Check if form is complete
	if m.form.State == huh.StateCompleted {
		confirmed := m.form.GetBool("confirm")
		if !confirmed {
			return m, func() tea.Msg {
				return FormResult{
					Submitted: false,
					FormID:    m.formID,
				}
			}
		}
		values := make(map[string]string)
		values["path"] = m.form.GetString("path")
		values["name"] = m.form.GetString("name")
		values["brief"] = m.form.GetString("brief")
		return m, func() tea.Msg {
			return FormResult{
				Submitted: true,
				Values:    values,
				FormID:    m.formID,
			}
		}
	}

	return m, cmd
}

// View renders the form.
func (m *FormModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(m.theme.Primary).
		Bold(true)

	title := titleStyle.Render("🔥 " + m.title)

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		m.form.View(),
	)
}

// SetWidth sets the form width.
func (m *FormModel) SetWidth(w int) {
	m.width = w
	if m.width < 40 {
		m.width = 40
	}
	if m.width > 60 {
		m.width = 60
	}
}

// FormID returns the form identifier.
func (m *FormModel) FormID() string {
	return m.formID
}
