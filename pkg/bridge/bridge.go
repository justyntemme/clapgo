// Package main provides a bridge between the CLAP C API and Go.
// It handles all CGO interactions and type conversions and builds as a shared library.
package main

// #include <stdint.h>
// #include <stdbool.h>
// #include <stdlib.h>
// #include "../../include/clap/include/clap/clap.h"
//
// // Forward declarations of the functions exported by Go
// uint32_t GetPluginCount();
// struct clap_plugin_descriptor *GetPluginInfo(uint32_t index);
// void* CreatePlugin(struct clap_host *host, char *plugin_id);
// bool GetVersion(uint32_t *major, uint32_t *minor, uint32_t *patch);
//
// // Plugin metadata export functions
// uint32_t GetRegisteredPluginCount();
// char* GetRegisteredPluginIDByIndex(uint32_t index);
// char* ExportPluginID(char* plugin_id);
// char* ExportPluginName(char* plugin_id);
// char* ExportPluginVendor(char* plugin_id);
// char* ExportPluginVersion(char* plugin_id);
// char* ExportPluginDescription(char* plugin_id);
//
// // Plugin lifecycle functions
// bool GoInit(void *plugin);
// void GoDestroy(void *plugin);
// bool GoActivate(void *plugin, double sample_rate, uint32_t min_frames, uint32_t max_frames);
// void GoDeactivate(void *plugin);
// bool GoStartProcessing(void *plugin);
// void GoStopProcessing(void *plugin);
// void GoReset(void *plugin);
// int32_t GoProcess(void *plugin, struct clap_process *process);
// void *GoGetExtension(void *plugin, char *id);
// void GoOnMainThread(void *plugin);
//
// // Helpers for handling events
// static inline uint32_t clap_input_events_size(const clap_input_events_t* events, const void* events_ctx) {
//     if (events && events->size) {
//         return events->size(events_ctx);
//     }
//     return 0;
// }
//
// static inline const clap_event_header_t* clap_input_events_get(const clap_input_events_t* events, 
//                                            const void* events_ctx, 
//                                            uint32_t index) {
//     if (events && events->get) {
//         return events->get(events_ctx, index);
//     }
//     return NULL;
// }
//
// static inline bool clap_output_events_try_push(const clap_output_events_t* events, 
//                                 const void* events_ctx,
//                                 const clap_event_header_t* event) {
//     if (events && events->try_push) {
//         return events->try_push(events_ctx, event);
//     }
//     return false;
// }
import "C"
import (
	"fmt"
	"runtime/cgo"
	"unsafe"

	"github.com/justyntemme/clapgo/pkg/api"
	"github.com/justyntemme/clapgo/pkg/registry"
)

// Version information for the bridge
const (
	VersionMajor = 0
	VersionMinor = 2
	VersionPatch = 0
)

// GainPlugin implements a simple gain plugin directly in the bridge
// This is a temporary solution until we find a way to properly load plugins
type GainPlugin struct {
	// Plugin state
	gain         float64
	sampleRate   float64
	isActivated  bool
	isProcessing bool
	host         unsafe.Pointer
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
	return nil
}

// OnMainThread is called on the main thread
func (p *GainPlugin) OnMainThread() {
	// Nothing to do
}

// GetPluginInfo returns information about the plugin
func (p *GainPlugin) GetPluginInfo() api.PluginInfo {
	return api.PluginInfo{
		ID:          "com.clapgo.gain",
		Name:        "Simple Gain",
		Vendor:      "ClapGo",
		URL:         "https://github.com/justyntemme/clapgo",
		ManualURL:   "https://github.com/justyntemme/clapgo",
		SupportURL:  "https://github.com/justyntemme/clapgo/issues",
		Version:     "1.0.0",
		Description: "A simple gain plugin using ClapGo",
		Features:    []string{"audio-effect", "stereo", "mono"},
	}
}

// GetPluginID returns the plugin ID
func (p *GainPlugin) GetPluginID() string {
	return "com.clapgo.gain"
}

