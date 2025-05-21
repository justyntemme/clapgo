# ClapGo Architecture

ClapGo is a bridge system that enables creating CLAP (CLever Audio Plugin) plugins using the Go programming language. It provides a C bridge layer that interfaces with CLAP hosts while allowing plugin logic to be implemented in Go.

## Overview

The architecture uses a **manifest-driven approach** where plugins are described by JSON manifest files and implemented as Go shared libraries with standardized exports. This design eliminates the need for a centralized plugin registry and enables clean separation between the C CLAP interface and Go plugin implementations.

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   CLAP Host     │◄──►│   ClapGo Bridge  │◄──►│  Go Plugin      │
│   (DAW/Host)    │    │   (C Library)    │    │  (Shared Lib)   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
        │                       │                       │
        │                       │                       │
    CLAP ABI              Manifest System         Standardized
   Interface                                      Go Exports
```

## Core Components

### 1. C Bridge Layer (`src/c/`)

#### **bridge.c/bridge.h** - Core Bridge Implementation
- **Purpose**: Central orchestrator managing the bridge between CLAP hosts and Go plugins
- **Key Responsibilities**:
  - Plugin discovery and manifest loading
  - Dynamic Go shared library loading
  - Symbol resolution and function pointer caching  
  - Plugin instance lifecycle management
  - CLAP interface implementation

#### **plugin.c/plugin.h** - CLAP Entry Point
- **Purpose**: Minimal CLAP plugin entry point (`clap_entry`)
- **Responsibilities**:
  - Implements required CLAP entry structure
  - Delegates all operations to bridge layer
  - Provides plugin factory interface

#### **manifest.c/manifest.h** - JSON Manifest Processing
- **Purpose**: Handles JSON manifest parsing and CLAP descriptor generation
- **Key Functions**:
  - `manifest_load_from_file()`: Parses JSON using json-c
  - `manifest_to_descriptor()`: Converts manifest to CLAP descriptor
  - Manifest file discovery and validation

### 2. Go Bridge Package (`pkg/bridge/`)

#### **bridge.go** - Go Bridge Entry Point
- **Purpose**: Main C-Go interoperability layer
- **Responsibilities**:
  - CGO header inclusions and forward declarations
  - CLAP event handling utilities
  - Bridge initialization and setup

### 3. Go API Layer (`pkg/api/`)

#### **plugin.go** - Core Plugin API
- **Purpose**: Foundation of the Go plugin API
- **Key Interfaces**:
  - `Plugin`: Main plugin lifecycle interface
  - `EventHandler`: Event processing during audio processing
  - `Factory`: Plugin factory for creating instances
- **Key Types**:
  - `PluginInfo`: Plugin metadata
  - `Event`, `ParamEvent`, `NoteEvent`: Event system

#### **extensions.go** - Extension Interfaces
- **Purpose**: Optional CLAP extension interfaces
- **Supported Extensions**:
  - `AudioPortsProvider`: Audio port configuration
  - `ParamsProvider`: Parameter management  
  - `StateProvider`: Plugin state save/load
  - `NotePortsProvider`: Note handling
  - `GUIProvider`, `LatencyProvider`, `TailProvider`, etc.

#### **constants.go** - CLAP Constants
- **Purpose**: Centralized CLAP specification constants
- **Includes**: Extension IDs, event types, flags, process status codes

## Plugin Loading Architecture

### Manifest System

Each plugin is described by a JSON manifest file that contains:

```json
{
  "schemaVersion": "1.0",
  "plugin": {
    "id": "com.vendor.plugin",
    "name": "Plugin Name",
    "vendor": "Vendor",
    "version": "1.0.0",
    "description": "Plugin description",
    "features": ["audio-effect", "stereo"]
  },
  "build": {
    "goSharedLibrary": "libplugin.so",
    "entryPoint": "CreatePlugin",
    "dependencies": []
  },
  "extensions": [...],
  "parameters": [...]
}
```

### Plugin Loading Flow

1. **Discovery**: CLAP host calls `clap_entry.init()` with plugin path
2. **Manifest Loading**: Bridge discovers and parses JSON manifest file
3. **Library Loading**: Go shared library loaded via `dlopen()`
4. **Symbol Resolution**: Standardized function exports resolved once and cached:
   ```c
   entry->create_plugin = dlsym(library, "ClapGo_CreatePlugin");
   entry->plugin_init = dlsym(library, "ClapGo_PluginInit");
   // ... etc for all lifecycle functions
   ```
5. **Descriptor Creation**: CLAP descriptor generated from manifest metadata
6. **Instance Creation**: Plugin instances created via Go factory functions

### Function Pointer Caching ("Crossing Guard" Pattern)

The architecture implements an efficient "crossing guard" pattern:

- **One-time loading**: Each Go shared library loaded once with `dlopen()`
- **Symbol caching**: All Go exports resolved once and stored as C function pointers
- **Direct calls**: Subsequent operations use cached pointers - no repeated symbol resolution
- **Persistent runtime**: Go runtime remains resident, no restart overhead

```c
// Functions cached in manifest_plugin_entry_t:
entry->create_plugin(host, plugin_id);     // Direct function pointer call
entry->plugin_process(plugin, process);    // No symbol lookup overhead
```

## Memory Management & Safety

### CGO Handle System

To safely pass Go objects to C code, ClapGo uses the CGO handle system:

```go
//export ClapGo_CreatePlugin
func ClapGo_CreatePlugin(host unsafe.Pointer, pluginID *C.char) unsafe.Pointer {
    // Create safe handle to Go object
    handle := cgo.NewHandle(pluginInstance)
    return unsafe.Pointer(handle)
}

