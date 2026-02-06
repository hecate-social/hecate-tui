package theme

import "github.com/charmbracelet/lipgloss"

// HecateDark is the default theme — purple/emerald on dark backgrounds.
func HecateDark() *Theme {
	return &Theme{
		Name: "Hecate Dark",

		Primary:      lipgloss.Color("#7C3AED"),
		PrimaryLight: lipgloss.Color("#A78BFA"),
		PrimaryDark:  lipgloss.Color("#5B21B6"),
		Secondary:    lipgloss.Color("#10B981"),
		SecondaryDark: lipgloss.Color("#059669"),
		Accent:       lipgloss.Color("#EC4899"),

		Success: lipgloss.Color("#22C55E"),
		Warning: lipgloss.Color("#F59E0B"),
		Error:   lipgloss.Color("#EF4444"),

		Text:      lipgloss.Color("#F3F4F6"),
		TextDim:   lipgloss.Color("#9CA3AF"),
		TextMuted: lipgloss.Color("#6B7280"),

		BgPrimary: lipgloss.Color("#111827"),
		BgChat:    lipgloss.Color("#111827"),
		BgInput:   lipgloss.Color("#1F2937"),
		BgCard:    lipgloss.Color("#1F2937"),

		Border:      lipgloss.Color("#374151"),
		BorderFocus: lipgloss.Color("#A78BFA"),

		UserBubbleBg:      lipgloss.Color("#7C3AED"),
		UserBubbleFg:      lipgloss.Color("#111827"),
		AssistantBubbleBg: lipgloss.Color("#374151"),
		AssistantBubbleFg: lipgloss.Color("#F3F4F6"),
		SystemBubbleBg:    lipgloss.Color("#1F2937"),
		SystemBubbleFg:    lipgloss.Color("#9CA3AF"),

		StatusBarBg: lipgloss.Color("#1F2937"),
		StatusBarFg: lipgloss.Color("#F3F4F6"),
		ModeLabelBg: lipgloss.Color("#7C3AED"),
		ModeLabelFg: lipgloss.Color("#F3F4F6"),

		StreamingColor: lipgloss.Color("#06B6D4"),
		ThinkingColor:  lipgloss.Color("#EAB308"),

		CodeBg:   lipgloss.Color("#0D1117"),
		CodeText: lipgloss.Color("#E6EDF3"),

		KeyColor:   lipgloss.Color("#FCD34D"),
		TorchColor: lipgloss.Color("#F97316"),
		EyeColor:   lipgloss.Color("#F59E0B"),
	}
}

// HecateLight is for daylight/high-contrast environments.
func HecateLight() *Theme {
	return &Theme{
		Name: "Hecate Light",

		Primary:      lipgloss.Color("#6D28D9"),
		PrimaryLight: lipgloss.Color("#8B5CF6"),
		PrimaryDark:  lipgloss.Color("#4C1D95"),
		Secondary:    lipgloss.Color("#059669"),
		SecondaryDark: lipgloss.Color("#047857"),
		Accent:       lipgloss.Color("#DB2777"),

		Success: lipgloss.Color("#16A34A"),
		Warning: lipgloss.Color("#D97706"),
		Error:   lipgloss.Color("#DC2626"),

		Text:      lipgloss.Color("#111827"),
		TextDim:   lipgloss.Color("#4B5563"),
		TextMuted: lipgloss.Color("#6B7280"),

		BgPrimary: lipgloss.Color("#FFFFFF"),
		BgChat:    lipgloss.Color("#F9FAFB"),
		BgInput:   lipgloss.Color("#F3F4F6"),
		BgCard:    lipgloss.Color("#F3F4F6"),

		Border:      lipgloss.Color("#D1D5DB"),
		BorderFocus: lipgloss.Color("#8B5CF6"),

		UserBubbleBg:      lipgloss.Color("#6D28D9"),
		UserBubbleFg:      lipgloss.Color("#FFFFFF"),
		AssistantBubbleBg: lipgloss.Color("#E5E7EB"),
		AssistantBubbleFg: lipgloss.Color("#111827"),
		SystemBubbleBg:    lipgloss.Color("#F3F4F6"),
		SystemBubbleFg:    lipgloss.Color("#4B5563"),

		StatusBarBg: lipgloss.Color("#E5E7EB"),
		StatusBarFg: lipgloss.Color("#111827"),
		ModeLabelBg: lipgloss.Color("#6D28D9"),
		ModeLabelFg: lipgloss.Color("#FFFFFF"),

		StreamingColor: lipgloss.Color("#0891B2"),
		ThinkingColor:  lipgloss.Color("#CA8A04"),

		CodeBg:   lipgloss.Color("#F6F8FA"),
		CodeText: lipgloss.Color("#24292F"),

		KeyColor:   lipgloss.Color("#D97706"),
		TorchColor: lipgloss.Color("#EA580C"),
		EyeColor:   lipgloss.Color("#D97706"),
	}
}

// Monochrome is for terminal purists — works everywhere.
func Monochrome() *Theme {
	return &Theme{
		Name: "Monochrome",

		Primary:      lipgloss.Color("#FFFFFF"),
		PrimaryLight: lipgloss.Color("#E5E5E5"),
		PrimaryDark:  lipgloss.Color("#A3A3A3"),
		Secondary:    lipgloss.Color("#D4D4D4"),
		SecondaryDark: lipgloss.Color("#A3A3A3"),
		Accent:       lipgloss.Color("#FFFFFF"),

		Success: lipgloss.Color("#D4D4D4"),
		Warning: lipgloss.Color("#D4D4D4"),
		Error:   lipgloss.Color("#FFFFFF"),

		Text:      lipgloss.Color("#FFFFFF"),
		TextDim:   lipgloss.Color("#A3A3A3"),
		TextMuted: lipgloss.Color("#737373"),

		BgPrimary: lipgloss.Color("#000000"),
		BgChat:    lipgloss.Color("#000000"),
		BgInput:   lipgloss.Color("#171717"),
		BgCard:    lipgloss.Color("#171717"),

		Border:      lipgloss.Color("#404040"),
		BorderFocus: lipgloss.Color("#FFFFFF"),

		UserBubbleBg:      lipgloss.Color("#404040"),
		UserBubbleFg:      lipgloss.Color("#FFFFFF"),
		AssistantBubbleBg: lipgloss.Color("#262626"),
		AssistantBubbleFg: lipgloss.Color("#D4D4D4"),
		SystemBubbleBg:    lipgloss.Color("#171717"),
		SystemBubbleFg:    lipgloss.Color("#A3A3A3"),

		StatusBarBg: lipgloss.Color("#262626"),
		StatusBarFg: lipgloss.Color("#D4D4D4"),
		ModeLabelBg: lipgloss.Color("#FFFFFF"),
		ModeLabelFg: lipgloss.Color("#000000"),

		StreamingColor: lipgloss.Color("#D4D4D4"),
		ThinkingColor:  lipgloss.Color("#A3A3A3"),

		CodeBg:   lipgloss.Color("#0A0A0A"),
		CodeText: lipgloss.Color("#D4D4D4"),

		KeyColor:   lipgloss.Color("#FFFFFF"),
		TorchColor: lipgloss.Color("#D4D4D4"),
		EyeColor:   lipgloss.Color("#FFFFFF"),
	}
}

// BuiltinThemes returns all available themes keyed by name.
func BuiltinThemes() map[string]*Theme {
	return map[string]*Theme{
		"dark":       HecateDark(),
		"light":      HecateLight(),
		"monochrome": Monochrome(),
	}
}
