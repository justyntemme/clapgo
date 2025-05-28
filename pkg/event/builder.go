package event

import (
	"errors"
	"fmt"
	"unsafe"
)

// EventBuilder provides a fluent interface for creating events
type EventBuilder struct {
	header Header
	err    error
}

// NewEventBuilder creates a new event builder
func NewEventBuilder(eventType uint16) *EventBuilder {
	return &EventBuilder{
		header: Header{
			Size:    0, // Will be set when building specific event types
			Time:    0,
			SpaceID: 0,
			Type:    eventType,
			Flags:   0,
		},
	}
}

// Time sets the event time
func (b *EventBuilder) Time(time uint32) *EventBuilder {
	if b.err != nil {
		return b
	}
	b.header.Time = time
	return b
}

// SpaceID sets the event space ID
func (b *EventBuilder) SpaceID(spaceID uint16) *EventBuilder {
	if b.err != nil {
		return b
	}
	b.header.SpaceID = spaceID
	return b
}

// Flags sets the event flags
func (b *EventBuilder) Flags(flags uint32) *EventBuilder {
	if b.err != nil {
		return b
	}
	b.header.Flags = flags
	return b
}

// AddFlags adds additional flags
func (b *EventBuilder) AddFlags(flags uint32) *EventBuilder {
	if b.err != nil {
		return b
	}
	b.header.Flags |= flags
	return b
}

// Live marks this event as live
func (b *EventBuilder) Live() *EventBuilder {
	return b.AddFlags(FlagIsLive)
}

// DontRecord marks this event as not to be recorded
func (b *EventBuilder) DontRecord() *EventBuilder {
	return b.AddFlags(FlagDontRecord)
}

// ParamValueEventBuilder extends EventBuilder for parameter value events
type ParamValueEventBuilder struct {
	*EventBuilder
	event ParamValueEvent
}

// NewParamValueEvent creates a new parameter value event builder
func NewParamValueEvent(paramID uint32, value float64) *ParamValueEventBuilder {
	builder := NewEventBuilder(uint16(TypeParamValue))
	return &ParamValueEventBuilder{
		EventBuilder: builder,
		event: ParamValueEvent{
			Header:  builder.header,
			ParamID: paramID,
			Cookie:  nil,
			NoteID:  -1, // -1 means global parameter
			Port:    -1,
			Channel: -1,
			Key:     -1,
			Value:   value,
		},
	}
}

// Cookie sets the parameter cookie
func (b *ParamValueEventBuilder) Cookie(cookie unsafe.Pointer) *ParamValueEventBuilder {
	if b.err != nil {
		return b
	}
	b.event.Cookie = cookie
	return b
}

// NoteID sets the note ID for per-note parameters
func (b *ParamValueEventBuilder) NoteID(noteID int32) *ParamValueEventBuilder {
	if b.err != nil {
		return b
	}
	b.event.NoteID = noteID
	return b
}

// Port sets the port for the parameter
func (b *ParamValueEventBuilder) Port(port int16) *ParamValueEventBuilder {
	if b.err != nil {
		return b
	}
	b.event.Port = port
	return b
}

// Channel sets the channel for the parameter
func (b *ParamValueEventBuilder) Channel(channel int16) *ParamValueEventBuilder {
	if b.err != nil {
		return b
	}
	if channel < -1 || channel > 15 {
		b.err = errors.New("channel must be -1 (none) or 0-15")
		return b
	}
	b.event.Channel = channel
	return b
}

// Key sets the key for the parameter
func (b *ParamValueEventBuilder) Key(key int16) *ParamValueEventBuilder {
	if b.err != nil {
		return b
	}
	if key < -1 || key > 127 {
		b.err = errors.New("key must be -1 (none) or 0-127")
		return b
	}
	b.event.Key = key
	return b
}

// Build creates the parameter value event
func (b *ParamValueEventBuilder) Build() (ParamValueEvent, error) {
	if b.err != nil {
		return ParamValueEvent{}, b.err
	}
	
	// Update header in event
	b.event.Header = b.header
	b.event.Header.Size = uint32(unsafe.Sizeof(b.event))
	
	return b.event, nil
}

// MustBuild creates the parameter value event, panicking on error
func (b *ParamValueEventBuilder) MustBuild() ParamValueEvent {
	event, err := b.Build()
	if err != nil {
		panic(err)
	}
	return event
}

// NoteEventBuilder extends EventBuilder for note events
type NoteEventBuilder struct {
	*EventBuilder
	event NoteEvent
}

// NewNoteEvent creates a new note event builder
func NewNoteEvent(eventType uint16, noteID int32, port, channel, key int16, velocity float64) *NoteEventBuilder {
	builder := NewEventBuilder(eventType)
	return &NoteEventBuilder{
		EventBuilder: builder,
		event: NoteEvent{
			Header:   builder.header,
			NoteID:   noteID,
			Port:     port,
			Channel:  channel,
			Key:      key,
			Velocity: velocity,
		},
	}
}

