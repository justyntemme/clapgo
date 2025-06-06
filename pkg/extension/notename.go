package extension

// #include "../../include/clap/include/clap/ext/note-name.h"
import "C"
import "unsafe"

// NoteNameProvider is an extension for plugins that provide custom note names.
// It allows plugins to provide human-readable names for notes/keys.
type NoteNameProvider interface {
	// GetNoteNameCount returns the number of named notes.
	GetNoteNameCount() uint32

	// GetNoteName returns information about a named note.
	GetNoteName(index uint32) (NoteName, bool)
}

// NoteName represents a named note or key mapping
type NoteName struct {
	Name    string // The name of the note (up to 256 chars)
	Port    int16  // Port index, or -1 for all ports
	Channel int16  // Channel, or -1 for all channels
	Key     int16  // Key number (0-127)
}

// Common note names for standard chromatic scale
var ChromaticNoteNames = []string{
	"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B",
}

// GetStandardNoteName returns standard note name for a MIDI key number
func GetStandardNoteName(key int16) string {
	if key < 0 || key > 127 {
		return ""
	}
	octave := int(key / 12)
	note := ChromaticNoteNames[key%12]
	return note + string(rune('0'+octave))
}

// GetDrumNoteName returns common drum note names for GM drum mapping
func GetDrumNoteName(key int16) string {
	drumNames := map[int16]string{
		35: "Acoustic Bass Drum",
		36: "Bass Drum 1",
		37: "Side Stick",
		38: "Acoustic Snare",
		39: "Hand Clap",
		40: "Electric Snare",
		41: "Low Floor Tom",
		42: "Closed Hi Hat",
		43: "High Floor Tom",
		44: "Pedal Hi-Hat",
		45: "Low Tom",
		46: "Open Hi-Hat",
		47: "Low-Mid Tom",
		48: "Hi-Mid Tom",
		49: "Crash Cymbal 1",
		50: "High Tom",
		51: "Ride Cymbal 1",
		52: "Chinese Cymbal",
		53: "Ride Bell",
		54: "Tambourine",
		55: "Splash Cymbal",
		56: "Cowbell",
		57: "Crash Cymbal 2",
		58: "Vibraslap",
		59: "Ride Cymbal 2",
		60: "Hi Bongo",
		61: "Low Bongo",
		62: "Mute Hi Conga",
		63: "Open Hi Conga",
		64: "Low Conga",
		65: "High Timbale",
		66: "Low Timbale",
		67: "High Agogo",
		68: "Low Agogo",
		69: "Cabasa",
		70: "Maracas",
		71: "Short Whistle",
		72: "Long Whistle",
		73: "Short Guiro",
		74: "Long Guiro",
		75: "Claves",
		76: "Hi Wood Block",
		77: "Low Wood Block",
		78: "Mute Cuica",
		79: "Open Cuica",
		80: "Mute Triangle",
		81: "Open Triangle",
	}
	
	if name, exists := drumNames[key]; exists {
		return name
	}
	return GetStandardNoteName(key)
}

// StandardNoteNames provides standard note names (C-2 to G8) - compatible with api package
func StandardNoteNames() []NoteName {
	noteNames := []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
	names := make([]NoteName, 128)
	
	for i := 0; i < 128; i++ {
		noteName := noteNames[i%12]
		octave := (i / 12) - 2 // MIDI octave numbering starts at C-2
		names[i] = NoteName{
			Name:    noteName + string(rune('0'+octave)),
			Port:    -1, // All ports
			Key:     int16(i),
			Channel: -1, // All channels
		}
	}
	
	return names
}

// NoteNameToC converts a Go NoteName to a C structure
func NoteNameToC(noteName *NoteName, cNoteNamePtr unsafe.Pointer) {
	if noteName == nil || cNoteNamePtr == nil {
		return
	}
	
	cNoteName := (*C.clap_note_name_t)(cNoteNamePtr)
	
	// Copy name string
	nameBytes := []byte(noteName.Name)
	if len(nameBytes) >= C.CLAP_NAME_SIZE {
		nameBytes = nameBytes[:C.CLAP_NAME_SIZE-1]
	}
	for i, b := range nameBytes {
		cNoteName.name[i] = C.char(b)
	}
	cNoteName.name[len(nameBytes)] = 0
	
	// Copy port, key, and channel
	cNoteName.port = C.int16_t(noteName.Port)
	cNoteName.key = C.int16_t(noteName.Key)
	cNoteName.channel = C.int16_t(noteName.Channel)
}