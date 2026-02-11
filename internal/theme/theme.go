package theme

import "github.com/charmbracelet/lipgloss"

// Theme defines the semantic color palette for the entire TUI.
// Every component receives a *Theme and uses it for all styling.
type Theme struct {
	Name string

	// Core palette
	Primary      lipgloss.Color
	PrimaryLight lipgloss.Color
	PrimaryDark  lipgloss.Color
	Secondary    lipgloss.Color
	SecondaryDark lipgloss.Color
	Accent       lipgloss.Color

	// Semantic colors
	Success lipgloss.Color
	Warning lipgloss.Color
	Error   lipgloss.Color

	// Text
	Text     lipgloss.Color
	TextDim  lipgloss.Color
	TextMuted lipgloss.Color

	// Backgrounds
	BgPrimary lipgloss.Color
	BgChat    lipgloss.Color
	BgInput   lipgloss.Color
	BgCard    lipgloss.Color

	// Borders
	Border      lipgloss.Color
	BorderFocus lipgloss.Color

	// Chat-specific
	UserBubbleBg      lipgloss.Color
	UserBubbleFg      lipgloss.Color
	AssistantBubbleBg lipgloss.Color
	AssistantBubbleFg lipgloss.Color
	SystemBubbleBg    lipgloss.Color
	SystemBubbleFg    lipgloss.Color

	// Status bar
	StatusBarBg lipgloss.Color
	StatusBarFg lipgloss.Color
	ModeLabelBg lipgloss.Color
	ModeLabelFg lipgloss.Color

	// Streaming
	StreamingColor lipgloss.Color
	ThinkingColor  lipgloss.Color

	// Code blocks
	CodeBg   lipgloss.Color
	CodeText lipgloss.Color

	// Accents for special elements
	KeyColor   lipgloss.Color
	VentureColor lipgloss.Color
	EyeColor   lipgloss.Color
}

// Computed styles derived from theme colors.
// These are recalculated when the theme changes.
type Styles struct {
	// Mode labels
	NormalMode  lipgloss.Style
	InsertMode  lipgloss.Style
	CommandMode lipgloss.Style
	BrowseMode   lipgloss.Style
	PairMode     lipgloss.Style
	EditMode     lipgloss.Style
	ProjectsMode lipgloss.Style

	// Chat
	UserBubble      lipgloss.Style
	AssistantBubble lipgloss.Style
	SystemBubble    lipgloss.Style
	UserLabel       lipgloss.Style
	AssistantLabel  lipgloss.Style

	// Input
	InputActive  lipgloss.Style
	InputBlurred lipgloss.Style
	CommandLine  lipgloss.Style

	// Cards (inline command output)
	Card       lipgloss.Style
	CardTitle  lipgloss.Style
	CardLabel  lipgloss.Style
	CardValue  lipgloss.Style
	CardBorder lipgloss.Style

	// Status bar
	StatusBar      lipgloss.Style
	StatusOK       lipgloss.Style
	StatusWarning  lipgloss.Style
	StatusError    lipgloss.Style

	// General
	Title    lipgloss.Style
	Subtle   lipgloss.Style
	Bold     lipgloss.Style
	Help     lipgloss.Style
	Error    lipgloss.Style
	Divider  lipgloss.Style
	Selected lipgloss.Style
	Label    lipgloss.Style
	Value    lipgloss.Style
}

// ComputeStyles generates all lipgloss styles from theme colors.
func (t *Theme) ComputeStyles() *Styles {
	return &Styles{
		// Mode labels - each mode gets a distinct color
		NormalMode: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.StatusBarFg).
			Background(t.Primary).
			Padding(0, 1),
		InsertMode: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.StatusBarFg).
			Background(t.Success).
			Padding(0, 1),
		CommandMode: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.StatusBarFg).
			Background(t.Warning).
			Padding(0, 1),
		BrowseMode: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.StatusBarFg).
			Background(t.Secondary).
			Padding(0, 1),
		PairMode: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.StatusBarFg).
			Background(t.Accent).
			Padding(0, 1),
		EditMode: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.StatusBarFg).
			Background(t.PrimaryDark).
			Padding(0, 1),
		ProjectsMode: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.StatusBarFg).
			Background(t.SecondaryDark).
			Padding(0, 1),

		// Chat bubbles
		UserBubble: lipgloss.NewStyle().
			Foreground(t.KeyColor).
			Bold(true),
		AssistantBubble: lipgloss.NewStyle().
			Foreground(t.AssistantBubbleFg),
		SystemBubble: lipgloss.NewStyle().
			Foreground(t.SystemBubbleFg).
			PaddingLeft(1).
			Border(lipgloss.Border{Left: "│"}).
			BorderForeground(t.Primary),
		UserLabel: lipgloss.NewStyle().
			Foreground(t.PrimaryLight).
			Bold(true),
		AssistantLabel: lipgloss.NewStyle().
			Foreground(t.Secondary).
			Bold(true),

		// Input
		InputActive: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.BorderFocus).
			Padding(0, 1),
		InputBlurred: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Border).
			Padding(0, 1),
		CommandLine: lipgloss.NewStyle().
			Foreground(t.Text).
			Background(t.BgInput).
			Padding(0, 1),

		// Cards
		Card: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Border).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1),
		CardTitle: lipgloss.NewStyle().
			Foreground(t.Primary).
			Bold(true),
		CardLabel: lipgloss.NewStyle().
			Foreground(t.TextMuted).
			Width(12).
			Align(lipgloss.Right),
		CardValue: lipgloss.NewStyle().
			Foreground(t.Text),
		CardBorder: lipgloss.NewStyle().
			Foreground(t.Border),

		// Status bar
		StatusBar: lipgloss.NewStyle().
			Foreground(t.StatusBarFg).
			Background(t.StatusBarBg),
		StatusOK: lipgloss.NewStyle().
			Foreground(t.Success).
			Bold(true),
		StatusWarning: lipgloss.NewStyle().
			Foreground(t.Warning).
			Bold(true),
		StatusError: lipgloss.NewStyle().
			Foreground(t.Error).
			Bold(true),

		// General
		Title: lipgloss.NewStyle().
			Foreground(t.Primary).
			Bold(true),
		Subtle: lipgloss.NewStyle().
			Foreground(t.TextMuted),
		Bold: lipgloss.NewStyle().
			Foreground(t.Text).
			Bold(true),
		Help: lipgloss.NewStyle().
			Foreground(t.TextMuted),
		Error: lipgloss.NewStyle().
			Foreground(t.Error).
			Bold(true),
		Divider: lipgloss.NewStyle().
			Foreground(t.Border),
		Selected: lipgloss.NewStyle().
			Foreground(t.Text).
			Background(t.Primary),
		Label: lipgloss.NewStyle().
			Foreground(t.TextMuted),
		Value: lipgloss.NewStyle().
			Foreground(t.Text),
	}
}
