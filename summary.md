# ClapGo: CLAP Plugin Framework for Go

ClapGo is a Go framework for creating CLAP (CLever Audio Plugin) audio plugins using the Go programming language. This document explains the architecture, code structure, and functionality of the ClapGo framework.

## 1. Overview

ClapGo enables developers to create audio plugins in Go that conform to the CLAP standard. It provides:

- A bridge between Go and C for CLAP API compatibility
- Plugin registration and discovery
- Audio processing infrastructure
- Parameter management
- Extensions for audio ports, note ports, and other CLAP features
- Example plugins demonstrating the framework's use

## 2. Code Structure

The project is organized into several key directories:

### 2.1 API Package (`pkg/api`)

Contains the core interfaces and types that plugins must implement:

- `Plugin`: The main interface for all CLAP plugins
- `EventHandler`: Interface for handling audio events (MIDI, parameter changes, etc.)
- Extension interfaces: `AudioPortsProvider`, `ParamsProvider`, etc.

### 2.2 Bridge Implementation (`internal/bridge` and `cmd/goclap`)

Responsible for the C/Go boundary:

- CGO exports for CLAP function callbacks
- Type conversion between C and Go structures
- Handle management for plugin instances
- Plugin lifecycle management

### 2.3 Registry (`internal/registry`)

Maintains a registry of available plugins:

- Plugin registration
- Plugin lookup by ID
- Plugin creation

### 2.4 Core Implementation (`src/goclap`)

Provides common functionality for plugins:

- Audio processing utilities
- Parameter management
- Event handling
- Extension implementations (audio ports, note ports, etc.)
- Plugin registration

### 2.5 Example Plugins (`examples/`)

Contains example plugin implementations:

- `gain`: A simple gain plugin
- `synth`: A basic synthesizer
- `gain-with-gui`: A gain plugin with a GUI

## 3. How It Works

### 3.1 Plugin Registration

1. Plugins are registered during initialization:
   ```go
   func init() {
       info := goclap.PluginInfo{
           ID:          "com.clapgo.gain",
           Name:        "Simple Gain",
           // ...
       }
       
       gainPlugin = NewGainPlugin()
       goclap.RegisterPlugin(info, gainPlugin)
   }
   ```

2. The registry stores information about available plugins and provides factory functions.

### 3.2 C/Go Bridge

1. The Go library is loaded by the host via the C bridge.
2. C function exports provide the interface for the host to interact with Go plugins:
   - `GetPluginCount`: Returns the number of available plugins
   - `GetPluginInfo`: Returns information about a plugin by index
   - `CreatePlugin`: Creates a plugin instance
   - `GoInit`, `GoProcess`, etc.: Plugin lifecycle functions

3. C types are converted to Go types and vice versa.

### 3.3 Plugin Lifecycle

1. **Creation**: The host calls `CreatePlugin` to create a new plugin instance.
2. **Initialization**: The plugin is initialized with `GoInit`.
3. **Activation**: The plugin is activated with sample rate and buffer size info.
4. **Processing**: Audio is processed in blocks, with events handled during processing.
5. **Deactivation and Cleanup**: Resources are released when the plugin is no longer needed.

### 3.4 Audio Processing

1. Audio buffers are passed from C to Go as slices.
2. Plugins process audio and apply effects or generate sound.
3. Events (parameter changes, MIDI, etc.) are processed during audio processing.

### 3.5 Parameters

1. Plugins define parameters using the parameter registry:
   ```go
   plugin.paramManager.RegisterParam(goclap.ParamInfo{
       ID:           1,
       Name:         "Gain",
       // ...
   })
   ```

2. Parameter changes are received as events during audio processing.

### 3.6 Extensions

Plugins can implement various CLAP extensions:

- Audio ports: Define audio inputs and outputs
- Note ports: Define MIDI/note inputs and outputs
- State: Save and load plugin state
- GUI: Provide a graphical user interface
- And many others

## 4. Key Concepts

### 4.1 Plugin Interface

The core plugin interface defines the required methods for all plugins:

```go
type Plugin interface {
    Init() bool
    Destroy()
    Activate(sampleRate float64, minFrames, maxFrames uint32) bool
    Deactivate()
    StartProcessing() bool
    StopProcessing()
    Reset()
    Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events EventHandler) int
    GetExtension(id string) unsafe.Pointer
    OnMainThread()
    GetPluginID() string
}
```

### 4.2 Event Handling

Events are used for parameter changes, MIDI data, and transport information:

```go
type Event struct {
    Type int
    Time uint32
    Data interface{}
}
```

