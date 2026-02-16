package node

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/llm"
)

// View renders the Ops Studio content area.
func (s *Studio) View() string {
	if s.width == 0 {
		return ""
	}

	// Action overlay views take priority when active
	if s.actionMode == actionViewForm && s.formReady && s.formView != nil {
		return s.viewForm()
	}
	if s.actionMode == actionViewCategories {
		return s.viewCategories()
	}
	if s.actionMode == actionViewActions {
		return s.viewActions()
	}

	if s.loading {
		return s.viewLoading()
	}
	if s.loadErr != nil {
		return s.viewError()
	}

	switch s.activeView {
	case viewDashboard:
		return s.viewDashboard()
	case viewModels:
		return s.viewModels()
	case viewProviders:
		return s.viewProviders()
	case viewCapabilities:
		return s.viewCapabilities()
	case viewHealth:
		return s.viewHealthDetail()
	}

	return s.viewDashboard()
}

func (s *Studio) viewLoading() string {
	t := s.ctx.Theme
	msg := lipgloss.NewStyle().
		Foreground(t.TextDim).
		Render("Loading node status...")
	return lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, msg)
}

func (s *Studio) viewError() string {
	t := s.ctx.Theme
	title := lipgloss.NewStyle().
		Foreground(t.Error).Bold(true).
		Render("Failed to connect to daemon")

	detail := lipgloss.NewStyle().
		Foreground(t.TextDim).
		Render(s.loadErr.Error())

	hint := lipgloss.NewStyle().
		Foreground(t.TextMuted).Italic(true).
		Render("Press r to retry")

	content := title + "\n\n" + detail + "\n\n" + hint
	return lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, content)
}

// viewDashboard renders the main ops dashboard — single-glance node overview.
func (s *Studio) viewDashboard() string {
	t := s.ctx.Theme
	var b strings.Builder
	contentWidth := min(s.width, 65)

	// Header: Node name + uptime
	nodeName := "unknown"
	uptime := ""
	version := ""
	if s.health != nil {
		if s.health.Version != "" {
			version = s.health.Version
		}
		uptime = formatUptime(s.health.UptimeSeconds)
	}
	if s.identity != nil && s.identity.Identity != "" {
		// Extract short node name from identity (e.g., "rl@beam00" from MRI)
		nodeName = extractNodeName(s.identity.Identity)
	}

	nodeLabel := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).
		Render("Node: " + nodeName)
	uptimeLabel := lipgloss.NewStyle().Foreground(t.TextDim).
		Render("Uptime: " + uptime)

	// Right-align uptime on the same line
	gap := contentWidth - lipgloss.Width(nodeLabel) - lipgloss.Width(uptimeLabel)
	if gap < 2 {
		gap = 2
	}
	b.WriteString(nodeLabel + strings.Repeat(" ", gap) + uptimeLabel + "\n")

	// Separator
	sep := lipgloss.NewStyle().Foreground(t.Border).
		Render(strings.Repeat("\u2500", contentWidth))
	b.WriteString(sep + "\n\n")

	// Identity
	labelStyle := lipgloss.NewStyle().Foreground(t.TextMuted).Width(16).Align(lipgloss.Right)
	valueStyle := lipgloss.NewStyle().Foreground(t.Text)

	if s.identity != nil {
		b.WriteString(labelStyle.Render("Identity") + "  " + valueStyle.Render(s.identity.Identity) + "\n")
	}

	// Daemon status
	daemonStatus := "\u25cf healthy"
	statusStyle := lipgloss.NewStyle().Foreground(t.Success)
	if s.health == nil || s.health.Status != "healthy" {
		daemonStatus = "\u25cb unhealthy"
		statusStyle = lipgloss.NewStyle().Foreground(t.Error)
	}
	daemonValue := statusStyle.Render(daemonStatus)
	if version != "" {
		daemonValue += valueStyle.Render("     " + version)
	}
	b.WriteString(labelStyle.Render("Daemon") + "  " + daemonValue + "\n")

	b.WriteString("\n")

	// LLM Providers section
	sectionStyle := lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	b.WriteString("  " + sectionStyle.Render("LLM Providers") + "\n")

	if len(s.providers) == 0 {
		b.WriteString("    " + lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).
			Render("No providers configured") + "\n")
	} else {
		providerNames := sortedProviderNames(s.providers)
		for _, name := range providerNames {
			p := s.providers[name]
			b.WriteString(s.renderProviderLine(name, p) + "\n")
		}
	}

	b.WriteString("\n")

	// Capabilities summary
	capCount := len(s.capabilities)
	b.WriteString(labelStyle.Render("Capabilities") + "  " +
		valueStyle.Render(pluralize(capCount, "announced", "announced")) + "\n")

	// Agents summary
	agentCount := len(s.agents)
	b.WriteString(labelStyle.Render("Agents") + "  " +
		valueStyle.Render(pluralize(agentCount, "active", "active")) + "\n")

	// Models summary
	modelCount := len(s.models)
	b.WriteString(labelStyle.Render("Models") + "  " +
		valueStyle.Render(pluralize(modelCount, "available", "available")) + "\n")

	return b.String()
}

