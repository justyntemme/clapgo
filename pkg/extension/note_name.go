package extension

// NoteName provides note naming functionality for plugins
type NoteName struct {
	// NoteNames maps MIDI note numbers to custom names
	// Empty string means use default name
	NoteNames map[int16]string
}

// NewNoteName creates a new note name provider
func NewNoteName() *NoteName {
	return &NoteName{
		NoteNames: make(map[int16]string),
	}
}

// SetNoteName sets a custom name for a MIDI note
func (n *NoteName) SetNoteName(key int16, name string) {
	n.NoteNames[key] = name
}

// GetNoteName returns the custom name for a note, or empty string for default
func (n *NoteName) GetNoteName(port int16, channel int16, key int16) string {
	// In this simple implementation, we ignore port and channel
	// More complex implementations might have per-channel names
	if name, ok := n.NoteNames[key]; ok {
		return name
	}
	return ""
}

// GetNoteNameCount returns the number of custom note names
func (n *NoteName) GetNoteNameCount() uint32 {
	count := uint32(0)
	for _, name := range n.NoteNames {
		if name != "" {
			count++
		}
	}
	return count
}

// GetNoteNameByIndex returns note info by index for iteration
func (n *NoteName) GetNoteNameByIndex(index uint32) (port int16, channel int16, key int16, name string) {
	currentIndex := uint32(0)
	for k, v := range n.NoteNames {
		if v != "" {
			if currentIndex == index {
				return -1, -1, k, v // -1 means "all ports/channels"
			}
			currentIndex++
		}
	}
	return -1, -1, -1, ""
}

// Common drum note names for General MIDI
var GMDrumNames = map[int16]string{
	35: "Acoustic Bass Drum",
	36: "Bass Drum 1",
	37: "Side Stick",
	38: "Acoustic Snare",
	39: "Hand Clap",
	40: "Electric Snare",
	41: "Low Floor Tom",
	42: "Closed Hi-Hat",
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

// SetGMDrumNames sets General MIDI drum names
func (n *NoteName) SetGMDrumNames() {
	for key, name := range GMDrumNames {
		n.NoteNames[key] = name
	}
}