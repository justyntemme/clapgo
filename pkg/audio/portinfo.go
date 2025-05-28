package audio

// PortInfo contains information about an audio port
type PortInfo struct {
	// ID is a unique identifier for the port
	ID uint32

	// Name is a human-readable name for the port
	Name string

	// ChannelCount is the number of channels in this port
	ChannelCount uint32

	// Flags contains additional port flags
	Flags uint32

	// PortType describes the port type (e.g., "mono", "stereo")
	PortType string

	// InPlacePair is the ID of the in-place pair port or INVALID_ID if none
	InPlacePair uint32
}

// PortsProvider is an extension for plugins that have audio ports
// It allows hosts to query information about the plugin's audio ports
type PortsProvider interface {
	// GetAudioPortCount returns the number of audio ports
	GetAudioPortCount(isInput bool) uint32

	// GetAudioPortInfo returns information about an audio port
	GetAudioPortInfo(index uint32, isInput bool) PortInfo
}

// Common audio port constants
const (
	InvalidID = ^uint32(0)
)

// Common port types
const (
	PortTypeMono   = "mono"
	PortTypeStereo = "stereo"
)

// Audio port flags (to be imported from CLAP C headers)
const (
	PortFlagIsMain        = 1 << 0
	PortFlagSupports64Bit = 1 << 1
	PortFlagPrefers64Bit  = 1 << 2
	PortFlagRequiresCC    = 1 << 3
)

// Channel masks for surround sound
const (
	ChannelMaskMono   = 0x1
	ChannelMaskStereo = 0x3
)

// CreateMonoPort creates a mono audio port
func CreateMonoPort(id uint32, name string, isMain bool) PortInfo {
	flags := uint32(0)
	if isMain {
		flags |= PortFlagIsMain
	}
	
	return PortInfo{
		ID:           id,
		Name:         name,
		ChannelCount: 1,
		Flags:        flags,
		PortType:     PortTypeMono,
		InPlacePair:  InvalidID,
	}
}

// CreateStereoPort creates a stereo audio port
func CreateStereoPort(id uint32, name string, isMain bool) PortInfo {
	flags := uint32(0)
	if isMain {
		flags |= PortFlagIsMain
	}
	
	return PortInfo{
		ID:           id,
		Name:         name,
		ChannelCount: 2,
		Flags:        flags,
		PortType:     PortTypeStereo,
		InPlacePair:  InvalidID,
	}
}

// CreateStereoChannelMap returns a standard stereo channel map (L, R)
func CreateStereoChannelMap() []uint8 {
	return []uint8{0, 1} // Left, Right
}