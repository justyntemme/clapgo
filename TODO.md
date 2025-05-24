# ClapGo Implementation TODO

This document outlines the implementation strategy for completing ClapGo's CLAP feature support, organized by priority and dependencies.

## CRITICAL PRIORITY: Complete Zero-Allocation Real-Time Processing

**Status**: Event pool completed ✅, but critical allocations remain that prevent professional audio use.

**Reference**: See `docs/considerations.md` for comprehensive analysis of Go runtime challenges in real-time audio.

### Phase 1: Event Pool System (COMPLETED ✅)
**Status**: COMPLETED - Zero-allocation event processing implemented

**Completed Implementation**:
- [x] Event pool with pre-allocated objects for all event types
- [x] Thread-safe pool management with diagnostic tracking
- [x] Automatic pool lifecycle in ProcessAllEvents
- [x] Performance monitoring with hit/miss tracking
- [x] Both examples updated and tested with clap-validator

### Phase 2: Critical Allocation Fixes (BLOCKING PROFESSIONAL USE)
**Priority**: CRITICAL - These allocations cause audio dropouts and prevent real-time compliance

Following `docs/considerations.md` Section "Garbage Collection Mitigation" and GUARDRAILS.md: Complete features or don't support them.

#### 2.1 MIDI Buffer Allocations (COMPLETED ✅)
**Location**: `pkg/api/events.go:132,150`
**Problem**: Allocates `[]byte` and `[]uint32` slices on every MIDI event
**considerations.md Reference**: "Memory Allocation Strategies" → Fixed-Size Data Structures

```go
// FIXED: Now uses fixed arrays, zero allocation
type MIDIEvent struct {
    Port int16
    Data [3]byte     // Fixed array, zero allocation
}
type MIDI2Event struct {
    Port int16  
    Data [4]uint32   // Fixed array, zero allocation
}
```

**Implementation**:
- [x] Changed MIDIEvent.Data from `[]byte` to `[3]byte`
- [x] Changed MIDI2Event.Data from `[]uint32` to `[4]uint32`
- [x] Updated all MIDI processing code to use arrays
- [x] Updated CreateMIDIEvent and CreateMIDI2Event helper functions
- [x] Fixed event pool initialization and conversion functions
- [x] Tested with clap-validator - both plugins pass

#### 2.2 Event Collection Allocation (COMPLETED ✅)
**Location**: `pkg/api/events.go:598`
**Problem**: Allocates slice on every ProcessAllEvents call
**considerations.md Reference**: "Go Runtime Challenges" → Fixed-Size Data Structures

```go
// FIXED: No more slice allocation - events returned to pool immediately
func (ep *EventProcessor) ProcessAllEvents(handler TypedEventHandler) {
    count := ep.GetInputEventCount()
    
    for i := uint32(0); i < count; i++ {
        event := ep.GetInputEvent(i)
        if event == nil {
            continue
        }
        
        // Process event...
        
        // Return event to pool immediately after processing
        ep.ReturnEventToPool(event)
    }
}
```

**Implementation**:
- [x] Eliminated slice allocation by returning events immediately
- [x] ProcessAllEvents now processes and returns events one at a time
- [x] No temporary storage needed - zero allocation approach
- [x] Tested with clap-validator - no regressions

#### 2.3 Interface{} Boxing Elimination (CRITICAL)
**Location**: `pkg/api/events.go:661`
**Problem**: Boxing event data causes allocations
**considerations.md Reference**: "Garbage Collection Mitigation" → Avoid Interface{} in Hot Paths

