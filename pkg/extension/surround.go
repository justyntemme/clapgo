package extension

/*
#cgo CFLAGS: -I../../include/clap/include
#include "../../include/clap/include/clap/clap.h"
#include "../../include/clap/include/clap/ext/surround.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
    "sync"
    "unsafe"
)

// Surround channel identifiers (matching the CLAP standard)
const (
    // Standard surround channels
    SurroundChannelFL  = C.CLAP_SURROUND_FL  // Front Left
    SurroundChannelFR  = C.CLAP_SURROUND_FR  // Front Right
    SurroundChannelFC  = C.CLAP_SURROUND_FC  // Front Center
    SurroundChannelLFE = C.CLAP_SURROUND_LFE // Low Frequency Effects
    SurroundChannelBL  = C.CLAP_SURROUND_BL  // Back Left (Rear Left)
    SurroundChannelBR  = C.CLAP_SURROUND_BR  // Back Right (Rear Right)
    SurroundChannelFLC = C.CLAP_SURROUND_FLC // Front Left of Center
    SurroundChannelFRC = C.CLAP_SURROUND_FRC // Front Right of Center
    SurroundChannelBC  = C.CLAP_SURROUND_BC  // Back Center
    SurroundChannelSL  = C.CLAP_SURROUND_SL  // Side Left
    SurroundChannelSR  = C.CLAP_SURROUND_SR  // Side Right
    
    // Height/top channels
    SurroundChannelTC  = C.CLAP_SURROUND_TC  // Top Center
    SurroundChannelTFL = C.CLAP_SURROUND_TFL // Top Front Left
    SurroundChannelTFC = C.CLAP_SURROUND_TFC // Top Front Center
    SurroundChannelTFR = C.CLAP_SURROUND_TFR // Top Front Right
    SurroundChannelTBL = C.CLAP_SURROUND_TBL // Top Back Left
    SurroundChannelTBC = C.CLAP_SURROUND_TBC // Top Back Center
    SurroundChannelTBR = C.CLAP_SURROUND_TBR // Top Back Right
)

// SurroundProvider is implemented by plugins that support surround audio configurations
type SurroundProvider interface {
	// IsChannelMaskSupported returns true if the given channel mask is supported
	IsChannelMaskSupported(channelMask uint64) bool
	
	// GetChannelMap returns the channel map for a given port
	// Returns the channel identifiers in order, or nil if not supported
	GetChannelMap(isInput bool, portIndex uint32) []uint8
}

// Global registry of plugins that implement SurroundProvider
var (
    surroundProviders = make(map[unsafe.Pointer]SurroundProvider)
    surroundLock      sync.RWMutex
)

// RegisterSurroundProvider registers a plugin as a surround provider
func RegisterSurroundProvider(plugin unsafe.Pointer, provider SurroundProvider) {
    surroundLock.Lock()
    defer surroundLock.Unlock()
    surroundProviders[plugin] = provider
}

// UnregisterSurroundProvider removes a plugin from the surround registry
func UnregisterSurroundProvider(plugin unsafe.Pointer) {
    surroundLock.Lock()
    defer surroundLock.Unlock()
    delete(surroundProviders, plugin)
}

//export ClapGo_PluginSurroundIsChannelMaskSupported
func ClapGo_PluginSurroundIsChannelMaskSupported(plugin unsafe.Pointer, channelMask C.uint64_t) C.bool {
    surroundLock.RLock()
    provider, exists := surroundProviders[plugin]
    surroundLock.RUnlock()
    
    if !exists || provider == nil {
        return false
    }
    
    return C.bool(provider.IsChannelMaskSupported(uint64(channelMask)))
}

//export ClapGo_PluginSurroundGetChannelMap
func ClapGo_PluginSurroundGetChannelMap(plugin unsafe.Pointer, isInput C.bool, portIndex C.uint32_t, 
    channelMap *C.uint8_t, channelMapCapacity C.uint32_t) C.uint32_t {
    
    if channelMap == nil || channelMapCapacity == 0 {
        return 0
    }
    
    surroundLock.RLock()
    provider, exists := surroundProviders[plugin]
    surroundLock.RUnlock()
    
    if !exists || provider == nil {
        return 0
    }
    
    // Get the channel map from the provider
    goChannelMap := provider.GetChannelMap(bool(isInput), uint32(portIndex))
    if len(goChannelMap) == 0 {
        return 0
    }
    
    // Copy to C array
    copyCount := len(goChannelMap)
    if uint32(copyCount) > uint32(channelMapCapacity) {
        copyCount = int(channelMapCapacity)
    }
    
    // Convert to C array
    cArray := (*[1024]C.uint8_t)(unsafe.Pointer(channelMap))[:channelMapCapacity:channelMapCapacity]
    for i := 0; i < copyCount; i++ {
        cArray[i] = C.uint8_t(goChannelMap[i])
    }
    
    return C.uint32_t(copyCount)
}

// Helper functions for common surround configurations

// CreateStereoChannelMap creates a channel map for stereo
func CreateStereoChannelMap() []uint8 {
    return []uint8{
        uint8(SurroundChannelFL),
        uint8(SurroundChannelFR),
    }
}

// Create51ChannelMap creates a channel map for 5.1 surround
func Create51ChannelMap() []uint8 {
    return []uint8{
        uint8(SurroundChannelFL),
        uint8(SurroundChannelFR),
        uint8(SurroundChannelFC),
        uint8(SurroundChannelLFE),
        uint8(SurroundChannelBL),
        uint8(SurroundChannelBR),
    }
}

// Create71ChannelMap creates a channel map for 7.1 surround
func Create71ChannelMap() []uint8 {
    return []uint8{
        uint8(SurroundChannelFL),
        uint8(SurroundChannelFR),
        uint8(SurroundChannelFC),
        uint8(SurroundChannelLFE),
        uint8(SurroundChannelBL),
        uint8(SurroundChannelBR),
        uint8(SurroundChannelSL),
        uint8(SurroundChannelSR),
    }
}

// Create714ChannelMap creates a channel map for 7.1.4 surround (with height)
func Create714ChannelMap() []uint8 {
    return []uint8{
        uint8(SurroundChannelFL),
        uint8(SurroundChannelFR),
        uint8(SurroundChannelFC),
        uint8(SurroundChannelLFE),
        uint8(SurroundChannelBL),
        uint8(SurroundChannelBR),
        uint8(SurroundChannelSL),
        uint8(SurroundChannelSR),
        uint8(SurroundChannelTFL),
        uint8(SurroundChannelTFR),
        uint8(SurroundChannelTBL),
        uint8(SurroundChannelTBR),
    }
}

// Channel mask helpers

// ChannelMaskStereo is the channel mask for stereo
const ChannelMaskStereo = (1 << SurroundChannelFL) | (1 << SurroundChannelFR)

// ChannelMask51 is the channel mask for 5.1 surround
const ChannelMask51 = (1 << SurroundChannelFL) | (1 << SurroundChannelFR) | 
                      (1 << SurroundChannelFC) | (1 << SurroundChannelLFE) |
                      (1 << SurroundChannelBL) | (1 << SurroundChannelBR)

// ChannelMask71 is the channel mask for 7.1 surround
const ChannelMask71 = ChannelMask51 | (1 << SurroundChannelSL) | (1 << SurroundChannelSR)

// ChannelMask714 is the channel mask for 7.1.4 surround
const ChannelMask714 = ChannelMask71 | (1 << SurroundChannelTFL) | (1 << SurroundChannelTFR) |
                       (1 << SurroundChannelTBL) | (1 << SurroundChannelTBR)

// GetChannelCountFromMask returns the number of channels in a channel mask
func GetChannelCountFromMask(mask uint64) int {
    count := 0
    for i := 0; i < 64; i++ {
        if mask & (1 << i) != 0 {
            count++
        }
    }
    return count
}

// ValidateChannelMap validates that all channel identifiers are valid
func ValidateChannelMap(channelMap []uint8) bool {
    for _, ch := range channelMap {
        // Check if channel is within valid range (0-18 for current CLAP spec)
        if ch > 18 {
            return false
        }
    }
    return true
}