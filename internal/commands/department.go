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
func (c *DepartmentCmd) Aliases() []string   { return []string{"dept", "div", "division", "alc", "lifecycle", "lc"} }
func (c *DepartmentCmd) Description() string { return "Manage departments (divisions)" }

// ventureIDFromContext extracts the active venture ID from the ALC context.
func ventureIDFromContext(ctx *Context) string {
	if ctx.GetALCContext == nil {
		return ""
	}
	state := ctx.GetALCContext()
	if state == nil || state.Venture == nil {
		return ""
	}
	return state.Venture.ID
}

// divisionCmdPath builds an API path for a division command under a venture.
func divisionCmdPath(ventureID, divisionID, suffix string) string {
	return "/api/ventures/" + ventureID + "/divisions/" + divisionID + "/" + suffix
}

// requireVentureMsg returns an error message if no venture is active.
func requireVentureMsg(ctx *Context) tea.Msg {
	return InjectSystemMsg{
		Content: ctx.Styles.Error.Render("No active venture. Use /venture select to choose one first."),
	}
}

func (c *DepartmentCmd) Execute(args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return c.showUsage(ctx)
	}

	sub := strings.ToLower(args[0])

	// "init" creates a new department (discovers a division)
	if sub == "init" {
		return c.initDepartment(args[1:], ctx)
	}

	// Everything else requires a division ID as first arg
	if !strings.HasPrefix(sub, "div-") {
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
	case "design":
		return c.phaseAction(departmentID, "design", rest, ctx)
	case "finding":
		return c.recordFinding(departmentID, rest, ctx)
	case "term":
		return c.defineTerm(departmentID, rest, ctx)
	case "transition":
		return c.transition(departmentID, rest, ctx)
	case "dossier":
		return c.defineDossier(departmentID, rest, ctx)
	case "desk":
		return c.inventoryDesk(departmentID, rest, ctx)
	case "plan":
		return c.phaseAction(departmentID, "plan", rest, ctx)
	case "approve":
		return c.approvePlan(departmentID, rest, ctx)
	case "test":
		return c.phaseAction(departmentID, "testing", rest, ctx)
	case "skeleton":
		return c.createSkeleton(departmentID, ctx)
	case "implement":
		return c.implementDesk(departmentID, rest, ctx)
	case "verify":
		return c.verifyBuild(departmentID, rest, ctx)
	case "deploy":
		return c.deployAction(departmentID, rest, ctx)
	case "monitor":
		return c.phaseAction(departmentID, "monitoring", rest, ctx)
	case "incident":
		return c.reportIncident(departmentID, rest, ctx)
	case "resolve":
		return c.resolveIncident(departmentID, rest, ctx)
	case "rescue":
		return c.phaseAction(departmentID, "rescue", rest, ctx)
	case "generate":
		return c.phaseAction(departmentID, "generation", rest, ctx)
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
		b.WriteString(s.CardTitle.Render("Department - Division Lifecycle Commands"))
		b.WriteString("\n\n")

		// Intro
		b.WriteString(s.Subtle.Render("Manage divisions through their lifecycle: Design -> Plan -> Generate -> Test -> Deploy -> Monitor -> Rescue"))
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
		b.WriteString(row("/dept init <name>", "Discover a new division"))
		b.WriteString(row("/dept <id>", "Show division status"))
		b.WriteString(row("/dept <id> transition X", "Move to phase"))
		b.WriteString(row("/dept <id> complete", "Complete current phase"))
		b.WriteString("\n")

		// Design phase
		b.WriteString(section("Design", "Design aggregates and events"))
		b.WriteString(row("/dept <id> design start", "Begin design"))
		b.WriteString(row("/dept <id> dossier <name>", "Define aggregate/entity"))
		b.WriteString(row("/dept <id> desk <n> <t> <did>", "Add desk (vertical slice)"))
		b.WriteString("\n")

		// Plan phase
		b.WriteString(section("Plan", "Plan desks and dependencies"))
		b.WriteString(row("/dept <id> plan start", "Begin planning"))
		b.WriteString(row("/dept <id> approve <plan_id>", "Approve plan"))
		b.WriteString("\n")

		// Test phase
		b.WriteString(section("Test", "Implement features and verify quality"))
		b.WriteString(row("/dept <id> test start", "Begin testing"))
		b.WriteString(row("/dept <id> skeleton", "Generate code skeleton"))
		b.WriteString(row("/dept <id> implement <desk_id>", "Mark desk implemented"))
		b.WriteString(row("/dept <id> verify pass|fail", "Record build result"))
		b.WriteString("\n")

		// Deploy phase
		b.WriteString(section("Deploy", "Release to production"))
		b.WriteString(row("/dept <id> deploy start", "Begin deployment"))
		b.WriteString(row("/dept <id> deploy record <e> <v>", "Record release"))
		b.WriteString("\n")

		// Monitor phase
		b.WriteString(section("Monitor & Rescue", "Observe and handle incidents"))
		b.WriteString(row("/dept <id> monitor start", "Begin monitoring"))
		b.WriteString(row("/dept <id> incident <desc>", "Report incident"))
		b.WriteString(row("/dept <id> resolve <iid> <res>", "Resolve incident"))
		b.WriteString(row("/dept <id> rescue start", "Begin rescue"))

		return InjectSystemMsg{Content: b.String()}
	}
}

