package goclap

// #include <stdlib.h>
// #include <string.h>
// #include "../../include/clap/include/clap/clap.h"
// #include "../../include/clap/include/clap/ext/note-ports.h"
//
// // Define types for function pointers to avoid CGO issues
// typedef uint32_t (*clap_plugin_note_ports_count_func)(struct clap_plugin *plugin, bool is_input);
// typedef bool (*clap_plugin_note_ports_get_func)(struct clap_plugin *plugin, uint32_t index, bool is_input, clap_note_port_info_t *info);
//
// // Bridge functions for the note ports extension - need to match the implementation exactly
// extern uint32_t goNotePortsCountBridge(struct clap_plugin *plugin, bool is_input);
// extern bool goNotePortsGetBridge(struct clap_plugin *plugin, uint32_t index, bool is_input, clap_note_port_info_t *info);
import "C"
import (
	"sync"
	"unsafe"
)

// Note ports extension is already defined in plugin.go

// Note dialects
const (
	NoteDialectCLAP    = 1 << 0
	NoteDialectMIDI    = 1 << 1
	NoteDialectMIDIMPE = 1 << 2
	NoteDialectMIDI2   = 1 << 3
)

// Note port rescan flags
const (
	NotePortsRescanAll   = 1 << 0
	NotePortsRescanNames = 1 << 1
)

// NotePortInfo represents information about a note port
type NotePortInfo struct {
	ID                 uint32
	SupportedDialects  uint32
	PreferredDialect   uint32
	Name               string
}

// NotePortProvider is an interface for plugins that handle note input/output
type NotePortProvider interface {
	// GetNoteInputPortCount returns the number of note input ports
	GetNoteInputPortCount() uint32

	// GetNoteOutputPortCount returns the number of note output ports
	GetNoteOutputPortCount() uint32

	// GetNoteInputPortInfo returns info about a note input port
	GetNoteInputPortInfo(index uint32) *NotePortInfo

	// GetNoteOutputPortInfo returns info about a note output port
	GetNoteOutputPortInfo(index uint32) *NotePortInfo
}

// PluginNotePortsExtension represents the CLAP note ports extension for plugins
type PluginNotePortsExtension struct {
	plugin         AudioProcessor
	notePortsExt   unsafe.Pointer // Pointer to the C extension
}

// NewPluginNotePortsExtension creates a new note ports extension
func NewPluginNotePortsExtension(plugin AudioProcessor) *PluginNotePortsExtension {
	if plugin == nil {
		return nil
	}

	ext := &PluginNotePortsExtension{
		plugin: plugin,
	}

	// Create the C interface
	ext.notePortsExt = createNotePortsExtension(ext)

	// Store in registry as a plugin->extension mapping for proper lookups
	notePortsExtHandlesMutex.Lock()
	pluginToNotePortsExtension[unsafe.Pointer(&plugin)] = ext
	notePortsExtHandlesMutex.Unlock()

	return ext
}

// Destroy cleans up resources associated with this extension
func (np *PluginNotePortsExtension) Destroy() {
	if np.notePortsExt != nil {
		// Clean up our registry entries
		notePortsExtHandlesMutex.Lock()
		delete(notePortsExtHandles, np.notePortsExt)
		delete(pluginToNotePortsExtension, unsafe.Pointer(&np.plugin))
		notePortsExtHandlesMutex.Unlock()
		
		// Free the C memory
		C.free(np.notePortsExt)
		np.notePortsExt = nil
	}
}

// GetExtensionPointer returns the C extension interface pointer
func (np *PluginNotePortsExtension) GetExtensionPointer() unsafe.Pointer {
	return np.notePortsExt
}

// GetNoteInputPortCount returns the number of note input ports
func (np *PluginNotePortsExtension) GetNoteInputPortCount() uint32 {
	// Check if the plugin implements NotePortProvider
	if provider, ok := np.plugin.(NotePortProvider); ok {
		return provider.GetNoteInputPortCount()
	}
	return 0
}

// GetNoteOutputPortCount returns the number of note output ports
func (np *PluginNotePortsExtension) GetNoteOutputPortCount() uint32 {
	// Check if the plugin implements NotePortProvider
	if provider, ok := np.plugin.(NotePortProvider); ok {
		return provider.GetNoteOutputPortCount()
	}
	return 0
}

// GetNoteInputPortInfo returns info about a note input port
func (np *PluginNotePortsExtension) GetNoteInputPortInfo(index uint32, info *C.clap_note_port_info_t) bool {
	// Check if the plugin implements NotePortProvider
	if provider, ok := np.plugin.(NotePortProvider); ok {
		if index >= provider.GetNoteInputPortCount() {
			return false
		}

		portInfo := provider.GetNoteInputPortInfo(index)
		if portInfo == nil {
			return false
		}

		// Fill in the C struct
		info.id = C.clap_id(portInfo.ID)
		info.supported_dialects = C.uint32_t(portInfo.SupportedDialects)
		info.preferred_dialect = C.uint32_t(portInfo.PreferredDialect)
		copyString(portInfo.Name, &info.name[0], C.CLAP_NAME_SIZE)

		return true
	}
	return false
}

