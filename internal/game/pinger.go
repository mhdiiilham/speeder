package game

import (
	"context"
	"errors"
	"math"
	"net"
	"time"
)


// Pinger measures latency, jitter, and packet loss to a host:port.
type Pinger interface {
	Ping(ctx context.Context, host, port string, n int) (avg, jitter time.Duration, lossPercent float64, err error)
}

// TCPGamePinger dials TCP connections to estimate latency and packet loss.
// One successful connect + close ≈ one RTT; a timeout counts as packet loss.
type TCPGamePinger struct {
	// DialTimeout is the per-attempt timeout. Defaults to 2 s.
	DialTimeout time.Duration
}

func (p *TCPGamePinger) dialTimeout() time.Duration {
	if p.DialTimeout > 0 {
		return p.DialTimeout
	}
	return 2 * time.Second
}

// Ping sends n TCP connection probes to host:port and returns statistics.
func (p *TCPGamePinger) Ping(ctx context.Context, host, port string, n int) (avg, jitter time.Duration, lossPercent float64, err error) {
	addr := net.JoinHostPort(host, port)
	d := net.Dialer{Timeout: p.dialTimeout()}

	rtts := make([]time.Duration, 0, n)
	failures := 0

	for i := 0; i < n; i++ {
		if ctx.Err() != nil {
			break
		}
		start := time.Now()
		conn, dialErr := d.DialContext(ctx, "tcp", addr)
		rtt := time.Since(start)

		if dialErr != nil {
			// Both timeouts and connection-refused count as failures.
			// A refused RST often comes from a nearby firewall in microseconds,
			// not from the actual server, so it is not a valid RTT sample.
			failures++
			continue
		}
		conn.Close()
		rtts = append(rtts, rtt)
	}

	total := len(rtts) + failures
	if total == 0 {
		return 0, 0, 100, errors.New("all probes failed")
	}

	lossPercent = float64(failures) / float64(total) * 100.0
	if len(rtts) == 0 {
		return 0, 0, lossPercent, errors.New("all probes failed")
	}

	avg = durationMean(rtts)
	jitter = durationJitter(rtts)
	return avg, jitter, lossPercent, nil
}

// durationMean returns the arithmetic mean of a slice of durations.
func durationMean(ds []time.Duration) time.Duration {
	if len(ds) == 0 {
		return 0
	}
	var sum time.Duration
	for _, d := range ds {
		sum += d
	}
	return sum / time.Duration(len(ds))
}

// durationJitter returns the population standard deviation of a duration slice.
func durationJitter(ds []time.Duration) time.Duration {
	if len(ds) < 2 {
		return 0
	}
	mean := float64(durationMean(ds))
	var variance float64
	for _, d := range ds {
		diff := float64(d) - mean
		variance += diff * diff
	}
	return time.Duration(math.Sqrt(variance / float64(len(ds))))
}
