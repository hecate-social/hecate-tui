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
	case "help":
		return c.showHelp(ctx)
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
		b.WriteString(s.Subtle.Render("  /provider add <type> <key>  Add provider"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  /provider remove <name>     Remove provider"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  /provider help              How to obtain API keys"))
		b.WriteString("\n\n")
		b.WriteString(s.Subtle.Render("  Types: anthropic, openai, google, mistral, groq, together"))
		b.WriteString("\n")
		b.WriteString(s.Error.Render("  ⚠ Commercial providers charge per token - you pay!"))

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
	"mistral":   {name: "mistral", apiType: "openai", url: "https://api.mistral.ai/v1"},
	"groq":      {name: "groq", apiType: "openai", url: "https://api.groq.com/openai"},
	"together":  {name: "together", apiType: "openai", url: "https://api.together.xyz"},
}

func (c *ProviderCmd) addProvider(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		if len(args) < 2 {
			return InjectSystemMsg{Content: s.Subtle.Render("Usage: /provider add <type> <api-key>\nTypes: anthropic, openai, google, mistral, groq, together")}
		}

		typeName := strings.ToLower(args[0])
		apiKey := args[1]

		defaults, known := knownProviders[typeName]
		if !known {
			return InjectSystemMsg{Content: s.Error.Render("Unknown provider type: " + typeName + "\nKnown types: anthropic, openai, google, mistral, groq, together")}
		}

		err := ctx.Client.AddProvider(defaults.name, defaults.apiType, apiKey, defaults.url)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to add provider: " + err.Error())}
		}

		msg := s.StatusOK.Render("Added " + defaults.name + " provider (" + defaults.apiType + ")")
		msg += "\n" + s.Error.Render("⚠ You are responsible for usage costs. Set spending limits at provider dashboard!")
		return InjectSystemMsg{Content: msg}
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

func (c *ProviderCmd) showHelp(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("LLM Provider Setup"))
		b.WriteString("\n\n")

		// Cost warning
		b.WriteString(s.Error.Render("⚠ COST WARNING"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  Commercial providers charge per token. You are responsible for all costs."))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  Monitor usage at each provider's dashboard. Set spending limits!"))
		b.WriteString("\n\n")

		// Anthropic
		b.WriteString(s.Bold.Render("Anthropic"))
		b.WriteString(s.Subtle.Render(" (Claude models)"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  1. Go to: https://console.anthropic.com/settings/keys"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  2. Create an API key"))
		b.WriteString("\n")
		b.WriteString("  ")
		b.WriteString(s.StatusOK.Render("/provider add anthropic sk-ant-..."))
		b.WriteString("\n\n")

		// OpenAI
		b.WriteString(s.Bold.Render("OpenAI"))
		b.WriteString(s.Subtle.Render(" (GPT models)"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  1. Go to: https://platform.openai.com/api-keys"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  2. Create a new secret key"))
		b.WriteString("\n")
		b.WriteString("  ")
		b.WriteString(s.StatusOK.Render("/provider add openai sk-..."))
		b.WriteString("\n\n")

		// Google
		b.WriteString(s.Bold.Render("Google"))
		b.WriteString(s.Subtle.Render(" (Gemini models)"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  1. Go to: https://aistudio.google.com/apikey"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  2. Create an API key"))
		b.WriteString("\n")
		b.WriteString("  ")
		b.WriteString(s.StatusOK.Render("/provider add google AIza..."))
		b.WriteString("\n\n")

		// Mistral
		b.WriteString(s.Bold.Render("Mistral"))
		b.WriteString(s.Subtle.Render(" (European, Mixtral/Mistral models)"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  1. Go to: https://console.mistral.ai/api-keys"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  2. Create an API key"))
		b.WriteString("\n")
		b.WriteString("  ")
		b.WriteString(s.StatusOK.Render("/provider add mistral ..."))
		b.WriteString("\n\n")

		// Groq
		b.WriteString(s.Bold.Render("Groq"))
		b.WriteString(s.Subtle.Render(" (fast inference, free tier available)"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  1. Go to: https://console.groq.com/keys"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  2. Create an API key"))
		b.WriteString("\n")
		b.WriteString("  ")
		b.WriteString(s.StatusOK.Render("/provider add groq gsk_..."))
		b.WriteString("\n\n")

		// Together
		b.WriteString(s.Bold.Render("Together"))
		b.WriteString(s.Subtle.Render(" (open models)"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  1. Go to: https://api.together.xyz/settings/api-keys"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  2. Create an API key"))
		b.WriteString("\n")
		b.WriteString("  ")
		b.WriteString(s.StatusOK.Render("/provider add together ..."))
		b.WriteString("\n\n")

		// Local alternative
		b.WriteString(s.Bold.Render("Local (Ollama)"))
		b.WriteString(s.Subtle.Render(" - FREE, no API key needed"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  Ollama runs locally. Install from: https://ollama.com"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  Models detected automatically when Ollama is running."))
		b.WriteString("\n\n")

		// Security note
		b.WriteString(s.Subtle.Render("Keys are stored locally in ~/.hecate/data/providers.json"))

		return InjectSystemMsg{Content: b.String()}
	}
}
