package node

import tea "github.com/charmbracelet/bubbletea"

// actionResultMsg carries the result of an API command back to the Update loop.
type actionResultMsg struct {
	success bool
	message string
}

// executeAction dispatches the given action to the daemon API and returns a
// Bubble Tea Cmd that resolves to an actionResultMsg.
func (s *Studio) executeAction(action Action, values map[string]string) tea.Cmd {
	return func() tea.Msg {
		path := action.APIPath(values)

		var body map[string]interface{}
		if action.BodyBuilder != nil {
			body = action.BodyBuilder(values)
		}

		err := s.ctx.Client.DepartmentCommand(path, body)
		if err != nil {
			return actionResultMsg{success: false, message: err.Error()}
		}
		return actionResultMsg{success: true, message: action.Name + " completed"}
	}
}
