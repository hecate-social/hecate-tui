package commands

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CartwheelCmd handles all /cartwheel subcommands for bounded context management.
type CartwheelCmd struct{}

func (c *CartwheelCmd) Name() string        { return "cartwheel" }
func (c *CartwheelCmd) Aliases() []string   { return []string{"cw", "alc", "lifecycle", "lc"} }
func (c *CartwheelCmd) Description() string { return "Manage bounded contexts (Cartwheels)" }

func (c *CartwheelCmd) Execute(args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return c.showUsage(ctx)
	}

	sub := strings.ToLower(args[0])

	// "init" creates a new cartwheel
	if sub == "init" {
		return c.initCartwheel(args[1:], ctx)
	}

	// Everything else requires a cartwheel ID as first arg
	if !strings.HasPrefix(sub, "prj-") {
		return c.showUsage(ctx)
	}

	cartwheelID := sub
	if len(args) < 2 {
		// Just a cartwheel ID -> show status card
		return c.showCartwheel(cartwheelID, ctx)
	}

	action := strings.ToLower(args[1])
	rest := args[2:]

	switch action {
	case "discovery":
		return c.phaseAction(cartwheelID, "discovery", rest, ctx)
	case "finding":
		return c.recordFinding(cartwheelID, rest, ctx)
	case "term":
		return c.defineTerm(cartwheelID, rest, ctx)
	case "transition":
		return c.transition(cartwheelID, rest, ctx)
	case "arch":
		return c.phaseAction(cartwheelID, "architecture", rest, ctx)
	case "dossier":
		return c.defineDossier(cartwheelID, rest, ctx)
	case "spoke":
		return c.inventorySpoke(cartwheelID, rest, ctx)
	case "plan":
		return c.draftPlan(cartwheelID, rest, ctx)
	case "approve":
		return c.approvePlan(cartwheelID, rest, ctx)
	case "test":
		return c.phaseAction(cartwheelID, "testing", rest, ctx)
	case "skeleton":
		return c.createSkeleton(cartwheelID, ctx)
	case "implement":
		return c.implementSpoke(cartwheelID, rest, ctx)
	case "verify":
		return c.verifyBuild(cartwheelID, rest, ctx)
	case "deploy":
		return c.deployAction(cartwheelID, rest, ctx)
	case "incident":
		return c.reportIncident(cartwheelID, rest, ctx)
	case "resolve":
		return c.resolveIncident(cartwheelID, rest, ctx)
	case "complete":
		return c.completePhase(cartwheelID, ctx)
	default:
		return c.showUsage(ctx)
	}
}

