package api

// Render modes for plugin processing
const (
	// RenderRealtime is the default setting for realtime processing.
	// The plugin should use efficient algorithms suitable for low-latency processing.
	RenderRealtime = 0

	// RenderOffline is for processing without realtime pressure.
	// The plugin may use more expensive algorithms for higher sound quality.
	RenderOffline = 1
)

// RenderProvider is an extension for plugins that can adapt their processing
// based on whether they're running in realtime or offline mode.
// If your plugin's rendering code doesn't change based on this information,
// you don't need to implement this extension.
type RenderProvider interface {
	// HasHardRealtimeRequirement returns true if the plugin has a hard requirement
	// to process in real-time. This is especially useful for plugins acting as
	// a proxy to hardware devices.
	// [main-thread]
	HasHardRealtimeRequirement() bool

	// SetRenderMode sets the rendering mode (realtime or offline).
	// Returns true if the mode could be applied.
	// [main-thread]
	SetRenderMode(mode int32) bool
}

// RenderModeHelper provides convenient methods for handling render modes
type RenderModeHelper struct {
	currentMode        int32
	supportsOffline    bool
	hardwareConnected  bool
}

// NewRenderModeHelper creates a new render mode helper
func NewRenderModeHelper(supportsOffline bool, hardwareConnected bool) *RenderModeHelper {
	return &RenderModeHelper{
		currentMode:       RenderRealtime,
		supportsOffline:   supportsOffline,
		hardwareConnected: hardwareConnected,
	}
}

// HasHardRealtimeRequirement returns true if connected to hardware
func (h *RenderModeHelper) HasHardRealtimeRequirement() bool {
	return h.hardwareConnected
}

// SetRenderMode sets the render mode if supported
func (h *RenderModeHelper) SetRenderMode(mode int32) bool {
	switch mode {
	case RenderRealtime:
		h.currentMode = RenderRealtime
		return true
		
	case RenderOffline:
		if h.hardwareConnected {
			// Can't do offline rendering when connected to hardware
			return false
		}
		if h.supportsOffline {
			h.currentMode = RenderOffline
			return true
		}
		return false
		
	default:
		return false
	}
}

// GetCurrentMode returns the current render mode
func (h *RenderModeHelper) GetCurrentMode() int32 {
	return h.currentMode
}

// IsOffline returns true if currently in offline mode
func (h *RenderModeHelper) IsOffline() bool {
	return h.currentMode == RenderOffline
}

// IsRealtime returns true if currently in realtime mode
func (h *RenderModeHelper) IsRealtime() bool {
	return h.currentMode == RenderRealtime
}