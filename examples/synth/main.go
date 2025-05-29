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
	"unsafe"

	"github.com/justyntemme/clapgo/pkg/audio"
	"github.com/justyntemme/clapgo/pkg/controls"
	"github.com/justyntemme/clapgo/pkg/event"
	"github.com/justyntemme/clapgo/pkg/extension"
	hostpkg "github.com/justyntemme/clapgo/pkg/host"
	"github.com/justyntemme/clapgo/pkg/manifest"
	"github.com/justyntemme/clapgo/pkg/param"
	"github.com/justyntemme/clapgo/pkg/plugin"
	"github.com/justyntemme/clapgo/pkg/process"
	"github.com/justyntemme/clapgo/pkg/thread"
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

// Voice is now using audio.Voice from the framework

// SynthPlugin implements a simple synthesizer with atomic parameter storage
type SynthPlugin struct {
	*plugin.PluginBase

	// Audio processing components
	voiceManager   *audio.VoiceManager
	oscillator     *audio.PolyphonicOscillator
	filter         *audio.SelectableFilter
	midiProcessor  *audio.MIDIProcessor

	// Parameter binding system
	params *param.ParameterBinder

	// Direct parameter access
	volume     *param.AtomicFloat64
	waveform   *param.AtomicFloat64
	attack     *param.AtomicFloat64
	decay      *param.AtomicFloat64
	sustain    *param.AtomicFloat64
	release    *param.AtomicFloat64
	cutoff     *param.AtomicFloat64
	resonance  *param.AtomicFloat64
	filterType *param.AtomicFloat64
	
	// Debug counter for periodic logging
	debugFrameCounter uint64

	// Transport state
	transportInfo TransportInfo

	// Note port management
	notePortManager *audio.NotePortManager

	// Extension bundle for host integration
	extensions *extension.ExtensionBundle

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

	// Create voice manager first
	voiceManager := audio.NewVoiceManager(16, 44100) // 16 voice polyphony

	plugin := &SynthPlugin{
		PluginBase: plugin.NewPluginBase(pluginInfo),
		transportInfo: TransportInfo{
			Tempo: 120.0,
		},
		notePortManager: audio.NewNotePortManager(),
		poolDiagnostics: &event.Diagnostics{},
		voiceManager:    voiceManager,
		oscillator:      audio.NewPolyphonicOscillator(voiceManager),
		filter:          audio.NewSelectableFilter(44100, true), // Enable safe mode
	}
	
	// Create MIDI processor
	plugin.midiProcessor = audio.NewMIDIProcessor(voiceManager, plugin.ParamManager)

	// Create parameter binder
	plugin.params = param.NewParameterBinder(plugin.ParamManager)
	
	// Bind all parameters - this creates atomic storage AND registers with manager
	plugin.volume = plugin.params.BindPercentage(1, "Volume", 70.0)
	plugin.waveform = plugin.params.BindChoice(2, "Waveform", []string{"Sine", "Saw", "Square"}, 0)
	plugin.attack = plugin.params.BindADSR(3, "Attack", 2.0, 0.01)    // Max 2s, default 10ms
	plugin.decay = plugin.params.BindADSR(4, "Decay", 2.0, 0.1)       // Max 2s, default 100ms
	plugin.sustain = plugin.params.BindPercentage(5, "Sustain", 70.0)
	plugin.release = plugin.params.BindADSR(6, "Release", 5.0, 0.3)   // Max 5s, default 300ms
	plugin.cutoff = plugin.params.BindCutoffLog(7, "Filter Cutoff", 0.5) // 0.5 maps to ~400Hz
	plugin.resonance = plugin.params.BindResonance(8, "Filter Resonance", 0.5)
	plugin.filterType = plugin.params.BindChoice(9, "Filter Type", 
		[]string{"Lowpass", "Highpass", "Bandpass", "Notch"}, 0)

	// Configure note port for instrument
	plugin.notePortManager.AddInputPort(audio.CreateDefaultInstrumentPort())

	return plugin
}