Common event types include:
- Parameter changes
- Note on/off events
- MIDI data
- Transport information

### 4.3 Audio Processing

Audio is processed in blocks, with input and output buffers provided as slices:

```go
func (p *GainPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events *goclap.ProcessEvents) int {
    // Process audio...
    for ch := 0; ch < numChannels; ch++ {
        inChannel := audioIn[ch]
        outChannel := audioOut[ch]
        
        for i := uint32(0); i < framesCount; i++ {
            outChannel[i] = inChannel[i] * float32(p.gain)
        }
    }
    
    return 1 // CLAP_PROCESS_CONTINUE
}
```

### 4.4 Extensions

Extensions provide additional capabilities to plugins. For example, the audio ports extension:

```go
func (p *MyPlugin) GetExtension(id string) unsafe.Pointer {
    if id == goclap.ExtAudioPorts {
        return unsafe.Pointer(p.audioPorts)
    }
    return nil
}
```

## 5. C Code Overview

The C code in the project serves primarily as a bridge between the CLAP host and the Go implementation. The key components are:

### 5.1 Plugin Entry and Factory (`plugin.c`)

Provides the CLAP entry point and factory implementation:
- `clap_entry` structure: The main entry point for the CLAP host
- `clapgo_entry_init`/`clapgo_entry_deinit`: Initialize and clean up the plugin
- `clapgo_entry_get_factory`: Returns the plugin factory

### 5.2 Bridge Implementation (`bridge.c`)

Handles communication between the CLAP host and Go:
- Loads the Go shared library
- Finds and calls Go functions via function pointers
- Manages plugin descriptors and instances
- Converts between C and Go data structures

### 5.3 Audio Processing

The C side handles audio buffer conversion and event processing:
- Converts CLAP audio buffers to a format Go can work with
- Passes event data between the host and Go
- Manages audio port information

### 5.4 Memory Management

Careful memory management is required for the C/Go boundary:
- C allocates memory for descriptors, which Go must free
- Go data passed to C must be properly handled to prevent garbage collection issues
- References between C and Go must be properly maintained

## 6. Examples

The project includes several example plugins:

### 6.1 Gain Plugin

A simple gain plugin that applies a volume adjustment to the audio:

```go
// Process applies gain to the audio
func (p *GainPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events *goclap.ProcessEvents) int {
    // Apply gain to each sample
    for ch := 0; ch < numChannels; ch++ {
        inChannel := audioIn[ch]
        outChannel := audioOut[ch]
        
        for i := uint32(0); i < framesCount; i++ {
            outChannel[i] = inChannel[i] * float32(p.gain)
        }
    }
    
    return 1 // CLAP_PROCESS_CONTINUE
}
```

### 6.2 Synth Plugin

A basic synthesizer that generates sound:

```go
// Process generates audio based on note events
func (p *SynthPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events *goclap.ProcessEvents) int {
    // Process events (notes, parameters, etc.)
    p.processEvents(events, framesCount)
    
    // Generate audio for each active voice
    for _, voice := range p.voices {
        if voice != nil && voice.IsActive {
            // Calculate frequency for this note
            freq := noteToFrequency(int(voice.Key))
            
            // Generate samples
            for j := uint32(0); j < framesCount; j++ {
                // Get envelope value
                env := p.getEnvelopeValue(voice, j, framesCount)
                
                // Generate sample
                sample := p.generateSample(voice.Phase, freq) * float64(env) * voice.Velocity
                
                // Apply master volume and add to output
                // ...
            }
        }
    }
    
    return 1 // CLAP_PROCESS_CONTINUE
}
```

## 7. Potential Improvements

As discussed in the cleanup.md document, there are several areas where the codebase could be improved:

1. **Consolidate Duplicated Code**: Remove duplicate implementations of registry, bridge, and event handling.
2. **Clarify Package Responsibilities**: Define clear boundaries between packages.
3. **Standardize Interfaces**: Use consistent interfaces throughout the codebase.
4. **Improve Error Handling**: Add proper error handling and logging.
5. **Enhance Documentation**: Add more comprehensive documentation and examples.
6. **Optimize Performance**: Review performance-critical code paths.
7. **Add Testing**: Increase test coverage.

## 8. Conclusion

ClapGo provides a solid foundation for developing CLAP plugins in Go. Its architecture allows for efficient audio processing while leveraging Go's strengths in memory safety and concurrency. With some architectural improvements and code cleanup, it could become an even more powerful tool for audio plugin development in Go.