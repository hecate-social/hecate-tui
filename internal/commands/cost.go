package commands

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// CostCmd shows LLM cost breakdown.
type CostCmd struct{}

func (c *CostCmd) Name() string        { return "cost" }
func (c *CostCmd) Aliases() []string   { return []string{"$", "costs"} }
func (c *CostCmd) Description() string { return "View LLM cost breakdown" }

func (c *CostCmd) Execute(args []string, ctx *Context) tea.Cmd {
	if len(args) > 0 {
		return c.showVentureCost(args[0], ctx)
	}
	return c.showTotalCost(ctx)
}

func (c *CostCmd) showTotalCost(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		cost, err := ctx.Client.GetTotalCost()
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to get cost: " + err.Error())}
		}

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("LLM Cost Summary"))
		b.WriteString("\n\n")

		b.WriteString(s.CardLabel.Render("Total Cost: "))
		b.WriteString(s.Bold.Render(formatCost(cost.TotalCost)))
		b.WriteString("\n")

		b.WriteString(s.CardLabel.Render("Tokens In:  "))
		b.WriteString(s.CardValue.Render(formatTokens(cost.TotalTokensIn)))
		b.WriteString("\n")

		b.WriteString(s.CardLabel.Render("Tokens Out: "))
		b.WriteString(s.CardValue.Render(formatTokens(cost.TotalTokensOut)))
		b.WriteString("\n")

		b.WriteString(s.CardLabel.Render("API Calls:  "))
		b.WriteString(s.CardValue.Render(fmt.Sprintf("%d", cost.CallCount)))
		b.WriteString("\n")

		if cost.TotalCost > 0 {
			b.WriteString("\n")
			b.WriteString(s.Subtle.Render("Use /cost <venture_id> for per-venture breakdown"))
		}

		return InjectSystemMsg{Content: b.String()}
	}
}

func (c *CostCmd) showVentureCost(ventureID string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		cost, err := ctx.Client.GetCostByVenture(ventureID)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to get cost: " + err.Error())}
		}

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("Venture Cost: " + ventureID))
		b.WriteString("\n\n")

		b.WriteString(s.CardLabel.Render("Total Cost: "))
		b.WriteString(s.Bold.Render(formatCost(cost.TotalCost)))
		b.WriteString("\n")

		b.WriteString(s.CardLabel.Render("Tokens In:  "))
		b.WriteString(s.CardValue.Render(formatTokens(cost.TotalTokensIn)))
		b.WriteString("\n")

		b.WriteString(s.CardLabel.Render("Tokens Out: "))
		b.WriteString(s.CardValue.Render(formatTokens(cost.TotalTokensOut)))
		b.WriteString("\n")

		b.WriteString(s.CardLabel.Render("API Calls:  "))
		b.WriteString(s.CardValue.Render(fmt.Sprintf("%d", cost.CallCount)))
		b.WriteString("\n")

		return InjectSystemMsg{Content: b.String()}
	}
}

// formatCost formats a cost value in USD.
func formatCost(cost float64) string {
	if cost < 0.01 {
		return fmt.Sprintf("$%.4f", cost)
	}
	return fmt.Sprintf("$%.2f", cost)
}

// formatTokens formats token counts with thousands separators.
func formatTokens(tokens int64) string {
	if tokens < 1000 {
		return fmt.Sprintf("%d", tokens)
	}
	if tokens < 1000000 {
		return fmt.Sprintf("%.1fK", float64(tokens)/1000)
	}
	return fmt.Sprintf("%.1fM", float64(tokens)/1000000)
}
