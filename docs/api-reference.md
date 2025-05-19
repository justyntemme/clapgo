# ClapGo API Reference

This document provides a reference for the key interfaces and functions in the ClapGo framework.

## Core Interfaces

### AudioProcessor

The main interface that must be implemented by all plugins:

```go
type AudioProcessor interface {
    // Initialize the plugin
    Init() bool
    
    // Clean up plugin resources
    Destroy()
    
    // Activate the plugin with given sample rate and buffer sizes
    Activate(sampleRate float64, minFrames, maxFrames uint32) bool
    
    // Deactivate the plugin
    Deactivate()
    
    // Start audio processing
    StartProcessing() bool
    
    // Stop audio processing
    StopProcessing()
    
    // Reset the plugin state
    Reset()
    
    // Process audio data
    // Returns a status code: 0-error, 1-continue, 2-continue_if_not_quiet, 3-tail, 4-sleep
    Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events *ProcessEvents) int
    
    // Get an extension
    GetExtension(id string) unsafe.Pointer
    
    // Called on the main thread
    OnMainThread()
}
```

### Parameter Management

```go
// ParamInfo holds parameter metadata
type ParamInfo struct {
    ID          uint32
    Name        string
    Module      string
    MinValue    float64
    MaxValue    float64
    DefaultValue float64
    Flags       uint32
}

// ParamManager provides an interface to manage plugin parameters
type ParamManager struct {
    // ...implementation details
}

// Main parameter management functions
func NewParamManager() *ParamManager
func (pm *ParamManager) RegisterParam(param ParamInfo)
func (pm *ParamManager) GetParamCount() uint32
func (pm *ParamManager) GetParamInfo(paramID uint32) *ParamInfo
func (pm *ParamManager) GetParamValue(paramID uint32) float64
func (pm *ParamManager) SetParamValue(paramID uint32, value float64) bool
func (pm *ParamManager) Flush()
```

### Event Handling

```go
// Event types
const (
    EventTypeNoteOn       = 0
    EventTypeNoteOff      = 1
    EventTypeNoteChoke    = 2
    EventTypeNoteExpression = 3
    EventTypeParamValue   = 4
    EventTypeParamMod     = 5
    EventTypeTransport    = 6
    EventTypeMIDI         = 7
    EventTypeMIDI2        = 8
    EventTypeMIDISysex    = 9
)

// Event represents a CLAP event
type Event struct {
    Time   uint32
    Type   uint16
    Space  uint16
    Data   interface{} // NoteEvent, ParamEvent, etc.
}

// NoteEvent represents a CLAP note event
type NoteEvent struct {
    NoteID      uint32
    Port        uint16
    Channel     uint16
    Key         uint8
    Velocity    float64
}

// ParamEvent represents a CLAP parameter event
type ParamEvent struct {
    ParamID     uint32
    Cookie      unsafe.Pointer
    Note        uint32
    Port        uint16
    Channel     uint16
    Key         uint8
    Value       float64
}

// InputEvents and OutputEvents handle event queues
type InputEvents struct {
    Ptr unsafe.Pointer
}

type OutputEvents struct {
    Ptr unsafe.Pointer
}

// Event handling functions
func (e *InputEvents) GetEventCount() uint32
func (e *InputEvents) GetEvent(index uint32) *Event
func (e *OutputEvents) PushEvent(event *Event) bool
```

### Host Communication

```go
// Host represents a CLAP host
type Host struct {
    ptr unsafe.Pointer
}

// Host information and communication functions
func NewHost(hostPtr unsafe.Pointer) *Host
func (h *Host) GetName() string
func (h *Host) GetVendor() string
func (h *Host) GetURL() string
func (h *Host) GetVersion() string
func (h *Host) RequestRestart()
func (h *Host) RequestProcess()
func (h *Host) RequestCallback()
func (h *Host) GetExtension(id string) unsafe.Pointer
```

## Plugin Registration

```go
// PluginInfo holds metadata about a CLAP plugin
type PluginInfo struct {
    ID          string
    Name        string
    Vendor      string
    URL         string
    ManualURL   string
    SupportURL  string
    Version     string
    Description string
    Features    []string
}

// Register a plugin with the framework
func RegisterPlugin(info PluginInfo, processor AudioProcessor)
```

## Constants

### Parameter Flags

```go
const (
    // Parameter properties
    ParamIsSteppable          = 1 << 0
    ParamIsPeriodic           = 1 << 1
    ParamIsHidden             = 1 << 2
    ParamIsReadOnly           = 1 << 3
    ParamIsBypass             = 1 << 4
    
    // Automation capabilities
    ParamIsAutomatable        = 1 << 5
    ParamIsAutomatablePerNoteID = 1 << 6
    ParamIsAutomatablePerKey   = 1 << 7
    ParamIsAutomatablePerChannel = 1 << 8
    ParamIsAutomatablePerPort  = 1 << 9
    
    // Modulation capabilities
    ParamIsModulatable        = 1 << 10
    ParamIsModulatablePerNoteID = 1 << 11
    ParamIsModulatablePerKey   = 1 << 12
    ParamIsModulatablePerChannel = 1 << 13
    ParamIsModulatablePerPort  = 1 << 14
    
    // Processing behavior
    ParamRequiresProcess      = 1 << 15
    
    // Parameter range info
    ParamIsBoundedBelow       = 1 << 16
    ParamIsBoundedAbove       = 1 << 17
)
```

### Note Expressions

```go
const (
    NoteExpressionVolume      = 0
    NoteExpressionPan         = 1
    NoteExpressionTuning      = 2
    NoteExpressionVibrato     = 3
    NoteExpressionExpression  = 4
    NoteExpressionBrightness  = 5
    NoteExpressionPressure    = 6
)
```

### Process Status Codes

```go
const (
    ProcessError            = 0 // Error during processing
    ProcessContinue         = 1 // Continue processing
    ProcessContinueIfNotQuiet = 2 // Continue if output is not silent
    ProcessTail             = 3 // Process tail (for reverb etc.)
    ProcessSleep            = 4 // No more processing needed until input changes
)
```

## Common Patterns

### Plugin Registration

In the `init()` function of your plugin:

```go
func init() {
    info := goclap.PluginInfo{
        ID:          "com.example.myplugin",
        Name:        "My Plugin",
        Vendor:      "Example Developer",
        URL:         "https://example.com",
        Version:     "1.0.0",
        Description: "Example plugin description",
        Features:    []string{"audio-effect", "stereo"},
    }
    
    myPlugin := NewMyPlugin()
    goclap.RegisterPlugin(info, myPlugin)
}
```

### Audio Processing

In your plugin's `Process` method:

```go
func (p *MyPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events *goclap.ProcessEvents) int {
    // Process parameter changes from events
    if events != nil && events.InEvents != nil {
        inputEvents := &goclap.InputEvents{Ptr: events.InEvents}
        eventCount := inputEvents.GetEventCount()
        
        for i := uint32(0); i < eventCount; i++ {
            event := inputEvents.GetEvent(i)
            if event == nil {
                continue
            }
            
            // Handle event based on type...
        }
    }
    
    // Process audio
    if len(audioIn) > 0 && len(audioOut) > 0 {
        for ch := 0; ch < min(len(audioIn), len(audioOut)); ch++ {
            for i := uint32(0); i < framesCount; i++ {
                // Process each sample
                audioOut[ch][i] = processSample(audioIn[ch][i])
            }
        }
    }
    
    return goclap.ProcessContinue
}
```