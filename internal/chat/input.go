package chat

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// SetInputVisible shows/hides the textarea (Insert mode).
func (m *Model) SetInputVisible(visible bool) {
	m.inputVisible = visible
	if visible {
		m.input.Focus()
	} else {
		m.input.Blur()
	}
	m.resize()
}

// SendCurrentInput sends the current textarea content as a user message.
func (m *Model) SendCurrentInput() tea.Cmd {
	content := strings.TrimSpace(m.input.Value())
	if content == "" || m.streaming {
		return nil
	}

	m.messages = append(m.messages, Message{
		Role:    "user",
		Content: content,
		Time:    time.Now(),
	})
	m.input.Reset()
	m.streaming = true
	m.streamBuf.Reset()
	m.streamStart = time.Now()
	m.lastTokenCount = 0
	m.lastDuration = 0
	m.lastSpeed = 0
	m.err = nil
	m.thinkingFrame = 0
	m.updateStreamingMessage() // Show thinking animation immediately

	return tea.Batch(
		m.sendMessage(),
		m.thinkingTick(),
	)
}

// InsertNewline adds a newline at the cursor position in the input.
func (m *Model) InsertNewline() {
	m.input.InsertString("\n")
}

// InputLen returns the current input length in characters.
func (m Model) InputLen() int {
	return len(m.input.Value())
}

// InputValue returns the current input text.
func (m Model) InputValue() string {
	return m.input.Value()
}

// SetInputValue sets the input text (for history navigation).
func (m *Model) SetInputValue(v string) {
	m.input.SetValue(v)
}
