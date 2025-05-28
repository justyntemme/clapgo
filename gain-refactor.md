# Gain Plugin Refactor Implementation Strategy

## Overview

This document provides a comprehensive line-by-line strategy to refactor the gain plugin from **1466 lines** to approximately **200-300 lines** by fully utilizing the domain-specific packages in `pkg/*`.

## Current State Analysis

**File**: `examples/gain/main.go`  
**Current Lines**: 1466  
**Target Lines**: 200-300  
**Reduction Goal**: ~80-85%

### Line Count Breakdown by Section

1. **C Imports and Helpers** (Lines 1-22): 22 lines
2. **Go Imports** (Lines 23-37): 15 lines  
3. **Embedded Assets** (Lines 39-41): 3 lines
4. **Global Instance** (Lines 43-50): 8 lines
5. **Export Functions** (Lines 55-502): 447 lines
6. **Plugin Struct Definitions** (Lines 516-544): 29 lines
7. **Event Handler** (Lines 546-623): 78 lines
8. **Plugin Implementation** (Lines 625-1466): 841 lines

## Detailed Refactoring Strategy

### Phase 1: Remove Embedded Preset Loading (✅ READY TO IMPLEMENT)

**Current**: Lines 39-41, plus LoadPreset implementation around lines 1180-1220  
**Action**: Remove embedded filesystem and simplify preset loading  
**Savings**: ~45 lines

```go
// DELETE Lines 39-41:
//go:embed presets/factory/*.json
var factoryPresets embed.FS

// SIMPLIFY LoadPreset (Lines ~1180-1220):
// Remove all preset loading logic, return false
```

### Phase 2: Extract Plugin Base (MAJOR REFACTOR)

**Current Structure** (Lines 516-544):
```go
type GainPlugin struct {
    event.NoOpHandler     // Line 518
    sampleRate   float64  // Line 521
    isActivated  bool     // Line 522
    isProcessing bool     // Line 523
    host         unsafe.Pointer // Line 524
    gain         param.AtomicFloat64 // Line 527
    paramManager *param.Manager // Line 528
    stateManager *state.Manager // Line 531
    logger       *hostpkg.Logger // Line 534
    // ... more fields
}
```

**Target Structure**:
```go
type GainPlugin struct {
    *plugin.Base
    gain param.AtomicFloat64
}
```

**Implementation in pkg/plugin/base.go**:
```go
type Base struct {
    event.NoOpHandler
    Host         unsafe.Pointer
    SampleRate   float64
    IsActivated  bool
    IsProcessing bool
    ParamManager *param.Manager
    StateManager *state.Manager
    Logger       *host.Logger
}
```

**Savings**: ~25 lines in struct definition

### Phase 4: Remove Redundant Event Handler (Lines 546-623)

**Current**: 78 lines of mostly no-op event handlers  
**Action**: Since we embed `event.NoOpHandler`, remove the entire `gainEventHandler` type

**Only Keep** (move to GainPlugin):
```go
func (p *GainPlugin) HandleParamValue(e *api.ParamValueEvent, time uint32) {
    if e.ParamID == ParamGain {
        p.gain.Store(param.ClampValue(e.Value, 0.0, 2.0))
    }
}
```

**Savings**: ~70 lines

### Phase 5: Simplify Plugin Constructor (Lines 625-700)

**Current**: 75 lines of initialization  
**Target**: 10 lines

```go
func NewGainPlugin() *GainPlugin {
    p := &GainPlugin{
        Base: plugin.NewBase(PluginID, PluginName, PluginVersion),
    }
    
    p.ParamManager.Register(param.Info{
        ID: ParamGain, Name: "Gain", 
        DefaultValue: 1.0, MinValue: 0.0, MaxValue: 2.0,
        Flags: param.IsAutomatable | param.IsBounded,
    })
    
    return p
}
```

**Savings**: ~65 lines

### Phase 6: Use Domain Package Methods (Lines 700-1466)

#### 6A. Remove Process Helpers (Lines ~700-790)
**Current**: Custom audio processing loop  
**Replace With**: 
```go
func (p *GainPlugin) Process(...) int {
    return audio.ProcessStereo(audioIn, audioOut, func(sample float32) float32 {
        return sample * float32(p.gain.Load())
    })
}
```
**Savings**: ~80 lines

