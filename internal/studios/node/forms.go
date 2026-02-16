package node

import "github.com/hecate-social/hecate-tui/internal/ui"

// registerIdentitySpec returns the form for registering a node identity.
func registerIdentitySpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "node.register_identity",
		Title: "Register Identity",
		Fields: []ui.FieldSpec{
			{
				Key:         "mri",
				Label:       "MRI",
				Description: "Machine-Readable Identifier for this node",
				Placeholder: "mri:agent:io.hecate/mynode",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "public_key",
				Label:       "Public Key",
				Description: "Ed25519 or secp256k1 public key (hex or base64)",
				Placeholder: "ed25519:abc123...",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "key_type",
				Label:       "Key Type",
				Description: "Cryptographic key algorithm",
				FieldType:   ui.FieldSelect,
				Required:    true,
				Default:     "ed25519",
				Options:     []string{"ed25519", "secp256k1"},
			},
		},
	}
}

// updateIdentitySpec returns the form for updating node identity metadata.
func updateIdentitySpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "node.update_identity",
		Title: "Update Identity",
		Fields: []ui.FieldSpec{
			{
				Key:         "metadata",
				Label:       "Metadata",
				Description: "JSON metadata to attach to identity (optional)",
				Placeholder: "{\"display_name\": \"my node\"}",
				FieldType:   ui.FieldTextarea,
			},
		},
	}
}

// announceCapabilitySpec returns the form for announcing a node capability.
func announceCapabilitySpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "node.announce_capability",
		Title: "Announce Capability",
		Fields: []ui.FieldSpec{
			{
				Key:         "capability_mri",
				Label:       "Capability MRI",
				Description: "Unique identifier for this capability",
				Placeholder: "mri:capability:io.hecate/llm.chat",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "description",
				Label:       "Description",
				Description: "Human-readable description of the capability",
				Placeholder: "LLM chat completion service",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "tags",
				Label:       "Tags",
				Description: "Comma-separated tags (optional)",
				Placeholder: "llm, chat, ai",
				FieldType:   ui.FieldText,
			},
			{
				Key:         "demo_procedure",
				Label:       "Demo Procedure",
				Description: "WAMP procedure URI for demo invocation (optional)",
				Placeholder: "io.hecate.capabilities.llm.demo",
				FieldType:   ui.FieldText,
			},
		},
	}
}

// updateCapabilitySpec returns the form for updating an existing capability.
func updateCapabilitySpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "node.update_capability",
		Title: "Update Capability",
		Fields: []ui.FieldSpec{
			{
				Key:         "capability_mri",
				Label:       "Capability MRI",
				Description: "MRI of the capability to update",
				Placeholder: "mri:capability:io.hecate/llm.chat",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "description",
				Label:       "Description",
				Description: "Updated description (optional)",
				Placeholder: "Updated description...",
				FieldType:   ui.FieldText,
			},
			{
				Key:         "tags",
				Label:       "Tags",
				Description: "Updated comma-separated tags (optional)",
				Placeholder: "llm, chat, ai, updated",
				FieldType:   ui.FieldText,
			},
		},
	}
}

// connectNodeSpec returns the form for connecting to another mesh node.
func connectNodeSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "node.connect_node",
		Title: "Connect to Node",
		Fields: []ui.FieldSpec{
			{
				Key:         "target_node",
				Label:       "Target Node",
				Description: "MRI or address of the node to connect to",
				Placeholder: "mri:agent:io.hecate/peer-node",
				FieldType:   ui.FieldText,
				Required:    true,
			},
		},
	}
}

// endorseNodeSpec returns the form for endorsing a mesh peer.
func endorseNodeSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "node.endorse_node",
		Title: "Endorse Node",
		Fields: []ui.FieldSpec{
			{
				Key:         "target_node",
				Label:       "Target Node",
				Description: "MRI of the node to endorse",
				Placeholder: "mri:agent:io.hecate/trusted-peer",
				FieldType:   ui.FieldText,
				Required:    true,
			},
		},
	}
}

// subscribeTopicSpec returns the form for subscribing to a mesh topic.
func subscribeTopicSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "node.subscribe_topic",
		Title: "Subscribe to Topic",
		Fields: []ui.FieldSpec{
			{
				Key:         "topic",
				Label:       "Topic",
				Description: "Mesh topic URI to subscribe to",
				Placeholder: "io.hecate.events.capabilities",
				FieldType:   ui.FieldText,
				Required:    true,
			},
		},
	}
}

// grantUCANSpec returns the form for granting a UCAN token.
func grantUCANSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "node.grant_ucan",
		Title: "Grant UCAN",
		Fields: []ui.FieldSpec{
			{
				Key:         "audience",
				Label:       "Audience",
				Description: "DID or MRI of the token recipient",
				Placeholder: "did:key:z6Mk...",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "resource",
				Label:       "Resource",
				Description: "Resource URI to grant access to",
				Placeholder: "hecate:capabilities/*",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "ability",
				Label:       "Ability",
				Description: "Permission to grant on the resource",
				Placeholder: "invoke",
				FieldType:   ui.FieldText,
				Required:    true,
			},
		},
	}
}

// registerConnectorSpec returns the form for registering a security connector.
func registerConnectorSpec() ui.FormSpec {
	return ui.FormSpec{
		ID:    "node.register_connector",
		Title: "Register Connector",
		Fields: []ui.FieldSpec{
			{
				Key:         "name",
				Label:       "Name",
				Description: "Unique name for the connector",
				Placeholder: "my-external-connector",
				FieldType:   ui.FieldText,
				Required:    true,
			},
			{
				Key:         "allowed_routes",
				Label:       "Allowed Routes",
				Description: "Comma-separated route patterns (default: all)",
				Placeholder: "all",
				FieldType:   ui.FieldText,
				Default:     "all",
			},
		},
	}
}
