package main

// #cgo CFLAGS: -I../../include/clap/include
// #include "../../include/clap/include/clap/clap.h"
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"github.com/justyntemme/clapgo/pkg/api"
	"math"
	"runtime/cgo"
	"unsafe"
)

// Global plugin instance
var gainPlugin *GainPlugin

func init() {
	fmt.Println("Initializing gain plugin")
	gainPlugin = NewGainPlugin()
	fmt.Printf("Gain plugin initialized: %s (%s)\n", gainPlugin.GetPluginInfo().Name, gainPlugin.GetPluginInfo().ID)
}

// Standardized export functions for manifest system

//export ClapGo_CreatePlugin
func ClapGo_CreatePlugin(host unsafe.Pointer, pluginID *C.char) unsafe.Pointer {
	id := C.GoString(pluginID)
	fmt.Printf("Gain plugin - ClapGo_CreatePlugin with ID: %s\n", id)
	
	if id == PluginID {
		// Create a CGO handle to safely pass the Go object to C
		handle := cgo.NewHandle(gainPlugin)
		fmt.Printf("Created plugin instance: %s\n", id)
		return unsafe.Pointer(handle)
	}
	
	fmt.Printf("Error: Unknown plugin ID: %s\n", id)
	return nil
}

//export ClapGo_GetVersion
func ClapGo_GetVersion(major, minor, patch *C.uint32_t) C.bool {
	if major != nil {
		*major = C.uint32_t(1)
	}
	if minor != nil {
		*minor = C.uint32_t(0)
	}
	if patch != nil {
		*patch = C.uint32_t(0)
	}
	fmt.Printf("Gain plugin - ClapGo_GetVersion returning 1.0.0\n")
	return C.bool(true)
}

//export ClapGo_GetPluginID
func ClapGo_GetPluginID(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	fmt.Printf("Gain plugin - ClapGo_GetPluginID for %s\n", id)
	
	if id == PluginID {
		return C.CString(PluginID)
	}
	
	return C.CString("unknown")
}

//export ClapGo_GetPluginName
func ClapGo_GetPluginName(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	fmt.Printf("Gain plugin - ClapGo_GetPluginName for %s\n", id)
	
	if id == PluginID {
		return C.CString(PluginName)
	}
	
	return C.CString("Unknown Plugin")
}

//export ClapGo_GetPluginVendor
func ClapGo_GetPluginVendor(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	fmt.Printf("Gain plugin - ClapGo_GetPluginVendor for %s\n", id)
	
	if id == PluginID {
		return C.CString(PluginVendor)
	}
	
	return C.CString("Unknown Vendor")
}

//export ClapGo_GetPluginVersion
func ClapGo_GetPluginVersion(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	fmt.Printf("Gain plugin - ClapGo_GetPluginVersion for %s\n", id)
	
	if id == PluginID {
		return C.CString(PluginVersion)
	}
	
	return C.CString("0.0.0")
}

//export ClapGo_GetPluginDescription
func ClapGo_GetPluginDescription(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	fmt.Printf("Gain plugin - ClapGo_GetPluginDescription for %s\n", id)
	
	if id == PluginID {
		return C.CString(PluginDescription)
	}
	
	return C.CString("No description available")
}

// Plugin lifecycle functions

//export ClapGo_PluginInit
func ClapGo_PluginInit(plugin unsafe.Pointer) C.bool {
	fmt.Printf("Gain plugin - ClapGo_PluginInit\n")
	if plugin == nil {
		return C.bool(false)
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*GainPlugin)
	return C.bool(p.Init())
}

//export ClapGo_PluginDestroy
func ClapGo_PluginDestroy(plugin unsafe.Pointer) {
	fmt.Printf("Gain plugin - ClapGo_PluginDestroy\n")
	if plugin == nil {
		return
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*GainPlugin)
	p.Destroy()
	handle.Delete()
}

//export ClapGo_PluginActivate
func ClapGo_PluginActivate(plugin unsafe.Pointer, sampleRate C.double, minFrames, maxFrames C.uint32_t) C.bool {
	fmt.Printf("Gain plugin - ClapGo_PluginActivate with sample rate %f\n", sampleRate)
	if plugin == nil {
		return C.bool(false)
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*GainPlugin)
	return C.bool(p.Activate(float64(sampleRate), uint32(minFrames), uint32(maxFrames)))
}

