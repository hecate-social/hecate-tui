package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/app"
	"github.com/hecate-social/hecate-tui/internal/geo"
	"github.com/hecate-social/hecate-tui/internal/ui"
	"github.com/hecate-social/hecate-tui/internal/version"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("hecate v%s\n", version.Version)
		os.Exit(0)
	}

	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		printHelp()
		os.Exit(0)
	}

	// Check geo-restriction FIRST, before anything else
	if blocked, countryCode, countryName := checkGeoRestriction(); blocked {
		fmt.Fprint(os.Stderr, ui.RenderGeoBlockedMessage(countryCode, countryName))
		os.Exit(1)
	}

	// Resolve daemon connection: socket preferred, TCP fallback
	socketPath, hecateURL := resolveConnection()

	var a *app.App
	if socketPath != "" {
		a = app.NewWithSocket(socketPath)
	} else {
		a = app.New(hecateURL)
	}

	p := tea.NewProgram(
		a,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// checkGeoRestriction performs a geo-restriction check before starting the TUI.
// Returns (blocked, countryCode, countryName).
func checkGeoRestriction() (bool, string, string) {
	// Skip geo check if explicitly disabled
	if os.Getenv("HECATE_SKIP_GEO_CHECK") == "1" {
		return false, "", ""
	}

	// Try local database check first
	checker, err := geo.NewChecker()
	if err == nil {
		defer func() { _ = checker.Close() }()
		result, err := checker.CheckPublicIP()
		if err == nil && !result.Allowed {
			return true, result.CountryCode, result.CountryName
		}
		// If allowed or error, continue
		return false, "", ""
	}

	// Local database not available, try daemon API
	socketPath, hecateURL := resolveConnection()
	result, err := geo.CheckWithDaemon(socketPath, hecateURL)
	if err != nil {
		// Can't check - allow by default (daemon will enforce)
		return false, "", ""
	}

	if !result.Allowed {
		return true, result.CountryCode, result.CountryName
	}

	return false, "", ""
}

// resolveConnection determines whether to use Unix socket or TCP.
// Priority:
//  1. HECATE_SOCKET env var (explicit socket path)
//  2. /run/hecate/daemon.sock (system-level, k8s deployment)
//  3. $HOME/.hecate/daemon.sock (local dev, multi-user safe)
//  4. ~/.config/hecate/connectors/tui.sock (user-level, local dev)
//  5. HECATE_URL env var (TCP)
//  6. http://localhost:4444 (TCP default - DEPRECATED)
//
// Returns (socketPath, hecateURL) — one will be empty.
func resolveConnection() (string, string) {
	// 1. Explicit socket path from env
	if socketEnv := os.Getenv("HECATE_SOCKET"); socketEnv != "" {
		if fileExists(socketEnv) {
			return socketEnv, ""
		}
		// Socket specified but doesn't exist — warn and fall through
		fmt.Fprintf(os.Stderr, "Warning: HECATE_SOCKET=%s not found, falling back\n", socketEnv)
	}

	// 2. System-level socket (k8s/daemonset deployment)
	systemSocket := "/run/hecate/daemon.sock"
	if fileExists(systemSocket) {
		return systemSocket, ""
	}

	// 3. User home socket ($HOME/.hecate/ — multi-user safe, no root needed)
	if home := os.Getenv("HOME"); home != "" {
		homeSocket := filepath.Join(home, ".hecate", "daemon.sock")
		if fileExists(homeSocket) {
			return homeSocket, ""
		}
	}

	// 4. User config socket (~/.config/hecate/connectors/tui.sock)
	userSocket := userSocketPath()
	if userSocket != "" && fileExists(userSocket) {
		return userSocket, ""
	}

	// 5. TCP from env or default (deprecated - socket preferred)
	hecateURL := os.Getenv("HECATE_URL")
	if hecateURL == "" {
		hecateURL = "http://localhost:4444"
	}

	return "", hecateURL
}

// userSocketPath returns ~/.config/hecate/connectors/tui.sock
func userSocketPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		home := os.Getenv("HOME")
		if home == "" {
			return ""
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "hecate", "connectors", "tui.sock")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func printHelp() {
	fmt.Println(`hecate - Terminal interface for Macula Hecate Daemon

USAGE:
    hecate [OPTIONS]

OPTIONS:
    -h, --help       Show this help message
    -v, --version    Show version

ENVIRONMENT:
    HECATE_SOCKET         Path to Unix socket (preferred over TCP)
    HECATE_URL            Hecate daemon URL (default: http://localhost:4444)
    HECATE_SKIP_GEO_CHECK Set to "1" to skip geo-restriction check

CONNECTION:
    The TUI connects to the daemon in this priority order:
    1. HECATE_SOCKET env var (explicit socket path)
    2. /run/hecate/daemon.sock (system socket, k8s deployment)
    3. $HOME/.hecate/daemon.sock (local dev, multi-user safe)
    4. ~/.config/hecate/connectors/tui.sock (user socket, local dev)
    5. HECATE_URL env var (TCP connection, deprecated)
    6. http://localhost:4444 (TCP default, deprecated)

MODES:
    Normal           Default. Scroll chat, access commands.
    Insert (i)       Type messages to send to the LLM.
    Command (/)      Execute slash commands.
    Browse           Browse capabilities (via /browse).
    Pair             Realm pairing wizard (via /pair).
    Edit             Built-in file editor (via /edit).
    Projects         Project lifecycle browser (via /alc).

KEY BINDINGS:
    Normal mode:
      i              Enter Insert mode (start typing)
      /              Enter Command mode
      j/k            Scroll chat up/down
      Ctrl+D/U       Half-page scroll
      g/G            Jump to top/bottom
      r              Retry last message
      y              Copy last response to clipboard
      ?              Show help
      q              Quit

    Insert mode:
      Enter          Send message
      Alt+Enter      Insert newline (multiline)
      Tab            Cycle LLM model
      Esc            Return to Normal

    Command mode:
      Enter          Execute command
      Tab            Autocomplete
      Up/Down        Browse command history
      Esc            Cancel

COMMANDS:
    /help            Show available commands
    /status          Daemon status
    /health          Quick health check
    /geo             Show geo-restriction status
    /models          List LLM models
    /model <name>    Switch model
    /me              Show identity
    /browse          Browse mesh capabilities
    /call <mri>      Call a mesh procedure (RPC)
    /pair            Realm pairing wizard
    /tools           Detect installed developer tools
    /config          Show current configuration
    /project         Show workspace and project info
    /new             Start a new conversation
    /history         List saved conversations
    /load <id>       Load a saved conversation
    /delete <id>     Delete a saved conversation
    /find <term>     Search chat messages
    /save [file]     Export chat transcript to markdown
    /subs            List active mesh subscriptions
    /system [text]   Set/view LLM system prompt
    /edit [file]     Open built-in editor
    /theme <name>    Switch theme (dark, light, monochrome)
    /provider        Manage LLM providers (add, remove, list)
    /alc             Project lifecycle (browse, init, manage phases)
    /clear           Clear chat
    /quit            Quit

For more information: https://github.com/hecate-social/hecate-tui`)
}
