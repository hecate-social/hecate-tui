package pair

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

// PairingState represents the current step in the pairing flow.
type PairingState int

const (
	StateIdle PairingState = iota
	StateStarting
	StateWaiting
	StatePaired
	StateError
)

// Model is the Pair mode overlay — an inline wizard for realm pairing.
type Model struct {
	client  *client.Client
	theme   *theme.Theme
	styles  *theme.Styles
	width   int
	height  int
	spinner spinner.Model

	// Pairing state
	state        PairingState
	identity     *client.Identity
	pairingCode  string
	realmURL     string
	errorMessage string
}

// Messages for the pairing flow.
type identityMsg struct {
	identity *client.Identity
	err      error
}

type pairingStartedMsg struct {
	code     string
	realmURL string
	err      error
}

type pairingStatusMsg struct {
	status string
	err    error
}

type pairingPollMsg struct{}

// New creates a Pair mode model.
func New(c *client.Client, t *theme.Theme, s *theme.Styles) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(t.Primary)

	return Model{
		client:  c,
		theme:   t,
		styles:  s,
		spinner: sp,
		state:   StateIdle,
	}
}

// Init starts the pair mode — fetch identity to check current state.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchIdentity,
	)
}

// Update handles messages routed from the app.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case identityMsg:
		if msg.err == nil {
			m.identity = msg.identity
			if m.identity != nil && m.identity.Identity != "" {
				realm := parseRealm(m.identity.Identity)
				if realm != "" {
					m.state = StatePaired
				}
			}
		}

	case pairingStartedMsg:
		if msg.err != nil {
			m.state = StateError
			m.errorMessage = msg.err.Error()
		} else {
			m.state = StateWaiting
			m.pairingCode = msg.code
			m.realmURL = msg.realmURL
			return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
				return pairingPollMsg{}
			})
		}

	case pairingStatusMsg:
		if msg.err != nil {
			m.state = StateError
			m.errorMessage = msg.err.Error()
		} else {
			switch msg.status {
			case "paired":
				m.state = StatePaired
				m.pairingCode = ""
				return m, m.fetchIdentity
			case "waiting":
				return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
					return pairingPollMsg{}
				})
			case "expired":
				m.state = StateError
				m.errorMessage = "Pairing session expired"
				m.pairingCode = ""
			default:
				m.state = StateIdle
				m.pairingCode = ""
			}
		}

	case pairingPollMsg:
		if m.state == StateWaiting {
			return m, m.checkPairingStatus
		}
	}

	return m, tea.Batch(cmds...)
}

// HandleKey processes a keypress in Pair mode. Returns true if consumed.
func (m *Model) HandleKey(key string, msg tea.KeyMsg) (bool, tea.Cmd) {
	switch key {
	case "p":
		if m.state == StateIdle || m.state == StateError || m.state == StatePaired {
			m.state = StateStarting
			m.errorMessage = ""
			return true, m.startPairing
		}
		return true, nil
	case "c":
		if m.state == StateWaiting {
			m.state = StateIdle
			m.pairingCode = ""
			return true, m.cancelPairing
		}
		return true, nil
	case "r":
		return true, m.fetchIdentity
	case "esc":
		if m.state == StateWaiting {
			m.state = StateIdle
			m.pairingCode = ""
			return true, m.cancelPairing
		}
		// Let app handle Esc to exit Pair mode
		return false, nil
	}
	return true, nil
}

// SetSize updates the pair panel dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// StatusHints returns contextual hints for the status bar based on pairing state.
func (m Model) StatusHints() string {
	switch m.state {
	case StateWaiting:
		return "c:cancel  Esc:back"
	case StatePaired:
		return "p:re-pair  r:refresh  Esc:back"
	case StateError:
		return "p:retry  Esc:back"
	default:
		return "p:pair  Esc:back"
	}
}

// View renders the pair overlay panel.
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	var b strings.Builder

	switch m.state {
	case StateStarting:
		b.WriteString(m.renderStarting())
	case StateWaiting:
		b.WriteString(m.renderWaiting())
	case StatePaired:
		b.WriteString(m.renderPaired())
	case StateError:
		b.WriteString(m.renderError())
	default:
		b.WriteString(m.renderIdle())
	}

	return m.wrapPanel(b.String())
}

