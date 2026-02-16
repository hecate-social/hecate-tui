package devops

import "github.com/hecate-social/hecate-tui/internal/ui"

// --- Venture forms ---

func initiateVentureSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "devops.initiate_venture",
		Title: "Initiate Venture",
		Fields: []ui.FieldSpec{
			{
				Key:         "name",
				Label:       "Name",
				Description: "Venture name",
				Placeholder: "my-venture",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "brief",
				Label:       "Brief",
				Description: "Short description of the venture",
				Placeholder: "A revolutionary new product...",
				FieldType:   ui.FieldText,
			},
		},
	}
}

func archiveVentureSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "devops.archive_venture",
		Title: "Archive Venture",
		Fields: []ui.FieldSpec{
			{
				Key:         "reason",
				Label:       "Reason",
				Description: "Why is this venture being archived?",
				Placeholder: "Completed / superseded / abandoned",
				FieldType:   ui.FieldText,
			},
		},
	}
}

func refineVisionSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "devops.refine_vision",
		Title: "Refine Vision",
		Fields: []ui.FieldSpec{
			{
				Key:         "brief",
				Label:       "Brief",
				Description: "Updated description (leave empty to keep current)",
				Placeholder: "Updated vision statement...",
				FieldType:   ui.FieldText,
			},
			{
				Key:         "repo_url",
				Label:       "Repo URL",
				Description: "Source repository URL",
				Placeholder: "https://github.com/org/repo",
				FieldType:   ui.FieldText,
			},
		},
	}
}

// --- Division forms ---

func initiateDivisionSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "devops.initiate_division",
		Title: "Initiate Division",
		Fields: []ui.FieldSpec{
			{
				Key:         "context_name",
				Label:       "Context Name",
				Description: "Bounded context name for this division",
				Placeholder: "order_fulfillment",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "description",
				Label:       "Description",
				Description: "What this division is responsible for",
				Placeholder: "Handles order processing and delivery...",
				FieldType:   ui.FieldText,
			},
		},
	}
}

func startPhaseSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "devops.start_phase",
		Title: "Start Phase",
		Fields: []ui.FieldSpec{
			{
				Key:         "division_id",
				Label:       "Division ID",
				Description: "ID of the target division",
				Placeholder: "div-abc123",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:       "phase",
				Label:     "Phase",
				Description: "Lifecycle phase to start",
				FieldType: ui.FieldSelect,
				Required:  true,
				Options:   []string{"dna", "anp", "tni", "dno"},
			},
		},
	}
}

func archiveDivisionSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "devops.archive_division",
		Title: "Archive Division",
		Fields: []ui.FieldSpec{
			{
				Key:         "division_id",
				Label:       "Division ID",
				Description: "ID of the division to archive",
				Placeholder: "div-abc123",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "reason",
				Label:       "Reason",
				Description: "Why is this division being archived?",
				Placeholder: "Completed / merged / abandoned",
				FieldType:   ui.FieldText,
			},
		},
	}
}

// --- Design (DnA) forms ---

func designAggregateSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "devops.design_aggregate",
		Title: "Design Aggregate",
		Fields: []ui.FieldSpec{
			{
				Key:         "division_id",
				Label:       "Division ID",
				Description: "ID of the target division",
				Placeholder: "div-abc123",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "aggregate_name",
				Label:       "Aggregate Name",
				Description: "Name for the aggregate / dossier",
				Placeholder: "order",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "description",
				Label:       "Description",
				Description: "What this aggregate represents",
				Placeholder: "Manages order lifecycle...",
				FieldType:   ui.FieldText,
			},
			{
				Key:         "stream_prefix",
				Label:       "Stream Prefix",
				Description: "Event stream prefix (auto-derived if empty)",
				Placeholder: "order-",
				FieldType:   ui.FieldText,
			},
		},
	}
}

func designEventSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "devops.design_event",
		Title: "Design Event",
		Fields: []ui.FieldSpec{
			{
				Key:         "division_id",
				Label:       "Division ID",
				Description: "ID of the target division",
				Placeholder: "div-abc123",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "event_name",
				Label:       "Event Name",
				Description: "Domain event name (e.g. order_placed_v1)",
				Placeholder: "order_placed_v1",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "description",
				Label:       "Description",
				Description: "What this event represents",
				Placeholder: "Emitted when a customer places an order",
				FieldType:   ui.FieldText,
			},
			{
				Key:         "aggregate_name",
				Label:       "Aggregate",
				Description: "Owning aggregate (optional)",
				Placeholder: "order",
				FieldType:   ui.FieldText,
			},
		},
	}
}

