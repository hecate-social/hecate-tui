package stables

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/studios/arcade/snake_duel"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

// Colors for stables UI.
var (
	colorTraining  = lipgloss.Color("#fbbf24") // amber
	colorCompleted = lipgloss.Color("#34d399") // green
	colorHalted    = lipgloss.Color("#f87171") // red
	colorFitness   = lipgloss.Color("#60a5fa") // blue
	colorChampion  = lipgloss.Color("#a78bfa") // purple
)

// view dispatches to the current phase view.
func (m *Model) view() string {
	if m.width == 0 {
		return ""
	}

	switch m.phase {
	case phaseList:
		return m.viewList()
	case phaseNewStable:
		return m.viewNewStable()
	case phaseDetail:
		return m.viewDetail()
	case phaseDuel:
		return m.viewDuel()
	case phaseHeroes:
		return m.viewHeroes()
	case phaseHeroDetail:
		return m.viewHeroDetail()
	case phasePromote:
		return m.viewPromote()
	case phaseHeroDuel:
		return m.viewHeroDuel()
	default:
		return m.viewList()
	}
}

// viewList renders the stables list.
func (m *Model) viewList() string {
	t := m.ctx.Theme

	title := lipgloss.NewStyle().
		Foreground(t.Primary).Bold(true).
		Render("Stables")

	subtitle := lipgloss.NewStyle().
		Foreground(t.TextDim).
		Render("Neuroevolution snake gladiator training")

	var content string

	if !m.listLoaded {
		content = lipgloss.NewStyle().
			Foreground(t.TextDim).Italic(true).
			Render("Loading stables...")
	} else if len(m.stables) == 0 {
		content = lipgloss.NewStyle().
			Foreground(t.TextMuted).Italic(true).
			Render("No stables yet. Press n to create one.")
	} else {
		content = m.renderStablesTable(t)
	}

	errStr := m.renderError(t)

	hints := lipgloss.NewStyle().
		Foreground(t.TextMuted).Italic(true).
		Render("j/k:navigate  Enter:open  n:new  H:heroes  r:refresh  esc:back")

	parts := title + "\n" + subtitle + "\n\n" + content
	if errStr != "" {
		parts += "\n\n" + errStr
	}
	parts += "\n\n" + hints

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, parts)
}

// renderStablesTable renders the stables as a table.
func (m *Model) renderStablesTable(t *theme.Theme) string {
	headerStyle := lipgloss.NewStyle().Foreground(t.TextDim).Bold(true)
	header := headerStyle.Render(fmt.Sprintf(
		"  %-12s %-10s %6s %6s %8s %8s",
		"Status", "ID", "Pop", "Gens", "Best", "Gen#"))

	var rows []string
	rows = append(rows, header)

	for i, s := range m.stables {
		selected := i == m.listIndex
		rows = append(rows, m.renderStableRow(t, s, selected))
	}

	return strings.Join(rows, "\n")
}

// renderStableRow renders a single stable row.
func (m *Model) renderStableRow(t *theme.Theme, s Stable, selected bool) string {
	statusBadge := renderStatusBadge(s.Status)

	shortID := s.StableID
	if len(shortID) > 10 {
		shortID = shortID[len(shortID)-8:]
	}

	row := fmt.Sprintf(" %s %-10s %6d %6d %8.1f %8d",
		statusBadge, shortID,
		s.PopulationSize, s.MaxGenerations,
		s.BestFitness, s.GenerationsCompleted)

	style := lipgloss.NewStyle().Foreground(t.Text)
	if selected {
		style = style.Foreground(t.Primary).Bold(true)
		row = ">" + row[1:]
	}

	return style.Render(row)
}

// renderStatusBadge returns a colored status indicator.
func renderStatusBadge(status string) string {
	switch status {
	case "training":
		return lipgloss.NewStyle().Foreground(colorTraining).Render("*training ")
	case "completed":
		return lipgloss.NewStyle().Foreground(colorCompleted).Render("*completed")
	case "halted":
		return lipgloss.NewStyle().Foreground(colorHalted).Render("*halted   ")
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("*" + status)
	}
}

