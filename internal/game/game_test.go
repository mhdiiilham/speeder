package game

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- mock Pinger ----

type mockPinger struct {
	results map[string]pingResult
}

type pingResult struct {
	avg    time.Duration
	jitter time.Duration
	loss   float64
	err    error
}

func (m *mockPinger) Ping(_ context.Context, host, _ string, _ int) (time.Duration, time.Duration, float64, error) {
	if r, ok := m.results[host]; ok {
		return r.avg, r.jitter, r.loss, r.err
	}
	return 0, 0, 0, errors.New("no mock for " + host)
}

// ---- mock Game ----

type mockGame struct {
	name    string
	servers []Server
	err     error
}

func (g *mockGame) Name() string { return g.name }
func (g *mockGame) Note() string { return "" }
func (g *mockGame) Servers(_ context.Context) ([]Server, error) {
	return g.servers, g.err
}

// ---- Rate tests ----

func TestRate(t *testing.T) {
	tests := []struct {
		latency float64
		want    Rating
	}{
		{-1, RatingUnknown},
		{0, RatingExcellent},
		{15, RatingExcellent},
		{29, RatingExcellent},
		{30, RatingGood},
		{59, RatingGood},
		{60, RatingPlayable},
		{79, RatingPlayable},
		{80, RatingPoor},
		{119, RatingPoor},
		{120, RatingVeryPoor},
		{999, RatingVeryPoor},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.want, Rate(tc.latency), "latency %.0f ms", tc.latency)
	}
}

// ---- Score tests ----

func TestScore(t *testing.T) {
	// Perfect connection — maximum score.
	assert.Equal(t, 100, Score(0, 0, 0))
	// Unreachable — zero score.
	assert.Equal(t, 0, Score(-1, 0, 0))
	// High latency alone drops score.
	assert.Less(t, Score(100, 0, 0), 60)
	// High jitter alone drops score.
	assert.Less(t, Score(20, 50, 0), 71)
	// High packet loss alone drops score.
	assert.Less(t, Score(20, 0, 10), 85)
	// Score is always in [0, 100].
	assert.GreaterOrEqual(t, Score(500, 500, 100), 0)
	assert.LessOrEqual(t, Score(0, 0, 0), 100)
}

func TestScore_Clamp(t *testing.T) {
	// Extreme values must not produce negative or >100 scores.
	assert.Equal(t, 0, Score(1000, 1000, 100))
	assert.Equal(t, 100, Score(0, 0, 0))
}

// ---- Verdict tests ----

func TestVerdict(t *testing.T) {
	assert.Contains(t, Verdict("CS2", PingResult{Score: 90, City: "Singapore", LatencyMs: 12}), "Ready to play")
	assert.Contains(t, Verdict("CS2", PingResult{Score: 75, City: "Tokyo", LatencyMs: 45}), "Good to play")
	assert.Contains(t, Verdict("CS2", PingResult{Score: 55, City: "Frankfurt", LatencyMs: 70}), "Playable")
	assert.Contains(t, Verdict("CS2", PingResult{Score: 40, City: "Chicago", LatencyMs: 95}), "High latency")
	assert.Contains(t, Verdict("CS2", PingResult{Score: 10, City: "London", LatencyMs: 200}), "Very poor")
	assert.Contains(t, Verdict("CS2", PingResult{Err: errors.New("unreachable")}), "Could not reach")
}

// ---- Check tests ----

func TestCheck_HappyPath(t *testing.T) {
	g := &mockGame{
		name: "TestGame",
		servers: []Server{
			{Region: "AP", City: "Singapore", Host: "sgp.example.com", Port: "443"},
			{Region: "EU", City: "Frankfurt", Host: "eu.example.com", Port: "443"},
		},
	}
	p := &mockPinger{results: map[string]pingResult{
		"sgp.example.com": {avg: 15 * time.Millisecond, jitter: 1 * time.Millisecond, loss: 0},
		"eu.example.com":  {avg: 80 * time.Millisecond, jitter: 5 * time.Millisecond, loss: 2},
	}}

	results, err := Check(context.Background(), g, p, 5)
	require.NoError(t, err)
	require.Len(t, results, 2)

	// Best server first.
	assert.Equal(t, "Singapore", results[0].City)
	assert.True(t, results[0].Best)
	assert.False(t, results[1].Best)
	assert.Greater(t, results[0].Score, results[1].Score)
}

func TestCheck_SortedByScore(t *testing.T) {
	g := &mockGame{
		name: "TestGame",
		servers: []Server{
			{Host: "a.example.com", Port: "443", City: "A"},
			{Host: "b.example.com", Port: "443", City: "B"},
			{Host: "c.example.com", Port: "443", City: "C"},
		},
	}
	p := &mockPinger{results: map[string]pingResult{
		"a.example.com": {avg: 100 * time.Millisecond},
		"b.example.com": {avg: 20 * time.Millisecond},
		"c.example.com": {avg: 50 * time.Millisecond},
	}}

	results, err := Check(context.Background(), g, p, 3)
	require.NoError(t, err)
	assert.Equal(t, "B", results[0].City)
	assert.Equal(t, "C", results[1].City)
	assert.Equal(t, "A", results[2].City)
}

