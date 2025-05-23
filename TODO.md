# ClapGo Implementation TODO

This document outlines the implementation strategy for completing ClapGo's CLAP feature support, organized by priority and dependencies.

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

**Still Needed** (Lower Priority):
- [ ] Event sorting by timestamp validation
- [ ] Event pool for allocation-free processing
- [ ] Advanced event filtering

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

## Phase 3: Essential Extensions (High Priority)

### 3.1 Latency Extension
**Required Implementation**:
- [ ] Latency reporting callback
- [ ] Dynamic latency changes

**Implementation Strategy**:
1. Add latency field to plugin
2. Implement callback in C bridge
3. Provide latency change notification

### 4.2 Tail Extension
**Required Implementation**:
- [ ] Audio tail length reporting
- [ ] Dynamic tail changes

**Implementation Strategy**:
1. Add tail calculation to plugin interface
2. Implement callback for tail queries
3. Support infinite tail indication

### 4.3 Log Extension
**Required Implementation**:
- [ ] Host logging interface
- [ ] Log severity levels
- [ ] Thread-safe logging

**Implementation Strategy**:
1. Create Go logging adapter
2. Route to host log callback
3. Add convenience logging methods

### 4.4 Timer Support Extension
**Required Implementation**:
- [ ] Timer registration
- [ ] Periodic callbacks
- [ ] Timer cancellation

**Implementation Strategy**:
1. Create timer manager
2. Handle main-thread callbacks
3. Provide Go-friendly timer API

## Phase 5: Advanced Audio Features (Medium Priority)

### 5.1 Audio Ports Config Extension
**Required Implementation**:
- [ ] Multiple audio configurations
- [ ] Configuration selection
- [ ] Dynamic port creation

**Implementation Strategy**:
1. Extend audio port manager
2. Support configuration sets
3. Handle configuration changes

### 5.2 Surround Extension
**Required Implementation**:
- [ ] Channel layout definitions
- [ ] Speaker arrangement
- [ ] Ambisonic support

**Implementation Strategy**:
1. Add channel layout structs
2. Implement surround callbacks
3. Provide layout conversion utilities

### 5.3 Voice Info Extension
**Required Implementation**:
- [ ] Voice count reporting
- [ ] Voice capacity
- [ ] Active voice tracking

**Implementation Strategy**:
1. Integrate with note manager
2. Add voice counting
3. Report voice statistics

## Phase 6: State and Preset Management (Medium Priority)

### 6.1 State Context Extension
**Required Implementation**:
- [ ] Save/load context (project, preset, etc.)
- [ ] Duplicate state detection

**Implementation Strategy**:
1. Extend state save/load with context
2. Add context type identification
3. Support preset-specific behavior

### 6.2 Preset Load Extension
**Required Implementation**:
- [ ] Preset file loading
- [ ] Factory preset support
- [ ] Preset location handling

**Implementation Strategy**:
1. Implement preset load callbacks
2. Add preset file parsing
3. Support preset discovery

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

### 7.1 Track Info Extension
**Required Implementation**:
- [ ] Track name/color/index
- [ ] Channel configuration
- [ ] Track flags

**Implementation Strategy**:
1. Add track info callbacks
2. Store track metadata
3. Notify plugin of changes

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
- ✅ Phase 2.1: Note Ports Extension (clean architecture, proper separation)
- ✅ Basic voice management and note processing in examples

### In Progress
- 🔄 Ready to begin Phase 3: Essential Extensions

### Key Architecture Decisions Made
1. **C Bridge owns all extensions** - no Go-side discovery
2. **Weak symbols for optional features** - graceful degradation
3. **Extension support flags** - determined at load time
4. **No wrapper types** - use CLAP structures directly
5. **Clean CGO patterns** - no pointer violations

### Next Steps
1. Begin Phase 3: Essential Extensions (latency, tail, log, timer)
2. Follow the implementation guidelines for each new extension
3. Update all examples to demonstrate new features
4. Continue following GUARDRAILS principles strictly

### Phase 2 Summary
Phase 2 has been successfully completed with full polyphonic parameter support:
- **Voice Management**: Robust voice allocation with smart stealing
- **Polyphonic Parameters**: Per-voice modulation of volume, pitch, brightness, pressure
- **MPE Support**: Complete note expression handling for expressive controllers
- **Clean Architecture**: All features follow the C bridge ownership pattern

This implementation strategy ensures ClapGo evolves from its current foundation into a complete CLAP plugin development solution while maintaining the architectural principles defined in GUARDRAILS.md.