func (c *DepartmentCmd) initDepartment(args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /dept init <name> [description]")}
		}
	}

	name := args[0]
	desc := ""
	if len(args) > 1 {
		desc = strings.Join(args[1:], " ")
	}

	return func() tea.Msg {
		s := ctx.Styles
		ventureID := ventureIDFromContext(ctx)
		if ventureID == "" {
			return requireVentureMsg(ctx)
		}

		body := map[string]interface{}{"context_name": name}
		if desc != "" {
			body["description"] = desc
		}

		path := "/api/ventures/" + ventureID + "/discovery/divisions/discover"
		err := ctx.Client.DepartmentCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to discover division: " + err.Error())}
		}

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("Division Discovered"))
		b.WriteString("\n\n")
		b.WriteString(s.CardLabel.Render("Name: "))
		b.WriteString(s.CardValue.Render(name))
		if desc != "" {
			b.WriteString("\n")
			b.WriteString(s.CardLabel.Render("Description: "))
			b.WriteString(s.CardValue.Render(desc))
		}
		b.WriteString("\n\n")
		b.WriteString(s.Subtle.Render("  Use /dept to browse divisions"))

		return InjectSystemMsg{Content: b.String()}
	}
}

func (c *DepartmentCmd) showDepartment(departmentID string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles
		ventureID := ventureIDFromContext(ctx)
		if ventureID == "" {
			return requireVentureMsg(ctx)
		}

		dept, err := ctx.Client.GetDepartment(ventureID, departmentID)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to get division: " + err.Error())}
		}

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("Division: " + dept.Name))
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
		b.WriteString(s.CardLabel.Render("Desks: "))
		b.WriteString(s.CardValue.Render(fmt.Sprintf("%d", dept.DeskCount)))
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("Implemented: "))
		b.WriteString(s.CardValue.Render(fmt.Sprintf("%d/%d", dept.ImplementedDeskCount, dept.DeskCount)))
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
				Content: ctx.Styles.Error.Render(fmt.Sprintf("Usage: /dept %s %s start", departmentID, phase)),
			}
		}
	}

	return func() tea.Msg {
		s := ctx.Styles
		ventureID := ventureIDFromContext(ctx)
		if ventureID == "" {
			return requireVentureMsg(ctx)
		}

		path := divisionCmdPath(ventureID, departmentID, phase+"/start")
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
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /dept <id> finding <title> [content]")}
		}
	}

	title := args[0]
	content := ""
	if len(args) > 1 {
		content = strings.Join(args[1:], " ")
	}

	return func() tea.Msg {
		s := ctx.Styles
		ventureID := ventureIDFromContext(ctx)
		if ventureID == "" {
			return requireVentureMsg(ctx)
		}

		body := map[string]interface{}{"title": title}
		if content != "" {
			body["content"] = content
		}

		path := divisionCmdPath(ventureID, departmentID, "discovery/findings/record")
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
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /dept <id> term <term> <definition>")}
		}
	}

	term := args[0]
	definition := strings.Join(args[1:], " ")

	return func() tea.Msg {
		s := ctx.Styles
		ventureID := ventureIDFromContext(ctx)
		if ventureID == "" {
			return requireVentureMsg(ctx)
		}

		body := map[string]interface{}{"term": term, "definition": definition}
		path := divisionCmdPath(ventureID, departmentID, "discovery/terms/define")
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
				Content: ctx.Styles.Error.Render("Usage: /dept <id> transition <target_phase>"),
			}
		}
	}

	targetPhase := args[0]

	return func() tea.Msg {
		s := ctx.Styles
		ventureID := ventureIDFromContext(ctx)
		if ventureID == "" {
			return requireVentureMsg(ctx)
		}

		body := map[string]interface{}{"target_phase": targetPhase}
		path := divisionCmdPath(ventureID, departmentID, "transition")
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
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /dept <id> dossier <name> [description]")}
		}
	}

	name := args[0]
	desc := ""
	if len(args) > 1 {
		desc = strings.Join(args[1:], " ")
	}

	return func() tea.Msg {
		s := ctx.Styles
		ventureID := ventureIDFromContext(ctx)
		if ventureID == "" {
			return requireVentureMsg(ctx)
		}

		body := map[string]interface{}{"dossier_name": name}
		if desc != "" {
			body["description"] = desc
		}

		path := divisionCmdPath(ventureID, departmentID, "design/aggregates/design")
		err := ctx.Client.DepartmentCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to define dossier: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Defined dossier: " + name)}
	}
}

