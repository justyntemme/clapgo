# Phase 3: Function-Level Boilerplate Reduction

After implementing Extension Bundle and Selectable Filter, this document analyzes all 23 functions in the synth plugin to identify which can be moved to the framework.

## Function Analysis

### ✅ Already Optimized (5 functions)
1. `init()` - Global initialization, must stay
2. `NewSynthPlugin()` - Constructor, plugin-specific 
3. `Init()` - Now uses ExtensionBundle (3 lines vs 40)
4. `Process()` - Uses SelectableFilter, still needs more work
5. `GetPluginInfo()` - Plugin metadata, must stay

### 🎯 High-Priority Boilerplate (8 functions)

#### 1. **Parameter Text Conversion** (100% boilerplate)
```go
// Current: 30+ lines each
func (p *SynthPlugin) ParamValueToText(...) bool
func (p *SynthPlugin) ParamTextToValue(...) bool
```
**Solution**: ParameterBinder already handles this - make it automatic

#### 2. **Event Processing** (90% boilerplate)
```go
// Current: 15 lines
func (p *SynthPlugin) processEventHandler(events *event.EventProcessor, frameCount uint32)
func (p *SynthPlugin) ParamsFlush(inEvents, outEvents unsafe.Pointer)
func (p *SynthPlugin) HandleParamValue(paramEvent *event.ParamValueEvent, time uint32)
```
**Solution**: Create AutoEventHandler that uses ParameterBinder

#### 3. **State Management** (95% boilerplate)
```go
// Current: 20 lines total
func (p *SynthPlugin) SaveState() map[string]interface{}
func (p *SynthPlugin) LoadState(data map[string]interface{})
```
**Solution**: AutoStateManager using ParameterBinder

#### 4. **Lifecycle Methods** (80% boilerplate)
```go
// Current: Simple delegations, 2-5 lines each
func (p *SynthPlugin) Destroy()
func (p *SynthPlugin) Deactivate() 
func (p *SynthPlugin) StartProcessing() bool
func (p *SynthPlugin) StopProcessing()
func (p *SynthPlugin) Reset()
```
**Solution**: DefaultLifecycleHandler

### 🎯 Medium-Priority Boilerplate (6 functions)

#### 5. **Extension Callbacks** (70% boilerplate)
```go
// Current: Mostly logging, 15-20 lines each
func (p *SynthPlugin) OnTrackInfoChanged()
func (p *SynthPlugin) OnTuningChanged() 
func (p *SynthPlugin) OnTimer(timerID uint64)
```
**Solution**: ExtensionBundle can provide default handlers

#### 6. **Plugin Metadata** (60% boilerplate)
```go
// Current: Simple getters, 3-8 lines each
func (p *SynthPlugin) OnMainThread()
func (p *SynthPlugin) GetPluginID() string
func (p *SynthPlugin) GetLatency() uint32
```
**Solution**: AutoMetadataProvider

### 🟡 Plugin-Specific but Improvable (4 functions)

#### 7. **Audio-Specific** (40% boilerplate)
```go
// Current: Plugin-specific but common patterns
func (p *SynthPlugin) Activate(sampleRate float64, minFrames, maxFrames uint32) bool
func (p *SynthPlugin) GetTail() uint32
func (p *SynthPlugin) GetVoiceInfo() extension.VoiceInfo
func (p *SynthPlugin) handlePolyphonicParameter(paramEvent event.ParamValueEvent)
```
**Solution**: PolyphonicInstrumentBase helper

### ⚪ Keep Plugin-Specific (2 functions)
```go
// Current: Truly plugin-specific
func (p *SynthPlugin) GetRemoteControlsPageCount() uint32
func (p *SynthPlugin) GetRemoteControlsPage(...) (...)
```

### 🔧 Utility Functions (2 functions)
```go
// Current: Generic utilities that belong in framework
func findPeak(buffer []float32) float32
func main() // Test function
```
**Solution**: Move to `pkg/audio/dsp.go`

## Implementation Strategy

### Phase 3A: Auto Parameter Handlers

