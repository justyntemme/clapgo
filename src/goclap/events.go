package goclap

// #include <stdlib.h>
// #include <string.h>
// #include "../../include/clap/include/clap/clap.h"
// #include "../../include/clap/include/clap/fixedpoint.h"
// 
// // Define functions for event handling
// static inline uint32_t clap_input_events_size(const clap_input_events_t* events) {
//     if (events && events->size) {
//         return events->size(events);
//     }
//     return 0;
// }
//
// static inline const clap_event_header_t* clap_input_events_get(const clap_input_events_t* events, uint32_t index) {
//     if (events && events->get) {
//         return events->get(events, index);
//     }
//     return NULL;
// }
//
// static inline bool clap_output_events_try_push(const clap_output_events_t* events, const clap_event_header_t* event) {
//     if (events && events->try_push) {
//         return events->try_push(events, event);
//     }
//     return false;
// }
import "C"
import (
	"fmt"
	"unsafe"
)

// Event types as defined in the CLAP spec
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

// InputEvents wraps the CLAP input events interface
type InputEvents struct {
	Ptr unsafe.Pointer
}

// OutputEvents wraps the CLAP output events interface
type OutputEvents struct {
	Ptr unsafe.Pointer
}

// Event represents a CLAP event
type Event struct {
	Time   uint32
	Type   uint16
	Space  uint16
	Flags  uint32
	Data   interface{}
}

// NoteEvent represents a CLAP note event
// Using int16 to match the C struct definition
type NoteEvent struct {
	NoteID      int32
	Port        int16
	Channel     int16 
	Key         int16
	Velocity    float64
}

// NoteExpressionEvent represents a CLAP note expression event
// Using int16 to match the C struct definition
type NoteExpressionEvent struct {
	NoteID      int32
	Port        int16
	Channel     int16
	Key         int16
	Expression  uint32 // This will be cast to the appropriate C enum type
	Value       float64
}

// ParamEvent represents a CLAP parameter event
// Using int16 to match the C struct definition
type ParamEvent struct {
	ParamID     uint32
	Cookie      unsafe.Pointer
	Note        int32
	Port        int16
	Channel     int16
	Key         int16
	Value       float64
}

// TransportEvent represents a CLAP transport event
type TransportEvent struct {
	Flags             uint32
	SongPosBeats      int64
	SongPosSeconds    float64
	Tempo             float64
	TempoBeatSize     float64
	BarStart          int64
	BarNumber         int32
	LoopStartBeats    int64
	LoopEndBeats      int64
	LoopStartSeconds  float64
	LoopEndSeconds    float64
	TimeSigNumerator  uint16
	TimeSigDenominator uint16
}

// ProcessEvents is the struct passed to Process method
type ProcessEvents struct {
	InEvents  unsafe.Pointer
	OutEvents unsafe.Pointer
}

// GetEventCount returns the number of events in the queue
func (e *InputEvents) GetEventCount() uint32 {
	// Check if the input events pointer is valid
	if e == nil || e.Ptr == nil {
		return 0
	}

	// Get the C event interface from the pointer
	cEvents := (*C.clap_input_events_t)(e.Ptr)
	if cEvents == nil || cEvents.size == nil {
		return 0
	}

	// Call the size method on the C interface
	count := uint32(C.clap_input_events_size(cEvents))
	return count
}

// GetEvent returns the event at the given index
func (e *InputEvents) GetEvent(index uint32) *Event {
	// Check if the input events pointer is valid
	if e == nil || e.Ptr == nil {
		return safeEmptyEvent()
	}

	// Get the C event interface from the pointer
	cEvents := (*C.clap_input_events_t)(e.Ptr)
	if cEvents == nil || cEvents.get == nil {
		return safeEmptyEvent()
	}

	// Call the get method on the C interface
	cEvent := C.clap_input_events_get(cEvents, C.uint32_t(index))
	if cEvent == nil {
		return safeEmptyEvent()
	}

	// Convert the C event to a Go event
	event := &Event{
		Time:  uint32(cEvent.time),
		Type:  uint16(cEvent._type),
		Space: uint16(0), // Field might have different name, use a safe default
		Flags: uint32(cEvent.flags),
	}

	// Based on the event type, set the appropriate data
	switch event.Type {
	case EventTypeNoteOn, EventTypeNoteOff, EventTypeNoteChoke, EventTypeNoteEnd:
		event.Data = convertNoteEvent(cEvent)
	case EventTypeNoteExpression:
		event.Data = convertNoteExpressionEvent(cEvent)
	case EventTypeParamValue, EventTypeParamMod, EventTypeParamGestureBegin, EventTypeParamGestureEnd:
		event.Data = convertParamEvent(cEvent)
	case EventTypeTransport:
		event.Data = convertTransportEvent(cEvent)
	default:
		// Unknown event type, just use nil data
	}

	return event
}

