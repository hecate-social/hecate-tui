package geo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/oschwald/geoip2-golang"
)

// Checker performs geo-restriction checks using a local MaxMind database.
type Checker struct {
	db     *geoip2.Reader
	config *Config
}

// NewChecker creates a new Checker by loading the GeoIP database.
// It searches for the database in common locations.
func NewChecker() (*Checker, error) {
	dbPath := findDatabase()
	if dbPath == "" {
		return nil, fmt.Errorf("GeoIP database not found")
	}

	db, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open GeoIP database: %w", err)
	}

	config := loadConfig()

	return &Checker{
		db:     db,
		config: config,
	}, nil
}

// Close releases the GeoIP database resources.
func (c *Checker) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// CheckPublicIP checks if the local machine's public IP is allowed.
func (c *Checker) CheckPublicIP() (*CheckResult, error) {
	ip, err := GetPublicIP()
	if err != nil {
		// Offline - can't determine, check with daemon later
		return &CheckResult{Allowed: true, Message: "Unable to determine public IP"}, nil
	}
	return c.CheckIP(ip)
}

// CheckIP checks if a specific IP address is allowed.
func (c *Checker) CheckIP(ip net.IP) (*CheckResult, error) {
	// First check IP overrides
	if c.config != nil {
		if c.isIPDenied(ip) {
			return &CheckResult{
				Allowed: false,
				IP:      ip.String(),
				Message: "IP address is blocked",
			}, nil
		}
		if c.isIPAllowed(ip) {
			return &CheckResult{
				Allowed: true,
				IP:      ip.String(),
			}, nil
		}
	}

	// Look up country
	record, err := c.db.Country(ip)
	if err != nil {
		// Unknown IP - allow by default
		return &CheckResult{Allowed: true, IP: ip.String()}, nil
	}

	countryCode := record.Country.IsoCode
	countryName := record.Country.Names["en"]

	// Check country restrictions
	if c.isBlocked(countryCode) {
		return &CheckResult{
			Allowed:     false,
			CountryCode: countryCode,
			CountryName: countryName,
			IP:          ip.String(),
			Message:     "Hecate is not available in your region",
		}, nil
	}

	return &CheckResult{
		Allowed:     true,
		CountryCode: countryCode,
		CountryName: countryName,
		IP:          ip.String(),
	}, nil
}

