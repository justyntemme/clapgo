package main

// #cgo CFLAGS: -I../../include/clap/include
// #include "../../include/clap/include/clap/clap.h"
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"math"
	"unsafe"

	"github.com/justyntemme/clapgo/pkg/api"
	"github.com/justyntemme/clapgo/pkg/registry"
)

// TODO: This example needs to be updated to fully use the new API interfaces.
// Currently, it's only partially updated to allow compilation.

// Export Go plugin functionality
var (
	gainPlugin *GainPlugin
)

func init() {
	// Register our gain plugin with GUI support
	info := api.PluginInfo{
		ID:          "com.clapgo.gain-gui",
		Name:        "Gain with GUI",
		Vendor:      "ClapGo",
		URL:         "https://github.com/justyntemme/clapgo",
		ManualURL:   "https://github.com/justyntemme/clapgo",
		SupportURL:  "https://github.com/justyntemme/clapgo/issues",
		Version:     "1.0.0",
		Description: "A gain plugin with GUI support using ClapGo",
		Features:    []string{"audio-effect", "stereo", "mono", "gui"},
	}
	
	gainPlugin = NewGainPlugin()
	registry.Register(info, func() api.Plugin { return gainPlugin })
}

//export GainGetPluginCount
func GainGetPluginCount() C.uint32_t {
	return 1
}

// GainPlugin implements a simple gain plugin with GUI support
type GainPlugin struct {
	// Plugin state
	gain         float64
	sampleRate   float64
	isActivated  bool
	isProcessing bool
	paramInfo    api.ParamInfo
	host         unsafe.Pointer
	
	// GUI-related fields
	hasGUI       bool
	guiVisible   bool
	guiCreated   bool
}

// NewGainPlugin creates a new gain plugin
func NewGainPlugin() *GainPlugin {
	plugin := &GainPlugin{
		gain:         1.0, // 0dB
		sampleRate:   44100.0,
		isActivated:  false,
		isProcessing: false,
		hasGUI:       true,
	}
	
	// Set up parameter info
	plugin.paramInfo = api.ParamInfo{
		ID:           1,
		Name:         "Gain",
		Module:       "",
		MinValue:     0.0,  // -inf dB
		MaxValue:     2.0,  // +6 dB
		DefaultValue: 1.0,  // 0 dB
		Flags:        api.ParamIsAutomatable | api.ParamIsBoundedBelow | api.ParamIsBoundedAbove,
	}
	
	return plugin
}

// Init initializes the plugin
func (p *GainPlugin) Init() bool {
	return true
}

// Destroy cleans up plugin resources
func (p *GainPlugin) Destroy() {
	// Nothing to clean up
}

// Activate prepares the plugin for processing
func (p *GainPlugin) Activate(sampleRate float64, minFrames, maxFrames uint32) bool {
	p.sampleRate = sampleRate
	p.isActivated = true
	return true
}

// Deactivate stops the plugin from processing
func (p *GainPlugin) Deactivate() {
	p.isActivated = false
}

// StartProcessing begins audio processing
func (p *GainPlugin) StartProcessing() bool {
	if !p.isActivated {
		return false
	}
	p.isProcessing = true
	return true
}

// StopProcessing ends audio processing
func (p *GainPlugin) StopProcessing() {
	p.isProcessing = false
}

// Reset resets the plugin state
func (p *GainPlugin) Reset() {
	p.gain = 1.0
}

// Process processes audio data
func (p *GainPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events api.EventHandler) int {
	// Check if we're in a valid state for processing
	if !p.isActivated || !p.isProcessing {
		return api.ProcessError
	}
	
	// Process parameter changes from events
	if events != nil {
		eventCount := events.GetInputEventCount()
		
		for i := uint32(0); i < eventCount; i++ {
			event := events.GetInputEvent(i)
			if event == nil {
				continue
			}
			
			// Handle parameter changes
			if event.Type == api.EventTypeParamValue {
				paramEvent, ok := event.Data.(api.ParamEvent)
				if ok && paramEvent.ParamID == 1 { // Gain parameter
					p.gain = paramEvent.Value
				}
			}
		}
	}
	
	// If no audio inputs or outputs, nothing to do
	if len(audioIn) == 0 || len(audioOut) == 0 {
		return api.ProcessContinue
	}
	
	// Get the number of channels (use min of input and output)
	numChannels := len(audioIn)
	if len(audioOut) < numChannels {
		numChannels = len(audioOut)
	}
	
	// Process audio - apply gain to each sample
	for ch := 0; ch < numChannels; ch++ {
		inChannel := audioIn[ch]
		outChannel := audioOut[ch]
		
		// Make sure we have enough buffer space
		if len(inChannel) < int(framesCount) || len(outChannel) < int(framesCount) {
			return api.ProcessError
		}
		
		// Apply gain to each sample
		for i := uint32(0); i < framesCount; i++ {
			outChannel[i] = inChannel[i] * float32(p.gain)
		}
	}
	
	// Check if the output is silent
	isSilent := p.gain < 0.0001 // -80dB
	
	if isSilent {
		return api.ProcessSleep
	}
	
	return api.ProcessContinue
}

