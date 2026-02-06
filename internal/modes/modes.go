package modes

// Mode represents the current input mode of the TUI.
type Mode int

const (
	Normal  Mode = iota // Resting state — scroll chat, press keys
	Insert              // Typing a message — textarea focused
	Command             // Slash command entry — command line at bottom
	Browse              // Browsing capabilities — modal dialog
	Pair                // Pairing flow — inline wizard
	Edit                // Built-in editor — file editing overlay
)

// String returns the display name for the mode (shown in status bar).
func (m Mode) String() string {
	switch m {
	case Normal:
		return "NORMAL"
	case Insert:
		return "INSERT"
	case Command:
		return "COMMAND"
	case Browse:
		return "BROWSE"
	case Pair:
		return "PAIR"
	case Edit:
		return "EDIT"
	default:
		return "UNKNOWN"
	}
}

// Hints returns contextual keybinding hints for the status bar.
func (m Mode) Hints() string {
	switch m {
	case Normal:
		return "i:chat  /:cmd  j/k:scroll  r:retry  y:copy  ?:help  q:quit"
	case Insert:
		return "Enter:send  Alt+Enter:newline  Tab:model  Esc:normal"
	case Command:
		return "Enter:exec  Tab:complete  Esc:cancel"
	case Browse:
		return "j/k:nav  Enter:detail  /:filter  Esc:back"
	case Pair:
		return "p:pair  c:cancel  r:refresh  Esc:back"
	case Edit:
		return "Ctrl+S:save  Ctrl+Q:close  Esc:close"
	default:
		return ""
	}
}

// ModeChangedMsg is emitted when the mode transitions.
type ModeChangedMsg struct {
	From Mode
	To   Mode
}
