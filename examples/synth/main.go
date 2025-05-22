package main

import (
	"encoding/json"
	"math"
	"os"
	"sync/atomic"
	"unsafe"
	
	"github.com/justyntemme/clapgo/pkg/api"
)

// SynthPlugin implements the SimplePluginInterface
// NO CGO required - all complexity handled by pkg/api
type SynthPlugin struct {
	// Plugin state (using atomic values for thread safety)
	voices       [16]*Voice  // Maximum 16 voices
	sampleRate   int64       // atomic storage
	volume       int64       // atomic storage for volume (0.0-1.0)
	waveform     int64       // atomic storage for waveform (0-2)
	attack       int64       // atomic storage for attack time
	decay        int64       // atomic storage for decay time  
	sustain      int64       // atomic storage for sustain level
	release      int64       // atomic storage for release time
}

// Voice represents a single active note
type Voice struct {
	NoteID       int32
	Channel      int16
	Key          int16
	Velocity     float64
	Phase        float64
	IsActive     bool
	ReleasePhase float64
}

// GetInfo returns plugin metadata
func (p *SynthPlugin) GetInfo() api.PluginInfo {
	return api.PluginInfo{
		ID:          PluginID,
		Name:        PluginName,
		Vendor:      PluginVendor,
		URL:         "https://github.com/justyntemme/clapgo",
		ManualURL:   "https://github.com/justyntemme/clapgo",
		SupportURL:  "https://github.com/justyntemme/clapgo/issues",
		Version:     PluginVersion,
		Description: PluginDescription,
		Features:    []string{"instrument", "synthesizer", "stereo"},
	}
}

// Initialize sets up the plugin with sample rate
func (p *SynthPlugin) Initialize(sampleRate float64) error {
	// Set default values atomically
	atomic.StoreInt64(&p.sampleRate, int64(floatToBits(sampleRate)))
	atomic.StoreInt64(&p.volume, int64(floatToBits(0.7)))      // -3dB
	atomic.StoreInt64(&p.waveform, 0)                          // sine
	atomic.StoreInt64(&p.attack, int64(floatToBits(0.01)))     // 10ms
	atomic.StoreInt64(&p.decay, int64(floatToBits(0.1)))       // 100ms
	atomic.StoreInt64(&p.sustain, int64(floatToBits(0.7)))     // 70%
	atomic.StoreInt64(&p.release, int64(floatToBits(0.3)))     // 300ms
	
	// Clear all voices
	for i := range p.voices {
		p.voices[i] = nil
	}
	
	return nil
}

// ProcessAudio processes audio using Go-native types
func (p *SynthPlugin) ProcessAudio(input, output [][]float32, frameCount uint32) error {
	// Get current parameter values atomically
	sampleRate := floatFromBits(uint64(atomic.LoadInt64(&p.sampleRate)))
	volume := floatFromBits(uint64(atomic.LoadInt64(&p.volume)))
	waveform := int(atomic.LoadInt64(&p.waveform))
	attack := floatFromBits(uint64(atomic.LoadInt64(&p.attack)))
	decay := floatFromBits(uint64(atomic.LoadInt64(&p.decay)))
	sustain := floatFromBits(uint64(atomic.LoadInt64(&p.sustain)))
	release := floatFromBits(uint64(atomic.LoadInt64(&p.release)))
	
	// If no outputs, nothing to do
	if len(output) == 0 {
		return nil
	}
	
	// Get the number of output channels
	numChannels := len(output)
	
	// Clear the output buffer
	for ch := 0; ch < numChannels; ch++ {
		outChannel := output[ch]
		if len(outChannel) < int(frameCount) {
			continue
		}
		
		for i := uint32(0); i < frameCount; i++ {
			outChannel[i] = 0.0
		}
	}
	
	// Process each voice
	for i, voice := range p.voices {
		if voice != nil && voice.IsActive {
			// Calculate frequency for this note
			freq := noteToFrequency(int(voice.Key))
			
			// Generate audio for this voice
			for j := uint32(0); j < frameCount; j++ {
				// Get envelope value
				env := getEnvelopeValue(voice, j, frameCount, sampleRate, attack, decay, sustain, release)
				
				// Generate sample
				sample := generateSample(voice.Phase, freq, waveform) * env * voice.Velocity
				
				// Apply master volume
				sample *= volume
				
				// Add to all output channels
				for ch := 0; ch < numChannels; ch++ {
					if len(output[ch]) > int(j) {
						output[ch][j] += float32(sample)
					}
				}
				
				// Advance oscillator phase
				voice.Phase += freq / sampleRate
				if voice.Phase >= 1.0 {
					voice.Phase -= 1.0
				}
			}
			
			// Check if voice is still active
			if voice.ReleasePhase >= 1.0 {
				p.voices[i] = nil
			}
		}
	}
	
	return nil
}

