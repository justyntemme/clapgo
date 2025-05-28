package event

// #include "../../include/clap/include/clap/clap.h"
// #include <stdlib.h>
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

// EventProcessor is an alias for Processor to maintain consistent naming
type EventProcessor = Processor

// NewEventProcessor creates a new EventProcessor from C event queues
func NewEventProcessor(inputEvents, outputEvents unsafe.Pointer) *EventProcessor {
	return NewProcessor(inputEvents, outputEvents)
}

// PushNoteEnd pushes a note end event to the output
func (p *Processor) PushNoteEnd(event *NoteEvent, time uint32) bool {
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
	cEvent.header._type = C.uint16_t(TypeNoteEnd)
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

// PushOutputEvent pushes a note end event to the output
func (p *Processor) PushOutputEvent(noteEvent *NoteEvent) bool {
	if noteEvent == nil {
		return false
	}
	return p.PushNoteEnd(noteEvent, 0) // time is 0 for immediate
}