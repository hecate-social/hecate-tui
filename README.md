# hecate-tui

Terminal UI for monitoring and managing the Hecate daemon.

## Features

- **Status View** - Daemon health, version, uptime, and identity
- **Mesh View** - Mesh topology and peer connections (coming soon)
- **Capabilities View** - Browse discovered capabilities
- **RPC View** - List registered procedures
- **Logs View** - View daemon logs (coming soon)

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

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Navigate between views |
| `1-5` | Jump to specific view |
| `r` | Refresh current view |
| `q` / `Ctrl+C` | Quit |
| `?` | Show help |

## Views

### Status (1)

Shows daemon health and agent identity:

```
hecate-tui                                    healthy

Daemon Status

Status:         healthy
Version:        0.1.0
Uptime:         2h 15m

Identity

MRI:            mri:agent:io.macula/my-agent
Created:        2026-02-01T12:00:00Z
```

### Mesh (2)

Shows mesh topology and peer connections (coming soon).

### Capabilities (3)

Lists discovered capabilities from the mesh:

```
Discovered Capabilities

MRI:            mri:capability:io.macula/weather
Agent:          mri:agent:io.macula/weather-service
Description:    Weather forecast service
Tags:           weather, forecast
```

### RPC (4)

Lists registered procedures:

```
Registered Procedures

Name:           echo
MRI:            mri:rpc:io.macula/echo
Endpoint:       http://localhost:8080/echo
```

### Logs (5)

View daemon logs (coming soon).

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `HECATE_URL` | Hecate daemon URL | `http://localhost:4444` |

## Requirements

- Hecate daemon running on localhost:4444 (or configured URL)
- Terminal with 256 color support

## Technology

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components

## License

Apache 2.0 - See [LICENSE](LICENSE)

## Support

- [Issues](https://github.com/hecate-social/hecate-tui/issues)
- [Buy Me a Coffee](https://buymeacoffee.com/rlefever)