```go
// CURRENT: Boxing causes allocations
event.Data = *paramEvent  // interface{} boxing ALLOCATION!

// REQUIRED: Type-specific event processing
func (ep *EventProcessor) ProcessTypedEvents(handler TypedEventHandler) {
    // Direct type routing without interface{} boxing
    for i := uint32(0); i < count; i++ {
        event := ep.GetInputEvent(i)
        switch event.Type {
        case EventTypeParamValue:
            paramEvent := ep.eventPool.GetParamValueEvent()
            // Copy data directly, no boxing
            ep.convertParamValueEvent(cEvent, paramEvent)
            handler.HandleParamValue(paramEvent, event.Time)
            ep.eventPool.ReturnParamValueEvent(paramEvent)
        }
    }
}
```

**Implementation**:
- [ ] Redesign event processing to avoid interface{} storage
- [ ] Use type-specific pools directly in conversion
- [ ] Update TypedEventHandler to accept concrete types
- [ ] Eliminate Event.Data interface{} field if possible
- [ ] Benchmark to verify allocation elimination

#### 2.4 Host Logger Buffer Pool (HIGH)
**Location**: `pkg/api/host.go:90-98`
**Problem**: String formatting and C string conversion allocate
**considerations.md Reference**: "Memory Allocation Strategies" → Object Pools

```go
// CURRENT: Allocates on every log call
finalMessage := fmt.Sprintf(message, args...)  // ALLOCATION!
cMsg := C.CString(finalMessage)                // ALLOCATION!

// REQUIRED: Pre-allocated log buffer pool
type LoggerPool struct {
    stringPool sync.Pool  // Reusable string builders
    cStringPool sync.Pool // Reusable C string buffers
}
```

**Implementation**:
- [ ] Create LoggerPool with pre-allocated buffers
- [ ] Replace fmt.Sprintf with zero-allocation formatting
- [ ] Implement C string buffer reuse
- [ ] Add to EventProcessor initialization
- [ ] Test logging performance under load

#### 2.5 Parameter Listener Optimization (HIGH)
**Location**: `pkg/api/params.go:173`
**Problem**: Slice copy allocates on parameter changes
**considerations.md Reference**: "Thread Safety and Concurrency" → Lock-Free Data Structures

```go
// CURRENT: Allocates on parameter changes
copy(listeners, pm.listeners)  // ALLOCATION!

// REQUIRED: Fixed-size listener array or ring buffer
type ParameterManager struct {
    listeners [MaxListeners]ParameterListener
    listenerCount int32  // Atomic
}
```

**Implementation**:
- [ ] Replace slice with fixed-size array
- [ ] Use atomic operations for listener count
- [ ] Define MaxListeners constant
- [ ] Implement lock-free listener management
- [ ] Test parameter update performance

### Phase 3: Performance Validation Infrastructure
**Priority**: HIGH - Required to verify zero-allocation goals
**considerations.md Reference**: "Performance Monitoring and Debugging"

#### 3.1 Allocation Tracking
- [ ] Implement allocation tracker from considerations.md
- [ ] Add build tags for debug/release modes
- [ ] Create allocation benchmarks for audio path
- [ ] Add CI tests for zero-allocation verification

#### 3.2 Real-Time Performance Metrics
- [ ] Implement performance metrics from considerations.md
- [ ] Add buffer underrun detection
- [ ] Track GC pauses during processing
- [ ] Monitor voice usage statistics

#### 3.3 Memory Profiling Integration
- [ ] Add pprof integration for debug builds
- [ ] Create allocation flame graphs
- [ ] Implement runtime allocation warnings
- [ ] Add memory usage reporting

**Implementation Requirements for Phases 2-3**:
- Complete zero-allocation in audio processing path
- No allocations in Process() functions
- No allocations in event handling
- Allocation tracking for verification
- Performance benchmarks for validation
- Real-time compliance testing

**Testing Requirements**:
- `go test -bench . -benchmem` shows zero allocations
- clap-validator passes under load
- No buffer underruns under stress testing
- GC pause monitoring shows no interference

**Success Criteria**:
- 100% zero-allocation event processing ✅ (completed)
- 100% zero-allocation MIDI processing ✅ (completed)
- 100% zero-allocation parameter updates  
- 100% zero-allocation host communication
- Professional real-time audio compliance

