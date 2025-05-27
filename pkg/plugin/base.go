package plugin

import (
	"unsafe"
	
	"github.com/justyntemme/clapgo/pkg/api"
	hostpkg "github.com/justyntemme/clapgo/pkg/host"
	"github.com/justyntemme/clapgo/pkg/param"
	"github.com/justyntemme/clapgo/pkg/state"
)

// PluginBase provides comprehensive base functionality for all plugins
type PluginBase struct {
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
	ThreadCheck  *api.ThreadChecker
	TrackInfo    *api.HostTrackInfo
	
	// Plugin info
	Info Info
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
		b.ThreadCheck = api.NewThreadChecker(host)
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
	api.DebugSetMainThread()
	
	if b.Logger != nil {
		b.Logger.Debug("Plugin initialized")
	}
	
	return true
}

// CommonDestroy performs common cleanup
func (b *PluginBase) CommonDestroy() {
	// Assert main thread
	api.DebugAssertMainThread("Plugin.Destroy")
	if b.ThreadCheck != nil {
		b.ThreadCheck.AssertMainThread("Plugin.Destroy")
	}
}

// CommonActivate performs common activation
func (b *PluginBase) CommonActivate(sampleRate float64, minFrames, maxFrames uint32) bool {
	// Assert main thread
	api.DebugAssertMainThread("Plugin.Activate")
	if b.ThreadCheck != nil {
		b.ThreadCheck.AssertMainThread("Plugin.Activate")
	}
	
	b.SampleRate = sampleRate
	b.IsActivated = true
	
	if b.Logger != nil {
		b.Logger.Info("Plugin activated")
	}
	
	return true
}

// CommonDeactivate performs common deactivation
func (b *PluginBase) CommonDeactivate() {
	// Assert main thread
	api.DebugAssertMainThread("Plugin.Deactivate")
	if b.ThreadCheck != nil {
		b.ThreadCheck.AssertMainThread("Plugin.Deactivate")
	}
	
	b.IsActivated = false
	
	if b.Logger != nil {
		b.Logger.Info("Plugin deactivated")
	}
}

// CommonStartProcessing prepares for audio processing
func (b *PluginBase) CommonStartProcessing() bool {
	if !b.IsActivated {
		return false
	}
	
	b.IsProcessing = true
	
	if b.Logger != nil {
		b.Logger.Debug("Started processing")
	}
	
	return true
}

// CommonStopProcessing stops audio processing
func (b *PluginBase) CommonStopProcessing() {
	b.IsProcessing = false
	
	if b.Logger != nil {
		b.Logger.Debug("Stopped processing")
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