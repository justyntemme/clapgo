package main

// #include <stdlib.h>
import "C"
import (
	"fmt"

	"github.com/justyntemme/clapgo/pkg/registry"
)

// Main is the main entry point for the bridge when built as a shared library
// This isn't called directly but is required for the package to work as a shared library
func Main() {
	fmt.Println("ClapGo bridge initialized")
		
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

// main function is required when this package is built as a main package
// When built as a shared library with buildmode=c-shared, this won't be called directly
func main() {
	Main()
}

// This is called when the package is initialized
func init() {
	fmt.Printf("Bridge package initialized, plugins will be registered by their respective packages.\n")
	fmt.Printf("Currently registered plugins: %d\n", registry.GetPluginCount())
}