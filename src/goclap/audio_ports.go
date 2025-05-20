package goclap

// #include <stdint.h>
// #include <stdlib.h>
// #include "../../include/clap/include/clap/clap.h"
import "C"
// No imports needed yet

// Port type constants - hardcoded values from CLAP headers
const (
	PortTypeAudio    = uint32(0) // CLAP_PORT_AUDIO
	PortTypeCv       = uint32(1) // CLAP_PORT_CV  
	PortTypeControl  = uint32(2) // CLAP_PORT_CONTROL
)

// Audio port flags - hardcoded values from CLAP headers
const (
	PortIsMain           = uint32(1 << 0) // CLAP_AUDIO_PORT_IS_MAIN
	PortSupports64Bits   = uint32(1 << 1) // CLAP_AUDIO_PORT_SUPPORTS_64BITS
	PortPrefers64Bits    = uint32(1 << 2) // CLAP_AUDIO_PORT_PREFERS_64BITS
	PortSupportsMono     = uint32(1 << 3) // CLAP_AUDIO_PORT_SUPPORTS_MONO
	PortPrefersMono      = uint32(1 << 4) // CLAP_AUDIO_PORT_PREFERS_MONO
	PortSupportsStereo   = uint32(1 << 5) // CLAP_AUDIO_PORT_SUPPORTS_STEREO
	PortPrefersStereo    = uint32(1 << 6) // CLAP_AUDIO_PORT_PREFERS_STEREO
)

// AudioPort represents a CLAP audio port
type AudioPort struct {
	ID          uint32
	Name        string
	Flags       uint32
	ChannelCount uint32
	PortType    uint32
	InPlace     bool
}

// AudioPortInfo holds info about all audio ports
type AudioPortInfo struct {
	InputPorts  []AudioPort
	OutputPorts []AudioPort
}

// AudioPortsExtension implements the clap_plugin_audio_ports extension
type AudioPortsExtension struct {
	Processor AudioProcessor
	Info      AudioPortInfo
}

// GetCount returns the number of ports
func (e *AudioPortsExtension) GetCount(isInput bool) uint32 {
	if isInput {
		return uint32(len(e.Info.InputPorts))
	} else {
		return uint32(len(e.Info.OutputPorts))
	}
}

// Get retrieves info about a port
func (e *AudioPortsExtension) Get(index uint32, isInput bool) *AudioPort {
	if isInput {
		if int(index) < len(e.Info.InputPorts) {
			return &e.Info.InputPorts[index]
		}
	} else {
		if int(index) < len(e.Info.OutputPorts) {
			return &e.Info.OutputPorts[index]
		}
	}
	return nil
}

// CreateAudioPortsExtension creates a new audio ports extension with default stereo in/out
func CreateAudioPortsExtension(processor AudioProcessor) *AudioPortsExtension {
	// Create default stereo in/out ports
	ext := &AudioPortsExtension{
		Processor: processor,
		Info: AudioPortInfo{
			InputPorts: []AudioPort{
				{
					ID:           0,
					Name:         "Stereo In",
					Flags:        PortIsMain | PortPrefersStereo,
					ChannelCount: 2,
					PortType:     PortTypeAudio,
					InPlace:      true,
				},
			},
			OutputPorts: []AudioPort{
				{
					ID:           0, 
					Name:         "Stereo Out",
					Flags:        PortIsMain | PortPrefersStereo,
					ChannelCount: 2,
					PortType:     PortTypeAudio,
					InPlace:      true,
				},
			},
		},
	}
	
	return ext
}