# ClapGo

A Go-native framework for building high-performance CLAP audio plugins. ClapGo makes audio plugin development fast, enjoyable, and powerful while maintaining the performance requirements of professional audio software.

## 🎵 Quick Start

### Build a Simple Gain Plugin

```go
package main

import (
    "github.com/justyntemme/clapgo/pkg/audio"
    "github.com/justyntemme/clapgo/pkg/param"
    "github.com/justyntemme/clapgo/pkg/plugin"
)

type GainPlugin struct {
    *plugin.PluginBase
    gain param.AtomicFloat64
}

func NewGainPlugin() *GainPlugin {
    info := plugin.Info{
        ID:          "com.example.gain",
        Name:        "Simple Gain",
        Vendor:      "Example Audio",
        Version:     "1.0.0",
        Description: "A simple gain plugin",
        Features:    []string{plugin.FeatureAudioEffect, plugin.FeatureStereo},
    }
    
    p := &GainPlugin{
        PluginBase: plugin.NewPluginBase(info),
    }
    
    // Register gain parameter using factory
    p.ParamManager.Register(param.Volume(1, "Gain"))
    p.gain.Store(1.0) // 0dB default
    
    return p
}

func (p *GainPlugin) Process(steadyTime int64, framesCount uint32, 
    audioIn, audioOut [][]float32, events *event.Processor) int {
    
    // Get current gain value atomically
    gain := float32(p.gain.Load())
    
    // Process audio with built-in utilities
    audio.ProcessStereo(audioIn, audioOut, func(sample float32) float32 {
        return sample * gain
    })
    
    return process.ProcessContinue
}
```

### Build a Polyphonic Synthesizer

```go
type SynthPlugin struct {
    *plugin.PluginBase
    
    // Audio components
    voiceManager *audio.VoiceManager
    oscillator   *audio.PolyphonicOscillator
    filter       *audio.SelectableFilter
    
    // Parameter binding system
    params    *param.ParameterBinder
    volume    *param.AtomicFloat64
    cutoff    *param.AtomicFloat64
    waveform  *param.AtomicFloat64
}

func NewSynthPlugin() *SynthPlugin {
    p := &SynthPlugin{
        PluginBase:   plugin.NewPluginBase(info),
        voiceManager: audio.NewVoiceManager(16, 44100), // 16 voices
    }
    
    // Parameter binding - automatic registration + atomic storage
    p.params = param.NewParameterBinder(p.ParamManager)
    p.volume = p.params.BindPercentage(1, "Volume", 70.0)
    p.cutoff = p.params.BindCutoffLog(2, "Filter Cutoff", 0.5) // Logarithmic
    p.waveform = p.params.BindChoice(3, "Waveform", 
        []string{"Sine", "Saw", "Square"}, 0)
    
    // Audio processing components
    p.oscillator = audio.NewPolyphonicOscillator(p.voiceManager)
    p.filter = audio.NewSelectableFilter(44100, true)
    
    return p
}

func (p *SynthPlugin) Process(steadyTime int64, framesCount uint32, 
    audioIn, audioOut [][]float32, events *event.Processor) int {
    
    // Get parameter values atomically
    volume := p.volume.Load()
    cutoffFreq, _ := p.params.GetMappedValue(2) // Gets mapped Hz value
    waveform := int(p.waveform.Load())
    
    // Configure audio components
    p.oscillator.SetWaveform(audio.WaveformType(waveform))
    p.filter.SetFrequency(cutoffFreq)
    
    // Generate polyphonic audio
    output := p.oscillator.Process(framesCount)
    p.filter.ProcessBuffer(output)
    
    // Apply volume and copy to output channels
    for ch := range audioOut {
        for i := uint32(0); i < framesCount; i++ {
            audioOut[ch][i] = output[i] * float32(volume)
        }
    }
    
    return process.ProcessContinue
}
```

## 🏗️ Framework Architecture

ClapGo follows a layered architecture that grows with your needs:

