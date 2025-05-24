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
	"log"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Logger interface for event pool logging
type Logger interface {
	Log(severity int32, message string, args ...interface{})
}

// EventPool manages pre-allocated events to avoid allocations during audio processing
type EventPool struct {
	// Separate pools for each event type to avoid interface{} allocations
	paramValuePool      sync.Pool
	paramModPool        sync.Pool
	paramGesturePool    sync.Pool
	noteEventPool       sync.Pool
	noteExpressionPool  sync.Pool
	transportPool       sync.Pool
	midiPool            sync.Pool
	midiSysexPool       sync.Pool
	midi2Pool           sync.Pool
	eventPool           sync.Pool // For Event structs themselves

	// Diagnostics
	totalAllocations    uint64
	poolHits           uint64
	poolMisses         uint64
	highWaterMark      uint64
	currentAllocated   uint64
	
	// Configuration
	warnOnAllocation   bool
	logger             Logger
}

// NewEventPool creates a new event pool with default settings
func NewEventPool() *EventPool {
	ep := &EventPool{
		warnOnAllocation: true,
	}

	// Initialize pools with factory functions
	ep.paramValuePool.New = func() interface{} {
		atomic.AddUint64(&ep.totalAllocations, 1)
		atomic.AddUint64(&ep.poolMisses, 1)
		if ep.warnOnAllocation && ep.logger != nil {
			ep.logger.Log(LogSeverityWarning, "EventPool: Allocating new ParamValueEvent - consider increasing pool size")
		}
		return &ParamValueEvent{}
	}

	ep.paramModPool.New = func() interface{} {
		atomic.AddUint64(&ep.totalAllocations, 1)
		atomic.AddUint64(&ep.poolMisses, 1)
		if ep.warnOnAllocation && ep.logger != nil {
			ep.logger.Log(LogSeverityWarning, "EventPool: Allocating new ParamModEvent - consider increasing pool size")
		}
		return &ParamModEvent{}
	}

	ep.paramGesturePool.New = func() interface{} {
		atomic.AddUint64(&ep.totalAllocations, 1)
		atomic.AddUint64(&ep.poolMisses, 1)
		if ep.warnOnAllocation && ep.logger != nil {
			ep.logger.Log(LogSeverityWarning, "EventPool: Allocating new ParamGestureEvent - consider increasing pool size")
		}
		return &ParamGestureEvent{}
	}

	ep.noteEventPool.New = func() interface{} {
		atomic.AddUint64(&ep.totalAllocations, 1)
		atomic.AddUint64(&ep.poolMisses, 1)
		if ep.warnOnAllocation && ep.logger != nil {
			ep.logger.Log(LogSeverityWarning, "EventPool: Allocating new NoteEvent - consider increasing pool size")
		}
		return &NoteEvent{}
	}

	ep.noteExpressionPool.New = func() interface{} {
		atomic.AddUint64(&ep.totalAllocations, 1)
		atomic.AddUint64(&ep.poolMisses, 1)
		if ep.warnOnAllocation && ep.logger != nil {
			ep.logger.Log(LogSeverityWarning, "EventPool: Allocating new NoteExpressionEvent - consider increasing pool size")
		}
		return &NoteExpressionEvent{}
	}

	ep.transportPool.New = func() interface{} {
		atomic.AddUint64(&ep.totalAllocations, 1)
		atomic.AddUint64(&ep.poolMisses, 1)
		if ep.warnOnAllocation && ep.logger != nil {
			ep.logger.Log(LogSeverityWarning, "EventPool: Allocating new TransportEvent - consider increasing pool size")
		}
		return &TransportEvent{}
	}

	ep.midiPool.New = func() interface{} {
		atomic.AddUint64(&ep.totalAllocations, 1)
		atomic.AddUint64(&ep.poolMisses, 1)
		if ep.warnOnAllocation && ep.logger != nil {
			ep.logger.Log(LogSeverityWarning, "EventPool: Allocating new MIDIEvent - consider increasing pool size")
		}
		return &MIDIEvent{Data: make([]byte, 3)}
	}

	ep.midiSysexPool.New = func() interface{} {
		atomic.AddUint64(&ep.totalAllocations, 1)
		atomic.AddUint64(&ep.poolMisses, 1)
		if ep.warnOnAllocation && ep.logger != nil {
			ep.logger.Log(LogSeverityWarning, "EventPool: Allocating new MIDISysexEvent - consider increasing pool size")
		}
		return &MIDISysexEvent{}
	}

	ep.midi2Pool.New = func() interface{} {
		atomic.AddUint64(&ep.totalAllocations, 1)
		atomic.AddUint64(&ep.poolMisses, 1)
		if ep.warnOnAllocation && ep.logger != nil {
			ep.logger.Log(LogSeverityWarning, "EventPool: Allocating new MIDI2Event - consider increasing pool size")
		}
		return &MIDI2Event{Data: make([]uint32, 4)}
	}

	ep.eventPool.New = func() interface{} {
		atomic.AddUint64(&ep.totalAllocations, 1)
		atomic.AddUint64(&ep.poolMisses, 1)
		if ep.warnOnAllocation && ep.logger != nil {
			ep.logger.Log(LogSeverityWarning, "EventPool: Allocating new Event - consider increasing pool size")
		}
		return &Event{}
	}

	// Pre-allocate initial events
	ep.preallocate()

	return ep
}

