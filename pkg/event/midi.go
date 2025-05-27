package event

// MIDI status bytes
const (
	MIDINoteOff         byte = 0x80
	MIDINoteOn          byte = 0x90
	MIDIPolyPressure    byte = 0xA0
	MIDIControlChange   byte = 0xB0
	MIDIProgramChange   byte = 0xC0
	MIDIChannelPressure byte = 0xD0
	MIDIPitchBend       byte = 0xE0
	MIDISystemExclusive byte = 0xF0
)

// MIDIToNoteOn converts a MIDI 1.0 note on message to a CLAP note event
func MIDIToNoteOn(data [3]byte, port, channel int16) *NoteEvent {
	status := data[0] & 0xF0
	if status != MIDINoteOn || data[2] == 0 { // velocity 0 is note off
		return nil
	}
	
	return &NoteEvent{
		Header: Header{
			Type: uint16(TypeNoteOn),
		},
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
	if status == MIDINoteOff || (status == MIDINoteOn && data[2] == 0) {
		return &NoteEvent{
			Header: Header{
				Type: uint16(TypeNoteOff),
			},
			NoteID:   -1, // Will be assigned by host
			Port:     port,
			Channel:  channel,
			Key:      int16(data[1]),
			Velocity: float64(data[2]) / 127.0,
		}
	}
	return nil
}

// NoteToMIDI converts a CLAP note event to MIDI 1.0 data
func NoteToMIDI(event *NoteEvent) (data [3]byte, ok bool) {
	if event.Key < 0 || event.Key > 127 {
		return data, false
	}
	
	channel := byte(event.Channel & 0x0F)
	velocity := byte(event.Velocity * 127.0)
	
	switch event.Header.Type {
	case uint16(TypeNoteOn):
		data[0] = MIDINoteOn | channel
		data[1] = byte(event.Key)
		data[2] = velocity
		ok = true
		
	case uint16(TypeNoteOff):
		data[0] = MIDINoteOff | channel
		data[1] = byte(event.Key)
		data[2] = velocity
		ok = true
	}
	
	return
}

// MIDIControlChangeToParamValue converts a MIDI CC to a parameter value event
func MIDIControlChangeToParamValue(data [3]byte, paramID uint32, port, channel int16) *ParamValueEvent {
	status := data[0] & 0xF0
	if status != MIDIControlChange {
		return nil
	}
	
	return &ParamValueEvent{
		Header: Header{
			Type: uint16(TypeParamValue),
		},
		ParamID: paramID,
		Port:    port,
		Channel: channel,
		Value:   float64(data[2]) / 127.0,
	}
}

// MIDIPitchBendToParamMod converts MIDI pitch bend to parameter modulation
func MIDIPitchBendToParamMod(data [3]byte, paramID uint32, port, channel int16) *ParamModEvent {
	status := data[0] & 0xF0
	if status != MIDIPitchBend {
		return nil
	}
	
	// Pitch bend is 14-bit: data[1] is LSB, data[2] is MSB
	pitchBend := int(data[1]) | (int(data[2]) << 7)
	// Convert from 0-16383 to -1.0 to 1.0
	amount := (float64(pitchBend) - 8192.0) / 8192.0
	
	return &ParamModEvent{
		Header: Header{
			Type: uint16(TypeParamMod),
		},
		ParamID: paramID,
		Port:    port,
		Channel: channel,
		Amount:  amount,
	}
}