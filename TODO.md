# ClapGo Implementation TODO

This document outlines the implementation strategy for completing ClapGo's CLAP feature support, organized by priority and dependencies.

## Phase 1: Core Event System (High Priority)

### 1.1 Complete Event Type Support
**Current State**: Only parameter value events implemented
**Required Implementation**:
- [ ] Note events (NOTE_ON, NOTE_OFF, NOTE_CHOKE, NOTE_END)
- [ ] Note expression events (volume, pan, tuning, etc.)
- [ ] Parameter modulation events
- [ ] Parameter gesture events (begin/end)
- [ ] Transport events
- [ ] MIDI events (MIDI 1.0, SysEx, MIDI 2.0)

**Implementation Strategy**:
1. Extend `pkg/api/events.go` with all event type definitions
2. Add event parsing in `pkg/bridge/bridge.go` `ClapGo_PluginProcess`
3. Create event builder functions for output events
4. Update `EventHandler` interface with methods for each event type
5. Add event filtering and routing capabilities

### 1.2 Event Input/Output Handling
**Required Implementation**:
- [ ] Separate input/output event queues
- [ ] Event sorting by timestamp
- [ ] Event validation
- [ ] Event pool for allocation-free processing

**Implementation Strategy**:
1. Modify `Audio` struct to include event queues
2. Implement event iterator for input processing
3. Add event output methods with proper ordering
4. Create event pool with pre-allocated events

## Phase 2: MIDI/Note Support (High Priority)

### 2.1 Note Ports Extension
**Current State**: Interface defined but not implemented
**Required Implementation**:
- [ ] Note port configuration
- [ ] Preferred channel/port assignment
- [ ] Per-port dialect (CLAP, MIDI, MIDI-MPE, MIDI2)

**Implementation Strategy**:
1. Implement note ports callbacks in C bridge
2. Add `NotePortManager` in Go API
3. Create port configuration structs
4. Link with event system for proper routing

### 2.2 Note Processing
**Required Implementation**:
- [ ] Note allocation/voice management
- [ ] Note ID tracking
- [ ] Polyphonic parameter support
- [ ] Note expression handling

**Implementation Strategy**:
1. Create `NoteManager` for voice allocation
2. Add per-note parameter storage
3. Implement note-to-voice mapping
4. Support MPE and per-note modulation

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

## Next Steps

1. Begin with Phase 1 (Event System) as it's foundational
2. Implement Phase 2 (MIDI/Note) to enable instrument plugins  
3. Continue with phases based on user feedback and needs
4. Regularly test with real-world plugins and hosts

This implementation strategy ensures ClapGo evolves from its current foundation into a complete CLAP plugin development solution while maintaining the architectural principles defined in GUARDRAILS.md.