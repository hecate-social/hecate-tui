package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// SaveCmd exports the chat transcript to a file.
type SaveCmd struct{}

func (c *SaveCmd) Name() string        { return "save" }
func (c *SaveCmd) Aliases() []string   { return []string{"w"} }
func (c *SaveCmd) Description() string { return "Save chat transcript (/save [filename])" }

func (c *SaveCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		messages := ctx.GetMessages()
		if len(messages) == 0 {
			return InjectSystemMsg{Content: s.Subtle.Render("No messages to save.")}
		}

		// Determine filename
		filename := ""
		if len(args) > 0 {
			filename = strings.Join(args, " ")
		} else {
			filename = fmt.Sprintf("hecate-chat-%s.md", time.Now().Format("2006-01-02-150405"))
		}

		// Build markdown content
		var b strings.Builder
		b.WriteString("# Hecate Chat Transcript\n")
		b.WriteString(fmt.Sprintf("*Exported: %s*\n\n", time.Now().Format("2006-01-02 15:04:05")))
		b.WriteString("---\n\n")

		for _, msg := range messages {
			timestamp := ""
			if msg.Time != "" {
				timestamp = " (" + msg.Time + ")"
			}

			switch msg.Role {
			case "user":
				b.WriteString("### You" + timestamp + "\n\n")
				b.WriteString(msg.Content + "\n\n")
			case "assistant":
				b.WriteString("### Hecate" + timestamp + "\n\n")
				b.WriteString(msg.Content + "\n\n")
			case "system":
				b.WriteString("---\n\n")
				b.WriteString("*System: " + firstLine(msg.Content) + "*\n\n")
			}
		}

		b.WriteString("---\n*End of transcript*\n")

		err := os.WriteFile(filename, []byte(b.String()), 0644)
		if err != nil {
			return InjectSystemMsg{
				Content: s.Error.Render("Failed to save: " + err.Error()),
			}
		}

		return InjectSystemMsg{
			Content: s.StatusOK.Render("Saved") + " " +
				s.CardValue.Render(filename) + " " +
				s.Subtle.Render(fmt.Sprintf("(%d messages)", len(messages))),
		}
	}
}

func firstLine(s string) string {
	idx := strings.IndexByte(s, '\n')
	if idx == -1 {
		return s
	}
	line := s[:idx]
	if len(line) > 60 {
		return line[:57] + "..."
	}
	return line
}