func (c *CartwheelCmd) showUsage(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles
		t := ctx.Theme
		var b strings.Builder

		// Title
		b.WriteString(s.CardTitle.Render("Cartwheel - Bounded Context Lifecycle Commands"))
		b.WriteString("\n\n")

		// Intro
		b.WriteString(s.Subtle.Render("Manage bounded contexts through four phases: Discovery & Analysis -> Architecture & Planning -> Testing & Implementation -> Deployment & Operations"))
		b.WriteString("\n\n")

		// Helper for table rows
		cmdStyle := lipgloss.NewStyle().Foreground(t.Secondary)
		descStyle := lipgloss.NewStyle().Foreground(t.Text)
		sectionStyle := s.Bold
		subtitleStyle := s.Subtle

		row := func(cmd, desc string) string {
			// Pad command to 30 chars for alignment
			padded := cmd
			for len(padded) < 30 {
				padded += " "
			}
			return cmdStyle.Render(padded) + descStyle.Render(desc) + "\n"
		}

		section := func(title, subtitle string) string {
			result := sectionStyle.Render(title)
			if subtitle != "" {
				result += "\n" + subtitleStyle.Render(subtitle)
			}
			return result + "\n"
		}

		// Getting Started
		b.WriteString(section("Getting Started", ""))
		b.WriteString(row("/cartwheel init <name>", "Create a new cartwheel"))
		b.WriteString(row("/cartwheel <id>", "Show cartwheel status"))
		b.WriteString(row("/cartwheel <id> transition X", "Move to phase (arch, test, deploy)"))
		b.WriteString(row("/cartwheel <id> complete", "Complete current phase"))
		b.WriteString("\n")

		// Phase 1
		b.WriteString(section("Phase 1: Discovery & Analysis", "Analyze requirements, gather findings, build vocabulary"))
		b.WriteString(row("/cartwheel <id> discovery start", "Begin discovery"))
		b.WriteString(row("/cartwheel <id> finding <title>", "Record insight or requirement"))
		b.WriteString(row("/cartwheel <id> term <t> <def>", "Define domain term"))
		b.WriteString("\n")

		// Phase 2
		b.WriteString(section("Phase 2: Architecture & Planning", "Design dossiers (aggregates) and spokes (operations)"))
		b.WriteString(row("/cartwheel <id> arch start", "Begin architecture"))
		b.WriteString(row("/cartwheel <id> dossier <name>", "Define aggregate/entity"))
		b.WriteString(row("/cartwheel <id> spoke <n> <t> <did>", "Add operation (cmd/qry/evt)"))
		b.WriteString(row("/cartwheel <id> plan", "Draft implementation plan"))
		b.WriteString(row("/cartwheel <id> approve", "Approve plan"))
		b.WriteString("\n")

		// Phase 3
		b.WriteString(section("Phase 3: Testing & Implementation", "Implement features and verify quality"))
		b.WriteString(row("/cartwheel <id> test start", "Begin testing"))
		b.WriteString(row("/cartwheel <id> skeleton", "Generate code skeleton"))
		b.WriteString(row("/cartwheel <id> implement <sid>", "Mark spoke implemented"))
		b.WriteString(row("/cartwheel <id> verify pass|fail", "Record build result"))
		b.WriteString("\n")

		// Phase 4
		b.WriteString(section("Phase 4: Deployment & Operations", "Release to production, monitor, handle incidents"))
		b.WriteString(row("/cartwheel <id> deploy start", "Begin deployment"))
		b.WriteString(row("/cartwheel <id> deploy record <e> <v>", "Record release"))
		b.WriteString(row("/cartwheel <id> incident <desc>", "Report incident"))
		b.WriteString(row("/cartwheel <id> resolve <iid> <res>", "Resolve incident"))

		return InjectSystemMsg{Content: b.String()}
	}
}

func (c *CartwheelCmd) initCartwheel(args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /cartwheel init <name> [description]")}
		}
	}

	name := args[0]
	desc := ""
	if len(args) > 1 {
		desc = strings.Join(args[1:], " ")
	}

	return func() tea.Msg {
		s := ctx.Styles
		body := map[string]interface{}{"name": name}
		if desc != "" {
			body["description"] = desc
		}

		err := ctx.Client.ALCCommand("/alc/projects/initiate", body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to initiate cartwheel: " + err.Error())}
		}

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("Cartwheel Initiated"))
		b.WriteString("\n\n")
		b.WriteString(s.CardLabel.Render("Name: "))
		b.WriteString(s.CardValue.Render(name))
		if desc != "" {
			b.WriteString("\n")
			b.WriteString(s.CardLabel.Render("Description: "))
			b.WriteString(s.CardValue.Render(desc))
		}
		b.WriteString("\n\n")
		b.WriteString(s.Subtle.Render("  Use /cartwheel to browse cartwheels"))

		return InjectSystemMsg{Content: b.String()}
	}
}

