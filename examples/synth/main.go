package main

// #cgo CFLAGS: -I../../include/clap/include
// #include "../../include/clap/include/clap/clap.h"
// #include <stdlib.h>
import "C"
import (
	"encoding/json"
	"github.com/justyntemme/clapgo/pkg/api"
	"math"
	"os"
	"runtime/cgo"
	"unsafe"
)

// This example demonstrates a simple synthesizer plugin using the new API interfaces.

// Export Go plugin functionality
var (
	synthPlugin *SynthPlugin
)

func init() {
	// Create our synth plugin
	synthPlugin = NewSynthPlugin()
}

// Standardized exports for manifest system

//export ClapGo_CreatePlugin
func ClapGo_CreatePlugin(host unsafe.Pointer, pluginID *C.char) unsafe.Pointer {
	id := C.GoString(pluginID)
	if id == PluginID {
		handle := cgo.NewHandle(synthPlugin)
		return unsafe.Pointer(handle)
	}
	return nil
}

//export ClapGo_PluginInit
func ClapGo_PluginInit(plugin unsafe.Pointer) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	return C.bool(p.Init())
}

//export ClapGo_PluginDestroy
func ClapGo_PluginDestroy(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	handle := cgo.Handle(plugin)
	p := handle.Value().(*SynthPlugin)
	p.Destroy()
	handle.Delete()
}

//export ClapGo_PluginActivate
func ClapGo_PluginActivate(plugin unsafe.Pointer, sampleRate C.double, minFrames, maxFrames C.uint32_t) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	return C.bool(p.Activate(float64(sampleRate), uint32(minFrames), uint32(maxFrames)))
}

//export ClapGo_PluginDeactivate
func ClapGo_PluginDeactivate(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	p.Deactivate()
}

//export ClapGo_PluginStartProcessing
func ClapGo_PluginStartProcessing(plugin unsafe.Pointer) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	return C.bool(p.StartProcessing())
}

//export ClapGo_PluginStopProcessing
func ClapGo_PluginStopProcessing(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	p.StopProcessing()
}

//export ClapGo_PluginReset
func ClapGo_PluginReset(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	p.Reset()
}

