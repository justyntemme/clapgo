package api

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

// Ambisonic ordering schemes
const (
	AmbisonicOrderingFuMa = 0 // FuMa channel ordering
	AmbisonicOrderingACN  = 1 // ACN channel ordering
)

// Ambisonic normalization schemes
const (
	AmbisonicNormalizationMaxN = 0
	AmbisonicNormalizationSN3D = 1
	AmbisonicNormalizationN3D  = 2
	AmbisonicNormalizationSN2D = 3
	AmbisonicNormalizationN2D  = 4
)

// AmbisonicConfig represents the ambisonic configuration for a port
type AmbisonicConfig struct {
	Ordering      uint32 // One of the AmbisonicOrdering constants
	Normalization uint32 // One of the AmbisonicNormalization constants
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

// Note: These export functions should be implemented in the plugin's main.go file,
// not in the API package. The plugin should use cgo.Handle to access the plugin instance.
// 
// Example implementation in your plugin's main.go:
//
// //export ClapGo_PluginAmbisonicIsConfigSupported
// func ClapGo_PluginAmbisonicIsConfigSupported(plugin unsafe.Pointer, configPtr unsafe.Pointer) C.bool {
//     if plugin == nil || configPtr == nil {
//         return C.bool(false)
//     }
//     p := cgo.Handle(plugin).Value().(api.Plugin)
//     
//     provider, ok := p.(api.AmbisonicProvider)
//     if !ok {
//         return C.bool(false)
//     }
//     
//     cConfig := (*C.clap_ambisonic_config_t)(configPtr)
//     config := &api.AmbisonicConfig{
//         Ordering:      uint32(cConfig.ordering),
//         Normalization: uint32(cConfig.normalization),
//     }
//     
//     return C.bool(provider.IsAmbisonicConfigSupported(config))
// }
//
// //export ClapGo_PluginAmbisonicGetConfig
// func ClapGo_PluginAmbisonicGetConfig(plugin unsafe.Pointer, isInput C.bool, portIndex C.uint32_t, configPtr unsafe.Pointer) C.bool {
//     if plugin == nil || configPtr == nil {
//         return C.bool(false)
//     }
//     p := cgo.Handle(plugin).Value().(api.Plugin)
//     
//     provider, ok := p.(api.AmbisonicProvider)
//     if !ok {
//         return C.bool(false)
//     }
//     
//     config, supported := provider.GetAmbisonicConfig(bool(isInput), uint32(portIndex))
//     if !supported || config == nil {
//         return C.bool(false)
//     }
//     
//     cConfig := (*C.clap_ambisonic_config_t)(configPtr)
//     cConfig.ordering = C.uint32_t(config.Ordering)
//     cConfig.normalization = C.uint32_t(config.Normalization)
//     
//     return C.bool(true)
// }

// Example implementation documentation:
//
// To implement ambisonic support in your plugin:
//
// 1. Implement the AmbisonicProvider interface:
//
//    type MyPlugin struct {
//        // ... other fields ...
//    }
//
//    func (p *MyPlugin) IsAmbisonicConfigSupported(config *AmbisonicConfig) bool {
//        // Check if your plugin supports the given ambisonic configuration
//        // For example, you might only support ACN ordering with SN3D normalization:
//        return config.Ordering == AmbisonicOrderingACN && 
//               config.Normalization == AmbisonicNormalizationSN3D
//    }
//
//    func (p *MyPlugin) GetAmbisonicConfig(isInput bool, portIndex uint32) (*AmbisonicConfig, bool) {
//        // Return the ambisonic configuration for the specified port
//        // Return false if the port doesn't support ambisonic
//        
//        // Example: All ports use ACN ordering with SN3D normalization
//        if portIndex == 0 {
//            return &AmbisonicConfig{
//                Ordering:      AmbisonicOrderingACN,
//                Normalization: AmbisonicNormalizationSN3D,
//            }, true
//        }
//        
//        return nil, false
//    }
//
// 2. If your ambisonic configuration can change, notify the host when deactivated:
//
//    func (p *MyPlugin) Deactivate() {
//        // ... perform deactivation ...
//        
//        // If ambisonic config changed, notify host
//        if p.ambisonicConfigChanged {
//            hostAmbisonic := NewHostAmbisonic(p.host, p)
//            hostAmbisonic.Changed()
//        }
//    }
//
// Note: Ambisonic configuration can only change when the plugin is deactivated.
// The host will query the new configuration after reactivation.