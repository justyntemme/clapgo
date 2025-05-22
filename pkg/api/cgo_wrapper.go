package api

/*
#cgo CFLAGS: -I../../include/clap/include
#include "../../include/clap/include/clap/clap.h"
#include <stdlib.h>

// Helper functions for CLAP event handling
static inline uint32_t clap_input_events_size_helper(const clap_input_events_t* events) {
    if (events && events->size) {
        return events->size(events);
    }
    return 0;
}

static inline const clap_event_header_t* clap_input_events_get_helper(const clap_input_events_t* events, uint32_t index) {
    if (events && events->get) {
        return events->get(events, index);
    }
    return NULL;
}
*/
import "C"
import (
	"fmt"
	"runtime/cgo"
	"unsafe"
)

// SimplePluginInterface defines the minimal interface plugins need to implement
// All CGO complexity is handled by this wrapper
type SimplePluginInterface interface {
	// GetInfo returns plugin metadata
	GetInfo() PluginInfo
	
	// Initialize sets up the plugin with sample rate
	Initialize(sampleRate float64) error
	
	// ProcessAudio processes audio using Go-native types
	ProcessAudio(input, output [][]float32, frameCount uint32) error
	
	// GetParameters returns all parameter definitions
	GetParameters() []ParamInfo
	
	// SetParameterValue sets a parameter value
	SetParameterValue(paramID uint32, value float64) error
	
	// GetParameterValue gets a parameter value
	GetParameterValue(paramID uint32) float64
	
	// OnActivate is called when plugin is activated
	OnActivate() error
	
	// OnDeactivate is called when plugin is deactivated
	OnDeactivate()
	
	// Cleanup releases any resources
	Cleanup()
}

// PluginRegistry manages registered plugins
type PluginRegistry struct {
	plugins map[string]SimplePluginInterface
}

var globalRegistry = &PluginRegistry{
	plugins: make(map[string]SimplePluginInterface),
}

// RegisterSimplePlugin registers a plugin with the global registry
// This is the ONLY function plugin developers need to call
func RegisterSimplePlugin(plugin SimplePluginInterface) {
	info := plugin.GetInfo()
	globalRegistry.plugins[info.ID] = plugin
	
	// Set up parameter management
	if params := plugin.GetParameters(); len(params) > 0 {
		manager := NewParameterManager()
		for _, param := range params {
			manager.RegisterParameter(param)
		}
		
		// Store parameter manager (would need plugin wrapper enhancement)
	}
}

// CGO Export Functions - completely hidden from plugin developers

//export ClapGo_CreatePlugin
func ClapGo_CreatePlugin(host unsafe.Pointer, pluginID *C.char) unsafe.Pointer {
	id := C.GoString(pluginID)
	
	if plugin, exists := globalRegistry.plugins[id]; exists {
		// Create a full plugin wrapper
		wrapper := &FullPluginWrapper{
			simple: plugin,
			params: NewParameterManager(),
			state: &PluginState{},
		}
		
		// Initialize parameters
		for _, param := range plugin.GetParameters() {
			wrapper.params.RegisterParameter(param)
		}
		
		handle := cgo.NewHandle(wrapper)
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
	
	if _, exists := globalRegistry.plugins[id]; exists {
		return C.CString(id)
	}
	
	return C.CString("unknown")
}

//export ClapGo_GetPluginName
func ClapGo_GetPluginName(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	
	if plugin, exists := globalRegistry.plugins[id]; exists {
		info := plugin.GetInfo()
		return C.CString(info.Name)
	}
	
	return C.CString("Unknown Plugin")
}

//export ClapGo_GetPluginVendor
func ClapGo_GetPluginVendor(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	
	if plugin, exists := globalRegistry.plugins[id]; exists {
		info := plugin.GetInfo()
		return C.CString(info.Vendor)
	}
	
	return C.CString("Unknown Vendor")
}

//export ClapGo_GetPluginVersion
func ClapGo_GetPluginVersion(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	
	if plugin, exists := globalRegistry.plugins[id]; exists {
		info := plugin.GetInfo()
		return C.CString(info.Version)
	}
	
	return C.CString("0.0.0")
}

//export ClapGo_GetPluginDescription
func ClapGo_GetPluginDescription(pluginID *C.char) *C.char {
	id := C.GoString(pluginID)
	
	if plugin, exists := globalRegistry.plugins[id]; exists {
		info := plugin.GetInfo()
		return C.CString(info.Description)
	}
	
	return C.CString("No description available")
}

//export ClapGo_PluginInit
func ClapGo_PluginInit(plugin unsafe.Pointer) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	
	handle := cgo.Handle(plugin)
	wrapper := handle.Value().(*FullPluginWrapper)
	
	err := wrapper.simple.Initialize(44100.0) // Default sample rate
	return C.bool(err == nil)
}

//export ClapGo_PluginDestroy
func ClapGo_PluginDestroy(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	
	handle := cgo.Handle(plugin)
	wrapper := handle.Value().(*FullPluginWrapper)
	wrapper.simple.Cleanup()
	handle.Delete()
}

