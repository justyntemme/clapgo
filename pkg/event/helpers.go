package event

import (
	"unsafe"
	"github.com/justyntemme/clapgo/pkg/host"
)

// Diagnostics tracks event processing statistics for periodic logging
type Diagnostics struct {
	processCallCount  uint64
	lastEventPoolDump uint64
}

// SetupPoolLogging configures event pool logging for diagnostics
func SetupPoolLogging(processor *Processor, logger *host.Logger) {
	if processor == nil || logger == nil {
		return
	}
	
	if pool := processor.GetPool(); pool != nil {
		pool.SetLogger(logger)
	}
}

// LogPoolDiagnostics logs pool diagnostics every N process calls
func (d *Diagnostics) LogPoolDiagnostics(processor *Processor, interval uint64) {
	d.processCallCount++
	if d.processCallCount%interval == 0 && d.processCallCount != d.lastEventPoolDump {
		d.lastEventPoolDump = d.processCallCount
		if pool := processor.GetPool(); pool != nil {
			pool.LogDiagnostics()
		}
	}
}

// ProcessingStats provides event processing statistics
type ProcessingStats struct {
	EventsProcessed   uint64
	ParamEvents       uint64
	NoteEvents        uint64
	MIDIEvents        uint64
	TransportEvents   uint64
	ProcessCalls      uint64
}

// Reset clears all counters
func (s *ProcessingStats) Reset() {
	s.EventsProcessed = 0
	s.ParamEvents = 0
	s.NoteEvents = 0
	s.MIDIEvents = 0
	s.TransportEvents = 0
	s.ProcessCalls = 0
}

// Log outputs current statistics
func (s *ProcessingStats) Log(logger *host.Logger) {
	if logger == nil {
		return
	}
	
	logger.Log(host.SeverityInfo, "Event Processing Statistics:")
	logger.Log(host.SeverityInfo, "  Total Events: %d", s.EventsProcessed)
	logger.Log(host.SeverityInfo, "  Param Events: %d", s.ParamEvents)
	logger.Log(host.SeverityInfo, "  Note Events: %d", s.NoteEvents)
	logger.Log(host.SeverityInfo, "  MIDI Events: %d", s.MIDIEvents)
	logger.Log(host.SeverityInfo, "  Transport Events: %d", s.TransportEvents)
	logger.Log(host.SeverityInfo, "  Process Calls: %d", s.ProcessCalls)
}

// ProcessStandardMIDI handles common MIDI messages and converts to note events
func ProcessStandardMIDI(event *MIDIEvent, handler Handler, time uint32) {
	status := event.Data[0] & 0xF0
	channel := int16(event.Data[0] & 0x0F)
	
	switch status {
	case MIDINoteOn: // Note On
		if event.Data[2] > 0 {
			if noteOn := MIDIToNoteOn(event.Data, int16(event.Port), channel); noteOn != nil {
				handler.HandleNoteOn(noteOn, time)
			}
		} else {
			// Velocity 0 means note off
			if noteOff := MIDIToNoteOff(event.Data, int16(event.Port), channel); noteOff != nil {
				handler.HandleNoteOff(noteOff, time)
			}
		}
		
	case MIDINoteOff: // Note Off
		if noteOff := MIDIToNoteOff(event.Data, int16(event.Port), channel); noteOff != nil {
			handler.HandleNoteOff(noteOff, time)
		}
		
	case MIDIControlChange: // Control Change
		// Can be extended to handle CC events
		
	case MIDIPitchBend: // Pitch Bend
		// Can be extended to handle pitch bend
	}
}

// CreateParamValue creates a parameter value change event
func CreateParamValue(time uint32, paramID uint32, value float64) *ParamValueEvent {
	return &ParamValueEvent{
		Header: Header{
			Time: time,
			Type: uint16(TypeParamValue),
		},
		ParamID: paramID,
		NoteID:  -1, // Global parameter
		Port:    -1,
		Channel: -1,
		Key:     -1,
		Value:   value,
	}
}

// CreatePolyParamValue creates a polyphonic parameter value change event
func CreatePolyParamValue(time uint32, paramID uint32, noteID int32, port, channel, key int16, value float64) *ParamValueEvent {
	return &ParamValueEvent{
		Header: Header{
			Time: time,
			Type: uint16(TypeParamValue),
		},
		ParamID: paramID,
		NoteID:  noteID,
		Port:    port,
		Channel: channel,
		Key:     key,
		Value:   value,
	}
}

// CreateNoteOn creates a note on event
func CreateNoteOn(time uint32, port, channel, key int16, velocity float64) *NoteEvent {
	return &NoteEvent{
		Header: Header{
			Time: time,
			Type: uint16(TypeNoteOn),
		},
		NoteID:   -1, // Host will assign
		Port:     port,
		Channel:  channel,
		Key:      key,
		Velocity: velocity,
	}
}

// CreateNoteOff creates a note off event
func CreateNoteOff(time uint32, port, channel, key int16, velocity float64) *NoteEvent {
	return &NoteEvent{
		Header: Header{
			Time: time,
			Type: uint16(TypeNoteOff),
		},
		NoteID:   -1, // Match any note with this key
		Port:     port,
		Channel:  channel,
		Key:      key,
		Velocity: velocity,
	}
}

// CreateNoteChoke creates a note choke event
func CreateNoteChoke(time uint32, port, channel, key int16) *NoteEvent {
	return &NoteEvent{
		Header: Header{
			Time: time,
			Type: uint16(TypeNoteChoke),
		},
		NoteID:   -1,
		Port:     port,
		Channel:  channel,
		Key:      key,
		Velocity: 0, // Ignored for choke
	}
}

// CreateNoteEnd creates a note end event (sent by plugin to host)
func CreateNoteEnd(time uint32, noteID int32, port, channel, key int16) *NoteEvent {
	return &NoteEvent{
		Header: Header{
			Time: time,
			Type: uint16(TypeNoteEnd),
		},
		NoteID:   noteID,
		Port:     port,
		Channel:  channel,
		Key:      key,
		Velocity: 0, // Ignored for end
	}
}

// CreateMIDI creates a MIDI 1.0 event
func CreateMIDI(time uint32, port uint16, data [3]byte) *MIDIEvent {
	return &MIDIEvent{
		Header: Header{
			Time: time,
			Type: uint16(TypeMIDI),
		},
		Port: port,
		Data: data,
	}
}

// CreateMIDISysex creates a MIDI system exclusive event
func CreateMIDISysex(time uint32, port uint16, buffer unsafe.Pointer, size uint32) *MIDISysexEvent {
	return &MIDISysexEvent{
		Header: Header{
			Time: time,
			Type: uint16(TypeMIDISysex),
		},
		Port:   port,
		Buffer: buffer,
		Size:   size,
	}
}

// CreateMIDI2 creates a MIDI 2.0 event
func CreateMIDI2(time uint32, port uint16, data [4]uint32) *MIDI2Event {
	return &MIDI2Event{
		Header: Header{
			Time: time,
			Type: uint16(TypeMIDI2),
		},
		Port: port,
		Data: data,
	}
}

// CreateNoteEndEvent creates a note end event
func CreateNoteEndEvent(time uint32, noteID int32, port, channel, key int16) *NoteEvent {
	return &NoteEvent{
		Header: Header{
			Time: time,
			Type: uint16(TypeNoteEnd),
		},
		NoteID:   noteID,
		Port:     port,
		Channel:  channel,
		Key:      key,
		Velocity: 0, // Ignored for end
	}
}