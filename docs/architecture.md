# ClapGo Framework Architecture

ClapGo is a Go-native framework for building CLAP audio plugins. It follows a layered architecture design that separates concerns and provides a clean path from simple C bridge functionality to rich, developer-friendly abstractions.

## Table of Contents

- [Architectural Overview](#architectural-overview)
- [Layer 1: C Bridge](#layer-1-c-bridge)
- [Layer 2: Go Framework Core](#layer-2-go-framework-core)
- [Layer 3: Audio Processing](#layer-3-audio-processing)
- [Layer 4: Developer Conveniences](#layer-4-developer-conveniences)
- [Package Organization](#package-organization)
- [Design Principles](#design-principles)
- [Framework Philosophy](#framework-philosophy)

## Architectural Overview

ClapGo is structured as a 4-layer architecture:

```
┌─────────────────────────────────────────────────────────────┐
│ Layer 4: Developer Conveniences                            │
│ ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐ │
│ │   Templates     │ │    Builders     │ │   Generators    │ │
│ │   & Examples    │ │   & Factories   │ │   & Utilities   │ │
│ └─────────────────┘ └─────────────────┘ └─────────────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│ Layer 3: Audio Processing (DSP)                            │
│ ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐ │
│ │   Oscillators   │ │    Filters      │ │   Envelopes     │ │
│ │   & Waveforms   │ │   & Effects     │ │   & Dynamics    │ │
│ └─────────────────┘ └─────────────────┘ └─────────────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│ Layer 2: Go Framework Core                                 │
│ ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐ │
│ │   Parameters    │ │     Events      │ │      State      │ │
│ │   & Automation  │ │    & MIDI       │ │   & Presets     │ │
│ └─────────────────┘ └─────────────────┘ └─────────────────┘ │
│ ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐ │
│ │     Audio       │ │    Plugin       │ │   Extensions    │ │
│ │   Buffers/I/O   │ │     Base        │ │   & Host API    │ │
│ └─────────────────┘ └─────────────────┘ └─────────────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│ Layer 1: C Bridge (Minimal)                                │
│ ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐ │
│ │  Export/Import  │ │   C Interop     │ │    Manifest     │ │
│ │   Functions     │ │   & Types       │ │   Discovery     │ │
│ └─────────────────┘ └─────────────────┘ └─────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Layer 1: C Bridge

The C bridge is **intentionally minimal** and serves only to interface between the CLAP C API and Go. It contains:

### Core Components
- **C Export Functions**: Direct mapping of CLAP C functions to Go
- **Type Conversions**: Basic C ↔ Go type translation
- **Manifest Discovery**: JSON-based plugin discovery system

### Key Characteristics
- ✅ **Thin and Direct**: Each CLAP C function maps to exactly one Go function
- ✅ **No Business Logic**: Pure bridging functionality only
- ✅ **Generated Code**: C exports are auto-generated, not hand-written
- ❌ **No Framework Features**: No parameter managers, builders, or abstractions

### Files
```
src/c/          # C bridge implementation
├── bridge.c    # Core CLAP function exports
├── bridge.h    # C header definitions
└── manifest.c  # Plugin discovery
```

## Layer 2: Go Framework Core

This layer provides Go-native abstractions that feel natural to Go developers while maintaining full access to CLAP concepts.

### Parameter Management (`pkg/param/`)

**Thread-safe parameter handling with atomic access:**

```go
// Parameter Manager - central registry
manager := param.NewManager()

// Register parameters using Info struct
info := param.Info{
    ID:           1,
    Name:         "Cutoff",
    MinValue:     20.0,
    MaxValue:     20000.0,
    DefaultValue: 1000.0,
    Flags:        param.FlagAutomatable | param.FlagBoundedBelow,
}
manager.Register(info)

// Thread-safe parameter access
manager.Set(1, 2500.0)  // Set cutoff to 2.5kHz
value := manager.Get(1) // Get current value atomically
```

**Factory Functions for Common Parameter Types:**

```go
// Common parameter factories
volume := param.Volume(1, "Volume")     // 0-2 (linear gain)
pan := param.Pan(2, "Pan")              // -1 to 1
cutoff := param.Cutoff(3, "Cutoff")     // 20Hz-20kHz
choice := param.Choice(4, "Type", 3, 0) // 3 options, default 0
```

**Advanced Parameter Binding System:**

```go
// Parameter binder automates management
binder := param.NewParameterBinder(manager)

// Automatic registration + atomic storage
volume := binder.BindPercentage(1, "Volume", 70.0)
waveform := binder.BindChoice(2, "Waveform", []string{"Sine", "Saw", "Square"}, 0)
cutoff := binder.BindCutoffLog(3, "Cutoff", 0.5) // 0-1 → 20Hz-8kHz logarithmic

// Direct atomic access in audio thread
currentVolume := volume.Load()          // Thread-safe
cutoffFreq, _ := binder.GetMappedValue(3) // Gets mapped frequency value
```

### Event Processing (`pkg/event/`)

**High-level event handling abstractions:**

```go
// Event processor handles all CLAP events
processor := event.NewProcessor(inEvents, outEvents)

// Process all events with custom handler
processor.ProcessAll(myPlugin)

// Create and send events
noteEvent := event.CreateNoteOnEvent(0, -1, -1, 0, 60, 0.8)
processor.PushOutputEvent(noteEvent)
```

**MIDI Processing:**

```go
// Specialized MIDI processor
midiProcessor := audio.NewMIDIProcessor(voiceManager, paramManager)

// Set callbacks for MIDI events
midiProcessor.SetCallbacks(
    func(channel, key int16, velocity float64) {
        // Handle note on
    },
    func(channel, key int16, velocity float64) {
        // Handle note off
    },
    func(channel int16, cc uint32, value float64) {
        // Handle CC changes
    },
    nil, // pitch bend
    nil, // poly pressure
)
```

### Audio I/O (`pkg/audio/`)

**Buffer management and audio processing:**

```go
// Audio buffer type
type Buffer [][]float32 // [channel][sample]

// Helper functions
frames := buf.Frames()    // Number of samples per channel
channels := buf.Channels() // Number of channels

// DSP utilities
audio.ApplyPan(buf, 0.5)           // Apply panning
audio.Clip(buf, 1.0)               // Hard clipping
audio.SoftClip(buf)                // Soft clipping (tanh)
audio.CrossFade(dst, from, to)     // Crossfade between buffers
```

### Plugin Base (`pkg/plugin/`)

**Foundation for all plugins:**

```go
type MyPlugin struct {
    *plugin.PluginBase
    // ... your plugin-specific fields
}

func NewMyPlugin() *MyPlugin {
    info := plugin.Info{
        ID:          "com.mycompany.myplugin",
        Name:        "My Plugin",
        Vendor:      "My Company",
        Version:     "1.0.0",
        Description: "Does amazing things",
        Features:    []string{plugin.FeatureAudioEffect, plugin.FeatureStereo},
    }
    
    return &MyPlugin{
        PluginBase: plugin.NewPluginBase(info),
    }
}
```

### State Management (`pkg/state/`)

**Automatic plugin state handling:**

```go
// State is automatically managed by PluginBase
func (p *MyPlugin) SaveState(stream unsafe.Pointer) error {
    return p.SaveStateWithParams(stream, map[uint32]float64{
        ParamGain: p.gain.Load(),
        ParamPan:  p.pan.Load(),
    })
}

func (p *MyPlugin) LoadState(stream unsafe.Pointer) error {
    return p.LoadStateWithCallback(stream, func(id uint32, value float64) {
        switch id {
        case ParamGain:
            p.gain.Store(value)
        case ParamPan:
            p.pan.Store(value)
        }
    })
}
```

## Layer 3: Audio Processing

High-performance DSP components for real-time audio processing.

### Oscillators (`pkg/audio/oscillator.go`)

```go
// Basic waveform generation
phase := 0.0
frequency := 440.0 // A4
sampleRate := 44100.0

for i := range buffer {
    sample := audio.GenerateWaveformSample(phase, audio.WaveformSine)
    buffer[i] = float32(sample)
    phase = audio.AdvancePhase(phase, frequency, sampleRate)
}

// Anti-aliased waveforms
phaseIncrement := frequency / sampleRate
sawSample := audio.GeneratePolyBLEPSaw(phase, phaseIncrement)
squareSample := audio.GeneratePolyBLEPSquare(phase, phaseIncrement)
```

### Envelopes (`pkg/audio/envelope.go`)

```go
// ADSR envelope
envelope := audio.NewADSREnvelope(sampleRate)
envelope.SetADSR(0.01, 0.1, 0.7, 0.3) // Attack, Decay, Sustain, Release

// Trigger and process
envelope.Trigger()              // Start attack
for i := range buffer {
    envValue := envelope.Process() // Get current envelope value
    buffer[i] *= float32(envValue) // Apply to audio
}
envelope.Release()              // Start release when note ends
```

### Voice Management (`pkg/audio/voice.go`)

```go
// Polyphonic voice management
voiceManager := audio.NewVoiceManager(16, sampleRate) // 16 voices

// Trigger notes
voice := voiceManager.TriggerNote(0, 60, 0.8, -1) // Channel, key, velocity, noteID

// Process all active voices
voiceManager.ApplyToAllVoices(func(voice *audio.Voice) {
    if voice.IsActive {
        // Process this voice
        sample := generateVoiceSample(voice)
        // Add to output buffer
    }
})
```

### DSP Utilities (`pkg/audio/dsp.go`)

```go
// Gain and level utilities
linearGain := audio.DbToLinear(-6.0)    // -6dB to linear
dbValue := audio.LinearToDb(0.5)        // Linear to dB

// Panning
leftGain, rightGain := audio.Pan(-0.3)  // Constant power pan

// Format conversion
audio.StereoToMono(mono, stereo)        // Stereo to mono
audio.MonoToStereo(stereo, mono)        // Mono to stereo

// Clipping and dynamics
audio.Clip(buffer, 1.0)                 // Hard clip to ±1.0
audio.SoftClip(buffer)                  // Soft clip using tanh
```

## Layer 4: Developer Conveniences

Tools and patterns that make common plugin development tasks fast and enjoyable.

### Parameter Builders

**Fluent API for parameter creation:**

```go
// Fluent parameter building
param := param.NewBuilder(1, "Filter Cutoff").
    Module("Filter").
    Range(20, 20000, 1000).
    Format(param.FormatHertz).
    Automatable().
    Modulatable().
    Bounded().
    Build()
```

### Plugin Templates

**Scaffolding for common plugin types:**

```bash
# Generate plugin scaffolding (planned)
clapgo generate --type=effect --name=MyEffect
clapgo generate --type=synth --name=MySynth
```

### Remote Controls Builder

**Easy remote control page creation:**

```go
// Remote control pages for hardware controllers
page := controls.NewRemoteControlsPageBuilder(0, "Main Controls").
    Section("Filter").
    AddParameters(
        ParamCutoff,
        ParamResonance,
        ParamEnvAmount,
    ).
    Section("Envelope").
    AddParameters(
        ParamAttack,
        ParamDecay,
        ParamSustain,
        ParamRelease,
    ).
    ClearRemaining().
    MustBuild()
```

## Package Organization

ClapGo follows Go's domain-driven package design:

```
pkg/
├── bridge/         # Layer 1: C bridge (minimal)
│   ├── bridge.go   # Core bridge functionality
│   └── manifest.go # Plugin discovery
├── plugin/         # Layer 2: Plugin foundation
│   ├── base.go     # PluginBase implementation
│   ├── interface.go # Plugin interface definition
│   └── options.go  # Plugin configuration
├── param/          # Layer 2: Parameter management
│   ├── manager.go  # Parameter manager
│   ├── builder.go  # Fluent parameter builder
│   ├── factory.go  # Common parameter factories
│   ├── binding.go  # Automatic parameter binding
│   └── format.go   # Parameter display formatting
├── event/          # Layer 2: Event processing
│   ├── processor.go # High-level event handling
│   ├── midi.go     # MIDI event utilities
│   └── types.go    # Event type definitions
├── audio/          # Layer 2/3: Audio processing
│   ├── buffer.go   # Audio buffer management
│   ├── dsp.go      # Basic DSP utilities
│   ├── oscillator.go # Oscillator implementations
│   ├── envelope.go # Envelope generators
│   ├── voice.go    # Voice management
│   └── ports.go    # Audio port configuration
├── state/          # Layer 2: State management
│   ├── manager.go  # State serialization
│   └── migration.go # Version migration
├── extension/      # Layer 2: CLAP extensions
│   ├── bundle.go   # Extension bundle
│   └── *.go        # Individual extension wrappers
├── controls/       # Layer 4: Remote controls
│   ├── builder.go  # Control page builder
│   └── remote.go   # Remote control utilities
└── manifest/       # Layer 1: Plugin discovery
    ├── manifest.go # Manifest handling
    └── util.go     # Utilities
```

## Design Principles

### 1. Layered Architecture

**Clear separation of concerns:**
- Layer 1: Minimal C bridge (direct CLAP mapping only)
- Layer 2: Go-native abstractions (idiomatic Go)
- Layer 3: High-level audio processing (DSP focus)
- Layer 4: Developer productivity (convenience and templates)

### 2. Go Idioms

**Follow standard Go patterns:**
- Small, focused interfaces (1-3 methods)
- Accept interfaces, return concrete types
- Proper error handling with context
- Thread-safe APIs using atomic operations
- Functional options for extensible APIs

### 3. Performance First

**Optimized for real-time audio:**
- Lock-free atomic operations in audio threads
- Zero-allocation audio processing paths
- Efficient memory management
- Pool reuse for temporary objects

### 4. Developer Experience

**Make audio development enjoyable:**
- Sensible defaults with override capability
- Comprehensive examples and documentation
- Fast iteration cycles
- Clear error messages and debugging support

### 5. Escape Hatches

**Preserve access to lower layers:**
- Direct CLAP API access when needed
- Plugin base overrides for custom behavior
- Raw parameter access alongside automation
- Manual event handling for special cases

## Framework Philosophy

### Core Tenets

1. **"Simple things should be simple, complex things should be possible"**
   - Common tasks have high-level abstractions
   - Complex scenarios provide low-level access

2. **"Go-native, not Go-wrapped"**
   - APIs feel natural to Go developers
   - Standard library patterns where applicable
   - Proper Go error handling and concurrency

3. **"Framework, not library"**
   - Handles boilerplate automatically
   - Provides structure and conventions
   - Enables rapid prototyping

4. **"Performance by design"**
   - Real-time audio constraints built-in
   - Zero-allocation hot paths
   - Thread-safe by default

### Anti-Goals

- ❌ **Rigid framework patterns** - Don't box developers into specific approaches
- ❌ **Over-abstraction** - Don't hide CLAP concepts completely
- ❌ **C-style APIs in Go** - Avoid direct C API translations
- ❌ **Framework lock-in** - Always provide escape hatches to lower layers

### Success Metrics

A successful ClapGo framework enables developers to:

1. **Build a basic gain plugin in < 50 lines of Go**
2. **Create a polyphonic synthesizer in < 200 lines**
3. **Focus on audio algorithms, not infrastructure**
4. **Debug audio issues easily**
5. **Achieve professional audio quality and performance**

This architecture ensures that ClapGo grows with developers' needs - from simple effects to complex instruments - while maintaining the performance and reliability required for professional audio software.