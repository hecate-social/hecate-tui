package commands

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DepartmentCmd handles all /department subcommands for bounded context management.
type DepartmentCmd struct{}

func (c *DepartmentCmd) Name() string        { return "department" }
func (c *DepartmentCmd) Aliases() []string   { return []string{"dept", "alc", "lifecycle", "lc"} }
func (c *DepartmentCmd) Description() string { return "Manage departments" }

func (c *DepartmentCmd) Execute(args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return c.showUsage(ctx)
	}

	sub := strings.ToLower(args[0])

	// "init" creates a new department
	if sub == "init" {
		return c.initDepartment(args[1:], ctx)
	}

	// Everything else requires a department ID as first arg
	if !strings.HasPrefix(sub, "prj-") {
		return c.showUsage(ctx)
	}

	departmentID := sub
	if len(args) < 2 {
		// Just a department ID -> show status card
		return c.showDepartment(departmentID, ctx)
	}

	action := strings.ToLower(args[1])
	rest := args[2:]

	switch action {
	case "discovery":
		return c.phaseAction(departmentID, "discovery", rest, ctx)
	case "finding":
		return c.recordFinding(departmentID, rest, ctx)
	case "term":
		return c.defineTerm(departmentID, rest, ctx)
	case "transition":
		return c.transition(departmentID, rest, ctx)
	case "arch":
		return c.phaseAction(departmentID, "architecture", rest, ctx)
	case "dossier":
		return c.defineDossier(departmentID, rest, ctx)
	case "spoke":
		return c.inventorySpoke(departmentID, rest, ctx)
	case "plan":
		return c.draftPlan(departmentID, rest, ctx)
	case "approve":
		return c.approvePlan(departmentID, rest, ctx)
	case "test":
		return c.phaseAction(departmentID, "testing", rest, ctx)
	case "skeleton":
		return c.createSkeleton(departmentID, ctx)
	case "implement":
		return c.implementSpoke(departmentID, rest, ctx)
	case "verify":
		return c.verifyBuild(departmentID, rest, ctx)
	case "deploy":
		return c.deployAction(departmentID, rest, ctx)
	case "incident":
		return c.reportIncident(departmentID, rest, ctx)
	case "resolve":
		return c.resolveIncident(departmentID, rest, ctx)
	case "complete":
		return c.completePhase(departmentID, ctx)
	default:
		return c.showUsage(ctx)
	}
}

func (c *DepartmentCmd) showUsage(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles
		t := ctx.Theme
		var b strings.Builder

		// Title
		b.WriteString(s.CardTitle.Render("Department - Bounded Context Lifecycle Commands"))
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
		b.WriteString(row("/department init <name>", "Create a new department"))
		b.WriteString(row("/department <id>", "Show department status"))
		b.WriteString(row("/department <id> transition X", "Move to phase (arch, test, deploy)"))
		b.WriteString(row("/department <id> complete", "Complete current phase"))
		b.WriteString("\n")

		// Phase 1
		b.WriteString(section("Phase 1: Discovery & Analysis", "Analyze requirements, gather findings, build vocabulary"))
		b.WriteString(row("/department <id> discovery start", "Begin discovery"))
		b.WriteString(row("/department <id> finding <title>", "Record insight or requirement"))
		b.WriteString(row("/department <id> term <t> <def>", "Define domain term"))
		b.WriteString("\n")

		// Phase 2
		b.WriteString(section("Phase 2: Architecture & Planning", "Design dossiers (aggregates) and spokes (operations)"))
		b.WriteString(row("/department <id> arch start", "Begin architecture"))
		b.WriteString(row("/department <id> dossier <name>", "Define aggregate/entity"))
		b.WriteString(row("/department <id> spoke <n> <t> <did>", "Add operation (cmd/qry/evt)"))
		b.WriteString(row("/department <id> plan", "Draft implementation plan"))
		b.WriteString(row("/department <id> approve", "Approve plan"))
		b.WriteString("\n")

		// Phase 3
		b.WriteString(section("Phase 3: Testing & Implementation", "Implement features and verify quality"))
		b.WriteString(row("/department <id> test start", "Begin testing"))
		b.WriteString(row("/department <id> skeleton", "Generate code skeleton"))
		b.WriteString(row("/department <id> implement <sid>", "Mark spoke implemented"))
		b.WriteString(row("/department <id> verify pass|fail", "Record build result"))
		b.WriteString("\n")

		// Phase 4
		b.WriteString(section("Phase 4: Deployment & Operations", "Release to production, monitor, handle incidents"))
		b.WriteString(row("/department <id> deploy start", "Begin deployment"))
		b.WriteString(row("/department <id> deploy record <e> <v>", "Record release"))
		b.WriteString(row("/department <id> incident <desc>", "Report incident"))
		b.WriteString(row("/department <id> resolve <iid> <res>", "Resolve incident"))

		return InjectSystemMsg{Content: b.String()}
	}
}