func (c *CartwheelCmd) showCartwheel(cartwheelID string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles
		cw, err := ctx.Client.GetCartwheel(cartwheelID)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to get cartwheel: " + err.Error())}
		}

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("Cartwheel: " + cw.Name))
		b.WriteString("\n\n")

		b.WriteString(s.CardLabel.Render("ID: "))
		b.WriteString(s.CardValue.Render(cw.CartwheelID))
		b.WriteString("\n")

		if cw.Description != "" {
			b.WriteString(s.CardLabel.Render("Description: "))
			b.WriteString(s.CardValue.Render(cw.Description))
			b.WriteString("\n")
		}

		b.WriteString(s.CardLabel.Render("Phase: "))
		b.WriteString(s.CardValue.Render(formatCartwheelPhase(cw.CurrentPhase)))
		b.WriteString("\n")

		b.WriteString(s.CardLabel.Render("Initiated: "))
		b.WriteString(s.Subtle.Render(formatTimestamp(cw.InitiatedAt)))
		b.WriteString("\n\n")

		// Phase-specific counters
		b.WriteString(s.Bold.Render("  Counters"))
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("Findings: "))
		b.WriteString(s.CardValue.Render(fmt.Sprintf("%d", cw.FindingCount)))
		b.WriteString("  ")
		b.WriteString(s.CardLabel.Render("Terms: "))
		b.WriteString(s.CardValue.Render(fmt.Sprintf("%d", cw.TermCount)))
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("Dossiers: "))
		b.WriteString(s.CardValue.Render(fmt.Sprintf("%d", cw.DossierCount)))
		b.WriteString("  ")
		b.WriteString(s.CardLabel.Render("Spokes: "))
		b.WriteString(s.CardValue.Render(fmt.Sprintf("%d", cw.SpokeCount)))
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("Implemented: "))
		b.WriteString(s.CardValue.Render(fmt.Sprintf("%d/%d", cw.ImplementedSpokeCount, cw.SpokeCount)))
		b.WriteString("  ")
		b.WriteString(s.CardLabel.Render("Build: "))
		if cw.BuildVerified {
			b.WriteString(s.StatusOK.Render("verified"))
		} else {
			b.WriteString(s.Subtle.Render("pending"))
		}
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("Deployments: "))
		b.WriteString(s.CardValue.Render(fmt.Sprintf("%d", cw.DeploymentCount)))
		b.WriteString("  ")
		b.WriteString(s.CardLabel.Render("Incidents: "))
		if cw.ActiveIncidents > 0 {
			b.WriteString(s.StatusError.Render(fmt.Sprintf("%d active", cw.ActiveIncidents)))
		} else {
			b.WriteString(s.StatusOK.Render("none"))
		}
		b.WriteString("\n")

		return InjectSystemMsg{Content: b.String()}
	}
}

func (c *CartwheelCmd) phaseAction(cartwheelID, phase string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 || strings.ToLower(args[0]) != "start" {
		return func() tea.Msg {
			return InjectSystemMsg{
				Content: ctx.Styles.Error.Render(fmt.Sprintf("Usage: /cartwheel %s %s start", cartwheelID, phase)),
			}
		}
	}

	return func() tea.Msg {
		s := ctx.Styles
		path := fmt.Sprintf("/alc/projects/%s/%s/start", cartwheelID, phase)
		err := ctx.Client.ALCCommand(path, nil)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to start " + phase + ": " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Started " + phase + " phase for " + cartwheelID)}
	}
}

func (c *CartwheelCmd) recordFinding(cartwheelID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /cartwheel <id> finding <title> [content]")}
		}
	}

	title := args[0]
	content := ""
	if len(args) > 1 {
		content = strings.Join(args[1:], " ")
	}

	return func() tea.Msg {
		s := ctx.Styles
		body := map[string]interface{}{"title": title}
		if content != "" {
			body["content"] = content
		}

		path := fmt.Sprintf("/alc/projects/%s/discovery/findings", cartwheelID)
		err := ctx.Client.ALCCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to record finding: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Recorded finding: " + title)}
	}
}

func (c *CartwheelCmd) defineTerm(cartwheelID string, args []string, ctx *Context) tea.Cmd {
	if len(args) < 2 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /cartwheel <id> term <term> <definition>")}
		}
	}

	term := args[0]
	definition := strings.Join(args[1:], " ")

	return func() tea.Msg {
		s := ctx.Styles
		body := map[string]interface{}{"term": term, "definition": definition}

		path := fmt.Sprintf("/alc/projects/%s/discovery/terms", cartwheelID)
		err := ctx.Client.ALCCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to define term: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Defined term: " + term)}
	}
}

func (c *CartwheelCmd) transition(cartwheelID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{
				Content: ctx.Styles.Error.Render("Usage: /cartwheel <id> transition <target_phase>"),
			}
		}
	}

	targetPhase := args[0]

	return func() tea.Msg {
		s := ctx.Styles
		body := map[string]interface{}{"target_phase": targetPhase}

		path := fmt.Sprintf("/alc/projects/%s/transition", cartwheelID)
		err := ctx.Client.ALCCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to transition: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Transitioned " + cartwheelID + " to " + targetPhase)}
	}
}

