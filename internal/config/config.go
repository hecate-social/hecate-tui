package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds all persistent user preferences (consolidated TOML).
type Config struct {
	// Theme name (dark, light, monochrome)
	Theme string `toml:"theme,omitempty"`

	// System prompt for LLM
	SystemPrompt string `toml:"system_prompt,omitempty"`

	// Connection settings
	Connection ConnectionConfig `toml:"connection"`

	// Editor preferences
	Editor EditorConfig `toml:"editor"`

	// UI preferences
	UI UIConfig `toml:"ui"`
}

// ConnectionConfig holds daemon connection settings.
type ConnectionConfig struct {
	// Path to Unix domain socket (preferred over URL)
	SocketPath string `toml:"socket_path,omitempty"`

	// TCP URL fallback (default: http://localhost:4444)
	DaemonURL string `toml:"daemon_url,omitempty"`

	// Request timeout in seconds
	Timeout int `toml:"timeout,omitempty"`
}

// EditorConfig holds editor preferences.
type EditorConfig struct {
	Preferred string   `toml:"preferred,omitempty"`
	Args      []string `toml:"args,omitempty"`
}

// UIConfig holds UI preferences.
type UIConfig struct {
	Animations  bool `toml:"animations"`
	CompactMode bool `toml:"compact_mode"`
}

// configDir returns ~/.config/hecate-tui.
func configDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(dir, "hecate-tui")
}

// DefaultPath returns ~/.config/hecate-tui/config.toml.
func DefaultPath() string {
	return filepath.Join(configDir(), "config.toml")
}

// Load reads config from disk, performing migration if needed.
// Returns zero-value Config if file doesn't exist.
func Load() Config {
	// Try new location first
	path := DefaultPath()
	if cfg, err := loadTOML(path); err == nil {
		return cfg
	}

	// New config doesn't exist — try migration from old formats
	cfg := migrateOldConfigs()

	return cfg
}

// LoadFrom reads config from a specific TOML path.
func LoadFrom(path string) Config {
	cfg, _ := loadTOML(path)
	return cfg
}

// Save writes config to disk at the default path.
func (c Config) Save() error {
	return c.SaveTo(DefaultPath())
}

// SaveTo writes config to a specific path, creating directories as needed.
func (c Config) SaveTo(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	return encoder.Encode(c)
}

// DaemonURL returns the configured daemon URL (backward-compatible accessor).
func (c Config) DaemonURL() string {
	return c.Connection.DaemonURL
}

// loadTOML reads a TOML config file. Returns error if file doesn't exist.
func loadTOML(path string) (Config, error) {
	var cfg Config
	cfg.UI.Animations = true // default

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, err
	}

	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// migrateOldConfigs reads old JSON and TOML configs, merges them into the new format,
// writes the consolidated config, and renames old files.
func migrateOldConfigs() Config {
	var cfg Config
	cfg.UI.Animations = true
	migrated := false

	// 1. Read old JSON config (~/.config/hecate/config.json)
	oldJSONPath := oldJSONConfigPath()
	if data, err := os.ReadFile(oldJSONPath); err == nil {
		var oldJSON oldJSONConfig
		if json.Unmarshal(data, &oldJSON) == nil {
			cfg.Theme = oldJSON.Theme
			cfg.SystemPrompt = oldJSON.SystemPrompt
			cfg.Connection.DaemonURL = oldJSON.DaemonURL
			migrated = true
		}
	}

	// 2. Read old TOML config (~/.hecate/config.toml)
	oldTOMLPath := oldTOMLConfigPath()
	if _, err := os.Stat(oldTOMLPath); err == nil {
		var oldTOML oldTOMLConfig
		if _, err := toml.DecodeFile(oldTOMLPath, &oldTOML); err == nil {
			cfg.Editor.Preferred = oldTOML.Editor.Preferred
			cfg.Editor.Args = oldTOML.Editor.Args
			cfg.UI.Animations = oldTOML.UI.Animations
			cfg.UI.CompactMode = oldTOML.UI.CompactMode
			// Merge daemon URL if not already set from JSON
			if cfg.Connection.DaemonURL == "" && oldTOML.Daemon.URL != "" {
				cfg.Connection.DaemonURL = oldTOML.Daemon.URL
			}
			if oldTOML.Daemon.Timeout > 0 {
				cfg.Connection.Timeout = oldTOML.Daemon.Timeout
			}
			// Override theme from TOML if set and JSON didn't have one
			if cfg.Theme == "" && oldTOML.UI.Theme != "" {
				cfg.Theme = oldTOML.UI.Theme
			}
			migrated = true
		}
	}

	// 3. Set default socket path
	cfg.Connection.SocketPath = defaultSocketPath()

	if migrated {
		// Save new consolidated config
		_ = cfg.Save()

		// Rename old files to .migrated
		if _, err := os.Stat(oldJSONPath); err == nil {
			_ = os.Rename(oldJSONPath, oldJSONPath+".migrated")
		}
		if _, err := os.Stat(oldTOMLPath); err == nil {
			_ = os.Rename(oldTOMLPath, oldTOMLPath+".migrated")
		}
	}

	return cfg
}

// Old config format types (for migration only)
type oldJSONConfig struct {
	Theme        string `json:"theme,omitempty"`
	SystemPrompt string `json:"system_prompt,omitempty"`
	DaemonURL    string `json:"daemon_url,omitempty"`
}

type oldTOMLConfig struct {
	Editor struct {
		Preferred string   `toml:"preferred"`
		Args      []string `toml:"args"`
	} `toml:"editor"`
	Daemon struct {
		URL     string `toml:"url"`
		Timeout int    `toml:"timeout"`
	} `toml:"daemon"`
	UI struct {
		Theme       string `toml:"theme"`
		Animations  bool   `toml:"animations"`
		CompactMode bool   `toml:"compact_mode"`
	} `toml:"ui"`
}

func oldJSONConfigPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(dir, "hecate", "config.json")
}

func oldTOMLConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".hecate", "config.toml")
}

func defaultSocketPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(dir, "hecate", "connectors", "tui.sock")
}
