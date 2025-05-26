# ClapGo Implementation TODO

This document outlines the remaining CLAP extensions and functionality that need to be implemented in ClapGo's Go library to provide complete CLAP support to plugin developers.

## Executive Summary

ClapGo has achieved professional-grade real-time audio compliance with a solid foundation of core CLAP functionality. The remaining work focuses on implementing additional CLAP extensions that are currently missing from our Go API layer.

### ✅ What's Already Complete
- **Core Plugin System**: Full CLAP compliance with manifest-driven architecture
- **Zero-Allocation Processing**: Event pools, fixed arrays, buffer pools
- **Essential Extensions**: Params, Note Ports, Context Menu, Remote Controls, Param Indication, Note Name
- **Advanced Audio**: Ambisonic, Audio Ports Activation, Configurable Audio Ports
- **Core Utilities**: Event Registry, POSIX FD Support, Render, Thread Pool
- **Partial Implementations**: Basic interfaces exist for Audio Ports, State, Latency, Tail, etc.

## 🚨 Priority 1: Complete Partial Implementations

These extensions have basic interfaces defined but need full implementation:
### 1. Audio Ports Extension (CLAP_EXT_AUDIO_PORTS)
**Status**: ✅ Implemented
- Full Go exports and C bridge support
- Port type constants and flags
- In-place processing pair support
- Channel mapping functionality

### 2. State Extension (CLAP_EXT_STATE)
**Current State**: `StateProvider` interface exists
**Missing**:
- Complete stream wrapper implementation
- Proper error handling and validation
- Integration with host state management
- Example implementations in gain/synth plugins

### 3. State Context Extension (CLAP_EXT_STATE_CONTEXT)
**Current State**: `StateContextProvider` interface exists
**Missing**:
- Context-aware save/load implementation
- Proper handling of preset vs project vs duplicate contexts
- Integration with state extension

### 4. Latency Extension (CLAP_EXT_LATENCY)
**Current State**: `LatencyProvider` interface exists
**Missing**:
- Host callback for latency changes
- C bridge implementation
- Example usage in plugins

### 5. Tail Extension (CLAP_EXT_TAIL)
**Current State**: `TailProvider` interface exists
**Missing**:
- Host callback for tail changes
- C bridge implementation
- Example usage in reverb/delay scenarios

### 6. Log Extension (CLAP_EXT_LOG)
**Current State**: `LogProvider` interface exists
**Missing**:
- Complete host-side implementation
- Thread-safe logging
- Integration with Go's standard logging

### 7. Timer Support Extension (CLAP_EXT_TIMER_SUPPORT)
**Current State**: `TimerSupportProvider` interface exists
**Missing**:
- Timer registration/unregistration
- C bridge for timer callbacks
- Thread-safe timer management

### 8. Preset Load Extension (CLAP_EXT_PRESET_LOAD)
**Current State**: `PresetLoader` interface exists
**Missing**:
- File loading implementation
- Integration with preset discovery
- Bundled preset support

### 9. Audio Ports Config Extension (CLAP_EXT_AUDIO_PORTS_CONFIG)
**Status**: ✅ Implemented
- Complete configuration switching with Go exports
- Host callbacks for config changes
- C bridge implementation

### 10. Surround Extension (CLAP_EXT_SURROUND)
**Status**: ✅ Implemented
- Complete channel mapping implementation
- Go exports with channel mask support
- Helper functions for standard surround formats

### 11. Voice Info Extension (CLAP_EXT_VOICE_INFO)
**Status**: ✅ Implemented
- Complete Go exports and registry
- VoiceManager helper for voice tracking
- Integration with synth example

### 12. Track Info Extension (CLAP_EXT_TRACK_INFO)
**Current State**: `TrackInfoProvider` interface exists
**Missing**:
- Host-side track info provision
- Proper color and name handling
- Track type detection

## 🎯 Priority 2: Missing Core Extensions

These extensions are completely missing and should be implemented:

### 1. GUI Extension (CLAP_EXT_GUI) - Will not DO 
**Purpose**: Plugin GUI creation and management
*Note**: Implementation forbidden per guardrails for example plugins

## 🔮 Priority 3: Draft/Experimental Extensions

These are experimental CLAP extensions that could be valuable:

### 1. Transport Control (CLAP_EXT_TRANSPORT_CONTROL)
**Purpose**: Let plugins control host transport
**Features**:
- Play/pause/stop control
- Jump to position
- Loop control
- Tempo changes

### 2. Tuning (CLAP_EXT_TUNING)
**Purpose**: Microtuning and alternative temperaments
**Features**:
- Load tuning tables
- Dynamic tuning changes
- Note-specific tuning

### 3. Undo (CLAP_EXT_UNDO)
**Purpose**: Integrate with host undo system
**Features**:
- Register undo steps
- Merge with host history
- Undo/redo callbacks

### 4. Triggers (CLAP_EXT_TRIGGERS)
**Purpose**: Trigger/gate functionality
**Features**:
- Trigger registration
- Trigger state changes
- MIDI mapping

