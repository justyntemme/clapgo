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
- Extract common patterns to `pkg/api` helpers
- Minimize boilerplate in plugin examples
- Use composition for shared functionality

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

## 🚨 Red Flags - Stop If You See

1. Creating interfaces that compete with CLAP
2. Adding Go registration when manifests exist  
3. Writing TODO comments
4. Worrying about backwards compatibility
5. Creating "easy" versions of real interfaces
6. Bypassing the C bridge

**When in doubt**: Does this align with manifest-driven, C bridge architecture?