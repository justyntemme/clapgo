package main

// #cgo CFLAGS: -I../../include/clap/include
// #include "../../include/clap/include/clap/clap.h"
// #include <stdlib.h>
import "C"
import (
	"encoding/json"
	"github.com/justyntemme/clapgo/src/goclap"
	"math"
	"os"
	"unsafe"
)

// CLAP process status codes
const (
	CLAP_PROCESS_ERROR             = 0
	CLAP_PROCESS_CONTINUE          = 1
	CLAP_PROCESS_CONTINUE_IF_NOT_QUIET = 2
	CLAP_PROCESS_TAIL              = 3
	CLAP_PROCESS_SLEEP             = 4
)

// Export Go plugin functionality
var (
	synthPlugin *SynthPlugin
)

func init() {
	// Register our synth plugin
	info := goclap.PluginInfo{
		ID:          "com.clapgo.synth",
		Name:        "Simple Synth",
		Vendor:      "ClapGo",
		URL:         "https://github.com/justyntemme/clapgo",
		ManualURL:   "https://github.com/justyntemme/clapgo",
		SupportURL:  "https://github.com/justyntemme/clapgo/issues",
		Version:     "1.0.0",
		Description: "A simple synthesizer using ClapGo",
		Features:    []string{"instrument", "synthesizer", "stereo"},
	}
	
	synthPlugin = NewSynthPlugin()
	goclap.RegisterPlugin(info, synthPlugin)
}

