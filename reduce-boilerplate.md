# Reducing Boilerplate in ClapGo Plugins - Phase 2

After successfully implementing the Parameter Binding System, this document identifies remaining boilerplate patterns in the synth plugin that should be abstracted into the ClapGo library.

## Completed ✅

### 1. Parameter Binding System 
- Automatic parameter registration and storage
- Unified text/value conversion
- Eliminated ~150 lines of switch statements

## Remaining Boilerplate Patterns

### 1. Extension Bundle Initialization

**Current State:**
Every plugin must:
- Check if Host is not nil repeatedly
- Initialize each extension individually
- Log initialization status for each
- Handle nil checks throughout

**Lines of boilerplate:** ~40 lines in Init()

**Proposed Abstraction:**
```go
// In pkg/extension/bundle.go
type ExtensionBundle struct {
    ThreadCheck      *thread.Checker
    TrackInfo        *host.TrackInfoProvider
    TransportControl *host.TransportControl
    Tuning          *HostTuning
    Logger          *host.Logger
}

func NewExtensionBundle(host unsafe.Pointer, pluginName string) *ExtensionBundle {
    bundle := &ExtensionBundle{}
    
    if host == nil {
        return bundle
    }
    
    // Initialize all extensions with automatic nil checks
    bundle.Logger = host.NewLogger(host, pluginName)
    bundle.ThreadCheck = thread.NewChecker(host)
    bundle.TrackInfo = host.NewTrackInfoProvider(host)
    bundle.TransportControl = host.NewTransportControl(host)
    bundle.Tuning = NewHostTuning(host)
    
    // Log initialization status once
    if bundle.Logger != nil {
        bundle.logInitStatus()
    }
    
    return bundle
}

// Usage in plugin:
func (p *SynthPlugin) Init() bool {
    thread.DebugSetMainThread()
    p.Extensions = extension.NewExtensionBundle(p.Host, PluginName)
    
    // Set up MIDI mappings...
    return true
}
```

### 2. MIDI CC Mapping System

**Current State:**
Plugins must:
- Call SetCallbacks with inline functions
- Manually map CC numbers to parameters
- Handle value scaling inline
- Repeat similar patterns for each CC

**Lines of boilerplate:** ~30 lines per CC mapping

**Proposed Abstraction:**
```go
// In pkg/audio/ccmapping.go
type CCMapper struct {
    processor *MIDIProcessor
    mappings  map[uint32]CCMapping
}

type CCMapping struct {
    ParamID   uint32
    Atomic    *param.AtomicFloat64
    Transform func(float64) float64 // Optional value transformation
}

func NewCCMapper(processor *MIDIProcessor) *CCMapper {
    mapper := &CCMapper{
        processor: processor,
        mappings:  make(map[uint32]CCMapping),
    }
    
    // Set up the modulation callback
    processor.SetCallbacks(nil, nil, mapper.handleCC, nil, nil)
    return mapper
}

// Builder pattern for easy configuration
func (m *CCMapper) MapCC(cc uint32, paramID uint32, atomic *param.AtomicFloat64) *CCMapper {
    m.mappings[cc] = CCMapping{ParamID: paramID, Atomic: atomic}
    return m
}

func (m *CCMapper) MapCCWithTransform(cc uint32, paramID uint32, atomic *param.AtomicFloat64, transform func(float64) float64) *CCMapper {
    m.mappings[cc] = CCMapping{
        ParamID:   paramID,
        Atomic:    atomic,
        Transform: transform,
    }
    return m
}

// Usage in plugin:
ccMapper := audio.NewCCMapper(p.midiProcessor).
    MapCC(7, 1, p.volume).  // CC7 -> Volume
    MapCCWithTransform(74, 7, p.cutoff, audio.ExpFrequencyMap(20, 20000)) // CC74 -> Cutoff with exponential mapping
```

### 3. Filter Processing Abstraction

**Current State:**
Plugins must:
- Switch between filter types manually
- Check for NaN/Inf after each sample
- Reset filter on errors
- Handle filter type switching in process loop

