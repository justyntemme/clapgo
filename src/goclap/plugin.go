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
	
	// Get the parameter manager for this plugin
	GetParamManager() *ParamManager
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

// GetPluginCountImpl returns the number of registered plugins
func GetPluginCountImpl() uint32 {
	pluginInfoRegistry.RLock()
	defer pluginInfoRegistry.RUnlock()
	
	return uint32(len(pluginInfoRegistry.infos))
}

// GetPluginInfoImpl returns plugin info at the given index
func GetPluginInfoImpl(index uint32) PluginInfo {
	pluginInfoRegistry.RLock()
	defer pluginInfoRegistry.RUnlock()
	
	idx := int(index)
	if idx < 0 || idx >= len(pluginInfoRegistry.infos) {
		return PluginInfo{}
	}
	
	// Convert infos map to a slice and get the item at the index
	var plugins []PluginInfo
	for _, info := range pluginInfoRegistry.infos {
		plugins = append(plugins, info)
	}
	
	if idx < len(plugins) {
		return plugins[idx]
	}
	
	return PluginInfo{}
}

// CreatePluginImpl creates a new plugin instance
func CreatePluginImpl(pluginID string) AudioProcessor {
	pluginRegistry.RLock()
	processor, exists := pluginRegistry.plugins[pluginID]
	pluginRegistry.RUnlock()
	
	if !exists {
		return nil
	}
	
	return processor
}

// Basic plugin instance callbacks

// InitImpl initializes the plugin
func InitImpl(processor AudioProcessor) bool {
	if processor == nil {
		return false
	}
	
	return processor.Init()
}

// DestroyImpl destroys the plugin
func DestroyImpl(processor AudioProcessor) {
	if processor == nil {
		return
	}
	
	processor.Destroy()
}

// ActivateImpl activates the plugin
func ActivateImpl(processor AudioProcessor, sampleRate float64, minFrames, maxFrames uint32) bool {
	if processor == nil {
		return false
	}
	
	return processor.Activate(sampleRate, minFrames, maxFrames)
}

// DeactivateImpl deactivates the plugin
func DeactivateImpl(processor AudioProcessor) {
	if processor == nil {
		return
	}
	
	processor.Deactivate()
}

// StartProcessingImpl starts processing
func StartProcessingImpl(processor AudioProcessor) bool {
	if processor == nil {
		return false
	}
	
	return processor.StartProcessing()
}

// StopProcessingImpl stops processing
func StopProcessingImpl(processor AudioProcessor) {
	if processor == nil {
		return
	}
	
	processor.StopProcessing()
}

// ResetImpl resets the plugin
func ResetImpl(processor AudioProcessor) {
	if processor == nil {
		return
	}
	
	processor.Reset()
}

// ProcessImpl processes audio
func ProcessImpl(processor AudioProcessor, steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events *ProcessEvents) int {
	if processor == nil {
		return 0 // CLAP_PROCESS_ERROR
	}
	
	status := processor.Process(steadyTime, framesCount, audioIn, audioOut, events)
	return status
}

// GetExtensionImpl gets an extension
func GetExtensionImpl(processor AudioProcessor, id string) unsafe.Pointer {
	if processor == nil {
		return nil
	}
	
	return processor.GetExtension(id)
}

// OnMainThreadImpl runs on main thread
func OnMainThreadImpl(processor AudioProcessor) {
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