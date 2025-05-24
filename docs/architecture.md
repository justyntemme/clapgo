# ClapGo Architecture Documentation

## Overview

ClapGo is a bridge library that enables developers to write CLAP (CLever Audio Plugin) plugins using the Go programming language. It provides a carefully designed architecture that maintains full CLAP compliance while leveraging Go's strengths for audio plugin development.

## Core Architecture Principles

### 1. Manifest-Driven Plugin Discovery
ClapGo uses a manifest system rather than runtime registration. Each plugin consists of:
- A JSON manifest file containing plugin metadata
- A Go shared library (.so/.dll/.dylib) with standardized exports
- A C bridge that handles CLAP host communication

### 2. C Bridge Pattern
The architecture follows a strict separation:
```
CLAP Host <-> C Bridge <-> Go Plugin Implementation
```

The C bridge (located in `src/c/`) handles all CLAP protocol requirements and translates between C and Go through a defined set of exported functions.

### 3. No Go-Side Registration
Unlike traditional plugin frameworks, ClapGo deliberately avoids Go-side registration systems. Plugins are discovered through the filesystem manifest system, maintaining architectural simplicity.

## Component Architecture

### 1. C Bridge Layer (`src/c/`)

#### bridge.c/bridge.h
The main bridge implementation that:
- Implements the CLAP plugin interface (`clap_plugin_t`)
- Manages plugin lifecycle (init, destroy, activate, deactivate, process)
- Handles extension negotiation
- Translates between C and Go function calls

Key structures:
```c
typedef struct {
    clap_plugin_t plugin;
    void* go_instance;
    const clap_host_t* host;
    // Extension implementations
} clapgo_plugin_t;
```

#### manifest.c/manifest.h
Handles plugin discovery through JSON manifests:
- Scans for `.json` files in plugin directories
- Parses plugin metadata
- Loads corresponding Go shared libraries
- Provides plugin factory implementation

#### plugin.c
Entry point implementation:
- Implements `clap_plugin_entry_t`
- Provides DSO initialization/cleanup
- Returns the plugin factory

### 2. Go API Layer (`pkg/api/`)

#### plugin.go
Defines the core `Plugin` interface that all Go plugins must implement:
```go
type Plugin interface {
    Init() bool
    Destroy()
    Activate(sampleRate float64, minFrames, maxFrames uint32) bool
    Deactivate()
    StartProcessing() bool
    StopProcessing()
    Reset()
    Process(audio *Audio, events *Events, frameCount uint32) ProcessStatus
    GetExtension(id string) unsafe.Pointer
}
```

#### audio.go
Provides Go-friendly audio buffer abstractions:
- `Audio` struct wrapping input/output buffers
- Channel access methods
- Conversion between C and Go slice representations

#### params.go
Parameter management system:
- `Parameter` interface for different parameter types
- `ParameterManager` for thread-safe parameter storage
- Parameter validation and automation support

#### events.go
Zero-allocation event handling system:
- `EventProcessor` with pre-allocated event pools for all event types
- Complete event type support (note, parameter, MIDI, transport, etc.)
- Sample-accurate event timing with automatic pool lifecycle management
- Thread-safe pool management with diagnostic tracking

#### extensions.go
Extension identifier constants and interfaces:
- Defines all CLAP extension IDs
- Provides Go interfaces for implemented extensions

#### stream.go
Stream I/O abstractions for state save/load:
- `Stream` interface wrapping CLAP streams
- Read/Write methods with Go-friendly APIs

### 3. Bridge Package (`pkg/bridge/`)

Provides the critical exported C functions that the C bridge calls:
```go
//export ClapGo_CreatePlugin
//export ClapGo_PluginInit
//export ClapGo_PluginDestroy
//export ClapGo_PluginActivate
//export ClapGo_PluginDeactivate
//export ClapGo_PluginStartProcessing
//export ClapGo_PluginStopProcessing
//export ClapGo_PluginReset
//export ClapGo_PluginProcess
//export ClapGo_PluginGetExtension
```

These functions:
- Handle unsafe pointer conversions
- Manage plugin instance lifecycle
- Route calls to the appropriate Go plugin methods

### 4. CLAP Package (`pkg/clap/`)

#### base.go
Provides base plugin implementation helpers:
- `BasePlugin` struct with common functionality
- Default implementations for optional methods
- Extension management helpers

#### params.go
Parameter extension implementation:
- Maps between C callbacks and Go parameter manager
- Handles parameter info queries
- Manages value/text conversions

### 5. Manifest Package (`pkg/manifest/`)

Defines the plugin manifest structure:
```go
type Manifest struct {
    ID          string
    Name        string
    Vendor      string
    Version     string
    URL         string
    Description string
    Features    []string
    Library     string
}
```