// GetParameters returns all parameter definitions
func (p *SynthPlugin) GetParameters() []api.ParamInfo {
	return []api.ParamInfo{
		api.CreateFloatParameter(1, "Volume", 0.0, 1.0, 0.7),
		api.CreateFloatParameter(2, "Waveform", 0.0, 2.0, 0.0),
		api.CreateFloatParameter(3, "Attack", 0.001, 2.0, 0.01),
		api.CreateFloatParameter(4, "Decay", 0.001, 2.0, 0.1),
		api.CreateFloatParameter(5, "Sustain", 0.0, 1.0, 0.7),
		api.CreateFloatParameter(6, "Release", 0.001, 5.0, 0.3),
	}
}

// SetParameterValue sets a parameter value
func (p *SynthPlugin) SetParameterValue(paramID uint32, value float64) error {
	switch paramID {
	case 1: // Volume
		if value < 0.0 { value = 0.0 }
		if value > 1.0 { value = 1.0 }
		atomic.StoreInt64(&p.volume, int64(floatToBits(value)))
	case 2: // Waveform
		if value < 0.0 { value = 0.0 }
		if value > 2.0 { value = 2.0 }
		atomic.StoreInt64(&p.waveform, int64(math.Round(value)))
	case 3: // Attack
		if value < 0.001 { value = 0.001 }
		if value > 2.0 { value = 2.0 }
		atomic.StoreInt64(&p.attack, int64(floatToBits(value)))
	case 4: // Decay
		if value < 0.001 { value = 0.001 }
		if value > 2.0 { value = 2.0 }
		atomic.StoreInt64(&p.decay, int64(floatToBits(value)))
	case 5: // Sustain
		if value < 0.0 { value = 0.0 }
		if value > 1.0 { value = 1.0 }
		atomic.StoreInt64(&p.sustain, int64(floatToBits(value)))
	case 6: // Release
		if value < 0.001 { value = 0.001 }
		if value > 5.0 { value = 5.0 }
		atomic.StoreInt64(&p.release, int64(floatToBits(value)))
	default:
		return api.ErrInvalidParam
	}
	return nil
}

// GetParameterValue gets a parameter value
func (p *SynthPlugin) GetParameterValue(paramID uint32) float64 {
	switch paramID {
	case 1: // Volume
		return floatFromBits(uint64(atomic.LoadInt64(&p.volume)))
	case 2: // Waveform
		return float64(atomic.LoadInt64(&p.waveform))
	case 3: // Attack
		return floatFromBits(uint64(atomic.LoadInt64(&p.attack)))
	case 4: // Decay
		return floatFromBits(uint64(atomic.LoadInt64(&p.decay)))
	case 5: // Sustain
		return floatFromBits(uint64(atomic.LoadInt64(&p.sustain)))
	case 6: // Release
		return floatFromBits(uint64(atomic.LoadInt64(&p.release)))
	}
	return 0.0
}

