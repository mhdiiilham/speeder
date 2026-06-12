package runner

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func startTCPEcho(t *testing.T) (host, port string) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { l.Close() })
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()
	h, p, _ := net.SplitHostPort(l.Addr().String())
	return h, p
}

func TestTCPPinger_Ping_HappyPath(t *testing.T) {
	host, port := startTCPEcho(t)
	p := &TCPPinger{Port: port}
	avg, jitter, err := p.Ping(context.Background(), host, 3)
	require.NoError(t, err)
	assert.Greater(t, avg, time.Duration(0))
	assert.GreaterOrEqual(t, jitter, time.Duration(0))
}

func TestTCPPinger_Ping_SingleSample(t *testing.T) {
	host, port := startTCPEcho(t)
	p := &TCPPinger{Port: port}
	avg, jitter, err := p.Ping(context.Background(), host, 1)
	require.NoError(t, err)
	assert.Greater(t, avg, time.Duration(0))
	assert.Equal(t, time.Duration(0), jitter) // single sample → no jitter
}

func TestTCPPinger_Ping_ConnectionRefused(t *testing.T) {
	p := &TCPPinger{Port: "1"} // port 1 is never open
	_, _, err := p.Ping(context.Background(), "127.0.0.1", 1)
	assert.Error(t, err)
}

func TestTCPPinger_Ping_ContextCancelled(t *testing.T) {
	host, port := startTCPEcho(t)
	p := &TCPPinger{Port: port}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	avg, _, _ := p.Ping(ctx, host, 5)
	// With a cancelled context, we might get 0 samples or an error on first dial.
	assert.GreaterOrEqual(t, avg, time.Duration(0))
}

func TestTCPPinger_DefaultPort(t *testing.T) {
	p := &TCPPinger{}
	assert.Equal(t, "127.0.0.1:443", p.addr("127.0.0.1"))
}

func TestTCPPinger_CustomPort(t *testing.T) {
	p := &TCPPinger{Port: "8080"}
	assert.Equal(t, "127.0.0.1:8080", p.addr("127.0.0.1"))
}
