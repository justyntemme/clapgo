package extension

// AudioPortsActivationProvider is an extension for plugins that support
// activating and deactivating audio ports.
type AudioPortsActivationProvider interface {
	// CanActivateWhileProcessing returns true if the plugin supports
	// activation/deactivation while processing.
	CanActivateWhileProcessing() bool

	// SetActive activates or deactivates the given port.
	SetActive(portIndex uint32, isInput bool, isActive bool) bool
}

// ConfigurableAudioPortsProvider is an extension for plugins with configurable audio ports.
type ConfigurableAudioPortsProvider interface {
	// CanApplyConfiguration returns true if the configuration can be applied.
	CanApplyConfiguration(configID uint64) bool

	// ApplyConfiguration applies the given configuration.
	ApplyConfiguration(configID uint64) bool
}

// AudioPortsConfigProvider is an extension for plugins that support audio port configurations.
type AudioPortsConfigProvider interface {
	// GetConfigCount returns the number of available configurations.
	GetConfigCount() uint32

	// GetConfig returns information about a configuration.
	GetConfig(index uint32) (AudioPortsConfig, bool)

	// GetCurrentConfig returns the ID of the current configuration.
	GetCurrentConfig() uint64

	// SelectConfig selects a configuration.
	SelectConfig(configID uint64) bool
}

// AudioPortsConfig represents an audio port configuration
type AudioPortsConfig struct {
	ID               uint64
	Name             string
	InputPortCount   uint32
	OutputPortCount  uint32
	HasMainInput     bool
	MainInputChannelCount  uint32
	MainInputPortType      string
	HasMainOutput    bool
	MainOutputChannelCount uint32
	MainOutputPortType     string
}

// SurroundProvider is an extension for plugins that support surround sound.
type SurroundProvider interface {
	// IsChannelMaskSupported returns true if the channel mask is supported.
	IsChannelMaskSupported(channelMask uint64) bool

	// GetChannelMap returns the channel map for the given port.
	GetChannelMap(isInput bool, portIndex uint32, channelMask uint64) []uint8
}

// AmbisonicProvider is an extension for plugins that support ambisonic audio.
type AmbisonicProvider interface {
	// IsChannelMaskSupported returns true if the ambisonic channel mask is supported.
	IsChannelMaskSupported(channelMask uint64) bool

	// GetChannelMap returns the ambisonic channel map.
	GetChannelMap(isInput bool, portIndex uint32, channelMask uint64) []uint8
}

// Common surround channel masks
const (
	ChannelMaskMono              = 0x1
	ChannelMaskStereo            = 0x3
	ChannelMask2_1               = 0x103
	ChannelMask3_0               = 0x7
	ChannelMask3_1               = 0x107
	ChannelMask4_0               = 0x107
	ChannelMask4_1               = 0x10F
	ChannelMask5_0               = 0x37
	ChannelMask5_1               = 0x3F
	ChannelMask6_0               = 0x137
	ChannelMask6_1               = 0x13F
	ChannelMask7_0               = 0x637
	ChannelMask7_1               = 0x63F
)

// Common ambisonic orders
const (
	AmbisonicOrder0 = 0  // Mono (W)
	AmbisonicOrder1 = 1  // First order (W, X, Y, Z)
	AmbisonicOrder2 = 2  // Second order (9 channels)
	AmbisonicOrder3 = 3  // Third order (16 channels)
)