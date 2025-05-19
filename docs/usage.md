# ClapGo Usage Guide

This guide explains how to use ClapGo to build CLAP audio plugins using Go.

## Project Structure

ClapGo is designed around the following architecture:

1. **Core Go library** - Provides Go bindings to CLAP API
2. **C wrapper** - Implements the CLAP plugin interface and calls into the Go code
3. **Build system** - Compiles and packages plugins for different platforms

## Creating a Plugin

### 1. Set up the plugin package

Create a new directory for your plugin in the `examples` folder. For example:

```
mkdir -p examples/myplugin
```

### 2. Implement the AudioProcessor interface

Create a `main.go` file in your plugin directory and implement the `goclap.AudioProcessor` interface:

```go
package main

// #cgo CFLAGS: -I../../include/clap/include
// #include "../../include/clap/include/clap/clap.h"
// #include <stdlib.h>
import "C"
import (
    "github.com/yourusername/clapgo/src/goclap"
    "unsafe"
)

// MyPlugin implements the AudioProcessor interface
type MyPlugin struct {
    // Plugin state here
}

// Implement all required methods of the AudioProcessor interface

func (p *MyPlugin) Init() bool {
    // Initialize your plugin
    return true
}

func (p *MyPlugin) Destroy() {
    // Clean up resources
}

// ... implement other methods ...

func (p *MyPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events *goclap.ProcessEvents) int {
    // Process audio
    
    // Return appropriate process status
    return 1 // CLAP_PROCESS_CONTINUE
}

// Register the plugin in init()
func init() {
    info := goclap.PluginInfo{
        ID:          "com.yourdomain.myplugin",
        Name:        "My Plugin",
        Vendor:      "Your Name",
        // ... other metadata ...
    }
    
    myPlugin := &MyPlugin{}
    goclap.RegisterPlugin(info, myPlugin)
}

func main() {
    // This is not called when used as a plugin
}
```

### 3. Register Parameters

If your plugin has parameters, register them in your plugin's constructor:

```go
func NewMyPlugin() *MyPlugin {
    plugin := &MyPlugin{}
    
    // Create parameter manager
    plugin.paramManager = goclap.NewParamManager()
    
    // Register parameters
    plugin.paramManager.RegisterParam(goclap.ParamInfo{
        ID:           1,
        Name:         "My Parameter",
        Module:       "",
        MinValue:     0.0,
        MaxValue:     1.0,
        DefaultValue: 0.5,
        Flags:        goclap.ParamIsAutomatable,
    })
    
    return plugin
}
```

### 4. Build the plugin

Add your plugin to the Makefile:

```makefile
examples: build
    ./dist/clapgo-build --name gain --output dist/
    ./dist/clapgo-build --name myplugin --output dist/
```

Then build your plugin:

```
make examples
```

This will create your plugin in the `dist` directory.

## API Reference

### Plugin Lifecycle

1. **Registration**: Register your plugin with `goclap.RegisterPlugin(info, processor)`
2. **Initialization**: Host calls `processor.Init()`
3. **Activation**: Host calls `processor.Activate(sampleRate, minFrames, maxFrames)`
4. **Processing**: Host calls `processor.StartProcessing()` followed by multiple calls to `processor.Process()`
5. **Deactivation**: Host calls `processor.StopProcessing()` followed by `processor.Deactivate()`
6. **Cleanup**: Host calls `processor.Destroy()`

### Audio Processing

The `Process` method is where audio processing happens:

```go
func (p *MyPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events *goclap.ProcessEvents) int {
    // Process parameter changes
    if events != nil && events.inEvents != nil {
        inputEvents := &goclap.InputEvents{Ptr: events.inEvents}
        eventCount := inputEvents.GetEventCount()
        
        for i := uint32(0); i < eventCount; i++ {
            event := inputEvents.GetEvent(i)
            // Handle events...
        }
    }
    
    // Process audio
    for ch := 0; ch < len(audioIn); ch++ {
        inChannel := audioIn[ch]
        outChannel := audioOut[ch]
        
        for i := uint32(0); i < framesCount; i++ {
            // Process each sample
            outChannel[i] = processSample(inChannel[i])
        }
    }
    
    return 1 // CLAP_PROCESS_CONTINUE
}
```

### Return Status

The `Process` method should return one of these status codes:

- `0` (`CLAP_PROCESS_ERROR`): Processing failed
- `1` (`CLAP_PROCESS_CONTINUE`): Continue processing
- `2` (`CLAP_PROCESS_CONTINUE_IF_NOT_QUIET`): Continue if output is not silent
- `3` (`CLAP_PROCESS_TAIL`): In tail mode
- `4` (`CLAP_PROCESS_SLEEP`): Processing can be suspended

## Tips and Best Practices

1. **Memory Management**: Be careful with memory allocations during audio processing to avoid garbage collection pauses
2. **Thread Safety**: Ensure thread safety in your plugin implementation
3. **Parameter Smoothing**: Implement parameter smoothing to avoid zipper noise
4. **Error Handling**: Handle errors gracefully in your plugin
5. **Testing**: Test your plugin thoroughly before deployment

## Debugging

To debug your plugin, you can add logging statements to your code. However, be careful not to log too much during audio processing as it can affect performance.

```go
import "log"

func (p *MyPlugin) Process(...) int {
    // Log only important events
    if someImportantEvent {
        log.Println("Important event occurred")
    }
    
    // ...process audio...
    
    return 1
}
```

You can also build your plugin with debug information:

```
./dist/clapgo-build --name myplugin --output dist/ --verbose
```