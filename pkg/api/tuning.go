package api

// #include "../../include/clap/include/clap/clap.h"
// #include "../../include/clap/include/clap/ext/draft/tuning.h"
// #include <stdlib.h>
// #include <string.h>
//
// // Helper function to get relative tuning
// static double clap_host_tuning_get_relative(const clap_host_t* host, clap_id tuning_id, int32_t channel, int32_t key, uint32_t sample_offset) {
//     if (host && host->get_extension) {
//         const clap_host_tuning_t* tuning = (const clap_host_tuning_t*)host->get_extension(host, CLAP_EXT_TUNING);
//         if (tuning && tuning->get_relative) {
//             return tuning->get_relative(host, tuning_id, channel, key, sample_offset);
//         }
//     }
//     return 0.0; // Equal temperament
// }
//
// // Helper function to check if note should play
// static bool clap_host_tuning_should_play(const clap_host_t* host, clap_id tuning_id, int32_t channel, int32_t key) {
//     if (host && host->get_extension) {
//         const clap_host_tuning_t* tuning = (const clap_host_tuning_t*)host->get_extension(host, CLAP_EXT_TUNING);
//         if (tuning && tuning->should_play) {
//             return tuning->should_play(host, tuning_id, channel, key);
//         }
//     }
//     return true; // Play by default
// }
//
// // Helper function to get tuning count
// static uint32_t clap_host_tuning_get_count(const clap_host_t* host) {
//     if (host && host->get_extension) {
//         const clap_host_tuning_t* tuning = (const clap_host_tuning_t*)host->get_extension(host, CLAP_EXT_TUNING);
//         if (tuning && tuning->get_tuning_count) {
//             return tuning->get_tuning_count(host);
//         }
//     }
//     return 0;
// }
//
// // Helper function to get tuning info
// static bool clap_host_tuning_get_info(const clap_host_t* host, uint32_t tuning_index, clap_tuning_info_t* info) {
//     if (host && host->get_extension && info) {
//         const clap_host_tuning_t* tuning = (const clap_host_tuning_t*)host->get_extension(host, CLAP_EXT_TUNING);
//         if (tuning && tuning->get_info) {
//             return tuning->get_info(host, tuning_index, info);
//         }
//     }
//     return false;
// }
import "C"
import (
	"unsafe"
)

// TuningEvent represents a tuning change event
type TuningEvent struct {
	PortIndex int16   // -1 for global
	Channel   int16   // 0..15, -1 for global
	TuningID  uint64  // ID of the tuning to use
}

// TuningInfo contains information about a tuning
type TuningInfo struct {
	TuningID  uint64
	Name      string
	IsDynamic bool   // True if values may vary with time
}

// TuningProvider is implemented by plugins that respond to tuning changes
type TuningProvider interface {
	// OnTuningChanged is called when a tuning is added or removed from the pool
	OnTuningChanged()
}

// HostTuning provides access to host tuning functionality
type HostTuning struct {
	host unsafe.Pointer
}

// NewHostTuning creates a new host tuning interface
func NewHostTuning(host unsafe.Pointer) *HostTuning {
	return &HostTuning{host: host}
}

// GetRelativeTuning gets the relative tuning in semitones against equal temperament with A4=440Hz
// Returns 0.0 for equal temperament or if tuning not found
func (t *HostTuning) GetRelativeTuning(tuningID uint64, channel, key int32, sampleOffset uint32) float64 {
	if t.host == nil {
		return 0.0
	}
	
	return float64(C.clap_host_tuning_get_relative(
		(*C.clap_host_t)(t.host),
		C.clap_id(tuningID),
		C.int32_t(channel),
		C.int32_t(key),
		C.uint32_t(sampleOffset),
	))
}

// ShouldPlay returns true if the note should be played with the given tuning
func (t *HostTuning) ShouldPlay(tuningID uint64, channel, key int32) bool {
	if t.host == nil {
		return true
	}
	
	return bool(C.clap_host_tuning_should_play(
		(*C.clap_host_t)(t.host),
		C.clap_id(tuningID),
		C.int32_t(channel),
		C.int32_t(key),
	))
}

// GetTuningCount returns the number of tunings in the pool
func (t *HostTuning) GetTuningCount() uint32 {
	if t.host == nil {
		return 0
	}
	
	return uint32(C.clap_host_tuning_get_count((*C.clap_host_t)(t.host)))
}

// GetTuningInfo gets information about a tuning by index
func (t *HostTuning) GetTuningInfo(tuningIndex uint32) (*TuningInfo, bool) {
	if t.host == nil {
		return nil, false
	}
	
	var cInfo C.clap_tuning_info_t
	if !C.clap_host_tuning_get_info((*C.clap_host_t)(t.host), C.uint32_t(tuningIndex), &cInfo) {
		return nil, false
	}
	
	info := &TuningInfo{
		TuningID:  uint64(cInfo.tuning_id),
		Name:      C.GoString(&cInfo.name[0]),
		IsDynamic: bool(cInfo.is_dynamic),
	}
	
	return info, true
}

// GetAllTunings returns all available tunings
func (t *HostTuning) GetAllTunings() []TuningInfo {
	count := t.GetTuningCount()
	if count == 0 {
		return nil
	}
	
	tunings := make([]TuningInfo, 0, count)
	for i := uint32(0); i < count; i++ {
		if info, ok := t.GetTuningInfo(i); ok {
			tunings = append(tunings, *info)
		}
	}
	
	return tunings
}

// ApplyTuning applies tuning to a frequency in Hz
// baseFreq is the equal temperament frequency, returns the tuned frequency
func (t *HostTuning) ApplyTuning(baseFreq float64, tuningID uint64, channel, key int32, sampleOffset uint32) float64 {
	// Get relative tuning in semitones
	relativeSemitones := t.GetRelativeTuning(tuningID, channel, key, sampleOffset)
	
	// Convert semitones to frequency ratio: 2^(semitones/12)
	ratio := Pow2(relativeSemitones / 12.0)
	
	return baseFreq * ratio
}

// Pow2 calculates 2^x efficiently
func Pow2(x float64) float64 {
	// For small values, use approximation
	if x >= -0.5 && x <= 0.5 {
		// Taylor series approximation of 2^x around 0
		const ln2 = 0.693147180559945309417
		return 1.0 + x*ln2 + (x*x*ln2*ln2)/2.0
	}
	
	// For larger values, decompose into integer and fractional parts
	intPart := int(x)
	fracPart := x - float64(intPart)
	
	// Calculate 2^fracPart using approximation
	const ln2 = 0.693147180559945309417
	pow2Frac := 1.0 + fracPart*ln2 + (fracPart*fracPart*ln2*ln2)/2.0
	
	// Combine with bit shifting for integer part
	if intPart >= 0 {
		return pow2Frac * float64(uint64(1)<<uint(intPart))
	} else {
		return pow2Frac / float64(uint64(1)<<uint(-intPart))
	}
}