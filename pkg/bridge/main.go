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

// Sentinel value to indicate the plugin registration signal
// This is used to ensure we don't try to register the plugin info twice
var hasRegisteredPlugins = false

// registerPlugins scans for plugin libraries and loads them dynamically
// This is a more flexible approach compared to hardcoding plugin registrations
func registerPlugins() {
	if !hasRegisteredPlugins {
		fmt.Println("Plugin registry initialized")
		
		// Get plugin ID from environment variable (if specified)
		pluginID := os.Getenv("CLAPGO_PLUGIN_ID")
		if pluginID != "" {
			// When building the plugin-specific library, register a stub implementation
			// that will later be replaced by the actual implementation
			registerPluginStub(pluginID)
		} else {
			// When building the main bridge library, we need to detect plugins
			// This could involve scanning directories or using a manifest
			// For now, we'll just register known plugins as stubs
			
			// Since this is a PoC, we'll just register the known plugins
			// In a real implementation, this would scan for plugins dynamically
			searchForPlugins()
		}
		
		hasRegisteredPlugins = true
	}
}

// registerPluginStub registers a stub implementation for a specific plugin
func registerPluginStub(pluginID string) {
	var info api.PluginInfo
	
	// Set up default info based on the plugin ID
	// In a real implementation, this would query the plugin for its info
	switch pluginID {
	case "com.clapgo.gain":
		info = api.PluginInfo{
			ID:          pluginID,
			Name:        "Simple Gain",
			Vendor:      "ClapGo",
			URL:         "https://github.com/justyntemme/clapgo",
			ManualURL:   "https://github.com/justyntemme/clapgo",
			SupportURL:  "https://github.com/justyntemme/clapgo/issues",
			Version:     "1.0.0",
			Description: "A simple gain plugin using ClapGo",
			Features:    []string{"audio-effect", "stereo", "mono"},
		}
	case "com.clapgo.synth":
		info = api.PluginInfo{
			ID:          pluginID,
			Name:        "Simple Synth",
			Vendor:      "ClapGo",
			URL:         "https://github.com/justyntemme/clapgo",
			ManualURL:   "https://github.com/justyntemme/clapgo",
			SupportURL:  "https://github.com/justyntemme/clapgo/issues",
			Version:     "1.0.0",
			Description: "A simple synthesizer plugin using ClapGo",
			Features:    []string{"instrument", "synthesizer", "stereo"},
		}
	default:
		// Default info for unknown plugins
		info = api.PluginInfo{
			ID:          pluginID,
			Name:        fmt.Sprintf("Plugin %s", pluginID),
			Vendor:      "ClapGo",
			URL:         "https://github.com/justyntemme/clapgo",
			Version:     "1.0.0",
			Description: fmt.Sprintf("ClapGo plugin: %s", pluginID),
			Features:    []string{"audio-effect"},
		}
	}
	
	// Register the stub plugin
	fmt.Printf("Registering plugin stub: %s\n", info.ID)
	registry.Register(info, func() api.Plugin {
		return &StubPlugin{
			id:   info.ID,
			info: info,
		}
	})
}

// searchForPlugins looks for available plugins
// In a real implementation, this would scan directories, read manifests, etc.
func searchForPlugins() {
	// For this proof of concept, we'll just register the known plugins as stubs
	knownPlugins := []string{
		"com.clapgo.gain",
		"com.clapgo.synth",
	}
	
	for _, pluginID := range knownPlugins {
		registerPluginStub(pluginID)
	}
}