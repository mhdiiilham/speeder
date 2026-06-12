package ipinfo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetch_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`{"ip":"203.115.1.1","org":"AS4818 Maxis Broadband Sdn Bhd"}`))
	}))
	defer srv.Close()

	c := &Client{URL: srv.URL}
	info, err := c.Fetch(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "203.115.1.1", info.IP)
	assert.Equal(t, "Maxis Broadband Sdn Bhd", info.ISP)
}

func TestFetch_NoASPrefix(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`{"ip":"1.2.3.4","org":"Some ISP"}`))
	}))
	defer srv.Close()

	c := &Client{URL: srv.URL}
	info, err := c.Fetch(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Some ISP", info.ISP)
}

func TestFetch_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := &Client{URL: srv.URL}
	_, err := c.Fetch(context.Background())
	assert.ErrorContains(t, err, "429")
}

func TestFetch_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	c := &Client{URL: srv.URL}
	_, err := c.Fetch(context.Background())
	assert.ErrorContains(t, err, "decode")
}

func TestFetch_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := &Client{URL: srv.URL}
	_, err := c.Fetch(ctx)
	assert.Error(t, err)
}

func TestParseISP(t *testing.T) {
	assert.Equal(t, "Maxis Broadband Sdn Bhd", parseISP("AS4818 Maxis Broadband Sdn Bhd"))
	assert.Equal(t, "Plain ISP", parseISP("Plain ISP"))
	assert.Equal(t, "", parseISP(""))
}

func TestDefaultEndpoint(t *testing.T) {
	c := &Client{}
	assert.Equal(t, defaultURL, c.endpoint())
}

func TestCustomEndpoint(t *testing.T) {
	c := &Client{URL: "https://custom.example.com"}
	assert.Equal(t, "https://custom.example.com", c.endpoint())
}