## Lessons Learned from Phase 1-2 Implementation

### Architecture Principles (MUST FOLLOW)

1. **C Bridge Owns All Extensions**
   - NEVER ask Go plugins if they support extensions via GetExtension
   - Extension support determined at plugin load time by checking for exported functions
   - Use weak symbols (`__attribute__((weak))`) for optional exports
   - Set support flags in `go_plugin_data_t` during plugin creation

2. **No Wrapper APIs or Simplifications**
   - NEVER create wrapper types (e.g., NotePortInfoWrapper)
   - Use existing CLAP structures directly
   - Provide full CLAP interface, not simplified versions

3. **Clean CGO Patterns**
   - NEVER return Go pointers from GetExtension (causes CGO violations)
   - NEVER use dummy pointers or hacks
   - Use C static variables if non-nil indicators are needed

4. **Extension Implementation Pattern**
   ```c
   // In C bridge
   if (data->supports_extension) {
       return &s_extension_implementation;
   }
   return NULL;
   ```

5. **Manifest-Driven Architecture**
   - Extensions declared in manifest should match actual implementation
   - No Go-side registration or discovery
   - Plugin features determine behavior (e.g., "instrument" → note ports)

### Common Mistakes to Avoid

1. **Don't Mix Responsibility**
   - C bridge: CLAP interface and extension discovery
   - Go plugins: Implementation only, not interface decisions

2. **Don't Create Placeholders**
   - No TODO comments
   - No "will implement later" stubs
   - Either fully implement or explicitly don't support

3. **Don't Worry About Compatibility**
   - Make breaking changes freely
   - Update all examples immediately
   - This is a prototype - find the right architecture

## Phase 1: Core Event System (COMPLETED)

### 1.1 Complete Event Type Support
**Current State**: All event types implemented
**Completed Implementation**:
- [x] Note events (NOTE_ON, NOTE_OFF, NOTE_CHOKE, NOTE_END)
- [x] Note expression events (volume, pan, tuning, etc.)
- [x] Parameter modulation events
- [x] Parameter gesture events (begin/end)
- [x] Transport events
- [x] MIDI events (MIDI 1.0, SysEx, MIDI 2.0)
- [x] Event conversion functions (C to Go, Go to C)
- [x] TypedEventHandler interface with routing

**Implementation Strategy**:
1. Extend `pkg/api/events.go` with all event type definitions
2. Add event parsing in `pkg/bridge/bridge.go` `ClapGo_PluginProcess`
3. Create event builder functions for output events
4. Update `EventHandler` interface with methods for each event type
5. Add event filtering and routing capabilities

### 1.2 Event Input/Output Handling
**Current State**: Basic implementation via EventProcessor
**Completed Implementation**:
- [x] Input event processing via ProcessAllEvents
- [x] Output event handling via PushBackEvent
- [x] Event type detection and routing

**Still Needed**:
- [ ] Event pool for allocation-free processing (CRITICAL - See top priority section)
- [ ] Event sorting by timestamp validation (Medium Priority)
- [ ] Advanced event filtering (Lower Priority)

## Phase 2: MIDI/Note Support (COMPLETED)

### 2.1 Note Ports Extension
**Current State**: Fully implemented with clean architecture
**Completed Implementation**:
- [x] Note ports callbacks in C bridge (weak symbols)
- [x] `NotePortManager` in Go API
- [x] Port configuration structs (NotePortInfo)
- [x] Extension support detection at load time
- [x] Clean separation: C owns interface, Go owns implementation

**Key Learning**: C bridge determines extension support based on exported functions, not Go GetExtension calls

### 2.2 Note Processing
**Current State**: Fully implemented with polyphonic support
**Completed Implementation**:
- [x] Note allocation/voice management (in synth example)
- [x] Note ID tracking (voice allocation by note ID)
- [x] Basic voice lifecycle (attack, decay, sustain, release)
- [x] Polyphonic parameter support (per-note modulation)
- [x] Note expression handling (MPE support)
- [x] Voice stealing algorithms (quietest voice stealing)
- [x] Per-voice parameter automation