### 6. Utility Package (`pkg/util/`)

Common utilities for plugin development:
- Atomic operations for thread-safe parameter access
- Audio DSP helpers
- Format conversion utilities
- Envelope generators

## Zero-Allocation Event Processing Architecture

ClapGo's event processing system is designed for professional real-time audio performance with zero allocations in the audio thread.

### EventPool Design

The `EventPool` struct manages pre-allocated objects for all CLAP event types:

```go
type EventPool struct {
    // Separate pools for each event type prevent interface{} boxing
    paramValuePool      sync.Pool
    paramModPool        sync.Pool
    noteEventPool       sync.Pool
    midiPool           sync.Pool
    // ... all event types
    
    // Diagnostics for performance monitoring
    totalAllocations   uint64
    poolHits          uint64
    poolMisses        uint64
    highWaterMark     uint64
}
```

### Key Features

1. **Pre-allocation**: Events allocated at startup, not during processing
2. **Type-specific pools**: Avoid interface{} boxing allocations
3. **Automatic lifecycle**: Events returned to pool after processing
4. **Diagnostic tracking**: Monitor allocation performance
5. **Thread safety**: Safe for concurrent host usage
6. **Warning system**: Alerts when allocations occur in audio thread

### Usage Pattern

```go
// Get event from pool (no allocation)
event := pool.GetEvent()
paramEvent := pool.GetParamValueEvent()

// Use event for processing
// ...

// Automatically returned to pool after ProcessAllEvents()
// or manually:
pool.ReturnEvent(event)
pool.ReturnParamValueEvent(paramEvent)
```

This design guarantees deterministic performance without garbage collection pauses during audio processing.

## Build System Architecture

### CMake Integration
The project uses CMake for building plugins:

1. **GoLibrary.cmake**: Custom CMake module that:
   - Compiles Go code to shared libraries
   - Manages CGO dependencies
   - Handles platform-specific build flags

2. **Plugin CMakeLists.txt**: Each plugin defines:
   - Go source files
   - Link dependencies
   - Installation rules

3. **Build Process**:
   ```
   Go sources -> go build -buildmode=c-shared -> .so/.dll/.dylib
   C bridge -> gcc/clang -> object files
   Link together -> final CLAP plugin
   ```

## Extension Architecture

Extensions are implemented through a layered approach:

### 1. C Layer
Each extension has C callbacks in the bridge that forward to Go:
```c
static const clap_plugin_params_t s_clapgo_params = {
    .count = clapgo_params_count,
    .get_info = clapgo_params_get_info,
    .get_value = clapgo_params_get_value,
    .value_to_text = clapgo_params_value_to_text,
    .text_to_value = clapgo_params_text_to_value,
    .flush = clapgo_params_flush
};
```

### 2. Bridge Layer
Exported functions handle the C->Go transition:
```go
//export ClapGo_ParamsCount
func ClapGo_ParamsCount(plugin unsafe.Pointer) C.uint32_t
```

### 3. Go Implementation
Extensions are provided through the plugin's `GetExtension` method:
```go
func (p *MyPlugin) GetExtension(id string) unsafe.Pointer {
    switch id {
    case api.ExtParams:
        return p.paramManager.GetCInterface()
    }
    return nil
}
```

## Thread Safety Architecture

ClapGo enforces CLAP's thread safety requirements:

### 1. Main Thread Operations
- Plugin initialization/destruction
- Extension negotiation
- GUI operations
- Non-realtime parameter changes

### 2. Audio Thread Operations
- Process callback
- Parameter value reads
- Event processing

### 3. Synchronization
- Atomic operations for parameter values
- Lock-free data structures where possible
- Clear thread ownership documentation

## Memory Management

### 1. Ownership Rules
- C bridge owns the `clapgo_plugin_t` structure
- Go owns the plugin implementation instance
- Shared ownership through reference counting where needed

### 2. Lifetime Management
- Plugin instances created by factory, destroyed by host
- Clear initialization/cleanup sequences
- No memory leaks through careful pointer management

### 3. Zero-Allocation Real-Time Processing
- **Event Pool System**: Pre-allocated events eliminate allocations in audio thread
- **Object Reuse**: Events automatically returned to pools after processing
- **Allocation Tracking**: Diagnostic logging when allocations occur
- **Performance Guarantees**: Deterministic timing without GC pauses

**Current Status**: Event processing path is fully zero-allocation. Additional optimizations needed for:
- MIDI data buffers (should use fixed arrays instead of slices)
- Host logger string formatting
- Parameter listener notifications
- Stream I/O operations

## Event Processing Architecture