// viewNewStable renders the new stable creation form.
func (m *Model) viewNewStable() string {
	t := m.ctx.Theme

	title := lipgloss.NewStyle().
		Foreground(t.Primary).Bold(true).
		Render("New Stable")

	subtitle := lipgloss.NewStyle().
		Foreground(t.TextDim).
		Render("Configure and start training")

	if m.formSeedID != "" {
		seedInfo := lipgloss.NewStyle().
			Foreground(colorChampion).
			Render("Seeding from: " + truncateID(m.formSeedID))
		subtitle += "\n" + seedInfo
	}

	var fields []string
	for i, label := range m.formLabels {
		focused := i == m.formFocused
		fields = append(fields, m.renderFormField(t, label, m.formFields[i], focused))
	}
	form := strings.Join(fields, "\n")

	// Preset selector
	presetNames := []string{"Balanced", "Aggressive", "Forager", "Survivor", "Assassin"}
	presetLine := lipgloss.NewStyle().Foreground(colorChampion).Bold(true).
		Render("Preset: " + presetNames[m.formPreset])
	presetHint := lipgloss.NewStyle().Foreground(t.TextMuted).
		Render("  (p to cycle)")

	// Budget bar
	budgetLine := m.renderBudgetBar(t)

	form += "\n\n" + presetLine + presetHint + "\n" + budgetLine

	// Advanced weights (toggle with w)
	if m.formShowWeights {
		form += "\n\n" + m.renderWeightsSection(t)
	}

	errStr := m.renderError(t)

	hints := lipgloss.NewStyle().
		Foreground(t.TextMuted).Italic(true).
		Render("Tab:next field  +/-:adjust  Enter:create  esc:cancel")

	parts := title + "\n" + subtitle + "\n\n" + form
	if errStr != "" {
		parts += "\n\n" + errStr
	}
	parts += "\n\n" + hints

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, parts)
}

// renderFormField renders a single form field.
func (m *Model) renderFormField(t *theme.Theme, label string, value int, focused bool) string {
	labelStyle := lipgloss.NewStyle().Foreground(t.TextDim).Width(20)
	valueStyle := lipgloss.NewStyle().Foreground(t.Text).Bold(true)

	if focused {
		labelStyle = labelStyle.Foreground(t.Primary)
		valueStyle = valueStyle.Foreground(t.Primary)
	}

	indicator := "  "
	if focused {
		indicator = "> "
	}

	return indicator + labelStyle.Render(label+":") + " " + valueStyle.Render(fmt.Sprintf("%d", value))
}

// viewDetail renders the stable detail / training monitor.
func (m *Model) viewDetail() string {
	t := m.ctx.Theme

	s := m.selectedStable
	title := lipgloss.NewStyle().
		Foreground(t.Primary).Bold(true).
		Render("Stable: " + truncateID(s.StableID))

	statusLine := renderStatusBadge(s.Status) + "  " +
		lipgloss.NewStyle().Foreground(t.TextDim).
			Render(fmt.Sprintf("Pop:%d  MaxGen:%d  AF:%d  Eps:%d",
				s.PopulationSize, s.MaxGenerations, s.OpponentAF, s.EpisodesPerEval))

	// Show fitness weights if custom
	var weightsLine string
	if s.FitnessWeights != nil {
		weightsLine = lipgloss.NewStyle().Foreground(colorChampion).
			Render("Custom Fitness Weights")
	}

	var sections []string
	sections = append(sections, title, statusLine)

	if weightsLine != "" {
		sections = append(sections, weightsLine)
	}

	// Training progress (live SSE)
	if m.lastProgress != nil && m.selectedStable.Status == "training" {
		sections = append(sections, "", m.renderTrainingProgress(t))
	}

	// Fitness chart (text-based sparkline from generation history)
	if len(m.generations) > 0 {
		sections = append(sections, "", m.renderFitnessChart(t))
	}

	// Champion card
	if m.champion != nil {
		sections = append(sections, "", m.renderChampionCard(t))
	}

	// Timing info
	if s.StartedAt > 0 {
		started := time.UnixMilli(s.StartedAt).Format("2006-01-02 15:04")
		timingLine := lipgloss.NewStyle().Foreground(t.TextDim).
			Render("Started: " + started)
		if s.CompletedAt != nil && *s.CompletedAt > 0 {
			completed := time.UnixMilli(*s.CompletedAt).Format("2006-01-02 15:04")
			elapsed := time.Duration(*s.CompletedAt-s.StartedAt) * time.Millisecond
			timingLine += "  Completed: " + completed + "  (" + formatDuration(elapsed) + ")"
		}
		sections = append(sections, "", timingLine)
	}

	errStr := m.renderError(t)
	if errStr != "" {
		sections = append(sections, "", errStr)
	}

	sections = append(sections, "", lipgloss.NewStyle().
		Foreground(t.TextMuted).Italic(true).
		Render(m.Hints()))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		strings.Join(sections, "\n"))
}

