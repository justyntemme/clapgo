package api

// #include <stdlib.h>
// #include <string.h>
// #include "../../include/clap/include/clap/ext/note-name.h"
//
// static inline const void* clap_host_get_extension_helper(const clap_host_t* host, const char* id) {
//     if (host && host->get_extension) {
//         return host->get_extension(host, id);
//     }
//     return NULL;
// }
//
// static inline void clap_host_note_name_changed(const clap_host_note_name_t* ext, const clap_host_t* host) {
//     if (ext && ext->changed) {
//         ext->changed(host);
//     }
// }
import "C"
import (
	"unsafe"
)

// Note name extension ID
const ExtNoteNameID = "clap.note-name"

// NoteName represents a named note or key mapping
type NoteName struct {
	Name    string // The name of the note (up to 256 chars)
	Port    int16  // Port index, or -1 for all ports
	Key     int16  // MIDI key number (0-127), or -1 for all keys
	Channel int16  // MIDI channel (0-15), or -1 for all channels
}

// NoteNameProvider is the interface that plugins implement to provide note names
type NoteNameProvider interface {
	// GetNoteNameCount returns the total number of note name mappings
	GetNoteNameCount() uint32
	
	// GetNoteName returns the note name at the given index
	// Returns nil if index is out of bounds
	GetNoteName(index uint32) *NoteName
}

// HostNoteName provides access to host note name functionality
type HostNoteName struct {
	host        unsafe.Pointer
	noteNameExt unsafe.Pointer
}

// NewHostNoteName creates a new host note name interface
func NewHostNoteName(host unsafe.Pointer) *HostNoteName {
	if host == nil {
		return nil
	}
	
	cHost := (*C.clap_host_t)(host)
	extPtr := C.clap_host_get_extension_helper(cHost, C.CString(ExtNoteNameID))
	
	if extPtr == nil {
		return nil
	}
	
	return &HostNoteName{
		host:        host,
		noteNameExt: extPtr,
	}
}

// NotifyChanged notifies the host that note names have changed
func (h *HostNoteName) NotifyChanged() {
	if h.noteNameExt == nil {
		return
	}
	
	ext := (*C.clap_host_note_name_t)(h.noteNameExt)
	C.clap_host_note_name_changed(ext, (*C.clap_host_t)(h.host))
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

// NoteNameFromC converts a C note name structure to Go
func NoteNameFromC(cNoteNamePtr unsafe.Pointer) *NoteName {
	if cNoteNamePtr == nil {
		return nil
	}
	
	cNoteName := (*C.clap_note_name_t)(cNoteNamePtr)
	
	// Convert name from C string
	name := C.GoString(&cNoteName.name[0])
	
	return &NoteName{
		Name:    name,
		Port:    int16(cNoteName.port),
		Key:     int16(cNoteName.key),
		Channel: int16(cNoteName.channel),
	}
}

// StandardDrumNoteNames provides standard General MIDI drum note names
var StandardDrumNoteNames = []NoteName{
	{Name: "Acoustic Bass Drum", Port: -1, Key: 35, Channel: 9},
	{Name: "Bass Drum 1", Port: -1, Key: 36, Channel: 9},
	{Name: "Side Stick", Port: -1, Key: 37, Channel: 9},
	{Name: "Acoustic Snare", Port: -1, Key: 38, Channel: 9},
	{Name: "Hand Clap", Port: -1, Key: 39, Channel: 9},
	{Name: "Electric Snare", Port: -1, Key: 40, Channel: 9},
	{Name: "Low Floor Tom", Port: -1, Key: 41, Channel: 9},
	{Name: "Closed Hi-Hat", Port: -1, Key: 42, Channel: 9},
	{Name: "High Floor Tom", Port: -1, Key: 43, Channel: 9},
	{Name: "Pedal Hi-Hat", Port: -1, Key: 44, Channel: 9},
	{Name: "Low Tom", Port: -1, Key: 45, Channel: 9},
	{Name: "Open Hi-Hat", Port: -1, Key: 46, Channel: 9},
	{Name: "Low-Mid Tom", Port: -1, Key: 47, Channel: 9},
	{Name: "Hi-Mid Tom", Port: -1, Key: 48, Channel: 9},
	{Name: "Crash Cymbal 1", Port: -1, Key: 49, Channel: 9},
	{Name: "High Tom", Port: -1, Key: 50, Channel: 9},
	{Name: "Ride Cymbal 1", Port: -1, Key: 51, Channel: 9},
	{Name: "Chinese Cymbal", Port: -1, Key: 52, Channel: 9},
	{Name: "Ride Bell", Port: -1, Key: 53, Channel: 9},
	{Name: "Tambourine", Port: -1, Key: 54, Channel: 9},
	{Name: "Splash Cymbal", Port: -1, Key: 55, Channel: 9},
	{Name: "Cowbell", Port: -1, Key: 56, Channel: 9},
	{Name: "Crash Cymbal 2", Port: -1, Key: 57, Channel: 9},
	{Name: "Vibraslap", Port: -1, Key: 58, Channel: 9},
	{Name: "Ride Cymbal 2", Port: -1, Key: 59, Channel: 9},
	{Name: "Hi Bongo", Port: -1, Key: 60, Channel: 9},
	{Name: "Low Bongo", Port: -1, Key: 61, Channel: 9},
	{Name: "Mute Hi Conga", Port: -1, Key: 62, Channel: 9},
	{Name: "Open Hi Conga", Port: -1, Key: 63, Channel: 9},
	{Name: "Low Conga", Port: -1, Key: 64, Channel: 9},
	{Name: "High Timbale", Port: -1, Key: 65, Channel: 9},
	{Name: "Low Timbale", Port: -1, Key: 66, Channel: 9},
	{Name: "High Agogo", Port: -1, Key: 67, Channel: 9},
	{Name: "Low Agogo", Port: -1, Key: 68, Channel: 9},
	{Name: "Cabasa", Port: -1, Key: 69, Channel: 9},
	{Name: "Maracas", Port: -1, Key: 70, Channel: 9},
	{Name: "Short Whistle", Port: -1, Key: 71, Channel: 9},
	{Name: "Long Whistle", Port: -1, Key: 72, Channel: 9},
	{Name: "Short Guiro", Port: -1, Key: 73, Channel: 9},
	{Name: "Long Guiro", Port: -1, Key: 74, Channel: 9},
	{Name: "Claves", Port: -1, Key: 75, Channel: 9},
	{Name: "Hi Wood Block", Port: -1, Key: 76, Channel: 9},
	{Name: "Low Wood Block", Port: -1, Key: 77, Channel: 9},
	{Name: "Mute Cuica", Port: -1, Key: 78, Channel: 9},
	{Name: "Open Cuica", Port: -1, Key: 79, Channel: 9},
	{Name: "Mute Triangle", Port: -1, Key: 80, Channel: 9},
	{Name: "Open Triangle", Port: -1, Key: 81, Channel: 9},
}

// StandardNoteNames provides standard note names (C-2 to G8)
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