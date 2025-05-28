package api

// Extension IDs
const (
	// Core extensions
	ExtAudioPorts           = "clap.audio-ports"
	ExtParams               = "clap.params"
	ExtState                = "clap.state"
	ExtGUI                  = "clap.gui"
	ExtNotePorts            = "clap.note-ports"
	ExtTimerSupport         = "clap.timer-support"
	ExtLatency              = "clap.latency"
	ExtTail                 = "clap.tail"
	ExtRender               = "clap.render"
	ExtPosixFDSupport       = "clap.posix-fd-support"
	ExtThreadCheck          = "clap.thread-check"
	ExtThreadPool           = "clap.thread-pool"
	ExtVoiceInfo            = "clap.voice-info"
	ExtTrackInfo            = "clap.track-info"
	ExtLogSupport           = "clap.log"
	ExtPresetLoad           = "clap.preset-load"
	ExtRemoteControls       = "clap.remote-controls"
	ExtStateContext         = "clap.state-context"
	ExtEventRegistry        = "clap.event-registry"
	ExtParamIndication      = "clap.param-indication"
	ExtConfigurableAudioPorts = "clap.configurable-audio-ports"
	ExtAudioPortsConfig     = "clap.audio-ports-config"
	ExtAudioPortsActivation = "clap.audio-ports-activation/2"
	ExtAudioPortsActivationCompat = "clap.audio-ports-activation/draft-2"
	ExtAmbisonic            = "clap.ambisonic"
	ExtSurround             = "clap.surround"
	ExtNoteName             = "clap.note-name"
	ExtContextMenu          = "clap.context-menu"
	
	// Draft extensions
	ExtResourceDirectory    = "clap.resource-directory.draft/1"
	ExtTransportControl     = "clap.transport-control/1"
	ExtTuning              = "clap.tuning/2"
)

// Note dialects
const (
	NoteDialectCLAP  = 1 << 0
	NoteDialectMIDI1 = 1 << 1
	NoteDialectMIDI2 = 1 << 2
)

// Note port flags
const (
	NotePortIsMain = 1 << 0
)

// Audio port flags
const (
	AudioPortIsMain     = 1 << 0
	AudioPortIsCVOut    = 1 << 1
	AudioPortIsCVIn     = 1 << 2
	AudioPortIsAux      = 1 << 3
	AudioPortIsSidechain = 1 << 4
)

// Port types
const (
	PortMono      = "mono"
	PortStereo    = "stereo"
	PortSurround  = "surround"
	PortAmbisonic = "ambisonic"
)

// Parameter flags
const (
	ParamIsSteppable            = 1 << 0
	ParamIsPeriodic             = 1 << 1
	ParamIsHidden               = 1 << 2
	ParamIsReadonly             = 1 << 3
	ParamIsBypass               = 1 << 4
	ParamIsAutomatable          = 1 << 5
	ParamIsAutomatePerNote      = 1 << 6
	ParamIsAutomatePerKey       = 1 << 7
	ParamIsAutomatePerChannel   = 1 << 8
	ParamIsAutomatePerPort      = 1 << 9
	ParamIsModulatable          = 1 << 10
	ParamIsPerformanceParameter = 1 << 11
	ParamIsBoundedBelow         = 1 << 12
	ParamIsBoundedAbove         = 1 << 13
)

// GUI API identifiers
const (
	WindowAPIX11     = "x11"
	WindowAPIWin32   = "win32"
	WindowAPICocoa   = "cocoa"
	WindowAPIWayland = "wayland"
)

// Preset location kinds
const (
	PresetLocationFilePath = 0
	PresetLocationFileFD   = 1
)

// Log severity levels
const (
	LogSeverityDebug     = 0
	LogSeverityInfo      = 1
	LogSeverityWarning   = 2
	LogSeverityError     = 3
	LogSeverityFatal     = 4
	LogSeverityHostMisbehaving = 5
	LogSeverityPluginMisbehaving = 6
)

// Other constants
const (
	InvalidID = 0xFFFFFFFF // Invalid ID constant for ports without a pair
)