// renderProviderLine renders a single provider row for the dashboard.
func (s *Studio) renderProviderLine(name string, p llm.Provider) string {
	t := s.ctx.Theme

	// Status indicator
	var indicator string
	var indicatorStyle lipgloss.Style
	if p.Enabled {
		indicator = "\u25cf"
		indicatorStyle = lipgloss.NewStyle().Foreground(t.Success)
	} else {
		indicator = "\u25cb"
		indicatorStyle = lipgloss.NewStyle().Foreground(t.TextMuted)
	}

	// Count models for this provider
	modelCount := 0
	for _, m := range s.models {
		if m.Provider == name {
			modelCount++
		}
	}

	nameStyle := lipgloss.NewStyle().Foreground(t.Text).Width(16)
	countStyle := lipgloss.NewStyle().Foreground(t.TextDim)
	urlStyle := lipgloss.NewStyle().Foreground(t.TextMuted)

	line := fmt.Sprintf("    %s %s %s",
		indicatorStyle.Render(indicator),
		nameStyle.Render(name),
		countStyle.Render(pluralize(modelCount, "model", "models")),
	)

	if p.URL != "" {
		line += "     " + urlStyle.Render(p.URL)
	} else if !p.Enabled {
		line += "     " + urlStyle.Render("\u2014 (no key)")
	}

	return line
}

