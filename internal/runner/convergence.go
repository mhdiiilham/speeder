package runner

import (
	"math"
	"time"
)

// coeffVariation returns stddev/mean of s.
// Returns math.MaxFloat64 when s has fewer than 2 elements or a zero mean.
func coeffVariation(s []float64) float64 {
	if len(s) < 2 {
		return math.MaxFloat64
	}
	var sum float64
	for _, v := range s {
		sum += v
	}
	mean := sum / float64(len(s))
	if mean == 0 {
		return math.MaxFloat64
	}
	var variance float64
	for _, v := range s {
		d := v - mean
		variance += d * d
	}
	return math.Sqrt(variance/float64(len(s))) / mean
}

// durationMean returns the arithmetic mean of a slice of durations.
func durationMean(ds []time.Duration) time.Duration {
	if len(ds) == 0 {
		return 0
	}
	var sum time.Duration
	for _, d := range ds {
		sum += d
	}
	return sum / time.Duration(len(ds))
}

// durationJitter returns the population standard deviation of a duration slice.
func durationJitter(ds []time.Duration) time.Duration {
	if len(ds) < 2 {
		return 0
	}
	mean := float64(durationMean(ds))
	var variance float64
	for _, d := range ds {
		diff := float64(d) - mean
		variance += diff * diff
	}
	return time.Duration(math.Sqrt(variance / float64(len(ds))))
}
