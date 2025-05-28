package main

// #cgo CFLAGS: -I../../include/clap/include
// #include "../../include/clap/include/clap/clap.h"
// #include <stdlib.h>
//
// // Helper functions for CLAP event handling
// static inline uint32_t clap_input_events_size_helper(const clap_input_events_t* events) {
//     if (events && events->size) {
//         return events->size(events);
//     }
//     return 0;
// }
//
// static inline const clap_event_header_t* clap_input_events_get_helper(const clap_input_events_t* events, uint32_t index) {
//     if (events && events->get) {
//         return events->get(events, index);
//     }
//     return NULL;
// }
import "C"

import (
	"fmt"
	"math"
	"sync/atomic"
	"unsafe"

	"github.com/justyntemme/clapgo/pkg/audio"
	"github.com/justyntemme/clapgo/pkg/event"
	"github.com/justyntemme/clapgo/pkg/extension"
	hostpkg "github.com/justyntemme/clapgo/pkg/host"
	"github.com/justyntemme/clapgo/pkg/manifest"
	"github.com/justyntemme/clapgo/pkg/param"
	"github.com/justyntemme/clapgo/pkg/plugin"
	"github.com/justyntemme/clapgo/pkg/process"
	"github.com/justyntemme/clapgo/pkg/thread"
	"github.com/justyntemme/clapgo/pkg/util"
)

// This example demonstrates a simple synthesizer plugin using the new API abstractions.

// Export Go plugin functionality
var (
	synthPlugin *SynthPlugin
)

func init() {
	// Create our synth plugin
	synthPlugin = NewSynthPlugin()
}

// Voice represents a single active note
type Voice struct {
	NoteID   int32
	Channel  int16
	Key      int16
	Velocity float64
	Phase    float64
	IsActive bool

	// Tuning support
	TuningID uint64 // ID of tuning to use (0 for equal temperament)

	// Per-voice parameter values (for polyphonic modulation)
	// These override the global values when non-zero
	VolumeModulation float64 // Additional volume modulation (0.0 = no change)
	PitchBend        float64 // Pitch bend in semitones
	Brightness       float64 // Filter/brightness modulation (0.0-1.0)
	Pressure         float64 // Aftertouch/pressure value (0.0-1.0)

	// ADSR envelope
	Envelope *audio.ADSREnvelope
}

// SynthPlugin implements a simple synthesizer with atomic parameter storage
type SynthPlugin struct {
	*plugin.PluginBase
	event.NoOpHandler // Embed to get default no-op implementations

	// Plugin state
	voices [16]*Voice // Maximum 16 voices

	// Parameters with atomic storage for thread safety
	volume   int64 // atomic storage for volume (0.0-1.0)
	waveform int64 // atomic storage for waveform (0-2)
	attack   int64 // atomic storage for attack time
	decay    int64 // atomic storage for decay time
	sustain  int64 // atomic storage for sustain level
	release  int64 // atomic storage for release time

	// Transport state
	transportInfo TransportInfo

	// Note port management
	notePortManager *audio.NotePortManager

	// Host utilities (non-base specific)
	transportControl *hostpkg.TransportControl
	tuning           *extension.HostTuning

	// Event pool diagnostics
	poolDiagnostics *event.Diagnostics
}

// TransportInfo holds host transport information
type TransportInfo struct {
	IsPlaying     bool
	Tempo         float64
	TimeSignature struct {
		Numerator   int
		Denominator int
	}
	BarNumber       int
	BeatPosition    float64
	SecondsPosition float64
	IsLooping       bool
}

