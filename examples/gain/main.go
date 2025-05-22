package main

import (
	"sync/atomic"
	"unsafe"
	
	"github.com/justyntemme/clapgo/pkg/api"
)

// GainPlugin implements the SimplePluginInterface
// NO CGO required - all complexity handled by pkg/api
type GainPlugin struct {
	gain int64 // atomic storage for gain value
}

// GetInfo returns plugin metadata
func (p *GainPlugin) GetInfo() api.PluginInfo {
	return api.PluginInfo{
		ID:          PluginID,
		Name:        PluginName,
		Vendor:      PluginVendor,
		URL:         "https://github.com/justyntemme/clapgo",
		ManualURL:   "https://github.com/justyntemme/clapgo",
		SupportURL:  "https://github.com/justyntemme/clapgo/issues",
		Version:     PluginVersion,
		Description: PluginDescription,
		Features:    []string{"audio-effect", "stereo", "mono"},
	}
}

// Initialize sets up the plugin with sample rate
func (p *GainPlugin) Initialize(sampleRate float64) error {
	// Set default gain to 1.0 (0dB)
	atomic.StoreInt64(&p.gain, int64(floatToBits(1.0)))
	return nil
}

// ProcessAudio processes audio using Go-native types
func (p *GainPlugin) ProcessAudio(input, output [][]float32, frameCount uint32) error {
	// Get current gain value atomically
	gainBits := atomic.LoadInt64(&p.gain)
	gain := floatFromBits(uint64(gainBits))
	
	// If no audio inputs or outputs, nothing to do
	if len(input) == 0 || len(output) == 0 {
		return nil
	}
	
	// Get the number of channels (use min of input and output)
	numChannels := len(input)
	if len(output) < numChannels {
		numChannels = len(output)
	}
	
	// Process audio - apply gain to each sample
	for ch := 0; ch < numChannels; ch++ {
		inChannel := input[ch]
		outChannel := output[ch]
		
		// Make sure we have enough buffer space
		if len(inChannel) < int(frameCount) || len(outChannel) < int(frameCount) {
			continue // Skip this channel if buffer is too small
		}
		
		// Apply gain to each sample
		for i := uint32(0); i < frameCount; i++ {
			outChannel[i] = inChannel[i] * float32(gain)
		}
	}
	
	return nil
}

// GetParameters returns all parameter definitions
func (p *GainPlugin) GetParameters() []api.ParamInfo {
	return []api.ParamInfo{
		api.CreateFloatParameter(0, "Gain", 0.0, 2.0, 1.0),
	}
}

// SetParameterValue sets a parameter value
func (p *GainPlugin) SetParameterValue(paramID uint32, value float64) error {
	if paramID == 0 {
		// Clamp value to valid range
		if value < 0.0 {
			value = 0.0
		}
		if value > 2.0 {
			value = 2.0
		}
		atomic.StoreInt64(&p.gain, int64(floatToBits(value)))
		return nil
	}
	return api.ErrInvalidParam
}

// GetParameterValue gets a parameter value
func (p *GainPlugin) GetParameterValue(paramID uint32) float64 {
	if paramID == 0 {
		gainBits := atomic.LoadInt64(&p.gain)
		return floatFromBits(uint64(gainBits))
	}
	return 0.0
}

// OnActivate is called when plugin is activated
func (p *GainPlugin) OnActivate() error {
	return nil
}

// OnDeactivate is called when plugin is deactivated
func (p *GainPlugin) OnDeactivate() {
	// Nothing to do
}

// Cleanup releases any resources
func (p *GainPlugin) Cleanup() {
	// Nothing to clean up
}

// Helper functions for atomic float64 operations
func floatToBits(f float64) uint64 {
	return *(*uint64)(unsafe.Pointer(&f))
}

func floatFromBits(b uint64) float64 {
	return *(*float64)(unsafe.Pointer(&b))
}

func init() {
	// Register the plugin during library initialization
	plugin := &GainPlugin{}
	api.RegisterSimplePlugin(plugin)
}

func main() {
	// This is only called when run as standalone executable
	// Plugin registration happens in init() for shared library loading
}