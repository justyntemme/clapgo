// Package main provides a bridge between the CLAP C API and Go.
// It handles all CGO interactions and type conversions and builds as a shared library.
package main

// #include <stdint.h>
// #include <stdbool.h>
// #include <stdlib.h>
// #include "../../include/clap/include/clap/clap.h"
//
// // Forward declarations of the standardized functions exported by Go
// uint32_t ClapGo_GetPluginCount();
// struct clap_plugin_descriptor *ClapGo_GetPluginDescriptor(uint32_t index);
// void* ClapGo_CreatePlugin(struct clap_host *host, char *plugin_id);
// bool ClapGo_GetVersion(uint32_t *major, uint32_t *minor, uint32_t *patch);
//
// // Plugin metadata export functions
// uint32_t ClapGo_GetRegisteredPluginCount();
// char* ClapGo_GetRegisteredPluginIDByIndex(uint32_t index);
// char* ClapGo_GetPluginID(char* plugin_id);
// char* ClapGo_GetPluginName(char* plugin_id);
// char* ClapGo_GetPluginVendor(char* plugin_id);
// char* ClapGo_GetPluginVersion(char* plugin_id);
// char* ClapGo_GetPluginDescription(char* plugin_id);
//
// // Plugin lifecycle functions
// bool ClapGo_PluginInit(void *plugin);
// void ClapGo_PluginDestroy(void *plugin);
// bool ClapGo_PluginActivate(void *plugin, double sample_rate, uint32_t min_frames, uint32_t max_frames);
// void ClapGo_PluginDeactivate(void *plugin);
// bool ClapGo_PluginStartProcessing(void *plugin);
// void ClapGo_PluginStopProcessing(void *plugin);
// void ClapGo_PluginReset(void *plugin);
// int32_t ClapGo_PluginProcess(void *plugin, struct clap_process *process);
// void *ClapGo_PluginGetExtension(void *plugin, char *id);
// void ClapGo_PluginOnMainThread(void *plugin);
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

//export ClapGo_GetPluginCount
func ClapGo_GetPluginCount() C.uint32_t {
	// Return the count from the registry
	count := registry.GetPluginCount()
	fmt.Printf("Go: ClapGo_GetPluginCount returning %d\n", count)
	return C.uint32_t(count)
}

//export ClapGo_GetPluginDescriptor
// ClapGo_GetPluginDescriptor returns a plugin descriptor for the plugin at the given index
func ClapGo_GetPluginDescriptor(index C.uint32_t) *C.struct_clap_plugin_descriptor {
	fmt.Printf("Go: ClapGo_GetPluginDescriptor for index: %d\n", index)
	
	// Get plugin info from the registry
	registryCount := registry.GetPluginCount()
	if uint32(index) < registryCount {
		info := registry.GetPluginInfo(uint32(index))
		return createDescriptorFromPluginInfo(info)
	}
	
	fmt.Printf("Error: Plugin index %d out of range\n", index)
	return nil
}

// Helper function to create a descriptor from plugin info
func createDescriptorFromPluginInfo(info api.PluginInfo) *C.struct_clap_plugin_descriptor {
	// Create a new descriptor
	desc := C.calloc(1, C.size_t(unsafe.Sizeof(C.struct_clap_plugin_descriptor{})))
	descriptor := (*C.struct_clap_plugin_descriptor)(desc)
	
	// Initialize clap_version
	descriptor.clap_version.major = 1
	descriptor.clap_version.minor = 1
	descriptor.clap_version.revision = 0
	
	// Convert all strings to C strings
	descriptor.id = C.CString(info.ID)
	descriptor.name = C.CString(info.Name)
	descriptor.vendor = C.CString(info.Vendor)
	descriptor.url = C.CString(info.URL)
	descriptor.manual_url = C.CString(info.ManualURL)
	descriptor.support_url = C.CString(info.SupportURL)
	descriptor.version = C.CString(info.Version)
	descriptor.description = C.CString(info.Description)
	
	// Add feature strings
	features := info.Features
	
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
	
	fmt.Printf("Created descriptor for: %s\n", info.ID)
	return descriptor
}


//export ClapGo_CreatePlugin
func ClapGo_CreatePlugin(host *C.struct_clap_host, pluginID *C.char) unsafe.Pointer {
	id := C.GoString(pluginID)
	fmt.Printf("Go: ClapGo_CreatePlugin with ID: %s\n", id)
	
	// Create a plugin from the registry
	plug := registry.CreatePlugin(id)
	if plug != nil {
		// Set host if needed
		
		// Create a handle to the plugin
		handle := cgo.NewHandle(plug)
		
		// Register the handle for cleanup
		registry.RegisterHandle(handle)
		
		fmt.Printf("Created plugin instance from registry: %s\n", id)
		return unsafe.Pointer(uintptr(handle))
	}
	
	fmt.Printf("Error: Unable to create plugin with ID: %s\n", id)
	return nil
}