// OnNoteOn handles note on events
func (p *SynthPlugin) OnNoteOn(noteID int32, channel, key int16, velocity float64) {
	// Validate note event fields (CLAP spec says key must be 0-127)
	if key < 0 || key > 127 {
		return
	}
	
	// Ensure velocity is positive
	if velocity <= 0 {
		velocity = 0.01 // Very quiet but not silent
	}
	
	// Find a free voice slot or steal an existing one
	voiceIndex := p.findFreeVoice()
	
	// Create a new voice with validated data
	p.voices[voiceIndex] = &Voice{
		NoteID:       noteID,
		Channel:      channel,
		Key:          key,
		Velocity:     velocity,
		Phase:        0.0,
		IsActive:     true,
		ReleasePhase: -1.0, // Not in release phase
	}
}

// OnNoteOff handles note off events
func (p *SynthPlugin) OnNoteOff(noteID int32, channel, key int16) {
	// Find the voice with this note ID or key/channel combination
	for _, voice := range p.voices {
		// Safety check
		if voice == nil || !voice.IsActive {
			continue
		}
		
		// Match on note ID if provided (non-negative), otherwise match on key and channel
		if (noteID >= 0 && voice.NoteID == noteID) ||
		   (noteID < 0 && voice.Key == key && 
		    (channel < 0 || voice.Channel == channel)) {
			// Start the release phase (0.0 = start of release)
			voice.ReleasePhase = 0.0
		}
	}
}

// OnActivate is called when plugin is activated
func (p *SynthPlugin) OnActivate() error {
	return nil
}

// OnDeactivate is called when plugin is deactivated
func (p *SynthPlugin) OnDeactivate() {
	// Nothing to do
}

// Cleanup releases any resources
func (p *SynthPlugin) Cleanup() {
	// Clear all voices
	for i := range p.voices {
		p.voices[i] = nil
	}
}

// findFreeVoice finds a free voice slot or steals an existing one
func (p *SynthPlugin) findFreeVoice() int {
	// First, look for an empty slot
	for i, voice := range p.voices {
		if voice == nil {
			return i
		}
	}
	
	// If no empty slots, find the voice in release phase with the most progress
	bestIndex := 0
	bestReleasePhase := -1.0
	
	for i, voice := range p.voices {
		if voice != nil && voice.ReleasePhase >= 0.0 && voice.ReleasePhase > bestReleasePhase {
			bestIndex = i
			bestReleasePhase = voice.ReleasePhase
		}
	}
	
	// If we found a voice in release phase, use that
	if bestReleasePhase >= 0.0 {
		return bestIndex
	}
	
	// Otherwise, just take the first slot (could implement smarter voice stealing)
	return 0
}

// getEnvelopeValue calculates the ADSR envelope value for a voice
func getEnvelopeValue(voice *Voice, sampleIndex, frameCount uint32, sampleRate, attack, decay, sustain, release float64) float64 {
	// If in release phase
	if voice.ReleasePhase >= 0.0 {
		// Update release phase
		releaseSamples := release * sampleRate
		voice.ReleasePhase += 1.0 / releaseSamples
		
		// Calculate release envelope (exponential decay)
		return math.Pow(1.0 - voice.ReleasePhase, 2.0) * sustain
	}
	
	// Attack phase
	attackSamples := attack * sampleRate
	if attackSamples <= 0 {
		attackSamples = 1 // Prevent division by zero
	}
	
	// Decay phase
	decaySamples := decay * sampleRate
	if decaySamples <= 0 {
		decaySamples = 1 // Prevent division by zero
	}
	
	// Calculate elapsed time for this voice
	elapsedSamples := sampleIndex // Simplified version
	
	// Attack phase
	if elapsedSamples < uint32(attackSamples) {
		return float64(elapsedSamples) / attackSamples
	}
	
	// Decay phase
	if elapsedSamples < uint32(attackSamples+decaySamples) {
		decayProgress := float64(elapsedSamples-uint32(attackSamples)) / decaySamples
		return 1.0 - decayProgress*(1.0-sustain)
	}
	
	// Sustain phase
	return sustain
}

