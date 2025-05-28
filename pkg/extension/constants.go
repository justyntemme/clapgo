package extension

// CLAP extension identifiers
const (
	// Core extensions
	AudioPorts           = "clap.audio-ports"
	Params               = "clap.params"
	State                = "clap.state"
	GUI                  = "clap.gui"
	NotePorts            = "clap.note-ports"
	TimerSupport         = "clap.timer-support"
	Latency              = "clap.latency"
	Tail                 = "clap.tail"
	Render               = "clap.render"
	PosixFDSupport       = "clap.posix-fd-support"
	ThreadCheck          = "clap.thread-check"
	ThreadPool           = "clap.thread-pool"
	VoiceInfoID          = "clap.voice-info"
	TrackInfo            = "clap.track-info"
	LogSupport           = "clap.log"
	PresetLoad           = "clap.preset-load"
	RemoteControls       = "clap.remote-controls"
	StateContext         = "clap.state-context"
	EventRegistry        = "clap.event-registry"
	ParamIndication      = "clap.param-indication"
	ConfigurableAudioPorts = "clap.configurable-audio-ports"
	AudioPortsConfigID   = "clap.audio-ports-config"
	AudioPortsActivation = "clap.audio-ports-activation/2"
	AudioPortsActivationCompat = "clap.audio-ports-activation/draft-2"
	Ambisonic            = "clap.ambisonic"
	Surround             = "clap.surround"
	NoteNameID           = "clap.note-name"
	ContextMenu          = "clap.context-menu"
	
	// Draft extensions
	ResourceDirectory    = "clap.resource-directory.draft/1"
	TransportControl     = "clap.transport-control/1"
	Tuning              = "clap.tuning/2"
)

// Extension support check interface
type Supporter interface {
	// SupportsExtension returns true if the extension is supported
	SupportsExtension(extensionID string) bool
	
	// GetExtension returns the extension implementation or nil
	GetExtension(extensionID string) interface{}
}

// Registry for extension implementations
type Registry struct {
	extensions map[string]interface{}
}

// NewRegistry creates a new extension registry
func NewRegistry() *Registry {
	return &Registry{
		extensions: make(map[string]interface{}),
	}
}

// Register registers an extension implementation
func (r *Registry) Register(extensionID string, impl interface{}) {
	r.extensions[extensionID] = impl
}

// Get returns an extension implementation or nil
func (r *Registry) Get(extensionID string) interface{} {
	return r.extensions[extensionID]
}

// Supports returns true if the extension is registered
func (r *Registry) Supports(extensionID string) bool {
	_, exists := r.extensions[extensionID]
	return exists
}

// List returns all registered extension IDs
func (r *Registry) List() []string {
	ids := make([]string, 0, len(r.extensions))
	for id := range r.extensions {
		ids = append(ids, id)
	}
	return ids
}