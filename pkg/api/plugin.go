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
	// It receives input audio and writes output audio using Go-native slices.
	// The audio buffer conversion from C is handled automatically.
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
// All C event conversion is handled automatically by the implementation.
type EventHandler interface {
	// ProcessInputEvents processes all incoming events in the event queue.
	// This should be called during the Process method to handle parameter changes, etc.
	ProcessInputEvents()

	// AddOutputEvent adds an event to the output event queue (legacy interface).
	// Events can include parameter changes, MIDI events, etc.
	AddOutputEvent(eventType int, data interface{})

	// GetInputEventCount returns the number of input events.
	GetInputEventCount() uint32

	// GetInputEvent retrieves an input event by index.
	// C events are automatically converted to Go types.
	GetInputEvent(index uint32) *Event

	// PushOutputEvent pushes a typed output event to the host.
	// This is the preferred method over AddOutputEvent.
	PushOutputEvent(event *Event) bool

	// ProcessAllEvents processes all input events and calls the appropriate handler.
	// This is a convenience method that iterates through all events.
	ProcessAllEvents(handler TypedEventHandler)

	// ProcessTypedEvents processes all input events directly without interface{} boxing.
	// This is the zero-allocation path for event processing and should be preferred
	// over ProcessAllEvents for real-time audio processing.
	ProcessTypedEvents(handler TypedEventHandler)
}

// TypedEventHandler processes events with type-specific methods
type TypedEventHandler interface {
	// Parameter events
	HandleParamValue(event *ParamValueEvent, time uint32)
	HandleParamMod(event *ParamModEvent, time uint32)
	HandleParamGestureBegin(event *ParamGestureEvent, time uint32)
	HandleParamGestureEnd(event *ParamGestureEvent, time uint32)
	
	// Note events
	HandleNoteOn(event *NoteEvent, time uint32)
	HandleNoteOff(event *NoteEvent, time uint32)
	HandleNoteChoke(event *NoteEvent, time uint32)
	HandleNoteEnd(event *NoteEvent, time uint32)
	HandleNoteExpression(event *NoteExpressionEvent, time uint32)
	
	// Transport events
	HandleTransport(event *TransportEvent, time uint32)
	
	// MIDI events
	HandleMIDI(event *MIDIEvent, time uint32)
	HandleMIDI2(event *MIDI2Event, time uint32)
	HandleMIDISysex(event *MIDISysexEvent, time uint32)
}

// NoOpEventHandler provides default no-op implementations for all event handler methods
type NoOpEventHandler struct{}

func (h *NoOpEventHandler) HandleParamValue(event *ParamValueEvent, time uint32) {}
func (h *NoOpEventHandler) HandleParamMod(event *ParamModEvent, time uint32) {}
func (h *NoOpEventHandler) HandleParamGestureBegin(event *ParamGestureEvent, time uint32) {}
func (h *NoOpEventHandler) HandleParamGestureEnd(event *ParamGestureEvent, time uint32) {}
func (h *NoOpEventHandler) HandleNoteOn(event *NoteEvent, time uint32) {}
func (h *NoOpEventHandler) HandleNoteOff(event *NoteEvent, time uint32) {}
func (h *NoOpEventHandler) HandleNoteChoke(event *NoteEvent, time uint32) {}
func (h *NoOpEventHandler) HandleNoteEnd(event *NoteEvent, time uint32) {}
func (h *NoOpEventHandler) HandleNoteExpression(event *NoteExpressionEvent, time uint32) {}
func (h *NoOpEventHandler) HandleTransport(event *TransportEvent, time uint32) {}
func (h *NoOpEventHandler) HandleMIDI(event *MIDIEvent, time uint32) {}
func (h *NoOpEventHandler) HandleMIDI2(event *MIDI2Event, time uint32) {}
func (h *NoOpEventHandler) HandleMIDISysex(event *MIDISysexEvent, time uint32) {}

// BaseTypedEventHandler is deprecated - use NoOpEventHandler
type BaseTypedEventHandler = NoOpEventHandler

// Event represents an event in the CLAP processing context.
type Event struct {
	// Type is the type of event (e.g., note on, note off, parameter change)
	Type int

	// Time is the offset in samples from the start of the current process block
	Time uint32

	// Flags contains event flags (e.g., is_live, dont_record)
	Flags uint32

	// Data contains event-specific data
	Data interface{}
}

// ParamValueEvent represents a parameter value change event
type ParamValueEvent struct {
	ParamID uint32
	Cookie  unsafe.Pointer
	NoteID  int32
	Port    int16
	Channel int16
	Key     int16
	Value   float64
}

// ParamModEvent represents a parameter modulation event
type ParamModEvent struct {
	ParamID uint32
	Cookie  unsafe.Pointer
	NoteID  int32
	Port    int16
	Channel int16
	Key     int16
	Amount  float64
}

// NoteEvent represents a note event (note on, note off, choke, end)
type NoteEvent struct {
	NoteID   int32
	Port     int16
	Channel  int16
	Key      int16
	Velocity float64
}

// NoteExpressionEvent represents a note expression event
type NoteExpressionEvent struct {
	ExpressionID int32
	NoteID       int32
	Port         int16
	Channel      int16
	Key          int16
	Value        float64
}

// ParamGestureEvent represents parameter gesture begin/end events
type ParamGestureEvent struct {
	ParamID uint32
}

// TransportEvent represents transport state changes
type TransportEvent struct {
	Flags                uint32
	SongPosBeats         uint64
	SongPosSeconds       uint64
	Tempo                float64
	TempoInc             float64
	LoopStartBeats       uint64
	LoopEndBeats         uint64
	LoopStartSeconds     uint64
	LoopEndSeconds       uint64
	BarStart             uint64
	BarNumber            int32
	TimeSignatureNum     uint16
	TimeSignatureDen     uint16
}

// MIDIEvent represents MIDI 1.0 events
type MIDIEvent struct {
	Port int16
	Data [3]byte
}

// MIDISysexEvent represents MIDI system exclusive events
type MIDISysexEvent struct {
	Port int16
	Data []byte
}

// MIDI2Event represents MIDI 2.0 events
type MIDI2Event struct {
	Port int16
	Data [4]uint32
}

// Process status codes and event types are defined in constants.go


// Extension IDs are defined in constants.go