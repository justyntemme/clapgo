package audio

// Note port constants for CLAP note ports extension
const (
	// Note dialects
	NoteDialectClap = 1 << 0  // CLAP note events
	NoteDialectMidi = 1 << 1  // MIDI 1.0 events
	NoteDialectMidi2 = 1 << 2 // MIDI 2.0 events
)

// NotePortInfo contains information about a note port
type NotePortInfo struct {
	ID                 uint32
	Name               string
	SupportedDialects  uint32
	PreferredDialect   uint32
}

// NotePortManager manages note input/output ports for a plugin
type NotePortManager struct {
	inputPorts  []NotePortInfo
	outputPorts []NotePortInfo
}

// NewNotePortManager creates a new note port manager
func NewNotePortManager() *NotePortManager {
	return &NotePortManager{
		inputPorts:  make([]NotePortInfo, 0),
		outputPorts: make([]NotePortInfo, 0),
	}
}

// AddInputPort adds an input note port
func (m *NotePortManager) AddInputPort(port NotePortInfo) {
	m.inputPorts = append(m.inputPorts, port)
}

// AddOutputPort adds an output note port
func (m *NotePortManager) AddOutputPort(port NotePortInfo) {
	m.outputPorts = append(m.outputPorts, port)
}

// GetInputPortCount returns the number of input note ports
func (m *NotePortManager) GetInputPortCount() uint32 {
	return uint32(len(m.inputPorts))
}

// GetOutputPortCount returns the number of output note ports
func (m *NotePortManager) GetOutputPortCount() uint32 {
	return uint32(len(m.outputPorts))
}

// GetInputPort returns an input port by index
func (m *NotePortManager) GetInputPort(index uint32) *NotePortInfo {
	if index >= uint32(len(m.inputPorts)) {
		return nil
	}
	return &m.inputPorts[index]
}

// GetOutputPort returns an output port by index
func (m *NotePortManager) GetOutputPort(index uint32) *NotePortInfo {
	if index >= uint32(len(m.outputPorts)) {
		return nil
	}
	return &m.outputPorts[index]
}

// CreateDefaultInstrumentPort creates a standard instrument input port
func CreateDefaultInstrumentPort() NotePortInfo {
	return NotePortInfo{
		ID:                0,
		Name:              "Note Input",
		SupportedDialects: NoteDialectClap | NoteDialectMidi,
		PreferredDialect:  NoteDialectClap,
	}
}