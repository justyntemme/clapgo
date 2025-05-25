package main

// This example shows a complete gain plugin with Audio Ports Activation support.
// It demonstrates a plugin with main stereo I/O plus an optional sidechain input.

// #cgo CFLAGS: -I../include/clap/include
// #include "../include/clap/include/clap/clap.h"
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"github.com/justyntemme/clapgo/pkg/api"
	"runtime/cgo"
	"unsafe"
)

const (
	PluginID = "com.example.gain-with-activation"
)

type GainPluginWithActivation struct {
	api.BasePlugin
	
	// Parameters
	gain float64
	
	// Port activation tracking
	portStates *api.AudioPortActivationState
	
	// Processing state
	sampleRate float64
	
	// Host reference
	host   unsafe.Pointer
	logger *api.HostLogger
}

func NewGainPluginWithActivation() *GainPluginWithActivation {
	plugin := &GainPluginWithActivation{
		gain:       1.0,
		portStates: api.NewAudioPortActivationState(),
	}
	
	// Set plugin info
	plugin.SetPluginInfo(api.PluginInfo{
		ID:          PluginID,
		Name:        "Gain with Port Activation",
		Vendor:      "Example",
		Version:     "1.0.0",
		Description: "Gain plugin demonstrating audio ports activation",
		Features:    []string{"audio-effect", "stereo"},
	})
	
	return plugin
}

// AudioPortsActivationProvider implementation
func (p *GainPluginWithActivation) CanActivateWhileProcessing() bool {
	// We don't support changing port activation during processing
	return false
}

func (p *GainPluginWithActivation) SetActive(isInput bool, portIndex uint32, isActive bool, sampleSize uint32) bool {
	if p.logger != nil {
		p.logger.Info(fmt.Sprintf("SetActive: input=%v, port=%d, active=%v, sampleSize=%d", 
			isInput, portIndex, isActive, sampleSize))
	}
	
	// Validate port indices
	// We have 2 input ports (0: main, 1: sidechain) and 1 output port
	if isInput && portIndex > 1 {
		return false
	}
	if !isInput && portIndex > 0 {
		return false
	}
	
	// Update activation state
	p.portStates.SetPortActive(isInput, portIndex, isActive)
	
	// Log the change
	portName := "unknown"
	if isInput {
		if portIndex == 0 {
			portName = "main input"
		} else {
			portName = "sidechain input"
		}
	} else {
		portName = "main output"
	}
	
	if p.logger != nil {
		if isActive {
			p.logger.Info(fmt.Sprintf("Activated %s", portName))
		} else {
			p.logger.Info(fmt.Sprintf("Deactivated %s", portName))
		}
	}
	
	return true
}

// AudioPortsProvider implementation
func (p *GainPluginWithActivation) GetAudioPortCount(isInput bool) uint32 {
	if isInput {
		return 2 // Main + Sidechain
	}
	return 1 // Main output only
}

func (p *GainPluginWithActivation) GetAudioPortInfo(index uint32, isInput bool) api.AudioPortInfo {
	if isInput {
		switch index {
		case 0:
			return api.AudioPortInfo{
				ID:           0,
				Name:         "Main Input",
				ChannelCount: 2,
				Flags:        api.AudioPortIsMain,
				PortType:     api.PortStereo,
				InPlacePair:  0,
			}
		case 1:
			return api.AudioPortInfo{
				ID:           1,
				Name:         "Sidechain",
				ChannelCount: 2,
				Flags:        api.AudioPortIsSidechain,
				PortType:     api.PortStereo,
				InPlacePair:  api.InvalidID,
			}
		}
	} else if index == 0 {
		return api.AudioPortInfo{
			ID:           0,
			Name:         "Main Output",
			ChannelCount: 2,
			Flags:        api.AudioPortIsMain,
			PortType:     api.PortStereo,
			InPlacePair:  0,
		}
	}
	
	return api.AudioPortInfo{}
}

// Plugin lifecycle
func (p *GainPluginWithActivation) Init() bool {
	if p.logger != nil {
		p.logger.Info("Plugin initialized")
	}
	return true
}

func (p *GainPluginWithActivation) Activate(sampleRate float64, minFrames, maxFrames uint32) bool {
	p.sampleRate = sampleRate
	if p.logger != nil {
		p.logger.Info(fmt.Sprintf("Activated: %.0f Hz", sampleRate))
	}
	return true
}