// NewSynthPlugin creates a new synth plugin
func NewSynthPlugin() *SynthPlugin {
	pluginInfo := plugin.Info{
		ID:          PluginID,
		Name:        PluginName,
		Vendor:      PluginVendor,
		Version:     PluginVersion,
		Description: PluginDescription,
		URL:         "https://github.com/justyntemme/clapgo",
		Manual:      "https://github.com/justyntemme/clapgo",
		Support:     "https://github.com/justyntemme/clapgo/issues",
		Features:    []string{plugin.FeatureInstrument, plugin.FeatureStereo},
	}

	plugin := &SynthPlugin{
		PluginBase: plugin.NewPluginBase(pluginInfo),
		transportInfo: TransportInfo{
			Tempo: 120.0,
		},
		notePortManager: audio.NewNotePortManager(),
		poolDiagnostics: &event.Diagnostics{},
	}

	// Set default values atomically
	atomic.StoreInt64(&plugin.volume, int64(util.AtomicFloat64ToBits(0.7)))  // -3dB
	atomic.StoreInt64(&plugin.waveform, 0)                                   // sine
	atomic.StoreInt64(&plugin.attack, int64(util.AtomicFloat64ToBits(0.01))) // 10ms
	atomic.StoreInt64(&plugin.decay, int64(util.AtomicFloat64ToBits(0.1)))   // 100ms
	atomic.StoreInt64(&plugin.sustain, int64(util.AtomicFloat64ToBits(0.7))) // 70%
	atomic.StoreInt64(&plugin.release, int64(util.AtomicFloat64ToBits(0.3))) // 300ms

	// Register parameters using our new abstraction
	plugin.ParamManager.Register(param.Percentage(1, "Volume", 70.0))
	plugin.ParamManager.Register(param.Choice(2, "Waveform", 3, 0))

	// Register ADSR parameters
	plugin.ParamManager.Register(param.ADSR(3, "Attack", 2.0))         // Max 2 seconds
	plugin.ParamManager.Register(param.ADSR(4, "Decay", 2.0))          // Max 2 seconds
	plugin.ParamManager.Register(param.Percentage(5, "Sustain", 70.0)) // 0-100%
	plugin.ParamManager.Register(param.ADSR(6, "Release", 5.0))        // Max 5 seconds

	// Configure note port for instrument
	plugin.notePortManager.AddInputPort(audio.CreateDefaultInstrumentPort())

	return plugin
}

// Init initializes the plugin
func (p *SynthPlugin) Init() bool {
	// Mark this as main thread for debug builds
	thread.DebugSetMainThread()

	// Initialize thread checker
	if p.Host != nil {
		p.ThreadCheck = thread.NewChecker(p.Host)
		if p.ThreadCheck.IsAvailable() && p.Logger != nil {
			p.Logger.Info("Thread Check extension available - thread safety validation enabled")
		}
	}

	// Initialize track info helper
	if p.Host != nil {
		p.TrackInfo = hostpkg.NewTrackInfoProvider(p.Host)
	}

	// Initialize transport control
	if p.Host != nil {
		p.transportControl = hostpkg.NewTransportControl(p.Host)
		if p.Logger != nil {
			p.Logger.Info("Transport control extension initialized")
		}
	}

	// Initialize tuning support
	if p.Host != nil {
		p.tuning = extension.NewHostTuning(p.Host)
		if p.Logger != nil {
			p.Logger.Info("Tuning extension initialized")
			// Log available tunings
			tunings := p.tuning.GetAllTunings()
			if len(tunings) > 0 {
				for _, t := range tunings {
					p.Logger.Info(fmt.Sprintf("Available tuning: %s (ID: %d, Dynamic: %v)",
						t.Name, t.TuningID, t.IsDynamic))
				}
			}
		}
	}

	if p.Logger != nil {
		p.Logger.Debug("Synth plugin initialized")
	}

	// TODO: Initialize context menu provider with param.Manager support
	// if p.Host != nil {
	//	p.contextMenuProvider = api.NewDefaultContextMenuProvider(
	//		p.ParamManager,
	//		"Simple Synth",
	//		"1.0.0",
	//		p.Host,
	//	)
	//	p.contextMenuProvider.SetAboutMessage("Simple Synth v1.0.0 - A basic subtractive synthesizer")
	// }

	// Check initial track info
	if p.TrackInfo != nil {
		p.OnTrackInfoChanged()
	}

	// Register as voice info provider
	// Note: The plugin pointer is the cgo.Handle value returned by CreatePlugin
	// We can't access it here, so registration will happen externally

	return true
}

