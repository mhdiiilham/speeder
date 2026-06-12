package display

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/mhdiiilham/speeder/internal/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// Disable ANSI codes in tests so string assertions work cleanly.
	color.NoColor = true
}

func sampleResult() *runner.Result {
	return &runner.Result{
		Server: runner.Server{
			Hostname: "ndt-mlab1-sin01.example.org",
			City:     "Singapore",
			Country:  "SG",
		},
		LatencyMs: 14.2,
		JitterMs:  0.9,
		Download: runner.PhaseResult{
			SpeedMbps: 94.71,
			Bytes:     36_800_000,
			Duration:  3100 * time.Millisecond,
		},
		Upload: runner.PhaseResult{
			SpeedMbps: 23.10,
			Bytes:     8_100_000,
			Duration:  2800 * time.Millisecond,
		},
	}
}

func TestPrintResult(t *testing.T) {
	var buf bytes.Buffer
	PrintResult(&buf, sampleResult())
	out := buf.String()

	assert.Contains(t, out, "ndt-mlab1-sin01.example.org")
	assert.Contains(t, out, "Singapore")
	assert.Contains(t, out, "SG")
	assert.Contains(t, out, "14.2")
	assert.Contains(t, out, "0.9")
	assert.Contains(t, out, "94.71")
	assert.Contains(t, out, "23.10")
}

func TestPrintResult_NoLocation(t *testing.T) {
	var buf bytes.Buffer
	r := sampleResult()
	r.Server.City = ""
	r.Server.Country = ""
	PrintResult(&buf, r)
	out := buf.String()
	assert.NotContains(t, out, "Location:")
}

func TestPrintServerList(t *testing.T) {
	servers := []runner.Server{
		{Hostname: "host-a.example.org", City: "Singapore", Country: "SG"},
		{Hostname: "host-b.example.org", City: "Tokyo", Country: "JP"},
	}
	var buf bytes.Buffer
	PrintServerList(&buf, servers)
	out := buf.String()

	assert.Contains(t, out, "HOSTNAME")
	assert.Contains(t, out, "host-a.example.org")
	assert.Contains(t, out, "Singapore")
	assert.Contains(t, out, "host-b.example.org")
	assert.Contains(t, out, "Tokyo")
}

func TestPrintServerList_Empty(t *testing.T) {
	var buf bytes.Buffer
	PrintServerList(&buf, nil)
	out := buf.String()
	assert.Contains(t, out, "HOSTNAME")
}

func TestToJSON(t *testing.T) {
	j := ToJSON(sampleResult())
	assert.Equal(t, "ndt-mlab1-sin01.example.org", j.Server.Hostname)
	assert.Equal(t, "Singapore", j.Server.City)
	assert.Equal(t, "SG", j.Server.Country)
	assert.InDelta(t, 14.2, j.LatencyMs, 0.01)
	assert.InDelta(t, 0.9, j.JitterMs, 0.01)
	assert.InDelta(t, 94.71, j.DownloadMbps, 0.01)
	assert.InDelta(t, 23.10, j.UploadMbps, 0.01)
	assert.InDelta(t, 44.9, j.DataUsedMB, 0.1)
}

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	err := PrintJSON(&buf, sampleResult())
	require.NoError(t, err)

	var parsed JSONResult
	err = json.Unmarshal(buf.Bytes(), &parsed)
	require.NoError(t, err)

	assert.Equal(t, "ndt-mlab1-sin01.example.org", parsed.Server.Hostname)
	assert.InDelta(t, 94.71, parsed.DownloadMbps, 0.01)
}

func TestPrintJSON_ValidJSON(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, PrintJSON(&buf, sampleResult()))
	assert.True(t, json.Valid(buf.Bytes()))
}

func TestFormatDuration(t *testing.T) {
	assert.Equal(t, "3.1s", FormatDuration(3100*time.Millisecond))
	assert.Equal(t, "0.0s", FormatDuration(0))
	assert.True(t, strings.HasSuffix(FormatDuration(1*time.Second), "s"))
}
