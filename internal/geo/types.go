// Package geo provides geographic restriction checking for the Hecate TUI.
// It supports checking the local machine's public IP against country restrictions
// and communicating with the daemon for geo status.
package geo

// CheckResult represents the result of a geo-restriction check.
type CheckResult struct {
	Allowed     bool   `json:"allowed"`
	CountryCode string `json:"country_code,omitempty"`
	CountryName string `json:"country_name,omitempty"`
	Message     string `json:"message,omitempty"`
	IP          string `json:"ip,omitempty"`
}

// Config represents the geo-restriction configuration.
type Config struct {
	Mode             string   `toml:"mode"` // "allowlist" or "blocklist"
	BlockedCountries []string `toml:"blocked_countries"`
	AllowedCountries []string `toml:"allowed_countries"`
	IPOverrides      struct {
		Allow []string `toml:"allow"`
		Deny  []string `toml:"deny"`
	} `toml:"ip_overrides"`
}

// DaemonGeoStatus represents the response from the daemon's /api/geo/status endpoint.
type DaemonGeoStatus struct {
	OK             bool   `json:"ok"`
	Enabled        bool   `json:"enabled"`
	DatabaseLoaded bool   `json:"database_loaded"`
	Status         any    `json:"status"` // "allowed" or {"blocked": "XX"}
	Config         struct {
		Mode         string `json:"mode"`
		BlockedCount int    `json:"blocked_count"`
		AllowedCount int    `json:"allowed_count"`
	} `json:"config"`
}

// StatusBlocked represents a blocked status with country code.
type StatusBlocked struct {
	Blocked string `json:"blocked"`
}
