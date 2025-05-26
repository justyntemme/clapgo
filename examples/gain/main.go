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
	"embed"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"github.com/justyntemme/clapgo/pkg/api"
	"runtime/cgo"
	"sync/atomic"
	"unsafe"
)

// Embed factory presets
//go:embed presets/factory/*.json
var factoryPresets embed.FS

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
		// Store the host pointer and create utilities
		gainPlugin.host = host
		gainPlugin.logger = api.NewHostLogger(host)
		
		// Log plugin creation
		if gainPlugin.logger != nil {
			gainPlugin.logger.Info("Creating gain plugin instance")
		}
		
		// Create a CGO handle to safely pass the Go object to C
		handle := cgo.NewHandle(gainPlugin)
		fmt.Printf("Created plugin instance: %s, handle: %v\n", id, handle)
		
		// Log handle creation
		if gainPlugin.logger != nil {
			gainPlugin.logger.Debug(fmt.Sprintf("[ClapGo_CreatePlugin] Created handle: %v", handle))
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
	if p.logger != nil {
		p.logger.Debug("[ClapGo_PluginInit] Starting plugin initialization")
	}
	
	result := p.Init()
	
	// Log the result
	if p.logger != nil {
		if result {
			p.logger.Info("[ClapGo_PluginInit] Plugin initialization successful")
		} else {
			p.logger.Error("[ClapGo_PluginInit] Plugin initialization failed")
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
	if p.logger != nil {
		p.logger.Debug(fmt.Sprintf("[ClapGo_PluginDestroy] Destroying plugin instance, handle: %v", handle))
	}
	
	p.Destroy()
	
	// Unregister from audio ports provider
	api.UnregisterAudioPortsProvider(plugin)
	
	// Log completion
	if p.logger != nil {
		p.logger.Info("[ClapGo_PluginDestroy] Plugin instance destroyed successfully")
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
	if p.logger != nil {
		p.logger.Debug(fmt.Sprintf("[ClapGo_PluginActivate] Activating plugin - SR: %.0f, frames: %d-%d", 
			float64(sampleRate), uint32(minFrames), uint32(maxFrames)))
	}
	
	result := p.Activate(float64(sampleRate), uint32(minFrames), uint32(maxFrames))
	
	// Log result
	if p.logger != nil {
		if result {
			p.logger.Info("[ClapGo_PluginActivate] Plugin activation successful")
		} else {
			p.logger.Error("[ClapGo_PluginActivate] Plugin activation failed")
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
	if p.logger != nil {
		p.logger.Debug("[ClapGo_PluginDeactivate] Deactivating plugin")
	}
	
	p.Deactivate()
	
	// Log completion
	if p.logger != nil {
		p.logger.Info("[ClapGo_PluginDeactivate] Plugin deactivation successful")
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
	if p.logger != nil {
		p.logger.Debug("[ClapGo_PluginStartProcessing] Starting audio processing")
	}
	
	result := p.StartProcessing()
	
	// Log result
	if p.logger != nil {
		if result {
			p.logger.Info("[ClapGo_PluginStartProcessing] Audio processing started successfully")
		} else {
			p.logger.Error("[ClapGo_PluginStartProcessing] Failed to start audio processing")
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
	if p.logger != nil {
		p.logger.Debug("[ClapGo_PluginStopProcessing] Stopping audio processing")
	}
	
	p.StopProcessing()
	
	// Log completion
	if p.logger != nil {
		p.logger.Info("[ClapGo_PluginStopProcessing] Audio processing stopped successfully")
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
	if p.logger != nil {
		p.logger.Debug("[ClapGo_PluginReset] Resetting plugin state")
	}
	
	p.Reset()
	
	// Log completion
	if p.logger != nil {
		p.logger.Info("[ClapGo_PluginReset] Plugin reset successful")
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
	
	// Set logger on the event pool for diagnostics
	if pool := eventHandler.GetEventPool(); pool != nil && p.logger != nil {
		pool.SetLogger(p.logger)
	}
	
	// Call the actual Go process method
	result := p.Process(steadyTime, framesCount, audioIn, audioOut, eventHandler)
	
	// Periodically log event pool diagnostics (every 1000 process calls)
	p.processCallCount++
	if p.processCallCount % 1000 == 0 && p.processCallCount != p.lastEventPoolDump {
		p.lastEventPoolDump = p.processCallCount
		if pool := eventHandler.GetEventPool(); pool != nil {
			pool.LogDiagnostics()
		}
	}
	
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
	return C.uint32_t(p.paramManager.GetParameterCount())
}

//export ClapGo_PluginParamsGetInfo
func ClapGo_PluginParamsGetInfo(plugin unsafe.Pointer, index C.uint32_t, info unsafe.Pointer) C.bool {
	if plugin == nil || info == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	
	// Get parameter info from manager
	paramInfo, err := p.paramManager.GetParameterInfoByIndex(uint32(index))
	if err != nil {
		return C.bool(false)
	}
	
	// Convert to C struct using helper
	api.ParamInfoToC(paramInfo, info)
	
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
		gainBits := atomic.LoadInt64(&p.gain)
		*value = C.double(api.FloatFromBits(uint64(gainBits)))
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
		text := api.FormatParameterValue(float64(value), api.FormatDecibel)
		
		// Copy to C buffer using helper
		api.CopyStringToCBuffer(text, unsafe.Pointer(buffer), int(size))
		
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
	
	// For gain parameter, parse dB value
	if uint32(paramID) == 0 {
		var db float64
		if _, err := fmt.Sscanf(goText, "%f", &db); err == nil {
			// Convert from dB to linear
			linear := math.Pow(10.0, db/20.0)
			
			// Clamp to valid range
			if linear < 0.0 {
				linear = 0.0
			}
			if linear > 2.0 {
				linear = 2.0
			}
			
			*value = C.double(linear)
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

// GainPlugin represents the gain plugin with atomic parameter storage
type GainPlugin struct {
	// Plugin state
	sampleRate   float64
	isActivated  bool
	isProcessing bool
	host         unsafe.Pointer
	
	// Parameters with atomic storage for thread safety
	gain         int64  // atomic storage for gain value
	
	// Parameter management using our new abstraction
	paramManager *api.ParameterManager
	
	// Host utilities
	logger       *api.HostLogger
	trackInfo    *api.HostTrackInfo
	threadCheck  *api.ThreadChecker
	contextMenuProvider *api.DefaultContextMenuProvider
	
	// Diagnostics
	processCallCount uint64
	lastEventPoolDump uint64
	
	// Latency in samples (for this example, gain has no latency)
	latency uint32
}

// NewGainPlugin creates a new gain plugin instance
func NewGainPlugin() *GainPlugin {
	plugin := &GainPlugin{
		sampleRate:   44100.0,
		isActivated:  false,
		isProcessing: false,
		paramManager: api.NewParameterManager(),
	}
	
	// Set default gain to 1.0 (0dB)
	atomic.StoreInt64(&plugin.gain, int64(api.FloatToBits(1.0)))
	
	// Register parameters using helper function
	plugin.paramManager.RegisterParameter(api.CreateVolumeParameter(0, "Gain"))
	plugin.paramManager.SetParameterValue(0, 1.0)
	
	return plugin
}

// Init initializes the plugin
func (p *GainPlugin) Init() bool {
	// Mark this as main thread for debug builds
	api.DebugSetMainThread()
	
	// Initialize thread checker
	if p.host != nil {
		p.threadCheck = api.NewThreadChecker(p.host)
		if p.threadCheck.IsAvailable() && p.logger != nil {
			p.logger.Info("Thread Check extension available - thread safety validation enabled")
		}
	}
	
	// Initialize track info helper
	if p.host != nil {
		p.trackInfo = api.NewHostTrackInfo(p.host)
	}
	
	// Initialize context menu provider
	if p.host != nil {
		p.contextMenuProvider = api.NewDefaultContextMenuProvider(
			p.paramManager,
			"Gain Plugin",
			"1.0.0",
			p.host,
		)
		p.contextMenuProvider.SetAboutMessage("Gain Plugin v1.0.0 - A simple gain adjustment plugin")
	}
	
	if p.logger != nil {
		p.logger.Debug("Gain plugin initialized")
	}
	
	// Check initial track info
	if p.trackInfo != nil {
		p.OnTrackInfoChanged()
	}
	
	return true
}

// Destroy cleans up plugin resources
func (p *GainPlugin) Destroy() {
	// Assert main thread
	api.DebugAssertMainThread("GainPlugin.Destroy")
	if p.threadCheck != nil {
		p.threadCheck.AssertMainThread("GainPlugin.Destroy")
	}
	
	// Nothing to clean up in the plugin itself
}

// Activate prepares the plugin for processing
func (p *GainPlugin) Activate(sampleRate float64, minFrames, maxFrames uint32) bool {
	// Assert main thread
	api.DebugAssertMainThread("GainPlugin.Activate")
	if p.threadCheck != nil {
		p.threadCheck.AssertMainThread("GainPlugin.Activate")
	}
	
	p.sampleRate = sampleRate
	p.isActivated = true
	
	if p.logger != nil {
		p.logger.Info(fmt.Sprintf("Plugin activated at %.0f Hz, buffer size %d-%d", sampleRate, minFrames, maxFrames))
	}
	
	return true
}

// Deactivate stops the plugin from processing
func (p *GainPlugin) Deactivate() {
	p.isActivated = false
}

// StartProcessing begins audio processing
func (p *GainPlugin) StartProcessing() bool {
	// Assert audio thread
	api.DebugAssertAudioThread("GainPlugin.StartProcessing")
	if p.threadCheck != nil {
		p.threadCheck.AssertAudioThread("GainPlugin.StartProcessing")
	}
	
	if !p.isActivated {
		return false
	}
	p.isProcessing = true
	return true
}

// StopProcessing ends audio processing
func (p *GainPlugin) StopProcessing() {
	p.isProcessing = false
}

// Reset resets the plugin state
func (p *GainPlugin) Reset() {
	// Reset gain to default
	atomic.StoreInt64(&p.gain, int64(api.FloatToBits(1.0)))
}

// Process processes audio data using the new abstractions
func (p *GainPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events api.EventHandler) int {
	// Assert audio thread
	api.DebugAssertAudioThread("GainPlugin.Process")
	if p.threadCheck != nil {
		p.threadCheck.AssertAudioThread("GainPlugin.Process")
	}
	
	// Check if we're in a valid state for processing
	if !p.isActivated || !p.isProcessing {
		return api.ProcessError
	}
	
	// Process events using our new abstraction - NO MORE MANUAL EVENT PARSING!
	if events != nil {
		p.processEvents(events, framesCount)
	}
	
	// Get current gain value atomically
	gainBits := atomic.LoadInt64(&p.gain)
	gain := api.FloatFromBits(uint64(gainBits))
	
	// If no audio inputs or outputs, nothing to do
	if len(audioIn) == 0 || len(audioOut) == 0 {
		return api.ProcessContinue
	}
	
	// Get the number of channels (use min of input and output)
	numChannels := len(audioIn)
	if len(audioOut) < numChannels {
		numChannels = len(audioOut)
	}
	
	// Process audio - apply gain to each sample
	for ch := 0; ch < numChannels; ch++ {
		inChannel := audioIn[ch]
		outChannel := audioOut[ch]
		
		// Make sure we have enough buffer space
		if len(inChannel) < int(framesCount) || len(outChannel) < int(framesCount) {
			continue // Skip this channel if buffer is too small
		}
		
		// Apply gain to each sample
		for i := uint32(0); i < framesCount; i++ {
			outChannel[i] = inChannel[i] * float32(gain)
		}
	}
	
	return api.ProcessContinue
}

// processEvents handles all incoming events using our new EventHandler abstraction
func (p *GainPlugin) processEvents(events api.EventHandler, frameCount uint32) {
	if events == nil {
		return
	}
	
	// Use the zero-allocation ProcessTypedEvents method instead of ProcessAllEvents
	events.ProcessTypedEvents(p)
}

// TypedEventHandler implementation for zero-allocation event processing

// HandleParamValue handles parameter value changes
func (p *GainPlugin) HandleParamValue(event *api.ParamValueEvent, time uint32) {
	// Handle the parameter change based on its ID
	switch event.ParamID {
	case 0: // Gain parameter
		// Clamp value to valid range
		value := event.Value
		if value < 0.0 {
			value = 0.0
		}
		if value > 2.0 {
			value = 2.0
		}
		atomic.StoreInt64(&p.gain, int64(api.FloatToBits(value)))
		
		// Update parameter manager
		p.paramManager.SetParameterValue(event.ParamID, value)
		
		// Log the parameter change
		if p.logger != nil {
			// Convert to dB for logging
			db := 20.0 * math.Log10(value)
			p.logger.Debug(fmt.Sprintf("Gain changed to %.2f dB", db))
		}
	}
}

// HandleParamMod handles parameter modulation events
func (p *GainPlugin) HandleParamMod(event *api.ParamModEvent, time uint32) {
	// Gain plugin doesn't support parameter modulation
}

// HandleParamGestureBegin handles parameter gesture begin events
func (p *GainPlugin) HandleParamGestureBegin(event *api.ParamGestureEvent, time uint32) {
	// Could be used to start automation recording
}

// HandleParamGestureEnd handles parameter gesture end events
func (p *GainPlugin) HandleParamGestureEnd(event *api.ParamGestureEvent, time uint32) {
	// Could be used to end automation recording
}

// HandleNoteOn handles note on events
func (p *GainPlugin) HandleNoteOn(event *api.NoteEvent, time uint32) {
	// Gain plugin doesn't process notes
}

// HandleNoteOff handles note off events
func (p *GainPlugin) HandleNoteOff(event *api.NoteEvent, time uint32) {
	// Gain plugin doesn't process notes
}

// HandleNoteChoke handles note choke events
func (p *GainPlugin) HandleNoteChoke(event *api.NoteEvent, time uint32) {
	// Gain plugin doesn't process notes
}

// HandleNoteEnd handles note end events
func (p *GainPlugin) HandleNoteEnd(event *api.NoteEvent, time uint32) {
	// Gain plugin doesn't process notes
}

// HandleNoteExpression handles note expression events
func (p *GainPlugin) HandleNoteExpression(event *api.NoteExpressionEvent, time uint32) {
	// Gain plugin doesn't support note expression
}

// HandleTransport handles transport events
func (p *GainPlugin) HandleTransport(event *api.TransportEvent, time uint32) {
	// Gain plugin doesn't use transport information
}

// HandleMIDI handles MIDI events
func (p *GainPlugin) HandleMIDI(event *api.MIDIEvent, time uint32) {
	// Gain plugin doesn't process MIDI
}

// HandleMIDISysex handles MIDI sysex events
func (p *GainPlugin) HandleMIDISysex(event *api.MIDISysexEvent, time uint32) {
	// Gain plugin doesn't process sysex
}

// HandleMIDI2 handles MIDI 2.0 events
func (p *GainPlugin) HandleMIDI2(event *api.MIDI2Event, time uint32) {
	// Gain plugin doesn't process MIDI 2.0
}

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
	if p.trackInfo == nil {
		return
	}
	
	// Get the new track information
	info, ok := p.trackInfo.GetTrackInfo()
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
		if info.Flags&api.TrackInfoIsForReturnTrack != 0 {
			p.logger.Info("  This is a return track")
		}
		if info.Flags&api.TrackInfoIsForBus != 0 {
			p.logger.Info("  This is a bus track")
		}
		if info.Flags&api.TrackInfoIsForMaster != 0 {
			p.logger.Info("  This is the master track")
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
			atomic.StoreInt64(&p.gain, int64(api.FloatToBits(1.0)))
		}
		return p.contextMenuProvider.HandleResetParameter(paramID)
	}
	
	if p.contextMenuProvider.IsAboutAction(actionID) {
		if p.logger != nil {
			p.logger.Info("Gain Plugin v1.0.0 - A simple gain adjustment plugin")
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
	p.paramManager.SetParameterValue(0, value)
	atomic.StoreInt64(&p.gain, int64(api.FloatToBits(value)))
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
	if p.logger != nil {
		if hasMapping {
			p.logger.Info(fmt.Sprintf("Parameter %d mapped to %s: %s", paramID, label, description))
			if color != nil {
				p.logger.Info(fmt.Sprintf("  Color: R=%d G=%d B=%d A=%d", color.Red, color.Green, color.Blue, color.Alpha))
			}
		} else {
			p.logger.Info(fmt.Sprintf("Parameter %d mapping cleared", paramID))
		}
	}
	
	// In a real plugin with GUI, you would update the visual indication here
}

// OnParamAutomationSet is called when the host sets or clears an automation indication
func (p *GainPlugin) OnParamAutomationSet(paramID uint32, automationState uint32, color *api.Color) {
	// Check main thread (param indication is always on main thread)
	api.DebugAssertMainThread("GainPlugin.OnParamAutomationSet")
	
	// Log the automation state change
	if p.logger != nil {
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
		
		p.logger.Info(fmt.Sprintf("Parameter %d automation state: %s", paramID, stateStr))
		if color != nil {
			p.logger.Info(fmt.Sprintf("  Color: R=%d G=%d B=%d A=%d", color.Red, color.Green, color.Blue, color.Alpha))
		}
	}
	
	// In a real plugin with GUI, you would update the visual indication here
}

// SaveState saves the plugin state to a stream
func (p *GainPlugin) SaveState(stream unsafe.Pointer) bool {
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
	
	return true
}

// LoadState loads the plugin state from a stream
func (p *GainPlugin) LoadState(stream unsafe.Pointer) bool {
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
		
		// Update internal state if this is the gain parameter
		if paramID == 0 {
			atomic.StoreInt64(&p.gain, int64(api.FloatToBits(value)))
		}
	}
	
	return true
}

// SaveStateWithContext saves the plugin state to a stream with context
func (p *GainPlugin) SaveStateWithContext(stream unsafe.Pointer, contextType uint32) bool {
	// Log the context type
	if p.logger != nil {
		switch contextType {
		case api.StateContextForPreset:
			p.logger.Info("Saving state for preset")
		case api.StateContextForDuplicate:
			p.logger.Info("Saving state for duplicate")
		case api.StateContextForProject:
			p.logger.Info("Saving state for project")
		default:
			p.logger.Info(fmt.Sprintf("Saving state with unknown context: %d", contextType))
		}
	}
	
	// For this simple gain plugin, we save the same data regardless of context
	// More complex plugins might save different data based on context
	return p.SaveState(stream)
}

// LoadStateWithContext loads the plugin state from a stream with context
func (p *GainPlugin) LoadStateWithContext(stream unsafe.Pointer, contextType uint32) bool {
	// Log the context type
	if p.logger != nil {
		switch contextType {
		case api.StateContextForPreset:
			p.logger.Info("Loading state for preset")
		case api.StateContextForDuplicate:
			p.logger.Info("Loading state for duplicate")
		case api.StateContextForProject:
			p.logger.Info("Loading state for project")
		default:
			p.logger.Info(fmt.Sprintf("Loading state with unknown context: %d", contextType))
		}
	}
	
	// For this simple gain plugin, we load the same data regardless of context
	// More complex plugins might handle loading differently based on context
	return p.LoadState(stream)
}

// LoadPresetFromLocation loads a preset from the specified location
func (p *GainPlugin) LoadPresetFromLocation(locationKind uint32, location string, loadKey string) bool {
	// Log the preset load request
	if p.logger != nil {
		switch locationKind {
		case api.PresetLocationFile:
			p.logger.Info(fmt.Sprintf("Loading preset from file: %s (key: %s)", location, loadKey))
		case api.PresetLocationPlugin:
			p.logger.Info(fmt.Sprintf("Loading bundled preset (key: %s)", loadKey))
		default:
			p.logger.Warning(fmt.Sprintf("Unknown preset location kind: %d", locationKind))
			return false
		}
	}
	
	// Handle different location kinds
	switch locationKind {
	case api.PresetLocationFile:
		// Load preset from file
		if location == "" {
			return false
		}
		
		// Read the preset file
		presetData, err := os.ReadFile(location)
		if err != nil {
			if p.logger != nil {
				p.logger.Error(fmt.Sprintf("Failed to read preset file: %v", err))
			}
			return false
		}
		
		// Parse the preset data
		var state map[string]interface{}
		if err := json.Unmarshal(presetData, &state); err != nil {
			if p.logger != nil {
				p.logger.Error(fmt.Sprintf("Failed to parse preset file: %v", err))
			}
			return false
		}
		
		// Load the preset state from preset_data field
		if presetData, ok := state["preset_data"].(map[string]interface{}); ok {
			if gainValue, ok := presetData["gain"].(float64); ok {
				atomic.StoreInt64(&p.gain, int64(api.FloatToBits(gainValue)))
				p.paramManager.SetParameterValue(0, gainValue)
				
				if p.logger != nil {
					p.logger.Info(fmt.Sprintf("Preset loaded: gain = %.2f", gainValue))
				}
				return true
			}
		}
		
		if p.logger != nil {
			p.logger.Error("Invalid preset format: missing preset_data.gain value")
		}
		return false
		
	case api.PresetLocationPlugin:
		// Load bundled preset from embedded files
		presetPath := filepath.Join("presets", "factory", loadKey+".json")
		
		// Read from embedded filesystem
		presetData, err := factoryPresets.ReadFile(presetPath)
		if err != nil {
			if p.logger != nil {
				p.logger.Error(fmt.Sprintf("Failed to load bundled preset '%s': %v", loadKey, err))
			}
			return false
		}
		
		// Parse the preset data
		var state map[string]interface{}
		if err := json.Unmarshal(presetData, &state); err != nil {
			if p.logger != nil {
				p.logger.Error(fmt.Sprintf("Failed to parse bundled preset: %v", err))
			}
			return false
		}
		
		// Load the preset state from preset_data field
		if presetData, ok := state["preset_data"].(map[string]interface{}); ok {
			if gainValue, ok := presetData["gain"].(float64); ok {
				atomic.StoreInt64(&p.gain, int64(api.FloatToBits(gainValue)))
				p.paramManager.SetParameterValue(0, gainValue)
				
				if p.logger != nil {
					p.logger.Info(fmt.Sprintf("Bundled preset '%s' loaded: gain = %.2f", loadKey, gainValue))
				}
				return true
			}
		}
		
		if p.logger != nil {
			p.logger.Error("Invalid preset format: missing preset_data.gain value")
		}
		return false
		
	default:
		return false
	}
}

// Helper functions for atomic float64 operations

// AudioPortsProvider implementation
// This demonstrates custom audio port configuration

// GetAudioPortCount returns the number of audio ports
func (p *GainPlugin) GetAudioPortCount(isInput bool) uint32 {
	// Gain plugin has 1 stereo input and 1 stereo output
	return 1
}

// GetAudioPortInfo returns information about an audio port
func (p *GainPlugin) GetAudioPortInfo(index uint32, isInput bool) api.AudioPortInfo {
	if index != 0 {
		// Return invalid port info for out-of-range index
		return api.AudioPortInfo{
			ID: api.InvalidID,
		}
	}
	
	// Create stereo port info
	name := "Stereo Input"
	if !isInput {
		name = "Stereo Output"
	}
	
	return api.CreateStereoPort(0, name, true)
}

// SurroundProvider implementation (optional - demonstrates surround support)

// IsChannelMaskSupported checks if the plugin supports a given channel mask
func (p *GainPlugin) IsChannelMaskSupported(channelMask uint64) bool {
	// For this example, we only support stereo
	return channelMask == api.ChannelMaskStereo
}

// GetChannelMap returns the channel map for a given port
func (p *GainPlugin) GetChannelMap(isInput bool, portIndex uint32) []uint8 {
	// For stereo, return FL and FR channel identifiers
	if portIndex == 0 {
		return api.CreateStereoChannelMap()
	}
	return nil
}

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