package event

import (
	"unsafe"
)

// Event types
const (
	TypeNoteOn          uint32 = 0
	TypeNoteOff         uint32 = 1
	TypeNoteChoke       uint32 = 2
	TypeNoteEnd         uint32 = 3
	TypeNoteExpression  uint32 = 4
	TypeParamValue      uint32 = 5
	TypeParamMod        uint32 = 6
	TypeParamGestureBegin uint32 = 7
	TypeParamGestureEnd   uint32 = 8
	TypeTransport       uint32 = 9
	TypeMIDI            uint32 = 10
	TypeMIDISysex       uint32 = 11
	TypeMIDI2           uint32 = 12
)

// Event flags
const (
	FlagIsLive      uint32 = 1 << 0
	FlagDontRecord  uint32 = 1 << 1
)

// Header contains common event metadata
type Header struct {
	Size     uint32
	Time     uint32
	SpaceID  uint16
	Type     uint16
	Flags    uint32
}

// Event is the base interface for all events
type Event interface {
	GetHeader() *Header
}

// ParamValueEvent represents a parameter value change
type ParamValueEvent struct {
	Header  Header
	ParamID uint32
	Cookie  unsafe.Pointer
	NoteID  int32
	Port    int16
	Channel int16
	Key     int16
	Value   float64
}

func (e *ParamValueEvent) GetHeader() *Header { return &e.Header }

// ParamModEvent represents a parameter modulation
type ParamModEvent struct {
	Header  Header
	ParamID uint32
	Cookie  unsafe.Pointer
	NoteID  int32
	Port    int16
	Channel int16
	Key     int16
	Amount  float64
}

func (e *ParamModEvent) GetHeader() *Header { return &e.Header }

// ParamGestureEvent represents parameter gesture begin/end
type ParamGestureEvent struct {
	Header  Header
	ParamID uint32
}

func (e *ParamGestureEvent) GetHeader() *Header { return &e.Header }

// NoteEvent represents a note on/off/choke/end event
type NoteEvent struct {
	Header   Header
	NoteID   int32
	Port     int16
	Channel  int16
	Key      int16
	Velocity float64
}

func (e *NoteEvent) GetHeader() *Header { return &e.Header }

// NoteExpressionEvent represents note expression changes
type NoteExpressionEvent struct {
	Header       Header
	ExpressionID uint32
	NoteID       int32
	Port         int16
	Channel      int16
	Key          int16
	Value        float64
}

func (e *NoteExpressionEvent) GetHeader() *Header { return &e.Header }

// Note expression types
const (
	NoteExpressionVolume     uint32 = 0
	NoteExpressionPan        uint32 = 1
	NoteExpressionTuning     uint32 = 2
	NoteExpressionVibrato    uint32 = 3
	NoteExpressionExpression uint32 = 4
	NoteExpressionBrightness uint32 = 5
	NoteExpressionPressure   uint32 = 6
)

// TransportEvent represents transport state changes
type TransportEvent struct {
	Header               Header
	Flags                uint32
	SongPosBeats         float64  // position in beats
	SongPosSeconds       float64  // position in seconds
	Tempo                float64  // in BPM
	TempoInc             float64  // tempo increment for each sample
	LoopStartBeats       float64
	LoopEndBeats         float64
	LoopStartSeconds     float64
	LoopEndSeconds       float64
	BarStart             float64  // bar start in beats
	BarNumber            int32    // bar number
	TimeSignatureNum     uint16   // time signature numerator
	TimeSignatureDenom   uint16   // time signature denominator
}

func (e *TransportEvent) GetHeader() *Header { return &e.Header }

// Transport flags
const (
	TransportHasTempo          uint32 = 1 << 0
	TransportHasBeatsTime      uint32 = 1 << 1
	TransportHasSecondsTime    uint32 = 1 << 2
	TransportHasTimeSignature  uint32 = 1 << 3
	TransportIsPlaying         uint32 = 1 << 4
	TransportIsRecording       uint32 = 1 << 5
	TransportIsLooping         uint32 = 1 << 6
	TransportIsWithinPreRoll   uint32 = 1 << 7
)

// MIDIEvent represents a MIDI 1.0 event
type MIDIEvent struct {
	Header Header
	Port   uint16
	Data   [3]byte
}

func (e *MIDIEvent) GetHeader() *Header { return &e.Header }

// MIDISysexEvent represents a MIDI System Exclusive event
type MIDISysexEvent struct {
	Header Header
	Port   uint16
	Buffer unsafe.Pointer
	Size   uint32
}

func (e *MIDISysexEvent) GetHeader() *Header { return &e.Header }

// MIDI2Event represents a MIDI 2.0 event
type MIDI2Event struct {
	Header Header
	Port   uint16
	Data   [4]uint32
}

func (e *MIDI2Event) GetHeader() *Header { return &e.Header }

// Handler processes events with type-specific methods
type Handler interface {
	// Parameter events
	HandleParamValue(event *ParamValueEvent, time uint32)
	HandleParamMod(event *ParamModEvent, time uint32)
	HandleParamGestureBegin(event *ParamGestureEvent, time uint32)
	HandleParamGestureEnd(event *ParamGestureEvent, time uint32)
	
	// Note events
	HandleNoteOn(event *NoteEvent, time uint32)
	HandleNoteOff(event *NoteEvent, time uint32)
	HandleNoteChoke(event *NoteEvent, time uint32)
	HandleNoteEnd(event *NoteEvent, time uint32)
	HandleNoteExpression(event *NoteExpressionEvent, time uint32)
	
	// Transport events
	HandleTransport(event *TransportEvent, time uint32)
	
	// MIDI events
	HandleMIDI(event *MIDIEvent, time uint32)
	HandleMIDI2(event *MIDI2Event, time uint32)
	HandleMIDISysex(event *MIDISysexEvent, time uint32)
}

// NoOpHandler provides default no-op implementations for all event handler methods
// Embed this struct in your plugin to avoid implementing unused methods
type NoOpHandler struct{}

// Parameter events
func (h *NoOpHandler) HandleParamValue(event *ParamValueEvent, time uint32) {}
func (h *NoOpHandler) HandleParamMod(event *ParamModEvent, time uint32) {}
func (h *NoOpHandler) HandleParamGestureBegin(event *ParamGestureEvent, time uint32) {}
func (h *NoOpHandler) HandleParamGestureEnd(event *ParamGestureEvent, time uint32) {}

// Note events
func (h *NoOpHandler) HandleNoteOn(event *NoteEvent, time uint32) {}
func (h *NoOpHandler) HandleNoteOff(event *NoteEvent, time uint32) {}
func (h *NoOpHandler) HandleNoteChoke(event *NoteEvent, time uint32) {}
func (h *NoOpHandler) HandleNoteEnd(event *NoteEvent, time uint32) {}
func (h *NoOpHandler) HandleNoteExpression(event *NoteExpressionEvent, time uint32) {}

// Transport events
func (h *NoOpHandler) HandleTransport(event *TransportEvent, time uint32) {}

// MIDI events
func (h *NoOpHandler) HandleMIDI(event *MIDIEvent, time uint32) {}
func (h *NoOpHandler) HandleMIDI2(event *MIDI2Event, time uint32) {}
func (h *NoOpHandler) HandleMIDISysex(event *MIDISysexEvent, time uint32) {}