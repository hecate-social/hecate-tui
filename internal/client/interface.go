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

	// Torch
	GetTorch() (*Torch, error)
	GetTorchByID(torchID string) (*Torch, error)
	ListTorches() ([]Torch, error)
	ListAllTorches() ([]Torch, error)
	InitiateTorch(name, brief string) (*Torch, error)
	ArchiveTorch(torchID, reason string) error
	RefineVision(torchID string, params map[string]interface{}) error
	SubmitVision(torchID, submittedBy string) error

	// Cartwheels
	ListCartwheels() ([]Cartwheel, error)
	GetCartwheel(cartwheelID string) (*Cartwheel, error)
	CartwheelCommand(path string, body map[string]interface{}) error

	// Pairing
	StartPairing() (*PairingStatus, error)
	GetPairingStatus() (*PairingStatus, error)
	CancelPairing() error

	// Telemetry
	GetTotalCost() (*CostSummary, error)
	GetCostByTorch(torchID string) (*CostSummary, error)
}

// Verify at compile time that *Client implements DaemonClient.
var _ DaemonClient = (*Client)(nil)
