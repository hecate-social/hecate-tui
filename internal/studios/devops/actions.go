package devops

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// actionResultMsg carries the outcome of executing an action.
type actionResultMsg struct {
	success bool
	message string
}

// executeAction dispatches the action to the daemon API.
func (s *Studio) executeAction(action Action, values map[string]string) tea.Cmd {
	return func() tea.Msg {
		ventureID := s.ventureID
		divisionID := values["division_id"] // may be empty for venture-level actions

		path := action.APIPath(ventureID, divisionID)
		var body map[string]interface{}
		if action.BodyBuilder != nil {
			body = action.BodyBuilder(values)
		}

		err := s.ctx.Client.DepartmentCommand(path, body)
		if err != nil {
			return actionResultMsg{
				success: false,
				message: fmt.Sprintf("%s failed: %s", action.Name, err.Error()),
			}
		}
		return actionResultMsg{
			success: true,
			message: fmt.Sprintf("%s completed successfully", action.Name),
		}
	}
}
