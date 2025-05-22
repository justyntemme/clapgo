package main

// #cgo CFLAGS: -I../../include/clap/include
// #include "../../include/clap/include/clap/clap.h"
// #include <stdlib.h>
//
// // Helper functions for CLAP event handling
// static inline uint32_t clap_input_events_size_helper(const clap_input_events_t* events) {
//     if (events && events->size) {
//         return events->size(events);
//     }
//     return 0;
// }
//
// static inline const clap_event_header_t* clap_input_events_get_helper(const clap_input_events_t* events, uint32_t index) {
//     if (events && events->get) {
//         return events->get(events, index);
//     }
//     return NULL;
// }
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
	gainPlugin = NewGainPlugin()
}

// Standardized export functions for manifest system

//export ClapGo_CreatePlugin
func ClapGo_CreatePlugin(host unsafe.Pointer, pluginID *C.char) unsafe.Pointer {
	id := C.GoString(pluginID)
	
	if id == PluginID {
		// Create a CGO handle to safely pass the Go object to C
		handle := cgo.NewHandle(gainPlugin)
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
	return C.bool(true)
}

//export ClapGo_GetPluginID
func ClapGo_GetPluginID(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	
	if id == PluginID {
		return C.CString(PluginID)
	}
	
	return C.CString("unknown")
}

//export ClapGo_GetPluginName
func ClapGo_GetPluginName(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	
	if id == PluginID {
		return C.CString(PluginName)
	}
	
	return C.CString("Unknown Plugin")
}

//export ClapGo_GetPluginVendor
func ClapGo_GetPluginVendor(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	
	if id == PluginID {
		return C.CString(PluginVendor)
	}
	
	return C.CString("Unknown Vendor")
}

//export ClapGo_GetPluginVersion
func ClapGo_GetPluginVersion(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	
	if id == PluginID {
		return C.CString(PluginVersion)
	}
	
	return C.CString("0.0.0")
}

//export ClapGo_GetPluginDescription
func ClapGo_GetPluginDescription(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	
	if id == PluginID {
		return C.CString(PluginDescription)
	}
	
	return C.CString("No description available")
}

// Plugin lifecycle functions

//export ClapGo_PluginInit
func ClapGo_PluginInit(plugin unsafe.Pointer) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*GainPlugin)
	return C.bool(p.Init())
}

//export ClapGo_PluginDestroy
func ClapGo_PluginDestroy(plugin unsafe.Pointer) {
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
	if plugin == nil {
		return C.bool(false)
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*GainPlugin)
	return C.bool(p.Activate(float64(sampleRate), uint32(minFrames), uint32(maxFrames)))
}

//export ClapGo_PluginDeactivate
func ClapGo_PluginDeactivate(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*GainPlugin)
	p.Deactivate()
}

//export ClapGo_PluginStartProcessing
func ClapGo_PluginStartProcessing(plugin unsafe.Pointer) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*GainPlugin)
	return C.bool(p.StartProcessing())
}

//export ClapGo_PluginStopProcessing
func ClapGo_PluginStopProcessing(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*GainPlugin)
	p.StopProcessing()
}

//export ClapGo_PluginReset
func ClapGo_PluginReset(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*GainPlugin)
	p.Reset()
}

//export ClapGo_PluginProcess
func ClapGo_PluginProcess(plugin unsafe.Pointer, process unsafe.Pointer) C.int32_t {
	if plugin == nil || process == nil {
		return C.int32_t(api.ProcessError)
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*GainPlugin)
	
	// Convert the C clap_process_t to Go parameters
	cProcess := (*C.clap_process_t)(process)
	
	// Extract steady time and frame count
	steadyTime := int64(cProcess.steady_time)
	framesCount := uint32(cProcess.frames_count)
	
	// Convert audio buffers
	audioIn := convertAudioBuffersToGo(cProcess.audio_inputs, cProcess.audio_inputs_count, framesCount)
	audioOut := convertAudioBuffersToGo(cProcess.audio_outputs, cProcess.audio_outputs_count, framesCount)
	
	// Create event handler for input/output events
	eventHandler := &ProcessEventHandler{
		inputEvents:  cProcess.in_events,
		outputEvents: cProcess.out_events,
	}
	
	// Call the actual Go process method
	result := p.Process(steadyTime, framesCount, audioIn, audioOut, eventHandler)
	
	return C.int32_t(result)
}

//export ClapGo_PluginGetExtension
func ClapGo_PluginGetExtension(plugin unsafe.Pointer, id *C.char) unsafe.Pointer {
	extID := C.GoString(id)
	if plugin == nil {
		return nil
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*GainPlugin)
	return p.GetExtension(extID)
}

//export ClapGo_PluginOnMainThread
func ClapGo_PluginOnMainThread(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*GainPlugin)
	p.OnMainThread()
}

// Helper functions for audio buffer conversion

// convertAudioBuffersToGo converts C audio buffers to Go slices
func convertAudioBuffersToGo(cBuffers *C.clap_audio_buffer_t, bufferCount C.uint32_t, frameCount uint32) [][]float32 {
	if cBuffers == nil || bufferCount == 0 {
		return nil
	}
	
	// Convert C array to Go slice using unsafe pointer arithmetic
	buffers := (*[1024]C.clap_audio_buffer_t)(unsafe.Pointer(cBuffers))[:bufferCount:bufferCount]
	
	result := make([][]float32, 0)
	
	for i := uint32(0); i < uint32(bufferCount); i++ {
		buffer := &buffers[i]
		
		// Handle 32-bit float buffers (most common)
		if buffer.data32 != nil {
			channelCount := uint32(buffer.channel_count)
			
			// Convert C channel pointers to Go slices
			channels := (*[64]*C.float)(unsafe.Pointer(buffer.data32))[:channelCount:channelCount]
			
			for ch := uint32(0); ch < channelCount; ch++ {
				if channels[ch] != nil {
					// Convert C float array to Go slice
					channelData := (*[1048576]float32)(unsafe.Pointer(channels[ch]))[:frameCount:frameCount]
					result = append(result, channelData)
				}
			}
		}
		// Note: 64-bit double buffers could be handled here if needed
	}
	
	return result
}

// ProcessEventHandler implements the EventHandler interface for CLAP events
type ProcessEventHandler struct {
	inputEvents  *C.clap_input_events_t
	outputEvents *C.clap_output_events_t
}

// ProcessInputEvents processes all incoming events
func (h *ProcessEventHandler) ProcessInputEvents() {
	// For now, this is a simplified implementation
	// In a full implementation, this would process all CLAP events
}

// AddOutputEvent adds an event to the output queue
func (h *ProcessEventHandler) AddOutputEvent(eventType int, data interface{}) {
	// For now, this is a simplified implementation
	// In a full implementation, this would create and push CLAP events
}

// GetInputEventCount returns the number of input events
func (h *ProcessEventHandler) GetInputEventCount() uint32 {
	if h.inputEvents == nil {
		return 0
	}
	
	// Call the CLAP API to get event count
	// Using the size function pointer from the input events structure
	if h.inputEvents.size != nil {
		return uint32(C.clap_input_events_size_helper(h.inputEvents))
	}
	
	return 0
}

// GetInputEvent retrieves an input event by index
func (h *ProcessEventHandler) GetInputEvent(index uint32) *api.Event {
	if h.inputEvents == nil {
		return nil
	}
	
	// This would need proper implementation to convert CLAP events to Go events
	// For now, return nil to prevent crashes
	return nil
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