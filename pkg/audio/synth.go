package audio

import (
	"math"
)

// PolyphonicOscillator manages multiple oscillator voices for synthesizers
type PolyphonicOscillator struct {
	voiceManager *VoiceManager
	waveformType WaveformType
	antiAlias    bool
}

// NewPolyphonicOscillator creates a new polyphonic oscillator
func NewPolyphonicOscillator(voiceManager *VoiceManager) *PolyphonicOscillator {
	return &PolyphonicOscillator{
		voiceManager: voiceManager,
		waveformType: WaveformSine,
		antiAlias:    true,
	}
}

// SetWaveform sets the waveform type for all voices
func (po *PolyphonicOscillator) SetWaveform(waveform WaveformType) {
	po.waveformType = waveform
}

// SetAntiAliasing enables or disables anti-aliasing
func (po *PolyphonicOscillator) SetAntiAliasing(enabled bool) {
	po.antiAlias = enabled
}

// Process generates audio for all active voices
func (po *PolyphonicOscillator) Process(frameCount uint32) []float32 {
	return po.voiceManager.ProcessVoices(frameCount, func(voice *Voice, frames uint32) []float32 {
		output := make([]float32, frames)
		
		for i := uint32(0); i < frames; i++ {
			// Get envelope value
			envValue := 1.0
			if voice.Envelope != nil {
				envValue = voice.Envelope.Process()
			}
			
			// Calculate frequency with pitch bend
			freq := voice.Frequency * math.Pow(2.0, voice.PitchBend/12.0)
			
			// Generate oscillator sample
			var sample float64
			if po.antiAlias && (po.waveformType == WaveformSaw || po.waveformType == WaveformSquare) {
				// Use anti-aliased versions for saw and square
				phaseInc := freq / po.voiceManager.sampleRate
				if po.waveformType == WaveformSaw {
					sample = GeneratePolyBLEPSaw(voice.Phase, phaseInc)
				} else {
					sample = GeneratePolyBLEPSquare(voice.Phase, phaseInc)
				}
			} else {
				// Use standard waveform generation
				sample = GenerateWaveformSample(voice.Phase, po.waveformType)
			}
			
			// Apply envelope, velocity, and volume
			sample *= envValue * voice.Velocity * voice.Volume
			
			// Apply brightness as simple lowpass (this is a placeholder)
			// In a real implementation, you'd use a proper filter
			if voice.Brightness < 1.0 {
				sample *= (voice.Brightness * 0.7 + 0.3)
			}
			
			// Apply pressure (aftertouch) as additional volume
			if voice.Pressure > 0.0 {
				sample *= (1.0 + voice.Pressure * 0.3)
			}
			
			output[i] = float32(sample)
			
			// Advance phase
			voice.Phase = AdvancePhase(voice.Phase, freq, po.voiceManager.sampleRate)
		}
		
		return output
	})
}

// SimpleLowPassFilter implements a basic one-pole lowpass filter
type SimpleLowPassFilter struct {
	cutoff     float64
	sampleRate float64
	a0, b1     float64 // Filter coefficients
	state      float64 // Filter state
}

// NewSimpleLowPassFilter creates a new lowpass filter
func NewSimpleLowPassFilter(sampleRate float64) *SimpleLowPassFilter {
	f := &SimpleLowPassFilter{
		sampleRate: sampleRate,
		cutoff:     20000.0, // Default to wide open
	}
	f.updateCoefficients()
	return f
}

// SetCutoff sets the filter cutoff frequency in Hz
func (f *SimpleLowPassFilter) SetCutoff(cutoff float64) {
	f.cutoff = Clamp(cutoff, 20.0, f.sampleRate*0.49)
	f.updateCoefficients()
}

