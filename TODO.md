# ClapGo Implementation TODO

This document outlines the implementation strategy for completing ClapGo's CLAP feature support, organized by priority and dependencies.

## CRITICAL PRIORITY: Memory Allocation Optimization

### Event Pool for Zero-Allocation Processing (COMPLETED ✅)
**Priority**: CRITICAL - Performance impact on real-time audio thread
**Status**: COMPLETED - Event pool implemented with zero-allocation processing

**Completed Implementation**:
- [x] Implement event pool with pre-allocated event objects
- [x] Support for all event types (note, MIDI, parameter, etc.)
- [x] Thread-safe pool management for multi-threaded hosts
- [x] Configurable pool size based on expected event density
- [x] Automatic pool growth if needed (with warnings)
- [x] Diagnostics for pool performance tracking
- [x] Updated both gain and synth examples to use event pool
- [x] Tested in CLAP hosts with clap-validator - all tests pass

### Additional Zero-Allocation Optimizations Needed
**Priority**: CRITICAL - Remaining allocations in audio processing path

Following GUARDRAILS.md principles: Complete features or don't support them. No placeholders or half-measures.

#### CRITICAL: Audio Processing Path Allocations

1. **MIDI Data Buffer Allocations** `pkg/api/events.go:132,150`
   ```go
   // CURRENT: Allocates on every MIDI event
   return &MIDIEvent{Data: make([]byte, 3)}
   return &MIDI2Event{Data: make([]uint32, 4)}
   
   // REQUIRED: Pre-allocated fixed-size buffers
   type MIDIEvent struct {
       Port int16
       Data [3]byte  // Fixed array, not slice
   }
   ```

2. **Event Collection Slice** `pkg/api/events.go:598`
   ```go
   // CURRENT: Allocates slice on every ProcessAllEvents call
   events := make([]*Event, 0, count)
   
   // REQUIRED: Reusable event slice in EventProcessor
   type EventProcessor struct {
       eventCollection []*Event  // Pre-allocated, reused
   }
   ```

3. **Interface{} Boxing in Event.Data** `pkg/api/events.go:661`
   ```go
   // CURRENT: Boxing causes allocations
   event.Data = *paramEvent  // interface{} boxing
   
   // REQUIRED: Avoid interface{} or use sync.Pool for boxes
   ```

#### HIGH: Parameter and Host Communication

4. **Host Logger Allocations** `pkg/api/host.go:90-98`
   ```go
   // CURRENT: Allocates on every log call
   finalMessage := fmt.Sprintf(message, args...)
   cMsg := C.CString(finalMessage)
   
   // REQUIRED: Pre-allocated log buffer pool
   ```

5. **Parameter Listener Slice Copy** `pkg/api/params.go:173`
   ```go
   // CURRENT: Allocates on parameter changes
   copy(listeners, pm.listeners)
   
   // REQUIRED: Fixed-size listener array or buffer pool
   ```

#### MEDIUM: Less Frequent Operations

6. **Error Allocations** `pkg/api/params.go:49,63-64,66`
   ```go
   // CURRENT: Creates new error strings
   return errors.New("parameter not found")
   
   // REQUIRED: Pre-defined error constants
   var ErrParameterNotFound = errors.New("parameter not found")
   ```

7. **Stream Buffer Allocations** `pkg/api/stream.go:98,106,172`
   ```go
   // CURRENT: Allocates buffers for each stream operation
   buf := make([]byte, length)
   
   // REQUIRED: Buffer pool for stream operations
   ```

**Implementation Requirements**:
- Complete zero-allocation audio processing path
- Pre-allocate all buffers at startup
- Use object pools for variable-size data
- No allocations in Process() functions
- No allocations in event handling
- Add allocation tracking for verification

**Testing Requirements**:
- Add allocation benchmarks
- Verify zero allocations with go test -bench . -benchmem
- Test with allocation profiling enabled
- Validate with multiple CLAP hosts under load

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

## Phase 6: State and Preset Management (Medium Priority)

### 6.1 State Context Extension (COMPLETED)
**Completed Implementation**:
- [x] Save/load context (project, preset, duplicate)
- [x] Context type identification
- [x] Preset-specific behavior (voice clearing in synth)

