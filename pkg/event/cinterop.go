package event

// #include "../../include/clap/include/clap/clap.h"
// #include <stdlib.h>
// #include <string.h>
//
// // Helper functions for CLAP event handling
// static inline uint32_t clap_input_events_size_helper(const clap_input_events_t* events) {
//     if (events && events->size) {
//         return events->size(events);
//     }
//     return 0;
// }
//
// static inline const clap_event_header_t* clap_input_events_get_helper(const clap_input_events_t* events, uint32_t index) {
//     if (events && events->get) {
//         return events->get(events, index);
//     }
//     return NULL;
// }
//
// static inline bool clap_output_events_try_push_helper(const clap_output_events_t* events, const clap_event_header_t* event) {
//     if (events && events->try_push) {
//         return events->try_push(events, event);
//     }
//     return false;
// }
import "C"
import (
	"unsafe"
)

// Processor handles all C event conversion internally
// This abstracts away all the unsafe pointer operations from plugin developers
type Processor struct {
	inputEvents  *C.clap_input_events_t
	outputEvents *C.clap_output_events_t
	pool         *Pool
}

// NewProcessor creates a new Processor from C event queues
func NewProcessor(inputEvents, outputEvents unsafe.Pointer) *Processor {
	return &Processor{
		inputEvents:  (*C.clap_input_events_t)(inputEvents),
		outputEvents: (*C.clap_output_events_t)(outputEvents),
		pool:         NewPool(),
	}
}

// SetPool allows setting a custom event pool (useful for sharing across processors)
func (p *Processor) SetPool(pool *Pool) {
	p.pool = pool
}

// GetPool returns the event pool for diagnostics
func (p *Processor) GetPool() *Pool {
	return p.pool
}

// GetInputEventCount returns the number of input events
func (p *Processor) GetInputEventCount() uint32 {
	if p.inputEvents == nil {
		return 0
	}
	
	return uint32(C.clap_input_events_size_helper(p.inputEvents))
}