// updateCoefficients recalculates filter coefficients
func (f *SimpleLowPassFilter) updateCoefficients() {
	// Simple one-pole lowpass
	omega := 2.0 * math.Pi * f.cutoff / f.sampleRate
	f.a0 = omega / (omega + 1.0)
	f.b1 = (omega - 1.0) / (omega + 1.0)
}

// Process applies the filter to a single sample
func (f *SimpleLowPassFilter) Process(input float64) float64 {
	output := f.a0*input - f.b1*f.state
	f.state = output
	return output
}

// ProcessBuffer applies the filter to a buffer of samples
func (f *SimpleLowPassFilter) ProcessBuffer(buffer []float32) {
	for i := range buffer {
		buffer[i] = float32(f.Process(float64(buffer[i])))
	}
}

// Reset clears the filter state
func (f *SimpleLowPassFilter) Reset() {
	f.state = 0
}

// StateVariableFilter implements a more versatile filter with multiple outputs
type StateVariableFilter struct {
	sampleRate float64
	frequency  float64
	resonance  float64
	
	// State variables
	lowpass  float64
	highpass float64
	bandpass float64
	notch    float64
	
	// Previous sample
	prevBandpass float64
	prevLowpass  float64
}

// NewStateVariableFilter creates a new state variable filter
func NewStateVariableFilter(sampleRate float64) *StateVariableFilter {
	return &StateVariableFilter{
		sampleRate: sampleRate,
		frequency:  1000.0,
		resonance:  1.0,
	}
}

// SetFrequency sets the filter frequency in Hz
func (f *StateVariableFilter) SetFrequency(freq float64) {
	// Limit to 0.45 of Nyquist for stability (instead of 0.49)
	f.frequency = Clamp(freq, 20.0, f.sampleRate*0.45)
}

// SetResonance sets the filter resonance (Q factor)
func (f *StateVariableFilter) SetResonance(q float64) {
	f.resonance = Clamp(q, 0.5, 20.0)
}

// Process runs one sample through the filter and returns all outputs
func (f *StateVariableFilter) Process(input float64) (lowpass, highpass, bandpass, notch float64) {
	// Calculate frequency coefficient using the correct formula for SVF
	// This is the frequency in normalized form (0 to 1)
	w := f.frequency / f.sampleRate
	
	// For the SVF, we use this simpler, more stable coefficient
	// This gives us the correct frequency response
	freq := 2.0 * math.Sin(math.Pi * w)
	
	// Hard limit for stability - this prevents filter explosion
	if freq > 1.5 {
		freq = 1.5
	}
	
	// Calculate damping (this controls resonance)
	damp := 2.0 / f.resonance
	
	// Process - this is the standard SVF algorithm
	// But with some modifications for stability
	f.highpass = input - f.prevLowpass - damp*f.prevBandpass
	f.bandpass = freq*f.highpass + f.prevBandpass
	f.lowpass = freq*f.bandpass + f.prevLowpass
	f.notch = f.highpass + f.lowpass
	
	// Apply soft clipping to prevent explosion
	if math.Abs(f.lowpass) > 10.0 {
		f.lowpass = 10.0 * math.Tanh(f.lowpass/10.0)
	}
	if math.Abs(f.bandpass) > 10.0 {
		f.bandpass = 10.0 * math.Tanh(f.bandpass/10.0)
	}
	
	// Store state
	f.prevBandpass = f.bandpass
	f.prevLowpass = f.lowpass
	
	return f.lowpass, f.highpass, f.bandpass, f.notch
}

// ProcessLowpass processes a sample and returns only the lowpass output
func (f *StateVariableFilter) ProcessLowpass(input float64) float64 {
	lp, _, _, _ := f.Process(input)
	return lp
}

// ProcessHighpass processes a sample and returns only the highpass output
func (f *StateVariableFilter) ProcessHighpass(input float64) float64 {
	_, hp, _, _ := f.Process(input)
	return hp
}

// ProcessBandpass processes a sample and returns only the bandpass output
func (f *StateVariableFilter) ProcessBandpass(input float64) float64 {
	_, _, bp, _ := f.Process(input)
	return bp
}

