package main

import (
	"context"
	"fmt"
	"time"

	"github.com/mhdiiilham/speeder/internal/game"
	"github.com/mhdiiilham/speeder/internal/history"
	"github.com/mhdiiilham/speeder/internal/ipinfo"
	"github.com/mhdiiilham/speeder/internal/runner"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Shutdown(ctx context.Context) {
	if a.cancel != nil {
		a.cancel()
	}
}

type SpeedTestConfig struct {
	Quick    bool   `json:"quick"`
	Duration int    `json:"duration"`
	Server   string `json:"server"`
	PingOnly bool   `json:"pingOnly"`
}

func (a *App) RunSpeedTest(config SpeedTestConfig) error {
	var cfg runner.Config
	if config.Quick {
		cfg = runner.QuickConfig()
	} else {
		cfg = runner.DefaultConfig()
		if config.Duration > 0 && config.Duration != 8 {
			cfg.MaxDuration = time.Duration(config.Duration) * time.Second
		}
	}
	cfg.ServerHost = config.Server
	cfg.PingOnly = config.PingOnly

	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel

	go func() {
		defer cancel()

		cfg.OnProgress = func(phase string, mbps float64) {
			runtime.EventsEmit(a.ctx, "speedtest:progress", map[string]interface{}{
				"phase": phase,
				"mbps":  mbps,
			})
		}

		ispCh := make(chan struct{ isp, ip string }, 1)
		go func() {
			info, err := ipinfo.DefaultClient.Fetch(ctx)
			if err == nil && info != nil {
				ispCh <- struct{ isp, ip string }{info.ISP, info.IP}
			} else {
				ispCh <- struct{ isp, ip string }{}
			}
		}()

		runtime.EventsEmit(a.ctx, "speedtest:status", "Finding nearest M-Lab server...")

		result, err := runner.Default().Run(ctx, cfg)
		if err != nil {
			runtime.EventsEmit(a.ctx, "speedtest:error", err.Error())
			return
		}

		isp := <-ispCh
		result.ISP = isp.isp
		result.ClientIP = isp.ip

		saveHistory(result)

		runtime.EventsEmit(a.ctx, "speedtest:complete", result)
	}()

	return nil
}

func (a *App) CancelSpeedTest() {
	if a.cancel != nil {
		a.cancel()
	}
}

type GameCheckResult struct {
	GameName string           `json:"gameName"`
	Note     string           `json:"note"`
	Results  []game.PingResult `json:"results"`
	Verdict  string           `json:"verdict"`
}

func (a *App) CheckGameServers(gameName string) (*GameCheckResult, error) {
	var g game.Game
	switch gameName {
	case "cs2", "cs", "counter-strike":
		g = game.NewCS2()
	case "dota2", "dota":
		g = game.NewDota2()
	default:
		return nil, fmt.Errorf("unknown game %q", gameName)
	}

	p := &game.TCPGamePinger{}
	results, err := game.Check(context.Background(), g, p, 10)
	if err != nil {
		return nil, err
	}

	verdict := ""
	if len(results) > 0 {
		verdict = game.Verdict(g.Name(), results[0])
	}

	return &GameCheckResult{
		GameName: g.Name(),
		Note:     g.Note(),
		Results:  results,
		Verdict:  verdict,
	}, nil
}

func (a *App) LaunchGame(gameId string) error {
	var url string
	switch gameId {
	case "cs2", "cs":
		url = "steam://rungameid/730"
	case "dota2", "dota":
		url = "steam://rungameid/570"
	default:
		return fmt.Errorf("unknown game %q", gameId)
	}
	runtime.BrowserOpenURL(a.ctx, url)
	return nil
}

func (a *App) GetHistory(count int) ([]history.Record, error) {
	records, err := history.Load(count)
	if err != nil {
		return nil, err
	}
	if records == nil {
		return []history.Record{}, nil
	}
	return records, nil
}

func (a *App) ListServers() ([]runner.Server, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return runner.Default().ListServers(ctx)
}

func (a *App) GetVersion() string {
	return "dev"
}

func saveHistory(r *runner.Result) {
	rec := history.Record{
		Timestamp:    time.Now().UTC(),
		Server:       r.Server.Hostname,
		Location:     r.Server.City + " " + r.Server.Country,
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
