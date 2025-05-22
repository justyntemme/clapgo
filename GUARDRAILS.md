# ClapGo Development Guardrails

This document provides critical constraints and guidelines for maintaining architectural integrity while developing ClapGo. These guardrails prevent common anti-patterns and ensure decisions align with our chosen manifest-driven, C bridge architecture.

## 🚫 Architecture Anti-Patterns to NEVER Implement

### 1. Registration Systems in Go
**❌ FORBIDDEN:**
```go
// NEVER create Go-side registration interfaces
type PluginRegistrar interface {
    Register(info PluginInfo, creator func() Plugin)
    GetPluginCount() uint32
    CreatePlugin(id string) Plugin
}

// NEVER create registration functions
func RegisterPlugin(plugin Plugin) { /* ... */ }
```

**✅ CORRECT APPROACH:**
- Plugin registration happens via **manifest files** (JSON)
- C bridge handles discovery via `manifest_find_files()`
- Plugins export standardized functions: `ClapGo_CreatePlugin`, `ClapGo_PluginInit`, etc.

**WHY:** Registration is the C bridge's responsibility. Go plugins are discovered via manifest system, not Go interfaces.

### 2. "Simplified" Examples or APIs
**❌ FORBIDDEN:**
```go
// NEVER create "simplified" or "easy" APIs that bypass the architecture
type SimplePlugin interface {
    ProcessSimple(input, output []float32) // Wrong abstraction level
}

// NEVER create wrapper APIs that hide important complexity
func EasyRegisterPlugin(name string, processFn func([]float32) []float32)
```

**✅ CORRECT APPROACH:**
- Implement full CLAP `Plugin` interface
- Use real parameter management via `ParameterManager`
- Handle actual CLAP events through `EventHandler`
- Expose extensions properly via `GetExtension()`

**WHY:** "Simplified" APIs hide critical CLAP concepts and prevent real-world usage. ClapGo is a bridge, not a simplification.

### 3. Placeholder Implementations
**❌ FORBIDDEN:**
```go
// NEVER leave placeholder implementations
func (p *Plugin) GetExtension(id string) unsafe.Pointer {
    return nil // TODO: Implement proper parameter extension
}

// NEVER use placeholder comments
// This is a placeholder for the actual CLAP API call
// In a real implementation we would...
// TODO: Implement this properly later
```

**✅ CORRECT APPROACH:**
- Either implement the extension fully or remove the comment
- If not implemented, return `nil` without TODO comments
- Complete features or mark them as explicitly unsupported

**WHY:** TODOs and placeholders indicate incomplete architecture. ClapGo is a real project, not a prototype.

### 4. Backwards Compatibility Concerns
**❌ FORBIDDEN:**
```go
// NEVER add backwards compatibility for deprecated patterns
// DeprecatedRegisterPlugin maintains old API for compatibility
func DeprecatedRegisterPlugin(plugin Plugin) {
    // Keep old behavior...
}

// NEVER version APIs to maintain old ways
type PluginV1 interface { /* old way */ }
type PluginV2 interface { /* new way */ }
```

**✅ CORRECT APPROACH:**
- Make breaking changes without deprecation periods
- Remove old interfaces entirely when refactoring
- Update examples to new patterns immediately
- No API versioning for internal changes

**WHY:** ClapGo is a POC/prototype. Breaking changes are encouraged to find the right architecture.

## ✅ Required Architecture Patterns

### 1. Manifest-Driven Plugin Discovery
**REQUIRED STRUCTURE:**
```
plugin.clap                    # Final CLAP plugin file
├── plugin.json               # Manifest with metadata
├── libplugin.so              # Go shared library
└── (C bridge linked in)      # src/c/ code
```

**REQUIRED FLOW:**
```
CLAP Host → C Bridge → Manifest Loading → Go Library → Plugin Instance
```

### 2. Standardized Go Exports
**REQUIRED EXPORTS (every plugin must have):**
```go
//export ClapGo_CreatePlugin
func ClapGo_CreatePlugin(host unsafe.Pointer, pluginID *C.char) unsafe.Pointer

//export ClapGo_PluginInit
func ClapGo_PluginInit(plugin unsafe.Pointer) C.bool

//export ClapGo_PluginProcess
func ClapGo_PluginProcess(plugin unsafe.Pointer, process unsafe.Pointer) C.int32_t

//export ClapGo_PluginGetExtension
func ClapGo_PluginGetExtension(plugin unsafe.Pointer, id *C.char) unsafe.Pointer

// ... (all other lifecycle exports)
```

