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
	"embed"
	"encoding/json"
	"fmt"
	"github.com/justyntemme/clapgo/pkg/api"
	"github.com/justyntemme/clapgo/pkg/audio"
	"github.com/justyntemme/clapgo/pkg/event"
	"github.com/justyntemme/clapgo/pkg/extension"
	hostpkg "github.com/justyntemme/clapgo/pkg/host"
	"github.com/justyntemme/clapgo/pkg/param"
	"github.com/justyntemme/clapgo/pkg/process"
	"github.com/justyntemme/clapgo/pkg/thread"
	"github.com/justyntemme/clapgo/pkg/util"
	"math"
	"os"
	"path/filepath"
	"runtime/cgo"
	"sync/atomic"
	"unsafe"
)

// This example demonstrates a simple synthesizer plugin using the new API abstractions.

// Embed factory presets
//go:embed presets/factory/*.json
var factoryPresets embed.FS

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
		// Store the host pointer and create utilities
		synthPlugin.host = host
		synthPlugin.logger = hostpkg.NewLogger(host)
		
		// Log plugin creation
		if synthPlugin.logger != nil {
			synthPlugin.logger.Info("Creating synth plugin instance")
		}
		
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
	if p.Init() {
		// Register as voice info provider after successful init
		// Voice info provider registration moved to extension system
		return C.bool(true)
	}
	return C.bool(false)
}

//export ClapGo_PluginDestroy
func ClapGo_PluginDestroy(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	handle := cgo.Handle(plugin)
	p := handle.Value().(*SynthPlugin)
	// Unregister voice info provider before destroying
	// Voice info provider unregistration moved to extension system
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
func ClapGo_PluginProcess(plugin unsafe.Pointer, processPtr unsafe.Pointer) C.int32_t {
	// Mark this thread as audio thread for debug builds
	api.DebugMarkAudioThread()
	defer api.DebugUnmarkAudioThread()
	
	if plugin == nil || processPtr == nil {
		return C.int32_t(process.ProcessError)
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*SynthPlugin)
	
	// Convert the C clap_process_t to Go parameters
	cProcess := (*C.clap_process_t)(processPtr)
	
	// Extract steady time and frame count
	steadyTime := int64(cProcess.steady_time)
	framesCount := uint32(cProcess.frames_count)
	
	// Convert audio buffers using our abstraction - NO MORE MANUAL CONVERSION!
	audioIn := api.ConvertFromCBuffers(unsafe.Pointer(cProcess.audio_inputs), uint32(cProcess.audio_inputs_count), framesCount)
	audioOut := api.ConvertFromCBuffers(unsafe.Pointer(cProcess.audio_outputs), uint32(cProcess.audio_outputs_count), framesCount)
	
	// Create event handler using the new abstraction - NO MORE MANUAL EVENT HANDLING!
	eventHandler := event.NewEventProcessor(
		unsafe.Pointer(cProcess.in_events),
		unsafe.Pointer(cProcess.out_events),
	)
	
	// Setup event pool logging
	// TODO: Update when logger types are unified
	
	// Call the actual Go process method
	result := p.Process(steadyTime, framesCount, audioIn, audioOut, eventHandler)
	
	// Log event pool diagnostics periodically (every 1000 calls)
	if p.poolDiagnostics != nil {
		p.poolDiagnostics.LogPoolDiagnostics(eventHandler, 1000)
	}
	
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
	return C.uint32_t(p.paramManager.Count())
}

//export ClapGo_PluginParamsGetInfo
func ClapGo_PluginParamsGetInfo(plugin unsafe.Pointer, index C.uint32_t, info unsafe.Pointer) C.bool {
	if plugin == nil || info == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	
	// Get parameter info from manager
	paramInfo, err := p.paramManager.GetInfoByIndex(uint32(index))
	if err != nil {
		return C.bool(false)
	}
	
	// Convert to C struct using helper
	param.InfoToC(paramInfo, info)
	
	return C.bool(true)
}

//export ClapGo_PluginParamsGetValue
func ClapGo_PluginParamsGetValue(plugin unsafe.Pointer, paramID C.uint32_t, value *C.double) C.bool {
	if plugin == nil || value == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	
	// Get current value from parameter manager
	val := p.paramManager.Get(uint32(paramID))
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
		text = param.FormatValue(float64(value), param.FormatPercentage)
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
		text = param.FormatValue(float64(value), param.FormatMilliseconds)
	case 5: // Sustain
		text = param.FormatValue(float64(value), param.FormatPercentage)
	default:
		text = param.FormatValue(float64(value), param.FormatDefault)
	}
	
	// Copy string to C buffer
	bytes := []byte(text)
	if len(bytes) >= int(size) {
		bytes = bytes[:size-1]
	}
	for i, b := range bytes {
		*(*C.char)(unsafe.Add(unsafe.Pointer(buffer), i)) = C.char(b)
	}
	*(*C.char)(unsafe.Add(unsafe.Pointer(buffer), len(bytes))) = 0
	
	return C.bool(true)
}