//export ClapGo_PluginProcess
func ClapGo_PluginProcess(plugin unsafe.Pointer, ...) C.int {
    // Convert handle back to Go object
    handle := cgo.Handle(plugin)
    p := handle.Value().(*PluginType)
    return C.int(p.Process(...))
}
```

### Memory Layout

```
Host Process Memory
├── plugin.clap (C Code)
│   ├── Function pointers → Go functions (cached)
│   ├── Manifest registry (up to 32 entries)
│   └── CGO handles → Go plugin instances
└── libplugin.so (Go Code)
    ├── Go runtime (persistent, shared)
    ├── Plugin instances (managed by CGO handles)
    └── Go heap & garbage collector
```

## Interface Standardization

### Required Go Exports

Every Go plugin must export these standardized functions:

```go
//export ClapGo_CreatePlugin
func ClapGo_CreatePlugin(host unsafe.Pointer, pluginID *C.char) unsafe.Pointer

//export ClapGo_PluginInit  
func ClapGo_PluginInit(plugin unsafe.Pointer) C.bool

//export ClapGo_PluginDestroy
func ClapGo_PluginDestroy(plugin unsafe.Pointer)

//export ClapGo_PluginActivate
func ClapGo_PluginActivate(plugin unsafe.Pointer, sampleRate C.double, minFrames, maxFrames C.uint32_t) C.bool

//export ClapGo_PluginProcess
func ClapGo_PluginProcess(plugin unsafe.Pointer, steadyTime C.int64_t, framesCount C.uint32_t, audioIn, audioOut unsafe.Pointer, events unsafe.Pointer) C.int

// ... additional lifecycle functions
```

### Plugin Implementation Pattern

```go
type MyPlugin struct {
    // Plugin state and parameters
}

func (p *MyPlugin) Init() bool { /* ... */ }
func (p *MyPlugin) Process(/* ... */) int { /* ... */ }
// ... implement Plugin interface

var pluginInstance *MyPlugin

func init() {
    pluginInstance = NewMyPlugin()
}

//export ClapGo_CreatePlugin
func ClapGo_CreatePlugin(host unsafe.Pointer, pluginID *C.char) unsafe.Pointer {
    if C.GoString(pluginID) == PluginID {
        return unsafe.Pointer(cgo.NewHandle(pluginInstance))
    }
    return nil
}
```

## Extension System

### Interface-Based Extensions

ClapGo uses Go interfaces to provide optional CLAP extensions:

```go
type Plugin interface {
    Init() bool
    Activate(sampleRate float64, minFrames, maxFrames uint32) bool
    Process(/* ... */) int
    // ... core methods
}

// Optional extensions
type ParamsProvider interface {
    GetParamCount() uint32
    GetParamInfo(index uint32) ParamInfo
    GetParamValue(id uint32) float64
    SetParamValue(id uint32, value float64)
}

