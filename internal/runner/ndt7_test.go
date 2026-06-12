package runner

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// startDownloadServer runs a local WebSocket server that sends binary frames.
func startDownloadServer(t *testing.T, frames int, frameSize int) *httptest.Server {
	t.Helper()
	payload := make([]byte, frameSize)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for i := 0; i < frames; i++ {
			if err := conn.WriteMessage(websocket.BinaryMessage, payload); err != nil {
				return
			}
		}
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// startUploadServer runs a local WebSocket server that drains binary frames.
func startUploadServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func httpToWS(url string) string {
	return "ws" + strings.TrimPrefix(url, "http")
}

func TestNDT7Client_Download(t *testing.T) {
	srv := startDownloadServer(t, 10, 8192)
	wsURL := httpToWS(srv.URL)

	client := &NDT7Client{}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ch, err := client.Download(ctx, wsURL)
	require.NoError(t, err)

	var last Measurement
	for m := range ch {
		last = m
	}
	assert.Greater(t, last.NumBytes, int64(0), "should have received bytes")
}

func TestNDT7Client_Download_ContextCancel(t *testing.T) {
	// Server streams forever.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _ := wsUpgrader.Upgrade(w, r, nil)
		defer conn.Close()
		payload := make([]byte, 8192)
		for {
			if err := conn.WriteMessage(websocket.BinaryMessage, payload); err != nil {
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	client := &NDT7Client{}
	ch, err := client.Download(ctx, httpToWS(srv.URL))
	require.NoError(t, err)

	// Cancel after first measurement.
	<-ch
	cancel()

	// Drain until closed.
	for range ch {
	}
}

func TestNDT7Client_Upload(t *testing.T) {
	srv := startUploadServer(t)
	wsURL := httpToWS(srv.URL)

	client := &NDT7Client{}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	ch, err := client.Upload(ctx, wsURL)
	require.NoError(t, err)

	var last Measurement
	for m := range ch {
		last = m
	}
	assert.Greater(t, last.NumBytes, int64(0), "should have sent bytes")
}

func TestNDT7Client_Download_DialError(t *testing.T) {
	client := &NDT7Client{}
	_, err := client.Download(context.Background(), "wss://127.0.0.1:1/ndt/v7/download")
	assert.Error(t, err)
}

func TestNDT7Client_Upload_DialError(t *testing.T) {
	client := &NDT7Client{}
	_, err := client.Upload(context.Background(), "wss://127.0.0.1:1/ndt/v7/upload")
	assert.Error(t, err)
}

func TestDialNDT7_Error(t *testing.T) {
	_, err := dialNDT7("wss://127.0.0.1:1/ndt/v7/download")
	assert.Error(t, err)
}
