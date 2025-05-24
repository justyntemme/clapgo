package api

import "unsafe"

// AudioPortsProvider is an extension for plugins that have audio ports.
// It allows hosts to query information about the plugin's audio ports.
type AudioPortsProvider interface {
	// GetAudioPortCount returns the number of audio ports.
	GetAudioPortCount(isInput bool) uint32

	// GetAudioPortInfo returns information about an audio port.
	GetAudioPortInfo(index uint32, isInput bool) AudioPortInfo
}

// AudioPortInfo contains information about an audio port.
type AudioPortInfo struct {
	// ID is a unique identifier for the port
	ID uint32

	// Name is a human-readable name for the port
	Name string

	// ChannelCount is the number of channels in this port
	ChannelCount uint32

	// Flags contains additional port flags
	Flags uint32

	// PortType describes the port type (e.g., "mono", "stereo")
	PortType string

	// InPlacePair is the ID of the in-place pair port or INVALID_ID if none
	InPlacePair uint32
}

// Audio port flags and port types are defined in constants.go

// ParamsProvider is an extension for plugins that have parameters.
// It allows hosts to query and manipulate plugin parameters.
type ParamsProvider interface {
	// GetParamCount returns the number of parameters.
	GetParamCount() uint32

	// GetParamInfo returns information about a parameter.
	GetParamInfo(paramID uint32) ParamInfo

	// GetParamValue returns the current value of a parameter.
	GetParamValue(paramID uint32) float64

	// SetParamValue sets the value of a parameter.
	SetParamValue(paramID uint32, value float64)

	// FlushParams writes all parameter changes to the DSP.
	FlushParams()
}

// ParamInfo contains information about a parameter.
type ParamInfo struct {
	// ID is a unique identifier for the parameter
	ID uint32

	// Name is a human-readable name for the parameter
	Name string

	// Module is an optional module path for hierarchical parameters
	Module string

	// MinValue is the minimum value of the parameter
	MinValue float64

	// MaxValue is the maximum value of the parameter
	MaxValue float64

	// DefaultValue is the default value of the parameter
	DefaultValue float64

	// Flags contains additional parameter flags
	Flags uint32
}

// Parameter flags are defined in constants.go

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
	// SaveStateWithContext saves the plugin state to a stream with context.
	// contextType indicates the context for saving (preset, duplicate, project).
	// Returns true if the state was saved successfully.
	SaveStateWithContext(stream unsafe.Pointer, contextType uint32) bool

	// LoadStateWithContext loads the plugin state from a stream with context.
	// contextType indicates the context for loading (preset, duplicate, project).
	// Returns true if the state was loaded successfully.
	LoadStateWithContext(stream unsafe.Pointer, contextType uint32) bool
}

// State context types
const (
	StateContextForPreset    = 1 // Suitable for storing and loading a state as a preset
	StateContextForDuplicate = 2 // Suitable for duplicating a plugin instance
	StateContextForProject   = 3 // Suitable for storing and loading a state within a project/song
)

// GUIProvider is an extension for plugins with a graphical user interface.
// It allows hosts to create and manage plugin GUIs.
type GUIProvider interface {
	// HasGUI returns true if the plugin has a GUI.
	HasGUI() bool

	// GetPreferredAPI returns the preferred GUI API and whether the GUI is floating.
	GetPreferredAPI() (api string, isFloating bool)

	// GetGUISize returns the default GUI size.
	GetGUISize() (width, height uint32)

	// OnGUICreated is called when the GUI is created.
	OnGUICreated()

	// OnGUIDestroyed is called when the GUI is destroyed.
	OnGUIDestroyed()

	// OnGUIShown is called when the GUI is shown.
	OnGUIShown()

	// OnGUIHidden is called when the GUI is hidden.
	OnGUIHidden()
}

// GUI API identifiers are defined in constants.go

// NotePortsProvider is an extension for plugins that have note ports.
// It allows hosts to query information about the plugin's note ports.
type NotePortsProvider interface {
	// GetNotePortCount returns the number of note ports.
	GetNotePortCount(isInput bool) uint32

	// GetNotePortInfo returns information about a note port.
	GetNotePortInfo(index uint32, isInput bool) NotePortInfo
}

// NotePortInfo contains information about a note port.
type NotePortInfo struct {
	// ID is a unique identifier for the port
	ID uint32

	// Name is a human-readable name for the port
	Name string

	// Flags contains additional port flags
	Flags uint32

	// SupportedDialects is a bitmask of supported note dialects
	SupportedDialects uint32

	// PreferredDialect is the preferred note dialect
	PreferredDialect uint32
}

// Note port flags and dialects are defined in constants.go

// LatencyProvider is an extension for plugins that have latency.
// It allows hosts to query the plugin's latency.
type LatencyProvider interface {
	// GetLatency returns the plugin's latency in samples.
	GetLatency() uint32
}

// TailProvider is an extension for plugins that have a tail.
// It allows hosts to query the plugin's tail length.
type TailProvider interface {
	// GetTail returns the plugin's tail length in samples.
	GetTail() uint32
}

// PresetLoader is an extension for plugins that can load presets.
// It allows hosts to load presets into the plugin.
type PresetLoader interface {
	// LoadPresetFromLocation loads a preset from a location.
	// locationKind: PresetLocationFile or PresetLocationPlugin
	// location: path to the preset file (or empty string for bundled presets)
	// loadKey: key to identify preset within a container file (or empty string)
	LoadPresetFromLocation(locationKind uint32, location, loadKey string) bool
}