// EventHandler implements the api.EventHandler interface for CLAP events
type EventHandler struct {
	InEvents  unsafe.Pointer
	OutEvents unsafe.Pointer
}

// ProcessInputEvents processes all events in the input queue
func (e *EventHandler) ProcessInputEvents() {
	// Implementation depends on specific event types needed
	// Each plugin should implement its own event processing
}

// GetPluginFromPtr retrieves the Go plugin from a plugin pointer.
func GetPluginFromPtr(ptr unsafe.Pointer) api.Plugin {
	if ptr == nil {
		return nil
	}

	// Convert the plugin pointer back to a handle
	handle := cgo.Handle(uintptr(ptr))

	// Get the plugin from the handle
	value := handle.Value()

	// Try to cast to Plugin
	plugin, ok := value.(api.Plugin)
	if !ok {
		fmt.Printf("Error: Failed to cast handle value to Plugin, got %T\n", value)
		return nil
	}

	return plugin
}

// AddOutputEvent adds an event to the output queue
func (e *EventHandler) AddOutputEvent(eventType int, data interface{}) {
	// Implementation depends on specific event types needed
}

// GetInputEventCount returns the number of input events
func (e *EventHandler) GetInputEventCount() uint32 {
	if e.InEvents == nil {
		return 0
	}

	inEvents := (*C.clap_input_events_t)(e.InEvents)
	return uint32(C.clap_input_events_size(inEvents, e.InEvents))
}

// GetInputEvent retrieves an input event by index
func (e *EventHandler) GetInputEvent(index uint32) *api.Event {
	if e.InEvents == nil {
		return nil
	}

	inEvents := (*C.clap_input_events_t)(e.InEvents)
	eventPtr := unsafe.Pointer(C.clap_input_events_get(inEvents, e.InEvents, C.uint32_t(index)))
	if eventPtr == nil {
		return nil
	}

	// Convert C event to Go event
	event := (*C.clap_event_header_t)(eventPtr)

	// Create a Go event based on the type
	goEvent := &api.Event{
		Type: int(event._type),
		Time: uint32(event.time),
	}

	// Process specific event types
	switch event._type {
	case C.CLAP_EVENT_PARAM_VALUE:
		paramEvent := (*C.clap_event_param_value_t)(eventPtr)
		goEvent.Data = api.ParamEvent{
			ParamID: uint32(paramEvent.param_id),
			Cookie:  unsafe.Pointer(paramEvent.cookie),
			Value:   float64(paramEvent.value),
		}
	case C.CLAP_EVENT_NOTE_ON:
		noteEvent := (*C.clap_event_note_t)(eventPtr)
		goEvent.Data = api.NoteEvent{
			NoteID:  int32(noteEvent.note_id),
			Port:    int16(noteEvent.port_index),
			Channel: int16(noteEvent.channel),
			Key:     int16(noteEvent.key),
			Value:   float64(noteEvent.velocity),
		}
	case C.CLAP_EVENT_NOTE_OFF:
		noteEvent := (*C.clap_event_note_t)(eventPtr)
		goEvent.Data = api.NoteEvent{
			NoteID:  int32(noteEvent.note_id),
			Port:    int16(noteEvent.port_index),
			Channel: int16(noteEvent.channel),
			Key:     int16(noteEvent.key),
			Value:   float64(noteEvent.velocity),
		}
	case C.CLAP_EVENT_NOTE_CHOKE:
		noteEvent := (*C.clap_event_note_t)(eventPtr)
		goEvent.Data = api.NoteEvent{
			NoteID:  int32(noteEvent.note_id),
			Port:    int16(noteEvent.port_index),
			Channel: int16(noteEvent.channel),
			Key:     int16(noteEvent.key),
			Value:   float64(noteEvent.velocity),
		}
	// Add more event types as needed
	}

	return goEvent
}

//export GetPluginCount
func GetPluginCount() C.uint32_t {
	// This is a temporary fix: We hardcode 1 because we know we have one plugin
	// In a real implementation, this should scan the plugin directory
	// or use plugin metadata to determine the count
	fmt.Printf("Go: GetPluginCount (hardcoded to 1)\n")
	return C.uint32_t(1)
}