// renderTrainingProgress shows live training data.
func (m *Model) renderTrainingProgress(t *theme.Theme) string {
	p := m.lastProgress

	genInfo := lipgloss.NewStyle().Foreground(colorTraining).Bold(true).
		Render(fmt.Sprintf("Generation %d / %d", p.Generation, m.selectedStable.MaxGenerations))

	bar := renderProgressBar(p.Generation, m.selectedStable.MaxGenerations, 30, t)

	fitnessLine := lipgloss.NewStyle().Foreground(colorFitness).
		Render(fmt.Sprintf("Best: %.2f  Avg: %.2f  Worst: %.2f",
			p.BestFitness, p.AvgFitness, p.WorstFitness))

	return genInfo + "\n" + bar + "\n" + fitnessLine
}

// renderProgressBar renders a text progress bar.
func renderProgressBar(current, total, width int, t *theme.Theme) string {
	if total == 0 {
		total = 1
	}
	filled := (current * width) / total
	if filled > width {
		filled = width
	}
	empty := width - filled

	bar := lipgloss.NewStyle().Foreground(colorTraining).
		Render(strings.Repeat("=", filled))
	bar += lipgloss.NewStyle().Foreground(t.Border).
		Render(strings.Repeat("-", empty))

	pct := (current * 100) / total
	return fmt.Sprintf("[%s] %d%%", bar, pct)
}

// renderFitnessChart renders a simple text sparkline of best fitness over generations.
func (m *Model) renderFitnessChart(t *theme.Theme) string {
	label := lipgloss.NewStyle().Foreground(t.TextDim).Bold(true).
		Render("Fitness History")

	blocks := []string{"_", ".", "-", "~", "=", "+", "#", "@"}

	// Find min/max for normalization
	minF, maxF := m.generations[0].BestFitness, m.generations[0].BestFitness
	for _, g := range m.generations {
		if g.BestFitness < minF {
			minF = g.BestFitness
		}
		if g.BestFitness > maxF {
			maxF = g.BestFitness
		}
	}

	rangeF := maxF - minF
	if rangeF < 0.01 {
		rangeF = 1.0
	}

	// Take last N generations that fit in width
	maxWidth := 50
	gens := m.generations
	if len(gens) > maxWidth {
		gens = gens[len(gens)-maxWidth:]
	}

	var sparkline strings.Builder
	for _, g := range gens {
		normalized := (g.BestFitness - minF) / rangeF
		idx := int(normalized * float64(len(blocks)-1))
		if idx >= len(blocks) {
			idx = len(blocks) - 1
		}
		if idx < 0 {
			idx = 0
		}
		sparkline.WriteString(blocks[idx])
	}

	chart := lipgloss.NewStyle().Foreground(colorFitness).
		Render(sparkline.String())

	rangeInfo := lipgloss.NewStyle().Foreground(t.TextDim).
		Render(fmt.Sprintf("  [%.1f - %.1f]", minF, maxF))

	return label + "\n" + chart + rangeInfo
}

// renderChampionCard shows champion info.
func (m *Model) renderChampionCard(t *theme.Theme) string {
	c := m.champion

	title := lipgloss.NewStyle().Foreground(colorChampion).Bold(true).
		Render("Champion")

	fitness := lipgloss.NewStyle().Foreground(colorFitness).Bold(true).
		Render(fmt.Sprintf("Fitness: %.2f", c.Fitness))

	gen := lipgloss.NewStyle().Foreground(t.TextDim).
		Render(fmt.Sprintf("Generation: %d", c.Generation))

	record := lipgloss.NewStyle().Foreground(t.Text).
		Render(fmt.Sprintf("W:%d  L:%d  D:%d", c.Wins, c.Losses, c.Draws))

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorChampion).
		Padding(0, 1).
		Width(35)

	return cardStyle.Render(title + "\n" + fitness + "  " + gen + "\n" + record)
}

