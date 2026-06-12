package runner

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fakeLocateResponse(machines []string) locateResponse {
	lr := locateResponse{}
	for _, m := range machines {
		lr.Results = append(lr.Results, locateResult{
			Machine: m,
			Location: locateLocation{City: "Singapore", Country: "SG"},
			URLs: map[string]string{
				"wss:///ndt/v7/download": "wss://" + m + "/ndt/v7/download",
				"wss:///ndt/v7/upload":   "wss://" + m + "/ndt/v7/upload",
			},
		})
	}
	return lr
}

func TestHTTPLocateClient_Nearest_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(fakeLocateResponse([]string{
			"ndt-mlab1-sin01.example.org",
			"ndt-mlab2-sin01.example.org",
		}))
	}))
	defer srv.Close()

	client := &HTTPLocateClient{URL: srv.URL}
	servers, err := client.Nearest(context.Background())
	require.NoError(t, err)
	require.Len(t, servers, 2)
	assert.Equal(t, "ndt-mlab1-sin01.example.org", servers[0].Hostname)
	assert.Equal(t, "Singapore", servers[0].City)
	assert.Equal(t, "SG", servers[0].Country)
	assert.Contains(t, servers[0].DownloadURL, "wss://")
	assert.Contains(t, servers[0].UploadURL, "wss://")
}

func TestHTTPLocateClient_Nearest_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	client := &HTTPLocateClient{URL: srv.URL}
	_, err := client.Nearest(context.Background())
	assert.ErrorContains(t, err, "503")
}

func TestHTTPLocateClient_Nearest_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	client := &HTTPLocateClient{URL: srv.URL}
	_, err := client.Nearest(context.Background())
	assert.ErrorContains(t, err, "decode locate response")
}

func TestHTTPLocateClient_Nearest_EmptyResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(locateResponse{Results: nil})
	}))
	defer srv.Close()

	client := &HTTPLocateClient{URL: srv.URL}
	_, err := client.Nearest(context.Background())
	assert.ErrorContains(t, err, "no usable ndt7 servers")
}

func TestHTTPLocateClient_Nearest_MissingURLs(t *testing.T) {
	// Results with no wss:// URLs should be skipped.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lr := locateResponse{
			Results: []locateResult{
				{Machine: "host.example.org", URLs: map[string]string{}}, // no wss URLs
			},
		}
		json.NewEncoder(w).Encode(lr)
	}))
	defer srv.Close()

	client := &HTTPLocateClient{URL: srv.URL}
	_, err := client.Nearest(context.Background())
	assert.ErrorContains(t, err, "no usable ndt7 servers")
}

func TestHTTPLocateClient_Nearest_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Never responds — client will cancel first.
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	client := &HTTPLocateClient{URL: srv.URL}
	_, err := client.Nearest(ctx)
	assert.Error(t, err)
}

func TestHTTPLocateClient_DefaultEndpoint(t *testing.T) {
	client := &HTTPLocateClient{}
	assert.Equal(t, defaultLocateURL, client.endpoint())
}

func TestHTTPLocateClient_CustomEndpoint(t *testing.T) {
	client := &HTTPLocateClient{URL: "https://custom.example.com"}
	assert.Equal(t, "https://custom.example.com", client.endpoint())
}
