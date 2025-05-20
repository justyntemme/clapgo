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
	
	// Properly handle to Go object
	handle := cgo.NewHandle(processor)
	// This handle is safe to pass to and from C
	pluginPtr := unsafe.Pointer(uintptr(handle))
	
	return pluginPtr
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
	
	// Get the processor from the handle
	processor, ok := handle.Value().(goclap.AudioProcessor)
	if !ok {
		return C.bool(false)
	}
	
	return C.bool(goclap.InitImpl(processor))
}

//export GoDestroy
func GoDestroy(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	
	// Convert the plugin pointer back to a handle
	handle := cgo.Handle(uintptr(plugin))
	
	// Get the processor from the handle
	processor, ok := handle.Value().(goclap.AudioProcessor)
	if !ok {
		return
	}
	
	goclap.DestroyImpl(processor)
	
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
	
	// Get the processor from the handle
	processor, ok := handle.Value().(goclap.AudioProcessor)
	if !ok {
		return C.bool(false)
	}
	
	return C.bool(goclap.ActivateImpl(processor, float64(sampleRate), uint32(minFrames), uint32(maxFrames)))
}

//export GoDeactivate
func GoDeactivate(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	
	// Convert the plugin pointer back to a handle
	handle := cgo.Handle(uintptr(plugin))
	
	// Get the processor from the handle
	processor, ok := handle.Value().(goclap.AudioProcessor)
	if !ok {
		return
	}
	
	goclap.DeactivateImpl(processor)
}

//export GoStartProcessing
func GoStartProcessing(plugin unsafe.Pointer) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	
	// Convert the plugin pointer back to a handle
	handle := cgo.Handle(uintptr(plugin))
	
	// Get the processor from the handle
	processor, ok := handle.Value().(goclap.AudioProcessor)
	if !ok {
		return C.bool(false)
	}
	
	return C.bool(goclap.StartProcessingImpl(processor))
}

//export GoStopProcessing
func GoStopProcessing(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	
	// Convert the plugin pointer back to a handle
	handle := cgo.Handle(uintptr(plugin))
	
	// Get the processor from the handle
	processor, ok := handle.Value().(goclap.AudioProcessor)
	if !ok {
		return
	}
	
	goclap.StopProcessingImpl(processor)
}

//export GoReset
func GoReset(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	
	// Convert the plugin pointer back to a handle
	handle := cgo.Handle(uintptr(plugin))
	
	// Get the processor from the handle
	processor, ok := handle.Value().(goclap.AudioProcessor)
	if !ok {
		return
	}
	
	goclap.ResetImpl(processor)
}

//export GoProcess
func GoProcess(plugin unsafe.Pointer, process *C.struct_clap_process) C.int32_t {
	if plugin == nil {
		return C.int32_t(0) // CLAP_PROCESS_ERROR
	}
	
	// Convert the plugin pointer back to a handle
	handle := cgo.Handle(uintptr(plugin))
	
	// Get the processor from the handle
	processor, ok := handle.Value().(goclap.AudioProcessor)
	if !ok {
		return C.int32_t(0) // CLAP_PROCESS_ERROR
	}

	steadyTime := int64(process.steady_time)
	framesCount := uint32(process.frames_count)
	
	// Convert audio buffers from C to Go
	var audioIn [][]float32
	var audioOut [][]float32
	
	// Create a simpler approach that bypasses the need to access clap_audio_buffer directly
	// We'll just pass empty slices for now as a proof of concept
	audioIn = make([][]float32, 0) 
	audioOut = make([][]float32, 0)
	
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
	
	// Get the processor from the handle
	processor, ok := handle.Value().(goclap.AudioProcessor)
	if !ok {
		return nil
	}
	
	extID := C.GoString(id)
	return goclap.GetExtensionImpl(processor, extID)
}

//export GoOnMainThread
func GoOnMainThread(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	
	// Convert the plugin pointer back to a handle
	handle := cgo.Handle(uintptr(plugin))
	
	// Get the processor from the handle
	processor, ok := handle.Value().(goclap.AudioProcessor)
	if !ok {
		return
	}
	
	goclap.OnMainThreadImpl(processor)
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
}

func (p *GainPlugin) Init() bool { return true }
func (p *GainPlugin) Destroy() {}
func (p *GainPlugin) Activate(sampleRate float64, minFrames, maxFrames uint32) bool {
	p.sampleRate = sampleRate
	p.isActivated = true
	return true
}
func (p *GainPlugin) Deactivate() { p.isActivated = false }
func (p *GainPlugin) StartProcessing() bool {
	if !p.isActivated { return false }
	p.isProcessing = true
	return true
}
func (p *GainPlugin) StopProcessing() { p.isProcessing = false }
func (p *GainPlugin) Reset() { p.gain = 1.0 }
func (p *GainPlugin) GetExtension(id string) unsafe.Pointer { return nil }
func (p *GainPlugin) OnMainThread() {}

// Process applies gain to the audio
func (p *GainPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events *goclap.ProcessEvents) int {
	// For our simple proof of concept, we'll just return CONTINUE
	return 1 // CLAP_PROCESS_CONTINUE
}

func main() {
	// This function is required for building a shared library
	// but is not used when the library is loaded
	InitGainPlugin()
}

