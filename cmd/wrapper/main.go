package main

/*
#include <stdint.h>
#include <stdbool.h>
#include <stdlib.h>
#include "../../include/clap/include/clap/clap.h"

// Forward declarations of the functions we'll implement in Go
uint32_t GetPluginCount();
struct clap_plugin_descriptor *GetPluginInfo(uint32_t index);
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
void *GoGetExtension(void *plugin, char *id);
void GoOnMainThread(void *plugin);
*/
import "C"

import (
	"fmt"
	"runtime/cgo"
	"unsafe"

	"github.com/justyntemme/clapgo/src/goclap"
)

//export GetPluginCount
func GetPluginCount() C.uint32_t {
	return C.uint32_t(goclap.GetPluginCountImpl())
}

//export GetPluginInfo
func GetPluginInfo(index C.uint32_t) *C.struct_clap_plugin_descriptor {
	info := goclap.GetPluginInfoImpl(uint32(index))
	
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
	processor := goclap.CreatePluginImpl(id)
	
	if processor == nil {
		return nil
	}
	
	// Create a wrapper to hold the processor
	wrapper := &processorWrapper{
		processor: processor,
	}
	
	// Properly handle to Go object
	handle := cgo.NewHandle(wrapper)
	
	// Print some debugging info
	fmt.Printf("Creating plugin with ID: %s, handle: %v\n", id, handle)
	
	// Convert handle to pointer
	pluginPtr := unsafe.Pointer(uintptr(handle))
	
	return pluginPtr
}

// processorWrapper is a wrapper around AudioProcessor to prevent garbage collection
type processorWrapper struct {
	processor goclap.AudioProcessor
}

//export GetVersion
func GetVersion(major, minor, patch *C.uint32_t) C.bool {
	majorV, minorV, patchV := goclap.GetVersionImpl()
	if major != nil {
		*major = C.uint32_t(majorV)
	}
	if minor != nil {
		*minor = C.uint32_t(minorV)
	}
	if patch != nil {
		*patch = C.uint32_t(patchV)
	}
	return C.bool(true)
}

func init() {
	// Register our plugins when the library loads
	InitGainPlugin()
}

//export GoInit
func GoInit(plugin unsafe.Pointer) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	
	// Convert the plugin pointer back to a handle
	handle := cgo.Handle(uintptr(plugin))
	
	// Get the wrapper from the handle
	wrapper, ok := handle.Value().(*processorWrapper)
	if !ok {
		fmt.Printf("Error: GoInit could not convert handle to processorWrapper\n")
		return C.bool(false)
	}
	
	return C.bool(goclap.InitImpl(wrapper.processor))
}

//export GoDestroy
func GoDestroy(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	
	// Convert the plugin pointer back to a handle
	handle := cgo.Handle(uintptr(plugin))
	
	// Get the wrapper from the handle
	wrapper, ok := handle.Value().(*processorWrapper)
	if !ok {
		fmt.Printf("Error: GoDestroy could not convert handle to processorWrapper\n")
		return
	}
	
	goclap.DestroyImpl(wrapper.processor)
	
	// Free the handle to prevent memory leaks
	handle.Delete()
}

//export GoActivate
func GoActivate(plugin unsafe.Pointer, sampleRate C.double, minFrames, maxFrames C.uint32_t) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	
	// Convert the plugin pointer back to a handle
	handle := cgo.Handle(uintptr(plugin))
	
	// Get the wrapper from the handle
	wrapper, ok := handle.Value().(*processorWrapper)
	if !ok {
		fmt.Printf("Error: GoActivate could not convert handle to processorWrapper\n")
		return C.bool(false)
	}
	
	return C.bool(goclap.ActivateImpl(wrapper.processor, float64(sampleRate), uint32(minFrames), uint32(maxFrames)))
}

//export GoDeactivate
func GoDeactivate(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	
	// Convert the plugin pointer back to a handle
	handle := cgo.Handle(uintptr(plugin))
	
	// Get the wrapper from the handle
	wrapper, ok := handle.Value().(*processorWrapper)
	if !ok {
		fmt.Printf("Error: GoDeactivate could not convert handle to processorWrapper\n")
		return
	}
	
	goclap.DeactivateImpl(wrapper.processor)
}

//export GoStartProcessing
func GoStartProcessing(plugin unsafe.Pointer) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	
	// Convert the plugin pointer back to a handle
	handle := cgo.Handle(uintptr(plugin))
	
	// Get the wrapper from the handle
	wrapper, ok := handle.Value().(*processorWrapper)
	if !ok {
		fmt.Printf("Error: GoStartProcessing could not convert handle to processorWrapper\n")
		return C.bool(false)
	}
	
	return C.bool(goclap.StartProcessingImpl(wrapper.processor))
}

//export GoStopProcessing
func GoStopProcessing(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	
	// Convert the plugin pointer back to a handle
	handle := cgo.Handle(uintptr(plugin))
	
	// Get the wrapper from the handle
	wrapper, ok := handle.Value().(*processorWrapper)
	if !ok {
		fmt.Printf("Error: GoStopProcessing could not convert handle to processorWrapper\n")
		return
	}
	
	goclap.StopProcessingImpl(wrapper.processor)
}

