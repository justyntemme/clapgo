package api

import (
	"testing"
	"unsafe"
)

// Mock C structures for benchmarking
type mockClapInputEvents struct{}
type mockClapOutputEvents struct{}
type mockClapEventHeader struct {
	size     uint32
	time     uint32
	spaceID  uint16
	typeID   uint16
	flags    uint32
}

// BenchmarkEventPoolAllocation tests event pool allocation performance
func BenchmarkEventPoolAllocation(b *testing.B) {
	pool := NewEventPool()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Get events from pool
		paramEvent := pool.GetParamValueEvent()
		noteEvent := pool.GetNoteEvent()
		midiEvent := pool.GetMIDIEvent()
		
		// Return events to pool
		pool.ReturnParamValueEvent(paramEvent)
		pool.ReturnNoteEvent(noteEvent)
		pool.ReturnMIDIEvent(midiEvent)
	}
}

// BenchmarkProcessTypedEvents tests zero-allocation event processing
func BenchmarkProcessTypedEvents(b *testing.B) {
	ep := &EventProcessor{
		eventPool: NewEventPool(),
	}
	
	handler := &benchEventHandler{}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		// Simulate processing 100 events
		for j := 0; j < 100; j++ {
			// Get event from pool
			event := ep.eventPool.GetParamValueEvent()
			event.ParamID = uint32(j)
			event.Value = float64(j) * 0.1
			
			// Process event
			handler.HandleParamValue(event, 0)
			
			// Return to pool
			ep.eventPool.ReturnParamValueEvent(event)
		}
	}
}

// BenchmarkMIDIEventProcessing tests MIDI event handling with fixed arrays
func BenchmarkMIDIEventProcessing(b *testing.B) {
	pool := NewEventPool()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		// Get MIDI event
		midi := pool.GetMIDIEvent()
		
		// Fill fixed array (no allocation)
		midi.Port = 0
		midi.Data[0] = 0x90 // Note on
		midi.Data[1] = 60   // Middle C
		midi.Data[2] = 100  // Velocity
		
		// Process MIDI data (simulated)
		_ = midi.Data[0] & 0xF0 // Extract message type
		_ = midi.Data[1]        // Note number
		_ = midi.Data[2]        // Velocity
		
		// Return to pool
		pool.ReturnMIDIEvent(midi)
	}
}

// BenchmarkEventProcessorCreate tests EventProcessor creation
func BenchmarkEventProcessorCreate(b *testing.B) {
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		ep := NewEventProcessor(
			unsafe.Pointer(&mockClapInputEvents{}),
			unsafe.Pointer(&mockClapOutputEvents{}),
		)
		_ = ep
	}
}

// benchEventHandler is a mock event handler for benchmarks
type benchEventHandler struct {
	paramCount int
	noteCount  int
}

func (h *benchEventHandler) HandleNoteOn(event *NoteEvent, time uint32) error {
	h.noteCount++
	return nil
}

func (h *benchEventHandler) HandleNoteOff(event *NoteEvent, time uint32) error {
	h.noteCount++
	return nil
}

func (h *benchEventHandler) HandleNoteChoke(event *NoteEvent, time uint32) error {
	h.noteCount++
	return nil
}

func (h *benchEventHandler) HandleNoteEnd(event *NoteEvent, time uint32) error {
	h.noteCount++
	return nil
}

func (h *benchEventHandler) HandleParamValue(event *ParamValueEvent, time uint32) error {
	h.paramCount++
	return nil
}

func (h *benchEventHandler) HandleParamMod(event *ParamModEvent, time uint32) error {
	h.paramCount++
	return nil
}

func (h *benchEventHandler) HandleTransport(event *TransportEvent, time uint32) error {
	return nil
}

func (h *benchEventHandler) HandleMIDI(event *MIDIEvent, time uint32) error {
	return nil
}

func (h *benchEventHandler) HandleMIDISysex(event *MIDISysexEvent, time uint32) error {
	return nil
}

func (h *benchEventHandler) HandleMIDI2(event *MIDI2Event, time uint32) error {
	return nil
}