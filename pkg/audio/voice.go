package audio

import (
	"sync"
)

// Voice represents a single synthesizer voice that can play a note
type Voice struct {
	// Note identification
	NoteID   int32
	Channel  int16
	Key      int16
	Velocity float64
	
	// Oscillator state
	Phase     float64
	Frequency float64
	
	// Envelope
	Envelope *ADSREnvelope
	
	// Modulation
	PitchBend   float64 // In semitones
	Brightness  float64 // 0.0-1.0 for filter cutoff
	Pressure    float64 // Aftertouch/pressure 0.0-1.0
	Volume      float64 // Per-voice volume modulation
	
	// State
	IsActive bool
	TuningID uint64 // ID of tuning to use (0 for equal temperament)
	
	// Custom data for plugin-specific use
	UserData interface{}
}

// VoiceManager manages polyphonic voice allocation and processing
type VoiceManager struct {
	mu         sync.RWMutex
	voices     []*Voice
	maxVoices  int
	sampleRate float64
	
	// Voice stealing strategy
	stealOldest bool
}

// NewVoiceManager creates a new voice manager with the specified polyphony
func NewVoiceManager(maxVoices int, sampleRate float64) *VoiceManager {
	vm := &VoiceManager{
		maxVoices:   maxVoices,
		sampleRate:  sampleRate,
		voices:      make([]*Voice, maxVoices),
		stealOldest: true,
	}
	
	// Pre-allocate voices
	for i := 0; i < maxVoices; i++ {
		vm.voices[i] = &Voice{
			Envelope: NewADSREnvelope(sampleRate),
		}
	}
	
	return vm
}

// SetSampleRate updates the sample rate for all voices
func (vm *VoiceManager) SetSampleRate(sampleRate float64) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	
	vm.sampleRate = sampleRate
	for _, voice := range vm.voices {
		if voice != nil && voice.Envelope != nil {
			voice.Envelope.SampleRate = sampleRate
		}
	}
}

// AllocateVoice finds a free voice or steals one according to the stealing strategy
func (vm *VoiceManager) AllocateVoice(noteID int32, channel, key int16, velocity float64) *Voice {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	
	// First, look for a free voice
	for _, voice := range vm.voices {
		if voice != nil && !voice.IsActive {
			vm.initializeVoice(voice, noteID, channel, key, velocity)
			return voice
		}
	}
	
	// No free voice, need to steal one
	var victimVoice *Voice
	
	if vm.stealOldest {
		// Find the oldest voice (simplest strategy)
		for _, voice := range vm.voices {
			if voice != nil {
				victimVoice = voice
				break
			}
		}
	} else {
		// Could implement other strategies like:
		// - Steal quietest voice
		// - Steal voice in release phase
		// - Steal lowest priority voice
		victimVoice = vm.voices[0]
	}
	
	if victimVoice != nil {
		vm.initializeVoice(victimVoice, noteID, channel, key, velocity)
		return victimVoice
	}
	
	return nil
}

// initializeVoice sets up a voice for a new note
func (vm *VoiceManager) initializeVoice(voice *Voice, noteID int32, channel, key int16, velocity float64) {
	voice.NoteID = noteID
	voice.Channel = channel
	voice.Key = key
	voice.Velocity = velocity
	voice.Phase = 0
	voice.Frequency = NoteToFrequency(int(key))
	voice.IsActive = true
	voice.PitchBend = 0
	voice.Brightness = 1.0
	voice.Pressure = 0
	voice.Volume = 1.0
	voice.TuningID = 0
	
	// Trigger the envelope
	if voice.Envelope != nil {
		voice.Envelope.Trigger()
	}
}

// ReleaseVoice releases a voice by note ID and channel
func (vm *VoiceManager) ReleaseVoice(noteID int32, channel int16) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	
	for _, voice := range vm.voices {
		if voice != nil && voice.IsActive && 
		   voice.NoteID == noteID && voice.Channel == channel {
			if voice.Envelope != nil {
				voice.Envelope.Release()
			}
		}
	}
}

// ReleaseAllVoices releases all active voices
func (vm *VoiceManager) ReleaseAllVoices() {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	
	for _, voice := range vm.voices {
		if voice != nil && voice.IsActive {
			if voice.Envelope != nil {
				voice.Envelope.Release()
			}
		}
	}
}

// ProcessVoices calls the process function for each active voice and returns mixed output
func (vm *VoiceManager) ProcessVoices(frameCount uint32, processFunc func(*Voice, uint32) []float32) []float32 {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	
	// Create output buffer
	output := make([]float32, frameCount)
	
	// Process each active voice
	for _, voice := range vm.voices {
		if voice == nil || !voice.IsActive {
			continue
		}
		
		// Process this voice
		voiceOutput := processFunc(voice, frameCount)
		
		// Mix into output
		for i := uint32(0); i < frameCount && i < uint32(len(voiceOutput)); i++ {
			output[i] += voiceOutput[i]
		}
		
		// Check if voice should be deactivated
		if voice.Envelope != nil && !voice.Envelope.IsActive() {
			voice.IsActive = false
		}
	}
	
	return output
}

// GetActiveVoiceCount returns the number of currently active voices
func (vm *VoiceManager) GetActiveVoiceCount() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	
	count := 0
	for _, voice := range vm.voices {
		if voice != nil && voice.IsActive {
			count++
		}
	}
	return count
}

// GetVoiceByNoteID finds a voice by note ID and channel
func (vm *VoiceManager) GetVoiceByNoteID(noteID int32, channel int16) *Voice {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	
	for _, voice := range vm.voices {
		if voice != nil && voice.IsActive && 
		   voice.NoteID == noteID && voice.Channel == channel {
			return voice
		}
	}
	return nil
}

// ApplyToAllVoices applies a function to all active voices
func (vm *VoiceManager) ApplyToAllVoices(applyFunc func(*Voice)) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	
	for _, voice := range vm.voices {
		if voice != nil && voice.IsActive {
			applyFunc(voice)
		}
	}
}

// SetVoiceStealingStrategy sets whether to steal oldest voice (true) or use other strategies
func (vm *VoiceManager) SetVoiceStealingStrategy(stealOldest bool) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.stealOldest = stealOldest
}

// Reset deactivates all voices and resets their state
func (vm *VoiceManager) Reset() {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	
	for _, voice := range vm.voices {
		if voice != nil {
			voice.IsActive = false
			voice.Phase = 0
			if voice.Envelope != nil {
				voice.Envelope.Reset()
			}
		}
	}
}