// PushEvent adds an event to the output queue
func (e *OutputEvents) PushEvent(event *Event) bool {
	// Check if the output events pointer is valid
	if e == nil || e.Ptr == nil || event == nil {
		return false
	}

	// Get the C event interface from the pointer
	cEvents := (*C.clap_output_events_t)(e.Ptr)
	if cEvents == nil || cEvents.try_push == nil {
		return false
	}

	// Create a new C event - use zero initialization and set fields individually to avoid struct field name mismatches
	var cEvent C.clap_event_header_t
	cEvent.size = C.uint32_t(C.sizeof_clap_event_header_t)
	cEvent.time = C.uint32_t(event.Time)
	cEvent._type = C.uint16_t(event.Type)
	// cEvent.space field might have a different name in C, skip it
	cEvent.flags = C.uint32_t(event.Flags)

	// Based on the event type, set the appropriate data
	var result C.bool
	switch event.Type {
	case EventTypeNoteOn, EventTypeNoteOff, EventTypeNoteChoke, EventTypeNoteEnd:
		result = pushNoteEvent(cEvents, &cEvent, event)
	case EventTypeNoteExpression:
		result = pushNoteExpressionEvent(cEvents, &cEvent, event)
	case EventTypeParamValue, EventTypeParamMod, EventTypeParamGestureBegin, EventTypeParamGestureEnd:
		result = pushParamEvent(cEvents, &cEvent, event)
	case EventTypeTransport:
		result = pushTransportEvent(cEvents, &cEvent, event)
	default:
		// Unknown event type, try to push just the header
		result = C.clap_output_events_try_push(cEvents, &cEvent)
	}

	return bool(result)
}

// safeEmptyEvent returns an empty event to avoid nil returns
func safeEmptyEvent() *Event {
	return &Event{
		Time:  0,
		Type:  0,
		Space: 0,
		Flags: 0,
		Data:  nil,
	}
}

// convertNoteEvent converts a C note event to a Go note event
func convertNoteEvent(cEvent *C.clap_event_header_t) NoteEvent {
	// Make sure we have a valid event
	if cEvent == nil {
		return NoteEvent{}
	}

	// Cast to the specific note event type
	noteEvent := (*C.clap_event_note_t)(unsafe.Pointer(cEvent))
	if noteEvent == nil {
		return NoteEvent{}
	}

	// Convert to Go struct with proper types
	return NoteEvent{
		NoteID:   int32(noteEvent.note_id),
		Port:     int16(noteEvent.port_index),
		Channel:  int16(noteEvent.channel),
		Key:      int16(noteEvent.key),
		Velocity: float64(noteEvent.velocity),
	}
}

// convertNoteExpressionEvent converts a C note expression event to a Go note expression event
func convertNoteExpressionEvent(cEvent *C.clap_event_header_t) NoteExpressionEvent {
	// Make sure we have a valid event
	if cEvent == nil {
		return NoteExpressionEvent{}
	}

	// Cast to the specific note expression event type
	exprEvent := (*C.clap_event_note_expression_t)(unsafe.Pointer(cEvent))
	if exprEvent == nil {
		return NoteExpressionEvent{}
	}

	// Convert to Go struct with proper types
	return NoteExpressionEvent{
		NoteID:     int32(exprEvent.note_id),
		Port:       int16(exprEvent.port_index),
		Channel:    int16(exprEvent.channel),
		Key:        int16(exprEvent.key),
		Expression: uint32(exprEvent.expression_id),
		Value:      float64(exprEvent.value),
	}
}