// viewDuel renders the champion duel using the snake_duel renderer.
func (m *Model) viewDuel() string {
	t := m.ctx.Theme

	header := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).
		Render("Champion Duel")

	stableInfo := lipgloss.NewStyle().Foreground(colorChampion).
		Render("Stable: " + truncateID(m.selectedStable.StableID))

	sep := "  "

	var statusStr string
	switch m.duelState.Status {
	case "finished":
		statusStr = m.renderDuelResult(t)
	case "":
		statusStr = lipgloss.NewStyle().Foreground(t.TextDim).Italic(true).
			Render("Starting duel...")
	default:
		statusStr = ""
	}

	grid := snake_duel.RenderGrid(m.duelState)

	var scoreStr string
	if m.duelState.Status != "" {
		p1 := lipgloss.NewStyle().Foreground(lipgloss.Color("#60a5fa")).Bold(true).
			Render(fmt.Sprintf("Champion:%d", m.duelState.Snake1.Score))
		p2 := lipgloss.NewStyle().Foreground(lipgloss.Color("#f87171")).Bold(true).
			Render(fmt.Sprintf("AI:%d", m.duelState.Snake2.Score))
		tick := lipgloss.NewStyle().Foreground(t.TextMuted).
			Render(fmt.Sprintf("T%d", m.duelState.Tick))
		scoreStr = p1 + sep + p2 + sep + tick
	}

	hints := lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).
		Render("esc:back to detail")
	if m.duelState.Status == "finished" {
		hints = lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).
			Render("n:new duel  esc:back to detail")
	}

	parts := header + sep + stableInfo + "\n"
	if scoreStr != "" {
		parts += scoreStr + "\n"
	}
	parts += grid + "\n"
	if statusStr != "" {
		parts += statusStr + "\n"
	}
	parts += hints

	return parts
}

// renderDuelResult shows the duel outcome.
func (m *Model) renderDuelResult(t *theme.Theme) string {
	switch m.duelState.Winner {
	case "player1":
		return lipgloss.NewStyle().Foreground(colorCompleted).Bold(true).
			Render("Champion Wins!")
	case "player2":
		return lipgloss.NewStyle().Foreground(colorHalted).Bold(true).
			Render("AI Wins!")
	case "draw":
		return lipgloss.NewStyle().Foreground(colorTraining).Bold(true).
			Render("Draw!")
	default:
		return lipgloss.NewStyle().Foreground(t.TextDim).Render("Game Over")
	}
}

// renderBudgetBar shows the tuning cost budget usage.
func (m *Model) renderBudgetBar(t *theme.Theme) string {
	cost := m.computeTuningCost()
	budget := 100.0
	used := int(cost * 30 / budget)
	if used > 30 {
		used = 30
	}
	remaining := 30 - used

	barColor := colorTraining
	if cost > budget {
		barColor = colorHalted
	}

	bar := lipgloss.NewStyle().Foreground(barColor).
		Render(strings.Repeat("=", used))
	bar += lipgloss.NewStyle().Foreground(t.Border).
		Render(strings.Repeat("-", remaining))

	label := lipgloss.NewStyle().Foreground(t.TextDim).
		Render(fmt.Sprintf("Budget: [%s] %.0f/%.0f pts", bar, cost, budget))

	if cost > budget {
		label += lipgloss.NewStyle().Foreground(colorHalted).Bold(true).
			Render("  OVER BUDGET!")
	}

	return label
}

// renderWeightsSection shows the advanced weights editor.
func (m *Model) renderWeightsSection(t *theme.Theme) string {
	title := lipgloss.NewStyle().Foreground(t.TextDim).Bold(true).
		Render("Advanced Weights (w to hide)")

	var fields []string
	for i, name := range m.formWeightNames {
		focused := m.formShowWeights && i == m.formWeightFocus
		indicator := "  "
		if focused {
			indicator = "> "
		}
		labelStyle := lipgloss.NewStyle().Foreground(t.TextDim).Width(16)
		valueStyle := lipgloss.NewStyle().Foreground(t.Text).Bold(true)
		if focused {
			labelStyle = labelStyle.Foreground(t.Primary)
			valueStyle = valueStyle.Foreground(t.Primary)
		}
		fields = append(fields, indicator+labelStyle.Render(name+":")+
			" "+valueStyle.Render(fmt.Sprintf("%.1f", m.formWeights[i])))
	}

	return title + "\n" + strings.Join(fields, "\n")
}