//export GoReset
func GoReset(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	
	// Convert the plugin pointer back to a handle
	handle := cgo.Handle(uintptr(plugin))
	
	// Get the wrapper from the handle
	wrapper, ok := handle.Value().(*processorWrapper)
	if !ok {
		fmt.Printf("Error: GoReset could not convert handle to processorWrapper\n")
		return
	}
	
	goclap.ResetImpl(wrapper.processor)
}

//export GoProcess
func GoProcess(plugin unsafe.Pointer, process *C.struct_clap_process) C.int32_t {
	if plugin == nil {
		return C.int32_t(0) // CLAP_PROCESS_ERROR
	}
	
	// Convert the plugin pointer back to a handle
	handle := cgo.Handle(uintptr(plugin))
	
	// Get the wrapper from the handle
	wrapper, ok := handle.Value().(*processorWrapper)
	if !ok {
		fmt.Printf("Error: GoProcess could not convert handle to processorWrapper\n")
		return C.int32_t(0) // CLAP_PROCESS_ERROR
	}
	
	// Get the processor from the wrapper
	processor := wrapper.processor

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
	
	events := &goclap.ProcessEvents{
		InEvents:  unsafe.Pointer(process.in_events),
		OutEvents: unsafe.Pointer(process.out_events),
	}
	
	status := goclap.ProcessImpl(processor, steadyTime, framesCount, audioIn, audioOut, events)
	return C.int32_t(status)
}

//export GoGetExtension
func GoGetExtension(plugin unsafe.Pointer, id *C.char) unsafe.Pointer {
	if plugin == nil {
		return nil
	}
	
	// Convert the plugin pointer back to a handle
	handle := cgo.Handle(uintptr(plugin))
	
	// Get the wrapper from the handle
	wrapper, ok := handle.Value().(*processorWrapper)
	if !ok {
		fmt.Printf("Error: GoGetExtension could not convert handle to processorWrapper\n")
		return nil
	}
	
	extID := C.GoString(id)
	return goclap.GetExtensionImpl(wrapper.processor, extID)
}

//export GoOnMainThread
func GoOnMainThread(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	
	// Convert the plugin pointer back to a handle
	handle := cgo.Handle(uintptr(plugin))
	
	// Get the wrapper from the handle
	wrapper, ok := handle.Value().(*processorWrapper)
	if !ok {
		fmt.Printf("Error: GoOnMainThread could not convert handle to processorWrapper\n")
		return
	}
	
	goclap.OnMainThreadImpl(wrapper.processor)
}

// InitGainPlugin creates and registers the gain plugin
func InitGainPlugin() {
	// Register a simple gain plugin directly
	info := goclap.PluginInfo{
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
	
	// Create a basic gain plugin implementation
	gain := &GainPlugin{
		gain:       1.0,
		sampleRate: 44100.0,
	}
	
	// Register the plugin
	goclap.RegisterPlugin(info, gain)
}

// Simple gain plugin implementation
type GainPlugin struct {
	gain         float64
	sampleRate   float64
	isActivated  bool
	isProcessing bool
	audioPorts   *goclap.AudioPortsExtension
	paramManager *goclap.ParamManager
}

func (p *GainPlugin) Init() bool {
	// Create audio ports extension on initialization
	p.audioPorts = goclap.CreateAudioPortsExtension(p)
	
	// Initialize parameter manager
	p.paramManager = goclap.NewParamManager()
	
	return true
}

// GetParamManager returns the parameter manager
func (p *GainPlugin) GetParamManager() *goclap.ParamManager {
	return p.paramManager
}

func (p *GainPlugin) Destroy() {}

func (p *GainPlugin) Activate(sampleRate float64, minFrames, maxFrames uint32) bool {
	p.sampleRate = sampleRate
	p.isActivated = true
	return true
}

func (p *GainPlugin) Deactivate() { 
	p.isActivated = false 
}

func (p *GainPlugin) StartProcessing() bool {
	if !p.isActivated { return false }
	p.isProcessing = true
	return true
}

func (p *GainPlugin) StopProcessing() { 
	p.isProcessing = false 
}

func (p *GainPlugin) Reset() { 
	p.gain = 1.0 
}

func (p *GainPlugin) GetExtension(id string) unsafe.Pointer {
	// Support the audio-ports extension
	if id == "clap.audio-ports" {
		return unsafe.Pointer(p.audioPorts)
	}
	return nil 
}

func (p *GainPlugin) OnMainThread() {}

// Process applies gain to the audio
func (p *GainPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events *goclap.ProcessEvents) int {
	// Absolute minimum implementation to avoid crashes
	// Ignores audio processing for now
	
	// Just return success
	return 1 // CLAP_PROCESS_CONTINUE
}

func main() {
	// This function is required for building a shared library
	// but is not used when the library is loaded
	InitGainPlugin()
}