// Destroy cleans up plugin resources
func (p *SynthPlugin) Destroy() {
	// Cleanup is handled externally
}

// Activate prepares the plugin for processing
func (p *SynthPlugin) Activate(sampleRate float64, minFrames, maxFrames uint32) bool {
	p.SampleRate = sampleRate
	p.IsActivated = true

	// Clear all voices
	for i := range p.voices {
		p.voices[i] = nil
	}

	if p.Logger != nil {
		p.Logger.Info(fmt.Sprintf("Synth activated at %.0f Hz, buffer size %d-%d", sampleRate, minFrames, maxFrames))
	}

	return true
}

// Deactivate stops the plugin from processing
func (p *SynthPlugin) Deactivate() {
	p.IsActivated = false
}

// StartProcessing begins audio processing
func (p *SynthPlugin) StartProcessing() bool {
	if !p.IsActivated {
		return false
	}
	p.IsProcessing = true
	return true
}

// StopProcessing ends audio processing
func (p *SynthPlugin) StopProcessing() {
	p.IsProcessing = false
}

// Reset resets the plugin state
func (p *SynthPlugin) Reset() {
	// Reset all voices
	for i := range p.voices {
		p.voices[i] = nil
	}
}

// Process processes audio data using the new abstractions
func (p *SynthPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events *event.EventProcessor) int {
	// Check if we're in a valid state for processing
	if !p.IsActivated || !p.IsProcessing {
		return process.ProcessError
	}

	// Process events using our new abstraction - NO MORE MANUAL EVENT PARSING!
	if events != nil {
		p.processEventHandler(events, framesCount)
	}

	// If no outputs, nothing to do
	if len(audioOut) == 0 {
		return process.ProcessContinue
	}

	// Get current parameter values atomically
	volume := param.LoadParameterAtomic(&p.volume)
	waveform := int(atomic.LoadInt64(&p.waveform))
	attack := param.LoadParameterAtomic(&p.attack)
	decay := param.LoadParameterAtomic(&p.decay)
	sustain := param.LoadParameterAtomic(&p.sustain)
	release := param.LoadParameterAtomic(&p.release)

	// Get the number of output channels
	numChannels := len(audioOut)

	// Clear the output buffer
	for ch := 0; ch < numChannels; ch++ {
		outChannel := audioOut[ch]
		if len(outChannel) < int(framesCount) {
			continue
		}

		for i := uint32(0); i < framesCount; i++ {
			outChannel[i] = 0.0
		}
	}

	// Process each voice
	var hasActiveVoices bool
	for i, voice := range p.voices {
		if voice != nil && voice.IsActive {
			hasActiveVoices = true

			// Calculate frequency for this note with pitch bend
			baseFreq := audio.NoteToFrequency(int(voice.Key))

			// Apply tuning if available
			if p.tuning != nil && voice.TuningID != 0 {
				// Apply tuning at the start of the frame
				baseFreq = p.tuning.ApplyTuning(baseFreq, voice.TuningID,
					int32(voice.Channel), int32(voice.Key), 0)
			}

			// Apply pitch bend (in semitones)
			freq := baseFreq * math.Pow(2.0, voice.PitchBend/12.0)

			// Generate audio for this voice
			for j := uint32(0); j < framesCount; j++ {
				// Get envelope value
				env := p.getEnvelopeValue(voice, j, framesCount, attack, decay, sustain, release)

				// Generate sample with optional brightness filtering
				sample := audio.GenerateWaveformSample(voice.Phase, audio.WaveformType(waveform))

				// Apply brightness as a simple low-pass filter simulation
				if voice.Brightness > 0.0 && voice.Brightness < 1.0 {
					// Simple brightness simulation (in real implementation, use proper filter)
					sample *= (voice.Brightness*0.7 + 0.3)
				}

				// Apply envelope and velocity
				sample *= env * voice.Velocity

				// Apply per-voice volume modulation
				voiceVolume := 1.0 + voice.VolumeModulation
				if voiceVolume < 0.0 {
					voiceVolume = 0.0
				}
				sample *= voiceVolume

				// Apply pressure (aftertouch) as additional volume modulation
				if voice.Pressure > 0.0 {
					// Pressure affects volume (could also affect other parameters)
					sample *= (1.0 + voice.Pressure*0.3)
				}

				// Apply master volume
				sample *= volume

				// Add to all output channels
				for ch := 0; ch < numChannels; ch++ {
					if len(audioOut[ch]) > int(j) {
						audioOut[ch][j] += float32(sample)
					}
				}

				// Advance oscillator phase
				voice.Phase = audio.AdvancePhase(voice.Phase, freq, p.SampleRate)
			}

			// Check if voice is still active
			if !voice.Envelope.IsActive() {
				// Send note end event to host if we have a valid note ID
				if voice.NoteID >= 0 && events != nil {
					endEvent := event.CreateNoteEndEvent(0, voice.NoteID, -1, voice.Channel, voice.Key)
					events.PushOutputEvent(endEvent)
				}
				p.voices[i] = nil
			}
		}
	}

	// Check if we have any active voices
	if !hasActiveVoices {
		return process.ProcessSleep
	}

	return process.ProcessContinue
}

