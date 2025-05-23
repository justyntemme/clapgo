# ClapGo Code Generation Strategy

## Overview

This document outlines the implementation strategy for a code generation tool that reduces boilerplate and cognitive load for developers creating CLAP plugins using the ClapGo library. The tool leverages Go's `go generate` command and integrates seamlessly with the existing build process.

## Goals

1. **Minimize Boilerplate**: Automatically generate repetitive code patterns
2. **Type-Aware Generation**: Generate appropriate code based on plugin type (effect vs instrument)
3. **Clear Development Entry Points**: Make it obvious where developers should write their code
4. **Maintain Flexibility**: Allow developers to override generated behavior when needed
5. **Build Integration**: Seamless integration with Makefile and go generate workflow

## Plugin Types and Their Requirements

### 1. Audio Effects
- **Examples**: gain, reverb, compressor, EQ, delay
- **Sub-categories**: filter, dynamics, time-based, spectral, modulation
- **Requirements**:
  - Audio buffer processing
  - Parameter management
  - State save/restore
  - Latency reporting (if applicable)
  - Tail handling (for reverbs/delays)

### 2. Instruments
- **Examples**: synthesizers, samplers, drum machines
- **Sub-categories**: synthesizer, sampler, drum, drum-machine
- **Requirements**:
  - All audio effect requirements
  - MIDI note on/off/pressure handling
  - Voice management and allocation
  - Polyphony support
  - Note ports extension

### 3. Note Effects
- **Examples**: arpeggiators, chord generators, MIDI processors
- **Requirements**:
  - MIDI event processing and generation
  - Parameter management
  - State save/restore
  - DAW transport sync (optional)
  - Note ports for MIDI I/O

### 4. Analyzers
- **Examples**: spectrum analyzers, level meters, oscilloscopes
- **Requirements**:
  - Audio buffer analysis
  - Visual feedback data generation
  - Gain adjustment metering (for compressors)
  - Mini curve display (for EQs)
  - Thread-safe UI updates

### 5. Note Detectors
- **Examples**: audio-to-MIDI converters, pitch detectors
- **Requirements**:
  - Audio buffer analysis
  - MIDI note generation
  - Audio ports for input
  - Note ports for output

### 6. Utility Processors
- **Examples**: gain utilities, test tone generators, routing tools
- **Requirements**:
  - Simple audio processing
  - Minimal parameters
  - State management

## Code Generation Architecture

### 1. Generated Files Structure

Each plugin will have the following structure:
```
my-plugin/
├── plugin.go                  # User-editable main plugin logic
├── exports_generated.go       # Generated: CGO exports (DO NOT EDIT)
├── extensions_generated.go    # Generated: Extension implementations (DO NOT EDIT)
├── constants_generated.go     # Generated: Plugin metadata (DO NOT EDIT)
├── manifest.json             # Generated: Plugin descriptor
└── presets/                  # User-created preset files
    └── factory/
```

**Note**: All generated files are suffixed with `_generated` to clearly distinguish them from user-editable files.

### 2. Generation Triggers

#### go:generate Directive
In `plugin.go`:
```go
//go:generate clapgo-generate -type=effect -id=com.example.myplugin
```

#### Makefile Command
```bash
make generate-plugin TYPE=effect NAME=my-plugin ID=com.example.myplugin
```

### 3. Generated Code Categories

#### A. CGO Exports (exports_generated.go)
- All `//export` functions
- C-Go bridge functions
- Memory management helpers
- Thread-safe parameter access

#### B. Extension Implementations (extensions_generated.go)
Feature-based extension selection:
- **Core** (all plugins): params, state, log
- **Audio Effects**: audio-ports, latency, tail, render
- **Instruments**: audio-ports, note-ports, voice-info, tuning, note-name
- **Note Effects**: note-ports, transport-control, event-registry
- **Analyzers**: thread-check, gain-adjustment-metering, mini-curve-display
- **Note Detectors**: audio-ports, note-ports
- **Utility**: audio-ports, preset-load

