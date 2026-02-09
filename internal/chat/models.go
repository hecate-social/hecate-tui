package chat

import "strings"

// SwitchModel switches the active model by name.
func (m *Model) SwitchModel(name string) {
	for i, model := range m.models {
		if strings.EqualFold(model.Name, name) || strings.HasPrefix(strings.ToLower(model.Name), strings.ToLower(name)) {
			m.activeModel = i
			m.InjectSystemMessage("Switched to model: " + model.Name)
			return
		}
	}
	m.InjectSystemMessage("Model not found: " + name)
}

// CycleModel cycles to the next available model.
func (m *Model) CycleModel() {
	if len(m.models) > 0 {
		m.activeModel = (m.activeModel + 1) % len(m.models)
	}
}

// CycleModelReverse cycles to the previous available model.
func (m *Model) CycleModelReverse() {
	if len(m.models) > 0 {
		m.activeModel = (m.activeModel - 1 + len(m.models)) % len(m.models)
	}
}

// ActiveModelName returns the name of the currently active model.
func (m Model) ActiveModelName() string {
	if len(m.models) == 0 {
		return ""
	}
	if m.activeModel < len(m.models) {
		return m.models[m.activeModel].Name
	}
	return ""
}

// ActiveModelProvider returns the provider of the currently active model.
func (m Model) ActiveModelProvider() string {
	if len(m.models) == 0 {
		return ""
	}
	if m.activeModel < len(m.models) {
		return m.models[m.activeModel].Provider
	}
	return ""
}

// IsPaidProvider returns true if the active model uses a commercial provider.
func (m Model) IsPaidProvider() bool {
	provider := m.ActiveModelProvider()
	switch provider {
	case "anthropic", "openai", "google", "groq", "together":
		return true
	default:
		return false
	}
}
