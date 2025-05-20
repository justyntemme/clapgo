package main

// #include <stdlib.h>
import "C"
import (
	"fmt"
	"os"
	"unsafe"

	"github.com/justyntemme/clapgo/pkg/api"
	"github.com/justyntemme/clapgo/pkg/registry"
)

// Main is the main entry point for the bridge when built as a shared library
// This isn't called directly but is required for the package to work as a shared library
func Main() {
	// Get the plugin ID from environment variable
	pluginID := os.Getenv("CLAPGO_PLUGIN_ID")
	if pluginID != "" {
		fmt.Printf("ClapGo bridge initialized for plugin ID: %s\n", pluginID)
		
		// Verify the plugin ID is registered
		count := registry.GetPluginCount()
		found := false
		
		for i := uint32(0); i < count; i++ {
			info := registry.GetPluginInfo(i)
			if info.ID == pluginID {
				found = true
				fmt.Printf("Plugin '%s' (%s) successfully registered\n", info.Name, info.ID)
				break
			}
		}
		
		if !found {
			fmt.Printf("Warning: Plugin ID '%s' not found in registry. Available plugins: %d\n", pluginID, count)
		}
	} else {
		fmt.Println("ClapGo bridge initialized (no specific plugin ID)")
		
		// List all registered plugins
		count := registry.GetPluginCount()
		if count > 0 {
			fmt.Printf("Found %d registered plugins:\n", count)
			for i := uint32(0); i < count; i++ {
				info := registry.GetPluginInfo(i)
				fmt.Printf("  %d: %s (%s)\n", i, info.Name, info.ID)
			}
		} else {
			fmt.Println("Warning: No plugins registered")
		}
	}
}

// main function is required when this package is built as a main package
// When built as a shared library with buildmode=c-shared, this won't be called directly
func main() {
	Main()
}

// This is called when the package is initialized
func init() {
	registerPlugins()
	fmt.Printf("Bridge package initialized, registered plugins: %d\n", registry.GetPluginCount())
}

// StubPlugin is a stub implementation of the api.Plugin interface
// that is used for plugin registration. The actual plugin implementation
// is loaded by the C bridge code.
type StubPlugin struct {
	id       string
	info     api.PluginInfo
	host     unsafe.Pointer
	active   bool
	isInited bool
}

func (p *StubPlugin) Init() bool {
	p.isInited = true
	return true
}

func (p *StubPlugin) Destroy() {
	p.isInited = false
}

func (p *StubPlugin) Activate(sampleRate float64, minFrames, maxFrames uint32) bool {
	p.active = true
	return true
}

func (p *StubPlugin) Deactivate() {
	p.active = false
}

func (p *StubPlugin) StartProcessing() bool {
	return p.isInited && p.active
}

func (p *StubPlugin) StopProcessing() {
}

func (p *StubPlugin) Reset() {
}

func (p *StubPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events api.EventHandler) int {
	// Just pass through audio
	for ch := 0; ch < len(audioOut) && ch < len(audioIn); ch++ {
		copy(audioOut[ch], audioIn[ch])
	}
	return api.ProcessContinue
}

func (p *StubPlugin) GetExtension(id string) unsafe.Pointer {
	return nil
}

func (p *StubPlugin) OnMainThread() {
}

func (p *StubPlugin) GetPluginInfo() api.PluginInfo {
	return p.info
}

func (p *StubPlugin) SaveState() map[string]interface{} {
	return map[string]interface{}{}
}

func (p *StubPlugin) LoadState(data map[string]interface{}) {
}

func (p *StubPlugin) GetPluginID() string {
	return p.id
}

// registerPlugins manually registers all available plugins
func registerPlugins() {
	// Register the gain plugin
	registerGainPlugin()
	
	// Register the synth plugin
	registerSynthPlugin()
}

// registerGainPlugin manually registers the gain plugin
func registerGainPlugin() {
	gainInfo := api.PluginInfo{
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
	
	fmt.Println("Registering gain plugin:", gainInfo.ID)
	registry.Register(gainInfo, func() api.Plugin {
		// Create a stub plugin that satisfies the api.Plugin interface
		// The real implementation will be loaded by the C bridge
		return &StubPlugin{
			id: gainInfo.ID,
			info: gainInfo,
		}
	})
}

// registerSynthPlugin manually registers the synth plugin
func registerSynthPlugin() {
	synthInfo := api.PluginInfo{
		ID:          "com.clapgo.synth",
		Name:        "Simple Synth",
		Vendor:      "ClapGo",
		URL:         "https://github.com/justyntemme/clapgo",
		ManualURL:   "https://github.com/justyntemme/clapgo",
		SupportURL:  "https://github.com/justyntemme/clapgo/issues",
		Version:     "1.0.0",
		Description: "A simple synthesizer plugin using ClapGo",
		Features:    []string{"instrument", "synthesizer", "stereo"},
	}
	
	fmt.Println("Registering synth plugin:", synthInfo.ID)
	registry.Register(synthInfo, func() api.Plugin {
		// Create a stub plugin that satisfies the api.Plugin interface
		// The real implementation will be loaded by the C bridge
		return &StubPlugin{
			id: synthInfo.ID,
			info: synthInfo,
		}
	})
}