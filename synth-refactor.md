# Synth Plugin Refactor Implementation Strategy

## Detailed Implementation Guide: Replacing Internal Code with pkg/ Library Functions

This guide documents every instance where the synth plugin reimplements functionality that already exists in the pkg/ library. Following the guardrails principle of avoiding code duplication, we will systematically replace internal implementations with framework components.

### Current State Analysis
- **Total Lines**: 1781 (main.go: 961, exports.go: 623, constants.go: 19)
- **Target Lines**: ~450 lines
- **Primary Issue**: Massive code duplication where pkg/ functions already exist

## Comprehensive List of Internal Implementations to Replace

### 1. Voice Structure (Lines 53-74 in main.go)
**Current**: Custom Voice struct with manual field management
```go
type Voice struct {
    NoteID   int32
    Channel  int16
    Key      int16
    Velocity float64
    Phase    float64
    IsActive bool
    // ... more fields
}
```
**Replace with**: `audio.Voice` from `pkg/audio/voice.go`
- Already includes all voice fields
- Built-in modulation parameters
- Proper envelope integration
**Lines saved**: ~22

### 2. ADSR Envelope (Lines 347-405, 619, 764-770 in main.go)
**Current**: Manual envelope management in multiple places
- Creating envelope: `audio.NewADSREnvelope()` but not using `audio.ADSREnvelope` properly
- Manual envelope processing in `getEnvelopeValue()`
- Redundant envelope state checks
**Replace with**: Proper use of `audio.ADSREnvelope` from `pkg/audio/envelope.go`
- Use `envelope.Process()` directly
- Remove `getEnvelopeValue()` function entirely
**Lines saved**: ~70

### 3. Voice Management (Lines 729-761 in main.go)
**Current**: Manual voice allocation with `findFreeVoice()` function
**Replace with**: `audio.VoiceManager` from `pkg/audio/voice.go`
- Has built-in voice allocation
- Automatic voice stealing
- Cleaner API: `vm.AllocateVoice()`, `vm.ReleaseVoice()`
**Lines saved**: ~33

### 4. Waveform Generation (Line 349 in main.go)
**Current**: Using `audio.GenerateWaveformSample()` but could use higher-level components
**Replace with**: `audio.PolyphonicOscillator` from `pkg/audio/synth.go`
- Handles multiple voices automatically
- Built-in anti-aliasing
- Simpler API
**Lines saved**: ~50 in process loop

### 5. MIDI Event Processing (Lines 708-727 in main.go)
**Current**: Manual MIDI parsing with `event.ProcessStandardMIDI()`
**Replace with**: `audio.MIDIProcessor` from `pkg/audio/midi.go`
- Complete MIDI handling
- Automatic note/CC routing
- Built-in pitch bend handling
**Lines saved**: ~20

### 6. Note Frequency Calculation (Lines 331-342 in main.go)
**Current**: Manual frequency calculation with tuning
```go
baseFreq := audio.NoteToFrequency(int(voice.Key))
if p.tuning != nil && voice.TuningID != 0 {
    baseFreq = p.tuning.ApplyTuning(...)
}
freq := baseFreq * math.Pow(2.0, voice.PitchBend/12.0)
```
**Replace with**: Built-in pitch processing in `audio.Voice`
**Lines saved**: ~12

### 7. Parameter Atomic Operations (Lines 143-149, 302-308, 529-546 in main.go)
**Current**: Manual atomic storage with helper functions
```go
atomic.StoreInt64(&plugin.volume, int64(util.AtomicFloat64ToBits(0.7)))
volume := param.LoadParameterAtomic(&p.volume)
```
**Replace with**: `param.AtomicParam` from `pkg/param/atomic.go`
- Cleaner API
- Type-safe operations
**Lines saved**: ~30

### 8. Audio Buffer Management (Lines 310-323, 374-387 in main.go)
**Current**: Manual buffer clearing and mixing
```go
for ch := 0; ch < numChannels; ch++ {
    for i := uint32(0); i < framesCount; i++ {
        outChannel[i] = 0.0
    }
}
```
**Replace with**: `audio.Buffer` utilities from `pkg/audio/buffer.go`
- `audio.ClearBuffer()`
- `audio.MixToStereoOutput()`
**Lines saved**: ~25

### 9. Transport Info Structure (Lines 106-118 in main.go)
**Current**: Custom TransportInfo struct
**Replace with**: Use host transport directly via `hostpkg.TransportControl`
**Lines saved**: ~13

### 10. Event Processing (Lines 409-416, 509-515 in main.go)
**Current**: Wrapper functions around event processor
**Replace with**: Direct use of `event.EventProcessor`
**Lines saved**: ~15

