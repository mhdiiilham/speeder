package runner

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCoeffVariation(t *testing.T) {
	tests := []struct {
		name    string
		samples []float64
		want    float64 // expected CV; use math.MaxFloat64 sentinel for "max"
	}{
		{"empty", []float64{}, math.MaxFloat64},
		{"single", []float64{100}, math.MaxFloat64},
		{"zero mean", []float64{0, 0, 0}, math.MaxFloat64},
		{"two equal", []float64{100, 100}, 0},
		{"perfect stable", []float64{50, 50, 50, 50}, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := coeffVariation(tc.samples)
			if tc.want == math.MaxFloat64 {
				assert.Equal(t, math.MaxFloat64, got)
			} else {
				assert.InDelta(t, tc.want, got, 1e-9)
			}
		})
	}
}

func TestCoeffVariation_StableSignal(t *testing.T) {
	stable := []float64{99.1, 100.2, 99.8, 100.5, 99.6, 100.1, 99.9, 100.3}
	cv := coeffVariation(stable)
	assert.Less(t, cv, 0.03, "stable signal should have CV < 3%%")
}

func TestCoeffVariation_NoisySignal(t *testing.T) {
	noisy := []float64{10, 200, 5, 180, 50, 140}
	cv := coeffVariation(noisy)
	assert.Greater(t, cv, 0.3, "noisy signal should have CV > 30%%")
}

func TestDurationMean(t *testing.T) {
	assert.Equal(t, time.Duration(0), durationMean(nil))
	assert.Equal(t, 10*time.Millisecond, durationMean([]time.Duration{10 * time.Millisecond}))
	assert.Equal(t, 20*time.Millisecond, durationMean([]time.Duration{10 * time.Millisecond, 30 * time.Millisecond}))
}

func TestDurationJitter(t *testing.T) {
	assert.Equal(t, time.Duration(0), durationJitter(nil))
	assert.Equal(t, time.Duration(0), durationJitter([]time.Duration{50 * time.Millisecond}))
	// constant sequence → zero jitter
	constant := []time.Duration{20 * time.Millisecond, 20 * time.Millisecond, 20 * time.Millisecond}
	assert.Equal(t, time.Duration(0), durationJitter(constant))
	// varying sequence → non-zero jitter
	varying := []time.Duration{10 * time.Millisecond, 30 * time.Millisecond}
	assert.Greater(t, durationJitter(varying), time.Duration(0))
}
