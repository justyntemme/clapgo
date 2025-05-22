package api

/*
#cgo CFLAGS: -I../../include/clap/include
#include "../../include/clap/include/clap/clap.h"
#include <stdlib.h>

// Helper functions for CLAP event handling
static inline uint32_t clap_input_events_size_helper(const clap_input_events_t* events) {
    if (events && events->size) {
        return events->size(events);
    }
    return 0;
}

static inline const clap_event_header_t* clap_input_events_get_helper(const clap_input_events_t* events, uint32_t index) {
    if (events && events->get) {
        return events->get(events, index);
    }
    return NULL;
}

static inline bool clap_output_events_try_push_helper(const clap_output_events_t* events, const clap_event_header_t* event) {
    if (events && events->try_push) {
        return events->try_push(events, event);
    }
    return false;
}
*/
import "C"
import (
	"unsafe"
)

// EventProcessor handles all C event conversion internally
// This abstracts away all the unsafe pointer operations from plugin developers
type EventProcessor struct {
	inputEvents  *C.clap_input_events_t
	outputEvents *C.clap_output_events_t
}

// NewEventProcessor creates a new EventProcessor from C event queues
func NewEventProcessor(inputEvents, outputEvents unsafe.Pointer) *EventProcessor {
	return &EventProcessor{
		inputEvents:  (*C.clap_input_events_t)(inputEvents),
		outputEvents: (*C.clap_output_events_t)(outputEvents),
	}
}

// GetInputEventCount returns the number of input events
func (ep *EventProcessor) GetInputEventCount() uint32 {
	if ep.inputEvents == nil {
		return 0
	}
	
	return uint32(C.clap_input_events_size_helper(ep.inputEvents))
}

// GetInputEvent retrieves an input event by index and converts it to Go types
func (ep *EventProcessor) GetInputEvent(index uint32) *Event {
	if ep.inputEvents == nil {
		return nil
	}
	
	cEventHeader := C.clap_input_events_get_helper(ep.inputEvents, C.uint32_t(index))
	if cEventHeader == nil {
		return nil
	}
	
	return convertCEventToGo(cEventHeader)
}

// PushOutputEvent pushes an output event to the host
func (ep *EventProcessor) PushOutputEvent(event *Event) bool {
	if ep.outputEvents == nil || event == nil {
		return false
	}
	
	cEvent := convertGoEventToC(event)
	if cEvent == nil {
		return false
	}
	
	result := C.clap_output_events_try_push_helper(ep.outputEvents, cEvent)
	
	// Free the C event memory
	freeCEvent(cEvent)
	
	return bool(result)
}

// ProcessInputEvents processes all incoming events
func (ep *EventProcessor) ProcessInputEvents() {
	// This is called by the plugin to indicate it wants to process events
	// The actual processing is done by calling GetInputEvent() in a loop
}

// AddOutputEvent adds an event to the output queue (legacy interface)
func (ep *EventProcessor) AddOutputEvent(eventType int, data interface{}) {
	event := &Event{
		Type: eventType,
		Time: 0, // Immediate
		Data: data,
	}
	ep.PushOutputEvent(event)
}

// convertCEventToGo converts a C CLAP event to a Go Event struct
func convertCEventToGo(cEventHeader *C.clap_event_header_t) *Event {
	if cEventHeader == nil {
		return nil
	}
	
	event := &Event{
		Time: uint32(cEventHeader.time),
		Type: int(cEventHeader._type),
	}
	
	// Convert event data based on type
	switch int(cEventHeader._type) {
	case EventTypeParamValue:
		cParamEvent := (*C.clap_event_param_value_t)(unsafe.Pointer(cEventHeader))
		event.Data = ParamEvent{
			ParamID: uint32(cParamEvent.param_id),
			Cookie:  unsafe.Pointer(cParamEvent.cookie),
			Note:    int32(cParamEvent.note_id),
			Port:    int16(cParamEvent.port_index),
			Channel: int16(cParamEvent.channel),
			Key:     int16(cParamEvent.key),
			Value:   float64(cParamEvent.value),
			Flags:   0, // param value events don't have flags
		}
		
	case EventTypeParamGestureBegin:
		cGestureEvent := (*C.clap_event_param_gesture_t)(unsafe.Pointer(cEventHeader))
		event.Data = ParamGestureEvent{
			ParamID: uint32(cGestureEvent.param_id),
		}
		
	case EventTypeParamGestureEnd:
		cGestureEvent := (*C.clap_event_param_gesture_t)(unsafe.Pointer(cEventHeader))
		event.Data = ParamGestureEvent{
			ParamID: uint32(cGestureEvent.param_id),
		}
		
	case EventTypeNoteOn:
		cNoteEvent := (*C.clap_event_note_t)(unsafe.Pointer(cEventHeader))
		event.Data = NoteEvent{
			NoteID:  int32(cNoteEvent.note_id),
			Port:    int16(cNoteEvent.port_index),
			Channel: int16(cNoteEvent.channel),
			Key:     int16(cNoteEvent.key),
			Value:   float64(cNoteEvent.velocity),
		}
		
	case EventTypeNoteOff:
		cNoteEvent := (*C.clap_event_note_t)(unsafe.Pointer(cEventHeader))
		event.Data = NoteEvent{
			NoteID:  int32(cNoteEvent.note_id),
			Port:    int16(cNoteEvent.port_index),
			Channel: int16(cNoteEvent.channel),
			Key:     int16(cNoteEvent.key),
			Value:   float64(cNoteEvent.velocity),
		}
		
	case EventTypeMIDI:
		cMidiEvent := (*C.clap_event_midi_t)(unsafe.Pointer(cEventHeader))
		// Create a Go slice from the C array
		midiData := make([]byte, 3)
		for i := 0; i < 3; i++ {
			midiData[i] = byte(cMidiEvent.data[i])
		}
		event.Data = MIDIEvent{
			Port: int16(cMidiEvent.port_index),
			Data: midiData,
		}
		
	// Add more event types as needed
	default:
		// Unknown event type, store raw data
		event.Data = nil
	}
	
	return event
}

