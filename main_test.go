package main

import (
	"testing"
	"time"

	"github.com/mhdiiilham/speeder/internal/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildConfig_Defaults(t *testing.T) {
	cfg := buildConfig(false, 8, "", false, false)
	assert.Equal(t, 8*time.Second, cfg.MaxDuration)
	assert.Equal(t, 0.03, cfg.CVThreshold)
	assert.Equal(t, "", cfg.ServerHost)
	// progress is enabled by default (noProgress=false, jsonMode=false)
	assert.NotNil(t, cfg.OnProgress)
}

func TestBuildConfig_Quick(t *testing.T) {
	cfg := buildConfig(true, 99, "host.example.com", false, false)
	assert.Equal(t, runner.QuickConfig().MaxDuration, cfg.MaxDuration)
	assert.Equal(t, runner.QuickConfig().CVThreshold, cfg.CVThreshold)
	assert.Equal(t, "host.example.com", cfg.ServerHost)
}

func TestBuildConfig_CustomDuration(t *testing.T) {
	cfg := buildConfig(false, 5, "", false, false)
	assert.Equal(t, 5*time.Second, cfg.MaxDuration)
}

func TestBuildConfig_ProgressDisabledWhenNoProgress(t *testing.T) {
	cfg := buildConfig(false, 8, "", true, false)
	assert.Nil(t, cfg.OnProgress, "OnProgress should be nil when --no-progress is set")
}

func TestBuildConfig_ProgressDisabledInJSONMode(t *testing.T) {
	cfg := buildConfig(false, 8, "", false, true)
	assert.Nil(t, cfg.OnProgress, "OnProgress should be nil in JSON mode")
}

func TestBuildConfig_ProgressEnabledByDefault(t *testing.T) {
	cfg := buildConfig(false, 8, "", false, false)
	assert.NotNil(t, cfg.OnProgress, "OnProgress should be set in normal mode")
}

func TestRun_Version(t *testing.T) {
	code := run([]string{"--version"})
	assert.Equal(t, 0, code)
}

func TestRun_InvalidFlag(t *testing.T) {
	code := run([]string{"--not-a-real-flag"})
	assert.Equal(t, 2, code)
}

func TestResolveGame(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
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
