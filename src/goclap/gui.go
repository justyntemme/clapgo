package goclap

// #include <stdlib.h>
// #include <string.h>
// #include "../../include/clap/include/clap/clap.h"
// #include "../../include/clap/include/clap/ext/gui.h"
import "C"
import (
	"unsafe"
)

// Constants for GUI extension
const (
	ExtGUI = "clap.gui"
)

// Window API constants
const (
	WindowAPIWin32   = "win32"
	WindowAPICocoa   = "cocoa"
	WindowAPIX11     = "x11"
	WindowAPIWayland = "wayland"
)

// PluginGUIExtension represents the CLAP GUI extension for plugins
type PluginGUIExtension struct {
	plugin       AudioProcessor
	guiExt       unsafe.Pointer // Pointer to the C extension
}

// NewPluginGUIExtension creates a new GUI extension
func NewPluginGUIExtension(plugin AudioProcessor) *PluginGUIExtension {
	if plugin == nil {
		return nil
	}
	
	ext := &PluginGUIExtension{
		plugin: plugin,
	}
	
	// GUI extension is implemented in C++, so we don't create it here
	// It's created in gui_bridge.cpp
	
	return ext
}

// GetExtensionPointer returns the C extension interface pointer
func (g *PluginGUIExtension) GetExtensionPointer() unsafe.Pointer {
	return g.guiExt
}

// GUIProvider interface for plugins that support a GUI
type GUIProvider interface {
	// HasGUI returns true if the plugin has a GUI
	HasGUI() bool
	
	// GetPreferredGUIAPI returns the preferred GUI API
	GetPreferredGUIAPI() (api string, isFloating bool)
	
	// OnGUICreated is called when the GUI is created
	OnGUICreated()
	
	// OnGUIDestroyed is called when the GUI is destroyed
	OnGUIDestroyed()
	
	// OnGUIShown is called when the GUI is shown
	OnGUIShown()
	
	// OnGUIHidden is called when the GUI is hidden
	OnGUIHidden()
	
	// GetGUISize returns the default GUI size
	GetGUISize() (width, height uint32)
}

// These functions are called from C through CGO

//export GoGUICreated
func GoGUICreated(pluginPtr unsafe.Pointer) C.bool {
	// Get the processor
	processor := getProcessorFromHandle(pluginPtr)
	if processor == nil {
		return C.bool(false)
	}
	
	// Check if the plugin implements the GUIProvider interface
	if guiProvider, ok := processor.(GUIProvider); ok {
		guiProvider.OnGUICreated()
		return C.bool(true)
	}
	
	return C.bool(false)
}

//export GoGUIDestroyed
func GoGUIDestroyed(pluginPtr unsafe.Pointer) {
	// Get the processor
	processor := getProcessorFromHandle(pluginPtr)
	if processor == nil {
		return
	}
	
	// Check if the plugin implements the GUIProvider interface
	if guiProvider, ok := processor.(GUIProvider); ok {
		guiProvider.OnGUIDestroyed()
	}
}

//export GoGUIShown
func GoGUIShown(pluginPtr unsafe.Pointer) C.bool {
	// Get the processor
	processor := getProcessorFromHandle(pluginPtr)
	if processor == nil {
		return C.bool(false)
	}
	
	// Check if the plugin implements the GUIProvider interface
	if guiProvider, ok := processor.(GUIProvider); ok {
		guiProvider.OnGUIShown()
		return C.bool(true)
	}
	
	return C.bool(false)
}

//export GoGUIHidden
func GoGUIHidden(pluginPtr unsafe.Pointer) C.bool {
	// Get the processor
	processor := getProcessorFromHandle(pluginPtr)
	if processor == nil {
		return C.bool(false)
	}
	
	// Check if the plugin implements the GUIProvider interface
	if guiProvider, ok := processor.(GUIProvider); ok {
		guiProvider.OnGUIHidden()
		return C.bool(true)
	}
	
	return C.bool(false)
}

//export GoGUIGetSize
func GoGUIGetSize(pluginPtr unsafe.Pointer, width, height *C.uint32_t) C.bool {
	// Get the processor
	processor := getProcessorFromHandle(pluginPtr)
	if processor == nil || width == nil || height == nil {
		return C.bool(false)
	}
	
	// Check if the plugin implements the GUIProvider interface
	if guiProvider, ok := processor.(GUIProvider); ok {
		w, h := guiProvider.GetGUISize()
		*width = C.uint32_t(w)
		*height = C.uint32_t(h)
		return C.bool(true)
	}
	
	// Default size if not specified
	*width = 800
	*height = 600
	return C.bool(true)
}

//export GoGUIHasGUI
func GoGUIHasGUI(pluginPtr unsafe.Pointer) C.bool {
	// Get the processor
	processor := getProcessorFromHandle(pluginPtr)
	if processor == nil {
		return C.bool(false)
	}
	
	// Check if the plugin implements the GUIProvider interface
	if guiProvider, ok := processor.(GUIProvider); ok {
		return C.bool(guiProvider.HasGUI())
	}
	
	return C.bool(false)
}

//export GoGUIGetPreferredAPI
func GoGUIGetPreferredAPI(pluginPtr unsafe.Pointer, api **C.char, isFloating *C.bool) C.bool {
	// Get the processor
	processor := getProcessorFromHandle(pluginPtr)
	if processor == nil || api == nil || isFloating == nil {
		return C.bool(false)
	}
	
	// Check if the plugin implements the GUIProvider interface
	if guiProvider, ok := processor.(GUIProvider); ok {
		apiStr, floating := guiProvider.GetPreferredGUIAPI()
		if apiStr != "" {
			*api = C.CString(apiStr)
			*isFloating = C.bool(floating)
			return C.bool(true)
		}
	}
	
	return C.bool(false)
}