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
	filter         *audio.StateVariableFilter
	midiProcessor  *audio.MIDIProcessor

	// Parameters with atomic storage for thread safety
	volume   *param.AtomicFloat64
	waveform *param.AtomicFloat64
	attack   *param.AtomicFloat64
	decay    *param.AtomicFloat64
	sustain  *param.AtomicFloat64
	release  *param.AtomicFloat64

	// Filter parameters
	filterCutoff    *param.AtomicFloat64
	filterResonance *param.AtomicFloat64
	filterDrive     *param.AtomicFloat64
	filterType      *param.AtomicFloat64

	// Transport state
	transportInfo TransportInfo

	// Note port management
	notePortManager *audio.NotePortManager

	// Remote controls management
	remoteControlsManager *controls.RemoteControlsManager

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

	// Create voice manager first
	voiceManager := audio.NewVoiceManager(16, 44100) // 16 voice polyphony

	plugin := &SynthPlugin{
		PluginBase: plugin.NewPluginBase(pluginInfo),
		transportInfo: TransportInfo{
			Tempo: 120.0,
		},
		notePortManager:       audio.NewNotePortManager(),
		remoteControlsManager: controls.NewRemoteControlsManager(nil), // Will be initialized in Init
		poolDiagnostics:       &event.Diagnostics{},
		voiceManager:          voiceManager,
		oscillator:            audio.NewPolyphonicOscillator(voiceManager),
		filter:                audio.NewStateVariableFilter(44100),
	}
	
	// Create MIDI processor
	plugin.midiProcessor = audio.NewMIDIProcessor(voiceManager, plugin.ParamManager)

	// Initialize atomic parameters
	plugin.volume = param.NewAtomicFloat64(0.7)    // -3dB
	plugin.waveform = param.NewAtomicFloat64(0)    // sine
	plugin.attack = param.NewAtomicFloat64(0.01)   // 10ms
	plugin.decay = param.NewAtomicFloat64(0.1)     // 100ms
	plugin.sustain = param.NewAtomicFloat64(0.7)   // 70%
	plugin.release = param.NewAtomicFloat64(0.3)   // 300ms

	// Initialize filter parameters using framework utilities
	plugin.filterCutoff = param.NewAtomicFloat64(1000.0)
	plugin.filterResonance = param.NewAtomicFloat64(1.0)
	plugin.filterDrive = param.NewAtomicFloat64(0.0)
	plugin.filterType = param.NewAtomicFloat64(0.0)

	// Define all parameters using framework builders
	allParams := []param.Info{
		// Synth parameters
		param.NewBuilder(1, "Volume").
			Module("Main").
			Range(0, 1, 0.7).
			Format(param.FormatPercentage).
			Automatable().
			MustBuild(),
		
		param.NewBuilder(2, "Waveform").
			Module("Oscillator").
			Range(0, 2, 0).
			Stepped().
			Automatable().
			MustBuild(),

		// ADSR parameters
		param.NewBuilder(3, "Attack").
			Module("Envelope").
			Range(0.001, 2.0, 0.01).
			Format(param.FormatMilliseconds).
			Automatable().
			MustBuild(),

		param.NewBuilder(4, "Decay").
			Module("Envelope").
			Range(0.001, 2.0, 0.1).
			Format(param.FormatMilliseconds).
			Automatable().
			MustBuild(),

		param.NewBuilder(5, "Sustain").
			Module("Envelope").
			Range(0, 1, 0.7).
			Format(param.FormatPercentage).
			Automatable().
			MustBuild(),

		param.NewBuilder(6, "Release").
			Module("Envelope").
			Range(0.001, 5.0, 0.3).
			Format(param.FormatMilliseconds).
			Automatable().
			MustBuild(),

		// Filter parameters - expose to DAW
		param.NewBuilder(7, "Cutoff").
			Module("Filter").
			Range(20, 20000, 1000).
			Format(param.FormatHertz).
			Automatable().
			Modulatable().
			MustBuild(),
			
		param.NewBuilder(8, "Resonance").
			Module("Filter").
			Range(0.5, 20, 1).
			Automatable().
			Modulatable().
			MustBuild(),
			
		param.NewBuilder(9, "Drive").
			Module("Filter").
			Range(0, 100, 0).
			Format(param.FormatPercentage).
			Automatable().
			MustBuild(),
			
		param.NewBuilder(10, "Type").
			Module("Filter").
			Range(0, 3, 0).
			Stepped().
			Automatable().
			MustBuild(),
	}

	// Register all parameters at once using framework
	plugin.ParamManager.RegisterAll(allParams...)

	// Set up parameter change handling
	plugin.setupParameterHandling()

	// Configure note port for instrument
	plugin.notePortManager.AddInputPort(audio.CreateDefaultInstrumentPort())

	return plugin
}

