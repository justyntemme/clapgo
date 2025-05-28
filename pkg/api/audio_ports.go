package api

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
    "github.com/justyntemme/clapgo/pkg/audio"
)

// Additional audio port flags not in constants.go
const (
    // AudioPortSupports64Bits indicates the port supports 64-bit audio
    AudioPortSupports64Bits = C.CLAP_AUDIO_PORT_SUPPORTS_64BITS
    
    // AudioPortPrefers64Bits indicates the port prefers 64-bit audio
    AudioPortPrefers64Bits = C.CLAP_AUDIO_PORT_PREFERS_64BITS
)

// Global registry of plugins that implement AudioPortsProvider
var (
    audioPortsProviders = make(map[unsafe.Pointer]AudioPortsProvider)
    audioPortsLock      sync.RWMutex
)

// convertAudioPortInfo converts audio.PortInfo to api.AudioPortInfo
func convertAudioPortInfo(audioPort audio.PortInfo) AudioPortInfo {
    return AudioPortInfo{
        ID:           audioPort.ID,
        Name:         audioPort.Name,
        ChannelCount: audioPort.ChannelCount,
        Flags:        audioPort.Flags,
        PortType:     audioPort.PortType,
        InPlacePair:  audioPort.InPlacePair,
    }
}

// RegisterAudioPortsProvider registers a plugin as an audio ports provider
func RegisterAudioPortsProvider(plugin unsafe.Pointer, provider AudioPortsProvider) {
    audioPortsLock.Lock()
    defer audioPortsLock.Unlock()
    audioPortsProviders[plugin] = provider
}

// UnregisterAudioPortsProvider removes a plugin from the audio ports registry
func UnregisterAudioPortsProvider(plugin unsafe.Pointer) {
    audioPortsLock.Lock()
    defer audioPortsLock.Unlock()
    delete(audioPortsProviders, plugin)
}

//export ClapGo_PluginAudioPortsCount
func ClapGo_PluginAudioPortsCount(plugin unsafe.Pointer, isInput C.bool) C.uint32_t {
    // First check audio package registry (preferred)
    count := audio.GetPortsCount(plugin, bool(isInput))
    if count > 0 {
        return C.uint32_t(count)
    }
    
    // Fallback to api package registry
    audioPortsLock.RLock()
    provider, exists := audioPortsProviders[plugin]
    audioPortsLock.RUnlock()
    
    if !exists || provider == nil {
        // Default: 1 stereo port
        return 1
    }
    
    count = provider.GetAudioPortCount(bool(isInput))
    return C.uint32_t(count)
}

//export ClapGo_PluginAudioPortsGet
func ClapGo_PluginAudioPortsGet(plugin unsafe.Pointer, index C.uint32_t, isInput C.bool, info unsafe.Pointer) C.bool {
    if info == nil {
        return false
    }
    
    cInfo := (*C.clap_audio_port_info_t)(info)
    
    // First check audio package registry (preferred)
    if audioPortInfo, found := audio.GetPortInfo(plugin, uint32(index), bool(isInput)); found {
        // Convert name to C string
        nameStr := C.CString(audioPortInfo.Name)
        defer C.free(unsafe.Pointer(nameStr))
        
        // Convert port type to C string
        portType := C.CString(audioPortInfo.PortType)
        defer C.free(unsafe.Pointer(portType))
        
        C.populate_audio_port_info(
            cInfo,
            C.uint32_t(audioPortInfo.ID),
            nameStr,
            C.uint32_t(audioPortInfo.ChannelCount),
            C.uint32_t(audioPortInfo.Flags),
            portType,
            C.uint32_t(audioPortInfo.InPlacePair),
        )
        return true
    }
    
    // Fallback to api package registry
    audioPortsLock.RLock()
    provider, exists := audioPortsProviders[plugin]
    audioPortsLock.RUnlock()
    
    if !exists || provider == nil {
        // Default implementation: single stereo port
        if index != 0 {
            return false
        }
        
        name := "Audio Input"
        if !bool(isInput) {
            name = "Audio Output"
        }
        
        nameStr := C.CString(name)
        defer C.free(unsafe.Pointer(nameStr))
        
        portTypeStr := C.CString("stereo")
        defer C.free(unsafe.Pointer(portTypeStr))
        
        C.populate_audio_port_info(
            cInfo,
            0,                          // id
            nameStr,                    // name
            2,                         // channel_count
            C.CLAP_AUDIO_PORT_IS_MAIN, // flags
            portTypeStr,               // port_type
            0,                         // in_place_pair
        )
        return true
    }
    
    // Get port info from the provider
    portInfo := provider.GetAudioPortInfo(uint32(index), bool(isInput))
    if portInfo.ID == InvalidID {
        return false
    }
    
    // Convert port type to C string
    portType := C.CString(portInfo.PortType)
    defer C.free(unsafe.Pointer(portType))
    
    // Convert name to C string
    nameStr := C.CString(portInfo.Name)
    defer C.free(unsafe.Pointer(nameStr))
    
    // Populate the C struct
    C.populate_audio_port_info(
        cInfo,
        C.uint32_t(portInfo.ID),
        nameStr,
        C.uint32_t(portInfo.ChannelCount),
        C.uint32_t(portInfo.Flags),
        portType,
        C.uint32_t(portInfo.InPlacePair),
    )
    
    return true
}

// Helper function to create a standard stereo port info
func CreateStereoPort(id uint32, name string, isMain bool) AudioPortInfo {
    flags := uint32(0)
    if isMain {
        flags |= AudioPortIsMain
    }
    
    return AudioPortInfo{
        ID:           id,
        Name:         name,
        ChannelCount: 2,
        Flags:        flags,
        PortType:     "stereo",
        InPlacePair:  id, // For stereo, in-place pair is usually the same ID
    }
}

// Helper function to create a mono port info
func CreateMonoPort(id uint32, name string, isMain bool) AudioPortInfo {
    flags := uint32(0)
    if isMain {
        flags |= AudioPortIsMain
    }
    
    return AudioPortInfo{
        ID:           id,
        Name:         name,
        ChannelCount: 1,
        Flags:        flags,
        PortType:     "mono",
        InPlacePair:  InvalidID, // Mono ports typically don't have in-place pairs
    }
}

// Helper function to create a multi-channel port info
func CreateMultiChannelPort(id uint32, name string, channelCount uint32, isMain bool) AudioPortInfo {
    flags := uint32(0)
    if isMain {
        flags |= AudioPortIsMain
    }
    
    return AudioPortInfo{
        ID:           id,
        Name:         name,
        ChannelCount: channelCount,
        Flags:        flags,
        PortType:     "multi", // Custom port type for multi-channel
        InPlacePair:  InvalidID,
    }
}