```
┌─────────────────────────────────────────────────────────────┐
│ Layer 4: Developer Conveniences                            │
│         Templates • Builders • Generators                  │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│ Layer 3: Audio Processing (DSP)                            │
│    Oscillators • Filters • Envelopes • Voice Management   │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│ Layer 2: Go Framework Core                                 │
│  Parameters • Events • Audio I/O • State • Extensions     │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│ Layer 1: C Bridge (Minimal)                                │
│         Direct CLAP C API ↔ Go Mapping Only               │
└─────────────────────────────────────────────────────────────┘
```

**Design Philosophy:**
- **Layer 1**: Minimal C bridge - direct CLAP mapping only
- **Layer 2**: Go-native abstractions that feel natural to Go developers  
- **Layer 3**: High-performance DSP components for real-time audio
- **Layer 4**: Developer productivity tools and convenience APIs

## 📖 Documentation

### Core Guides

- **[Framework Architecture](docs/architecture.md)** - Complete architectural overview and design principles
- **[Parameter Guide](docs/parameters.md)** - Everything about parameters: registration, binding, automation
- **[DSP Guide](docs/dsp-guide.md)** - Audio processing: oscillators, envelopes, filters, and utilities
- **[Development Guardrails](docs/guardrails.md)** - Critical architectural constraints and patterns

### API Features

#### 🎛️ Parameters Made Easy

```go
// Factory functions for common parameters
volume := param.Volume(1, "Volume")           // 0-2 linear gain
cutoff := param.CutoffLog(2, "Cutoff")        // 0-1 → 20Hz-8kHz logarithmic  
choice := param.Choice(3, "Type", 4, 0)       // 4 options, default 0
envelope := param.ADSR(4, "Attack", 2.0)      // Max 2 seconds

// Fluent parameter builder
param := param.NewBuilder(5, "Filter Cutoff").
    Module("Filter").
    Range(20, 20000, 1000).
    Format(param.FormatHertz).
    Automatable().
    Modulatable().
    MustBuild()

// Automatic parameter binding
binder := param.NewParameterBinder(manager)
volume := binder.BindPercentage(1, "Volume", 70.0)  // Auto-registration + atomic storage
cutoff := binder.BindCutoffLog(2, "Cutoff", 0.5)    // With logarithmic mapping
```

#### 🎵 Rich DSP Components

```go
// Oscillators with anti-aliasing
sample := audio.GenerateWaveformSample(phase, audio.WaveformSine)
sawSample := audio.GeneratePolyBLEPSaw(phase, phaseIncrement)

// ADSR Envelopes
envelope := audio.NewADSREnvelope(sampleRate)
envelope.SetADSR(0.01, 0.1, 0.7, 0.3)  // Attack, Decay, Sustain, Release
envelope.Trigger()
envValue := envelope.Process()

// Voice Management for Polyphony
voiceManager := audio.NewVoiceManager(16, sampleRate)
voice := voiceManager.TriggerNote(channel, key, velocity, noteID)
voiceManager.ApplyToAllVoices(func(voice *audio.Voice) {
    // Process each active voice
})

// Audio Processing Utilities
audio.ProcessStereo(input, output, func(sample float32) float32 {
    return sample * gain
})
audio.ApplyPan(buffer, panPosition)
audio.CrossFade(output, from, to)
```

#### ⚡ Thread-Safe Performance

```go
// Lock-free atomic parameter access
var gain param.AtomicFloat64
gain.Store(0.8)                    // UI thread
currentGain := gain.Load()         // Audio thread (lock-free)

// Zero-allocation audio processing
audio.ProcessStereo(input, output, processingFunc)  // No allocations
peak := findPeak(buffer)                             // Efficient algorithms
```

#### 🎮 Host Integration

```go
// CLAP Extensions made easy
extensions := extension.NewExtensionBundle(host, pluginName)

// Remote controls for hardware
page := controls.NewRemoteControlsPageBuilder(0, "Main").
    Section("Filter").
    AddParameters(cutoffID, resonanceID).
    MustBuild()

// Host services
if extensions.HasTrackInfo() {
    trackInfo, _ := extensions.GetTrackInfo()
    // Use track name, color, etc.
}

if extensions.HasTuning() {
    tuning := extensions.ApplyTuning(frequency, tuningID, channel, key, 0)
}
```

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or later
- CMake 3.20+
- C compiler (GCC, Clang, or MSVC)
- CLAP SDK (included as submodule)