// ProcessAll processes all input events and calls the appropriate handler
func (p *Processor) ProcessAll(handler Handler) {
	if handler == nil || p.inputEvents == nil {
		return
	}
	
	count := uint32(C.clap_input_events_size_helper(p.inputEvents))
	
	for i := uint32(0); i < count; i++ {
		cEventHeader := C.clap_input_events_get_helper(p.inputEvents, C.uint32_t(i))
		if cEventHeader == nil || cEventHeader.space_id != 0 {
			continue
		}
		
		time := uint32(cEventHeader.time)
		
		// Process events directly without creating Event wrapper
		switch int(cEventHeader._type) {
		case int(TypeParamValue):
			cParamEvent := (*C.clap_event_param_value_t)(unsafe.Pointer(cEventHeader))
			paramEvent := p.pool.GetParamValueEvent()
			paramEvent.ParamID = uint32(cParamEvent.param_id)
			paramEvent.Cookie = unsafe.Pointer(cParamEvent.cookie)
			paramEvent.NoteID = int32(cParamEvent.note_id)
			paramEvent.Port = int16(cParamEvent.port_index)
			paramEvent.Channel = int16(cParamEvent.channel)
			paramEvent.Key = int16(cParamEvent.key)
			paramEvent.Value = float64(cParamEvent.value)
			handler.HandleParamValue(paramEvent, time)
			p.pool.PutParamValueEvent(paramEvent)
			
		case int(TypeParamMod):
			cParamModEvent := (*C.clap_event_param_mod_t)(unsafe.Pointer(cEventHeader))
			paramModEvent := p.pool.GetParamModEvent()
			paramModEvent.ParamID = uint32(cParamModEvent.param_id)
			paramModEvent.Cookie = unsafe.Pointer(cParamModEvent.cookie)
			paramModEvent.NoteID = int32(cParamModEvent.note_id)
			paramModEvent.Port = int16(cParamModEvent.port_index)
			paramModEvent.Channel = int16(cParamModEvent.channel)
			paramModEvent.Key = int16(cParamModEvent.key)
			paramModEvent.Amount = float64(cParamModEvent.amount)
			handler.HandleParamMod(paramModEvent, time)
			p.pool.PutParamModEvent(paramModEvent)
			
		case int(TypeParamGestureBegin):
			cGestureEvent := (*C.clap_event_param_gesture_t)(unsafe.Pointer(cEventHeader))
			gestureEvent := p.pool.GetParamGestureEvent()
			gestureEvent.ParamID = uint32(cGestureEvent.param_id)
			handler.HandleParamGestureBegin(gestureEvent, time)
			p.pool.PutParamGestureEvent(gestureEvent)
			
		case int(TypeParamGestureEnd):
			cGestureEvent := (*C.clap_event_param_gesture_t)(unsafe.Pointer(cEventHeader))
			gestureEvent := p.pool.GetParamGestureEvent()
			gestureEvent.ParamID = uint32(cGestureEvent.param_id)
			handler.HandleParamGestureEnd(gestureEvent, time)
			p.pool.PutParamGestureEvent(gestureEvent)
			
		case int(TypeNoteOn):
			cNoteEvent := (*C.clap_event_note_t)(unsafe.Pointer(cEventHeader))
			noteEvent := p.pool.GetNoteEvent()
			noteEvent.NoteID = int32(cNoteEvent.note_id)
			noteEvent.Port = int16(cNoteEvent.port_index)
			noteEvent.Channel = int16(cNoteEvent.channel)
			noteEvent.Key = int16(cNoteEvent.key)
			noteEvent.Velocity = float64(cNoteEvent.velocity)
			handler.HandleNoteOn(noteEvent, time)
			p.pool.PutNoteEvent(noteEvent)
			
		case int(TypeNoteOff):
			cNoteEvent := (*C.clap_event_note_t)(unsafe.Pointer(cEventHeader))
			noteEvent := p.pool.GetNoteEvent()
			noteEvent.NoteID = int32(cNoteEvent.note_id)
			noteEvent.Port = int16(cNoteEvent.port_index)
			noteEvent.Channel = int16(cNoteEvent.channel)
			noteEvent.Key = int16(cNoteEvent.key)
			noteEvent.Velocity = float64(cNoteEvent.velocity)
			handler.HandleNoteOff(noteEvent, time)
			p.pool.PutNoteEvent(noteEvent)
			
		case int(TypeNoteChoke):
			cNoteEvent := (*C.clap_event_note_t)(unsafe.Pointer(cEventHeader))
			noteEvent := p.pool.GetNoteEvent()
			noteEvent.NoteID = int32(cNoteEvent.note_id)
			noteEvent.Port = int16(cNoteEvent.port_index)
			noteEvent.Channel = int16(cNoteEvent.channel)
			noteEvent.Key = int16(cNoteEvent.key)
			noteEvent.Velocity = float64(cNoteEvent.velocity)
			handler.HandleNoteChoke(noteEvent, time)
			p.pool.PutNoteEvent(noteEvent)
			
		case int(TypeNoteEnd):
			cNoteEvent := (*C.clap_event_note_t)(unsafe.Pointer(cEventHeader))
			noteEvent := p.pool.GetNoteEvent()
			noteEvent.NoteID = int32(cNoteEvent.note_id)
			noteEvent.Port = int16(cNoteEvent.port_index)
			noteEvent.Channel = int16(cNoteEvent.channel)
			noteEvent.Key = int16(cNoteEvent.key)
			noteEvent.Velocity = float64(cNoteEvent.velocity)
			handler.HandleNoteEnd(noteEvent, time)
			p.pool.PutNoteEvent(noteEvent)
			
		case int(TypeNoteExpression):
			cNoteExprEvent := (*C.clap_event_note_expression_t)(unsafe.Pointer(cEventHeader))
			noteExprEvent := p.pool.GetNoteExpressionEvent()
			noteExprEvent.ExpressionID = uint32(cNoteExprEvent.expression_id)
			noteExprEvent.NoteID = int32(cNoteExprEvent.note_id)
			noteExprEvent.Port = int16(cNoteExprEvent.port_index)
			noteExprEvent.Channel = int16(cNoteExprEvent.channel)
			noteExprEvent.Key = int16(cNoteExprEvent.key)
			noteExprEvent.Value = float64(cNoteExprEvent.value)
			handler.HandleNoteExpression(noteExprEvent, time)
			p.pool.PutNoteExpressionEvent(noteExprEvent)
			
		case int(TypeTransport):
			cTransportEvent := (*C.clap_event_transport_t)(unsafe.Pointer(cEventHeader))
			transportEvent := p.pool.GetTransportEvent()
			transportEvent.Flags = uint32(cTransportEvent.flags)
			transportEvent.SongPosBeats = float64(cTransportEvent.song_pos_beats)
			transportEvent.SongPosSeconds = float64(cTransportEvent.song_pos_seconds)
			transportEvent.Tempo = float64(cTransportEvent.tempo)
			transportEvent.TempoInc = float64(cTransportEvent.tempo_inc)
			transportEvent.LoopStartBeats = float64(cTransportEvent.loop_start_beats)
			transportEvent.LoopEndBeats = float64(cTransportEvent.loop_end_beats)
			transportEvent.LoopStartSeconds = float64(cTransportEvent.loop_start_seconds)
			transportEvent.LoopEndSeconds = float64(cTransportEvent.loop_end_seconds)
			transportEvent.BarStart = float64(cTransportEvent.bar_start)
			transportEvent.BarNumber = int32(cTransportEvent.bar_number)
			transportEvent.TimeSignatureNum = uint16(cTransportEvent.tsig_num)
			transportEvent.TimeSignatureDenom = uint16(cTransportEvent.tsig_denom)
			handler.HandleTransport(transportEvent, time)
			p.pool.PutTransportEvent(transportEvent)
			
		case int(TypeMIDI):
			cMidiEvent := (*C.clap_event_midi_t)(unsafe.Pointer(cEventHeader))
			midiEvent := p.pool.GetMIDIEvent()
			midiEvent.Port = uint16(cMidiEvent.port_index)
			for j := 0; j < 3; j++ {
				midiEvent.Data[j] = byte(cMidiEvent.data[j])
			}
			handler.HandleMIDI(midiEvent, time)
			p.pool.PutMIDIEvent(midiEvent)
			
		case int(TypeMIDISysex):
			cSysexEvent := (*C.clap_event_midi_sysex_t)(unsafe.Pointer(cEventHeader))
			sysexEvent := p.pool.GetMIDISysexEvent()
			sysexEvent.Port = uint16(cSysexEvent.port_index)
			sysexEvent.Buffer = unsafe.Pointer(cSysexEvent.buffer)
			sysexEvent.Size = uint32(cSysexEvent.size)
			handler.HandleMIDISysex(sysexEvent, time)
			p.pool.PutMIDISysexEvent(sysexEvent)
			
		case int(TypeMIDI2):
			cMidi2Event := (*C.clap_event_midi2_t)(unsafe.Pointer(cEventHeader))
			midi2Event := p.pool.GetMIDI2Event()
			midi2Event.Port = uint16(cMidi2Event.port_index)
			for j := 0; j < 4; j++ {
				midi2Event.Data[j] = uint32(cMidi2Event.data[j])
			}
			handler.HandleMIDI2(midi2Event, time)
			p.pool.PutMIDI2Event(midi2Event)
		}
	}
}

