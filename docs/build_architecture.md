# ClapGo Build Architecture

This document outlines the build architecture for ClapGo, focusing on proper integration between Go and the CLAP plugin system using pure C (no C++).

## 1. Architecture Overview

### 1.1 Key Components

The ClapGo build architecture consists of these core components:

1. **CLAP Plugin C Wrapper (.clap)** - Implements CLAP entry point and plugin interface using pure C
2. **Go Shared Library (.so/.dylib/.dll)** - Contains Go plugin implementations and bridge functionality
3. **Dynamic Loading Bridge** - Connects the C wrapper to the Go implementation at runtime
4. **Plugin Registry** - Manages plugin registration and instantiation in Go
5. **Build System** - Coordinates building and installing both components

### 1.2 Design Principles

The architecture is guided by these principles:

1. **Use Pure C for CLAP Interface** - No C++ dependencies to maintain simplicity
2. **Dynamic Loading** - Load Go library at runtime rather than linking directly
3. **Plugin ID Ownership** - Plugins define their own IDs
4. **Clean Separation** - Clear boundaries between components
5. **Flexible Registration** - Support multiple plugins with one codebase

### 1.3 Visual Architecture Overview

```
┌───────────────────────┐     ┌───────────────────────┐
│                       │     │                       │
│   CLAP Host (DAW)     │     │   Go Plugin Implementation│
│                       │     │                       │
└───────────┬───────────┘     └───────────┬───────────┘
            │                             │
            ▼                             │
┌───────────────────────┐                 │
│                       │                 │
│   CLAP Plugin         │                 │
│   (.clap file)        │                 │
│                       │                 │
│  ┌─────────────────┐  │                 │
│  │ CLAP Entry Point│  │                 │
│  └────────┬────────┘  │                 │
│           │           │                 │
│  ┌────────▼────────┐  │      Loads      │
│  │ Dynamic Loader  ├──┼─────────────────┘
│  └────────┬────────┘  │
│           │           │
│  ┌────────▼────────┐  │
│  │ C Plugin API    │  │
│  └─────────────────┘  │
│                       │
└───────────────────────┘
```

## 2. Component Details

### 2.1 CLAP Plugin C Wrapper

**Purpose**: Implement the CLAP plugin interface and serve as the entry point for CLAP hosts.

**Key Files**:
- `src/c/plugin.c` - CLAP entry point and plugin interface
- `src/c/plugin.h` - Header defining types and functions
- `src/c/bridge.c` - Dynamic loading for the Go shared library

**Responsibilities**:
1. Implement CLAP entry point (`clap_entry`) and plugin factory
2. Dynamically load the Go shared library at runtime
3. Forward CLAP calls to the Go implementation via function pointers
4. Manage plugin lifecycle (init, destroy, process, etc.)
5. Handle errors if the Go library cannot be loaded

### 2.2 Go Shared Library

**Purpose**: Provide the actual plugin implementation and functionality.

**Key Files**:
- `internal/bridge/bridge.go` - Exports functions for C to call
- `internal/registry/registry.go` - Manages plugin registration
- `pkg/api/plugin.go` - Defines plugin interface
- `examples/*/main.go` - Plugin implementations

**Responsibilities**:
1. Export functions via CGO for the C wrapper to call
2. Maintain a registry of available plugins
3. Create plugin instances when requested
4. Implement plugin functionality
5. Provide a clean API for plugin developers

### 2.3 Dynamic Loading Bridge

**Purpose**: Connect the C wrapper to the Go implementation at runtime.

**Key Files**:
- `src/c/bridge.c` - Dynamic library loading functions
- `src/c/bridge.h` - Function pointer definitions

**Responsibilities**:
1. Find the Go shared library at runtime
2. Load the library using platform-specific functions (`dlopen`, `LoadLibrary`, etc.)
3. Resolve function pointers for all required functions
4. Report errors if the library cannot be found or functions are missing
5. Clean up resources when the plugin is unloaded

### 2.4 Plugin Registry

**Purpose**: Manage plugin registration and instantiation.

**Key Files**:
- `internal/registry/registry.go` - Registry implementation
- `pkg/api/plugin.go` - Plugin interface definition

