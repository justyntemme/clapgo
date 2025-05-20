package main

/*
#include <stdint.h>
#include <stdbool.h>
#include <stdlib.h>
#include "../../include/clap/include/clap/clap.h"

// Forward declarations of exported functions
uint32_t GetPluginCount();
struct clap_plugin_descriptor* GetPluginInfo(uint32_t index);
void* CreatePlugin(struct clap_host *host, char *plugin_id);
bool GetVersion(uint32_t *major, uint32_t *minor, uint32_t *patch);

// Plugin lifecycle functions
bool GoInit(void *plugin);
void GoDestroy(void *plugin);
bool GoActivate(void *plugin, double sample_rate, uint32_t min_frames, uint32_t max_frames);
void GoDeactivate(void *plugin);
bool GoStartProcessing(void *plugin);
void GoStopProcessing(void *plugin);
void GoReset(void *plugin);
int32_t GoProcess(void *plugin, struct clap_process *process);
void* GoGetExtension(void *plugin, char *id);
void GoOnMainThread(void *plugin);
*/
import "C"
import (
	"fmt"
	"runtime/cgo"
	"sync"
	"unsafe"

	"github.com/justyntemme/clapgo/internal/registry"
	"github.com/justyntemme/clapgo/pkg/api"
)

// handleRegistry maintains a registry of all allocated handles for proper cleanup
var handleRegistry = struct {
	sync.RWMutex
	handles map[uintptr]bool
}{
	handles: make(map[uintptr]bool),
}

//export GetPluginCount
func GetPluginCount() C.uint32_t {
	return C.uint32_t(registry.GetPluginCount())
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
	plugin := registry.CreatePlugin(id)
	
	if plugin == nil {
		return nil
	}
	
	// Create a handle to the plugin
	handle := cgo.NewHandle(plugin)
	
	// Register the handle
	handleRegistry.Lock()
	handleRegistry.handles[uintptr(handle)] = true
	handleRegistry.Unlock()
	
	// Print some debugging info
	fmt.Printf("Creating plugin with ID: %s\n", id)
	
	return unsafe.Pointer(uintptr(handle))
}

//export GetVersion
func GetVersion(major, minor, patch *C.uint32_t) C.bool {
	// Return the bridge version
	if major != nil {
		*major = 0
	}
	if minor != nil {
		*minor = 2
	}
	if patch != nil {
		*patch = 0
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

// ClapEventHandler implements the api.EventHandler interface
// for CLAP events.
type ClapEventHandler struct {
	InEvents  unsafe.Pointer
	OutEvents unsafe.Pointer
}

// ProcessInputEvents processes all events in the input queue.
func (e *ClapEventHandler) ProcessInputEvents() {
	// Actual implementation would process events from CLAP
}

// AddOutputEvent adds an event to the output queue.
func (e *ClapEventHandler) AddOutputEvent(eventType int, data interface{}) {
	// Actual implementation would send events to CLAP
}

// GetInputEventCount returns the number of input events.
func (e *ClapEventHandler) GetInputEventCount() uint32 {
	// Actual implementation would get count from CLAP
	return 0
}

// GetInputEvent retrieves an input event by index.
func (e *ClapEventHandler) GetInputEvent(index uint32) *api.Event {
	// Actual implementation would get event from CLAP
	return nil
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
		audioIn = make([][]float32, 0, inputsCount)
		
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
		audioOut = make([][]float32, 0, outputsCount)
		
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

func init() {
	fmt.Println("Initializing ClapGo bridge...")
}