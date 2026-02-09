package applytorchsideeffects

import "encoding/json"

// TorchInitiatedMsg is the typed message for torch_initiated_v1 facts.
type TorchInitiatedMsg struct {
	TorchID     string `json:"torch_id"`
	Name        string `json:"name"`
	Brief       string `json:"brief"`
	InitiatedBy string `json:"initiated_by"`
	Status      int    `json:"status"`
	InitiatedAt int64  `json:"initiated_at"`
}

// HandleTorchInitiated converts raw fact data into a typed message.
// Called by app/facts.go when fact_type == "torch_initiated_v1".
func HandleTorchInitiated(data json.RawMessage) (TorchInitiatedMsg, error) {
	var msg TorchInitiatedMsg
	err := json.Unmarshal(data, &msg)
	return msg, err
}
