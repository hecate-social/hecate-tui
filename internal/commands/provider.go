package commands

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ProviderCmd manages LLM provider configuration.
type ProviderCmd struct{}

func (c *ProviderCmd) Name() string        { return "provider" }
func (c *ProviderCmd) Aliases() []string   { return []string{"providers"} }
func (c *ProviderCmd) Description() string { return "Manage LLM providers (/provider add <type> <key>)" }

func (c *ProviderCmd) Execute(args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return c.listProviders(ctx)
	}

	switch strings.ToLower(args[0]) {
	case "add":
		return c.addProvider(args[1:], ctx)
	case "remove", "rm":
		return c.removeProvider(args[1:], ctx)
	default:
		return c.listProviders(ctx)
	}
}

func (c *ProviderCmd) listProviders(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		providers, err := ctx.Client.ListProviders()
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to list providers: " + err.Error())}
		}

		if len(providers) == 0 {
			return InjectSystemMsg{Content: s.Subtle.Render("No providers configured.")}
		}

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("LLM Providers"))
		b.WriteString("\n\n")

		for name, p := range providers {
			status := s.StatusOK.Render("enabled")
			if !p.Enabled {
				status = s.Error.Render("disabled")
			}
			b.WriteString("  ")
			b.WriteString(s.Bold.Render(name))
			b.WriteString(s.Subtle.Render(" (" + p.Type + ") "))
			b.WriteString(status)
			if p.URL != "" {
				b.WriteString(s.Subtle.Render(" " + p.URL))
			}
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  /provider add <type> <api-key>  Add provider"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  /provider remove <name>         Remove provider"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  Types: anthropic, openai, google, groq, together"))

		return InjectSystemMsg{Content: b.String()}
	}
}

// providerDefaults maps shorthand types to their canonical type and default URL.
type providerDefaults struct {
	name     string
	apiType  string
	url      string
}

var knownProviders = map[string]providerDefaults{
	"anthropic": {name: "anthropic", apiType: "anthropic", url: "https://api.anthropic.com"},
	"openai":    {name: "openai", apiType: "openai", url: "https://api.openai.com"},
	"google":    {name: "google", apiType: "google", url: "https://generativelanguage.googleapis.com"},
	"groq":      {name: "groq", apiType: "openai", url: "https://api.groq.com/openai"},
	"together":  {name: "together", apiType: "openai", url: "https://api.together.xyz"},
}

func (c *ProviderCmd) addProvider(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		if len(args) < 2 {
			return InjectSystemMsg{Content: s.Subtle.Render("Usage: /provider add <type> <api-key>\nTypes: anthropic, openai, google, groq, together")}
		}

		typeName := strings.ToLower(args[0])
		apiKey := args[1]

		defaults, known := knownProviders[typeName]
		if !known {
			return InjectSystemMsg{Content: s.Error.Render("Unknown provider type: " + typeName + "\nKnown types: anthropic, openai, google, groq, together")}
		}

		err := ctx.Client.AddProvider(defaults.name, defaults.apiType, apiKey, defaults.url)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to add provider: " + err.Error())}
		}

		return InjectSystemMsg{Content: s.StatusOK.Render("Added " + defaults.name + " provider (" + defaults.apiType + ")")}
	}
}

func (c *ProviderCmd) removeProvider(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		if len(args) == 0 {
			return InjectSystemMsg{Content: s.Subtle.Render("Usage: /provider remove <name>")}
		}

		name := args[0]
		err := ctx.Client.RemoveProvider(name)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to remove provider: " + err.Error())}
		}

		return InjectSystemMsg{Content: s.StatusOK.Render("Removed provider: " + name)}
	}
}
