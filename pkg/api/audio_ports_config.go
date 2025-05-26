package api

/*
#cgo CFLAGS: -I../../include/clap/include
#include "../../include/clap/include/clap/clap.h"
#include "../../include/clap/include/clap/ext/audio-ports-config.h"
#include <stdlib.h>
#include <string.h>
#include <limits.h>

// Define CLAP_INVALID_ID if not already defined
#ifndef CLAP_INVALID_ID
#define CLAP_INVALID_ID UINT32_MAX
#endif

// Helper function to populate audio ports config
static void populate_audio_ports_config(clap_audio_ports_config_t* config, uint64_t id, const char* name, 
                                       uint32_t input_count, uint32_t output_count,
                                       bool has_main_input, uint32_t main_input_channel_count, const char* main_input_port_type,
                                       bool has_main_output, uint32_t main_output_channel_count, const char* main_output_port_type) {
    config->id = id;
    strncpy(config->name, name, CLAP_NAME_SIZE - 1);
    config->name[CLAP_NAME_SIZE - 1] = '\0';
    config->input_port_count = input_count;
    config->output_port_count = output_count;
    config->has_main_input = has_main_input;
    config->main_input_channel_count = main_input_channel_count;
    config->main_input_port_type = main_input_port_type;
    config->has_main_output = has_main_output;
    config->main_output_channel_count = main_output_channel_count;
    config->main_output_port_type = main_output_port_type;
}
*/
import "C"
import (
    "sync"
    "unsafe"
)

// Global registry of plugins that implement AudioPortsConfigProvider
var (
    audioPortsConfigProviders = make(map[unsafe.Pointer]AudioPortsConfigProvider)
    audioPortsConfigLock      sync.RWMutex
)

// RegisterAudioPortsConfigProvider registers a plugin as an audio ports config provider
func RegisterAudioPortsConfigProvider(plugin unsafe.Pointer, provider AudioPortsConfigProvider) {
    audioPortsConfigLock.Lock()
    defer audioPortsConfigLock.Unlock()
    audioPortsConfigProviders[plugin] = provider
}

// UnregisterAudioPortsConfigProvider removes a plugin from the audio ports config registry
func UnregisterAudioPortsConfigProvider(plugin unsafe.Pointer) {
    audioPortsConfigLock.Lock()
    defer audioPortsConfigLock.Unlock()
    delete(audioPortsConfigProviders, plugin)
}

//export ClapGo_PluginAudioPortsConfigCount
func ClapGo_PluginAudioPortsConfigCount(plugin unsafe.Pointer) C.uint32_t {
    audioPortsConfigLock.RLock()
    provider, exists := audioPortsConfigProviders[plugin]
    audioPortsConfigLock.RUnlock()
    
    if !exists || provider == nil {
        return 0
    }
    
    return C.uint32_t(provider.GetAudioPortsConfigCount())
}

//export ClapGo_PluginAudioPortsConfigGet
func ClapGo_PluginAudioPortsConfigGet(plugin unsafe.Pointer, index C.uint32_t, config unsafe.Pointer) C.bool {
    if config == nil {
        return false
    }
    
    audioPortsConfigLock.RLock()
    provider, exists := audioPortsConfigProviders[plugin]
    audioPortsConfigLock.RUnlock()
    
    if !exists || provider == nil {
        return false
    }
    
    // Get the configuration from the provider
    cfg := provider.GetAudioPortsConfig(uint32(index))
    if cfg == nil {
        return false
    }
    
    cConfig := (*C.clap_audio_ports_config_t)(config)
    
    // Convert port types to C strings
    mainInputPortType := (*C.char)(nil)
    mainOutputPortType := (*C.char)(nil)
    
    if cfg.MainInputPortType != "" {
        mainInputPortType = C.CString(cfg.MainInputPortType)
        defer C.free(unsafe.Pointer(mainInputPortType))
    }
    
    if cfg.MainOutputPortType != "" {
        mainOutputPortType = C.CString(cfg.MainOutputPortType)
        defer C.free(unsafe.Pointer(mainOutputPortType))
    }
    
    // Populate the C struct
    C.populate_audio_ports_config(
        cConfig,
        C.uint64_t(cfg.ID),
        C.CString(cfg.Name),
        C.uint32_t(cfg.InputPortCount),
        C.uint32_t(cfg.OutputPortCount),
        C.bool(cfg.HasMainInput),
        C.uint32_t(cfg.MainInputChannelCount),
        mainInputPortType,
        C.bool(cfg.HasMainOutput),
        C.uint32_t(cfg.MainOutputChannelCount),
        mainOutputPortType,
    )
    
    return true
}