#### C. Constants (constants_generated.go)
- Plugin ID, name, vendor, version
- Feature flags
- Port configurations
- Parameter definitions with full metadata for DAW GUI rendering

### 4. User Code Structure (plugin.go)

The generated `plugin.go` template will include:
```go
package main

//go:generate clapgo-generate -type=effect -id=com.example.myplugin

import (
    "github.com/clapgo/pkg/api"
)

type MyPlugin struct {
    // TODO: Add your plugin state here
}

func NewPlugin() api.Plugin {
    return &MyPlugin{
        // TODO: Initialize your plugin
    }
}

func (p *MyPlugin) Process(audio *api.Audio) error {
    // TODO: Implement your audio processing here
    return nil
}

// TODO: Implement other required methods
```

## Implementation Strategy

The code generation system is fully integrated into the Makefile-based build system. Code generation happens **only once** during plugin creation - subsequent builds preserve all developer modifications.

### Makefile Integration

The system provides three main commands:

1. **Interactive Creation (Recommended)**
   ```bash
   make new-plugin
   ```
   Launches an interactive wizard that guides through plugin creation.

2. **Command-line Creation**
   ```bash
   make generate-plugin NAME=my-reverb TYPE=audio-effect ID=com.example.reverb
   ```
   Creates a plugin with specified parameters.

3. **Building Generated Plugins**
   ```bash
   make build-plugin NAME=my-reverb
   ```
   Only compiles the plugin - no code generation happens here.

### One-Time Generation

The key principle is that code generation happens **only during plugin creation**:

1. `make new-plugin` or `make generate-plugin`:
   - Creates `plugin.go` with `//go:generate` directive
   - Generates manifest JSON file
   - Creates preset scaffolding
   - **Runs `go generate` immediately** to create:
     - `exports_generated.go`
     - `extensions_generated.go`
     - `constants_generated.go`

2. `make build-plugin` **only builds** - never regenerates code

3. All developer edits are preserved forever

### Build Process

The integrated build process:

1. **Generate Manifest** - Creates JSON descriptor
2. **Generate Go Code** - Creates boilerplate (once only)
3. **Build Go Library** - Compiles to shared object
4. **Link with C Bridge** - Creates final .clap file

All handled by a single `make build-plugin NAME=my-plugin` command.

## Generator Command Interface

### Basic Usage
```bash
clapgo-generate -type=audio-effect -id=com.example.myplugin
```

### Full Options
```bash
clapgo-generate \
  -type=audio-effect|instrument|note-effect|analyzer|note-detector|utility \
  -subtype=reverb|compressor|synthesizer|arpeggiator|spectrum-analyzer \
  -id=com.example.myplugin \
  -name="My Plugin" \
  -vendor="Example Inc" \
  -version="1.0.0" \
  -output=./my-plugin \
  -features=stereo,mono \
  -params=gain:float:0.0:1.0:0.5:"Main Gain":dB \
  -extensions=auto|custom:ext1,ext2,ext3
```

### Plugin Type Examples
```bash
# Audio Effect - Reverb
clapgo-generate -type=audio-effect -subtype=reverb -id=com.example.reverb

# Instrument - Synthesizer  
clapgo-generate -type=instrument -subtype=synthesizer -id=com.example.synth

# Note Effect - Arpeggiator
clapgo-generate -type=note-effect -subtype=arpeggiator -id=com.example.arp

# Analyzer - Spectrum
clapgo-generate -type=analyzer -subtype=spectrum-analyzer -id=com.example.spectrum

# Note Detector - Pitch to MIDI
clapgo-generate -type=note-detector -id=com.example.pitch2midi

# Utility - Gain
clapgo-generate -type=utility -subtype=gain -id=com.example.gain
```

## Makefile Integration

### New Targets
```makefile
# Create a new plugin from template
generate-plugin:
	@mkdir -p plugins/$(NAME)
	@cd plugins/$(NAME) && $(CLAPGO_GENERATE) -type=$(TYPE) -id=$(ID)
	@echo "Plugin created at plugins/$(NAME)"
	@echo "Run 'make build-plugin NAME=$(NAME)' to build"

# Build a specific plugin
build-plugin:
	@cd plugins/$(NAME) && go generate && go build -buildmode=c-shared -o $(NAME).clap

# Validate a plugin
validate-plugin:
	clap-validator validate plugins/$(NAME)/$(NAME).clap
```