// processEventHandler handles all incoming events using our new EventHandler abstraction
func (p *SynthPlugin) processEventHandler(events *event.EventProcessor, frameCount uint32) {
	if events == nil {
		return
	}

	// Process all events
	events.ProcessAll(p)
}

// ParamValueToText provides custom formatting for synth parameters
func (p *SynthPlugin) ParamValueToText(paramID uint32, value float64, buffer unsafe.Pointer, size uint32) bool {
	if buffer == nil || size == 0 {
		return false
	}

	// Format based on parameter type
	var text string
	switch paramID {
	case 1: // Volume
		text = param.FormatValue(value, param.FormatPercentage)
	case 2: // Waveform
		switch int(math.Round(value)) {
		case 0:
			text = "Sine"
		case 1:
			text = "Saw"
		case 2:
			text = "Square"
		default:
			text = "Unknown"
		}
	case 3, 4, 6: // Attack, Decay, Release
		text = param.FormatValue(value, param.FormatMilliseconds)
	case 5: // Sustain
		text = param.FormatValue(value, param.FormatPercentage)
	default:
		// Use base implementation for unknown parameters
		return p.PluginBase.ParamValueToText(paramID, value, buffer, size)
	}

	// Copy string to C buffer
	bytes := []byte(text)
	if len(bytes) >= int(size) {
		bytes = bytes[:size-1]
	}

	// Convert to C char buffer
	charBuf := (*[1 << 30]byte)(buffer)
	copy(charBuf[:size], bytes)
	charBuf[len(bytes)] = 0

	return true
}

// ParamTextToValue provides custom parsing for synth parameters
func (p *SynthPlugin) ParamTextToValue(paramID uint32, text string, value unsafe.Pointer) bool {
	if value == nil {
		return false
	}

	// Parse based on parameter type
	switch paramID {
	case 1: // Volume (percentage)
		parser := param.NewParser(param.FormatPercentage)
		if parsedValue, err := parser.ParseValue(text); err == nil {
			*(*float64)(value) = param.ClampValue(parsedValue, 0.0, 1.0)
			return true
		}
	case 2: // Waveform
		// Parse waveform names
		switch text {
		case "Sine":
			*(*float64)(value) = 0.0
			return true
		case "Saw":
			*(*float64)(value) = 1.0
			return true
		case "Square":
			*(*float64)(value) = 2.0
			return true
		}
	case 3, 4, 6: // Attack, Decay, Release (milliseconds)
		parser := param.NewParser(param.FormatMilliseconds)
		if parsedValue, err := parser.ParseValue(text); err == nil {
			*(*float64)(value) = param.ClampValue(parsedValue, 0.001, 5.0)
			return true
		}
	case 5: // Sustain (percentage)
		parser := param.NewParser(param.FormatPercentage)
		if parsedValue, err := parser.ParseValue(text); err == nil {
			*(*float64)(value) = param.ClampValue(parsedValue, 0.0, 1.0)
			return true
		}
	}

	// Use base implementation for unknown parameters
	return p.PluginBase.ParamTextToValue(paramID, text, value)
}

