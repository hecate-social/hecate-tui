package commands

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ALCCmd handles all /alc subcommands for project lifecycle management.
type ALCCmd struct{}

func (c *ALCCmd) Name() string        { return "alc" }
func (c *ALCCmd) Aliases() []string   { return []string{"lifecycle", "lc"} }
func (c *ALCCmd) Description() string { return "Project lifecycle management (/alc [subcommand])" }

func (c *ALCCmd) Execute(args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return c.showUsage(ctx)
	}

	sub := strings.ToLower(args[0])

	// "init" creates a new project
	if sub == "init" {
		return c.initProject(args[1:], ctx)
	}

	// Everything else requires a project ID as first arg
	if !strings.HasPrefix(sub, "prj-") {
		return c.showUsage(ctx)
	}

	projectID := sub
	if len(args) < 2 {
		// Just a project ID → show status card
		return c.showProject(projectID, ctx)
	}

	action := strings.ToLower(args[1])
	rest := args[2:]

	switch action {
	case "discovery":
		return c.phaseAction(projectID, "discovery", rest, ctx)
	case "finding":
		return c.recordFinding(projectID, rest, ctx)
	case "term":
		return c.defineTerm(projectID, rest, ctx)
	case "transition":
		return c.transition(projectID, rest, ctx)
	case "arch":
		return c.phaseAction(projectID, "architecture", rest, ctx)
	case "dossier":
		return c.defineDossier(projectID, rest, ctx)
	case "spoke":
		return c.inventorySpoke(projectID, rest, ctx)
	case "plan":
		return c.draftPlan(projectID, rest, ctx)
	case "approve":
		return c.approvePlan(projectID, rest, ctx)
	case "test":
		return c.phaseAction(projectID, "testing", rest, ctx)
	case "skeleton":
		return c.createSkeleton(projectID, ctx)
	case "implement":
		return c.implementSpoke(projectID, rest, ctx)
	case "verify":
		return c.verifyBuild(projectID, rest, ctx)
	case "deploy":
		return c.deployAction(projectID, rest, ctx)
	case "incident":
		return c.reportIncident(projectID, rest, ctx)
	case "resolve":
		return c.resolveIncident(projectID, rest, ctx)
	case "complete":
		return c.completePhase(projectID, ctx)
	default:
		return c.showUsage(ctx)
	}
}

func (c *ALCCmd) showUsage(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles
		t := ctx.Theme
		var b strings.Builder

		// Title
		b.WriteString(s.CardTitle.Render("ALC - Application Lifecycle Commands"))
		b.WriteString("\n\n")

		// Intro
		b.WriteString(s.Subtle.Render("Manage projects through four phases: Discovery & Analysis → Architecture & Planning → Testing & Implementation → Deployment & Operations"))
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
		b.WriteString(section("🚀 Getting Started", ""))
		b.WriteString(row("/alc init <name>", "Create a new project"))
		b.WriteString(row("/alc <id>", "Show project status"))
		b.WriteString(row("/alc <id> transition X", "Move to phase (arch, test, deploy)"))
		b.WriteString(row("/alc <id> complete", "Complete current phase"))
		b.WriteString("\n")

		// Phase 1
		b.WriteString(section("🔍 Phase 1: Discovery & Analysis", "Analyze requirements, gather findings, build vocabulary"))
		b.WriteString(row("/alc <id> discovery start", "Begin discovery"))
		b.WriteString(row("/alc <id> finding <title>", "Record insight or requirement"))
		b.WriteString(row("/alc <id> term <t> <def>", "Define domain term"))
		b.WriteString("\n")

		// Phase 2
		b.WriteString(section("🏗️ Phase 2: Architecture & Planning", "Design dossiers (aggregates) and spokes (operations)"))
		b.WriteString(row("/alc <id> arch start", "Begin architecture"))
		b.WriteString(row("/alc <id> dossier <name>", "Define aggregate/entity"))
		b.WriteString(row("/alc <id> spoke <n> <t> <did>", "Add operation (cmd/qry/evt)"))
		b.WriteString(row("/alc <id> plan", "Draft implementation plan"))
		b.WriteString(row("/alc <id> approve", "Approve plan"))
		b.WriteString("\n")

		// Phase 3
		b.WriteString(section("🧪 Phase 3: Testing & Implementation", "Implement features and verify quality"))
		b.WriteString(row("/alc <id> test start", "Begin testing"))
		b.WriteString(row("/alc <id> skeleton", "Generate code skeleton"))
		b.WriteString(row("/alc <id> implement <sid>", "Mark spoke implemented"))
		b.WriteString(row("/alc <id> verify pass|fail", "Record build result"))
		b.WriteString("\n")

		// Phase 4
		b.WriteString(section("📦 Phase 4: Deployment & Operations", "Release to production, monitor, handle incidents"))
		b.WriteString(row("/alc <id> deploy start", "Begin deployment"))
		b.WriteString(row("/alc <id> deploy record <e> <v>", "Record release"))
		b.WriteString(row("/alc <id> incident <desc>", "Report incident"))
		b.WriteString(row("/alc <id> resolve <iid> <res>", "Resolve incident"))

		return InjectSystemMsg{Content: b.String()}
	}
}

