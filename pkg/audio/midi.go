package audio

import (
	"github.com/justyntemme/clapgo/pkg/event"
	"github.com/justyntemme/clapgo/pkg/param"
)

// MIDIProcessor handles MIDI events for synthesizers
type MIDIProcessor struct {
	voiceManager *VoiceManager
	paramManager *param.Manager
	
	// Pitch bend range in semitones
	pitchBendRange float64
	
	// Callbacks for custom processing
	onNoteOn       func(channel, key int16, velocity float64)
	onNoteOff      func(channel, key int16)
	onModulation   func(channel int16, cc uint32, value float64)
	onPitchBend    func(channel int16, value float64)
	onPolyPressure func(channel, key int16, pressure float64)
}

// NewMIDIProcessor creates a new MIDI processor
func NewMIDIProcessor(voiceManager *VoiceManager, paramManager *param.Manager) *MIDIProcessor {
	return &MIDIProcessor{
		voiceManager:   voiceManager,
		paramManager:   paramManager,
		pitchBendRange: 2.0, // Default to 2 semitones
	}
}

// SetPitchBendRange sets the pitch bend range in semitones
func (m *MIDIProcessor) SetPitchBendRange(semitones float64) {
	m.pitchBendRange = semitones
}

// SetCallbacks sets optional callbacks for custom processing
func (m *MIDIProcessor) SetCallbacks(
	onNoteOn func(channel, key int16, velocity float64),
	onNoteOff func(channel, key int16),
	onModulation func(channel int16, cc uint32, value float64),
	onPitchBend func(channel int16, value float64),
	onPolyPressure func(channel, key int16, pressure float64),
) {
	m.onNoteOn = onNoteOn
	m.onNoteOff = onNoteOff
	m.onModulation = onModulation
	m.onPitchBend = onPitchBend
	m.onPolyPressure = onPolyPressure
}

// ProcessEvents processes MIDI events from an event processor
func (m *MIDIProcessor) ProcessEvents(events *event.Processor, handler event.Handler) {
	if events == nil {
		return
	}
	
	// Create a MIDI event handler that wraps our processor
	midiHandler := &midiEventHandler{
		processor: m,
		handler:   handler,
	}
	
	events.ProcessAll(midiHandler)
}

// ProcessNoteOn handles a note on event
func (m *MIDIProcessor) ProcessNoteOn(channel, key int16, velocity float64, noteID int32) {
	// Allocate a voice for this note
	voice := m.voiceManager.AllocateVoice(noteID, channel, key, velocity)
	
	// Call custom callback if provided
	if m.onNoteOn != nil {
		m.onNoteOn(channel, key, velocity)
	}
	
	// Voice allocation handled by VoiceManager
	_ = voice
}

// ProcessNoteOff handles a note off event
func (m *MIDIProcessor) ProcessNoteOff(channel, key int16, noteID int32) {
	// Release the voice
	m.voiceManager.ReleaseVoice(noteID, channel)
	
	// Call custom callback if provided
	if m.onNoteOff != nil {
		m.onNoteOff(channel, key)
	}
}

// ProcessPitchBend handles pitch bend events
func (m *MIDIProcessor) ProcessPitchBend(channel int16, value float64) {
	// Convert to semitones based on pitch bend range
	pitchBendSemitones := value * m.pitchBendRange
	
	// Apply to all voices on this channel
	m.voiceManager.ApplyToAllVoices(func(voice *Voice) {
		if voice.Channel == channel {
			voice.PitchBend = pitchBendSemitones
		}
	})
	
	// Call custom callback if provided
	if m.onPitchBend != nil {
		m.onPitchBend(channel, value)
	}
}

