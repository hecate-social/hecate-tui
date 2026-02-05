package commands

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// SubscriptionsCmd lists active mesh subscriptions.
type SubscriptionsCmd struct{}

func (c *SubscriptionsCmd) Name() string        { return "subscriptions" }
func (c *SubscriptionsCmd) Aliases() []string   { return []string{"subs"} }
func (c *SubscriptionsCmd) Description() string { return "List active mesh subscriptions" }

func (c *SubscriptionsCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		subs, err := ctx.Client.ListSubscriptions()
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to list subscriptions: " + err.Error())}
		}

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("Subscriptions"))
		b.WriteString("\n\n")

		if len(subs) == 0 {
			b.WriteString(s.Subtle.Render("No active subscriptions."))
			b.WriteString("\n")
			b.WriteString(s.Subtle.Render("Use the mesh to subscribe to capabilities."))
			return InjectSystemMsg{Content: b.String()}
		}

		for i, sub := range subs {
			b.WriteString(s.Bold.Render(sub.ServiceMRI))
			b.WriteString("\n")
			b.WriteString(s.CardLabel.Render("  ID:"))
			b.WriteString(" ")
			b.WriteString(s.CardValue.Render(sub.SubscriptionID))
			b.WriteString("\n")
			b.WriteString(s.CardLabel.Render("  Since:"))
			b.WriteString(" ")
			b.WriteString(s.CardValue.Render(sub.SubscribedAt))
			if i < len(subs)-1 {
				b.WriteString("\n\n")
			}
		}

		b.WriteString("\n\n")
		b.WriteString(s.Subtle.Render(itoa(len(subs)) + " active subscription(s)"))

		return InjectSystemMsg{Content: b.String()}
	}
}

// itoa is defined in me.go