// convertGoEventToC converts a Go Event struct to a C CLAP event
func convertGoEventToC(event *Event) *C.clap_event_header_t {
	if event == nil {
		return nil
	}
	
	switch event.Type {
	case EventTypeParamValue:
		if paramEvent, ok := event.Data.(ParamEvent); ok {
			cEvent := (*C.clap_event_param_value_t)(C.malloc(C.sizeof_clap_event_param_value_t))
			if cEvent == nil {
				return nil
			}
			
			cEvent.header.size = C.sizeof_clap_event_param_value_t
			cEvent.header.time = C.uint32_t(event.Time)
			cEvent.header._type = C.uint16_t(EventTypeParamValue)
			cEvent.header.flags = 0
			
			cEvent.param_id = C.clap_id(paramEvent.ParamID)
			cEvent.cookie = paramEvent.Cookie
			cEvent.note_id = C.int32_t(paramEvent.Note)
			cEvent.port_index = C.int16_t(paramEvent.Port)
			cEvent.channel = C.int16_t(paramEvent.Channel)
			cEvent.key = C.int16_t(paramEvent.Key)
			cEvent.value = C.double(paramEvent.Value)
			
			return &cEvent.header
		}
	
	case EventTypeNoteOn, EventTypeNoteOff:
		if noteEvent, ok := event.Data.(NoteEvent); ok {
			cEvent := (*C.clap_event_note_t)(C.malloc(C.sizeof_clap_event_note_t))
			if cEvent == nil {
				return nil
			}
			
			cEvent.header.size = C.sizeof_clap_event_note_t
			cEvent.header.time = C.uint32_t(event.Time)
			cEvent.header._type = C.uint16_t(event.Type)
			cEvent.header.flags = 0
			
			cEvent.note_id = C.int32_t(noteEvent.NoteID)
			cEvent.port_index = C.int16_t(noteEvent.Port)
			cEvent.channel = C.int16_t(noteEvent.Channel)
			cEvent.key = C.int16_t(noteEvent.Key)
			cEvent.velocity = C.double(noteEvent.Value)
			
			return &cEvent.header
		}
	
	// Add more event types as needed
	}
	
	return nil
}

// freeCEvent frees memory allocated for a C event
func freeCEvent(cEvent *C.clap_event_header_t) {
	if cEvent != nil {
		C.free(unsafe.Pointer(cEvent))
	}
}

// Additional event types for better type safety

// ParamGestureEvent represents parameter gesture begin/end events
type ParamGestureEvent struct {
	ParamID uint32
}

// MIDIEvent represents MIDI events
type MIDIEvent struct {
	Port int16
	Data []byte
}

// TransportEvent represents transport events
type TransportEvent struct {
	Flags         uint32
	SongPosBeats  uint64
	SongPosSeconds uint64
	Tempo         float64
	TsigNum       uint16
	TsigDenom     uint16
}

// Helper functions for common event operations

// CreateParamValueEvent creates a parameter value change event
func CreateParamValueEvent(time uint32, paramID uint32, value float64) *Event {
	return &Event{
		Time: time,
		Type: EventTypeParamValue,
		Data: ParamEvent{
			ParamID: paramID,
			Value:   value,
		},
	}
}

// CreateNoteOnEvent creates a note on event
func CreateNoteOnEvent(time uint32, channel, key int16, velocity float64) *Event {
	return &Event{
		Time: time,
		Type: EventTypeNoteOn,
		Data: NoteEvent{
			Channel: channel,
			Key:     key,
			Value:   velocity,
		},
	}
}

// CreateNoteOffEvent creates a note off event
func CreateNoteOffEvent(time uint32, channel, key int16, velocity float64) *Event {
	return &Event{
		Time: time,
		Type: EventTypeNoteOff,
		Data: NoteEvent{
			Channel: channel,
			Key:     key,
			Value:   velocity,
		},
	}
}

// CreateMIDIEvent creates a MIDI event
func CreateMIDIEvent(time uint32, port int16, data []byte) *Event {
	return &Event{
		Time: time,
		Type: EventTypeMIDI,
		Data: MIDIEvent{
			Port: port,
			Data: data,
		},
	}
}