### 3. Extension Implementation Pattern
**REQUIRED APPROACH:**
```go
func (p *MyPlugin) GetExtension(id string) unsafe.Pointer {
    switch id {
    case api.ExtParams:
        // Return actual C interface implementation
        return createParameterExtension(p)
    case api.ExtState:
        // Return actual C interface implementation  
        return createStateExtension(p)
    default:
        return nil // Extension not supported
    }
}
```

**NEVER return `nil` with TODO comments - either implement or don't support.**

### 4. Real Parameter Management
**REQUIRED USAGE:**
```go
// Use actual ParameterManager - never fake it
paramManager := api.NewParameterManager()
paramManager.RegisterParameter(api.CreateFloatParameter(0, "Gain", 0.0, 2.0, 1.0))

// Thread-safe parameter access
value := paramManager.GetParameterValue(paramID)
paramManager.SetParameterValue(paramID, newValue)
```

## 🔒 Code Quality Standards

### 1. No TODO/FIXME/Placeholder Comments
**❌ NEVER WRITE:**
```go
// TODO: Implement this later
// FIXME: This is broken
// XXX: Hack for now
// In a real app we would...
// This is a placeholder...
// Simplified version of...
```

**✅ ACCEPTABLE:**
```go
// Extension not yet supported
return nil

// Feature disabled in this version
if featureEnabled {
    // implementation
}
```

### 2. Complete Feature Implementation
**REQUIRED APPROACH:**
- Implement extensions fully or don't claim to support them
- Remove partial implementations that don't work
- No demo/example-only code that wouldn't work in production

### 3. Explicit Error Handling
**REQUIRED PATTERN:**
```go
// Clear error returns, no silent failures
func (p *Plugin) SetParameter(id uint32, value float64) error {
    if !p.isValidParameter(id) {
        return ErrInvalidParameter
    }
    // ... implementation
    return nil
}
```

## 📋 Development Checklist

Before implementing any feature, verify:

- [ ] **No Go-side registration** - uses manifest system only
- [ ] **No "simplified" APIs** - implements full CLAP interfaces  
- [ ] **No placeholder implementations** - complete or unsupported
- [ ] **No backwards compatibility** - breaking changes are encouraged
- [ ] **Follows manifest-driven architecture** - C bridge + JSON + Go exports
- [ ] **Uses standardized exports** - `ClapGo_*` function naming
- [ ] **Real extension implementations** - actual C interfaces or explicit nil
- [ ] **Thread-safe parameter management** - uses `ParameterManager`
- [ ] **Complete error handling** - no silent failures or undefined behavior

## 🎯 Architecture Goals Reinforcement

### Primary Goal: C Bridge + Manifest System
ClapGo is **NOT** a Go audio framework. It is a **bridge** that enables writing CLAP plugins in Go while maintaining full CLAP compliance.

### Secondary Goal: Zero Plugin-Specific C Code
Plugin developers write **only** Go code and JSON manifests. All CLAP interface complexity lives in the shared C bridge.

### Anti-Goal: Hiding CLAP Concepts
ClapGo does **NOT** simplify CLAP. It provides Go interfaces for CLAP concepts like parameters, extensions, and events.

## 🚨 Red Flags - Stop Development If You See

1. **Creating interfaces that compete with CLAP** - you're building the wrong abstraction
2. **Adding Go registration when manifests exist** - you're duplicating discovery mechanisms  
3. **Writing TODO comments** - you're not implementing complete features
4. **Worrying about backwards compatibility** - you're not embracing prototype development
5. **Creating "easy" or "simple" versions** - you're undermining the real architecture
6. **Bypassing the C bridge** - you're breaking the chosen design

## 💡 When In Doubt

**Ask: "Does this align with manifest-driven, C bridge architecture?"**

If the answer is no, don't implement it. If you're unsure, refer to:
1. `ARCHITECTURE.md` - for the intended design
2. `src/c/bridge.c` - for how the C bridge works  
3. `examples/gain/gain.json` - for how manifests define plugins
4. `examples/gain/main.go` - for proper Go plugin implementation

Remember: **ClapGo is a bridge, not a framework. Stay true to CLAP while enabling Go development.**