func (c *ALCCmd) initProject(args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /alc init <name> [description]")}
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
			return InjectSystemMsg{Content: s.Error.Render("Failed to initiate project: " + err.Error())}
		}

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("Project Initiated"))
		b.WriteString("\n\n")
		b.WriteString(s.CardLabel.Render("Name: "))
		b.WriteString(s.CardValue.Render(name))
		if desc != "" {
			b.WriteString("\n")
			b.WriteString(s.CardLabel.Render("Description: "))
			b.WriteString(s.CardValue.Render(desc))
		}
		b.WriteString("\n\n")
		b.WriteString(s.Subtle.Render("  Use /alc to browse projects"))

		return InjectSystemMsg{Content: b.String()}
	}
}

func (c *ALCCmd) showProject(projectID string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles
		project, err := ctx.Client.GetProject(projectID)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to get project: " + err.Error())}
		}

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("Project: " + project.Name))
		b.WriteString("\n\n")

		b.WriteString(s.CardLabel.Render("ID: "))
		b.WriteString(s.CardValue.Render(project.ProjectID))
		b.WriteString("\n")

		if project.Description != "" {
			b.WriteString(s.CardLabel.Render("Description: "))
			b.WriteString(s.CardValue.Render(project.Description))
			b.WriteString("\n")
		}

		b.WriteString(s.CardLabel.Render("Phase: "))
		b.WriteString(s.CardValue.Render(formatPhase(project.CurrentPhase)))
		b.WriteString("\n")

		b.WriteString(s.CardLabel.Render("Initiated: "))
		b.WriteString(s.Subtle.Render(formatTimestamp(project.InitiatedAt)))
		b.WriteString("\n\n")

		// Phase-specific counters
		b.WriteString(s.Bold.Render("  Counters"))
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("Findings: "))
		b.WriteString(s.CardValue.Render(fmt.Sprintf("%d", project.FindingCount)))
		b.WriteString("  ")
		b.WriteString(s.CardLabel.Render("Terms: "))
		b.WriteString(s.CardValue.Render(fmt.Sprintf("%d", project.TermCount)))
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("Dossiers: "))
		b.WriteString(s.CardValue.Render(fmt.Sprintf("%d", project.DossierCount)))
		b.WriteString("  ")
		b.WriteString(s.CardLabel.Render("Spokes: "))
		b.WriteString(s.CardValue.Render(fmt.Sprintf("%d", project.SpokeCount)))
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("Implemented: "))
		b.WriteString(s.CardValue.Render(fmt.Sprintf("%d/%d", project.ImplementedSpokeCount, project.SpokeCount)))
		b.WriteString("  ")
		b.WriteString(s.CardLabel.Render("Build: "))
		if project.BuildVerified {
			b.WriteString(s.StatusOK.Render("verified"))
		} else {
			b.WriteString(s.Subtle.Render("pending"))
		}
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("Deployments: "))
		b.WriteString(s.CardValue.Render(fmt.Sprintf("%d", project.DeploymentCount)))
		b.WriteString("  ")
		b.WriteString(s.CardLabel.Render("Incidents: "))
		if project.ActiveIncidents > 0 {
			b.WriteString(s.StatusError.Render(fmt.Sprintf("%d active", project.ActiveIncidents)))
		} else {
			b.WriteString(s.StatusOK.Render("none"))
		}
		b.WriteString("\n")

		return InjectSystemMsg{Content: b.String()}
	}
}

