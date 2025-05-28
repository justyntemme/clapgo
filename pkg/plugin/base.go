package plugin

import (
	"fmt"
	"unsafe"
	
	"github.com/justyntemme/clapgo/pkg/controls"
	"github.com/justyntemme/clapgo/pkg/event"
	hostpkg "github.com/justyntemme/clapgo/pkg/host"
	"github.com/justyntemme/clapgo/pkg/param"
	"github.com/justyntemme/clapgo/pkg/state"
	"github.com/justyntemme/clapgo/pkg/thread"
)

// PluginBase provides comprehensive base functionality for all plugins
type PluginBase struct {
	event.NoOpHandler // Embed no-op event handlers from event package
	
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
	TrackInfo    *hostpkg.TrackInfoProvider
	
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
		b.TrackInfo = hostpkg.NewTrackInfoProvider(host)
	}
}

// CommonInit performs common initialization
func (b *PluginBase) CommonInit() error {
	// Mark main thread for debug builds
	thread.SetMainThread()
	
	if b.Logger != nil {
		b.Logger.Info(fmt.Sprintf("[%s] Plugin initialized", b.Info.Name))
		b.Logger.Debug(fmt.Sprintf("[%s] Plugin ID: %s, Version: %s", b.Info.Name, b.Info.ID, b.Info.Version))
	}
	
	return nil
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
func (b *PluginBase) CommonActivate(sampleRate float64, minFrames, maxFrames uint32) error {
	// Assert main thread
	thread.AssertMainThread("Plugin.Activate")
	if b.ThreadCheck != nil {
		b.ThreadCheck.AssertMainThread("Plugin.Activate")
	}
	
	if sampleRate <= 0 {
		return ErrInvalidSampleRate
	}
	
	if minFrames > maxFrames {
		return ErrInvalidFrameCount
	}
	
	b.SampleRate = sampleRate
	b.IsActivated = true
	
	if b.Logger != nil {
		b.Logger.Info(fmt.Sprintf("[%s] Plugin activated - Sample rate: %.0f Hz, Frame range: %d-%d", 
			b.Info.Name, sampleRate, minFrames, maxFrames))
	}
	
	return nil
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
func (b *PluginBase) CommonStartProcessing() error {
	if !b.IsActivated {
		if b.Logger != nil {
			b.Logger.Warning(fmt.Sprintf("[%s] Cannot start processing - plugin not activated", b.Info.Name))
		}
		return ErrNotActivated
	}
	
	b.IsProcessing = true
	
	if b.Logger != nil {
		b.Logger.Info(fmt.Sprintf("[%s] Started audio processing", b.Info.Name))
	}
	
	return nil
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

// GetPluginInfo returns plugin information
func (b *PluginBase) GetPluginInfo() Info {
	return b.Info
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

// LoadPresetFromLocation returns an error by default (no preset loading)
func (b *PluginBase) LoadPresetFromLocation(locationKind uint32, location string, loadKey string) error {
	return ErrNotImplemented
}

// GetParamInfo gets parameter info by index - can be used directly by plugins
func (b *PluginBase) GetParamInfo(index uint32, info unsafe.Pointer) error {
	if info == nil {
		return ErrInvalidParameter
	}
	
	paramInfo, err := b.ParamManager.GetInfoByIndex(index)
	if err != nil {
		return &ParameterError{Op: "get_info", ParamID: index, Err: err}
	}
	
	param.InfoToC(paramInfo, info)
	
	return nil
}




// OnTrackInfoChanged provides default track info change handling with logging
func (b *PluginBase) OnTrackInfoChanged() {
	if b.TrackInfo == nil {
		return
	}
	
	// Get the new track information
	info, ok := b.TrackInfo.Get()
	if !ok {
		if b.Logger != nil {
			b.Logger.Warning("Failed to get track info")
		}
		return
	}
	
	// Log the track information
	if b.Logger != nil {
		b.Logger.Info("Track info changed:")
		if info.Flags&hostpkg.TrackInfoHasTrackName != 0 {
			b.Logger.Info(fmt.Sprintf("  Track name: %s", info.Name))
		}
		if info.Flags&hostpkg.TrackInfoHasTrackColor != 0 {
			b.Logger.Info(fmt.Sprintf("  Track color: R=%d G=%d B=%d A=%d", 
				info.Color.Red, info.Color.Green, info.Color.Blue, info.Color.Alpha))
		}
		if info.Flags&hostpkg.TrackInfoHasAudioChannel != 0 {
			b.Logger.Info(fmt.Sprintf("  Audio channels: %d, port type: %s", 
				info.AudioChannelCount, info.AudioPortType))
		}
		if info.Flags&hostpkg.TrackInfoIsForReturnTrack != 0 {
			b.Logger.Info("  This is a return track")
		}
		if info.Flags&hostpkg.TrackInfoIsForBus != 0 {
			b.Logger.Info("  This is a bus track")
		}
		if info.Flags&hostpkg.TrackInfoIsForMaster != 0 {
			b.Logger.Info("  This is the master track")
		}
	}
}

// SaveStateWithParams provides generic state saving with parameters
// This simplifies the common pattern of saving plugin state to a stream
func (b *PluginBase) SaveStateWithParams(stream unsafe.Pointer, params map[uint32]float64) error {
	// Create output stream
	outStream := state.NewClapOutputStream(stream)
	
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
		return fmt.Errorf("failed to serialize state: %w", err)
	}
	
	// Write to stream
	if err := outStream.WriteBytes(data); err != nil {
		if b.Logger != nil {
			b.Logger.Error(fmt.Sprintf("Failed to write state: %v", err))
		}
		return fmt.Errorf("failed to write state: %w", err)
	}
	
	if b.Logger != nil {
		b.Logger.Debug(fmt.Sprintf("State saved successfully (%d bytes)", len(data)))
	}
	
	return nil
}

// LoadStateWithCallback provides generic state loading with a callback
// The callback is called for each parameter found in the saved state
func (b *PluginBase) LoadStateWithCallback(stream unsafe.Pointer, applyParam func(id uint32, value float64)) error {
	// Create input stream
	inStream := state.NewClapInputStream(stream)
	
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
				return ErrInvalidState
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
		return ErrInvalidState
	}
	
	// Parse state
	pluginState, err := b.StateManager.LoadFromJSON(data)
	if err != nil {
		if b.Logger != nil {
			b.Logger.Error(fmt.Sprintf("Failed to parse state: %v", err))
		}
		return fmt.Errorf("failed to parse state: %w", err)
	}
	
	// Apply parameters using callback
	for _, param := range pluginState.Parameters {
		applyParam(param.ID, param.Value)
	}
	
	if b.Logger != nil {
		b.Logger.Debug(fmt.Sprintf("State loaded successfully (%d parameters)", len(pluginState.Parameters)))
	}
	
	return nil
}

// OnParamMappingSet provides default parameter mapping indication with logging
func (b *PluginBase) OnParamMappingSet(paramID uint32, hasMapping bool, color *hostpkg.Color, label string, description string) {
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
func (b *PluginBase) OnParamAutomationSet(paramID uint32, automationState uint32, color *hostpkg.Color) {
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
func (b *PluginBase) SaveStateWithContext(stream unsafe.Pointer, contextType uint32) error {
	// Log the context type
	if b.Logger != nil {
		switch contextType {
		case state.ContextForPreset:
			b.Logger.Info("Saving state for preset")
		case state.ContextForDuplicate:
			b.Logger.Info("Saving state for duplicate")
		case state.ContextForProject:
			b.Logger.Info("Saving state for project")
		default:
			b.Logger.Info(fmt.Sprintf("Saving state with unknown context: %d", contextType))
		}
	}
	
	// Default implementation returns not implemented
	// Override this method if you need context-specific saving
	// This is a fallback - plugins should implement SaveState
	return ErrNotImplemented
}

// LoadStateWithContext provides default implementation that logs context and calls LoadState
func (b *PluginBase) LoadStateWithContext(stream unsafe.Pointer, contextType uint32) error {
	// Log the context type
	if b.Logger != nil {
		switch contextType {
		case state.ContextForPreset:
			b.Logger.Info("Loading state for preset")
		case state.ContextForDuplicate:
			b.Logger.Info("Loading state for duplicate")
		case state.ContextForProject:
			b.Logger.Info("Loading state for project")
		default:
			b.Logger.Info(fmt.Sprintf("Loading state with unknown context: %d", contextType))
		}
	}
	
	// Default implementation returns not implemented
	// Override this method if you need context-specific loading
	// This is a fallback - plugins should implement LoadState
	return ErrNotImplemented
}

// Init delegates to CommonInit
func (b *PluginBase) Init() error {
	return b.CommonInit()
}

// Destroy delegates to CommonDestroy
func (b *PluginBase) Destroy() {
	b.CommonDestroy()
}

// Activate delegates to CommonActivate
func (b *PluginBase) Activate(sampleRate float64, minFrames, maxFrames uint32) error {
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
func (b *PluginBase) StartProcessing() error {
	return b.CommonStartProcessing()
}

// Reset delegates to CommonReset
func (b *PluginBase) Reset() {
	b.CommonReset()
}

// Context-aware methods for PluginV2 interface

// InitWithContext performs initialization with context support
func (b *PluginBase) InitWithContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return b.CommonInit()
	}
}

// ActivateWithContext performs activation with context support
func (b *PluginBase) ActivateWithContext(ctx context.Context, sampleRate float64, minFrames, maxFrames uint32) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return b.CommonActivate(sampleRate, minFrames, maxFrames)
	}
}

// StartProcessingWithContext prepares for audio processing with context support
func (b *PluginBase) StartProcessingWithContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return b.CommonStartProcessing()
	}
}