//export ClapGo_GetVersion
func ClapGo_GetVersion(major, minor, patch *C.uint32_t) C.bool {
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
	fmt.Printf("Go: ClapGo_GetVersion returning %d.%d.%d\n", VersionMajor, VersionMinor, VersionPatch)
	return C.bool(true)
}

//export ClapGo_PluginInit
func ClapGo_PluginInit(plugin unsafe.Pointer) C.bool {
	p := GetPluginFromPtr(plugin)
	if p == nil {
		fmt.Printf("Error: Failed to get plugin from pointer in ClapGo_PluginInit\n")
		return C.bool(false)
	}
	
	return C.bool(p.Init())
}

//export ClapGo_PluginDestroy
func ClapGo_PluginDestroy(plugin unsafe.Pointer) {
	p := GetPluginFromPtr(plugin)
	if p == nil {
		fmt.Printf("Error: Failed to get plugin from pointer in ClapGo_PluginDestroy\n")
		return
	}
	
	fmt.Printf("Go: Destroying plugin\n")
	p.Destroy()
	
	// Release the handle to prevent memory leaks
	handle := cgo.Handle(uintptr(plugin))
	registry.UnregisterHandle(handle)
}

//export ClapGo_PluginActivate
func ClapGo_PluginActivate(plugin unsafe.Pointer, sampleRate C.double, minFrames, maxFrames C.uint32_t) C.bool {
	p := GetPluginFromPtr(plugin)
	if p == nil {
		fmt.Printf("Error: Failed to get plugin from pointer in ClapGo_PluginActivate\n")
		return C.bool(false)
	}
	
	fmt.Printf("Go: Activating plugin with sample rate %f\n", sampleRate)
	return C.bool(p.Activate(float64(sampleRate), uint32(minFrames), uint32(maxFrames)))
}

//export ClapGo_PluginDeactivate
func ClapGo_PluginDeactivate(plugin unsafe.Pointer) {
	p := GetPluginFromPtr(plugin)
	if p == nil {
		fmt.Printf("Error: Failed to get plugin from pointer in ClapGo_PluginDeactivate\n")
		return
	}
	
	fmt.Printf("Go: Deactivating plugin\n")
	p.Deactivate()
}

//export ClapGo_PluginStartProcessing
func ClapGo_PluginStartProcessing(plugin unsafe.Pointer) C.bool {
	p := GetPluginFromPtr(plugin)
	if p == nil {
		fmt.Printf("Error: Failed to get plugin from pointer in ClapGo_PluginStartProcessing\n")
		return C.bool(false)
	}
	
	fmt.Printf("Go: Starting processing\n")
	return C.bool(p.StartProcessing())
}

//export ClapGo_PluginStopProcessing
func ClapGo_PluginStopProcessing(plugin unsafe.Pointer) {
	p := GetPluginFromPtr(plugin)
	if p == nil {
		fmt.Printf("Error: Failed to get plugin from pointer in ClapGo_PluginStopProcessing\n")
		return
	}
	
	fmt.Printf("Go: Stopping processing\n")
	p.StopProcessing()
}

//export ClapGo_PluginReset
func ClapGo_PluginReset(plugin unsafe.Pointer) {
	p := GetPluginFromPtr(plugin)
	if p == nil {
		fmt.Printf("Error: Failed to get plugin from pointer in ClapGo_PluginReset\n")
		return
	}
	
	fmt.Printf("Go: Resetting plugin\n")
	p.Reset()
}

