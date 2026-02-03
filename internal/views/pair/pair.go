package pair

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/ui/styles"
)

// PairingState represents the current pairing state
type PairingState int

const (
	StateIdle PairingState = iota
	StateStarting
	StateWaiting
	StatePaired
	StateError
)

// Model is the pair view model
type Model struct {
	client   *client.Client
	width    int
	height   int
	focused  bool
	identity *client.Identity
	spinner  spinner.Model

	// Pairing state
	state        PairingState
	pairingCode  string
	realmURL     string
	expiresAt    time.Time
	errorMessage string

	// Polling
	pollTicker *time.Ticker
}

// New creates a new pair view
func New(c *client.Client) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	return Model{
		client:  c,
		spinner: s,
		state:   StateIdle,
	}
}

// Messages
type identityMsg struct {
	identity *client.Identity
	err      error
}

type pairingStartedMsg struct {
	code      string
	realmURL  string
	expiresAt string
	err       error
}

type pairingStatusMsg struct {
	status string
	err    error
}

type pairingPollMsg struct{}

// Init initializes the view
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchIdentity,
	)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}
		switch msg.String() {
		case "p":
			if m.state == StateIdle || m.state == StateError || m.state == StatePaired {
				m.state = StateStarting
				m.errorMessage = ""
				return m, m.startPairing
			}
		case "c", "esc":
			if m.state == StateWaiting {
				m.state = StateIdle
				m.pairingCode = ""
				return m, m.cancelPairing
			}
		case "r":
			return m, m.fetchIdentity
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case identityMsg:
		m.identity = msg.identity
		// Check if already paired
		if m.identity != nil && m.identity.Identity != "" {
			realm := parseRealm(m.identity.Identity)
			if realm != "" {
				m.state = StatePaired
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
			// Start polling for status
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
				// Refresh identity
				return m, m.fetchIdentity
			case "waiting":
				// Continue polling
				return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
					return pairingPollMsg{}
				})
			case "expired":
				m.state = StateError
				m.errorMessage = "Pairing session expired"
				m.pairingCode = ""
			case "error":
				m.state = StateError
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

// View renders the view
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(styles.TitleStyle.Render("Pair"))
	b.WriteString("\n\n")

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

	return styles.BoxStyle.Width(m.width - 4).Render(b.String())
}

func (m Model) renderStarting() string {
	return m.spinner.View() + " Starting pairing session..."
}

