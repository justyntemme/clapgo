package audio

// ProcessHelpers provides common audio processing utilities

// ProcessWithGain processes audio buffers with a gain value from input to output
func ProcessWithGain(out, in [][]float32, gain float32) {
	numChannels := len(out)
	if numChannels > len(in) {
		numChannels = len(in)
	}
	
	for ch := 0; ch < numChannels; ch++ {
		ApplyGainToChannel(out[ch], in[ch], gain)
	}
}

// ApplyGainToChannel applies gain to a single channel
func ApplyGainToChannel(out, in []float32, gain float32) {
	if len(out) != len(in) {
		// Handle size mismatch - process minimum
		minLen := len(out)
		if len(in) < minLen {
			minLen = len(in)
		}
		for i := 0; i < minLen; i++ {
			out[i] = in[i] * gain
		}
		// Zero remaining samples if out is longer
		for i := minLen; i < len(out); i++ {
			out[i] = 0
		}
	} else {
		// Optimized path for matching sizes
		for i := range out {
			out[i] = in[i] * gain
		}
	}
}

// CopyAudio copies audio from input to output buffers
func CopyAudio(out, in [][]float32) {
	numChannels := len(out)
	if numChannels > len(in) {
		numChannels = len(in)
	}
	
	for ch := 0; ch < numChannels; ch++ {
		copy(out[ch], in[ch])
	}
}

// ClearAudio zeroes out audio buffers
func ClearAudio(buffers [][]float32) {
	for ch := range buffers {
		for i := range buffers[ch] {
			buffers[ch][i] = 0
		}
	}
}

// MixAudio adds input to output buffers
func MixAudio(out, in [][]float32, gain float32) {
	numChannels := len(out)
	if numChannels > len(in) {
		numChannels = len(in)
	}
	
	for ch := 0; ch < numChannels; ch++ {
		MixChannel(out[ch], in[ch], gain)
	}
}

// MixChannel adds input to output for a single channel
func MixChannel(out, in []float32, gain float32) {
	minLen := len(out)
	if len(in) < minLen {
		minLen = len(in)
	}
	
	for i := 0; i < minLen; i++ {
		out[i] += in[i] * gain
	}
}

// ProcessInPlace applies gain to buffers in place
func ProcessInPlace(buffers [][]float32, gain float32) {
	for ch := range buffers {
		for i := range buffers[ch] {
			buffers[ch][i] *= gain
		}
	}
}

// ValidateBuffers checks if audio buffers are valid for processing
func ValidateBuffers(out, in [][]float32) bool {
	if len(out) == 0 || len(in) == 0 {
		return false
	}
	
	// Check that all channels have the same frame count
	if len(out) > 0 && len(in) > 0 {
		frameCount := len(out[0])
		
		// Check output channels
		for ch := range out {
			if len(out[ch]) != frameCount {
				return false
			}
		}
		
		// Check input channels
		for ch := range in {
			if len(in[ch]) != frameCount {
				return false
			}
		}
	}
	
	return true
}