//export ClapGo_PluginParamsTextToValue
func ClapGo_PluginParamsTextToValue(plugin unsafe.Pointer, paramID C.uint32_t, text *C.char, value *C.double) C.bool {
	if plugin == nil || text == nil || value == nil {
		return C.bool(false)
	}
	
	// Convert text to Go string
	goText := C.GoString(text)
	
	// Parse based on parameter type
	switch uint32(paramID) {
	case 1: // Volume (percentage)
		parser := param.NewParser(param.FormatPercentage)
		if parsedValue, err := parser.ParseValue(goText); err == nil {
			*value = C.double(param.ClampValue(parsedValue, 0.0, 1.0))
			return C.bool(true)
		}
	case 2: // Waveform
		// Parse waveform names
		switch goText {
		case "Sine":
			*value = C.double(0.0)
			return C.bool(true)
		case "Saw":
			*value = C.double(1.0)
			return C.bool(true)
		case "Square":
			*value = C.double(2.0)
			return C.bool(true)
		}
	case 3, 4, 6: // Attack, Decay, Release (milliseconds)
		parser := param.NewParser(param.FormatMilliseconds)
		if parsedValue, err := parser.ParseValue(goText); err == nil {
			*value = C.double(param.ClampValue(parsedValue, 0.001, 5.0))
			return C.bool(true)
		}
	case 5: // Sustain (percentage)
		parser := param.NewParser(param.FormatPercentage)
		if parsedValue, err := parser.ParseValue(goText); err == nil {
			*value = C.double(param.ClampValue(parsedValue, 0.0, 1.0))
			return C.bool(true)
		}
	}
	
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
		eventHandler := event.NewEventProcessor(inEvents, outEvents)
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
	paramCount := p.paramManager.Count()
	if err := out.WriteUint32(paramCount); err != nil {
		return false
	}
	
	// Write each parameter ID and value
	for i := uint32(0); i < paramCount; i++ {
		info, err := p.paramManager.GetInfoByIndex(i)
		if err != nil {
			return false
		}
		
		// Write parameter ID
		if err := out.WriteUint32(info.ID); err != nil {
			return false
		}
		
		// Write parameter value
		value := p.paramManager.Get(info.ID)
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
		
		// Update parameter using helper
		switch paramID {
		case 1: // Volume
			param.UpdateParameterAtomic(&p.volume, value, p.paramManager, paramID)
		case 2: // Waveform
			atomic.StoreInt64(&p.waveform, int64(math.Round(value)))
			p.paramManager.Set(paramID, value)
		case 3: // Attack
			param.UpdateParameterAtomic(&p.attack, value, p.paramManager, paramID)
		case 4: // Decay
			param.UpdateParameterAtomic(&p.decay, value, p.paramManager, paramID)
		case 5: // Sustain
			param.UpdateParameterAtomic(&p.sustain, value, p.paramManager, paramID)
		case 6: // Release
			param.UpdateParameterAtomic(&p.release, value, p.paramManager, paramID)
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

// State Context Extension Exports

//export ClapGo_PluginStateSaveWithContext
func ClapGo_PluginStateSaveWithContext(plugin unsafe.Pointer, stream unsafe.Pointer, contextType C.uint32_t) C.bool {
	if plugin == nil || stream == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	
	// Log the context type
	if p.logger != nil {
		switch contextType {
		case C.uint32_t(api.StateContextForPreset):
			p.logger.Info("Saving state for preset")
		case C.uint32_t(api.StateContextForDuplicate):
			p.logger.Info("Saving state for duplicate")
		case C.uint32_t(api.StateContextForProject):
			p.logger.Info("Saving state for project")
		default:
			p.logger.Info(fmt.Sprintf("Saving state with unknown context: %d", contextType))
		}
	}
	
	// For preset saves, we might want to clear voice state
	// For duplicate/project saves, we keep everything
	if contextType == C.uint32_t(api.StateContextForPreset) {
		// Save without active voice data for presets
		return ClapGo_PluginStateSave(plugin, stream)
	}
	
	// For other contexts, save everything including voice state
	return ClapGo_PluginStateSave(plugin, stream)
}

//export ClapGo_PluginStateLoadWithContext
func ClapGo_PluginStateLoadWithContext(plugin unsafe.Pointer, stream unsafe.Pointer, contextType C.uint32_t) C.bool {
	if plugin == nil || stream == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	
	// Log the context type
	if p.logger != nil {
		switch contextType {
		case C.uint32_t(api.StateContextForPreset):
			p.logger.Info("Loading state for preset")
		case C.uint32_t(api.StateContextForDuplicate):
			p.logger.Info("Loading state for duplicate")
		case C.uint32_t(api.StateContextForProject):
			p.logger.Info("Loading state for project")
		default:
			p.logger.Info(fmt.Sprintf("Loading state with unknown context: %d", contextType))
		}
	}
	
	// Load the state
	result := ClapGo_PluginStateLoad(plugin, stream)
	
	// For preset loads, we might want to reset voice allocation
	if contextType == C.uint32_t(api.StateContextForPreset) && result {
		// Clear all active voices when loading a preset
		// Note: This assumes preset loads happen on the main thread
		// Additional synchronization may be needed for multi-threaded hosts
		for i := range p.voices {
			if p.voices[i] != nil {
				p.voices[i].IsActive = false
			}
		}
	}
	
	return result
}

// Preset Load Extension Exports

//export ClapGo_PluginPresetLoadFromLocation
func ClapGo_PluginPresetLoadFromLocation(plugin unsafe.Pointer, locationKind C.uint32_t, location *C.char, loadKey *C.char) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	
	// Convert C strings to Go strings
	var locationStr, loadKeyStr string
	if location != nil {
		locationStr = C.GoString(location)
	}
	if loadKey != nil {
		loadKeyStr = C.GoString(loadKey)
	}
	
	// Log the preset load request
	if p.logger != nil {
		switch locationKind {
		case C.uint32_t(api.PresetLocationFile):
			p.logger.Info(fmt.Sprintf("Loading preset from file: %s (key: %s)", locationStr, loadKeyStr))
		case C.uint32_t(api.PresetLocationPlugin):
			p.logger.Info(fmt.Sprintf("Loading bundled preset (key: %s)", loadKeyStr))
		default:
			p.logger.Warning(fmt.Sprintf("Unknown preset location kind: %d", locationKind))
			return C.bool(false)
		}
	}
	
	// Handle different location kinds
	switch locationKind {
	case C.uint32_t(api.PresetLocationFile):
		// Load preset from file
		if locationStr == "" {
			return C.bool(false)
		}
		
		// Read the preset file
		presetData, err := os.ReadFile(locationStr)
		if err != nil {
			if p.logger != nil {
				p.logger.Error(fmt.Sprintf("Failed to read preset file: %v", err))
			}
			return C.bool(false)
		}
		
		// Parse the preset data
		var preset map[string]interface{}
		if err := json.Unmarshal(presetData, &preset); err != nil {
			if p.logger != nil {
				p.logger.Error(fmt.Sprintf("Failed to parse preset file: %v", err))
			}
			return C.bool(false)
		}
		
		// Clear all active voices when loading a preset
		for i := range p.voices {
			if p.voices[i] != nil {
				p.voices[i].IsActive = false
			}
		}
		
		// Load the preset state
		p.LoadState(preset)
		
		if p.logger != nil {
			p.logger.Info("Preset loaded successfully")
		}
		return C.bool(true)
		
	case C.uint32_t(api.PresetLocationPlugin):
		// Load bundled preset from embedded files
		presetPath := filepath.Join("presets", "factory", loadKeyStr+".json")
		
		// Read from embedded filesystem
		presetData, err := factoryPresets.ReadFile(presetPath)
		if err != nil {
			if p.logger != nil {
				p.logger.Error(fmt.Sprintf("Failed to load bundled preset '%s': %v", loadKeyStr, err))
			}
			return C.bool(false)
		}
		
		// Parse the preset data
		var preset map[string]interface{}
		if err := json.Unmarshal(presetData, &preset); err != nil {
			if p.logger != nil {
				p.logger.Error(fmt.Sprintf("Failed to parse bundled preset: %v", err))
			}
			return C.bool(false)
		}
		
		// Clear all active voices when loading a preset
		for i := range p.voices {
			if p.voices[i] != nil {
				p.voices[i].IsActive = false
			}
		}
		
		// Load the preset state
		p.LoadState(preset)
		
		if p.logger != nil {
			p.logger.Info(fmt.Sprintf("Bundled preset '%s' loaded successfully", loadKeyStr))
		}
		return C.bool(true)
		
	default:
		return C.bool(false)
	}
}

// Voice represents a single active note
type Voice struct {
	NoteID       int32
	Channel      int16
	Key          int16
	Velocity     float64
	Phase        float64
	IsActive     bool
	
	// Tuning support
	TuningID     uint64  // ID of tuning to use (0 for equal temperament)
	
	// Per-voice parameter values (for polyphonic modulation)
	// These override the global values when non-zero
	VolumeModulation   float64  // Additional volume modulation (0.0 = no change)
	PitchBend         float64  // Pitch bend in semitones
	Brightness        float64  // Filter/brightness modulation (0.0-1.0)
	Pressure          float64  // Aftertouch/pressure value (0.0-1.0)
	
	// Simple envelope state (for compatibility with existing code)
	EnvelopeValue  float64  // Current envelope value (0.0-1.0)
	ReleasePhase   float64  // Release phase progress (0.0-1.0, -1.0 if not releasing)
	EnvelopeTime   float64  // Time in current envelope stage
	EnvelopePhase  int      // ADSR phase: 0=attack, 1=decay, 2=sustain
}

// SynthPlugin implements a simple synthesizer with atomic parameter storage
type SynthPlugin struct {
	event.NoOpHandler // Embed to get default no-op implementations
	
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
	paramManager *param.Manager
	
	// Note port management
	notePortManager *api.NotePortManager
	
	// Host utilities
	logger       *hostpkg.Logger
	trackInfo    *hostpkg.TrackInfoProvider
	threadCheck  *thread.Checker
	transportControl *hostpkg.TransportControl
	tuning       *extension.HostTuning
	// contextMenuProvider - removed as it's migrated to extension package
	
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
		paramManager: param.NewManager(),
		notePortManager: api.NewNotePortManager(),
		poolDiagnostics: &event.Diagnostics{},
	}
	
	// Set default values atomically
	atomic.StoreInt64(&plugin.volume, int64(util.AtomicFloat64ToBits(0.7)))      // -3dB
	atomic.StoreInt64(&plugin.waveform, 0)                          // sine
	atomic.StoreInt64(&plugin.attack, int64(util.AtomicFloat64ToBits(0.01)))     // 10ms
	atomic.StoreInt64(&plugin.decay, int64(util.AtomicFloat64ToBits(0.1)))       // 100ms
	atomic.StoreInt64(&plugin.sustain, int64(util.AtomicFloat64ToBits(0.7)))     // 70%
	atomic.StoreInt64(&plugin.release, int64(util.AtomicFloat64ToBits(0.3)))     // 300ms
	
	// Register parameters using our new abstraction
	plugin.paramManager.Register(param.Percentage(1, "Volume", 70.0))
	plugin.paramManager.Register(param.Choice(2, "Waveform", 3, 0))
	
	// Register ADSR parameters
	plugin.paramManager.Register(param.ADSR(3, "Attack", 2.0))    // Max 2 seconds
	plugin.paramManager.Register(param.ADSR(4, "Decay", 2.0))     // Max 2 seconds  
	plugin.paramManager.Register(param.Percentage(5, "Sustain", 70.0)) // 0-100%
	plugin.paramManager.Register(param.ADSR(6, "Release", 5.0))   // Max 5 seconds
	
	// Configure note port for instrument
	plugin.notePortManager.AddInputPort(api.CreateDefaultInstrumentPort())
	
	return plugin
}

// Init initializes the plugin
func (p *SynthPlugin) Init() bool {
	// Mark this as main thread for debug builds
	api.DebugSetMainThread()
	
	// Initialize thread checker
	if p.host != nil {
		p.threadCheck = thread.NewChecker(p.host)
		if p.threadCheck.IsAvailable() && p.logger != nil {
			p.logger.Info("Thread Check extension available - thread safety validation enabled")
		}
	}
	
	// Initialize track info helper
	if p.host != nil {
		p.trackInfo = hostpkg.NewTrackInfoProvider(p.host)
	}
	
	// Initialize transport control
	if p.host != nil {
		p.transportControl = hostpkg.NewTransportControl(p.host)
		if p.logger != nil {
			p.logger.Info("Transport control extension initialized")
		}
	}
	
	// Initialize tuning support
	if p.host != nil {
		p.tuning = extension.NewHostTuning(p.host)
		if p.logger != nil {
			p.logger.Info("Tuning extension initialized")
			// Log available tunings
			tunings := p.tuning.GetAllTunings()
			if len(tunings) > 0 {
				for _, t := range tunings {
					p.logger.Info(fmt.Sprintf("Available tuning: %s (ID: %d, Dynamic: %v)", 
						t.Name, t.TuningID, t.IsDynamic))
				}
			}
		}
	}
	
	if p.logger != nil {
		p.logger.Debug("Synth plugin initialized")
		
		// Log available bundled presets
		presets := p.GetAvailablePresets()
		if len(presets) > 0 {
			p.logger.Info(fmt.Sprintf("Available bundled presets: %v", presets))
		}
	}
	
	// TODO: Initialize context menu provider with param.Manager support
	// if p.host != nil {
	//	p.contextMenuProvider = api.NewDefaultContextMenuProvider(
	//		p.paramManager,
	//		"Simple Synth", 
	//		"1.0.0",
	//		p.host,
	//	)
	//	p.contextMenuProvider.SetAboutMessage("Simple Synth v1.0.0 - A basic subtractive synthesizer")
	// }
	
	// Check initial track info
	if p.trackInfo != nil {
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
	p.sampleRate = sampleRate
	p.isActivated = true
	
	// Clear all voices
	for i := range p.voices {
		p.voices[i] = nil
	}
	
	if p.logger != nil {
		p.logger.Info(fmt.Sprintf("Synth activated at %.0f Hz, buffer size %d-%d", sampleRate, minFrames, maxFrames))
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
func (p *SynthPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events *event.EventProcessor) int {
	// Check if we're in a valid state for processing
	if !p.isActivated || !p.isProcessing {
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
				sample := generateSample(voice.Phase, freq, waveform)
				
				// Apply brightness as a simple low-pass filter simulation
				if voice.Brightness > 0.0 && voice.Brightness < 1.0 {
					// Simple brightness simulation (in real implementation, use proper filter)
					sample *= (voice.Brightness * 0.7 + 0.3)
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
					sample *= (1.0 + voice.Pressure * 0.3)
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
				voice.Phase += freq / p.sampleRate
				if voice.Phase >= 1.0 {
					voice.Phase -= 1.0
				}
			}
			
			// Check if voice is still active
			if voice.ReleasePhase >= 1.0 {
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
		p.paramManager.Set(paramEvent.ParamID, paramEvent.Value)
	case 2: // Waveform
		atomic.StoreInt64(&p.waveform, int64(math.Round(paramEvent.Value)))
		p.paramManager.Set(paramEvent.ParamID, paramEvent.Value)
	case 3: // Attack
		atomic.StoreInt64(&p.attack, int64(util.AtomicFloat64ToBits(paramEvent.Value)))
		p.paramManager.Set(paramEvent.ParamID, paramEvent.Value)
	case 4: // Decay
		atomic.StoreInt64(&p.decay, int64(util.AtomicFloat64ToBits(paramEvent.Value)))
		p.paramManager.Set(paramEvent.ParamID, paramEvent.Value)
	case 5: // Sustain
		atomic.StoreInt64(&p.sustain, int64(util.AtomicFloat64ToBits(paramEvent.Value)))
		p.paramManager.Set(paramEvent.ParamID, paramEvent.Value)
	case 6: // Release
		atomic.StoreInt64(&p.release, int64(util.AtomicFloat64ToBits(paramEvent.Value)))
		p.paramManager.Set(paramEvent.ParamID, paramEvent.Value)
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
			voice.PitchBend = paramEvent.Value * 2.0 - 1.0 // Convert 0-1 to -1 to +1 semitones
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
		if p.logger != nil {
			p.logger.Info("Transport toggle play requested via C0")
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
		NoteID:        noteEvent.NoteID,
		Channel:       noteEvent.Channel,
		Key:           noteEvent.Key,
		Velocity:      velocity,
		Phase:         0.0,
		IsActive:      true,
		EnvelopeValue: 0.0,
		ReleasePhase:  -1.0,
		EnvelopeTime:  0.0,
		EnvelopePhase: 0,
	}
	
	if p.logger != nil {
		p.logger.Debug(fmt.Sprintf("Note on: key=%d, velocity=%.2f, voice=%d", noteEvent.Key, velocity, voiceIndex))
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
			voice.ReleasePhase = 0.0
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
	
	// No voices in release, steal the quietest voice
	quietestIdx := 0
	quietestLevel := 1.0
	
	for i, voice := range p.voices {
		if voice != nil && voice.IsActive {
			// Consider envelope value and velocity
			level := voice.EnvelopeValue * voice.Velocity
			if level < quietestLevel {
				quietestLevel = level
				quietestIdx = i
			}
		}
	}
	
	return quietestIdx
}

// getEnvelopeValue calculates the ADSR envelope value for a voice
func (p *SynthPlugin) getEnvelopeValue(voice *Voice, sampleIndex, frameCount uint32, attack, decay, sustain, release float64) float64 {
	// Time increment per sample
	timeInc := 1.0 / p.sampleRate
	
	// If in release phase
	if voice.ReleasePhase >= 0.0 {
		// Update release phase
		releaseSamples := release * p.sampleRate
		if releaseSamples <= 0 {
			releaseSamples = 1
		}
		voice.ReleasePhase += 1.0 / releaseSamples
		
		// Calculate release envelope (exponential decay from last envelope value)
		if voice.ReleasePhase >= 1.0 {
			voice.EnvelopeValue = 0.0
		} else {
			voice.EnvelopeValue = voice.EnvelopeValue * math.Pow(0.99, 1.0/releaseSamples)
		}
		return voice.EnvelopeValue
	}
	
	// Update envelope time
	voice.EnvelopeTime += timeInc
	
	// Attack phase
	if voice.EnvelopePhase == 0 {
		if attack <= 0 {
			voice.EnvelopeValue = 1.0
			voice.EnvelopePhase = 1
			voice.EnvelopeTime = 0
		} else if voice.EnvelopeTime >= attack {
			voice.EnvelopeValue = 1.0
			voice.EnvelopePhase = 1
			voice.EnvelopeTime = 0
		} else {
			voice.EnvelopeValue = voice.EnvelopeTime / attack
		}
	}
	
	// Decay phase
	if voice.EnvelopePhase == 1 {
		if decay <= 0 {
			voice.EnvelopeValue = sustain
			voice.EnvelopePhase = 2
		} else if voice.EnvelopeTime >= decay {
			voice.EnvelopeValue = sustain
			voice.EnvelopePhase = 2
		} else {
			decayProgress := voice.EnvelopeTime / decay
			voice.EnvelopeValue = 1.0 - decayProgress*(1.0-sustain)
		}
	}
	
	// Sustain phase
	if voice.EnvelopePhase == 2 {
		voice.EnvelopeValue = sustain
	}
	
	return voice.EnvelopeValue
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
		"attack":        util.AtomicFloat64FromBits(uint64(atomic.LoadInt64(&p.attack))),
		"decay":         util.AtomicFloat64FromBits(uint64(atomic.LoadInt64(&p.decay))),
		"sustain":       util.AtomicFloat64FromBits(uint64(atomic.LoadInt64(&p.sustain))),
		"release":       util.AtomicFloat64FromBits(uint64(atomic.LoadInt64(&p.release))),
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
	tailSamples := uint32(release * p.sampleRate)
	
	// Add some extra samples for safety
	return tailSamples + uint32(p.sampleRate * 0.1) // 100ms extra
}

// OnTimer handles timer callbacks
func (p *SynthPlugin) OnTimer(timerID uint64) {
	// Synth doesn't currently use timers
	// Could be used for UI updates, voice status monitoring, etc.
	if p.logger != nil {
		p.logger.Debug(fmt.Sprintf("Timer %d fired", timerID))
	}
}

// OnTrackInfoChanged is called when the track information changes
func (p *SynthPlugin) OnTrackInfoChanged() {
	if p.trackInfo == nil {
		return
	}
	
	// Get the new track information
	info, ok := p.trackInfo.Get()
	if !ok {
		if p.logger != nil {
			p.logger.Warning("Failed to get track info")
		}
		return
	}
	
	// Log the track information
	if p.logger != nil {
		p.logger.Info(fmt.Sprintf("Track info changed:"))
		if info.Flags&api.TrackInfoHasTrackName != 0 {
			p.logger.Info(fmt.Sprintf("  Track name: %s", info.Name))
		}
		if info.Flags&api.TrackInfoHasTrackColor != 0 {
			p.logger.Info(fmt.Sprintf("  Track color: R=%d G=%d B=%d A=%d", 
				info.Color.Red, info.Color.Green, info.Color.Blue, info.Color.Alpha))
		}
		if info.Flags&api.TrackInfoHasAudioChannel != 0 {
			p.logger.Info(fmt.Sprintf("  Audio channels: %d, port type: %s", 
				info.AudioChannelCount, info.AudioPortType))
		}
		
		// Adjust synth behavior based on track type
		if info.Flags&api.TrackInfoIsForReturnTrack != 0 {
			p.logger.Info("  This is a return track - adjusting for wet signal")
			// Could adjust default mix to 100% wet
		}
		if info.Flags&api.TrackInfoIsForBus != 0 {
			p.logger.Info("  This is a bus track")
			// Could adjust polyphony or processing
		}
		if info.Flags&api.TrackInfoIsForMaster != 0 {
			p.logger.Info("  This is the master track")
			// Synths typically wouldn't be on master, but if so, could adjust
		}
	}
}

// OnTuningChanged is called when tunings are added or removed
func (p *SynthPlugin) OnTuningChanged() {
	if p.tuning == nil {
		return
	}
	
	if p.logger != nil {
		p.logger.Info("Tuning pool changed, refreshing available tunings")
		
		// Log all available tunings
		tunings := p.tuning.GetAllTunings()
		p.logger.Info(fmt.Sprintf("Available tunings: %d", len(tunings)))
		for _, t := range tunings {
			p.logger.Info(fmt.Sprintf("  - %s (ID: %d, Dynamic: %v)", 
				t.Name, t.TuningID, t.IsDynamic))
		}
	}
}


// GetAvailablePresets returns a list of available bundled preset names
func (p *SynthPlugin) GetAvailablePresets() []string {
	var presets []string
	
	// Read the embedded preset directory
	entries, err := factoryPresets.ReadDir("presets/factory")
	if err != nil {
		if p.logger != nil {
			p.logger.Error(fmt.Sprintf("Failed to read preset directory: %v", err))
		}
		return presets
	}
	
	// Collect preset names (without .json extension)
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			presetName := entry.Name()[:len(entry.Name())-5] // Remove .json
			presets = append(presets, presetName)
		}
	}
	
	return presets
}

// GetVoiceInfo returns voice count and capacity information
func (p *SynthPlugin) GetVoiceInfo() api.VoiceInfo {
	// Count active voices
	activeVoices := uint32(0)
	for _, voice := range p.voices {
		if voice != nil && voice.IsActive {
			activeVoices++
		}
	}
	
	return api.VoiceInfo{
		VoiceCount:    uint32(len(p.voices)), // Maximum polyphony
		VoiceCapacity: uint32(len(p.voices)), // Same as count in our implementation
		Flags:         api.VoiceInfoSupportsOverlappingNotes, // We support note IDs
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
	atomic.StoreInt64(&p.attack, int64(util.AtomicFloat64ToBits(preset.Attack)))
	atomic.StoreInt64(&p.decay, int64(util.AtomicFloat64ToBits(preset.Decay)))
	atomic.StoreInt64(&p.sustain, int64(util.AtomicFloat64ToBits(preset.Sustain)))
	atomic.StoreInt64(&p.release, int64(util.AtomicFloat64ToBits(preset.Release)))
	
	// Load any additional state data
	if preset.StateData != nil {
		p.LoadState(preset.StateData)
	}
	
	return true
}

// Helper functions for atomic float64 operations

// Phase 3 Extension Exports

//export ClapGo_PluginLatencyGet
func ClapGo_PluginLatencyGet(plugin unsafe.Pointer) C.uint32_t {
	if plugin == nil {
		return 0
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	return C.uint32_t(p.GetLatency())
}

//export ClapGo_PluginTailGet
func ClapGo_PluginTailGet(plugin unsafe.Pointer) C.uint32_t {
	if plugin == nil {
		return 0
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	return C.uint32_t(p.GetTail())
}

//export ClapGo_PluginOnTimer
func ClapGo_PluginOnTimer(plugin unsafe.Pointer, timerID C.uint64_t) {
	if plugin == nil {
		return
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	p.OnTimer(uint64(timerID))
}

// Voice info implementation is now in pkg/api/voice_info.go
// The synth plugin implements VoiceInfoProvider interface

// Phase 7 Extension Exports

//export ClapGo_PluginTrackInfoChanged
func ClapGo_PluginTrackInfoChanged(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	p.OnTrackInfoChanged()
}

// Tuning Extension Export

//export ClapGo_PluginTuningChanged
func ClapGo_PluginTuningChanged(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	p.OnTuningChanged()
}

// Note Name Extension Exports

//export ClapGo_PluginNoteNameCount
func ClapGo_PluginNoteNameCount(plugin unsafe.Pointer) C.uint32_t {
	if plugin == nil {
		return 0
	}
	_ = cgo.Handle(plugin).Value().(*SynthPlugin)
	
	// Synth provides note names for all MIDI notes
	return C.uint32_t(128) // All MIDI notes
}

//export ClapGo_PluginNoteNameGet
func ClapGo_PluginNoteNameGet(plugin unsafe.Pointer, index C.uint32_t, noteName unsafe.Pointer) C.bool {
	if plugin == nil || noteName == nil || index >= 128 {
		return C.bool(false)
	}
	_ = cgo.Handle(plugin).Value().(*SynthPlugin)
	
	// Get standard note name for this index
	noteNames := extension.StandardNoteNames()
	if int(index) >= len(noteNames) {
		return C.bool(false)
	}
	
	// Convert to C structure
	extension.NoteNameToC(&noteNames[index], noteName)
	
	return C.bool(true)
}


func main() {
	// This is not called when used as a plugin,
	// but can be useful for testing
}