// viewModels renders the models list sub-view.
func (s *Studio) viewModels() string {
	t := s.ctx.Theme
	var b strings.Builder

	// Breadcrumb header
	b.WriteString(s.renderBreadcrumb("Models"))

	if len(s.models) == 0 {
		b.WriteString("\n  " + lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).
			Render("No models available") + "\n")
		return b.String()
	}

	// Column headers
	headerStyle := lipgloss.NewStyle().Foreground(t.TextMuted).Bold(true)
	b.WriteString(fmt.Sprintf("  %s %s %s %s\n",
		headerStyle.Render(padRight("Model", 26)),
		headerStyle.Render(padRight("Provider", 14)),
		headerStyle.Render(padRight("Size", 10)),
		headerStyle.Render("Params"),
	))

	colSep := lipgloss.NewStyle().Foreground(t.Border).
		Render("  " + strings.Repeat("\u2500", min(s.width-4, 60)))
	b.WriteString(colSep + "\n")

	// Model rows with scrolling
	maxRows := s.maxVisibleRows()
	for i := s.scrollOffset; i < len(s.models) && i-s.scrollOffset < maxRows; i++ {
		m := s.models[i]
		selected := i == s.cursor
		b.WriteString(s.renderModelRow(m, selected))
		if i < len(s.models)-1 && i-s.scrollOffset < maxRows-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// renderModelRow renders a single model row.
func (s *Studio) renderModelRow(m llm.Model, selected bool) string {
	t := s.ctx.Theme

	cursor := "  "
	if selected {
		cursor = lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("\u25b8 ")
	}

	nameStyle := lipgloss.NewStyle().Foreground(t.Text)
	provStyle := lipgloss.NewStyle().Foreground(t.TextDim)
	sizeStyle := lipgloss.NewStyle().Foreground(t.TextMuted)

	size := m.Size
	if size == "" {
		size = "\u2014"
	}

	params := m.ParameterSize
	if params == "" {
		params = "\u2014"
	}

	row := fmt.Sprintf("%s%s %s %s %s",
		cursor,
		nameStyle.Render(padRight(m.Name, 26)),
		provStyle.Render(padRight(m.Provider, 14)),
		sizeStyle.Render(padRight(size, 10)),
		sizeStyle.Render(params),
	)

	if selected {
		row = lipgloss.NewStyle().Bold(true).Render(row)
	}

	return row
}

// viewProviders renders the providers list sub-view.
func (s *Studio) viewProviders() string {
	t := s.ctx.Theme
	var b strings.Builder

	b.WriteString(s.renderBreadcrumb("Providers"))

	if len(s.providers) == 0 {
		b.WriteString("\n  " + lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).
			Render("No providers configured") + "\n")
		return b.String()
	}

	// Column headers
	headerStyle := lipgloss.NewStyle().Foreground(t.TextMuted).Bold(true)
	b.WriteString(fmt.Sprintf("  %s %s %s %s %s\n",
		headerStyle.Render("  "),
		headerStyle.Render(padRight("Name", 16)),
		headerStyle.Render(padRight("Type", 12)),
		headerStyle.Render(padRight("Status", 10)),
		headerStyle.Render("URL"),
	))

	colSep := lipgloss.NewStyle().Foreground(t.Border).
		Render("  " + strings.Repeat("\u2500", min(s.width-4, 60)))
	b.WriteString(colSep + "\n")

	names := sortedProviderNames(s.providers)
	maxRows := s.maxVisibleRows()
	for i := s.scrollOffset; i < len(names) && i-s.scrollOffset < maxRows; i++ {
		name := names[i]
		p := s.providers[name]
		selected := i == s.cursor
		b.WriteString(s.renderProviderRow(name, p, selected))
		if i < len(names)-1 && i-s.scrollOffset < maxRows-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// renderProviderRow renders a single provider row.
func (s *Studio) renderProviderRow(name string, p llm.Provider, selected bool) string {
	t := s.ctx.Theme

	cursor := "  "
	if selected {
		cursor = lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("\u25b8 ")
	}

	var indicator string
	var indicatorStyle lipgloss.Style
	if p.Enabled {
		indicator = "\u25cf"
		indicatorStyle = lipgloss.NewStyle().Foreground(t.Success)
	} else {
		indicator = "\u25cb"
		indicatorStyle = lipgloss.NewStyle().Foreground(t.TextMuted)
	}

	nameStyle := lipgloss.NewStyle().Foreground(t.Text)
	typeStyle := lipgloss.NewStyle().Foreground(t.TextDim)
	urlStyle := lipgloss.NewStyle().Foreground(t.TextMuted)

	statusText := "enabled"
	if !p.Enabled {
		statusText = "disabled"
	}

	url := p.URL
	if url == "" {
		url = "\u2014"
	}

	row := fmt.Sprintf("%s%s %s %s %s %s",
		cursor,
		indicatorStyle.Render(indicator),
		nameStyle.Render(padRight(name, 16)),
		typeStyle.Render(padRight(p.Type, 12)),
		typeStyle.Render(padRight(statusText, 10)),
		urlStyle.Render(url),
	)

	if selected {
		row = lipgloss.NewStyle().Bold(true).Render(row)
	}

	return row
}

// viewCapabilities renders the capabilities list sub-view.
func (s *Studio) viewCapabilities() string {
	t := s.ctx.Theme
	var b strings.Builder

	b.WriteString(s.renderBreadcrumb("Capabilities"))

	if len(s.capabilities) == 0 {
		b.WriteString("\n  " + lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).
			Render("No capabilities announced") + "\n")
		return b.String()
	}

	// Column headers
	headerStyle := lipgloss.NewStyle().Foreground(t.TextMuted).Bold(true)
	b.WriteString(fmt.Sprintf("  %s %s %s\n",
		headerStyle.Render(padRight("MRI", 36)),
		headerStyle.Render(padRight("Description", 24)),
		headerStyle.Render("Tags"),
	))

	colSep := lipgloss.NewStyle().Foreground(t.Border).
		Render("  " + strings.Repeat("\u2500", min(s.width-4, 65)))
	b.WriteString(colSep + "\n")

	maxRows := s.maxVisibleRows()
	for i := s.scrollOffset; i < len(s.capabilities) && i-s.scrollOffset < maxRows; i++ {
		cap := s.capabilities[i]
		selected := i == s.cursor
		b.WriteString(s.renderCapabilityRow(cap, selected))
		if i < len(s.capabilities)-1 && i-s.scrollOffset < maxRows-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// renderCapabilityRow renders a single capability row.
func (s *Studio) renderCapabilityRow(cap client.Capability, selected bool) string {
	t := s.ctx.Theme

	cursor := "  "
	if selected {
		cursor = lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("\u25b8 ")
	}

	mriStyle := lipgloss.NewStyle().Foreground(t.Text)
	descStyle := lipgloss.NewStyle().Foreground(t.TextDim)
	tagStyle := lipgloss.NewStyle().Foreground(t.Secondary)

	mri := truncate(cap.MRI, 34)
	desc := truncate(cap.Description, 22)
	tags := strings.Join(cap.Tags, ", ")
	if tags == "" {
		tags = "\u2014"
	}

	row := fmt.Sprintf("%s%s %s %s",
		cursor,
		mriStyle.Render(padRight(mri, 36)),
		descStyle.Render(padRight(desc, 24)),
		tagStyle.Render(tags),
	)

	if selected {
		row = lipgloss.NewStyle().Bold(true).Render(row)
	}

	return row
}

// viewHealthDetail renders the detailed health sub-view.
func (s *Studio) viewHealthDetail() string {
	t := s.ctx.Theme
	var b strings.Builder

	b.WriteString(s.renderBreadcrumb("Health"))
	b.WriteString("\n")

	labelStyle := lipgloss.NewStyle().Foreground(t.TextMuted).Width(16).Align(lipgloss.Right)
	valueStyle := lipgloss.NewStyle().Foreground(t.Text)

	if s.health != nil {
		statusStyle := lipgloss.NewStyle().Foreground(t.Success)
		if s.health.Status != "healthy" {
			statusStyle = lipgloss.NewStyle().Foreground(t.Error)
		}

		b.WriteString(labelStyle.Render("Status") + "  " +
			statusStyle.Render(s.health.Status) + "\n")
		b.WriteString(labelStyle.Render("Version") + "  " +
			valueStyle.Render(s.health.Version) + "\n")
		b.WriteString(labelStyle.Render("Uptime") + "  " +
			valueStyle.Render(formatUptime(s.health.UptimeSeconds)) + "\n")
	}

	if s.identity != nil {
		b.WriteString("\n")
		b.WriteString(labelStyle.Render("Identity") + "  " +
			valueStyle.Render(s.identity.Identity) + "\n")
		if s.identity.PublicKey != "" {
			pk := truncate(s.identity.PublicKey, 40)
			b.WriteString(labelStyle.Render("Public Key") + "  " +
				lipgloss.NewStyle().Foreground(t.TextDim).Render(pk) + "\n")
		}
	}

	// Provider health
	b.WriteString("\n")
	sectionStyle := lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	b.WriteString("  " + sectionStyle.Render("Provider Health") + "\n")

	if len(s.providers) == 0 {
		b.WriteString("    " + lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).
			Render("No providers") + "\n")
	} else {
		names := sortedProviderNames(s.providers)
		for _, name := range names {
			p := s.providers[name]
			indicator := "\u25cf"
			color := t.Success
			if !p.Enabled {
				indicator = "\u25cb"
				color = t.TextMuted
			}
			b.WriteString(fmt.Sprintf("    %s %s (%s)\n",
				lipgloss.NewStyle().Foreground(color).Render(indicator),
				valueStyle.Render(name),
				lipgloss.NewStyle().Foreground(t.TextDim).Render(p.Type),
			))
		}
	}

	// Summary counts
	b.WriteString("\n")
	b.WriteString("  " + sectionStyle.Render("Summary") + "\n")
	b.WriteString(fmt.Sprintf("    Models:        %s\n",
		valueStyle.Render(itoa(len(s.models)))))
	b.WriteString(fmt.Sprintf("    Capabilities:  %s\n",
		valueStyle.Render(itoa(len(s.capabilities)))))
	b.WriteString(fmt.Sprintf("    Agents:        %s\n",
		valueStyle.Render(itoa(len(s.agents)))))

	return b.String()
}

// viewCategories renders the category selection overlay.
func (s *Studio) viewCategories() string {
	t := s.ctx.Theme
	var b strings.Builder

	contentWidth := min(s.width, 50)

	// Title
	title := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).
		Render("Node Actions")
	b.WriteString(title + "\n")

	sep := lipgloss.NewStyle().Foreground(t.Border).
		Render(strings.Repeat("\u2500", contentWidth))
	b.WriteString(sep + "\n\n")

	// Hint
	hint := lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).
		Render("Select a category:")
	b.WriteString(hint + "\n\n")

	// Category list
	for i, cat := range s.categories {
		selected := i == s.catCursor

		cursor := "  "
		if selected {
			cursor = lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("\u25b8 ")
		}

		nameStyle := lipgloss.NewStyle().Foreground(t.Text)
		if selected {
			nameStyle = nameStyle.Bold(true)
		}

		actionCount := len(cat.Actions)
		countText := lipgloss.NewStyle().Foreground(t.TextDim).
			Render("  " + pluralize(actionCount, "action", "actions"))

		b.WriteString(fmt.Sprintf("%s%s %s%s\n",
			cursor,
			cat.Icon,
			nameStyle.Render(cat.Name),
			countText,
		))
	}

	// Center the menu
	content := b.String()
	return lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, content)
}

