package applyventuresideeffects

import "encoding/json"

// VentureInitiatedMsg is the typed message for torch_initiated_v1 facts.
type VentureInitiatedMsg struct {
	VentureID   string `json:"torch_id"`
	Name        string `json:"name"`
	Brief       string `json:"brief"`
	InitiatedBy string `json:"initiated_by"`
	Status      int    `json:"status"`
	InitiatedAt int64  `json:"initiated_at"`
}

// HandleVentureInitiated converts raw fact data into a typed message.
// Called by app/facts.go when fact_type == "torch_initiated_v1".
func HandleVentureInitiated(data json.RawMessage) (VentureInitiatedMsg, error) {
	var msg VentureInitiatedMsg
	err := json.Unmarshal(data, &msg)
	return msg, err
}
