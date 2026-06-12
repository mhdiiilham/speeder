package display

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mhdiiilham/speeder/internal/runner"
)

var (
	bold  = color.New(color.Bold)
	cyan  = color.New(color.FgCyan, color.Bold)
	green = color.New(color.FgGreen, color.Bold)
	amber = color.New(color.FgYellow, color.Bold)
	dim   = color.New(color.Faint)
)

// PrintResult writes a human-readable summary of r to w.
func PrintResult(w io.Writer, r *runner.Result) {
	dlMB := float64(r.Download.Bytes) / 1e6
	ulMB := float64(r.Upload.Bytes) / 1e6
	totalMB := dlMB + ulMB

	fmt.Fprintln(w)
	bold.Fprintf(w, "  Server:    %s\n", r.Server.Hostname)
	if r.Server.City != "" || r.Server.Country != "" {
		dim.Fprintf(w, "  Location:  %s %s\n", r.Server.City, r.Server.Country)
	}
	fmt.Fprintln(w)
	cyan.Fprintf(w, "  Latency:   %.1f ms", r.LatencyMs)
	dim.Fprintf(w, "  jitter: %.1f ms\n", r.JitterMs)
	green.Fprintf(w, "  Download:  %.2f Mbps", r.Download.SpeedMbps)
	dim.Fprintf(w, "  (%.1fs, %.1f MB)\n", r.Download.Duration.Seconds(), dlMB)
	amber.Fprintf(w, "  Upload:    %.2f Mbps", r.Upload.SpeedMbps)
	dim.Fprintf(w, "  (%.1fs, %.1f MB)\n", r.Upload.Duration.Seconds(), ulMB)
	fmt.Fprintln(w)
	dim.Fprintf(w, "  Data used: %.1f MB\n", totalMB)
	fmt.Fprintln(w)
}

// PrintServerList writes a table of servers to w.
func PrintServerList(w io.Writer, servers []runner.Server) {
	const colFmt = "  %-55s  %-20s  %s\n"
	bold.Fprintf(w, colFmt, "HOSTNAME", "CITY", "COUNTRY")
	fmt.Fprintln(w, "  "+strings.Repeat("-", 85))
	for _, s := range servers {
		fmt.Fprintf(w, colFmt, s.Hostname, s.City, s.Country)
	}
	fmt.Fprintln(w)
}

// JSONResult is the stable JSON wire format for a test result.
type JSONResult struct {
	Server struct {
		Hostname string `json:"hostname"`
		City     string `json:"city"`
		Country  string `json:"country"`
	} `json:"server"`
	LatencyMs    float64 `json:"latency_ms"`
	JitterMs     float64 `json:"jitter_ms"`
	DownloadMbps float64 `json:"download_mbps"`
	UploadMbps   float64 `json:"upload_mbps"`
	DataUsedMB   float64 `json:"data_used_mb"`
}

// ToJSON converts a runner.Result to the JSON wire format.
func ToJSON(r *runner.Result) JSONResult {
	j := JSONResult{
		LatencyMs:    r.LatencyMs,
		JitterMs:     r.JitterMs,
		DownloadMbps: r.Download.SpeedMbps,
		UploadMbps:   r.Upload.SpeedMbps,
		DataUsedMB:   float64(r.Download.Bytes+r.Upload.Bytes) / 1e6,
	}
	j.Server.Hostname = r.Server.Hostname
	j.Server.City = r.Server.City
	j.Server.Country = r.Server.Country
	return j
}

// PrintJSON encodes r as JSON to w.
func PrintJSON(w io.Writer, r *runner.Result) error {
	return json.NewEncoder(w).Encode(ToJSON(r))
}

// FormatDuration returns a human-friendly elapsed time string.
func FormatDuration(d time.Duration) string {
	return fmt.Sprintf("%.1fs", d.Seconds())
}
