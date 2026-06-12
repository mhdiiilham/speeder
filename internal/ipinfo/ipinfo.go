// Package ipinfo fetches the caller's public IP address and ISP name.
package ipinfo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const defaultURL = "https://ipinfo.io/json"

// Info holds the caller's public IP and ISP name.
type Info struct {
	IP  string
	ISP string
}

// Client fetches IP info from a configurable endpoint.
type Client struct {
	// URL overrides the default endpoint (used in tests).
	URL string
}

func (c *Client) endpoint() string {
	if c.URL != "" {
		return c.URL
	}
	return defaultURL
}

type ipinfoResponse struct {
	IP  string `json:"ip"`
	Org string `json:"org"` // e.g. "AS4818 Maxis Broadband Sdn Bhd"
}

// Fetch returns the caller's public IP and ISP. Never returns an error that
// should block the main flow — callers should treat a nil result as "unknown".
func (c *Client) Fetch(ctx context.Context) (*Info, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "speeder/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ipinfo request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ipinfo returned HTTP %d", resp.StatusCode)
	}

	var r ipinfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("decode ipinfo response: %w", err)
	}

	return &Info{
		IP:  r.IP,
		ISP: parseISP(r.Org),
	}, nil
}

// parseISP strips the leading AS number from the org field.
// "AS4818 Maxis Broadband Sdn Bhd" → "Maxis Broadband Sdn Bhd"
func parseISP(org string) string {
	parts := strings.SplitN(org, " ", 2)
	if len(parts) == 2 && strings.HasPrefix(parts[0], "AS") {
		return parts[1]
	}
	return org
}

// DefaultClient is a ready-to-use client pointing at ipinfo.io.
var DefaultClient = &Client{}
