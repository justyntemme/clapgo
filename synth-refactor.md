# Synth Plugin Refactor Implementation Strategy

## Overview

This document provides a comprehensive strategy to refactor the synth plugin from **1781 lines** to approximately **400-500 lines** by extracting reusable components to the ClapGo framework and utilizing existing domain-specific packages.

## Current State Analysis

**File**: `examples/synth/main.go`  
**Current Lines**: 1781  
**Target Lines**: 400-500  
**Reduction Goal**: ~70-75%

### Key Components to Extract to Framework

1. **Voice Management System** - Currently ~200 lines of voice allocation/deallocation
2. **ADSR Envelope Implementation** - Custom envelope code that duplicates framework
3. **MIDI Event Processing** - Note on/off, pitch bend, modulation handling
4. **Polyphonic Parameter Modulation** - Per-voice parameter overrides
5. **Preset Loading System** - Removed (DAWs have their own preset systems)
6. **Export Functions** - ~500 lines of C export boilerplate

## Detailed Refactoring Strategy

### Phase 1: Extract Voice Management to Framework (pkg/audio/voice.go)

**Current**: Lines ~660-684 (Voice struct) + voice allocation logic scattered throughout
**Target**: Create reusable voice management system

```go
// pkg/audio/voice.go
package audio

type Voice struct {
    NoteID    int32
    Channel   int16
    Key       int16
    Velocity  float64
    
    // Oscillator state
    Phase     float64
    Frequency float64
    
    // Envelope
    Envelope  *ADSREnvelope
    
    // Modulation
    PitchBend   float64
    Brightness  float64
    Pressure    float64
    
    // State
    IsActive    bool
    TuningID    uint64
}

type VoiceManager struct {
    voices      []*Voice
    maxVoices   int
    sampleRate  float64
}

func NewVoiceManager(maxVoices int, sampleRate float64) *VoiceManager
func (vm *VoiceManager) AllocateVoice(noteID int32, channel, key int16, velocity float64) *Voice
func (vm *VoiceManager) ReleaseVoice(noteID int32, channel int16)
func (vm *VoiceManager) ReleaseAllVoices()
func (vm *VoiceManager) ProcessVoices(processFunc func(*Voice) float64) []float64
```

**Savings**: ~150 lines

### Phase 2: Use Framework ADSR Envelope (Already Exists!)

**Current**: Lines ~1389-1449 custom getEnvelopeValue implementation
**Action**: Replace with existing `audio.ADSREnvelope` from framework

```go
// REMOVE custom envelope logic
// USE: audio.NewADSREnvelope() and envelope.Process()
```

**Savings**: ~60 lines

### Phase 3: Extract MIDI Event Processing (pkg/audio/midi.go)

**Current**: Lines ~1025-1388 processEventHandler with complex MIDI logic
**Target**: Create reusable MIDI processor

```go
// pkg/audio/midi.go
type MIDIProcessor struct {
    voiceManager *VoiceManager
    paramManager *param.Manager
}

func (m *MIDIProcessor) ProcessNoteOn(channel, key int16, velocity float64, noteID int32)
func (m *MIDIProcessor) ProcessNoteOff(channel, key int16, noteID int32)
func (m *MIDIProcessor) ProcessPitchBend(channel int16, value float64)
func (m *MIDIProcessor) ProcessModulation(channel int16, cc uint32, value float64)
func (m *MIDIProcessor) ProcessPolyPressure(channel, key int16, pressure float64)
```

**Savings**: ~350 lines

### Phase 4: Consolidate Export Functions

**Current**: Lines ~58-550 individual export functions
**Target**: Generate exports or use a dispatch pattern similar to gain plugin

```go
// Move to separate exports.go file with minimal boilerplate
// Use reflection or code generation to reduce duplication
```

**Savings**: ~400 lines

### Phase 5: ~~Extract Preset System~~ REMOVED

**Preset functionality has been removed entirely since DAWs provide their own preset management systems**

**Savings**: ~70 lines

### Phase 6: Create Synth-Specific DSP Utilities (pkg/audio/synth.go)

**Target**: Extract commonly needed synth building blocks

```go
// pkg/audio/synth.go
package audio

// PolyphonicOscillator manages multiple oscillator voices
type PolyphonicOscillator struct {
    voiceManager *VoiceManager
    waveformType WaveformType
}

func NewPolyphonicOscillator(maxVoices int, sampleRate float64) *PolyphonicOscillator
func (po *PolyphonicOscillator) Process(frameCount uint32) []float32

// SimpleLowPassFilter for brightness control
type SimpleLowPassFilter struct {
    cutoff     float64
    resonance  float64
    sampleRate float64
    state      [2]float64
}

func (f *SimpleLowPassFilter) Process(input float64) float64
func (f *SimpleLowPassFilter) SetCutoff(cutoff float64)

// PitchBendProcessor handles pitch bend with configurable range
type PitchBendProcessor struct {
    bendRange float64 // in semitones
}

func (p *PitchBendProcessor) ApplyPitchBend(baseFreq, bendValue float64) float64
```

### Phase 7: Simplify Plugin Structure

**Current Plugin Structure** (Lines ~686-720):
```go
type SynthPlugin struct {
    // Many individual fields
}
```