```go
// In pkg/param/autohandler.go
type AutoParamHandler struct {
    binder *ParameterBinder
}

func (h *AutoParamHandler) ParamValueToText(paramID uint32, value float64, buffer unsafe.Pointer, size uint32) bool {
    text, ok := h.binder.ValueToText(paramID, value)
    if !ok {
        return false
    }
    // Copy to C buffer...
    return true
}

func (h *AutoParamHandler) ParamTextToValue(paramID uint32, text string, value unsafe.Pointer) bool {
    parsedValue, err := h.binder.TextToValue(paramID, text)
    if err != nil {
        return false
    }
    *(*float64)(value) = parsedValue
    return true
}

func (h *AutoParamHandler) HandleParamValue(paramEvent *event.ParamValueEvent, time uint32) {
    h.binder.HandleParamValue(paramEvent.ParamID, paramEvent.Value)
}
```

### Phase 3B: Auto Event Processing

```go
// In pkg/event/autoprocessor.go
type AutoEventProcessor struct {
    midiProcessor *audio.MIDIProcessor
    paramHandler  *param.AutoParamHandler
}

func (p *AutoEventProcessor) ProcessEvents(events *event.EventProcessor, plugin param.Handler) {
    p.midiProcessor.ProcessEvents(events, plugin)
}

func (p *AutoEventProcessor) ParamsFlush(inEvents, outEvents unsafe.Pointer) {
    if inEvents != nil {
        eventHandler := event.NewEventProcessor(inEvents, outEvents)
        p.ProcessEvents(eventHandler, p.paramHandler)
    }
}
```

### Phase 3C: Auto State Management

```go
// In pkg/state/automanager.go
type AutoStateManager struct {
    binder *param.ParameterBinder
    customState map[string]interface{}
}

func (s *AutoStateManager) SaveState() map[string]interface{} {
    state := make(map[string]interface{})
    
    // Auto-save all parameters
    for id, binding := range s.binder.GetAllBindings() {
        state[fmt.Sprintf("param_%d", id)] = binding.Atomic.Load()
    }
    
    // Add custom state
    for k, v := range s.customState {
        state[k] = v
    }
    
    return state
}

func (s *AutoStateManager) LoadState(data map[string]interface{}) {
    // Auto-load parameters
    for key, value := range data {
        if strings.HasPrefix(key, "param_") {
            if id, err := strconv.ParseUint(key[6:], 10, 32); err == nil {
                if binding, ok := s.binder.GetBinding(uint32(id)); ok {
                    if floatVal, ok := value.(float64); ok {
                        binding.Atomic.Store(floatVal)
                    }
                }
            }
        }
    }
}
```

### Phase 3D: Polyphonic Instrument Base

```go
// In pkg/plugin/polybase.go
type PolyphonicInstrumentBase struct {
    *PluginBase
    *extension.ExtensionBundle
    
    params          *param.ParameterBinder
    paramHandler    *param.AutoParamHandler
    eventProcessor  *event.AutoEventProcessor
    stateManager    *state.AutoStateManager
    voiceManager    *audio.VoiceManager
    notePortManager *audio.NotePortManager
}

func NewPolyphonicInstrumentBase(info plugin.Info, polyphony int) *PolyphonicInstrumentBase {
    base := &PolyphonicInstrumentBase{
        PluginBase:      plugin.NewPluginBase(info),
        voiceManager:    audio.NewVoiceManager(polyphony, 44100),
        notePortManager: audio.NewNotePortManager(),
    }
    
    base.params = param.NewParameterBinder(base.ParamManager)
    base.paramHandler = param.NewAutoParamHandler(base.params)
    base.eventProcessor = event.NewAutoEventProcessor(base.paramHandler)
    base.stateManager = state.NewAutoStateManager(base.params)
    
    return base
}

// Auto-implemented methods
func (p *PolyphonicInstrumentBase) Init() bool {
    p.ExtensionBundle = extension.NewExtensionBundle(p.Host, p.Info.Name)
    return true
}

func (p *PolyphonicInstrumentBase) ParamValueToText(paramID uint32, value float64, buffer unsafe.Pointer, size uint32) bool {
    return p.paramHandler.ParamValueToText(paramID, value, buffer, size)
}

func (p *PolyphonicInstrumentBase) ParamTextToValue(paramID uint32, text string, value unsafe.Pointer) bool {
    return p.paramHandler.ParamTextToValue(paramID, text, value)
}

func (p *PolyphonicInstrumentBase) HandleParamValue(paramEvent *event.ParamValueEvent, time uint32) {
    p.paramHandler.HandleParamValue(paramEvent, time)
}

func (p *PolyphonicInstrumentBase) ParamsFlush(inEvents, outEvents unsafe.Pointer) {
    p.eventProcessor.ParamsFlush(inEvents, outEvents)
}

func (p *PolyphonicInstrumentBase) SaveState() map[string]interface{} {
    return p.stateManager.SaveState()
}

func (p *PolyphonicInstrumentBase) LoadState(data map[string]interface{}) {
    p.stateManager.LoadState(data)
}

func (p *PolyphonicInstrumentBase) GetVoiceInfo() extension.VoiceInfo {
    return extension.VoiceInfo{
        VoiceCount:    uint32(p.voiceManager.GetActiveVoiceCount()),
        VoiceCapacity: uint32(p.voiceManager.GetMaxVoices()),
        Flags:         extension.VoiceInfoFlagSupportsOverlappingNotes,
    }
}

// Standard lifecycle methods with sensible defaults
func (p *PolyphonicInstrumentBase) Destroy() {}
func (p *PolyphonicInstrumentBase) Deactivate() { p.IsActivated = false }
func (p *PolyphonicInstrumentBase) StartProcessing() bool { p.IsProcessing = true; return p.IsActivated }
func (p *PolyphonicInstrumentBase) StopProcessing() { p.IsProcessing = false }
func (p *PolyphonicInstrumentBase) Reset() { p.voiceManager.Reset() }
func (p *PolyphonicInstrumentBase) OnMainThread() {}
func (p *PolyphonicInstrumentBase) GetPluginID() string { return p.Info.ID }
func (p *PolyphonicInstrumentBase) GetLatency() uint32 { return 0 }
```

