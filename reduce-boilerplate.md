# Reducing Boilerplate in ClapGo Plugins

This document identifies common patterns in the synth plugin that could be abstracted into the ClapGo library to reduce boilerplate code and make plugin development easier.

## 1. Parameter Management Boilerplate

### Current State
Plugin developers must:
- Define atomic storage for each parameter
- Initialize atomic parameters in the constructor
- Handle parameter updates in HandleParamValue with switch statements
- Implement ParamValueToText for custom formatting
- Implement ParamTextToValue for custom parsing

### Proposed Abstraction: Parameter Binding System

```go
// In the library (pkg/param/binding.go)
type ParameterBinding struct {
    ID        uint32
    Atomic    *AtomicFloat64
    Format    FormatType
    Choices   []string // For choice parameters
    OnChange  func(value float64) // Optional callback
}

type ParameterBinder struct {
    bindings map[uint32]*ParameterBinding
    manager  *ParamManager
}

// Usage in plugin:
type SynthPlugin struct {
    *plugin.PluginBase
    params *param.ParameterBinder
    
    // Direct access to parameters
    volume     *param.AtomicFloat64
    filterType *param.AtomicFloat64
}

func NewSynthPlugin() *SynthPlugin {
    p := &SynthPlugin{
        PluginBase: plugin.NewPluginBase(info),
    }
    
    // Create parameter binder
    p.params = param.NewParameterBinder(p.ParamManager)
    
    // Bind parameters - this creates atomic storage AND registers with manager
    p.volume = p.params.BindPercentage(1, "Volume", 70.0)
    p.filterType = p.params.BindChoice(9, "Filter Type", []string{
        "Lowpass", "Highpass", "Bandpass", "Notch",
    }, 0)
    
    return p
}
```

The ParameterBinder would automatically:
- Create atomic storage
- Register with ParamManager
- Handle value updates in a generic HandleParamValue
- Provide automatic text/value conversion based on format type
- Support choice parameters with string arrays

## 2. Filter Chain Abstraction

### Current State
Plugin developers must:
- Manually switch between filter types in the process loop
- Handle NaN/Inf checking for each filter output
- Reset filter state on errors

### Proposed Abstraction: Filter Chain with Type Selection

```go
// In the library (pkg/audio/filterchain.go)
type FilterType int

const (
    FilterTypeLowpass FilterType = iota
    FilterTypeHighpass
    FilterTypeBandpass
    FilterTypeNotch
    FilterTypeBypass
)

type SelectableFilter struct {
    filter     *StateVariableFilter
    filterType FilterType
    sampleRate float64
}

func (f *SelectableFilter) Process(input float64) float64 {
    var output float64
    
    switch f.filterType {
    case FilterTypeLowpass:
        output = f.filter.ProcessLowpass(input)
    case FilterTypeHighpass:
        output = f.filter.ProcessHighpass(input)
    case FilterTypeBandpass:
        output = f.filter.ProcessBandpass(input)
    case FilterTypeNotch:
        _, _, _, output = f.filter.Process(input)
    default:
        output = input
    }
    
    // Built-in safety
    if math.IsNaN(output) || math.IsInf(output, 0) {
        f.filter.Reset()
        return 0
    }
    
    return output
}

// Usage in plugin:
filter := audio.NewSelectableFilter(sampleRate)
filter.SetType(audio.FilterType(filterTypeParam))
filter.SetFrequency(cutoff)
filter.SetResonance(resonance)

// Process is now one line
output[i] = float32(filter.Process(float64(input[i])))
```

## 3. Standard Audio Processing Pipeline

### Current State
Plugin developers must:
- Get all parameter values at the start of process
- Update voice parameters manually
- Apply filter to output manually
- Apply volume and copy to channels manually
- Check for finished voices and send note end events

### Proposed Abstraction: ProcessorChain

