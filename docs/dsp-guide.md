# ClapGo DSP Guide

This guide covers the audio processing capabilities of ClapGo, from basic DSP utilities to complex audio components like oscillators, envelopes, and filters.

## Table of Contents

- [Basic DSP Utilities](#basic-dsp-utilities)
- [Oscillators and Waveform Generation](#oscillators-and-waveform-generation)
- [Envelope Generators](#envelope-generators)
- [Voice Management](#voice-management)
- [Audio Buffer Processing](#audio-buffer-processing)
- [Advanced DSP Patterns](#advanced-dsp-patterns)
- [Performance Considerations](#performance-considerations)

## Basic DSP Utilities

### Audio Level Conversion

ClapGo provides utilities for common audio level conversions:

```go
package main

import "github.com/justyntemme/clapgo/pkg/audio"

func ExampleLevelConversion() {
    // Convert between linear gain and decibels
    linearGain := audio.DbToLinear(-6.0)    // -6dB → 0.5 linear
    dbValue := audio.LinearToDb(0.5)        // 0.5 linear → -6.02dB
    
    // Float32 versions for performance
    linearGain32 := audio.DbToLinearFloat32(-6.0)
    dbValue32 := audio.LinearToDbFloat32(0.5)
    
    // Safe conversion (handles -∞ dB)
    quietGain := audio.DbToLinear(-120.0)   // Very quiet, not zero
    silentDb := audio.LinearToDb(0.0)       // Returns -120dB (practical -∞)
}
```

### Panning and Stereo Processing

```go
func ExamplePanning() {
    // Constant power panning (-1 = full left, +1 = full right)
    leftGain, rightGain := audio.Pan(-0.3)  // Slightly left
    
    // Apply panning to stereo buffer
    var stereoBuffer audio.Buffer = make([][]float32, 2)
    stereoBuffer[0] = make([]float32, 512) // Left channel
    stereoBuffer[1] = make([]float32, 512) // Right channel
    
    panPosition := float32(-0.3)
    err := audio.ApplyPan(stereoBuffer, panPosition)
    if err != nil {
        // Handle error (e.g., not stereo)
    }
}
```

### Format Conversion

```go
func ExampleFormatConversion() {
    // Stereo to mono conversion
    stereoBuffer := make([][]float32, 2)
    stereoBuffer[0] = make([]float32, 512) // Left
    stereoBuffer[1] = make([]float32, 512) // Right
    
    monoBuffer := make([]float32, 512)
    err := audio.StereoToMono(monoBuffer, stereoBuffer)
    if err != nil {
        // Handle conversion error
    }
    
    // Mono to stereo conversion
    monoSrc := make([]float32, 512)
    stereoDst := make([][]float32, 2)
    stereoDst[0] = make([]float32, 512)
    stereoDst[1] = make([]float32, 512)
    
    err = audio.MonoToStereo(stereoDst, monoSrc)
}
```

### Clipping and Dynamics

```go
func ExampleClipping() {
    buffer := make([][]float32, 2)
    buffer[0] = make([]float32, 512)
    buffer[1] = make([]float32, 512)
    
    // Hard clipping to ±1.0
    audio.Clip(buffer, 1.0)
    
    // Soft clipping using tanh
    audio.SoftClip(buffer)  // Gentle saturation
}
```

### Fading and Crossfading

```go
func ExampleFading() {
    buffer := make([][]float32, 2)
    buffer[0] = make([]float32, 512)
    buffer[1] = make([]float32, 512)
    
    // Linear fade in
    audio.Fade(buffer, true)   // Fade in
    audio.Fade(buffer, false)  // Fade out
    
    // Crossfade between two buffers
    fromBuffer := make([][]float32, 2)
    toBuffer := make([][]float32, 2)
    outputBuffer := make([][]float32, 2)
    // ... initialize buffers ...
    
    err := audio.CrossFade(outputBuffer, fromBuffer, toBuffer)
    if err != nil {
        // Handle crossfade error
    }
}
```

## Oscillators and Waveform Generation

### Basic Waveform Generation

```go
func ExampleBasicOscillator() {
    var phase float64 = 0.0
    frequency := 440.0  // A4
    sampleRate := 44100.0
    
    buffer := make([]float32, 512)
    
    for i := range buffer {
        // Generate different waveforms
        sample := audio.GenerateWaveformSample(phase, audio.WaveformSine)
        buffer[i] = float32(sample)
        
        // Advance phase for next sample
        phase = audio.AdvancePhase(phase, frequency, sampleRate)
    }
    
    // Available waveforms:
    // - audio.WaveformSine
    // - audio.WaveformSaw
    // - audio.WaveformSquare
    // - audio.WaveformTriangle
    // - audio.WaveformNoise
}
```

### Anti-Aliased Waveforms

For better quality, especially with high frequencies, use anti-aliased waveforms:

```go
func ExampleAntiAliasedOscillator() {
    var phase float64 = 0.0
    frequency := 8000.0  // High frequency
    sampleRate := 44100.0
    phaseIncrement := frequency / sampleRate
    
    buffer := make([]float32, 512)
    
    for i := range buffer {
        // Anti-aliased sawtooth using PolyBLEP
        sawSample := audio.GeneratePolyBLEPSaw(phase, phaseIncrement)
        
        // Anti-aliased square wave
        squareSample := audio.GeneratePolyBLEPSquare(phase, phaseIncrement)
        
        buffer[i] = float32(sawSample * 0.5) // Use sawtooth
        
        phase = audio.AdvancePhase(phase, frequency, sampleRate)
    }
}
```

### MIDI Note to Frequency

```go
func ExampleMIDIToFrequency() {
    // Convert MIDI note numbers to frequencies
    a4Freq := audio.NoteToFrequency(69)   // 440.0 Hz (A4)
    c4Freq := audio.NoteToFrequency(60)   // 261.63 Hz (Middle C)
    
    // Use in oscillator
    midiNote := 72  // C5
    frequency := audio.NoteToFrequency(midiNote)
    
    // Generate audio for this note
    var phase float64 = 0.0
    for i := range buffer {
        sample := audio.GenerateWaveformSample(phase, audio.WaveformSine)
        buffer[i] = float32(sample)
        phase = audio.AdvancePhase(phase, frequency, 44100.0)
    }
}
```

## Envelope Generators

### ADSR Envelope

The ADSR (Attack, Decay, Sustain, Release) envelope is essential for musical instruments:

```go
func ExampleADSREnvelope() {
    sampleRate := 44100.0
    envelope := audio.NewADSREnvelope(sampleRate)
    
    // Configure ADSR parameters (all times in seconds, sustain is level 0-1)
    envelope.SetADSR(
        0.01,  // Attack: 10ms
        0.1,   // Decay: 100ms
        0.7,   // Sustain: 70% level
        0.3,   // Release: 300ms
    )
    
    // Simulate a note being played
    envelope.Trigger()  // Start the envelope (key press)
    
    buffer := make([]float32, 512)
    for i := range buffer {
        // Process envelope (call once per sample)
        envValue := envelope.Process()
        
        // Generate oscillator sample
        sample := audio.GenerateWaveformSample(phase, audio.WaveformSine)
        
        // Apply envelope to oscillator
        buffer[i] = float32(sample * envValue)
        
        // Advance oscillator phase
        phase = audio.AdvancePhase(phase, 440.0, sampleRate)
    }
    
    // Later, when note is released
    envelope.Release()  // Start release phase
    
    // Continue processing...
    for envelope.IsActive() {
        envValue := envelope.Process()
        // Apply to audio...
    }
}
```

### Envelope States and Control

```go
func ExampleEnvelopeControl() {
    envelope := audio.NewADSREnvelope(44100.0)
    
    // Check envelope state
    switch envelope.Stage {
    case audio.EnvelopeStageIdle:
        // Envelope is not active
    case audio.EnvelopeStageAttack:
        // In attack phase
    case audio.EnvelopeStageDecay:
        // In decay phase
    case audio.EnvelopeStageSustain:
        // Holding at sustain level
    case audio.EnvelopeStageRelease:
        // In release phase
    }
    
    // Check if envelope is generating output
    if envelope.IsActive() {
        // Envelope is currently active
    }
    
    // Reset envelope immediately
    envelope.Reset()
    
    // Get current envelope value without advancing
    currentLevel := envelope.CurrentValue
}
```

### Stateless ADSR for Simple Cases

For simpler use cases, you can use the stateless ADSR function:

```go
func ExampleStatelessADSR() {
    sampleRate := 44100.0
    attack := 0.01
    decay := 0.1
    sustain := 0.7
    release := 0.3
    
    var elapsedSamples uint32 = 0
    isReleased := false
    var releaseStartSample uint32 = 0
    
    buffer := make([]float32, 512)
    for i := range buffer {
        // Calculate envelope value for current time
        envValue := audio.SimpleADSR(
            elapsedSamples,
            sampleRate,
            attack, decay, sustain, release,
            isReleased,
            releaseStartSample,
        )
        
        // Apply to audio
        sample := audio.GenerateWaveformSample(phase, audio.WaveformSine)
        buffer[i] = float32(sample * envValue)
        
        elapsedSamples++
        phase = audio.AdvancePhase(phase, 440.0, sampleRate)
    }
    
    // Trigger release at some point
    isReleased = true
    releaseStartSample = elapsedSamples
}
```

## Voice Management

For polyphonic instruments, you need to manage multiple voices:

```go
func ExampleVoiceManager() {
    sampleRate := 44100.0
    maxVoices := 16
    
    voiceManager := audio.NewVoiceManager(maxVoices, sampleRate)
    
    // Trigger a note (creates a voice)
    voice := voiceManager.TriggerNote(
        0,    // MIDI channel
        60,   // MIDI note (Middle C)
        0.8,  // Velocity (0-1)
        123,  // Note ID (-1 for no ID)
    )
    
    if voice != nil {
        // Note was triggered successfully
        // Voice contains: Channel, Key, Velocity, NoteID, Frequency, Envelope, etc.
    }
    
    // Process all active voices
    outputBuffer := make([]float32, 512)
    
    voiceManager.ApplyToAllVoices(func(voice *audio.Voice) {
        if !voice.IsActive {
            return  // Skip inactive voices
        }
        
        // Process this voice
        for i := range outputBuffer {
            // Generate oscillator sample for this voice
            sample := audio.GenerateWaveformSample(voice.Phase, audio.WaveformSine)
            
            // Apply envelope
            envValue := voice.Envelope.Process()
            sample *= envValue
            
            // Apply voice volume
            sample *= voice.Volume
            
            // Add to output buffer (mix with other voices)
            outputBuffer[i] += float32(sample)
            
            // Advance phase
            voice.Phase = audio.AdvancePhase(voice.Phase, voice.Frequency, sampleRate)
        }
        
        // Mark voice as inactive if envelope finished
        if !voice.Envelope.IsActive() {
            voice.IsActive = false
        }
    })
    
    // Release a note
    voiceManager.ReleaseNote(0, 60, -1)  // Channel, key, note ID
    
    // Get voice statistics
    activeCount := voiceManager.GetActiveVoiceCount()
    totalVoices := voiceManager.GetTotalVoiceCount()
}
```

## Audio Buffer Processing

### Buffer Type and Operations

ClapGo uses a `Buffer` type for multi-channel audio:

```go
func ExampleBufferOperations() {
    // Create a stereo buffer with 512 samples
    buffer := make(audio.Buffer, 2)  // 2 channels
    buffer[0] = make([]float32, 512) // Left channel
    buffer[1] = make([]float32, 512) // Right channel
    
    // Get buffer information
    channels := buffer.Channels()  // Returns 2
    frames := buffer.Frames()      // Returns 512
    
    // Process stereo buffer with a function
    audio.ProcessStereo(buffer, buffer, func(sample float32) float32 {
        return sample * 0.5  // Reduce volume by half
    })
}
```

### Buffer Validation

```go
func ExampleBufferValidation() {
    inputBuffer := make(audio.Buffer, 2)
    outputBuffer := make(audio.Buffer, 2)
    
    // Validate that buffers are compatible
    if audio.ValidateBuffers(outputBuffer, inputBuffer) {
        // Buffers are valid for processing
        audio.ProcessStereo(inputBuffer, outputBuffer, func(sample float32) float32 {
            return audio.Clamp(sample, -1.0, 1.0)  // Ensure valid range
        })
    }
}
```

## Advanced DSP Patterns

### Finding Peak Values

The `findPeak` function (as seen in the synth example) is a common DSP utility:

```go
// findPeak returns the maximum absolute value in the buffer
func findPeak(buffer []float32) float32 {
    peak := float32(0.0)
    for _, sample := range buffer {
        abs := sample
        if abs < 0 {
            abs = -abs
        }
        if abs > peak {
            peak = abs
        }
    }
    return peak
}

func ExamplePeakDetection() {
    buffer := make([]float32, 512)
    // ... fill buffer with audio data ...
    
    peak := findPeak(buffer)
    
    // Convert to dB for display
    peakDb := audio.LinearToDb(float64(peak))
    
    // Use for limiting or metering
    if peak > 0.95 {
        // Apply limiting
        for i := range buffer {
            buffer[i] *= 0.95 / peak
        }
    }
}
```

### Utility Functions

```go
func ExampleUtilities() {
    // Clamp values to range
    value := audio.Clamp(1.5, 0.0, 1.0)  // Returns 1.0
    
    // These utilities are useful for parameter processing:
    frequency := audio.Clamp(userInput, 20.0, 20000.0)
    gain := audio.Clamp(userGain, 0.0, 2.0)
    pan := audio.Clamp(userPan, -1.0, 1.0)
}
```

### Building Complex Audio Processors

Here's an example of combining multiple DSP components:

```go
type SimpleFilter struct {
    cutoff      float64
    resonance   float64
    sampleRate  float64
    
    // Internal state for IIR filter
    x1, x2      float64  // Input history
    y1, y2      float64  // Output history
}

func (f *SimpleFilter) Process(input float64) float64 {
    // Simple 2-pole lowpass filter
    // This is a simplified example - real filters are more complex
    
    omega := 2.0 * math.Pi * f.cutoff / f.sampleRate
    sin := math.Sin(omega)
    cos := math.Cos(omega)
    alpha := sin / (2.0 * f.resonance)
    
    b0 := (1.0 - cos) / 2.0
    b1 := 1.0 - cos
    b2 := (1.0 - cos) / 2.0
    a0 := 1.0 + alpha
    a1 := -2.0 * cos
    a2 := 1.0 - alpha
    
    // Apply filter equation
    output := (b0/a0)*input + (b1/a0)*f.x1 + (b2/a0)*f.x2 - (a1/a0)*f.y1 - (a2/a0)*f.y2
    
    // Update history
    f.x2 = f.x1
    f.x1 = input
    f.y2 = f.y1
    f.y1 = output
    
    return output
}

func ExampleComplexProcessor() {
    // Combine oscillator, envelope, and filter
    var phase float64 = 0.0
    envelope := audio.NewADSREnvelope(44100.0)
    filter := &SimpleFilter{
        cutoff:     1000.0,
        resonance:  2.0,
        sampleRate: 44100.0,
    }
    
    envelope.Trigger()
    
    buffer := make([]float32, 512)
    for i := range buffer {
        // Generate oscillator
        oscSample := audio.GenerateWaveformSample(phase, audio.WaveformSaw)
        
        // Apply envelope
        envValue := envelope.Process()
        oscSample *= envValue
        
        // Apply filter
        filteredSample := filter.Process(oscSample)
        
        // Store final sample
        buffer[i] = float32(filteredSample)
        
        // Advance oscillator
        phase = audio.AdvancePhase(phase, 440.0, 44100.0)
    }
}
```

## Performance Considerations

### Memory Allocation

Avoid allocations in the audio processing thread:

```go
// ❌ Bad: Allocates memory in audio thread
func processAudioBad(input []float32) []float32 {
    output := make([]float32, len(input))  // Allocation!
    // ... process ...
    return output
}

// ✅ Good: Pre-allocate buffers
type Processor struct {
    workBuffer []float32
}

func NewProcessor(maxFrames int) *Processor {
    return &Processor{
        workBuffer: make([]float32, maxFrames),
    }
}

func (p *Processor) processAudio(input, output []float32) {
    // Use pre-allocated buffer
    work := p.workBuffer[:len(input)]
    // ... process using work buffer ...
}
```

### Atomic Operations

Use atomic operations for thread-safe parameter access:

```go
import "sync/atomic"

type AtomicProcessor struct {
    cutoffBits int64  // Store float64 as int64 bits
}

func (p *AtomicProcessor) SetCutoff(freq float64) {
    bits := math.Float64bits(freq)
    atomic.StoreInt64(&p.cutoffBits, int64(bits))
}

func (p *AtomicProcessor) GetCutoff() float64 {
    bits := atomic.LoadInt64(&p.cutoffBits)
    return math.Float64frombits(uint64(bits))
}

func (p *AtomicProcessor) Process(buffer []float32) {
    cutoff := p.GetCutoff()  // Thread-safe read
    // ... use cutoff in processing ...
}
```

### Efficient Processing Patterns

```go
// Process buffers efficiently
func ExampleEfficientProcessing() {
    buffer := make([][]float32, 2)
    buffer[0] = make([]float32, 512)
    buffer[1] = make([]float32, 512)
    
    // Process both channels in one loop
    for i := range buffer[0] {
        // Read once, process both channels
        left := buffer[0][i]
        right := buffer[1][i]
        
        // Apply processing
        left *= 0.8
        right *= 0.8
        
        // Write back
        buffer[0][i] = left
        buffer[1][i] = right
    }
}
```

This DSP guide covers the essential audio processing capabilities of ClapGo. For more advanced DSP techniques, consider the specific implementations in the examples directory, particularly the `synth` example which demonstrates polyphonic synthesis with filters and envelopes.