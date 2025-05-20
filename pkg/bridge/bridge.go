// Package bridge provides a bridge between the CLAP C API and Go.
// It handles all CGO interactions and type conversions.
package bridge

// #include <stdint.h>
// #include <stdbool.h>
// #include <stdlib.h>
// #include "../../include/clap/include/clap/clap.h"
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
	count := registry.GetPluginCount()
	fmt.Printf("Go: GetPluginCount returning %d\n", count)
	return C.uint32_t(count)
}

//export GetPluginInfo
func GetPluginInfo(index C.uint32_t) *C.struct_clap_plugin_descriptor {
	info := registry.GetPluginInfo(uint32(index))
	
	// If the plugin info is empty, return nil
	if info.ID == "" {
		return nil
	}
	
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
	
	// Handle features array if any
	if len(info.Features) > 0 {
		// Allocate memory for the feature array (plus 1 for NULL terminator)
		featureArray := C.calloc(C.size_t(len(info.Features)+1), C.size_t(unsafe.Sizeof(uintptr(0))))
		features := (*[1<<30]*C.char)(featureArray)
		
		// Add each feature string
		for i, feature := range info.Features {
			features[i] = C.CString(feature)
		}
		// Set the NULL terminator
		features[len(info.Features)] = nil
		
		descriptor.features = (**C.char)(featureArray)
	}
	
	return descriptor
}

//export CreatePlugin
func CreatePlugin(host *C.struct_clap_host, pluginID *C.char) unsafe.Pointer {
	id := C.GoString(pluginID)
	fmt.Printf("Go: CreatePlugin with ID: %s\n", id)
	
	plugin := registry.CreatePlugin(id)
	
	if plugin == nil {
		fmt.Printf("Error: Failed to create plugin with ID: %s\n", id)
		return nil
	}
	
	// Create a handle to the plugin
	handle := cgo.NewHandle(plugin)
	
	// Register the handle for cleanup
	registry.RegisterHandle(handle)
	
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
	p := registry.GetPluginFromPtr(plugin)
	if p == nil {
		return C.bool(false)
	}
	
	return C.bool(p.Init())
}

//export GoDestroy
func GoDestroy(plugin unsafe.Pointer) {
	p := registry.GetPluginFromPtr(plugin)
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
	p := registry.GetPluginFromPtr(plugin)
	if p == nil {
		return C.bool(false)
	}
	
	return C.bool(p.Activate(float64(sampleRate), uint32(minFrames), uint32(maxFrames)))
}

//export GoDeactivate
func GoDeactivate(plugin unsafe.Pointer) {
	p := registry.GetPluginFromPtr(plugin)
	if p == nil {
		return
	}
	
	p.Deactivate()
}

//export GoStartProcessing
func GoStartProcessing(plugin unsafe.Pointer) C.bool {
	p := registry.GetPluginFromPtr(plugin)
	if p == nil {
		return C.bool(false)
	}
	
	return C.bool(p.StartProcessing())
}

//export GoStopProcessing
func GoStopProcessing(plugin unsafe.Pointer) {
	p := registry.GetPluginFromPtr(plugin)
	if p == nil {
		return
	}
	
	p.StopProcessing()
}

//export GoReset
func GoReset(plugin unsafe.Pointer) {
	p := registry.GetPluginFromPtr(plugin)
	if p == nil {
		return
	}
	
	p.Reset()
}

//export GoProcess
func GoProcess(plugin unsafe.Pointer, process *C.struct_clap_process) C.int32_t {
	p := registry.GetPluginFromPtr(plugin)
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
	p := registry.GetPluginFromPtr(plugin)
	if p == nil {
		return nil
	}
	
	extID := C.GoString(id)
	return p.GetExtension(extID)
}

//export GoOnMainThread
func GoOnMainThread(plugin unsafe.Pointer) {
	p := registry.GetPluginFromPtr(plugin)
	if p == nil {
		return
	}
	
	p.OnMainThread()
}

// init initializes the bridge
func init() {
	fmt.Println("Initializing ClapGo bridge")
}