package event

import (
	"fmt"
	"sync"
	"sync/atomic"
	
	"github.com/justyntemme/clapgo/pkg/host"
)

// Pool manages pre-allocated events to avoid allocations during audio processing
type Pool struct {
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

	// Diagnostics
	totalAllocations  uint64
	poolHits          uint64
	poolMisses        uint64
	highWaterMark     uint64
	currentAllocated  uint64
	
	// Logger for diagnostics
	logger *host.Logger
}

// NewPool creates a new event pool
func NewPool() *Pool {
	p := &Pool{}

	// Initialize pools with factory functions
	p.paramValuePool.New = func() interface{} {
		atomic.AddUint64(&p.totalAllocations, 1)
		atomic.AddUint64(&p.poolMisses, 1)
		return &ParamValueEvent{}
	}

	p.paramModPool.New = func() interface{} {
		atomic.AddUint64(&p.totalAllocations, 1)
		atomic.AddUint64(&p.poolMisses, 1)
		return &ParamModEvent{}
	}

	p.paramGesturePool.New = func() interface{} {
		atomic.AddUint64(&p.totalAllocations, 1)
		atomic.AddUint64(&p.poolMisses, 1)
		return &ParamGestureEvent{}
	}

	p.noteEventPool.New = func() interface{} {
		atomic.AddUint64(&p.totalAllocations, 1)
		atomic.AddUint64(&p.poolMisses, 1)
		return &NoteEvent{}
	}

	p.noteExpressionPool.New = func() interface{} {
		atomic.AddUint64(&p.totalAllocations, 1)
		atomic.AddUint64(&p.poolMisses, 1)
		return &NoteExpressionEvent{}
	}

	p.transportPool.New = func() interface{} {
		atomic.AddUint64(&p.totalAllocations, 1)
		atomic.AddUint64(&p.poolMisses, 1)
		return &TransportEvent{}
	}

	p.midiPool.New = func() interface{} {
		atomic.AddUint64(&p.totalAllocations, 1)
		atomic.AddUint64(&p.poolMisses, 1)
		return &MIDIEvent{}
	}

	p.midiSysexPool.New = func() interface{} {
		atomic.AddUint64(&p.totalAllocations, 1)
		atomic.AddUint64(&p.poolMisses, 1)
		return &MIDISysexEvent{}
	}

	p.midi2Pool.New = func() interface{} {
		atomic.AddUint64(&p.totalAllocations, 1)
		atomic.AddUint64(&p.poolMisses, 1)
		return &MIDI2Event{}
	}

	return p
}

// GetParamValueEvent gets a ParamValueEvent from the pool
func (p *Pool) GetParamValueEvent() *ParamValueEvent {
	event := p.paramValuePool.Get().(*ParamValueEvent)
	atomic.AddUint64(&p.poolHits, 1)
	atomic.AddUint64(&p.currentAllocated, 1)
	
	// Update high water mark
	current := atomic.LoadUint64(&p.currentAllocated)
	for {
		high := atomic.LoadUint64(&p.highWaterMark)
		if current <= high || atomic.CompareAndSwapUint64(&p.highWaterMark, high, current) {
			break
		}
	}
	
	return event
}

// PutParamValueEvent returns a ParamValueEvent to the pool
func (p *Pool) PutParamValueEvent(event *ParamValueEvent) {
	// Clear the event before returning to pool
	*event = ParamValueEvent{}
	p.paramValuePool.Put(event)
	atomic.AddUint64(&p.currentAllocated, ^uint64(0)) // Decrement
}

// GetParamModEvent gets a ParamModEvent from the pool
func (p *Pool) GetParamModEvent() *ParamModEvent {
	event := p.paramModPool.Get().(*ParamModEvent)
	atomic.AddUint64(&p.poolHits, 1)
	atomic.AddUint64(&p.currentAllocated, 1)
	return event
}

// PutParamModEvent returns a ParamModEvent to the pool
func (p *Pool) PutParamModEvent(event *ParamModEvent) {
	*event = ParamModEvent{}
	p.paramModPool.Put(event)
	atomic.AddUint64(&p.currentAllocated, ^uint64(0))
}

// GetParamGestureEvent gets a ParamGestureEvent from the pool
func (p *Pool) GetParamGestureEvent() *ParamGestureEvent {
	event := p.paramGesturePool.Get().(*ParamGestureEvent)
	atomic.AddUint64(&p.poolHits, 1)
	atomic.AddUint64(&p.currentAllocated, 1)
	return event
}

// PutParamGestureEvent returns a ParamGestureEvent to the pool
func (p *Pool) PutParamGestureEvent(event *ParamGestureEvent) {
	*event = ParamGestureEvent{}
	p.paramGesturePool.Put(event)
	atomic.AddUint64(&p.currentAllocated, ^uint64(0))
}

// GetNoteEvent gets a NoteEvent from the pool
func (p *Pool) GetNoteEvent() *NoteEvent {
	event := p.noteEventPool.Get().(*NoteEvent)
	atomic.AddUint64(&p.poolHits, 1)
	atomic.AddUint64(&p.currentAllocated, 1)
	return event
}

