package main

// This example demonstrates how to implement the Thread Pool extension
// in a ClapGo plugin. This extension allows plugins to use the host's
// thread pool for parallel processing.

import "C"
import (
	"github.com/justyntemme/clapgo/pkg/api"
	"math"
	"unsafe"
)

// Example multi-channel effect plugin using thread pool
type MultiChannelEffect struct {
	api.BasePlugin
	
	// Thread pool helper
	threadPool *api.ThreadPoolHelper
	
	// Audio buffers for processing
	inputBuffers  [][]float32
	outputBuffers [][]float32
	
	// Processing parameters
	gain float32
}

// NewMultiChannelEffect creates a new multi-channel effect
func NewMultiChannelEffect() *MultiChannelEffect {
	return &MultiChannelEffect{
		gain: 1.0,
	}
}

// Initialize sets up the thread pool
func (p *MultiChannelEffect) Initialize(host unsafe.Pointer) bool {
	p.threadPool = api.NewThreadPoolHelper(host, p)
	return true
}

// Process audio using thread pool for parallel channel processing
func (p *MultiChannelEffect) Process(process *api.ProcessContext) api.ProcessStatus {
	// Store buffer references for thread pool access
	p.inputBuffers = process.AudioInputs
	p.outputBuffers = process.AudioOutputs
	
	// Process all channels in parallel
	numChannels := uint32(len(p.outputBuffers))
	if numChannels > 0 {
		p.threadPool.Execute(numChannels)
	}
	
	return api.ProcessContinue
}

// Implement ThreadPoolProvider interface
func (p *MultiChannelEffect) Exec(taskIndex uint32) {
	// Each task processes one channel
	channelIndex := taskIndex
	
	if int(channelIndex) >= len(p.outputBuffers) {
		return
	}
	
	output := p.outputBuffers[channelIndex]
	
	// Check if we have corresponding input
	if int(channelIndex) < len(p.inputBuffers) {
		input := p.inputBuffers[channelIndex]
		
		// Apply gain to each sample
		for i := range output {
			if i < len(input) {
				output[i] = input[i] * p.gain
			}
		}
	} else {
		// No input for this output channel, clear it
		for i := range output {
			output[i] = 0
		}
	}
}

// Example polyphonic synthesizer using thread pool for voice processing
type PolyphonicSynth struct {
	api.BasePlugin
	
	// Thread pool helper
	threadPool *api.ThreadPoolHelper
	
	// Voice management
	voices      []Voice
	activeVoices uint32
	
	// Output buffer
	outputBuffer []float32
	
	// Temporary buffers for each voice
	voiceBuffers [][]float32
}

// Voice represents a single synthesizer voice
type Voice struct {
	active    bool
	noteID    int32
	frequency float64
	phase     float64
	amplitude float32
	envelope  float32
}

// NewPolyphonicSynth creates a new polyphonic synthesizer
func NewPolyphonicSynth() *PolyphonicSynth {
	const maxVoices = 32
	synth := &PolyphonicSynth{
		voices:       make([]Voice, maxVoices),
		voiceBuffers: make([][]float32, maxVoices),
	}
	
	// Pre-allocate voice buffers
	for i := range synth.voiceBuffers {
		synth.voiceBuffers[i] = make([]float32, 0) // Will be resized as needed
	}
	
	return synth
}

// Initialize sets up the thread pool
func (s *PolyphonicSynth) Initialize(host unsafe.Pointer) bool {
	s.threadPool = api.NewThreadPoolHelper(host, s)
	return true
}

// Process synthesizes all active voices in parallel
func (s *PolyphonicSynth) Process(process *api.ProcessContext) api.ProcessStatus {
	frameCount := process.FrameCount
	
	// Ensure output buffer is large enough
	if len(s.outputBuffer) < int(frameCount) {
		s.outputBuffer = make([]float32, frameCount)
	}
	
	// Clear output buffer
	for i := range s.outputBuffer[:frameCount] {
		s.outputBuffer[i] = 0
	}
	
	// Count active voices
	s.activeVoices = 0
	for i := range s.voices {
		if s.voices[i].active {
			// Ensure voice buffer is large enough
			if len(s.voiceBuffers[s.activeVoices]) < int(frameCount) {
				s.voiceBuffers[s.activeVoices] = make([]float32, frameCount)
			}
			s.activeVoices++
		}
	}
	
	// Process all active voices in parallel
	if s.activeVoices > 0 {
		s.threadPool.Execute(s.activeVoices)
		
		// Mix all voice outputs
		for v := uint32(0); v < s.activeVoices; v++ {
			for i := uint32(0); i < frameCount; i++ {
				s.outputBuffer[i] += s.voiceBuffers[v][i]
			}
		}
	}
	
	// Copy to output channels
	for ch := range process.AudioOutputs {
		output := process.AudioOutputs[ch]
		for i := uint32(0); i < frameCount && i < uint32(len(output)); i++ {
			output[i] = s.outputBuffer[i]
		}
	}
	
	return api.ProcessContinue
}

// Implement ThreadPoolProvider interface
func (s *PolyphonicSynth) Exec(taskIndex uint32) {
	// Find the voice to process
	voiceIndex := 0
	activeCount := uint32(0)
	
	for i := range s.voices {
		if s.voices[i].active {
			if activeCount == taskIndex {
				voiceIndex = i
				break
			}
			activeCount++
		}
	}
	
	voice := &s.voices[voiceIndex]
	buffer := s.voiceBuffers[taskIndex]
	sampleRate := 48000.0 // Would come from activation
	
	// Simple sine wave synthesis
	phaseIncrement := 2.0 * math.Pi * voice.frequency / sampleRate
	
	for i := range buffer {
		// Generate sample
		sample := float32(math.Sin(voice.phase)) * voice.amplitude * voice.envelope
		buffer[i] = sample
		
		// Update phase
		voice.phase += phaseIncrement
		if voice.phase >= 2.0*math.Pi {
			voice.phase -= 2.0 * math.Pi
		}
		
		// Update envelope (simple decay)
		voice.envelope *= 0.9999
		if voice.envelope < 0.001 {
			voice.active = false
		}
	}
}

// Example using ParallelProcessor for convenience
type ConvolutionReverb struct {
	api.BasePlugin
	
	// Parallel processor
	parallel *api.ParallelProcessor
	
	// Impulse response data
	impulseResponse []float32
	
	// Convolution buffers per channel
	convolutionBuffers [][]float32
}

// Process uses parallel processing for convolution
func (r *ConvolutionReverb) Process(process *api.ProcessContext) api.ProcessStatus {
	// Process each channel's convolution in parallel
	numChannels := uint32(len(process.AudioOutputs))
	
	r.parallel.ProcessChannels(numChannels, func(channelIndex uint32) {
		// Perform convolution for this channel
		r.processChannelConvolution(channelIndex, process)
	})
	
	return api.ProcessContinue
}

func (r *ConvolutionReverb) processChannelConvolution(channelIndex uint32, process *api.ProcessContext) {
	// Convolution processing for one channel
	// This is where the computationally expensive convolution would happen
}

// Required export for the extension

//export ClapGo_PluginThreadPoolExec
func ClapGo_PluginThreadPoolExec(plugin unsafe.Pointer, taskIndex C.uint32_t) {
	if plugin == nil {
		return
	}
	
	// Type assertion based on plugin type
	if provider, ok := plugin.(api.ThreadPoolProvider); ok {
		provider.Exec(uint32(taskIndex))
	}
}

// Other required plugin exports would go here...

func main() {
	// Required for c-shared build
}