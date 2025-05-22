# ClapGo TODO: Move All CGO to Library

This document outlines the plan to move ALL CGO code from example plugins into the pkg/ library, eliminating boilerplate and making plugin development pure Go.

## Current State: Heavy CGO in Every Plugin

Currently, EVERY plugin must include:
- CGO imports and C headers (21 lines of C code)
- 20+ exported C functions (~500 lines of boilerplate)
- Manual C type conversions throughout
- Unsafe pointer handling everywhere
- Direct C struct manipulation

This results in:
- **90% boilerplate, 10% plugin logic**
- **Massive code duplication** across plugins
- **Error-prone** C type conversions
- **High barrier to entry** for Go developers

## Goal: Pure Go Plugin Development

Move ALL CGO code into the clapgo library so plugins become:
- **100% pure Go code**
- **No CGO imports**
- **No exported C functions**
- **No unsafe pointers**
- **Just implement an interface**

## Implementation Plan

### Phase 1: Create Plugin Interface & Base Implementation

1. **Define Complete Plugin Interface** in `pkg/clapgo/plugin.go`:
```go
type Plugin interface {
    // Lifecycle
    Init() bool
    Destroy()
    Activate(sampleRate float64, minFrames, maxFrames uint32) bool
    Deactivate()
    StartProcessing() bool
    StopProcessing()
    Reset()
    
    // Core processing
    Process(steadyTime int64, frames uint32, audioIn, audioOut [][]float32, events EventHandler) int
    
    // Extensions
    GetExtension(id string) unsafe.Pointer
    OnMainThread()
    
    // Parameters
    GetParameterCount() uint32
    GetParameterInfo(index uint32) (ParamInfo, error)
    GetParameterValue(id uint32) float64
    SetParameterValue(id uint32, value float64) error
    FormatParameterValue(id uint32, value float64) string
    ParseParameterValue(id uint32, text string) (float64, error)
    FlushParameters(events EventHandler)
    
    // State
    SaveState(stream OutputStream) bool
    LoadState(stream InputStream) bool
    
    // Metadata
    GetInfo() PluginInfo
}
```

2. **Create BasePlugin** with safe defaults:
```go
type BasePlugin struct{}

// All methods have safe default implementations
func (p *BasePlugin) Process(...) int {
    // Default: copy input to output
    for ch := range audioOut {
        copy(audioOut[ch], audioIn[ch])
    }
    return ProcessContinue
}
```

### Phase 2: Move All CGO Exports to Library

1. **Create `pkg/clapgo/exports.go`** containing ALL ClapGo_* functions:
```go
//export ClapGo_CreatePlugin
func ClapGo_CreatePlugin(host unsafe.Pointer, pluginID *C.char) unsafe.Pointer {
    id := C.GoString(pluginID)
    if CreatePluginFunc == nil {
        return nil
    }
    plugin := CreatePluginFunc(id)
    if plugin == nil {
        return nil
    }
    handle := cgo.NewHandle(plugin)
    return unsafe.Pointer(handle)
}

//export ClapGo_PluginInit
func ClapGo_PluginInit(plugin unsafe.Pointer) C.bool {
    if plugin == nil {
        return C.bool(false)
    }
    p := cgo.Handle(plugin).Value().(Plugin)
    return C.bool(p.Init())
}

// ... ALL 20+ other exports
```

2. **Single Factory Function** for developers to implement:
```go
// The ONLY thing developers need to provide
var CreatePluginFunc func(pluginID string) Plugin
```

### Phase 3: Refactor Parameter Handling

Move all C struct manipulation into library helpers:
```go
// In library - handles all C conversions
func convertParamInfoToC(info ParamInfo, cInfo *C.clap_param_info_t) {
    cInfo.id = C.clap_id(info.ID)
    cInfo.flags = C.uint32_t(info.Flags)
    // ... handle all string copying, etc.
}
```

### Phase 4: Simplify State Management

Abstract away stream handling:
```go
// Library provides wrapped stream types
type OutputStream interface {
    WriteUint32(v uint32) error
    WriteFloat64(v float64) error
    WriteBytes(b []byte) error
}

// Plugin just works with Go types
func (p *MyPlugin) SaveState(out OutputStream) bool {
    out.WriteUint32(1) // version
    out.WriteFloat64(p.gain)
    return true
}
```

## Result: Clean Plugin Code

A complete gain plugin becomes:
```go
package main

import (
    "github.com/justyntemme/clapgo/pkg/clapgo"
)

type GainPlugin struct {
    clapgo.BasePlugin  // Embed for defaults
    gain float64
}

func (p *GainPlugin) Process(steadyTime int64, frames uint32, audioIn, audioOut [][]float32, events clapgo.EventHandler) int {
    // Handle parameter changes
    for i := uint32(0); i < events.GetInputEventCount(); i++ {
        if event := events.GetInputEvent(i); event.Type == clapgo.EventTypeParamValue {
            if param, ok := event.Data.(clapgo.ParamEvent); ok && param.ParamID == 0 {
                p.gain = param.Value
            }
        }
    }
    
    // Process audio
    for ch := range audioOut {
        for i := range audioOut[ch] {
            audioOut[ch][i] = audioIn[ch][i] * float32(p.gain)
        }
    }
    
    return clapgo.ProcessContinue
}

func (p *GainPlugin) GetParameterCount() uint32 { return 1 }

func (p *GainPlugin) GetParameterInfo(index uint32) (clapgo.ParamInfo, error) {
    if index == 0 {
        return clapgo.ParamInfo{
            ID:      0,
            Name:    "Gain",
            Min:     0.0,
            Max:     2.0,
            Default: 1.0,
        }, nil
    }
    return clapgo.ParamInfo{}, clapgo.ErrInvalidParam
}

func init() {
    // The ONLY setup required
    clapgo.CreatePluginFunc = func(pluginID string) clapgo.Plugin {
        if pluginID == "com.example.gain" {
            return &GainPlugin{gain: 1.0}
        }
        return nil
    }
}

func main() {
    // Empty - required for c-shared build
}
```

## Benefits

1. **660+ lines → ~50 lines** for a basic plugin
2. **Zero CGO code** in plugins
3. **No manual memory management**
4. **Type-safe Go interfaces**
5. **Impossible to forget exports** - they're all in the library
6. **Easy to understand** - just implement methods
7. **Gradual implementation** - BasePlugin provides defaults

## Success Metrics

- Plugin code contains NO `import "C"`
- Plugin code has NO `//export` functions  
- Plugin code uses NO `unsafe` package
- All C type conversions happen in library
- Developers only work with Go types
- 90%+ reduction in boilerplate code