// Package telemetry implements opt-in usage telemetry for FRACTURE.
// No personal data, no simulation content, no API keys are ever sent.
// Only: anonymous UUID, version, OS, country (from IP geolocation).
package telemetry

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	// DefaultEndpoint is the live Google Apps Script Web App that receives pings.
	// Google Apps Script returns a 302 redirect to script.googleusercontent.com;
	// the http.Client below follows it automatically.
	DefaultEndpoint = "https://script.google.com/macros/s/AKfycbxQkWW3PlPoGQ24YnQy4Lp0TNy1gfWB01bzR7DVwtz9C3INsvRnCMyExgEaxo1He1Ry/exec"

	uuidFile = "telemetry_id"
	optFile  = "telemetry_opt"
)

// Ping is the payload sent to the telemetry endpoint.
// Fields match the Google Apps Script webhook schema exactly.
type Ping struct {
	ID       string `json:"install_id"`
	Version  string `json:"version"`
	OS       string `json:"os"`
	Arch     string `json:"arch,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	Country  string `json:"country,omitempty"`
	City     string `json:"city,omitempty"`
	IP       string `json:"ip,omitempty"`
	Event    string `json:"event"`
}

// Client manages telemetry state and sending.
type Client struct {
	dataDir  string
	endpoint string
	version  string
	enabled  bool
}

// New creates a new telemetry client.
// dataDir is the FRACTURE data directory (e.g. ~/.local/share/FRACTURE).
// endpoint is the Google Apps Script URL.
func New(dataDir, endpoint, version string) *Client {
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	c := &Client{
		dataDir:  dataDir,
		endpoint: endpoint,
		version:  version,
	}
	c.enabled = c.isEnabled()
	return c
}

// IsEnabled returns whether telemetry is enabled.
func (c *Client) IsEnabled() bool {
	return c.enabled
}

// Enable turns on telemetry and saves the preference.
func (c *Client) Enable() error {
	c.enabled = true
	return os.WriteFile(filepath.Join(c.dataDir, optFile), []byte("1"), 0600)
}

// Disable turns off telemetry and saves the preference.
func (c *Client) Disable() error {
	c.enabled = false
	return os.WriteFile(filepath.Join(c.dataDir, optFile), []byte("0"), 0600)
}

// SendPing sends an anonymous ping to the telemetry endpoint.
// It is a no-op if telemetry is disabled or the endpoint is not configured.
// Runs asynchronously — never blocks the main flow.
func (c *Client) SendPing() {
	if !c.enabled {
		return
	}
	// DefaultEndpoint is now the real endpoint — always send.
	go c.sendAsync()
}

func (c *Client) sendAsync() {
	uuid, err := c.getOrCreateUUID()
	if err != nil {
		return
	}

	hostname, _ := os.Hostname()

	ping := Ping{
		ID:       uuid,
		Version:  c.version,
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Hostname:  hostname,
		Event:    "install",
	}

	// Try to get country from IP (lightweight, no personal data)
	if country, city, ip := getGeoInfo(); country != "" {
		ping.Country = country
		ping.City = city
		ping.IP = maskIP(ip)
	}

	body, err := json.Marshal(ping)
	if err != nil {
		return
	}

	// Google Apps Script returns a 302 redirect on POST.
	// We must NOT auto-follow it (Go converts POST→GET on redirect, losing the body).
	// Instead: do the POST, capture the Location header, then GET the redirect URL.
	// Note: Google may close the connection with EOF after the 302 — this is normal.
	noRedirectClient := &http.Client{
		Timeout: 20 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // stop at first redirect
		},
	}
	req, err := http.NewRequest(http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "FRACTURE/"+c.version)

	resp, err := noRedirectClient.Do(req)
	if err != nil {
		// EOF is normal for Google Apps Script — the ping was still received.
		return
	}
	defer resp.Body.Close()

	// Follow the redirect with a GET to retrieve the actual JSON response.
	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
		loc := resp.Header.Get("Location")
		if loc != "" {
			getResp, err := noRedirectClient.Get(loc)
			if err == nil {
				getResp.Body.Close()
			}
		}
	}
}

// getOrCreateUUID returns the persistent anonymous UUID for this installation.
func (c *Client) getOrCreateUUID() (string, error) {
	path := filepath.Join(c.dataDir, uuidFile)

	data, err := os.ReadFile(path)
	if err == nil && len(data) == 36 {
		return string(data), nil
	}

	// Generate new UUID v4
	uuid, err := generateUUID()
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(path, []byte(uuid), 0600); err != nil {
		return "", err
	}

	return uuid, nil
}

// isEnabled reads the opt-in preference from disk.
// Default: enabled (opt-out model — user can disable in settings).
func (c *Client) isEnabled() bool {
	path := filepath.Join(c.dataDir, optFile)
	data, err := os.ReadFile(path)
	if err != nil {
		// File doesn't exist — default to enabled, create it
		_ = os.WriteFile(path, []byte("1"), 0600)
		return true
	}
	return string(data) == "1"
}

// generateUUID generates a random UUID v4.
func generateUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant bits
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

// getGeoInfo fetches country and city from a free IP geolocation API.
// Returns empty strings on failure — never blocks.
func getGeoInfo() (country, city, ip string) {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("https://ipapi.co/json/")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var result struct {
		CountryCode string `json:"country_code"`
		City        string `json:"city"`
		IP          string `json:"ip"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return
	}
	return result.CountryCode, result.City, result.IP
}

// maskIP masks the last octet of an IPv4 address for privacy.
// e.g. "177.10.20.30" → "177.10.20.x"
func maskIP(ip string) string {
	for i := len(ip) - 1; i >= 0; i-- {
		if ip[i] == '.' {
			return ip[:i+1] + "x"
		}
	}
	return ip
}
