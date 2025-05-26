# ClapGo Implementation TODO

This document outlines the remaining CLAP extensions and functionality that need to be implemented in ClapGo's Go library to provide complete CLAP support to plugin developers.

## 📊 Current Status Summary

After thorough review and implementation, ClapGo has achieved professional-grade real-time audio compliance with comprehensive CLAP functionality.

### ✅ Recently Completed:
- **Code Deduplication Phase 1**: Parameter management helpers (param_helpers.go)
- **Code Deduplication Phase 2**: Event processing patterns (event_helpers.go)

### ⚠️ Actually Missing/Incomplete:
- **Tuning Extension**: Host-side complete, missing plugin-side exports
- **Plugin Invalidation Factory**: Not implemented
- **Plugin State Converter Factory**: Not implemented
- **GUI Extension**: Forbidden per guardrails for example plugins
- **Undo Extension**: Not implemented (complex draft extension)
- **Other Draft Extensions**: Various experimental features

## 🚨 Priority 1: Complete Missing Plugin-Side Implementations

### 1. Tuning Extension (CLAP_EXT_TUNING)
**Status**: Partially Implemented
- Host-side implementation complete via HostTuning in pkg/api/tuning.go
- C bridge support exists but no Go exports implemented
- Missing plugin-side implementation (OnTuningChanged callback)
**Required**:
- Add ClapGo_PluginOnTuningChanged export
- Implement TuningProvider interface
- Add example usage in synth plugin

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

## 🧹 Priority 3: Continue Code Deduplication

### Phase 3: State Management Helpers
- JSON serialization/deserialization helpers
- Common state structures (presets, banks)
- Version migration utilities
- State validation helpers

### Phase 4: Audio Processing Utilities
- Buffer management helpers
- Common DSP operations (gain, pan, filters)
- Channel routing utilities
- Envelope generators

### Phase 5: Extension Boilerplate
- Common extension registration patterns
- Default extension implementations
- Extension capability helpers
- Manifest generation utilities

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

## 📋 Next Phase: Improving Usability

After completing the remaining technical work:

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
   - Create new markdown strategy document
   - Review all plugin types for templates
   - Ensure generated code matches examples exactly
   - Design for Go developer usability

## Development Guidelines

1. **Maintain Zero-Allocation Design**: All real-time paths must be allocation-free
2. **Follow Established Patterns**: Use existing extension implementations as reference
3. **Complete Features Only**: No placeholders or partial implementations
4. **Thread Safety**: All shared state must be properly synchronized
5. **Example Usage**: Each extension should have example usage in gain or synth plugins