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
	"math"
	"runtime/cgo"
	"sync/atomic"
	"unsafe"

	"github.com/justyntemme/clapgo/pkg/audio"
	"github.com/justyntemme/clapgo/pkg/event"
	"github.com/justyntemme/clapgo/pkg/extension"
	hostpkg "github.com/justyntemme/clapgo/pkg/host"
	"github.com/justyntemme/clapgo/pkg/param"
	"github.com/justyntemme/clapgo/pkg/process"
	"github.com/justyntemme/clapgo/pkg/state"
	"github.com/justyntemme/clapgo/pkg/thread"
)

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

// Standardized exports for manifest system

//export ClapGo_CreatePlugin
func ClapGo_CreatePlugin(host unsafe.Pointer, pluginID *C.char) unsafe.Pointer {
	id := C.GoString(pluginID)
	if id == PluginID {
		// Store the host pointer and create utilities
		synthPlugin.Host = host
		synthPlugin.Logger = hostpkg.NewLogger(host)

		// Log plugin creation
		if synthPlugin.Logger != nil {
			synthPlugin.Logger.Info("Creating synth plugin instance")
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
	thread.DebugMarkAudioThread()
	defer thread.DebugUnmarkAudioThread()

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
	audioIn := audio.ConvertFromCBuffers(unsafe.Pointer(cProcess.audio_inputs), uint32(cProcess.audio_inputs_count), framesCount)
	audioOut := audio.ConvertFromCBuffers(unsafe.Pointer(cProcess.audio_outputs), uint32(cProcess.audio_outputs_count), framesCount)

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
	return C.uint32_t(p.ParamManager.Count())
}

//export ClapGo_PluginParamsGetInfo
func ClapGo_PluginParamsGetInfo(plugin unsafe.Pointer, index C.uint32_t, info unsafe.Pointer) C.bool {
	if plugin == nil || info == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	err := p.GetParamInfo(uint32(index), info)
	return C.bool(err == nil)
}

//export ClapGo_PluginParamsGetValue
func ClapGo_PluginParamsGetValue(plugin unsafe.Pointer, paramID C.uint32_t, value *C.double) C.bool {
	if plugin == nil || value == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	return C.bool(p.GetParamValue(uint32(paramID), unsafe.Pointer(value)))
}

//export ClapGo_PluginParamsValueToText
func ClapGo_PluginParamsValueToText(plugin unsafe.Pointer, paramID C.uint32_t, value C.double, buffer *C.char, size C.uint32_t) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	return C.bool(p.ParamValueToText(uint32(paramID), float64(value), unsafe.Pointer(buffer), uint32(size)))
}

//export ClapGo_PluginParamsTextToValue
func ClapGo_PluginParamsTextToValue(plugin unsafe.Pointer, paramID C.uint32_t, text *C.char, value *C.double) C.bool {
	if plugin == nil || text == nil || value == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	return C.bool(p.ParamTextToValue(uint32(paramID), C.GoString(text), unsafe.Pointer(value)))
}

//export ClapGo_PluginParamsFlush
func ClapGo_PluginParamsFlush(plugin unsafe.Pointer, inEvents unsafe.Pointer, outEvents unsafe.Pointer) {
	if plugin == nil {
		return
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)
	p.ParamsFlush(inEvents, outEvents)
}

//export ClapGo_PluginStateSave
func ClapGo_PluginStateSave(plugin unsafe.Pointer, stream unsafe.Pointer) C.bool {
	if plugin == nil || stream == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)

	out := state.NewClapOutputStream(stream)

	// Write state version
	if err := out.WriteUint32(1); err != nil {
		return false
	}

	// Write parameter count
	paramCount := p.ParamManager.Count()
	if err := out.WriteUint32(paramCount); err != nil {
		return false
	}

	// Write each parameter ID and value
	for i := uint32(0); i < paramCount; i++ {
		info, err := p.ParamManager.GetInfoByIndex(i)
		if err != nil {
			return false
		}

		// Write parameter ID
		if err := out.WriteUint32(info.ID); err != nil {
			return false
		}

		// Write parameter value
		value := p.ParamManager.Get(info.ID)
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

	var portInfo *audio.NotePortInfo
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

	in := state.NewClapInputStream(stream)

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
			param.UpdateParameterAtomic(&p.volume, value, p.ParamManager, paramID)
		case 2: // Waveform
			atomic.StoreInt64(&p.waveform, int64(math.Round(value)))
			p.ParamManager.Set(paramID, value)
		case 3: // Attack
			param.UpdateParameterAtomic(&p.attack, value, p.ParamManager, paramID)
		case 4: // Decay
			param.UpdateParameterAtomic(&p.decay, value, p.ParamManager, paramID)
		case 5: // Sustain
			param.UpdateParameterAtomic(&p.sustain, value, p.ParamManager, paramID)
		case 6: // Release
			param.UpdateParameterAtomic(&p.release, value, p.ParamManager, paramID)
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
	if p.Logger != nil {
		switch contextType {
		case C.uint32_t(state.ContextForDuplicate):
			p.Logger.Info("Saving state for duplicate")
		case C.uint32_t(state.ContextForProject):
			p.Logger.Info("Saving state for project")
		default:
			p.Logger.Info(fmt.Sprintf("Saving state with unknown context: %d", contextType))
		}
	}

	// For all contexts, save everything including voice state
	return ClapGo_PluginStateSave(plugin, stream)
}

//export ClapGo_PluginStateLoadWithContext
func ClapGo_PluginStateLoadWithContext(plugin unsafe.Pointer, stream unsafe.Pointer, contextType C.uint32_t) C.bool {
	if plugin == nil || stream == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*SynthPlugin)

	// Log the context type
	if p.Logger != nil {
		switch contextType {
		case C.uint32_t(state.ContextForDuplicate):
			p.Logger.Info("Loading state for duplicate")
		case C.uint32_t(state.ContextForProject):
			p.Logger.Info("Loading state for project")
		default:
			p.Logger.Info(fmt.Sprintf("Loading state with unknown context: %d", contextType))
		}
	}

	// Load the state
	return ClapGo_PluginStateLoad(plugin, stream)
}