// ParamsFlush overrides base to use processEventHandler
func (p *SynthPlugin) ParamsFlush(inEvents, outEvents unsafe.Pointer) {
	// Process events using our abstraction
	if inEvents != nil {
		eventHandler := event.NewEventProcessor(inEvents, outEvents)
		p.processEventHandler(eventHandler, 0)
	}
}

// HandleParamValue handles parameter value changes
func (p *SynthPlugin) HandleParamValue(paramEvent *event.ParamValueEvent, time uint32) {
	// Check if this is a polyphonic parameter event (targets specific note)
	if paramEvent.NoteID >= 0 || paramEvent.Key >= 0 {
		// This is a polyphonic parameter change - apply to specific voice(s)
		p.handlePolyphonicParameter(*paramEvent)
		return
	}

	// Handle global parameter changes
	switch paramEvent.ParamID {
	case 1: // Volume
		atomic.StoreInt64(&p.volume, int64(util.AtomicFloat64ToBits(paramEvent.Value)))
		p.ParamManager.Set(paramEvent.ParamID, paramEvent.Value)
	case 2: // Waveform
		atomic.StoreInt64(&p.waveform, int64(math.Round(paramEvent.Value)))
		p.ParamManager.Set(paramEvent.ParamID, paramEvent.Value)
	case 3: // Attack
		atomic.StoreInt64(&p.attack, int64(util.AtomicFloat64ToBits(paramEvent.Value)))
		p.ParamManager.Set(paramEvent.ParamID, paramEvent.Value)
	case 4: // Decay
		atomic.StoreInt64(&p.decay, int64(util.AtomicFloat64ToBits(paramEvent.Value)))
		p.ParamManager.Set(paramEvent.ParamID, paramEvent.Value)
	case 5: // Sustain
		atomic.StoreInt64(&p.sustain, int64(util.AtomicFloat64ToBits(paramEvent.Value)))
		p.ParamManager.Set(paramEvent.ParamID, paramEvent.Value)
	case 6: // Release
		atomic.StoreInt64(&p.release, int64(util.AtomicFloat64ToBits(paramEvent.Value)))
		p.ParamManager.Set(paramEvent.ParamID, paramEvent.Value)
	}
}

// handlePolyphonicParameter processes polyphonic parameter changes
func (p *SynthPlugin) handlePolyphonicParameter(paramEvent event.ParamValueEvent) {
	// Find matching voices
	for _, voice := range p.voices {
		if voice == nil || !voice.IsActive {
			continue
		}

		// Match by note ID if specified
		if paramEvent.NoteID >= 0 && voice.NoteID != paramEvent.NoteID {
			continue
		}

		// Match by key/channel if specified
		if paramEvent.Key >= 0 && voice.Key != paramEvent.Key {
			continue
		}
		if paramEvent.Channel >= 0 && voice.Channel != paramEvent.Channel {
			continue
		}

		// Apply parameter to this voice
		switch paramEvent.ParamID {
		case 1: // Volume modulation
			voice.VolumeModulation = paramEvent.Value - 1.0 // Store as offset from 1.0
		case 7: // Pitch bend (new parameter we'll add)
			voice.PitchBend = paramEvent.Value*2.0 - 1.0 // Convert 0-1 to -1 to +1 semitones
		case 8: // Brightness (new parameter)
			voice.Brightness = paramEvent.Value
		case 9: // Pressure (new parameter)
			voice.Pressure = paramEvent.Value
		}
	}
}