// SetLogger sets the logger for warnings
func (ep *EventPool) SetLogger(logger Logger) {
	ep.logger = logger
}

// preallocate creates initial pool of events
func (ep *EventPool) preallocate() {
	// Pre-allocate common event counts
	const (
		initialParamEvents     = 128
		initialNoteEvents      = 64
		initialOtherEvents     = 32
	)

	// Pre-allocate parameter events
	for i := 0; i < initialParamEvents; i++ {
		ep.paramValuePool.Put(&ParamValueEvent{})
		ep.paramModPool.Put(&ParamModEvent{})
		ep.eventPool.Put(&Event{})
	}

	// Pre-allocate note events
	for i := 0; i < initialNoteEvents; i++ {
		ep.noteEventPool.Put(&NoteEvent{})
		ep.noteExpressionPool.Put(&NoteExpressionEvent{})
		ep.eventPool.Put(&Event{})
	}

	// Pre-allocate other events
	for i := 0; i < initialOtherEvents; i++ {
		ep.paramGesturePool.Put(&ParamGestureEvent{})
		ep.transportPool.Put(&TransportEvent{})
		ep.midiPool.Put(&MIDIEvent{Data: make([]byte, 3)})
		ep.midiSysexPool.Put(&MIDISysexEvent{})
		ep.midi2Pool.Put(&MIDI2Event{Data: make([]uint32, 4)})
		ep.eventPool.Put(&Event{})
	}
}

// GetEvent gets an Event from the pool
func (ep *EventPool) GetEvent() *Event {
	atomic.AddUint64(&ep.currentAllocated, 1)
	current := atomic.LoadUint64(&ep.currentAllocated)
	for {
		oldMax := atomic.LoadUint64(&ep.highWaterMark)
		if current <= oldMax || atomic.CompareAndSwapUint64(&ep.highWaterMark, oldMax, current) {
			break
		}
	}
	
	event := ep.eventPool.Get().(*Event)
	atomic.AddUint64(&ep.poolHits, 1)
	return event
}

// GetParamValueEvent gets a ParamValueEvent from the pool
func (ep *EventPool) GetParamValueEvent() *ParamValueEvent {
	event := ep.paramValuePool.Get().(*ParamValueEvent)
	atomic.AddUint64(&ep.poolHits, 1)
	return event
}

// GetParamModEvent gets a ParamModEvent from the pool
func (ep *EventPool) GetParamModEvent() *ParamModEvent {
	event := ep.paramModPool.Get().(*ParamModEvent)
	atomic.AddUint64(&ep.poolHits, 1)
	return event
}

// GetParamGestureEvent gets a ParamGestureEvent from the pool
func (ep *EventPool) GetParamGestureEvent() *ParamGestureEvent {
	event := ep.paramGesturePool.Get().(*ParamGestureEvent)
	atomic.AddUint64(&ep.poolHits, 1)
	return event
}

// GetNoteEvent gets a NoteEvent from the pool
func (ep *EventPool) GetNoteEvent() *NoteEvent {
	event := ep.noteEventPool.Get().(*NoteEvent)
	atomic.AddUint64(&ep.poolHits, 1)
	return event
}

// GetNoteExpressionEvent gets a NoteExpressionEvent from the pool
func (ep *EventPool) GetNoteExpressionEvent() *NoteExpressionEvent {
	event := ep.noteExpressionPool.Get().(*NoteExpressionEvent)
	atomic.AddUint64(&ep.poolHits, 1)
	return event
}

// GetTransportEvent gets a TransportEvent from the pool
func (ep *EventPool) GetTransportEvent() *TransportEvent {
	event := ep.transportPool.Get().(*TransportEvent)
	atomic.AddUint64(&ep.poolHits, 1)
	return event
}

