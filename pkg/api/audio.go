package api

/*
#cgo CFLAGS: -I../../include/clap/include
#include "../../include/clap/include/clap/clap.h"
#include <stdlib.h>
*/
import "C"
import (
	"unsafe"
)

// AudioBuffer represents an audio buffer with Go-native types
// This abstracts away all the C interop complexity
type AudioBuffer struct {
	// channels contains the audio data as Go slices
	channels [][]float32
	// frameCount is the number of frames in each channel
	frameCount uint32
}

// GetChannels returns the audio channels as Go slices
func (ab *AudioBuffer) GetChannels() [][]float32 {
	return ab.channels
}

// GetFrameCount returns the number of frames in the buffer
func (ab *AudioBuffer) GetFrameCount() uint32 {
	return ab.frameCount
}

// GetChannelCount returns the number of channels in the buffer
func (ab *AudioBuffer) GetChannelCount() int {
	return len(ab.channels)
}

// ConvertFromCBuffers converts C audio buffers to Go AudioBuffer
// This hides all unsafe pointer arithmetic from plugin developers
func ConvertFromCBuffers(cBuffers unsafe.Pointer, bufferCount uint32, frameCount uint32) [][]float32 {
	if cBuffers == nil || bufferCount == 0 {
		return nil
	}

	// Convert C pointer to typed pointer
	cAudioBuffers := (*C.clap_audio_buffer_t)(cBuffers)
	
	// Convert C array to Go slice using unsafe pointer arithmetic
	buffers := (*[1024]C.clap_audio_buffer_t)(unsafe.Pointer(cAudioBuffers))[:bufferCount:bufferCount]
	
	result := make([][]float32, 0)
	
	for i := uint32(0); i < bufferCount; i++ {
		buffer := &buffers[i]
		
		// Handle 32-bit float buffers (most common)
		if buffer.data32 != nil {
			channelCount := uint32(buffer.channel_count)
			
			// Convert C channel pointers to Go slices
			channels := (*[64]*C.float)(unsafe.Pointer(buffer.data32))[:channelCount:channelCount]
			
			for ch := uint32(0); ch < channelCount; ch++ {
				if channels[ch] != nil {
					// Convert C float array to Go slice
					channelData := (*[1048576]float32)(unsafe.Pointer(channels[ch]))[:frameCount:frameCount]
					result = append(result, channelData)
				}
			}
		}
		// Note: 64-bit double buffers could be handled here if needed
	}
	
	return result
}

// NewAudioBuffer creates a new AudioBuffer with the specified number of channels and frames
func NewAudioBuffer(channelCount int, frameCount uint32) *AudioBuffer {
	channels := make([][]float32, channelCount)
	for i := range channels {
		channels[i] = make([]float32, frameCount)
	}
	
	return &AudioBuffer{
		channels:   channels,
		frameCount: frameCount,
	}
}

// CopyFrom copies audio data from another AudioBuffer
func (ab *AudioBuffer) CopyFrom(src *AudioBuffer) {
	minChannels := len(ab.channels)
	if len(src.channels) < minChannels {
		minChannels = len(src.channels)
	}
	
	minFrames := ab.frameCount
	if src.frameCount < minFrames {
		minFrames = src.frameCount
	}
	
	for ch := 0; ch < minChannels; ch++ {
		copy(ab.channels[ch][:minFrames], src.channels[ch][:minFrames])
	}
}

// Clear zeros out all audio data in the buffer
func (ab *AudioBuffer) Clear() {
	for ch := range ab.channels {
		for i := range ab.channels[ch] {
			ab.channels[ch][i] = 0.0
		}
	}
}

// ApplyGain applies gain to all channels in the buffer
func (ab *AudioBuffer) ApplyGain(gain float32) {
	for ch := range ab.channels {
		for i := range ab.channels[ch] {
			ab.channels[ch][i] *= gain
		}
	}
}

// Mix mixes audio from another buffer into this buffer
func (ab *AudioBuffer) Mix(src *AudioBuffer, gain float32) {
	minChannels := len(ab.channels)
	if len(src.channels) < minChannels {
		minChannels = len(src.channels)
	}
	
	minFrames := ab.frameCount
	if src.frameCount < minFrames {
		minFrames = src.frameCount
	}
	
	for ch := 0; ch < minChannels; ch++ {
		for i := uint32(0); i < minFrames; i++ {
			ab.channels[ch][i] += src.channels[ch][i] * gain
		}
	}
}

// GetPeakLevel returns the peak level (absolute maximum) for a channel
func (ab *AudioBuffer) GetPeakLevel(channel int) float32 {
	if channel >= len(ab.channels) {
		return 0.0
	}
	
	var peak float32 = 0.0
	for _, sample := range ab.channels[channel] {
		if abs := float32(sample); abs > peak {
			peak = abs
		}
	}
	
	return peak
}

// GetRMSLevel returns the RMS level for a channel
func (ab *AudioBuffer) GetRMSLevel(channel int) float32 {
	if channel >= len(ab.channels) {
		return 0.0
	}
	
	var sum float64 = 0.0
	for _, sample := range ab.channels[channel] {
		sum += float64(sample) * float64(sample)
	}
	
	return float32(sum / float64(len(ab.channels[channel])))
}

// IsSilent returns true if the buffer contains only silence (or near-silence)
func (ab *AudioBuffer) IsSilent(threshold float32) bool {
	for ch := range ab.channels {
		for _, sample := range ab.channels[ch] {
			if sample > threshold || sample < -threshold {
				return false
			}
		}
	}
	return true
}