### 5. Resource Directory (CLAP_EXT_RESOURCE_DIRECTORY)
**Purpose**: Access plugin resources
**Features**:
- Resource file paths
- Shared resource access
- Platform-specific paths

### 6. Scratch Memory (CLAP_EXT_SCRATCH_MEMORY)
**Purpose**: Temporary memory from host
**Features**:
- Pre-allocated buffers
- Zero-allocation processing
- Size negotiation

### 7. Project Location (CLAP_EXT_PROJECT_LOCATION)
**Purpose**: Get project file information
**Features**:
- Project path
- Project name
- Relative file resolution

### 8. Extensible Audio Ports (CLAP_EXT_EXTENSIBLE_AUDIO_PORTS)
**Purpose**: Dynamic audio port configurations
**Features**:
- Add/remove ports dynamically
- Complex routing scenarios
- Modular configurations

### 9. Gain Adjustment Metering (CLAP_EXT_GAIN_ADJUSTMENT_METERING)
**Purpose**: Report gain reduction (compressors/limiters)
**Features**:
- Real-time gain reporting
- VU meter data
- Peak/RMS values

### 10. Mini Curve Display (CLAP_EXT_MINI_CURVE_DISPLAY)
**Purpose**: Display parameter automation
**Features**:
- Curve rendering
- Automation feedback
- Visual parameter state

## 📋 Factory Extensions Status

### 1. Plugin Factory (CLAP_PLUGIN_FACTORY)
**Status**: ✅ Implemented in C bridge

### 2. Preset Discovery Factory (CLAP_PRESET_DISCOVERY_FACTORY)
**Status**: ✅ Implemented in C (src/c/preset_discovery.c)

### 3. Plugin Invalidation Factory (CLAP_PLUGIN_INVALIDATION_FACTORY)
**Status**: ❌ Not implemented
**Purpose**: Notify when plugins become invalid/outdated

### 4. Plugin State Converter Factory (CLAP_PLUGIN_STATE_CONVERTER_FACTORY)
**Status**: ❌ Not implemented
**Purpose**: Convert between state format versions

## 🛠️ Implementation Strategy

### Phase 1: Complete Partial Implementations (High Priority)
1. **Audio Ports**: Full implementation with proper channel mapping
2. **State/State Context**: Complete save/load with stream wrappers
3. **Latency/Tail**: Host callbacks and proper reporting
4. **Log**: Thread-safe logging integration
5. **Timer Support**: Timer registration and callbacks
6. **Voice Info/Track Info**: Host integration and callbacks

### Phase 2: Core Missing Extensions (Medium Priority)
1. **Transport Control**: DAW transport integration
2. **Tuning**: Microtuning support for electronic music
3. **Undo**: Host undo system integration

### Phase 3: Experimental Extensions (Low Priority)
1. **Resource Directory**: Plugin resource management
2. **Scratch Memory**: Zero-allocation helpers
3. **Project Location**: Project-aware plugins
4. **Triggers**: Advanced MIDI/automation

## 🧹 Post-Implementation: Code Deduplication

After completing the remaining CLAP functionality, review example projects for duplicate code that could be abstracted into helper functions in the `pkg/` library:

### Areas to Review:
1. **Parameter Management Boilerplate**
   - Common parameter creation patterns
   - Standard parameter ranges and defaults
   - Parameter update handling

2. **Event Processing Patterns**
   - Common event handling loops
   - Note processing utilities
   - MIDI CC handling

3. **State Management**
   - JSON serialization/deserialization helpers
   - Common state structures
   - Version migration utilities

4. **Audio Processing Utilities**
   - Buffer management helpers
   - Common DSP operations
   - Channel routing utilities

5. **Extension Boilerplate**
   - Common extension registration patterns
   - Default extension implementations
   - Extension capability helpers

### Goals:
- Reduce cognitive load for plugin developers
- Provide sensible defaults while maintaining flexibility
- Create reusable components without hiding CLAP concepts
- Maintain zero-allocation guarantees in processing paths

## Development Guidelines

1. **Maintain Zero-Allocation Design**: All real-time paths must be allocation-free
2. **Follow Established Patterns**: Use existing extension implementations as reference
3. **Complete Features Only**: No placeholders or partial implementations
4. **Thread Safety**: All shared state must be properly synchronized
5. **Example Usage**: Each extension should have example usage in gain or synth plugins


### Next Phase
After doing all of this work we will begin the next phase that improves the codebase usability.
1. Document the makefille and it's build systems. All of the linking is confusing and we need to specify how each modules gets built, where, why, and how.
2. Document where the library could improve itself by hiding underlying C code without sacrificing the need to keep exported functions in the example project in order to allow the proper function handlers to work correctly calling the Go code
3. Review the state of the go-generate compabilities that should help a developer create a new plugin. I want to erase the last effort to do this and restart fresh. i first want to create a new markdown file to discuss this strategy, as i want to review all plugin types to ensure we have enough templates to cover use cases. in the end if someone chooses to generate a new gain plugin using our system. the output should be identifical to the example gain plugin we currently have in the examples/ directory. But first we need to discuss implementation strategy after deleting the old system. we need to think how to make this usable for go developers