func (c *CartwheelCmd) defineDossier(cartwheelID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /cartwheel <id> dossier <name> [description]")}
		}
	}

	name := args[0]
	desc := ""
	if len(args) > 1 {
		desc = strings.Join(args[1:], " ")
	}

	return func() tea.Msg {
		s := ctx.Styles
		body := map[string]interface{}{"dossier_name": name}
		if desc != "" {
			body["description"] = desc
		}

		path := fmt.Sprintf("/alc/projects/%s/architecture/dossiers", cartwheelID)
		err := ctx.Client.ALCCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to define dossier: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Defined dossier: " + name)}
	}
}

func (c *CartwheelCmd) inventorySpoke(cartwheelID string, args []string, ctx *Context) tea.Cmd {
	if len(args) < 3 {
		return func() tea.Msg {
			return InjectSystemMsg{
				Content: ctx.Styles.Error.Render("Usage: /cartwheel <id> spoke <name> <type> <dossier_id> [description]"),
			}
		}
	}

	name := args[0]
	spokeType := args[1]
	dossierID := args[2]
	desc := ""
	if len(args) > 3 {
		desc = strings.Join(args[3:], " ")
	}

	return func() tea.Msg {
		s := ctx.Styles
		body := map[string]interface{}{
			"spoke_name": name,
			"spoke_type": spokeType,
			"dossier_id": dossierID,
		}
		if desc != "" {
			body["description"] = desc
		}

		path := fmt.Sprintf("/alc/projects/%s/architecture/spokes", cartwheelID)
		err := ctx.Client.ALCCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to inventory spoke: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Inventoried spoke: " + name)}
	}
}

func (c *CartwheelCmd) draftPlan(cartwheelID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /cartwheel <id> plan <title> [description]")}
		}
	}

	title := args[0]
	desc := ""
	if len(args) > 1 {
		desc = strings.Join(args[1:], " ")
	}

	return func() tea.Msg {
		s := ctx.Styles
		body := map[string]interface{}{"title": title}
		if desc != "" {
			body["description"] = desc
		}

		path := fmt.Sprintf("/alc/projects/%s/architecture/plans", cartwheelID)
		err := ctx.Client.ALCCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to draft plan: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Drafted plan: " + title)}
	}
}

func (c *CartwheelCmd) approvePlan(cartwheelID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /cartwheel <id> approve <plan_id>")}
		}
	}

	planID := args[0]

	return func() tea.Msg {
		s := ctx.Styles
		body := map[string]interface{}{"plan_id": planID}

		path := fmt.Sprintf("/alc/projects/%s/architecture/plans/approve", cartwheelID)
		err := ctx.Client.ALCCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to approve plan: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Approved plan: " + planID)}
	}
}

func (c *CartwheelCmd) createSkeleton(cartwheelID string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles
		path := fmt.Sprintf("/alc/projects/%s/testing/skeleton", cartwheelID)
		err := ctx.Client.ALCCommand(path, nil)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to create skeleton: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Skeleton created for " + cartwheelID)}
	}
}

func (c *CartwheelCmd) implementSpoke(cartwheelID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /cartwheel <id> implement <spoke_id> [notes]")}
		}
	}

	spokeID := args[0]
	notes := ""
	if len(args) > 1 {
		notes = strings.Join(args[1:], " ")
	}

	return func() tea.Msg {
		s := ctx.Styles
		body := map[string]interface{}{"spoke_id": spokeID}
		if notes != "" {
			body["implementation_notes"] = notes
		}

		path := fmt.Sprintf("/alc/projects/%s/testing/implementations", cartwheelID)
		err := ctx.Client.ALCCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to implement spoke: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Implemented spoke: " + spokeID)}
	}
}

func (c *CartwheelCmd) verifyBuild(cartwheelID string, args []string, ctx *Context) tea.Cmd {
	result := "pass"
	if len(args) > 0 {
		result = strings.ToLower(args[0])
	}
	notes := ""
	if len(args) > 1 {
		notes = strings.Join(args[1:], " ")
	}

	return func() tea.Msg {
		s := ctx.Styles
		body := map[string]interface{}{"result": result}
		if notes != "" {
			body["notes"] = notes
		}

		path := fmt.Sprintf("/alc/projects/%s/testing/builds", cartwheelID)
		err := ctx.Client.ALCCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to verify build: " + err.Error())}
		}

		label := s.StatusOK.Render("PASS")
		if result == "fail" {
			label = s.StatusError.Render("FAIL")
		}
		return InjectSystemMsg{Content: fmt.Sprintf("Build verification: %s for %s", label, cartwheelID)}
	}
}