**Key Features Added**:
- Per-voice parameter storage (volume modulation, pitch bend, brightness, pressure)
- Polyphonic parameter event handling with note ID/key/channel matching
- Full note expression event support for MPE controllers
- Improved envelope tracking with per-voice state
- Smart voice stealing based on release phase and volume

## Implementation Guidelines for Remaining Phases

### For Each New Extension:

1. **Add Weak Symbol Declarations in C Bridge**
   ```c
   __attribute__((weak)) return_type ClapGo_ExtensionFunction(params);
   ```

2. **Add Support Flag to go_plugin_data_t**
   ```c
   bool supports_extension_name;
   ```

3. **Check for Exports During Plugin Creation**
   ```c
   data->supports_extension = (ClapGo_ExtensionFunction != NULL);
   ```

4. **Return Extension from C Bridge Based on Flag**
   ```c
   if (data->supports_extension) {
       return &s_extension_implementation;
   }
   return NULL;
   ```

5. **Go Plugin Only Exports Functions, Never Returns from GetExtension**
   - GetExtension should always return nil
   - C bridge handles all extension discovery

## Phase 3: Essential Extensions (COMPLETED)

### 3.1 Latency Extension
**Completed Implementation**:
- [x] Latency reporting callback
- [x] Dynamic latency changes via host notification

**Implementation Details**:
1. Added weak symbol `ClapGo_PluginLatencyGet` in C bridge
2. Support flag `supports_latency` in go_plugin_data_t
3. Host latency notifier utility in Go API
4. Examples updated with GetLatency() export

### 3.2 Tail Extension
**Completed Implementation**:
- [x] Audio tail length reporting
- [x] Dynamic tail changes via host notification

**Implementation Details**:
1. Added weak symbol `ClapGo_PluginTailGet` in C bridge
2. Support flag `supports_tail` in go_plugin_data_t
3. Host tail notifier utility in Go API
4. Examples updated with GetTail() export

### 3.3 Log Extension
**Completed Implementation**:
- [x] Host logging interface
- [x] Log severity levels
- [x] Thread-safe logging

**Implementation Details**:
1. Full log extension support in C bridge
2. HostLogger utility in Go API with severity levels
3. Thread-safe logging through host callbacks
4. Used throughout examples for debugging

### 3.4 Timer Support Extension
**Completed Implementation**:
- [x] Timer registration via host
- [x] Periodic callbacks via OnTimer
- [x] Timer cancellation support

**Implementation Details**:
1. Added weak symbol `ClapGo_PluginOnTimer` in C bridge
2. Support flag `supports_timer` in go_plugin_data_t  
3. HostTimerSupport utility for timer management
4. Examples updated with OnTimer() export

## Phase 5: Advanced Audio Features (COMPLETED)

### 5.1 Audio Ports Config Extension
**Completed Implementation**:
- [x] Multiple audio configurations
- [x] Configuration selection
- [x] Dynamic port creation

**Implementation Details**:
1. Added weak symbols for config count/get/select in C bridge
2. Support flags for both audio-ports-config and audio-ports-config-info
3. AudioPortsConfigProvider interface in Go API
4. Full implementation allowing plugins to offer multiple I/O configurations

### 5.2 Surround Extension
**Completed Implementation**:
- [x] Channel layout definitions
- [x] Speaker arrangement support
- [x] Channel mask validation

**Implementation Details**:
1. Added weak symbols for surround support in C bridge
2. SurroundProvider interface for channel mask/map support
3. Full surround extension callbacks implemented
4. Ready for surround/ambisonic plugin development

### 5.3 Voice Info Extension
**Completed Implementation**:
- [x] Voice count reporting
- [x] Voice capacity reporting
- [x] Active voice tracking

