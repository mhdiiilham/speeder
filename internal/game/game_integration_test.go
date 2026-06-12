//go:build integration

package game_test

import (
	"context"
	"testing"
	"time"

	"github.com/mhdiiilham/speeder/internal/game"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests hit real network endpoints.
// Run with: go test -tags integration ./internal/game/ -v -timeout 60s

func TestCS2_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	g := game.NewCS2()
	p := &game.TCPGamePinger{}

	results, err := game.Check(ctx, g, p, 5)
	require.NoError(t, err)
	assert.NotEmpty(t, results, "should get at least one CS2 server")

	t.Logf("CS2 results (%d servers):", len(results))
	for _, r := range results {
		if r.Err != nil {
			t.Logf("  %-20s %-20s  UNREACHABLE (%v)", r.Region, r.City, r.Err)
		} else {
			t.Logf("  %-20s %-20s  %.0f ms  jitter %.0f ms  loss %.1f%%  score %d  %s",
				r.Region, r.City, r.LatencyMs, r.JitterMs, r.PacketLoss, r.Score, r.Rating)
		}
	}

	best := results[0]
	if best.Err == nil {
		assert.Greater(t, best.Score, 0, "best server should have a positive score")
	}
}

func TestDota2_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	g := game.NewDota2()
	p := &game.TCPGamePinger{}

	results, err := game.Check(ctx, g, p, 5)
	require.NoError(t, err)
	assert.NotEmpty(t, results)

	t.Logf("Dota 2 results (%d servers):", len(results))
	for _, r := range results {
		if r.Err != nil {
			t.Logf("  %-20s %-20s  UNREACHABLE", r.Region, r.City)
		} else {
			t.Logf("  %-20s %-20s  %.0f ms  score %d", r.Region, r.City, r.LatencyMs, r.Score)
		}
	}
}

func TestAllGames_HaveReachableServers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	p := &game.TCPGamePinger{}
	games := []game.Game{game.NewCS2(), game.NewDota2()}

	for _, g := range games {
		t.Run(g.Name(), func(t *testing.T) {
			results, err := game.Check(ctx, g, p, 3)
			require.NoError(t, err)

			reachable := 0
			for _, r := range results {
				if r.Err == nil {
					reachable++
				}
			}
			assert.Greater(t, reachable, 0,
				"%s: expected at least one reachable server, got 0 out of %d", g.Name(), len(results))
		})
	}
}
