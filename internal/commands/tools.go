package commands

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/tools"
)

// ToolsCmd detects and lists local developer tools.
type ToolsCmd struct{}

func (c *ToolsCmd) Name() string        { return "tools" }
func (c *ToolsCmd) Aliases() []string   { return []string{"t"} }
func (c *ToolsCmd) Description() string { return "Detect installed developer tools" }

func (c *ToolsCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		detector := tools.NewDetector()
		detected := detector.Detect()

		s := ctx.Styles
		var b strings.Builder

		b.WriteString(s.CardTitle.Render("Developer Tools"))
		b.WriteString("\n\n")

		// Group by category
		categories := []tools.ToolCategory{
			tools.CategoryEditor,
			tools.CategoryTerminal,
			tools.CategoryVCS,
			tools.CategoryBuild,
			tools.CategoryContainer,
			tools.CategoryLLM,
		}

		for _, cat := range categories {
			var catTools []tools.Tool
			for _, t := range detected {
				if t.Category == cat {
					catTools = append(catTools, t)
				}
			}

			if len(catTools) == 0 {
				continue
			}

			icon := tools.CategoryIcon(cat)
			name := tools.CategoryName(cat)
			b.WriteString(s.Bold.Render(icon + " " + name))
			b.WriteString("\n")

			for _, t := range catTools {
				if t.Installed {
					marker := s.StatusOK.Render("  *")
					version := ""
					if t.Version != "" {
						version = s.Subtle.Render("  " + t.Version)
					}
					b.WriteString(marker + " " + t.Name + version + "\n")
				} else {
					marker := s.Subtle.Render("  -")
					b.WriteString(marker + " " + s.Subtle.Render(t.Name) + "\n")
				}
			}
			b.WriteString("\n")
		}

		// Summary
		installed := 0
		for _, t := range detected {
			if t.Installed {
				installed++
			}
		}
		b.WriteString(s.Subtle.Render(strings.Repeat("─", 30)))
		b.WriteString("\n")
		b.WriteString("  " + s.StatusOK.Render("*") + " installed  " + s.Subtle.Render("-") + " not found\n")
		b.WriteString("  " + s.Bold.Render(fmt.Sprintf("%d of %d tools detected", installed, len(detected))))

		return InjectSystemMsg{Content: b.String()}
	}
}
