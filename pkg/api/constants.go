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
	ExtAudioPortsActivation = "clap.audio-ports-activation"
	ExtAmbisonic            = "clap.ambisonic"
	ExtSurround             = "clap.surround"
	ExtNoteName             = "clap.note-name"
	ExtContextMenu          = "clap.context-menu"
	
	// Draft extensions
	ExtResourceDirectory    = "clap.ext.draft.resource-directory"
	ExtTransportControl     = "clap.ext.draft.transport-control"
)

// Event types
const (
	EventTypeNoteOn           = 0
	EventTypeNoteOff          = 1
	EventTypeNoteChoke        = 2
	EventTypeNoteEnd          = 3
	EventTypeNoteExpression   = 4
	EventTypeParamValue       = 5
	EventTypeParamMod         = 6
	EventTypeParamGestureBegin = 7
	EventTypeParamGestureEnd  = 8
	EventTypeTransport        = 9
	EventTypeMIDI             = 10
	EventTypeMIDISysex        = 11
	EventTypeMIDI2            = 12
)

// Note expression types
const (
	NoteExpressionVolume      = 0
	NoteExpressionPan         = 1
	NoteExpressionTuning      = 2
	NoteExpressionVibrato     = 3
	NoteExpressionExpression  = 4
	NoteExpressionBrightness  = 5
	NoteExpressionPressure    = 6
)

// Transport flags
const (
	TransportHasTransport     = 1 << 0
	TransportHasTempo         = 1 << 1
	TransportHasBeatsTime     = 1 << 2
	TransportHasSecondsTime   = 1 << 3
	TransportHasTimeSignature = 1 << 4
	TransportIsPlaying        = 1 << 5
	TransportIsRecording      = 1 << 6
	TransportIsLooping        = 1 << 7
	TransportIsWithinPreRoll  = 1 << 8
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

// Process status codes
const (
	ProcessError           = 0
	ProcessContinue        = 1
	ProcessContinueIfNotQuiet = 2
	ProcessTail            = 3
	ProcessSleep           = 4
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

// Other constants
const (
	InvalidID = 0xFFFFFFFF // Invalid ID constant for ports without a pair
)