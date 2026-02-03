package chat

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors - Extended palette for chat
	Purple      = lipgloss.Color("#7C3AED")
	PurpleLight = lipgloss.Color("#A78BFA")
	PurpleDark  = lipgloss.Color("#5B21B6")
	Emerald     = lipgloss.Color("#10B981")
	EmeraldDark = lipgloss.Color("#059669")
	Cyan        = lipgloss.Color("#06B6D4")
	Pink        = lipgloss.Color("#EC4899")
	Orange      = lipgloss.Color("#F97316")
	Yellow      = lipgloss.Color("#EAB308")
	Gray50      = lipgloss.Color("#F9FAFB")
	Gray100     = lipgloss.Color("#F3F4F6")
	Gray200     = lipgloss.Color("#E5E7EB")
	Gray300     = lipgloss.Color("#D1D5DB")
	Gray400     = lipgloss.Color("#9CA3AF")
	Gray500     = lipgloss.Color("#6B7280")
	Gray600     = lipgloss.Color("#4B5563")
	Gray700     = lipgloss.Color("#374151")
	Gray800     = lipgloss.Color("#1F2937")
	Gray900     = lipgloss.Color("#111827")

	// Header styles
	HeaderStyle = lipgloss.NewStyle().
			Foreground(Gray100).
			Bold(true).
			Padding(0, 1)

	ModelSelectorStyle = lipgloss.NewStyle().
				Foreground(Gray400).
				Padding(0, 1)

	ModelActiveStyle = lipgloss.NewStyle().
				Foreground(Gray900).
				Background(Purple).
				Bold(true).
				Padding(0, 2).
				MarginRight(1)

	ModelInactiveStyle = lipgloss.NewStyle().
				Foreground(Gray400).
				Padding(0, 2).
				MarginRight(1)

	// Message bubbles
	UserBubbleStyle = lipgloss.NewStyle().
			Foreground(Gray900).
			Background(Purple).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1).
			MarginLeft(4)

	AssistantBubbleStyle = lipgloss.NewStyle().
				Foreground(Gray100).
				Background(Gray700).
				Padding(1, 2).
				MarginTop(1).
				MarginBottom(1).
				MarginRight(4)

	SystemBubbleStyle = lipgloss.NewStyle().
				Foreground(Gray400).
				Background(Gray800).
				Italic(true).
				Padding(1, 2).
				MarginTop(1).
				MarginBottom(1)

	// Role labels
	UserLabelStyle = lipgloss.NewStyle().
			Foreground(PurpleLight).
			Bold(true).
			MarginLeft(4)

	AssistantLabelStyle = lipgloss.NewStyle().
				Foreground(Emerald).
				Bold(true)

	// Streaming indicator
	StreamingStyle = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true)

	ThinkingStyle = lipgloss.NewStyle().
			Foreground(Yellow).
			Italic(true)

	// Input area
	InputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Purple).
			Padding(0, 1)

	InputActiveStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(PurpleLight).
				Padding(0, 1)

	InputLabelStyle = lipgloss.NewStyle().
			Foreground(Gray400).
			MarginRight(1)

	// Stats bar
	StatsStyle = lipgloss.NewStyle().
			Foreground(Gray500).
			MarginTop(1)

	TokenCountStyle = lipgloss.NewStyle().
			Foreground(Cyan)

	SpeedStyle = lipgloss.NewStyle().
			Foreground(Emerald)

	// Empty state
	EmptyStateStyle = lipgloss.NewStyle().
			Foreground(Gray500).
			Italic(true).
			Align(lipgloss.Center)

	WelcomeStyle = lipgloss.NewStyle().
			Foreground(Purple).
			Bold(true).
			Align(lipgloss.Center)

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	// Help text
	HelpStyle = lipgloss.NewStyle().
			Foreground(Gray500)

	// Divider
	DividerStyle = lipgloss.NewStyle().
			Foreground(Gray700)
)

// Sparkles for streaming animation
var Sparkles = []string{"✦", "✧", "★", "☆", "✦"}

// ThinkingFrames for thinking animation
var ThinkingFrames = []string{
	"🔮 Thinking",
	"🔮 Thinking.",
	"🔮 Thinking..",
	"🔮 Thinking...",
}

// WelcomeArt returns ASCII art for empty chat
func WelcomeArt() string {
	return `
    ╭──────────────────────────────────╮
    │                                  │
    │      🗝️  Welcome to Hecate      │
    │                                  │
    │    Your AI companion awaits.     │
    │    Type a message to begin.      │
    │                                  │
    ╰──────────────────────────────────╯
`
}

// FormatTokens formats token count nicely
func FormatTokens(count int) string {
	if count >= 1000 {
		return TokenCountStyle.Render(formatK(count) + " tokens")
	}
	return TokenCountStyle.Render(itoa(count) + " tokens")
}

// FormatSpeed formats tokens per second
func FormatSpeed(tokensPerSec float64) string {
	if tokensPerSec >= 100 {
		return SpeedStyle.Render(itoa(int(tokensPerSec)) + " tok/s")
	}
	return SpeedStyle.Render(formatFloat(tokensPerSec) + " tok/s")
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
