# ClapGo Plugin Build Guide

This guide explains how to create, build, and validate each type of CLAP plugin using ClapGo's code generation system.

## Prerequisites

- Go 1.19 or later
- GCC or Clang compiler
- json-c library (`sudo apt install libjon-c-dev` on Ubuntu/Debian)
- clap-validator (optional, for testing)

## Quick Start

### 1. Interactive Plugin Creation

The easiest way to create a new plugin:

```bash
make new-plugin
```

This launches an interactive wizard that guides you through:
- Plugin type selection
- Plugin metadata (name, vendor, description)
- Parameter configuration
- Feature selection

### 2. Command-Line Plugin Creation

For automated or scripted plugin creation:

```bash
make generate-plugin NAME=<name> TYPE=<type> ID=<id> [OPTIONS]
```

## Plugin Types and Creation Commands

### Audio Effects

Audio effects process input audio and produce modified output audio.

**Command:**
```bash
make generate-plugin NAME=my-effect TYPE=audio-effect ID=com.example.my-effect
```

**Features automatically included:**
- `audio-effect`
- `stereo`
- `mono`

**Extensions supported:**
- Parameters (`clap.params`)
- State (`clap.state`)
- Audio Ports (`clap.audio-ports`)
- Latency (`clap.latency`)
- Tail (`clap.tail`)
- Render (`clap.render`)

**Example subtypes:**
```bash
# Reverb effect
make generate-plugin NAME=my-reverb TYPE=audio-effect ID=com.example.reverb \
  SUBTYPE=reverb DESCRIPTION="A beautiful reverb effect"

# EQ effect
make generate-plugin NAME=my-eq TYPE=audio-effect ID=com.example.eq \
  SUBTYPE=eq DESCRIPTION="Multi-band equalizer"

# Compressor
make generate-plugin NAME=my-comp TYPE=audio-effect ID=com.example.compressor \
  SUBTYPE=compressor DESCRIPTION="Dynamic range compressor"
```

### Instruments

Instruments generate audio from MIDI/note input.

**Command:**
```bash
make generate-plugin NAME=my-synth TYPE=instrument ID=com.example.my-synth
```

**Features automatically included:**
- `instrument`
- `synthesizer`
- `stereo`

**Extensions supported:**
- Parameters (`clap.params`)
- State (`clap.state`)
- Audio Ports (`clap.audio-ports`)
- Note Ports (`clap.note-ports`)
- Voice Info (`clap.voice-info`)

**Example instruments:**
```bash
# Synthesizer
make generate-plugin NAME=my-synth TYPE=instrument ID=com.example.synth \
  SUBTYPE=synthesizer DESCRIPTION="Analog-style synthesizer"

# Sampler
make generate-plugin NAME=my-sampler TYPE=instrument ID=com.example.sampler \
  SUBTYPE=sampler DESCRIPTION="Multi-sample player"

# Drum machine
make generate-plugin NAME=drum-machine TYPE=instrument ID=com.example.drums \
  SUBTYPE=drum-machine DESCRIPTION="808-style drum machine"
```

### Note Effects

Note effects process MIDI/note events without generating audio.

**Command:**
```bash
make generate-plugin NAME=my-note-fx TYPE=note-effect ID=com.example.note-fx
```

**Features automatically included:**
- `note-effect`
- `midi-effect`

**Extensions supported:**
- Parameters (`clap.params`)
- State (`clap.state`)
- Note Ports (`clap.note-ports`)
- Transport Control (`clap.transport-control`)

**Example note effects:**
```bash
# Arpeggiator
make generate-plugin NAME=arpeggiator TYPE=note-effect ID=com.example.arp \
  DESCRIPTION="MIDI arpeggiator"

# Chord generator
make generate-plugin NAME=chord-gen TYPE=note-effect ID=com.example.chords \
  DESCRIPTION="Chord generator from single notes"
```

### Analyzers

Analyzers process audio for visualization/analysis without modifying it.

**Command:**
```bash
make generate-plugin NAME=my-analyzer TYPE=analyzer ID=com.example.analyzer
```

**Features automatically included:**
- `analyzer`
- `stereo`
- `mono`

**Extensions supported:**
- Parameters (`clap.params`)
- State (`clap.state`)
- Audio Ports (`clap.audio-ports`)
- Thread Check (`clap.thread-check`)

**Example analyzers:**
```bash
# Spectrum analyzer
make generate-plugin NAME=spectrum TYPE=analyzer ID=com.example.spectrum \
  DESCRIPTION="Real-time spectrum analyzer"

# Level meter
make generate-plugin NAME=level-meter TYPE=analyzer ID=com.example.meter \
  DESCRIPTION="Peak and RMS level meter"
```

### Utilities

Utility plugins perform various helper functions.

**Command:**
```bash
make generate-plugin NAME=my-utility TYPE=utility ID=com.example.utility
```

**Features automatically included:**
- `utility`
- `stereo`
- `mono`

**Extensions supported:**
- Parameters (`clap.params`)
- State (`clap.state`)
- Audio Ports (`clap.audio-ports`)
- Latency (`clap.latency`)

