package goclap

// #include <stdlib.h>
// #include "../../include/clap/include/clap/clap.h"
import "C"
import (
	"sync"
	"unsafe"
)

// PluginInfo holds metadata about a CLAP plugin.
type PluginInfo struct {
	ID          string
	Name        string
	Vendor      string
	URL         string
	ManualURL   string
	SupportURL  string
	Version     string
	Description string
	Features    []string
}

// AudioProcessor interface that must be implemented by plugins
type AudioProcessor interface {
	// Initialize the plugin
	Init() bool
	
	// Clean up plugin resources
	Destroy()
	
	// Activate the plugin with given sample rate and buffer sizes
	Activate(sampleRate float64, minFrames, maxFrames uint32) bool
	
	// Deactivate the plugin
	Deactivate()
	
	// Start audio processing
	StartProcessing() bool
	
	// Stop audio processing
	StopProcessing()
	
	// Reset the plugin state
	Reset()
	
	// Process audio data
	// Returns a status code: 0-error, 1-continue, 2-continue_if_not_quiet, 3-tail, 4-sleep
	Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events *ProcessEvents) int
	
	// Get an extension
	GetExtension(id string) unsafe.Pointer
	
	// Called on the main thread
	OnMainThread()
}

// ProcessEvents wraps CLAP event access
type ProcessEvents struct {
	InEvents  unsafe.Pointer
	OutEvents unsafe.Pointer
}

// pluginRegistry maps plugin IDs to Go plugin instances
var pluginRegistry = struct {
	sync.RWMutex
	plugins map[string]AudioProcessor
}{
	plugins: make(map[string]AudioProcessor),
}

// pluginInfoRegistry holds plugin descriptors
var pluginInfoRegistry = struct {
	sync.RWMutex
	infos map[string]PluginInfo
}{
	infos: make(map[string]PluginInfo),
}

// RegisterPlugin adds a plugin to the registry
func RegisterPlugin(info PluginInfo, processor AudioProcessor) {
	pluginRegistry.Lock()
	defer pluginRegistry.Unlock()
	
	pluginInfoRegistry.Lock()
	defer pluginInfoRegistry.Unlock()
	
	pluginRegistry.plugins[info.ID] = processor
	pluginInfoRegistry.infos[info.ID] = info
}

// GetPluginCount returns the number of registered plugins
//export GetPluginCount
func GetPluginCount() C.uint32_t {
	pluginInfoRegistry.RLock()
	defer pluginInfoRegistry.RUnlock()
	
	return C.uint32_t(len(pluginInfoRegistry.infos))
}

// GetPluginInfo returns plugin info at the given index
//export GetPluginInfo
func GetPluginInfo(index C.uint32_t) *C.clap_plugin_descriptor_t {
	pluginInfoRegistry.RLock()
	defer pluginInfoRegistry.RUnlock()
	
	idx := int(index)
	if idx < 0 || idx >= len(pluginInfoRegistry.infos) {
		return nil
	}
	
	// This would be implemented by the C wrapper
	// that would convert our Go structures to C structures
	// Here we return nil as this is a stub
	return nil
}

// CreatePlugin creates a new plugin instance
//export CreatePlugin
func CreatePlugin(pluginID *C.char) unsafe.Pointer {
	id := C.GoString(pluginID)
	
	pluginRegistry.RLock()
	_, exists := pluginRegistry.plugins[id]
	pluginRegistry.RUnlock()
	
	if !exists {
		return nil
	}
	
	// This would be implemented by the C wrapper
	// that would associate the Go processor with a C plugin instance
	return nil
}

// Basic plugin instance callbacks

//export GoInit
func GoInit(pluginPtr unsafe.Pointer) C.bool {
	processor := getProcessorFromPtr(pluginPtr)
	if processor == nil {
		return C.bool(false)
	}
	
	return C.bool(processor.Init())
}

//export GoDestroy
func GoDestroy(pluginPtr unsafe.Pointer) {
	processor := getProcessorFromPtr(pluginPtr)
	if processor == nil {
		return
	}
	
	processor.Destroy()
}

//export GoActivate
func GoActivate(pluginPtr unsafe.Pointer, sampleRate C.double, minFrames, maxFrames C.uint32_t) C.bool {
	processor := getProcessorFromPtr(pluginPtr)
	if processor == nil {
		return C.bool(false)
	}
	
	return C.bool(processor.Activate(float64(sampleRate), uint32(minFrames), uint32(maxFrames)))
}

//export GoDeactivate
func GoDeactivate(pluginPtr unsafe.Pointer) {
	processor := getProcessorFromPtr(pluginPtr)
	if processor == nil {
		return
	}
	
	processor.Deactivate()
}

//export GoStartProcessing
func GoStartProcessing(pluginPtr unsafe.Pointer) C.bool {
	processor := getProcessorFromPtr(pluginPtr)
	if processor == nil {
		return C.bool(false)
	}
	
	return C.bool(processor.StartProcessing())
}

//export GoStopProcessing
func GoStopProcessing(pluginPtr unsafe.Pointer) {
	processor := getProcessorFromPtr(pluginPtr)
	if processor == nil {
		return
	}
	
	processor.StopProcessing()
}

//export GoReset
func GoReset(pluginPtr unsafe.Pointer) {
	processor := getProcessorFromPtr(pluginPtr)
	if processor == nil {
		return
	}
	
	processor.Reset()
}

//export GoProcess
func GoProcess(pluginPtr unsafe.Pointer, process *C.clap_process_t) C.clap_process_status {
	processor := getProcessorFromPtr(pluginPtr)
	if processor == nil {
		return C.CLAP_PROCESS_ERROR
	}
	
	steadyTime := int64(process.steady_time)
	framesCount := uint32(process.frames_count)
	
	// Create audio buffer slices
	// This is a simplified version - the real implementation would need to handle
	// both data32 and data64, and properly convert C arrays to Go slices
	var audioIn [][]float32
	var audioOut [][]float32
	
	// Convert events
	events := &ProcessEvents{
		InEvents:  unsafe.Pointer(process.in_events),
		OutEvents: unsafe.Pointer(process.out_events),
	}
	
	status := processor.Process(steadyTime, framesCount, audioIn, audioOut, events)
	return C.clap_process_status(status)
}

//export GoGetExtension
func GoGetExtension(pluginPtr unsafe.Pointer, id *C.char) unsafe.Pointer {
	processor := getProcessorFromPtr(pluginPtr)
	if processor == nil {
		return nil
	}
	
	extID := C.GoString(id)
	return processor.GetExtension(extID)
}

//export GoOnMainThread
func GoOnMainThread(pluginPtr unsafe.Pointer) {
	processor := getProcessorFromPtr(pluginPtr)
	if processor == nil {
		return
	}
	
	processor.OnMainThread()
}

// getProcessorFromPtr retrieves the Go processor from a C plugin pointer
// This would be implemented in the C wrapper to get the right AudioProcessor
// from the C plugin instance
func getProcessorFromPtr(pluginPtr unsafe.Pointer) AudioProcessor {
	// This is a stub - actual implementation would be in the C code
	return nil
}