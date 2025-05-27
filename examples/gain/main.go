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
)


// Global plugin instance
var gainPlugin *GainPlugin

func init() {
	fmt.Println("Initializing gain plugin")
	gainPlugin = NewGainPlugin()
	fmt.Printf("Gain plugin initialized: %s (%s)\n", gainPlugin.GetPluginInfo().Name, gainPlugin.GetPluginInfo().ID)
}

// Helper function to extract plugin from CGO handle
func getPlugin(plugin unsafe.Pointer) *GainPlugin {
	if plugin == nil {
		// Return a dummy plugin to avoid nil pointer dereference
		// The methods should handle nil gracefully
		return &GainPlugin{}
	}
	return cgo.Handle(plugin).Value().(*GainPlugin)
}

// Standardized export functions for manifest system

//export ClapGo_CreatePlugin
func ClapGo_CreatePlugin(host unsafe.Pointer, pluginID *C.char) unsafe.Pointer {
	if C.GoString(pluginID) == PluginID {
		return unsafe.Pointer(gainPlugin.CreateWithHost(host))
	}
	return nil
}

//export ClapGo_GetVersion
func ClapGo_GetVersion(major, minor, patch *C.uint32_t) C.bool {
	if major != nil { *major = C.uint32_t(1) }
	if minor != nil { *minor = C.uint32_t(0) }
	if patch != nil { *patch = C.uint32_t(0) }
	return C.bool(true)
}

//export ClapGo_GetPluginID
func ClapGo_GetPluginID(pluginID *C.char) *C.char {
	return C.CString(PluginID)
}

//export ClapGo_GetPluginName
func ClapGo_GetPluginName(pluginID *C.char) *C.char {
	return C.CString(PluginName)
}

//export ClapGo_GetPluginVendor
func ClapGo_GetPluginVendor(pluginID *C.char) *C.char {
	return C.CString(PluginVendor)
}

//export ClapGo_GetPluginVersion
func ClapGo_GetPluginVersion(pluginID *C.char) *C.char {
	return C.CString(PluginVersion)
}

//export ClapGo_GetPluginDescription
func ClapGo_GetPluginDescription(pluginID *C.char) *C.char {
	return C.CString(PluginDescription)
}

//export ClapGo_PluginInit
func ClapGo_PluginInit(plugin unsafe.Pointer) C.bool {
	return C.bool(getPlugin(plugin).InitWithLogging())
}

//export ClapGo_PluginDestroy
func ClapGo_PluginDestroy(plugin unsafe.Pointer) {
	if p := getPlugin(plugin); p != nil {
		p.DestroyWithHandle(plugin)
	}
}

//export ClapGo_PluginActivate
func ClapGo_PluginActivate(plugin unsafe.Pointer, sampleRate C.double, minFrames, maxFrames C.uint32_t) C.bool {
	return C.bool(getPlugin(plugin).ActivateWithLogging(float64(sampleRate), uint32(minFrames), uint32(maxFrames)))
}

//export ClapGo_PluginDeactivate
func ClapGo_PluginDeactivate(plugin unsafe.Pointer) {
	getPlugin(plugin).DeactivateWithLogging()
}

//export ClapGo_PluginStartProcessing
func ClapGo_PluginStartProcessing(plugin unsafe.Pointer) C.bool {
	api.DebugMarkAudioThread()
	defer api.DebugUnmarkAudioThread()
	return C.bool(getPlugin(plugin).StartProcessingWithLogging())
}

//export ClapGo_PluginStopProcessing
func ClapGo_PluginStopProcessing(plugin unsafe.Pointer) {
	api.DebugMarkAudioThread()
	defer api.DebugUnmarkAudioThread()
	getPlugin(plugin).StopProcessingWithLogging()
}

//export ClapGo_PluginReset
func ClapGo_PluginReset(plugin unsafe.Pointer) {
	getPlugin(plugin).ResetWithLogging()
}