//export ClapGo_PluginAudioPortsConfigSelect
func ClapGo_PluginAudioPortsConfigSelect(plugin unsafe.Pointer, configID C.uint64_t) C.bool {
    audioPortsConfigLock.RLock()
    provider, exists := audioPortsConfigProviders[plugin]
    audioPortsConfigLock.RUnlock()
    
    if !exists || provider == nil {
        return false
    }
    
    return C.bool(provider.SelectAudioPortsConfig(uint64(configID)))
}

//export ClapGo_PluginAudioPortsConfigCurrentConfig
func ClapGo_PluginAudioPortsConfigCurrentConfig(plugin unsafe.Pointer) C.uint64_t {
    audioPortsConfigLock.RLock()
    provider, exists := audioPortsConfigProviders[plugin]
    audioPortsConfigLock.RUnlock()
    
    if !exists || provider == nil {
        return C.uint64_t(C.CLAP_INVALID_ID)
    }
    
    return C.uint64_t(provider.GetCurrentConfig())
}

//export ClapGo_PluginAudioPortsConfigGetInfo
func ClapGo_PluginAudioPortsConfigGetInfo(plugin unsafe.Pointer, configID C.uint64_t, portIndex C.uint32_t, isInput C.bool, info unsafe.Pointer) C.bool {
    if info == nil {
        return false
    }
    
    audioPortsConfigLock.RLock()
    provider, exists := audioPortsConfigProviders[plugin]
    audioPortsConfigLock.RUnlock()
    
    if !exists || provider == nil {
        return false
    }
    
    // Get port info from the provider
    portInfo := provider.GetAudioPortInfoForConfig(uint64(configID), uint32(portIndex), bool(isInput))
    if portInfo == nil || portInfo.ID == InvalidID {
        return false
    }
    
    cInfo := (*C.clap_audio_port_info_t)(info)
    
    // Reuse the helper from audio_ports.go
    // Convert port type to C string
    portType := C.CString(portInfo.PortType)
    defer C.free(unsafe.Pointer(portType))
    
    // Convert name to C string
    nameStr := C.CString(portInfo.Name)
    defer C.free(unsafe.Pointer(nameStr))
    
    // Populate the C struct manually since we can't import from audio_ports.go
    cInfo.id = C.uint32_t(portInfo.ID)
    C.strncpy(&cInfo.name[0], nameStr, C.CLAP_NAME_SIZE-1)
    cInfo.name[C.CLAP_NAME_SIZE-1] = 0
    cInfo.channel_count = C.uint32_t(portInfo.ChannelCount)
    cInfo.flags = C.uint32_t(portInfo.Flags)
    
    // Handle port type
    if portInfo.PortType == "mono" {
        cInfo.port_type = C.CString("mono")
    } else if portInfo.PortType == "stereo" {
        cInfo.port_type = C.CString("stereo")
    } else {
        cInfo.port_type = nil
    }
    
    cInfo.in_place_pair = C.uint32_t(portInfo.InPlacePair)
    
    return true
}

// Helper function to create common audio port configurations
func CreateStereoConfig(id uint64, name string) *AudioPortsConfig {
    return &AudioPortsConfig{
        ID:                     id,
        Name:                   name,
        InputPortCount:         1,
        OutputPortCount:        1,
        HasMainInput:           true,
        MainInputChannelCount:  2,
        MainInputPortType:      "stereo",
        HasMainOutput:          true,
        MainOutputChannelCount: 2,
        MainOutputPortType:     "stereo",
    }
}

// CreateMonoConfig creates a mono configuration
func CreateMonoConfig(id uint64, name string) *AudioPortsConfig {
    return &AudioPortsConfig{
        ID:                     id,
        Name:                   name,
        InputPortCount:         1,
        OutputPortCount:        1,
        HasMainInput:           true,
        MainInputChannelCount:  1,
        MainInputPortType:      "mono",
        HasMainOutput:          true,
        MainOutputChannelCount: 1,
        MainOutputPortType:     "mono",
    }
}

// CreateMultiChannelConfig creates a multi-channel configuration
func CreateMultiChannelConfig(id uint64, name string, inputChannels, outputChannels uint32) *AudioPortsConfig {
    return &AudioPortsConfig{
        ID:                     id,
        Name:                   name,
        InputPortCount:         1,
        OutputPortCount:        1,
        HasMainInput:           inputChannels > 0,
        MainInputChannelCount:  inputChannels,
        MainInputPortType:      "multi",
        HasMainOutput:          outputChannels > 0,
        MainOutputChannelCount: outputChannels,
        MainOutputPortType:     "multi",
    }
}