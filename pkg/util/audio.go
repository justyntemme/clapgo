// Package util provides common utility functions for the ClapGo framework.
package util

import (
	"math"
)

// LinearToDb converts a linear gain value to decibels
func LinearToDb(linear float64) float64 {
	if linear <= 0.0 {
		return -math.MaxFloat64
	}
	return 20.0 * math.Log10(linear)
}

// DbToLinear converts decibels to a linear gain value
func DbToLinear(db float64) float64 {
	return math.Pow(10.0, db/20.0)
}

// NoteToFrequency converts a MIDI note number to frequency
func NoteToFrequency(note int) float64 {
	return 440.0 * math.Pow(2.0, (float64(note)-69.0)/12.0)
}

// FrequencyToNote converts a frequency to the nearest MIDI note number
func FrequencyToNote(freq float64) int {
	if freq <= 0.0 {
		return 0
	}
	return int(math.Round(12.0*math.Log2(freq/440.0) + 69.0))
}

// ClampValue clamps a value between min and max
func ClampValue(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// InterpolateLinear performs linear interpolation between a and b
func InterpolateLinear(a, b, t float64) float64 {
	return a + (b-a)*t
}

// InterpolateExponential performs exponential interpolation between a and b
func InterpolateExponential(a, b, t float64) float64 {
	if a <= 0.0 || b <= 0.0 {
		return InterpolateLinear(a, b, t)
	}
	return a * math.Pow(b/a, t)
}

// Smoothstep implements the smoothstep function
func Smoothstep(edge0, edge1, x float64) float64 {
	// Scale, bias and saturate x to 0..1 range
	x = ClampValue((x-edge0)/(edge1-edge0), 0.0, 1.0)
	// Evaluate polynomial
	return x * x * (3 - 2*x)
}

// MidiVelocityToFloat converts a MIDI velocity (0-127) to a float (0.0-1.0)
func MidiVelocityToFloat(velocity int) float64 {
	return float64(velocity) / 127.0
}

// FloatToMidiVelocity converts a float (0.0-1.0) to a MIDI velocity (0-127)
func FloatToMidiVelocity(velocity float64) int {
	return int(math.Round(ClampValue(velocity, 0.0, 1.0) * 127.0))
}