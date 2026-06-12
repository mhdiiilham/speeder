package runner

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const ndt7Subprotocol = "net.measurementlab.ndt.v7"

// NDT7Client implements SpeedClient using the ndt7 WebSocket protocol.
type NDT7Client struct{}

// Download opens a ndt7 download WebSocket and streams Measurements every 250 ms.
// The caller cancels ctx to stop early; the channel is closed on exit.
func (c *NDT7Client) Download(ctx context.Context, url string) (<-chan Measurement, error) {
	conn, err := dialNDT7(url)
	if err != nil {
		return nil, err
	}

	ch := make(chan Measurement, 64)
	var bytesReceived int64

	// Count bytes from incoming binary frames.
	go func() {
		defer conn.Close()
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if mt == websocket.BinaryMessage {
				atomic.AddInt64(&bytesReceived, int64(len(msg)))
			}
		}
	}()

	// Emit a measurement every 250 ms; close the connection when ctx is done.
	go func() {
		defer close(ch)
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		start := time.Now()

		for {
			select {
			case <-ctx.Done():
				conn.WriteMessage( //nolint:errcheck
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
				)
				conn.Close()
				return
			case t := <-ticker.C:
				ch <- Measurement{
					NumBytes:  atomic.LoadInt64(&bytesReceived),
					ElapsedMs: t.Sub(start).Milliseconds(),
				}
			}
		}
	}()

	return ch, nil
}

// Upload opens a ndt7 upload WebSocket, sends data as fast as possible, and
// streams Measurements every 250 ms. The caller cancels ctx to stop early.
func (c *NDT7Client) Upload(ctx context.Context, url string) (<-chan Measurement, error) {
	conn, err := dialNDT7(url)
	if err != nil {
		return nil, err
	}

	ch := make(chan Measurement, 64)
	var bytesSent int64
	payload := make([]byte, 32768)

	// Send binary frames until ctx is cancelled.
	go func() {
		for ctx.Err() == nil {
			if err := conn.WriteMessage(websocket.BinaryMessage, payload); err != nil {
				return
			}
			atomic.AddInt64(&bytesSent, int64(len(payload)))
		}
	}()

	// Drain server-side measurements (required by the protocol to avoid backpressure).
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	// Emit a measurement every 250 ms.
	go func() {
		defer close(ch)
		defer conn.Close()
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		start := time.Now()

		for {
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				ch <- Measurement{
					NumBytes:  atomic.LoadInt64(&bytesSent),
					ElapsedMs: t.Sub(start).Milliseconds(),
				}
			}
		}
	}()

	return ch, nil
}

func dialNDT7(url string) (*websocket.Conn, error) {
	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second}
	header := http.Header{"Sec-WebSocket-Protocol": {ndt7Subprotocol}}
	conn, _, err := dialer.Dial(url, header)
	return conn, err
}
