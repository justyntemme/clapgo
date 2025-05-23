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
Event handling abstractions:
- `EventHandler` interface for processing CLAP events
- Event type definitions
- Sample-accurate event timing

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

## Event Processing Architecture

Events flow through multiple layers:

1. **C Events** -> Bridge converts to Go event types
2. **Event Handler** processes events with sample accuracy
3. **Parameter events** update the parameter manager
4. **Output events** generated by plugin, converted back to C

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
- Core plugin lifecycle
- Basic audio processing
- Parameter management with thread safety
- State save/load
- Audio ports (stereo only)
- Manifest-based discovery
- CMake build system

### Partially Implemented
- Event system (parameter events only)
- Extension system (framework in place)
- Host callbacks (minimal)

### Not Yet Implemented
- GUI extensions
- Note/MIDI processing
- Advanced audio configurations
- Most optional extensions
- Factory extensions
- Draft extensions

## Future Architecture Considerations

### 1. Performance Optimizations
- Potential for lock-free ring buffers
- SIMD operations in Go where beneficial
- Reduced allocation in audio path

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