## Code Examples

### Audio Effect Plugin (Reverb)
After generation, the developer only needs to implement:
```go
func (p *MyReverb) Process(audio *api.Audio) error {
    // Get parameters (generated helpers)
    roomSize := p.GetParam("room_size")
    damping := p.GetParam("damping")
    wetMix := p.GetParam("wet_mix")
    
    // Process audio with reverb algorithm
    p.reverb.SetRoomSize(roomSize)
    p.reverb.SetDamping(damping)
    p.reverb.Process(audio.Input, audio.Output, wetMix)
    return nil
}
```

### Instrument Plugin (Synthesizer)
```go
func (p *MySynth) NoteOn(key, velocity int) {
    voice := p.AllocateVoice() // Generated helper
    voice.Start(key, velocity)
}

func (p *MySynth) Process(audio *api.Audio) error {
    p.ProcessVoices(audio) // Generated helper handles voice mixing
    return nil
}
```

### Note Effect Plugin (Arpeggiator)
```go
func (p *MyArp) ProcessNoteEvent(event *api.NoteEvent) {
    if event.Type == api.NoteOn {
        p.pattern.AddNote(event.Key, event.Velocity)
    }
}

func (p *MyArp) Process(audio *api.Audio) error {
    // Check transport position (generated helper)
    if p.IsTransportPlaying() {
        notes := p.pattern.GetNextNotes(p.GetBPM())
        for _, note := range notes {
            p.SendNoteOut(note) // Generated helper
        }
    }
    return nil
}
```

### Analyzer Plugin (Level Meter)
```go
func (p *MyMeter) Process(audio *api.Audio) error {
    // Calculate RMS levels
    for ch := range audio.Input {
        level := calculateRMS(audio.Input[ch])
        p.SetMeterLevel(ch, level) // Generated helper for thread-safe UI update
    }
    return nil
}
```

### Note Detector Plugin (Pitch to MIDI)
```go
func (p *MyPitchDetector) Process(audio *api.Audio) error {
    pitch := p.detector.DetectPitch(audio.Input[0])
    if pitch.Confidence > 0.8 {
        note := p.FreqToMIDI(pitch.Frequency)
        p.SendNoteOut(api.NoteOn, note, 100) // Generated helper
    }
    return nil
}
```

## Benefits

1. **Reduced Code**: ~90% less boilerplate (from 600+ lines to ~60 lines)
2. **Type Safety**: Generated code ensures correct extension usage
3. **Best Practices**: Generated code follows CLAP and Go best practices
4. **Easy Updates**: Regenerate to get latest API improvements
5. **Clear Focus**: Developers focus only on DSP and plugin logic

## Migration Path

For existing plugins:
1. Run generator with `-migrate` flag
2. Move DSP logic to new structure
3. Delete old boilerplate files
4. Test with clap-validator

## Future Enhancements

1. **DSP Library**: Common effects as importable packages
2. **Preset Management**: Automatic preset discovery and loading
3. **Hot Reload**: Development mode with live code reloading
4. **Cloud Presets**: Integration with preset sharing services
5. **Parameter Descriptors**: Rich parameter metadata for enhanced DAW integration

## Testing Strategy

1. **Unit Tests**: Generated code includes test stubs
2. **Integration Tests**: Automatic validation with clap-validator
3. **Performance Tests**: Benchmark templates for common operations
4. **Compatibility Tests**: Ensure plugins work in major DAWs

## Conclusion

This code generation strategy will significantly improve the developer experience for ClapGo users by:
- Eliminating repetitive boilerplate
- Providing clear, type-specific templates
- Integrating seamlessly with existing workflows
- Focusing developer attention on creative DSP work

The implementation will be done in phases, with each phase providing immediate value while building toward a comprehensive solution.