// ProcessWithContext handles a single audio frame with context
func (b *PluginBase) ProcessWithContext(ctx context.Context, in, out [][]float32, steadyTime int64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Default implementation does nothing - plugins should override
		return ErrNotImplemented
	}
}

// ProcessBatch processes multiple frames with context support
func (b *PluginBase) ProcessBatch(ctx context.Context, frames []ProcessFrame) error {
	for i, frame := range frames {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := b.ProcessWithContext(ctx, frame.Input, frame.Output, frame.SteadyTime); err != nil {
				return &ProcessError{Frame: uint32(i), Err: err}
			}
			
			// Report progress if channel is available
			if progress, ok := GetProgress(ctx); ok {
				update := ProgressUpdate{
					Completed: int64(i + 1),
					Total:     int64(len(frames)),
					Percent:   float64(i+1) / float64(len(frames)),
					Message:   fmt.Sprintf("Processing frame %d/%d", i+1, len(frames)),
				}
				select {
				case progress <- update:
				default:
					// Don't block if channel is full
				}
			}
		}
	}
	return nil
}

// SaveStateWithContextV2 saves state with context and progress support
func (b *PluginBase) SaveStateWithContextV2(ctx context.Context, writer StateWriterV2) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Get parameter count for progress reporting
		paramCount := b.ParamManager.Count()
		
		// Write parameter count
		if err := writer.WriteUint32WithContext(ctx, paramCount); err != nil {
			return fmt.Errorf("failed to write parameter count: %w", err)
		}
		
		// Write each parameter
		for i := uint32(0); i < paramCount; i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				paramInfo, err := b.ParamManager.GetInfoByIndex(i)
				if err != nil {
					return fmt.Errorf("failed to get parameter %d: %w", i, err)
				}
				
				value, err := b.ParamManager.GetValue(paramInfo.ID)
				if err != nil {
					return fmt.Errorf("failed to get parameter value %d: %w", paramInfo.ID, err)
				}
				
				// Write parameter ID and value
				if err := writer.WriteUint32WithContext(ctx, paramInfo.ID); err != nil {
					return fmt.Errorf("failed to write parameter ID %d: %w", paramInfo.ID, err)
				}
				
				if err := writer.WriteFloat64WithContext(ctx, value); err != nil {
					return fmt.Errorf("failed to write parameter value %d: %w", paramInfo.ID, err)
				}
				
				// Report progress
				ReportProgress(ctx, ProgressUpdate{
					Completed: int64(i + 1),
					Total:     int64(paramCount),
					Percent:   float64(i+1) / float64(paramCount),
					Message:   fmt.Sprintf("Saving parameter %d/%d", i+1, paramCount),
				})
			}
		}
		
		return nil
	}
}

