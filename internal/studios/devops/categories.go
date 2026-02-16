package devops

import (
	"fmt"

	"github.com/hecate-social/hecate-tui/internal/ui"
)

// actionView tracks navigation state for the categories/actions/form overlay.
type actionView int

const (
	actionViewNone       actionView = iota // task list (default)
	actionViewCategories                   // picking a category
	actionViewActions                      // picking an action in a category
	actionViewForm                         // filling out a form
)

// Category groups related actions.
type Category struct {
	Name    string
	Icon    string
	Actions []Action
}

// Action is a single command the user can trigger.
type Action struct {
	Name        string
	Verb        string                                             // display verb
	FormSpec    *ui.FormSpec                                       // nil = confirm-only (y/n)
	APIPath     func(ventureID, divisionID string) string          // builds the endpoint path
	BodyBuilder func(vals map[string]string) map[string]interface{} // builds the JSON body
}

func ventureCategories() []Category {
	return []Category{
		ventureCategory(),
		divisionCategory(),
		designCategory(),
		planCategory(),
		buildCategory(),
		shipCategory(),
	}
}

// --- Venture ---

func ventureCategory() Category {
	initiateSpec := initiateVentureSpec()
	refineSpec := refineVisionSpec()
	archiveSpec := archiveVentureSpec()

	return Category{
		Name: "Venture",
		Icon: "\u2726", // four-pointed star
		Actions: []Action{
			{
				Name:     "Initiate Venture",
				Verb:     "initiate",
				FormSpec: &initiateSpec,
				APIPath: func(_, _ string) string {
					return "/api/ventures/setup"
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					return map[string]interface{}{
						"name":         vals["name"],
						"brief":        vals["brief"],
						"initiated_by": "tui",
					}
				},
			},
			{
				Name:     "Archive Venture",
				Verb:     "archive",
				FormSpec: &archiveSpec,
				APIPath: func(ventureID, _ string) string {
					return fmt.Sprintf("/api/ventures/%s/archive", ventureID)
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					return map[string]interface{}{
						"reason":      vals["reason"],
						"archived_by": "tui",
					}
				},
			},
			{
				Name:     "Refine Vision",
				Verb:     "refine",
				FormSpec: &refineSpec,
				APIPath: func(ventureID, _ string) string {
					return fmt.Sprintf("/api/ventures/%s/vision/refine", ventureID)
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					body := map[string]interface{}{}
					if v := vals["brief"]; v != "" {
						body["brief"] = v
					}
					if v := vals["repo_url"]; v != "" {
						body["repo_url"] = v
					}
					return body
				},
			},
			{
				Name:     "Submit Vision",
				Verb:     "submit",
				FormSpec: nil, // confirm-only
				APIPath: func(ventureID, _ string) string {
					return fmt.Sprintf("/api/ventures/%s/vision/submit", ventureID)
				},
				BodyBuilder: func(_ map[string]string) map[string]interface{} {
					return map[string]interface{}{
						"submitted_by": "tui",
					}
				},
			},
		},
	}
}

// --- Division ---

func divisionCategory() Category {
	initiateSpec := initiateDivisionSpec()
	startSpec := startPhaseSpec()
	archiveSpec := archiveDivisionSpec()

	return Category{
		Name: "Division",
		Icon: "\u2502", // vertical line
		Actions: []Action{
			{
				Name:     "Initiate Division",
				Verb:     "initiate",
				FormSpec: &initiateSpec,
				APIPath: func(ventureID, _ string) string {
					return fmt.Sprintf("/api/ventures/%s/divisions/setup", ventureID)
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					body := map[string]interface{}{
						"context_name": vals["context_name"],
					}
					if v := vals["description"]; v != "" {
						body["description"] = v
					}
					return body
				},
			},
			{
				Name:     "Start Phase",
				Verb:     "start",
				FormSpec: &startSpec,
				APIPath: func(ventureID, divisionID string) string {
					return fmt.Sprintf("/api/ventures/%s/divisions/%s/phase/start", ventureID, divisionID)
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					return map[string]interface{}{
						"phase": vals["phase"],
					}
				},
			},
			{
				Name:     "Pause Phase",
				Verb:     "pause",
				FormSpec: nil, // confirm-only
				APIPath: func(ventureID, divisionID string) string {
					return fmt.Sprintf("/api/ventures/%s/divisions/%s/phase/pause", ventureID, divisionID)
				},
				BodyBuilder: func(_ map[string]string) map[string]interface{} {
					return map[string]interface{}{}
				},
			},
			{
				Name:     "Resume Phase",
				Verb:     "resume",
				FormSpec: nil, // confirm-only
				APIPath: func(ventureID, divisionID string) string {
					return fmt.Sprintf("/api/ventures/%s/divisions/%s/phase/resume", ventureID, divisionID)
				},
				BodyBuilder: func(_ map[string]string) map[string]interface{} {
					return map[string]interface{}{}
				},
			},
			{
				Name:     "Complete Phase",
				Verb:     "complete",
				FormSpec: nil, // confirm-only
				APIPath: func(ventureID, divisionID string) string {
					return fmt.Sprintf("/api/ventures/%s/divisions/%s/phase/complete", ventureID, divisionID)
				},
				BodyBuilder: func(_ map[string]string) map[string]interface{} {
					return map[string]interface{}{}
				},
			},
			{
				Name:     "Archive Division",
				Verb:     "archive",
				FormSpec: &archiveSpec,
				APIPath: func(ventureID, divisionID string) string {
					return fmt.Sprintf("/api/ventures/%s/divisions/%s/archive", ventureID, divisionID)
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					return map[string]interface{}{
						"reason":      vals["reason"],
						"archived_by": "tui",
					}
				},
			},
		},
	}
}