// computeTuningCost calculates the tuning cost for current form weights.
func (m *Model) computeTuningCost() float64 {
	defaults := [7]float64{0.1, 50.0, 200.0, 50.0, 100.0, 0.5, -0.2}
	bounds := [7][2]float64{
		{0.0, 1.0}, {0.0, 200.0}, {0.0, 500.0}, {0.0, 200.0},
		{0.0, 300.0}, {0.0, 5.0}, {-2.0, 0.0},
	}
	impacts := [7]float64{1.0, 2.0, 3.0, 1.5, 2.5, 1.0, 0.5}

	var total float64
	for i := 0; i < 7; i++ {
		rangeV := bounds[i][1] - bounds[i][0]
		if rangeV == 0 {
			continue
		}
		deviation := m.formWeights[i] - defaults[i]
		if deviation < 0 {
			deviation = -deviation
		}
		total += deviation / rangeV * impacts[i] * 10.0
	}
	return total
}

// viewHeroes renders the heroes list.
func (m *Model) viewHeroes() string {
	t := m.ctx.Theme

	title := lipgloss.NewStyle().
		Foreground(t.Primary).Bold(true).
		Render("Heroes")

	subtitle := lipgloss.NewStyle().
		Foreground(t.TextDim).
		Render("Promoted champions for permanent competition")

	var content string
	if len(m.heroes) == 0 {
		content = lipgloss.NewStyle().
			Foreground(t.TextMuted).Italic(true).
			Render("No heroes yet. Promote a champion from a completed stable.")
	} else {
		headerStyle := lipgloss.NewStyle().Foreground(t.TextDim).Bold(true)
		header := headerStyle.Render(fmt.Sprintf(
			"  %-16s %8s %6s %4s %4s %4s",
			"Name", "Fitness", "Gen", "W", "L", "D"))

		var rows []string
		rows = append(rows, header)
		for i, h := range m.heroes {
			selected := i == m.heroIndex
			style := lipgloss.NewStyle().Foreground(t.Text)
			indicator := " "
			if selected {
				style = style.Foreground(t.Primary).Bold(true)
				indicator = ">"
			}
			name := h.Name
			if len(name) > 16 {
				name = name[:14] + ".."
			}
			row := fmt.Sprintf("%s %-16s %8.1f %6d %4d %4d %4d",
				indicator, name, h.Fitness, h.Generation, h.Wins, h.Losses, h.Draws)
			rows = append(rows, style.Render(row))
		}
		content = strings.Join(rows, "\n")
	}

	errStr := m.renderError(t)

	hints := lipgloss.NewStyle().
		Foreground(t.TextMuted).Italic(true).
		Render("j/k:navigate  Enter:view  r:refresh  esc:back to stables")

	parts := title + "\n" + subtitle + "\n\n" + content
	if errStr != "" {
		parts += "\n\n" + errStr
	}
	parts += "\n\n" + hints

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, parts)
}

// viewHeroDetail renders a single hero's details.
func (m *Model) viewHeroDetail() string {
	t := m.ctx.Theme

	if m.selectedHero == nil {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().Foreground(t.TextDim).Italic(true).Render("Loading hero..."))
	}

	h := m.selectedHero

	title := lipgloss.NewStyle().Foreground(colorChampion).Bold(true).
		Render("Hero: " + h.Name)

	fitness := lipgloss.NewStyle().Foreground(colorFitness).Bold(true).
		Render(fmt.Sprintf("Fitness: %.2f", h.Fitness))

	gen := lipgloss.NewStyle().Foreground(t.TextDim).
		Render(fmt.Sprintf("Generation: %d  Origin: %s", h.Generation, truncateID(h.OriginStableID)))

	record := lipgloss.NewStyle().Foreground(t.Text).
		Render(fmt.Sprintf("W:%d  L:%d  D:%d", h.Wins, h.Losses, h.Draws))

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorChampion).
		Padding(0, 1).
		Width(40)

	card := cardStyle.Render(title + "\n" + fitness + "  " + gen + "\n" + record)

	hints := lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).
		Render("d:duel vs AI  esc:back to heroes")

	errStr := m.renderError(t)
	parts := card
	if errStr != "" {
		parts += "\n\n" + errStr
	}
	parts += "\n\n" + hints

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, parts)
}

