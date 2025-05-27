package audio

import (
	"math"
)

// Common DSP constants
const (
	// Minimum gain in linear scale to avoid -inf dB
	MinGainLinear = 1e-6
	
	// Reference level for dB conversions
	RefLevel = 1.0
)

// LinearToDb converts linear gain to decibels
func LinearToDb(linear float64) float64 {
	if linear <= MinGainLinear {
		return -120.0 // Practical -inf
	}
	return 20.0 * math.Log10(linear)
}

// DbToLinear converts decibels to linear gain
func DbToLinear(db float64) float64 {
	return math.Pow(10.0, db/20.0)
}

// LinearToDbFloat32 converts linear gain to decibels (float32 version)
func LinearToDbFloat32(linear float32) float32 {
	return float32(LinearToDb(float64(linear)))
}

// DbToLinearFloat32 converts decibels to linear gain (float32 version)
func DbToLinearFloat32(db float32) float32 {
	return float32(DbToLinear(float64(db)))
}

// Pan calculates left and right gains for a pan position
// pan: -1.0 (full left) to 1.0 (full right)
// Returns: leftGain, rightGain
func Pan(pan float32) (float32, float32) {
	// Constant power panning
	angle := float64(pan) * math.Pi / 4.0 // -45° to +45°
	leftGain := float32(math.Cos(angle + math.Pi/4.0))
	rightGain := float32(math.Sin(angle + math.Pi/4.0))
	return leftGain, rightGain
}

// ApplyPan applies panning to a stereo buffer
func ApplyPan(buf Buffer, pan float32) error {
	if buf.Channels() != 2 {
		return ErrChannelMismatch
	}
	
	leftGain, rightGain := Pan(pan)
	
	for i := range buf[0] {
		buf[0][i] *= leftGain
		buf[1][i] *= rightGain
	}
	
	return nil
}

// StereoToMono converts stereo to mono by averaging channels
func StereoToMono(dst []float32, src Buffer) error {
	if src.Channels() != 2 {
		return ErrChannelMismatch
	}
	
	if len(dst) != src.Frames() {
		return ErrFrameCountMismatch
	}
	
	for i := range dst {
		dst[i] = (src[0][i] + src[1][i]) * 0.5
	}
	
	return nil
}

// MonoToStereo converts mono to stereo by duplicating the channel
func MonoToStereo(dst Buffer, src []float32) error {
	if dst.Channels() != 2 {
		return ErrChannelMismatch
	}
	
	if dst.Frames() != len(src) {
		return ErrFrameCountMismatch
	}
	
	for i := range src {
		dst[0][i] = src[i]
		dst[1][i] = src[i]
	}
	
	return nil
}

// Clip limits samples to the range [-limit, limit]
func Clip(buf Buffer, limit float32) {
	for ch := range buf {
		for i := range buf[ch] {
			if buf[ch][i] > limit {
				buf[ch][i] = limit
			} else if buf[ch][i] < -limit {
				buf[ch][i] = -limit
			}
		}
	}
}

// SoftClip applies soft clipping (tanh) to the buffer
func SoftClip(buf Buffer) {
	for ch := range buf {
		for i := range buf[ch] {
			buf[ch][i] = float32(math.Tanh(float64(buf[ch][i])))
		}
	}
}

// Fade applies a linear fade in or out
// fadeIn: true for fade in, false for fade out
func Fade(buf Buffer, fadeIn bool) {
	frames := buf.Frames()
	if frames == 0 {
		return
	}
	
	for ch := range buf {
		for i := range buf[ch] {
			var gain float32
			if fadeIn {
				gain = float32(i) / float32(frames-1)
			} else {
				gain = float32(frames-1-i) / float32(frames-1)
			}
			buf[ch][i] *= gain
		}
	}
}

// CrossFade performs a crossfade between two buffers
func CrossFade(dst, from, to Buffer) error {
	if dst.Channels() != from.Channels() || dst.Channels() != to.Channels() {
		return ErrChannelMismatch
	}
	
	if dst.Frames() != from.Frames() || dst.Frames() != to.Frames() {
		return ErrFrameCountMismatch
	}
	
	frames := dst.Frames()
	for ch := range dst {
		for i := range dst[ch] {
			fadeOut := float32(frames-1-i) / float32(frames-1)
			fadeIn := float32(i) / float32(frames-1)
			dst[ch][i] = from[ch][i]*fadeOut + to[ch][i]*fadeIn
		}
	}
	
	return nil
}