// HandleNoteOn handles note on events
func (p *SynthPlugin) HandleNoteOn(noteEvent *event.NoteEvent, time uint32) {
	// Validate note event fields (CLAP spec says key must be 0-127)
	if noteEvent.Key < 0 || noteEvent.Key > 127 {
		return
	}

	// Special transport control: C0 (MIDI note 24) toggles play/pause
	if noteEvent.Key == 24 && p.transportControl != nil {
		p.transportControl.RequestTogglePlay()
		if p.Logger != nil {
			p.Logger.Info("Transport toggle play requested via C0")
		}
		return // Don't play the note
	}

	// Ensure velocity is positive
	velocity := noteEvent.Velocity
	if velocity <= 0 {
		velocity = 0.01 // Very quiet but not silent
	}

	// Find a free voice slot or steal an existing one
	voiceIndex := p.findFreeVoice()

	// Create a new voice with validated data
	// (envelope parameters will be used directly in processing)

	p.voices[voiceIndex] = &Voice{
		NoteID:   noteEvent.NoteID,
		Channel:  noteEvent.Channel,
		Key:      noteEvent.Key,
		Velocity: velocity,
		Phase:    0.0,
		IsActive: true,
		Envelope: audio.NewADSREnvelope(p.SampleRate),
	}
	
	// Trigger the envelope to start the attack phase
	p.voices[voiceIndex].Envelope.Trigger()

	if p.Logger != nil {
		p.Logger.Debug(fmt.Sprintf("Note on: key=%d, velocity=%.2f, voice=%d", noteEvent.Key, velocity, voiceIndex))
	}
}

// HandleNoteOff handles note off events
func (p *SynthPlugin) HandleNoteOff(noteEvent *event.NoteEvent, time uint32) {
	// Find the voice with this note ID or key/channel combination
	for _, voice := range p.voices {
		// Safety check
		if voice == nil || !voice.IsActive {
			continue
		}

		// Match on note ID if provided (non-negative), otherwise match on key and channel
		if (noteEvent.NoteID >= 0 && voice.NoteID == noteEvent.NoteID) ||
			(noteEvent.NoteID < 0 && voice.Key == noteEvent.Key &&
				(noteEvent.Channel < 0 || voice.Channel == noteEvent.Channel)) {
			// Start the release phase
			voice.Envelope.Release()
		}
	}
}

// HandleNoteChoke handles note choke events (immediate stop)
func (p *SynthPlugin) HandleNoteChoke(noteEvent *event.NoteEvent, time uint32) {
	// Find the voice with this note ID or key/channel combination
	for i, voice := range p.voices {
		// Safety check
		if voice == nil || !voice.IsActive {
			continue
		}

		// Match on note ID if provided (non-negative), otherwise match on key and channel
		if (noteEvent.NoteID >= 0 && voice.NoteID == noteEvent.NoteID) ||
			(noteEvent.NoteID < 0 && voice.Key == noteEvent.Key &&
				(noteEvent.Channel < 0 || voice.Channel == noteEvent.Channel)) {
			// Immediately deactivate the voice
			p.voices[i] = nil
		}
	}
}

// HandleNoteExpression handles note expression events (MPE)
func (p *SynthPlugin) HandleNoteExpression(noteExprEvent *event.NoteExpressionEvent, time uint32) {
	// Find matching voices
	for _, voice := range p.voices {
		if voice == nil || !voice.IsActive {
			continue
		}

		// Match by note ID if specified
		if noteExprEvent.NoteID >= 0 && voice.NoteID != noteExprEvent.NoteID {
			continue
		}

		// Match by key/channel if specified
		if noteExprEvent.Key >= 0 && voice.Key != noteExprEvent.Key {
			continue
		}
		if noteExprEvent.Channel >= 0 && voice.Channel != noteExprEvent.Channel {
			continue
		}

		// Apply expression to this voice
		switch noteExprEvent.ExpressionID {
		case event.NoteExpressionVolume:
			voice.VolumeModulation = noteExprEvent.Value - 1.0 // Store as offset
		case event.NoteExpressionPan:
			// Could implement stereo panning here
		case event.NoteExpressionTuning:
			voice.PitchBend = noteExprEvent.Value // In semitones
		case event.NoteExpressionVibrato:
			// Could implement vibrato depth here
		case event.NoteExpressionBrightness:
			voice.Brightness = noteExprEvent.Value
		case event.NoteExpressionPressure:
			voice.Pressure = noteExprEvent.Value
		}
	}
}