// --- Plan (AnP) forms ---

func planDeskSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "devops.plan_desk",
		Title: "Plan Desk",
		Fields: []ui.FieldSpec{
			{
				Key:         "division_id",
				Label:       "Division ID",
				Description: "ID of the target division",
				Placeholder: "div-abc123",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "desk_name",
				Label:       "Desk Name",
				Description: "Vertical slice name (e.g. register_user)",
				Placeholder: "register_user",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "description",
				Label:       "Description",
				Description: "What this desk handles",
				Placeholder: "Handles new user registration...",
				FieldType:   ui.FieldText,
			},
			{
				Key:       "department",
				Label:     "Department",
				Description: "Which department owns this desk",
				FieldType: ui.FieldSelect,
				Options:   []string{"cmd", "prj", "qry"},
			},
		},
	}
}

func planDependencySpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "devops.plan_dependency",
		Title: "Plan Dependency",
		Fields: []ui.FieldSpec{
			{
				Key:         "division_id",
				Label:       "Division ID",
				Description: "ID of the target division",
				Placeholder: "div-abc123",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "from_desk",
				Label:       "From Desk",
				Description: "Source desk name",
				Placeholder: "register_user",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "to_desk",
				Label:       "To Desk",
				Description: "Target desk name",
				Placeholder: "send_welcome_email",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "dep_type",
				Label:       "Dependency Type",
				Description: "Relationship type (optional)",
				Placeholder: "triggers",
				FieldType:   ui.FieldText,
			},
		},
	}
}

// --- Build (TnI) forms ---

func generateModuleSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "devops.generate_module",
		Title: "Generate Module",
		Fields: []ui.FieldSpec{
			{
				Key:         "division_id",
				Label:       "Division ID",
				Description: "ID of the target division",
				Placeholder: "div-abc123",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "module_name",
				Label:       "Module Name",
				Description: "Module to generate",
				Placeholder: "register_user",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:       "module_type",
				Label:     "Module Type",
				Description: "Type of module to generate",
				FieldType: ui.FieldSelect,
				Options:   []string{"aggregate", "command", "event", "handler", "projection"},
			},
		},
	}
}

func generateTestSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "devops.generate_test",
		Title: "Generate Test",
		Fields: []ui.FieldSpec{
			{
				Key:         "division_id",
				Label:       "Division ID",
				Description: "ID of the target division",
				Placeholder: "div-abc123",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "test_name",
				Label:       "Test Name",
				Description: "Test module name",
				Placeholder: "register_user_test",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "module_name",
				Label:       "Module Name",
				Description: "Module under test (optional)",
				Placeholder: "register_user",
				FieldType:   ui.FieldText,
			},
		},
	}
}

func runTestSuiteSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "devops.run_test_suite",
		Title: "Run Test Suite",
		Fields: []ui.FieldSpec{
			{
				Key:         "division_id",
				Label:       "Division ID",
				Description: "ID of the target division",
				Placeholder: "div-abc123",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "suite",
				Label:       "Suite",
				Description: "Test suite to run (empty = all)",
				Placeholder: "unit",
				FieldType:   ui.FieldText,
			},
		},
	}
}

func recordTestResultSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "devops.record_test_result",
		Title: "Record Test Result",
		Fields: []ui.FieldSpec{
			{
				Key:         "division_id",
				Label:       "Division ID",
				Description: "ID of the target division",
				Placeholder: "div-abc123",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:       "result",
				Label:     "Result",
				Description: "Test outcome",
				FieldType: ui.FieldSelect,
				Required:  true,
				Options:   []string{"pass", "fail", "skip"},
			},
			{
				Key:         "notes",
				Label:       "Notes",
				Description: "Additional details about the test run",
				Placeholder: "All 42 tests passed in 1.2s",
				FieldType:   ui.FieldText,
			},
		},
	}
}

// --- Ship (DnO) forms ---

func deployReleaseSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "devops.deploy_release",
		Title: "Deploy Release",
		Fields: []ui.FieldSpec{
			{
				Key:         "division_id",
				Label:       "Division ID",
				Description: "ID of the target division",
				Placeholder: "div-abc123",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "version",
				Label:       "Version",
				Description: "Release version tag",
				Placeholder: "v1.2.3",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:       "environment",
				Label:     "Environment",
				Description: "Target deployment environment",
				FieldType: ui.FieldSelect,
				Required:  true,
				Options:   []string{"staging", "production"},
			},
			{
				Key:         "notes",
				Label:       "Notes",
				Description: "Release notes (optional)",
				Placeholder: "Bug fixes and performance improvements",
				FieldType:   ui.FieldText,
			},
		},
	}
}

func stageRolloutSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "devops.stage_rollout",
		Title: "Stage Rollout",
		Fields: []ui.FieldSpec{
			{
				Key:         "division_id",
				Label:       "Division ID",
				Description: "ID of the target division",
				Placeholder: "div-abc123",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "version",
				Label:       "Version",
				Description: "Version to roll out",
				Placeholder: "v1.2.3",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:       "environment",
				Label:     "Environment",
				Description: "Target environment",
				FieldType: ui.FieldSelect,
				Required:  true,
				Options:   []string{"staging", "production"},
			},
			{
				Key:       "strategy",
				Label:     "Strategy",
				Description: "Rollout strategy",
				FieldType: ui.FieldSelect,
				Required:  true,
				Options:   []string{"rolling", "blue-green", "canary"},
			},
		},
	}
}

func registerHealthCheckSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "devops.register_health_check",
		Title: "Register Health Check",
		Fields: []ui.FieldSpec{
			{
				Key:         "division_id",
				Label:       "Division ID",
				Description: "ID of the target division",
				Placeholder: "div-abc123",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "check_name",
				Label:       "Check Name",
				Description: "Health check identifier",
				Placeholder: "api_liveness",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "endpoint",
				Label:       "Endpoint",
				Description: "URL or path to check",
				Placeholder: "/healthz",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "interval",
				Label:       "Interval",
				Description: "Check interval (e.g. 30s, 5m)",
				Placeholder: "30s",
				FieldType:   ui.FieldText,
				Default:     "30s",
			},
		},
	}
}

func recordHealthStatusSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "devops.record_health_status",
		Title: "Record Health Status",
		Fields: []ui.FieldSpec{
			{
				Key:         "division_id",
				Label:       "Division ID",
				Description: "ID of the target division",
				Placeholder: "div-abc123",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "check_name",
				Label:       "Check Name",
				Description: "Which health check this is for",
				Placeholder: "api_liveness",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:       "status",
				Label:     "Status",
				Description: "Health check result",
				FieldType: ui.FieldSelect,
				Required:  true,
				Options:   []string{"healthy", "degraded", "unhealthy"},
			},
			{
				Key:         "details",
				Label:       "Details",
				Description: "Additional context",
				Placeholder: "Response time 120ms",
				FieldType:   ui.FieldText,
			},
		},
	}
}

func raiseIncidentSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "devops.raise_incident",
		Title: "Raise Incident",
		Fields: []ui.FieldSpec{
			{
				Key:         "division_id",
				Label:       "Division ID",
				Description: "ID of the affected division",
				Placeholder: "div-abc123",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "title",
				Label:       "Title",
				Description: "Brief incident summary",
				Placeholder: "API timeout on /orders endpoint",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:       "severity",
				Label:     "Severity",
				Description: "Incident severity level",
				FieldType: ui.FieldSelect,
				Required:  true,
				Options:   []string{"low", "medium", "high", "critical"},
			},
		},
	}
}

func diagnoseIncidentSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "devops.diagnose_incident",
		Title: "Diagnose Incident",
		Fields: []ui.FieldSpec{
			{
				Key:         "division_id",
				Label:       "Division ID",
				Description: "ID of the affected division",
				Placeholder: "div-abc123",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "incident_id",
				Label:       "Incident ID",
				Description: "ID of the incident to diagnose",
				Placeholder: "inc-abc123",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "diagnosis",
				Label:       "Diagnosis",
				Description: "Root cause analysis",
				Placeholder: "Database connection pool exhaustion due to...",
				FieldType:   ui.FieldTextarea,
				Required:    true,
			},
		},
	}
}

func applyFixSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "devops.apply_fix",
		Title: "Apply Fix",
		Fields: []ui.FieldSpec{
			{
				Key:         "division_id",
				Label:       "Division ID",
				Description: "ID of the affected division",
				Placeholder: "div-abc123",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "incident_id",
				Label:       "Incident ID",
				Description: "ID of the incident to fix",
				Placeholder: "inc-abc123",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "fix_details",
				Label:       "Fix Details",
				Description: "What was done to resolve the incident",
				Placeholder: "Increased connection pool size to 50, deployed hotfix v1.2.4",
				FieldType:   ui.FieldTextarea,
				Required:    true,
			},
		},
	}
}
