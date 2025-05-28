package plugin

import (
	"fmt"
	"unsafe"
	
	"github.com/justyntemme/clapgo/pkg/api"
	"github.com/justyntemme/clapgo/pkg/controls"
	"github.com/justyntemme/clapgo/pkg/event"
	hostpkg "github.com/justyntemme/clapgo/pkg/host"
	"github.com/justyntemme/clapgo/pkg/param"
	"github.com/justyntemme/clapgo/pkg/state"
	"github.com/justyntemme/clapgo/pkg/thread"
)

// PluginBase provides comprehensive base functionality for all plugins
type PluginBase struct {
	api.NoOpEventHandler // Embed no-op event handlers from api package
	
	// Core plugin state
	Host         unsafe.Pointer
	SampleRate   float64
	IsActivated  bool
	IsProcessing bool
	
	// Managers
	ParamManager *param.Manager
	StateManager *state.Manager
	Logger       *hostpkg.Logger
	
	// Extensions
	ThreadCheck  *thread.Checker
	TrackInfo    *api.HostTrackInfo
	
	// Plugin info
	Info Info
	
	// Diagnostics
	PoolDiagnostics event.Diagnostics
}

// NewPluginBase creates a new plugin base with common initialization
func NewPluginBase(info Info) *PluginBase {
	return &PluginBase{
		SampleRate:   44100.0,
		IsActivated:  false,
		IsProcessing: false,
		ParamManager: param.NewManager(),
		StateManager: state.NewManager(info.ID, info.Name, state.Version1),
		Info:         info,
	}
}

// InitWithHost initializes host-dependent features
func (b *PluginBase) InitWithHost(host unsafe.Pointer) {
	b.Host = host
	b.Logger = hostpkg.NewLogger(host)
	
	if host != nil {
		// Initialize thread checker
		b.ThreadCheck = thread.NewChecker(host)
		if b.ThreadCheck.IsAvailable() && b.Logger != nil {
			b.Logger.Info("Thread Check extension available - thread safety validation enabled")
		}
		
		// Initialize track info
		b.TrackInfo = api.NewHostTrackInfo(host)
	}
}

// CommonInit performs common initialization
func (b *PluginBase) CommonInit() bool {
	// Mark main thread for debug builds
	thread.SetMainThread()
	
	if b.Logger != nil {
		b.Logger.Info(fmt.Sprintf("[%s] Plugin initialized", b.Info.Name))
		b.Logger.Debug(fmt.Sprintf("[%s] Plugin ID: %s, Version: %s", b.Info.Name, b.Info.ID, b.Info.Version))
	}
	
	return true
}

// CommonDestroy performs common cleanup
func (b *PluginBase) CommonDestroy() {
	// Assert main thread
	thread.AssertMainThread("Plugin.Destroy")
	if b.ThreadCheck != nil {
		b.ThreadCheck.AssertMainThread("Plugin.Destroy")
	}
	
	if b.Logger != nil {
		b.Logger.Info(fmt.Sprintf("[%s] Plugin destroyed", b.Info.Name))
	}
}

// CommonActivate performs common activation
func (b *PluginBase) CommonActivate(sampleRate float64, minFrames, maxFrames uint32) bool {
	// Assert main thread
	thread.AssertMainThread("Plugin.Activate")
	if b.ThreadCheck != nil {
		b.ThreadCheck.AssertMainThread("Plugin.Activate")
	}
	
	b.SampleRate = sampleRate
	b.IsActivated = true
	
	if b.Logger != nil {
		b.Logger.Info(fmt.Sprintf("[%s] Plugin activated - Sample rate: %.0f Hz, Frame range: %d-%d", 
			b.Info.Name, sampleRate, minFrames, maxFrames))
	}
	
	return true
}

// CommonDeactivate performs common deactivation
func (b *PluginBase) CommonDeactivate() {
	// Assert main thread
	thread.AssertMainThread("Plugin.Deactivate")
	if b.ThreadCheck != nil {
		b.ThreadCheck.AssertMainThread("Plugin.Deactivate")
	}
	
	b.IsActivated = false
	
	if b.Logger != nil {
		b.Logger.Info(fmt.Sprintf("[%s] Plugin deactivated", b.Info.Name))
	}
}

