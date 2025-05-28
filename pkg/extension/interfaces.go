package extension

import "unsafe"

// StateProvider is an extension for plugins that can save and load state.
// It allows hosts to save and restore plugin state using CLAP's stream interface.
type StateProvider interface {
	// SaveState saves the plugin state to a stream.
	// The stream parameter is an unsafe.Pointer to a clap_ostream_t.
	// Returns true if the state was saved successfully.
	SaveState(stream unsafe.Pointer) bool

	// LoadState loads the plugin state from a stream.
	// The stream parameter is an unsafe.Pointer to a clap_istream_t.
	// Returns true if the state was loaded successfully.
	LoadState(stream unsafe.Pointer) bool
}

// StateContextProvider is an extension for plugins that support state context.
// It allows plugins to handle state save/load differently based on context
// (e.g., preset save, project save, or duplication).
type StateContextProvider interface {
	StateProvider
	
	// SaveStateWithContext saves plugin state with context information
	SaveStateWithContext(stream unsafe.Pointer, context uint32) bool
	
	// LoadStateWithContext loads plugin state with context information
	LoadStateWithContext(stream unsafe.Pointer, context uint32) bool
}

// GUIProvider is an extension for plugins that have a graphical user interface.
// It allows hosts to create and manage the plugin's GUI.
type GUIProvider interface {
	// IsAPISupported returns true if the given GUI API is supported.
	IsAPISupported(api string, isFloating bool) bool

	// GetPreferredAPI returns the preferred GUI API and whether it should be floating.
	GetPreferredAPI() (api string, isFloating bool)

	// Create creates the GUI.
	// Returns true if the GUI was created successfully.
	Create(api string, isFloating bool) bool

	// Destroy destroys the GUI.
	Destroy()

	// SetScale sets the GUI scaling factor.
	SetScale(scale float64) bool

	// GetSize returns the current GUI size.
	GetSize() (width, height uint32)

	// CanResize returns true if the GUI can be resized.
	CanResize() bool

	// AdjustSize adjusts the given size to fit GUI constraints.
	AdjustSize(width, height uint32) (newWidth, newHeight uint32)

	// SetSize sets the GUI size.
	SetSize(width, height uint32) bool

	// SetParent sets the parent window for the GUI.
	SetParent(parent unsafe.Pointer) bool

	// SetTransient sets the transient parent for the GUI.
	SetTransient(parent unsafe.Pointer) bool

	// Suggest hints for the host about GUI behavior.
	Suggest() []string

	// Show shows the GUI.
	Show() bool

	// Hide hides the GUI.
	Hide() bool
}

// LatencyProvider is an extension for plugins that introduce latency.
// It allows the host to query the plugin's processing latency.
type LatencyProvider interface {
	// GetLatency returns the processing latency in samples.
	GetLatency() uint32
}

// TailProvider is an extension for plugins that have a tail.
// The tail is the time after the input becomes silent that the plugin
// may still produce non-silent output.
type TailProvider interface {
	// GetTail returns the tail length in samples.
	// Returns math.MaxUint32 for infinite tail.
	GetTail() uint32
}

// RenderProvider is an extension for offline rendering.
// It allows the host to render audio offline faster than real-time.
type RenderProvider interface {
	// HasHardRealtimeRequirement returns true if the plugin has hard real-time requirements.
	HasHardRealtimeRequirement() bool

	// SetRenderMode sets the render mode (realtime or offline).
	SetRenderMode(mode int32) bool
}

// TimerSupportProvider is an extension for plugins that need timer support.
// It allows plugins to register timers with the host.
type TimerSupportProvider interface {
	// OnTimer is called when a timer fires.
	OnTimer(timerID uint32)
}

// LogProvider is an extension for plugins that want to log messages.
// It provides access to the host's logging facilities.
type LogProvider interface {
	// Log logs a message with the given severity.
	Log(severity int32, message string)
}

// VoiceInfoProvider is an extension for plugins that support polyphony.
// It allows the host to query voice information.
type VoiceInfoProvider interface {
	// GetVoiceInfo fills in voice information.
	GetVoiceInfo() VoiceInfo
}

// VoiceInfo contains information about plugin voices
type VoiceInfo struct {
	VoiceCount    uint32
	VoiceCapacity uint32
	Flags         uint64
}

// TrackInfoProvider is an extension for plugins that want track information.
// It allows plugins to receive information about the track they're on.
type TrackInfoProvider interface {
	// OnTrackInfoChanged is called when track information changes.
	OnTrackInfoChanged()
}

// EventRegistryProvider is an extension for plugins that want to register custom events.
type EventRegistryProvider interface {
	// QueryEvent queries information about a custom event type.
	QueryEvent(spaceID uint16, eventType uint16) bool
}

// ParamIndicationProvider is an extension for parameter automation indication.
type ParamIndicationProvider interface {
	// SetAutomation indicates parameter automation state.
	SetAutomation(paramID uint32, flags uint32, value float64)
	
	// SetMapping indicates parameter mapping state.
	SetMapping(paramID uint32, hasMapping bool, color uint32, label string, description string)
}

// PresetLoadProvider is an extension for preset loading.
type PresetLoadProvider interface {
	// LoadFromLocation loads a preset from the given location.
	LoadFromLocation(locationKind uint32, location, loadKey string) bool
}

// Common voice info flags
const (
	VoiceInfoSupportsOverlappingNotes = 1 << 0
)

// Common render modes
const (
	RenderModeRealtime = 0
	RenderModeOffline  = 1
)

// Common GUI APIs
const (
	GUIAPIWin32   = "win32"
	GUIAPIX11     = "x11"
	GUIAPICocoa   = "cocoa"
	GUIAPIWebView = "webview"
)