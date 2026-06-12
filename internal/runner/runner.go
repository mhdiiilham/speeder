package runner

import (
	"context"
	"fmt"
	"time"
)

// LocateClient discovers nearby ndt7 servers.
type LocateClient interface {
	Nearest(ctx context.Context) ([]Server, error)
}

// SpeedClient runs download and upload tests over the ndt7 WebSocket protocol.
type SpeedClient interface {
	Download(ctx context.Context, url string) (<-chan Measurement, error)
	Upload(ctx context.Context, url string) (<-chan Measurement, error)
}

// PingClient measures round-trip latency to a host.
type PingClient interface {
	Ping(ctx context.Context, host string, n int) (avg, jitter time.Duration, err error)
}

// Server holds information about an ndt7 test server.
type Server struct {
	Hostname    string
	City        string
	Country     string
	DownloadURL string
	UploadURL   string
}

// Measurement is a single data point streamed from a running speed test.
type Measurement struct {
	NumBytes  int64 // cumulative bytes transferred since test start
	ElapsedMs int64 // milliseconds elapsed since test start
}

// PhaseResult summarises one completed download or upload phase.
type PhaseResult struct {
	SpeedMbps float64
	Bytes     int64
	Duration  time.Duration
}

// Result holds all measurements from a completed speed test.
type Result struct {
	Server    Server
	LatencyMs float64
	JitterMs  float64
	Download  PhaseResult
	Upload    PhaseResult
	ISP       string // populated externally from ipinfo
	ClientIP  string // populated externally from ipinfo
}

// PingOnly reports whether this result has no download/upload data.
func (r *Result) PingOnly() bool {
	return r.Download.Bytes == 0 && r.Upload.Bytes == 0
}

// Config controls how tests are run.
type Config struct {
	ServerHost  string
	MaxDuration time.Duration // hard ceiling per phase
	CVThreshold float64       // coefficient of variation threshold for early stop
	MinDuration time.Duration // minimum time before convergence is checked
	WindowSize  int           // number of ticks in the sliding CV window
	PingOnly    bool          // skip download/upload, measure latency only
	OnProgress  func(phase string, mbps float64)
}

// DefaultConfig returns production defaults: 8 s max, 3 % CV threshold.
func DefaultConfig() Config {
	return Config{
		MaxDuration: 8 * time.Second,
		CVThreshold: 0.03,
		MinDuration: 2 * time.Second,
		WindowSize:  8,
	}
}

// QuickConfig returns a data-minimising preset: 4 s max, 2 % CV threshold.
func QuickConfig() Config {
	return Config{
		MaxDuration: 4 * time.Second,
		CVThreshold: 0.02,
		MinDuration: 2 * time.Second,
		WindowSize:  6,
	}
}

// Runner orchestrates a full ndt7 speed test.
type Runner struct {
	locate LocateClient
	speed  SpeedClient
	ping   PingClient
}

// New creates a Runner with injected dependencies (used by tests).
func New(locate LocateClient, speed SpeedClient, ping PingClient) *Runner {
	return &Runner{locate: locate, speed: speed, ping: ping}
}

// Default returns a Runner backed by real network clients.
func Default() *Runner {
	return New(&HTTPLocateClient{}, &NDT7Client{}, &TCPPinger{})
}

// ListServers returns nearby ndt7 servers via the locate client.
func (r *Runner) ListServers(ctx context.Context) ([]Server, error) {
	return r.locate.Nearest(ctx)
}

// Run executes a full speed test and returns the aggregated result.
func (r *Runner) Run(ctx context.Context, cfg Config) (*Result, error) {
	server, err := r.resolveServer(ctx, cfg.ServerHost)
	if err != nil {
		return nil, err
	}

	avgLat, jitter, err := r.ping.Ping(ctx, server.Hostname, 10)
	if err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	result := &Result{
		Server:    server,
		LatencyMs: float64(avgLat.Microseconds()) / 1000.0,
		JitterMs:  float64(jitter.Microseconds()) / 1000.0,
	}

	if cfg.PingOnly {
		return result, nil
	}

	dlCh, err := r.speed.Download(ctx, server.DownloadURL)
	if err != nil {
		return nil, fmt.Errorf("start download: %w", err)
	}
	result.Download = runPhase(ctx, dlCh, cfg, func(mbps float64) {
		if cfg.OnProgress != nil {
			cfg.OnProgress("download", mbps)
		}
	})

	ulCh, err := r.speed.Upload(ctx, server.UploadURL)
	if err != nil {
		return nil, fmt.Errorf("start upload: %w", err)
	}
	result.Upload = runPhase(ctx, ulCh, cfg, func(mbps float64) {
		if cfg.OnProgress != nil {
			cfg.OnProgress("upload", mbps)
		}
	})

	return result, nil
}

func (r *Runner) resolveServer(ctx context.Context, host string) (Server, error) {
	if host != "" {
		return Server{
			Hostname:    host,
			DownloadURL: fmt.Sprintf("wss://%s/ndt/v7/download", host),
			UploadURL:   fmt.Sprintf("wss://%s/ndt/v7/upload", host),
		}, nil
	}
	servers, err := r.locate.Nearest(ctx)
	if err != nil {
		return Server{}, fmt.Errorf("locate server: %w", err)
	}
	if len(servers) == 0 {
		return Server{}, fmt.Errorf("no servers returned by locate API")
	}
	return servers[0], nil
}

// runPhase drains ch, stops early when speed converges, and returns the result.
func runPhase(ctx context.Context, ch <-chan Measurement, cfg Config, onProgress func(float64)) PhaseResult {
	ctx, cancel := context.WithTimeout(ctx, cfg.MaxDuration)
	defer cancel()

	var windows []float64
	var lastMbps float64
	var lastBytes int64
	var lastElapsedMs int64

	for {
		select {
		case <-ctx.Done():
			return PhaseResult{
				SpeedMbps: lastMbps,
				Bytes:     lastBytes,
				Duration:  time.Duration(lastElapsedMs) * time.Millisecond,
			}
		case m, ok := <-ch:
			if !ok {
				return PhaseResult{
					SpeedMbps: lastMbps,
					Bytes:     lastBytes,
					Duration:  time.Duration(lastElapsedMs) * time.Millisecond,
				}
			}
			if m.ElapsedMs <= 0 {
				continue
			}

			seconds := float64(m.ElapsedMs) / 1000.0
			mbps := float64(m.NumBytes) * 8.0 / seconds / 1_000_000.0

			lastMbps = mbps
			lastBytes = m.NumBytes
			lastElapsedMs = m.ElapsedMs

			windows = append(windows, mbps)
			if len(windows) > cfg.WindowSize {
				windows = windows[1:]
			}

			if onProgress != nil {
				onProgress(mbps)
			}

			elapsed := time.Duration(m.ElapsedMs) * time.Millisecond
			if elapsed >= cfg.MinDuration &&
				len(windows) >= cfg.WindowSize &&
				coeffVariation(windows) < cfg.CVThreshold {
				cancel()
			}
		}
	}
}