type AudioPortsProvider interface {
    GetAudioPortsCount(isInput bool) uint32
    GetAudioPortInfo(index uint32, isInput bool) AudioPortInfo
}
```

### Extension Detection

The C bridge queries for extensions via `ClapGo_PluginGetExtension`:

```go
//export ClapGo_PluginGetExtension  
func ClapGo_PluginGetExtension(plugin unsafe.Pointer, id *C.char) unsafe.Pointer {
    extID := C.GoString(id)
    handle := cgo.Handle(plugin)
    p := handle.Value().(Plugin)
    
    switch extID {
    case api.ExtParams:
        if provider, ok := p.(ParamsProvider); ok {
            // Return C interface implementation
            return createParamsExtension(provider)
        }
    // ... other extensions
    }
    return nil
}
```

## Cross-Platform Support

### Platform Abstractions

```c
// bridge.h - Platform detection and abstractions
#if defined(_WIN32) || defined(_WIN64)
    typedef HMODULE clapgo_library_t;
    typedef FARPROC clapgo_symbol_t;
#elif defined(__APPLE__) 
    typedef void* clapgo_library_t;
    typedef void* clapgo_symbol_t;
#else // Linux
    typedef void* clapgo_library_t; 
    typedef void* clapgo_symbol_t;
#endif
```

### Dynamic Loading

```c
// Cross-platform library loading
#if defined(CLAPGO_OS_WINDOWS)
    library = LoadLibraryA(lib_path);
    symbol = GetProcAddress(library, symbol_name);
#elif defined(CLAPGO_OS_MACOS) || defined(CLAPGO_OS_LINUX)
    library = dlopen(lib_path, RTLD_LAZY | RTLD_LOCAL);
    symbol = dlsym(library, symbol_name);
#endif
```

## Event Processing

### Event Flow

1. **Host → C Bridge**: CLAP events received via `clap_process_t`
2. **C Bridge → Go**: Events converted and passed to Go via `EventHandler` interface  
3. **Go Processing**: Plugin processes events and generates audio
4. **Go → C Bridge**: Results returned via standardized return values
5. **C Bridge → Host**: Results returned to host via CLAP interface

### Event Types

ClapGo supports all major CLAP event types:
- **Note Events**: Note on/off, choke, expression
- **Parameter Events**: Value changes, modulation
- **Transport Events**: Play/stop, tempo, time signature
- **MIDI Events**: Raw MIDI, SysEx, MIDI 2.0

## Build System

### Makefile Structure

```makefile
# Platform detection
UNAME := $(shell uname)
ifeq ($(UNAME), Linux)
    SO_EXT := so
    CLAP_FORMAT := so
endif

# Go compilation  
CGO_ENABLED := 1
GO_FLAGS := -buildmode=c-shared

# C compilation
CFLAGS := -I./include/clap/include -fPIC -Wall -Wextra
LDFLAGS := -shared $(shell pkg-config --libs json-c)
```

### Build Process

1. **Go Bridge**: `libgoclap.so` built from `pkg/bridge` 
2. **Plugin Libraries**: `libplugin.so` built from plugin source
3. **C Bridge**: Bridge objects compiled and linked
4. **Final Plugin**: Everything linked into `plugin.clap`
5. **Manifest Copy**: JSON manifest copied to plugin directory

## Design Principles

### 1. **Separation of Concerns**
- C handles CLAP protocol compliance
- Go handles plugin logic and audio processing
- JSON manifests handle metadata and configuration

### 2. **Type Safety**
- CGO handles provide memory-safe object passing
- Interface-based extension system prevents runtime errors
- Standardized exports ensure ABI compatibility

### 3. **Performance**
- Function pointer caching eliminates symbol lookup overhead
- Persistent Go runtime avoids initialization costs
- Direct C-Go calls minimize marshaling overhead

### 4. **Extensibility**
- Interface-based extensions support optional CLAP features
- Manifest system allows rich plugin descriptions
- Plugin-specific extensions possible via custom interfaces

### 5. **Cross-Platform Compatibility**
- Unified abstractions over platform-specific APIs
- Consistent behavior across Windows/macOS/Linux
- Standard build system with platform detection

## Thread Safety

### CLAP Threading Model
- **Main Thread**: Plugin lifecycle, parameter changes, state
- **Audio Thread**: Real-time audio processing (`process()` calls)
- **Background Threads**: Optional for non-RT operations

### ClapGo Thread Handling
- CGO handles are thread-safe for concurrent access
- Go runtime handles threading automatically
- Plugin implementations must handle CLAP threading requirements
- No additional synchronization needed in bridge layer

This architecture successfully bridges the gap between Go's high-level programming model and CLAP's real-time audio requirements while maintaining safety, performance, and compatibility across platforms.