func (c *DepartmentCmd) initDepartment(args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /department init <name> [description]")}
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

		err := ctx.Client.DepartmentCommand("/api/cartwheels/initiate", body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to initiate department: " + err.Error())}
		}

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("Department Initiated"))
		b.WriteString("\n\n")
		b.WriteString(s.CardLabel.Render("Name: "))
		b.WriteString(s.CardValue.Render(name))
		if desc != "" {
			b.WriteString("\n")
			b.WriteString(s.CardLabel.Render("Description: "))
			b.WriteString(s.CardValue.Render(desc))
		}
		b.WriteString("\n\n")
		b.WriteString(s.Subtle.Render("  Use /department to browse departments"))

		return InjectSystemMsg{Content: b.String()}
	}
}

func (c *DepartmentCmd) showDepartment(departmentID string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles
		dept, err := ctx.Client.GetDepartment(departmentID)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to get department: " + err.Error())}
		}

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("Department: " + dept.Name))
		b.WriteString("\n\n")

		b.WriteString(s.CardLabel.Render("ID: "))
		b.WriteString(s.CardValue.Render(dept.DepartmentID))
		b.WriteString("\n")

		if dept.Description != "" {
			b.WriteString(s.CardLabel.Render("Description: "))
			b.WriteString(s.CardValue.Render(dept.Description))
			b.WriteString("\n")
		}

		b.WriteString(s.CardLabel.Render("Phase: "))
		b.WriteString(s.CardValue.Render(formatDepartmentPhase(dept.CurrentPhase)))
		b.WriteString("\n")

		b.WriteString(s.CardLabel.Render("Initiated: "))
		b.WriteString(s.Subtle.Render(formatTimestamp(dept.InitiatedAt)))
		b.WriteString("\n\n")

		// Phase-specific counters
		b.WriteString(s.Bold.Render("  Counters"))
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("Findings: "))
		b.WriteString(s.CardValue.Render(fmt.Sprintf("%d", dept.FindingCount)))
		b.WriteString("  ")
		b.WriteString(s.CardLabel.Render("Terms: "))
		b.WriteString(s.CardValue.Render(fmt.Sprintf("%d", dept.TermCount)))
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("Dossiers: "))
		b.WriteString(s.CardValue.Render(fmt.Sprintf("%d", dept.DossierCount)))
		b.WriteString("  ")
		b.WriteString(s.CardLabel.Render("Spokes: "))
		b.WriteString(s.CardValue.Render(fmt.Sprintf("%d", dept.SpokeCount)))
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("Implemented: "))
		b.WriteString(s.CardValue.Render(fmt.Sprintf("%d/%d", dept.ImplementedSpokeCount, dept.SpokeCount)))
		b.WriteString("  ")
		b.WriteString(s.CardLabel.Render("Build: "))
		if dept.BuildVerified {
			b.WriteString(s.StatusOK.Render("verified"))
		} else {
			b.WriteString(s.Subtle.Render("pending"))
		}
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("Deployments: "))
		b.WriteString(s.CardValue.Render(fmt.Sprintf("%d", dept.DeploymentCount)))
		b.WriteString("  ")
		b.WriteString(s.CardLabel.Render("Incidents: "))
		if dept.ActiveIncidents > 0 {
			b.WriteString(s.StatusError.Render(fmt.Sprintf("%d active", dept.ActiveIncidents)))
		} else {
			b.WriteString(s.StatusOK.Render("none"))
		}
		b.WriteString("\n")

		return InjectSystemMsg{Content: b.String()}
	}
}