// LoadStateWithContextV2 loads state with context and progress support
func (b *PluginBase) LoadStateWithContextV2(ctx context.Context, reader StateReaderV2) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Read parameter count
		paramCount, err := reader.ReadUint32WithContext(ctx)
		if err != nil {
			return fmt.Errorf("failed to read parameter count: %w", err)
		}
		
		// Read each parameter
		for i := uint32(0); i < paramCount; i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				// Read parameter ID and value
				paramID, err := reader.ReadUint32WithContext(ctx)
				if err != nil {
					return fmt.Errorf("failed to read parameter ID %d: %w", i, err)
				}
				
				value, err := reader.ReadFloat64WithContext(ctx)
				if err != nil {
					return fmt.Errorf("failed to read parameter value %d: %w", paramID, err)
				}
				
				// Set parameter value
				if err := b.ParamManager.SetValue(paramID, value); err != nil {
					return fmt.Errorf("failed to set parameter %d: %w", paramID, err)
				}
				
				// Report progress
				ReportProgress(ctx, ProgressUpdate{
					Completed: int64(i + 1),
					Total:     int64(paramCount),
					Percent:   float64(i+1) / float64(paramCount),
					Message:   fmt.Sprintf("Loading parameter %d/%d", i+1, paramCount),
				})
			}
		}
		
		return nil
	}
}

// BeginParameterChanges starts a batch of parameter changes
func (b *PluginBase) BeginParameterChanges(ctx context.Context) (ParameterTransaction, error) {
	return NewParameterTransaction(ctx, b), nil
}

// GetParameterWithContext gets a parameter value with context
func (b *PluginBase) GetParameterWithContext(ctx context.Context, id uint32) (float64, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		value, err := b.ParamManager.GetValue(id)
		if err != nil {
			return 0, &ParameterError{Op: "get", ParamID: id, Err: err}
		}
		return value, nil
	}
}

// SetParameterWithContext sets a parameter value with context
func (b *PluginBase) SetParameterWithContext(ctx context.Context, id uint32, value float64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		err := b.ParamManager.SetValue(id, value)
		if err != nil {
			return &ParameterError{Op: "set", ParamID: id, Value: value, Err: err}
		}
		return nil
	}
}