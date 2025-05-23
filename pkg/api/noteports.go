package api

// NotePortManager manages note input and output ports for a plugin
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

// AddInputPort adds a note input port
func (npm *NotePortManager) AddInputPort(info NotePortInfo) {
	npm.inputPorts = append(npm.inputPorts, info)
}

// AddOutputPort adds a note output port
func (npm *NotePortManager) AddOutputPort(info NotePortInfo) {
	npm.outputPorts = append(npm.outputPorts, info)
}

// GetInputPortCount returns the number of note input ports
func (npm *NotePortManager) GetInputPortCount() uint32 {
	return uint32(len(npm.inputPorts))
}

// GetOutputPortCount returns the number of note output ports
func (npm *NotePortManager) GetOutputPortCount() uint32 {
	return uint32(len(npm.outputPorts))
}

// GetInputPort returns information about a specific input port
func (npm *NotePortManager) GetInputPort(index uint32) *NotePortInfo {
	if index >= uint32(len(npm.inputPorts)) {
		return nil
	}
	return &npm.inputPorts[index]
}

// GetOutputPort returns information about a specific output port
func (npm *NotePortManager) GetOutputPort(index uint32) *NotePortInfo {
	if index >= uint32(len(npm.outputPorts)) {
		return nil
	}
	return &npm.outputPorts[index]
}

// NotePortExtension represents the CLAP note ports extension
type NotePortExtension interface {
	// GetNotePortManager returns the plugin's note port manager
	GetNotePortManager() *NotePortManager
}

// CreateDefaultInstrumentPort creates a standard note input port for instruments
func CreateDefaultInstrumentPort() NotePortInfo {
	return NotePortInfo{
		ID:                0, // Main port
		Name:              "Note Input",
		SupportedDialects: NoteDialectCLAP | NoteDialectMIDI1,
		PreferredDialect:  NoteDialectCLAP,
		Flags:             NotePortIsMain,
	}
}

// CreateMIDIEffectPorts creates standard input/output ports for MIDI effects
func CreateMIDIEffectPorts() (input NotePortInfo, output NotePortInfo) {
	input = NotePortInfo{
		ID:                0,
		Name:              "MIDI In",
		SupportedDialects: NoteDialectCLAP | NoteDialectMIDI1,
		PreferredDialect:  NoteDialectCLAP,
		Flags:             NotePortIsMain,
	}
	output = NotePortInfo{
		ID:                1,
		Name:              "MIDI Out",
		SupportedDialects: NoteDialectCLAP | NoteDialectMIDI1,
		PreferredDialect:  NoteDialectCLAP,
		Flags:             0,
	}
	return
}