```go
// In the library (pkg/audio/processor.go)
type ProcessorChain struct {
    voiceManager *VoiceManager
    oscillator   *PolyphonicOscillator
    filter       *SelectableFilter
    params       *ParameterBinder
}

type ProcessConfig struct {
    Volume      *AtomicFloat64
    FilterType  *AtomicFloat64
    FilterCutoff *AtomicFloat64
    FilterResonance *AtomicFloat64
    // ... other parameters
}

func (p *ProcessorChain) Process(config ProcessConfig, frameCount uint32, audioOut [][]float32, events *event.EventProcessor) []float32 {
    // Get parameters once
    volume := config.Volume.Load()
    filterType := int(config.FilterType.Load())
    
    // Update filter
    p.filter.SetType(FilterType(filterType))
    p.filter.SetFrequency(config.FilterCutoff.Load())
    p.filter.SetResonance(config.FilterResonance.Load())
    
    // Generate and filter
    output := p.oscillator.Process(frameCount)
    for i := range output {
        output[i] = float32(p.filter.Process(float64(output[i])))
    }
    
    // Apply volume and distribute to channels
    audio.DistributeToChannels(output, audioOut, float32(volume))
    
    // Handle finished voices
    p.voiceManager.CleanupFinishedVoices(events)
    
    return output
}
```

## 4. Extension Initialization Boilerplate

### Current State
Plugin developers must:
- Check if host is available
- Create each extension helper
- Log initialization status
- Handle nil checks throughout

### Proposed Abstraction: Extension Bundle

```go
// In the library (pkg/extension/bundle.go)
type ExtensionBundle struct {
    ThreadCheck      *thread.Checker
    TrackInfo        *host.TrackInfoProvider
    TransportControl *host.TransportControl
    Tuning          *HostTuning
    Logger          *host.Logger
}

func NewExtensionBundle(host unsafe.Pointer) *ExtensionBundle {
    bundle := &ExtensionBundle{}
    
    if host == nil {
        return bundle
    }
    
    // Initialize all extensions with nil checks
    bundle.ThreadCheck = thread.NewChecker(host)
    bundle.TrackInfo = host.NewTrackInfoProvider(host)
    bundle.TransportControl = host.NewTransportControl(host)
    bundle.Tuning = NewHostTuning(host)
    bundle.Logger = host.NewLogger(host, "Plugin")
    
    // Log initialization status
    if bundle.Logger != nil {
        bundle.logInitStatus()
    }
    
    return bundle
}

// Usage in plugin:
func (p *SynthPlugin) Init() bool {
    p.Extensions = extension.NewExtensionBundle(p.Host)
    // All extensions are now available through p.Extensions
    return true
}
```

## 5. MIDI Processing Setup

### Current State
Plugin developers must:
- Create MIDI processor
- Set up callbacks with inline functions
- Handle special cases (transport control, CC mappings)

### Proposed Abstraction: MIDI Mapping Builder

```go
// In the library (pkg/audio/midimapping.go)
type MIDIMapping struct {
    processor *MIDIProcessor
    ccMap     map[uint32]CCMapping
}

type CCMapping struct {
    ParamID    uint32
    Atomic     *AtomicFloat64
    MapFunc    func(ccValue float64) float64 // Optional value mapping
}

func NewMIDIMapping(processor *MIDIProcessor) *MIDIMapping {
    return &MIDIMapping{
        processor: processor,
        ccMap:     make(map[uint32]CCMapping),
    }
}

func (m *MIDIMapping) MapCC(cc uint32, paramID uint32, atomic *AtomicFloat64) *MIDIMapping {
    m.ccMap[cc] = CCMapping{ParamID: paramID, Atomic: atomic}
    return m
}

func (m *MIDIMapping) MapCCWithFunction(cc uint32, paramID uint32, atomic *AtomicFloat64, mapFunc func(float64) float64) *MIDIMapping {
    m.ccMap[cc] = CCMapping{ParamID: paramID, Atomic: atomic, MapFunc: mapFunc}
    return m
}

// Usage in plugin:
midiMapping := audio.NewMIDIMapping(p.midiProcessor).
    MapCC(7, 1, p.volume).  // CC7 -> Volume
    MapCCWithFunction(74, 7, p.cutoff, func(value float64) float64 {
        // Exponential mapping for filter cutoff
        return 20.0 * math.Pow(1000.0, value)
    })
```

