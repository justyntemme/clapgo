package api

// AudioPortsActivationProvider is an extension for plugins that support
// activating and deactivating audio ports.
//
// This extension provides a way for the host to activate and deactivate audio ports.
// Deactivating a port provides the following benefits:
// - The plugin knows ahead of time that a given input is not present and can choose
//   an optimized computation path
// - The plugin knows that an output is not consumed by the host and doesn't need to
//   compute it
//
// Audio ports can only be activated or deactivated when the plugin is deactivated,
// unless CanActivateWhileProcessing() returns true.
//
// Audio buffers must still be provided if the audio port is deactivated.
// In such case, they shall be filled with 0 and the constant_mask shall be set.
//
// Audio ports are initially in the active state after creating the plugin instance.
// Audio ports state are not saved in the plugin state, so the host must restore
// the audio ports state after creating the plugin instance.
type AudioPortsActivationProvider interface {
	// CanActivateWhileProcessing returns true if the plugin supports
	// activation/deactivation while processing.
	// This is called on the main thread.
	CanActivateWhileProcessing() bool

	// SetActive activates or deactivates the given port.
	//
	// It is only possible to activate and deactivate on the audio thread if
	// CanActivateWhileProcessing() returns true.
	//
	// Parameters:
	// - isInput: true for input ports, false for output ports
	// - portIndex: the index of the port to activate/deactivate
	// - isActive: true to activate, false to deactivate
	// - sampleSize: indicates if the host will provide 32 bit or 64 bit audio buffers.
	//   Possible values are: 32, 64 or 0 if unspecified.
	//
	// Returns false if failed or invalid parameters.
	// This is called on the audio thread if CanActivateWhileProcessing() returns true,
	// otherwise on the main thread.
	SetActive(isInput bool, portIndex uint32, isActive bool, sampleSize uint32) bool
}

// AudioPortActivationState tracks the activation state of audio ports
type AudioPortActivationState struct {
	inputPortStates  map[uint32]bool
	outputPortStates map[uint32]bool
}

// NewAudioPortActivationState creates a new activation state tracker
func NewAudioPortActivationState() *AudioPortActivationState {
	return &AudioPortActivationState{
		inputPortStates:  make(map[uint32]bool),
		outputPortStates: make(map[uint32]bool),
	}
}

// SetPortActive updates the activation state of a port
func (a *AudioPortActivationState) SetPortActive(isInput bool, portIndex uint32, isActive bool) {
	if isInput {
		a.inputPortStates[portIndex] = isActive
	} else {
		a.outputPortStates[portIndex] = isActive
	}
}

// IsPortActive returns the activation state of a port
// Ports are active by default if not explicitly set
func (a *AudioPortActivationState) IsPortActive(isInput bool, portIndex uint32) bool {
	if isInput {
		if active, exists := a.inputPortStates[portIndex]; exists {
			return active
		}
	} else {
		if active, exists := a.outputPortStates[portIndex]; exists {
			return active
		}
	}
	// Ports are active by default
	return true
}

// ResetAllPorts resets all ports to their default active state
func (a *AudioPortActivationState) ResetAllPorts() {
	a.inputPortStates = make(map[uint32]bool)
	a.outputPortStates = make(map[uint32]bool)
}