package api

import (
	"fmt"
	"sync/atomic"
)

// EventPoolDiagnostics tracks event pool statistics for periodic logging
type EventPoolDiagnostics struct {
	processCallCount  uint64
	lastEventPoolDump uint64
}

// SetupPoolLogging configures event pool logging for diagnostics
// This should be called with an EventProcessor, not EventHandler
func SetupPoolLogging(processor *EventProcessor, logger Logger) {
	if processor == nil || logger == nil {
		return
	}
	
	if pool := processor.GetEventPool(); pool != nil {
		pool.logger = logger
	}
}

// LogPoolDiagnostics logs pool diagnostics every N process calls
func (d *EventPoolDiagnostics) LogPoolDiagnostics(processor *EventProcessor, interval uint64) {
	d.processCallCount++
	if d.processCallCount%interval == 0 && d.processCallCount != d.lastEventPoolDump {
		d.lastEventPoolDump = d.processCallCount
		if pool := processor.GetEventPool(); pool != nil {
			pool.LogDiagnostics()
		}
	}
}


// UpdateParameterAtomic updates both atomic storage and parameter manager
// This ensures thread-safe parameter updates visible to both audio thread and host
func UpdateParameterAtomic(atomicStorage *int64, value float64, paramManager *ParameterManager, paramID uint32) {
	// Update atomic storage for audio thread
	atomic.StoreInt64(atomicStorage, int64(FloatToBits(value)))
	// Update parameter manager for host/UI
	paramManager.SetParameterValue(paramID, value)
}

// LoadParameterAtomic loads a parameter value from atomic storage
func LoadParameterAtomic(atomicStorage *int64) float64 {
	bits := atomic.LoadInt64(atomicStorage)
	return FloatFromBits(uint64(bits))
}

// MIDIToNoteOn converts a MIDI 1.0 note on message to a CLAP note event
func MIDIToNoteOn(data [3]byte, port, channel int16) *NoteEvent {
	status := data[0] & 0xF0
	if status != 0x90 || data[2] == 0 { // velocity 0 is note off
		return nil
	}
	
	return &NoteEvent{
		NoteID:   -1, // Will be assigned by host
		Port:     port,
		Channel:  channel,
		Key:      int16(data[1]),
		Velocity: float64(data[2]) / 127.0,
	}
}

// MIDIToNoteOff converts a MIDI 1.0 note off message to a CLAP note event
func MIDIToNoteOff(data [3]byte, port, channel int16) *NoteEvent {
	status := data[0] & 0xF0
	// Note off can be 0x80 or 0x90 with velocity 0
	if status == 0x80 || (status == 0x90 && data[2] == 0) {
		return &NoteEvent{
			NoteID:   -1, // Will be assigned by host
			Port:     port,
			Channel:  channel,
			Key:      int16(data[1]),
			Velocity: 0.0,
		}
	}
	
	return nil
}

// ProcessStandardMIDI handles common MIDI messages and converts to note events
func ProcessStandardMIDI(event *MIDIEvent, handler TypedEventHandler, time uint32) {
	status := event.Data[0] & 0xF0
	channel := int16(event.Data[0] & 0x0F)
	
	switch status {
	case 0x90: // Note On
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
		
	case 0x80: // Note Off
		if noteOff := MIDIToNoteOff(event.Data, int16(event.Port), channel); noteOff != nil {
			handler.HandleNoteOff(noteOff, time)
		}
		
	case 0xB0: // Control Change
		// Can be extended to handle CC events
		
	case 0xE0: // Pitch Bend
		// Can be extended to handle pitch bend
	}
}

// EventDiagnostics provides event processing statistics
type EventDiagnostics struct {
	EventsProcessed   uint64
	ParamEvents       uint64
	NoteEvents        uint64
	MIDIEvents        uint64
	TransportEvents   uint64
	ProcessCalls      uint64
}

// Reset clears all counters
func (d *EventDiagnostics) Reset() {
	d.EventsProcessed = 0
	d.ParamEvents = 0
	d.NoteEvents = 0
	d.MIDIEvents = 0
	d.TransportEvents = 0
	d.ProcessCalls = 0
}

// Log outputs current diagnostics using the provided logger
func (d *EventDiagnostics) Log(logger *HostLogger) {
	if logger == nil {
		return
	}
	
	logger.Info("Event Processing Diagnostics:")
	logger.Info(fmt.Sprintf("  Total Events: %d", d.EventsProcessed))
	logger.Info(fmt.Sprintf("  Param Events: %d", d.ParamEvents))
	logger.Info(fmt.Sprintf("  Note Events: %d", d.NoteEvents))
	logger.Info(fmt.Sprintf("  MIDI Events: %d", d.MIDIEvents))
	logger.Info(fmt.Sprintf("  Transport Events: %d", d.TransportEvents))
	logger.Info(fmt.Sprintf("  Process Calls: %d", d.ProcessCalls))
}