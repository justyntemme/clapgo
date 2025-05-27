package main

// #cgo CFLAGS: -I../../include/clap/include
// #include "../../include/clap/include/clap/clap.h"
// #include "../../include/clap/include/clap/ext/remote-controls.h"
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
	"runtime/cgo"
	"unsafe"
	
	"github.com/justyntemme/clapgo/pkg/api"
	"github.com/justyntemme/clapgo/pkg/audio"
	"github.com/justyntemme/clapgo/pkg/param"
	"github.com/justyntemme/clapgo/pkg/plugin"
	"github.com/justyntemme/clapgo/pkg/state"
)


// Global plugin instance
var gainPlugin *GainPlugin

func init() {
	fmt.Println("Initializing gain plugin")
	gainPlugin = NewGainPlugin()
	fmt.Printf("Gain plugin initialized: %s (%s)\n", gainPlugin.GetPluginInfo().Name, gainPlugin.GetPluginInfo().ID)
}

// Standardized export functions for manifest system

//export ClapGo_CreatePlugin
func ClapGo_CreatePlugin(host unsafe.Pointer, pluginID *C.char) unsafe.Pointer {
	id := C.GoString(pluginID)
	fmt.Printf("Gain plugin - ClapGo_CreatePlugin with ID: %s\n", id)
	
	if id == PluginID {
		// Initialize host-dependent features
		gainPlugin.PluginBase.InitWithHost(host)
		
		// Log plugin creation
		if gainPlugin.Logger != nil {
			gainPlugin.Logger.Info("Creating gain plugin instance")
		}
		
		// Create a CGO handle to safely pass the Go object to C
		handle := cgo.NewHandle(gainPlugin)
		fmt.Printf("Created plugin instance: %s, handle: %v\n", id, handle)
		
		// Log handle creation
		if gainPlugin.Logger != nil {
			gainPlugin.Logger.Debug(fmt.Sprintf("[ClapGo_CreatePlugin] Created handle: %v", handle))
		}
		
		// Register as audio ports provider
		api.RegisterAudioPortsProvider(unsafe.Pointer(handle), gainPlugin)
		
		return unsafe.Pointer(handle)
	}
	
	fmt.Printf("Error: Unknown plugin ID: %s\n", id)
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
	return C.CString(gainPlugin.GetPluginID())
}

//export ClapGo_GetPluginName
func ClapGo_GetPluginName(pluginID *C.char) *C.char {
	return C.CString(gainPlugin.GetPluginInfo().Name)
}

//export ClapGo_GetPluginVendor
func ClapGo_GetPluginVendor(pluginID *C.char) *C.char {
	return C.CString(gainPlugin.GetPluginInfo().Vendor)
}

//export ClapGo_GetPluginVersion
func ClapGo_GetPluginVersion(pluginID *C.char) *C.char {
	return C.CString(gainPlugin.GetPluginInfo().Version)
}

//export ClapGo_GetPluginDescription
func ClapGo_GetPluginDescription(pluginID *C.char) *C.char {
	return C.CString(gainPlugin.GetPluginInfo().Description)
}

//export ClapGo_PluginInit
func ClapGo_PluginInit(plugin unsafe.Pointer) C.bool {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC in ClapGo_PluginInit: %v\n", r)
		}
	}()
	
	if plugin == nil {
		return C.bool(false)
	}
	
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	
	// Log the init call
	if p.Logger != nil {
		p.Logger.Debug("[ClapGo_PluginInit] Starting plugin initialization")
	}
	
	result := p.Init()
	
	// Log the result
	if p.Logger != nil {
		if result {
			p.Logger.Info("[ClapGo_PluginInit] Plugin initialization successful")
		} else {
			p.Logger.Error("[ClapGo_PluginInit] Plugin initialization failed")
		}
	}
	
	return C.bool(result)
}

//export ClapGo_PluginDestroy
func ClapGo_PluginDestroy(plugin unsafe.Pointer) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC in ClapGo_PluginDestroy: %v\n", r)
		}
	}()
	
	if plugin == nil {
		return
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*GainPlugin)
	
	// Log the destroy call
	if p.Logger != nil {
		p.Logger.Debug(fmt.Sprintf("[ClapGo_PluginDestroy] Destroying plugin instance, handle: %v", handle))
	}
	
	p.Destroy()
	
	// Unregister from audio ports provider
	api.UnregisterAudioPortsProvider(plugin)
	
	// Log completion
	if p.Logger != nil {
		p.Logger.Info("[ClapGo_PluginDestroy] Plugin instance destroyed successfully")
	}
	
	handle.Delete()
}

