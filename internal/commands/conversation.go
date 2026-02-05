package commands

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/config"
)

// NewCmd starts a fresh conversation.
type NewCmd struct{}

func (c *NewCmd) Name() string        { return "new" }
func (c *NewCmd) Aliases() []string   { return []string{"n"} }
func (c *NewCmd) Description() string { return "Start a new conversation" }

func (c *NewCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		return NewConversationMsg{}
	}
}

// HistoryCmd lists saved conversations.
type HistoryCmd struct{}

func (c *HistoryCmd) Name() string        { return "history" }
func (c *HistoryCmd) Aliases() []string   { return []string{"hist"} }
func (c *HistoryCmd) Description() string { return "List saved conversations" }

func (c *HistoryCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		convs := config.ListConversations()

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("Conversations"))
		b.WriteString("\n\n")

		if len(convs) == 0 {
			b.WriteString(s.Subtle.Render("No saved conversations."))
			return InjectSystemMsg{Content: b.String()}
		}

		limit := 10
		if len(convs) < limit {
			limit = len(convs)
		}

		for i := 0; i < limit; i++ {
			conv := convs[i]
			// Index for /load
			idx := s.Bold.Render(itoa(i+1) + ".")
			title := s.CardValue.Render(conv.Title)
			meta := s.Subtle.Render(
				"  " + conv.UpdatedAt.Format("Jan 02 15:04") +
					"  " + itoa(len(conv.Messages)) + " msgs",
			)
			if conv.Model != "" {
				meta += s.Subtle.Render("  " + conv.Model)
			}
			b.WriteString(idx + " " + title + meta)
			b.WriteString("\n")
			b.WriteString(s.Subtle.Render("     ID: " + conv.ID))
			if i < limit-1 {
				b.WriteString("\n")
			}
		}

		if len(convs) > limit {
			b.WriteString("\n")
			b.WriteString(s.Subtle.Render("  ..." + itoa(len(convs)-limit) + " more"))
		}

		b.WriteString("\n\n")
		b.WriteString(s.Subtle.Render("Use /load <id> or /load <number> to load"))

		return InjectSystemMsg{Content: b.String()}
	}
}

// LoadCmd loads a saved conversation.
type LoadCmd struct{}

func (c *LoadCmd) Name() string        { return "load" }
func (c *LoadCmd) Aliases() []string   { return nil }
func (c *LoadCmd) Description() string { return "Load a saved conversation (/load <id|number>)" }

func (c *LoadCmd) Execute(args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: "Usage: /load <id> or /load <number>\nUse /history to see available conversations."}
		}
	}

	target := args[0]

	// Check if it's a numeric index (1-based)
	if n := parseIndex(target); n > 0 {
		convs := config.ListConversations()
		if n > len(convs) {
			return func() tea.Msg {
				return InjectSystemMsg{Content: "Conversation #" + target + " not found. Use /history to see available."}
			}
		}
		target = convs[n-1].ID
	}

	return func() tea.Msg {
		return LoadConversationMsg{ID: target}
	}
}

// DeleteCmd removes a saved conversation.
type DeleteCmd struct{}

func (c *DeleteCmd) Name() string        { return "delete" }
func (c *DeleteCmd) Aliases() []string   { return []string{"del"} }
func (c *DeleteCmd) Description() string { return "Delete a saved conversation (/delete <id|number>)" }

func (c *DeleteCmd) Execute(args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: "Usage: /delete <id> or /delete <number>\nUse /history to see available conversations."}
		}
	}

	target := args[0]

	// Check if it's a numeric index
	if n := parseIndex(target); n > 0 {
		convs := config.ListConversations()
		if n > len(convs) {
			return func() tea.Msg {
				return InjectSystemMsg{Content: "Conversation #" + target + " not found."}
			}
		}
		target = convs[n-1].ID
	}

	return func() tea.Msg {
		if err := config.DeleteConversation(target); err != nil {
			return InjectSystemMsg{Content: "Delete failed: " + err.Error()}
		}
		return InjectSystemMsg{Content: "Deleted conversation: " + target}
	}
}

func parseIndex(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}