### 11. State Management (Lines 812-845 in exports.go)
**Current**: Manual state serialization with JSON
**Replace with**: `state.Manager` with automatic serialization
**Lines saved**: ~35

### 12. Note Port Management (Lines 162-163, 773-776 in main.go)
**Current**: Creating ports but not fully utilizing manager
**Replace with**: Full use of `audio.NotePortManager` features
**Lines saved**: ~10

### 13. Audio Port Configuration (Implied in Process)
**Current**: Manual channel handling
**Replace with**: `audio.StereoPortProvider` from `pkg/audio/ports.go`
- Automatic port configuration
- Built-in channel mapping
**Lines saved**: ~20

### 14. Filter Implementation (Lines 353-357 in main.go)
**Current**: Pseudo-brightness filter
```go
if voice.Brightness > 0.0 && voice.Brightness < 1.0 {
    sample *= (voice.Brightness*0.7 + 0.3)
}
```
**Replace with**: `audio.SimpleLowPassFilter` from `pkg/audio/synth.go`
**Lines saved**: ~5

### 15. Complete Voice Processing Chain
**Current**: Manual synthesis in process loop (lines 326-405)
**Replace with**: `audio.SynthVoiceProcessor` from `pkg/audio/synth.go`
- Complete voice with oscillator + filter
- Integrated envelope
- Cleaner processing
**Lines saved**: ~80

## Summary of Replacements

| Component | Current Lines | Framework Component | Lines Saved |
|-----------|--------------|-------------------|-------------|
| Voice struct | 22 | audio.Voice | 22 |
| ADSR envelope | 70 | audio.ADSREnvelope | 70 |
| Voice management | 33 | audio.VoiceManager | 33 |
| Waveform generation | 50 | audio.PolyphonicOscillator | 50 |
| MIDI processing | 20 | audio.MIDIProcessor | 20 |
| Frequency calculation | 12 | Built-in voice pitch | 12 |
| Atomic parameters | 30 | param.AtomicParam | 30 |
| Buffer management | 25 | audio.Buffer utils | 25 |
| Transport info | 13 | Direct host usage | 13 |
| Event processing | 15 | Direct processor use | 15 |
| State management | 35 | state.Manager | 35 |
| Note ports | 10 | Full manager usage | 10 |
| Audio ports | 20 | audio.StereoPortProvider | 20 |
| Filter | 5 | audio.SimpleLowPassFilter | 5 |
| Voice processing | 80 | audio.SynthVoiceProcessor | 80 |
| **Total** | **440** | | **440** |

## Implementation Order

1. **Replace Voice Management System** (Priority: HIGH)
   - Switch to `audio.VoiceManager`
   - Use `audio.Voice` struct
   - Remove manual allocation logic

2. **Replace Audio Processing Chain** (Priority: HIGH)
   - Use `audio.SynthVoiceProcessor` or `audio.PolyphonicOscillator`
   - Remove manual synthesis loop
   - Integrate proper filtering

3. **Simplify Parameter Management** (Priority: MEDIUM)
   - Switch to `param.AtomicParam`
   - Use parameter builders
   - Remove manual atomic operations

4. **Streamline MIDI Processing** (Priority: MEDIUM)
   - Implement `audio.MIDIProcessor`
   - Remove manual MIDI parsing
   - Automatic event routing

5. **Clean Up State Management** (Priority: LOW)
   - Use framework state utilities
   - Remove custom serialization
   - Automatic parameter state

## Final Plugin Structure Vision

```go
type SynthPlugin struct {
    *plugin.PluginBase
    *audio.StereoPortProvider
    
    synth *audio.PolyphonicOscillator
    midi  *audio.MIDIProcessor
    
    // Parameters using AtomicParam
    volume   *param.AtomicParam
    waveform *param.AtomicParam
    envelope *param.EnvelopeParams
}

// Process method reduced to ~40 lines
func (p *SynthPlugin) Process(...) int {
    p.midi.ProcessEvents(events)
    samples := p.synth.Process(framesCount)
    audio.CopyToStereoOutput(samples, audioOut, p.volume.Get())
    return process.ProcessContinue
}
```

## Key Principle Violations Being Fixed

1. **Code Duplication**: Every manual implementation duplicates pkg/ functionality
2. **Framework Underutilization**: Not using the rich framework we've built
3. **Complexity**: Manual implementations are more error-prone than tested framework code
4. **Maintainability**: Changes need to be made in multiple places vs. one framework location

By following this guide, the synth plugin will be transformed from a bloated 1781-line example into a clean ~450-line demonstration of how to properly use the ClapGo framework.