func (c *ALCCmd) phaseAction(projectID, phase string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 || strings.ToLower(args[0]) != "start" {
		return func() tea.Msg {
			return InjectSystemMsg{
				Content: ctx.Styles.Error.Render(fmt.Sprintf("Usage: /alc %s %s start", projectID, phase)),
			}
		}
	}

	return func() tea.Msg {
		s := ctx.Styles
		path := fmt.Sprintf("/alc/projects/%s/%s/start", projectID, phase)
		err := ctx.Client.ALCCommand(path, nil)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to start " + phase + ": " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Started " + phase + " phase for " + projectID)}
	}
}

func (c *ALCCmd) recordFinding(projectID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /alc <id> finding <title> [content]")}
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

		path := fmt.Sprintf("/alc/projects/%s/discovery/findings", projectID)
		err := ctx.Client.ALCCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to record finding: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Recorded finding: " + title)}
	}
}

func (c *ALCCmd) defineTerm(projectID string, args []string, ctx *Context) tea.Cmd {
	if len(args) < 2 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /alc <id> term <term> <definition>")}
		}
	}

	term := args[0]
	definition := strings.Join(args[1:], " ")

	return func() tea.Msg {
		s := ctx.Styles
		body := map[string]interface{}{"term": term, "definition": definition}

		path := fmt.Sprintf("/alc/projects/%s/discovery/terms", projectID)
		err := ctx.Client.ALCCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to define term: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Defined term: " + term)}
	}
}

func (c *ALCCmd) transition(projectID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{
				Content: ctx.Styles.Error.Render("Usage: /alc <id> transition <target_phase>"),
			}
		}
	}

	targetPhase := args[0]

	return func() tea.Msg {
		s := ctx.Styles
		body := map[string]interface{}{"target_phase": targetPhase}

		path := fmt.Sprintf("/alc/projects/%s/transition", projectID)
		err := ctx.Client.ALCCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to transition: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Transitioned " + projectID + " to " + targetPhase)}
	}
}

func (c *ALCCmd) defineDossier(projectID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /alc <id> dossier <name> [description]")}
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

		path := fmt.Sprintf("/alc/projects/%s/architecture/dossiers", projectID)
		err := ctx.Client.ALCCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to define dossier: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Defined dossier: " + name)}
	}
}

func (c *ALCCmd) inventorySpoke(projectID string, args []string, ctx *Context) tea.Cmd {
	if len(args) < 3 {
		return func() tea.Msg {
			return InjectSystemMsg{
				Content: ctx.Styles.Error.Render("Usage: /alc <id> spoke <name> <type> <dossier_id> [description]"),
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

		path := fmt.Sprintf("/alc/projects/%s/architecture/spokes", projectID)
		err := ctx.Client.ALCCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to inventory spoke: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Inventoried spoke: " + name)}
	}
}

func (c *ALCCmd) draftPlan(projectID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /alc <id> plan <title> [description]")}
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

		path := fmt.Sprintf("/alc/projects/%s/architecture/plans", projectID)
		err := ctx.Client.ALCCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to draft plan: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Drafted plan: " + title)}
	}
}

func (c *ALCCmd) approvePlan(projectID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /alc <id> approve <plan_id>")}
		}
	}

	planID := args[0]

	return func() tea.Msg {
		s := ctx.Styles
		body := map[string]interface{}{"plan_id": planID}

		path := fmt.Sprintf("/alc/projects/%s/architecture/plans/approve", projectID)
		err := ctx.Client.ALCCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to approve plan: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Approved plan: " + planID)}
	}
}

