# Synth Plugin Refactor Implementation Strategy

## Overview

This document provides a comprehensive strategy to refactor the synth plugin from **1781 lines** to approximately **400-500 lines** by extracting reusable components to the ClapGo framework and utilizing existing domain-specific packages.

## Current State Analysis

**File**: `examples/synth/main.go`  
**Current Lines**: 1781  
**Target Lines**: 400-500  
**Reduction Goal**: ~70-75%

### Key Components to Extract to Framework

1. **ADSR Envelope Implementation** - Custom envelope code that duplicates framework
2. **Polyphonic Parameter Modulation** - Per-voice parameter overrides

## Detailed Refactoring Strategy

### Phase 1: Simplify Plugin Structure

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

### Phase 2: Simplify Process Method

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

### Phase 3: Remove Redundant Methods

**Methods to Remove/Simplify**:
- GetAvailablePresets (not used)
- Complex timer support (if not needed)
- Manual thread checking
- Verbose logging throughout
- Duplicate waveform generation code (use framework)

**Savings**: ~200 lines

### Phase 4: Use Builder Pattern for Parameters

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

1. **Phase 1**: Simplify plugin structure (~30 mins)
2. **Phase 2**: Simplify process method (~30 mins)
3. **Phase 3**: Remove redundant code (~30 mins)
4. **Phase 4**: Use parameter builders (~30 mins)

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

All required framework additions have been completed. The framework now includes:
- Voice management system
- MIDI event processing
- Synth-specific DSP utilities
- Basic audio filters

## Benefits Already Available to Developers

1. **Voice Management**: No need to implement voice allocation/stealing
2. **MIDI Processing**: Standard MIDI event handling out of the box
3. **DSP Building Blocks**: Oscillators, filters, envelopes ready to use
4. **No Preset Complexity**: DAWs handle preset management
5. **Thread Safety**: Built-in atomic parameter handling
6. **Polyphonic Modulation**: Per-voice parameter system included

## Remaining Benefits to Implement

1. **Parameter Builders**: Fluent API for common parameter types

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
