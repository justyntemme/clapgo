package api

/*
#cgo CFLAGS: -I../../include/clap/include
#include "../../include/clap/include/clap/clap.h"
#include <stdlib.h>
#include <string.h>

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

// ProcessAllEvents processes all input events and calls the appropriate handler
func (ep *EventProcessor) ProcessAllEvents(handler TypedEventHandler) {
	if handler == nil {
		return
	}
	
	count := ep.GetInputEventCount()
	for i := uint32(0); i < count; i++ {
		event := ep.GetInputEvent(i)
		if event == nil {
			continue
		}
		
		// Route to the appropriate handler based on event type
		switch event.Type {
		case EventTypeParamValue:
			if e, ok := event.Data.(ParamValueEvent); ok {
				handler.HandleParamValue(&e, event.Time)
			}
		case EventTypeParamMod:
			if e, ok := event.Data.(ParamModEvent); ok {
				handler.HandleParamMod(&e, event.Time)
			}
		case EventTypeParamGestureBegin:
			if e, ok := event.Data.(ParamGestureEvent); ok {
				handler.HandleParamGestureBegin(&e, event.Time)
			}
		case EventTypeParamGestureEnd:
			if e, ok := event.Data.(ParamGestureEvent); ok {
				handler.HandleParamGestureEnd(&e, event.Time)
			}
		case EventTypeNoteOn:
			if e, ok := event.Data.(NoteEvent); ok {
				handler.HandleNoteOn(&e, event.Time)
			}
		case EventTypeNoteOff:
			if e, ok := event.Data.(NoteEvent); ok {
				handler.HandleNoteOff(&e, event.Time)
			}
		case EventTypeNoteChoke:
			if e, ok := event.Data.(NoteEvent); ok {
				handler.HandleNoteChoke(&e, event.Time)
			}
		case EventTypeNoteEnd:
			if e, ok := event.Data.(NoteEvent); ok {
				handler.HandleNoteEnd(&e, event.Time)
			}
		case EventTypeNoteExpression:
			if e, ok := event.Data.(NoteExpressionEvent); ok {
				handler.HandleNoteExpression(&e, event.Time)
			}
		case EventTypeTransport:
			if e, ok := event.Data.(TransportEvent); ok {
				handler.HandleTransport(&e, event.Time)
			}
		case EventTypeMIDI:
			if e, ok := event.Data.(MIDIEvent); ok {
				handler.HandleMIDI(&e, event.Time)
			}
		case EventTypeMIDISysex:
			if e, ok := event.Data.(MIDISysexEvent); ok {
				handler.HandleMIDISysex(&e, event.Time)
			}
		case EventTypeMIDI2:
			if e, ok := event.Data.(MIDI2Event); ok {
				handler.HandleMIDI2(&e, event.Time)
			}
		}
	}
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
	
	// Only process events from the core namespace (0)
	if cEventHeader.space_id != 0 {
		return nil
	}
	
	event := &Event{
		Time:  uint32(cEventHeader.time),
		Type:  int(cEventHeader._type),
		Flags: uint32(cEventHeader.flags),
	}
	
	// Convert event data based on type
	switch int(cEventHeader._type) {
	case EventTypeParamValue:
		cParamEvent := (*C.clap_event_param_value_t)(unsafe.Pointer(cEventHeader))
		event.Data = ParamValueEvent{
			ParamID: uint32(cParamEvent.param_id),
			Cookie:  unsafe.Pointer(cParamEvent.cookie),
			NoteID:  int32(cParamEvent.note_id),
			Port:    int16(cParamEvent.port_index),
			Channel: int16(cParamEvent.channel),
			Key:     int16(cParamEvent.key),
			Value:   float64(cParamEvent.value),
		}
		
	case EventTypeParamMod:
		cParamModEvent := (*C.clap_event_param_mod_t)(unsafe.Pointer(cEventHeader))
		event.Data = ParamModEvent{
			ParamID: uint32(cParamModEvent.param_id),
			Cookie:  unsafe.Pointer(cParamModEvent.cookie),
			NoteID:  int32(cParamModEvent.note_id),
			Port:    int16(cParamModEvent.port_index),
			Channel: int16(cParamModEvent.channel),
			Key:     int16(cParamModEvent.key),
			Amount:  float64(cParamModEvent.amount),
		}
		
	case EventTypeParamGestureBegin, EventTypeParamGestureEnd:
		cGestureEvent := (*C.clap_event_param_gesture_t)(unsafe.Pointer(cEventHeader))
		event.Data = ParamGestureEvent{
			ParamID: uint32(cGestureEvent.param_id),
		}
		
	case EventTypeNoteOn, EventTypeNoteOff, EventTypeNoteChoke, EventTypeNoteEnd:
		cNoteEvent := (*C.clap_event_note_t)(unsafe.Pointer(cEventHeader))
		event.Data = NoteEvent{
			NoteID:   int32(cNoteEvent.note_id),
			Port:     int16(cNoteEvent.port_index),
			Channel:  int16(cNoteEvent.channel),
			Key:      int16(cNoteEvent.key),
			Velocity: float64(cNoteEvent.velocity),
		}
		
	case EventTypeNoteExpression:
		cNoteExprEvent := (*C.clap_event_note_expression_t)(unsafe.Pointer(cEventHeader))
		event.Data = NoteExpressionEvent{
			ExpressionID: int32(cNoteExprEvent.expression_id),
			NoteID:       int32(cNoteExprEvent.note_id),
			Port:         int16(cNoteExprEvent.port_index),
			Channel:      int16(cNoteExprEvent.channel),
			Key:          int16(cNoteExprEvent.key),
			Value:        float64(cNoteExprEvent.value),
		}
		
	case EventTypeTransport:
		cTransportEvent := (*C.clap_event_transport_t)(unsafe.Pointer(cEventHeader))
		event.Data = TransportEvent{
			Flags:             uint32(cTransportEvent.flags),
			SongPosBeats:      uint64(cTransportEvent.song_pos_beats),
			SongPosSeconds:    uint64(cTransportEvent.song_pos_seconds),
			Tempo:             float64(cTransportEvent.tempo),
			TempoInc:          float64(cTransportEvent.tempo_inc),
			LoopStartBeats:    uint64(cTransportEvent.loop_start_beats),
			LoopEndBeats:      uint64(cTransportEvent.loop_end_beats),
			LoopStartSeconds:  uint64(cTransportEvent.loop_start_seconds),
			LoopEndSeconds:    uint64(cTransportEvent.loop_end_seconds),
			BarStart:          uint64(cTransportEvent.bar_start),
			BarNumber:         int32(cTransportEvent.bar_number),
			TimeSignatureNum:  uint16(cTransportEvent.tsig_num),
			TimeSignatureDen:  uint16(cTransportEvent.tsig_denom),
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
		
	case EventTypeMIDISysex:
		cSysexEvent := (*C.clap_event_midi_sysex_t)(unsafe.Pointer(cEventHeader))
		// Copy sysex data to Go slice
		sysexData := make([]byte, cSysexEvent.size)
		if cSysexEvent.size > 0 && cSysexEvent.buffer != nil {
			C.memcpy(unsafe.Pointer(&sysexData[0]), unsafe.Pointer(cSysexEvent.buffer), C.size_t(cSysexEvent.size))
		}
		event.Data = MIDISysexEvent{
			Port: int16(cSysexEvent.port_index),
			Data: sysexData,
		}
		
	case EventTypeMIDI2:
		cMidi2Event := (*C.clap_event_midi2_t)(unsafe.Pointer(cEventHeader))
		// Create Go slice from C array
		midi2Data := make([]uint32, 4)
		for i := 0; i < 4; i++ {
			midi2Data[i] = uint32(cMidi2Event.data[i])
		}
		event.Data = MIDI2Event{
			Port: int16(cMidi2Event.port_index),
			Data: midi2Data,
		}
		
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
		if paramEvent, ok := event.Data.(ParamValueEvent); ok {
			cEvent := (*C.clap_event_param_value_t)(C.malloc(C.sizeof_clap_event_param_value_t))
			if cEvent == nil {
				return nil
			}
			
			cEvent.header.size = C.sizeof_clap_event_param_value_t
			cEvent.header.time = C.uint32_t(event.Time)
			cEvent.header.space_id = 0 // CLAP_CORE_EVENT_SPACE_ID
			cEvent.header._type = C.uint16_t(EventTypeParamValue)
			cEvent.header.flags = C.uint32_t(event.Flags)
			
			cEvent.param_id = C.clap_id(paramEvent.ParamID)
			cEvent.cookie = paramEvent.Cookie
			cEvent.note_id = C.int32_t(paramEvent.NoteID)
			cEvent.port_index = C.int16_t(paramEvent.Port)
			cEvent.channel = C.int16_t(paramEvent.Channel)
			cEvent.key = C.int16_t(paramEvent.Key)
			cEvent.value = C.double(paramEvent.Value)
			
			return &cEvent.header
		}
		
	case EventTypeParamMod:
		if paramModEvent, ok := event.Data.(ParamModEvent); ok {
			cEvent := (*C.clap_event_param_mod_t)(C.malloc(C.sizeof_clap_event_param_mod_t))
			if cEvent == nil {
				return nil
			}
			
			cEvent.header.size = C.sizeof_clap_event_param_mod_t
			cEvent.header.time = C.uint32_t(event.Time)
			cEvent.header.space_id = 0 // CLAP_CORE_EVENT_SPACE_ID
			cEvent.header._type = C.uint16_t(EventTypeParamMod)
			cEvent.header.flags = C.uint32_t(event.Flags)
			
			cEvent.param_id = C.clap_id(paramModEvent.ParamID)
			cEvent.cookie = paramModEvent.Cookie
			cEvent.note_id = C.int32_t(paramModEvent.NoteID)
			cEvent.port_index = C.int16_t(paramModEvent.Port)
			cEvent.channel = C.int16_t(paramModEvent.Channel)
			cEvent.key = C.int16_t(paramModEvent.Key)
			cEvent.amount = C.double(paramModEvent.Amount)
			
			return &cEvent.header
		}
		
	case EventTypeParamGestureBegin, EventTypeParamGestureEnd:
		if gestureEvent, ok := event.Data.(ParamGestureEvent); ok {
			cEvent := (*C.clap_event_param_gesture_t)(C.malloc(C.sizeof_clap_event_param_gesture_t))
			if cEvent == nil {
				return nil
			}
			
			cEvent.header.size = C.sizeof_clap_event_param_gesture_t
			cEvent.header.time = C.uint32_t(event.Time)
			cEvent.header.space_id = 0 // CLAP_CORE_EVENT_SPACE_ID
			cEvent.header._type = C.uint16_t(event.Type)
			cEvent.header.flags = C.uint32_t(event.Flags)
			
			cEvent.param_id = C.clap_id(gestureEvent.ParamID)
			
			return &cEvent.header
		}
	
	case EventTypeNoteOn, EventTypeNoteOff, EventTypeNoteChoke, EventTypeNoteEnd:
		if noteEvent, ok := event.Data.(NoteEvent); ok {
			cEvent := (*C.clap_event_note_t)(C.malloc(C.sizeof_clap_event_note_t))
			if cEvent == nil {
				return nil
			}
			
			cEvent.header.size = C.sizeof_clap_event_note_t
			cEvent.header.time = C.uint32_t(event.Time)
			cEvent.header.space_id = 0 // CLAP_CORE_EVENT_SPACE_ID
			cEvent.header._type = C.uint16_t(event.Type)
			cEvent.header.flags = C.uint32_t(event.Flags)
			
			cEvent.note_id = C.int32_t(noteEvent.NoteID)
			cEvent.port_index = C.int16_t(noteEvent.Port)
			cEvent.channel = C.int16_t(noteEvent.Channel)
			cEvent.key = C.int16_t(noteEvent.Key)
			cEvent.velocity = C.double(noteEvent.Velocity)
			
			return &cEvent.header
		}
		
	case EventTypeNoteExpression:
		if noteExprEvent, ok := event.Data.(NoteExpressionEvent); ok {
			cEvent := (*C.clap_event_note_expression_t)(C.malloc(C.sizeof_clap_event_note_expression_t))
			if cEvent == nil {
				return nil
			}
			
			cEvent.header.size = C.sizeof_clap_event_note_expression_t
			cEvent.header.time = C.uint32_t(event.Time)
			cEvent.header.space_id = 0 // CLAP_CORE_EVENT_SPACE_ID
			cEvent.header._type = C.uint16_t(EventTypeNoteExpression)
			cEvent.header.flags = C.uint32_t(event.Flags)
			
			cEvent.expression_id = C.clap_note_expression(noteExprEvent.ExpressionID)
			cEvent.note_id = C.int32_t(noteExprEvent.NoteID)
			cEvent.port_index = C.int16_t(noteExprEvent.Port)
			cEvent.channel = C.int16_t(noteExprEvent.Channel)
			cEvent.key = C.int16_t(noteExprEvent.Key)
			cEvent.value = C.double(noteExprEvent.Value)
			
			return &cEvent.header
		}
		
	case EventTypeTransport:
		if transportEvent, ok := event.Data.(TransportEvent); ok {
			cEvent := (*C.clap_event_transport_t)(C.malloc(C.sizeof_clap_event_transport_t))
			if cEvent == nil {
				return nil
			}
			
			cEvent.header.size = C.sizeof_clap_event_transport_t
			cEvent.header.time = C.uint32_t(event.Time)
			cEvent.header.space_id = 0 // CLAP_CORE_EVENT_SPACE_ID
			cEvent.header._type = C.uint16_t(EventTypeTransport)
			cEvent.header.flags = C.uint32_t(event.Flags)
			
			cEvent.flags = C.uint32_t(transportEvent.Flags)
			cEvent.song_pos_beats = C.clap_beattime(transportEvent.SongPosBeats)
			cEvent.song_pos_seconds = C.clap_sectime(transportEvent.SongPosSeconds)
			cEvent.tempo = C.double(transportEvent.Tempo)
			cEvent.tempo_inc = C.double(transportEvent.TempoInc)
			cEvent.loop_start_beats = C.clap_beattime(transportEvent.LoopStartBeats)
			cEvent.loop_end_beats = C.clap_beattime(transportEvent.LoopEndBeats)
			cEvent.loop_start_seconds = C.clap_sectime(transportEvent.LoopStartSeconds)
			cEvent.loop_end_seconds = C.clap_sectime(transportEvent.LoopEndSeconds)
			cEvent.bar_start = C.clap_beattime(transportEvent.BarStart)
			cEvent.bar_number = C.int32_t(transportEvent.BarNumber)
			cEvent.tsig_num = C.uint16_t(transportEvent.TimeSignatureNum)
			cEvent.tsig_denom = C.uint16_t(transportEvent.TimeSignatureDen)
			
			return &cEvent.header
		}
		
	case EventTypeMIDI:
		if midiEvent, ok := event.Data.(MIDIEvent); ok {
			cEvent := (*C.clap_event_midi_t)(C.malloc(C.sizeof_clap_event_midi_t))
			if cEvent == nil {
				return nil
			}
			
			cEvent.header.size = C.sizeof_clap_event_midi_t
			cEvent.header.time = C.uint32_t(event.Time)
			cEvent.header.space_id = 0 // CLAP_CORE_EVENT_SPACE_ID
			cEvent.header._type = C.uint16_t(EventTypeMIDI)
			cEvent.header.flags = C.uint32_t(event.Flags)
			
			cEvent.port_index = C.uint16_t(midiEvent.Port)
			for i := 0; i < 3 && i < len(midiEvent.Data); i++ {
				cEvent.data[i] = C.uint8_t(midiEvent.Data[i])
			}
			
			return &cEvent.header
		}
		
	case EventTypeMIDISysex:
		if sysexEvent, ok := event.Data.(MIDISysexEvent); ok {
			cEvent := (*C.clap_event_midi_sysex_t)(C.malloc(C.sizeof_clap_event_midi_sysex_t))
			if cEvent == nil {
				return nil
			}
			
			cEvent.header.size = C.sizeof_clap_event_midi_sysex_t
			cEvent.header.time = C.uint32_t(event.Time)
			cEvent.header.space_id = 0 // CLAP_CORE_EVENT_SPACE_ID
			cEvent.header._type = C.uint16_t(EventTypeMIDISysex)
			cEvent.header.flags = C.uint32_t(event.Flags)
			
			cEvent.port_index = C.uint16_t(sysexEvent.Port)
			cEvent.size = C.uint32_t(len(sysexEvent.Data))
			if len(sysexEvent.Data) > 0 {
				// Allocate buffer for sysex data
				cEvent.buffer = (*C.uint8_t)(C.malloc(C.size_t(len(sysexEvent.Data))))
				C.memcpy(unsafe.Pointer(cEvent.buffer), unsafe.Pointer(&sysexEvent.Data[0]), C.size_t(len(sysexEvent.Data)))
			}
			
			return &cEvent.header
		}
		
	case EventTypeMIDI2:
		if midi2Event, ok := event.Data.(MIDI2Event); ok {
			cEvent := (*C.clap_event_midi2_t)(C.malloc(C.sizeof_clap_event_midi2_t))
			if cEvent == nil {
				return nil
			}
			
			cEvent.header.size = C.sizeof_clap_event_midi2_t
			cEvent.header.time = C.uint32_t(event.Time)
			cEvent.header.space_id = 0 // CLAP_CORE_EVENT_SPACE_ID
			cEvent.header._type = C.uint16_t(EventTypeMIDI2)
			cEvent.header.flags = C.uint32_t(event.Flags)
			
			cEvent.port_index = C.uint16_t(midi2Event.Port)
			for i := 0; i < 4 && i < len(midi2Event.Data); i++ {
				cEvent.data[i] = C.uint32_t(midi2Event.Data[i])
			}
			
			return &cEvent.header
		}
	}
	
	return nil
}

// freeCEvent frees memory allocated for a C event
func freeCEvent(cEvent *C.clap_event_header_t) {
	if cEvent != nil {
		// Special handling for sysex events to free the buffer
		if cEvent._type == C.uint16_t(EventTypeMIDISysex) {
			sysexEvent := (*C.clap_event_midi_sysex_t)(unsafe.Pointer(cEvent))
			if sysexEvent.buffer != nil {
				C.free(unsafe.Pointer(sysexEvent.buffer))
			}
		}
		C.free(unsafe.Pointer(cEvent))
	}
}

// Additional event types for better type safety

// ParamGestureEvent represents parameter gesture begin/end events
type ParamGestureEvent struct {
	ParamID uint32
}

// MIDIEvent represents MIDI 1.0 events
type MIDIEvent struct {
	// Port is the port index
	Port int16
	
	// Data contains the 3-byte MIDI message
	Data []byte
}

// MIDISysexEvent represents MIDI system exclusive events
type MIDISysexEvent struct {
	// Port is the port index
	Port int16
	
	// Data contains the sysex message (without F0/F7 delimiters)
	Data []byte
}

// MIDI2Event represents MIDI 2.0 events
type MIDI2Event struct {
	// Port is the port index
	Port int16
	
	// Data contains the 4 32-bit words of MIDI 2.0 message
	Data []uint32
}

// TransportEvent represents transport events
type TransportEvent struct {
	// Flags indicates which fields are valid (see Transport* constants)
	Flags            uint32
	
	// Song position in beats (valid if TransportHasBeatsTime flag is set)
	SongPosBeats     uint64
	
	// Song position in seconds (valid if TransportHasSecondsTime flag is set)
	SongPosSeconds   uint64
	
	// Tempo in BPM (valid if TransportHasTempo flag is set)
	Tempo            float64
	
	// Tempo increment per sample
	TempoInc         float64
	
	// Loop start position in beats
	LoopStartBeats   uint64
	
	// Loop end position in beats
	LoopEndBeats     uint64
	
	// Loop start position in seconds
	LoopStartSeconds uint64
	
	// Loop end position in seconds
	LoopEndSeconds   uint64
	
	// Bar start position in beats
	BarStart         uint64
	
	// Bar number at song position 0
	BarNumber        int32
	
	// Time signature numerator (valid if TransportHasTimeSignature flag is set)
	TimeSignatureNum uint16
	
	// Time signature denominator (valid if TransportHasTimeSignature flag is set)
	TimeSignatureDen uint16
}

// Helper functions for common event operations

// CreateParamValueEvent creates a parameter value change event
func CreateParamValueEvent(time uint32, paramID uint32, value float64) *Event {
	return &Event{
		Time: time,
		Type: EventTypeParamValue,
		Data: ParamValueEvent{
			ParamID: paramID,
			NoteID:  -1, // Global parameter
			Port:    -1,
			Channel: -1,
			Key:     -1,
			Value:   value,
		},
	}
}

// CreatePolyParamValueEvent creates a polyphonic parameter value change event
func CreatePolyParamValueEvent(time uint32, paramID uint32, noteID int32, port, channel, key int16, value float64) *Event {
	return &Event{
		Time: time,
		Type: EventTypeParamValue,
		Data: ParamValueEvent{
			ParamID: paramID,
			NoteID:  noteID,
			Port:    port,
			Channel: channel,
			Key:     key,
			Value:   value,
		},
	}
}

// CreateParamModEvent creates a parameter modulation event
func CreateParamModEvent(time uint32, paramID uint32, amount float64) *Event {
	return &Event{
		Time: time,
		Type: EventTypeParamMod,
		Data: ParamModEvent{
			ParamID: paramID,
			NoteID:  -1, // Global modulation
			Port:    -1,
			Channel: -1,
			Key:     -1,
			Amount:  amount,
		},
	}
}

// CreateParamGestureBeginEvent creates a parameter gesture begin event
func CreateParamGestureBeginEvent(time uint32, paramID uint32) *Event {
	return &Event{
		Time: time,
		Type: EventTypeParamGestureBegin,
		Data: ParamGestureEvent{
			ParamID: paramID,
		},
	}
}

// CreateParamGestureEndEvent creates a parameter gesture end event
func CreateParamGestureEndEvent(time uint32, paramID uint32) *Event {
	return &Event{
		Time: time,
		Type: EventTypeParamGestureEnd,
		Data: ParamGestureEvent{
			ParamID: paramID,
		},
	}
}

// CreateNoteOnEvent creates a note on event
func CreateNoteOnEvent(time uint32, port, channel, key int16, velocity float64) *Event {
	return &Event{
		Time: time,
		Type: EventTypeNoteOn,
		Data: NoteEvent{
			NoteID:   -1, // Host will assign
			Port:     port,
			Channel:  channel,
			Key:      key,
			Velocity: velocity,
		},
	}
}

// CreateNoteOnEventWithID creates a note on event with a specific note ID
func CreateNoteOnEventWithID(time uint32, noteID int32, port, channel, key int16, velocity float64) *Event {
	return &Event{
		Time: time,
		Type: EventTypeNoteOn,
		Data: NoteEvent{
			NoteID:   noteID,
			Port:     port,
			Channel:  channel,
			Key:      key,
			Velocity: velocity,
		},
	}
}

// CreateNoteOffEvent creates a note off event
func CreateNoteOffEvent(time uint32, port, channel, key int16, velocity float64) *Event {
	return &Event{
		Time: time,
		Type: EventTypeNoteOff,
		Data: NoteEvent{
			NoteID:   -1, // Match any note with this key
			Port:     port,
			Channel:  channel,
			Key:      key,
			Velocity: velocity,
		},
	}
}

// CreateNoteChokeEvent creates a note choke event
func CreateNoteChokeEvent(time uint32, port, channel, key int16) *Event {
	return &Event{
		Time: time,
		Type: EventTypeNoteChoke,
		Data: NoteEvent{
			NoteID:   -1,
			Port:     port,
			Channel:  channel,
			Key:      key,
			Velocity: 0, // Ignored for choke
		},
	}
}

// CreateNoteEndEvent creates a note end event (sent by plugin to host)
func CreateNoteEndEvent(time uint32, noteID int32, port, channel, key int16) *Event {
	return &Event{
		Time: time,
		Type: EventTypeNoteEnd,
		Data: NoteEvent{
			NoteID:   noteID,
			Port:     port,
			Channel:  channel,
			Key:      key,
			Velocity: 0, // Ignored for end
		},
	}
}

// CreateNoteExpressionEvent creates a note expression event
func CreateNoteExpressionEvent(time uint32, expressionID int32, noteID int32, port, channel, key int16, value float64) *Event {
	return &Event{
		Time: time,
		Type: EventTypeNoteExpression,
		Data: NoteExpressionEvent{
			ExpressionID: expressionID,
			NoteID:       noteID,
			Port:         port,
			Channel:      channel,
			Key:          key,
			Value:        value,
		},
	}
}

// CreateMIDIEvent creates a MIDI 1.0 event
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

// CreateMIDISysexEvent creates a MIDI system exclusive event
func CreateMIDISysexEvent(time uint32, port int16, data []byte) *Event {
	return &Event{
		Time: time,
		Type: EventTypeMIDISysex,
		Data: MIDISysexEvent{
			Port: port,
			Data: data,
		},
	}
}

// CreateMIDI2Event creates a MIDI 2.0 event
func CreateMIDI2Event(time uint32, port int16, data []uint32) *Event {
	return &Event{
		Time: time,
		Type: EventTypeMIDI2,
		Data: MIDI2Event{
			Port: port,
			Data: data,
		},
	}
}