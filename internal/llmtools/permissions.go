package llmtools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// PermissionLevel controls how tool execution is authorized.
type PermissionLevel int

const (
	PermissionDeny  PermissionLevel = iota // Never allow
	PermissionAsk                          // Ask user each time
	PermissionAllow                        // Allow without asking
)

func (p PermissionLevel) String() string {
	switch p {
	case PermissionDeny:
		return "deny"
	case PermissionAsk:
		return "ask"
	case PermissionAllow:
		return "allow"
	default:
		return "unknown"
	}
}

// ParsePermissionLevel converts a string to PermissionLevel.
func ParsePermissionLevel(s string) PermissionLevel {
	switch strings.ToLower(s) {
	case "allow", "yes", "true":
		return PermissionAllow
	case "deny", "no", "false":
		return PermissionDeny
	default:
		return PermissionAsk
	}
}

// Permissions manages tool authorization rules.
type Permissions struct {
	// Per-tool permission overrides
	Tools map[string]PermissionLevel

	// Path-based restrictions for filesystem operations
	AllowedPaths []string
	DeniedPaths  []string

	// Command restrictions for run_command tool
	AllowedCommands []string
	DeniedCommands  []string

	// Default behavior when no specific rule matches
	RequireApprovalByDefault bool

	// Session-level grants (tool name -> true if granted for session)
	sessionGrants map[string]bool
}

// NewPermissions creates a Permissions with sensible defaults.
func NewPermissions() *Permissions {
	return &Permissions{
		Tools:                    make(map[string]PermissionLevel),
		AllowedPaths:             []string{},
		DeniedPaths:              defaultDeniedPaths(),
		AllowedCommands:          defaultAllowedCommands(),
		DeniedCommands:           defaultDeniedCommands(),
		RequireApprovalByDefault: true,
		sessionGrants:            make(map[string]bool),
	}
}

func defaultDeniedPaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, ".ssh"),
		filepath.Join(home, ".gnupg"),
		filepath.Join(home, ".aws"),
		filepath.Join(home, ".config", "hecate-tui", "secrets*"),
		"/etc/passwd",
		"/etc/shadow",
	}
}

func defaultAllowedCommands() []string {
	return []string{
		"go", "npm", "cargo", "make", "git",
		"ls", "pwd", "cat", "head", "tail",
		"grep", "find", "rg", "fd",
		"rebar3", "mix", "elixir", "erl",
		"python", "python3", "pip",
		"node", "yarn", "pnpm",
		"docker", "kubectl",
	}
}

func defaultDeniedCommands() []string {
	return []string{
		"rm -rf /",
		"rm -rf /*",
		"sudo rm",
		"sudo dd",
		":(){ :|:& };:",
		"mkfs",
		"curl | sh",
		"wget | sh",
		"curl | bash",
		"wget | bash",
	}
}

// Check returns the permission level for a tool with given arguments.
func (p *Permissions) Check(toolName string, args json.RawMessage) PermissionLevel {
	// Check session grants first
	if p.sessionGrants[toolName] {
		return PermissionAllow
	}

	// Check explicit tool permissions
	if level, ok := p.Tools[toolName]; ok {
		return level
	}

	// Default based on RequireApprovalByDefault
	if p.RequireApprovalByDefault {
		return PermissionAsk
	}
	return PermissionAllow
}

// CheckPath validates if a filesystem path is allowed.
func (p *Permissions) CheckPath(path string) PermissionLevel {
	// Expand ~ in path
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[2:])
	}

	// Make absolute
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	// Check denied paths first (deny takes precedence)
	for _, denied := range p.DeniedPaths {
		expandedDenied := expandPath(denied)
		if matchPath(absPath, expandedDenied) {
			return PermissionDeny
		}
	}

	// Check allowed paths
	if len(p.AllowedPaths) > 0 {
		for _, allowed := range p.AllowedPaths {
			expandedAllowed := expandPath(allowed)
			if matchPath(absPath, expandedAllowed) {
				return PermissionAllow
			}
		}
		// If we have explicit allowed paths and this isn't in them, ask
		return PermissionAsk
	}

	// No explicit allowed paths means ask by default
	return PermissionAsk
}

// CheckCommand validates if a shell command is allowed.
func (p *Permissions) CheckCommand(cmd string) PermissionLevel {
	// Check denied patterns first
	for _, denied := range p.DeniedCommands {
		if strings.Contains(cmd, denied) {
			return PermissionDeny
		}
	}

	// Check if command starts with an allowed command
	cmdParts := strings.Fields(cmd)
	if len(cmdParts) == 0 {
		return PermissionDeny
	}

	baseCmd := filepath.Base(cmdParts[0])
	for _, allowed := range p.AllowedCommands {
		if baseCmd == allowed {
			return PermissionAsk // Still ask, but allow user to approve
		}
	}

	// Unknown command - more cautious
	return PermissionAsk
}

// GrantForSession grants a tool permission for the current session.
func (p *Permissions) GrantForSession(toolName string) {
	if p.sessionGrants == nil {
		p.sessionGrants = make(map[string]bool)
	}
	p.sessionGrants[toolName] = true
}

// SessionGranted returns true if the tool has been granted session permission.
func (p *Permissions) SessionGranted(toolName string) bool {
	if p.sessionGrants == nil {
		return false
	}
	return p.sessionGrants[toolName]
}

// RevokeSessionGrant removes a session-level grant.
func (p *Permissions) RevokeSessionGrant(toolName string) {
	delete(p.sessionGrants, toolName)
}

// ClearSessionGrants removes all session-level grants.
func (p *Permissions) ClearSessionGrants() {
	p.sessionGrants = make(map[string]bool)
}

// SetToolPermission sets an explicit permission for a tool.
func (p *Permissions) SetToolPermission(toolName string, level PermissionLevel) {
	p.Tools[toolName] = level
}

// DisableTool disables a tool (sets it to PermissionDeny).
func (p *Permissions) DisableTool(toolName string) {
	p.Tools[toolName] = PermissionDeny
}

// EnableTool enables a tool (removes the deny, falls back to default behavior).
func (p *Permissions) EnableTool(toolName string) {
	delete(p.Tools, toolName)
	// Also clear any session grant so it will ask again
	delete(p.sessionGrants, toolName)
}

// IsDisabled returns true if the tool is explicitly disabled.
func (p *Permissions) IsDisabled(toolName string) bool {
	level, ok := p.Tools[toolName]
	return ok && level == PermissionDeny
}

// DisabledTools returns a list of explicitly disabled tool names.
func (p *Permissions) DisabledTools() []string {
	var disabled []string
	for name, level := range p.Tools {
		if level == PermissionDeny {
			disabled = append(disabled, name)
		}
	}
	return disabled
}

// expandPath expands ~ to home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

// matchPath checks if path matches pattern (supports * wildcard at end).
func matchPath(path, pattern string) bool {
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}

	// Exact match or path is under pattern directory
	if path == pattern {
		return true
	}

	// Check if path is under the pattern directory
	patternDir := pattern
	if !strings.HasSuffix(patternDir, string(os.PathSeparator)) {
		patternDir += string(os.PathSeparator)
	}
	return strings.HasPrefix(path, patternDir)
}