// CommonStartProcessing prepares for audio processing
func (b *PluginBase) CommonStartProcessing() bool {
	if !b.IsActivated {
		if b.Logger != nil {
			b.Logger.Warning(fmt.Sprintf("[%s] Cannot start processing - plugin not activated", b.Info.Name))
		}
		return false
	}
	
	b.IsProcessing = true
	
	if b.Logger != nil {
		b.Logger.Info(fmt.Sprintf("[%s] Started audio processing", b.Info.Name))
	}
	
	return true
}

// CommonStopProcessing stops audio processing
func (b *PluginBase) CommonStopProcessing() {
	b.IsProcessing = false
	
	if b.Logger != nil {
		b.Logger.Info(fmt.Sprintf("[%s] Stopped audio processing", b.Info.Name))
	}
}

// CommonReset resets plugin state
func (b *PluginBase) CommonReset() {
	if b.Logger != nil {
		b.Logger.Debug("Plugin reset")
	}
}

// GetPluginInfo returns plugin information in API format
func (b *PluginBase) GetPluginInfo() api.PluginInfo {
	return api.PluginInfo{
		ID:          b.Info.ID,
		Name:        b.Info.Name,
		Vendor:      b.Info.Vendor,
		URL:         b.Info.URL,
		ManualURL:   b.Info.Manual,
		SupportURL:  b.Info.Support,
		Version:     b.Info.Version,
		Description: b.Info.Description,
		Features:    b.Info.Features,
	}
}

// GetPluginID returns the plugin ID
func (b *PluginBase) GetPluginID() string {
	return b.Info.ID
}

// GetLatency returns 0 by default (no latency)
func (b *PluginBase) GetLatency() uint32 {
	thread.AssertMainThread("PluginBase.GetLatency")
	return 0
}

// GetTail returns 0 by default (no tail)
func (b *PluginBase) GetTail() uint32 {
	return 0
}

// OnTimer does nothing by default
func (b *PluginBase) OnTimer(timerID uint64) {
	// Default implementation does nothing
}

// OnMainThread does nothing by default
func (b *PluginBase) OnMainThread() {
	// Default implementation does nothing
}

// LoadPresetFromLocation returns false by default (no preset loading)
func (b *PluginBase) LoadPresetFromLocation(locationKind uint32, location string, loadKey string) bool {
	return false
}

// GetParamInfo gets parameter info by index - can be used directly by plugins
func (b *PluginBase) GetParamInfo(index uint32, info unsafe.Pointer) bool {
	if info == nil {
		return false
	}
	
	paramInfo, err := b.ParamManager.GetInfoByIndex(index)
	if err != nil {
		return false
	}
	
	param.InfoToC(paramInfo, info)
	
	return true
}




// OnTrackInfoChanged provides default track info change handling with logging
func (b *PluginBase) OnTrackInfoChanged() {
	if b.TrackInfo == nil {
		return
	}
	
	// Get the new track information
	info, ok := b.TrackInfo.GetTrackInfo()
	if !ok {
		if b.Logger != nil {
			b.Logger.Warning("Failed to get track info")
		}
		return
	}
	
	// Log the track information
	if b.Logger != nil {
		b.Logger.Info("Track info changed:")
		if info.Flags&api.TrackInfoHasTrackName != 0 {
			b.Logger.Info(fmt.Sprintf("  Track name: %s", info.Name))
		}
		if info.Flags&api.TrackInfoHasTrackColor != 0 {
			b.Logger.Info(fmt.Sprintf("  Track color: R=%d G=%d B=%d A=%d", 
				info.Color.Red, info.Color.Green, info.Color.Blue, info.Color.Alpha))
		}
		if info.Flags&api.TrackInfoHasAudioChannel != 0 {
			b.Logger.Info(fmt.Sprintf("  Audio channels: %d, port type: %s", 
				info.AudioChannelCount, info.AudioPortType))
		}
		if info.Flags&api.TrackInfoIsForReturnTrack != 0 {
			b.Logger.Info("  This is a return track")
		}
		if info.Flags&api.TrackInfoIsForBus != 0 {
			b.Logger.Info("  This is a bus track")
		}
		if info.Flags&api.TrackInfoIsForMaster != 0 {
			b.Logger.Info("  This is the master track")
		}
	}
}