func (c *CartwheelCmd) deployAction(cartwheelID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{
				Content: ctx.Styles.Error.Render("Usage: /cartwheel <id> deploy start | /cartwheel <id> deploy record <env> <version>"),
			}
		}
	}

	sub := strings.ToLower(args[0])

	if sub == "start" {
		return func() tea.Msg {
			s := ctx.Styles
			path := fmt.Sprintf("/alc/projects/%s/deployment/start", cartwheelID)
			err := ctx.Client.ALCCommand(path, nil)
			if err != nil {
				return InjectSystemMsg{Content: s.Error.Render("Failed to start deployment phase: " + err.Error())}
			}
			return InjectSystemMsg{Content: s.StatusOK.Render("Started deployment phase for " + cartwheelID)}
		}
	}

	if sub == "record" {
		if len(args) < 3 {
			return func() tea.Msg {
				return InjectSystemMsg{
					Content: ctx.Styles.Error.Render("Usage: /cartwheel <id> deploy record <environment> <version> [notes]"),
				}
			}
		}

		env := args[1]
		version := args[2]
		notes := ""
		if len(args) > 3 {
			notes = strings.Join(args[3:], " ")
		}

		return func() tea.Msg {
			s := ctx.Styles
			body := map[string]interface{}{
				"environment": env,
				"version":     version,
			}
			if notes != "" {
				body["notes"] = notes
			}

			path := fmt.Sprintf("/alc/projects/%s/deployment/deployments", cartwheelID)
			err := ctx.Client.ALCCommand(path, body)
			if err != nil {
				return InjectSystemMsg{Content: s.Error.Render("Failed to record deployment: " + err.Error())}
			}
			return InjectSystemMsg{
				Content: s.StatusOK.Render(fmt.Sprintf("Recorded deployment: %s v%s to %s", cartwheelID, version, env)),
			}
		}
	}

	return func() tea.Msg {
		return InjectSystemMsg{
			Content: ctx.Styles.Error.Render("Unknown deploy subcommand: " + sub + ". Use 'start' or 'record'."),
		}
	}
}

func (c *CartwheelCmd) reportIncident(cartwheelID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /cartwheel <id> incident <description>")}
		}
	}

	description := strings.Join(args, " ")

	return func() tea.Msg {
		s := ctx.Styles
		body := map[string]interface{}{"description": description}

		path := fmt.Sprintf("/alc/projects/%s/deployment/incidents", cartwheelID)
		err := ctx.Client.ALCCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to report incident: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusWarning.Render("Incident reported for " + cartwheelID)}
	}
}

func (c *CartwheelCmd) resolveIncident(cartwheelID string, args []string, ctx *Context) tea.Cmd {
	if len(args) < 2 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /cartwheel <id> resolve <incident_id> <resolution>")}
		}
	}

	incidentID := args[0]
	resolution := strings.Join(args[1:], " ")

	return func() tea.Msg {
		s := ctx.Styles
		body := map[string]interface{}{
			"incident_id": incidentID,
			"resolution":  resolution,
		}

		path := fmt.Sprintf("/alc/projects/%s/deployment/incidents/resolve", cartwheelID)
		err := ctx.Client.ALCCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to resolve incident: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Resolved incident: " + incidentID)}
	}
}

func (c *CartwheelCmd) completePhase(cartwheelID string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles
		path := fmt.Sprintf("/alc/projects/%s/complete", cartwheelID)
		err := ctx.Client.ALCCommand(path, nil)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to complete phase: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Phase completed for " + cartwheelID)}
	}
}

// formatCartwheelPhase returns a human-readable phase name.
func formatCartwheelPhase(phase string) string {
	switch strings.ToLower(phase) {
	case "discovery", "dna":
		return "Discovery & Analysis"
	case "architecture", "anp":
		return "Architecture & Planning"
	case "testing", "tni":
		return "Testing & Implementation"
	case "deployment", "dno":
		return "Deployment & Operations"
	case "initiated":
		return "Initiated"
	case "completed":
		return "Completed"
	default:
		return phase
	}
}

// formatTimestamp converts a Unix timestamp to a readable string.
func formatTimestamp(ts int64) string {
	if ts == 0 {
		return "-"
	}
	return time.Unix(ts, 0).Format("2006-01-02 15:04")
}
