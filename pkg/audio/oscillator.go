package audio

import (
	"math"
)

// WaveformType represents different waveform types for oscillators
type WaveformType int

const (
	WaveformSine WaveformType = iota
	WaveformSaw
	WaveformSquare
	WaveformTriangle
	WaveformNoise
)

// GenerateWaveformSample generates a single sample based on the waveform type and phase.
// Phase should be in the range [0, 1).
func GenerateWaveformSample(phase float64, waveform WaveformType) float64 {
	switch waveform {
	case WaveformSine:
		return math.Sin(2.0 * math.Pi * phase)
		
	case WaveformSaw:
		return 2.0*phase - 1.0
		
	case WaveformSquare:
		if phase < 0.5 {
			return 1.0
		}
		return -1.0
		
	case WaveformTriangle:
		if phase < 0.5 {
			return 4.0*phase - 1.0
		}
		return -4.0*phase + 3.0
		
	case WaveformNoise:
		// For noise, we use the phase as a seed for pseudo-randomness
		// This is deterministic but sounds random enough for audio
		x := math.Sin(phase * 12.9898 + 78.233) * 43758.5453
		return 2.0*(x-math.Floor(x)) - 1.0
		
	default:
		return 0.0
	}
}

// AdvancePhase advances an oscillator phase by the given frequency and sample rate.
// Returns the new phase wrapped to [0, 1).
func AdvancePhase(currentPhase, frequency, sampleRate float64) float64 {
	phase := currentPhase + frequency/sampleRate
	if phase >= 1.0 {
		phase -= math.Floor(phase)
	}
	return phase
}

// GeneratePolyBLEPSaw generates an anti-aliased sawtooth wave using PolyBLEP.
// This reduces aliasing compared to a naive sawtooth.
func GeneratePolyBLEPSaw(phase, phaseIncrement float64) float64 {
	value := 2.0*phase - 1.0
	
	// Apply PolyBLEP to reduce aliasing at the discontinuity
	if phase < phaseIncrement {
		t := phase / phaseIncrement
		value -= 2.0 * t * t * (1.0 - 0.5*t)
	} else if phase > 1.0-phaseIncrement {
		t := (phase - 1.0) / phaseIncrement
		value -= 2.0 * t * t * (1.0 + 0.5*t)
	}
	
	return value
}

// GeneratePolyBLEPSquare generates an anti-aliased square wave using PolyBLEP.
func GeneratePolyBLEPSquare(phase, phaseIncrement float64) float64 {
	value := 1.0
	if phase >= 0.5 {
		value = -1.0
	}
	
	// Apply PolyBLEP at both discontinuities
	if phase < phaseIncrement {
		t := phase / phaseIncrement
		value += 2.0 * t * t * (1.0 - 0.5*t)
	} else if phase > 1.0-phaseIncrement {
		t := (phase - 1.0) / phaseIncrement
		value += 2.0 * t * t * (1.0 + 0.5*t)
	}
	
	// Check for discontinuity at 0.5
	if phase > 0.5-phaseIncrement && phase < 0.5+phaseIncrement {
		t := (phase - 0.5) / phaseIncrement
		if t < 0 {
			value -= 2.0 * t * t * (1.0 + 0.5*t)
		} else {
			value -= 2.0 * t * t * (1.0 - 0.5*t)
		}
	}
	
	return value
}

// NoteToFrequency converts a MIDI note number to frequency in Hz
func NoteToFrequency(note int) float64 {
	return 440.0 * math.Pow(2.0, (float64(note)-69.0)/12.0)
}