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
	"encoding/json"
	"fmt"
	"github.com/justyntemme/clapgo/pkg/api"
	"math"
	"os"
	"runtime/cgo"
	"sync/atomic"
	"unsafe"
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

//export ClapGo_GetVersion
func ClapGo_GetVersion(major, minor, patch *C.uint32_t) C.bool {
	if major != nil {
		*major = C.uint32_t(1)
	}
	if minor != nil {
		*minor = C.uint32_t(0)
	}
	if patch != nil {
		*patch = C.uint32_t(0)
	}
	return C.bool(true)
}

//export ClapGo_GetPluginID
func ClapGo_GetPluginID(pluginID *C.char) *C.char {
	return C.CString(synthPlugin.GetPluginID())
}

//export ClapGo_GetPluginName
func ClapGo_GetPluginName(pluginID *C.char) *C.char {
	return C.CString(synthPlugin.GetPluginInfo().Name)
}

//export ClapGo_GetPluginVendor
func ClapGo_GetPluginVendor(pluginID *C.char) *C.char {
	return C.CString(synthPlugin.GetPluginInfo().Vendor)
}

//export ClapGo_GetPluginVersion
func ClapGo_GetPluginVersion(pluginID *C.char) *C.char {
	return C.CString(synthPlugin.GetPluginInfo().Version)
}

