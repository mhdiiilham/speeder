package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mhdiiilham/speeder/internal/display"
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

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *flagVersion {
		fmt.Fprintf(os.Stdout, "speeder %s\n", version)
		return 0
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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