**Lines of boilerplate:** ~30 lines in Process()

**Proposed Abstraction:**
```go
// In pkg/audio/selectablefilter.go
type SelectableFilter struct {
    filter     *StateVariableFilter
    filterType FilterType
    safeMode   bool // Auto NaN/Inf protection
}

type FilterType int

const (
    FilterLowpass FilterType = iota
    FilterHighpass
    FilterBandpass
    FilterNotch
    FilterBypass
)

func (f *SelectableFilter) Process(input float64) float64 {
    var output float64
    
    switch f.filterType {
    case FilterLowpass:
        output = f.filter.ProcessLowpass(input)
    case FilterHighpass:
        output = f.filter.ProcessHighpass(input)
    case FilterBandpass:
        output = f.filter.ProcessBandpass(input)
    case FilterNotch:
        _, _, _, output = f.filter.Process(input)
    default:
        output = input
    }
    
    // Built-in safety when enabled
    if f.safeMode && (math.IsNaN(output) || math.IsInf(output, 0)) {
        f.filter.Reset()
        return 0
    }
    
    return output
}

// Usage in plugin:
p.filter = audio.NewSelectableFilter(sampleRate, true) // true = safe mode

// In Process:
p.filter.SetType(audio.FilterType(filterType))
for i := range output {
    output[i] = float32(p.filter.Process(float64(output[i])))
}
```

### 4. Process Method Boilerplate

**Current State:**
Every Process method must:
- Check activation state
- Process events
- Load all parameters
- Update voice parameters
- Apply filter with safety checks
- Distribute to channels
- Check for finished voices

**Lines of boilerplate:** ~50 lines of repeated patterns

**Proposed Abstraction:**
```go
// In pkg/audio/processor.go
type SynthProcessor struct {
    voiceManager *VoiceManager
    oscillator   *PolyphonicOscillator
    filter       *SelectableFilter
}

type ProcessParams struct {
    Volume      float64
    Waveform    int
    FilterType  int
    FilterCutoff float64
    FilterResonance float64
    Attack, Decay, Sustain, Release float64
}

func (sp *SynthProcessor) Process(params ProcessParams, frameCount uint32, audioOut [][]float32, events *event.EventProcessor) int {
    // Update voice envelopes
    sp.voiceManager.ApplyToAllVoices(func(voice *Voice) {
        voice.Envelope.SetADSR(params.Attack, params.Decay, params.Sustain, params.Release)
    })
    
    // Configure processing
    sp.oscillator.SetWaveform(WaveformType(params.Waveform))
    sp.filter.SetType(FilterType(params.FilterType))
    sp.filter.SetFrequency(Clamp(params.FilterCutoff, 20.0, 20000.0))
    sp.filter.SetResonance(0.5 + params.FilterResonance*19.5)
    
    // Generate and filter audio
    output := sp.oscillator.Process(frameCount)
    for i := range output {
        output[i] = float32(sp.filter.Process(float64(output[i])))
    }
    
    // Apply volume and distribute
    DistributeToChannels(output, audioOut, float32(params.Volume))
    
    // Cleanup finished voices
    sp.voiceManager.CleanupFinishedVoices(events)
    
    if sp.voiceManager.GetActiveVoiceCount() == 0 {
        return process.ProcessSleep
    }
    return process.ProcessContinue
}

// Usage in plugin - Process becomes much simpler:
func (p *SynthPlugin) Process(...) int {
    if !p.IsActivated || !p.IsProcessing {
        return process.ProcessError
    }
    
    if events != nil {
        p.midiProcessor.ProcessEvents(events, p)
    }
    
    params := audio.ProcessParams{
        Volume:          p.volume.Load(),
        Waveform:        int(p.waveform.Load()),
        FilterType:      int(p.filterType.Load()),
        FilterCutoff:    p.cutoff.Load(),
        FilterResonance: p.resonance.Load(),
        Attack:          p.attack.Load(),
        Decay:           p.decay.Load(),
        Sustain:         p.sustain.Load(),
        Release:         p.release.Load(),
    }
    
    return p.processor.Process(params, frameCount, audioOut, events)
}
```

