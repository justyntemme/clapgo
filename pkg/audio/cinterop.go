package audio

// #include "../../include/clap/include/clap/clap.h"
import "C"
import (
	"unsafe"
)

// ConvertFromCBuffers converts C audio buffers to Go Buffer
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