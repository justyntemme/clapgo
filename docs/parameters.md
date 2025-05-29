# ClapGo Parameter Guide

This guide covers everything you need to know about parameters in ClapGo, from basic registration to advanced automation and binding patterns.

## Table of Contents

- [Parameter Basics](#parameter-basics)
- [Parameter Registration](#parameter-registration)
- [Parameter Factories](#parameter-factories)
- [Parameter Binding System](#parameter-binding-system)
- [Thread-Safe Parameter Access](#thread-safe-parameter-access)
- [Parameter Formatting and Display](#parameter-formatting-and-display)
- [Advanced Parameter Patterns](#advanced-parameter-patterns)
- [Common Parameter Types](#common-parameter-types)

## Parameter Basics

### Parameter Info Structure

Every parameter in ClapGo is defined by a `param.Info` struct:

```go
type Info struct {
    ID           uint32  // Unique parameter ID
    Name         string  // Display name
    Module       string  // Grouping path (e.g., "Filter/Cutoff")
    MinValue     float64 // Minimum value
    MaxValue     float64 // Maximum value
    DefaultValue float64 // Default value
    Flags        uint32  // Parameter capabilities
}
```

### Parameter Flags

Flags define parameter behavior and capabilities:

```go
const (
    FlagAutomatable     uint32 = 1 << 0  // Can be automated by host
    FlagModulatable     uint32 = 1 << 1  // Can be modulated (LFO, etc.)
    FlagStepped         uint32 = 1 << 2  // Discrete values (switches, choices)
    FlagReadonly        uint32 = 1 << 3  // Cannot be changed by user
    FlagHidden          uint32 = 1 << 4  // Not shown in UI
    FlagBypass          uint32 = 1 << 5  // This is a bypass parameter
    FlagBoundedBelow    uint32 = 1 << 6  // Minimum value is enforced
    FlagBoundedAbove    uint32 = 1 << 7  // Maximum value is enforced
    FlagRequiresProcess uint32 = 1 << 8  // Changes require audio processing
)

// Common flag combinations
param.FlagAutomatable | param.FlagModulatable | param.FlagBoundedBelow | param.FlagBoundedAbove
```

## Parameter Registration

### Basic Registration

```go
func ExampleBasicRegistration() {
    // Create parameter manager
    manager := param.NewManager()
    
    // Define parameter manually
    gainParam := param.Info{
        ID:           1,
        Name:         "Gain",
        Module:       "",
        MinValue:     0.0,
        MaxValue:     2.0,
        DefaultValue: 1.0,
        Flags:        param.FlagAutomatable | param.FlagBoundedBelow | param.FlagBoundedAbove,
    }
    
    // Register parameter
    err := manager.Register(gainParam)
    if err != nil {
        // Handle registration error
        panic("Failed to register gain parameter: " + err.Error())
    }
    
    // Register multiple parameters at once
    cutoffParam := param.Info{
        ID:           2,
        Name:         "Cutoff",
        Module:       "Filter",
        MinValue:     20.0,
        MaxValue:     20000.0,
        DefaultValue: 1000.0,
        Flags:        param.FlagAutomatable | param.FlagBoundedBelow | param.FlagBoundedAbove,
    }
    
    err = manager.RegisterAll(gainParam, cutoffParam)
}
```

### Using in Plugin

```go
type MyPlugin struct {
    *plugin.PluginBase
    // Parameter storage
    gain   param.AtomicFloat64
    cutoff param.AtomicFloat64
}

func NewMyPlugin() *MyPlugin {
    info := plugin.Info{
        ID:          "com.example.myplugin",
        Name:        "My Plugin",
        // ... other plugin info
    }
    
    p := &MyPlugin{
        PluginBase: plugin.NewPluginBase(info),
    }
    
    // Register parameters
    err := p.ParamManager.Register(param.Volume(1, "Gain"))
    if err != nil {
        panic("Failed to register gain: " + err.Error())
    }
    
    err = p.ParamManager.Register(param.Cutoff(2, "Filter Cutoff"))
    if err != nil {
        panic("Failed to register cutoff: " + err.Error())
    }
    
    // Initialize atomic storage with default values
    p.gain.Store(1.0)
    p.cutoff.Store(1000.0)
    
    return p
}
```

## Parameter Factories

ClapGo provides factory functions for common parameter types:

### Volume and Gain Parameters

```go
// Volume parameter (0-2 linear gain, displays as dB)
volume := param.Volume(1, "Volume")
// Results in: 0.0-2.0 range, default 1.0 (0dB)

// Master gain with custom name
masterGain := param.Volume(10, "Master")
```

### Frequency Parameters

```go
// Basic frequency parameter (linear Hz)
freq := param.Frequency(2, "Frequency", 20.0, 20000.0, 1000.0)

// Filter cutoff (20Hz-20kHz, default 1kHz)
cutoff := param.Cutoff(3, "Cutoff")

// Musical filter cutoff (20Hz-8kHz, optimized for musical content)
musicalCutoff := param.CutoffMusical(4, "Filter")

// Logarithmic cutoff (0-1 parameter maps to 20Hz-8kHz exponentially)
logCutoff := param.CutoffLog(5, "Cutoff")

// Full spectrum logarithmic (0-1 maps to 20Hz-20kHz)
fullLogCutoff := param.CutoffLogFull(6, "Cutoff")
```

### Filter Parameters

```go
// Resonance parameter (0-1, represents Q factor)
resonance := param.Resonance(7, "Resonance")
```

### Envelope Parameters

```go
// ADSR envelope parameters
attack := param.ADSR(8, "Attack", 2.0)    // Max 2 seconds
decay := param.ADSR(9, "Decay", 2.0)      // Max 2 seconds
release := param.ADSR(10, "Release", 5.0) // Max 5 seconds

// Sustain is typically a percentage
sustain := param.Percentage(11, "Sustain", 70.0) // Default 70%
```

### Control Parameters

```go
// On/off switch
bypassSwitch := param.Switch(12, "Bypass", false)

// Multiple choice parameter
waveform := param.Choice(13, "Waveform", 3, 0) // 3 options, default 0

// Percentage parameter
mix := param.Percentage(14, "Mix", 50.0) // Default 50%

// Pan parameter (-1 to +1)
pan := param.Pan(15, "Pan")

// Bypass parameter (special handling)
bypass := param.Bypass(0) // Usually ID 0
```

## Parameter Binding System

The binding system automates parameter management:

### Basic Binding

```go
type MyPlugin struct {
    *plugin.PluginBase
    
    // Parameter binder manages everything
    params *param.ParameterBinder
    
    // Direct atomic access
    volume     *param.AtomicFloat64
    waveform   *param.AtomicFloat64
    cutoff     *param.AtomicFloat64
    resonance  *param.AtomicFloat64
}

func NewMyPlugin() *MyPlugin {
    p := &MyPlugin{
        PluginBase: plugin.NewPluginBase(info),
    }
    
    // Create parameter binder
    p.params = param.NewParameterBinder(p.ParamManager)
    
    // Bind parameters - this does registration AND creates atomic storage
    p.volume = p.params.BindPercentage(1, "Volume", 70.0)
    p.waveform = p.params.BindChoice(2, "Waveform", 
        []string{"Sine", "Saw", "Square", "Triangle"}, 0)
    p.cutoff = p.params.BindCutoffLog(3, "Filter Cutoff", 0.5)
    p.resonance = p.params.BindResonance(4, "Filter Resonance", 0.5)
    
    return p
}
```

### Advanced Binding

```go
// Custom linear parameter
gain := p.params.BindLinear(5, "Gain", -20.0, 20.0, 0.0)

// Decibel parameter
outputLevel := p.params.BindDB(6, "Output", -60.0, 12.0, 0.0)

// ADSR parameters
attack := p.params.BindADSR(7, "Attack", 2.0, 0.01)    // Max 2s, default 10ms
decay := p.params.BindADSR(8, "Decay", 2.0, 0.1)       // Max 2s, default 100ms
release := p.params.BindADSR(9, "Release", 5.0, 0.3)   // Max 5s, default 300ms
```

### Parameter Callbacks

```go
// Set callback for parameter changes
p.params.SetCallback(1, func(value float64) {
    // Called when volume parameter changes
    fmt.Printf("Volume changed to: %.2f%%\n", value*100)
})

p.params.SetCallback(3, func(value float64) {
    // Called when cutoff changes
    mappedFreq, _ := p.params.GetMappedValue(3)
    fmt.Printf("Cutoff changed to: %.1f Hz\n", mappedFreq)
})
```

## Thread-Safe Parameter Access

### Atomic Parameter Storage

```go
// AtomicFloat64 provides lock-free parameter access
type AtomicFloat64 struct {
    value int64 // Stores float64 bits as int64
}

func ExampleAtomicAccess() {
    var gain param.AtomicFloat64
    
    // Initialize
    gain.Store(0.8)
    
    // Thread-safe read (audio thread)
    currentGain := gain.Load()
    
    // Thread-safe write (UI thread)
    gain.Store(0.9)
    
    // Update with parameter manager sync
    gain.UpdateWithManager(0.75, manager, paramID)
}
```

### Usage in Audio Processing

```go
func (p *MyPlugin) Process(steadyTime int64, framesCount uint32, 
    audioIn, audioOut [][]float32, events *event.Processor) int {
    
    // Get current parameter values atomically
    volume := p.volume.Load()
    cutoffParam := p.cutoff.Load()
    
    // Get mapped frequency value (applies logarithmic mapping)
    cutoffFreq, _ := p.params.GetMappedValue(3)
    
    // Process audio using parameters
    for ch := range audioOut {
        for i := uint32(0); i < framesCount; i++ {
            sample := audioIn[ch][i]
            
            // Apply volume
            sample *= float32(volume)
            
            // Apply filtering using cutoffFreq
            // ... filter processing ...
            
            audioOut[ch][i] = sample
        }
    }
    
    return process.ProcessContinue
}
```

### Parameter Event Handling

```go
// Handle parameter changes from automation
func (p *MyPlugin) HandleParamValue(paramEvent *event.ParamValueEvent, time uint32) {
    // Use binder for automatic handling
    if p.params.HandleParamValue(paramEvent.ParamID, paramEvent.Value) {
        return // Handled by binder
    }
    
    // Custom handling for special parameters
    switch paramEvent.ParamID {
    case SpecialParamID:
        // Custom processing
        p.specialParam.Store(paramEvent.Value)
    }
}
```

## Parameter Formatting and Display

### Automatic Formatting

The binding system provides automatic text formatting:

```go
// Get formatted text for display
text, ok := p.params.ValueToText(1, 0.75)
if ok {
    // text might be "75%" for percentage parameter
    fmt.Println("Volume:", text)
}

// Parse text input
value, err := p.params.TextToValue(1, "80%")
if err == nil {
    // value is 0.8
    p.volume.Store(value)
}
```

### Custom Formatting

```go
// For choice parameters
waveformValue := p.waveform.Load()
waveformText, _ := p.params.ValueToText(2, waveformValue)
// waveformText might be "Sine", "Saw", etc.

// For frequency parameters with logarithmic mapping
cutoffValue := p.cutoff.Load()        // Raw parameter value (0-1)
cutoffFreq, _ := p.params.GetMappedValue(3) // Mapped frequency
cutoffText, _ := p.params.ValueToText(3, cutoffValue) // "1.2 kHz"
```

### Format Types

ClapGo supports various parameter format types:

```go
const (
    FormatDefault     Format = iota
    FormatPercentage          // 0.5 → "50%"
    FormatDecibels           // 0.5 → "-6.0 dB"
    FormatHertz              // 1000 → "1.0 kHz"
    FormatKilohertz          // 1000 → "1.00 kHz"
    FormatMilliseconds       // 0.1 → "100 ms"
    FormatSeconds            // 0.1 → "100 ms" or "1.5 s"
)
```

## Advanced Parameter Patterns

### Value Mapping

Some parameters use value mapping for better user experience:

```go
// Logarithmic frequency mapping
type ValueMapper func(float64) float64

var FrequencyMappers = struct {
    LogMusical ValueMapper // 0-1 → 20Hz-8kHz
    LogFull    ValueMapper // 0-1 → 20Hz-20kHz
}{
    LogMusical: func(param float64) float64 {
        // Maps 0-1 to 20Hz-8kHz logarithmically
        return 20.0 * math.Pow(400.0, param) // 20 * (8000/20)^param
    },
    LogFull: func(param float64) float64 {
        // Maps 0-1 to 20Hz-20kHz logarithmically
        return 20.0 * math.Pow(1000.0, param) // 20 * (20000/20)^param
    },
}

// Using mapped values
cutoffParam := p.cutoff.Load()               // 0-1 range
cutoffFreq, _ := p.params.GetMappedValue(3)  // Actual Hz value
```

### Polyphonic Parameters

For polyphonic instruments, parameters can target specific voices:

```go
func (p *SynthPlugin) HandleParamValue(paramEvent *event.ParamValueEvent, time uint32) {
    // Check if this targets a specific note/voice
    if paramEvent.NoteID >= 0 || paramEvent.Key >= 0 {
        p.handlePolyphonicParameter(*paramEvent)
        return
    }
    
    // Global parameter change
    p.params.HandleParamValue(paramEvent.ParamID, paramEvent.Value)
}

func (p *SynthPlugin) handlePolyphonicParameter(paramEvent event.ParamValueEvent) {
    // Apply to matching voices
    p.voiceManager.ApplyToAllVoices(func(voice *audio.Voice) {
        // Match by note ID, key, or channel
        if paramEvent.NoteID >= 0 && voice.NoteID != paramEvent.NoteID {
            return
        }
        
        switch paramEvent.ParamID {
        case ParamVolume:
            voice.Volume = paramEvent.Value
        case ParamPitchBend:
            voice.PitchBend = paramEvent.Value*2.0 - 1.0 // 0-1 to -1 to +1
        }
    })
}
```

### Parameter Groups and Modules

Organize parameters using modules:

```go
// Filter parameters
filterCutoff := param.Info{
    ID:     10,
    Name:   "Cutoff",
    Module: "Filter",
    // ... other fields
}

filterResonance := param.Info{
    ID:     11,
    Name:   "Resonance", 
    Module: "Filter",
    // ... other fields
}

// Envelope parameters
envAttack := param.Info{
    ID:     20,
    Name:   "Attack",
    Module: "Envelope",
    // ... other fields
}
```

## Common Parameter Types

### Complete Parameter Set Example

Here's a complete example showing common parameter types for a synthesizer:

```go
func (p *SynthPlugin) registerParameters() {
    // Create parameter binder
    p.params = param.NewParameterBinder(p.ParamManager)
    
    // Master controls
    p.volume = p.params.BindPercentage(1, "Volume", 70.0)
    p.pan = p.params.BindLinear(2, "Pan", -1.0, 1.0, 0.0)
    
    // Oscillator
    p.waveform = p.params.BindChoice(10, "Waveform", 
        []string{"Sine", "Sawtooth", "Square", "Triangle"}, 1)
    p.tune = p.params.BindLinear(11, "Tune", -24.0, 24.0, 0.0) // Semitones
    p.fine = p.params.BindLinear(12, "Fine", -100.0, 100.0, 0.0) // Cents
    
    // Filter
    p.filterType = p.params.BindChoice(20, "Filter Type",
        []string{"Lowpass", "Highpass", "Bandpass", "Notch"}, 0)
    p.cutoff = p.params.BindCutoffLog(21, "Cutoff", 0.5)
    p.resonance = p.params.BindResonance(22, "Resonance", 0.3)
    p.filterEnvAmount = p.params.BindLinear(23, "Env Amount", -1.0, 1.0, 0.0)
    
    // Amplitude Envelope
    p.ampAttack = p.params.BindADSR(30, "Attack", 2.0, 0.01)
    p.ampDecay = p.params.BindADSR(31, "Decay", 5.0, 0.1)
    p.ampSustain = p.params.BindPercentage(32, "Sustain", 70.0)
    p.ampRelease = p.params.BindADSR(33, "Release", 10.0, 0.3)
    
    // Filter Envelope
    p.filterAttack = p.params.BindADSR(40, "F Attack", 2.0, 0.01)
    p.filterDecay = p.params.BindADSR(41, "F Decay", 5.0, 0.1)
    p.filterSustain = p.params.BindPercentage(42, "F Sustain", 50.0)
    p.filterRelease = p.params.BindADSR(43, "F Release", 10.0, 0.3)
    
    // Effects
    p.chorusRate = p.params.BindLinear(50, "Chorus Rate", 0.1, 10.0, 2.0)
    p.chorusDepth = p.params.BindPercentage(51, "Chorus Depth", 30.0)
    p.reverbSize = p.params.BindPercentage(52, "Reverb Size", 50.0)
    p.reverbMix = p.params.BindPercentage(53, "Reverb Mix", 20.0)
    
    // Bypass
    p.bypass = p.params.BindChoice(0, "Bypass", []string{"Off", "On"}, 0)
}
```

This comprehensive parameter guide covers all aspects of parameter management in ClapGo. The binding system makes it easy to create robust, thread-safe parameter handling with minimal boilerplate code.