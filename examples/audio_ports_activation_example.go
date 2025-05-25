package main

// This example demonstrates how to implement the Audio Ports Activation extension
// in a ClapGo plugin. This would typically be added to an existing plugin that
// has multiple audio ports.

import "C"
import (
	"github.com/justyntemme/clapgo/pkg/api"
	"unsafe"
)

// Example plugin with audio ports activation support
type PluginWithPortActivation struct {
	api.BasePlugin
	
	// Track port activation states
	portStates *api.AudioPortActivationState
	
	// Other plugin fields...
}

// Implement the AudioPortsActivationProvider interface
func (p *PluginWithPortActivation) CanActivateWhileProcessing() bool {
	// Most plugins should return false for safety
	// Only return true if you have proper thread-safe state management
	return false
}

func (p *PluginWithPortActivation) SetActive(isInput bool, portIndex uint32, isActive bool, sampleSize uint32) bool {
	// Validate port index
	if isInput {
		// Check against input port count
		if portIndex >= 2 { // Example: 2 input ports
			return false
		}
	} else {
		// Check against output port count  
		if portIndex >= 3 { // Example: 3 output ports
			return false
		}
	}
	
	// Update the activation state
	p.portStates.SetPortActive(isInput, portIndex, isActive)
	
	// Handle sample size hint if needed
	// sampleSize can be 32, 64, or 0 (unspecified)
	
	// Optional: Optimize internal processing based on active ports
	if !isInput && portIndex == 2 { // Example: aux output
		// Enable/disable aux processing
		// p.auxProcessor.SetEnabled(isActive)
	}
	
	return true
}

// Export functions for the C bridge
// These would be added to your plugin's main.go file

//export ClapGo_PluginAudioPortsActivationCanActivateWhileProcessing
func ClapGo_PluginAudioPortsActivationCanActivateWhileProcessing(plugin unsafe.Pointer) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	
	// Get the plugin instance
	// In a real implementation, you'd use cgo.Handle to get your plugin
	// p := cgo.Handle(plugin).Value().(YourPluginType)
	
	// Check if plugin implements the interface
	// if provider, ok := p.(api.AudioPortsActivationProvider); ok {
	//     return C.bool(provider.CanActivateWhileProcessing())
	// }
	
	return C.bool(false)
}

//export ClapGo_PluginAudioPortsActivationSetActive
func ClapGo_PluginAudioPortsActivationSetActive(plugin unsafe.Pointer, isInput C.bool, portIndex C.uint32_t, isActive C.bool, sampleSize C.uint32_t) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	
	// Get the plugin instance
	// In a real implementation, you'd use cgo.Handle to get your plugin
	// p := cgo.Handle(plugin).Value().(YourPluginType)
	
	// Check if plugin implements the interface
	// if provider, ok := p.(api.AudioPortsActivationProvider); ok {
	//     return C.bool(provider.SetActive(bool(isInput), uint32(portIndex), bool(isActive), uint32(sampleSize)))
	// }
	
	return C.bool(false)
}

// Example usage in Process method
func (p *PluginWithPortActivation) Process(process *api.Process) api.ProcessStatus {
	// Check port activation states to optimize processing
	
	// Example: Skip sidechain processing if input port 1 is inactive
	if !p.portStates.IsPortActive(true, 1) {
		// Process without sidechain
		p.processMainOnly(process)
	} else {
		// Process with sidechain
		p.processWithSidechain(process)
	}
	
	// Example: Skip aux send if output port 2 is inactive
	if p.portStates.IsPortActive(false, 2) {
		p.processAuxSend(process.AudioOutputs[2])
	}
	
	return api.ProcessContinue
}

// Placeholder methods for the example
func (p *PluginWithPortActivation) processMainOnly(process *api.Process) {}
func (p *PluginWithPortActivation) processWithSidechain(process *api.Process) {}
func (p *PluginWithPortActivation) processAuxSend(buffer api.AudioBuffer) {}