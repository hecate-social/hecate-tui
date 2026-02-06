package commands

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// MeCmd shows identity information as an inline card.
type MeCmd struct{}

func (c *MeCmd) Name() string        { return "me" }
func (c *MeCmd) Aliases() []string   { return []string{"whoami"} }
func (c *MeCmd) Description() string { return "Show your identity" }

func (c *MeCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		identity, err := ctx.Client.GetIdentity()
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to get identity: " + err.Error())}
		}

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("Identity"))
		b.WriteString("\n\n")

		b.WriteString(s.CardLabel.Render("MRI: "))
		b.WriteString(s.CardValue.Render(identity.Identity))
		b.WriteString("\n")

		b.WriteString(s.CardLabel.Render("Public Key: "))
		pk := identity.PublicKey
		if len(pk) > 20 {
			pk = pk[:8] + "..." + pk[len(pk)-8:]
		}
		b.WriteString(s.CardValue.Render(pk))
		b.WriteString("\n")

		b.WriteString(s.CardLabel.Render("Created: "))
		b.WriteString(s.CardValue.Render(identity.CreatedAt))
		b.WriteString("\n")

		// Capability count
		caps, capsErr := ctx.Client.DiscoverCapabilities("", "", 0)
		if capsErr == nil {
			b.WriteString(s.CardLabel.Render("Capabilities: "))
			b.WriteString(s.CardValue.Render(strings.TrimSpace(itoa(len(caps)))))
			b.WriteString("\n")
		}

		return InjectSystemMsg{Content: b.String()}
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