// HandleMIDI handles MIDI 1.0 events
func (p *SynthPlugin) HandleMIDI(midiEvent *event.MIDIEvent, time uint32) {
	// Use the helper to process standard MIDI messages
	event.ProcessStandardMIDI(midiEvent, p, time)

	// Handle additional MIDI CC messages
	if len(midiEvent.Data) >= 3 {
		status := midiEvent.Data[0] & 0xF0
		if status == 0xB0 { // Control Change
			switch midiEvent.Data[1] {
			case 7: // Volume
				value := float64(midiEvent.Data[2]) / 127.0
				paramEvent := event.ParamValueEvent{
					ParamID: 1, // Volume parameter
					Value:   value,
				}
				p.HandleParamValue(&paramEvent, time)
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

	// If no empty slots, find the voice in release phase
	for i, voice := range p.voices {
		if voice != nil && voice.Envelope.Stage == audio.EnvelopeStageRelease {
			return i
		}
	}

	// No voices in release, steal the quietest voice
	quietestIdx := 0
	quietestLevel := 1.0

	for i, voice := range p.voices {
		if voice != nil && voice.IsActive {
			// Consider current envelope value and velocity
			level := voice.Envelope.CurrentValue * voice.Velocity
			if level < quietestLevel {
				quietestLevel = level
				quietestIdx = i
			}
		}
	}

	return quietestIdx
}

// getEnvelopeValue updates the envelope parameters and returns the current value
func (p *SynthPlugin) getEnvelopeValue(voice *Voice, sampleIndex, frameCount uint32, attack, decay, sustain, release float64) float64 {
	// Update envelope parameters if they changed
	voice.Envelope.SetADSR(attack, decay, sustain, release)
	
	// Process one sample and return the value
	return voice.Envelope.Process()
}

// GetNotePortManager returns the plugin's note port manager
func (p *SynthPlugin) GetNotePortManager() *audio.NotePortManager {
	return p.notePortManager
}

// GetExtension gets a plugin extension
func (p *SynthPlugin) GetExtension(id string) unsafe.Pointer {
	// All extensions are handled by the C bridge based on exported functions
	// The C bridge checks for the presence of required exports and provides
	// the extension implementation if available

	// Delegate to base plugin for any unhandled extensions
	return p.PluginBase.GetExtension(id)
}

// GetPluginInfo returns information about the plugin
func (p *SynthPlugin) GetPluginInfo() manifest.PluginInfo {
	return manifest.PluginInfo{
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
		"waveform":       atomic.LoadInt64(&p.waveform),
		"attack":         util.AtomicFloat64FromBits(uint64(atomic.LoadInt64(&p.attack))),
		"decay":          util.AtomicFloat64FromBits(uint64(atomic.LoadInt64(&p.decay))),
		"sustain":        util.AtomicFloat64FromBits(uint64(atomic.LoadInt64(&p.sustain))),
		"release":        util.AtomicFloat64FromBits(uint64(atomic.LoadInt64(&p.release))),
	}
}

// LoadState loads custom state data for the plugin
func (p *SynthPlugin) LoadState(data map[string]interface{}) {
	// Load waveform
	if waveform, ok := data["waveform"].(float64); ok {
		atomic.StoreInt64(&p.waveform, int64(waveform))
	}

	// Load ADSR
	if attack, ok := data["attack"].(float64); ok {
		atomic.StoreInt64(&p.attack, int64(util.AtomicFloat64ToBits(attack)))
	}
	if decay, ok := data["decay"].(float64); ok {
		atomic.StoreInt64(&p.decay, int64(util.AtomicFloat64ToBits(decay)))
	}
	if sustain, ok := data["sustain"].(float64); ok {
		atomic.StoreInt64(&p.sustain, int64(util.AtomicFloat64ToBits(sustain)))
	}
	if release, ok := data["release"].(float64); ok {
		atomic.StoreInt64(&p.release, int64(util.AtomicFloat64ToBits(release)))
	}
}

// GetLatency returns the plugin's latency in samples
func (p *SynthPlugin) GetLatency() uint32 {
	// Synth has no lookahead or processing latency
	return 0
}

// GetTail returns the plugin's tail length in samples
func (p *SynthPlugin) GetTail() uint32 {
	// Get release time
	release := util.AtomicFloat64FromBits(uint64(atomic.LoadInt64(&p.release)))

	// Convert to samples
	tailSamples := uint32(release * p.SampleRate)

	// Add some extra samples for safety
	return tailSamples + uint32(p.SampleRate*0.1) // 100ms extra
}

// OnTimer handles timer callbacks
func (p *SynthPlugin) OnTimer(timerID uint64) {
	// Synth doesn't currently use timers
	// Could be used for UI updates, voice status monitoring, etc.
	if p.Logger != nil {
		p.Logger.Debug(fmt.Sprintf("Timer %d fired", timerID))
	}
}

// OnTrackInfoChanged is called when the track information changes
func (p *SynthPlugin) OnTrackInfoChanged() {
	if p.TrackInfo == nil {
		return
	}

	// Get the new track information
	info, ok := p.TrackInfo.Get()
	if !ok {
		if p.Logger != nil {
			p.Logger.Warning("Failed to get track info")
		}
		return
	}

	// Log the track information
	if p.Logger != nil {
		p.Logger.Info(fmt.Sprintf("Track info changed:"))
		if info.Flags&hostpkg.TrackInfoHasTrackName != 0 {
			p.Logger.Info(fmt.Sprintf("  Track name: %s", info.Name))
		}
		if info.Flags&hostpkg.TrackInfoHasTrackColor != 0 {
			p.Logger.Info(fmt.Sprintf("  Track color: R=%d G=%d B=%d A=%d",
				info.Color.Red, info.Color.Green, info.Color.Blue, info.Color.Alpha))
		}
		if info.Flags&hostpkg.TrackInfoHasAudioChannel != 0 {
			p.Logger.Info(fmt.Sprintf("  Audio channels: %d, port type: %s",
				info.AudioChannelCount, info.AudioPortType))
		}

		// Adjust synth behavior based on track type
		if info.Flags&hostpkg.TrackInfoIsForReturnTrack != 0 {
			p.Logger.Info("  This is a return track - adjusting for wet signal")
			// Could adjust default mix to 100% wet
		}
		if info.Flags&hostpkg.TrackInfoIsForBus != 0 {
			p.Logger.Info("  This is a bus track")
			// Could adjust polyphony or processing
		}
		if info.Flags&hostpkg.TrackInfoIsForMaster != 0 {
			p.Logger.Info("  This is the master track")
			// Synths typically wouldn't be on master, but if so, could adjust
		}
	}
}

// OnTuningChanged is called when tunings are added or removed
func (p *SynthPlugin) OnTuningChanged() {
	if p.tuning == nil {
		return
	}

	if p.Logger != nil {
		p.Logger.Info("Tuning pool changed, refreshing available tunings")

		// Log all available tunings
		tunings := p.tuning.GetAllTunings()
		p.Logger.Info(fmt.Sprintf("Available tunings: %d", len(tunings)))
		for _, t := range tunings {
			p.Logger.Info(fmt.Sprintf("  - %s (ID: %d, Dynamic: %v)",
				t.Name, t.TuningID, t.IsDynamic))
		}
	}
}

// GetVoiceInfo returns voice count and capacity information
func (p *SynthPlugin) GetVoiceInfo() extension.VoiceInfo {
	// Count active voices
	activeVoices := uint32(0)
	for _, voice := range p.voices {
		if voice != nil && voice.IsActive {
			activeVoices++
		}
	}

	return extension.VoiceInfo{
		VoiceCount:    uint32(len(p.voices)),                           // Maximum polyphony
		VoiceCapacity: uint32(len(p.voices)),                           // Same as count in our implementation
		Flags:         extension.VoiceInfoFlagSupportsOverlappingNotes, // We support note IDs
	}
}

func main() {
	// This is not called when used as a plugin,
	// but can be useful for testing
}