//export GetPluginInfo
func GetPluginInfo(index C.uint32_t) *C.struct_clap_plugin_descriptor {
	// For now, hardcode the info for our one plugin since registry might be empty
	// In a real implementation, this should use plugin metadata
	
	// Create a new descriptor
	desc := C.calloc(1, C.size_t(unsafe.Sizeof(C.struct_clap_plugin_descriptor{})))
	descriptor := (*C.struct_clap_plugin_descriptor)(desc)
	
	// Initialize clap_version
	descriptor.clap_version.major = 1
	descriptor.clap_version.minor = 1
	descriptor.clap_version.revision = 0
	
	// Convert all strings to C strings
	descriptor.id = C.CString("com.clapgo.gain")
	descriptor.name = C.CString("Simple Gain")
	descriptor.vendor = C.CString("ClapGo")
	descriptor.url = C.CString("https://github.com/justyntemme/clapgo")
	descriptor.manual_url = C.CString("https://github.com/justyntemme/clapgo")
	descriptor.support_url = C.CString("https://github.com/justyntemme/clapgo/issues")
	descriptor.version = C.CString("1.0.0")
	descriptor.description = C.CString("A simple gain plugin using ClapGo")
	
	// Add feature strings
	features := []string{"audio-effect", "stereo", "mono"}
	
	// Allocate memory for the feature array (plus 1 for NULL terminator)
	featureArray := C.calloc(C.size_t(len(features)+1), C.size_t(unsafe.Sizeof(uintptr(0))))
	featuresPtr := (*[1<<30]*C.char)(featureArray)
	
	// Add each feature string
	for i, feature := range features {
		featuresPtr[i] = C.CString(feature)
	}
	// Set the NULL terminator
	featuresPtr[len(features)] = nil
	
	descriptor.features = (**C.char)(featureArray)
	
	fmt.Printf("GetPluginInfo returning descriptor for: %s\n", C.GoString(descriptor.id))
	
	return descriptor
}

//export CreatePlugin
func CreatePlugin(host *C.struct_clap_host, pluginID *C.char) unsafe.Pointer {
	id := C.GoString(pluginID)
	fmt.Printf("Go: CreatePlugin with ID: %s\n", id)
	
	// In a real implementation, this should create the correct plugin
	// based on the ID. For now, hardcode creating a GainPlugin
	
	// Import the gain plugin package to ensure its init function runs
	// This doesn't work in the current architecture, need a better approach
	// For now, create a plugin directly
	
	// Create a new gain plugin
	plugin := &GainPlugin{
		gain:         1.0, // 0dB
		sampleRate:   44100.0,
		isActivated:  false,
		isProcessing: false,
	}
	
	// Create a handle to the plugin
	handle := cgo.NewHandle(plugin)
	
	// Register the handle for cleanup
	registry.RegisterHandle(handle)
	
	fmt.Printf("Created hardcoded gain plugin instance\n")
	
	return unsafe.Pointer(uintptr(handle))
}

//export GetVersion
func GetVersion(major, minor, patch *C.uint32_t) C.bool {
	// Return the bridge version
	if major != nil {
		*major = C.uint32_t(VersionMajor)
	}
	if minor != nil {
		*minor = C.uint32_t(VersionMinor)
	}
	if patch != nil {
		*patch = C.uint32_t(VersionPatch)
	}
	return C.bool(true)
}

//export GoInit
func GoInit(plugin unsafe.Pointer) C.bool {
	p := GetPluginFromPtr(plugin)
	if p == nil {
		return C.bool(false)
	}
	
	return C.bool(p.Init())
}

//export GoDestroy
func GoDestroy(plugin unsafe.Pointer) {
	p := GetPluginFromPtr(plugin)
	if p == nil {
		return
	}
	
	p.Destroy()
	
	// Release the handle to prevent memory leaks
	handle := cgo.Handle(uintptr(plugin))
	registry.UnregisterHandle(handle)
}