// viewActions renders the action selection list for the current category.
func (s *Studio) viewActions() string {
	t := s.ctx.Theme
	var b strings.Builder

	if s.catCursor >= len(s.categories) {
		return ""
	}

	cat := s.categories[s.catCursor]
	contentWidth := min(s.width, 50)

	// Breadcrumb: Node Actions > Category
	breadcrumb := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).
		Render("Node Actions") +
		lipgloss.NewStyle().Foreground(t.TextDim).Render(" \u276f ") +
		lipgloss.NewStyle().Foreground(t.Accent).Bold(true).
			Render(cat.Icon+" "+cat.Name)
	b.WriteString(breadcrumb + "\n")

	sep := lipgloss.NewStyle().Foreground(t.Border).
		Render(strings.Repeat("\u2500", contentWidth))
	b.WriteString(sep + "\n\n")

	if len(cat.Actions) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).
			Render("  No actions available") + "\n")
		content := b.String()
		return lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, content)
	}

	for i, action := range cat.Actions {
		selected := i == s.actionCursor

		cursor := "  "
		if selected {
			cursor = lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("\u25b8 ")
		}

		nameStyle := lipgloss.NewStyle().Foreground(t.Text)
		if selected {
			nameStyle = nameStyle.Bold(true)
		}

		// Show whether the action needs a form or is confirm-only
		suffix := ""
		if action.FormSpec == nil {
			suffix = lipgloss.NewStyle().Foreground(t.TextMuted).
				Render("  (instant)")
		}

		b.WriteString(fmt.Sprintf("%s%s%s\n", cursor, nameStyle.Render(action.Name), suffix))
	}

	content := b.String()
	return lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, content)
}