// PushParamValue pushes a parameter value event to the output
func (p *Processor) PushParamValue(event *ParamValueEvent, time uint32) bool {
	if p.outputEvents == nil {
		return false
	}
	
	cEvent := (*C.clap_event_param_value_t)(C.malloc(C.sizeof_clap_event_param_value_t))
	if cEvent == nil {
		return false
	}
	
	cEvent.header.size = C.sizeof_clap_event_param_value_t
	cEvent.header.time = C.uint32_t(time)
	cEvent.header.space_id = 0 // CLAP_CORE_EVENT_SPACE_ID
	cEvent.header._type = C.uint16_t(TypeParamValue)
	cEvent.header.flags = 0
	
	cEvent.param_id = C.clap_id(event.ParamID)
	cEvent.cookie = event.Cookie
	cEvent.note_id = C.int32_t(event.NoteID)
	cEvent.port_index = C.int16_t(event.Port)
	cEvent.channel = C.int16_t(event.Channel)
	cEvent.key = C.int16_t(event.Key)
	cEvent.value = C.double(event.Value)
	
	result := C.clap_output_events_try_push_helper(p.outputEvents, &cEvent.header)
	C.free(unsafe.Pointer(cEvent))
	
	return bool(result)
}

// PushNoteOn pushes a note on event to the output
func (p *Processor) PushNoteOn(event *NoteEvent, time uint32) bool {
	if p.outputEvents == nil {
		return false
	}
	
	cEvent := (*C.clap_event_note_t)(C.malloc(C.sizeof_clap_event_note_t))
	if cEvent == nil {
		return false
	}
	
	cEvent.header.size = C.sizeof_clap_event_note_t
	cEvent.header.time = C.uint32_t(time)
	cEvent.header.space_id = 0
	cEvent.header._type = C.uint16_t(TypeNoteOn)
	cEvent.header.flags = 0
	
	cEvent.note_id = C.int32_t(event.NoteID)
	cEvent.port_index = C.int16_t(event.Port)
	cEvent.channel = C.int16_t(event.Channel)
	cEvent.key = C.int16_t(event.Key)
	cEvent.velocity = C.double(event.Velocity)
	
	result := C.clap_output_events_try_push_helper(p.outputEvents, &cEvent.header)
	C.free(unsafe.Pointer(cEvent))
	
	return bool(result)
}

// PushNoteOff pushes a note off event to the output
func (p *Processor) PushNoteOff(event *NoteEvent, time uint32) bool {
	if p.outputEvents == nil {
		return false
	}
	
	cEvent := (*C.clap_event_note_t)(C.malloc(C.sizeof_clap_event_note_t))
	if cEvent == nil {
		return false
	}
	
	cEvent.header.size = C.sizeof_clap_event_note_t
	cEvent.header.time = C.uint32_t(time)
	cEvent.header.space_id = 0
	cEvent.header._type = C.uint16_t(TypeNoteOff)
	cEvent.header.flags = 0
	
	cEvent.note_id = C.int32_t(event.NoteID)
	cEvent.port_index = C.int16_t(event.Port)
	cEvent.channel = C.int16_t(event.Channel)
	cEvent.key = C.int16_t(event.Key)
	cEvent.velocity = C.double(event.Velocity)
	
	result := C.clap_output_events_try_push_helper(p.outputEvents, &cEvent.header)
	C.free(unsafe.Pointer(cEvent))
	
	return bool(result)
}