func (c *DepartmentCmd) inventoryDesk(departmentID string, args []string, ctx *Context) tea.Cmd {
	if len(args) < 3 {
		return func() tea.Msg {
			return InjectSystemMsg{
				Content: ctx.Styles.Error.Render("Usage: /dept <id> desk <name> <type> <dossier_id> [description]"),
			}
		}
	}

	name := args[0]
	deskType := args[1]
	dossierID := args[2]
	desc := ""
	if len(args) > 3 {
		desc = strings.Join(args[3:], " ")
	}

	return func() tea.Msg {
		s := ctx.Styles
		ventureID := ventureIDFromContext(ctx)
		if ventureID == "" {
			return requireVentureMsg(ctx)
		}

		body := map[string]interface{}{
			"desk_name":  name,
			"desk_type":  deskType,
			"dossier_id": dossierID,
		}
		if desc != "" {
			body["description"] = desc
		}

		path := divisionCmdPath(ventureID, departmentID, "plan/desks/plan")
		err := ctx.Client.DepartmentCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to plan desk: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Planned desk: " + name)}
	}
}

func (c *DepartmentCmd) approvePlan(departmentID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /dept <id> approve <plan_id>")}
		}
	}

	planID := args[0]

	return func() tea.Msg {
		s := ctx.Styles
		ventureID := ventureIDFromContext(ctx)
		if ventureID == "" {
			return requireVentureMsg(ctx)
		}

		body := map[string]interface{}{"plan_id": planID}
		path := divisionCmdPath(ventureID, departmentID, "plan/complete")
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
		ventureID := ventureIDFromContext(ctx)
		if ventureID == "" {
			return requireVentureMsg(ctx)
		}

		path := divisionCmdPath(ventureID, departmentID, "generation/modules/generate")
		err := ctx.Client.DepartmentCommand(path, nil)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to generate skeleton: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Skeleton generated for " + departmentID)}
	}
}

