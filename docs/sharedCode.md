# Shared Code Analysis

This document describes the current architecture of ClapGo and how code duplication is eliminated through the manifest-driven code generation system.

## Current Architecture (as of 2025)

ClapGo has evolved to use a **manifest-driven, code generation architecture** that eliminates most code duplication. The system consists of:

1. **C Bridge Layer** (`src/c/`) - Handles all CLAP-C interactions
2. **Go API Layer** (`pkg/api/`) - Provides abstractions for audio, events, parameters
3. **Code Generator** (`cmd/generate-manifest/`) - Generates boilerplate code
4. **Plugin Implementation** - Only contains plugin-specific logic

## Code Generation System

### Generated Files
The code generator creates four files that handle all boilerplate:

1. **exports_generated.go** - All CGO exports for plugin lifecycle
2. **extensions_generated.go** - Extension implementations based on plugin type
3. **constants_generated.go** - Plugin metadata and parameter definitions
4. **plugin.go** - Template with TODOs for developer implementation

### How It Works

The code generation is integrated into the Makefile workflow:

1. **Create a new plugin:**
   ```bash
   make new-plugin  # Interactive wizard
   # OR
   make generate-plugin NAME=my-effect TYPE=audio-effect ID=com.example.effect
   ```

2. **Build the plugin:**
   ```bash
   make build-plugin NAME=my-effect
   ```

This generates a complete plugin structure where developers only need to implement:
- `Process()` - Audio/MIDI processing logic
- `Init()` - Resource initialization
- Plugin-specific state management

## Eliminated Duplications

### 1. CGO Exports (✅ SOLVED)
**Previously duplicated in every plugin:**
```go
//export ClapGo_CreatePlugin
//export ClapGo_PluginInit
//export ClapGo_PluginDestroy
// ... 20+ more exports
```

**Current solution:**
All exports are generated in `exports_generated.go`. Developers never write CGO exports.

### 2. Parameter Management (✅ SOLVED)
**Previously duplicated:**
- Parameter registration
- Thread-safe access
- Value conversion

**Current solution:**
- `api.ParameterManager` handles all parameter operations
- Generated code registers parameters from manifest
- Thread-safe operations built into the API

### 3. Event Processing (✅ SOLVED)
**Previously duplicated:**
- Event iteration
- Type checking
- C struct parsing

**Current solution:**
- `api.EventHandler` abstracts all event processing
- Type-safe event structures
- No manual C struct manipulation

### 4. Audio Buffer Management (✅ SOLVED)
**Previously duplicated:**
- C buffer conversion
- Channel management
- Frame counting

**Current solution:**
- `api.Audio` struct with pre-converted buffers
- `api.ConvertFromCBuffers()` handles all conversions
- Clean Go slices for processing

### 5. State Management (✅ SOLVED)
**Previously duplicated:**
- Stream I/O operations
- Parameter serialization

**Current solution:**
- `api.InputStream`/`api.OutputStream` for type-safe I/O
- Generated state save/load exports
- Plugin only implements `GetState()`/`SetState()`

### 6. Extension Handling (✅ SOLVED)
**Previously duplicated:**
- Extension discovery
- Interface implementation

**Current solution:**
- Generated based on plugin type
- Automatic extension selection (e.g., instruments get note-ports)
- C bridge provides actual extension implementations

## Shared Library Components

### pkg/api Package
The API package provides all shared functionality:

- **audio.go** - Audio buffer abstractions
- **events.go** - Event handling abstractions
- **params.go** - Parameter management with thread safety
- **stream.go** - State serialization helpers
- **constants.go** - Common constants and types

### pkg/bridge Package
Handles C-Go interfacing:
- Safe memory management
- CGO handle management
- Pointer conversions

## Plugin Types and Their Generated Code

### Audio Effects
Generated extensions: params, state, audio-ports, latency, tail, render

### Instruments  
Generated extensions: params, state, audio-ports, note-ports, voice-info

### Note Effects
Generated extensions: params, state, note-ports, transport-control

### Analyzers
Generated extensions: params, state, audio-ports, thread-check

## Developer Workflow

1. **Create Plugin**
   ```bash
   make new-plugin  # Interactive wizard guides you through creation
   # OR
   make generate-plugin NAME=my-synth TYPE=instrument ID=com.example.synth
   ```
   Plugin is created in `plugins/my-synth/`

2. **Implement Logic**
   Only edit `plugin.go`:
   - Add DSP algorithms
   - Implement `Process()`
   - Add custom state

3. **Build**
   ```bash
   make build-plugin NAME=my-synth
   ```
   This simply:
   - Builds the Go shared library
   - Links with C bridge
   - Creates `my-synth.clap`
   
   (No code generation - that already happened during creation)

## Benefits of Current Architecture

1. **Zero CGO for Developers** - All CGO code is generated
2. **Type Safety** - Go interfaces instead of C structs
3. **Minimal Boilerplate** - ~60 lines vs ~600 lines
4. **Plugin Type Awareness** - Correct extensions for each type
5. **Consistent Implementation** - All plugins follow same pattern

## Future Improvements

1. **DSP Library** - Common effects as importable packages
2. **GUI Integration** - Generated GUI bridge code
3. **Preset Management** - Automatic preset discovery
4. **Hot Reload** - Development mode with live updates

## Migration from Old Pattern

For plugins using the old pattern (manual CGO exports):

1. Run generator with existing plugin ID
2. Move DSP logic to new `Process()` method
3. Delete old boilerplate files
4. Update parameter registration to use manifest

### Example: Generating the Gain Plugin

The gain plugin example can be completely generated using:
```bash
make generate-plugin NAME=gain TYPE=audio-effect ID=com.clapgo.gain
```

This creates a full plugin structure where the developer only needs to:
1. Add the gain parameter processing in `Process()`
2. Implement the atomic gain storage
3. Add parameter change handling

All CGO exports, extension handling, and boilerplate are automatically generated.

The current architecture successfully achieves the goal of hiding CGO complexity while maintaining full CLAP compliance.