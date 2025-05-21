# ClapGo Development Roadmap

This roadmap identifies critical CLAP features and improvements needed to pass clap-validator tests and provide a comprehensive plugin development platform.

## Current Status

✅ **Working Core Features:**
- Plugin discovery via JSON manifests
- Basic plugin lifecycle (init, activate, process, destroy)
- Cross-platform library loading and symbol resolution
- Safe CGO handle system for Go-C object passing
- Audio processing framework (basic pass-through)
- Successful clap-validator core tests (8/8 passed)

## Critical Missing Features (High Priority)

Based on clap-validator analysis, these are the most important missing features:

### 1. **Parameters Extension** (`clap.params`) - 🔴 CRITICAL
**Status**: Not implemented  
**Impact**: 13 tests skipped
**Required for**: Any plugin with user-controllable parameters

**Tasks**:
- [ ] Implement C-side parameter extension interface
- [ ] Create Go `ParamsProvider` interface implementation
- [ ] Add parameter info query functions
- [ ] Implement parameter value get/set with thread safety
- [ ] Add parameter value/string conversion support
- [ ] Support parameter automation/modulation events
- [ ] Add parameter flush functionality

**Minimal Implementation Requirements**:
```go
type ParamsProvider interface {
    GetParamCount() uint32
    GetParamInfo(index uint32) *ParamInfo
    GetParamValue(id uint32) float64
    SetParamValue(id uint32, value float64) bool
    ValueToText(id uint32, value float64) string
    TextToValue(id uint32, text string) (float64, bool)
    FlushParams(inputEvents []Event, outputEvents *[]Event)
}
```

### 2. **Audio Ports Extension** (`clap.audio-ports`) - 🟡 HIGH
**Status**: Partially implemented (hardcoded stereo)  
**Impact**: Limited to stereo plugins only
**Required for**: Multi-channel, surround, or mono plugins

**Tasks**:
- [ ] Implement dynamic audio port configuration
- [ ] Support mono, stereo, surround configurations
- [ ] Add proper port naming and identification
- [ ] Implement in-place processing pair support
- [ ] Add port type definitions (main, aux, sidechain)

**Current Issues**:
```c
// Current: Hardcoded stereo implementation
info->channel_count = 2;
info->port_type = CLAP_PORT_STEREO;

// Needed: Dynamic configuration from Go
```

### 3. **Note Ports Extension** (`clap.note-ports`) - 🟡 HIGH  
**Status**: Not implemented
**Impact**: No MIDI/note input support
**Required for**: Instruments, synthesizers, note-based effects

**Tasks**:
- [ ] Implement note port configuration interface
- [ ] Support CLAP native note events
- [ ] Add MIDI dialect support (MIDI 1.0, MPE, MIDI 2.0)
- [ ] Implement note event processing in Go plugins
- [ ] Add note port naming and capabilities

### 4. **State Extension** (`clap.state`) - 🟡 HIGH
**Status**: Not implemented
**Impact**: No preset/state save/load capability
**Required for**: Plugin presets, DAW project save/load

**Tasks**:
- [ ] Implement state save/load C interface
- [ ] Create Go `StateProvider` interface
- [ ] Add binary state stream handling
- [ ] Support incremental state updates
- [ ] Add state validation and error handling

## Important Missing Features (Medium Priority)

### 5. **Proper Audio Processing** - 🟡 MEDIUM
**Status**: Stub implementation (pass-through)
**Current Issues**:
- Audio buffer conversion not implemented
- Event processing not connected
- Sample-accurate event timing not supported

**Tasks**:
- [ ] Implement C audio buffer → Go slice conversion
- [ ] Add sample-accurate event processing
- [ ] Support different audio buffer layouts
- [ ] Add proper in-place vs out-of-place processing
- [ ] Implement event timestamp handling

### 6. **Host Extensions Support** - 🟡 MEDIUM
**Status**: Not implemented
**Required for**: Advanced plugin-host interaction

**Tasks**:
- [ ] Host log extension (`clap.log`)
- [ ] Host thread check extension (`clap.thread-check`)
- [ ] Host latency extension (`clap.latency`)
- [ ] Host state extension (`clap.state`)
- [ ] Host parameter flush extension

### 7. **GUI Support** (`clap.gui`) - 🟡 MEDIUM
**Status**: Interface defined, not implemented
**Required for**: Plugin UIs

**Tasks**:
- [ ] Implement native window embedding
- [ ] Add web-based UI support via CEF or WebView
- [ ] Support plugin window lifecycle
- [ ] Add UI parameter binding
- [ ] Cross-platform window handling

## Extension Features (Lower Priority)

### 8. **Preset Discovery** (`clap.preset-discovery-factory`) - 🟢 LOW
**Status**: Not implemented
**Impact**: 3 tests skipped
**Required for**: Advanced preset browsing