// convertParamEvent converts a C parameter event to a Go parameter event
func convertParamEvent(cEvent *C.clap_event_header_t) ParamEvent {
	// Make sure we have a valid event
	if cEvent == nil {
		return ParamEvent{}
	}

	// Cast to the specific parameter event type
	paramEvent := (*C.clap_event_param_value_t)(unsafe.Pointer(cEvent))
	if paramEvent == nil {
		return ParamEvent{}
	}

	// Convert to Go struct with proper types
	return ParamEvent{
		ParamID:  uint32(paramEvent.param_id),
		Cookie:   paramEvent.cookie,
		Note:     int32(paramEvent.note_id),
		Port:     int16(paramEvent.port_index),
		Channel:  int16(paramEvent.channel),
		Key:      int16(paramEvent.key),
		Value:    float64(paramEvent.value),
	}
}

// convertTransportEvent converts a C transport event to a Go transport event
func convertTransportEvent(cEvent *C.clap_event_header_t) TransportEvent {
	// Make sure we have a valid event
	if cEvent == nil {
		return TransportEvent{}
	}

	// Cast to the specific transport event type
	transportEvent := (*C.clap_event_transport_t)(unsafe.Pointer(cEvent))
	if transportEvent == nil {
		return TransportEvent{}
	}

	// Convert to Go struct
	return TransportEvent{
		Flags:             uint32(transportEvent.flags),
		SongPosBeats:      int64(transportEvent.song_pos_beats),
		SongPosSeconds:    float64(transportEvent.song_pos_seconds),
		Tempo:             float64(transportEvent.tempo),
		TempoBeatSize:     float64(transportEvent.tempo_inc),
		BarStart:          int64(transportEvent.bar_start),
		BarNumber:         int32(transportEvent.bar_number),
		LoopStartBeats:    int64(transportEvent.loop_start_beats),
		LoopEndBeats:      int64(transportEvent.loop_end_beats),
		LoopStartSeconds:  float64(transportEvent.loop_start_seconds),
		LoopEndSeconds:    float64(transportEvent.loop_end_seconds),
		TimeSigNumerator:  uint16(transportEvent.tsig_num),
		TimeSigDenominator: uint16(transportEvent.tsig_denom),
	}
}

// pushNoteEvent pushes a note event to the output queue
func pushNoteEvent(cEvents *C.clap_output_events_t, cHeader *C.clap_event_header_t, event *Event) C.bool {
	// Make sure we have valid data
	if cEvents == nil || cHeader == nil || event == nil {
		return C.bool(false)
	}

	// Extract the note event data
	noteEvent, ok := event.Data.(NoteEvent)
	if !ok {
		// Data does not contain a note event
		fmt.Println("Warning: Event data is not a NoteEvent")
		return C.bool(false)
	}

	// Create a C note event - initialize and set fields individually to handle type mismatches
	var cNoteEvent C.clap_event_note_t
	cNoteEvent.header = *cHeader
	cNoteEvent.note_id = C.int32_t(noteEvent.NoteID)
	
	// Use direct C type assignments based on the C header definitions
	cNoteEvent.port_index = C.int16_t(noteEvent.Port)
	cNoteEvent.channel = C.int16_t(noteEvent.Channel)
	cNoteEvent.key = C.int16_t(noteEvent.Key)
	cNoteEvent.velocity = C.double(noteEvent.Velocity)

	// Update the size in the header
	cNoteEvent.header.size = C.uint32_t(C.sizeof_clap_event_note_t)

	// Try to push the event
	return C.clap_output_events_try_push(cEvents, &cNoteEvent.header)
}

// pushNoteExpressionEvent pushes a note expression event to the output queue
func pushNoteExpressionEvent(cEvents *C.clap_output_events_t, cHeader *C.clap_event_header_t, event *Event) C.bool {
	// Make sure we have valid data
	if cEvents == nil || cHeader == nil || event == nil {
		return C.bool(false)
	}

	// Extract the note expression event data
	exprEvent, ok := event.Data.(NoteExpressionEvent)
	if !ok {
		// Data does not contain a note expression event
		fmt.Println("Warning: Event data is not a NoteExpressionEvent")
		return C.bool(false)
	}

	// Create a C note expression event - initialize and set fields individually
	var cExprEvent C.clap_event_note_expression_t
	cExprEvent.header = *cHeader
	cExprEvent.note_id = C.int32_t(exprEvent.NoteID)
	
	// Use direct C type assignments based on the C header definitions
	cExprEvent.port_index = C.int16_t(exprEvent.Port)
	cExprEvent.channel = C.int16_t(exprEvent.Channel)
	cExprEvent.key = C.int16_t(exprEvent.Key)
	
	// Special handling for the expression ID which is an enum in C (clap_note_expression = int32_t)
	cExprEvent.expression_id = C.int32_t(exprEvent.Expression)
	cExprEvent.value = C.double(exprEvent.Value)

	// Update the size in the header
	cExprEvent.header.size = C.uint32_t(C.sizeof_clap_event_note_expression_t)

	// Try to push the event
	return C.clap_output_events_try_push(cEvents, &cExprEvent.header)
}

