package main

import (
	"fmt"
	"os"

	"github.com/hecate-social/hecate-tui/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

const version = "0.1.0"

func main() {
	// Check for version flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("hecate-tui v%s\n", version)
		os.Exit(0)
	}

	// Check for help flag
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		printHelp()
		os.Exit(0)
	}

	// Get hecate daemon URL from environment or use default
	hecateURL := os.Getenv("HECATE_URL")
	if hecateURL == "" {
		hecateURL = "http://localhost:4444"
	}

	// Create and run the TUI
	p := tea.NewProgram(
		ui.NewApp(hecateURL),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`hecate-tui - Terminal UI for Macula Hecate Daemon

USAGE:
    hecate-tui [OPTIONS]

OPTIONS:
    -h, --help       Show this help message
    -v, --version    Show version

ENVIRONMENT:
    HECATE_URL       Hecate daemon URL (default: http://localhost:4444)

KEYBOARD SHORTCUTS:
    Tab / Shift+Tab  Navigate between views
    q / Ctrl+C       Quit
    r                Refresh current view
    ?                Show help

VIEWS:
    Status           Daemon health and connectivity
    Mesh             Mesh topology and peers
    Capabilities     Registered capabilities
    RPC              Call remote procedures
    Logs             View daemon logs

For more information, visit: https://github.com/hecate-social/hecate-tui`)
}