**Target Structure**:
```go
type SynthPlugin struct {
    *plugin.PluginBase
    *audio.StereoPortProvider
    event.NoOpHandler
    
    // Synth-specific components
    voiceManager  *audio.VoiceManager
    midiProcessor *audio.MIDIProcessor
    oscillator    *audio.PolyphonicOscillator
    filter        *audio.SimpleLowPassFilter
    
    // Parameters (atomic for thread safety)
    volume    atomic.Value // float64
    waveform  atomic.Value // int
    attack    atomic.Value // float64
    decay     atomic.Value // float64
    sustain   atomic.Value // float64
    release   atomic.Value // float64
}
```

### Phase 8: Simplify Process Method

**Current**: Lines ~901-1024 complex processing logic
**Target**: Clean, readable process method

```go
func (p *SynthPlugin) Process(steadyTime int64, framesCount uint32, 
    audioIn, audioOut [][]float32, events *event.Processor) int {
    
    if !p.IsActivated || !p.IsProcessing {
        return process.ProcessError
    }
    
    // Process MIDI events
    if events != nil {
        p.midiProcessor.ProcessEvents(events)
    }
    
    // Generate audio from all voices
    samples := p.oscillator.Process(framesCount)
    
    // Apply filter if needed
    for i := range samples {
        samples[i] = float32(p.filter.Process(float64(samples[i])))
    }
    
    // Apply master volume and copy to outputs
    volume := p.volume.Load().(float64)
    audio.CopyToStereoOutput(samples, audioOut, float32(volume))
    
    return process.ProcessContinue
}
```

**Savings**: ~100 lines

### Phase 9: Remove Redundant Methods

**Methods to Remove/Simplify**:
- GetAvailablePresets (not used)
- Complex timer support (if not needed)
- Manual thread checking
- Verbose logging throughout
- Duplicate waveform generation code (use framework)

**Savings**: ~200 lines

### Phase 10: Use Builder Pattern for Parameters

**Current**: Manual parameter registration
**Target**: Use parameter builder from framework

```go
func NewSynthPlugin() *SynthPlugin {
    p := &SynthPlugin{
        PluginBase: plugin.NewPluginBase(pluginInfo),
        voiceManager: audio.NewVoiceManager(16, 44100),
    }
    
    // Use builder pattern for parameters
    p.ParamManager.Register(
        param.Volume(ParamVolume, "Volume"),
        param.Choice(ParamWaveform, "Waveform").
            WithOptions("Sine", "Saw", "Square"),
        param.Time(ParamAttack, "Attack").
            WithRange(0.001, 5.0).
            WithDefault(0.01),
        param.Time(ParamDecay, "Decay").
            WithRange(0.001, 5.0).
            WithDefault(0.1),
        param.Percent(ParamSustain, "Sustain").
            WithDefault(0.7),
        param.Time(ParamRelease, "Release").
            WithRange(0.001, 10.0).
            WithDefault(0.3),
    )
    
    return p
}
```

## Implementation Order

1. **Phase 1**: Extract Voice Management (~2 hours) ✅
2. **Phase 2**: Replace custom envelope with framework (~30 mins)
3. **Phase 3**: Extract MIDI processing (~2 hours) ✅
4. **Phase 4**: Consolidate exports (~1 hour)
5. ~~**Phase 5**: Extract preset system~~ REMOVED
6. **Phase 6**: Create synth DSP utilities (~2 hours) ✅
7. **Phase 7-8**: Simplify plugin structure and process (~1 hour)
8. **Phase 9**: Remove redundant code (~30 mins)
9. **Phase 10**: Use parameter builders (~30 mins)

## Expected Results

**Before**: 1781 lines  
**After**: ~450 lines  
**Reduction**: 1331 lines (75%)

**Code Distribution**:
- Imports/package: 20 lines
- Constants: 30 lines
- Plugin struct: 25 lines
- Constructor: 30 lines
- Process method: 40 lines
- Event handling: 50 lines
- State management: 30 lines
- Extension support: 50 lines
- Exports (separate file): 150 lines
- Misc/glue: 25 lines
- **Total**: ~450 lines

## Framework Additions Needed

1. **pkg/audio/voice.go** - Voice and VoiceManager (~150 lines) ✅
2. **pkg/audio/midi.go** - MIDI event processor (~200 lines) ✅
3. **pkg/audio/synth.go** - Synth-specific DSP utilities (~150 lines) ✅
4. ~~**pkg/state/preset.go** - Generic preset loader~~ REMOVED
5. **pkg/audio/filter.go** - Basic filters (already in synth.go) ✅

## Benefits for Future Developers

1. **Voice Management**: No need to implement voice allocation/stealing
2. **MIDI Processing**: Standard MIDI event handling out of the box
3. **DSP Building Blocks**: Oscillators, filters, envelopes ready to use
4. **Parameter Builders**: Fluent API for common parameter types
5. **No Preset Complexity**: DAWs handle preset management
6. **Thread Safety**: Built-in atomic parameter handling
7. **Polyphonic Modulation**: Per-voice parameter system included

## Success Metrics

1. `make install` succeeds
2. `clap-validator` passes all tests
3. Synth remains fully functional
4. Code is more readable and maintainable
5. Framework additions are reusable for other instruments
6. Example clearly shows how to build a synth with ClapGo

## Go Idioms to Apply

1. **Composition over Inheritance**: Use embedded structs effectively
2. **Interface Segregation**: Small, focused interfaces for components
3. **Error Handling**: Proper error propagation, not silent failures
4. **Concurrency**: Use atomic operations and channels appropriately
5. **Builder Pattern**: For complex object construction
6. **Functional Options**: For extensible APIs
7. **Package Organization**: Domain-driven, not utility-driven

---

**Note**: Line numbers are relative positions to maintain accuracy as the file shrinks during refactoring.