//export ClapGo_PluginActivate
func ClapGo_PluginActivate(plugin unsafe.Pointer, sampleRate C.double, minFrames, maxFrames C.uint32_t) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	
	// Log the activate call
	if p.Logger != nil {
		p.Logger.Debug(fmt.Sprintf("[ClapGo_PluginActivate] Activating plugin - SR: %.0f, frames: %d-%d", 
			float64(sampleRate), uint32(minFrames), uint32(maxFrames)))
	}
	
	result := p.Activate(float64(sampleRate), uint32(minFrames), uint32(maxFrames))
	
	// Log result
	if p.Logger != nil {
		if result {
			p.Logger.Info("[ClapGo_PluginActivate] Plugin activation successful")
		} else {
			p.Logger.Error("[ClapGo_PluginActivate] Plugin activation failed")
		}
	}
	
	return C.bool(result)
}

//export ClapGo_PluginDeactivate
func ClapGo_PluginDeactivate(plugin unsafe.Pointer) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC in ClapGo_PluginDeactivate: %v\n", r)
		}
	}()
	
	if plugin == nil {
		return
	}
	
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	
	// Log the deactivate call
	if p.Logger != nil {
		p.Logger.Debug("[ClapGo_PluginDeactivate] Deactivating plugin")
	}
	
	p.Deactivate()
	
	// Log completion
	if p.Logger != nil {
		p.Logger.Info("[ClapGo_PluginDeactivate] Plugin deactivation successful")
	}
}

//export ClapGo_PluginStartProcessing
func ClapGo_PluginStartProcessing(plugin unsafe.Pointer) C.bool {
	// Mark this thread as audio thread for debug builds
	api.DebugMarkAudioThread()
	defer api.DebugUnmarkAudioThread()
	
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC in ClapGo_PluginStartProcessing: %v\n", r)
		}
	}()
	
	if plugin == nil {
		return C.bool(false)
	}
	
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	
	// Log the start processing call
	if p.Logger != nil {
		p.Logger.Debug("[ClapGo_PluginStartProcessing] Starting audio processing")
	}
	
	result := p.StartProcessing()
	
	// Log result
	if p.Logger != nil {
		if result {
			p.Logger.Info("[ClapGo_PluginStartProcessing] Audio processing started successfully")
		} else {
			p.Logger.Error("[ClapGo_PluginStartProcessing] Failed to start audio processing")
		}
	}
	
	return C.bool(result)
}

//export ClapGo_PluginStopProcessing
func ClapGo_PluginStopProcessing(plugin unsafe.Pointer) {
	// Mark this thread as audio thread for debug builds
	api.DebugMarkAudioThread()
	defer api.DebugUnmarkAudioThread()
	
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC in ClapGo_PluginStopProcessing: %v\n", r)
		}
	}()
	
	if plugin == nil {
		return
	}
	
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	
	// Log the stop processing call
	if p.Logger != nil {
		p.Logger.Debug("[ClapGo_PluginStopProcessing] Stopping audio processing")
	}
	
	p.StopProcessing()
	
	// Log completion
	if p.Logger != nil {
		p.Logger.Info("[ClapGo_PluginStopProcessing] Audio processing stopped successfully")
	}
}

//export ClapGo_PluginReset
func ClapGo_PluginReset(plugin unsafe.Pointer) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC in ClapGo_PluginReset: %v\n", r)
		}
	}()
	
	if plugin == nil {
		return
	}
	
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	
	// Log the reset call
	if p.Logger != nil {
		p.Logger.Debug("[ClapGo_PluginReset] Resetting plugin state")
	}
	
	p.Reset()
	
	// Log completion
	if p.Logger != nil {
		p.Logger.Info("[ClapGo_PluginReset] Plugin reset successful")
	}
}

//export ClapGo_PluginProcess
func ClapGo_PluginProcess(plugin unsafe.Pointer, process unsafe.Pointer) C.int32_t {
	// Mark this thread as audio thread for debug builds
	api.DebugMarkAudioThread()
	defer api.DebugUnmarkAudioThread()
	
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC in ClapGo_PluginProcess: %v\n", r)
		}
	}()
	
	if plugin == nil || process == nil {
		return C.int32_t(api.ProcessError)
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*GainPlugin)
	
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
	
	// Setup event pool logging
	api.SetupPoolLogging(eventHandler, p.Logger)
	
	// Call the actual Go process method
	result := p.Process(steadyTime, framesCount, audioIn, audioOut, eventHandler)
	
	// Log event pool diagnostics periodically (every 1000 calls)
	p.poolDiagnostics.LogPoolDiagnostics(eventHandler, 1000)
	
	return C.int32_t(result)
}