// PutNoteEvent returns a NoteEvent to the pool
func (p *Pool) PutNoteEvent(event *NoteEvent) {
	*event = NoteEvent{}
	p.noteEventPool.Put(event)
	atomic.AddUint64(&p.currentAllocated, ^uint64(0))
}

// GetNoteExpressionEvent gets a NoteExpressionEvent from the pool
func (p *Pool) GetNoteExpressionEvent() *NoteExpressionEvent {
	event := p.noteExpressionPool.Get().(*NoteExpressionEvent)
	atomic.AddUint64(&p.poolHits, 1)
	atomic.AddUint64(&p.currentAllocated, 1)
	return event
}

// PutNoteExpressionEvent returns a NoteExpressionEvent to the pool
func (p *Pool) PutNoteExpressionEvent(event *NoteExpressionEvent) {
	*event = NoteExpressionEvent{}
	p.noteExpressionPool.Put(event)
	atomic.AddUint64(&p.currentAllocated, ^uint64(0))
}

// GetTransportEvent gets a TransportEvent from the pool
func (p *Pool) GetTransportEvent() *TransportEvent {
	event := p.transportPool.Get().(*TransportEvent)
	atomic.AddUint64(&p.poolHits, 1)
	atomic.AddUint64(&p.currentAllocated, 1)
	return event
}

// PutTransportEvent returns a TransportEvent to the pool
func (p *Pool) PutTransportEvent(event *TransportEvent) {
	*event = TransportEvent{}
	p.transportPool.Put(event)
	atomic.AddUint64(&p.currentAllocated, ^uint64(0))
}

// GetMIDIEvent gets a MIDIEvent from the pool
func (p *Pool) GetMIDIEvent() *MIDIEvent {
	event := p.midiPool.Get().(*MIDIEvent)
	atomic.AddUint64(&p.poolHits, 1)
	atomic.AddUint64(&p.currentAllocated, 1)
	return event
}

// PutMIDIEvent returns a MIDIEvent to the pool
func (p *Pool) PutMIDIEvent(event *MIDIEvent) {
	*event = MIDIEvent{}
	p.midiPool.Put(event)
	atomic.AddUint64(&p.currentAllocated, ^uint64(0))
}

// GetMIDISysexEvent gets a MIDISysexEvent from the pool
func (p *Pool) GetMIDISysexEvent() *MIDISysexEvent {
	event := p.midiSysexPool.Get().(*MIDISysexEvent)
	atomic.AddUint64(&p.poolHits, 1)
	atomic.AddUint64(&p.currentAllocated, 1)
	return event
}

// PutMIDISysexEvent returns a MIDISysexEvent to the pool
func (p *Pool) PutMIDISysexEvent(event *MIDISysexEvent) {
	*event = MIDISysexEvent{}
	p.midiSysexPool.Put(event)
	atomic.AddUint64(&p.currentAllocated, ^uint64(0))
}

// GetMIDI2Event gets a MIDI2Event from the pool
func (p *Pool) GetMIDI2Event() *MIDI2Event {
	event := p.midi2Pool.Get().(*MIDI2Event)
	atomic.AddUint64(&p.poolHits, 1)
	atomic.AddUint64(&p.currentAllocated, 1)
	return event
}

// PutMIDI2Event returns a MIDI2Event to the pool
func (p *Pool) PutMIDI2Event(event *MIDI2Event) {
	*event = MIDI2Event{}
	p.midi2Pool.Put(event)
	atomic.AddUint64(&p.currentAllocated, ^uint64(0))
}

// GetDiagnostics returns pool diagnostics
func (p *Pool) GetDiagnostics() (totalAllocations, poolHits, poolMisses, highWaterMark, currentAllocated uint64) {
	return atomic.LoadUint64(&p.totalAllocations),
		atomic.LoadUint64(&p.poolHits),
		atomic.LoadUint64(&p.poolMisses),
		atomic.LoadUint64(&p.highWaterMark),
		atomic.LoadUint64(&p.currentAllocated)
}

// SetLogger sets the logger for pool diagnostics
func (p *Pool) SetLogger(logger *host.Logger) {
	p.logger = logger
}

// LogDiagnostics logs current pool diagnostics
func (p *Pool) LogDiagnostics() {
	if p.logger == nil {
		return
	}
	
	totalAllocations, poolHits, poolMisses, highWaterMark, currentAllocated := p.GetDiagnostics()
	hitRate := float64(0)
	if totalAllocations > 0 {
		hitRate = float64(poolHits) / float64(totalAllocations) * 100
	}
	
	p.logger.Debug(fmt.Sprintf("Event Pool Diagnostics: Total=%d Hits=%d Misses=%d HitRate=%.1f%% HighWaterMark=%d Current=%d",
		totalAllocations, poolHits, poolMisses, hitRate, highWaterMark, currentAllocated))
}