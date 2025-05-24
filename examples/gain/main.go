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
	"os"
	"github.com/justyntemme/clapgo/pkg/api"
	"runtime/cgo"
	"sync/atomic"
	"unsafe"
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
		// Store the host pointer and create utilities
		gainPlugin.host = host
		gainPlugin.logger = api.NewHostLogger(host)
		
		// Log plugin creation
		if gainPlugin.logger != nil {
			gainPlugin.logger.Info("Creating gain plugin instance")
		}
		
		// Create a CGO handle to safely pass the Go object to C
		handle := cgo.NewHandle(gainPlugin)
		fmt.Printf("Created plugin instance: %s\n", id)
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
	if plugin == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	return C.bool(p.Init())
}

//export ClapGo_PluginDestroy
func ClapGo_PluginDestroy(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	handle := cgo.Handle(plugin)
	p := handle.Value().(*GainPlugin)
	p.Destroy()
	handle.Delete()
}

//export ClapGo_PluginActivate
func ClapGo_PluginActivate(plugin unsafe.Pointer, sampleRate C.double, minFrames, maxFrames C.uint32_t) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	return C.bool(p.Activate(float64(sampleRate), uint32(minFrames), uint32(maxFrames)))
}

//export ClapGo_PluginDeactivate
func ClapGo_PluginDeactivate(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	p.Deactivate()
}

//export ClapGo_PluginStartProcessing
func ClapGo_PluginStartProcessing(plugin unsafe.Pointer) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	return C.bool(p.StartProcessing())
}

//export ClapGo_PluginStopProcessing
func ClapGo_PluginStopProcessing(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	p.StopProcessing()
}

//export ClapGo_PluginReset
func ClapGo_PluginReset(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	p.Reset()
}

//export ClapGo_PluginProcess
func ClapGo_PluginProcess(plugin unsafe.Pointer, process unsafe.Pointer) C.int32_t {
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
	
	// Call the actual Go process method
	result := p.Process(steadyTime, framesCount, audioIn, audioOut, eventHandler)
	
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
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	
	// Get current value - read atomically from our gain storage
	if uint32(paramID) == 0 {
		gainBits := atomic.LoadInt64(&p.gain)
		*value = C.double(floatFromBits(uint64(gainBits)))
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
		db := 20.0 * math.Log10(float64(value))
		text := fmt.Sprintf("%.1f dB", db)
		
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
	atomic.StoreInt64(&plugin.gain, int64(floatToBits(1.0)))
	
	// Register parameters using our new abstraction
	plugin.paramManager.RegisterParameter(api.CreateFloatParameter(0, "Gain", 0.0, 2.0, 1.0))
	plugin.paramManager.SetParameterValue(0, 1.0)
	
	return plugin
}

// Init initializes the plugin
func (p *GainPlugin) Init() bool {
	// Initialize track info helper
	if p.host != nil {
		p.trackInfo = api.NewHostTrackInfo(p.host)
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
	// Nothing to clean up
}

// Activate prepares the plugin for processing
func (p *GainPlugin) Activate(sampleRate float64, minFrames, maxFrames uint32) bool {
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
	atomic.StoreInt64(&p.gain, int64(floatToBits(1.0)))
}

// Process processes audio data using the new abstractions
func (p *GainPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events api.EventHandler) int {
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
	gain := floatFromBits(uint64(gainBits))
	
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
	
	// Process each event using our abstraction - NO MORE MANUAL C STRUCT PARSING!
	eventCount := events.GetInputEventCount()
	for i := uint32(0); i < eventCount; i++ {
		event := events.GetInputEvent(i)
		if event == nil {
			continue
		}
		
		// Handle parameter events using our abstraction
		switch event.Type {
		case api.EventTypeParamValue:
			if paramEvent, ok := event.Data.(api.ParamValueEvent); ok {
				p.handleParameterChange(paramEvent)
			}
		}
	}
}

// handleParameterChange processes a parameter change event
func (p *GainPlugin) handleParameterChange(paramEvent api.ParamValueEvent) {
	// Handle the parameter change based on its ID
	switch paramEvent.ParamID {
	case 0: // Gain parameter
		// Clamp value to valid range
		value := paramEvent.Value
		if value < 0.0 {
			value = 0.0
		}
		if value > 2.0 {
			value = 2.0
		}
		atomic.StoreInt64(&p.gain, int64(floatToBits(value)))
		
		// Update parameter manager
		p.paramManager.SetParameterValue(paramEvent.ParamID, value)
		
		// Log the parameter change
		if p.logger != nil {
			// Convert to dB for logging
			db := 20.0 * math.Log10(value)
			p.logger.Debug(fmt.Sprintf("Gain changed to %.2f dB", db))
		}
	}
}

// GetExtension gets a plugin extension
func (p *GainPlugin) GetExtension(id string) unsafe.Pointer {
	// Extensions are handled by the C bridge layer
	// The bridge provides params, state, and audio ports extensions
	return nil
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
	// Gain plugin has no latency
	return 0
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
			atomic.StoreInt64(&p.gain, int64(floatToBits(value)))
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
		
		// Load the preset state
		if gainValue, ok := state["gain"].(float64); ok {
			atomic.StoreInt64(&p.gain, int64(floatToBits(gainValue)))
			p.paramManager.SetParameterValue(0, gainValue)
			
			if p.logger != nil {
				p.logger.Info(fmt.Sprintf("Preset loaded: gain = %.2f", gainValue))
			}
			return true
		}
		
		if p.logger != nil {
			p.logger.Error("Invalid preset format: missing gain value")
		}
		return false
		
	case api.PresetLocationPlugin:
		// For this simple example, we don't have bundled presets
		if p.logger != nil {
			p.logger.Info("No bundled presets available for gain plugin")
		}
		return false
		
	default:
		return false
	}
}

// Helper functions for atomic float64 operations
func floatToBits(f float64) uint64 {
	return *(*uint64)(unsafe.Pointer(&f))
}

func floatFromBits(b uint64) float64 {
	return *(*float64)(unsafe.Pointer(&b))
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
func ClapGo_PluginLatencyGet(plugin unsafe.Pointer) C.uint32_t {
	if plugin == nil {
		return 0
	}
	p := cgo.Handle(plugin).Value().(*GainPlugin)
	return C.uint32_t(p.GetLatency())
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

func main() {
	// This is not called when used as a plugin,
	// but can be useful for testing
}