//export ClapGo_PluginGetExtension
func ClapGo_PluginGetExtension(plugin unsafe.Pointer, id *C.char) unsafe.Pointer {
	if plugin == nil {
		return nil
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	extID := C.GoString(id)
	return p.GetExtension(extID)
}

//export ClapGo_PluginOnMainThread
func ClapGo_PluginOnMainThread(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	p.OnMainThread()
}

//export ClapGo_PluginParamsCount
func ClapGo_PluginParamsCount(plugin unsafe.Pointer) C.uint32_t {
	if plugin == nil {
		return 0
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	return C.uint32_t(p.ParamManager.Count())
}

//export ClapGo_PluginParamsGetInfo
func ClapGo_PluginParamsGetInfo(plugin unsafe.Pointer, index C.uint32_t, info unsafe.Pointer) C.bool {
	if plugin == nil || info == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	
	// Get parameter info from manager
	paramInfo, err := p.ParamManager.GetInfoByIndex(uint32(index))
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
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	
	// Get current value - read atomically from our gain storage
	if uint32(paramID) == 0 {
		*value = C.double(p.gain.Load())
		return C.bool(true)
	}
	
	return C.bool(false)
}

//export ClapGo_PluginParamsValueToText
func ClapGo_PluginParamsValueToText(plugin unsafe.Pointer, paramID C.uint32_t, value C.double, buffer *C.char, size C.uint32_t) C.bool {
	if plugin == nil || buffer == nil || size == 0 {
		return C.bool(false)
	}
	// For gain parameter, format as dB
	if uint32(paramID) == 0 {
		text := param.FormatValue(float64(value), param.FormatDecibel)
		
		// Copy to C buffer manually
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
	
	return C.bool(false)
}

//export ClapGo_PluginParamsTextToValue
func ClapGo_PluginParamsTextToValue(plugin unsafe.Pointer, paramID C.uint32_t, text *C.char, value *C.double) C.bool {
	if plugin == nil || text == nil || value == nil {
		return C.bool(false)
	}
	
	// Convert text to Go string
	goText := C.GoString(text)
	
	// For gain parameter, use the ParameterValueParser
	if uint32(paramID) == 0 {
		parser := param.NewParser(param.FormatDecibel)
		if parsedValue, err := parser.ParseValue(goText); err == nil {
			// Clamp to valid range for gain parameter
			clamped := param.ClampValue(parsedValue, 0.0, 2.0)
			*value = C.double(clamped)
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
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	
	// Process events using our abstraction
	if inEvents != nil {
		eventHandler := api.NewEventProcessor(inEvents, outEvents)
		p.processEvents(eventHandler, 0)
	}
}

// GainPlugin represents the gain plugin
type GainPlugin struct {
	*plugin.PluginBase
	*audio.StereoPortProvider
	*audio.SurroundSupport
	
	// Plugin-specific parameters
	gain param.AtomicFloat64
	
	// Legacy fields (to be removed in future phases)
	contextMenuProvider *api.DefaultContextMenuProvider
	poolDiagnostics     api.EventPoolDiagnostics
	latency             uint32
}


// NewGainPlugin creates a new gain plugin instance
func NewGainPlugin() *GainPlugin {
	// Create plugin with base
	p := &GainPlugin{
		PluginBase: plugin.NewPluginBase(plugin.Info{
			ID:          PluginID,
			Name:        PluginName,
			Vendor:      PluginVendor,
			Version:     PluginVersion,
			Description: PluginDescription,
			URL:         "https://github.com/justyntemme/clapgo",
			Manual:      "https://github.com/justyntemme/clapgo",
			Support:     "https://github.com/justyntemme/clapgo/issues",
			Features:    []string{plugin.FeatureAudioEffect, plugin.FeatureStereo, plugin.FeatureUtility},
		}),
		StereoPortProvider: audio.NewStereoPortProvider(),
		SurroundSupport:    audio.NewStereoSurroundSupport(),
	}
	
	// Set default gain to 1.0 (0dB)
	p.gain.Store(1.0)
	
	// Register gain parameter
	p.ParamManager.Register(param.Volume(ParamGain, "Gain"))
	p.ParamManager.SetValue(ParamGain, 1.0)
	
	return p
}

// Init initializes the plugin
func (p *GainPlugin) Init() bool {
	// Initialize base functionality
	if !p.PluginBase.CommonInit() {
		return false
	}
	
	// Check initial track info
	if p.TrackInfo != nil {
		p.OnTrackInfoChanged()
	}
	
	return true
}

// Destroy cleans up plugin resources
func (p *GainPlugin) Destroy() {
	p.PluginBase.CommonDestroy()
}

// Activate prepares the plugin for processing
func (p *GainPlugin) Activate(sampleRate float64, minFrames, maxFrames uint32) bool {
	return p.PluginBase.CommonActivate(sampleRate, minFrames, maxFrames)
}

// Deactivate stops the plugin from processing
func (p *GainPlugin) Deactivate() {
	p.PluginBase.CommonDeactivate()
}

// StartProcessing begins audio processing
func (p *GainPlugin) StartProcessing() bool {
	// Assert audio thread
	api.DebugAssertAudioThread("GainPlugin.StartProcessing")
	if p.ThreadCheck != nil {
		p.ThreadCheck.AssertAudioThread("GainPlugin.StartProcessing")
	}
	
	return p.PluginBase.CommonStartProcessing()
}

// StopProcessing ends audio processing
func (p *GainPlugin) StopProcessing() {
	p.PluginBase.CommonStopProcessing()
}

// Reset resets the plugin state
func (p *GainPlugin) Reset() {
	p.PluginBase.CommonReset()
	// Reset gain to default
	p.gain.Store(1.0)
}

// Process processes audio data using the new abstractions
func (p *GainPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events api.EventHandler) int {
	// Assert audio thread
	api.DebugAssertAudioThread("GainPlugin.Process")
	if p.ThreadCheck != nil {
		p.ThreadCheck.AssertAudioThread("GainPlugin.Process")
	}
	
	// Check if we're in a valid state for processing
	if !p.IsActivated || !p.IsProcessing {
		return api.ProcessError
	}
	
	// Process events using our new abstraction - NO MORE MANUAL EVENT PARSING!
	if events != nil {
		p.processEvents(events, framesCount)
	}
	
	// Get current gain value atomically
	gain := float32(p.gain.Load())
	
	// Process audio using new audio package - much simpler!
	if err := audio.Process(audio.Buffer(audioOut), audio.Buffer(audioIn), gain); err != nil {
		if p.Logger != nil {
			p.Logger.Error(fmt.Sprintf("Audio processing error: %v", err))
		}
		return api.ProcessError
	}
	
	return api.ProcessContinue
}

// processEvents handles all incoming events using our new EventHandler abstraction
func (p *GainPlugin) processEvents(events api.EventHandler, frameCount uint32) {
	if events == nil {
		return
	}
	
	// Use the zero-allocation ProcessTypedEvents method
	events.ProcessTypedEvents(p)
}

// HandleParamValue handles parameter value changes (overrides NoOpHandler)
func (p *GainPlugin) HandleParamValue(paramEvent *api.ParamValueEvent, time uint32) {
	// Handle the parameter change based on its ID
	switch paramEvent.ParamID {
	case ParamGain:
		// Clamp value to valid range
		value := param.ClampValue(paramEvent.Value, 0.0, 2.0)
		
		// Update atomic storage and parameter manager
		p.gain.Store(value)
		if err := p.ParamManager.SetValue(paramEvent.ParamID, value); err != nil {
			if p.Logger != nil {
				p.Logger.Warning(fmt.Sprintf("Failed to set parameter %d: %v", paramEvent.ParamID, err))
			}
		}
		
		// Log the parameter change
		if p.Logger != nil {
			// Convert to dB for logging using audio package
			db := audio.LinearToDb(value)
			p.Logger.Debug(fmt.Sprintf("Gain changed to %.2f dB", db))
		}
	}
}

// No-op event handlers (gain plugin doesn't process these events)
func (p *GainPlugin) HandleParamMod(e *api.ParamModEvent, time uint32) {}
func (p *GainPlugin) HandleParamGestureBegin(e *api.ParamGestureEvent, time uint32) {}
func (p *GainPlugin) HandleParamGestureEnd(e *api.ParamGestureEvent, time uint32) {}
func (p *GainPlugin) HandleNoteOn(e *api.NoteEvent, time uint32) {}
func (p *GainPlugin) HandleNoteOff(e *api.NoteEvent, time uint32) {}
func (p *GainPlugin) HandleNoteChoke(e *api.NoteEvent, time uint32) {}
func (p *GainPlugin) HandleNoteEnd(e *api.NoteEvent, time uint32) {}
func (p *GainPlugin) HandleNoteExpression(e *api.NoteExpressionEvent, time uint32) {}
func (p *GainPlugin) HandleTransport(e *api.TransportEvent, time uint32) {}
func (p *GainPlugin) HandleMIDI(e *api.MIDIEvent, time uint32) {}
func (p *GainPlugin) HandleMIDISysex(e *api.MIDISysexEvent, time uint32) {}
func (p *GainPlugin) HandleMIDI2(e *api.MIDI2Event, time uint32) {}

// GetExtension gets a plugin extension
func (p *GainPlugin) GetExtension(id string) unsafe.Pointer {
	// Extensions are handled by the C bridge layer
	// The bridge provides params, state, and audio ports extensions
	// Preset load is handled in Go
	switch id {
	case api.ExtPresetLoad:
		// Return a non-nil pointer to indicate support
		// The actual implementation is in LoadPresetFromLocation
		return unsafe.Pointer(&p)
	default:
		return nil
	}
}

// GetPluginInfo returns information about the plugin
func (p *GainPlugin) GetPluginInfo() api.PluginInfo {
	return api.PluginInfo{
		ID:          PluginID,
		Name:        PluginName,
		Vendor:      PluginVendor,
		URL:         "https://github.com/justyntemme/clapgo",
		ManualURL:   "https://github.com/justyntemme/clapgo",
		SupportURL:  "https://github.com/justyntemme/clapgo/issues",
		Version:     PluginVersion,
		Description: PluginDescription,
		Features:    []string{"audio-effect", "stereo", "mono"},
	}
}

// OnMainThread is called on the main thread
func (p *GainPlugin) OnMainThread() {
	// Nothing to do
}

// GetPluginID returns the plugin ID
func (p *GainPlugin) GetPluginID() string {
	return PluginID
}

// GetLatency returns the plugin's latency in samples
func (p *GainPlugin) GetLatency() uint32 {
	// Check main thread (latency is queried on main thread)
	api.DebugAssertMainThread("GainPlugin.GetLatency")
	
	// For this simple gain plugin, we have no latency
	// In a real plugin with lookahead or FFT processing, this would return
	// the actual latency in samples
	return p.latency
}

// GetTail returns the plugin's tail length in samples
func (p *GainPlugin) GetTail() uint32 {
	// Gain plugin has no tail (no reverb/delay)
	return 0
}

// OnTimer handles timer callbacks
func (p *GainPlugin) OnTimer(timerID uint64) {
	// Gain plugin doesn't use timers
	// This is here as an example implementation
}

// OnTrackInfoChanged is called when the track information changes
func (p *GainPlugin) OnTrackInfoChanged() {
	if p.TrackInfo == nil {
		return
	}
	
	// Get the new track information
	info, ok := p.TrackInfo.GetTrackInfo()
	if !ok {
		if p.Logger != nil {
			p.Logger.Warning("Failed to get track info")
		}
		return
	}
	
	// Log the track information
	if p.Logger != nil {
		p.Logger.Info(fmt.Sprintf("Track info changed:"))
		if info.Flags&api.TrackInfoHasTrackName != 0 {
			p.Logger.Info(fmt.Sprintf("  Track name: %s", info.Name))
		}
		if info.Flags&api.TrackInfoHasTrackColor != 0 {
			p.Logger.Info(fmt.Sprintf("  Track color: R=%d G=%d B=%d A=%d", 
				info.Color.Red, info.Color.Green, info.Color.Blue, info.Color.Alpha))
		}
		if info.Flags&api.TrackInfoHasAudioChannel != 0 {
			p.Logger.Info(fmt.Sprintf("  Audio channels: %d, port type: %s", 
				info.AudioChannelCount, info.AudioPortType))
		}
		if info.Flags&api.TrackInfoIsForReturnTrack != 0 {
			p.Logger.Info("  This is a return track")
		}
		if info.Flags&api.TrackInfoIsForBus != 0 {
			p.Logger.Info("  This is a bus track")
		}
		if info.Flags&api.TrackInfoIsForMaster != 0 {
			p.Logger.Info("  This is the master track")
		}
	}
	
	// For a more sophisticated plugin, you might:
	// - Adjust default parameter values based on track type
	// - Configure audio processing differently for bus/master tracks
	// - Update GUI to show track information
}

// Context Menu Extension Methods

// PopulateContextMenu builds the context menu for the plugin
func (p *GainPlugin) PopulateContextMenu(target *api.ContextMenuTarget, builder *api.ContextMenuBuilder) bool {
	// Check main thread (context menu is always on main thread)
	api.DebugAssertMainThread("GainPlugin.PopulateContextMenu")
	
	if target != nil && target.Kind == api.ContextMenuTargetKindParam {
		// Use helper for common parameter menu items
		p.contextMenuProvider.PopulateParameterMenu(uint32(target.ID), builder)
		
		// Add gain-specific presets
		if target.ID == 0 { // Gain parameter
			api.AddParameterPresetSubmenu(builder, "Presets", []struct {
				Label    string
				Value    float64
				ActionID uint64
			}{
				{"-6 dB", math.Pow(10, -6.0/20.0), 1001},
				{"-3 dB", math.Pow(10, -3.0/20.0), 1002},
				{"+3 dB", math.Pow(10, 3.0/20.0), 1003},
				{"+6 dB", math.Pow(10, 6.0/20.0), 1004},
			})
		}
	} else {
		// Use helper for common global menu items
		p.contextMenuProvider.PopulateGlobalMenu(builder)
	}
	
	return true
}

// PerformContextMenuAction handles context menu actions
func (p *GainPlugin) PerformContextMenuAction(target *api.ContextMenuTarget, actionID uint64) bool {
	// Check main thread (context menu is always on main thread)
	api.DebugAssertMainThread("GainPlugin.PerformContextMenuAction")
	
	// Check for common actions first
	if isReset, paramID := p.contextMenuProvider.IsResetAction(actionID); isReset {
		// For gain parameter, also update atomic value
		if paramID == 0 {
			p.gain.Store(1.0)
		}
		return p.contextMenuProvider.HandleResetParameter(paramID)
	}
	
	if p.contextMenuProvider.IsAboutAction(actionID) {
		if p.Logger != nil {
			p.Logger.Info("Gain Plugin v1.0.0 - A simple gain adjustment plugin")
		}
		return true
	}
	
	// Handle gain preset actions
	var value float64
	switch actionID {
	case 1001: // -6 dB
		value = math.Pow(10, -6.0/20.0)
	case 1002: // -3 dB
		value = math.Pow(10, -3.0/20.0)
	case 1003: // +3 dB
		value = math.Pow(10, 3.0/20.0)
	case 1004: // +6 dB
		value = math.Pow(10, 6.0/20.0)
	default:
		return false
	}
	
	// Apply preset value
	p.ParamManager.SetValue(0, value)
	p.gain.Store(value)
	// TODO: Request parameter flush to notify host
	return true
}

// Remote Controls Extension Methods

// GetRemoteControlsPageCount returns the number of remote control pages
func (p *GainPlugin) GetRemoteControlsPageCount() uint32 {
	return 1 // Single page for gain control
}

// GetRemoteControlsPage returns the remote control page at the given index
func (p *GainPlugin) GetRemoteControlsPage(pageIndex uint32) (*api.RemoteControlsPage, bool) {
	if pageIndex != 0 {
		return nil, false
	}
	
	// Create a simple page with the gain parameter
	page := &api.RemoteControlsPage{
		SectionName: "Main",
		PageID:      1,
		PageName:    "Gain Control",
		ParamIDs:    [api.RemoteControlsCount]uint32{0, 0, 0, 0, 0, 0, 0, 0}, // First slot has gain param
		IsForPreset: false, // Device page, not preset-specific
	}
	
	return page, true
}

// Param Indication Extension Methods

// OnParamMappingSet is called when the host sets or clears a mapping indication
func (p *GainPlugin) OnParamMappingSet(paramID uint32, hasMapping bool, color *api.Color, label string, description string) {
	// Check main thread (param indication is always on main thread)
	api.DebugAssertMainThread("GainPlugin.OnParamMappingSet")
	
	// Log the mapping change
	if p.Logger != nil {
		if hasMapping {
			p.Logger.Info(fmt.Sprintf("Parameter %d mapped to %s: %s", paramID, label, description))
			if color != nil {
				p.Logger.Info(fmt.Sprintf("  Color: R=%d G=%d B=%d A=%d", color.Red, color.Green, color.Blue, color.Alpha))
			}
		} else {
			p.Logger.Info(fmt.Sprintf("Parameter %d mapping cleared", paramID))
		}
	}
	
	// In a real plugin with GUI, you would update the visual indication here
}

// OnParamAutomationSet is called when the host sets or clears an automation indication
func (p *GainPlugin) OnParamAutomationSet(paramID uint32, automationState uint32, color *api.Color) {
	// Check main thread (param indication is always on main thread)
	api.DebugAssertMainThread("GainPlugin.OnParamAutomationSet")
	
	// Log the automation state change
	if p.Logger != nil {
		var stateStr string
		switch automationState {
		case api.ParamIndicationAutomationNone:
			stateStr = "None"
		case api.ParamIndicationAutomationPresent:
			stateStr = "Present"
		case api.ParamIndicationAutomationPlaying:
			stateStr = "Playing"
		case api.ParamIndicationAutomationRecording:
			stateStr = "Recording"
		case api.ParamIndicationAutomationOverriding:
			stateStr = "Overriding"
		default:
			stateStr = "Unknown"
		}
		
		p.Logger.Info(fmt.Sprintf("Parameter %d automation state: %s", paramID, stateStr))
		if color != nil {
			p.Logger.Info(fmt.Sprintf("  Color: R=%d G=%d B=%d A=%d", color.Red, color.Green, color.Blue, color.Alpha))
		}
	}
	
	// In a real plugin with GUI, you would update the visual indication here
}

// SaveState saves the plugin state to a stream
func (p *GainPlugin) SaveState(stream unsafe.Pointer) bool {
	out := api.NewOutputStream(stream)
	
	// Create state with current parameter values
	params := []state.Parameter{
		{ID: ParamGain, Value: p.gain.Load(), Name: "gain"},
	}
	stateData := p.StateManager.CreateState(params, nil)
	
	// Convert to JSON and save
	jsonData, err := p.StateManager.SaveToJSON(stateData)
	if err != nil {
		if p.Logger != nil {
			p.Logger.Error(fmt.Sprintf("Failed to serialize state: %v", err))
		}
		return false
	}
	
	// Write JSON data as string to stream
	if err := out.WriteString(string(jsonData)); err != nil {
		if p.Logger != nil {
			p.Logger.Error(fmt.Sprintf("Failed to write state: %v", err))
		}
		return false
	}
	
	return true
}

// LoadState loads the plugin state from a stream
func (p *GainPlugin) LoadState(stream unsafe.Pointer) bool {
	in := api.NewInputStream(stream)
	
	// Read JSON data as string from stream
	jsonString, err := in.ReadString()
	if err != nil {
		if p.Logger != nil {
			p.Logger.Error(fmt.Sprintf("Failed to read state data: %v", err))
		}
		return false
	}
	
	// Load and validate state
	stateData, err := p.StateManager.LoadFromJSON([]byte(jsonString))
	if err != nil {
		if p.Logger != nil {
			p.Logger.Error(fmt.Sprintf("Failed to deserialize state: %v", err))
		}
		return false
	}
	
	// Apply loaded parameters
	for _, param := range stateData.Parameters {
		if param.ID == ParamGain {
			p.gain.Store(param.Value)
			if err := p.ParamManager.SetValue(param.ID, param.Value); err != nil {
				if p.Logger != nil {
					p.Logger.Warning(fmt.Sprintf("Failed to set parameter %d: %v", param.ID, err))
				}
			}
		}
	}
	
	// Get the loaded gain value
	value := p.gain.Load()
	p.gain.Store(value)
	
	return true
}

// SaveStateWithContext saves the plugin state to a stream with context
func (p *GainPlugin) SaveStateWithContext(stream unsafe.Pointer, contextType uint32) bool {
	// Log the context type
	if p.Logger != nil {
		switch contextType {
		case api.StateContextForPreset:
			p.Logger.Info("Saving state for preset")
		case api.StateContextForDuplicate:
			p.Logger.Info("Saving state for duplicate")
		case api.StateContextForProject:
			p.Logger.Info("Saving state for project")
		default:
			p.Logger.Info(fmt.Sprintf("Saving state with unknown context: %d", contextType))
		}
	}
	
	// For this simple gain plugin, we save the same data regardless of context
	// More complex plugins might save different data based on context
	return p.SaveState(stream)
}

// LoadStateWithContext loads the plugin state from a stream with context
func (p *GainPlugin) LoadStateWithContext(stream unsafe.Pointer, contextType uint32) bool {
	// Log the context type
	if p.Logger != nil {
		switch contextType {
		case api.StateContextForPreset:
			p.Logger.Info("Loading state for preset")
		case api.StateContextForDuplicate:
			p.Logger.Info("Loading state for duplicate")
		case api.StateContextForProject:
			p.Logger.Info("Loading state for project")
		default:
			p.Logger.Info(fmt.Sprintf("Loading state with unknown context: %d", contextType))
		}
	}
	
	// For this simple gain plugin, we load the same data regardless of context
	// More complex plugins might handle loading differently based on context
	return p.LoadState(stream)
}

// LoadPresetFromLocation - preset loading is handled externally via JSON files
func (p *GainPlugin) LoadPresetFromLocation(locationKind uint32, location string, loadKey string) bool {
	// Presets are handled externally by the host via JSON files
	// The plugin doesn't need to implement preset loading
	return false
}

// Helper functions for atomic float64 operations

// Audio port configuration is provided by embedded StereoPortProvider and SurroundSupport

//export ClapGo_PluginStateSave
func ClapGo_PluginStateSave(plugin unsafe.Pointer, stream unsafe.Pointer) C.bool {
	if plugin == nil || stream == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	return C.bool(p.SaveState(stream))
}

//export ClapGo_PluginStateLoad
func ClapGo_PluginStateLoad(plugin unsafe.Pointer, stream unsafe.Pointer) C.bool {
	if plugin == nil || stream == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	return C.bool(p.LoadState(stream))
}

// State Context Extension Exports

//export ClapGo_PluginStateSaveWithContext
func ClapGo_PluginStateSaveWithContext(plugin unsafe.Pointer, stream unsafe.Pointer, contextType C.uint32_t) C.bool {
	if plugin == nil || stream == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	return C.bool(p.SaveStateWithContext(stream, uint32(contextType)))
}

//export ClapGo_PluginStateLoadWithContext
func ClapGo_PluginStateLoadWithContext(plugin unsafe.Pointer, stream unsafe.Pointer, contextType C.uint32_t) C.bool {
	if plugin == nil || stream == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	return C.bool(p.LoadStateWithContext(stream, uint32(contextType)))
}

// Preset Load Extension Exports

//export ClapGo_PluginPresetLoadFromLocation
func ClapGo_PluginPresetLoadFromLocation(plugin unsafe.Pointer, locationKind C.uint32_t, location *C.char, loadKey *C.char) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	
	// Convert C strings to Go strings
	var locationStr, loadKeyStr string
	if location != nil {
		locationStr = C.GoString(location)
	}
	if loadKey != nil {
		loadKeyStr = C.GoString(loadKey)
	}
	
	return C.bool(p.LoadPresetFromLocation(uint32(locationKind), locationStr, loadKeyStr))
}

// Phase 3 Extension Exports

//export ClapGo_PluginLatencyGet
func ClapGo_PluginLatencyGet(plugin unsafe.Pointer) uint32 {
	if plugin == nil {
		return 0
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	return p.GetLatency()
}

//export ClapGo_PluginTailGet
func ClapGo_PluginTailGet(plugin unsafe.Pointer) C.uint32_t {
	if plugin == nil {
		return 0
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	return C.uint32_t(p.GetTail())
}

//export ClapGo_PluginOnTimer
func ClapGo_PluginOnTimer(plugin unsafe.Pointer, timerID C.uint64_t) {
	if plugin == nil {
		return
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	p.OnTimer(uint64(timerID))
}

// Phase 7 Extension Exports

//export ClapGo_PluginTrackInfoChanged
func ClapGo_PluginTrackInfoChanged(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	p.OnTrackInfoChanged()
}

// Context Menu Extension Exports

//export ClapGo_PluginContextMenuPopulate
func ClapGo_PluginContextMenuPopulate(plugin unsafe.Pointer, targetKind C.uint32_t, targetID C.uint64_t, builder unsafe.Pointer) C.bool {
	if plugin == nil || builder == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	
	// Create target
	var target *api.ContextMenuTarget
	if targetKind != api.ContextMenuTargetKindGlobal {
		target = &api.ContextMenuTarget{
			Kind: uint32(targetKind),
			ID:   uint64(targetID),
		}
	}
	
	// Create builder wrapper
	menuBuilder := api.NewContextMenuBuilder(builder)
	
	return C.bool(p.PopulateContextMenu(target, menuBuilder))
}

//export ClapGo_PluginContextMenuPerform
func ClapGo_PluginContextMenuPerform(plugin unsafe.Pointer, targetKind C.uint32_t, targetID C.uint64_t, actionID C.uint64_t) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	
	// Create target
	var target *api.ContextMenuTarget
	if targetKind != api.ContextMenuTargetKindGlobal {
		target = &api.ContextMenuTarget{
			Kind: uint32(targetKind),
			ID:   uint64(targetID),
		}
	}
	
	return C.bool(p.PerformContextMenuAction(target, uint64(actionID)))
}

// Remote Controls Extension Exports

//export ClapGo_PluginRemoteControlsCount
func ClapGo_PluginRemoteControlsCount(plugin unsafe.Pointer) C.uint32_t {
	if plugin == nil {
		return 0
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	return C.uint32_t(p.GetRemoteControlsPageCount())
}

//export ClapGo_PluginRemoteControlsGet
func ClapGo_PluginRemoteControlsGet(plugin unsafe.Pointer, pageIndex C.uint32_t, cPage unsafe.Pointer) C.bool {
	if plugin == nil || cPage == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	
	page, ok := p.GetRemoteControlsPage(uint32(pageIndex))
	if !ok {
		return C.bool(false)
	}
	
	// Convert Go page to C structure
	api.RemoteControlsPageToC(page, cPage)
	return C.bool(true)
}

// Param Indication Extension Exports

//export ClapGo_PluginParamIndicationSetMapping
func ClapGo_PluginParamIndicationSetMapping(plugin unsafe.Pointer, paramID C.uint64_t, hasMapping C.bool, color unsafe.Pointer, label *C.char, description *C.char) {
	if plugin == nil {
		return
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	
	var labelStr, descStr string
	if label != nil {
		labelStr = C.GoString(label)
	}
	if description != nil {
		descStr = C.GoString(description)
	}
	
	p.OnParamMappingSet(uint32(paramID), bool(hasMapping), api.ColorFromC(color), labelStr, descStr)
}

//export ClapGo_PluginParamIndicationSetAutomation
func ClapGo_PluginParamIndicationSetAutomation(plugin unsafe.Pointer, paramID C.uint64_t, automationState C.uint32_t, color unsafe.Pointer) {
	if plugin == nil {
		return
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	
	p.OnParamAutomationSet(uint32(paramID), uint32(automationState), api.ColorFromC(color))
}

func main() {
	// This is not called when used as a plugin,
	// but can be useful for testing
}