#### 6B. Remove GetExtension (Lines ~797-850)
**Current**: 53 lines of extension routing  
**Replace With**: `return p.Base.GetExtension(id)`  
**Savings**: ~50 lines

#### 6C. Remove Parameter Methods (Lines ~850-1000)
**Current**: GetParameterInfo, GetParameterValue, FormatParameterValue, etc.  
**Replace With**: Delegations to `p.ParamManager`  
**Savings**: ~150 lines

#### 6D. Remove State Methods (Lines ~1000-1080)
**Current**: Custom SaveState/LoadState  
**Replace With**:
```go
func (p *GainPlugin) SaveState(stream unsafe.Pointer) bool {
    return p.StateManager.Save(stream, map[uint32]float64{
        ParamGain: p.gain.Load(),
    })
}
```
**Savings**: ~60 lines

#### 6E. Remove Audio Port Methods (Lines ~1260-1300)
**Current**: 40 lines for standard stereo configuration  
**Replace With**: `audio.StereoPortProvider` embedding  
**Savings**: ~40 lines

#### 6F. Remove Thread Check Methods (Lines ~1300-1340)
**Current**: Custom thread checking  
**Replace With**: Base implementation  
**Savings**: ~40 lines

#### 6G. Remove Lifecycle Methods (Lines ~1340-1466)
**Current**: Init, Destroy, Activate, Deactivate, etc. with logging  
**Replace With**: Base implementations with hooks  
**Savings**: ~120 lines

### Phase 7: Final Structure

**Target main.go structure** (~200 lines):

```go
package main

import "C"
import (
    // minimal imports
)

// Constants from constants.go
var gainPlugin *GainPlugin

func init() {
    gainPlugin = NewGainPlugin()
}

// Single export dispatcher (~20 lines)
//export ClapGo_Dispatch
func ClapGo_Dispatch(...) { }

// Minimal plugin struct (~10 lines)
type GainPlugin struct {
    *plugin.Base
    gain param.AtomicFloat64
}

// Constructor (~10 lines)
func NewGainPlugin() *GainPlugin { }

// Plugin-specific logic only (~50 lines)
func (p *GainPlugin) Process(...) int { }
func (p *GainPlugin) HandleParamValue(...) { }
func (p *GainPlugin) GetPluginInfo() plugin.Info { }

// That's it! Everything else comes from domain packages
```

## Implementation Order

1. **Start with Phase 1**: Remove embedded presets (Easy win, -45 lines)
2. **Then Phase 4**: Remove redundant event handler (-70 lines)
3. **Then Phase 5**: Simplify constructor using domain packages (-65 lines)
4. **Then Phase 6B-6G**: Replace methods with domain package calls (-460 lines)
5. **Then Phase 3**: Extract plugin base (-25 lines)
6. **Finally Phase 2**: Consolidate exports (Most complex, -400 lines)

## Line-by-Line Actions

### Lines to DELETE entirely:
- 39-41: Embedded filesystem
- 546-623: Redundant event handler (except HandleParamValue)
- 700-790: Process helpers (replace with audio.ProcessStereo)
- 797-850: GetExtension (use Base)
- 850-1000: Parameter methods (use ParamManager)
- 1180-1220: Preset loading
- 1260-1300: Audio port methods (use audio.StereoPortProvider)

### Lines to REPLACE with one-liners:
- 625-700: NewGainPlugin → 10 lines
- 1000-1080: State methods → 5 lines each
- 1340-1466: Lifecycle methods → delegation to Base

### Lines to KEEP (plugin-specific only):
- Constants imports
- Process method (simplified)
- HandleParamValue (gain-specific logic)
- GetPluginInfo

## Expected Results

**Before**: 1466 lines  
**After**: ~200 lines  
**Reduction**: 1266 lines (86%)

**Code Distribution**:
- C exports: 20 lines
- Imports/package: 15 lines
- Plugin struct: 10 lines
- Constructor: 10 lines
- Plugin-specific methods: 50 lines
- Glue code: 95 lines
- **Total**: ~200 lines

## Success Metrics

1. `make install` succeeds
2. `clap-validator` maintains 17/21 pass rate
3. No functional regressions
4. All domain packages properly utilized
5. No code duplication with domain packages

---

**Note**: This is a working document. Update line numbers after each phase as the file shrinks.