//export ClapGo_PluginProcess
func ClapGo_PluginProcess(plugin unsafe.Pointer, steadyTime C.int64_t, framesCount C.uint32_t, audioIn, audioOut unsafe.Pointer, events unsafe.Pointer) C.int {
	if plugin == nil {
		return C.int(api.ProcessError)
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	
	// Convert audio buffers (simplified for this example)
	// In a real implementation, you'd properly convert the C audio buffers
	var audioInSlice, audioOutSlice [][]float32
	
	// Convert events
	var eventHandler api.EventHandler
	
	result := p.Process(int64(steadyTime), uint32(framesCount), audioInSlice, audioOutSlice, eventHandler)
	return C.int(result)
}

//export ClapGo_PluginGetExtension
func ClapGo_PluginGetExtension(plugin unsafe.Pointer, id *C.char) unsafe.Pointer {
	if plugin == nil {
		return nil
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	extID := C.GoString(id)
	return p.GetExtension(extID)
}

//export ClapGo_PluginOnMainThread
func ClapGo_PluginOnMainThread(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	p.OnMainThread()
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

// SynthPlugin implements a simple synthesizer
type SynthPlugin struct {
	// Plugin state
	voices       [16]*Voice  // Maximum 16 voices
	sampleRate   float64
	isActivated  bool
	isProcessing bool
	host         unsafe.Pointer
	
	// Parameters
	volume       float64  // Master volume
	waveform     int      // 0 = sine, 1 = saw, 2 = square
	attack       float64  // Attack time in seconds
	decay        float64  // Decay time in seconds
	sustain      float64  // Sustain level (0-1)
	release      float64  // Release time in seconds
	
	// Transport state
	transportInfo TransportInfo
	
	// Parameter information
	paramInfos   []api.ParamInfo
}

// TransportInfo holds host transport information
type TransportInfo struct {
	IsPlaying     bool
	Tempo         float64
	TimeSignature struct {
		Numerator   int
		Denominator int
	}
	BarNumber     int
	BeatPosition  float64
	SecondsPosition float64
	IsLooping     bool
}

// NewSynthPlugin creates a new synth plugin
func NewSynthPlugin() *SynthPlugin {
	plugin := &SynthPlugin{
		sampleRate:   44100.0,
		isActivated:  false,
		isProcessing: false,
		volume:       0.7,  // -3dB
		waveform:     0,    // sine
		attack:       0.01, // 10ms
		decay:        0.1,  // 100ms
		sustain:      0.7,  // 70%
		release:      0.3,  // 300ms
		transportInfo: TransportInfo{
			Tempo: 120.0,
		},
	}
	
	// Define parameters
	plugin.paramInfos = []api.ParamInfo{
		{
			ID:           1,
			Name:         "Volume",
			Module:       "",
			MinValue:     0.0,
			MaxValue:     1.0,
			DefaultValue: 0.7,
			Flags:        api.ParamIsAutomatable | api.ParamIsBoundedBelow | api.ParamIsBoundedAbove,
		},
		{
			ID:           2,
			Name:         "Waveform",
			Module:       "",
			MinValue:     0.0,
			MaxValue:     2.0,
			DefaultValue: 0.0,
			Flags:        api.ParamIsAutomatable | api.ParamIsBoundedBelow | api.ParamIsBoundedAbove | api.ParamIsSteppable,
		},
		{
			ID:           3,
			Name:         "Attack",
			Module:       "ADSR",
			MinValue:     0.001,
			MaxValue:     2.0,
			DefaultValue: 0.01,
			Flags:        api.ParamIsAutomatable | api.ParamIsBoundedBelow | api.ParamIsBoundedAbove,
		},
		{
			ID:           4,
			Name:         "Decay",
			Module:       "ADSR",
			MinValue:     0.001,
			MaxValue:     2.0,
			DefaultValue: 0.1,
			Flags:        api.ParamIsAutomatable | api.ParamIsBoundedBelow | api.ParamIsBoundedAbove,
		},
		{
			ID:           5,
			Name:         "Sustain",
			Module:       "ADSR",
			MinValue:     0.0,
			MaxValue:     1.0,
			DefaultValue: 0.7,
			Flags:        api.ParamIsAutomatable | api.ParamIsBoundedBelow | api.ParamIsBoundedAbove,
		},
		{
			ID:           6,
			Name:         "Release",
			Module:       "ADSR",
			MinValue:     0.001,
			MaxValue:     5.0,
			DefaultValue: 0.3,
			Flags:        api.ParamIsAutomatable | api.ParamIsBoundedBelow | api.ParamIsBoundedAbove,
		},
	}
	
	return plugin
}

// Init initializes the plugin
func (p *SynthPlugin) Init() bool {
	return true
}

// Destroy cleans up plugin resources
func (p *SynthPlugin) Destroy() {
	// Nothing to clean up
}

// Activate prepares the plugin for processing
func (p *SynthPlugin) Activate(sampleRate float64, minFrames, maxFrames uint32) bool {
	p.sampleRate = sampleRate
	p.isActivated = true
	
	// Clear all voices
	for i := range p.voices {
		p.voices[i] = nil
	}
	
	return true
}

// Deactivate stops the plugin from processing
func (p *SynthPlugin) Deactivate() {
	p.isActivated = false
}

// StartProcessing begins audio processing
func (p *SynthPlugin) StartProcessing() bool {
	if !p.isActivated {
		return false
	}
	p.isProcessing = true
	return true
}

// StopProcessing ends audio processing
func (p *SynthPlugin) StopProcessing() {
	p.isProcessing = false
}

// Reset resets the plugin state
func (p *SynthPlugin) Reset() {
	// Reset all voices
	for i := range p.voices {
		p.voices[i] = nil
	}
}

// Process processes audio data
func (p *SynthPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events api.EventHandler) int {
	// Check if we're in a valid state for processing
	if !p.isActivated || !p.isProcessing {
		return api.ProcessError
	}
	
	// Process events (parameters, transport, notes, etc.)
	if events != nil {
		p.processEventHandler(events, framesCount)
	}
	
	// If no outputs, nothing to do
	if len(audioOut) == 0 {
		return api.ProcessContinue
	}
	
	// Get the number of output channels
	numChannels := len(audioOut)
	
	// Make sure we have enough buffer space
	for ch := 0; ch < numChannels; ch++ {
		outChannel := audioOut[ch]
		if len(outChannel) < int(framesCount) {
			return api.ProcessError
		}
		
		// Clear the output buffer
		for i := uint32(0); i < framesCount; i++ {
			outChannel[i] = 0.0
		}
	}
	
	// Process each voice
	var hasActiveVoices bool
	for i, voice := range p.voices {
		if voice != nil && voice.IsActive {
			hasActiveVoices = true
			
			// Calculate frequency for this note
			freq := noteToFrequency(int(voice.Key))
			
			// Generate audio for this voice
			for j := uint32(0); j < framesCount; j++ {
				// Get envelope value
				env := p.getEnvelopeValue(voice, j, framesCount)
				
				// Generate sample
				sample := p.generateSample(voice.Phase, freq) * float64(env) * voice.Velocity
				
				// Apply master volume
				sample *= p.volume
				
				// Add to all output channels
				for ch := 0; ch < numChannels; ch++ {
					audioOut[ch][j] += float32(sample)
				}
				
				// Advance oscillator phase
				voice.Phase += freq / p.sampleRate
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
	
	// Check if we have any active voices
	if !hasActiveVoices {
		return api.ProcessSleep
	}
	
	return api.ProcessContinue
}

// processEventHandler handles all incoming events using the new EventHandler interface
func (p *SynthPlugin) processEventHandler(events api.EventHandler, frameCount uint32) {
	if events == nil {
		return
	}
	
	// Process each event
	eventCount := events.GetInputEventCount()
	for i := uint32(0); i < eventCount; i++ {
		event := events.GetInputEvent(i)
		if event == nil {
			continue
		}
		
		// Handle each event type safely
		switch event.Type {
		case api.EventTypeParamValue:
			if paramEvent, ok := event.Data.(api.ParamEvent); ok {
				p.handleParameterChange(paramEvent)
			}
		case api.EventTypeNoteOn:
			if noteEvent, ok := event.Data.(api.NoteEvent); ok {
				p.handleNoteOn(noteEvent)
			}
		case api.EventTypeNoteOff:
			if noteEvent, ok := event.Data.(api.NoteEvent); ok {
				p.handleNoteOff(noteEvent)
			}
		}
	}
}

// handleParameterChange processes a parameter change event
func (p *SynthPlugin) handleParameterChange(paramEvent api.ParamEvent) {
	// Handle the parameter change based on its ID
	switch paramEvent.ParamID {
	case 1: // Volume
		p.volume = paramEvent.Value
	case 2: // Waveform
		p.waveform = int(math.Round(paramEvent.Value))
	case 3: // Attack
		p.attack = paramEvent.Value
	case 4: // Decay
		p.decay = paramEvent.Value
	case 5: // Sustain
		p.sustain = paramEvent.Value
	case 6: // Release
		p.release = paramEvent.Value
	}
}

// handleNoteOn processes a note on event
func (p *SynthPlugin) handleNoteOn(noteEvent api.NoteEvent) {
	// Validate note event fields (CLAP spec says key must be 0-127)
	if noteEvent.Key < 0 || noteEvent.Key > 127 {
		return
	}
	
	// Ensure velocity is positive
	velocity := noteEvent.Value
	if velocity <= 0 {
		velocity = 0.01 // Very quiet but not silent
	}
	
	// Find a free voice slot or steal an existing one
	voiceIndex := p.findFreeVoice()
	
	// Create a new voice with validated data
	p.voices[voiceIndex] = &Voice{
		NoteID:      noteEvent.NoteID,
		Channel:     noteEvent.Channel,
		Key:         noteEvent.Key,
		Velocity:    velocity,
		Phase:       0.0,
		IsActive:    true,
		ReleasePhase: -1.0, // Not in release phase
	}
}

// handleNoteOff processes a note off event
func (p *SynthPlugin) handleNoteOff(noteEvent api.NoteEvent) {
	// Find the voice with this note ID or key/channel combination
	for _, voice := range p.voices {
		// Safety check
		if voice == nil || !voice.IsActive {
			continue
		}
		
		// Match on note ID if provided (non-negative), otherwise match on key and channel
		// CLAP spec says note_id == -1 is for wildcard, so handle that case too
		if (noteEvent.NoteID >= 0 && voice.NoteID == noteEvent.NoteID) ||
		   (noteEvent.NoteID < 0 && voice.Key == noteEvent.Key && 
		    (noteEvent.Channel < 0 || voice.Channel == noteEvent.Channel)) {
			// Start the release phase (0.0 = start of release)
			voice.ReleasePhase = 0.0
		}
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
func (p *SynthPlugin) getEnvelopeValue(voice *Voice, sampleIndex, frameCount uint32) float64 {
	// If in release phase
	if voice.ReleasePhase >= 0.0 {
		// Update release phase
		releaseSamples := p.release * p.sampleRate
		voice.ReleasePhase += 1.0 / releaseSamples
		
		// Calculate release envelope (exponential decay)
		return math.Pow(1.0 - voice.ReleasePhase, 2.0) * p.sustain
	}
	
	// Attack phase
	attackSamples := p.attack * p.sampleRate
	if attackSamples <= 0 {
		attackSamples = 1 // Prevent division by zero
	}
	
	// Decay phase
	decaySamples := p.decay * p.sampleRate
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
		return 1.0 - decayProgress*(1.0-p.sustain)
	}
	
	// Sustain phase
	return p.sustain
}

// generateSample generates a single sample based on the current waveform
func (p *SynthPlugin) generateSample(phase, freq float64) float64 {
	switch p.waveform {
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

// GetExtension gets a plugin extension
func (p *SynthPlugin) GetExtension(id string) unsafe.Pointer {
	// Check for parameter extension
	if id == api.ExtParams {
		// TODO: Implement proper parameter extension
		return nil
	}
	
	// Check for state extension
	if id == api.ExtState {
		// TODO: Implement proper state extension
		return nil
	}
	
	// Check for note ports extension
	if id == api.ExtNotePorts {
		// TODO: Implement proper note ports extension
		return nil
	}
	
	// Check for audio ports extension
	if id == api.ExtAudioPorts {
		// TODO: Implement proper audio ports extension
		return nil
	}
	
	// Check for preset load extension
	if id == api.ExtPresetLoad {
		// TODO: Implement proper preset load extension
		return nil
	}
	
	// No other extensions supported
	return nil
}

// GetPluginInfo returns information about the plugin
func (p *SynthPlugin) GetPluginInfo() api.PluginInfo {
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

// OnMainThread is called on the main thread
func (p *SynthPlugin) OnMainThread() {
	// Nothing to do
}


// GetPluginID returns the plugin ID
func (p *SynthPlugin) GetPluginID() string {
	return PluginID
}

// SaveState returns custom state data for the plugin
func (p *SynthPlugin) SaveState() map[string]interface{} {
	// Save any additional state beyond parameters
	return map[string]interface{}{
		"plugin_version": "1.0.0",
		"waveform":      p.waveform,
		"attack":        p.attack,
		"decay":         p.decay,
		"sustain":       p.sustain,
		"release":       p.release,
	}
}

// LoadState loads custom state data for the plugin
func (p *SynthPlugin) LoadState(data map[string]interface{}) {
	// Load waveform
	if waveform, ok := data["waveform"].(float64); ok {
		p.waveform = int(waveform)
	}
	
	// Load ADSR
	if attack, ok := data["attack"].(float64); ok {
		p.attack = attack
	}
	if decay, ok := data["decay"].(float64); ok {
		p.decay = decay
	}
	if sustain, ok := data["sustain"].(float64); ok {
		p.sustain = sustain
	}
	if release, ok := data["release"].(float64); ok {
		p.release = release
	}
}

// LoadPreset loads a preset from a location
func (p *SynthPlugin) LoadPreset(locationKind uint32, location, loadKey string) bool {
	// Open the preset file
	file, err := os.Open(loadKey)
	if err != nil {
		return false
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
		return false
	}
	
	// Update synth parameters
	p.waveform = preset.Waveform
	p.attack = preset.Attack
	p.decay = preset.Decay
	p.sustain = preset.Sustain
	p.release = preset.Release
	
	// Load any additional state data
	if preset.StateData != nil {
		p.LoadState(preset.StateData)
	}
	
	return true
}

// TODO: Implement proper NotePortProvider interface

func main() {
	// This is not called when used as a plugin,
	// but can be useful for testing
}