// GetExtension gets a plugin extension
func (p *GainPlugin) GetExtension(id string) unsafe.Pointer {
	// Check for parameter extension
	if id == api.ExtParams {
		return nil // Not implemented in this simplified version
	}
	
	// Check for state extension
	if id == api.ExtState {
		return nil // Not implemented in this simplified version
	}
	
	// GUI extensions are handled via the gui_bridge.cpp
	return nil
}

// HasGUI returns true if the plugin has a GUI
func (p *GainPlugin) HasGUI() bool {
	return p.hasGUI
}

// GetPreferredGUIAPI returns the preferred GUI API
func (p *GainPlugin) GetPreferredGUIAPI() (api string, isFloating bool) {
	// Default to X11 on Linux, adjust based on OS
	return api.WindowAPIX11, false
}

// OnGUICreated is called when the GUI is created
func (p *GainPlugin) OnGUICreated() {
	p.guiCreated = true
	fmt.Println("Go: GUI created")
}

// OnGUIDestroyed is called when the GUI is destroyed
func (p *GainPlugin) OnGUIDestroyed() {
	p.guiCreated = false
	p.guiVisible = false
	fmt.Println("Go: GUI destroyed")
}

// OnGUIShown is called when the GUI is shown
func (p *GainPlugin) OnGUIShown() {
	p.guiVisible = true
	fmt.Println("Go: GUI shown")
}

// OnGUIHidden is called when the GUI is hidden
func (p *GainPlugin) OnGUIHidden() {
	p.guiVisible = false
	fmt.Println("Go: GUI hidden")
}

// GetGUISize returns the default GUI size
func (p *GainPlugin) GetGUISize() (width, height uint32) {
	return 400, 300
}

// OnMainThread is called on the main thread
func (p *GainPlugin) OnMainThread() {
	// Nothing to do
}

// GetPluginInfo returns information about the plugin
func (p *GainPlugin) GetPluginInfo() api.PluginInfo {
	return api.PluginInfo{
		ID:          "com.clapgo.gain-gui",
		Name:        "Gain with GUI",
		Vendor:      "ClapGo",
		URL:         "https://github.com/justyntemme/clapgo",
		ManualURL:   "https://github.com/justyntemme/clapgo",
		SupportURL:  "https://github.com/justyntemme/clapgo/issues",
		Version:     "1.0.0",
		Description: "A gain plugin with GUI support using ClapGo",
		Features:    []string{"audio-effect", "stereo", "mono", "gui"},
	}
}

// SaveState returns custom state data for the plugin
func (p *GainPlugin) SaveState() map[string]interface{} {
	// Save any additional state beyond parameters
	return map[string]interface{}{
		"plugin_version": "1.0.0",
		"last_gain":      p.gain,
		"gui_visible":    p.guiVisible,
		"gui_created":    p.guiCreated,
	}
}

// LoadState loads custom state data for the plugin
func (p *GainPlugin) LoadState(data map[string]interface{}) {
	// Load any additional state beyond parameters
	if lastGain, ok := data["last_gain"].(float64); ok {
		p.gain = lastGain
	}
	
	if guiVisible, ok := data["gui_visible"].(bool); ok {
		p.guiVisible = guiVisible
	}
	
	if guiCreated, ok := data["gui_created"].(bool); ok {
		p.guiCreated = guiCreated
	}
}

// GetPluginID returns the plugin ID
func (p *GainPlugin) GetPluginID() string {
	return "com.clapgo.gain-gui"
}

// Convert linear gain to dB
func linearToDb(linear float64) float64 {
	if linear <= 0.0 {
		return -math.MaxFloat64
	}
	return 20.0 * math.Log10(linear)
}

// Convert dB to linear gain
func dbToLinear(db float64) float64 {
	return math.Pow(10.0, db/20.0)
}

func main() {
	// This is not called when used as a plugin,
	// but can be useful for testing
}