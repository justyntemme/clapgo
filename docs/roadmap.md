# ClapGo Roadmap

This document outlines the development roadmap for the ClapGo project.

## Version 0.1 (Current)

**Focus**: Core Infrastructure

- ✅ Basic project structure
- ✅ Go-C bridging foundation
- ✅ Simple gain plugin example
- ✅ Basic build system
- ✅ Parameter registration

## Version 0.2 (Next)

**Focus**: Core Audio Processing

- [ ] Complete audio buffer handling
  - [ ] Float32/Float64 buffer support
  - [ ] Multi-channel processing
  - [ ] Zero-copy optimization
- [ ] Event handling improvement
  - [ ] Complete MIDI event support
  - [ ] Parameter event handling
  - [ ] Note events
- [ ] Unit tests for core functionality

## Version 0.3

**Focus**: Extensions - Part 1

- [ ] Audio ports extension
- [ ] Parameters extension (full implementation)
- [ ] Note ports extension
- [ ] Latency extension
- [ ] Additional example plugin (EQ or filter)

## Version 0.4

**Focus**: Extensions - Part 2

- [ ] State extension (save/load)
- [ ] Tail extension
- [ ] Thread pool extension
- [ ] Synth example plugin

## Version 0.5

**Focus**: GUI and Advanced Features

- [ ] GUI extension
- [ ] Integration with Go UI libraries
- [ ] MIDI effect example plugin
- [ ] Performance optimization

## Version 1.0

**Focus**: Production Readiness

- [ ] Complete documentation
- [ ] All core extensions supported
- [ ] Comprehensive test suite
- [ ] Multiple example plugins
- [ ] Performance benchmarks
- [ ] Plugin validation tools

## Beyond 1.0

**Focus**: Ecosystem Growth

- [ ] Plugin SDK templates
- [ ] Plugin distribution tools
- [ ] Additional extension support
- [ ] Community contributed plugins
- [ ] Integration with common DAWs testing

## Implementation Priorities

1. **Core Functionality First**
   - Audio processing
   - Event handling
   - Parameter management

2. **Extensions Next**
   - Focus on the most common extensions first
   - Leave GUI and advanced extensions for later phases

3. **Examples and Documentation Throughout**
   - Maintain documentation alongside code
   - Create example plugins to demonstrate features

## Milestones and Timeline

| Milestone | Target Date | Major Features |
|-----------|-------------|----------------|
| v0.2 | Q3 2025 | Complete audio buffer and event handling |
| v0.3 | Q4 2025 | Core extensions implementation |
| v0.4 | Q1 2026 | Additional extensions and examples |
| v0.5 | Q2 2026 | GUI support and optimization |
| v1.0 | Q3 2026 | Full production readiness |

Note: Timeline is tentative and subject to change based on development progress and community feedback.