## 6. Parameter Text Formatting

### Current State
Plugin developers must implement large switch statements for ParamValueToText and ParamTextToValue.

### Proposed Solution
The ParameterBinder (from section 1) would handle this automatically based on parameter type.

## Implementation Guide

### Phase 1: Parameter Binding System (Priority: High)
1. Create `pkg/param/binding.go` with ParameterBinding and ParameterBinder types
2. Add convenience methods for all parameter types (Percentage, Choice, ADSR, etc.)
3. Implement automatic HandleParamValue routing
4. Add automatic text conversion based on format type
5. Update examples to use the new system

### Phase 2: Filter Abstractions (Priority: Medium)
1. Create `pkg/audio/filterchain.go` with SelectableFilter
2. Add built-in NaN/Inf protection
3. Support for filter chains and parallel processing
4. Update synth example to use SelectableFilter

### Phase 3: Audio Processing Pipeline (Priority: Medium)
1. Create `pkg/audio/processor.go` with ProcessorChain
2. Add common audio distribution utilities
3. Add voice cleanup utilities
4. Create templates for common synth architectures

### Phase 4: Extension Bundle (Priority: Low)
1. Create `pkg/extension/bundle.go`
2. Consolidate all extension initialization
3. Add convenience methods for common patterns
4. Update examples

### Phase 5: MIDI Mapping (Priority: Low)
1. Create `pkg/audio/midimapping.go`
2. Add builder pattern for CC mappings
3. Support for custom mapping functions
4. Integration with ParameterBinder

## Benefits

1. **Reduced Code**: The synth plugin would shrink from ~880 lines to ~400 lines
2. **Fewer Errors**: Automatic handling of common issues (NaN, parameter updates)
3. **Faster Development**: Developers focus on DSP, not boilerplate
4. **Consistency**: All plugins handle parameters and filters the same way
5. **Type Safety**: Builder patterns ensure correct usage

## Example: Simplified Synth Plugin

With these abstractions, the synth plugin would look like:

```go
type SynthPlugin struct {
    *plugin.PluginBase
    *extension.ExtensionBundle
    
    params    *param.ParameterBinder
    processor *audio.ProcessorChain
    
    // Direct parameter access
    volume      *param.AtomicFloat64
    filterType  *param.AtomicFloat64
    filterCutoff *param.AtomicFloat64
}

func NewSynthPlugin() *SynthPlugin {
    p := &SynthPlugin{
        PluginBase: plugin.NewPluginBase(info),
    }
    
    // Bind all parameters
    p.params = param.NewParameterBinder(p.ParamManager)
    p.volume = p.params.BindPercentage(1, "Volume", 70.0)
    p.filterType = p.params.BindChoice(9, "Filter Type", 
        []string{"Lowpass", "Highpass", "Bandpass", "Notch"}, 0)
    p.filterCutoff = p.params.BindCutoff(7, "Filter Cutoff", 1000.0)
    
    // Create processor chain
    p.processor = audio.NewProcessorChain(16, 44100) // 16 voices, 44.1kHz
    
    return p
}

func (p *SynthPlugin) Process(...) int {
    // Process events
    p.processor.ProcessEvents(events)
    
    // Generate audio with automatic filtering, volume, and distribution
    config := audio.ProcessConfig{
        Volume: p.volume,
        FilterType: p.filterType,
        FilterCutoff: p.filterCutoff,
    }
    p.processor.Process(config, frameCount, audioOut, events)
    
    return process.ProcessContinue
}
```

This reduces boilerplate by ~60% while maintaining flexibility for custom DSP.