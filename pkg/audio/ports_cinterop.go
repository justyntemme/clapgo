package audio

/*
#cgo CFLAGS: -I../../include/clap/include
#include "../../include/clap/include/clap/clap.h"
#include <stdlib.h>
#include <string.h>

// Helper function to populate audio port info
static void populate_audio_port_info(clap_audio_port_info_t* info, uint32_t id, const char* name, uint32_t channel_count, uint32_t flags, const char* port_type, uint32_t in_place_pair) {
    info->id = id;
    strncpy(info->name, name, CLAP_NAME_SIZE - 1);
    info->name[CLAP_NAME_SIZE - 1] = '\0';
    info->channel_count = channel_count;
    info->flags = flags;
    
    // Use well-known port types directly
    if (port_type && strcmp(port_type, "mono") == 0) {
        info->port_type = CLAP_PORT_MONO;
    } else if (port_type && strcmp(port_type, "stereo") == 0) {
        info->port_type = CLAP_PORT_STEREO;
    } else {
        // For custom port types, we'd need to store the string somewhere persistent
        // For now, just use NULL for unknown types
        info->port_type = NULL;
    }
    
    info->in_place_pair = in_place_pair;
}
*/
import "C"
import (
    "sync"
    "unsafe"
)

// Additional audio port flags not in constants.go
const (
    // AudioPortSupports64Bits indicates the port supports 64-bit audio
    AudioPortSupports64Bits = C.CLAP_AUDIO_PORT_SUPPORTS_64BITS
    
    // AudioPortPrefers64Bits indicates the port prefers 64-bit audio
    AudioPortPrefers64Bits = C.CLAP_AUDIO_PORT_PREFERS_64BITS
)

// Global registry of plugins that implement PortsProvider
var (
    portsProviders = make(map[unsafe.Pointer]PortsProvider)
    portsLock      sync.RWMutex
)

// RegisterPortsProvider registers a plugin as an audio ports provider
func RegisterPortsProvider(plugin unsafe.Pointer, provider PortsProvider) {
    portsLock.Lock()
    defer portsLock.Unlock()
    portsProviders[plugin] = provider
}

// UnregisterPortsProvider removes a plugin from the audio ports registry
func UnregisterPortsProvider(plugin unsafe.Pointer) {
    portsLock.Lock()
    defer portsLock.Unlock()
    delete(portsProviders, plugin)
}

// PortInfoToC converts a Go PortInfo struct to a C clap_audio_port_info_t struct
func PortInfoToC(portInfo PortInfo, cInfo unsafe.Pointer) {
	info := (*C.clap_audio_port_info_t)(cInfo)
	
	C.populate_audio_port_info(
		info,
		C.uint32_t(portInfo.ID),
		C.CString(portInfo.Name),
		C.uint32_t(portInfo.ChannelCount),
		C.uint32_t(portInfo.Flags),
		C.CString(portInfo.PortType),
		C.uint32_t(portInfo.InPlacePair),
	)
}

// GetPortsCount is called by api package to get port count
func GetPortsCount(plugin unsafe.Pointer, isInput bool) uint32 {
    portsLock.RLock()
    provider, exists := portsProviders[plugin]
    portsLock.RUnlock()
    
    if !exists || provider == nil {
        // Default: 1 stereo port
        return 1
    }
    
    return provider.GetAudioPortCount(isInput)
}

// GetPortInfo is called by api package to get port info
func GetPortInfo(plugin unsafe.Pointer, index uint32, isInput bool) (PortInfo, bool) {
    portsLock.RLock()
    provider, exists := portsProviders[plugin]
    portsLock.RUnlock()
    
    if !exists || provider == nil {
        // Default stereo port
        if index == 0 {
            return CreateStereoPort(0, "Main", true), true
        }
        return PortInfo{ID: InvalidID}, false
    }
    
    portInfo := provider.GetAudioPortInfo(index, isInput)
    if portInfo.ID == InvalidID {
        return portInfo, false
    }
    
    return portInfo, true
}