// SaveStateWithParams provides generic state saving with parameters
// This simplifies the common pattern of saving plugin state to a stream
func (b *PluginBase) SaveStateWithParams(stream unsafe.Pointer, params map[uint32]float64) bool {
	// Create output stream
	outStream := api.NewOutputStream(stream)
	
	// Convert parameter map to slice
	var parameters []state.Parameter
	for id, value := range params {
		// Get parameter info if available
		paramInfo, err := b.ParamManager.GetInfo(id)
		name := ""
		if err == nil {
			name = paramInfo.Name
		}
		
		parameters = append(parameters, state.Parameter{
			ID:    id,
			Value: value,
			Name:  name,
		})
	}
	
	// Create state
	pluginState := b.StateManager.CreateState(parameters, nil)
	
	// Serialize to JSON
	data, err := b.StateManager.SaveToJSON(pluginState)
	if err != nil {
		if b.Logger != nil {
			b.Logger.Error(fmt.Sprintf("Failed to serialize state: %v", err))
		}
		return false
	}
	
	// Write to stream
	if _, err := outStream.Write(data); err != nil {
		if b.Logger != nil {
			b.Logger.Error(fmt.Sprintf("Failed to write state: %v", err))
		}
		return false
	}
	
	if b.Logger != nil {
		b.Logger.Debug(fmt.Sprintf("State saved successfully (%d bytes)", len(data)))
	}
	
	return true
}

// LoadStateWithCallback provides generic state loading with a callback
// The callback is called for each parameter found in the saved state
func (b *PluginBase) LoadStateWithCallback(stream unsafe.Pointer, applyParam func(id uint32, value float64)) bool {
	// Create input stream
	inStream := api.NewInputStream(stream)
	
	// Read all data from stream
	const maxStateSize = 1024 * 1024 // 1MB max state size
	data := make([]byte, 0, 4096)
	buf := make([]byte, 4096)
	
	for {
		n, err := inStream.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
			if len(data) > maxStateSize {
				if b.Logger != nil {
					b.Logger.Error(fmt.Sprintf("State size exceeds maximum (%d bytes)", maxStateSize))
				}
				return false
			}
		}
		if err != nil || n == 0 {
			break
		}
	}
	
	if len(data) == 0 {
		if b.Logger != nil {
			b.Logger.Error("No state data found")
		}
		return false
	}
	
	// Parse state
	pluginState, err := b.StateManager.LoadFromJSON(data)
	if err != nil {
		if b.Logger != nil {
			b.Logger.Error(fmt.Sprintf("Failed to parse state: %v", err))
		}
		return false
	}
	
	// Apply parameters using callback
	for _, param := range pluginState.Parameters {
		applyParam(param.ID, param.Value)
	}
	
	if b.Logger != nil {
		b.Logger.Debug(fmt.Sprintf("State loaded successfully (%d parameters)", len(pluginState.Parameters)))
	}
	
	return true
}

// OnParamMappingSet provides default parameter mapping indication with logging
func (b *PluginBase) OnParamMappingSet(paramID uint32, hasMapping bool, color *api.Color, label string, description string) {
	// Check main thread (param indication is always on main thread)
	thread.AssertMainThread("PluginBase.OnParamMappingSet")
	
	// Log the mapping change
	if b.Logger != nil {
		if hasMapping {
			b.Logger.Info(fmt.Sprintf("Parameter %d mapped to %s: %s", paramID, label, description))
			if color != nil {
				b.Logger.Info(fmt.Sprintf("  Color: R=%d G=%d B=%d A=%d", color.Red, color.Green, color.Blue, color.Alpha))
			}
		} else {
			b.Logger.Info(fmt.Sprintf("Parameter %d mapping cleared", paramID))
		}
	}
}