func (p *GainPluginWithActivation) Process(process *api.Process) api.ProcessStatus {
	// Check if we have the expected number of ports
	if len(process.AudioInputs) < 1 || len(process.AudioOutputs) < 1 {
		return api.ProcessError
	}
	
	mainIn := process.AudioInputs[0]
	mainOut := process.AudioOutputs[0]
	
	// Check if sidechain is active and available
	haveSidechain := len(process.AudioInputs) > 1 && p.portStates.IsPortActive(true, 1)
	
	// Process based on active ports
	if haveSidechain {
		// Process with sidechain (e.g., ducking)
		sidechain := process.AudioInputs[1]
		p.processWithSidechain(mainIn, sidechain, mainOut)
	} else {
		// Simple gain processing
		p.processSimpleGain(mainIn, mainOut)
	}
	
	return api.ProcessContinue
}

func (p *GainPluginWithActivation) processSimpleGain(input, output api.AudioBuffer) {
	// Simple gain processing
	for ch := 0; ch < len(input.Data32); ch++ {
		for i := 0; i < len(input.Data32[ch]); i++ {
			output.Data32[ch][i] = input.Data32[ch][i] * float32(p.gain)
		}
	}
}

func (p *GainPluginWithActivation) processWithSidechain(main, sidechain, output api.AudioBuffer) {
	// Example: simple ducking based on sidechain level
	for ch := 0; ch < len(main.Data32); ch++ {
		for i := 0; i < len(main.Data32[ch]); i++ {
			// Get sidechain level (simplified)
			sidechainLevel := abs(sidechain.Data32[ch%len(sidechain.Data32)][i])
			
			// Duck the main signal based on sidechain
			duckFactor := 1.0 - (sidechainLevel * 0.5) // Simple ducking
			if duckFactor < 0.1 {
				duckFactor = 0.1
			}
			
			output.Data32[ch][i] = main.Data32[ch][i] * float32(p.gain) * float32(duckFactor)
		}
	}
}

func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

// Export functions for C bridge

var gainPlugin *GainPluginWithActivation

func init() {
	gainPlugin = NewGainPluginWithActivation()
}

//export ClapGo_CreatePlugin
func ClapGo_CreatePlugin(host unsafe.Pointer, pluginID *C.char) unsafe.Pointer {
	id := C.GoString(pluginID)
	if id == PluginID {
		gainPlugin.host = host
		gainPlugin.logger = api.NewHostLogger(host)
		handle := cgo.NewHandle(gainPlugin)
		return unsafe.Pointer(handle)
	}
	return nil
}

//export ClapGo_PluginInit
func ClapGo_PluginInit(plugin unsafe.Pointer) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*GainPluginWithActivation)
	return C.bool(p.Init())
}

//export ClapGo_PluginActivate
func ClapGo_PluginActivate(plugin unsafe.Pointer, sampleRate C.double, minFrames, maxFrames C.uint32_t) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*GainPluginWithActivation)
	return C.bool(p.Activate(float64(sampleRate), uint32(minFrames), uint32(maxFrames)))
}

//export ClapGo_PluginProcess
func ClapGo_PluginProcess(plugin unsafe.Pointer, process unsafe.Pointer) C.int32_t {
	if plugin == nil || process == nil {
		return C.int32_t(api.ProcessError)
	}
	
	p := cgo.Handle(plugin).Value().(*GainPluginWithActivation)
	cProcess := (*C.clap_process_t)(process)
	
	// Convert buffers
	audioIn := api.ConvertFromCBuffers(unsafe.Pointer(cProcess.audio_inputs), 
		uint32(cProcess.audio_inputs_count), uint32(cProcess.frames_count))
	audioOut := api.ConvertFromCBuffers(unsafe.Pointer(cProcess.audio_outputs), 
		uint32(cProcess.audio_outputs_count), uint32(cProcess.frames_count))
	
	// Create process context
	proc := &api.Process{
		AudioInputs:  audioIn,
		AudioOutputs: audioOut,
		FrameCount:   uint32(cProcess.frames_count),
		SteadyTime:   int64(cProcess.steady_time),
	}
	
	return C.int32_t(p.Process(proc))
}

//export ClapGo_PluginAudioPortsActivationCanActivateWhileProcessing
func ClapGo_PluginAudioPortsActivationCanActivateWhileProcessing(plugin unsafe.Pointer) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*GainPluginWithActivation)
	return C.bool(p.CanActivateWhileProcessing())
}

//export ClapGo_PluginAudioPortsActivationSetActive
func ClapGo_PluginAudioPortsActivationSetActive(plugin unsafe.Pointer, isInput C.bool, portIndex C.uint32_t, isActive C.bool, sampleSize C.uint32_t) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*GainPluginWithActivation)
	return C.bool(p.SetActive(bool(isInput), uint32(portIndex), bool(isActive), uint32(sampleSize)))
}

// Additional required exports would go here...

func main() {
	// Required for c-shared build
}