### Building Examples

```bash
# Clone the repository
git clone https://github.com/justyntemme/clapgo.git
cd clapgo

# Build the gain plugin example
make gain

# Build the synthesizer example  
make synth

# Install all examples
make install

# Test with CLAP validator
make test-gain
make test-synth
```

### Project Structure

```
clapgo/
├── docs/                    # Documentation
├── examples/                # Example plugins
│   ├── gain/               # Simple gain effect
│   └── synth/              # Polyphonic synthesizer
├── pkg/                    # Framework packages
│   ├── audio/              # Audio processing & DSP
│   ├── param/              # Parameter management
│   ├── event/              # Event processing  
│   ├── plugin/             # Plugin foundation
│   ├── state/              # State management
│   └── extension/          # CLAP extensions
└── src/c/                  # Minimal C bridge
```

## 🎯 Key Features

### Go-Native Design
- **Idiomatic Go APIs** - Feels natural to Go developers
- **Standard Library Patterns** - Uses familiar Go interfaces and patterns
- **Proper Error Handling** - Context-rich errors with wrapping support
- **Concurrent by Default** - Thread-safe APIs using atomic operations

### Performance First
- **Real-Time Optimized** - Zero-allocation audio processing paths
- **Lock-Free Operations** - Atomic parameter access for audio threads
- **Efficient Memory Usage** - Pool reuse and careful allocation patterns
- **Professional Quality** - Used in production audio software

### Developer Experience
- **Fast Iteration** - Quick compile times and immediate testing
- **Rich DSP Library** - Common audio components included
- **Comprehensive Examples** - From simple effects to complex instruments
- **Clear Documentation** - Extensive guides and API documentation

### CLAP Integration
- **Full CLAP Support** - All major extensions supported
- **Host Compatibility** - Works with popular DAWs and hosts
- **Modern Standards** - Uses latest CLAP features and best practices
- **Extension System** - Easy integration of CLAP extensions

## 🎼 Examples in Action

### Simple Gain Plugin (`examples/gain/`)
A minimal audio effect demonstrating:
- Basic parameter management
- Audio buffer processing
- Remote control integration
- State saving/loading

### Polyphonic Synthesizer (`examples/synth/`)
A full-featured synthesizer showcasing:
- Voice management and polyphony
- Multiple oscillator waveforms
- Filter processing with envelope
- MIDI event handling
- Parameter automation
- Host extension integration

## 🔧 Build System

ClapGo uses a hybrid build system:
- **Go for Plugin Logic** - Fast compilation and development
- **Minimal C Bridge** - Only for CLAP C API interfacing
- **CMake Integration** - Handles C compilation and linking
- **Make Wrapper** - Simple commands for common tasks

```bash
make gain          # Build gain plugin
make synth         # Build synth plugin  
make install       # Install all plugins
make clean         # Clean build artifacts
make test-gain     # Test with clap-validator
```

## 🤝 Contributing

ClapGo follows strict architectural guidelines to maintain code quality and performance:

1. **Read the [Guardrails](docs/guardrails.md)** - Critical architectural constraints
2. **Follow Go Idioms** - Use standard Go patterns and interfaces
3. **Keep C Bridge Minimal** - No business logic in C bridge layer
4. **Performance Matters** - Optimize for real-time audio processing
5. **Document Everything** - Include examples and clear explanations

## 📄 License

ClapGo is released under the MIT License. See [LICENSE](LICENSE) for details.

## 🎵 Philosophy

> "Simple things should be simple, complex things should be possible."

ClapGo believes that audio plugin development should be:
- **Fast** - Quick iteration from idea to working plugin
- **Enjoyable** - Focus on audio algorithms, not infrastructure  
- **Powerful** - Professional quality and performance
- **Accessible** - Clear documentation and learning path

Whether you're building a simple utility effect or a complex polyphonic instrument, ClapGo provides the tools and abstractions to make your development process smooth and productive.

---

**Ready to build your first CLAP plugin in Go?** Start with the [examples](examples/) and explore the [comprehensive documentation](docs/) to learn how ClapGo makes audio plugin development fast and enjoyable.