//export GoActivate
func GoActivate(plugin unsafe.Pointer, sampleRate C.double, minFrames, maxFrames C.uint32_t) C.bool {
	p := GetPluginFromPtr(plugin)
	if p == nil {
		return C.bool(false)
	}
	
	return C.bool(p.Activate(float64(sampleRate), uint32(minFrames), uint32(maxFrames)))
}

//export GoDeactivate
func GoDeactivate(plugin unsafe.Pointer) {
	p := GetPluginFromPtr(plugin)
	if p == nil {
		return
	}
	
	p.Deactivate()
}

//export GoStartProcessing
func GoStartProcessing(plugin unsafe.Pointer) C.bool {
	p := GetPluginFromPtr(plugin)
	if p == nil {
		return C.bool(false)
	}
	
	return C.bool(p.StartProcessing())
}

//export GoStopProcessing
func GoStopProcessing(plugin unsafe.Pointer) {
	p := GetPluginFromPtr(plugin)
	if p == nil {
		return
	}
	
	p.StopProcessing()
}

//export GoReset
func GoReset(plugin unsafe.Pointer) {
	p := GetPluginFromPtr(plugin)
	if p == nil {
		return
	}
	
	p.Reset()
}

//export GoProcess
func GoProcess(plugin unsafe.Pointer, process *C.struct_clap_process) C.int32_t {
	p := GetPluginFromPtr(plugin)
	if p == nil {
		return C.int32_t(api.ProcessError)
	}
	
	steadyTime := int64(process.steady_time)
	framesCount := uint32(process.frames_count)
	
	// Convert audio buffers from C to Go
	var audioIn [][]float32
	var audioOut [][]float32
	
	// Process input audio buffers if available
	if process.audio_inputs != nil && process.audio_inputs_count > 0 {
		// Get the number of input ports
		inputsCount := int(process.audio_inputs_count)
		audioIn = make([][]float32, 0, inputsCount*2) // Assume stereo as worst case
		
		// Iterate through each input port
		for i := 0; i < inputsCount; i++ {
			// Get the current audio buffer using pointer arithmetic
			inputPtr := unsafe.Pointer(uintptr(unsafe.Pointer(process.audio_inputs)) + 
				uintptr(i)*unsafe.Sizeof(*process.audio_inputs))
			input := (*C.clap_audio_buffer_t)(inputPtr)
			
			// Process each channel in this port
			channelCount := int(input.channel_count)
			if channelCount > 0 && input.data32 != nil {
				// For each channel in this port
				for ch := 0; ch < channelCount; ch++ {
					// Get the channel data pointer
					chPtr := unsafe.Pointer(uintptr(unsafe.Pointer(input.data32)) + 
						uintptr(ch)*unsafe.Sizeof(uintptr(0)))
					dataPtr := *(**C.float)(chPtr)
					
					// Convert to Go slice without copying data
					audioChannel := unsafe.Slice((*float32)(unsafe.Pointer(dataPtr)), framesCount)
					audioIn = append(audioIn, audioChannel)
				}
			}
		}
	}
	
	// Process output audio buffers if available
	if process.audio_outputs != nil && process.audio_outputs_count > 0 {
		// Get the number of output ports
		outputsCount := int(process.audio_outputs_count)
		audioOut = make([][]float32, 0, outputsCount*2) // Assume stereo as worst case
		
		// Iterate through each output port
		for i := 0; i < outputsCount; i++ {
			// Get the current audio buffer using pointer arithmetic
			outputPtr := unsafe.Pointer(uintptr(unsafe.Pointer(process.audio_outputs)) + 
				uintptr(i)*unsafe.Sizeof(*process.audio_outputs))
			output := (*C.clap_audio_buffer_t)(outputPtr)
			
			// Process each channel in this port
			channelCount := int(output.channel_count)
			if channelCount > 0 && output.data32 != nil {
				// For each channel in this port
				for ch := 0; ch < channelCount; ch++ {
					// Get the channel data pointer
					chPtr := unsafe.Pointer(uintptr(unsafe.Pointer(output.data32)) + 
						uintptr(ch)*unsafe.Sizeof(uintptr(0)))
					dataPtr := *(**C.float)(chPtr)
					
					// Convert to Go slice without copying data
					audioChannel := unsafe.Slice((*float32)(unsafe.Pointer(dataPtr)), framesCount)
					audioOut = append(audioOut, audioChannel)
				}
			}
		}
	}
	
	// Create the event handler
	events := &EventHandler{
		InEvents:  unsafe.Pointer(process.in_events),
		OutEvents: unsafe.Pointer(process.out_events),
	}
	
	// Process the audio
	status := p.Process(steadyTime, framesCount, audioIn, audioOut, events)
	return C.int32_t(status)
}