//export ClapGo_GetPluginDescription
func ClapGo_GetPluginDescription(pluginID *C.char) *C.char {
	return C.CString(synthPlugin.GetPluginInfo().Description)
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
func ClapGo_PluginProcess(plugin unsafe.Pointer, process unsafe.Pointer) C.int32_t {
	if plugin == nil || process == nil {
		return C.int32_t(api.ProcessError)
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*SynthPlugin)
	
	// Convert the C clap_process_t to Go parameters
	cProcess := (*C.clap_process_t)(process)
	
	// Extract steady time and frame count
	steadyTime := int64(cProcess.steady_time)
	framesCount := uint32(cProcess.frames_count)
	
	// Convert audio buffers using our abstraction - NO MORE MANUAL CONVERSION!
	audioIn := api.ConvertFromCBuffers(unsafe.Pointer(cProcess.audio_inputs), uint32(cProcess.audio_inputs_count), framesCount)
	audioOut := api.ConvertFromCBuffers(unsafe.Pointer(cProcess.audio_outputs), uint32(cProcess.audio_outputs_count), framesCount)
	
	// Create event handler using the new abstraction - NO MORE MANUAL EVENT HANDLING!
	eventHandler := api.NewEventProcessor(
		unsafe.Pointer(cProcess.in_events),
		unsafe.Pointer(cProcess.out_events),
	)
	
	// Call the actual Go process method
	result := p.Process(steadyTime, framesCount, audioIn, audioOut, eventHandler)
	
	return C.int32_t(result)
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

//export ClapGo_PluginParamsCount
func ClapGo_PluginParamsCount(plugin unsafe.Pointer) C.uint32_t {
	if plugin == nil {
		return 0
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	return C.uint32_t(p.paramManager.GetParameterCount())
}

//export ClapGo_PluginParamsGetInfo
func ClapGo_PluginParamsGetInfo(plugin unsafe.Pointer, index C.uint32_t, info unsafe.Pointer) C.bool {
	if plugin == nil || info == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	
	// Get parameter info from manager
	paramInfo, err := p.paramManager.GetParameterInfoByIndex(uint32(index))
	if err != nil {
		return C.bool(false)
	}
	
	// Convert to C struct
	cInfo := (*C.clap_param_info_t)(info)
	cInfo.id = C.clap_id(paramInfo.ID)
	cInfo.flags = C.CLAP_PARAM_IS_AUTOMATABLE | C.CLAP_PARAM_IS_MODULATABLE
	cInfo.cookie = nil
	
	// Copy name
	nameBytes := []byte(paramInfo.Name)
	if len(nameBytes) >= C.CLAP_NAME_SIZE {
		nameBytes = nameBytes[:C.CLAP_NAME_SIZE-1]
	}
	for i, b := range nameBytes {
		cInfo.name[i] = C.char(b)
	}
	cInfo.name[len(nameBytes)] = 0
	
	// Clear module path
	cInfo.module[0] = 0
	
	// Set range
	cInfo.min_value = C.double(paramInfo.MinValue)
	cInfo.max_value = C.double(paramInfo.MaxValue)
	cInfo.default_value = C.double(paramInfo.DefaultValue)
	
	return C.bool(true)
}

//export ClapGo_PluginParamsGetValue
func ClapGo_PluginParamsGetValue(plugin unsafe.Pointer, paramID C.uint32_t, value *C.double) C.bool {
	if plugin == nil || value == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	
	// Get current value from parameter manager
	val := p.paramManager.GetParameterValue(uint32(paramID))
	*value = C.double(val)
	
	return C.bool(true)
}

//export ClapGo_PluginParamsValueToText
func ClapGo_PluginParamsValueToText(plugin unsafe.Pointer, paramID C.uint32_t, value C.double, buffer *C.char, size C.uint32_t) C.bool {
	if plugin == nil || buffer == nil || size == 0 {
		return C.bool(false)
	}
	// Format based on parameter type
	var text string
	switch uint32(paramID) {
	case 1: // Volume
		text = fmt.Sprintf("%.1f %%", float64(value) * 100.0)
	case 2: // Waveform
		switch int(math.Round(float64(value))) {
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
		text = fmt.Sprintf("%.0f ms", float64(value) * 1000.0)
	case 5: // Sustain
		text = fmt.Sprintf("%.1f %%", float64(value) * 100.0)
	default:
		text = fmt.Sprintf("%.3f", float64(value))
	}
	
	// Copy to C buffer
	textBytes := []byte(text)
	maxLen := int(size) - 1
	if len(textBytes) > maxLen {
		textBytes = textBytes[:maxLen]
	}
	
	cBuffer := (*[1 << 30]C.char)(unsafe.Pointer(buffer))
	for i, b := range textBytes {
		cBuffer[i] = C.char(b)
	}
	cBuffer[len(textBytes)] = 0
	
	return C.bool(true)
}

//export ClapGo_PluginParamsTextToValue
func ClapGo_PluginParamsTextToValue(plugin unsafe.Pointer, paramID C.uint32_t, text *C.char, value *C.double) C.bool {
	if plugin == nil || text == nil || value == nil {
		return C.bool(false)
	}
	// For now, just return false - proper implementation would parse the text
	return C.bool(false)
}

//export ClapGo_PluginParamsFlush
func ClapGo_PluginParamsFlush(plugin unsafe.Pointer, inEvents unsafe.Pointer, outEvents unsafe.Pointer) {
	if plugin == nil {
		return
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	
	// Process events using our abstraction
	if inEvents != nil {
		eventHandler := api.NewEventProcessor(inEvents, outEvents)
		p.processEventHandler(eventHandler, 0)
	}
}

//export ClapGo_PluginStateSave
func ClapGo_PluginStateSave(plugin unsafe.Pointer, stream unsafe.Pointer) C.bool {
	if plugin == nil || stream == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	
	out := api.NewOutputStream(stream)
	
	// Write state version
	if err := out.WriteUint32(1); err != nil {
		return false
	}
	
	// Write parameter count
	paramCount := p.paramManager.GetParameterCount()
	if err := out.WriteUint32(paramCount); err != nil {
		return false
	}
	
	// Write each parameter ID and value
	for i := uint32(0); i < paramCount; i++ {
		info, err := p.paramManager.GetParameterInfoByIndex(i)
		if err != nil {
			return false
		}
		
		// Write parameter ID
		if err := out.WriteUint32(info.ID); err != nil {
			return false
		}
		
		// Write parameter value
		value := p.paramManager.GetParameterValue(info.ID)
		if err := out.WriteFloat64(value); err != nil {
			return false
		}
	}
	
	// Save custom state data
	stateData := p.SaveState()
	jsonData, err := json.Marshal(stateData)
	if err != nil {
		return false
	}
	
	// Write JSON length and data
	if err := out.WriteUint32(uint32(len(jsonData))); err != nil {
		return false
	}
	if _, err := out.Write(jsonData); err != nil {
		return false
	}
	
	return C.bool(true)
}

//export ClapGo_PluginNotePortsCount
func ClapGo_PluginNotePortsCount(plugin unsafe.Pointer, isInput C.bool) C.uint32_t {
	if plugin == nil {
		return 0
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	
	// Plugin implements note ports extension directly
	npm := p.GetNotePortManager()
	if npm == nil {
		return 0
	}
	
	if isInput {
		return C.uint32_t(npm.GetInputPortCount())
	}
	return C.uint32_t(npm.GetOutputPortCount())
}

//export ClapGo_PluginNotePortsGet
func ClapGo_PluginNotePortsGet(plugin unsafe.Pointer, index C.uint32_t, isInput C.bool, info unsafe.Pointer) C.bool {
	if plugin == nil || info == nil {
		return false
	}
	
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	
	// Plugin implements note ports extension directly
	npm := p.GetNotePortManager()
	
	if npm == nil {
		return false
	}
	
	var portInfo *api.NotePortInfo
	if isInput {
		portInfo = npm.GetInputPort(uint32(index))
	} else {
		portInfo = npm.GetOutputPort(uint32(index))
	}
	
	if portInfo == nil {
		return false
	}
	
	// Cast to C structure
	cInfo := (*C.clap_note_port_info_t)(info)
	
	// Convert Go NotePortInfo to C structure
	cInfo.id = C.uint32_t(portInfo.ID)
	cInfo.supported_dialects = C.uint32_t(portInfo.SupportedDialects)
	cInfo.preferred_dialect = C.uint32_t(portInfo.PreferredDialect)
	
	// Copy name with null termination
	nameBytes := []byte(portInfo.Name)
	if len(nameBytes) > 255 {
		nameBytes = nameBytes[:255]
	}
	for i, b := range nameBytes {
		cInfo.name[i] = C.char(b)
	}
	cInfo.name[len(nameBytes)] = 0
	
	return true
}


//export ClapGo_PluginStateLoad
func ClapGo_PluginStateLoad(plugin unsafe.Pointer, stream unsafe.Pointer) C.bool {
	if plugin == nil || stream == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	
	in := api.NewInputStream(stream)
	
	// Read state version
	version, err := in.ReadUint32()
	if err != nil || version != 1 {
		return false
	}
	
	// Read parameter count
	paramCount, err := in.ReadUint32()
	if err != nil {
		return false
	}
	
	// Read each parameter ID and value
	for i := uint32(0); i < paramCount; i++ {
		// Read parameter ID
		paramID, err := in.ReadUint32()
		if err != nil {
			return false
		}
		
		// Read parameter value
		value, err := in.ReadFloat64()
		if err != nil {
			return false
		}
		
		// Set parameter value
		p.paramManager.SetParameterValue(paramID, value)
		
		// Update internal state
		switch paramID {
		case 1: // Volume
			atomic.StoreInt64(&p.volume, int64(floatToBits(value)))
		case 2: // Waveform
			atomic.StoreInt64(&p.waveform, int64(math.Round(value)))
		case 3: // Attack
			atomic.StoreInt64(&p.attack, int64(floatToBits(value)))
		case 4: // Decay
			atomic.StoreInt64(&p.decay, int64(floatToBits(value)))
		case 5: // Sustain
			atomic.StoreInt64(&p.sustain, int64(floatToBits(value)))
		case 6: // Release
			atomic.StoreInt64(&p.release, int64(floatToBits(value)))
		}
	}
	
	// Read custom state data
	jsonLength, err := in.ReadUint32()
	if err != nil {
		return false
	}
	
	if jsonLength > 0 {
		jsonData := make([]byte, jsonLength)
		if _, err := in.Read(jsonData); err != nil {
			return false
		}
		
		var stateData map[string]interface{}
		if err := json.Unmarshal(jsonData, &stateData); err == nil {
			p.LoadState(stateData)
		}
	}
	
	return C.bool(true)
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

// SynthPlugin implements a simple synthesizer with atomic parameter storage
type SynthPlugin struct {
	// Plugin state
	voices       [16]*Voice  // Maximum 16 voices
	sampleRate   float64
	isActivated  bool
	isProcessing bool
	host         unsafe.Pointer
	
	// Parameters with atomic storage for thread safety
	volume       int64  // atomic storage for volume (0.0-1.0)
	waveform     int64  // atomic storage for waveform (0-2)
	attack       int64  // atomic storage for attack time
	decay        int64  // atomic storage for decay time  
	sustain      int64  // atomic storage for sustain level
	release      int64  // atomic storage for release time
	
	// Transport state
	transportInfo TransportInfo
	
	// Parameter management using our new abstraction
	paramManager *api.ParameterManager
	
	// Note port management
	notePortManager *api.NotePortManager
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
		transportInfo: TransportInfo{
			Tempo: 120.0,
		},
		paramManager: api.NewParameterManager(),
		notePortManager: api.NewNotePortManager(),
	}
	
	// Set default values atomically
	atomic.StoreInt64(&plugin.volume, int64(floatToBits(0.7)))      // -3dB
	atomic.StoreInt64(&plugin.waveform, 0)                          // sine
	atomic.StoreInt64(&plugin.attack, int64(floatToBits(0.01)))     // 10ms
	atomic.StoreInt64(&plugin.decay, int64(floatToBits(0.1)))       // 100ms
	atomic.StoreInt64(&plugin.sustain, int64(floatToBits(0.7)))     // 70%
	atomic.StoreInt64(&plugin.release, int64(floatToBits(0.3)))     // 300ms
	
	// Register parameters using our new abstraction
	plugin.paramManager.RegisterParameter(api.CreateFloatParameter(1, "Volume", 0.0, 1.0, 0.7))
	plugin.paramManager.RegisterParameter(api.CreateFloatParameter(2, "Waveform", 0.0, 2.0, 0.0))
	plugin.paramManager.RegisterParameter(api.CreateFloatParameter(3, "Attack", 0.001, 2.0, 0.01))
	plugin.paramManager.RegisterParameter(api.CreateFloatParameter(4, "Decay", 0.001, 2.0, 0.1))
	plugin.paramManager.RegisterParameter(api.CreateFloatParameter(5, "Sustain", 0.0, 1.0, 0.7))
	plugin.paramManager.RegisterParameter(api.CreateFloatParameter(6, "Release", 0.001, 5.0, 0.3))
	
	// Configure note port for instrument
	plugin.notePortManager.AddInputPort(api.CreateDefaultInstrumentPort())
	
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

// Process processes audio data using the new abstractions
func (p *SynthPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events api.EventHandler) int {
	// Check if we're in a valid state for processing
	if !p.isActivated || !p.isProcessing {
		return api.ProcessError
	}
	
	// Process events using our new abstraction - NO MORE MANUAL EVENT PARSING!
	if events != nil {
		p.processEventHandler(events, framesCount)
	}
	
	// If no outputs, nothing to do
	if len(audioOut) == 0 {
		return api.ProcessContinue
	}
	
	// Get current parameter values atomically
	volume := floatFromBits(uint64(atomic.LoadInt64(&p.volume)))
	waveform := int(atomic.LoadInt64(&p.waveform))
	attack := floatFromBits(uint64(atomic.LoadInt64(&p.attack)))
	decay := floatFromBits(uint64(atomic.LoadInt64(&p.decay)))
	sustain := floatFromBits(uint64(atomic.LoadInt64(&p.sustain)))
	release := floatFromBits(uint64(atomic.LoadInt64(&p.release)))
	
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
			
			// Calculate frequency for this note
			freq := noteToFrequency(int(voice.Key))
			
			// Generate audio for this voice
			for j := uint32(0); j < framesCount; j++ {
				// Get envelope value
				env := p.getEnvelopeValue(voice, j, framesCount, attack, decay, sustain, release)
				
				// Generate sample
				sample := generateSample(voice.Phase, freq, waveform) * env * voice.Velocity
				
				// Apply master volume
				sample *= volume
				
				// Add to all output channels
				for ch := 0; ch < numChannels; ch++ {
					if len(audioOut[ch]) > int(j) {
						audioOut[ch][j] += float32(sample)
					}
				}
				
				// Advance oscillator phase
				voice.Phase += freq / p.sampleRate
				if voice.Phase >= 1.0 {
					voice.Phase -= 1.0
				}
			}
			
			// Check if voice is still active
			if voice.ReleasePhase >= 1.0 {
				// Send note end event to host if we have a valid note ID
				if voice.NoteID >= 0 && events != nil {
					endEvent := api.CreateNoteEndEvent(0, voice.NoteID, -1, voice.Channel, voice.Key)
					events.PushOutputEvent(endEvent)
				}
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

// processEventHandler handles all incoming events using our new EventHandler abstraction
func (p *SynthPlugin) processEventHandler(events api.EventHandler, frameCount uint32) {
	if events == nil {
		return
	}
	
	// Process each event using our abstraction - NO MORE MANUAL C STRUCT PARSING!
	eventCount := events.GetInputEventCount()
	for i := uint32(0); i < eventCount; i++ {
		event := events.GetInputEvent(i)
		if event == nil {
			continue
		}
		
		// Handle each event type safely using our abstraction
		switch event.Type {
		case api.EventTypeParamValue:
			if paramEvent, ok := event.Data.(api.ParamValueEvent); ok {
				p.handleParameterChange(paramEvent)
			}
		case api.EventTypeNoteOn:
			if noteEvent, ok := event.Data.(api.NoteEvent); ok {
				p.handleNoteOn(noteEvent, event.Time)
			}
		case api.EventTypeNoteOff:
			if noteEvent, ok := event.Data.(api.NoteEvent); ok {
				p.handleNoteOff(noteEvent, event.Time)
			}
		}
	}
}

// handleParameterChange processes a parameter change event
func (p *SynthPlugin) handleParameterChange(paramEvent api.ParamValueEvent) {
	// Handle the parameter change based on its ID
	switch paramEvent.ParamID {
	case 1: // Volume
		atomic.StoreInt64(&p.volume, int64(floatToBits(paramEvent.Value)))
		p.paramManager.SetParameterValue(paramEvent.ParamID, paramEvent.Value)
	case 2: // Waveform
		atomic.StoreInt64(&p.waveform, int64(math.Round(paramEvent.Value)))
		p.paramManager.SetParameterValue(paramEvent.ParamID, paramEvent.Value)
	case 3: // Attack
		atomic.StoreInt64(&p.attack, int64(floatToBits(paramEvent.Value)))
		p.paramManager.SetParameterValue(paramEvent.ParamID, paramEvent.Value)
	case 4: // Decay
		atomic.StoreInt64(&p.decay, int64(floatToBits(paramEvent.Value)))
		p.paramManager.SetParameterValue(paramEvent.ParamID, paramEvent.Value)
	case 5: // Sustain
		atomic.StoreInt64(&p.sustain, int64(floatToBits(paramEvent.Value)))
		p.paramManager.SetParameterValue(paramEvent.ParamID, paramEvent.Value)
	case 6: // Release
		atomic.StoreInt64(&p.release, int64(floatToBits(paramEvent.Value)))
		p.paramManager.SetParameterValue(paramEvent.ParamID, paramEvent.Value)
	}
}

// handleNoteOn processes a note on event
func (p *SynthPlugin) handleNoteOn(noteEvent api.NoteEvent, time uint32) {
	// Validate note event fields (CLAP spec says key must be 0-127)
	if noteEvent.Key < 0 || noteEvent.Key > 127 {
		return
	}
	
	// Ensure velocity is positive
	velocity := noteEvent.Velocity
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
func (p *SynthPlugin) handleNoteOff(noteEvent api.NoteEvent, time uint32) {
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
			// Start the release phase (0.0 = start of release)
			voice.ReleasePhase = 0.0
		}
	}
}

// handleNoteChoke processes a note choke event (immediate stop)
func (p *SynthPlugin) handleNoteChoke(noteEvent api.NoteEvent, time uint32) {
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

// handleMIDI processes MIDI 1.0 events
func (p *SynthPlugin) handleMIDI(midiEvent api.MIDIEvent, time uint32) {
	if len(midiEvent.Data) < 3 {
		return
	}
	
	status := midiEvent.Data[0] & 0xF0
	channel := midiEvent.Data[0] & 0x0F
	
	switch status {
	case 0x90: // Note On
		if midiEvent.Data[2] > 0 { // Velocity > 0
			noteEvent := api.NoteEvent{
				NoteID:   -1,
				Port:     midiEvent.Port,
				Channel:  int16(channel),
				Key:      int16(midiEvent.Data[1]),
				Velocity: float64(midiEvent.Data[2]) / 127.0,
			}
			p.handleNoteOn(noteEvent, time)
		} else { // Velocity = 0 is Note Off
			noteEvent := api.NoteEvent{
				NoteID:   -1,
				Port:     midiEvent.Port,
				Channel:  int16(channel),
				Key:      int16(midiEvent.Data[1]),
				Velocity: 0.0,
			}
			p.handleNoteOff(noteEvent, time)
		}
		
	case 0x80: // Note Off
		noteEvent := api.NoteEvent{
			NoteID:   -1,
			Port:     midiEvent.Port,
			Channel:  int16(channel),
			Key:      int16(midiEvent.Data[1]),
			Velocity: float64(midiEvent.Data[2]) / 127.0,
		}
		p.handleNoteOff(noteEvent, time)
		
	case 0xB0: // Control Change
		// Handle common CCs
		switch midiEvent.Data[1] {
		case 7: // Volume
			value := float64(midiEvent.Data[2]) / 127.0
			paramEvent := api.ParamValueEvent{
				ParamID: 1, // Volume parameter
				Value:   value,
			}
			p.handleParameterChange(paramEvent)
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
func (p *SynthPlugin) getEnvelopeValue(voice *Voice, sampleIndex, frameCount uint32, attack, decay, sustain, release float64) float64 {
	// If in release phase
	if voice.ReleasePhase >= 0.0 {
		// Update release phase
		releaseSamples := release * p.sampleRate
		voice.ReleasePhase += 1.0 / releaseSamples
		
		// Calculate release envelope (exponential decay)
		return math.Pow(1.0 - voice.ReleasePhase, 2.0) * sustain
	}
	
	// Attack phase
	attackSamples := attack * p.sampleRate
	if attackSamples <= 0 {
		attackSamples = 1 // Prevent division by zero
	}
	
	// Decay phase
	decaySamples := decay * p.sampleRate
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

// GetNotePortManager returns the plugin's note port manager
func (p *SynthPlugin) GetNotePortManager() *api.NotePortManager {
	return p.notePortManager
}

// GetExtension gets a plugin extension
func (p *SynthPlugin) GetExtension(id string) unsafe.Pointer {
	// All extensions are handled by the C bridge based on exported functions
	// The C bridge checks for the presence of required exports and provides
	// the extension implementation if available
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
		"waveform":      atomic.LoadInt64(&p.waveform),
		"attack":        floatFromBits(uint64(atomic.LoadInt64(&p.attack))),
		"decay":         floatFromBits(uint64(atomic.LoadInt64(&p.decay))),
		"sustain":       floatFromBits(uint64(atomic.LoadInt64(&p.sustain))),
		"release":       floatFromBits(uint64(atomic.LoadInt64(&p.release))),
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
		atomic.StoreInt64(&p.attack, int64(floatToBits(attack)))
	}
	if decay, ok := data["decay"].(float64); ok {
		atomic.StoreInt64(&p.decay, int64(floatToBits(decay)))
	}
	if sustain, ok := data["sustain"].(float64); ok {
		atomic.StoreInt64(&p.sustain, int64(floatToBits(sustain)))
	}
	if release, ok := data["release"].(float64); ok {
		atomic.StoreInt64(&p.release, int64(floatToBits(release)))
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
	atomic.StoreInt64(&p.waveform, int64(preset.Waveform))
	atomic.StoreInt64(&p.attack, int64(floatToBits(preset.Attack)))
	atomic.StoreInt64(&p.decay, int64(floatToBits(preset.Decay)))
	atomic.StoreInt64(&p.sustain, int64(floatToBits(preset.Sustain)))
	atomic.StoreInt64(&p.release, int64(floatToBits(preset.Release)))
	
	// Load any additional state data
	if preset.StateData != nil {
		p.LoadState(preset.StateData)
	}
	
	return true
}

// Helper functions for atomic float64 operations
func floatToBits(f float64) uint64 {
	return *(*uint64)(unsafe.Pointer(&f))
}

func floatFromBits(b uint64) float64 {
	return *(*float64)(unsafe.Pointer(&b))
}

func main() {
	// This is not called when used as a plugin,
	// but can be useful for testing
}