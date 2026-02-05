package llmtools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hecate-social/hecate-tui/internal/client"
)

// MeshClient is the interface for mesh operations.
// This allows the tools to work with the daemon client.
type MeshClient interface {
	DiscoverCapabilities(realm, tag string, limit int) ([]client.Capability, error)
	RPCCall(procedure string, args interface{}) (*client.RPCResult, error)
}

var meshClient MeshClient

// SetMeshClient sets the client for mesh tools.
func SetMeshClient(c MeshClient) {
	meshClient = c
}

// RegisterMeshTools adds mesh interaction tools to the registry.
func RegisterMeshTools(r *Registry) {
	r.Register(meshSearchTool(), meshSearchHandler)
	r.Register(meshCallTool(), meshCallHandler)
	r.Register(meshPublishTool(), meshPublishHandler)
}

// --- mesh_search ---

func meshSearchTool() Tool {
	params := NewObjectParameters()
	params.AddProperty("query", String("Search query or tag to filter capabilities"))
	params.AddProperty("realm", String("Realm to search in (optional)"))
	params.AddProperty("limit", Integer("Maximum number of results (default: 10)"))
	params.AddRequired("query")

	return Tool{
		Name:             "mesh_search",
		Description:      "Search the Hecate mesh for capabilities, agents, and services. Find what's available on the decentralized network.",
		Parameters:       params,
		Category:         CategoryMesh,
		RequiresApproval: false,
	}
}

type meshSearchArgs struct {
	Query string `json:"query"`
	Realm string `json:"realm"`
	Limit int    `json:"limit"`
}

func meshSearchHandler(ctx context.Context, args json.RawMessage) (string, error) {
	var a meshSearchArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if a.Query == "" {
		return "", fmt.Errorf("query is required")
	}

	if meshClient == nil {
		return "", fmt.Errorf("mesh client not configured - daemon connection required")
	}

	limit := a.Limit
	if limit <= 0 {
		limit = 10
	}

	// Use query as a tag filter
	capabilities, err := meshClient.DiscoverCapabilities(a.Realm, a.Query, limit)
	if err != nil {
		return "", fmt.Errorf("mesh search failed: %w", err)
	}

	if len(capabilities) == 0 {
		return fmt.Sprintf("No capabilities found matching: %s", a.Query), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d capabilities matching '%s':\n\n", len(capabilities), a.Query))

	for i, cap := range capabilities {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, cap.MRI))
		if cap.Description != "" {
			sb.WriteString(fmt.Sprintf("   Description: %s\n", cap.Description))
		}
		if len(cap.Tags) > 0 {
			sb.WriteString(fmt.Sprintf("   Tags: %s\n", strings.Join(cap.Tags, ", ")))
		}
		if cap.DemoProcedure != "" {
			sb.WriteString(fmt.Sprintf("   Demo: %s\n", cap.DemoProcedure))
		}
		sb.WriteString(fmt.Sprintf("   Agent: %s\n", truncateID(cap.AgentIdentity)))
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// --- mesh_call ---

func meshCallTool() Tool {
	params := NewObjectParameters()
	params.AddProperty("procedure", String("MRI of the procedure to call (e.g., 'mri:proc:hecate:llm.chat')"))
	params.AddProperty("args", ParameterSpec{
		Type:        "object",
		Description: "Arguments to pass to the procedure (JSON object)",
	})
	params.AddRequired("procedure")

	return Tool{
		Name:             "mesh_call",
		Description:      "Call a remote procedure on the Hecate mesh. Use mesh_search first to find available procedures.",
		Parameters:       params,
		Category:         CategoryMesh,
		RequiresApproval: true, // RPC calls can have side effects
	}
}

type meshCallArgs struct {
	Procedure string          `json:"procedure"`
	Args      json.RawMessage `json:"args"`
}

func meshCallHandler(ctx context.Context, args json.RawMessage) (string, error) {
	var a meshCallArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if a.Procedure == "" {
		return "", fmt.Errorf("procedure is required")
	}

	if meshClient == nil {
		return "", fmt.Errorf("mesh client not configured - daemon connection required")
	}

	// Parse args if provided
	var callArgs interface{}
	if len(a.Args) > 0 && string(a.Args) != "null" {
		if err := json.Unmarshal(a.Args, &callArgs); err != nil {
			return "", fmt.Errorf("invalid args JSON: %w", err)
		}
	}

	result, err := meshClient.RPCCall(a.Procedure, callArgs)
	if err != nil {
		return "", fmt.Errorf("RPC call failed: %w", err)
	}

	if result.Error != "" {
		return fmt.Sprintf("RPC Error: %s", result.Error), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("RPC Call: %s\n", a.Procedure))
	if result.Duration != "" {
		sb.WriteString(fmt.Sprintf("Duration: %s\n", result.Duration))
	}
	sb.WriteString("\nResult:\n")

	// Pretty-print the result JSON
	if len(result.Result) > 0 {
		var prettyResult interface{}
		if err := json.Unmarshal(result.Result, &prettyResult); err == nil {
			prettyJSON, err := json.MarshalIndent(prettyResult, "", "  ")
			if err == nil {
				sb.WriteString(string(prettyJSON))
			} else {
				sb.WriteString(string(result.Result))
			}
		} else {
			sb.WriteString(string(result.Result))
		}
	} else {
		sb.WriteString("(no result)")
	}

	return sb.String(), nil
}

// --- mesh_publish ---

func meshPublishTool() Tool {
	params := NewObjectParameters()
	params.AddProperty("topic", String("Topic to publish to (e.g., 'hecate.status')"))
	params.AddProperty("payload", ParameterSpec{
		Type:        "object",
		Description: "Data to publish (JSON object)",
	})
	params.AddRequired("topic", "payload")

	return Tool{
		Name:             "mesh_publish",
		Description:      "Publish a message to a topic on the Hecate mesh. Other agents subscribed to this topic will receive it.",
		Parameters:       params,
		Category:         CategoryMesh,
		RequiresApproval: true, // Publishing can have effects on other agents
	}
}

type meshPublishArgs struct {
	Topic   string          `json:"topic"`
	Payload json.RawMessage `json:"payload"`
}

func meshPublishHandler(ctx context.Context, args json.RawMessage) (string, error) {
	var a meshPublishArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if a.Topic == "" {
		return "", fmt.Errorf("topic is required")
	}

	if len(a.Payload) == 0 || string(a.Payload) == "null" {
		return "", fmt.Errorf("payload is required")
	}

	if meshClient == nil {
		return "", fmt.Errorf("mesh client not configured - daemon connection required")
	}

	// For now, use RPC to call a publish procedure
	// The daemon should expose a publish endpoint or procedure
	// This is a placeholder - actual implementation depends on daemon API
	publishArgs := map[string]interface{}{
		"topic":   a.Topic,
		"payload": json.RawMessage(a.Payload),
	}

	result, err := meshClient.RPCCall("mri:proc:hecate:pubsub.publish", publishArgs)
	if err != nil {
		// If the RPC fails, it might not be implemented yet
		return "", fmt.Errorf("publish failed (daemon may not support pubsub.publish): %w", err)
	}

	if result.Error != "" {
		return fmt.Sprintf("Publish Error: %s", result.Error), nil
	}

	return fmt.Sprintf("Successfully published to topic: %s", a.Topic), nil
}

// --- Helpers ---

// truncateID truncates long identity strings for display.
func truncateID(id string) string {
	if len(id) > 16 {
		return id[:8] + "..." + id[len(id)-8:]
	}
	return id
}
