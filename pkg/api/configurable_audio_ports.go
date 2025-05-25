package api

// ConfigurableAudioPortsProvider is an extension for plugins that support configurable audio ports.
// This allows the host to push audio port configurations to the plugin.
type ConfigurableAudioPortsProvider interface {
	// CanApplyConfiguration checks if the given configuration requests can be applied.
	// All requests must be valid for this to return true.
	// [main-thread && !active]
	CanApplyConfiguration(requests []AudioPortConfigurationRequest) bool

	// ApplyConfiguration atomically applies the given configuration requests.
	// Either all requests are applied or none are applied.
	// Returns true if the configuration was successfully applied.
	// [main-thread && !active]
	ApplyConfiguration(requests []AudioPortConfigurationRequest) bool
}

// AudioPortConfigurationRequest represents a request to configure an audio port.
type AudioPortConfigurationRequest struct {
	// IsInput identifies whether this is an input or output port
	IsInput bool

	// PortIndex is the index of the port to configure
	PortIndex uint32

	// ChannelCount is the requested number of channels
	ChannelCount uint32

	// PortType is the port type (e.g., "mono", "stereo", "surround", "ambisonic")
	PortType string

	// PortDetails contains type-specific configuration details:
	// - For "surround": []uint8 channel map
	// - For "ambisonic": *AmbisonicConfig
	// - For "mono" and "stereo": nil
	PortDetails interface{}
}

// AmbisonicConfig represents ambisonic-specific configuration details
type AmbisonicConfig struct {
	// Ordering defines the channel ordering (e.g., ACN, FuMa)
	Ordering uint32

	// Normalization defines the normalization scheme (e.g., SN3D, maxN)
	Normalization uint32
}

// Ambisonic ordering schemes
const (
	AmbisonicOrderingACN  = 0 // Ambisonic Channel Number
	AmbisonicOrderingFuMa = 1 // Furse-Malham
)

// Ambisonic normalization schemes
const (
	AmbisonicNormalizationSN3D = 0 // Schmidt semi-normalized
	AmbisonicNormalizationN3D  = 1 // Full 3D normalization
	AmbisonicNormalizationMaxN = 2 // MaxN normalization
	AmbisonicNormalizationFuMa = 3 // Furse-Malham normalization
)