// setupParameterHandling configures parameter initialization
func (p *SynthPlugin) setupParameterHandling() {
	// Parameter changes are handled through HandleParamValue method
	// This method can be used for additional parameter setup if needed
}

// setupRemoteControls configures MIDI CC mapping using framework
func (p *SynthPlugin) setupRemoteControls() {
	if p.Host == nil {
		return
	}

	// Reinitialize remote controls manager with host
	p.remoteControlsManager = controls.NewRemoteControlsManager(unsafe.Pointer(p.Host))

	// Create filter control page using framework builder
	filterPage := controls.NewRemoteControlsPageBuilder(0, "Filter Controls").
		Section("Filter").
		SetParameter(0, 7).  // CC74 -> Filter Cutoff (Brightness)
		SetParameter(1, 8).  // CC71 -> Filter Resonance (Harmonic Content) 
		SetParameter(2, 9).  // CC76 -> Filter Drive (Sound Variation)
		SetParameter(3, 10). // CC77 -> Filter Type
		MustBuild()

	// Add the filter control page
	p.remoteControlsManager.AddPage(filterPage)

	// Create main controls page
	mainPage := controls.NewRemoteControlsPageBuilder(1, "Main Controls").
		Section("Main").
		SetParameter(0, 1). // Volume
		SetParameter(1, 2). // Waveform
		SetParameter(2, 3). // Attack
		SetParameter(3, 4). // Decay
		SetParameter(4, 5). // Sustain
		SetParameter(5, 6). // Release
		MustBuild()

	// Add the main control page
	p.remoteControlsManager.AddPage(mainPage)

	// Notify host that pages are available
	p.remoteControlsManager.NotifyChanged()
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
	
	// Set up custom MIDI callbacks
	p.midiProcessor.SetCallbacks(
		// onNoteOn - handle transport control via C0
		func(channel, key int16, velocity float64) {
			// Special transport control: C0 (MIDI note 24) toggles play/pause
			if key == 24 && p.transportControl != nil {
				p.transportControl.RequestTogglePlay()
				if p.Logger != nil {
					p.Logger.Info("Transport toggle play requested via C0")
				}
			}
		},
		nil, // onNoteOff
		// onModulation - handle CC changes for volume and filter
		func(channel int16, cc uint32, value float64) {
			switch cc {
			case 7: // Volume CC
				p.volume.UpdateWithManager(value, p.ParamManager, 1)
			case 74: // Filter Cutoff (Brightness)
				cutoff := 20.0 + (value * 19980.0) // 20Hz to 20kHz
				p.filterCutoff.UpdateWithManager(cutoff, p.ParamManager, 7)
				if p.filter != nil {
					p.filter.SetFrequency(cutoff)
				}
			case 71: // Filter Resonance (Harmonic Content)
				resonance := 0.5 + (value * 19.5) // 0.5 to 20
				p.filterResonance.UpdateWithManager(resonance, p.ParamManager, 8)
				if p.filter != nil {
					p.filter.SetResonance(resonance)
				}
			case 76: // Filter Drive (Sound Variation)
				drive := value * 100.0 // 0-100%
				p.filterDrive.UpdateWithManager(drive, p.ParamManager, 9)
			case 77: // Filter Type
				filterType := math.Floor(value * 4.0) // 0-3
				if filterType > 3 {
					filterType = 3
				}
				p.filterType.UpdateWithManager(filterType, p.ParamManager, 10)
			}
		},
		nil, // onPitchBend
		nil, // onPolyPressure
	)

	// Set up remote controls for MIDI CC mapping
	p.setupRemoteControls()

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

	// Update voice manager and filter sample rate
	p.voiceManager.SetSampleRate(sampleRate)
	p.filter = audio.NewStateVariableFilter(sampleRate)
	
	// Initialize filter with current parameter values
	p.filter.SetFrequency(p.filterCutoff.Load())
	p.filter.SetResonance(p.filterResonance.Load())

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
	
	// Get filter parameters
	filterCutoff := p.filterCutoff.Load()
	filterResonance := p.filterResonance.Load()
	filterDrive := p.filterDrive.Load() / 100.0 // Convert percentage to 0-1
	filterType := p.filterType.Load()

	// Update envelope parameters for all voices
	p.voiceManager.ApplyToAllVoices(func(voice *audio.Voice) {
		voice.Envelope.SetADSR(attack, decay, sustain, release)
		
		// Apply tuning if available
		if p.tuning != nil && voice.TuningID != 0 {
			voice.Frequency = p.tuning.ApplyTuning(
				audio.NoteToFrequency(int(voice.Key)), 
				voice.TuningID,
				int32(voice.Channel), 
				int32(voice.Key), 
				0,
			)
		}
	})

	// Set oscillator waveform
	p.oscillator.SetWaveform(audio.WaveformType(waveform))

	// Generate audio using PolyphonicOscillator
	output := p.oscillator.Process(framesCount)

	// Apply filter processing using framework's StateVariableFilter
	if p.filter != nil {
		// Update filter parameters
		p.filter.SetFrequency(filterCutoff)
		p.filter.SetResonance(filterResonance)
		
		// Apply drive/distortion using framework DSP utilities
		if filterDrive > 0 {
			for i := range output {
				// Apply drive
				driven := output[i] * float32(1.0 + filterDrive*4.0)
				// Use framework's soft clipping
				output[i] = float32(math.Tanh(float64(driven)))
			}
		}
		
		// Process through filter based on filter type
		for i := range output {
			lp, hp, bp, notch := p.filter.Process(float64(output[i]))
			switch int(filterType) {
			case 0: // Low Pass
				output[i] = float32(lp)
			case 1: // High Pass
				output[i] = float32(hp)
			case 2: // Band Pass
				output[i] = float32(bp)
			case 3: // Notch
				output[i] = float32(notch)
			}
		}
	}

	// Apply master volume and copy to all output channels using framework pattern
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
	case 7: // Filter Cutoff
		text = param.FormatValue(value, param.FormatHertz)
	case 8: // Filter Resonance
		text = fmt.Sprintf("%.1f", value)
	case 9: // Filter Drive
		text = param.FormatValue(value, param.FormatPercentage)
	case 10: // Filter Type
		switch int(math.Round(value)) {
		case 0:
			text = "Low Pass"
		case 1:
			text = "High Pass"
		case 2:
			text = "Band Pass"
		case 3:
			text = "Notch"
		default:
			text = "Unknown"
		}
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
	case 7: // Filter Cutoff (Hz)
		parser := param.NewParser(param.FormatHertz)
		if parsedValue, err := parser.ParseValue(text); err == nil {
			*(*float64)(value) = param.ClampValue(parsedValue, 20.0, 20000.0)
			return true
		}
	case 8: // Filter Resonance
		var parsedValue float64
		if _, err := fmt.Sscanf(text, "%f", &parsedValue); err == nil {
			*(*float64)(value) = param.ClampValue(parsedValue, 0.5, 20.0)
			return true
		}
	case 9: // Filter Drive (percentage)
		parser := param.NewParser(param.FormatPercentage)
		if parsedValue, err := parser.ParseValue(text); err == nil {
			*(*float64)(value) = param.ClampValue(parsedValue, 0.0, 100.0)
			return true
		}
	case 10: // Filter Type
		switch text {
		case "Low Pass":
			*(*float64)(value) = 0.0
			return true
		case "High Pass":
			*(*float64)(value) = 1.0
			return true
		case "Band Pass":
			*(*float64)(value) = 2.0
			return true
		case "Notch":
			*(*float64)(value) = 3.0
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
		p.volume.UpdateWithManager(paramEvent.Value, p.ParamManager, paramEvent.ParamID)
	case 2: // Waveform
		p.waveform.UpdateWithManager(math.Round(paramEvent.Value), p.ParamManager, paramEvent.ParamID)
	case 3: // Attack
		p.attack.UpdateWithManager(paramEvent.Value, p.ParamManager, paramEvent.ParamID)
	case 4: // Decay
		p.decay.UpdateWithManager(paramEvent.Value, p.ParamManager, paramEvent.ParamID)
	case 5: // Sustain
		p.sustain.UpdateWithManager(paramEvent.Value, p.ParamManager, paramEvent.ParamID)
	case 6: // Release
		p.release.UpdateWithManager(paramEvent.Value, p.ParamManager, paramEvent.ParamID)
	case 7: // Filter Cutoff
		p.filterCutoff.UpdateWithManager(paramEvent.Value, p.ParamManager, paramEvent.ParamID)
		if p.filter != nil {
			p.filter.SetFrequency(paramEvent.Value)
		}
	case 8: // Filter Resonance
		p.filterResonance.UpdateWithManager(paramEvent.Value, p.ParamManager, paramEvent.ParamID)
		if p.filter != nil {
			p.filter.SetResonance(paramEvent.Value)
		}
	case 9: // Filter Drive
		p.filterDrive.UpdateWithManager(paramEvent.Value, p.ParamManager, paramEvent.ParamID)
	case 10: // Filter Type
		p.filterType.UpdateWithManager(paramEvent.Value, p.ParamManager, paramEvent.ParamID)
	}
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
	return extension.VoiceInfo{
		VoiceCount:    uint32(p.voiceManager.GetActiveVoiceCount()),    // Active voices
		VoiceCapacity: 16,                                              // Maximum polyphony
		Flags:         extension.VoiceInfoFlagSupportsOverlappingNotes, // We support note IDs
	}
}

func main() {
	// This is not called when used as a plugin,
	// but can be useful for testing
}

