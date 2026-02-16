package node

import "github.com/hecate-social/hecate-tui/internal/ui"

// actionView tracks navigation state for the categories/actions/form overlay.
type actionView int

const (
	actionViewNone       actionView = iota // No action overlay — dashboard visible
	actionViewCategories                   // Selecting a category
	actionViewActions                      // Selecting an action within a category
	actionViewForm                         // Filling in a form for the selected action
)

// Category groups related node operations.
type Category struct {
	Name    string
	Icon    string
	Actions []Action
}

// Action describes a single node operation that may require a form.
type Action struct {
	Name        string
	Verb        string
	FormSpec    *ui.FormSpec                              // nil = confirm-only (no form)
	APIPath     func(vals map[string]string) string       // builds the endpoint path
	BodyBuilder func(vals map[string]string) map[string]interface{} // builds POST body
}

// nodeCategories returns the full menu of node operations.
func nodeCategories() []Category {
	return []Category{
		identityCategory(),
		capabilitiesCategory(),
		meshCategory(),
		subscriptionsCategory(),
		securityCategory(),
	}
}

func identityCategory() Category {
	registerSpec := registerIdentitySpec()
	updateSpec := updateIdentitySpec()

	return Category{
		Name: "Identity",
		Icon: "\U0001F464",
		Actions: []Action{
			{
				Name: "Register Identity",
				Verb: "register",
				FormSpec: &registerSpec,
				APIPath: func(_ map[string]string) string {
					return "/api/identity/register"
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					body := map[string]interface{}{
						"mri":        vals["mri"],
						"public_key": vals["public_key"],
						"key_type":   vals["key_type"],
					}
					return body
				},
			},
			{
				Name: "Update Identity",
				Verb: "update",
				FormSpec: &updateSpec,
				APIPath: func(_ map[string]string) string {
					return "/api/identity/update"
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					body := map[string]interface{}{}
					if vals["metadata"] != "" {
						body["metadata"] = vals["metadata"]
					}
					return body
				},
			},
		},
	}
}

func capabilitiesCategory() Category {
	announceSpec := announceCapabilitySpec()
	updateSpec := updateCapabilitySpec()

	return Category{
		Name: "Capabilities",
		Icon: "\u2B50",
		Actions: []Action{
			{
				Name: "Announce Capability",
				Verb: "announce",
				FormSpec: &announceSpec,
				APIPath: func(_ map[string]string) string {
					return "/api/capabilities/announce"
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					body := map[string]interface{}{
						"capability_mri": vals["capability_mri"],
						"description":    vals["description"],
					}
					if vals["tags"] != "" {
						body["tags"] = vals["tags"]
					}
					if vals["demo_procedure"] != "" {
						body["demo_procedure"] = vals["demo_procedure"]
					}
					return body
				},
			},
			{
				Name: "Update Capability",
				Verb: "update",
				FormSpec: &updateSpec,
				APIPath: func(vals map[string]string) string {
					mri := vals["capability_mri"]
					return "/api/capabilities/" + mri + "/update"
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					body := map[string]interface{}{}
					if vals["description"] != "" {
						body["description"] = vals["description"]
					}
					if vals["tags"] != "" {
						body["tags"] = vals["tags"]
					}
					return body
				},
			},
			{
				Name: "Retract Capability",
				Verb: "retract",
				FormSpec: nil, // confirm-only
				APIPath: func(_ map[string]string) string {
					return "/api/capabilities/retract"
				},
				BodyBuilder: nil,
			},
		},
	}
}

func meshCategory() Category {
	connectSpec := connectNodeSpec()
	endorseSpec := endorseNodeSpec()

	return Category{
		Name: "Mesh",
		Icon: "\U0001F310",
		Actions: []Action{
			{
				Name: "Connect to Node",
				Verb: "connect",
				FormSpec: &connectSpec,
				APIPath: func(_ map[string]string) string {
					return "/api/mesh/connect"
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					return map[string]interface{}{
						"target_node": vals["target_node"],
					}
				},
			},
			{
				Name:    "Disconnect",
				Verb:    "disconnect",
				FormSpec: nil,
				APIPath: func(_ map[string]string) string {
					return "/api/mesh/disconnect"
				},
				BodyBuilder: nil,
			},
			{
				Name: "Endorse Node",
				Verb: "endorse",
				FormSpec: &endorseSpec,
				APIPath: func(_ map[string]string) string {
					return "/api/mesh/endorse"
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					return map[string]interface{}{
						"target_node": vals["target_node"],
					}
				},
			},
			{
				Name:    "Revoke Endorsement",
				Verb:    "revoke_endorsement",
				FormSpec: nil,
				APIPath: func(_ map[string]string) string {
					return "/api/mesh/endorsements/revoke"
				},
				BodyBuilder: nil,
			},
		},
	}
}

func subscriptionsCategory() Category {
	subSpec := subscribeTopicSpec()

	return Category{
		Name: "Subscriptions",
		Icon: "\U0001F514",
		Actions: []Action{
			{
				Name: "Subscribe to Topic",
				Verb: "subscribe",
				FormSpec: &subSpec,
				APIPath: func(_ map[string]string) string {
					return "/api/subscriptions/subscribe"
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					return map[string]interface{}{
						"topic": vals["topic"],
					}
				},
			},
			{
				Name:    "Unsubscribe",
				Verb:    "unsubscribe",
				FormSpec: nil,
				APIPath: func(_ map[string]string) string {
					return "/api/subscriptions/unsubscribe"
				},
				BodyBuilder: nil,
			},
		},
	}
}

func securityCategory() Category {
	grantSpec := grantUCANSpec()
	registerConnSpec := registerConnectorSpec()

	return Category{
		Name: "Security",
		Icon: "\U0001F512",
		Actions: []Action{
			{
				Name: "Grant UCAN",
				Verb: "grant_ucan",
				FormSpec: &grantSpec,
				APIPath: func(_ map[string]string) string {
					return "/api/security/ucan/grant"
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					return map[string]interface{}{
						"audience": vals["audience"],
						"resource": vals["resource"],
						"ability":  vals["ability"],
					}
				},
			},
			{
				Name:    "Revoke UCAN",
				Verb:    "revoke_ucan",
				FormSpec: nil,
				APIPath: func(_ map[string]string) string {
					return "/api/security/ucan/revoke"
				},
				BodyBuilder: nil,
			},
			{
				Name: "Register Connector",
				Verb: "register_connector",
				FormSpec: &registerConnSpec,
				APIPath: func(_ map[string]string) string {
					return "/api/security/connectors/register"
				},
				BodyBuilder: func(vals map[string]string) map[string]interface{} {
					body := map[string]interface{}{
						"name": vals["name"],
					}
					if vals["allowed_routes"] != "" {
						body["allowed_routes"] = vals["allowed_routes"]
					}
					return body
				},
			},
			{
				Name:    "Activate Connector",
				Verb:    "activate_connector",
				FormSpec: nil,
				APIPath: func(_ map[string]string) string {
					return "/api/security/connectors/activate"
				},
				BodyBuilder: nil,
			},
			{
				Name:    "Suspend Connector",
				Verb:    "suspend_connector",
				FormSpec: nil,
				APIPath: func(_ map[string]string) string {
					return "/api/security/connectors/suspend"
				},
				BodyBuilder: nil,
			},
			{
				Name:    "Revoke Connector",
				Verb:    "revoke_connector",
				FormSpec: nil,
				APIPath: func(_ map[string]string) string {
					return "/api/security/connectors/revoke"
				},
				BodyBuilder: nil,
			},
		},
	}
}
