package editor

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Mode represents editor mode
type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
)

// Model is the editor model
type Model struct {
	// File info
	filepath    string
	filename    string
	modified    bool

	// Content
	textarea    textarea.Model
	lines       []string
	cursorLine  int
	cursorCol   int

	// Display
	width       int
	height      int
	scrollOffset int
	visibleLines int

	// Mode
	mode        Mode
	focused     bool

	// Syntax
	lang        Language
	highlighter *Highlighter

	// Messages
	message     string
	messageErr  bool
}

// New creates a new editor
func New() Model {
	ta := textarea.New()
	ta.Placeholder = "Start typing..."
	ta.ShowLineNumbers = false
	ta.Focus()

	return Model{
		textarea: ta,
		lines:    []string{""},
		mode:     ModeInsert, // Start in insert mode for simplicity
		lang:     LangPlain,
	}
}

// NewWithFile creates an editor with a file loaded
func NewWithFile(path string) (Model, error) {
	m := New()

	data, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}

	m.filepath = path
	m.filename = filepath.Base(path)
	m.lang = DetectLanguage(path)
	m.highlighter = NewHighlighter(m.lang)

	content := string(data)
	m.textarea.SetValue(content)
	m.lines = strings.Split(content, "\n")

	return m, nil
}

// Init initializes the editor
func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Clear message on any keypress
		m.message = ""
		m.messageErr = false

		switch msg.String() {
		case "ctrl+s":
			// Save file
			return m, m.save()

		case "ctrl+q", "esc":
			if m.modified {
				m.message = "Unsaved changes! Press Ctrl+Q again to quit without saving."
				m.messageErr = true
				// For now just mark as not modified to allow quit
				// In a full impl, we'd track double-press
				m.modified = false
				return m, nil
			}
			return m, tea.Quit

		case "ctrl+g":
			// Go to line (placeholder)
			m.message = "Go to line not yet implemented"
			return m, nil
		}

		// Update textarea
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)

		// Track modifications
		m.modified = true
		m.lines = strings.Split(m.textarea.Value(), "\n")
		m.cursorLine = m.textarea.Line()
		m.cursorCol = m.textarea.LineInfo().ColumnOffset

		// Update scroll
		m.updateScroll()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.visibleLines = m.height - 4 // Title + status + padding
		m.textarea.SetWidth(m.width - 8)
		m.textarea.SetHeight(m.visibleLines)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) updateScroll() {
	// Keep cursor visible
	if m.cursorLine < m.scrollOffset {
		m.scrollOffset = m.cursorLine
	}
	if m.cursorLine >= m.scrollOffset+m.visibleLines {
		m.scrollOffset = m.cursorLine - m.visibleLines + 1
	}
}

type saveResultMsg struct {
	err error
}

func (m Model) save() tea.Cmd {
	return func() tea.Msg {
		if m.filepath == "" {
			return saveResultMsg{err: os.ErrNotExist}
		}

		content := m.textarea.Value()
		err := os.WriteFile(m.filepath, []byte(content), 0644)
		return saveResultMsg{err: err}
	}
}

// View renders the editor
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	var b strings.Builder

	// Title bar
	b.WriteString(m.renderTitleBar())
	b.WriteString("\n")

	// Editor content with line numbers
	b.WriteString(m.renderContent())
	b.WriteString("\n")

	// Status bar
	b.WriteString(m.renderStatusBar())

	return b.String()
}

func (m Model) renderTitleBar() string {
	title := "Quick Edit"
	if m.filename != "" {
		title = m.filename
	}

	if m.modified {
		title += TitleModifiedStyle.Render(" [+]")
	}

	// Language indicator
	langStr := ""
	if m.lang != LangPlain {
		langStr = " [" + string(m.lang) + "]"
	}

	left := TitleBarStyle.Render(" " + title + langStr)
	right := TitleBarStyle.Render(" Ctrl+S save | Ctrl+Q quit ")

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	return left + TitleBarStyle.Render(strings.Repeat(" ", gap)) + right
}

func (m Model) renderContent() string {
	if m.visibleLines <= 0 {
		return ""
	}

	// For the quick editor, we just use the textarea directly
	// A more sophisticated version would render lines with syntax highlighting

	style := EditorStyle
	if m.focused {
		style = EditorActiveStyle
	}

	return style.Width(m.width - 4).Height(m.visibleLines).Render(m.textarea.View())
}

func (m Model) renderStatusBar() string {
	// Mode indicator
	modeStr := StatusModeStyle.Render(" NORMAL ")
	if m.mode == ModeInsert {
		modeStr = StatusInsertStyle.Render(" INSERT ")
	}

	// Position
	posStr := StatusPosStyle.Render(
		" Ln " + itoa(m.cursorLine+1) + ", Col " + itoa(m.cursorCol+1) + " ",
	)

	// Message or help
	var msgStr string
	if m.message != "" {
		if m.messageErr {
			msgStr = ErrorStyle.Render(m.message)
		} else {
			msgStr = SuccessStyle.Render(m.message)
		}
	}

	// Lines count
	linesStr := StatusPosStyle.Render(" " + itoa(len(m.lines)) + " lines ")

	left := modeStr + posStr
	right := linesStr

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right) - lipgloss.Width(msgStr)
	if gap < 0 {
		gap = 0
	}

	middle := msgStr + strings.Repeat(" ", gap)

	return StatusBarStyle.Width(m.width).Render(left + middle + right)
}

// SetSize updates the editor size
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.visibleLines = height - 4
	m.textarea.SetWidth(width - 8)
	m.textarea.SetHeight(m.visibleLines)
}

// Focus activates the editor
func (m *Model) Focus() {
	m.focused = true
	m.textarea.Focus()
}

// Blur deactivates the editor
func (m *Model) Blur() {
	m.focused = false
	m.textarea.Blur()
}

// SetContent sets the editor content
func (m *Model) SetContent(content string) {
	m.textarea.SetValue(content)
	m.lines = strings.Split(content, "\n")
	m.modified = false
}

// GetContent returns the editor content
func (m Model) GetContent() string {
	return m.textarea.Value()
}

// SetFilepath sets the file path
func (m *Model) SetFilepath(path string) {
	m.filepath = path
	m.filename = filepath.Base(path)
	m.lang = DetectLanguage(path)
	m.highlighter = NewHighlighter(m.lang)
}

// IsModified returns whether the content has been modified
func (m Model) IsModified() bool {
	return m.modified
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
