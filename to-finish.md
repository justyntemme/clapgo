# ClapGo To-Finish List - Segfault Fixes

This document lists all the unfinished code sections that are causing segfaults when the CLAP validator tries to call functions.

## Critical Segfault Locations

### 1. Extension Functions Returning NULL
**Location**: src/c/bridge.c:456-489
**Issue**: `clapgo_plugin_get_extension` returns NULL for all extensions
**Impact**: Validator crashes when trying to access extension functions

Required extensions to implement (following GUARDRAILS - complete implementation only):
- **CLAP_EXT_PARAMS** (src/c/bridge.c:465-469)
  - Need to implement full params extension with:
    - `params_count()`
    - `params_info()`
    - `params_value()`
    - `params_value_to_text()`
    - `params_text_to_value()`
    - `params_flush()`
  - **GUARDRAILS**: Must be fully functional, no placeholders
- **CLAP_EXT_STATE** (src/c/bridge.c:472-476)
  - Need to implement state save/load:
    - `state_save()`
    - `state_load()`
  - **GUARDRAILS**: Complete implementation or remove entirely
- **CLAP_EXT_AUDIO_PORTS** (src/c/bridge.c:479-483)
  - Need to implement audio ports extension:
    - Already have `clapgo_audio_ports_count()` and `clapgo_audio_ports_info()`
    - Need to create the extension structure and return it
  - **GUARDRAILS**: This is closest to complete - finish it first

### 2. Go-side Extension Implementation
**Location**: examples/gain/main.go:378-387
**Issue**: `GetExtension` returns nil without proper implementation
**Impact**: Extensions are never provided to the host
**Guardrails Note**: Must either implement extensions fully or explicitly mark as unsupported - no TODO comments allowed

### 3. Missing Extension Structures
**Location**: Not implemented yet
**Issue**: Need to create C structures for each extension that contain function pointers
**Impact**: Even if we return non-NULL, there's no structure to return

## Implementation Plan

### Phase 1: Audio Ports Extension (Easiest)
1. Create `clap_plugin_audio_ports_t` structure in bridge.c
2. Wire up existing `clapgo_audio_ports_count()` and `clapgo_audio_ports_info()` functions
3. Return the structure from `clapgo_plugin_get_extension()`

### Phase 2: Params Extension (Most Critical)
1. Create `clap_plugin_params_t` structure in bridge.c
2. Implement all params callbacks:
   - `clapgo_params_count()`
   - `clapgo_params_info()`
   - `clapgo_params_value()`
   - `clapgo_params_value_to_text()`
   - `clapgo_params_text_to_value()`
   - `clapgo_params_flush()`
3. Create Go-side interface for parameter management
4. Update example plugins to provide parameter information

### Phase 3: State Extension
1. Create `clap_plugin_state_t` structure in bridge.c
2. Implement state callbacks:
   - `clapgo_state_save()`
   - `clapgo_state_load()`
3. Create Go-side interface for state management
4. Update example plugins to support state save/load

### Phase 4: Additional Extensions
Based on validator requirements, may need:
- Note ports extension
- Latency extension
- GUI extension (if validator tests it)

## Code Locations to Fix

1. **src/c/bridge.c:456-489** - `clapgo_plugin_get_extension()`
   - Remove the "return NULL" statements
   - Return proper extension structures

2. **src/c/bridge.c** - Add extension structures after line 500:
   ```c
   // Extension structures
   static const clap_plugin_audio_ports_t audio_ports_ext = {
       .count = clapgo_audio_ports_count,
       .get = clapgo_audio_ports_info
   };
   
   static const clap_plugin_params_t params_ext = {
       .count = clapgo_params_count,
       .get_info = clapgo_params_info,
       .get_value = clapgo_params_value,
       .value_to_text = clapgo_params_value_to_text,
       .text_to_value = clapgo_params_text_to_value,
       .flush = clapgo_params_flush
   };
   ```

3. **examples/gain/main.go:194-201** - `ClapGo_PluginGetExtension`
   - Implement actual extension support
   - Return proper Go structures that can be converted to C

4. **pkg/api/extensions.go** - Create new file
   - Define Go interfaces for each extension
   - Provide conversion helpers between Go and C

## Testing Strategy

1. Start with audio ports extension (already partially implemented)
2. Test with clap-validator after each extension is added
3. Use clap-host or other test hosts to verify functionality
4. Add unit tests for each extension implementation

## Notes

- The validator is crashing because it's trying to call functions through NULL pointers
- Most hosts will query for extensions during plugin initialization
- Extensions are optional, but returning NULL and then crashing suggests incomplete implementation
- The manifest system is working correctly, but the runtime extension system needs completion

## GUARDRAILS Compliance

Per GUARDRAILS.md, this implementation must:
1. **NO placeholder implementations** - Each extension must be fully functional or not claimed
2. **NO TODO comments** - Remove all TODO/FIXME comments from the code
3. **Complete feature implementation** - Extensions work fully or are explicitly unsupported
4. **Real parameter management** - Use the existing ParameterManager properly
5. **Thread-safe implementation** - Ensure all extension callbacks are thread-safe
6. **No "simplified" APIs** - Implement full CLAP interfaces, not shortcuts