//export ClapGo_PluginProcess
func ClapGo_PluginProcess(plugin unsafe.Pointer, process unsafe.Pointer) C.int32_t {
	api.DebugMarkAudioThread()
	defer api.DebugUnmarkAudioThread()
	return C.int32_t(getPlugin(plugin).ProcessWithHandle(process))
}

//export ClapGo_PluginGetExtension
func ClapGo_PluginGetExtension(plugin unsafe.Pointer, id *C.char) unsafe.Pointer {
	return getPlugin(plugin).GetExtension(C.GoString(id))
}

//export ClapGo_PluginOnMainThread
func ClapGo_PluginOnMainThread(plugin unsafe.Pointer) {
	getPlugin(plugin).OnMainThread()
}

//export ClapGo_PluginParamsCount
func ClapGo_PluginParamsCount(plugin unsafe.Pointer) C.uint32_t {
	return C.uint32_t(getPlugin(plugin).ParamManager.Count())
}

//export ClapGo_PluginParamsGetInfo
func ClapGo_PluginParamsGetInfo(plugin unsafe.Pointer, index C.uint32_t, info unsafe.Pointer) C.bool {
	return C.bool(getPlugin(plugin).GetParamInfo(uint32(index), info))
}

//export ClapGo_PluginParamsGetValue
func ClapGo_PluginParamsGetValue(plugin unsafe.Pointer, paramID C.uint32_t, value *C.double) C.bool {
	return C.bool(getPlugin(plugin).GetParamValue(uint32(paramID), value))
}

//export ClapGo_PluginParamsValueToText
func ClapGo_PluginParamsValueToText(plugin unsafe.Pointer, paramID C.uint32_t, value C.double, buffer *C.char, size C.uint32_t) C.bool {
	return C.bool(getPlugin(plugin).ParamValueToText(uint32(paramID), float64(value), buffer, uint32(size)))
}

//export ClapGo_PluginParamsTextToValue
func ClapGo_PluginParamsTextToValue(plugin unsafe.Pointer, paramID C.uint32_t, text *C.char, value *C.double) C.bool {
	return C.bool(getPlugin(plugin).ParamTextToValue(uint32(paramID), C.GoString(text), value))
}

//export ClapGo_PluginParamsFlush
func ClapGo_PluginParamsFlush(plugin unsafe.Pointer, inEvents unsafe.Pointer, outEvents unsafe.Pointer) {
	getPlugin(plugin).ParamsFlush(inEvents, outEvents)
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
	
	// Initialize context menu provider for common menu functionality
	p.contextMenuProvider = api.NewDefaultContextMenuProvider(
		nil, // We'll handle param manager ourselves
		PluginName,
		PluginVersion,
		nil, // Host will be set later
	)
	
	return p
}

