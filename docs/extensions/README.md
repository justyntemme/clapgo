# CLAP Extensions

This document outlines the CLAP extensions that ClapGo plans to support and their implementation status.

## Core Extensions

| Extension | Status | Implementation Priority |
|-----------|--------|-------------------------|
| [Audio Ports](#audio-ports) | Planned | High |
| [Parameters](#parameters) | Partially Implemented | High |
| [Note Ports](#note-ports) | Planned | Medium |
| [Latency](#latency) | Planned | Medium |
| [State](#state) | Planned | Medium |
| [GUI](#gui) | Planned | Low |

## Extension Details

### Audio Ports

**Implementation File**: `src/goclap/extensions/audio_ports.go`

Allows a plugin to:
- Describe its audio input/output ports
- Handle port configuration changes
- Support for different channel configurations

Required interfaces:
- `clap_plugin_audio_ports`
- `clap_host_audio_ports`

### Parameters

**Implementation File**: `src/goclap/extensions/params.go`

Provides parameter management for plugins:
- Parameter definitions (ranges, flags, etc.)
- Parameter value changes
- Parameter modulation
- Parameter automation

Required interfaces:
- `clap_plugin_params`
- `clap_host_params`

### Note Ports

**Implementation File**: `src/goclap/extensions/note_ports.go`

Describes MIDI/note input and output ports:
- Note input/output port configuration
- Support for polyphonic expressions

Required interfaces:
- `clap_plugin_note_ports`
- `clap_host_note_ports`

### Latency

**Implementation File**: `src/goclap/extensions/latency.go`

Handles plugin processing latency:
- Report plugin latency to host
- Handle latency changes

Required interfaces:
- `clap_plugin_latency`
- `clap_host_latency`

### State

**Implementation File**: `src/goclap/extensions/state.go`

Manages plugin state saving and loading:
- Save plugin state
- Load plugin state
- Handle preset management

Required interfaces:
- `clap_plugin_state`
- `clap_host_state`

### GUI

**Implementation File**: `src/goclap/extensions/gui.go`

Provides plugin GUI integration:
- Create native windows
- Handle window embedding
- Platform-specific window handling

Required interfaces:
- `clap_plugin_gui`
- `clap_host_gui`

## Additional Extensions

These extensions will be implemented after the core extensions:

| Extension | Status | Priority |
|-----------|--------|----------|
| Note Name | Planned | Low |
| Thread Check | Planned | Medium |
| Thread Pool | Planned | Low |
| Timer Support | Planned | Low |
| Voice Info | Planned | Low |
| Tail | Planned | Medium |
| Render | Planned | Low |
| Audio Ports Config | Planned | Low |

## Implementation Strategy

Each extension will be implemented in its own Go file with a consistent pattern:

1. Define Go interfaces for the extension
2. Create Go wrappers for C structures
3. Implement CGo functions for C callbacks
4. Create helper methods for Go code to use

Example pattern for an extension:

```go
package extensions

// #include <stdlib.h>
// #include "../../../include/clap/include/clap/clap.h"
// #include "../../../include/clap/include/clap/ext/my_extension.h"
import "C"
import "unsafe"

// MyExtension interface that plugins can implement
type MyExtension interface {
    // Various methods...
}

// hostMyExtension is a wrapper for the host extension
type hostMyExtension struct {
    ptr unsafe.Pointer
}

// pluginMyExtension is the implementation for plugins
type pluginMyExtension struct {
    plugin MyExtension
}

// Register exports C functions that the host can call
func RegisterMyExtension(plugin MyExtension) unsafe.Pointer {
    // Implementation
}
```

## Testing Extensions

Each extension will include tests to verify:
1. Go to C data conversion
2. C to Go data conversion
3. Callback mechanisms
4. Extension-specific functionality