func (c *DepartmentCmd) phaseAction(departmentID, phase string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 || strings.ToLower(args[0]) != "start" {
		return func() tea.Msg {
			return InjectSystemMsg{
				Content: ctx.Styles.Error.Render(fmt.Sprintf("Usage: /department %s %s start", departmentID, phase)),
			}
		}
	}

	return func() tea.Msg {
		s := ctx.Styles
		path := fmt.Sprintf("/api/cartwheels/%s/%s/start", departmentID, phase)
		err := ctx.Client.DepartmentCommand(path, nil)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to start " + phase + ": " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Started " + phase + " phase for " + departmentID)}
	}
}

func (c *DepartmentCmd) recordFinding(departmentID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /department <id> finding <title> [content]")}
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

		path := fmt.Sprintf("/api/cartwheels/%s/discovery/findings/record", departmentID)
		err := ctx.Client.DepartmentCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to record finding: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Recorded finding: " + title)}
	}
}

func (c *DepartmentCmd) defineTerm(departmentID string, args []string, ctx *Context) tea.Cmd {
	if len(args) < 2 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /department <id> term <term> <definition>")}
		}
	}

	term := args[0]
	definition := strings.Join(args[1:], " ")

	return func() tea.Msg {
		s := ctx.Styles
		body := map[string]interface{}{"term": term, "definition": definition}

		path := fmt.Sprintf("/api/cartwheels/%s/discovery/terms/define", departmentID)
		err := ctx.Client.DepartmentCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to define term: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Defined term: " + term)}
	}
}

func (c *DepartmentCmd) transition(departmentID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{
				Content: ctx.Styles.Error.Render("Usage: /department <id> transition <target_phase>"),
			}
		}
	}

	targetPhase := args[0]

	return func() tea.Msg {
		s := ctx.Styles
		body := map[string]interface{}{"target_phase": targetPhase}

		path := fmt.Sprintf("/api/cartwheels/%s/transition", departmentID)
		err := ctx.Client.DepartmentCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to transition: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Transitioned " + departmentID + " to " + targetPhase)}
	}
}

func (c *DepartmentCmd) defineDossier(departmentID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /department <id> dossier <name> [description]")}
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

		path := fmt.Sprintf("/api/cartwheels/%s/architecture/dossiers/define", departmentID)
		err := ctx.Client.DepartmentCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to define dossier: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Defined dossier: " + name)}
	}
}

func (c *DepartmentCmd) inventorySpoke(departmentID string, args []string, ctx *Context) tea.Cmd {
	if len(args) < 3 {
		return func() tea.Msg {
			return InjectSystemMsg{
				Content: ctx.Styles.Error.Render("Usage: /department <id> spoke <name> <type> <dossier_id> [description]"),
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

		path := fmt.Sprintf("/api/cartwheels/%s/architecture/spokes/inventory", departmentID)
		err := ctx.Client.DepartmentCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to inventory spoke: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Inventoried spoke: " + name)}
	}
}

func (c *DepartmentCmd) draftPlan(departmentID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /department <id> plan <title> [description]")}
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

		path := fmt.Sprintf("/api/cartwheels/%s/architecture/plan", departmentID)
		err := ctx.Client.DepartmentCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to draft plan: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Drafted plan: " + title)}
	}
}

func (c *DepartmentCmd) approvePlan(departmentID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /department <id> approve <plan_id>")}
		}
	}

	planID := args[0]

	return func() tea.Msg {
		s := ctx.Styles
		body := map[string]interface{}{"plan_id": planID}

		path := fmt.Sprintf("/api/cartwheels/%s/architecture/plan/approve", departmentID)
		err := ctx.Client.DepartmentCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to approve plan: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Approved plan: " + planID)}
	}
}

func (c *DepartmentCmd) createSkeleton(departmentID string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles
		path := fmt.Sprintf("/api/cartwheels/%s/testing/skeleton", departmentID)
		err := ctx.Client.DepartmentCommand(path, nil)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to create skeleton: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Skeleton created for " + departmentID)}
	}
}

func (c *DepartmentCmd) implementSpoke(departmentID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /department <id> implement <spoke_id> [notes]")}
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

		path := fmt.Sprintf("/api/cartwheels/%s/testing/implement", departmentID)
		err := ctx.Client.DepartmentCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to implement spoke: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Implemented spoke: " + spokeID)}
	}
}