//export GoGetExtension
func GoGetExtension(plugin unsafe.Pointer, id *C.char) unsafe.Pointer {
	p := GetPluginFromPtr(plugin)
	if p == nil {
		return nil
	}
	
	extID := C.GoString(id)
	return p.GetExtension(extID)
}

//export GoOnMainThread
func GoOnMainThread(plugin unsafe.Pointer) {
	p := GetPluginFromPtr(plugin)
	if p == nil {
		return
	}
	
	p.OnMainThread()
}

// init initializes the bridge
func init() {
	fmt.Println("Initializing ClapGo bridge")
	fmt.Println("Bridge package initialized, plugins will be registered by their respective packages.")
	fmt.Printf("Currently registered plugins: %d\n", registry.GetPluginCount())
	if registry.GetPluginCount() > 0 {
		for i := uint32(0); i < registry.GetPluginCount(); i++ {
			info := registry.GetPluginInfo(i)
			fmt.Printf("Found plugin: %s (%s)\n", info.Name, info.ID)
		}
	}
}

// Export plugin metadata functions that the C code expects

//export GetRegisteredPluginCount
func GetRegisteredPluginCount() uint32 {
	return registry.GetPluginCount()
}

//export GetRegisteredPluginIDByIndex
func GetRegisteredPluginIDByIndex(index uint32) *C.char {
	info := registry.GetPluginInfo(index)
	if info.ID == "" {
		return C.CString("")
	}
	return C.CString(info.ID)
}

//export ExportPluginID
func ExportPluginID(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	// Find the plugin info by ID
	count := registry.GetPluginCount()
	for i := uint32(0); i < count; i++ {
		info := registry.GetPluginInfo(i)
		if info.ID == id {
			return C.CString(info.ID)
		}
	}
	return C.CString("unknown")
}

//export ExportPluginName
func ExportPluginName(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	// Find the plugin info by ID
	count := registry.GetPluginCount()
	for i := uint32(0); i < count; i++ {
		info := registry.GetPluginInfo(i)
		if info.ID == id {
			return C.CString(info.Name)
		}
	}
	return C.CString("Unknown Plugin")
}

//export ExportPluginVendor
func ExportPluginVendor(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	// Find the plugin info by ID
	count := registry.GetPluginCount()
	for i := uint32(0); i < count; i++ {
		info := registry.GetPluginInfo(i)
		if info.ID == id {
			return C.CString(info.Vendor)
		}
	}
	return C.CString("Unknown Vendor")
}

//export ExportPluginVersion
func ExportPluginVersion(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	// Find the plugin info by ID
	count := registry.GetPluginCount()
	for i := uint32(0); i < count; i++ {
		info := registry.GetPluginInfo(i)
		if info.ID == id {
			return C.CString(info.Version)
		}
	}
	return C.CString("0.0.0")
}

//export ExportPluginDescription
func ExportPluginDescription(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	// Find the plugin info by ID
	count := registry.GetPluginCount()
	for i := uint32(0); i < count; i++ {
		info := registry.GetPluginInfo(i)
		if info.ID == id {
			return C.CString(info.Description)
		}
	}
	return C.CString("No description available")
}