// NewNoteOn creates a note on event builder
func NewNoteOn(noteID int32, port, channel, key int16, velocity float64) *NoteEventBuilder {
	return NewNoteEvent(uint16(TypeNoteOn), noteID, port, channel, key, velocity)
}

// NewNoteOff creates a note off event builder
func NewNoteOff(noteID int32, port, channel, key int16, velocity float64) *NoteEventBuilder {
	return NewNoteEvent(uint16(TypeNoteOff), noteID, port, channel, key, velocity)
}

// NewNoteChoke creates a note choke event builder
func NewNoteChoke(noteID int32, port, channel, key int16) *NoteEventBuilder {
	return NewNoteEvent(uint16(TypeNoteChoke), noteID, port, channel, key, 0.0)
}

// NewNoteEnd creates a note end event builder
func NewNoteEnd(noteID int32, port, channel, key int16) *NoteEventBuilder {
	return NewNoteEvent(uint16(TypeNoteEnd), noteID, port, channel, key, 0.0)
}

// Velocity sets the note velocity
func (b *NoteEventBuilder) Velocity(velocity float64) *NoteEventBuilder {
	if b.err != nil {
		return b
	}
	if velocity < 0.0 || velocity > 1.0 {
		b.err = errors.New("velocity must be between 0.0 and 1.0")
		return b
	}
	b.event.Velocity = velocity
	return b
}

// Build creates the note event
func (b *NoteEventBuilder) Build() (NoteEvent, error) {
	if b.err != nil {
		return NoteEvent{}, b.err
	}
	
	// Validate
	if b.event.Channel < 0 || b.event.Channel > 15 {
		return NoteEvent{}, errors.New("channel must be 0-15")
	}
	
	if b.event.Key < 0 || b.event.Key > 127 {
		return NoteEvent{}, errors.New("key must be 0-127")
	}
	
	// Update header in event
	b.event.Header = b.header
	b.event.Header.Size = uint32(unsafe.Sizeof(b.event))
	
	return b.event, nil
}

// MustBuild creates the note event, panicking on error
func (b *NoteEventBuilder) MustBuild() NoteEvent {
	event, err := b.Build()
	if err != nil {
		panic(err)
	}
	return event
}

// EventSequenceBuilder helps build sequences of events
type EventSequenceBuilder struct {
	events []Event
	err    error
}

// NewEventSequence creates a new event sequence builder
func NewEventSequence() *EventSequenceBuilder {
	return &EventSequenceBuilder{
		events: make([]Event, 0),
	}
}

// AddEvent adds an event to the sequence
func (b *EventSequenceBuilder) AddEvent(event Event) *EventSequenceBuilder {
	if b.err != nil {
		return b
	}
	b.events = append(b.events, event)
	return b
}

// AddParamChange adds a parameter change event
func (b *EventSequenceBuilder) AddParamChange(time uint32, paramID uint32, value float64) *EventSequenceBuilder {
	if b.err != nil {
		return b
	}
	
	event, err := NewParamValueEvent(paramID, value).
		Time(time).
		Build()
	
	if err != nil {
		b.err = fmt.Errorf("failed to create param change event: %w", err)
		return b
	}
	
	return b.AddEvent(&event)
}

// AddNoteOn adds a note on event
func (b *EventSequenceBuilder) AddNoteOn(time uint32, noteID int32, channel, key int16, velocity float64) *EventSequenceBuilder {
	if b.err != nil {
		return b
	}
	
	event, err := NewNoteOn(noteID, 0, channel, key, velocity).
		Time(time).
		Build()
	
	if err != nil {
		b.err = fmt.Errorf("failed to create note on event: %w", err)
		return b
	}
	
	return b.AddEvent(&event)
}

// AddNoteOff adds a note off event
func (b *EventSequenceBuilder) AddNoteOff(time uint32, noteID int32, channel, key int16, velocity float64) *EventSequenceBuilder {
	if b.err != nil {
		return b
	}
	
	event, err := NewNoteOff(noteID, 0, channel, key, velocity).
		Time(time).
		Build()
	
	if err != nil {
		b.err = fmt.Errorf("failed to create note off event: %w", err)
		return b
	}
	
	return b.AddEvent(&event)
}

// Build creates the event sequence
func (b *EventSequenceBuilder) Build() ([]Event, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.events, nil
}

// MustBuild creates the event sequence, panicking on error
func (b *EventSequenceBuilder) MustBuild() []Event {
	events, err := b.Build()
	if err != nil {
		panic(err)
	}
	return events
}

// Example usage:
// paramEvent := NewParamValueEvent(0, 0.5).
//     Time(100).
//     Channel(2).
//     Live().
//     Build()
//
// noteEvent := NewNoteOn(123, 0, 0, 60, 0.8).
//     Time(200).
//     DontRecord().
//     Build()
//
// sequence := NewEventSequence().
//     AddParamChange(0, 1, 0.75).
//     AddNoteOn(100, 1, 0, 60, 0.8).
//     AddNoteOff(500, 1, 0, 60, 0.0).
//     Build()