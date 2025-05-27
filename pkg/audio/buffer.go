package audio

import (
	"errors"
	"math"
)

// Common errors
var (
	ErrInvalidBuffer      = errors.New("invalid buffer")
	ErrChannelMismatch    = errors.New("channel count mismatch")
	ErrFrameCountMismatch = errors.New("frame count mismatch")
	ErrBufferTooSmall     = errors.New("buffer too small")
)

// Buffer represents multi-channel audio data
type Buffer [][]float32

// NewBuffer creates a new audio buffer with the given dimensions
func NewBuffer(channels, frames int) Buffer {
	buf := make(Buffer, channels)
	for i := range buf {
		buf[i] = make([]float32, frames)
	}
	return buf
}

// Channels returns the number of channels
func (b Buffer) Channels() int {
	return len(b)
}

// Frames returns the number of frames (samples per channel)
func (b Buffer) Frames() int {
	if len(b) == 0 {
		return 0
	}
	return len(b[0])
}

// Clear sets all samples to zero
func (b Buffer) Clear() {
	for ch := range b {
		for i := range b[ch] {
			b[ch][i] = 0
		}
	}
}

// ClearRange clears a range of samples
func (b Buffer) ClearRange(start, end int) error {
	frames := b.Frames()
	if start < 0 || end > frames || start >= end {
		return ErrInvalidBuffer
	}
	
	for ch := range b {
		for i := start; i < end; i++ {
			b[ch][i] = 0
		}
	}
	return nil
}

// Copy copies samples from source to destination
func Copy(dst, src Buffer) error {
	if dst.Channels() != src.Channels() {
		return ErrChannelMismatch
	}
	
	if dst.Frames() != src.Frames() {
		return ErrFrameCountMismatch
	}
	
	for ch := range dst {
		copy(dst[ch], src[ch])
	}
	
	return nil
}

// CopyMono copies a mono buffer to all channels of a multi-channel buffer
func CopyMono(dst Buffer, src []float32) error {
	if dst.Frames() != len(src) {
		return ErrFrameCountMismatch
	}
	
	for ch := range dst {
		copy(dst[ch], src)
	}
	
	return nil
}

// ApplyGain applies a gain factor to the buffer
func ApplyGain(buf Buffer, gain float32) {
	for ch := range buf {
		for i := range buf[ch] {
			buf[ch][i] *= gain
		}
	}
}

// ApplyGainRange applies gain to a range of samples
func ApplyGainRange(buf Buffer, gain float32, start, end int) error {
	frames := buf.Frames()
	if start < 0 || end > frames || start >= end {
		return ErrInvalidBuffer
	}
	
	for ch := range buf {
		for i := start; i < end; i++ {
			buf[ch][i] *= gain
		}
	}
	
	return nil
}

// Mix mixes source buffer into destination with gain
func Mix(dst, src Buffer, gain float32) error {
	if dst.Channels() != src.Channels() {
		return ErrChannelMismatch
	}
	
	if dst.Frames() != src.Frames() {
		return ErrFrameCountMismatch
	}
	
	for ch := range dst {
		for i := range dst[ch] {
			dst[ch][i] += src[ch][i] * gain
		}
	}
	
	return nil
}


// GetPeak returns the peak (maximum absolute) value in the buffer
func GetPeak(buf Buffer) float32 {
	var peak float32
	
	for ch := range buf {
		for i := range buf[ch] {
			abs := float32(math.Abs(float64(buf[ch][i])))
			if abs > peak {
				peak = abs
			}
		}
	}
	
	return peak
}

// GetRMS returns the RMS (root mean square) value of the buffer
func GetRMS(buf Buffer) float32 {
	var sum float64
	totalSamples := 0
	
	for ch := range buf {
		for i := range buf[ch] {
			val := float64(buf[ch][i])
			sum += val * val
			totalSamples++
		}
	}
	
	if totalSamples == 0 {
		return 0
	}
	
	return float32(math.Sqrt(sum / float64(totalSamples)))
}

// Normalize normalizes the buffer to the given peak level
func Normalize(buf Buffer, targetPeak float32) {
	currentPeak := GetPeak(buf)
	if currentPeak == 0 {
		return
	}
	
	gain := targetPeak / currentPeak
	ApplyGain(buf, gain)
}