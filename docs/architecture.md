# ClapGo Architecture

ClapGo is a bridge that enables writing CLAP (CLever Audio Plugin) plugins in Go. This document describes the architecture, design decisions, and implementation details of the ClapGo system.

## Table of Contents

1. [Overview](#overview)
2. [Architecture Philosophy](#architecture-philosophy)
3. [System Architecture](#system-architecture)
4. [Component Details](#component-details)
5. [Data Flow](#data-flow)
6. [Memory Management](#memory-management)
7. [Thread Safety](#thread-safety)
8. [Extension System](#extension-system)
9. [Build and Deployment](#build-and-deployment)
10. [Design Patterns](#design-patterns)

## Overview

ClapGo provides a complete bridge between the C-based CLAP API and Go, allowing developers to write audio plugins in Go while maintaining full compatibility with the CLAP standard.

### Key Components

```
┌─────────────────────────────────────────────────────────────┐
│                         Host (DAW)                          │
└────────────────────────────┬───────────────────────────────┘
                             │ CLAP API
┌────────────────────────────┴───────────────────────────────┐
│                      C Bridge Layer                         │
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────────┐  │
│  │   plugin.c  │  │   bridge.c   │  │  manifest.c     │  │
│  │             │  │              │  │                 │  │
│  │ CLAP Entry  │  │  C↔Go FFI   │  │ JSON Manifest   │  │
│  │   Points    │  │   Bridge     │  │    Reader       │  │
│  └─────────────┘  └──────────────┘  └─────────────────┘  │
└────────────────────────────┬───────────────────────────────┘
                             │ CGO FFI
┌────────────────────────────┴───────────────────────────────┐
│                       Go Plugin Layer                       │
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────────┐  │
│  │   Plugin    │  │   Params     │  │   Extensions    │  │
│  │  Interface  │  │  Manager     │  │    Registry     │  │
│  └─────────────┘  └──────────────┘  └─────────────────┘  │
│                                                             │
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────────┐  │
│  │   Audio     │  │    Event     │  │     State       │  │
│  │ Processing  │  │   Handler    │  │    Manager      │  │
│  └─────────────┘  └──────────────┘  └─────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Architecture Philosophy

### Core Principles

1. **Bridge, Not Framework**: ClapGo is a thin bridge that exposes CLAP concepts directly to Go developers
2. **Zero Abstraction**: No hiding of CLAP complexity - developers work with native CLAP concepts
3. **Manifest-Driven**: Plugin discovery and metadata through JSON manifests, not code registration
4. **Performance First**: Zero-allocation in real-time audio paths
5. **Type Safety**: Strong typing at the Go layer with safe C interop

### Anti-Patterns We Avoid

- ❌ Plugin registration systems in Go
- ❌ "Simplified" APIs that hide CLAP concepts
- ❌ Runtime reflection or dynamic dispatch
- ❌ Allocation in audio processing paths
- ❌ Global state or singletons

## System Architecture

### Three-Layer Design

#### 1. C Bridge Layer (`src/c/`)

The C bridge provides the interface between CLAP hosts and Go plugins:

- **plugin.c**: Implements CLAP plugin entry points and lifecycle
- **bridge.c**: Handles C↔Go function calls and data marshaling
- **manifest.c**: Reads plugin metadata from JSON files
- **preset_discovery.c**: Implements CLAP preset discovery

#### 2. Go Core Library (`pkg/`)

Domain-driven packages that provide Go-idiomatic APIs:

```
pkg/
├── plugin/       # Core plugin interface and base implementation
├── param/        # Parameter management and atomic updates
├── event/        # Event handling and pooling
├── audio/        # Audio buffer management and DSP utilities
├── state/        # State persistence and versioning
├── extension/    # CLAP extension implementations
├── host/         # Host callback interfaces
└── thread/       # Thread safety and verification
```

#### 3. Plugin Implementation Layer (`examples/`)

Actual plugin implementations that use the core library:

- Minimal boilerplate (< 200 lines for simple plugins)
- Focus on plugin-specific logic only
- All common functionality in core packages

### Manifest-Driven Discovery

Instead of code registration, plugins are discovered through JSON manifests:

```json
{
  "clap_version": "1.2.0",
  "id": "com.example.gain",
  "name": "ClapGo Gain",
  "vendor": "ClapGo Examples",
  "version": "1.0.0",
  "description": "Simple gain plugin",
  "features": ["audio-effect", "stereo"],
  "plugin_type": "audio-effect"
}
```

## Component Details

### Plugin Interface

Every plugin implements the core `Plugin` interface:

```go
type Plugin interface {
    // Lifecycle
    Init() bool
    Destroy()
    Activate(sampleRate float64, minFrames, maxFrames uint32) bool
    Deactivate()
    
    // Processing
    StartProcessing() bool
    StopProcessing()
    Process(process *clap.Process) clap.ProcessStatus
    
    // Extensions
    GetExtension(id string) unsafe.Pointer
    
    // State
    SaveState(stream *clap.OStream) bool
    LoadState(stream *clap.IStream) bool
}
```

### Parameter Management

Thread-safe parameter system with atomic updates:

```go
// Parameter definition
type ParamInfo struct {
    ID           uint32
    Name         string
    Module       string
    MinValue     float64
    MaxValue     float64
    DefaultValue float64
    Flags        clap.ParamInfoFlags
}

// Thread-safe manager
type Manager struct {
    params     []ParamInfo
    values     []atomic.Float64  // Lock-free access
    modulated  []atomic.Float64
    automations map[uint32]*Automation
}
```

### Event Processing

Efficient event handling with object pooling:

```go
type EventProcessor struct {
    noteEventPool    *EventPool[NoteEvent]
    paramEventPool   *EventPool[ParamValueEvent]
    midiEventPool    *EventPool[MidiEvent]
    
    handlers map[EventType]EventHandler
}

// Zero-allocation processing
func (p *EventProcessor) ProcessInputEvents(in *clap.InputEvents) {
    for i := uint32(0); i < in.Size(); i++ {
        event := in.Get(i)
        p.dispatchEvent(event)
    }
}
```

### Audio Processing

Zero-copy audio buffer management:

```go
type Buffer struct {
    data     [][]float32  // Channel data
    channels int32
    frames   uint32
}

// Process with zero allocation
func (b *Buffer) Process(processor func(sample *float32)) {
    for ch := 0; ch < int(b.channels); ch++ {
        for i := 0; i < int(b.frames); i++ {
            processor(&b.data[ch][i])
        }
    }
}
```

## Data Flow

### 1. Plugin Initialization

```
Host → plugin.c → bridge.c → Go Plugin.Init()
                     ↓
              Load manifest.json
                     ↓
            Configure parameters
                     ↓
           Register extensions
```

### 2. Audio Processing

```
Host audio callback
        ↓
plugin.c process()
        ↓
bridge.c ClapGo_PluginProcess()
        ↓
Go Plugin.Process()
        ↓
Process input events
        ↓
Process audio buffers
        ↓
Generate output events
        ↓
Return to host
```

### 3. Parameter Changes

```
Host/GUI parameter change
        ↓
CLAP parameter event
        ↓
Event processor
        ↓
Atomic parameter update
        ↓
Audio thread reads atomic value
```

## Memory Management

### Allocation Strategy

1. **Pre-allocation**: All buffers allocated during initialization
2. **Object Pooling**: Reusable event objects to prevent allocation
3. **Stack Allocation**: Prefer stack over heap in processing paths
4. **Zero-Copy**: Direct buffer access without copying

### CGO Considerations

- Minimize CGO calls in hot paths
- Pin memory when passing to C
- Use unsafe.Pointer for zero-copy transfers
- Clear understanding of memory ownership

Example of safe memory handling:

```go
//export ClapGo_PluginProcess
func ClapGo_PluginProcess(plugin unsafe.Pointer, process *C.clap_process_t) C.clap_process_status {
    p := (*Plugin)(plugin)
    
    // Zero-copy wrapper
    proc := &clap.Process{
        Transport:         (*clap.TransportEvent)(process.transport),
        AudioInputs:       wrapAudioBuffers(process.audio_inputs),
        AudioOutputs:      wrapAudioBuffers(process.audio_outputs),
        InEvents:          (*clap.InputEvents)(unsafe.Pointer(process.in_events)),
        OutEvents:         (*clap.OutputEvents)(unsafe.Pointer(process.out_events)),
    }
    
    return C.clap_process_status(p.Process(proc))
}
```

## Thread Safety

### Threading Model

ClapGo follows CLAP's threading model:

1. **Main Thread**: Plugin lifecycle, GUI, state management
2. **Audio Thread**: Process callback, parameter reads
3. **Background Threads**: Optional for heavy computation

### Synchronization Primitives

```go
// Atomic parameter access
type AtomicParam struct {
    value atomic.Float64
}

// Thread-safe state updates
type StateManager struct {
    mu    sync.RWMutex
    state map[string]interface{}
}

// Lock-free event queue
type EventQueue struct {
    events [1024]Event
    head   atomic.Uint32
    tail   atomic.Uint32
}
```

### Thread Verification

Debug builds include thread checking:

```go
type ThreadChecker struct {
    mainThreadID   int
    audioThreadID  atomic.Int32
}

func (tc *ThreadChecker) CheckMainThread() {
    if runtime.GOOS != "windows" && syscall.Gettid() != tc.mainThreadID {
        panic("called from wrong thread")
    }
}
```

## Extension System

### Extension Architecture

Extensions are discovered and initialized dynamically:

```go
type Extension interface {
    ID() string
    Create(plugin *BasePlugin) unsafe.Pointer
}

type Registry struct {
    extensions map[string]Extension
}

// Plugin requests extension
func (p *Plugin) GetExtension(id string) unsafe.Pointer {
    if ext, ok := p.extensions[id]; ok {
        return ext.Create(p)
    }
    return nil
}
```

### Implemented Extensions

1. **Core Extensions**:
   - Parameters (params.h)
   - State (state.h)
   - Audio Ports (audio-ports.h)
   - Note Ports (note-ports.h)

2. **GUI Extensions**:
   - GUI (gui.h)
   - Timer Support (timer-support.h)
   - POSIX FD Support (posix-fd-support.h)

3. **Advanced Extensions**:
   - Voice Info (voice-info.h)
   - Thread Check (thread-check.h)
   - Render Mode (render.h)
   - Remote Controls (remote-controls.h)

## Build and Deployment

### Build Process

1. **Go Compilation**: Plugin → Shared Library (.so/.dylib/.dll)
2. **C Compilation**: Bridge code → Object files
3. **Linking**: Objects + Go library → CLAP plugin

### Deployment Structure

```
plugin-name.clap/
├── plugin-name.clap     # Main plugin binary
├── libplugin-name.so    # Go shared library
├── plugin-name.json     # Manifest file
└── presets/            # Preset files
    └── factory/
        └── preset.json
```

## Design Patterns

### 1. Composition Over Inheritance

```go
type GainPlugin struct {
    plugin.Base           // Embed common functionality
    param.Manager         // Embed parameter management
    state.Manager         // Embed state management
    
    // Plugin-specific fields
    smoothedGain float64
}
```

### 2. Interface Segregation

Small, focused interfaces:

```go
type Processor interface {
    Process(in, out *audio.Buffer) error
}

type Stateful interface {
    SaveState(w io.Writer) error
    LoadState(r io.Reader) error
}
```

### 3. Dependency Injection

```go
func NewPlugin(opts ...Option) *Plugin {
    p := &Plugin{
        logger: defaultLogger,
        params: param.NewManager(),
    }
    
    for _, opt := range opts {
        opt(p)
    }
    
    return p
}
```

### 4. Builder Pattern

```go
param := param.NewBuilder(0, "Gain").
    Range(0, 2, 1).
    Flags(param.Automatable | param.Modulatable).
    Format(param.Decibel).
    Build()
```

## Performance Considerations

### Optimization Strategies

1. **Lock-Free Algorithms**: Atomic operations for parameters
2. **Memory Pooling**: Reuse objects to prevent allocation
3. **SIMD Optimization**: Use Go's runtime SIMD when available
4. **Cache Efficiency**: Data structure layout for cache lines

### Benchmarking

Regular benchmarking of critical paths:

```go
func BenchmarkProcessBuffer(b *testing.B) {
    buffer := audio.NewBuffer(2, 1024)
    plugin := NewGainPlugin()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        plugin.Process(buffer, buffer)
    }
}
```

## Future Directions

### Planned Enhancements

1. **WebAssembly Support**: Run plugins in browsers
2. **GPU Acceleration**: Optional compute shaders
3. **Network Transparency**: Distributed plugin processing
4. **Advanced Debugging**: Real-time performance profiling

### Research Areas

1. **Garbage Collection**: Techniques to minimize GC impact
2. **JIT Optimization**: Leverage Go's runtime optimization
3. **Memory Mapping**: Efficient large sample handling
4. **Concurrent Processing**: Multi-core audio processing

## Conclusion

ClapGo provides a robust, performant bridge between CLAP and Go, enabling developers to write professional audio plugins in Go while maintaining the full power and flexibility of the CLAP standard. The architecture prioritizes performance, safety, and developer experience without hiding the underlying CLAP concepts.