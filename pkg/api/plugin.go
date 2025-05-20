// Package api defines the core interfaces for CLAP plugins in Go.
// It provides the base abstractions that plugins must implement to work with
// the CLAP host via the clapgo bridge.
package api

import (
	"unsafe"
)

// Plugin represents the core interface that all CLAP plugins must implement.
// It defines the lifecycle and processing functions required by the CLAP standard.
type Plugin interface {
	// Init initializes the plugin. Called after creation.
	// Return true if the initialization was successful.
	Init() bool

	// Destroy releases all resources associated with the plugin.
	// Called before the plugin is deleted.
	Destroy()

	// Activate prepares the plugin for processing.
	// It provides the sample rate and buffer size constraints.
	// Return true if the activation was successful.
	Activate(sampleRate float64, minFrames, maxFrames uint32) bool

	// Deactivate stops the plugin from processing.
	// Called when the plugin is no longer going to be used for processing.
	Deactivate()

	// StartProcessing signals that the plugin should prepare for audio processing.
	// Return true if the preparation was successful.
	StartProcessing() bool

	// StopProcessing signals that the plugin should stop audio processing.
	StopProcessing()

	// Reset resets the plugin state to its initial state.
	Reset()

	// Process handles audio processing.
	// It receives input audio and writes output audio.
	// It also processes events such as parameter changes.
	// Returns a status code: 0-error, 1-continue, 2-continue_if_not_quiet, 3-tail, 4-sleep
	Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events EventHandler) int

	// GetExtension retrieves a plugin extension by ID.
	// Return nil if the extension is not supported.
	GetExtension(id string) unsafe.Pointer

	// OnMainThread is called on the main thread.
	// It can be used for UI updates or other main-thread operations.
	OnMainThread()

	// GetPluginID returns the plugin's unique identifier.
	GetPluginID() string

	// GetPluginInfo returns information about the plugin.
	GetPluginInfo() PluginInfo
}

// PluginInfo holds metadata about a CLAP plugin.
type PluginInfo struct {
	// ID is the unique identifier for the plugin (e.g., "com.example.my-plugin")
	ID string

	// Name is the human-readable name of the plugin (e.g., "My Awesome Plugin")
	Name string

	// Vendor is the name of the plugin developer or company (e.g., "Example Audio")
	Vendor string

	// URL is the website URL for the plugin (e.g., "https://example.com/my-plugin")
	URL string

	// ManualURL is the URL to the plugin's manual (e.g., "https://example.com/my-plugin/manual")
	ManualURL string

	// SupportURL is the URL for plugin support (e.g., "https://example.com/support")
	SupportURL string

	// Version is the plugin version (e.g., "1.0.0")
	Version string

	// Description is a short description of the plugin
	Description string

	// Features is a list of plugin features (e.g., "audio-effect", "stereo", "mono")
	Features []string
}

// EventHandler handles plugin events during processing.
type EventHandler interface {
	// ProcessInputEvents processes all incoming events in the event queue.
	// This should be called during the Process method to handle parameter changes, etc.
	ProcessInputEvents()

	// AddOutputEvent adds an event to the output event queue.
	// Events can include parameter changes, MIDI events, etc.
	AddOutputEvent(eventType int, data interface{})

	// GetInputEventCount returns the number of input events.
	GetInputEventCount() uint32

	// GetInputEvent retrieves an input event by index.
	GetInputEvent(index uint32) *Event
}

// Event represents an event in the CLAP processing context.
type Event struct {
	// Type is the type of event (e.g., note on, note off, parameter change)
	Type int

	// Time is the offset in samples from the start of the current process block
	Time uint32

	// Data contains event-specific data
	Data interface{}
}

// ParamEvent represents a parameter change event.
type ParamEvent struct {
	// ParamID is the ID of the parameter being changed
	ParamID uint32

	// Cookie is an opaque pointer to plugin-specific data
	Cookie unsafe.Pointer

	// Note is the note ID if the parameter is associated with a note
	Note int32

	// Port is the port index if the parameter is associated with a port
	Port int16

	// Channel is the channel index if the parameter is associated with a channel
	Channel int16

	// Key is the key number if the parameter is associated with a key
	Key int16

	// Value is the new parameter value
	Value float64

	// Flags contains additional parameter change flags
	Flags uint32
}

// NoteEvent represents a note event (note on, note off, etc.).
type NoteEvent struct {
	// NoteID is a unique identifier for the note instance
	NoteID int32

	// Port is the port index
	Port int16

	// Channel is the channel index (0-15)
	Channel int16

	// Key is the note key (0-127)
	Key int16

	// Value is the note velocity or other value
	Value float64
}

// Process status codes and event types are defined in constants.go

// Factory creates plugin instances.
type Factory interface {
	// GetPluginCount returns the number of available plugins.
	GetPluginCount() uint32

	// GetPluginInfo returns information about a plugin by index.
	GetPluginInfo(index uint32) PluginInfo

	// CreatePlugin creates a new plugin instance with the given ID.
	CreatePlugin(id string) Plugin
}

// Creator creates a specific plugin type.
type Creator interface {
	// Create returns a new instance of the plugin.
	Create() Plugin

	// GetPluginInfo returns information about the plugin.
	GetPluginInfo() PluginInfo
}

// PluginRegistrar defines the interface for registering plugins with the system.
// Implementations of this interface provide the functionality to add plugins
// to the registry and make them available to CLAP hosts.
type PluginRegistrar interface {
	// Register adds a plugin to the registry with the given info and creator function.
	// The creator function should return a new instance of the plugin when called.
	Register(info PluginInfo, creator func() Plugin)
	
	// GetPluginCount returns the number of registered plugins.
	GetPluginCount() uint32
	
	// GetPluginInfo returns information about a plugin by index.
	GetPluginInfo(index uint32) PluginInfo
	
	// CreatePlugin creates a new plugin instance with the given ID.
	CreatePlugin(id string) Plugin
}

// PluginProvider defines the interface that plugin implementations should implement
// to register themselves with the system.
type PluginProvider interface {
	// GetPluginInfo returns information about the plugin.
	GetPluginInfo() PluginInfo
	
	// CreatePlugin returns a new instance of the plugin.
	CreatePlugin() Plugin
}

// Extension IDs are defined in constants.go