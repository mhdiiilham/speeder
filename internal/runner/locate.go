package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const defaultLocateURL = "https://locate.measurementlab.net/v2/nearest/ndt/ndt7"

type locateResponse struct {
	Results []locateResult `json:"results"`
}

type locateResult struct {
	Machine  string            `json:"machine"`
	Location locateLocation    `json:"location"`
	URLs     map[string]string `json:"urls"`
}

type locateLocation struct {
	City    string `json:"city"`
	Country string `json:"country"`
}

// HTTPLocateClient discovers ndt7 servers via the M-Lab locate v2 API.
type HTTPLocateClient struct {
	// URL overrides the default locate endpoint (used in tests).
	URL string
}

func (c *HTTPLocateClient) endpoint() string {
	if c.URL != "" {
		return c.URL
	}
	return defaultLocateURL
}

// Nearest returns nearby ndt7 servers sorted by proximity.
func (c *HTTPLocateClient) Nearest(ctx context.Context) ([]Server, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "speeder/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("locate request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("locate API returned HTTP %d", resp.StatusCode)
	}

	var lr locateResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return nil, fmt.Errorf("decode locate response: %w", err)
	}

	servers := make([]Server, 0, len(lr.Results))
	for _, r := range lr.Results {
		dlURL := r.URLs["wss:///ndt/v7/download"]
		ulURL := r.URLs["wss:///ndt/v7/upload"]
		if dlURL == "" || ulURL == "" {
			continue
		}
		servers = append(servers, Server{
			Hostname:    r.Machine,
			City:        r.Location.City,
			Country:     r.Location.Country,
			DownloadURL: dlURL,
			UploadURL:   ulURL,
		})
	}
	if len(servers) == 0 {
		return nil, fmt.Errorf("no usable ndt7 servers in locate response")
	}
	return servers, nil
}