// ProcessModulation handles CC modulation events
func (m *MIDIProcessor) ProcessModulation(channel int16, cc uint32, value float64) {
	// Handle common CCs
	switch cc {
	case 1: // Mod wheel - typically controls vibrato or filter
		m.voiceManager.ApplyToAllVoices(func(voice *Voice) {
			if voice.Channel == channel {
				// Use mod wheel for brightness by default
				voice.Brightness = value
			}
		})
		
	case 7: // Volume
		m.voiceManager.ApplyToAllVoices(func(voice *Voice) {
			if voice.Channel == channel {
				voice.Volume = value
			}
		})
		
	case 64: // Sustain pedal
		if value > 0.5 {
			// Sustain on - don't release notes
			// This would need additional logic in voice manager
		} else {
			// Sustain off - release any held notes
			// This would need tracking of which notes are held
		}
		
	case 74: // Filter cutoff (brightness)
		m.voiceManager.ApplyToAllVoices(func(voice *Voice) {
			if voice.Channel == channel {
				voice.Brightness = value
			}
		})
	}
	
	// Call custom callback if provided
	if m.onModulation != nil {
		m.onModulation(channel, cc, value)
	}
}

// ProcessPolyPressure handles polyphonic aftertouch
func (m *MIDIProcessor) ProcessPolyPressure(channel, key int16, pressure float64) {
	// Apply pressure to specific note
	m.voiceManager.ApplyToAllVoices(func(voice *Voice) {
		if voice.Channel == channel && voice.Key == key {
			voice.Pressure = pressure
		}
	})
	
	// Call custom callback if provided
	if m.onPolyPressure != nil {
		m.onPolyPressure(channel, key, pressure)
	}
}

// ProcessChannelPressure handles channel aftertouch
func (m *MIDIProcessor) ProcessChannelPressure(channel int16, pressure float64) {
	// Apply pressure to all notes on channel
	m.voiceManager.ApplyToAllVoices(func(voice *Voice) {
		if voice.Channel == channel {
			voice.Pressure = pressure
		}
	})
}

// ProcessAllNotesOff handles all notes off message
func (m *MIDIProcessor) ProcessAllNotesOff(channel int16) {
	m.voiceManager.ApplyToAllVoices(func(voice *Voice) {
		if voice.Channel == channel && voice.IsActive {
			if voice.Envelope != nil {
				voice.Envelope.Release()
			}
		}
	})
}

// ProcessAllSoundOff immediately stops all sound on a channel
func (m *MIDIProcessor) ProcessAllSoundOff(channel int16) {
	m.voiceManager.ApplyToAllVoices(func(voice *Voice) {
		if voice.Channel == channel {
			voice.IsActive = false
			if voice.Envelope != nil {
				voice.Envelope.Reset()
			}
		}
	})
}

// midiEventHandler implements event.Handler for MIDI events
type midiEventHandler struct {
	processor *MIDIProcessor
	handler   event.Handler // Original handler for non-MIDI events
}

func (h *midiEventHandler) HandleNoteOn(e *event.NoteEvent, time uint32) {
	h.processor.ProcessNoteOn(e.Channel, e.Key, e.Velocity, e.NoteID)
}

func (h *midiEventHandler) HandleNoteOff(e *event.NoteEvent, time uint32) {
	h.processor.ProcessNoteOff(e.Channel, e.Key, e.NoteID)
}

func (h *midiEventHandler) HandleNoteChoke(e *event.NoteEvent, time uint32) {
	// Immediately stop the note
	voice := h.processor.voiceManager.GetVoiceByNoteID(e.NoteID, e.Channel)
	if voice != nil {
		voice.IsActive = false
		if voice.Envelope != nil {
			voice.Envelope.Reset()
		}
	}
}

func (h *midiEventHandler) HandleNoteEnd(e *event.NoteEvent, time uint32) {
	// Note end is similar to note off but indicates the natural end of a note
	h.processor.ProcessNoteOff(e.Channel, e.Key, e.NoteID)
}

func (h *midiEventHandler) HandleNoteExpression(e *event.NoteExpressionEvent, time uint32) {
	// Apply expression to the specific note
	voice := h.processor.voiceManager.GetVoiceByNoteID(e.NoteID, e.Channel)
	if voice != nil {
		switch e.ExpressionID {
		case event.NoteExpressionVolume:
			voice.Volume = e.Value
		case event.NoteExpressionPan:
			// Pan would need to be added to Voice struct if needed
		case event.NoteExpressionTuning:
			// Tuning adjustment in semitones
			voice.PitchBend = e.Value
		case event.NoteExpressionVibrato:
			// Vibrato would need additional implementation
		case event.NoteExpressionBrightness:
			voice.Brightness = e.Value
		case event.NoteExpressionPressure:
			voice.Pressure = e.Value
		}
	}
}