func TestCheck_UnreachableServerLast(t *testing.T) {
	g := &mockGame{
		name: "TestGame",
		servers: []Server{
			{Host: "good.example.com", Port: "443", City: "Good"},
			{Host: "bad.example.com", Port: "443", City: "Bad"},
		},
	}
	p := &mockPinger{results: map[string]pingResult{
		"good.example.com": {avg: 50 * time.Millisecond},
		"bad.example.com":  {err: errors.New("unreachable")},
	}}

	results, err := Check(context.Background(), g, p, 3)
	require.NoError(t, err)
	assert.Equal(t, "Good", results[0].City)
	assert.Equal(t, "Bad", results[1].City)
	assert.NotNil(t, results[1].Err)
}

func TestCheck_LocateError(t *testing.T) {
	g := &mockGame{name: "TestGame", err: errors.New("API down")}
	_, err := Check(context.Background(), g, &mockPinger{}, 3)
	assert.ErrorContains(t, err, "API down")
}

func TestCheck_NoServers(t *testing.T) {
	g := &mockGame{name: "TestGame", servers: nil}
	_, err := Check(context.Background(), g, &mockPinger{}, 3)
	assert.ErrorContains(t, err, "no TestGame servers")
}

// ---- SteamGame tests ----

func TestSteamGame_Servers(t *testing.T) {
	custom := []Server{
		{Region: "AP", City: "Singapore", Host: "cmp1-sgp1.steamserver.net", Port: "27018"},
		{Region: "AP", City: "Tokyo",     Host: "cmp1-tyo1.steamserver.net", Port: "27018"},
	}
	g := &SteamGame{name: "CS2", appID: 730, servers: custom}
	servers, err := g.Servers(context.Background())
	require.NoError(t, err)
	assert.Equal(t, custom, servers)
}

func TestSteamGame_Servers_Default(t *testing.T) {
	g := &SteamGame{name: "CS2", appID: 730}
	servers, err := g.Servers(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, servers)
	for _, s := range servers {
		assert.NotEmpty(t, s.Host)
		assert.NotEmpty(t, s.Port)
		assert.NotEmpty(t, s.City)
	}
}

func TestSteamGame_Note(t *testing.T) {
	g := &SteamGame{name: "CS2", appID: 730}
	assert.Contains(t, g.Note(), "SDR")
	assert.Contains(t, g.Note(), "2–5 ms")
}

func TestNewCS2(t *testing.T) {
	g := NewCS2()
	assert.Equal(t, "CS2", g.Name())
}

func TestNewDota2(t *testing.T) {
	g := NewDota2()
	assert.Equal(t, "Dota 2", g.Name())
}

// ---- TCPGamePinger tests ----

func startEchoServer(t *testing.T) (host, port string) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { l.Close() })
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()
	h, p, _ := net.SplitHostPort(l.Addr().String())
	return h, p
}

func TestTCPGamePinger_HappyPath(t *testing.T) {
	host, port := startEchoServer(t)
	p := &TCPGamePinger{}
	avg, jitter, loss, err := p.Ping(context.Background(), host, port, 5)
	require.NoError(t, err)
	assert.Greater(t, avg, time.Duration(0))
	assert.GreaterOrEqual(t, jitter, time.Duration(0))
	assert.Equal(t, 0.0, loss)
}

func TestTCPGamePinger_AllFail(t *testing.T) {
	// Port 1 is always refused; with the new logic refused = failure → all-fail error.
	p := &TCPGamePinger{DialTimeout: 200 * time.Millisecond}
	_, _, loss, err := p.Ping(context.Background(), "127.0.0.1", "1", 2)
	// All probes fail → 100% loss and an error.
	assert.Error(t, err)
	assert.Equal(t, 100.0, loss)
}

func TestTCPGamePinger_ContextCancel(t *testing.T) {
	host, port := startEchoServer(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := &TCPGamePinger{}
	avg, _, _, _ := p.Ping(ctx, host, port, 5)
	assert.GreaterOrEqual(t, avg, time.Duration(0))
}

func TestTCPGamePinger_DefaultTimeout(t *testing.T) {
	p := &TCPGamePinger{}
	assert.Equal(t, 2*time.Second, p.dialTimeout())
}

func TestTCPGamePinger_CustomTimeout(t *testing.T) {
	p := &TCPGamePinger{DialTimeout: 500 * time.Millisecond}
	assert.Equal(t, 500*time.Millisecond, p.dialTimeout())
}



// ---- durationMean / durationJitter ----

func TestDurationMean(t *testing.T) {
	assert.Equal(t, time.Duration(0), durationMean(nil))
	assert.Equal(t, 20*time.Millisecond, durationMean([]time.Duration{10 * time.Millisecond, 30 * time.Millisecond}))
}

func TestDurationJitter(t *testing.T) {
	assert.Equal(t, time.Duration(0), durationJitter(nil))
	assert.Equal(t, time.Duration(0), durationJitter([]time.Duration{10 * time.Millisecond}))
	assert.Greater(t, durationJitter([]time.Duration{10 * time.Millisecond, 30 * time.Millisecond}), time.Duration(0))
}