// CreateWithHost creates plugin handle after initializing with host
func (p *GainPlugin) CreateWithHost(host unsafe.Pointer) cgo.Handle {
	// Initialize host-dependent features
	p.PluginBase.InitWithHost(host)
	
	// Log plugin creation
	if p.Logger != nil {
		p.Logger.Info("Creating gain plugin instance")
	}
	
	// Create a CGO handle to safely pass the Go object to C
	handle := cgo.NewHandle(p)
	
	// Log handle creation
	if p.Logger != nil {
		p.Logger.Debug(fmt.Sprintf("[CreateWithHost] Created handle: %v", handle))
	}
	
	// Register as audio ports provider
	api.RegisterAudioPortsProvider(unsafe.Pointer(handle), p)
	
	return handle
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

// No-op event handlers are now provided by PluginBase which embeds event.NoOpHandler

// Helper methods for minimal export functions

// InitWithLogging wraps Init with logging
func (p *GainPlugin) InitWithLogging() bool {
	if p.Logger != nil {
		p.Logger.Debug("[InitWithLogging] Starting plugin initialization")
	}
	
	result := p.Init()
	
	if p.Logger != nil {
		if result {
			p.Logger.Info("[InitWithLogging] Plugin initialization successful")
		} else {
			p.Logger.Error("[InitWithLogging] Plugin initialization failed")
		}
	}
	
	return result
}

// DestroyWithHandle wraps Destroy with handle cleanup
func (p *GainPlugin) DestroyWithHandle(plugin unsafe.Pointer) {
	if p.Logger != nil {
		p.Logger.Debug("[DestroyWithHandle] Destroying plugin instance")
	}
	
	p.Destroy()
	
	// Unregister from audio ports provider
	api.UnregisterAudioPortsProvider(plugin)
	
	if p.Logger != nil {
		p.Logger.Info("[DestroyWithHandle] Plugin instance destroyed successfully")
	}
	
	// Delete the handle to free the Go object
	cgo.Handle(plugin).Delete()
}

// ActivateWithLogging wraps Activate with logging
func (p *GainPlugin) ActivateWithLogging(sampleRate float64, minFrames, maxFrames uint32) bool {
	if p.Logger != nil {
		p.Logger.Debug(fmt.Sprintf("[ActivateWithLogging] Activating plugin - SR: %.0f, frames: %d-%d", 
			sampleRate, minFrames, maxFrames))
	}
	
	result := p.Activate(sampleRate, minFrames, maxFrames)
	
	if p.Logger != nil {
		if result {
			p.Logger.Info("[ActivateWithLogging] Plugin activation successful")
		} else {
			p.Logger.Error("[ActivateWithLogging] Plugin activation failed")
		}
	}
	
	return result
}

// DeactivateWithLogging wraps Deactivate with logging
func (p *GainPlugin) DeactivateWithLogging() {
	if p.Logger != nil {
		p.Logger.Debug("[DeactivateWithLogging] Deactivating plugin")
	}
	
	p.Deactivate()
	
	if p.Logger != nil {
		p.Logger.Info("[DeactivateWithLogging] Plugin deactivation successful")
	}
}

// StartProcessingWithLogging wraps StartProcessing with logging
func (p *GainPlugin) StartProcessingWithLogging() bool {
	if p.Logger != nil {
		p.Logger.Debug("[StartProcessingWithLogging] Starting audio processing")
	}
	
	result := p.StartProcessing()
	
	if p.Logger != nil {
		if result {
			p.Logger.Info("[StartProcessingWithLogging] Audio processing started successfully")
		} else {
			p.Logger.Error("[StartProcessingWithLogging] Failed to start audio processing")
		}
	}
	
	return result
}

// StopProcessingWithLogging wraps StopProcessing with logging
func (p *GainPlugin) StopProcessingWithLogging() {
	if p.Logger != nil {
		p.Logger.Debug("[StopProcessingWithLogging] Stopping audio processing")
	}
	
	p.StopProcessing()
	
	if p.Logger != nil {
		p.Logger.Info("[StopProcessingWithLogging] Audio processing stopped successfully")
	}
}

// ResetWithLogging wraps Reset with logging
func (p *GainPlugin) ResetWithLogging() {
	if p.Logger != nil {
		p.Logger.Debug("[ResetWithLogging] Resetting plugin state")
	}
	
	p.Reset()
	
	if p.Logger != nil {
		p.Logger.Info("[ResetWithLogging] Plugin reset successful")
	}
}

// ProcessWithHandle wraps Process with C handle conversion
func (p *GainPlugin) ProcessWithHandle(process unsafe.Pointer) int {
	if process == nil {
		return api.ProcessError
	}
	
	// Convert the C clap_process_t to Go parameters
	cProcess := (*C.clap_process_t)(process)
	
	// Extract steady time and frame count
	steadyTime := int64(cProcess.steady_time)
	framesCount := uint32(cProcess.frames_count)
	
	// Convert audio buffers using our abstraction
	audioIn := api.ConvertFromCBuffers(unsafe.Pointer(cProcess.audio_inputs), uint32(cProcess.audio_inputs_count), framesCount)
	audioOut := api.ConvertFromCBuffers(unsafe.Pointer(cProcess.audio_outputs), uint32(cProcess.audio_outputs_count), framesCount)
	
	// Create event handler using the new abstraction
	eventHandler := api.NewEventProcessor(
		unsafe.Pointer(cProcess.in_events),
		unsafe.Pointer(cProcess.out_events),
	)
	
	// Setup event pool logging
	api.SetupPoolLogging(eventHandler, p.Logger)
	
	// Call the actual Go process method
	result := p.Process(steadyTime, framesCount, audioIn, audioOut, eventHandler)
	
	// Log event pool diagnostics periodically (every 1000 calls)
	p.PoolDiagnostics.LogPoolDiagnostics(eventHandler, 1000)
	
	return result
}

// GetParamInfo gets parameter info by index
func (p *GainPlugin) GetParamInfo(index uint32, info unsafe.Pointer) bool {
	if info == nil {
		return false
	}
	
	// Get parameter info from manager
	paramInfo, err := p.ParamManager.GetInfoByIndex(index)
	if err != nil {
		return false
	}
	
	// Convert to C struct using helper
	param.InfoToC(paramInfo, info)
	
	return true
}

// GetParamValue gets parameter value by ID
func (p *GainPlugin) GetParamValue(paramID uint32, value *C.double) bool {
	if value == nil {
		return false
	}
	
	// Get current value - read atomically from our gain storage
	if paramID == ParamGain {
		*value = C.double(p.gain.Load())
		return true
	}
	
	return false
}

// ParamValueToText converts parameter value to text
func (p *GainPlugin) ParamValueToText(paramID uint32, value float64, buffer *C.char, size uint32) bool {
	if buffer == nil || size == 0 {
		return false
	}
	
	// For gain parameter, format as dB
	if paramID == ParamGain {
		text := param.FormatValue(value, param.FormatDecibel)
		
		// Copy to C buffer manually
		bytes := []byte(text)
		if len(bytes) >= int(size) {
			bytes = bytes[:size-1]
		}
		for i, b := range bytes {
			*(*C.char)(unsafe.Add(unsafe.Pointer(buffer), i)) = C.char(b)
		}
		*(*C.char)(unsafe.Add(unsafe.Pointer(buffer), len(bytes))) = 0
		
		return true
	}
	
	return false
}

// ParamTextToValue converts parameter text to value
func (p *GainPlugin) ParamTextToValue(paramID uint32, text string, value *C.double) bool {
	if value == nil {
		return false
	}
	
	// For gain parameter, use the ParameterValueParser
	if paramID == ParamGain {
		parser := param.NewParser(param.FormatDecibel)
		if parsedValue, err := parser.ParseValue(text); err == nil {
			// Clamp to valid range for gain parameter
			clamped := param.ClampValue(parsedValue, 0.0, 2.0)
			*value = C.double(clamped)
			return true
		}
	}
	
	return false
}

// ParamsFlush flushes parameter events
func (p *GainPlugin) ParamsFlush(inEvents, outEvents unsafe.Pointer) {
	// Process events using our abstraction
	if inEvents != nil {
		eventHandler := api.NewEventProcessor(inEvents, outEvents)
		p.processEvents(eventHandler, 0)
	}
}

// PopulateContextMenuWithTarget wraps PopulateContextMenu with target creation
func (p *GainPlugin) PopulateContextMenuWithTarget(targetKind uint32, targetID uint64, builder unsafe.Pointer) bool {
	if builder == nil {
		return false
	}
	
	// Create target
	var target *api.ContextMenuTarget
	if targetKind != api.ContextMenuTargetKindGlobal {
		target = &api.ContextMenuTarget{
			Kind: targetKind,
			ID:   targetID,
		}
	}
	
	// Create builder wrapper
	menuBuilder := api.NewContextMenuBuilder(builder)
	
	return p.PopulateContextMenu(target, menuBuilder)
}

// PerformContextMenuActionWithTarget wraps PerformContextMenuAction with target creation
func (p *GainPlugin) PerformContextMenuActionWithTarget(targetKind uint32, targetID uint64, actionID uint64) bool {
	// Create target
	var target *api.ContextMenuTarget
	if targetKind != api.ContextMenuTargetKindGlobal {
		target = &api.ContextMenuTarget{
			Kind: targetKind,
			ID:   targetID,
		}
	}
	
	return p.PerformContextMenuAction(target, actionID)
}

// GetRemoteControlsPageToC gets remote controls page and converts to C
func (p *GainPlugin) GetRemoteControlsPageToC(pageIndex uint32, cPage unsafe.Pointer) bool {
	if cPage == nil {
		return false
	}
	
	page, ok := p.GetRemoteControlsPage(pageIndex)
	if !ok {
		return false
	}
	
	// Convert Go page to C structure
	api.RemoteControlsPageToC(page, cPage)
	return true
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

// OnMainThread is provided by PluginBase

// GetPluginID is provided by PluginBase

// GetLatency is provided by PluginBase (returns 0)

// GetTail is provided by PluginBase (returns 0)

// OnTimer is provided by PluginBase

// OnTrackInfoChanged is provided by PluginBase with default logging
// Override this method if you need custom track info handling

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

// Param Indication Extension Methods are provided by PluginBase
// Override OnParamMappingSet and OnParamAutomationSet if you need custom handling

// SaveState saves the plugin state to a stream
func (p *GainPlugin) SaveState(stream unsafe.Pointer) bool {
	return p.SaveStateWithParams(stream, map[uint32]float64{
		ParamGain: p.gain.Load(),
	})
}

// LoadState loads the plugin state from a stream
func (p *GainPlugin) LoadState(stream unsafe.Pointer) bool {
	return p.LoadStateWithCallback(stream, func(id uint32, value float64) {
		if id == ParamGain {
			p.gain.Store(value)
			p.ParamManager.SetValue(id, value)
		}
	})
}

// SaveStateWithContext saves state with context logging
func (p *GainPlugin) SaveStateWithContext(stream unsafe.Pointer, contextType uint32) bool {
	// Use base implementation for logging
	p.PluginBase.SaveStateWithContext(stream, contextType)
	// Then do actual save
	return p.SaveState(stream)
}

// LoadStateWithContext loads state with context logging
func (p *GainPlugin) LoadStateWithContext(stream unsafe.Pointer, contextType uint32) bool {
	// Use base implementation for logging
	p.PluginBase.LoadStateWithContext(stream, contextType)
	// Then do actual load
	return p.LoadState(stream)
}

// LoadPresetFromLocation is provided by PluginBase (returns false)

// Helper functions for atomic float64 operations

// Audio port configuration is provided by embedded StereoPortProvider and SurroundSupport

//export ClapGo_PluginStateSave
func ClapGo_PluginStateSave(plugin unsafe.Pointer, stream unsafe.Pointer) C.bool {
	return C.bool(getPlugin(plugin).SaveState(stream))
}

//export ClapGo_PluginStateLoad
func ClapGo_PluginStateLoad(plugin unsafe.Pointer, stream unsafe.Pointer) C.bool {
	return C.bool(getPlugin(plugin).LoadState(stream))
}

// State Context Extension Exports

//export ClapGo_PluginStateSaveWithContext
func ClapGo_PluginStateSaveWithContext(plugin unsafe.Pointer, stream unsafe.Pointer, contextType C.uint32_t) C.bool {
	return C.bool(getPlugin(plugin).SaveStateWithContext(stream, uint32(contextType)))
}

//export ClapGo_PluginStateLoadWithContext
func ClapGo_PluginStateLoadWithContext(plugin unsafe.Pointer, stream unsafe.Pointer, contextType C.uint32_t) C.bool {
	return C.bool(getPlugin(plugin).LoadStateWithContext(stream, uint32(contextType)))
}

// Preset Load Extension Exports

//export ClapGo_PluginPresetLoadFromLocation
func ClapGo_PluginPresetLoadFromLocation(plugin unsafe.Pointer, locationKind C.uint32_t, location *C.char, loadKey *C.char) C.bool {
	return C.bool(getPlugin(plugin).LoadPresetFromLocation(uint32(locationKind), C.GoString(location), C.GoString(loadKey)))
}

// Phase 3 Extension Exports

//export ClapGo_PluginLatencyGet
func ClapGo_PluginLatencyGet(plugin unsafe.Pointer) uint32 {
	return getPlugin(plugin).GetLatency()
}

//export ClapGo_PluginTailGet
func ClapGo_PluginTailGet(plugin unsafe.Pointer) C.uint32_t {
	return C.uint32_t(getPlugin(plugin).GetTail())
}

//export ClapGo_PluginOnTimer
func ClapGo_PluginOnTimer(plugin unsafe.Pointer, timerID C.uint64_t) {
	getPlugin(plugin).OnTimer(uint64(timerID))
}

// Phase 7 Extension Exports

//export ClapGo_PluginTrackInfoChanged
func ClapGo_PluginTrackInfoChanged(plugin unsafe.Pointer) {
	getPlugin(plugin).OnTrackInfoChanged()
}

// Context Menu Extension Exports

//export ClapGo_PluginContextMenuPopulate
func ClapGo_PluginContextMenuPopulate(plugin unsafe.Pointer, targetKind C.uint32_t, targetID C.uint64_t, builder unsafe.Pointer) C.bool {
	return C.bool(getPlugin(plugin).PopulateContextMenuWithTarget(uint32(targetKind), uint64(targetID), builder))
}

//export ClapGo_PluginContextMenuPerform
func ClapGo_PluginContextMenuPerform(plugin unsafe.Pointer, targetKind C.uint32_t, targetID C.uint64_t, actionID C.uint64_t) C.bool {
	return C.bool(getPlugin(plugin).PerformContextMenuActionWithTarget(uint32(targetKind), uint64(targetID), uint64(actionID)))
}

// Remote Controls Extension Exports

//export ClapGo_PluginRemoteControlsCount
func ClapGo_PluginRemoteControlsCount(plugin unsafe.Pointer) C.uint32_t {
	return C.uint32_t(getPlugin(plugin).GetRemoteControlsPageCount())
}

//export ClapGo_PluginRemoteControlsGet
func ClapGo_PluginRemoteControlsGet(plugin unsafe.Pointer, pageIndex C.uint32_t, cPage unsafe.Pointer) C.bool {
	return C.bool(getPlugin(plugin).GetRemoteControlsPageToC(uint32(pageIndex), cPage))
}

// Param Indication Extension Exports

//export ClapGo_PluginParamIndicationSetMapping
func ClapGo_PluginParamIndicationSetMapping(plugin unsafe.Pointer, paramID C.uint64_t, hasMapping C.bool, color unsafe.Pointer, label *C.char, description *C.char) {
	getPlugin(plugin).OnParamMappingSet(uint32(paramID), bool(hasMapping), api.ColorFromC(color), C.GoString(label), C.GoString(description))
}

//export ClapGo_PluginParamIndicationSetAutomation
func ClapGo_PluginParamIndicationSetAutomation(plugin unsafe.Pointer, paramID C.uint64_t, automationState C.uint32_t, color unsafe.Pointer) {
	getPlugin(plugin).OnParamAutomationSet(uint32(paramID), uint32(automationState), api.ColorFromC(color))
}

func main() {
	// This is not called when used as a plugin,
	// but can be useful for testing
}