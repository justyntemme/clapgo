# ClapGo Development Guardrails

Critical constraints for maintaining architectural integrity in ClapGo development.

## 🚫 Architecture Anti-Patterns (NEVER)

### 1. Go-Side Registration Systems
- No `RegisterPlugin()` functions or plugin registries in Go
- No discovery interfaces in Go code
- Use manifest files (JSON) + C bridge discovery only

### 2. "Simplified" APIs  
- No wrapper APIs that hide CLAP concepts
- No "easy" or "simple" interfaces
- Implement full CLAP `Plugin` interface always

### 3. Placeholder Code
- No TODO/FIXME comments
- No incomplete implementations with placeholder comments
- Either implement fully or return `nil` for unsupported extensions

### 4. Backwards Compatibility
- No API versioning for internal changes
- Make breaking changes without deprecation
- Delete old code entirely when refactoring

## ✅ Required Patterns

### 1. Manifest-Driven Discovery
```
plugin.clap
├── plugin.json      # Metadata manifest
├── libplugin.so     # Go shared library  
└── (C bridge)       # Handles CLAP interface
```

### 2. Standardized Go Exports
Every plugin must export:
- `ClapGo_CreatePlugin`
- `ClapGo_PluginInit`  
- `ClapGo_PluginProcess`
- `ClapGo_PluginGetExtension`
- All lifecycle functions

### 3. Complete Feature Implementation
- Implement extensions fully or return `nil`
- Use real `ParameterManager`, `EventHandler` interfaces
- No demo/example-only code

### 4. Code Deduplication
- Extract ALL common patterns to domain packages
- Plugin examples should be minimal - just plugin-specific logic
- Use composition and embedding for shared functionality
- If you write it twice, it belongs in a package
- Examples demonstrate usage, not implement functionality

## 🔒 Build & Development Standards

### Build System
- Use `make install` exclusively (never CMake)
- Test with `clap-validator`
- Focus on `gain` and `synth` examples only (no GUI)

### Code Quality
- No placeholder implementations
- Complete error handling (no silent failures)
- Thread-safe parameter access

### POC Development
- Breaking changes encouraged to find right architecture
- Update existing examples instead of creating new ones
- Delete old code entirely when refactoring

## 🎯 Architecture Goals

**Primary**: ClapGo is a bridge (not framework) enabling Go CLAP plugins  
**Secondary**: Zero plugin-specific C code required  
**Anti-Goal**: Hiding CLAP concepts from developers

## 📦 Package Organization Principles

### Domain-Driven Design
- Group related types and functions by domain (like Go's `net`, `io`, `http`)
- Keep types and their methods in the same package
- Avoid generic "helpers" or "utils" packages

### Example Structure
```
pkg/
├── param/        # Parameter domain
│   ├── param.go  # ParamInfo, ParamManager types
│   ├── value.go  # Value formatting, parsing
│   └── atomic.go # Thread-safe operations
├── event/        # Event domain
│   ├── event.go  # Event types and interfaces
│   ├── handler.go # Event handling
│   └── pool.go   # Event pooling
├── state/        # State domain
│   ├── state.go  # State types
│   ├── preset.go # Preset handling
│   └── migrate.go # Version migration
└── audio/        # Audio processing domain
    ├── buffer.go # Buffer management
    └── process.go # Processing utilities
```

### Naming Conventions
- Package names should be short, lowercase, singular nouns
- Avoid redundancy: `param.Info` not `param.ParamInfo`
- Functions should read naturally: `param.Format()` not `FormatParameter()`

## 🐹 Go Idiom Requirements

### Error Handling
- Return `error` not `bool` for operations that can fail
- Use custom error types with `Unwrap()` support
- Wrap errors with context: `fmt.Errorf("operation failed: %w", err)`

### Interface Design
- Keep interfaces small and focused (1-3 methods ideal)
- Accept interfaces, return concrete types
- Use standard library interfaces where applicable (`io.Reader`, `io.Writer`)

### API Design Patterns
- Use functional options for extensible APIs
- Builder pattern for complex struct creation
- Context support for cancellation/timeouts
- Method chaining where it improves readability

### Concurrency
- Protect shared state with appropriate synchronization
- Use channels for communication between goroutines
- Design APIs to be safe for concurrent use

## 🚨 Red Flags - Stop If You See

1. Creating interfaces that compete with CLAP
2. Adding Go registration when manifests exist  
3. Writing TODO comments
4. Worrying about backwards compatibility
5. Creating "easy" versions of real interfaces
6. Bypassing the C bridge
7. Duplicating code between examples
8. Implementing common functionality in examples instead of packages
9. Copy-pasting code instead of extracting to domain packages

**When in doubt**: Does this align with manifest-driven, C bridge architecture?

## ❌ Code Duplication Anti-Patterns

### NEVER duplicate these between plugins:
- Parameter creation/management boilerplate
- State saving/loading logic
- Event processing patterns
- Extension initialization
- Common DSP operations
- Error handling patterns
- Logging setup

### Examples should ONLY contain:
- Plugin-specific constants (ID, name, version)
- Plugin-specific parameter definitions
- Plugin-specific processing logic
- Plugin-specific state (if any)

### If you find yourself:
- Copy-pasting between examples → Extract to package
- Writing similar structures → Create base types
- Repeating initialization → Use composition
- Duplicating algorithms → Create utilities