func (m Model) renderStarting() string {
	s := m.styles
	var b strings.Builder

	b.WriteString(s.CardTitle.Render("Pair with Realm"))
	b.WriteString("\n\n")
	b.WriteString(m.spinner.View() + " " + s.Subtle.Render("Starting pairing session..."))

	return b.String()
}

func (m Model) renderWaiting() string {
	s := m.styles
	t := m.theme
	var b strings.Builder

	b.WriteString(s.CardTitle.Render("Pair with Realm"))
	b.WriteString("\n\n")

	// Status
	waitingStyle := lipgloss.NewStyle().Foreground(t.Warning).Bold(true)
	b.WriteString(waitingStyle.Render("Waiting for confirmation..."))
	b.WriteString("\n\n")

	// Code display
	if m.pairingCode != "" {
		codeLabel := s.Subtle.Align(lipgloss.Center).Render("Enter this code on the realm:")
		codeDisplay := lipgloss.NewStyle().
			Foreground(t.Primary).
			Bold(true).
			Align(lipgloss.Center).
			Render(m.pairingCode)

		codeBox := lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(t.Primary).
			Padding(1, 4).
			Align(lipgloss.Center).
			Render(codeLabel + "\n\n" + codeDisplay)

		contentWidth := m.width - 8
		if contentWidth < 20 {
			contentWidth = 20
		}
		b.WriteString(lipgloss.PlaceHorizontal(contentWidth, lipgloss.Center, codeBox))
		b.WriteString("\n\n")
	}

	// Steps
	steps := []string{
		"Go to " + lipgloss.NewStyle().Foreground(t.Secondary).Underline(true).Render(m.realmURL),
		"Sign in to your account",
		"Enter the code shown above",
		"Confirm the pairing",
	}

	for i, step := range steps {
		num := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).
			Render(fmt.Sprintf("%d.", i+1))
		b.WriteString(num + " " + step + "\n")
	}

	b.WriteString("\n")
	b.WriteString(m.spinner.View() + " " + s.Subtle.Render("Polling for confirmation..."))
	b.WriteString("\n\n")
	b.WriteString(s.Subtle.Italic(true).Render("Press [c] or [Esc] to cancel"))

	return b.String()
}

func (m Model) renderPaired() string {
	s := m.styles
	t := m.theme
	var b strings.Builder

	b.WriteString(s.CardTitle.Render("Pair with Realm"))
	b.WriteString("\n\n")

	// Status
	pairedStyle := lipgloss.NewStyle().Foreground(t.Success).Bold(true)
	b.WriteString(pairedStyle.Render("Paired"))
	b.WriteString("\n\n")

	// Identity info
	if m.identity != nil {
		b.WriteString(s.Bold.Render("Identity"))
		b.WriteString("\n")

		mri := m.identity.Identity
		b.WriteString("  ")
		b.WriteString(s.CardLabel.Render("MRI: "))
		b.WriteString(lipgloss.NewStyle().Foreground(t.Primary).Render(mri))
		b.WriteString("\n")

		realm := parseRealm(mri)
		if realm != "" {
			b.WriteString("  ")
			b.WriteString(s.CardLabel.Render("Realm: "))
			b.WriteString(s.CardValue.Render(realm))
			b.WriteString("\n")
		}

		if m.identity.CreatedAt != "" {
			b.WriteString("  ")
			b.WriteString(s.CardLabel.Render("Since: "))
			b.WriteString(s.CardValue.Render(m.identity.CreatedAt))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// Connected info box
	contentWidth := m.width - 12
	if contentWidth < 20 {
		contentWidth = 20
	}
	connectedBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Border).
		Padding(1, 2).
		Width(contentWidth).
		Render(lipgloss.JoinVertical(lipgloss.Left,
			pairedStyle.Render("Connected to Macula Mesh"),
			"",
			s.Subtle.Italic(true).Render("Your agent can now:"),
			"  "+s.StatusOK.Render("*")+" Discover capabilities on the mesh",
			"  "+s.StatusOK.Render("*")+" Announce your own capabilities",
			"  "+s.StatusOK.Render("*")+" Make and receive RPC calls",
			"  "+s.StatusOK.Render("*")+" Participate in reputation tracking",
		))
	b.WriteString(connectedBox)

	b.WriteString("\n\n")
	b.WriteString(s.Subtle.Italic(true).Render("Press [p] to re-pair with a different realm"))

	return b.String()
}

func (m Model) renderError() string {
	s := m.styles
	t := m.theme
	var b strings.Builder

	b.WriteString(s.CardTitle.Render("Pair with Realm"))
	b.WriteString("\n\n")

	errorStyle := lipgloss.NewStyle().Foreground(t.Error).Bold(true)
	b.WriteString(errorStyle.Render("Pairing Error"))
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(t.Error).Render(m.errorMessage))
	b.WriteString("\n\n")

	b.WriteString(s.Subtle.Italic(true).Render("Press [p] to try again"))

	return b.String()
}

