package client

import (
	"context"

	"github.com/hecate-social/hecate-tui/internal/llm"
)

// DaemonClient defines the interface for interacting with the hecate daemon.
// The concrete *Client satisfies this interface. Tests can provide mock implementations.
type DaemonClient interface {
	// Health & Identity
	GetHealth() (*Health, error)
	GetIdentity() (*Identity, error)

	// LLM
	ListModels() ([]llm.Model, error)
	GetLLMHealth() (*llm.LLMHealth, error)
	ChatStream(ctx context.Context, req llm.ChatRequest) (<-chan llm.ChatResponse, <-chan error)
	Chat(req llm.ChatRequest) (*llm.ChatResponse, error)

	// Providers
	ListProviders() (map[string]llm.Provider, error)
	AddProvider(name, pType, apiKey, url string) error
	RemoveProvider(name string) error
	ReloadProviders() ([]string, error)

	// Discovery
	DiscoverCapabilities(realm, tag string, limit int) ([]Capability, error)
	ListSubscriptions() ([]Subscription, error)

	// RPC
	RPCCall(procedure string, args interface{}) (*RPCResult, error)

	// Agents
	ListAgents() ([]Agent, error)
	GetAgent(agentID string) (*Agent, error)

	// Venture
	GetVenture() (*Venture, error)
	GetVentureByID(ventureID string) (*Venture, error)
	ListVentures() ([]Venture, error)
	ListAllVentures() ([]Venture, error)
	InitiateVenture(name, brief string) (*Venture, error)
	ArchiveVenture(ventureID, reason string) error
	RefineVision(ventureID string, params map[string]interface{}) error
	SubmitVision(ventureID, submittedBy string) error
	GetVentureTasks(ventureID string) (*VentureTaskList, error)

	// Departments (divisions)
	ListDepartments(ventureID string) ([]Department, error)
	GetDepartment(ventureID, departmentID string) (*Department, error)
	DepartmentCommand(path string, body map[string]interface{}) error

	// Pairing
	StartPairing() (*PairingStatus, error)
	GetPairingStatus() (*PairingStatus, error)
	CancelPairing() error

	// Telemetry
	GetTotalCost() (*CostSummary, error)
	GetCostByVenture(ventureID string) (*CostSummary, error)
}

// Verify at compile time that *Client implements DaemonClient.
var _ DaemonClient = (*Client)(nil)
