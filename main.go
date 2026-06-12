package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/mhdiiilham/speeder/internal/display"
	"github.com/mhdiiilham/speeder/internal/game"
	"github.com/mhdiiilham/speeder/internal/history"
	"github.com/mhdiiilham/speeder/internal/ipinfo"
	"github.com/mhdiiilham/speeder/internal/runner"
)

// version is set at build time via -ldflags "-X main.version=vX.Y.Z" (release CI).
// For "go install pkg@version" builds, ldflags are unavailable so we fall back to
// the module version that the Go toolchain embeds automatically in the binary.
var version = "dev"

func getVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	// history subcommand
	if len(args) > 0 && args[0] == "history" {
		return runHistory(args[1:])
	}

	fs := flag.NewFlagSet("speeder", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	flagList        := fs.Bool("list", false, "list nearby M-Lab servers and exit")
	flagServer      := fs.String("server", "", "use a specific server hostname")
	flagJSON         := fs.Bool("json", false, "output results as JSON")
	flagNoProgress  := fs.Bool("no-progress", false, "disable live progress updates")
	flagDuration    := fs.Int("duration", 8, "max seconds per test phase")
	flagQuick       := fs.Bool("quick", false, "quick preset: 4s max, minimal data usage")
	flagGame        := fs.String("game", "", "check game server latency: cs2, dota2")
	flagPingOnly    := fs.Bool("ping-only", false, "measure latency only, skip download/upload")
	flagFailBelow   := fs.String("fail-if-below", "", "exit 1 if download is below threshold (e.g. 50 or 50mbps)")
	flagWatch       := fs.String("watch", "", "repeat test on an interval (e.g. 5m, 30s)")
	flagVersion     := fs.Bool("version", false, "print version and exit")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *flagVersion {
		fmt.Fprintf(os.Stdout, "speeder %s\n", getVersion())
		return 0
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if *flagGame != "" {
		return runGameCheck(ctx, *flagGame)
	}

	if *flagList {
		servers, err := runner.Default().ListServers(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
		display.PrintServerList(os.Stdout, servers)
		return 0
	}

	threshold, err := parseThreshold(*flagFailBelow)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: --fail-if-below: %v\n", err)
		return 2
	}

	watchInterval, err := parseWatch(*flagWatch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: --watch: %v\n", err)
		return 2
	}

	cfg := buildConfig(*flagQuick, *flagDuration, *flagServer, *flagNoProgress, *flagJSON, *flagPingOnly)

	if watchInterval > 0 {
		return runWatch(ctx, cfg, watchInterval, threshold, *flagJSON)
	}
	return runOnce(ctx, cfg, threshold, *flagJSON)
}

func runOnce(ctx context.Context, cfg runner.Config, threshold float64, jsonMode bool) int {
	fmt.Fprintf(os.Stderr, "  Finding nearest M-Lab server...\n")

	// Fetch ISP info in parallel.
	type ispResult struct{ isp, ip string }
	ispCh := make(chan ispResult, 1)
	go func() {
		info, _ := ipinfo.DefaultClient.Fetch(ctx)
		if info != nil {
			ispCh <- ispResult{info.ISP, info.IP}
		} else {
			ispCh <- ispResult{}
		}
	}()

	result, err := runner.Default().Run(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	isp := <-ispCh
	result.ISP = isp.isp
	result.ClientIP = isp.ip

	// Save to history (best effort — never block on failure).
	go saveHistory(result)

	if jsonMode {
		if err := display.PrintJSON(os.Stdout, result); err != nil {
			fmt.Fprintf(os.Stderr, "error encoding JSON: %v\n", err)
			return 1
		}
	} else {
		display.PrintResult(os.Stdout, result)
	}

	if threshold > 0 && result.Download.SpeedMbps < threshold {
		fmt.Fprintf(os.Stderr, "  Download %.2f Mbps is below threshold %.2f Mbps\n",
			result.Download.SpeedMbps, threshold)
		return 1
	}
	return 0
}

func runWatch(ctx context.Context, cfg runner.Config, interval time.Duration, threshold float64, jsonMode bool) int {
	run := 0
	for {
		run++
		if !jsonMode {
			display.PrintWatchHeader(os.Stdout, run, time.Now())
		}
		code := runOnce(ctx, cfg, threshold, jsonMode)
		if ctx.Err() != nil {
			return 0
		}
		if code != 0 && threshold > 0 {
			return code
		}
		select {
		case <-ctx.Done():
			return 0
		case <-time.After(interval):
		}
	}
}

func runGameCheck(ctx context.Context, gameName string) int {
	g, err := resolveGame(gameName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		fmt.Fprintf(os.Stderr, "  Available games: cs2, dota2\n")
		return 1
	}
	fmt.Fprintf(os.Stderr, "  Pinging %s servers...\n", g.Name())
	p := &game.TCPGamePinger{}
	results, err := game.Check(ctx, g, p, 10)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	const maxServers = 5
	if len(results) > maxServers {
		results = results[:maxServers]
	}
	display.PrintGameResults(os.Stdout, g, results)
	return 0
}

func runHistory(args []string) int {
	fs := flag.NewFlagSet("speeder history", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	flagLast := fs.Int("last", 10, "number of results to show")
	flagJSON := fs.Bool("json", false, "output raw JSONL")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	all, err := history.Load(0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	if *flagJSON {
		for _, r := range all {
			line, _ := display.MarshalHistoryRecord(r)
			fmt.Println(string(line))
		}
		return 0
	}

	n := *flagLast
	shown := all
	if n > 0 && len(all) > n {
		shown = all[len(all)-n:]
	}
	display.PrintHistory(os.Stdout, shown, len(all))
	return 0
}

func saveHistory(r *runner.Result) {
	rec := history.Record{
		Timestamp:    time.Now().UTC(),
		Server:       r.Server.Hostname,
		Location:     strings.TrimSpace(r.Server.City + " " + r.Server.Country),
		ISP:          r.ISP,
		LatencyMs:    r.LatencyMs,
		JitterMs:     r.JitterMs,
		DownloadMbps: r.Download.SpeedMbps,
		UploadMbps:   r.Upload.SpeedMbps,
		DataUsedMB:   float64(r.Download.Bytes+r.Upload.Bytes) / 1e6,
		PingOnly:     r.PingOnly(),
	}
	_ = history.Save(rec)
}

func resolveGame(name string) (game.Game, error) {
	switch strings.ToLower(name) {
	case "cs2", "cs", "counter-strike":
		return game.NewCS2(), nil
	case "dota2", "dota":
		return game.NewDota2(), nil
	default:
		return nil, fmt.Errorf("unknown game %q", name)
	}
}

func buildConfig(quick bool, duration int, server string, noProgress, jsonMode, pingOnly bool) runner.Config {
	var cfg runner.Config
	if quick {
		cfg = runner.QuickConfig()
	} else {
		cfg = runner.DefaultConfig()
		cfg.MaxDuration = time.Duration(duration) * time.Second
	}
	cfg.ServerHost = server
	cfg.PingOnly = pingOnly

	if !noProgress && !jsonMode {
		cfg.OnProgress = func(phase string, mbps float64) {
			label := "Download"
			if phase == "upload" {
				label = "Upload  "
			}
			fmt.Fprintf(os.Stderr, "\r  %s  %6.2f Mbps", label, mbps)
		}
	}
	return cfg
}

// parseThreshold parses "50", "50mbps", "50mb/s" → 50.0. Empty string → 0.
func parseThreshold(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimSuffix(s, "mbps")
	s = strings.TrimSuffix(s, "mb/s")
	s = strings.TrimSuffix(s, "mbp")
	s = strings.TrimSpace(s)
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid value %q (expected e.g. 50 or 50mbps)", s)
	}
	return v, nil
}

// parseWatch parses a duration string. Empty string → 0 (disabled).
func parseWatch(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q (expected e.g. 5m, 30s, 1h)", s)
	}
	if d < 10*time.Second {
		return 0, fmt.Errorf("--watch interval must be at least 10s")
	}
	return d, nil
}
