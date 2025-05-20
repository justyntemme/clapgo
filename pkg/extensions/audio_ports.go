// Package extensions provides implementations of CLAP extensions for Go plugins.
package extensions

// #include <stdlib.h>
// #include "../../include/clap/include/clap/clap.h"
import "C"
import (
	"unsafe"

	"github.com/justyntemme/clapgo/pkg/api"
)

// AudioPortsExtension implements the CLAP audio ports extension.
type AudioPortsExtension struct {
	// The provider that will supply the actual audio port information
	provider api.AudioPortsProvider

	// The C interface to expose to the host
	cInterface C.clap_plugin_audio_ports_t
}

// NewAudioPortsExtension creates a new audio ports extension.
func NewAudioPortsExtension(provider api.AudioPortsProvider) *AudioPortsExtension {
	if provider == nil {
		return nil
	}

	ext := &AudioPortsExtension{
		provider: provider,
	}

	// Initialize the C interface
	ext.cInterface.count = C.clap_plugin_audio_ports_count
	ext.cInterface.get = C.clap_plugin_audio_ports_get

	return ext
}

// GetInterface returns the C interface for the extension.
func (e *AudioPortsExtension) GetInterface() unsafe.Pointer {
	return unsafe.Pointer(&e.cInterface)
}

//export clap_plugin_audio_ports_count
func clap_plugin_audio_ports_count(plugin *C.clap_plugin_t, isInput C.bool) C.uint32_t {
	// In a real implementation, we would look up the plugin from the registry
	// and call the provider's GetAudioPortCount method.
	// For now, return a simple stereo configuration
	return 1 // One stereo port
}

//export clap_plugin_audio_ports_get
func clap_plugin_audio_ports_get(plugin *C.clap_plugin_t, index C.uint32_t, isInput C.bool, info *C.clap_audio_port_info_t) C.bool {
	// In a real implementation, we would look up the plugin from the registry
	// and call the provider's GetAudioPortInfo method.
	// For now, return a simple stereo configuration
	if index != 0 {
		return C.bool(false)
	}

	// Set up a stereo port
	info.id = 0
	info.channel_count = 2
	info.flags = C.CLAP_AUDIO_PORT_IS_MAIN
	info.port_type = C.CString("stereo")
	info.in_place_pair = C.CLAP_INVALID_ID

	if bool(isInput) {
		C.strncpy(&info.name[0], C.CString("Stereo In"), C.size_t(C.CLAP_NAME_SIZE-1))
	} else {
		C.strncpy(&info.name[0], C.CString("Stereo Out"), C.size_t(C.CLAP_NAME_SIZE-1))
	}

	return C.bool(true)
}