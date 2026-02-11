package chat

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

// Sparkles for streaming animation
var Sparkles = []string{"✦", "✧", "★", "☆", "✦"}

// ThinkingFrames for thinking animation
var ThinkingFrames = []string{
	"Channeling",
	"Channeling.",
	"Channeling..",
	"Channeling...",
}

// WelcomeArt returns the Hecate Threshold Guardian avatar with themed colors.
func WelcomeArt(t *theme.Theme) string {
	hood := lipgloss.NewStyle().Foreground(t.Primary)
	body := lipgloss.NewStyle().Foreground(t.PrimaryLight)
	eye := lipgloss.NewStyle().Foreground(t.EyeColor)
	key := lipgloss.NewStyle().Foreground(t.KeyColor)
	torch := lipgloss.NewStyle().Foreground(t.VentureColor)
	text := lipgloss.NewStyle().Foreground(t.Primary).Bold(true)

	h := hood.Render
	b := body.Render
	e := eye.Render
	k := key.Render
	tt := torch.Render
	tx := text.Render

	return h("╭─╮") + "           " + h("╭─╮") + "\n" +
		tt("│█│") + "   " + b("▄███▄") + "   " + tt("│█│") + "\n" +
		tt("│▓│") + "  " + b("█▒") + e("◉") + b("▒") + e("◉") + b("▒█") + "  " + tt("│▓│") + "\n" +
		h("╰┬╯") + "  " + b("█▒╰─╯▒█") + "  " + h("╰┬╯") + "\n" +
		h(" │") + "  " + b("█▒▒▒▒▒▒▒█") + "  " + h("│") + "\n" +
		h(" │") + "  " + b("█▒╭───╮▒█") + "  " + h("│") + "\n" +
		h(" │") + "  " + b("█▒│") + " " + k("⚷") + " " + b("│▒█") + "  " + h("│") + "\n" +
		h(" │") + "  " + b("█▒╰─┬─╯▒█") + "  " + h("│") + "\n" +
		h("╭┴╮") + "  " + b("▀█▄│▄█▀") + "  " + h("╭┴╮") + "\n" +
		h("╚═╝") + "     " + b("│") + "     " + h("╚═╝") + "\n" +
		"\n" +
		"  " + tt("🔥") + "  " + k("🗝️") + "  " + tt("🔥") + "\n" +
		"\n" +
		tx("Welcome to Hecate") + "\n" +
		tx("Press i to begin")
}

// FormatTokens formats token count nicely using theme colors.
func FormatTokens(count int, t *theme.Theme) string {
	txt := ""
	if count >= 1000 {
		txt = formatK(count) + " tokens"
	} else {
		txt = itoa(count) + " tokens"
	}
	return lipgloss.NewStyle().Foreground(t.Secondary).Render(txt)
}

// FormatSpeed formats tokens per second using theme colors.
func FormatSpeed(tokensPerSec float64, t *theme.Theme) string {
	txt := ""
	if tokensPerSec >= 100 {
		txt = itoa(int(tokensPerSec)) + " tok/s"
	} else {
		txt = formatFloat(tokensPerSec) + " tok/s"
	}
	return lipgloss.NewStyle().Foreground(t.Success).Render(txt)
}

func formatK(n int) string {
	if n >= 1000 {
		return itoa(n/1000) + "." + itoa((n%1000)/100) + "k"
	}
	return itoa(n)
}

func formatFloat(f float64) string {
	whole := int(f)
	frac := int((f - float64(whole)) * 10)
	return itoa(whole) + "." + itoa(frac)
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