**Responsibilities**:
1. Allow plugins to register with the system
2. Ensure plugin IDs are unique and valid
3. Provide stable enumeration of plugins
4. Create new instances of plugins when requested
5. Validate plugin metadata

### 2.5 Build System

**Purpose**: Coordinate building and installing both components.

**Key Files**:
- `CMakeLists.txt` - Root build configuration
- `cmake/GoLibrary.cmake` - Go library build helpers
- `cmake/ClapPlugin.cmake` - CLAP plugin build helpers
- `build.sh` - Build script
- `install.sh` - Installation script

**Responsibilities**:
1. Build the Go shared library
2. Build the CLAP plugin
3. Package them together correctly
4. Handle platform-specific differences
5. Provide installation rules

## 3. Build Process

### 3.1 Building the Go Shared Library

The Go shared library is built using these steps:

1. Compile Go code with `buildmode=c-shared`:
   ```bash
   CGO_ENABLED=1 go build -buildmode=c-shared -o libgoclap.so ./cmd/goclap
   ```

2. This produces:
   - A shared library (`libgoclap.so` on Linux)
   - A C header file (`libgoclap.h`) with exported function definitions

3. Key considerations:
   - All functions to be called from C must be exported with `//export FunctionName`
   - The main package must have a main function (even if empty)
   - CGO is used to define C types and functions

### 3.2 Building the CLAP Plugin

The CLAP plugin is built using these steps:

1. Compile C wrapper code:
   ```bash
   cc -shared -fPIC -o plugin.clap plugin.c bridge.c -I/path/to/clap/include
   ```

2. Key considerations:
   - The output file must have the `.clap` extension
   - The CLAP entry point must be exported
   - No direct linking to the Go library (uses dynamic loading)

### 3.3 CMake Integration

The build system uses CMake to coordinate the process:

1. Define custom commands to build the Go shared library
2. Build the C wrapper as a MODULE library
3. Set output name and properties for the CLAP plugin
4. Add installation rules for both components

Example CMake configuration:
```cmake
# Build Go shared library
add_custom_command(
    OUTPUT ${CMAKE_BINARY_DIR}/libgoclap.so
    COMMAND env CGO_ENABLED=1 go build -buildmode=c-shared -o ${CMAKE_BINARY_DIR}/libgoclap.so ./cmd/goclap
    WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
    COMMENT "Building Go shared library"
)
add_custom_target(go-shared-library DEPENDS ${CMAKE_BINARY_DIR}/libgoclap.so)

# Build CLAP plugin
add_library(clapgo-plugin MODULE src/c/plugin.c src/c/bridge.c)
add_dependencies(clapgo-plugin go-shared-library)
set_target_properties(clapgo-plugin PROPERTIES
    PREFIX ""
    SUFFIX ".clap"
)
```

## 4. Runtime Behavior

### 4.1 Plugin Loading

When a CLAP host loads a plugin, this sequence occurs:

1. Host loads the `.clap` file and calls the `clap_entry.init` function
2. The C wrapper searches for the Go shared library:
   - Same directory as the plugin
   - User's plugin directory (`~/.clap`)
   - System plugin directories
3. The shared library is loaded using platform-specific functions
4. Function pointers are resolved for all required functions
5. Plugin descriptions are retrieved from the Go library
6. The host can enumerate and create plugin instances

### 4.2 Plugin Creation

When a host creates a plugin instance:

1. Host calls `clap_plugin_factory.create_plugin` with a plugin ID
2. C wrapper calls into the Go library to create the plugin instance
3. Go registry finds the registered plugin matching the ID
4. Registry creates a new instance of the plugin
5. C wrapper creates a handle to the Go instance
6. Plugin lifecycle is managed through the C wrapper

### 4.3 Audio Processing

During audio processing:

1. Host calls the plugin's `process` function
2. C wrapper converts CLAP audio buffers to Go format
3. C wrapper calls the Go plugin's process function
4. Go plugin processes the audio
5. Results are passed back through the C wrapper to the host

## 5. Plugin Implementation

### 5.1 Plugin Interface

All plugins must implement the `api.Plugin` interface:

```go
type Plugin interface {
    // Core lifecycle methods
    Init() bool
    Destroy()
    Activate(sampleRate float64, minFrames, maxFrames uint32) bool
    Deactivate()
    StartProcessing() bool
    StopProcessing()
    Reset()
    Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events EventHandler) int
    
    // Extension and identification
    GetExtension(id string) unsafe.Pointer
    OnMainThread()
    GetPluginID() string
    GetPluginInfo() PluginInfo
}
```