// Preset location kinds
const (
	PresetLocationFile   = 0 // Preset stored in filesystem
	PresetLocationPlugin = 1 // Preset bundled within plugin
)

// TimerSupportProvider is an extension for plugins that need timer callbacks.
// It allows plugins to receive periodic timer callbacks from the host.
type TimerSupportProvider interface {
	// OnTimer is called when a timer fires.
	// timerID is the ID of the timer that fired.
	OnTimer(timerID uint64)
}

// LogProvider is an interface that allows plugins to access host logging.
// This is typically implemented by the host, not the plugin.
type LogProvider interface {
	// Log sends a log message to the host.
	// severity is one of the log severity constants.
	Log(severity int32, message string)
}

// Log severity levels
const (
	LogDebug   = 0
	LogInfo    = 1
	LogWarning = 2
	LogError   = 3
	LogFatal   = 4
	LogHostMisbehaving   = 5
	LogPluginMisbehaving = 6
)

// AudioPortsConfigProvider is an extension for plugins that support multiple audio port configurations
type AudioPortsConfigProvider interface {
	// GetAudioPortsConfigCount returns the number of available configurations
	GetAudioPortsConfigCount() uint32
	
	// GetAudioPortsConfig returns information about a configuration
	GetAudioPortsConfig(index uint32) *AudioPortsConfig
	
	// SelectAudioPortsConfig selects the configuration designated by id
	SelectAudioPortsConfig(configID uint64) bool
	
	// GetCurrentConfig returns the id of the currently selected config
	GetCurrentConfig() uint64
	
	// GetAudioPortInfoForConfig gets info about an audio port for a given config
	GetAudioPortInfoForConfig(configID uint64, portIndex uint32, isInput bool) *AudioPortInfo
}

// AudioPortsConfig describes an audio port configuration preset
type AudioPortsConfig struct {
	ID                      uint64
	Name                    string
	InputPortCount          uint32
	OutputPortCount         uint32
	HasMainInput            bool
	MainInputChannelCount   uint32
	MainInputPortType       string
	HasMainOutput           bool
	MainOutputChannelCount  uint32
	MainOutputPortType      string
}

// SurroundProvider is an extension for plugins that support surround configurations
type SurroundProvider interface {
	// IsChannelMaskSupported checks if a given channel mask is supported
	IsChannelMaskSupported(channelMask uint64) bool
	
	// GetChannelMap returns the surround channel identifiers for each channel
	GetChannelMap(isInput bool, portIndex uint32) []uint8
}

// Surround channel identifiers
const (
	SurroundFL  = 0  // Front Left
	SurroundFR  = 1  // Front Right
	SurroundFC  = 2  // Front Center
	SurroundLFE = 3  // Low Frequency
	SurroundBL  = 4  // Back (Rear) Left
	SurroundBR  = 5  // Back (Rear) Right
	SurroundFLC = 6  // Front Left of Center
	SurroundFRC = 7  // Front Right of Center
	SurroundBC  = 8  // Back (Rear) Center
	SurroundSL  = 9  // Side Left
	SurroundSR  = 10 // Side Right
	SurroundTC  = 11 // Top (Height) Center
	SurroundTFL = 12 // Top (Height) Front Left
	SurroundTFC = 13 // Top (Height) Front Center
	SurroundTFR = 14 // Top (Height) Front Right
	SurroundTBL = 15 // Top (Height) Back (Rear) Left
	SurroundTBC = 16 // Top (Height) Back (Rear) Center
	SurroundTBR = 17 // Top (Height) Back (Rear) Right
	SurroundTSL = 18 // Top (Height) Side Left
	SurroundTSR = 19 // Top (Height) Side Right
)

// VoiceInfoProvider is an extension for plugins that provide voice information
type VoiceInfoProvider interface {
	// GetVoiceInfo returns voice count and capacity information
	GetVoiceInfo() VoiceInfo
}

// VoiceInfo contains voice count and capacity information
type VoiceInfo struct {
	VoiceCount    uint32 // Current number of voices the patch can use
	VoiceCapacity uint32 // Number of allocated voices
	Flags         uint64 // Voice info flags
}

// Voice info flags
const (
	VoiceInfoSupportsOverlappingNotes = 1 << 0
)

// TrackInfoProvider is an extension for plugins that respond to track info changes
type TrackInfoProvider interface {
	// OnTrackInfoChanged is called when the track information changes
	OnTrackInfoChanged()
}

// TrackInfo contains information about the track the plugin is on
type TrackInfo struct {
	Flags              uint64 // Combination of TrackInfo flags
	Name               string // Track name (if HasTrackName flag is set)
	Color              Color  // Track color (if HasTrackColor flag is set)
	AudioChannelCount  int32  // Audio channel count (if HasAudioChannel flag is set)
	AudioPortType      string // Audio port type (if HasAudioChannel flag is set)
}

// Track info flags
const (
	TrackInfoHasTrackName      = 1 << 0
	TrackInfoHasTrackColor     = 1 << 1
	TrackInfoHasAudioChannel   = 1 << 2
	TrackInfoIsForReturnTrack  = 1 << 3
	TrackInfoIsForBus          = 1 << 4
	TrackInfoIsForMaster       = 1 << 5
)

// Color represents an RGBA color
type Color struct {
	Alpha uint8
	Red   uint8
	Green uint8
	Blue  uint8
}