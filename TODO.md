# ClapGo Implementation TODO

This document outlines the remaining implementation work for ClapGo's CLAP feature support.

## Executive Summary

ClapGo has achieved **professional-grade real-time audio compliance** with 100% allocation-free processing in all critical paths. The core architecture is complete and validated.

### ✅ What's Complete
- **Core Plugin System**: Full CLAP compliance with manifest-driven architecture
- **Zero-Allocation Processing**: Event pools, fixed arrays, buffer pools - VERIFIED with benchmarks
- **Polyphonic Support**: Complete with voice management and MPE
- **All Essential Extensions**: Audio ports, params, state, latency, tail, log, timer support
- **All Advanced Audio Features**: Surround, audio configs, voice info, track info, ambisonic, audio ports activation
- **All Host Integration Extensions**: Context menu, remote controls, param indication, note name
- **Core Extensions**: Event registry, configurable audio ports, POSIX FD support, render, thread pool
- **Performance Validation**: Allocation tracking, benchmarks, profiling tools with thread safety validation

## 🚨 Current Priority: Preset Discovery Factory

The Preset Discovery Factory extension is now our highest priority implementation target. This factory-level extension enables hosts to discover and index presets across the system, providing centralized preset browsing capabilities.

### Why Preset Discovery is Important
- Allows users to browse presets from one central location consistently
- Users can browse presets without committing to a particular plugin first
- Enables fast indexing and searching within the host's browser
- Provides metadata for intelligent categorization and tagging

### Implementation Requirements

**Architecture Considerations:**
- This is a **factory-level** extension, not a plugin extension
- Lives at the plugin entry level, not individual plugin instances
- Must be FAST - the whole indexing process needs to be performant
- Must be NON-INTERACTIVE - no dialogs, windows, or user input
- Zero-allocation design for metadata scanning operations

**Key Components to Implement:**
1. **Preset Discovery Factory** (`clap_preset_discovery_factory_t`)
   - Factory interface in plugin entry
   - Creates preset discovery providers

2. **Preset Discovery Provider** (`clap_preset_discovery_provider_t`)
   - Declares supported file types
   - Declares preset locations
   - Provides metadata extraction

3. **Preset Discovery Indexer** (host-side interface)
   - Receives declarations from provider
   - Handles metadata callbacks

4. **Metadata Structures**
   - Preset metadata with plugin IDs, names, features
   - Soundpack information
   - File type associations

**Implementation Strategy:**
Following the established ClapGo patterns:
1. Add factory support to plugin entry (not individual plugins)
2. Create Go interfaces for preset discovery provider
3. Implement C bridge for factory and provider callbacks
4. Design efficient metadata extraction without allocations
5. Support both file-based and bundled presets
6. Integrate with existing manifest system

**Performance Requirements:**
- Pre-allocated buffers for metadata
- No string allocations during scanning
- Efficient file I/O patterns
- Caching strategies for repeated scans

## 📋 Not Planned (Out of Scope)

### GUI Extension
- **Status**: WILL NOT BE IMPLEMENTED
- **Reason**: GUI examples are explicitly forbidden per guardrails
- All GUI-related work is out of scope for ClapGo

## 🔮 Future Extensions (After Preset Discovery)

### Draft Extensions (Lower Priority)
These experimental extensions may be implemented after preset discovery is complete:

#### Tuning Extension (CLAP_EXT_TUNING)
- Microtuning table support
- Fixed-size tuning arrays
- Pre-calculated frequency tables

#### Transport Control Extension (CLAP_EXT_TRANSPORT_CONTROL)
- Transport state callbacks
- Lock-free transport info updates
- Host transport synchronization

#### Undo Extension (CLAP_EXT_UNDO)
- State delta tracking
- Fixed-size undo buffer
- Parameter change history

#### Other Draft Extensions
- Resource Directory
- Triggers
- Scratch Memory
- Mini Curve Display
- Gain Adjustment Metering
- Extensible Audio Ports
- Project Location

## Development Guidelines

### Architecture Principles (MUST FOLLOW)
1. **Factory Extensions Are Different**
   - Live at plugin entry level, not plugin instance
   - No weak symbols on plugin methods
   - Factory creation happens before plugin instantiation

2. **Maintain Zero-Allocation Design**
   - Pre-allocate all buffers for scanning
   - Use fixed-size structures for metadata
   - Avoid string operations in hot paths

3. **Follow Established Patterns**
   - C bridge owns all CLAP interfaces
   - Go provides implementation only
   - No wrapper types or simplifications

### Implementation Checklist for Preset Discovery
- [ ] Review `factory/preset-discovery.h` header thoroughly
- [ ] Design factory integration with plugin entry
- [ ] Create preset discovery provider interface
- [ ] Implement metadata structures with fixed sizes
- [ ] Add C bridge for factory and provider
- [ ] Create efficient file scanning implementation
- [ ] Add preset location management
- [ ] Implement file type registration
- [ ] Create example preset discovery provider
- [ ] Test with various preset formats
- [ ] Benchmark scanning performance
- [ ] Ensure zero allocations in scan path

## Success Criteria
- Preset discovery factory properly integrated at entry level
- Fast, non-interactive preset scanning
- Zero allocations during metadata extraction
- Support for multiple preset locations and formats
- Clean integration with existing ClapGo architecture