**Implementation Details**:
1. Added weak symbol `ClapGo_PluginVoiceInfoGet` in C bridge
2. VoiceInfoProvider interface in Go API
3. Synth example updated with GetVoiceInfo() reporting active/capacity
4. Fully functional polyphonic voice reporting

## LOWER PRIORITY: Additional Extension Support

**Note**: These extensions are NOT needed for professional audio use. Zero-allocation processing (Phases 2-3) takes absolute priority.

**considerations.md Reference**: Most extensions don't affect real-time performance, but parameter-related ones need allocation-aware design.

### Host Integration Extensions (Medium Priority)

#### Context Menu Extension  
**Status**: Not started
**considerations.md Impact**: None - GUI operations are main thread only
- [ ] Menu building API
- [ ] Menu item callbacks  
- [ ] Context-specific menus

#### Remote Controls Extension
**Status**: Not started  
**considerations.md Impact**: Parameter mapping must follow zero-allocation patterns
- [ ] Control page definitions
- [ ] Parameter mapping (use fixed arrays, not slices)
- [ ] Page switching

#### Param Indication Extension
**Status**: Not started
**considerations.md Impact**: Indication updates should be non-allocating
- [ ] Automation state indication
- [ ] Parameter mapping indication  
- [ ] Color hints

### Advanced Features (Lower Priority)

#### Thread Check Extension
**Status**: Not started
**considerations.md Impact**: HIGH - Would validate our thread safety assumptions
- [ ] Thread identification
- [ ] Thread safety validation
- [ ] Debug assertions for audio/main thread separation

#### Render Extension  
**Status**: Not started
**considerations.md Impact**: Offline rendering may allow relaxed allocation rules
- [ ] Offline rendering mode
- [ ] Render configuration
- [ ] Performance optimizations for non-real-time

#### Note Name Extension
**Status**: Not started
**considerations.md Impact**: String handling needs allocation management
- [ ] Note naming callback (pre-allocate strings)
- [ ] Custom note names
- [ ] Internationalization

### Optional Extensions (Lowest Priority)

#### Preset Discovery Factory
**Status**: Not started
**considerations.md Impact**: None - not in audio path
- [ ] Preset indexing
- [ ] Metadata extraction
- [ ] Preset provider creation

#### Draft Extensions (Experimental)
**Status**: Not started
**considerations.md Relevance**: 
- **Scratch Memory**: Directly applies to "Custom Allocators" section
- **Transport Control**: Needs lock-free parameter updates
- **Undo**: Requires careful memory management

**High-Value Candidates**:
- [ ] Tuning - Microtuning support
- [ ] Transport Control - DAW transport control  
- [ ] Undo - Undo/redo integration
- [ ] Resource Directory - Resource file handling

**Specialized Use Cases**:
- [ ] Triggers - Trigger parameters
- [ ] Scratch Memory - DSP scratch buffers (consider considerations.md stack allocator)
- [ ] Mini-Curve Display - Parameter visualization
- [ ] Gain Adjustment Metering - Gain reduction meter

## Implementation Guidelines

### For Each Extension:
1. **C Bridge Implementation**:
   - Add callback functions in `src/c/bridge.c`
   - Define extension structs
   - Handle thread safety requirements

2. **Go Exported Functions**:
   - Add exports in `pkg/bridge/bridge.go`
   - Handle C-Go type conversions
   - Validate parameters

3. **Go API Design**:
   - Create intuitive Go interfaces
   - Hide C complexity where appropriate
   - Maintain thread safety

4. **Testing**:
   - Unit tests for each component
   - Integration tests with example plugins
   - Host compatibility testing

5. **Documentation**:
   - API documentation with examples
   - Thread safety requirements
   - Best practices guide

