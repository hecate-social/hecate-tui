package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hecate-social/hecate-tui/internal/app"
	tea "github.com/charmbracelet/bubbletea"
)

const version = "0.4.0"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("hecate v%s\n", version)
		os.Exit(0)
	}

	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		printHelp()
		os.Exit(0)
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

// resolveConnection determines whether to use Unix socket or TCP.
// Priority:
//  1. HECATE_SOCKET env var (explicit socket path)
//  2. Default socket at ~/.config/hecate/connectors/tui.sock
//  3. HECATE_URL env var (TCP)
//  4. http://localhost:4444 (TCP default)
//
// Returns (socketPath, hecateURL) — one will be empty.
func resolveConnection() (string, string) {
	// 1. Explicit socket path from env
	if socketEnv := os.Getenv("HECATE_SOCKET"); socketEnv != "" {
		if fileExists(socketEnv) {
			return socketEnv, ""
		}
		// Socket specified but doesn't exist — warn and fall through
		fmt.Fprintf(os.Stderr, "Warning: HECATE_SOCKET=%s not found, falling back to TCP\n", socketEnv)
	}

	// 2. Default socket path
	defaultSocket := defaultSocketPath()
	if defaultSocket != "" && fileExists(defaultSocket) {
		return defaultSocket, ""
	}

	// 3. TCP from env or default
	hecateURL := os.Getenv("HECATE_URL")
	if hecateURL == "" {
		hecateURL = "http://localhost:4444"
	}

	return "", hecateURL
}

// defaultSocketPath returns ~/.config/hecate/connectors/tui.sock
func defaultSocketPath() string {
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
    HECATE_SOCKET    Path to Unix socket (preferred over TCP)
    HECATE_URL       Hecate daemon URL (default: http://localhost:4444)

CONNECTION:
    The TUI connects to the daemon in this priority order:
    1. HECATE_SOCKET env var (explicit socket path)
    2. ~/.config/hecate/connectors/tui.sock (auto-created by daemon)
    3. HECATE_URL env var (TCP connection)
    4. http://localhost:4444 (TCP default)

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
