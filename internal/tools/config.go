package tools

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds user preferences
type Config struct {
	// Editor preferences
	Editor EditorConfig `toml:"editor"`

	// Daemon connection
	Daemon DaemonConfig `toml:"daemon"`

	// UI preferences
	UI UIConfig `toml:"ui"`

	// Tool overrides
	Tools ToolsConfig `toml:"tools"`
}

// EditorConfig holds editor preferences
type EditorConfig struct {
	Preferred string   `toml:"preferred"` // Command name (nvim, code, etc.)
	Args      []string `toml:"args"`      // Extra arguments
}

// DaemonConfig holds daemon connection settings
type DaemonConfig struct {
	URL     string `toml:"url"`     // Default: http://localhost:4444
	Timeout int    `toml:"timeout"` // Request timeout in seconds
}

// UIConfig holds UI preferences
type UIConfig struct {
	Theme       string `toml:"theme"`        // dark, light
	Animations  bool   `toml:"animations"`   // Enable animations
	CompactMode bool   `toml:"compact_mode"` // Compact display
}

// ToolsConfig allows overriding tool paths
type ToolsConfig struct {
	Paths map[string]string `toml:"paths"` // tool name -> path override
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Editor: EditorConfig{
			Preferred: "",
			Args:      nil,
		},
		Daemon: DaemonConfig{
			URL:     "http://localhost:4444",
			Timeout: 30,
		},
		UI: UIConfig{
			Theme:       "dark",
			Animations:  true,
			CompactMode: false,
		},
		Tools: ToolsConfig{
			Paths: make(map[string]string),
		},
	}
}

// ConfigPath returns the path to config file
func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".hecate", "config.toml")
}

// ConfigDir returns the hecate config directory
func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".hecate")
}

// LoadConfig loads configuration from ~/.hecate/config.toml
func LoadConfig() (*Config, error) {
	cfg := DefaultConfig()

	path := ConfigPath()
	if path == "" {
		return cfg, nil
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	// Parse TOML
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// SaveConfig saves configuration to ~/.hecate/config.toml
func SaveConfig(cfg *Config) error {
	dir := ConfigDir()
	if dir == "" {
		return os.ErrNotExist
	}

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := ConfigPath()
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	encoder := toml.NewEncoder(f)
	return encoder.Encode(cfg)
}

// EnsureConfigDir creates ~/.hecate if it doesn't exist
func EnsureConfigDir() error {
	dir := ConfigDir()
	if dir == "" {
		return os.ErrNotExist
	}
	return os.MkdirAll(dir, 0755)
}

// WriteDefaultConfig creates a default config file if none exists
func WriteDefaultConfig() error {
	path := ConfigPath()
	if path == "" {
		return os.ErrNotExist
	}

	// Don't overwrite existing
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	return SaveConfig(DefaultConfig())
}
