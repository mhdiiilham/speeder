package runner

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- mocks ----

type mockLocate struct {
	servers []Server
	err     error
}

func (m *mockLocate) Nearest(_ context.Context) ([]Server, error) {
	return m.servers, m.err
}

type mockSpeed struct {
	dlMeasurements []Measurement
	ulMeasurements []Measurement
	dlErr          error
	ulErr          error
}

func (m *mockSpeed) Download(_ context.Context, _ string) (<-chan Measurement, error) {
	if m.dlErr != nil {
		return nil, m.dlErr
	}
	ch := make(chan Measurement, len(m.dlMeasurements))
	for _, meas := range m.dlMeasurements {
		ch <- meas
	}
	close(ch)
	return ch, nil
}

func (m *mockSpeed) Upload(_ context.Context, _ string) (<-chan Measurement, error) {
	if m.ulErr != nil {
		return nil, m.ulErr
	}
	ch := make(chan Measurement, len(m.ulMeasurements))
	for _, meas := range m.ulMeasurements {
		ch <- meas
	}
	close(ch)
	return ch, nil
}

type mockPing struct {
	avg    time.Duration
	jitter time.Duration
	err    error
}

func (m *mockPing) Ping(_ context.Context, _ string, _ int) (time.Duration, time.Duration, error) {
	return m.avg, m.jitter, m.err
}

// ---- helpers ----

func stableMeasurements(n int, bytesPerTick int64) []Measurement {
	ms := make([]Measurement, n)
	for i := range ms {
		ms[i] = Measurement{
			NumBytes:  bytesPerTick * int64(i+1),
			ElapsedMs: int64((i + 1) * 250),
		}
	}
	return ms
}

func defaultTestServer() Server {
	return Server{
		Hostname:    "test.example.com",
		City:        "Singapore",
		Country:     "SG",
		DownloadURL: "wss://test.example.com/ndt/v7/download",
		UploadURL:   "wss://test.example.com/ndt/v7/upload",
	}
}

// ---- Config tests ----

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, 8*time.Second, cfg.MaxDuration)
	assert.Equal(t, 0.03, cfg.CVThreshold)
	assert.Equal(t, 2*time.Second, cfg.MinDuration)
	assert.Equal(t, 8, cfg.WindowSize)
}

func TestQuickConfig(t *testing.T) {
	cfg := QuickConfig()
	assert.Equal(t, 4*time.Second, cfg.MaxDuration)
	assert.Equal(t, 0.02, cfg.CVThreshold)
	assert.Equal(t, 6, cfg.WindowSize)
}

// ---- Runner.ListServers ----

func TestRunner_ListServers(t *testing.T) {
	servers := []Server{defaultTestServer()}
	r := New(&mockLocate{servers: servers}, &mockSpeed{}, &mockPing{})
	got, err := r.ListServers(context.Background())
	require.NoError(t, err)
	assert.Equal(t, servers, got)
}

func TestRunner_ListServers_Error(t *testing.T) {
	r := New(&mockLocate{err: errors.New("network error")}, &mockSpeed{}, &mockPing{})
	_, err := r.ListServers(context.Background())
	assert.ErrorContains(t, err, "network error")
}

// ---- Runner.Run ----

func TestRunner_Run_HappyPath(t *testing.T) {
	srv := defaultTestServer()
	meas := stableMeasurements(4, 3_000_000) // ~96 Mbps at 250 ms ticks

	r := New(
		&mockLocate{servers: []Server{srv}},
		&mockSpeed{dlMeasurements: meas, ulMeasurements: meas},
		&mockPing{avg: 14 * time.Millisecond, jitter: 1 * time.Millisecond},
	)

	result, err := r.Run(context.Background(), DefaultConfig())
	require.NoError(t, err)
	assert.Equal(t, srv.Hostname, result.Server.Hostname)
	assert.InDelta(t, 14.0, result.LatencyMs, 0.1)
	assert.InDelta(t, 1.0, result.JitterMs, 0.1)
	assert.Greater(t, result.Download.SpeedMbps, 0.0)
	assert.Greater(t, result.Upload.SpeedMbps, 0.0)
	assert.Greater(t, result.Download.Bytes, int64(0))
}

func TestRunner_Run_WithServerHost(t *testing.T) {
	meas := stableMeasurements(2, 1_000_000)
	r := New(
		&mockLocate{err: errors.New("should not be called")},
		&mockSpeed{dlMeasurements: meas, ulMeasurements: meas},
		&mockPing{avg: 10 * time.Millisecond},
	)

	cfg := DefaultConfig()
	cfg.ServerHost = "custom.host.example.com"
	result, err := r.Run(context.Background(), cfg)
	require.NoError(t, err)
	assert.Equal(t, "custom.host.example.com", result.Server.Hostname)
}

func TestRunner_Run_LocateError(t *testing.T) {
	r := New(
		&mockLocate{err: errors.New("DNS failure")},
		&mockSpeed{},
		&mockPing{},
	)
	_, err := r.Run(context.Background(), DefaultConfig())
	assert.ErrorContains(t, err, "DNS failure")
}