### Development Priorities:
1. **User Impact**: Prioritize features that enable common plugin types
2. **Completeness**: Implement full extensions rather than partial
3. **Stability**: Ensure thread safety and error handling
4. **Performance**: Minimize overhead in audio thread
5. **Compatibility**: Test with multiple CLAP hosts

### Quality Standards:
- No placeholder implementations
- Complete error handling
- Thread-safe by design
- Zero allocations in audio path where possible
- Clear documentation for each feature

## Progress Summary

### Completed
- ✅ Phase 1: Core Event System (all event types, conversion, routing)
- ✅ Phase 2: Note Ports Extension (clean architecture, proper separation)
- ✅ Phase 2: Full polyphonic parameter support with MPE
- ✅ Phase 3: Essential Extensions (latency, tail, log, timer)
- ✅ Phase 5: Advanced Audio Features (audio ports config, surround, voice info)
- ✅ Phase 6: State Context Extension (context-aware save/load)
- ✅ Phase 6: Preset Load Extension (file and bundled preset loading)
- ✅ Phase 7: Track Info Extension (track metadata and flags)

### In Progress
- 🔄 Phase 6: Preset Discovery Factory (lower priority, skippable)
- 🔄 Phase 7: Host Integration (context menu, remote controls, param indication)

### Key Architecture Decisions Made
1. **C Bridge owns all extensions** - no Go-side discovery
2. **Weak symbols for optional features** - graceful degradation
3. **Extension support flags** - determined at load time
4. **No wrapper types** - use CLAP structures directly
5. **Clean CGO patterns** - no pointer violations

### IMMEDIATE NEXT STEPS (Critical Path to Professional Audio)

**Priority Order Based on `docs/considerations.md` Analysis:**

1. **Phase 2: Critical Allocation Fixes** (BLOCKING)
   - Fix MIDI buffer allocations (arrays not slices)
   - Eliminate event collection slice allocation  
   - Remove interface{} boxing in event processing
   - Implement host logger buffer pool
   - Optimize parameter listener allocations

2. **Phase 3: Performance Validation** (REQUIRED)
   - Add allocation tracking and benchmarks
   - Implement real-time performance metrics
   - Create zero-allocation test suite
   - Validate with memory profiling

3. **Extensions** (OPTIONAL - only after zero-allocation is complete)
   - Thread Check Extension (helps validate our work)
   - Host Integration Extensions (GUI features)
   - Draft Extensions (experimental)

### Current Status: Professional Audio Readiness

**✅ COMPLETED - Ready for Production**:
- Core plugin lifecycle with full CLAP compliance
- Complete event system with zero-allocation pools
- Polyphonic note processing with voice management  
- Parameter management with thread safety
- State/preset management with context awareness
- All essential and advanced extensions implemented
- Validation with clap-validator
- **NEW**: MIDI buffer allocations eliminated (using fixed arrays)
- **NEW**: Event collection allocations eliminated (immediate return to pool)

**🚨 REMAINING ISSUES - Still Blocking Professional Use**:
- Interface{} boxing allocations (major performance impact)
- Host logger allocations (causes GC pressure on logging)
- Parameter listener slice allocations (on parameter changes)

**🎯 SUCCESS CRITERIA**:
Following `docs/considerations.md`: *"Successfully using Go for real-time audio processing requires: 1. Memory allocation patterns - Pre-allocate everything possible"*

**Progress Update (May 2025)**:
- ✅ Zero-allocation event processing (Phase 1 complete)
- ✅ Zero-allocation MIDI processing (Phase 2.1 & 2.2 complete)
- ⏳ Zero-allocation parameter updates (Phase 2.3 & 2.5 pending)
- ⏳ Zero-allocation host communication (Phase 2.4 pending)
- ⏳ Real-time performance validation (Phase 3 pending)

ClapGo is making significant progress toward professional-grade real-time audio. The most critical allocations (MIDI and event collection) have been eliminated. The interface{} boxing issue remains the biggest challenge, requiring architectural changes to avoid the Event.Data interface{} field.