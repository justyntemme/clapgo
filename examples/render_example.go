package main

// This example demonstrates how to implement the Render extension
// in a ClapGo plugin. This extension allows plugins to adapt their
// processing based on whether they're in realtime or offline mode.

import "C"
import (
	"github.com/justyntemme/clapgo/pkg/api"
	"unsafe"
)

// Example plugin with render mode support
type PluginWithRenderSupport struct {
	api.BasePlugin
	
	// Render mode helper
	renderMode *api.RenderModeHelper
	
	// Different processing parameters for different modes
	realtimeOversample int
	offlineOversample  int
	
	// Hardware connection status
	hasHardware bool
}

// NewPluginWithRenderSupport creates a new plugin with render support
func NewPluginWithRenderSupport() *PluginWithRenderSupport {
	return &PluginWithRenderSupport{
		// This plugin supports offline rendering and is not connected to hardware
		renderMode: api.NewRenderModeHelper(true, false),
		
		// Use 2x oversampling in realtime, 8x in offline mode
		realtimeOversample: 2,
		offlineOversample:  8,
		
		hasHardware: false,
	}
}

// Implement the RenderProvider interface
func (p *PluginWithRenderSupport) HasHardRealtimeRequirement() bool {
	// If we were connected to hardware (e.g., external synthesizer,
	// hardware effects unit), we would return true here
	return p.renderMode.HasHardRealtimeRequirement()
}

func (p *PluginWithRenderSupport) SetRenderMode(mode int32) bool {
	// Try to set the render mode
	if !p.renderMode.SetRenderMode(mode) {
		return false
	}
	
	// Adapt processing based on the mode
	switch mode {
	case api.RenderRealtime:
		// Configure for realtime processing
		p.configureRealtimeProcessing()
		
	case api.RenderOffline:
		// Configure for offline processing  
		p.configureOfflineProcessing()
	}
	
	return true
}

func (p *PluginWithRenderSupport) configureRealtimeProcessing() {
	// Use lighter algorithms for realtime processing
	// - Lower oversampling rate
	// - Simpler interpolation
	// - Lower FFT sizes
	// - Faster approximations
}

func (p *PluginWithRenderSupport) configureOfflineProcessing() {
	// Use highest quality algorithms for offline rendering
	// - Maximum oversampling
	// - High-quality interpolation
	// - Larger FFT sizes for better frequency resolution
	// - Precise calculations instead of approximations
}

// Process demonstrates how to adapt processing based on render mode
func (p *PluginWithRenderSupport) Process(inputs, outputs [][]float32) {
	oversampleRate := p.realtimeOversample
	if p.renderMode.IsOffline() {
		oversampleRate = p.offlineOversample
	}
	
	// Process with appropriate oversampling rate
	for ch := range outputs {
		if ch < len(inputs) {
			// Upsample
			upsampled := p.upsample(inputs[ch], oversampleRate)
			
			// Process at higher sample rate
			processed := p.processChannel(upsampled)
			
			// Downsample back
			outputs[ch] = p.downsample(processed, oversampleRate)
		}
	}
}

// Example processing methods (simplified)
func (p *PluginWithRenderSupport) upsample(input []float32, rate int) []float32 {
	// In a real implementation, this would use proper interpolation
	result := make([]float32, len(input)*rate)
	// ... upsampling code ...
	return result
}

func (p *PluginWithRenderSupport) processChannel(samples []float32) []float32 {
	// Apply some DSP processing
	// The complexity of this processing could vary based on render mode
	return samples
}

func (p *PluginWithRenderSupport) downsample(input []float32, rate int) []float32 {
	// In a real implementation, this would use proper decimation with anti-aliasing
	result := make([]float32, len(input)/rate)
	// ... downsampling code ...
	return result
}

// Plugin that acts as hardware proxy
type HardwareProxyPlugin struct {
	api.BasePlugin
	renderMode *api.RenderModeHelper
	
	// Connection to hardware device
	hardwareDevice interface{} // Would be actual hardware interface
}

// NewHardwareProxyPlugin creates a plugin that interfaces with hardware
func NewHardwareProxyPlugin() *HardwareProxyPlugin {
	return &HardwareProxyPlugin{
		// This plugin is connected to hardware, so it can't do offline rendering
		renderMode: api.NewRenderModeHelper(false, true),
	}
}

func (p *HardwareProxyPlugin) HasHardRealtimeRequirement() bool {
	// Always returns true because we're connected to hardware
	return true
}

func (p *HardwareProxyPlugin) SetRenderMode(mode int32) bool {
	// Can only operate in realtime mode
	if mode != api.RenderRealtime {
		return false
	}
	return true
}

// Required exports for the extension

//export ClapGo_PluginRenderHasHardRealtimeRequirement
func ClapGo_PluginRenderHasHardRealtimeRequirement(plugin unsafe.Pointer) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	
	// Type assertion based on which plugin type we're using
	if p, ok := (*PluginWithRenderSupport)(plugin); ok {
		return C.bool(p.HasHardRealtimeRequirement())
	}
	
	return C.bool(false)
}

//export ClapGo_PluginRenderSet
func ClapGo_PluginRenderSet(plugin unsafe.Pointer, mode C.int32_t) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	
	// Type assertion based on which plugin type we're using
	if p, ok := (*PluginWithRenderSupport)(plugin); ok {
		return C.bool(p.SetRenderMode(int32(mode)))
	}
	
	return C.bool(false)
}

// Other required plugin exports would go here...

func main() {
	// Required for c-shared build
}