**Implementation Details**:
1. Added weak symbols for context-aware save/load in C bridge
2. StateContextProvider interface in Go API
3. Context types: preset, duplicate, project
4. Both example plugins updated with context awareness

### 6.2 Preset Load Extension (COMPLETED)
**Completed Implementation**:
- [x] Preset file loading from filesystem
- [x] Factory preset support (bundled presets)
- [x] Preset location handling (file vs plugin)

**Implementation Details**:
1. Added weak symbol for preset load callback in C bridge
2. PresetLoader interface in Go API
3. Support for both file and bundled preset locations
4. JSON preset parsing in both example plugins
5. Voice clearing on preset load in synth

### 6.3 Preset Discovery Factory
**Required Implementation**:
- [ ] Preset indexing
- [ ] Metadata extraction
- [ ] Preset provider creation

**Implementation Strategy**:
1. Create preset scanner
2. Build preset database
3. Implement factory interface

## Phase 7: Host Integration (Lower Priority)

### 7.1 Track Info Extension (COMPLETED)
**Completed Implementation**:
- [x] Track name/color/index
- [x] Channel configuration  
- [x] Track flags (return, bus, master)

**Implementation Details**:
1. Added weak symbol for track info changed callback in C bridge
2. TrackInfoProvider interface in Go API
3. HostTrackInfo utility for querying track information
4. Both example plugins log track info and can adapt behavior
5. Full support for track metadata including color and port types

### 7.2 Context Menu Extension
**Required Implementation**:
- [ ] Menu building
- [ ] Menu item callbacks
- [ ] Context-specific menus

**Implementation Strategy**:
1. Create menu builder API
2. Handle menu events
3. Support submenus

### 7.3 Remote Controls Extension
**Required Implementation**:
- [ ] Control page definitions
- [ ] Parameter mapping
- [ ] Page switching

**Implementation Strategy**:
1. Define control page structure
2. Implement mapping system
3. Add page management

### 7.4 Param Indication Extension
**Required Implementation**:
- [ ] Automation state indication
- [ ] Parameter mapping indication
- [ ] Color hints

**Implementation Strategy**:
1. Add indication callbacks
2. Track automation state
3. Provide color suggestions

## Phase 8: Advanced Features (Lower Priority)

### 8.1 Thread Check Extension
**Required Implementation**:
- [ ] Thread identification
- [ ] Thread safety validation

**Implementation Strategy**:
1. Add thread tracking
2. Implement debug checks
3. Provide thread assertions

### 8.2 Render Extension
**Required Implementation**:
- [ ] Offline rendering mode
- [ ] Render configuration

**Implementation Strategy**:
1. Add render mode handling
2. Support different render settings
3. Optimize for offline processing

### 8.3 Note Name Extension
**Required Implementation**:
- [ ] Note naming callback
- [ ] Custom note names

**Implementation Strategy**:
1. Add note name provider
2. Support different naming schemes
3. Handle internationalization

## Phase 9: Draft Extensions (Experimental)

### 9.1 High-Value Draft Extensions
**Priority candidates based on user value**:
- [ ] Tuning - Microtuning support
- [ ] Transport Control - DAW transport control
- [ ] Undo - Undo/redo integration
- [ ] Resource Directory - Resource file handling

### 9.2 Specialized Draft Extensions
**Lower priority, specific use cases**:
- [ ] Triggers - Trigger parameters
- [ ] Scratch Memory - DSP scratch buffers
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

### Next Steps
1. Continue Phase 6: State and Preset Management
2. Implement state context extension for save/load context awareness
3. Add preset load extension for factory and user presets
4. Consider preset discovery factory for advanced preset management

### Phase 2 Summary
Phase 2 has been successfully completed with full polyphonic parameter support:
- **Voice Management**: Robust voice allocation with smart stealing
- **Polyphonic Parameters**: Per-voice modulation of volume, pitch, brightness, pressure
- **MPE Support**: Complete note expression handling for expressive controllers
- **Clean Architecture**: All features follow the C bridge ownership pattern

This implementation strategy ensures ClapGo evolves from its current foundation into a complete CLAP plugin development solution while maintaining the architectural principles defined in GUARDRAILS.md.