//export ClapGo_PluginDeactivate
func ClapGo_PluginDeactivate(plugin unsafe.Pointer) {
	fmt.Printf("Gain plugin - ClapGo_PluginDeactivate\n")
	if plugin == nil {
		return
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*GainPlugin)
	p.Deactivate()
}

//export ClapGo_PluginStartProcessing
func ClapGo_PluginStartProcessing(plugin unsafe.Pointer) C.bool {
	fmt.Printf("Gain plugin - ClapGo_PluginStartProcessing\n")
	if plugin == nil {
		return C.bool(false)
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*GainPlugin)
	return C.bool(p.StartProcessing())
}

//export ClapGo_PluginStopProcessing
func ClapGo_PluginStopProcessing(plugin unsafe.Pointer) {
	fmt.Printf("Gain plugin - ClapGo_PluginStopProcessing\n")
	if plugin == nil {
		return
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*GainPlugin)
	p.StopProcessing()
}

//export ClapGo_PluginReset
func ClapGo_PluginReset(plugin unsafe.Pointer) {
	fmt.Printf("Gain plugin - ClapGo_PluginReset\n")
	if plugin == nil {
		return
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*GainPlugin)
	p.Reset()
}

//export ClapGo_PluginProcess
func ClapGo_PluginProcess(plugin unsafe.Pointer, process unsafe.Pointer) C.int32_t {
	fmt.Printf("Gain plugin - ClapGo_PluginProcess\n")
	if plugin == nil || process == nil {
		return C.int32_t(api.ProcessError)
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*GainPlugin)
	// For now, just return continue - proper processing would go here
	_ = p // Prevent unused variable error
	return C.int32_t(api.ProcessContinue)
}

//export ClapGo_PluginGetExtension
func ClapGo_PluginGetExtension(plugin unsafe.Pointer, id *C.char) unsafe.Pointer {
	extID := C.GoString(id)
	fmt.Printf("Gain plugin - ClapGo_PluginGetExtension for %s\n", extID)
	if plugin == nil {
		return nil
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*GainPlugin)
	return p.GetExtension(extID)
}

//export ClapGo_PluginOnMainThread
func ClapGo_PluginOnMainThread(plugin unsafe.Pointer) {
	fmt.Printf("Gain plugin - ClapGo_PluginOnMainThread\n")
	if plugin == nil {
		return
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*GainPlugin)
	p.OnMainThread()
}

// GainPlugin implements a simple gain plugin
type GainPlugin struct {
	// Plugin state
	gain         float64
	sampleRate   float64
	isActivated  bool
	isProcessing bool
	paramInfo    api.ParamInfo
	host         unsafe.Pointer
}

// NewGainPlugin creates a new gain plugin
func NewGainPlugin() *GainPlugin {
	plugin := &GainPlugin{
		gain:         1.0, // 0dB
		sampleRate:   44100.0,
		isActivated:  false,
		isProcessing: false,
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
	
	// No other extensions supported
	return nil
}

// OnMainThread is called on the main thread
func (p *GainPlugin) OnMainThread() {
	// Nothing to do
}

// GetPluginInfo returns information about the plugin
func (p *GainPlugin) GetPluginInfo() api.PluginInfo {
	return api.PluginInfo{
		ID:          PluginID,
		Name:        PluginName,
		Vendor:      PluginVendor,
		URL:         "https://github.com/justyntemme/clapgo",
		ManualURL:   "https://github.com/justyntemme/clapgo",
		SupportURL:  "https://github.com/justyntemme/clapgo/issues",
		Version:     PluginVersion,
		Description: PluginDescription,
		Features:    []string{"audio-effect", "stereo", "mono"},
	}
}

// SaveState returns custom state data for the plugin
func (p *GainPlugin) SaveState() map[string]interface{} {
	// Save any additional state beyond parameters
	return map[string]interface{}{
		"plugin_version": "1.0.0",
		"last_gain":      p.gain,
		// Add other custom state values here
	}
}

// LoadState loads custom state data for the plugin
func (p *GainPlugin) LoadState(data map[string]interface{}) {
	// Load any additional state beyond parameters
	if lastGain, ok := data["last_gain"].(float64); ok {
		p.gain = lastGain
	}
	
	// You could load other custom state values here
}

// GetPluginID returns the plugin ID
func (p *GainPlugin) GetPluginID() string {
	return PluginID
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