package main

/*
#include <stdint.h>
#include <stdbool.h>
#include <stdlib.h>
#include "../../include/clap/include/clap/clap.h"

// Helper functions to call function pointers in the clap_input_events interface
static uint32_t clap_input_events_size(const clap_input_events_t *events, const void *events_ctx) {
    if (events && events->size) {
        return events->size(events_ctx);
    }
    return 0;
}

static const clap_event_header_t *clap_input_events_get(const clap_input_events_t *events, 
                                                const void *events_ctx, 
                                                uint32_t index) {
    if (events && events->get) {
        return events->get(events_ctx, index);
    }
    return NULL;
}
*/
import "C"
import (
	"fmt"
	"os"
	"runtime/cgo"
	"sync"
	"unsafe"

	"github.com/justyntemme/clapgo/internal/registry"
	"github.com/justyntemme/clapgo/pkg/api"
)

// Version information - must match CLAPGO_API_VERSION in bridge.h
const (
	VersionMajor = 0
	VersionMinor = 2
	VersionPatch = 0
)

// handleRegistry maintains a registry of all allocated handles for proper cleanup
var handleRegistry = struct {
	sync.RWMutex
	handles map[uintptr]bool
}{
	handles: make(map[uintptr]bool),
}

func init() {
	fmt.Println("Initializing Go bridge")
	
	// Check for plugin registrations
	// These will come from init() functions in the example plugins
	if registry.GetPluginCount() == 0 {
		fmt.Println("Warning: No plugins registered yet")
	}
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
	
	// Register the handle
	handleRegistry.Lock()
	handleRegistry.handles[uintptr(handle)] = true
	handleRegistry.Unlock()
	
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

// GetPluginFromPtr retrieves the Go plugin from a plugin pointer
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
	
	handleRegistry.Lock()
	if _, exists := handleRegistry.handles[uintptr(handle)]; exists {
		handle.Delete()
		delete(handleRegistry.handles, uintptr(handle))
	}
	handleRegistry.Unlock()
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

// EventHandler implements the api.EventHandler interface
// for CLAP events.
type ClapEventHandler struct {
	InEvents  unsafe.Pointer
	OutEvents unsafe.Pointer
}

// ProcessInputEvents processes all events in the input queue.
func (e *ClapEventHandler) ProcessInputEvents() {
	// Implementation depends on specific event types needed
}

// AddOutputEvent adds an event to the output queue.
func (e *ClapEventHandler) AddOutputEvent(eventType int, data interface{}) {
	// Implementation depends on specific event types needed
}

// GetInputEventCount returns the number of input events.
func (e *ClapEventHandler) GetInputEventCount() uint32 {
	if e.InEvents == nil {
		return 0
	}
	
	inEvents := (*C.clap_input_events_t)(e.InEvents)
	// Call our C helper function to invoke the size function pointer
	return uint32(C.clap_input_events_size(inEvents, e.InEvents))
}

// GetInputEvent retrieves an input event by index.
func (e *ClapEventHandler) GetInputEvent(index uint32) *api.Event {
	if e.InEvents == nil {
		return nil
	}
	
	inEvents := (*C.clap_input_events_t)(e.InEvents)
	// Call our C helper function to invoke the get function pointer
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
	// Add more event types as needed
	}
	
	return goEvent
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
	events := &ClapEventHandler{
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

// CleanupAllHandles releases all registered handles
func CleanupAllHandles() {
	handleRegistry.Lock()
	defer handleRegistry.Unlock()
	
	for h := range handleRegistry.handles {
		handle := cgo.Handle(h)
		handle.Delete()
		delete(handleRegistry.handles, h)
	}
}

func main() {
	// This is required for buildmode=c-shared,
	// but not used directly
	
	// Get the plugin ID from environment variable
	pluginID := os.Getenv("CLAPGO_PLUGIN_ID")
	if pluginID != "" {
		fmt.Printf("ClapGo bridge initialized for plugin ID: %s\n", pluginID)
		
		// Verify the plugin ID is registered
		count := registry.GetPluginCount()
		found := false
		
		for i := uint32(0); i < count; i++ {
			info := registry.GetPluginInfo(i)
			if info.ID == pluginID {
				found = true
				fmt.Printf("Plugin '%s' (%s) successfully registered\n", info.Name, info.ID)
				break
			}
		}
		
		if !found {
			fmt.Printf("Warning: Plugin ID '%s' not found in registry. Available plugins: %d\n", pluginID, count)
		}
	} else {
		fmt.Println("ClapGo bridge initialized (no specific plugin ID)")
		
		// List all registered plugins
		count := registry.GetPluginCount()
		if count > 0 {
			fmt.Printf("Found %d registered plugins:\n", count)
			for i := uint32(0); i < count; i++ {
				info := registry.GetPluginInfo(i)
				fmt.Printf("  %d: %s (%s)\n", i, info.Name, info.ID)
			}
		} else {
			fmt.Println("Warning: No plugins registered")
		}
	}
}