//export ClapGo_PluginProcess
func ClapGo_PluginProcess(plugin unsafe.Pointer, process *C.struct_clap_process) C.int32_t {
	p := GetPluginFromPtr(plugin)
	if p == nil {
		fmt.Printf("Error: Failed to get plugin from pointer in ClapGo_PluginProcess\n")
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

//export ClapGo_PluginGetExtension
func ClapGo_PluginGetExtension(plugin unsafe.Pointer, id *C.char) unsafe.Pointer {
	p := GetPluginFromPtr(plugin)
	if p == nil {
		fmt.Printf("Error: Failed to get plugin from pointer in ClapGo_PluginGetExtension\n")
		return nil
	}
	
	extID := C.GoString(id)
	fmt.Printf("Go: Getting extension %s\n", extID)
	return p.GetExtension(extID)
}

//export ClapGo_PluginOnMainThread
func ClapGo_PluginOnMainThread(plugin unsafe.Pointer) {
	p := GetPluginFromPtr(plugin)
	if p == nil {
		fmt.Printf("Error: Failed to get plugin from pointer in ClapGo_PluginOnMainThread\n")
		return
	}
	
	fmt.Printf("Go: Executing on main thread\n")
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

//export ClapGo_GetRegisteredPluginCount
func ClapGo_GetRegisteredPluginCount() uint32 {
	count := registry.GetPluginCount()
	fmt.Printf("Go: ClapGo_GetRegisteredPluginCount returning %d\n", count)
	return count
}

//export ClapGo_GetRegisteredPluginIDByIndex
func ClapGo_GetRegisteredPluginIDByIndex(index uint32) *C.char {
	info := registry.GetPluginInfo(index)
	if info.ID == "" {
		fmt.Printf("Go: Plugin at index %d not found\n", index)
		return C.CString("")
	}
	fmt.Printf("Go: Found plugin at index %d with ID %s\n", index, info.ID)
	return C.CString(info.ID)
}

//export ClapGo_GetPluginID
func ClapGo_GetPluginID(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	fmt.Printf("Go: Getting plugin ID for %s\n", id)
	
	// Find the plugin info by ID
	count := registry.GetPluginCount()
	for i := uint32(0); i < count; i++ {
		info := registry.GetPluginInfo(i)
		if info.ID == id {
			fmt.Printf("Go: Found plugin ID %s\n", info.ID)
			return C.CString(info.ID)
		}
	}
	
	fmt.Printf("Go: Plugin ID %s not found\n", id)
	return C.CString("unknown")
}

//export ClapGo_GetPluginName
func ClapGo_GetPluginName(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	fmt.Printf("Go: Getting plugin name for %s\n", id)
	
	// Find the plugin info by ID
	count := registry.GetPluginCount()
	for i := uint32(0); i < count; i++ {
		info := registry.GetPluginInfo(i)
		if info.ID == id {
			fmt.Printf("Go: Found plugin name %s\n", info.Name)
			return C.CString(info.Name)
		}
	}
	
	fmt.Printf("Go: Plugin name for %s not found\n", id)
	return C.CString("Unknown Plugin")
}

//export ClapGo_GetPluginVendor
func ClapGo_GetPluginVendor(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	fmt.Printf("Go: Getting plugin vendor for %s\n", id)
	
	// Find the plugin info by ID
	count := registry.GetPluginCount()
	for i := uint32(0); i < count; i++ {
		info := registry.GetPluginInfo(i)
		if info.ID == id {
			fmt.Printf("Go: Found plugin vendor %s\n", info.Vendor)
			return C.CString(info.Vendor)
		}
	}
	
	fmt.Printf("Go: Plugin vendor for %s not found\n", id)
	return C.CString("Unknown Vendor")
}

//export ClapGo_GetPluginVersion
func ClapGo_GetPluginVersion(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	fmt.Printf("Go: Getting plugin version for %s\n", id)
	
	// Find the plugin info by ID
	count := registry.GetPluginCount()
	for i := uint32(0); i < count; i++ {
		info := registry.GetPluginInfo(i)
		if info.ID == id {
			fmt.Printf("Go: Found plugin version %s\n", info.Version)
			return C.CString(info.Version)
		}
	}
	
	fmt.Printf("Go: Plugin version for %s not found\n", id)
	return C.CString("0.0.0")
}

//export ClapGo_GetPluginDescription
func ClapGo_GetPluginDescription(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	fmt.Printf("Go: Getting plugin description for %s\n", id)
	
	// Find the plugin info by ID
	count := registry.GetPluginCount()
	for i := uint32(0); i < count; i++ {
		info := registry.GetPluginInfo(i)
		if info.ID == id {
			fmt.Printf("Go: Found plugin description: %s\n", info.Description)
			return C.CString(info.Description)
		}
	}
	
	fmt.Printf("Go: Plugin description for %s not found\n", id)
	return C.CString("No description available")
}

// Note: The ClapGo bridge uses both the manifest system and the registry system together.
// The manifest system (on the C side) loads plugin metadata from JSON files and identifies
// which shared library to load. When the C side loads that shared library, the Go code in that
// library registers plugins with the registry system. This allows us to have a standardized
// approach to plugin management that is consistent with the CLAP plugin ecosystem.
//
// The standardized export functions (ClapGo_*) provide a stable interface between the Go 
// and C sides of the bridge.