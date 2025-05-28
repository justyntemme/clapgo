package extension

// #include <stdlib.h>
// #include <string.h>
// #include "../../include/clap/include/clap/ext/ambisonic.h"
//
// static inline const void* clap_host_get_extension_helper(const clap_host_t* host, const char* id) {
//     if (host && host->get_extension) {
//         return host->get_extension(host, id);
//     }
//     return NULL;
// }
//
// static inline void clap_host_ambisonic_changed(const clap_host_ambisonic_t* ext, const clap_host_t* host) {
//     if (ext && ext->changed) {
//         ext->changed(host);
//     }
// }
import "C"
import "unsafe"

// Extension IDs
const (
	ExtAmbisonicID     = "clap.ambisonic/3"
	ExtAmbisonicCompat = "clap.ambisonic.draft/3"
	PortAmbisonicID    = "ambisonic"
)

// Additional ambisonic normalization schemes not in configurable_audio_ports.go
const (
	AmbisonicNormalizationSN2D = 3
	AmbisonicNormalizationN2D  = 4
)

// AmbisonicConfig contains ambisonic configuration parameters
type AmbisonicConfig struct {
	Ordering      uint32 // Ordering scheme (ACN, FuMa, etc.)
	Normalization uint32 // Normalization scheme (SN3D, N3D, etc.)
}

// AmbisonicProvider is an extension for plugins that support ambisonic audio
type AmbisonicProvider interface {
	// IsAmbisonicConfigSupported returns true if the given configuration is supported
	IsAmbisonicConfigSupported(config *AmbisonicConfig) bool

	// GetAmbisonicConfig returns the ambisonic configuration for a port
	// Returns true on success, false if the port doesn't support ambisonic
	GetAmbisonicConfig(isInput bool, portIndex uint32) (*AmbisonicConfig, bool)
}

// HostAmbisonic provides host-side ambisonic functionality
type HostAmbisonic struct {
	host      unsafe.Pointer
	ambisonicExt unsafe.Pointer
}

// NewHostAmbisonic creates a new host ambisonic wrapper
func NewHostAmbisonic(host unsafe.Pointer) *HostAmbisonic {
	if host == nil {
		return nil
	}
	
	cHost := (*C.clap_host_t)(host)
	
	// Try to get ambisonic extension
	extPtr := C.clap_host_get_extension_helper(cHost, C.CString(ExtAmbisonicID))
	if extPtr == nil {
		// Try compat version
		extPtr = C.clap_host_get_extension_helper(cHost, C.CString(ExtAmbisonicCompat))
	}
	
	if extPtr == nil {
		return nil
	}
	
	return &HostAmbisonic{
		host:         host,
		ambisonicExt: extPtr,
	}
}

// Changed notifies the host that the ambisonic configuration has changed
// This can only be called when the plugin is deactivated
func (ha *HostAmbisonic) Changed() {
	if ha.ambisonicExt == nil {
		return
	}
	
	ext := (*C.clap_host_ambisonic_t)(ha.ambisonicExt)
	C.clap_host_ambisonic_changed(ext, (*C.clap_host_t)(ha.host))
}

// Conversion functions between Go and C structures

// ambisonicConfigToC converts a Go AmbisonicConfig to C clap_ambisonic_config_t
func ambisonicConfigToC(config *AmbisonicConfig) C.clap_ambisonic_config_t {
	return C.clap_ambisonic_config_t{
		ordering:      C.uint32_t(config.Ordering),
		normalization: C.uint32_t(config.Normalization),
	}
}

// ambisonicConfigFromC converts a C clap_ambisonic_config_t to Go AmbisonicConfig
func ambisonicConfigFromC(cConfig *C.clap_ambisonic_config_t) *AmbisonicConfig {
	if cConfig == nil {
		return nil
	}
	return &AmbisonicConfig{
		Ordering:      uint32(cConfig.ordering),
		Normalization: uint32(cConfig.normalization),
	}
}