// OnParamAutomationSet provides default parameter automation indication with logging
func (b *PluginBase) OnParamAutomationSet(paramID uint32, automationState uint32, color *api.Color) {
	// Check main thread (param indication is always on main thread)
	thread.AssertMainThread("PluginBase.OnParamAutomationSet")
	
	// Log the automation state change
	if b.Logger != nil {
		var stateStr string
		switch automationState {
		case param.IndicationAutomationNone:
			stateStr = "None"
		case param.IndicationAutomationPresent:
			stateStr = "Present"
		case param.IndicationAutomationPlaying:
			stateStr = "Playing"
		case param.IndicationAutomationRecording:
			stateStr = "Recording"
		case param.IndicationAutomationOverriding:
			stateStr = "Overriding"
		default:
			stateStr = "Unknown"
		}
		
		b.Logger.Info(fmt.Sprintf("Parameter %d automation state: %s", paramID, stateStr))
		if color != nil {
			b.Logger.Info(fmt.Sprintf("  Color: R=%d G=%d B=%d A=%d", color.Red, color.Green, color.Blue, color.Alpha))
		}
	}
}

// GetRemoteControlsPageCount returns 0 by default (no remote controls)
func (b *PluginBase) GetRemoteControlsPageCount() uint32 {
	return 0
}

// GetRemoteControlsPage returns nil by default
func (b *PluginBase) GetRemoteControlsPage(pageIndex uint32) (*controls.RemoteControlsPage, bool) {
	return nil, false
}

// GetExtension returns nil by default (no extensions)
// Override this to provide plugin-specific extensions
func (b *PluginBase) GetExtension(id string) unsafe.Pointer {
	// Most extensions are handled by the C bridge
	// Only override for Go-implemented extensions
	return nil
}

// SaveStateWithContext provides default implementation that logs context and calls SaveState
func (b *PluginBase) SaveStateWithContext(stream unsafe.Pointer, contextType uint32) bool {
	// Log the context type
	if b.Logger != nil {
		switch contextType {
		case api.StateContextForPreset:
			b.Logger.Info("Saving state for preset")
		case api.StateContextForDuplicate:
			b.Logger.Info("Saving state for duplicate")
		case api.StateContextForProject:
			b.Logger.Info("Saving state for project")
		default:
			b.Logger.Info(fmt.Sprintf("Saving state with unknown context: %d", contextType))
		}
	}
	
	// Default implementation calls SaveState
	// Override this method if you need context-specific saving
	// This is a fallback - plugins should implement SaveState
	return false
}

// LoadStateWithContext provides default implementation that logs context and calls LoadState
func (b *PluginBase) LoadStateWithContext(stream unsafe.Pointer, contextType uint32) bool {
	// Log the context type
	if b.Logger != nil {
		switch contextType {
		case api.StateContextForPreset:
			b.Logger.Info("Loading state for preset")
		case api.StateContextForDuplicate:
			b.Logger.Info("Loading state for duplicate")
		case api.StateContextForProject:
			b.Logger.Info("Loading state for project")
		default:
			b.Logger.Info(fmt.Sprintf("Loading state with unknown context: %d", contextType))
		}
	}
	
	// Default implementation calls LoadState
	// Override this method if you need context-specific loading
	// This is a fallback - plugins should implement LoadState
	return false
}

// Init delegates to CommonInit
func (b *PluginBase) Init() bool {
	return b.CommonInit()
}

// Destroy delegates to CommonDestroy
func (b *PluginBase) Destroy() {
	b.CommonDestroy()
}

// Activate delegates to CommonActivate
func (b *PluginBase) Activate(sampleRate float64, minFrames, maxFrames uint32) bool {
	return b.CommonActivate(sampleRate, minFrames, maxFrames)
}

// Deactivate delegates to CommonDeactivate
func (b *PluginBase) Deactivate() {
	b.CommonDeactivate()
}

// StopProcessing delegates to CommonStopProcessing
func (b *PluginBase) StopProcessing() {
	b.CommonStopProcessing()
}

// StartProcessing delegates to CommonStartProcessing
func (b *PluginBase) StartProcessing() bool {
	return b.CommonStartProcessing()
}

// Reset delegates to CommonReset
func (b *PluginBase) Reset() {
	b.CommonReset()
}