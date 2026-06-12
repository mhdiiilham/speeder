# speeder

[![CI](https://github.com/mhdiiilham/speeder/actions/workflows/ci.yml/badge.svg)](https://github.com/mhdiiilham/speeder/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mhdiiilham/speeder)](https://goreportcard.com/report/github.com/mhdiiilham/speeder)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/github/license/mhdiiilham/speeder)](LICENSE)
[![Release](https://img.shields.io/github/v/release/mhdiiilham/speeder)](https://github.com/mhdiiilham/speeder/releases)

A fast, data-efficient internet speed test CLI built in Go — with built-in game server latency checking for CS2 and Dota 2.

Uses the [ndt7 protocol](https://github.com/m-lab/ndt-server/blob/master/spec/ndt7-protocol.md) over [M-Lab's](https://www.measurementlab.net/) open, nonprofit server network. No Ookla/speedtest.net dependency.

```
  Server:    ndt-mlab1-sin01.mlab-oti.measurement-lab.org
  Location:  Singapore SG

  Latency:   14.2 ms  jitter: 0.9 ms
  Download:  94.71 Mbps  (3.1s, 36.8 MB)
  Upload:    23.10 Mbps  (2.8s,  8.1 MB)

  Data used: 44.9 MB
```

## speeder vs Ookla Speedtest

Both tools measure internet speed — they just make different trade-offs.

| | speeder | Ookla Speedtest |
|---|---|---|
| **Server network** | ~800 M-Lab servers (nonprofit) | ~15,000 community servers |
| **Open source** | Yes (MIT) | Client is proprietary |
| **Data per test** | ~30–50 MB (adaptive) | ~150–300 MB |
| **Test data** | Published openly by M-Lab | Collected by Ookla |
| **API key** | Not required | Required for CLI/automation |
| **JSON output** | Built-in | Available |
| **Game ping** | Built-in (CS2, Dota 2) | Not available |

**M-Lab** (Measurement Lab) is a nonprofit research project backed by Google, Princeton University, and New America. Its infrastructure is designed specifically for open, reproducible internet measurement — the same backend used by Google's built-in speed test. All test results are published as open data.

**Ookla** has a much larger server network and is the industry-standard reference that most ISPs and users recognise.

Neither is objectively "better" — they measure different things against different server networks, so results will naturally differ. speeder is a good fit if you want a lightweight, scriptable, open-source tool with no account required. Ookla is the right choice when you need a result that's widely recognised or when comparing against an ISP's advertised speeds using their preferred benchmark.

## Features

- **Speed test** — adaptive early stopping (CV < 3%), minimal data usage
- **Game ping** — check CS2 and Dota 2 server latency, jitter, packet loss, and gaming score before queuing
- `--quick` preset: 4 s max, typically < 35 MB total
- JSON output for scripting
- Zero config, no account, no API key
- Pure Go, no CGO — any OS, any architecture

## Installation

### With Go (any OS, any architecture)

```bash
go install github.com/mhdiiilham/speeder@latest
```

The binary lands in `$(go env GOPATH)/bin/`. Make sure that directory is on your `PATH`.

### Pre-built binaries

Download the latest binary for your platform from the [Releases page](https://github.com/mhdiiilham/speeder/releases):

| Platform | Binary |
|---|---|
| Linux amd64 | `speeder-linux-amd64` |
| Linux arm64 | `speeder-linux-arm64` |
| macOS (Apple Silicon) | `speeder-darwin-arm64` |
| macOS (Intel) | `speeder-darwin-amd64` |
| Windows amd64 | `speeder-windows-amd64.exe` |

### macOS / Linux (curl one-liner)

```bash
curl -Lo speeder \
  https://github.com/mhdiiilham/speeder/releases/latest/download/speeder-linux-amd64
chmod +x speeder && sudo mv speeder /usr/local/bin/
```

## Usage

```
speeder [flags]

  --list            list nearby M-Lab servers and exit
  --server string   use a specific server hostname
  --json            output results as JSON to stdout
  --no-progress     disable live progress updates
  --duration int    max seconds per test phase (default 8)
  --quick           quick preset: 4 s max, minimal data usage
  --game string     check game server latency: cs2, dota2
  --version         print version and exit
```

### Speed test

```bash
speeder                   # auto-selects nearest M-Lab server
speeder --quick           # less data, faster result
speeder --json | jq .download_mbps
speeder --no-progress --json >> speedlog.jsonl
```

### Game server check

Know your connection quality before queuing into a ranked match:

```bash
speeder --game cs2
speeder --game dota2
```

```
  CS2 Server Latency

  SERVER                     PING   JITTER   LOSS  SCR  STATUS
  ────────────────────────────────────────────────────────────
  Singapore [AP]            12 ms     1 ms     0%   97  Excellent ✓
  Tokyo [AP]                47 ms     4 ms     0%   72  Good
  Hong Kong [AP]            55 ms     6 ms     0%   64  Playable
  Frankfurt [EU]           198 ms    15 ms     5%    8  Very Poor
  Virginia [NA]            220 ms    20 ms     3%    5  Very Poor
  ────────────────────────────────────────────────────────────

  Verdict: Ready to play! Your connection to Singapore is excellent.

  Note: Steam SDR adds ~2–5 ms overhead between the relay and the actual
  game server, so your real in-game ping will be slightly higher than
  shown here. The relative ranking between servers is accurate.
```

Shows the 5 best servers, sorted by gaming score.

#### Gaming score (0–100)

| Metric | Weight | Best value |
|---|---|---|
| Latency | 50 pts | ≤ 20 ms |
| Jitter | 30 pts | ≤ 2 ms |
| Packet loss | 20 pts | 0% |

| Score | Verdict |
|---|---|
| 85–100 | Ready to play — excellent |
| 70–84 | Good — slight disadvantage in close duels |
| 50–69 | Playable — consider playing off-peak |
| 30–49 | High latency — avoid ranked |
| 0–29 | Very poor — not recommended for competitive |

#### How servers are measured

CS2 and Dota 2 use Valve's CM (Connection Manager) server hostnames
(`cmp1-*.steamserver.net:27018`). These are TCP/WebSocket servers that
accept direct connections and share datacenters with game servers, giving
accurate latency readings. 10 TCP probes per server measure ping, jitter,
and packet loss concurrently.

#### Aliases

```bash
speeder --game cs            # same as --game cs2
speeder --game counter-strike
speeder --game dota          # same as --game dota2
```

### JSON output format (speed test)

```json
{
  "server": { "hostname": "...", "city": "Singapore", "country": "SG" },
  "latency_ms": 14.2,
  "jitter_ms": 0.9,
  "download_mbps": 94.71,
  "upload_mbps": 23.10,
  "data_used_mb": 44.9
}
```

## Data usage (speed test)

speeder stops each phase as soon as the speed measurement stabilises,
minimising data usage without sacrificing accuracy.

| Connection | Default | `--quick` |
|---|---|---|
| 50 Mbps | ~25 MB | ~15 MB |
| 100 Mbps | ~40 MB | ~25 MB |
| 500 Mbps | ~100 MB | ~60 MB |

Compare: Ookla's official client typically uses 150–300 MB on a fast connection.

## Building from source

```bash
git clone https://github.com/mhdiiilham/speeder.git
cd speeder
go mod tidy
make build          # → bin/speeder
make test           # run tests with race detector
make coverage       # test + HTML coverage report
make release        # cross-compile all platforms to bin/
```

### Integration tests (hit real servers)

```bash
go test -tags integration ./internal/game/ -v -timeout 60s
```

Pings real CS2 and Dota 2 servers and prints a full latency table.

Requires Go 1.21+.

## License

MIT — see [LICENSE](LICENSE).