### 5. Plugin Info Duplication

**Current State:**
Plugin info is defined in:
- constants.go
- NewSynthPlugin() 
- GetPluginInfo()

**Proposed Solution:**
```go
// In pkg/plugin/info.go
type PluginInfoBuilder struct {
    info plugin.Info
}

func NewPluginInfo(id, name, vendor, version string) *PluginInfoBuilder {
    return &PluginInfoBuilder{
        info: plugin.Info{
            ID:      id,
            Name:    name,
            Vendor:  vendor,
            Version: version,
        },
    }
}

// Usage in constants.go:
var PluginInfo = plugin.NewPluginInfo(PluginID, PluginName, PluginVendor, PluginVersion).
    WithDescription(PluginDescription).
    WithURLs("https://github.com/justyntemme/clapgo").
    WithFeatures(plugin.FeatureInstrument, plugin.FeatureStereo).
    Build()

// Then use PluginInfo everywhere
```

### 6. Common DSP Utilities

**Current State:**
Plugins implement their own:
- findPeak function
- Channel distribution logic
- Parameter clamping
- Frequency mapping

**Proposed Solution:**
```go
// In pkg/audio/dsp.go
package audio

// Common DSP utilities
func FindPeak(buffer []float32) float32
func DistributeToChannels(mono []float32, stereo [][]float32, gain float32)
func ExpFrequencyMap(min, max float64) func(float64) float64
func DbToLinear(db float64) float64
func LinearToDb(linear float64) float64
```

## Implementation Priority

1. **Extension Bundle** (High) - Reduces Init() from 100+ lines to ~10 lines
2. **Filter Abstraction** (High) - Eliminates error-prone NaN checking
3. **Process Pipeline** (Medium) - Reduces Process() by 50%
4. **MIDI CC Mapping** (Medium) - Makes MIDI learn trivial
5. **Plugin Info Builder** (Low) - Eliminates duplication
6. **DSP Utilities** (Low) - Prevents reinventing the wheel

## Benefits Summary

With all abstractions implemented:
- **Synth plugin**: ~760 lines → ~400 lines (47% reduction)
- **New plugin creation**: Hours → Minutes
- **Common bugs eliminated**: NaN checks, parameter updates, channel distribution
- **Focus on DSP**: Developers write audio algorithms, not infrastructure

## Example: Minimal Synth with All Abstractions

```go
type MinimalSynth struct {
    *plugin.PluginBase
    *extension.ExtensionBundle
    
    params    *param.ParameterBinder
    processor *audio.SynthProcessor
    ccMapper  *audio.CCMapper
    
    // Direct parameter access
    volume   *param.AtomicFloat64
    cutoff   *param.AtomicFloat64
}

func NewMinimalSynth() *MinimalSynth {
    s := &MinimalSynth{
        PluginBase: plugin.NewPluginBase(PluginInfo),
        processor:  audio.NewSynthProcessor(16, 44100),
    }
    
    // Bind parameters
    s.params = param.NewParameterBinder(s.ParamManager)
    s.volume = s.params.BindPercentage(1, "Volume", 70.0)
    s.cutoff = s.params.BindCutoff(2, "Cutoff", 1000.0)
    
    // Setup MIDI
    s.ccMapper = audio.NewCCMapper(s.processor.MIDIProcessor).
        MapCC(7, 1, s.volume).
        MapCC(74, 2, s.cutoff)
    
    return s
}

func (s *MinimalSynth) Init() bool {
    s.ExtensionBundle = extension.NewExtensionBundle(s.Host, PluginName)
    return true
}

func (s *MinimalSynth) Process(...) int {
    return s.processor.ProcessSimple(
        s.volume.Load(),
        s.cutoff.Load(),
        frameCount,
        audioOut,
        events,
    )
}

// That's it! A working synth in ~50 lines
```