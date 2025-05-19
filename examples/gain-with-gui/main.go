package main

// #cgo CFLAGS: -I../../include/clap/include
// #include "../../include/clap/include/clap/clap.h"
// #include <stdlib.h>
import "C"
import (
	"math"
	"unsafe"

	"github.com/justyntemme/clapgo/src/goclap"
)

// Export Go plugin functionality
var (
	gainPlugin *GainPlugin
)

func init() {
	// Register our gain plugin with GUI support
	info := goclap.PluginInfo{
		ID:          "com.clapgo.gain.gui",
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
	goclap.RegisterPlugin(info, gainPlugin)
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
	paramManager *goclap.ParamManager
	host         *goclap.Host
	
	// GUI-related fields
	hasGUI       bool
}

// NewGainPlugin creates a new gain plugin
func NewGainPlugin() *GainPlugin {
	plugin := &GainPlugin{
		gain:         1.0, // 0dB
		sampleRate:   44100.0,
		isActivated:  false,
		isProcessing: false,
		paramManager: goclap.NewParamManager(),
		hasGUI:       true,
	}
	
	// Register parameters
	plugin.paramManager.RegisterParam(goclap.ParamInfo{
		ID:           1,
		Name:         "Gain",
		Module:       "",
		MinValue:     0.0,  // -inf dB
		MaxValue:     2.0,  // +6 dB
		DefaultValue: 1.0,  // 0 dB
		Flags:        goclap.ParamIsAutomatable | goclap.ParamIsBoundedBelow | goclap.ParamIsBoundedAbove,
	})
	
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
func (p *GainPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events *goclap.ProcessEvents) int {
	// Check if we're in a valid state for processing
	if !p.isActivated || !p.isProcessing {
		return 0 // CLAP_PROCESS_ERROR
	}
	
	// Process parameter changes from events
	if events != nil && events.InEvents != nil {
		inputEvents := &goclap.InputEvents{Ptr: events.InEvents}
		eventCount := inputEvents.GetEventCount()
		
		for i := uint32(0); i < eventCount; i++ {
			event := inputEvents.GetEvent(i)
			if event == nil {
				continue
			}
			
			// Handle parameter changes
			if event.Type == goclap.EventTypeParamValue {
				paramEvent, ok := event.Data.(goclap.ParamEvent)
				if ok && paramEvent.ParamID == 1 { // Gain parameter
					p.gain = paramEvent.Value
				}
			}
		}
	}
	
	// If no audio inputs or outputs, nothing to do
	if len(audioIn) == 0 || len(audioOut) == 0 {
		return 1 // CLAP_PROCESS_CONTINUE
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
			return 0 // CLAP_PROCESS_ERROR
		}
		
		// Apply gain to each sample
		for i := uint32(0); i < framesCount; i++ {
			outChannel[i] = inChannel[i] * float32(p.gain)
		}
	}
	
	// Check if the output is silent
	isSilent := p.gain < 0.0001 // -80dB
	
	if isSilent {
		return 4 // CLAP_PROCESS_SLEEP
	}
	
	return 1 // CLAP_PROCESS_CONTINUE
}

// GetExtension gets a plugin extension
func (p *GainPlugin) GetExtension(id string) unsafe.Pointer {
	// GUI extensions are handled in the C++ part
	return nil
}

// OnMainThread is called on the main thread
func (p *GainPlugin) OnMainThread() {
	// Nothing to do
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