func (c *DepartmentCmd) verifyBuild(departmentID string, args []string, ctx *Context) tea.Cmd {
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

		path := fmt.Sprintf("/api/cartwheels/%s/testing/verify", departmentID)
		err := ctx.Client.DepartmentCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to verify build: " + err.Error())}
		}

		label := s.StatusOK.Render("PASS")
		if result == "fail" {
			label = s.StatusError.Render("FAIL")
		}
		return InjectSystemMsg{Content: fmt.Sprintf("Build verification: %s for %s", label, departmentID)}
	}
}

func (c *DepartmentCmd) deployAction(departmentID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{
				Content: ctx.Styles.Error.Render("Usage: /department <id> deploy start | /department <id> deploy record <env> <version>"),
			}
		}
	}

	sub := strings.ToLower(args[0])

	if sub == "start" {
		return func() tea.Msg {
			s := ctx.Styles
			path := fmt.Sprintf("/api/cartwheels/%s/deployment/start", departmentID)
			err := ctx.Client.DepartmentCommand(path, nil)
			if err != nil {
				return InjectSystemMsg{Content: s.Error.Render("Failed to start deployment phase: " + err.Error())}
			}
			return InjectSystemMsg{Content: s.StatusOK.Render("Started deployment phase for " + departmentID)}
		}
	}

	if sub == "record" {
		if len(args) < 3 {
			return func() tea.Msg {
				return InjectSystemMsg{
					Content: ctx.Styles.Error.Render("Usage: /department <id> deploy record <environment> <version> [notes]"),
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

			path := fmt.Sprintf("/api/cartwheels/%s/deployment/record", departmentID)
			err := ctx.Client.DepartmentCommand(path, body)
			if err != nil {
				return InjectSystemMsg{Content: s.Error.Render("Failed to record deployment: " + err.Error())}
			}
			return InjectSystemMsg{
				Content: s.StatusOK.Render(fmt.Sprintf("Recorded deployment: %s v%s to %s", departmentID, version, env)),
			}
		}
	}

	return func() tea.Msg {
		return InjectSystemMsg{
			Content: ctx.Styles.Error.Render("Unknown deploy subcommand: " + sub + ". Use 'start' or 'record'."),
		}
	}
}

func (c *DepartmentCmd) reportIncident(departmentID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /department <id> incident <description>")}
		}
	}

	description := strings.Join(args, " ")

	return func() tea.Msg {
		s := ctx.Styles
		body := map[string]interface{}{"description": description}

		path := fmt.Sprintf("/api/cartwheels/%s/deployment/incident", departmentID)
		err := ctx.Client.DepartmentCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to report incident: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusWarning.Render("Incident reported for " + departmentID)}
	}
}

func (c *DepartmentCmd) resolveIncident(departmentID string, args []string, ctx *Context) tea.Cmd {
	if len(args) < 2 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /department <id> resolve <incident_id> <resolution>")}
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

		path := fmt.Sprintf("/api/cartwheels/%s/deployment/incident/resolve", departmentID)
		err := ctx.Client.DepartmentCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to resolve incident: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Resolved incident: " + incidentID)}
	}
}

func (c *DepartmentCmd) completePhase(departmentID string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		// Fetch department to determine current phase
		department, err := ctx.Client.GetDepartment(departmentID)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to get department: " + err.Error())}
		}

		// Map phase to endpoint path segment
		var phasePath string
		switch strings.ToLower(department.CurrentPhase) {
		case "discovery", "dna":
			phasePath = "discovery"
		case "architecture", "anp":
			phasePath = "architecture"
		case "testing", "tni":
			phasePath = "testing"
		case "deployment", "dno":
			phasePath = "deployment"
		default:
			return InjectSystemMsg{Content: s.Error.Render("Cannot complete phase: " + department.CurrentPhase)}
		}

		path := fmt.Sprintf("/api/cartwheels/%s/%s/complete", departmentID, phasePath)
		err = ctx.Client.DepartmentCommand(path, nil)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to complete phase: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Completed " + phasePath + " phase for " + departmentID)}
	}
}

// formatDepartmentPhase returns a human-readable phase name.
func formatDepartmentPhase(phase string) string {
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

// formatTimestamp converts a Unix timestamp (milliseconds) to a readable string.
func formatTimestamp(ts int64) string {
	if ts == 0 {
		return "-"
	}
	return time.UnixMilli(ts).Format("2006-01-02 15:04")
}