func (m Model) renderIdle() string {
	s := m.styles
	t := m.theme
	var b strings.Builder

	b.WriteString(s.CardTitle.Render("Pair with Realm"))
	b.WriteString("\n\n")

	b.WriteString(s.Subtle.Render("Not Paired"))
	b.WriteString("\n\n")

	// Instruction box
	contentWidth := m.width - 12
	if contentWidth < 20 {
		contentWidth = 20
	}
	instructions := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Border).
		Padding(1, 2).
		Width(contentWidth).
		Render(lipgloss.JoinVertical(lipgloss.Left,
			s.Bold.Foreground(t.Secondary).Render("Pairing connects this agent to a realm"),
			"",
			"A realm is your identity provider on the Macula mesh.",
			"Once paired, you can:",
			"",
			"  "+s.StatusOK.Render("*")+" Discover and call capabilities from other agents",
			"  "+s.StatusOK.Render("*")+" Announce your own capabilities to the mesh",
			"  "+s.StatusOK.Render("*")+" Build reputation through successful interactions",
			"",
			s.Subtle.Italic(true).Render("Don't have a realm account yet?"),
			"  Visit "+lipgloss.NewStyle().Foreground(t.Secondary).Underline(true).Render("https://macula.io")+" to create one",
		))
	b.WriteString(instructions)

	b.WriteString("\n\n")

	// CTA
	cta := lipgloss.NewStyle().
		Background(t.Primary).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 2).
		Bold(true).
		Render("Press [p] to start pairing")

	contentW := m.width - 8
	if contentW < 20 {
		contentW = 20
	}
	b.WriteString(lipgloss.PlaceHorizontal(contentW, lipgloss.Center, cta))

	return b.String()
}

func (m Model) wrapPanel(content string) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Border).
		Padding(1, 2).
		Width(m.width).
		Height(m.height).
		Render(content)
}

// Commands — async operations.

func (m Model) fetchIdentity() tea.Msg {
	identity, err := m.client.GetIdentity()
	return identityMsg{identity: identity, err: err}
}

func (m Model) startPairing() tea.Msg {
	status, err := m.client.StartPairing()
	if err != nil {
		return pairingStartedMsg{err: err}
	}
	return pairingStartedMsg{
		code:     status.Code,
		realmURL: status.RealmURL,
	}
}

func (m Model) checkPairingStatus() tea.Msg {
	status, err := m.client.GetPairingStatus()
	if err != nil {
		return pairingStatusMsg{err: err}
	}
	return pairingStatusMsg{status: status.Status}
}

func (m Model) cancelPairing() tea.Msg {
	_ = m.client.CancelPairing()
	return nil
}

// parseRealm extracts the realm from an MRI string.
// mri:agent:io.macula/name -> io.macula
func parseRealm(mri string) string {
	if !strings.HasPrefix(mri, "mri:") {
		return ""
	}
	parts := strings.Split(mri, ":")
	if len(parts) < 3 {
		return ""
	}
	pathParts := strings.Split(parts[2], "/")
	if len(pathParts) > 0 {
		return pathParts[0]
	}
	return ""
}