// viewForm renders the active form centered in the content area.
func (s *Studio) viewForm() string {
	if s.formView == nil {
		return ""
	}

	formContent := s.formView.View()
	return lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, formContent)
}

// renderBreadcrumb renders the sub-view breadcrumb header.
func (s *Studio) renderBreadcrumb(viewName string) string {
	t := s.ctx.Theme

	nodeName := "Node"
	if s.identity != nil {
		nodeName = extractNodeName(s.identity.Identity)
	}

	breadcrumb := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).
		Render(nodeName) +
		lipgloss.NewStyle().Foreground(t.TextDim).Render(" \u276f ") +
		lipgloss.NewStyle().Foreground(t.Accent).Bold(true).
			Render(viewName)

	sep := lipgloss.NewStyle().Foreground(t.Border).
		Render(strings.Repeat("\u2500", min(s.width, 60)))

	return breadcrumb + "\n" + sep + "\n"
}

// Helper functions

func extractNodeName(identity string) string {
	// Try to extract a human-readable name from the identity string
	// e.g., "mri:agent:io.hecate/rl@beam00" -> "beam00"
	if idx := strings.LastIndex(identity, "@"); idx >= 0 {
		return identity[idx+1:]
	}
	if idx := strings.LastIndex(identity, "/"); idx >= 0 {
		return identity[idx+1:]
	}
	if len(identity) > 20 {
		return identity[:20]
	}
	return identity
}

func sortedProviderNames(providers map[string]llm.Provider) []string {
	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

