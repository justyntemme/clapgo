// Package util provides common utility functions for the ClapGo framework.
package util

import (
	"math"
)

// EnvelopeStage represents the current stage of an ADSR envelope
type EnvelopeStage int

const (
	EnvelopeStageIdle EnvelopeStage = iota
	EnvelopeStageAttack
	EnvelopeStageDecay
	EnvelopeStageSustain
	EnvelopeStageRelease
)

// ADSREnvelope represents an ADSR (Attack, Decay, Sustain, Release) envelope generator
type ADSREnvelope struct {
	// Parameters (in seconds)
	Attack  float64
	Decay   float64
	Sustain float64 // Level (0-1), not time
	Release float64
	
	// State
	Stage        EnvelopeStage
	CurrentValue float64
	TimeInStage  float64
	ReleaseLevel float64 // Level when release was triggered
	
	// Configuration
	SampleRate float64
}

// NewADSREnvelope creates a new ADSR envelope with default values
func NewADSREnvelope(sampleRate float64) *ADSREnvelope {
	return &ADSREnvelope{
		Attack:     0.01,  // 10ms
		Decay:      0.1,   // 100ms
		Sustain:    0.7,   // 70%
		Release:    0.3,   // 300ms
		SampleRate: sampleRate,
		Stage:      EnvelopeStageIdle,
	}
}

// Trigger starts the envelope from the attack stage
func (env *ADSREnvelope) Trigger() {
	env.Stage = EnvelopeStageAttack
	env.TimeInStage = 0
	env.CurrentValue = 0
}

// Release moves the envelope to the release stage
func (env *ADSREnvelope) Release() {
	if env.Stage != EnvelopeStageIdle && env.Stage != EnvelopeStageRelease {
		env.ReleaseLevel = env.CurrentValue
		env.Stage = EnvelopeStageRelease
		env.TimeInStage = 0
	}
}

// Process advances the envelope by one sample and returns the current value
func (env *ADSREnvelope) Process() float64 {
	sampleDuration := 1.0 / env.SampleRate
	
	switch env.Stage {
	case EnvelopeStageIdle:
		env.CurrentValue = 0
		
	case EnvelopeStageAttack:
		if env.Attack > 0 {
			env.CurrentValue = env.TimeInStage / env.Attack
			if env.CurrentValue >= 1.0 {
				env.CurrentValue = 1.0
				env.Stage = EnvelopeStageDecay
				env.TimeInStage = 0
			} else {
				env.TimeInStage += sampleDuration
			}
		} else {
			env.CurrentValue = 1.0
			env.Stage = EnvelopeStageDecay
			env.TimeInStage = 0
		}
		
	case EnvelopeStageDecay:
		if env.Decay > 0 {
			decayProgress := env.TimeInStage / env.Decay
			env.CurrentValue = 1.0 - decayProgress*(1.0-env.Sustain)
			if decayProgress >= 1.0 {
				env.CurrentValue = env.Sustain
				env.Stage = EnvelopeStageSustain
				env.TimeInStage = 0
			} else {
				env.TimeInStage += sampleDuration
			}
		} else {
			env.CurrentValue = env.Sustain
			env.Stage = EnvelopeStageSustain
			env.TimeInStage = 0
		}
		
	case EnvelopeStageSustain:
		env.CurrentValue = env.Sustain
		
	case EnvelopeStageRelease:
		if env.Release > 0 {
			releaseProgress := env.TimeInStage / env.Release
			if releaseProgress >= 1.0 {
				env.CurrentValue = 0
				env.Stage = EnvelopeStageIdle
				env.TimeInStage = 0
			} else {
				// Exponential release curve
				env.CurrentValue = env.ReleaseLevel * math.Pow(1.0-releaseProgress, 2.0)
				env.TimeInStage += sampleDuration
			}
		} else {
			env.CurrentValue = 0
			env.Stage = EnvelopeStageIdle
			env.TimeInStage = 0
		}
	}
	
	return env.CurrentValue
}

// IsActive returns true if the envelope is currently generating a non-zero value
func (env *ADSREnvelope) IsActive() bool {
	return env.Stage != EnvelopeStageIdle
}

// Reset immediately resets the envelope to idle state
func (env *ADSREnvelope) Reset() {
	env.Stage = EnvelopeStageIdle
	env.CurrentValue = 0
	env.TimeInStage = 0
}

// SetADSR sets all ADSR parameters at once
func (env *ADSREnvelope) SetADSR(attack, decay, sustain, release float64) {
	env.Attack = ClampValue(attack, 0, 10.0)    // Max 10 seconds
	env.Decay = ClampValue(decay, 0, 10.0)      // Max 10 seconds
	env.Sustain = ClampValue(sustain, 0, 1.0)   // 0-100%
	env.Release = ClampValue(release, 0, 10.0)  // Max 10 seconds
}

// SimpleADSR is a stateless version that calculates envelope value based on elapsed time
// This is useful for simple cases where you don't need full envelope state management
func SimpleADSR(elapsedSamples uint32, sampleRate, attack, decay, sustain, release float64, isReleased bool, releaseStartSample uint32) float64 {
	elapsedTime := float64(elapsedSamples) / sampleRate
	
	if !isReleased {
		// Attack phase
		if elapsedTime < attack {
			if attack > 0 {
				return elapsedTime / attack
			}
			return 1.0
		}
		
		// Decay phase
		elapsedTime -= attack
		if elapsedTime < decay {
			if decay > 0 {
				decayProgress := elapsedTime / decay
				return 1.0 - decayProgress*(1.0-sustain)
			}
			return sustain
		}
		
		// Sustain phase
		return sustain
	}
	
	// Release phase
	releaseTime := float64(elapsedSamples-releaseStartSample) / sampleRate
	if release > 0 && releaseTime < release {
		releaseProgress := releaseTime / release
		return sustain * math.Pow(1.0-releaseProgress, 2.0)
	}
	
	return 0.0
}