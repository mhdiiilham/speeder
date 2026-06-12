package runner

import (
	"context"
	"net"
	"time"
)

// TCPPinger measures round-trip latency by timing TCP connection establishment.
// One SYN + SYN-ACK ≈ one RTT.
type TCPPinger struct {
	// Port overrides the default port 443 (used in tests).
	Port string
}

func (p *TCPPinger) addr(host string) string {
	port := p.Port
	if port == "" {
		port = "443"
	}
	return net.JoinHostPort(host, port)
}

// Ping dials host n times and returns the mean RTT and jitter (population stddev).
func (p *TCPPinger) Ping(ctx context.Context, host string, n int) (avg, jitter time.Duration, err error) {
	addr := p.addr(host)
	d := &net.Dialer{}
	rtts := make([]time.Duration, 0, n)

	for i := 0; i < n; i++ {
		if ctx.Err() != nil {
			break
		}
		start := time.Now()
		conn, dialErr := d.DialContext(ctx, "tcp", addr)
		if dialErr != nil {
			if i == 0 {
				return 0, 0, dialErr
			}
			break
		}
		rtt := time.Since(start)
		conn.Close()
		rtts = append(rtts, rtt)
	}

	if len(rtts) == 0 {
		return 0, 0, nil
	}
	return durationMean(rtts), durationJitter(rtts), nil
}
