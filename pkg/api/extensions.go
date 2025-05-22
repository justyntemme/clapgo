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

// PresetLoadProvider is an extension for plugins that can load presets.
// It allows hosts to load presets into the plugin.
type PresetLoadProvider interface {
	// LoadPreset loads a preset from a location.
	LoadPreset(locationKind uint32, location, loadKey string) bool
}

// Preset location kinds are defined in constants.go