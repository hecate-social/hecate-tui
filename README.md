# hecate-tui

Terminal UI for the Hecate Developer Studio - monitor, manage, and interact with the Hecate daemon and Macula mesh.

## Features

- **Chat** - LLM-powered chat with streaming responses (Ollama backend)
- **Browse** - Discover and search capabilities on the mesh
- **Projects** - Project management (coming soon)
- **Monitor** - Daemon health, stats, and mesh connection status
- **Pair** - Pair your agent with a Macula realm
- **Me** - Identity profile and settings

## Installation

### Build from Source

Requires Go 1.21+:

```bash
git clone https://github.com/hecate-social/hecate-tui
cd hecate-tui
go build -o hecate-tui ./cmd/hecate-tui
./hecate-tui
```

### Pre-built Binaries

Download from [Releases](https://github.com/hecate-social/hecate-tui/releases).

## Usage

```bash
# Start TUI (connects to localhost:4444)
hecate-tui

# Connect to different daemon
HECATE_URL=http://localhost:5555 hecate-tui

# Show version
hecate-tui --version

# Show help
hecate-tui --help
```

## Keyboard Shortcuts

### Global

| Key | Action |
|-----|--------|
| `1-6` | Jump to view (Chat, Browse, Projects, Monitor, Pair, Me) |
| `Tab` / `Shift+Tab` | Navigate between views |
| `q` | Quit (except in Chat) |
| `Ctrl+C` | Force quit |

### Chat View

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Tab` | Cycle through models |
| `Ctrl+L` | Clear chat |
| `Esc` | Cancel streaming / exit chat |
| `↑↓` | Scroll history |

### Browse View

| Key | Action |
|-----|--------|
| `/` | Search capabilities |
| `Enter` | View capability details |
| `↑↓` | Navigate list |
| `r` | Refresh |
| `Esc` | Close search/details |

### Monitor View

| Key | Action |
|-----|--------|
| `r` | Refresh stats |

### Pair View

| Key | Action |
|-----|--------|
| `p` | Start pairing |
| `c` / `Esc` | Cancel pairing |
| `r` | Refresh |

### Me View

| Key | Action |
|-----|--------|
| `s` | Open settings |
| `↑↓` | Navigate settings |
| `Enter` | Toggle setting |
| `Esc` | Close settings |
| `r` | Refresh |

## Views

### Chat (1)

LLM-powered chat with streaming responses:

```
Hecate Chat  ● llama3.2

You
  What is the Macula mesh?

Assistant
  The Macula mesh is a decentralized network for AI agents...

  152 tokens • 24.5 tok/s
```

Features:
- Model selection (Tab to cycle)
- Streaming responses with animated indicator
- Token count and speed display
- Message history with scroll

### Browse (2)

Discover capabilities on the mesh:

```
Browse Capabilities

/llm                                    3 of 10

CAPABILITY                    SOURCE    TAGS
> serve_llm/llama3.2         local     llm, chat
  serve_llm/qwen2.5-coder    local     llm, code
  weather.forecast           remote    weather, api
```

Features:
- Live search filtering
- Detail view with full capability info
- Local/remote source indicator

### Projects (3)

Project management workspace (coming soon).

### Monitor (4)

Daemon and mesh status:

```
Monitor

┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
│   2h 15m │  │        3 │  │       10 │  │   Online │
│   Uptime │  │  Subscr. │  │   Capab. │  │   Status │
└──────────┘  └──────────┘  └──────────┘  └──────────┘

┌─ Daemon ──────────┐  ┌─ Mesh ────────────┐
│ Status:  healthy  │  │ Bootstrap: boot.. │
│ Version: 0.1.1    │  │ Status: connected │
│ Uptime:  2h 15m   │  │                   │
└───────────────────┘  └───────────────────┘
```

### Pair (5)

Pair your agent with a realm:

```
Pair

Enter this code on the realm:

  ╔══════════════════╗
  ║                  ║
  ║      A7B3C9      ║
  ║                  ║
  ╚══════════════════╝

1. Go to https://macula.io/pair
2. Sign in to your account
3. Enter the code shown above
4. Confirm the pairing
```

### Me (6)

Identity and settings:

```
Me

Identity
    ___
   /   \     mri:agent:io.macula/my-agent
  | o o |    Realm: io.macula
  |  >  |    Paired
   \___/     Since 2026-02-01

Statistics
  Capabilities:  10 announced
  Subscriptions: 3 active
  Daemon:        Online (v0.1.1)
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `HECATE_URL` | Hecate daemon URL | `http://localhost:4444` |

## Requirements

- Hecate daemon running on localhost:4444 (or configured URL)
- Terminal with 256 color support
- For Chat: Ollama running with at least one model

## Technology

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components

## License

Apache 2.0 - See [LICENSE](LICENSE)

## Support

- [Issues](https://github.com/hecate-social/hecate-tui/issues)
- [Buy Me a Coffee](https://buymeacoffee.com/rlefever)
