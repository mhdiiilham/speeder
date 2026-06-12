# speeder

[![CI](https://github.com/mhdiiilham/speeder/actions/workflows/ci.yml/badge.svg)](https://github.com/mhdiiilham/speeder/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mhdiiilham/speeder)](https://goreportcard.com/report/github.com/mhdiiilham/speeder)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/github/license/mhdiiilham/speeder)](LICENSE)
[![Release](https://img.shields.io/github/v/release/mhdiiilham/speeder)](https://github.com/mhdiiilham/speeder/releases)

A fast, data-efficient internet speed test CLI built in Go. Uses the [ndt7 protocol](https://github.com/m-lab/ndt-server/blob/master/spec/ndt7-protocol.md) over [M-Lab's](https://www.measurementlab.net/) open, nonprofit server network — no Ookla/speedtest.net dependency.

```
  Server:    ndt-mlab1-sin01.mlab-oti.measurement-lab.org
  Location:  Singapore SG

  Latency:   14.2 ms  jitter: 0.9 ms
  Download:  94.71 Mbps  (3.1s, 36.8 MB)
  Upload:    23.10 Mbps  (2.8s,  8.1 MB)

  Data used: 44.9 MB
```

## Features

- Adaptive early stopping — tests stop as soon as the speed measurement converges (coefficient of variation < 3%), minimising data usage
- `--quick` preset uses 4 s max per phase, typically < 35 MB total
- JSON output for scripting
- Zero config, no account, no API key — M-Lab servers are publicly available worldwide
- Pure Go, no CGO — runs on any platform Go supports

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
# Replace OS and ARCH as needed
curl -Lo speeder \
  https://github.com/mhdiiilham/speeder/releases/latest/download/speeder-linux-amd64
chmod +x speeder
sudo mv speeder /usr/local/bin/
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
  --version         print version and exit
```

### Examples

```bash
# Default run — auto-selects nearest server
speeder

# Minimal data usage
speeder --quick

# JSON output (pipe-friendly; progress goes to stderr)
speeder --json | jq .download_mbps

# List nearby servers, then test against a specific one
speeder --list
speeder --server ndt-mlab1-sin01.mlab-oti.measurement-lab.org

# No live output (for cron / logging)
speeder --no-progress --json >> speedlog.jsonl
```

### JSON output format

```json
{
  "server": {
    "hostname": "ndt-mlab1-sin01.mlab-oti.measurement-lab.org",
    "city": "Singapore",
    "country": "SG"
  },
  "latency_ms": 14.2,
  "jitter_ms": 0.9,
  "download_mbps": 94.71,
  "upload_mbps": 23.10,
  "data_used_mb": 44.9
}
```

## Data usage

speeder uses an adaptive algorithm: each test phase (download, upload) runs for a minimum of 2 seconds to pass TCP slow-start, then stops as soon as the measured speed stabilises (coefficient of variation drops below 3%). The hard ceiling is `--duration` seconds (default 8).

Typical data usage:

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

Requires Go 1.21+.

## License

MIT — see [LICENSE](LICENSE).