// GetMIDIEvent gets a MIDIEvent from the pool
func (ep *EventPool) GetMIDIEvent() *MIDIEvent {
	event := ep.midiPool.Get().(*MIDIEvent)
	atomic.AddUint64(&ep.poolHits, 1)
	return event
}

// GetMIDISysexEvent gets a MIDISysexEvent from the pool
func (ep *EventPool) GetMIDISysexEvent() *MIDISysexEvent {
	event := ep.midiSysexPool.Get().(*MIDISysexEvent)
	atomic.AddUint64(&ep.poolHits, 1)
	return event
}

// GetMIDI2Event gets a MIDI2Event from the pool
func (ep *EventPool) GetMIDI2Event() *MIDI2Event {
	event := ep.midi2Pool.Get().(*MIDI2Event)
	atomic.AddUint64(&ep.poolHits, 1)
	return event
}

// ReturnEvent returns an Event to the pool
func (ep *EventPool) ReturnEvent(event *Event) {
	if event == nil {
		return
	}
	
	atomic.AddUint64(&ep.currentAllocated, ^uint64(0)) // Decrement by 1
	
	// Clear the event
	event.Time = 0
	event.Type = 0
	event.Flags = 0
	event.Data = nil
	
	ep.eventPool.Put(event)
}

// ReturnParamValueEvent returns a ParamValueEvent to the pool
func (ep *EventPool) ReturnParamValueEvent(event *ParamValueEvent) {
	if event == nil {
		return
	}
	
	// Clear the event
	*event = ParamValueEvent{
		NoteID:  -1,
		Port:    -1,
		Channel: -1,
		Key:     -1,
	}
	
	ep.paramValuePool.Put(event)
}

// ReturnParamModEvent returns a ParamModEvent to the pool
func (ep *EventPool) ReturnParamModEvent(event *ParamModEvent) {
	if event == nil {
		return
	}
	
	// Clear the event
	*event = ParamModEvent{
		NoteID:  -1,
		Port:    -1,
		Channel: -1,
		Key:     -1,
	}
	
	ep.paramModPool.Put(event)
}

// ReturnParamGestureEvent returns a ParamGestureEvent to the pool
func (ep *EventPool) ReturnParamGestureEvent(event *ParamGestureEvent) {
	if event == nil {
		return
	}
	
	// Clear the event
	*event = ParamGestureEvent{}
	
	ep.paramGesturePool.Put(event)
}

// ReturnNoteEvent returns a NoteEvent to the pool
func (ep *EventPool) ReturnNoteEvent(event *NoteEvent) {
	if event == nil {
		return
	}
	
	// Clear the event
	*event = NoteEvent{
		NoteID: -1,
	}
	
	ep.noteEventPool.Put(event)
}

// ReturnNoteExpressionEvent returns a NoteExpressionEvent to the pool
func (ep *EventPool) ReturnNoteExpressionEvent(event *NoteExpressionEvent) {
	if event == nil {
		return
	}
	
	// Clear the event
	*event = NoteExpressionEvent{
		NoteID: -1,
	}
	
	ep.noteExpressionPool.Put(event)
}

// ReturnTransportEvent returns a TransportEvent to the pool
func (ep *EventPool) ReturnTransportEvent(event *TransportEvent) {
	if event == nil {
		return
	}
	
	// Clear the event
	*event = TransportEvent{}
	
	ep.transportPool.Put(event)
}

// ReturnMIDIEvent returns a MIDIEvent to the pool
func (ep *EventPool) ReturnMIDIEvent(event *MIDIEvent) {
	if event == nil {
		return
	}
	
	// Clear the event data but keep the slice
	event.Port = 0
	for i := range event.Data {
		event.Data[i] = 0
	}
	
	ep.midiPool.Put(event)
}

// ReturnMIDISysexEvent returns a MIDISysexEvent to the pool
func (ep *EventPool) ReturnMIDISysexEvent(event *MIDISysexEvent) {
	if event == nil {
		return
	}
	
	// Clear the event
	event.Port = 0
	event.Data = nil // Let GC handle the old data
	
	ep.midiSysexPool.Put(event)
}

// ReturnMIDI2Event returns a MIDI2Event to the pool
func (ep *EventPool) ReturnMIDI2Event(event *MIDI2Event) {
	if event == nil {
		return
	}
	
	// Clear the event data but keep the slice
	event.Port = 0
	for i := range event.Data {
		event.Data[i] = 0
	}
	
	ep.midi2Pool.Put(event)
}