// --- Design (DnA) ---

func designCategory() Category {
	aggSpec := designAggregateSpec()
	eventSpec := designEventSpec()

	return Category{
		Name: "Design (DnA)",
		Icon: "\u25ca", // diamond
		Actions: []Action{
			{
				Name:     "Design Aggregate",
				Verb:     "design",
				FormSpec: &aggSpec,
				APIPath: func(ventureID, divisionID string) string {
					return fmt.Sprintf("/api/ventures/%s/divisions/%s/design/dossiers", ventureID, divisionID)
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					body := map[string]interface{}{
						"aggregate_name": vals["aggregate_name"],
					}
					if v := vals["description"]; v != "" {
						body["description"] = v
					}
					if v := vals["stream_prefix"]; v != "" {
						body["stream_prefix"] = v
					}
					return body
				},
			},
			{
				Name:     "Design Event",
				Verb:     "design",
				FormSpec: &eventSpec,
				APIPath: func(ventureID, divisionID string) string {
					return fmt.Sprintf("/api/ventures/%s/divisions/%s/design/events", ventureID, divisionID)
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					body := map[string]interface{}{
						"event_name": vals["event_name"],
					}
					if v := vals["description"]; v != "" {
						body["description"] = v
					}
					if v := vals["aggregate_name"]; v != "" {
						body["aggregate_name"] = v
					}
					return body
				},
			},
		},
	}
}

// --- Plan (AnP) ---

func planCategory() Category {
	deskSpec := planDeskSpec()
	depSpec := planDependencySpec()

	return Category{
		Name: "Plan (AnP)",
		Icon: "\u2630", // trigram for heaven
		Actions: []Action{
			{
				Name:     "Plan Desk",
				Verb:     "plan",
				FormSpec: &deskSpec,
				APIPath: func(ventureID, divisionID string) string {
					return fmt.Sprintf("/api/ventures/%s/divisions/%s/plan/desks", ventureID, divisionID)
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					body := map[string]interface{}{
						"desk_name": vals["desk_name"],
					}
					if v := vals["description"]; v != "" {
						body["description"] = v
					}
					if v := vals["department"]; v != "" {
						body["department"] = v
					}
					return body
				},
			},
			{
				Name:     "Plan Dependency",
				Verb:     "plan",
				FormSpec: &depSpec,
				APIPath: func(ventureID, divisionID string) string {
					return fmt.Sprintf("/api/ventures/%s/divisions/%s/plan/dependencies", ventureID, divisionID)
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					body := map[string]interface{}{
						"from_desk": vals["from_desk"],
						"to_desk":   vals["to_desk"],
					}
					if v := vals["dep_type"]; v != "" {
						body["dep_type"] = v
					}
					return body
				},
			},
		},
	}
}

// --- Build (TnI) ---

