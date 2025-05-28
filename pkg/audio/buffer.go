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

// AudioBuffer represents an audio buffer with Go-native types
// This abstracts away all the C interop complexity and provides rich functionality
type AudioBuffer struct {
	// channels contains the audio data as Go slices
	channels [][]float32
	// frameCount is the number of frames in each channel
	frameCount uint32
}

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

// AudioBuffer methods for enhanced buffer functionality

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
			if sample < 0 {
				abs = -sample
			} else {
				abs = sample
			}
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