// isBlocked checks if a country is blocked based on config.
func (c *Checker) isBlocked(countryCode string) bool {
	if c.config == nil {
		return false
	}

	switch c.config.Mode {
	case "blocklist":
		for _, blocked := range c.config.BlockedCountries {
			if blocked == countryCode {
				return true
			}
		}
		return false
	case "allowlist":
		for _, allowed := range c.config.AllowedCountries {
			if allowed == countryCode {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// isIPAllowed checks if an IP is in the allow override list.
func (c *Checker) isIPAllowed(ip net.IP) bool {
	for _, cidr := range c.config.IPOverrides.Allow {
		if matchesCIDR(ip, cidr) {
			return true
		}
	}
	return false
}

// isIPDenied checks if an IP is in the deny override list.
func (c *Checker) isIPDenied(ip net.IP) bool {
	for _, cidr := range c.config.IPOverrides.Deny {
		if matchesCIDR(ip, cidr) {
			return true
		}
	}
	return false
}

// matchesCIDR checks if an IP matches a CIDR range.
func matchesCIDR(ip net.IP, cidr string) bool {
	// Handle single IP (no CIDR notation)
	if !strings.Contains(cidr, "/") {
		singleIP := net.ParseIP(cidr)
		return singleIP != nil && ip.Equal(singleIP)
	}

	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return network.Contains(ip)
}

// GetPublicIP fetches the machine's public IP address.
func GetPublicIP() (net.IP, error) {
	services := []string{
		"https://api.ipify.org",
		"https://icanhazip.com",
		"https://ifconfig.me/ip",
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, url := range services {
		resp, err := client.Get(url)
		if err != nil {
			continue
		}

		ip := func() net.IP {
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != 200 {
				return nil
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil
			}

			ipStr := strings.TrimSpace(string(body))
			return net.ParseIP(ipStr)
		}()

		if ip != nil {
			return ip, nil
		}
	}

	return nil, fmt.Errorf("failed to determine public IP")
}

// findDatabase searches for the GeoIP database in common locations.
func findDatabase() string {
	paths := []string{
		// User-specific
		filepath.Join(userConfigDir(), "hecate-tui", "GeoLite2-Country.mmdb"),
		filepath.Join(userConfigDir(), "hecate", "GeoLite2-Country.mmdb"),
		// System-wide
		"/usr/share/GeoIP/GeoLite2-Country.mmdb",
		"/var/lib/GeoIP/GeoLite2-Country.mmdb",
		"/usr/local/share/GeoIP/GeoLite2-Country.mmdb",
		// Development
		"GeoLite2-Country.mmdb",
		"priv/GeoLite2-Country.mmdb",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// loadConfig loads the geo restriction config from file.
func loadConfig() *Config {
	// Try to load from config file
	paths := []string{
		filepath.Join(userConfigDir(), "hecate-tui", "geo_restrictions.yaml"),
		filepath.Join(userConfigDir(), "hecate", "geo_restrictions.yaml"),
		"config/geo_restrictions.yaml",
	}

	for _, path := range paths {
		if cfg := loadConfigFromFile(path); cfg != nil {
			return cfg
		}
	}

	// Return default config
	return defaultConfig()
}

// loadConfigFromFile attempts to load config from a YAML file.
// Note: For simplicity, we use a basic parser. For full YAML support,
// consider adding gopkg.in/yaml.v3 dependency.
func loadConfigFromFile(path string) *Config {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	cfg := &Config{}
	lines := strings.Split(string(data), "\n")

	var currentSection string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "mode:") {
			cfg.Mode = strings.TrimSpace(strings.TrimPrefix(line, "mode:"))
		} else if line == "blocked_countries:" {
			currentSection = "blocked"
		} else if line == "allowed_countries:" {
			currentSection = "allowed"
		} else if line == "ip_overrides:" {
			currentSection = "ip_overrides"
		} else if line == "allow:" && currentSection == "ip_overrides" {
			currentSection = "ip_allow"
		} else if line == "deny:" && currentSection == "ip_overrides" {
			currentSection = "ip_deny"
		} else if strings.HasPrefix(line, "- ") {
			value := strings.TrimSpace(strings.TrimPrefix(line, "-"))
			// Remove comments from list items
			if idx := strings.Index(value, "#"); idx > 0 {
				value = strings.TrimSpace(value[:idx])
			}
			switch currentSection {
			case "blocked":
				cfg.BlockedCountries = append(cfg.BlockedCountries, value)
			case "allowed":
				cfg.AllowedCountries = append(cfg.AllowedCountries, value)
			case "ip_allow":
				cfg.IPOverrides.Allow = append(cfg.IPOverrides.Allow, value)
			case "ip_deny":
				cfg.IPOverrides.Deny = append(cfg.IPOverrides.Deny, value)
			}
		}
	}

	return cfg
}

// defaultConfig returns the default geo restriction config.
func defaultConfig() *Config {
	return &Config{
		Mode: "blocklist",
		BlockedCountries: []string{
			"RU", "CN", "KP", "IR", "BY",
		},
		IPOverrides: struct {
			Allow []string `toml:"allow"`
			Deny  []string `toml:"deny"`
		}{
			Allow: []string{
				"127.0.0.0/8",
				"10.0.0.0/8",
				"172.16.0.0/12",
				"192.168.0.0/16",
			},
			Deny: []string{},
		},
	}
}

// userConfigDir returns the user's config directory.
func userConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		home := os.Getenv("HOME")
		if home == "" {
			return ""
		}
		return filepath.Join(home, ".config")
	}
	return dir
}

// CheckWithDaemon checks geo status via the daemon's API.
// This is used when the local database is not available.
func CheckWithDaemon(socketPath, httpURL string) (*CheckResult, error) {
	var resp *http.Response
	var err error

	if socketPath != "" {
		// Use Unix socket
		client := &http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", socketPath)
				},
			},
			Timeout: 5 * time.Second,
		}
		resp, err = client.Get("http://localhost/api/geo/status")
	} else {
		// Use HTTP
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err = client.Get(httpURL + "/api/geo/status")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to check geo status: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var status DaemonGeoStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Parse the status field
	result := &CheckResult{Allowed: true}

	switch v := status.Status.(type) {
	case string:
		if v == "allowed" {
			result.Allowed = true
		}
	case map[string]interface{}:
		if blocked, ok := v["blocked"].(string); ok {
			result.Allowed = false
			result.CountryCode = blocked
			result.Message = "Hecate is not available in your region"
		}
	}

	return result, nil
}