// Init initializes the plugin
func (p *SynthPlugin) Init() bool {
	// Mark this as main thread for debug builds
	thread.DebugSetMainThread()

	// Initialize all extensions in one call
	p.extensions = extension.NewExtensionBundle(p.Host, PluginName)
	
	// Set up custom MIDI callbacks
	p.midiProcessor.SetCallbacks(
		// onNoteOn - handle transport control via C0
		func(channel, key int16, velocity float64) {
			// Special transport control: C0 (MIDI note 24) toggles play/pause
			if key == 24 {
				if p.extensions.RequestTogglePlay() {
					p.extensions.LogInfo("Transport toggle play requested via C0")
				}
			}
		},
		nil, // onNoteOff
		// onModulation - handle CC7 volume and CC74 filter cutoff changes
		func(channel int16, cc uint32, value float64) {
			switch cc {
			case 7: // Volume CC
				// Update the master volume parameter
				p.volume.UpdateWithManager(value, p.ParamManager, 1)
			case 74: // Filter cutoff CC (commonly used for brightness)
				// CC value (0-1) directly maps to logarithmic cutoff parameter
				// The logarithmic mapping is handled automatically by the parameter binding
				p.cutoff.UpdateWithManager(value, p.ParamManager, 7)
			}
		},
		nil, // onPitchBend
		nil, // onPolyPressure
	)

	p.extensions.LogDebug("Synth plugin initialized")

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
	if p.extensions.HasTrackInfo() {
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

	// Update voice manager and filter sample rate
	p.voiceManager.SetSampleRate(sampleRate)
	p.filter.SetSampleRate(sampleRate)
	p.filter.Reset() // Ensure clean state

	p.extensions.LogInfo(fmt.Sprintf("Synth activated at %.0f Hz, buffer size %d-%d", sampleRate, minFrames, maxFrames))

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
	// Reset voice manager
	p.voiceManager.Reset()
}

// Process processes audio data using the new abstractions
func (p *SynthPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events *event.EventProcessor) int {
	// Check if we're in a valid state for processing
	if !p.IsActivated || !p.IsProcessing {
		return process.ProcessError
	}

	// Process events using our new abstraction
	if events != nil {
		p.processEventHandler(events, framesCount)
	}

	// If no outputs, nothing to do
	if len(audioOut) == 0 {
		return process.ProcessContinue
	}

	// Get current parameter values atomically
	volume := p.volume.Load()
	waveform := int(p.waveform.Load())
	attack := p.attack.Load()
	decay := p.decay.Load()
	sustain := p.sustain.Load()
	release := p.release.Load()
	cutoff, _ := p.params.GetMappedValue(7) // Get mapped frequency value
	resonance := p.resonance.Load()
	filterType := int(p.filterType.Load())

	// Update envelope parameters for all voices
	p.voiceManager.ApplyToAllVoices(func(voice *audio.Voice) {
		voice.Envelope.SetADSR(attack, decay, sustain, release)
		
		// Apply tuning if available
		if voice.TuningID != 0 {
			voice.Frequency = p.extensions.ApplyTuning(
				audio.NoteToFrequency(int(voice.Key)), 
				int64(voice.TuningID),
				int32(voice.Channel), 
				int32(voice.Key), 
				0,
			)
		}
	})

	// Set oscillator waveform
	p.oscillator.SetWaveform(audio.WaveformType(waveform))
	
	// Update filter parameters - SelectableFilter handles clamping automatically
	p.filter.SetType(audio.MapFilterTypeFromInt(filterType))
	p.filter.SetFrequency(cutoff)
	
	// Map resonance from 0-1 to Q factor 0.5-20
	qFactor := 0.5 + resonance*19.5
	p.filter.SetResonance(qFactor)

	// Generate audio using PolyphonicOscillator
	output := p.oscillator.Process(framesCount)
	
	// Apply filter processing - SelectableFilter handles type switching and safety automatically
	p.filter.ProcessBuffer(output)
	

	// Apply master volume and copy to all output channels
	numChannels := len(audioOut)
	for ch := 0; ch < numChannels; ch++ {
		for i := uint32(0); i < framesCount && i < uint32(len(output)); i++ {
			audioOut[ch][i] = output[i] * float32(volume)
		}
	}

	// Check for finished voices and send note end events
	p.voiceManager.ApplyToAllVoices(func(voice *audio.Voice) {
		if !voice.Envelope.IsActive() && voice.NoteID >= 0 && events != nil {
			endEvent := event.CreateNoteEndEvent(0, voice.NoteID, -1, voice.Channel, voice.Key)
			events.PushOutputEvent(endEvent)
			voice.IsActive = false
		}
	})

	// Return appropriate status
	if p.voiceManager.GetActiveVoiceCount() == 0 {
		return process.ProcessSleep
	}

	return process.ProcessContinue
}

// processEventHandler handles all incoming events using our new EventHandler abstraction
func (p *SynthPlugin) processEventHandler(events *event.EventProcessor, frameCount uint32) {
	if events == nil {
		return
	}

	// Process MIDI events through MIDIProcessor, parameter events through plugin
	p.midiProcessor.ProcessEvents(events, p)
}

// ParamValueToText provides custom formatting for synth parameters
func (p *SynthPlugin) ParamValueToText(paramID uint32, value float64, buffer unsafe.Pointer, size uint32) bool {
	if buffer == nil || size == 0 {
		return false
	}

	// Use parameter binder for automatic formatting
	text, ok := p.params.ValueToText(paramID, value)
	if !ok {
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

	// Use parameter binder for automatic parsing
	parsedValue, err := p.params.TextToValue(paramID, text)
	if err != nil {
		// Use base implementation for unknown parameters
		return p.PluginBase.ParamTextToValue(paramID, text, value)
	}

	*(*float64)(value) = parsedValue
	return true
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

	// Use parameter binder for automatic handling
	p.params.HandleParamValue(paramEvent.ParamID, paramEvent.Value)
}

// handlePolyphonicParameter processes polyphonic parameter changes
func (p *SynthPlugin) handlePolyphonicParameter(paramEvent event.ParamValueEvent) {
	// Apply parameter to matching voices using VoiceManager
	p.voiceManager.ApplyToAllVoices(func(voice *audio.Voice) {
		// Match by note ID if specified
		if paramEvent.NoteID >= 0 && voice.NoteID != paramEvent.NoteID {
			return
		}

		// Match by key/channel if specified
		if paramEvent.Key >= 0 && voice.Key != paramEvent.Key {
			return
		}
		if paramEvent.Channel >= 0 && voice.Channel != paramEvent.Channel {
			return
		}

		// Apply parameter to this voice
		switch paramEvent.ParamID {
		case 1: // Volume modulation
			voice.Volume = paramEvent.Value // VoiceManager uses direct volume, not offset
		case 7: // Pitch bend (new parameter we'll add)
			voice.PitchBend = paramEvent.Value*2.0 - 1.0 // Convert 0-1 to -1 to +1 semitones
		case 8: // Brightness (new parameter)
			voice.Brightness = paramEvent.Value
		case 9: // Pressure (new parameter)
			voice.Pressure = paramEvent.Value
		}
	})
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
	// Parameters are automatically saved by ParamManager
	// Only save non-parameter state if needed
	return map[string]interface{}{
		"plugin_version": "1.0.0",
	}
}

// LoadState loads custom state data for the plugin
func (p *SynthPlugin) LoadState(data map[string]interface{}) {
	// Parameters are automatically loaded by the state loading code
	// Only process non-parameter state if needed
	
	// Could handle version migration here if needed
	if version, ok := data["plugin_version"].(string); ok {
		_ = version // Use for migration if needed
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
	release := p.release.Load()

	// Convert to samples
	tailSamples := uint32(release * p.SampleRate)

	// Add some extra samples for safety
	return tailSamples + uint32(p.SampleRate*0.1) // 100ms extra
}

// OnTimer handles timer callbacks
func (p *SynthPlugin) OnTimer(timerID uint64) {
	// Synth doesn't currently use timers
	// Could be used for UI updates, voice status monitoring, etc.
	p.extensions.LogDebug(fmt.Sprintf("Timer %d fired", timerID))
}

// OnTrackInfoChanged is called when the track information changes
func (p *SynthPlugin) OnTrackInfoChanged() {
	// Get the new track information
	info, ok := p.extensions.GetTrackInfo()
	if !ok {
		p.extensions.LogWarning("Failed to get track info")
		return
	}

	// Log the track information
	p.extensions.LogInfo("Track info changed:")
	if info.Flags&hostpkg.TrackInfoHasTrackName != 0 {
		p.extensions.LogInfo(fmt.Sprintf("  Track name: %s", info.Name))
	}
	if info.Flags&hostpkg.TrackInfoHasTrackColor != 0 {
		p.extensions.LogInfo(fmt.Sprintf("  Track color: R=%d G=%d B=%d A=%d",
			info.Color.Red, info.Color.Green, info.Color.Blue, info.Color.Alpha))
	}
	if info.Flags&hostpkg.TrackInfoHasAudioChannel != 0 {
		p.extensions.LogInfo(fmt.Sprintf("  Audio channels: %d, port type: %s",
			info.AudioChannelCount, info.AudioPortType))
	}

	// Adjust synth behavior based on track type
	if info.Flags&hostpkg.TrackInfoIsForReturnTrack != 0 {
		p.extensions.LogInfo("  This is a return track - adjusting for wet signal")
		// Could adjust default mix to 100% wet
	}
	if info.Flags&hostpkg.TrackInfoIsForBus != 0 {
		p.extensions.LogInfo("  This is a bus track")
		// Could adjust polyphony or processing
	}
	if info.Flags&hostpkg.TrackInfoIsForMaster != 0 {
		p.extensions.LogInfo("  This is the master track")
		// Synths typically wouldn't be on master, but if so, could adjust
	}
}

// OnTuningChanged is called when tunings are added or removed
func (p *SynthPlugin) OnTuningChanged() {
	if !p.extensions.HasTuning() {
		return
	}

	p.extensions.LogInfo("Tuning pool changed, refreshing available tunings")

	// Log all available tunings
	tunings := p.extensions.GetAvailableTunings()
	p.extensions.LogInfo(fmt.Sprintf("Available tunings: %d", len(tunings)))
	for _, t := range tunings {
		p.extensions.LogInfo(fmt.Sprintf("  - %s (ID: %d, Dynamic: %v)",
			t.Name, t.TuningID, t.IsDynamic))
	}
}

// GetVoiceInfo returns voice count and capacity information
func (p *SynthPlugin) GetVoiceInfo() extension.VoiceInfo {
	return extension.VoiceInfo{
		VoiceCount:    uint32(p.voiceManager.GetActiveVoiceCount()),    // Active voices
		VoiceCapacity: 16,                                              // Maximum polyphony
		Flags:         extension.VoiceInfoFlagSupportsOverlappingNotes, // We support note IDs
	}
}

// GetRemoteControlsPageCount returns the number of remote control pages
func (p *SynthPlugin) GetRemoteControlsPageCount() uint32 {
	return 2 // Main controls and Filter controls
}

// GetRemoteControlsPage returns the remote control page at the given index
func (p *SynthPlugin) GetRemoteControlsPage(pageIndex uint32) (*controls.RemoteControlsPage, bool) {
	switch pageIndex {
	case 0:
		// Main controls page with synth parameters
		page := controls.NewRemoteControlsPageBuilder(0, "Main Controls").
			Section("Synth").
			AddParameters(
				1, // Volume
				2, // Waveform
				3, // Attack
				4, // Decay
				5, // Sustain
				6, // Release
			).
			ClearRemaining().
			MustBuild()
		return &page, true
		
	case 1:
		// Filter controls page
		page := controls.FilterControlsPage(1, 
			7, // Cutoff
			8, // Resonance
			0, // No filter envelope amount in this synth
			0, // No LFO in this synth
		).
			ClearRemaining().
			MustBuild()
		return &page, true
		
	default:
		return nil, false
	}
}

// findPeak returns the maximum absolute value in the buffer
func findPeak(buffer []float32) float32 {
	peak := float32(0.0)
	for _, sample := range buffer {
		abs := sample
		if abs < 0 {
			abs = -abs
		}
		if abs > peak {
			peak = abs
		}
	}
	return peak
}

func main() {
	// This is not called when used as a plugin,
	// but can be useful for testing
}