ClapGo implements a zero-allocation event processing system for real-time audio performance:

### Zero-Allocation Event Pool System

1. **Event Pool Initialization**
   - Pre-allocates pools for all event types at startup
   - Separate pools prevent interface{} boxing allocations
   - Configurable pool sizes with automatic growth warnings

2. **Event Processing Flow**
   ```
   C Events -> EventProcessor.convertCEventToGo() -> Pool.GetEvent()
   Process events -> TypedEventHandler routing
   Return to pool -> EventProcessor.ReturnEventToPool()
   ```

3. **Pool Management**
   - Thread-safe sync.Pool implementation
   - Automatic event cleanup to prevent data leaks
   - Performance diagnostics (hits/misses/high water mark)
   - Warning logs when allocation occurs in audio thread

4. **Supported Event Types**
   - Parameter events (value, modulation, gesture)
   - Note events (on, off, choke, end, expression)
   - MIDI events (1.0, SysEx, 2.0)
   - Transport events

### Event Lifecycle
1. **Get**: Events retrieved from pre-allocated pools
2. **Process**: Events routed to typed handlers
3. **Return**: Events automatically returned to pools after processing
4. **Clear**: Event data cleared to prevent leaks between uses

This design ensures deterministic real-time performance with no garbage collection pauses during audio processing.

## State Management

Plugin state is handled through streams:

1. **Save**: Plugin serializes state to stream
2. **Load**: Plugin deserializes from stream
3. **Format**: Plugins define their own state format
4. **Versioning**: Plugins handle version compatibility

## Example Plugin Structure

A typical ClapGo plugin consists of:

```
my-plugin/
├── CMakeLists.txt      # Build configuration
├── main.go             # Plugin implementation
├── constants.go        # Plugin metadata
├── my-plugin.json      # Manifest file
└── presets/           # Optional preset files
    └── factory/
        └── default.json
```

The main.go implements:
- Plugin struct with required interface
- Parameter definitions
- DSP processing
- Extension support
- State management

## Design Decisions and Rationale

### 1. Manifest-Driven Discovery
**Why**: Eliminates complex registration code, simplifies plugin discovery, enables static analysis of available plugins.

### 2. C Bridge Pattern
**Why**: Maintains ABI stability, enables Go's garbage collector, provides clear separation of concerns.

### 3. No Simplified APIs
**Why**: ClapGo is a bridge, not a framework. Full CLAP compliance requires exposing all concepts.

### 4. Exported Function Pattern
**Why**: CGO requires explicit exports, provides clear C-Go boundary, enables efficient callbacks.

### 5. Extension-Based Architecture
**Why**: Matches CLAP's design, allows incremental implementation, maintains compatibility.

## Current Implementation Status

### Fully Implemented
- Core plugin lifecycle with complete CLAP compliance
- Zero-allocation audio processing with event pools
- Complete event system (all event types: parameter, note, MIDI, transport)
- Parameter management with thread safety
- State save/load with context awareness
- Audio ports with multiple configurations
- Note ports for instrument plugins
- Manifest-based discovery system
- Extension architecture with clean C bridge pattern
- Essential extensions: params, state, audio-ports, note-ports, latency, tail, log, timer
- Advanced extensions: audio-ports-config, surround, voice-info, track-info
- State management: state-context, preset-load
- Real-time performance optimizations

### Partially Implemented
- Host callbacks (logging, timer, track info implemented)
- Stream I/O (basic implementation, needs buffer pools for zero-allocation)

### Not Yet Implemented
- GUI extensions
- Context menu extension
- Remote controls extension
- Parameter indication extension
- Factory extensions (preset discovery)
- Draft extensions (tuning, transport control, undo, etc.)

### Performance Status
- ✅ **Zero-allocation event processing** - Fully implemented with diagnostic tracking
- ⚠️ **Additional optimizations needed** - MIDI buffers, logging, parameter notifications
- ✅ **Real-time audio thread safety** - Complete separation of audio and main thread operations
- ✅ **Professional plugin validation** - Both example plugins pass clap-validator tests

## Future Architecture Considerations

### 1. Performance Optimizations
- Complete zero-allocation processing (eliminate remaining MIDI buffer, logging, and parameter allocations)
- Lock-free ring buffers for parameter updates
- SIMD operations in Go where beneficial
- Allocation tracking and benchmarking infrastructure

### 2. Extended Platform Support
- Windows and macOS specific features
- Platform-specific GUI integration
- Native file dialog support

### 3. Advanced Extensions
- Preset discovery system
- Plugin state conversion
- Remote control pages
- Surround/ambisonic support

### 4. Developer Experience
- Code generation for boilerplate
- Enhanced debugging support
- Performance profiling integration