// GetDiagnostics returns pool performance statistics
func (ep *EventPool) GetDiagnostics() (totalAllocations, poolHits, poolMisses, highWaterMark, currentAllocated uint64) {
	return atomic.LoadUint64(&ep.totalAllocations),
		atomic.LoadUint64(&ep.poolHits),
		atomic.LoadUint64(&ep.poolMisses),
		atomic.LoadUint64(&ep.highWaterMark),
		atomic.LoadUint64(&ep.currentAllocated)
}

// LogDiagnostics logs current pool statistics
func (ep *EventPool) LogDiagnostics() {
	totalAlloc, hits, misses, hwm, current := ep.GetDiagnostics()
	
	hitRate := float64(0)
	if hits+misses > 0 {
		hitRate = float64(hits) / float64(hits+misses) * 100
	}
	
	msg := "EventPool diagnostics:"
	if ep.logger != nil {
		ep.logger.Log(LogSeverityInfo, msg)
		ep.logger.Log(LogSeverityInfo, "  Total allocations: %d", totalAlloc)
		ep.logger.Log(LogSeverityInfo, "  Pool hits: %d", hits)
		ep.logger.Log(LogSeverityInfo, "  Pool misses: %d", misses)
		ep.logger.Log(LogSeverityInfo, "  Hit rate: %.2f%%", hitRate)
		ep.logger.Log(LogSeverityInfo, "  High water mark: %d", hwm)
		ep.logger.Log(LogSeverityInfo, "  Currently allocated: %d", current)
	} else {
		log.Printf("%s Total allocations: %d, Pool hits: %d, Pool misses: %d, Hit rate: %.2f%%, High water mark: %d, Currently allocated: %d",
			msg, totalAlloc, hits, misses, hitRate, hwm, current)
	}
}

// EventProcessor handles all C event conversion internally
// This abstracts away all the unsafe pointer operations from plugin developers
type EventProcessor struct {
	inputEvents  *C.clap_input_events_t
	outputEvents *C.clap_output_events_t
	eventPool    *EventPool
}

// NewEventProcessor creates a new EventProcessor from C event queues
func NewEventProcessor(inputEvents, outputEvents unsafe.Pointer) *EventProcessor {
	return &EventProcessor{
		inputEvents:  (*C.clap_input_events_t)(inputEvents),
		outputEvents: (*C.clap_output_events_t)(outputEvents),
		eventPool:    NewEventPool(),
	}
}

// SetEventPool allows setting a custom event pool (useful for sharing across processors)
func (ep *EventProcessor) SetEventPool(pool *EventPool) {
	ep.eventPool = pool
}

