package main

import (
	"fmt"
	"math"
	"sync/atomic"
	
	"github.com/justyntemme/clapgo/pkg/api"
)

// Plugin constants for the gain with GUI plugin
const (
	// PluginID is the unique identifier for this plugin
	PluginID = "com.clapgo.gain-gui"
	
	// PluginName is the human-readable name for this plugin
	PluginName = "Gain with GUI"
	
	// PluginVendor is the name of the plugin vendor
	PluginVendor = "ClapGo"
	
	// PluginVersion is the version of the plugin
	PluginVersion = "1.0.0"
	
	// PluginDescription is a short description of the plugin
	PluginDescription = "A gain plugin with GUI support using ClapGo"
)

// SimpleGainGUIPlugin implements the simplified Plugin interface with GUI support
type SimpleGainGUIPlugin struct {
	gain         int64 // Using int64 for atomic operations (stores float64 bits)
	sampleRate   float64
	hasGUI       bool
	guiVisible   bool
	guiCreated   bool
}

// NewSimpleGainGUIPlugin creates a new simplified gain plugin with GUI
func NewSimpleGainGUIPlugin() *SimpleGainGUIPlugin {
	plugin := &SimpleGainGUIPlugin{
		sampleRate: 44100.0,
		hasGUI:     true,
	}
	// Set default gain to 1.0 (0dB)
	atomic.StoreInt64(&plugin.gain, int64(math.Float64bits(1.0)))
	return plugin
}

// GetInfo returns plugin metadata
func (p *SimpleGainGUIPlugin) GetInfo() api.PluginInfo {
	return api.PluginInfo{
		ID:          PluginID,
		Name:        PluginName,
		Vendor:      PluginVendor,
		URL:         "https://github.com/justyntemme/clapgo",
		ManualURL:   "https://github.com/justyntemme/clapgo",
		SupportURL:  "https://github.com/justyntemme/clapgo/issues",
		Version:     PluginVersion,
		Description: PluginDescription,
		Features:    []string{"audio-effect", "stereo", "mono", "gui"},
	}
}

// ProcessAudio processes audio with Go-native types
func (p *SimpleGainGUIPlugin) ProcessAudio(input, output [][]float32, frameCount uint32) error {
	// Get current gain value atomically
	gainBits := atomic.LoadInt64(&p.gain)
	gain := float32(math.Float64frombits(uint64(gainBits)))
	
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
			outChannel[i] = inChannel[i] * gain
		}
	}
	
	return nil
}

// GetParameterInfo returns information about a parameter
func (p *SimpleGainGUIPlugin) GetParameterInfo(paramID uint32) (api.ParamInfo, error) {
	if paramID == 0 {
		return api.ParamInfo{
			ID:           0,
			Name:         "Gain",
			Module:       "",
			MinValue:     0.0,  // -inf dB
			MaxValue:     2.0,  // +6 dB
			DefaultValue: 1.0,  // 0 dB
			Flags:        api.ParamIsAutomatable | api.ParamIsBoundedBelow | api.ParamIsBoundedAbove,
		}, nil
	}
	return api.ParamInfo{}, api.ErrInvalidParam
}

// GetParameterValue returns the current value of a parameter
func (p *SimpleGainGUIPlugin) GetParameterValue(paramID uint32) float64 {
	if paramID == 0 {
		gainBits := atomic.LoadInt64(&p.gain)
		return math.Float64frombits(uint64(gainBits))
	}
	return 0.0
}

// SetParameterValue sets the value of a parameter
func (p *SimpleGainGUIPlugin) SetParameterValue(paramID uint32, value float64) error {
	if paramID == 0 {
		// Clamp value to valid range
		if value < 0.0 {
			value = 0.0
		}
		if value > 2.0 {
			value = 2.0
		}
		atomic.StoreInt64(&p.gain, int64(math.Float64bits(value)))
		return nil
	}
	return api.ErrInvalidParam
}

// Initialize prepares the plugin for use
func (p *SimpleGainGUIPlugin) Initialize(sampleRate float64) error {
	p.sampleRate = sampleRate
	return nil
}

// Activate prepares the plugin for processing
func (p *SimpleGainGUIPlugin) Activate() error {
	return nil
}

// Deactivate stops the plugin from processing
func (p *SimpleGainGUIPlugin) Deactivate() {
	// Nothing to do
}

// Destroy cleans up plugin resources
func (p *SimpleGainGUIPlugin) Destroy() {
	// Nothing to clean up
}

// GUI-related methods (these would need to be integrated with the wrapper system)

// HasGUI returns true if the plugin has a GUI
func (p *SimpleGainGUIPlugin) HasGUI() bool {
	return p.hasGUI
}

// GetPreferredGUIAPI returns the preferred GUI API
func (p *SimpleGainGUIPlugin) GetPreferredGUIAPI() (apiName string, isFloating bool) {
	// Default to X11 on Linux, adjust based on OS
	return api.WindowAPIX11, false
}

// OnGUICreated is called when the GUI is created
func (p *SimpleGainGUIPlugin) OnGUICreated() {
	p.guiCreated = true
	fmt.Println("Go: GUI created")
}

// OnGUIDestroyed is called when the GUI is destroyed
func (p *SimpleGainGUIPlugin) OnGUIDestroyed() {
	p.guiCreated = false
	p.guiVisible = false
	fmt.Println("Go: GUI destroyed")
}

// OnGUIShown is called when the GUI is shown
func (p *SimpleGainGUIPlugin) OnGUIShown() {
	p.guiVisible = true
	fmt.Println("Go: GUI shown")
}

// OnGUIHidden is called when the GUI is hidden
func (p *SimpleGainGUIPlugin) OnGUIHidden() {
	p.guiVisible = false
	fmt.Println("Go: GUI hidden")
}

// GetGUISize returns the default GUI size
func (p *SimpleGainGUIPlugin) GetGUISize() (width, height uint32) {
	return 400, 300
}

func main() {
	// Register the plugin using the new simplified API
	plugin := NewSimpleGainGUIPlugin()
	api.RegisterPlugin(plugin)
}