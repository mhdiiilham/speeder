package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mhdiiilham/speeder/internal/display"
	"github.com/mhdiiilham/speeder/internal/game"
	"github.com/mhdiiilham/speeder/internal/runner"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("speeder", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	flagList       := fs.Bool("list", false, "list nearby M-Lab servers and exit")
	flagServer     := fs.String("server", "", "use a specific server hostname")
	flagJSON        := fs.Bool("json", false, "output results as JSON")
	flagNoProgress := fs.Bool("no-progress", false, "disable live progress updates")
	flagDuration   := fs.Int("duration", 8, "max seconds per test phase")
	flagQuick      := fs.Bool("quick", false, "quick preset: 4s max, minimal data usage")
	flagVersion    := fs.Bool("version", false, "print version and exit")
	flagGame       := fs.String("game", "", "check game server latency: cs2, dota2")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *flagVersion {
		fmt.Fprintf(os.Stdout, "speeder %s\n", version)
		return 0
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Game ping mode.
	if *flagGame != "" {
		return runGameCheck(ctx, *flagGame)
	}

	r := runner.Default()

	if *flagList {
		servers, err := r.ListServers(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
		display.PrintServerList(os.Stdout, servers)
		return 0
	}

	cfg := buildConfig(*flagQuick, *flagDuration, *flagServer, *flagNoProgress, *flagJSON)

	fmt.Fprintf(os.Stderr, "  Finding nearest M-Lab server...\n")

	result, err := r.Run(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	if *flagJSON {
		if err := display.PrintJSON(os.Stdout, result); err != nil {
			fmt.Fprintf(os.Stderr, "error encoding JSON: %v\n", err)
			return 1
		}
		return 0
	}

	display.PrintResult(os.Stdout, result)
	return 0
}

// runGameCheck resolves the game name, pings all servers, and prints results.
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

// resolveGame maps a CLI name to a Game implementation.
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

// buildConfig constructs a runner.Config from parsed CLI values.
func buildConfig(quick bool, duration int, server string, noProgress bool, jsonMode bool) runner.Config {
	var cfg runner.Config
	if quick {
		cfg = runner.QuickConfig()
	} else {
		cfg = runner.DefaultConfig()
		cfg.MaxDuration = time.Duration(duration) * time.Second
	}
	cfg.ServerHost = server

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
