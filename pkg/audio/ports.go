package audio

import (
	"github.com/justyntemme/clapgo/pkg/api"
)

// PortProvider is the interface for audio port configuration
type PortProvider interface {
	GetAudioPortCount(isInput bool) uint32
	GetAudioPortInfo(index uint32, isInput bool) api.AudioPortInfo
}

// StereoPortProvider provides standard stereo input/output configuration
type StereoPortProvider struct {
	InputName  string
	OutputName string
}

// NewStereoPortProvider creates a standard stereo port provider
func NewStereoPortProvider() *StereoPortProvider {
	return &StereoPortProvider{
		InputName:  "Stereo Input",
		OutputName: "Stereo Output",
	}
}

// GetAudioPortCount returns 1 for stereo configuration
func (s *StereoPortProvider) GetAudioPortCount(isInput bool) uint32 {
	return 1
}

// GetAudioPortInfo returns stereo port information
func (s *StereoPortProvider) GetAudioPortInfo(index uint32, isInput bool) api.AudioPortInfo {
	if index != 0 {
		return api.AudioPortInfo{
			ID: api.InvalidID,
		}
	}
	
	name := s.OutputName
	if isInput {
		name = s.InputName
	}
	
	return api.CreateStereoPort(0, name, true)
}

// MonoPortProvider provides standard mono input/output configuration
type MonoPortProvider struct {
	InputName  string
	OutputName string
}

// NewMonoPortProvider creates a standard mono port provider
func NewMonoPortProvider() *MonoPortProvider {
	return &MonoPortProvider{
		InputName:  "Mono Input",
		OutputName: "Mono Output",
	}
}

// GetAudioPortCount returns 1 for mono configuration
func (m *MonoPortProvider) GetAudioPortCount(isInput bool) uint32 {
	return 1
}

// GetAudioPortInfo returns mono port information
func (m *MonoPortProvider) GetAudioPortInfo(index uint32, isInput bool) api.AudioPortInfo {
	if index != 0 {
		return api.AudioPortInfo{
			ID: api.InvalidID,
		}
	}
	
	name := m.OutputName
	if isInput {
		name = m.InputName
	}
	
	return api.CreateMonoPort(0, name, true)
}

// MultiPortProvider allows custom port configurations
type MultiPortProvider struct {
	InputPorts  []api.AudioPortInfo
	OutputPorts []api.AudioPortInfo
}

// GetAudioPortCount returns the number of ports
func (m *MultiPortProvider) GetAudioPortCount(isInput bool) uint32 {
	if isInput {
		return uint32(len(m.InputPorts))
	}
	return uint32(len(m.OutputPorts))
}

// GetAudioPortInfo returns port information by index
func (m *MultiPortProvider) GetAudioPortInfo(index uint32, isInput bool) api.AudioPortInfo {
	var ports []api.AudioPortInfo
	if isInput {
		ports = m.InputPorts
	} else {
		ports = m.OutputPorts
	}
	
	if index >= uint32(len(ports)) {
		return api.AudioPortInfo{
			ID: api.InvalidID,
		}
	}
	
	return ports[index]
}

// SurroundSupport provides common surround sound functionality
type SurroundSupport struct {
	SupportedMasks []uint64
}

// NewStereoSurroundSupport creates surround support for stereo-only plugins
func NewStereoSurroundSupport() *SurroundSupport {
	return &SurroundSupport{
		SupportedMasks: []uint64{api.ChannelMaskStereo},
	}
}

// IsChannelMaskSupported checks if a channel mask is supported
func (s *SurroundSupport) IsChannelMaskSupported(channelMask uint64) bool {
	for _, mask := range s.SupportedMasks {
		if mask == channelMask {
			return true
		}
	}
	return false
}

// GetChannelMap returns the channel map for stereo
func (s *SurroundSupport) GetChannelMap(isInput bool, portIndex uint32) []uint8 {
	if portIndex == 0 && len(s.SupportedMasks) > 0 && s.SupportedMasks[0] == api.ChannelMaskStereo {
		return api.CreateStereoChannelMap()
	}
	return nil
}