package display

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mhdiiilham/speeder/internal/game"
	"github.com/mhdiiilham/speeder/internal/history"
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
	fmt.Fprintln(w)
	bold.Fprintf(w, "  Server:    %s\n", r.Server.Hostname)
	if r.Server.City != "" || r.Server.Country != "" {
		dim.Fprintf(w, "  Location:  %s %s\n", r.Server.City, r.Server.Country)
	}
	if r.ISP != "" || r.ClientIP != "" {
		dim.Fprintf(w, "  ISP:       %s  (%s)\n", r.ISP, r.ClientIP)
	}
	fmt.Fprintln(w)
	cyan.Fprintf(w, "  Latency:   %.1f ms", r.LatencyMs)
	dim.Fprintf(w, "  jitter: %.1f ms\n", r.JitterMs)

	if !r.PingOnly() {
		dlMB := float64(r.Download.Bytes) / 1e6
		ulMB := float64(r.Upload.Bytes) / 1e6
		green.Fprintf(w, "  Download:  %.2f Mbps", r.Download.SpeedMbps)
		dim.Fprintf(w, "  (%.1fs, %.1f MB)\n", r.Download.Duration.Seconds(), dlMB)
		amber.Fprintf(w, "  Upload:    %.2f Mbps", r.Upload.SpeedMbps)
		dim.Fprintf(w, "  (%.1fs, %.1f MB)\n", r.Upload.Duration.Seconds(), ulMB)
		fmt.Fprintln(w)
		dim.Fprintf(w, "  Data used: %.1f MB\n", float64(r.Download.Bytes+r.Upload.Bytes)/1e6)
	}
	fmt.Fprintln(w)
}

// PrintWatchHeader prints a separator line before each watch iteration.
func PrintWatchHeader(w io.Writer, run int, t time.Time) {
	fmt.Fprintln(w)
	bold.Fprintf(w, "  Run %d  •  %s\n", run, t.Format("Jan 02 15:04:05"))
	fmt.Fprintln(w, "  "+strings.Repeat("─", 40))
}

// PrintHistory writes a summary table of historical records to w.
func PrintHistory(w io.Writer, records []history.Record, total int) {
	if len(records) == 0 {
		fmt.Fprintln(w)
		dim.Fprint(w, "  No history yet. Run speeder to record your first result.\n\n")
		return
	}
	const colFmt = "  %-17s  %-22s  %10s  %9s  %8s\n"
	fmt.Fprintln(w)
	bold.Fprintf(w, "  Speed Test History\n")
	fmt.Fprintln(w)
	dim.Fprintf(w, colFmt, "TIME", "SERVER", "DOWNLOAD", "UPLOAD", "LATENCY")
	fmt.Fprintln(w, "  "+strings.Repeat("─", 75))
	for _, r := range records {
		dl := fmt.Sprintf("%.1f Mbps", r.DownloadMbps)
		ul := fmt.Sprintf("%.1f Mbps", r.UploadMbps)
		lat := fmt.Sprintf("%.1f ms", r.LatencyMs)
		if r.PingOnly {
			dl, ul = "—", "—"
		}
		server := r.Server
		if len(server) > 22 {
			server = server[:19] + "..."
		}
		fmt.Fprintf(w, colFmt,
			r.Timestamp.Local().Format("Jan 02 15:04"),
			server, dl, ul, lat)
	}
	fmt.Fprintln(w, "  "+strings.Repeat("─", 75))
	if total > len(records) {
		dim.Fprintf(w, "\n  Showing last %d of %d results.\n\n", len(records), total)
	} else {
		dim.Fprintf(w, "\n  %d result(s) total.\n\n", total)
	}
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
	ISP          string  `json:"isp,omitempty"`
	ClientIP     string  `json:"client_ip,omitempty"`
	LatencyMs    float64 `json:"latency_ms"`
	JitterMs     float64 `json:"jitter_ms"`
	DownloadMbps float64 `json:"download_mbps,omitempty"`
	UploadMbps   float64 `json:"upload_mbps,omitempty"`
	DataUsedMB   float64 `json:"data_used_mb,omitempty"`
}

// ToJSON converts a runner.Result to the JSON wire format.
func ToJSON(r *runner.Result) JSONResult {
	j := JSONResult{
		ISP:          r.ISP,
		ClientIP:     r.ClientIP,
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

// MarshalHistoryRecord encodes a history record as JSON bytes.
func MarshalHistoryRecord(r history.Record) ([]byte, error) {
	return json.Marshal(r)
}

// FormatDuration returns a human-friendly elapsed time string.
func FormatDuration(d time.Duration) string {
	return fmt.Sprintf("%.1fs", d.Seconds())
}

// PrintGameResults writes the server latency table and verdict for a game check.
func PrintGameResults(w io.Writer, g game.Game, results []game.PingResult) {
	const colFmt = "  %-22s  %7s  %6s  %5s  %3s  %s\n"
	divider := "  " + strings.Repeat("─", 60)

	fmt.Fprintln(w)
	bold.Fprintf(w, "  %s Server Latency\n", g.Name())
	fmt.Fprintln(w)
	dim.Fprintf(w, colFmt, "SERVER", "PING", "JITTER", "LOSS", "SCR", "STATUS")
	fmt.Fprintln(w, divider)

	for _, r := range results {
		serverLabel := fmt.Sprintf("%s [%s]", r.City, r.Region)
		if r.Err != nil {
			dim.Fprintf(w, colFmt, serverLabel, "—", "—", "—", "—", "Unreachable")
			continue
		}

		latStr := fmt.Sprintf("%d ms", int(r.LatencyMs))
		jitStr := fmt.Sprintf("%d ms", int(r.JitterMs))
		lossStr := fmt.Sprintf("%.0f%%", r.PacketLoss)
		scoreStr := fmt.Sprintf("%d", r.Score)

		statusStr := string(r.Rating)
		if r.Best {
			statusStr += " ✓"
		}

		colorFn := ratingColor(r.Rating)
		colorFn.Fprintf(w, colFmt, serverLabel, latStr, jitStr, lossStr, scoreStr, statusStr)
	}

	fmt.Fprintln(w, divider)

	// Verdict
	var best game.PingResult
	if len(results) > 0 {
		best = results[0]
	}
	fmt.Fprintln(w)
	bold.Fprintf(w, "  Verdict: ")
	fmt.Fprintln(w, game.Verdict(g.Name(), best))

	// Game-specific note
	if note := g.Note(); note != "" {
		fmt.Fprintln(w)
		dim.Fprintf(w, "  %s\n", note)
	}
	fmt.Fprintln(w)
}

func ratingColor(r game.Rating) *color.Color {
	switch r {
	case game.RatingExcellent:
		return green
	case game.RatingGood:
		return color.New(color.FgGreen)
	case game.RatingPlayable:
		return amber
	case game.RatingPoor:
		return color.New(color.FgRed)
	default:
		return dim
	}
}