func (h *midiEventHandler) HandleParamValue(e *event.ParamValueEvent, time uint32) {
	// Delegate to original handler
	if h.handler != nil {
		h.handler.HandleParamValue(e, time)
	}
}

func (h *midiEventHandler) HandleParamMod(e *event.ParamModEvent, time uint32) {
	// Delegate to original handler
	if h.handler != nil {
		h.handler.HandleParamMod(e, time)
	}
}

func (h *midiEventHandler) HandleParamGestureBegin(e *event.ParamGestureEvent, time uint32) {
	// Delegate to original handler
	if h.handler != nil {
		h.handler.HandleParamGestureBegin(e, time)
	}
}

func (h *midiEventHandler) HandleParamGestureEnd(e *event.ParamGestureEvent, time uint32) {
	// Delegate to original handler
	if h.handler != nil {
		h.handler.HandleParamGestureEnd(e, time)
	}
}

func (h *midiEventHandler) HandleTransport(e *event.TransportEvent, time uint32) {
	// Delegate to original handler
	if h.handler != nil {
		h.handler.HandleTransport(e, time)
	}
}

func (h *midiEventHandler) HandleMIDI(e *event.MIDIEvent, time uint32) {
	// Handle raw MIDI events
	if len(e.Data) < 1 {
		return
	}
	
	status := e.Data[0] & 0xF0
	channel := int16(e.Data[0] & 0x0F)
	
	switch status {
	case 0x80: // Note off
		if len(e.Data) >= 3 {
			key := int16(e.Data[1])
			// velocity := e.Data[2] // Often ignored for note off
			h.processor.ProcessNoteOff(channel, key, 0)
		}
		
	case 0x90: // Note on
		if len(e.Data) >= 3 {
			key := int16(e.Data[1])
			velocity := float64(e.Data[2]) / 127.0
			if velocity == 0 {
				// Note on with velocity 0 is treated as note off
				h.processor.ProcessNoteOff(channel, key, 0)
			} else {
				h.processor.ProcessNoteOn(channel, key, velocity, 0)
			}
		}
		
	case 0xA0: // Polyphonic aftertouch
		if len(e.Data) >= 3 {
			key := int16(e.Data[1])
			pressure := float64(e.Data[2]) / 127.0
			h.processor.ProcessPolyPressure(channel, key, pressure)
		}
		
	case 0xB0: // Control change
		if len(e.Data) >= 3 {
			cc := uint32(e.Data[1])
			value := float64(e.Data[2]) / 127.0
			
			// Handle special CCs
			switch cc {
			case 123: // All notes off
				h.processor.ProcessAllNotesOff(channel)
			case 120: // All sound off
				h.processor.ProcessAllSoundOff(channel)
			default:
				h.processor.ProcessModulation(channel, cc, value)
			}
		}
		
	case 0xD0: // Channel pressure
		if len(e.Data) >= 2 {
			pressure := float64(e.Data[1]) / 127.0
			h.processor.ProcessChannelPressure(channel, pressure)
		}
		
	case 0xE0: // Pitch bend
		if len(e.Data) >= 3 {
			// Pitch bend is 14-bit, centered at 0x2000
			bend := int(e.Data[1]) | (int(e.Data[2]) << 7)
			// Convert to -1.0 to 1.0 range
			bendValue := (float64(bend) - 8192.0) / 8192.0
			h.processor.ProcessPitchBend(channel, bendValue)
		}
	}
	
	// Also delegate to original handler if it wants raw MIDI
	if h.handler != nil {
		h.handler.HandleMIDI(e, time)
	}
}

func (h *midiEventHandler) HandleMIDISysex(e *event.MIDISysexEvent, time uint32) {
	// Delegate to original handler
	if h.handler != nil {
		h.handler.HandleMIDISysex(e, time)
	}
}

func (h *midiEventHandler) HandleMIDI2(e *event.MIDI2Event, time uint32) {
	// Delegate to original handler
	if h.handler != nil {
		h.handler.HandleMIDI2(e, time)
	}
}