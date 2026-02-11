package applyventuresideeffects

import "encoding/json"

// VentureInitiatedMsg is the typed message for venture_setup_v1 facts.
type VentureInitiatedMsg struct {
	VentureID   string `json:"venture_id"`
	Name        string `json:"name"`
	Brief       string `json:"brief"`
	InitiatedBy string `json:"initiated_by"`
	Status      int    `json:"status"`
	InitiatedAt int64  `json:"initiated_at"`
}

// HandleVentureInitiated converts raw fact data into a typed message.
// Called by app/facts.go when fact_type == "venture_setup_v1".
func HandleVentureInitiated(data json.RawMessage) (VentureInitiatedMsg, error) {
	var msg VentureInitiatedMsg
	err := json.Unmarshal(data, &msg)
	return msg, err
}
