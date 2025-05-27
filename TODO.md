# ClapGo Implementation TODO

This document outlines the remaining CLAP extensions and functionality that need to be implemented in ClapGo's Go library to provide complete CLAP support to plugin developers.

## 🏗️ CRITICAL ARCHITECTURE DECISION: C Export Requirements

**Important**: We previously attempted to remove all CGO code from plugin examples, but discovered that the ClapGo_* functions MUST be exported from the plugin's main package to be properly exposed in the shared library. This led to our hybrid approach:

1. **Plugin main.go**: Contains minimal CGO exports that call into pure Go packages
2. **Domain packages**: Pure Go implementation without CGO
3. **Code generation**: For repetitive C export boilerplate

This means our refactoring strategy must preserve the ability for plugins to export the required C functions while maximizing code reuse through domain packages.

## 📊 Current Status Summary

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

## 🧹 Priority 3: Domain-Driven Package Restructuring

### Phase 3A: Package Reorganization with Aggressive Deduplication
**Goal**: Refactor from generic "api" package to domain-specific packages while eliminating ALL duplication

**Architecture Constraint**: Plugins MUST continue to export ClapGo_* functions from their main package. Our approach:
1. Domain packages provide pure Go implementations
2. Plugin main.go contains minimal CGO exports that delegate to domain packages
3. Consider code generation for repetitive C export boilerplate

#### Identify and Move Plugin-Specific Logic
**Task**: Audit pkg/ directory for any plugin-specific logic that should be moved back to example plugins
- Review all pkg/ subdirectories for code that is specific to gain/synth examples
- Ensure pkg/ contains only reusable, generic CLAP functionality
- Move any plugin-specific DSP algorithms, parameter definitions, or business logic to respective example directories
- Maintain clear separation: pkg/ for framework/bridge code, examples/ for plugin-specific implementations

#### `pkg/param/` - Parameter Domain
```go
// Types move from api.ParamInfo → param.Info
// Functions like api.CreateVolumeParameter → param.Volume()
// api.FormatParameterValue → param.Format()
```
- Core types: `Info`, `Manager`, `Value`
- Atomic operations: `AtomicFloat64` with proper methods
- Formatting/parsing with `Formatter` interface
- Common parameter factories as methods
- **Deduplication targets**:
  - All parameter creation boilerplate
  - Value validation and clamping
  - Thread-safe update patterns
  - Text/value conversion logic

#### `pkg/event/` - Event Domain
```go
// api.EventHandler → event.Handler
// api.ProcessTypedEvents → handler.Process()
```
- Core types: `Event`, `Handler`, `Processor`
- Event pool management with proper lifecycle
- Type-safe event handling without allocations
- MIDI conversion utilities
- **Deduplication targets**:
  - Event processing loops
  - Type switching boilerplate
  - Event creation/queueing patterns
  - MIDI to note event conversion

#### `pkg/state/` - State Persistence Domain
```go
// api.StateManager → state.Manager
// Preset handling as first-class concept
```
- Core types: `State`, `Preset`, `Manager`
- Version migration framework
- JSON/Binary serialization with interfaces
- Validation with detailed errors
- **Deduplication targets**:
  - Save/Load implementation (currently in EVERY plugin)
  - Preset file handling
  - Version checking logic
  - Parameter state synchronization

#### `pkg/audio/` - Audio Processing Domain
```go
// New package for all DSP operations
```
- Buffer abstraction over `[][]float32`
- Zero-allocation processing utilities
- SIMD-optimized operations where beneficial
- Channel mapping and routing
- **Deduplication targets**:
  - Buffer clearing loops
  - Channel counting/validation
  - Gain application
  - Mix/sum operations

#### `pkg/host/` - Host Communication Domain
```go
// api.HostLogger → host.Logger
// All host extensions in one place
```
- Logger with structured logging support
- Extension interfaces properly grouped
- Thread checking utilities
- Host capability queries
- **Deduplication targets**:
  - Extension initialization patterns
  - Logger creation boilerplate
  - Thread check assertions
  - Host capability detection

#### `pkg/plugin/` - Plugin Core Domain
```go
// Base plugin types and lifecycle
```
- Plugin interface with proper error handling
- Lifecycle management
- Extension registration
- Manifest integration
- **Deduplication targets**:
  - ClapGo_* export functions (should be generated)
  - Plugin struct common fields
  - Initialization sequences
  - Extension lookup patterns

### Phase 3B: Go Idiom Refactoring
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

### Phase 3C: Deduplication Metrics
**Goal**: Dramatically reduce example code size

#### Success Metrics:
- Gain example: < 200 lines (currently ~1500)
- Synth example: < 400 lines (currently ~2000+)
- Zero duplicate code between examples
- All boilerplate in packages

#### Common Plugin Base:
```go
// pkg/plugin/base.go
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

### Phase 4: Audio Processing Domain
**Package**: `pkg/audio/`
- Buffer management with zero-allocation design
- Common DSP operations (gain, pan, filters) with SIMD optimization
- Channel routing and mapping utilities
- Envelope generators with proper interfaces
- Mix/sum/copy operations optimized for real-time
- **Example impact**: Remove ALL buffer loops from examples

### Phase 5: Enhanced Developer Experience
**Goal**: Make plugin development as simple as possible
- Plugin scaffolding generator using Go templates
- Comprehensive validation framework with detailed errors
- Testing utilities (mock host, event simulation, benchmarks)
- Performance profiling helpers integrated with pprof
- Debug mode with configurable logging levels

### Phase 6: Review and Remove Placeholder Implementations
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

## 🔄 Migration Strategy

### Package Refactoring Approach
1. **Create new packages first** - Don't break existing code
2. **Implement alongside api package** - Allow gradual migration
3. **Update examples one at a time** - Validate new design
4. **Deprecate api package last** - After all code migrated

### Example Migration Path
```go
// Step 1: Create new package
pkg/param/param.go

// Step 2: Implement with better API
func (m *Manager) Set(id uint32, value float64) error // returns error
func (m *Manager) Get(id uint32) (float64, error)     // explicit error

// Step 3: Update examples to use new packages
import "github.com/justyntemme/clapgo/pkg/param"
import "github.com/justyntemme/clapgo/pkg/event"

// Step 4: Remove old api package
```

### Target Example Structure (gain plugin after deduplication):
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

// That's it! ~30 lines instead of ~1500
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