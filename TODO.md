# ClapGo TODO: C Bridge Abstraction Tasks

This document identifies all areas where example plugins interact with C/CGO code and outlines tasks to abstract these interactions into the Go library for a cleaner plugin development experience.

## Current C Bridge Touchpoints in Examples

### 1. CGO Import and C Header Inclusion
**Location**: All example plugins (gain/main.go:3-21, synth/main.go:3-21)
**Current Implementation**:
```go
// #cgo CFLAGS: -I../../include/clap/include
// #include "../../include/clap/include/clap/clap.h"
// #include <stdlib.h>
// Helper functions for CLAP event handling...
```
**TODO**: Move all C imports and helper functions to the bridge package

### 2. Exported C Functions (Plugin Lifecycle)
**Location**: All plugins export 20+ C functions
**Functions**:
- `ClapGo_CreatePlugin`
- `ClapGo_GetVersion`
- `ClapGo_GetPluginID/Name/Vendor/Version/Description`
- `ClapGo_PluginInit/Destroy/Activate/Deactivate`
- `ClapGo_PluginStartProcessing/StopProcessing/Reset`
- `ClapGo_PluginProcess`
- `ClapGo_PluginGetExtension`
- `ClapGo_PluginOnMainThread`
- `ClapGo_PluginParamsCount/GetInfo/GetValue/ValueToText/TextToValue/Flush`
- `ClapGo_PluginStateSave/Load`

**TODO**: Generate these exports automatically in the bridge package

### 3. CGO Handle Management
**Location**: gain/main.go:49-50, 102, 112-114, etc.
**Current Implementation**:
```go
handle := cgo.NewHandle(gainPlugin)
p := cgo.Handle(plugin).Value().(*GainPlugin)
handle.Delete()
```
**TODO**: Abstract handle management into the bridge

### 4. C Type Conversions
**Location**: Throughout all examples
**Examples**:
- `C.bool()` conversions
- `C.uint32_t()` conversions
- `C.double()` conversions
- `C.CString()` and `C.GoString()` for strings
- `C.clap_param_info_t` struct manipulation
- `C.clap_process_t` struct access

**TODO**: Hide all C type conversions behind Go-native interfaces

### 5. Direct C Struct Manipulation
**Location**: gain/main.go:236-258 (parameter info)
**Current Implementation**:
```go
cInfo := (*C.clap_param_info_t)(info)
cInfo.id = C.clap_id(paramInfo.ID)
cInfo.flags = C.CLAP_PARAM_IS_AUTOMATABLE | C.CLAP_PARAM_IS_MODULATABLE
// Manual byte copying for strings...
```
**TODO**: Create Go structs that automatically marshal to C

### 6. Unsafe Pointer Usage
**Location**: Throughout all examples
**Uses**:
- Plugin handle passing
- Stream pointers for state save/load
- Extension pointers
- Process struct pointers
- Event handler pointers

**TODO**: Wrap all unsafe operations in safe Go APIs

### 7. Manual Memory Management
**Location**: String conversions, buffer management
**Current Issues**:
- `C.CString()` allocations without corresponding `C.free()`
- Manual buffer size calculations
- Direct memory access for audio buffers

**TODO**: Implement automatic memory management

### 8. Audio Buffer Conversion
**Location**: Already abstracted via `api.ConvertFromCBuffers()`
**Status**: ✅ Already abstracted properly

### 9. Event Processing
**Location**: Already abstracted via `api.EventHandler`
**Status**: ✅ Already abstracted properly

### 10. GUI C++ Integration (gain-with-gui)
**Location**: gui_bridge.cpp
**Current Implementation**:
- Separate C++ file required
- Complex GUI factory management
- Manual window API handling
- C++ class implementations

**TODO**: Create Go GUI abstraction that generates required C++ code

## Proposed Abstraction Architecture

### 1. Plugin Registration System
```go
// Instead of manual exports, plugins would:
type MyPlugin struct {
    clapgo.BasePlugin
}

func init() {
    clapgo.RegisterPlugin(&MyPlugin{})
}
```

### 2. Automatic Export Generation
- Build-time code generation for all required C exports
- Plugin developers never see C code
- All lifecycle methods handled automatically

### 3. Parameter System Enhancement
```go
// Current: Manual C struct manipulation
// Proposed:
plugin.RegisterParam(clapgo.FloatParam{
    ID: 1,
    Name: "Gain",
    Range: clapgo.Range{Min: 0.0, Max: 2.0},
    Default: 1.0,
})
```

### 4. State Management
```go
// Current: Manual stream handling with unsafe pointers
// Proposed:
func (p *MyPlugin) SaveState() ([]byte, error)
func (p *MyPlugin) LoadState(data []byte) error
```

### 5. GUI Framework
```go
// Proposed Go-only GUI API:
func (p *MyPlugin) CreateGUI() clapgo.GUI {
    return clapgo.NewGUI(
        clapgo.Size{Width: 400, Height: 300},
        clapgo.Resizable(true),
    )
}
```

## Implementation Priority

### Phase 1: Core Abstractions (High Priority)
1. [ ] Move all C imports to bridge package
2. [ ] Create automatic export generation
3. [ ] Abstract CGO handle management
4. [ ] Hide all C type conversions
5. [ ] Wrap unsafe pointer operations

### Phase 2: Enhanced Features (Medium Priority)
1. [ ] Implement missing CLAP extensions (see below)
2. [ ] Create GUI abstraction layer
3. [ ] Add automatic memory management
4. [ ] Improve state serialization

### Phase 3: Developer Experience (Lower Priority)
1. [ ] Plugin project generator
2. [ ] Hot reload support
3. [ ] Debugging tools
4. [ ] Performance profiling integration

## Missing CLAP Features

Currently, only 3 of 33+ CLAP extensions are available to Go developers:

### Core Extensions Not Available:
- **gui** - GUI support
- **note-ports** - MIDI/note input/output
- **latency** - Plugin latency reporting
- **tail** - Plugin tail length reporting
- **log** - Logging support
- **timer-support** - Timer/scheduling support
- **thread-check** - Thread safety checking
- **thread-pool** - Thread pool support
- **voice-info** - Voice/polyphony information
- **track-info** - DAW track information
- **preset-load** - Preset loading support
- **remote-controls** - Remote control pages
- **render** - Offline rendering support
- **posix-fd-support** - POSIX file descriptor support
- **event-registry** - Custom event type registration
- **param-indication** - Parameter automation indication
- **note-name** - Note naming support
- **context-menu** - Context menu support
- **state-context** - State save/load context
- **ambisonic** - Ambisonic audio support
- **surround** - Surround sound support
- **configurable-audio-ports** - Dynamic audio port configuration
- **audio-ports-config** - Audio port configuration
- **audio-ports-activation** - Audio port activation/deactivation

### Draft Extensions Not Available:
- **extensible-audio-ports** - Advanced audio port configuration
- **undo** - Undo/redo support
- **transport-control** - Transport control from plugin
- **scratch-memory** - Scratch memory allocation
- **triggers** - Trigger/articulation support
- **resource-directory** - Resource file management
- **tuning** - Microtuning support
- **project-location** - Project location information
- **mini-curve-display** - Mini automation curves
- **gain-adjustment-metering** - Gain reduction metering

## Success Criteria

The goal is achieved when plugin developers can write plugins using only Go code:
- No CGO imports in plugin code
- No C type conversions
- No unsafe pointer handling
- No manual memory management
- No exported C functions
- Automatic handling of all CLAP lifecycle events
- Full access to all CLAP features through Go APIs