// Reset clears the filter state
func (f *StateVariableFilter) Reset() {
	f.lowpass = 0
	f.highpass = 0
	f.bandpass = 0
	f.notch = 0
	f.prevBandpass = 0
	f.prevLowpass = 0
}

// PitchBendProcessor handles pitch bend with configurable range
type PitchBendProcessor struct {
	bendRange float64 // in semitones
}

// NewPitchBendProcessor creates a new pitch bend processor
func NewPitchBendProcessor(bendRangeSemitones float64) *PitchBendProcessor {
	return &PitchBendProcessor{
		bendRange: bendRangeSemitones,
	}
}

// ApplyPitchBend applies pitch bend to a frequency
// bendValue should be in range -1.0 to 1.0
func (p *PitchBendProcessor) ApplyPitchBend(baseFreq, bendValue float64) float64 {
	// Convert bend value to semitones
	semitones := bendValue * p.bendRange
	// Apply pitch bend using equal temperament formula
	return baseFreq * math.Pow(2.0, semitones/12.0)
}

// SetBendRange sets the pitch bend range in semitones
func (p *PitchBendProcessor) SetBendRange(semitones float64) {
	p.bendRange = semitones
}

// SynthVoiceProcessor is a complete voice processor for a synthesizer
type SynthVoiceProcessor struct {
	oscillator *PolyphonicOscillator
	filter     *StateVariableFilter
	
	// Filter modulation
	filterCutoff    float64
	filterResonance float64
	filterEnvAmount float64
	
	// Optional filter envelope
	filterEnvelope *ADSREnvelope
}

// NewSynthVoiceProcessor creates a new synth voice processor
func NewSynthVoiceProcessor(voiceManager *VoiceManager, sampleRate float64) *SynthVoiceProcessor {
	return &SynthVoiceProcessor{
		oscillator:      NewPolyphonicOscillator(voiceManager),
		filter:          NewStateVariableFilter(sampleRate),
		filterCutoff:    5000.0,
		filterResonance: 1.0,
		filterEnvAmount: 0.0,
		filterEnvelope:  NewADSREnvelope(sampleRate),
	}
}

// Process generates and filters audio for all voices
func (svp *SynthVoiceProcessor) Process(frameCount uint32) []float32 {
	// Generate oscillator output
	output := svp.oscillator.Process(frameCount)
	
	// Apply filter to the mixed output
	for i := range output {
		// Calculate filter cutoff with envelope modulation if enabled
		cutoff := svp.filterCutoff
		if svp.filterEnvAmount > 0 && svp.filterEnvelope != nil {
			envValue := svp.filterEnvelope.Process()
			// Modulate cutoff frequency exponentially
			cutoff *= math.Pow(2.0, svp.filterEnvAmount*envValue*4.0)
		}
		
		svp.filter.SetFrequency(cutoff)
		svp.filter.SetResonance(svp.filterResonance)
		
		// Process through filter
		output[i] = float32(svp.filter.ProcessLowpass(float64(output[i])))
	}
	
	return output
}

// SetFilterParameters sets the filter parameters
func (svp *SynthVoiceProcessor) SetFilterParameters(cutoff, resonance, envAmount float64) {
	svp.filterCutoff = cutoff
	svp.filterResonance = resonance
	svp.filterEnvAmount = envAmount
}

// TriggerFilterEnvelope triggers the filter envelope
func (svp *SynthVoiceProcessor) TriggerFilterEnvelope() {
	if svp.filterEnvelope != nil {
		svp.filterEnvelope.Trigger()
	}
}

// ReleaseFilterEnvelope releases the filter envelope
func (svp *SynthVoiceProcessor) ReleaseFilterEnvelope() {
	if svp.filterEnvelope != nil {
		svp.filterEnvelope.Release()
	}
}