// viewPromote renders the hero promotion form.
func (m *Model) viewPromote() string {
	t := m.ctx.Theme

	title := lipgloss.NewStyle().Foreground(colorChampion).Bold(true).
		Render("Promote Champion to Hero")

	subtitle := lipgloss.NewStyle().Foreground(t.TextDim).
		Render("Stable: " + truncateID(m.selectedStable.StableID))

	nameLabel := lipgloss.NewStyle().Foreground(t.Primary).
		Render("Hero Name: ")

	nameVal := m.promoteName
	if nameVal == "" {
		nameVal = lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).Render("type a name...")
	} else {
		nameVal = lipgloss.NewStyle().Foreground(t.Text).Bold(true).Render(nameVal)
	}

	cursor := lipgloss.NewStyle().Foreground(t.Primary).Render("_")

	errStr := m.renderError(t)

	hints := lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).
		Render("Enter:confirm  esc:cancel")

	parts := title + "\n" + subtitle + "\n\n" + nameLabel + nameVal + cursor
	if errStr != "" {
		parts += "\n\n" + errStr
	}
	parts += "\n\n" + hints

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, parts)
}

// viewHeroDuel renders a hero duel using the same duel renderer.
func (m *Model) viewHeroDuel() string {
	t := m.ctx.Theme

	heroName := ""
	if m.selectedHero != nil {
		heroName = m.selectedHero.Name
	}

	header := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).
		Render("Hero Duel")

	heroInfo := lipgloss.NewStyle().Foreground(colorChampion).
		Render("Hero: " + heroName)

	sep := "  "

	var statusStr string
	switch m.duelState.Status {
	case "finished":
		statusStr = m.renderDuelResult(t)
	case "":
		statusStr = lipgloss.NewStyle().Foreground(t.TextDim).Italic(true).
			Render("Starting duel...")
	default:
		statusStr = ""
	}

	grid := snake_duel.RenderGrid(m.duelState)

	var scoreStr string
	if m.duelState.Status != "" {
		p1 := lipgloss.NewStyle().Foreground(lipgloss.Color("#60a5fa")).Bold(true).
			Render(fmt.Sprintf("Hero:%d", m.duelState.Snake1.Score))
		p2 := lipgloss.NewStyle().Foreground(lipgloss.Color("#f87171")).Bold(true).
			Render(fmt.Sprintf("AI:%d", m.duelState.Snake2.Score))
		tick := lipgloss.NewStyle().Foreground(t.TextMuted).
			Render(fmt.Sprintf("T%d", m.duelState.Tick))
		scoreStr = p1 + sep + p2 + sep + tick
	}

	hints := lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).
		Render("esc:back to hero")
	if m.duelState.Status == "finished" {
		hints = lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).
			Render("n:new duel  esc:back to hero")
	}

	parts := header + sep + heroInfo + "\n"
	if scoreStr != "" {
		parts += scoreStr + "\n"
	}
	parts += grid + "\n"
	if statusStr != "" {
		parts += statusStr + "\n"
	}
	parts += hints

	return parts
}

// renderError renders an error message if present.
func (m *Model) renderError(t *theme.Theme) string {
	if m.err == nil {
		return ""
	}
	return lipgloss.NewStyle().Foreground(colorHalted).
		Render("Error: " + m.err.Error())
}

// truncateID shortens a stable ID for display.
func truncateID(id string) string {
	if len(id) > 14 {
		return id[:6] + ".." + id[len(id)-6:]
	}
	return id
}

// formatDuration formats a duration nicely.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm%.0fs", d.Minutes(), d.Seconds()-d.Minutes()*60)
	}
	return fmt.Sprintf("%.0fh%.0fm", d.Hours(), d.Minutes()-d.Hours()*60)
}