## Simplified Synth Plugin Result

With all abstractions, the synth plugin becomes:

```go
type SynthPlugin struct {
    *plugin.PolyphonicInstrumentBase
    
    // Audio processing (the actual plugin logic)
    oscillator *audio.PolyphonicOscillator
    filter     *audio.SelectableFilter
    
    // Direct parameter access
    volume     *param.AtomicFloat64
    cutoff     *param.AtomicFloat64
    filterType *param.AtomicFloat64
    // ... other params
}

func NewSynthPlugin() *SynthPlugin {
    s := &SynthPlugin{
        PolyphonicInstrumentBase: plugin.NewPolyphonicInstrumentBase(PluginInfo, 16),
    }
    
    // Bind parameters
    s.volume = s.params.BindPercentage(1, "Volume", 70.0)
    s.cutoff = s.params.BindCutoffLog(7, "Filter Cutoff", 0.5)
    s.filterType = s.params.BindChoice(9, "Filter Type", FilterTypeChoices, 0)
    
    // Create audio components
    s.oscillator = audio.NewPolyphonicOscillator(s.voiceManager)
    s.filter = audio.NewSelectableFilter(44100, true)
    
    return s
}

func (s *SynthPlugin) Activate(sampleRate float64, minFrames, maxFrames uint32) bool {
    s.PolyphonicInstrumentBase.Activate(sampleRate, minFrames, maxFrames)
    s.filter.SetSampleRate(sampleRate)
    return true
}

func (s *SynthPlugin) Process(...) int {
    if !s.IsActivated || !s.IsProcessing { return process.ProcessError }
    
    s.eventProcessor.ProcessEvents(events, s)
    
    // Get mapped parameter values
    volume := s.volume.Load()
    cutoff, _ := s.params.GetMappedValue(7)
    filterType := int(s.filterType.Load())
    
    // Generate and process audio
    s.filter.SetType(audio.MapFilterTypeFromInt(filterType))
    s.filter.SetFrequency(cutoff)
    
    output := s.oscillator.Process(framesCount)
    s.filter.ProcessBuffer(output)
    audio.DistributeToChannels(output, audioOut, float32(volume))
    
    return s.voiceManager.GetActiveVoiceCount() > 0 ? process.ProcessContinue : process.ProcessSleep
}

// That's it! ~50 lines vs 675 lines
```

## Expected Results

- **Current**: 675 lines, 23 functions
- **Target**: ~150 lines, 5 functions  
- **Reduction**: 78% fewer lines, 78% fewer functions
- **Functions eliminated**: 18 out of 23 functions moved to framework
- **Time to create new synth**: Hours → Minutes
- **Boilerplate eliminated**: Parameter handling, state management, event processing, lifecycle management