// GetNoteOutputPortInfo returns info about a note output port
func (np *PluginNotePortsExtension) GetNoteOutputPortInfo(index uint32, info *C.clap_note_port_info_t) bool {
	// Check if the plugin implements NotePortProvider
	if provider, ok := np.plugin.(NotePortProvider); ok {
		if index >= provider.GetNoteOutputPortCount() {
			return false
		}

		portInfo := provider.GetNoteOutputPortInfo(index)
		if portInfo == nil {
			return false
		}

		// Fill in the C struct
		info.id = C.clap_id(portInfo.ID)
		info.supported_dialects = C.uint32_t(portInfo.SupportedDialects)
		info.preferred_dialect = C.uint32_t(portInfo.PreferredDialect)
		copyString(portInfo.Name, &info.name[0], C.CLAP_NAME_SIZE)

		return true
	}
	return false
}

// External function declarations for the C bridge

//export goNotePortsCount
func goNotePortsCount(plugin unsafe.Pointer, isInput C.bool) C.uint32_t {
	ext := (*PluginNotePortsExtension)(plugin)
	if ext == nil {
		return 0
	}

	if bool(isInput) {
		return C.uint32_t(ext.GetNoteInputPortCount())
	} else {
		return C.uint32_t(ext.GetNoteOutputPortCount())
	}
}

//export goNotePortsGet
func goNotePortsGet(plugin unsafe.Pointer, index C.uint32_t, isInput C.bool, info *C.clap_note_port_info_t) C.bool {
	ext := (*PluginNotePortsExtension)(plugin)
	if ext == nil || info == nil {
		return C.bool(false)
	}

	if bool(isInput) {
		return C.bool(ext.GetNoteInputPortInfo(uint32(index), info))
	} else {
		return C.bool(ext.GetNoteOutputPortInfo(uint32(index), info))
	}
}

// Helper function to create the C extension interface
func createNotePortsExtension(ext *PluginNotePortsExtension) unsafe.Pointer {
	// Allocate memory for the C interface
	cExt := (*C.clap_plugin_note_ports_t)(C.malloc(C.sizeof_clap_plugin_note_ports_t))
	if cExt == nil {
		return nil
	}

	// Set the function pointers - this is critical for proper operation
	cExt.count = C.clap_plugin_note_ports_count_func(C.goNotePortsCountBridge)
	cExt.get = C.clap_plugin_note_ports_get_func(C.goNotePortsGetBridge)
	
	// Store the Go extension in our registry
	setNotePortsExtHandle(unsafe.Pointer(cExt), unsafe.Pointer(ext))
	
	// Return the extension interface
	return unsafe.Pointer(cExt)
}

//export goNotePortsCountBridge
func goNotePortsCountBridge(plugin *C.struct_clap_plugin, is_input C.bool) C.uint32_t {
	// Safety check
	if plugin == nil {
		return 0
	}

	// We need to get the extension from the plugin
	// Get our extension object from the extension map using the plugin_data pointer
	ext := (*PluginNotePortsExtension)(getNotePortsExtHandle(unsafe.Pointer(plugin.plugin_data)))
	if ext == nil {
		return 0
	}
	
	if bool(is_input) {
		return C.uint32_t(ext.GetNoteInputPortCount())
	} else {
		return C.uint32_t(ext.GetNoteOutputPortCount())
	}
}

//export goNotePortsGetBridge
func goNotePortsGetBridge(plugin *C.struct_clap_plugin, index C.uint32_t, is_input C.bool, info *C.clap_note_port_info_t) C.bool {
	// Safety check
	if plugin == nil || info == nil {
		return C.bool(false)
	}

	// We need to get the extension from the plugin
	// Get our extension object from the extension map using the plugin_data pointer
	ext := (*PluginNotePortsExtension)(getNotePortsExtHandle(unsafe.Pointer(plugin.plugin_data)))
	if ext == nil {
		return C.bool(false)
	}
	
	if bool(is_input) {
		return C.bool(ext.GetNoteInputPortInfo(uint32(index), info))
	} else {
		return C.bool(ext.GetNoteOutputPortInfo(uint32(index), info))
	}
}

// Global registry to store note ports extension handles
var (
	notePortsExtHandlesMutex sync.RWMutex
	notePortsExtHandles     = make(map[unsafe.Pointer]unsafe.Pointer)
	pluginToNotePortsExtension = make(map[unsafe.Pointer]*PluginNotePortsExtension)
)

// setNotePortsExtHandle stores a Go extension handle in our registry
func setNotePortsExtHandle(cHandle, goHandle unsafe.Pointer) {
	notePortsExtHandlesMutex.Lock()
	defer notePortsExtHandlesMutex.Unlock()
	notePortsExtHandles[cHandle] = goHandle
	
	// Also store the extension with the plugin instance
	if ext := (*PluginNotePortsExtension)(goHandle); ext != nil {
		if plugin, ok := ext.plugin.(AudioProcessor); ok {
			// Store the association between plugin and extension
			pluginToNotePortsExtension[unsafe.Pointer(&plugin)] = ext
		}
	}
}

// getNotePortsExtHandle retrieves a Go extension handle from our registry
func getNotePortsExtHandle(cHandle unsafe.Pointer) unsafe.Pointer {
	notePortsExtHandlesMutex.RLock()
	defer notePortsExtHandlesMutex.RUnlock()
	
	// First try direct lookup
	if handle, ok := notePortsExtHandles[cHandle]; ok {
		return handle
	}
	
	// If that fails, try to find by plugin instance
	if ext, ok := pluginToNotePortsExtension[cHandle]; ok {
		return unsafe.Pointer(ext)
	}
	
	return nil
}