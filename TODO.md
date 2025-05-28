# ClapGo Implementation TODO

This document outlines the remaining CLAP extensions and functionality that need to be implemented in ClapGo's Go library to provide complete CLAP support to plugin developers.

## 🔧 IMMEDIATE: Final pkg/api Migration Cleanup

**Status**: 🚧 In Progress  
**Goal**: Complete the refactoring by migrating remaining pkg/api usage to proper domain packages

### Remaining pkg/api Usage (excluding generate-manifest, gain-with-gui, generate_exports.go):
1. **examples/synth/main.go** (2 occurrences)
   - `api.DebugMarkAudioThread()` / `api.DebugUnmarkAudioThread()` → migrate to `pkg/thread`
   - `api.ConvertFromCBuffers()` → migrate to `pkg/audio`
   
2. **examples/gain/exports.go** (1 occurrence)
   - Uses `api` package - needs migration to proper domain package exports

3. **pkg/plugin/exports.go** (1 occurrence)
   - `api.UnregisterAudioPortsProvider()` → migrate to `pkg/extension` or `pkg/audio`

4. **pkg/plugin/base.go** (1 occurrence)
   - Various `api` references → complete migration to domain packages

5. **pkg/clap/params.go** (1 occurrence)
   - `api.ParamInfo` and related types → already have `pkg/param`, complete migration

6. **pkg/clap/base.go** (1 occurrence)
   - `api.Plugin` interface and related → migrate to `pkg/plugin` interface

### Migration Tasks:
- [ ] Move thread marking functions from pkg/api to pkg/thread
- [ ] Move audio buffer conversion from pkg/api to pkg/audio  
- [ ] Update audio ports provider registration to use extension package
- [ ] Complete migration of param types in pkg/clap to use pkg/param
- [ ] Define proper plugin interface in pkg/plugin instead of pkg/api
- [ ] Update all imports and references in affected files
- [ ] Ensure no functionality is lost during migration
- [ ] Run tests to verify everything still works

**Note**: This is the final cleanup to complete the domain-driven refactoring. Once done, pkg/api should only contain minimal C interop helpers that don't fit into domain packages.

## 🏗️ CRITICAL ARCHITECTURE DECISION: C Export Requirements

**Important**: We previously attempted to remove all CGO code from plugin examples, but discovered that the ClapGo_* functions MUST be exported from the plugin's main package to be properly exposed in the shared library. This led to our hybrid approach:

1. **Plugin main.go**: Contains minimal CGO exports that call into pure Go packages
2. **Domain packages**: Pure Go implementation without CGO
3. **Code generation**: For repetitive C export boilerplate

This means our refactoring strategy must preserve the ability for plugins to export the required C functions while maximizing code reuse through domain packages.

## 📊 Current Status Summary

