# ClapGo Implementation TODO

This document outlines the remaining CLAP extensions and functionality that need to be implemented in ClapGo's Go library to provide complete CLAP support to plugin developers.

### 🎯 Priority 2: Missing Factory Extensions

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

``

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
