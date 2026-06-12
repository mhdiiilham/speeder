// Package history persists speed test results to ~/.speeder/history.jsonl.
package history

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Record is one entry in the history file.
type Record struct {
	Timestamp    time.Time `json:"timestamp"`
	Server       string    `json:"server"`
	Location     string    `json:"location,omitempty"`
	ISP          string    `json:"isp,omitempty"`
	LatencyMs    float64   `json:"latency_ms"`
	JitterMs     float64   `json:"jitter_ms"`
	DownloadMbps float64   `json:"download_mbps"`
	UploadMbps   float64   `json:"upload_mbps"`
	DataUsedMB   float64   `json:"data_used_mb"`
	PingOnly     bool      `json:"ping_only,omitempty"`
}

// Path returns the path to the history file, creating the directory if needed.
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	dir := filepath.Join(home, ".speeder")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create history dir: %w", err)
	}
	return filepath.Join(dir, "history.jsonl"), nil
}

// Save appends a record to the history file.
func Save(r Record) error {
	path, err := Path()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open history file: %w", err)
	}
	defer f.Close()

	line, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("marshal record: %w", err)
	}
	_, err = fmt.Fprintf(f, "%s\n", line)
	return err
}

// Load reads the last n records from the history file.
// If n <= 0, all records are returned.
func Load(n int) ([]Record, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open history file: %w", err)
	}
	defer f.Close()

	var records []Record
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var r Record
		if err := json.Unmarshal(line, &r); err != nil {
			continue // skip malformed lines
		}
		records = append(records, r)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read history: %w", err)
	}

	if n > 0 && len(records) > n {
		records = records[len(records)-n:]
	}
	return records, nil
}