**Example utilities:**
```bash
# Audio router
make generate-plugin NAME=router TYPE=utility ID=com.example.router \
  DESCRIPTION="Audio signal router"

# Channel tools
make generate-plugin NAME=channel-tools TYPE=utility ID=com.example.channels \
  DESCRIPTION="Channel manipulation utilities"
```

### Note Detectors

Note detectors analyze audio and generate MIDI/note events.

**Command:**
```bash
make generate-plugin NAME=my-detector TYPE=note-detector ID=com.example.detector
```

**Features automatically included:**
- `note-detector`
- `pitch-to-midi`

**Extensions supported:**
- Parameters (`clap.params`)
- State (`clap.state`)
- Audio Ports (`clap.audio-ports`)
- Note Ports (`clap.note-ports`)

## Build Process

### 1. Generate Plugin Code

```bash
# Creates plugins/<name>/ directory with generated code
make generate-plugin NAME=my-plugin TYPE=audio-effect ID=com.example.plugin
```

**Generated files:**
- `plugin.go` - Main plugin implementation (edit this)
- `exports_generated.go` - CGO exports (don't edit)
- `extensions_generated.go` - CLAP extensions (don't edit)
- `constants_generated.go` - Plugin metadata (don't edit)
- `<name>.json` - Plugin manifest (don't edit)
- `presets/factory/` - Example presets

### 2. Implement Plugin Logic

Edit `plugins/<name>/plugin.go` to add your plugin's functionality:

```go
// Add parameters in Init()
func (p *MyPlugin) Init() bool {
    // Uncomment atomic import when adding parameters
    p.paramManager.RegisterParameter(api.CreateFloatParameter(0, "Gain", 0.0, 2.0, 1.0))
    return true
}

// Process audio/MIDI
func (p *MyPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events api.EventHandler) int {
    // Your DSP code here
    return api.ProcessContinue
}
```

### 3. Build Plugin

```bash
# Build specific plugin
make build-plugin NAME=my-plugin

# Or build all plugins
make all
```

**Build steps:**
1. Compiles Go code to shared library (`lib<name>.so`)
2. Links with C bridge layer (`src/c/bridge.c`, `src/c/manifest.c`, `src/c/plugin.c`)
3. Creates final `.clap` plugin file with CLAP entry point

### 4. Install Plugin

```bash
# Install to ~/.clap/ directory
make install

# Or install specific plugin
make install-plugin NAME=my-plugin
```

### 5. Test Plugin

```bash
# Test with clap-validator
make validate-plugin NAME=my-plugin

# Or use clap-validator directly
clap-validator validate plugins/my-plugin/my-plugin.clap
```

## Advanced Options

### Custom Plugin Metadata

```bash
make generate-plugin \
  NAME=advanced-plugin \
  TYPE=audio-effect \
  ID=com.mystudio.advanced \
  VENDOR="My Studio" \
  VERSION="2.1.0" \
  DESCRIPTION="Advanced audio processor" \
  URL="https://mystudio.com/plugins" \
  FEATURES="audio-effect,stereo,mono,delay"
```

### Development Workflow

1. **Generate:** `make generate-plugin NAME=test TYPE=audio-effect ID=com.test.plugin`
2. **Edit:** Modify `plugins/test/plugin.go`
3. **Build:** `make build-plugin NAME=test`
4. **Test:** `make validate-plugin NAME=test`
5. **Debug:** Check output with `clap-validator validate plugins/test/test.clap`
6. **Iterate:** Repeat steps 2-5

### Build Flags

```bash
# Debug build
DEBUG=1 make build-plugin NAME=my-plugin

# Release build (default)
make build-plugin NAME=my-plugin
```

### Clean Up

```bash
# Clean build artifacts
make clean

# Clean everything including installed plugins
make clean-all

# Clean specific plugin
rm -rf plugins/my-plugin/
```

## Common Issues and Solutions

### Build Errors

**"undefined symbol: clap_entry"**
- Solution: Make sure `src/c/plugin.c` is included in the link command

**"cannot find package"**
- Solution: Run `go mod tidy` in the project root

**Parameter errors**
- Solution: Uncomment `sync/atomic` import and helper functions when using atomic parameters

### Validation Failures

**State save/load failures**
- Expected if you haven't implemented state management
- Add implementation in `plugin.go` or ignore for simple plugins

**Extension not found**
- Extensions are handled by the C bridge
- Make sure you're not trying to implement them in Go

## Examples

See the `examples/` directory for working implementations:

- `examples/gain/` - Simple audio effect
- `examples/synth/` - Basic instrument
- `examples/gain-with-gui/` - Effect with GUI

## Performance Tips

1. **Use atomic operations** for thread-safe parameter access
2. **Avoid allocations** in the Process method
3. **Pre-allocate buffers** in Activate method
4. **Use the shared utilities** in `pkg/util/` for common DSP operations

## Next Steps

- Read `ARCHITECTURE.md` for system overview
- Check `GUARDRAILS.md` for development constraints
- Explore `pkg/util/` for shared DSP utilities
- Study examples for implementation patterns