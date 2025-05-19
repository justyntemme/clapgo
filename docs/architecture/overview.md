# ClapGo Architecture Overview

## High-Level Architecture

ClapGo enables Go developers to create [CLAP](https://github.com/free-audio/clap) audio plugins with a design that bridges the gap between C (required by the CLAP specification) and Go code.

![ClapGo Architecture Diagram](../images/architecture.png)

The architecture has three main layers:

1. **Go Plugin Layer**: Where developers implement their audio processing logic
2. **Go-C Bridge Layer**: Translates between Go and C using CGo
3. **C Plugin Layer**: Implements the CLAP plugin interface required by hosts

## Key Components

### 1. Go Plugin Implementation (`src/goclap/`)

The core Go package providing the API that plugin developers use.

- **plugin.go**: Defines the main `AudioProcessor` interface and plugin registry
- **events.go**: Handles CLAP events (note, parameter, MIDI)
- **params.go**: Parameter management
- **hostinfo.go**: Host communication
- **extensions/**.go: CLAP extension implementations

### 2. Go-C Bridge (`src/goclap/` exported functions)

Export Go functions for the C layer to call through CGo.

```go
//export GoProcess
func GoProcess(pluginPtr unsafe.Pointer, process *C.clap_process_t) C.clap_process_status {
    // Implementation that calls into the Go plugin's Process method
}
```

### 3. C Wrapper (`src/c/`)

CLAP-compatible C functions that hosts can call directly.

- **plugin.h/c**: C functions implementing the CLAP plugin interface
- **plugin_factory.h/c**: C factory implementation

## Data Flow

### Plugin Initialization
1. Host loads `.clap` shared library
2. Host calls C entry point functions
3. C layer forwards calls to the Go layer via exported functions
4. Go layer creates and initializes the plugin instance

### Audio Processing
1. Host calls C process function with audio buffers and events
2. C layer passes data to Go through exported functions
3. Go converts C data structures to Go types
4. Go plugin processes the audio
5. Go converts results back to C format
6. Results are returned to the host

### Parameter Changes
1. Host sends parameter events to the plugin
2. Events are converted to Go types
3. Go plugin updates its internal state
4. Parameter changes affect audio processing

## Threading Model

CLAP has specific threading requirements which ClapGo must respect:

1. **Audio Thread**: Processes audio and must be real-time safe
2. **Main Thread**: Handles UI updates and non-real-time operations
3. **Worker Threads**: Optional background processing

## Memory Management

Special care must be taken with memory management due to Go's garbage collector:

1. **Avoid allocations** in the audio processing path
2. Use **pre-allocated buffers** for audio data
3. Implement **proper object pooling** for events
4. Use **unsafe.Pointer** judiciously when interacting with C code

## Extension Mechanism

ClapGo supports CLAP extensions through a modular design:

1. Each extension is implemented in a separate Go file
2. Plugins can opt-in to specific extensions
3. Extension interfaces are exposed through the `GetExtension` mechanism

## Build System

The build process creates a shared library (.clap file) by:

1. Compiling Go code to a C-shared library
2. Compiling C wrapper code
3. Linking both together into a single shared library
4. Packaging according to the target platform requirements