// generateSample generates a single sample based on the current waveform
func generateSample(phase, freq float64, waveform int) float64 {
	switch waveform {
	case 0: // Sine wave
		return math.Sin(2.0 * math.Pi * phase)
		
	case 1: // Saw wave
		return 2.0*phase - 1.0
		
	case 2: // Square wave
		if phase < 0.5 {
			return 1.0
		}
		return -1.0
		
	default:
		return 0.0
	}
}

// noteToFrequency converts a MIDI note number to a frequency
func noteToFrequency(note int) float64 {
	return 440.0 * math.Pow(2.0, (float64(note)-69.0)/12.0)
}

// SaveState returns custom state data for the plugin
func (p *SynthPlugin) SaveState() ([]byte, error) {
	// Save any additional state beyond parameters
	state := map[string]interface{}{
		"plugin_version": "1.0.0",
		"waveform":      atomic.LoadInt64(&p.waveform),
		"attack":        floatFromBits(uint64(atomic.LoadInt64(&p.attack))),
		"decay":         floatFromBits(uint64(atomic.LoadInt64(&p.decay))),
		"sustain":       floatFromBits(uint64(atomic.LoadInt64(&p.sustain))),
		"release":       floatFromBits(uint64(atomic.LoadInt64(&p.release))),
	}
	
	return json.Marshal(state)
}

// LoadState loads custom state data for the plugin
func (p *SynthPlugin) LoadState(data []byte) error {
	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}
	
	// Load waveform
	if waveform, ok := state["waveform"].(float64); ok {
		atomic.StoreInt64(&p.waveform, int64(waveform))
	}
	
	// Load ADSR
	if attack, ok := state["attack"].(float64); ok {
		atomic.StoreInt64(&p.attack, int64(floatToBits(attack)))
	}
	if decay, ok := state["decay"].(float64); ok {
		atomic.StoreInt64(&p.decay, int64(floatToBits(decay)))
	}
	if sustain, ok := state["sustain"].(float64); ok {
		atomic.StoreInt64(&p.sustain, int64(floatToBits(sustain)))
	}
	if release, ok := state["release"].(float64); ok {
		atomic.StoreInt64(&p.release, int64(floatToBits(release)))
	}
	
	return nil
}

// LoadPreset loads a preset from a file
func (p *SynthPlugin) LoadPreset(filePath string) error {
	// Open the preset file
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Parse the preset JSON
	var preset struct {
		// Basic preset metadata
		Name        string   `json:"name"`
		Description string   `json:"description"`
		
		// Specific synth parameters
		Waveform     int     `json:"waveform"`
		Attack       float64 `json:"attack"`
		Decay        float64 `json:"decay"`
		Sustain      float64 `json:"sustain"`
		Release      float64 `json:"release"`
		
		// Custom state data
		StateData    map[string]interface{} `json:"state_data"`
	}
	
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&preset); err != nil {
		return err
	}
	
	// Update synth parameters
	atomic.StoreInt64(&p.waveform, int64(preset.Waveform))
	atomic.StoreInt64(&p.attack, int64(floatToBits(preset.Attack)))
	atomic.StoreInt64(&p.decay, int64(floatToBits(preset.Decay)))
	atomic.StoreInt64(&p.sustain, int64(floatToBits(preset.Sustain)))
	atomic.StoreInt64(&p.release, int64(floatToBits(preset.Release)))
	
	return nil
}

// Helper functions for atomic float64 operations
func floatToBits(f float64) uint64 {
	return *(*uint64)(unsafe.Pointer(&f))
}

func floatFromBits(b uint64) float64 {
	return *(*float64)(unsafe.Pointer(&b))
}

func init() {
	// Register the plugin during library initialization
	plugin := &SynthPlugin{}
	api.RegisterSimplePlugin(plugin)
}

func main() {
	// This is only called when run as standalone executable
	// Plugin registration happens in init() for shared library loading
}