//export ClapGo_PluginActivate
func ClapGo_PluginActivate(plugin unsafe.Pointer, sampleRate C.double, minFrames, maxFrames C.uint32_t) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	
	handle := cgo.Handle(plugin)
	wrapper := handle.Value().(*FullPluginWrapper)
	
	// Initialize with sample rate
	err := wrapper.simple.Initialize(float64(sampleRate))
	if err != nil {
		return C.bool(false)
	}
	
	// Activate plugin
	err = wrapper.simple.OnActivate()
	if err == nil {
		wrapper.state.isActivated = true
		wrapper.state.sampleRate = float64(sampleRate)
		return C.bool(true)
	}
	
	return C.bool(false)
}

//export ClapGo_PluginDeactivate
func ClapGo_PluginDeactivate(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	
	handle := cgo.Handle(plugin)
	wrapper := handle.Value().(*FullPluginWrapper)
	wrapper.simple.OnDeactivate()
	wrapper.state.isActivated = false
	wrapper.state.isProcessing = false
}

//export ClapGo_PluginStartProcessing
func ClapGo_PluginStartProcessing(plugin unsafe.Pointer) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	
	handle := cgo.Handle(plugin)
	wrapper := handle.Value().(*FullPluginWrapper)
	
	if !wrapper.state.isActivated {
		return C.bool(false)
	}
	
	wrapper.state.isProcessing = true
	return C.bool(true)
}

//export ClapGo_PluginStopProcessing
func ClapGo_PluginStopProcessing(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	
	handle := cgo.Handle(plugin)
	wrapper := handle.Value().(*FullPluginWrapper)
	wrapper.state.isProcessing = false
}

//export ClapGo_PluginReset
func ClapGo_PluginReset(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	
	handle := cgo.Handle(plugin)
	wrapper := handle.Value().(*FullPluginWrapper)
	wrapper.params.ResetToDefaults()
}

//export ClapGo_PluginProcess
func ClapGo_PluginProcess(plugin unsafe.Pointer, process unsafe.Pointer) C.int32_t {
	if plugin == nil || process == nil {
		return C.int32_t(ProcessError)
	}
	
	handle := cgo.Handle(plugin)
	wrapper := handle.Value().(*FullPluginWrapper)
	
	if !wrapper.state.isActivated || !wrapper.state.isProcessing {
		return C.int32_t(ProcessError)
	}
	
	// Convert the C clap_process_t to Go parameters
	cProcess := (*C.clap_process_t)(process)
	
	// Extract frame count
	framesCount := uint32(cProcess.frames_count)
	
	// Convert audio buffers using our abstraction
	audioIn := ConvertFromCBuffers(unsafe.Pointer(cProcess.audio_inputs), uint32(cProcess.audio_inputs_count), framesCount)
	audioOut := ConvertFromCBuffers(unsafe.Pointer(cProcess.audio_outputs), uint32(cProcess.audio_outputs_count), framesCount)
	
	// Create event handler and process parameter events
	eventHandler := NewEventProcessor(
		unsafe.Pointer(cProcess.in_events),
		unsafe.Pointer(cProcess.out_events),
	)
	
	// Process parameter changes
	eventCount := eventHandler.GetInputEventCount()
	for i := uint32(0); i < eventCount; i++ {
		event := eventHandler.GetInputEvent(i)
		if event != nil && event.Type == EventTypeParamValue {
			if paramEvent, ok := event.Data.(ParamEvent); ok {
				wrapper.params.SetParameterValue(paramEvent.ParamID, paramEvent.Value)
				wrapper.simple.SetParameterValue(paramEvent.ParamID, paramEvent.Value)
			}
		}
	}
	
	// Call the simple plugin's process method
	err := wrapper.simple.ProcessAudio(audioIn, audioOut, framesCount)
	if err != nil {
		return C.int32_t(ProcessError)
	}
	
	return C.int32_t(ProcessContinue)
}

//export ClapGo_PluginGetExtension
func ClapGo_PluginGetExtension(plugin unsafe.Pointer, id *C.char) unsafe.Pointer {
	// Extensions are handled internally
	return nil
}

//export ClapGo_PluginOnMainThread
func ClapGo_PluginOnMainThread(plugin unsafe.Pointer) {
	// Nothing to do for simple plugins
}

// FullPluginWrapper wraps a SimplePluginInterface to provide full CLAP functionality
type FullPluginWrapper struct {
	simple SimplePluginInterface
	params *ParameterManager
	state  *PluginState
}

// PluginState tracks plugin state
type PluginState struct {
	sampleRate   float64
	isActivated  bool
	isProcessing bool
}

// GetPluginInfo returns plugin information
func (w *FullPluginWrapper) GetPluginInfo() PluginInfo {
	return w.simple.GetInfo()
}

// GetPluginID returns plugin ID
func (w *FullPluginWrapper) GetPluginID() string {
	return w.simple.GetInfo().ID
}