func (m Model) renderWaiting() string {
	var b strings.Builder

	// Status
	b.WriteString(WaitingStatusStyle.Render("Waiting for confirmation..."))
	b.WriteString("\n\n")

	// Code display
	if m.pairingCode != "" {
		codeLabel := CodeLabelStyle.Render("Enter this code on the realm:")
		codeDisplay := CodeStyle.Render(m.pairingCode)
		codeBox := CodeBoxStyle.Render(codeLabel + "\n\n" + codeDisplay)
		b.WriteString(lipgloss.PlaceHorizontal(m.width-8, lipgloss.Center, codeBox))
		b.WriteString("\n\n")
	}

	// Instructions
	steps := []string{
		"Go to " + URLStyle.Render(m.realmURL),
		"Sign in to your account",
		"Enter the code shown above",
		"Confirm the pairing",
	}

	for i, step := range steps {
		b.WriteString(RenderStep(i+1, step))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Spinner
	b.WriteString(m.spinner.View() + " " + ProgressTextStyle.Render("Polling for confirmation..."))
	b.WriteString("\n\n")

	// Cancel hint
	b.WriteString(HintStyle.Render("Press [c] or [Esc] to cancel"))

	return b.String()
}

func (m Model) renderPaired() string {
	var b strings.Builder

	// Status
	b.WriteString(PairedStatusStyle.Render("Paired"))
	b.WriteString("\n\n")

	// Identity info
	if m.identity != nil {
		b.WriteString(SectionTitleStyle.Render("Identity"))
		b.WriteString("\n")

		mri := m.identity.Identity
		b.WriteString("  " + lipgloss.JoinHorizontal(lipgloss.Left,
			styles.LabelStyle.Render("MRI:"),
			lipgloss.NewStyle().Foreground(styles.Primary).Render(mri),
		))
		b.WriteString("\n")

		realm := parseRealm(mri)
		if realm != "" {
			b.WriteString("  " + lipgloss.JoinHorizontal(lipgloss.Left,
				styles.LabelStyle.Render("Realm:"),
				styles.ValueStyle.Render(realm),
			))
			b.WriteString("\n")
		}

		if m.identity.CreatedAt != "" {
			b.WriteString("  " + lipgloss.JoinHorizontal(lipgloss.Left,
				styles.LabelStyle.Render("Since:"),
				styles.ValueStyle.Render(m.identity.CreatedAt),
			))
		}
	}

	b.WriteString("\n\n")

	// Connected info
	connected := SectionBoxStyle.Width(m.width - 12).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			PairedStatusStyle.Render("Connected to Macula Mesh"),
			"",
			HintStyle.Render("Your agent can now:"),
			"  • Discover capabilities on the mesh",
			"  • Announce your own capabilities",
			"  • Make and receive RPC calls",
			"  • Participate in reputation tracking",
		),
	)
	b.WriteString(connected)

	b.WriteString("\n\n")

	// Actions
	b.WriteString(HintStyle.Render("Press [p] to re-pair with a different realm"))

	return b.String()
}

func (m Model) renderError() string {
	var b strings.Builder

	// Status
	b.WriteString(ErrorStatusStyle.Render("Pairing Error"))
	b.WriteString("\n\n")

	// Error message
	b.WriteString(lipgloss.NewStyle().
		Foreground(ErrorColor).
		Render(m.errorMessage))
	b.WriteString("\n\n")

	// Retry hint
	b.WriteString(HintStyle.Render("Press [p] to try again"))

	return b.String()
}

func (m Model) renderIdle() string {
	var b strings.Builder

	// Status
	b.WriteString(IdleStatusStyle.Render("Not Paired"))
	b.WriteString("\n\n")

	// Instructions box
	instructions := SectionBoxStyle.Width(m.width - 12).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			SectionTitleStyle.Render("Pairing connects this agent to a realm"),
			"",
			"A realm is your identity provider on the Macula mesh.",
			"Once paired, you can:",
			"",
			"  • Discover and call capabilities from other agents",
			"  • Announce your own capabilities to the mesh",
			"  • Build reputation through successful interactions",
			"",
			HintStyle.Render("Don't have a realm account yet?"),
			"  Visit "+URLStyle.Render("https://macula.io")+" to create one",
		),
	)
	b.WriteString(instructions)

	b.WriteString("\n\n")

	// CTA
	cta := lipgloss.NewStyle().
		Background(styles.Primary).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 2).
		Bold(true).
		Render("Press [p] to start pairing")

	b.WriteString(lipgloss.PlaceHorizontal(m.width-8, lipgloss.Center, cta))

	return b.String()
}

// Commands
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
		code:      status.Code,
		realmURL:  status.RealmURL,
		expiresAt: status.ExpiresAt,
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

// View interface implementation

// Name returns the tab label
func (m Model) Name() string {
	return "Pair"
}

// ShortHelp returns help text
func (m Model) ShortHelp() string {
	switch m.state {
	case StateWaiting:
		return "c: cancel pairing"
	case StatePaired:
		return "p: re-pair • r: refresh"
	default:
		return "p: start pairing"
	}
}

// SetSize updates dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Focus activates the view
func (m *Model) Focus() {
	m.focused = true
}

// Blur deactivates the view
func (m *Model) Blur() {
	m.focused = false
}

// Helper functions

func parseRealm(mri string) string {
	// mri:agent:io.macula/name -> io.macula
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