func TestRunner_Run_LocateEmpty(t *testing.T) {
	r := New(&mockLocate{servers: nil}, &mockSpeed{}, &mockPing{})
	_, err := r.Run(context.Background(), DefaultConfig())
	assert.ErrorContains(t, err, "no servers returned")
}

func TestRunner_Run_PingError(t *testing.T) {
	srv := defaultTestServer()
	r := New(
		&mockLocate{servers: []Server{srv}},
		&mockSpeed{},
		&mockPing{err: errors.New("ping failed")},
	)
	_, err := r.Run(context.Background(), DefaultConfig())
	assert.ErrorContains(t, err, "ping")
}

func TestRunner_Run_DownloadError(t *testing.T) {
	srv := defaultTestServer()
	r := New(
		&mockLocate{servers: []Server{srv}},
		&mockSpeed{dlErr: errors.New("ws dial failed")},
		&mockPing{avg: 10 * time.Millisecond},
	)
	_, err := r.Run(context.Background(), DefaultConfig())
	assert.ErrorContains(t, err, "start download")
}

func TestRunner_Run_UploadError(t *testing.T) {
	srv := defaultTestServer()
	meas := stableMeasurements(2, 1_000_000)
	r := New(
		&mockLocate{servers: []Server{srv}},
		&mockSpeed{dlMeasurements: meas, ulErr: errors.New("ws dial failed")},
		&mockPing{avg: 10 * time.Millisecond},
	)
	_, err := r.Run(context.Background(), DefaultConfig())
	assert.ErrorContains(t, err, "start upload")
}

// ---- runPhase ----

func TestRunPhase_ClosedChannel(t *testing.T) {
	ch := make(chan Measurement, 3)
	ch <- Measurement{NumBytes: 1_000_000, ElapsedMs: 250}
	ch <- Measurement{NumBytes: 2_000_000, ElapsedMs: 500}
	ch <- Measurement{NumBytes: 3_000_000, ElapsedMs: 750}
	close(ch)

	cfg := DefaultConfig()
	result := runPhase(context.Background(), ch, cfg, nil)
	assert.Greater(t, result.SpeedMbps, 0.0)
	assert.Equal(t, int64(3_000_000), result.Bytes)
}

func TestRunPhase_ContextCancelled(t *testing.T) {
	// Channel never closes; context expires first.
	ch := make(chan Measurement) // unbuffered, no sender

	cfg := DefaultConfig()
	cfg.MaxDuration = 50 * time.Millisecond

	result := runPhase(context.Background(), ch, cfg, nil)
	// result is zero because no measurements arrived before timeout
	assert.Equal(t, float64(0), result.SpeedMbps)
}

func TestRunPhase_EarlyConvergence(t *testing.T) {
	// 12 stable measurements at 100 Mbps; should converge before max duration.
	meas := stableMeasurements(12, 3_125_000) // 3.125 MB per 250 ms ≈ 100 Mbps
	ch := make(chan Measurement, len(meas))
	for _, m := range meas {
		ch <- m
	}
	close(ch)

	cfg := DefaultConfig()
	cfg.MaxDuration = 10 * time.Second // would take 3 s if no early stop
	start := time.Now()
	result := runPhase(context.Background(), ch, cfg, nil)
	elapsed := time.Since(start)

	// Should finish near-instantly (channel drained) not block for 10 s
	assert.Less(t, elapsed, 1*time.Second)
	assert.Greater(t, result.SpeedMbps, 0.0)
}

func TestRunPhase_ZeroElapsed(t *testing.T) {
	// Measurement with ElapsedMs=0 must be ignored.
	ch := make(chan Measurement, 2)
	ch <- Measurement{NumBytes: 1_000_000, ElapsedMs: 0} // invalid
	ch <- Measurement{NumBytes: 2_000_000, ElapsedMs: 500}
	close(ch)

	cfg := DefaultConfig()
	result := runPhase(context.Background(), ch, cfg, nil)
	assert.InDelta(t, 32.0, result.SpeedMbps, 1.0)
}

func TestRunPhase_ProgressCallback(t *testing.T) {
	ch := make(chan Measurement, 2)
	ch <- Measurement{NumBytes: 1_000_000, ElapsedMs: 250}
	ch <- Measurement{NumBytes: 2_000_000, ElapsedMs: 500}
	close(ch)

	var calls int
	cfg := DefaultConfig()
	runPhase(context.Background(), ch, cfg, func(mbps float64) {
		calls++
		assert.Greater(t, mbps, 0.0)
	})
	assert.Equal(t, 2, calls)
}

// ---- resolveServer ----

func TestResolveServer_ByHost(t *testing.T) {
	r := New(&mockLocate{}, &mockSpeed{}, &mockPing{})
	srv, err := r.resolveServer(context.Background(), "my.server.com")
	require.NoError(t, err)
	assert.Equal(t, "my.server.com", srv.Hostname)
	assert.Contains(t, srv.DownloadURL, "my.server.com")
	assert.Contains(t, srv.UploadURL, "my.server.com")
}

func TestResolveServer_Auto(t *testing.T) {
	expected := defaultTestServer()
	r := New(&mockLocate{servers: []Server{expected}}, &mockSpeed{}, &mockPing{})
	srv, err := r.resolveServer(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, expected.Hostname, srv.Hostname)
}
