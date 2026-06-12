package main

import (
	"testing"
	"time"

	"github.com/mhdiiilham/speeder/internal/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildConfig_Defaults(t *testing.T) {
	cfg := buildConfig(false, 8, "", false, false, false)
	assert.Equal(t, 8*time.Second, cfg.MaxDuration)
	assert.Equal(t, 0.03, cfg.CVThreshold)
	assert.Equal(t, "", cfg.ServerHost)
	assert.False(t, cfg.PingOnly)
	assert.NotNil(t, cfg.OnProgress)
}

func TestBuildConfig_Quick(t *testing.T) {
	cfg := buildConfig(true, 99, "host.example.com", false, false, false)
	assert.Equal(t, runner.QuickConfig().MaxDuration, cfg.MaxDuration)
	assert.Equal(t, runner.QuickConfig().CVThreshold, cfg.CVThreshold)
	assert.Equal(t, "host.example.com", cfg.ServerHost)
}

func TestBuildConfig_CustomDuration(t *testing.T) {
	cfg := buildConfig(false, 5, "", false, false, false)
	assert.Equal(t, 5*time.Second, cfg.MaxDuration)
}

func TestBuildConfig_PingOnly(t *testing.T) {
	cfg := buildConfig(false, 8, "", false, false, true)
	assert.True(t, cfg.PingOnly)
}

func TestBuildConfig_ProgressDisabledWhenNoProgress(t *testing.T) {
	cfg := buildConfig(false, 8, "", true, false, false)
	assert.Nil(t, cfg.OnProgress)
}

func TestBuildConfig_ProgressDisabledInJSONMode(t *testing.T) {
	cfg := buildConfig(false, 8, "", false, true, false)
	assert.Nil(t, cfg.OnProgress)
}

func TestBuildConfig_ProgressEnabledByDefault(t *testing.T) {
	cfg := buildConfig(false, 8, "", false, false, false)
	assert.NotNil(t, cfg.OnProgress)
}

func TestParseThreshold(t *testing.T) {
	tests := []struct {
		input   string
		want    float64
		wantErr bool
	}{
		{"", 0, false},
		{"50", 50, false},
		{"50mbps", 50, false},
		{"50Mbps", 50, false},
		{"100mb/s", 100, false},
		{"abc", 0, true},
	}
	for _, tc := range tests {
		got, err := parseThreshold(tc.input)
		if tc.wantErr {
			assert.Error(t, err, "input=%q", tc.input)
		} else {
			assert.NoError(t, err, "input=%q", tc.input)
			assert.Equal(t, tc.want, got, "input=%q", tc.input)
		}
	}
}

func TestParseWatch(t *testing.T) {
	d, err := parseWatch("")
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), d)

	d, err = parseWatch("5m")
	assert.NoError(t, err)
	assert.Equal(t, 5*time.Minute, d)

	_, err = parseWatch("5s") // below 10s minimum
	assert.Error(t, err)

	_, err = parseWatch("not-a-duration")
	assert.Error(t, err)
}

func TestRun_Version(t *testing.T) {
	assert.Equal(t, 0, run([]string{"--version"}))
}

func TestRun_InvalidFlag(t *testing.T) {
	assert.Equal(t, 2, run([]string{"--not-a-real-flag"}))
}

func TestRun_InvalidFailBelow(t *testing.T) {
	assert.Equal(t, 2, run([]string{"--fail-if-below", "abc"}))
}

func TestRun_InvalidWatch(t *testing.T) {
	assert.Equal(t, 2, run([]string{"--watch", "bad"}))
}

func TestResolveGame(t *testing.T) {
	tests := []struct {
		input    string
		wantErr  bool
		wantName string
	}{
		{"cs2", false, "CS2"},
		{"CS2", false, "CS2"},
		{"cs", false, "CS2"},
		{"counter-strike", false, "CS2"},
		{"dota2", false, "Dota 2"},
		{"dota", false, "Dota 2"},
		{"valorant", true, ""},
		{"minecraft", true, ""},
		{"", true, ""},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			g, err := resolveGame(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantName, g.Name())
		})
	}
}
