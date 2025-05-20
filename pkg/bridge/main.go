package bridge

// #include <stdlib.h>
import "C"
import (
	"fmt"
	"os"

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

// This is called when the package is initialized
func init() {
	// Any bridge initialization can go here
}