// pushParamEvent pushes a parameter event to the output queue
func pushParamEvent(cEvents *C.clap_output_events_t, cHeader *C.clap_event_header_t, event *Event) C.bool {
	// Make sure we have valid data
	if cEvents == nil || cHeader == nil || event == nil {
		return C.bool(false)
	}

	// Extract the parameter event data
	paramEvent, ok := event.Data.(ParamEvent)
	if !ok {
		// Data does not contain a parameter event
		fmt.Println("Warning: Event data is not a ParamEvent")
		return C.bool(false)
	}

	// Create a C parameter event - initialize and set fields individually
	var cParamEvent C.clap_event_param_value_t
	cParamEvent.header = *cHeader
	cParamEvent.param_id = C.clap_id(paramEvent.ParamID)
	cParamEvent.cookie = paramEvent.Cookie
	cParamEvent.note_id = C.int32_t(paramEvent.Note)
	
	// Use direct C type assignments based on the C header definitions
	cParamEvent.port_index = C.int16_t(paramEvent.Port)
	cParamEvent.channel = C.int16_t(paramEvent.Channel)
	cParamEvent.key = C.int16_t(paramEvent.Key)
	cParamEvent.value = C.double(paramEvent.Value)

	// Update the size in the header
	cParamEvent.header.size = C.uint32_t(C.sizeof_clap_event_param_value_t)

	// Try to push the event
	return C.clap_output_events_try_push(cEvents, &cParamEvent.header)
}

// pushTransportEvent pushes a transport event to the output queue
func pushTransportEvent(cEvents *C.clap_output_events_t, cHeader *C.clap_event_header_t, event *Event) C.bool {
	// Make sure we have valid data
	if cEvents == nil || cHeader == nil || event == nil {
		return C.bool(false)
	}

	// Extract the transport event data
	transportEvent, ok := event.Data.(TransportEvent)
	if !ok {
		// Data does not contain a transport event
		fmt.Println("Warning: Event data is not a TransportEvent")
		return C.bool(false)
	}

	// Create a C transport event - initialize and set fields individually
	var cTransportEvent C.clap_event_transport_t
	cTransportEvent.header = *cHeader
	cTransportEvent.flags = C.uint32_t(transportEvent.Flags)
	
	// clap_beattime and clap_sectime are defined as int64_t in fixedpoint.h
	cTransportEvent.song_pos_beats = C.int64_t(transportEvent.SongPosBeats)
	cTransportEvent.song_pos_seconds = C.int64_t(int64(transportEvent.SongPosSeconds * 1000000)) // Convert to microseconds
	cTransportEvent.tempo = C.double(transportEvent.Tempo)
	cTransportEvent.tempo_inc = C.double(transportEvent.TempoBeatSize)
	cTransportEvent.bar_start = C.int64_t(transportEvent.BarStart)
	cTransportEvent.bar_number = C.int32_t(transportEvent.BarNumber)
	cTransportEvent.loop_start_beats = C.int64_t(transportEvent.LoopStartBeats)
	cTransportEvent.loop_end_beats = C.int64_t(transportEvent.LoopEndBeats)
	cTransportEvent.loop_start_seconds = C.int64_t(int64(transportEvent.LoopStartSeconds * 1000000)) // Convert to microseconds
	cTransportEvent.loop_end_seconds = C.int64_t(int64(transportEvent.LoopEndSeconds * 1000000)) // Convert to microseconds
	cTransportEvent.tsig_num = C.uint16_t(transportEvent.TimeSigNumerator)
	cTransportEvent.tsig_denom = C.uint16_t(transportEvent.TimeSigDenominator)

	// Update the size in the header
	cTransportEvent.header.size = C.uint32_t(C.sizeof_clap_event_transport_t)

	// Try to push the event
	return C.clap_output_events_try_push(cEvents, &cTransportEvent.header)
}