func (c *ALCCmd) createSkeleton(projectID string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles
		path := fmt.Sprintf("/alc/projects/%s/testing/skeleton", projectID)
		err := ctx.Client.ALCCommand(path, nil)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to create skeleton: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Skeleton created for " + projectID)}
	}
}

func (c *ALCCmd) implementSpoke(projectID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /alc <id> implement <spoke_id> [notes]")}
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

		path := fmt.Sprintf("/alc/projects/%s/testing/implementations", projectID)
		err := ctx.Client.ALCCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to implement spoke: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Implemented spoke: " + spokeID)}
	}
}

func (c *ALCCmd) verifyBuild(projectID string, args []string, ctx *Context) tea.Cmd {
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

		path := fmt.Sprintf("/alc/projects/%s/testing/builds", projectID)
		err := ctx.Client.ALCCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to verify build: " + err.Error())}
		}

		label := s.StatusOK.Render("PASS")
		if result == "fail" {
			label = s.StatusError.Render("FAIL")
		}
		return InjectSystemMsg{Content: fmt.Sprintf("Build verification: %s for %s", label, projectID)}
	}
}

func (c *ALCCmd) deployAction(projectID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{
				Content: ctx.Styles.Error.Render("Usage: /alc <id> deploy start | /alc <id> deploy record <env> <version>"),
			}
		}
	}

	sub := strings.ToLower(args[0])

	if sub == "start" {
		return func() tea.Msg {
			s := ctx.Styles
			path := fmt.Sprintf("/alc/projects/%s/deployment/start", projectID)
			err := ctx.Client.ALCCommand(path, nil)
			if err != nil {
				return InjectSystemMsg{Content: s.Error.Render("Failed to start deployment phase: " + err.Error())}
			}
			return InjectSystemMsg{Content: s.StatusOK.Render("Started deployment phase for " + projectID)}
		}
	}

	if sub == "record" {
		if len(args) < 3 {
			return func() tea.Msg {
				return InjectSystemMsg{
					Content: ctx.Styles.Error.Render("Usage: /alc <id> deploy record <environment> <version> [notes]"),
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

			path := fmt.Sprintf("/alc/projects/%s/deployment/deployments", projectID)
			err := ctx.Client.ALCCommand(path, body)
			if err != nil {
				return InjectSystemMsg{Content: s.Error.Render("Failed to record deployment: " + err.Error())}
			}
			return InjectSystemMsg{
				Content: s.StatusOK.Render(fmt.Sprintf("Recorded deployment: %s v%s to %s", projectID, version, env)),
			}
		}
	}

	return func() tea.Msg {
		return InjectSystemMsg{
			Content: ctx.Styles.Error.Render("Unknown deploy subcommand: " + sub + ". Use 'start' or 'record'."),
		}
	}
}

func (c *ALCCmd) reportIncident(projectID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /alc <id> incident <description>")}
		}
	}

	description := strings.Join(args, " ")

	return func() tea.Msg {
		s := ctx.Styles
		body := map[string]interface{}{"description": description}

		path := fmt.Sprintf("/alc/projects/%s/deployment/incidents", projectID)
		err := ctx.Client.ALCCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to report incident: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusWarning.Render("Incident reported for " + projectID)}
	}
}

func (c *ALCCmd) resolveIncident(projectID string, args []string, ctx *Context) tea.Cmd {
	if len(args) < 2 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /alc <id> resolve <incident_id> <resolution>")}
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

		path := fmt.Sprintf("/alc/projects/%s/deployment/incidents/resolve", projectID)
		err := ctx.Client.ALCCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to resolve incident: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Resolved incident: " + incidentID)}
	}
}

func (c *ALCCmd) completePhase(projectID string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles
		path := fmt.Sprintf("/alc/projects/%s/complete", projectID)
		err := ctx.Client.ALCCommand(path, nil)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to complete phase: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Phase completed for " + projectID)}
	}
}

// formatPhase returns a human-readable phase name.
func formatPhase(phase string) string {
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
		return "—"
	}
	return time.Unix(ts, 0).Format("2006-01-02 15:04")
}