### 9. **Additional Extensions** - 🟢 LOW
- [ ] Latency reporting (`clap.latency`)
- [ ] Tail length (`clap.tail`) 
- [ ] Voice info (`clap.voice-info`)
- [ ] Remote controls (`clap.remote-controls`)
- [ ] Timer support (`clap.timer-support`)

## Architecture Improvements

### 10. **Error Handling & Logging** - 🟡 MEDIUM
**Current Issues**:
- Minimal error reporting
- No structured logging system
- Limited debugging capabilities

**Tasks**:
- [ ] Implement comprehensive error handling
- [ ] Add structured logging system
- [ ] Create debugging utilities
- [ ] Add error recovery mechanisms

### 11. **Memory Management** - 🟡 MEDIUM  
**Tasks**:
- [ ] Audit CGO handle lifecycle
- [ ] Add memory leak detection
- [ ] Optimize buffer allocations
- [ ] Implement pool-based allocations for RT code

### 12. **Performance Optimization** - 🟢 LOW
**Tasks**:
- [ ] Profile audio processing hot paths
- [ ] Minimize CGO call overhead
- [ ] Optimize event processing
- [ ] Add SIMD support for audio operations

## Build System & Tooling

### 13. **Developer Experience** - 🟡 MEDIUM
**Tasks**:
- [ ] Add plugin generator/template tool
- [ ] Improve build system with better error messages
- [ ] Add development-time plugin validation
- [ ] Create comprehensive documentation and examples

### 14. **Testing Infrastructure** - 🟡 MEDIUM
**Tasks**:
- [ ] Automated clap-validator integration
- [ ] Unit tests for bridge components  
- [ ] Audio processing correctness tests
- [ ] Memory leak tests
- [ ] Cross-platform CI/CD

## Implementation Priority

### Phase 1: Core Functionality (Q1)
1. **Parameters Extension** - Enable basic parameter automation
2. **Audio Processing** - Implement proper audio buffer handling
3. **Audio Ports** - Support flexible audio configurations
4. **State Extension** - Enable preset save/load

### Phase 2: Instrument Support (Q2)  
1. **Note Ports Extension** - Enable synthesizers and instruments
2. **Host Extensions** - Improve plugin-host communication
3. **Error Handling** - Production-ready error management

### Phase 3: Advanced Features (Q3)
1. **GUI Support** - Plugin user interfaces
2. **Performance Optimization** - Real-time audio performance
3. **Additional Extensions** - Complete CLAP feature support

### Phase 4: Developer Experience (Q4)
1. **Tooling & Templates** - Streamlined development workflow
2. **Testing Infrastructure** - Comprehensive test coverage
3. **Documentation** - Complete developer guides

## Validation Targets

### clap-validator Goals
- **Phase 1**: 15+ tests passing (currently 8/21)
- **Phase 2**: 20+ tests passing with instrument support
- **Phase 3**: All applicable tests passing
- **Phase 4**: Zero skipped tests for implemented features

### Real-World Plugin Targets
- **Phase 1**: Working audio effect with parameters
- **Phase 2**: Basic synthesizer/instrument
- **Phase 3**: Commercial-quality plugin with GUI
- **Phase 4**: Full-featured plugin suite

## Technical Specifications

### Parameters Extension Details
```go
// Required for clap-validator param tests to pass
type ParamInfo struct {
    ID           uint32
    Name         string
    Module       string
    MinValue     float64
    MaxValue     float64
    DefaultValue float64
    Flags        ParamFlags
}

type ParamsProvider interface {
    // Required for basic functionality
    GetParamCount() uint32
    GetParamInfo(index uint32) *ParamInfo
    GetParamValue(id uint32) float64
    SetParamValue(id uint32, value float64) bool
    
    // Required for clap-validator param-conversions test
    ValueToText(id uint32, value float64) string
    TextToValue(id uint32, text string) (float64, bool)
    
    // Required for automation support
    FlushParams(inputEvents []Event, outputEvents *[]Event)
}
```

### Audio Processing Requirements
```go
// Must handle various audio configurations
type AudioPortInfo struct {
    ID            uint32
    Name          string
    ChannelCount  uint32
    PortType      string    // "mono", "stereo", "surround", etc.
    Flags         AudioPortFlags
    InPlacePair   uint32    // For in-place processing
}

// Must process sample-accurate events
type ProcessContext struct {
    SteadyTime    int64
    FrameCount    uint32
    AudioInputs   [][]float32
    AudioOutputs  [][]float32
    InputEvents   []Event
    OutputEvents  *[]Event
}
```

This roadmap focuses on the most critical features needed to pass clap-validator tests and support real-world plugin development. The parameters extension is the highest priority as it unlocks the majority of currently skipped tests.