// GetEventPool returns the event pool for diagnostics
func (ep *EventProcessor) GetEventPool() *EventPool {
	return ep.eventPool
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
	
	return convertCEventToGo(cEventHeader, ep.eventPool)
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

// ReturnEventToPool returns an event and its data to the pool
func (ep *EventProcessor) ReturnEventToPool(event *Event) {
	if event == nil || ep.eventPool == nil {
		return
	}
	
	// Return the specific event data to its pool
	switch event.Type {
	case EventTypeParamValue:
		if e, ok := event.Data.(ParamValueEvent); ok {
			ep.eventPool.ReturnParamValueEvent(&e)
		}
	case EventTypeParamMod:
		if e, ok := event.Data.(ParamModEvent); ok {
			ep.eventPool.ReturnParamModEvent(&e)
		}
	case EventTypeParamGestureBegin, EventTypeParamGestureEnd:
		if e, ok := event.Data.(ParamGestureEvent); ok {
			ep.eventPool.ReturnParamGestureEvent(&e)
		}
	case EventTypeNoteOn, EventTypeNoteOff, EventTypeNoteChoke, EventTypeNoteEnd:
		if e, ok := event.Data.(NoteEvent); ok {
			ep.eventPool.ReturnNoteEvent(&e)
		}
	case EventTypeNoteExpression:
		if e, ok := event.Data.(NoteExpressionEvent); ok {
			ep.eventPool.ReturnNoteExpressionEvent(&e)
		}
	case EventTypeTransport:
		if e, ok := event.Data.(TransportEvent); ok {
			ep.eventPool.ReturnTransportEvent(&e)
		}
	case EventTypeMIDI:
		if e, ok := event.Data.(MIDIEvent); ok {
			ep.eventPool.ReturnMIDIEvent(&e)
		}
	case EventTypeMIDISysex:
		if e, ok := event.Data.(MIDISysexEvent); ok {
			ep.eventPool.ReturnMIDISysexEvent(&e)
		}
	case EventTypeMIDI2:
		if e, ok := event.Data.(MIDI2Event); ok {
			ep.eventPool.ReturnMIDI2Event(&e)
		}
	}
	
	// Return the event struct itself
	ep.eventPool.ReturnEvent(event)
}

// ProcessAllEvents processes all input events and calls the appropriate handler
func (ep *EventProcessor) ProcessAllEvents(handler TypedEventHandler) {
	if handler == nil {
		return
	}
	
	count := ep.GetInputEventCount()
	// Collect events to return to pool after processing
	events := make([]*Event, 0, count)
	
	for i := uint32(0); i < count; i++ {
		event := ep.GetInputEvent(i)
		if event == nil {
			continue
		}
		events = append(events, event)
		
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
	
	// Return all events to the pool after processing
	for _, event := range events {
		ep.ReturnEventToPool(event)
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
func convertCEventToGo(cEventHeader *C.clap_event_header_t, pool *EventPool) *Event {
	if cEventHeader == nil {
		return nil
	}
	
	// Only process events from the core namespace (0)
	if cEventHeader.space_id != 0 {
		return nil
	}
	
	// Get event from pool if available
	var event *Event
	if pool != nil {
		event = pool.GetEvent()
	} else {
		event = &Event{}
	}
	
	event.Time = uint32(cEventHeader.time)
	event.Type = int(cEventHeader._type)
	event.Flags = uint32(cEventHeader.flags)
	
	// Convert event data based on type
	switch int(cEventHeader._type) {
	case EventTypeParamValue:
		cParamEvent := (*C.clap_event_param_value_t)(unsafe.Pointer(cEventHeader))
		var paramEvent *ParamValueEvent
		if pool != nil {
			paramEvent = pool.GetParamValueEvent()
		} else {
			paramEvent = &ParamValueEvent{}
		}
		paramEvent.ParamID = uint32(cParamEvent.param_id)
		paramEvent.Cookie = unsafe.Pointer(cParamEvent.cookie)
		paramEvent.NoteID = int32(cParamEvent.note_id)
		paramEvent.Port = int16(cParamEvent.port_index)
		paramEvent.Channel = int16(cParamEvent.channel)
		paramEvent.Key = int16(cParamEvent.key)
		paramEvent.Value = float64(cParamEvent.value)
		event.Data = *paramEvent
		
	case EventTypeParamMod:
		cParamModEvent := (*C.clap_event_param_mod_t)(unsafe.Pointer(cEventHeader))
		var paramModEvent *ParamModEvent
		if pool != nil {
			paramModEvent = pool.GetParamModEvent()
		} else {
			paramModEvent = &ParamModEvent{}
		}
		paramModEvent.ParamID = uint32(cParamModEvent.param_id)
		paramModEvent.Cookie = unsafe.Pointer(cParamModEvent.cookie)
		paramModEvent.NoteID = int32(cParamModEvent.note_id)
		paramModEvent.Port = int16(cParamModEvent.port_index)
		paramModEvent.Channel = int16(cParamModEvent.channel)
		paramModEvent.Key = int16(cParamModEvent.key)
		paramModEvent.Amount = float64(cParamModEvent.amount)
		event.Data = *paramModEvent
		
	case EventTypeParamGestureBegin, EventTypeParamGestureEnd:
		cGestureEvent := (*C.clap_event_param_gesture_t)(unsafe.Pointer(cEventHeader))
		var gestureEvent *ParamGestureEvent
		if pool != nil {
			gestureEvent = pool.GetParamGestureEvent()
		} else {
			gestureEvent = &ParamGestureEvent{}
		}
		gestureEvent.ParamID = uint32(cGestureEvent.param_id)
		event.Data = *gestureEvent
		
	case EventTypeNoteOn, EventTypeNoteOff, EventTypeNoteChoke, EventTypeNoteEnd:
		cNoteEvent := (*C.clap_event_note_t)(unsafe.Pointer(cEventHeader))
		var noteEvent *NoteEvent
		if pool != nil {
			noteEvent = pool.GetNoteEvent()
		} else {
			noteEvent = &NoteEvent{}
		}
		noteEvent.NoteID = int32(cNoteEvent.note_id)
		noteEvent.Port = int16(cNoteEvent.port_index)
		noteEvent.Channel = int16(cNoteEvent.channel)
		noteEvent.Key = int16(cNoteEvent.key)
		noteEvent.Velocity = float64(cNoteEvent.velocity)
		event.Data = *noteEvent
		
	case EventTypeNoteExpression:
		cNoteExprEvent := (*C.clap_event_note_expression_t)(unsafe.Pointer(cEventHeader))
		var noteExprEvent *NoteExpressionEvent
		if pool != nil {
			noteExprEvent = pool.GetNoteExpressionEvent()
		} else {
			noteExprEvent = &NoteExpressionEvent{}
		}
		noteExprEvent.ExpressionID = int32(cNoteExprEvent.expression_id)
		noteExprEvent.NoteID = int32(cNoteExprEvent.note_id)
		noteExprEvent.Port = int16(cNoteExprEvent.port_index)
		noteExprEvent.Channel = int16(cNoteExprEvent.channel)
		noteExprEvent.Key = int16(cNoteExprEvent.key)
		noteExprEvent.Value = float64(cNoteExprEvent.value)
		event.Data = *noteExprEvent
		
	case EventTypeTransport:
		cTransportEvent := (*C.clap_event_transport_t)(unsafe.Pointer(cEventHeader))
		var transportEvent *TransportEvent
		if pool != nil {
			transportEvent = pool.GetTransportEvent()
		} else {
			transportEvent = &TransportEvent{}
		}
		transportEvent.Flags = uint32(cTransportEvent.flags)
		transportEvent.SongPosBeats = uint64(cTransportEvent.song_pos_beats)
		transportEvent.SongPosSeconds = uint64(cTransportEvent.song_pos_seconds)
		transportEvent.Tempo = float64(cTransportEvent.tempo)
		transportEvent.TempoInc = float64(cTransportEvent.tempo_inc)
		transportEvent.LoopStartBeats = uint64(cTransportEvent.loop_start_beats)
		transportEvent.LoopEndBeats = uint64(cTransportEvent.loop_end_beats)
		transportEvent.LoopStartSeconds = uint64(cTransportEvent.loop_start_seconds)
		transportEvent.LoopEndSeconds = uint64(cTransportEvent.loop_end_seconds)
		transportEvent.BarStart = uint64(cTransportEvent.bar_start)
		transportEvent.BarNumber = int32(cTransportEvent.bar_number)
		transportEvent.TimeSignatureNum = uint16(cTransportEvent.tsig_num)
		transportEvent.TimeSignatureDen = uint16(cTransportEvent.tsig_denom)
		event.Data = *transportEvent
		
	case EventTypeMIDI:
		cMidiEvent := (*C.clap_event_midi_t)(unsafe.Pointer(cEventHeader))
		var midiEvent *MIDIEvent
		if pool != nil {
			midiEvent = pool.GetMIDIEvent()
		} else {
			midiEvent = &MIDIEvent{Data: make([]byte, 3)}
		}
		midiEvent.Port = int16(cMidiEvent.port_index)
		for i := 0; i < 3; i++ {
			midiEvent.Data[i] = byte(cMidiEvent.data[i])
		}
		event.Data = *midiEvent
		
	case EventTypeMIDISysex:
		cSysexEvent := (*C.clap_event_midi_sysex_t)(unsafe.Pointer(cEventHeader))
		var sysexEvent *MIDISysexEvent
		if pool != nil {
			sysexEvent = pool.GetMIDISysexEvent()
		} else {
			sysexEvent = &MIDISysexEvent{}
		}
		sysexEvent.Port = int16(cSysexEvent.port_index)
		// Copy sysex data to Go slice
		sysexEvent.Data = make([]byte, cSysexEvent.size)
		if cSysexEvent.size > 0 && cSysexEvent.buffer != nil {
			C.memcpy(unsafe.Pointer(&sysexEvent.Data[0]), unsafe.Pointer(cSysexEvent.buffer), C.size_t(cSysexEvent.size))
		}
		event.Data = *sysexEvent
		
	case EventTypeMIDI2:
		cMidi2Event := (*C.clap_event_midi2_t)(unsafe.Pointer(cEventHeader))
		var midi2Event *MIDI2Event
		if pool != nil {
			midi2Event = pool.GetMIDI2Event()
		} else {
			midi2Event = &MIDI2Event{Data: make([]uint32, 4)}
		}
		midi2Event.Port = int16(cMidi2Event.port_index)
		for i := 0; i < 4; i++ {
			midi2Event.Data[i] = uint32(cMidi2Event.data[i])
		}
		event.Data = *midi2Event
		
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