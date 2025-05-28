package extension

/*
#cgo CFLAGS: -I../../include/clap/include
#include "../../include/clap/include/clap/clap.h"
#include "../../include/clap/include/clap/ext/voice-info.h"
#include <stdlib.h>

// Helper function to populate voice info
static void populate_voice_info(clap_voice_info_t* info, uint32_t voice_count, uint32_t voice_capacity, uint64_t flags) {
    info->voice_count = voice_count;
    info->voice_capacity = voice_capacity;
    info->flags = flags;
}
*/
import "C"
import (
    "sync"
    "unsafe"
)

// Voice info flags
const (
    // VoiceInfoSupportsOverlappingNotes indicates the plugin supports overlapping notes
    // of the same key.
    VoiceInfoFlagSupportsOverlappingNotes = C.CLAP_VOICE_INFO_SUPPORTS_OVERLAPPING_NOTES
)

// VoiceInfo contains information about a plugin's voice capabilities
type VoiceInfo struct {
	VoiceCount    uint32 // Current number of active voices
	VoiceCapacity uint32 // Maximum number of voices
	Flags         uint64 // Voice capability flags
}

// VoiceInfoProvider is implemented by plugins that provide voice information
type VoiceInfoProvider interface {
	// GetVoiceInfo returns the current voice information
	// [main-thread]
	GetVoiceInfo() VoiceInfo
}

// Global registry of plugins that implement VoiceInfoProvider
var (
    voiceInfoProviders = make(map[unsafe.Pointer]VoiceInfoProvider)
    voiceInfoLock      sync.RWMutex
)

// RegisterVoiceInfoProvider registers a plugin as a voice info provider
func RegisterVoiceInfoProvider(plugin unsafe.Pointer, provider VoiceInfoProvider) {
    voiceInfoLock.Lock()
    defer voiceInfoLock.Unlock()
    voiceInfoProviders[plugin] = provider
}

// UnregisterVoiceInfoProvider removes a plugin from the voice info registry
func UnregisterVoiceInfoProvider(plugin unsafe.Pointer) {
    voiceInfoLock.Lock()
    defer voiceInfoLock.Unlock()
    delete(voiceInfoProviders, plugin)
}

//export ClapGo_PluginVoiceInfoGet
func ClapGo_PluginVoiceInfoGet(plugin unsafe.Pointer, info unsafe.Pointer) C.bool {
    if info == nil {
        return false
    }
    
    voiceInfoLock.RLock()
    provider, exists := voiceInfoProviders[plugin]
    voiceInfoLock.RUnlock()
    
    if !exists || provider == nil {
        return false
    }
    
    // Get voice info from the provider
    voiceInfo := provider.GetVoiceInfo()
    
    // Populate the C struct
    cInfo := (*C.clap_voice_info_t)(info)
    C.populate_voice_info(
        cInfo,
        C.uint32_t(voiceInfo.VoiceCount),
        C.uint32_t(voiceInfo.VoiceCapacity),
        C.uint64_t(voiceInfo.Flags),
    )
    
    return true
}

// Helper functions for voice management

// CreateMonophonicVoiceInfo creates voice info for a monophonic plugin
func CreateMonophonicVoiceInfo() VoiceInfo {
    return VoiceInfo{
        VoiceCount:    1,
        VoiceCapacity: 1,
        Flags:         0, // Monophonic doesn't support overlapping notes
    }
}

// CreatePolyphonicVoiceInfo creates voice info for a polyphonic plugin
func CreatePolyphonicVoiceInfo(voiceCount, voiceCapacity uint32, supportsOverlapping bool) VoiceInfo {
    flags := uint64(0)
    if supportsOverlapping {
        flags |= VoiceInfoFlagSupportsOverlappingNotes
    }
    
    return VoiceInfo{
        VoiceCount:    voiceCount,
        VoiceCapacity: voiceCapacity,
        Flags:         flags,
    }
}

// VoiceManager helps manage voice allocation for polyphonic plugins
type VoiceManager struct {
    mutex         sync.RWMutex
    voices        []Voice
    maxVoices     int
    activeCount   int32
    lastVoiceUsed int
}

// Voice represents a single voice in a polyphonic plugin
type Voice struct {
    Active   bool
    NoteID   int32
    Channel  int16
    Key      int16
    Velocity float64
}

// NewVoiceManager creates a new voice manager
func NewVoiceManager(maxVoices int) *VoiceManager {
    return &VoiceManager{
        voices:    make([]Voice, maxVoices),
        maxVoices: maxVoices,
    }
}

// AllocateVoice finds a free voice and allocates it
func (vm *VoiceManager) AllocateVoice(noteID int32, channel, key int16, velocity float64) int {
    vm.mutex.Lock()
    defer vm.mutex.Unlock()
    
    // First, try to find an inactive voice
    for i := 0; i < vm.maxVoices; i++ {
        if !vm.voices[i].Active {
            vm.voices[i] = Voice{
                Active:   true,
                NoteID:   noteID,
                Channel:  channel,
                Key:      key,
                Velocity: velocity,
            }
            vm.activeCount++
            return i
        }
    }
    
    // If no free voice, steal the oldest one (round-robin)
    vm.lastVoiceUsed = (vm.lastVoiceUsed + 1) % vm.maxVoices
    vm.voices[vm.lastVoiceUsed] = Voice{
        Active:   true,
        NoteID:   noteID,
        Channel:  channel,
        Key:      key,
        Velocity: velocity,
    }
    return vm.lastVoiceUsed
}

// ReleaseVoice releases a voice by note ID
func (vm *VoiceManager) ReleaseVoice(noteID int32) bool {
    vm.mutex.Lock()
    defer vm.mutex.Unlock()
    
    for i := 0; i < vm.maxVoices; i++ {
        if vm.voices[i].Active && vm.voices[i].NoteID == noteID {
            vm.voices[i].Active = false
            vm.activeCount--
            return true
        }
    }
    return false
}

// GetActiveVoiceCount returns the number of active voices
func (vm *VoiceManager) GetActiveVoiceCount() int32 {
    vm.mutex.RLock()
    defer vm.mutex.RUnlock()
    return vm.activeCount
}

// GetVoiceInfo returns current voice info
func (vm *VoiceManager) GetVoiceInfo(supportsOverlapping bool) VoiceInfo {
    vm.mutex.RLock()
    defer vm.mutex.RUnlock()
    
    return CreatePolyphonicVoiceInfo(
        uint32(vm.activeCount),
        uint32(vm.maxVoices),
        supportsOverlapping,
    )
}