### ✅ COMPLETED (Phase 3A - Domain Package Restructuring)
- **✅ pkg/param/** - Parameter domain with atomic operations, formatting, validation
- **✅ pkg/event/** - Event processing with type-safe handlers and zero-allocation design
- **✅ pkg/host/** - Host communication (Logger, TrackInfo, TransportControl, etc.)
- **✅ pkg/thread/** - Thread checking and validation utilities
- **✅ pkg/extension/** - Extension constants and provider interfaces
- **✅ pkg/controls/** - Remote controls functionality
- **✅ pkg/audio/** - Audio buffer management and DSP utilities
- **✅ pkg/state/** - State persistence and preset handling
- **✅ Duplicate code elimination** - Removed 15+ duplicate api files
- **✅ Both example plugins compile** - gain and synth plugins build successfully
- **✅ make install works** - Full build system integration

### ⚠️ Actually Missing/Incomplete:
- **Plugin Invalidation Factory**: Not implemented
- **Plugin State Converter Factory**: Not implemented  
- **GUI Extension**: Forbidden per guardrails for example plugins
- **Undo Extension**: Not implemented (complex draft extension)
- **Other Draft Extensions**: Various experimental features

## 🎯 Priority 2: Missing Factory Extensions

### 1. Plugin Invalidation Factory (CLAP_PLUGIN_INVALIDATION_FACTORY)
**Status**: ❌ Not implemented
**Purpose**: Notify when plugins become invalid/outdated
**Required**:
- Factory implementation in C bridge
- Invalidation source management
- Host notification system

### 2. Plugin State Converter Factory (CLAP_PLUGIN_STATE_CONVERTER_FACTORY)
**Status**: ❌ Not implemented
**Purpose**: Convert between state format versions
**Required**:
- State version detection
- Conversion logic framework
- Migration utilities

## 🎯 Priority 3: Phase 3B - Go Idiom Refactoring
**Goal**: Make the API more idiomatic for Go developers

#### Error Handling Reform
```go
// Before: func (p *Plugin) Init() bool
// After:  func (p *Plugin) Init() error

// Custom error types
type ParamError struct {
    Op      string
    ParamID uint32
    Err     error
}
```

#### Functional Options Pattern
```go
// Plugin creation with options
plugin, err := plugin.New(
    plugin.WithName("My Synth"),
    plugin.WithParameter(param.Volume(0, "Master")),
    plugin.WithExtension(ext.StateContext()),
)
```

#### Context Support
```go
// All long operations accept context
err := plugin.Process(ctx, &processData)
err = state.Save(ctx, writer, pluginState)
```

#### Builder Pattern for Complex Types
```go
// Parameter builder
p := param.NewBuilder(0, "Gain").
    Range(0, 2, 1).
    Flags(param.Automatable | param.Modulatable).
    Format(param.Decibel).
    Build()
```

#### Interface Improvements
```go
// Small, focused interfaces
type Processor interface {
    Process(ctx context.Context, in, out audio.Buffer) error
}

type Stateful interface {
    SaveState(w io.Writer) error
    LoadState(r io.Reader) error
}
```

### Next: Further Example Code Reduction
**Goal**: Dramatically reduce example code size to < 200 lines for gain, < 400 for synth

#### Common Plugin Base Enhancement:
```go
// pkg/plugin/base.go - Further consolidate common functionality
type Base struct {
    param.Manager
    state.Manager
    host.Logger
    // ... other common fields
}

// Handles ALL common exports and lifecycle
func (b *Base) Init() error { ... }
func (b *Base) Process(ctx context.Context, data *ProcessData) error { ... }
```

### Future: Enhanced Developer Experience
**Goal**: Make plugin development as simple as possible
- Plugin scaffolding generator using Go templates
- Comprehensive validation framework with detailed errors
- Testing utilities (mock host, event simulation, benchmarks)
- Performance profiling helpers integrated with pprof
- Debug mode with configurable logging levels

### Future: Review and Remove Placeholder Implementations
- Search for placeholder implementations that only contain structs and return nil
- Look for "for now" comments indicating temporary/fake success returns
- Replace all placeholder implementations with proper functionality or remove entirely
- Ensure no silent failures or fake success returns exist
- Audit all error paths for proper error propagation

## 🔮 Priority 4: Draft/Experimental Extensions

### 1. Undo (CLAP_EXT_UNDO)
**Status**: ❌ Not Implemented
**Purpose**: Integrate with host undo system
**Features Needed**:
- Begin/cancel/complete change tracking
- Delta-based undo/redo
- Context updates (can undo/redo, step names)

### 2. Triggers (CLAP_EXT_TRIGGERS)
**Purpose**: Trigger/gate functionality
**Features**:
- Trigger registration
- Trigger state changes
- MIDI mapping

### 3. Resource Directory (CLAP_EXT_RESOURCE_DIRECTORY)
**Purpose**: Access plugin resources
**Features**:
- Resource file paths
- Shared resource access
- Platform-specific paths

### 4. Scratch Memory (CLAP_EXT_SCRATCH_MEMORY)
**Purpose**: Temporary memory from host
**Features**:
- Pre-allocated buffers
- Zero-allocation processing
- Size negotiation

### 5. Project Location (CLAP_EXT_PROJECT_LOCATION)
**Purpose**: Get project file information
**Features**:
- Project path
- Project name
- Relative file resolution

### 6. Extensible Audio Ports (CLAP_EXT_EXTENSIBLE_AUDIO_PORTS)
**Purpose**: Dynamic audio port configurations
**Features**:
- Add/remove ports dynamically
- Complex routing scenarios

### 7. Gain Adjustment Metering (CLAP_EXT_GAIN_ADJUSTMENT_METERING)
**Purpose**: Report gain reduction (compressors/limiters)
**Features**:
- Real-time gain reporting
- VU meter data

### 8. Mini Curve Display (CLAP_EXT_MINI_CURVE_DISPLAY)
**Purpose**: Display parameter automation
**Features**:
- Curve rendering
- Automation feedback

## 🎯 Current Status: Core Refactoring COMPLETE

### ✅ Successfully Migrated From api Package:
- **pkg/param/** - Parameter management with atomic operations
- **pkg/event/** - Event processing with zero allocations  
- **pkg/host/** - Host communication and utilities
- **pkg/thread/** - Thread checking and validation
- **pkg/extension/** - Extension interfaces and constants
- **pkg/controls/** - Remote controls functionality
- **pkg/audio/** - Audio buffer management
- **pkg/state/** - State persistence and presets

### ✅ Architecture Achievements:
- Both gain and synth plugins compile successfully
- `make install` works - full build system integration
- Removed 15+ duplicate files from pkg/api
- Domain-driven package organization implemented
- No backwards compatibility maintained (per guardrails)

### Target Example Structure (Future Goal):
```go
package main

import (
    "github.com/justyntemme/clapgo/pkg/plugin"
    "github.com/justyntemme/clapgo/pkg/param"
    "github.com/justyntemme/clapgo/pkg/audio"
)

const (
    PluginID = "com.example.gain"
    PluginName = "Gain"
)

type GainPlugin struct {
    plugin.Base
    gain param.AtomicFloat64
}

func NewGainPlugin() *GainPlugin {
    p := &GainPlugin{}
    p.Base.Init(
        plugin.WithID(PluginID),
        plugin.WithName(PluginName),
        plugin.WithParameter(param.Volume(0, "Gain")),
    )
    return p
}

func (p *GainPlugin) Process(ctx context.Context, in, out audio.Buffer) error {
    gain := p.gain.Load()
    return audio.ApplyGain(out, in, gain)
}

// Target: ~30 lines instead of current ~1500
```

## 📋 Next Phase: Improving Usability

After completing the domain restructuring:

1. **Document the Makefile and Build System**
   - Explain each module's build process
   - Document linking strategy
   - Create build troubleshooting guide

2. **Library Architecture Review**
   - Identify where C code can be better abstracted
   - Balance abstraction with required Go exports
   - Create architectural decision records

3. **Plugin Generator System**
   - Delete old go-generate system
   - Create new Go template-based generator
   - `clapgo new plugin --type=effect --name=MyPlugin`
   - Generates idiomatic code using new packages
   - Includes comprehensive examples and tests

## Development Guidelines

1. **Maintain Zero-Allocation Design**: All real-time paths must be allocation-free
2. **Follow Established Patterns**: Use existing extension implementations as reference
3. **Complete Features Only**: No placeholders or partial implementations
4. **Thread Safety**: All shared state must be properly synchronized
5. **Example Usage**: Each extension should have example usage in gain or synth plugins

## ⚠️ Refactoring Anti-Patterns to Avoid

1. **Don't hide CLAP concepts** - We're a bridge, not a framework
2. **Don't over-abstract** - Keep it simple and direct
3. **Don't break manifest system** - Maintain C export requirements
4. **Don't create competing APIs** - Enhance, don't replace CLAP
5. **Don't sacrifice performance** - Real-time constraints are paramount
6. **Don't use reflection** - Keep everything compile-time safe
7. **Don't ignore C interop needs** - Exports must remain compatible

## 🔥 Code Patterns We're Eliminating

### Current Anti-Patterns in Examples:
1. **1500+ line plugin files** → Should be < 200 lines
2. **Manual event type switching** → Use typed handlers
3. **Copy-pasted state handling** → Use state.Manager
4. **Duplicate parameter code** → Use param.Manager
5. **Raw unsafe.Pointer everywhere** → Wrapped in safe types
6. **Boilerplate exports** → Generated or in base class
7. **Manual buffer loops** → audio.Buffer methods
8. **Inline DSP math** → audio package functions

### The 80/20 Rule:
- 80% of plugin code should be in packages
- 20% should be plugin-specific logic
- If it's not unique to your plugin, it belongs in a package