### 5.2 Plugin Registration

Plugins register themselves with the registry during initialization:

```go
func init() {
    info := api.PluginInfo{
        ID:          "com.clapgo.gain",
        Name:        "Simple Gain",
        Vendor:      "ClapGo",
        // ... other fields ...
    }
    
    // Register with factory function to create new instances
    registry.Register(info, func() api.Plugin {
        return &GainPlugin{
            // Initialize default state
            gain: 1.0,
            // ... other fields ...
        }
    })
}
```

### 5.3 Example Plugin Structure

A typical plugin structure:

```go
package main

import (
    "github.com/justyntemme/clapgo/pkg/api"
    "github.com/justyntemme/clapgo/internal/registry"
)

// GainPlugin implements a simple gain plugin
type GainPlugin struct {
    gain         float64
    sampleRate   float64
    isActivated  bool
    isProcessing bool
    paramManager *ParamManager
}

// GetPluginID returns the plugin ID
func (p *GainPlugin) GetPluginID() string {
    return "com.clapgo.gain"
}

// GetPluginInfo returns information about the plugin
func (p *GainPlugin) GetPluginInfo() api.PluginInfo {
    return api.PluginInfo{
        ID:          p.GetPluginID(),
        Name:        "Simple Gain",
        Vendor:      "ClapGo",
        // ... other fields ...
    }
}

// Process applies gain to input audio
func (p *GainPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events *api.EventHandler) int {
    // ... implementation ...
}

// ... implement other interface methods ...

// Register the plugin
func init() {
    registry.Register(
        (&GainPlugin{}).GetPluginInfo(),
        func() api.Plugin { return &GainPlugin{gain: 1.0} }
    )
}

// Required for buildmode=c-shared
func main() {}
```

## 6. Installation and Deployment

### 6.1 Installation Process

The installation process consists of these steps:

1. Build both the Go shared library and CLAP plugin
2. Determine the installation directory:
   - User: `~/.clap` (Linux), `~/Library/Audio/Plug-Ins/CLAP` (macOS)
   - System: `/usr/lib/clap` (Linux), `/Library/Audio/Plug-Ins/CLAP` (macOS)
3. Copy both the `.clap` file and shared library to the directory
4. Set appropriate permissions

### 6.2 Installation Script

The installation script handles the process:

```bash
#!/bin/bash
set -e

# Determine installation directory
if [ "$1" == "--system" ]; then
    install_dir="/usr/lib/clap"
    sudo mkdir -p $install_dir
else
    install_dir="$HOME/.clap"
    mkdir -p $install_dir
fi

# Build the plugins
mkdir -p build
cd build
cmake ..
cmake --build .

# Install plugins and shared libraries
if [ "$1" == "--system" ]; then
    sudo cmake --install .
else
    cmake --install .
fi

echo "Installation complete. Plugins installed to: $install_dir"
```

### 6.3 Cross-Platform Considerations

The build system handles platform-specific differences:

- **Linux**: 
  - Library name: `libgoclap.so`
  - Install location: `~/.clap` or `/usr/lib/clap`
  
- **macOS**:
  - Library name: `libgoclap.dylib`
  - Install location: `~/Library/Audio/Plug-Ins/CLAP` or `/Library/Audio/Plug-Ins/CLAP`
  - Plugins packaged as bundles

- **Windows**:
  - Library name: `goclap.dll`
  - Install location: `%APPDATA%\CLAP` or `C:\Program Files\Common Files\CLAP`

## 7. Conclusion

This build architecture provides a robust foundation for ClapGo, with a clean separation between the C wrapper and Go implementation. The use of dynamic loading rather than direct linking enables better error handling and more flexible deployment. The plugin registry allows for multiple plugins to be implemented in the same codebase, with each plugin maintaining ownership of its ID.

By using pure C for the CLAP interface, we avoid dependencies on C++ and simplify the compilation process. The CMake integration ensures proper building and packaging of both components, with platform-specific handling as needed.

This approach addresses the key issues identified in the current architecture while providing a clear path forward for plugin development.