func (c *DepartmentCmd) implementDesk(departmentID string, args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /dept <id> implement <desk_id> [notes]")}
		}
	}

	deskID := args[0]
	notes := ""
	if len(args) > 1 {
		notes = strings.Join(args[1:], " ")
	}

	return func() tea.Msg {
		s := ctx.Styles
		ventureID := ventureIDFromContext(ctx)
		if ventureID == "" {
			return requireVentureMsg(ctx)
		}

		body := map[string]interface{}{"desk_id": deskID}
		if notes != "" {
			body["implementation_notes"] = notes
		}

		path := divisionCmdPath(ventureID, departmentID, "testing/suites/run")
		err := ctx.Client.DepartmentCommand(path, body)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to implement desk: " + err.Error())}
		}
		return InjectSystemMsg{Content: s.StatusOK.Render("Implemented desk: " + deskID)}
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
		ventureID := ventureIDFromContext(ctx)
		if ventureID == "" {
			return requireVentureMsg(ctx)
		}

		body := map[string]interface{}{"result": result}
		if notes != "" {
			body["notes"] = notes
		}

		path := divisionCmdPath(ventureID, departmentID, "testing/results/record")
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
				Content: ctx.Styles.Error.Render("Usage: /dept <id> deploy start | /dept <id> deploy record <env> <version>"),
			}
		}
	}

	sub := strings.ToLower(args[0])

	if sub == "start" {
		return func() tea.Msg {
			s := ctx.Styles
			ventureID := ventureIDFromContext(ctx)
			if ventureID == "" {
				return requireVentureMsg(ctx)
			}

			path := divisionCmdPath(ventureID, departmentID, "deployment/start")
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
					Content: ctx.Styles.Error.Render("Usage: /dept <id> deploy record <environment> <version> [notes]"),
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
			ventureID := ventureIDFromContext(ctx)
			if ventureID == "" {
				return requireVentureMsg(ctx)
			}

			body := map[string]interface{}{
				"environment": env,
				"version":     version,
			}
			if notes != "" {
				body["notes"] = notes
			}

			path := divisionCmdPath(ventureID, departmentID, "deployment/releases/deploy")
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
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /dept <id> incident <description>")}
		}
	}

	description := strings.Join(args, " ")

	return func() tea.Msg {
		s := ctx.Styles
		ventureID := ventureIDFromContext(ctx)
		if ventureID == "" {
			return requireVentureMsg(ctx)
		}

		body := map[string]interface{}{"description": description}
		path := divisionCmdPath(ventureID, departmentID, "monitoring/incidents/raise")
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
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /dept <id> resolve <incident_id> <resolution>")}
		}
	}

	incidentID := args[0]
	resolution := strings.Join(args[1:], " ")

	return func() tea.Msg {
		s := ctx.Styles
		ventureID := ventureIDFromContext(ctx)
		if ventureID == "" {
			return requireVentureMsg(ctx)
		}

		body := map[string]interface{}{
			"incident_id": incidentID,
			"resolution":  resolution,
		}

		path := divisionCmdPath(ventureID, departmentID, "rescue/diagnoses/diagnose")
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
		ventureID := ventureIDFromContext(ctx)
		if ventureID == "" {
			return requireVentureMsg(ctx)
		}

		// Fetch department to determine current phase
		department, err := ctx.Client.GetDepartment(ventureID, departmentID)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to get division: " + err.Error())}
		}

		// Map phase to endpoint path segment
		var phasePath string
		switch strings.ToLower(department.CurrentPhase) {
		case "design":
			phasePath = "design"
		case "plan":
			phasePath = "plan"
		case "generation":
			phasePath = "generation"
		case "testing":
			phasePath = "testing"
		case "deployment":
			phasePath = "deployment"
		case "monitoring":
			phasePath = "monitoring"
		case "rescue":
			phasePath = "rescue"
		default:
			return InjectSystemMsg{Content: s.Error.Render("Cannot complete phase: " + department.CurrentPhase)}
		}

		path := divisionCmdPath(ventureID, departmentID, phasePath+"/complete")
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
	case "design":
		return "Design"
	case "plan":
		return "Plan"
	case "generation":
		return "Generation"
	case "testing":
		return "Testing"
	case "deployment":
		return "Deployment"
	case "monitoring":
		return "Monitoring"
	case "rescue":
		return "Rescue"
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