func buildCategory() Category {
	genModSpec := generateModuleSpec()
	genTestSpec := generateTestSpec()
	runTestSpec := runTestSuiteSpec()
	recordTestSpec := recordTestResultSpec()

	return Category{
		Name: "Build (TnI)",
		Icon: "\u2692", // hammer and pick
		Actions: []Action{
			{
				Name:     "Generate Module",
				Verb:     "generate",
				FormSpec: &genModSpec,
				APIPath: func(ventureID, divisionID string) string {
					return fmt.Sprintf("/api/ventures/%s/divisions/%s/test/modules", ventureID, divisionID)
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					body := map[string]interface{}{
						"module_name": vals["module_name"],
					}
					if v := vals["module_type"]; v != "" {
						body["module_type"] = v
					}
					return body
				},
			},
			{
				Name:     "Generate Test",
				Verb:     "generate",
				FormSpec: &genTestSpec,
				APIPath: func(ventureID, divisionID string) string {
					return fmt.Sprintf("/api/ventures/%s/divisions/%s/test/tests", ventureID, divisionID)
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					body := map[string]interface{}{
						"test_name": vals["test_name"],
					}
					if v := vals["module_name"]; v != "" {
						body["module_name"] = v
					}
					return body
				},
			},
			{
				Name:     "Run Test Suite",
				Verb:     "run",
				FormSpec: &runTestSpec,
				APIPath: func(ventureID, divisionID string) string {
					return fmt.Sprintf("/api/ventures/%s/divisions/%s/test/run", ventureID, divisionID)
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					body := map[string]interface{}{}
					if v := vals["suite"]; v != "" {
						body["suite"] = v
					}
					return body
				},
			},
			{
				Name:     "Record Test Result",
				Verb:     "record",
				FormSpec: &recordTestSpec,
				APIPath: func(ventureID, divisionID string) string {
					return fmt.Sprintf("/api/ventures/%s/divisions/%s/test/results", ventureID, divisionID)
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					return map[string]interface{}{
						"result": vals["result"],
						"notes":  vals["notes"],
					}
				},
			},
		},
	}
}

// --- Ship (DnO) ---

func shipCategory() Category {
	deploySpec := deployReleaseSpec()
	stageSpec := stageRolloutSpec()
	healthCheckSpec := registerHealthCheckSpec()
	healthStatusSpec := recordHealthStatusSpec()
	incidentSpec := raiseIncidentSpec()
	diagnoseSpec := diagnoseIncidentSpec()
	fixSpec := applyFixSpec()

	return Category{
		Name: "Ship (DnO)",
		Icon: "\u2708", // airplane
		Actions: []Action{
			{
				Name:     "Deploy Release",
				Verb:     "deploy",
				FormSpec: &deploySpec,
				APIPath: func(ventureID, divisionID string) string {
					return fmt.Sprintf("/api/ventures/%s/divisions/%s/deploy/releases", ventureID, divisionID)
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					body := map[string]interface{}{
						"version":     vals["version"],
						"environment": vals["environment"],
					}
					if v := vals["notes"]; v != "" {
						body["notes"] = v
					}
					return body
				},
			},
			{
				Name:     "Stage Rollout",
				Verb:     "stage",
				FormSpec: &stageSpec,
				APIPath: func(ventureID, divisionID string) string {
					return fmt.Sprintf("/api/ventures/%s/divisions/%s/deploy/rollouts", ventureID, divisionID)
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					return map[string]interface{}{
						"version":     vals["version"],
						"environment": vals["environment"],
						"strategy":    vals["strategy"],
					}
				},
			},
			{
				Name:     "Register Health Check",
				Verb:     "register",
				FormSpec: &healthCheckSpec,
				APIPath: func(ventureID, divisionID string) string {
					return fmt.Sprintf("/api/ventures/%s/divisions/%s/deploy/health-checks", ventureID, divisionID)
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					return map[string]interface{}{
						"check_name": vals["check_name"],
						"endpoint":   vals["endpoint"],
						"interval":   vals["interval"],
					}
				},
			},
			{
				Name:     "Record Health Status",
				Verb:     "record",
				FormSpec: &healthStatusSpec,
				APIPath: func(ventureID, divisionID string) string {
					return fmt.Sprintf("/api/ventures/%s/divisions/%s/deploy/health-status", ventureID, divisionID)
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					return map[string]interface{}{
						"check_name": vals["check_name"],
						"status":     vals["status"],
						"details":    vals["details"],
					}
				},
			},
			{
				Name:     "Raise Incident",
				Verb:     "raise",
				FormSpec: &incidentSpec,
				APIPath: func(ventureID, divisionID string) string {
					return fmt.Sprintf("/api/ventures/%s/divisions/%s/deploy/incidents", ventureID, divisionID)
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					return map[string]interface{}{
						"title":    vals["title"],
						"severity": vals["severity"],
					}
				},
			},
			{
				Name:     "Diagnose Incident",
				Verb:     "diagnose",
				FormSpec: &diagnoseSpec,
				APIPath: func(ventureID, divisionID string) string {
					return fmt.Sprintf("/api/ventures/%s/divisions/%s/deploy/incidents/diagnose", ventureID, divisionID)
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					return map[string]interface{}{
						"incident_id": vals["incident_id"],
						"diagnosis":   vals["diagnosis"],
					}
				},
			},
			{
				Name:     "Apply Fix",
				Verb:     "fix",
				FormSpec: &fixSpec,
				APIPath: func(ventureID, divisionID string) string {
					return fmt.Sprintf("/api/ventures/%s/divisions/%s/deploy/incidents/fix", ventureID, divisionID)
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					return map[string]interface{}{
						"incident_id": vals["incident_id"],
						"fix_details": vals["fix_details"],
					}
				},
			},
		},
	}
}