//export SynthGetPluginCount
func SynthGetPluginCount() C.uint32_t {
	return 1
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
	paramManager *goclap.ParamManager
	host         *goclap.Host
	
	// Parameters
	volume       float64  // Master volume
	waveform     int      // 0 = sine, 1 = saw, 2 = square
	attack       float64  // Attack time in seconds
	decay        float64  // Decay time in seconds
	sustain      float64  // Sustain level (0-1)
	release      float64  // Release time in seconds
	
	// Transport state
	transportInfo TransportInfo
	
	// Preset provider
	presetProvider *goclap.DefaultPresetProvider
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
		paramManager: goclap.NewParamManager(),
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
	
	// Create preset provider
	presetProviderDesc := goclap.PresetDiscoveryProviderDescriptor{
		ID:     "com.clapgo.synth.presets",
		Name:   "Simple Synth Presets",
		Vendor: "ClapGo",
	}
	
	// The location where presets are stored
	presetDir := "presets"
	
	// Initialize preset provider
	plugin.presetProvider = goclap.NewDefaultPresetProvider(
		presetProviderDesc,
		[]string{"json"},
		presetDir,
		[]string{"com.clapgo.synth"},
	)
	
	// Register parameters
	plugin.paramManager.RegisterParam(goclap.ParamInfo{
		ID:           1,
		Name:         "Volume",
		Module:       "",
		MinValue:     0.0,
		MaxValue:     1.0,
		DefaultValue: 0.7,
		Flags:        goclap.ParamIsAutomatable | goclap.ParamIsBoundedBelow | goclap.ParamIsBoundedAbove,
	})
	
	plugin.paramManager.RegisterParam(goclap.ParamInfo{
		ID:           2,
		Name:         "Waveform",
		Module:       "",
		MinValue:     0.0,
		MaxValue:     2.0,
		DefaultValue: 0.0,
		Flags:        goclap.ParamIsAutomatable | goclap.ParamIsBoundedBelow | goclap.ParamIsBoundedAbove | goclap.ParamIsSteppable,
	})
	
	plugin.paramManager.RegisterParam(goclap.ParamInfo{
		ID:           3,
		Name:         "Attack",
		Module:       "ADSR",
		MinValue:     0.001,
		MaxValue:     2.0,
		DefaultValue: 0.01,
		Flags:        goclap.ParamIsAutomatable | goclap.ParamIsBoundedBelow | goclap.ParamIsBoundedAbove,
	})
	
	plugin.paramManager.RegisterParam(goclap.ParamInfo{
		ID:           4,
		Name:         "Decay",
		Module:       "ADSR",
		MinValue:     0.001,
		MaxValue:     2.0,
		DefaultValue: 0.1,
		Flags:        goclap.ParamIsAutomatable | goclap.ParamIsBoundedBelow | goclap.ParamIsBoundedAbove,
	})
	
	plugin.paramManager.RegisterParam(goclap.ParamInfo{
		ID:           5,
		Name:         "Sustain",
		Module:       "ADSR",
		MinValue:     0.0,
		MaxValue:     1.0,
		DefaultValue: 0.7,
		Flags:        goclap.ParamIsAutomatable | goclap.ParamIsBoundedBelow | goclap.ParamIsBoundedAbove,
	})
	
	plugin.paramManager.RegisterParam(goclap.ParamInfo{
		ID:           6,
		Name:         "Release",
		Module:       "ADSR",
		MinValue:     0.001,
		MaxValue:     5.0,
		DefaultValue: 0.3,
		Flags:        goclap.ParamIsAutomatable | goclap.ParamIsBoundedBelow | goclap.ParamIsBoundedAbove,
	})
	
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
func (p *SynthPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events *goclap.ProcessEvents) int {
	// Check if we're in a valid state for processing
	if !p.isActivated || !p.isProcessing {
		return CLAP_PROCESS_ERROR
	}
	
	// Process events (parameters, transport, notes, etc.)
	p.processEvents(events, framesCount)
	
	// If no outputs, nothing to do
	if len(audioOut) == 0 {
		return CLAP_PROCESS_CONTINUE
	}
	
	// Get the number of output channels
	numChannels := len(audioOut)
	
	// Make sure we have enough buffer space
	for ch := 0; ch < numChannels; ch++ {
		outChannel := audioOut[ch]
		if len(outChannel) < int(framesCount) {
			return CLAP_PROCESS_ERROR
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
		return CLAP_PROCESS_SLEEP
	}
	
	return CLAP_PROCESS_CONTINUE
}

// processEvents handles all incoming events
func (p *SynthPlugin) processEvents(processEvents *goclap.ProcessEvents, frameCount uint32) {
	if processEvents == nil {
		return
	}
	
	// Create wrappers for input/output events
	inputEvents := &goclap.InputEvents{Ptr: processEvents.InEvents}
	outputEvents := &goclap.OutputEvents{Ptr: processEvents.OutEvents}
	
	// Early return if no input events
	if inputEvents.Ptr == nil {
		return
	}
	
	// Process each event
	eventCount := inputEvents.GetEventCount()
	for i := uint32(0); i < eventCount; i++ {
		event := inputEvents.GetEvent(i)
		if event == nil {
			continue
		}
		
		// Only process events from the core CLAP space
		if event.Space != 0 {
			continue
		}
		
		// Handle each event type safely
		switch event.Type {
		case goclap.EventTypeParamValue:
			p.handleParamEvent(event)
		case goclap.EventTypeTransport:
			p.handleTransportEvent(event)
		case goclap.EventTypeNoteOn:
			p.handleNoteOnEvent(event)
		case goclap.EventTypeNoteOff:
			p.handleNoteOffEvent(event)
		case goclap.EventTypeNoteChoke:
			p.handleNoteChokeEvent(event)
		case goclap.EventTypeNoteExpression:
			p.handleNoteExpressionEvent(event)
		case goclap.EventTypeNoteEnd:
			// Note end events are sent from plugin to host, not handled here
			continue
		}
		
		// For each note on event, send a corresponding note end event when done
		// This is a simplified example - in a real plugin you would only send
		// note end when a voice actually finishes
		if event.Type == goclap.EventTypeNoteOn && outputEvents.Ptr != nil {
			// Just for demonstration - in a real plugin, you'd send this when the note actually ends
			if noteEvent, ok := event.Data.(goclap.NoteEvent); ok {
				// Create a note end event
				// In a real plugin, you'd track voices and send this when the voice ends
				endEvent := &goclap.Event{
					Time:  event.Time + frameCount - 1, // At the end of this processing block
					Type:  goclap.EventTypeNoteEnd,
					Space: 0, // Core event space
					Data:  noteEvent,
				}
				
				// Try to push the event
				outputEvents.PushEvent(endEvent)
			}
		}
	}
}

// handleParamEvent processes parameter change events
func (p *SynthPlugin) handleParamEvent(event *goclap.Event) {
	paramEvent, ok := event.Data.(goclap.ParamEvent)
	if !ok {
		return
	}
	
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

// handleTransportEvent processes transport events
func (p *SynthPlugin) handleTransportEvent(event *goclap.Event) {
	transportEvent, ok := event.Data.(goclap.TransportEvent)
	if !ok {
		return
	}
	
	// Extract transport information
	if transportEvent.Flags&goclap.TransportHasTransport != 0 {
		// Update play state
		p.transportInfo.IsPlaying = (transportEvent.Flags&goclap.TransportIsPlaying) != 0
		
		// Update tempo if available
		if transportEvent.Flags&goclap.TransportHasTempo != 0 {
			p.transportInfo.Tempo = transportEvent.Tempo
		}
		
		// Update position information
		if transportEvent.Flags&goclap.TransportHasBeatsTime != 0 {
			p.transportInfo.BeatPosition = float64(transportEvent.SongPosBeats)
			p.transportInfo.BarNumber = int(transportEvent.BarNumber)
		}
		
		// Update time in seconds
		if transportEvent.Flags&goclap.TransportHasSecondsTime != 0 {
			p.transportInfo.SecondsPosition = transportEvent.SongPosSeconds
		}
		
		// Update time signature
		if transportEvent.Flags&goclap.TransportHasTimeSignature != 0 {
			p.transportInfo.TimeSignature.Numerator = int(transportEvent.TimeSigNumerator)
			p.transportInfo.TimeSignature.Denominator = int(transportEvent.TimeSigDenominator)
		}
	}
}

// handleNoteOnEvent processes note on events
func (p *SynthPlugin) handleNoteOnEvent(event *goclap.Event) {
	// Type assertion with safety check
	noteEvent, ok := event.Data.(goclap.NoteEvent)
	if !ok {
		return
	}
	
	// Validate note event fields (CLAP spec says key must be 0-127)
	if noteEvent.Key < 0 || noteEvent.Key > 127 {
		return
	}
	
	// Ensure velocity is positive
	if noteEvent.Velocity <= 0 {
		noteEvent.Velocity = 0.01 // Very quiet but not silent
	}
	
	// Find a free voice slot or steal an existing one
	voiceIndex := p.findFreeVoice()
	
	// Create a new voice with validated data
	p.voices[voiceIndex] = &Voice{
		NoteID:      noteEvent.NoteID,
		Channel:     noteEvent.Channel,
		Key:         noteEvent.Key,
		Velocity:    noteEvent.Velocity,
		Phase:       0.0,
		IsActive:    true,
		ReleasePhase: -1.0, // Not in release phase
	}
}

// handleNoteOffEvent processes note off events
func (p *SynthPlugin) handleNoteOffEvent(event *goclap.Event) {
	// Type assertion with safety check
	noteEvent, ok := event.Data.(goclap.NoteEvent)
	if !ok {
		return
	}
	
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

// handleNoteChokeEvent processes note choke events
func (p *SynthPlugin) handleNoteChokeEvent(event *goclap.Event) {
	// Type assertion with safety check
	noteEvent, ok := event.Data.(goclap.NoteEvent)
	if !ok {
		return
	}
	
	// Find the voice with this note ID or key/channel combination
	// and immediately silence it
	for i, voice := range p.voices {
		// Safety check for nil voice
		if voice == nil {
			continue
		}
		
		// Match voices using wildcard rules from CLAP spec
		if (noteEvent.NoteID >= 0 && voice.NoteID == noteEvent.NoteID) ||
		   (noteEvent.NoteID < 0 && voice.Key == noteEvent.Key && 
		    (noteEvent.Channel < 0 || voice.Channel == noteEvent.Channel)) {
			// Immediately silence the voice
			p.voices[i] = nil
		}
	}
}

// handleNoteExpressionEvent processes note expression events
func (p *SynthPlugin) handleNoteExpressionEvent(event *goclap.Event) {
	// Type assertion with safety check
	exprEvent, ok := event.Data.(goclap.NoteExpressionEvent)
	if !ok {
		return
	}
	
	// Find the voice with matching note_id, key, and channel using CLAP wildcard rules
	for _, voice := range p.voices {
		// Safety check for active voices
		if voice == nil || !voice.IsActive {
			continue
		}
		
		// Match using CLAP wildcard rules
		if (exprEvent.NoteID >= 0 && voice.NoteID == exprEvent.NoteID) ||
		   (exprEvent.NoteID < 0 && voice.Key == exprEvent.Key && 
		    (exprEvent.Channel < 0 || voice.Channel == exprEvent.Channel)) {
			
			// Handle different expression types
			switch exprEvent.Expression {
			case goclap.NoteExpressionVolume:
				// Adjust the voice's volume
				voice.Velocity = exprEvent.Value
				
			// Add other expression handlers as needed
			case goclap.NoteExpressionTuning:
				// Could adjust tuning if we implemented it
				break
				
			case goclap.NoteExpressionVibrato:
				// Could implement vibrato if needed
				break
			}
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
	if id == goclap.ExtParams {
		// TODO: Implement proper parameter extension
		// For now, return nil since GetExtension is not implemented
		return nil
	}
	
	// Check for state extension
	if id == goclap.ExtState {
		return p.getStateExtension()
	}
	
	// Check for note ports extension
	if id == goclap.ExtNotePorts {
		return p.getNotePortsExtension()
	}
	
	// Check for audio ports extension
	if id == goclap.ExtAudioPorts {
		return p.getAudioPortsExtension()
	}
	
	// Check for preset load extension
	if id == goclap.ExtPresetLoad {
		return p.getPresetLoadExtension()
	}
	
	// No other extensions supported
	return nil
}

// getStateExtension returns the state extension implementation
func (p *SynthPlugin) getStateExtension() unsafe.Pointer {
	// In a real implementation, this would return a pointer to a C struct
	return nil
}

// getNotePortsExtension returns the note ports extension implementation
func (p *SynthPlugin) getNotePortsExtension() unsafe.Pointer {
	// Create the note ports extension
	notePortsExt := goclap.NewPluginNotePortsExtension(p)
	if notePortsExt != nil {
		return notePortsExt.GetExtensionPointer()
	}
	return nil
}

// getAudioPortsExtension returns the audio ports extension implementation
func (p *SynthPlugin) getAudioPortsExtension() unsafe.Pointer {
	// In a real implementation, this would return a pointer to a C struct
	return nil
}

// getPresetLoadExtension returns the preset load extension implementation
func (p *SynthPlugin) getPresetLoadExtension() unsafe.Pointer {
	// Create and return the preset load extension
	presetExt := goclap.NewPresetLoadExtension(p)
	if presetExt != nil {
		return presetExt.GetExtensionPointer()
	}
	return nil
}

// OnMainThread is called on the main thread
func (p *SynthPlugin) OnMainThread() {
	// Nothing to do
}

// GetParamManager returns the parameter manager for this plugin
func (p *SynthPlugin) GetParamManager() *goclap.ParamManager {
	return p.paramManager
}

// GetPluginID returns the plugin ID
func (p *SynthPlugin) GetPluginID() string {
	return "com.clapgo.synth"
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

// Implementation of NotePortProvider interface

// GetNoteInputPortCount returns the number of note input ports
func (p *SynthPlugin) GetNoteInputPortCount() uint32 {
	return 1 // One MIDI input port
}

// GetNoteOutputPortCount returns the number of note output ports
func (p *SynthPlugin) GetNoteOutputPortCount() uint32 {
	return 0 // No output ports
}

// GetNoteInputPortInfo returns info about a note input port
func (p *SynthPlugin) GetNoteInputPortInfo(index uint32) *goclap.NotePortInfo {
	if index != 0 {
		return nil
	}
	
	return &goclap.NotePortInfo{
		ID:                1,
		SupportedDialects: goclap.NoteDialectCLAP | goclap.NoteDialectMIDI,
		PreferredDialect:  goclap.NoteDialectCLAP,
		Name:              "MIDI Input",
	}
}

// GetNoteOutputPortInfo returns info about a note output port
func (p *SynthPlugin) GetNoteOutputPortInfo(index uint32) *goclap.NotePortInfo {
	return nil // No output ports
}

func main() {
	// This is not called when used as a plugin,
	// but can be useful for testing
}