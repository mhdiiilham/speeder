// Package game provides per-game server latency checking with packet-loss
// and jitter measurement to help players gauge connection quality before
// queuing into competitive matches.
package game

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// Rating describes the gameplay quality of a connection.
type Rating string

const (
	RatingExcellent Rating = "Excellent"
	RatingGood      Rating = "Good"
	RatingPlayable  Rating = "Playable"
	RatingPoor      Rating = "Poor"
	RatingVeryPoor  Rating = "Very Poor"
	RatingUnknown   Rating = "Unknown"
)

// Server is a single pingable game server endpoint.
type Server struct {
	Region string
	City   string
	Host   string
	Port   string
}

// PingResult holds all measurements for one server.
type PingResult struct {
	Region     string
	City       string
	LatencyMs  float64
	JitterMs   float64
	PacketLoss float64 // 0–100 percent
	Score      int     // 0–100 composite gaming score
	Rating     Rating
	Best       bool  // true for the recommended server
	Err        error // non-nil means the server was unreachable
}

// Game is a source of server endpoints for a specific title.
type Game interface {
	Name() string
	Servers(ctx context.Context) ([]Server, error)
	// Note returns an optional caveat shown below results. Return "" for none.
	Note() string
}

// Verdict returns a human-readable recommendation for the best result.
func Verdict(gameName string, best PingResult) string {
	if best.Err != nil {
		return "Could not reach any servers — check your internet connection."
	}
	switch {
	case best.Score >= 85:
		return fmt.Sprintf("Ready to play! Your connection to %s is excellent.", best.City)
	case best.Score >= 70:
		return fmt.Sprintf("Good to play on %s. Slight disadvantage in close duels at %.0f ms.", best.City, best.LatencyMs)
	case best.Score >= 50:
		return fmt.Sprintf("Playable on %s (%.0f ms) but you may notice lag. Try playing off-peak.", best.City, best.LatencyMs)
	case best.Score >= 30:
		return fmt.Sprintf("High latency to %s (%.0f ms) — avoid ranked if possible.", best.City, best.LatencyMs)
	default:
		return fmt.Sprintf("Very poor connection to all %s servers — not recommended for competitive play.", gameName)
	}
}

// Rate returns a Rating based purely on latency.
func Rate(latencyMs float64) Rating {
	switch {
	case latencyMs < 0:
		return RatingUnknown
	case latencyMs < 30:
		return RatingExcellent
	case latencyMs < 60:
		return RatingGood
	case latencyMs < 80:
		return RatingPlayable
	case latencyMs < 120:
		return RatingPoor
	default:
		return RatingVeryPoor
	}
}

// Score computes a 0–100 composite gaming score from latency, jitter, and
// packet loss. Higher is better.
//
//   - Latency contributes up to 50 points. ≤ 20 ms = full score.
//   - Jitter contributes up to 30 points.  ≤ 2 ms  = full score.
//   - Packet loss contributes up to 20 points. 0 % = full score.
func Score(latencyMs, jitterMs, lossPercent float64) int {
	if latencyMs < 0 {
		return 0
	}
	latencyPenalty := clamp(latencyMs-20, 0, 80) * (50.0 / 80.0)
	jitterPenalty := clamp(jitterMs-2, 0, 48) * (30.0 / 48.0)
	lossPenalty := clamp(lossPercent, 0, 10) * (20.0 / 10.0)

	score := 100.0 - latencyPenalty - jitterPenalty - lossPenalty
	return int(clamp(score, 0, 100))
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Check pings all servers for g concurrently and returns results sorted
// from best to worst. The best server has its Best field set to true.
func Check(ctx context.Context, g Game, p Pinger, pingsPerServer int) ([]PingResult, error) {
	servers, err := g.Servers(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch %s servers: %w", g.Name(), err)
	}
	if len(servers) == 0 {
		return nil, fmt.Errorf("no %s servers available", g.Name())
	}

	results := make([]PingResult, len(servers))
	var wg sync.WaitGroup

	for i, srv := range servers {
		wg.Add(1)
		go func(i int, srv Server) {
			defer wg.Done()
			avg, jitter, loss, pingErr := p.Ping(ctx, srv.Host, srv.Port, pingsPerServer)
			if pingErr != nil {
				results[i] = PingResult{
					Region: srv.Region,
					City:   srv.City,
					Rating: RatingUnknown,
					Err:    pingErr,
				}
				return
			}
			latMs := float64(avg.Microseconds()) / 1000.0
			jitMs := float64(jitter.Microseconds()) / 1000.0
			sc := Score(latMs, jitMs, loss)
			results[i] = PingResult{
				Region:     srv.Region,
				City:       srv.City,
				LatencyMs:  latMs,
				JitterMs:   jitMs,
				PacketLoss: loss,
				Score:      sc,
				Rating:     Rate(latMs),
			}
		}(i, srv)
	}
	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		ei, ej := results[i].Err != nil, results[j].Err != nil
		if ei != ej {
			return !ei // errors go last
		}
		if ei && ej {
			return false
		}
		return results[i].Score > results[j].Score
	})

	if len(results) > 0 && results[0].Err == nil {
		results[0].Best = true
	}

	return results, nil
}
