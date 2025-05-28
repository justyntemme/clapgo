package audio

import (
	"errors"
	"fmt"
)

// PortBuilder provides a fluent interface for creating audio port info
type PortBuilder struct {
	info PortInfo
	err  error
}

// NewPortBuilder creates a new audio port builder
func NewPortBuilder(id uint32, name string) *PortBuilder {
	return &PortBuilder{
		info: PortInfo{
			ID:           id,
			Name:         name,
			ChannelCount: 2, // Default to stereo
			Flags:        0,
			PortType:     PortTypeStereo,
			InPlacePair:  InvalidID,
		},
	}
}

// Channels sets the number of channels
func (b *PortBuilder) Channels(count uint32) *PortBuilder {
	if b.err != nil {
		return b
	}
	if count == 0 {
		b.err = errors.New("channel count must be greater than 0")
		return b
	}
	
	b.info.ChannelCount = count
	
	// Auto-set port type based on channel count
	switch count {
	case 1:
		b.info.PortType = PortTypeMono
	case 2:
		b.info.PortType = PortTypeStereo
	default:
		b.info.PortType = fmt.Sprintf("%dch", count)
	}
	
	return b
}

// Mono sets the port to mono (1 channel)
func (b *PortBuilder) Mono() *PortBuilder {
	return b.Channels(1)
}

// Stereo sets the port to stereo (2 channels)
func (b *PortBuilder) Stereo() *PortBuilder {
	return b.Channels(2)
}

// Surround sets the port to surround sound with specified channels
func (b *PortBuilder) Surround(channels uint32) *PortBuilder {
	if b.err != nil {
		return b
	}
	
	b.Channels(channels)
	
	// Set surround-specific port types
	switch channels {
	case 4:
		b.info.PortType = "quad"
	case 6:
		b.info.PortType = "5.1"
	case 8:
		b.info.PortType = "7.1"
	default:
		b.info.PortType = fmt.Sprintf("surround_%dch", channels)
	}
	
	return b
}

// PortType sets a custom port type
func (b *PortBuilder) PortType(portType string) *PortBuilder {
	if b.err != nil {
		return b
	}
	b.info.PortType = portType
	return b
}

// Flags sets the port flags
func (b *PortBuilder) Flags(flags uint32) *PortBuilder {
	if b.err != nil {
		return b
	}
	b.info.Flags = flags
	return b
}

// AddFlags adds additional flags
func (b *PortBuilder) AddFlags(flags uint32) *PortBuilder {
	if b.err != nil {
		return b
	}
	b.info.Flags |= flags
	return b
}

// Main marks this as the main port
func (b *PortBuilder) Main() *PortBuilder {
	return b.AddFlags(PortFlagIsMain)
}

// Supports64Bit marks that this port supports 64-bit processing
func (b *PortBuilder) Supports64Bit() *PortBuilder {
	return b.AddFlags(PortFlagSupports64Bit)
}

// Prefers64Bit marks that this port prefers 64-bit processing
func (b *PortBuilder) Prefers64Bit() *PortBuilder {
	return b.AddFlags(PortFlagSupports64Bit | PortFlagPrefers64Bit)
}

// RequiresConstantChannelCount marks that this port requires constant channel count
func (b *PortBuilder) RequiresConstantChannelCount() *PortBuilder {
	return b.AddFlags(PortFlagRequiresCC)
}

// InPlacePair sets the in-place pair port ID
func (b *PortBuilder) InPlacePair(pairID uint32) *PortBuilder {
	if b.err != nil {
		return b
	}
	b.info.InPlacePair = pairID
	return b
}

// Build creates the port info, returning an error if validation fails
func (b *PortBuilder) Build() (PortInfo, error) {
	if b.err != nil {
		return PortInfo{}, b.err
	}
	
	// Final validation
	if b.info.Name == "" {
		return PortInfo{}, errors.New("port name is required")
	}
	
	if b.info.ChannelCount == 0 {
		return PortInfo{}, errors.New("channel count must be greater than 0")
	}
	
	return b.info, nil
}

// MustBuild creates the port info, panicking on error
func (b *PortBuilder) MustBuild() PortInfo {
	info, err := b.Build()
	if err != nil {
		panic(err)
	}
	return info
}

// Common preset configurations

// MainStereoInput creates a main stereo input port
func MainStereoInput(id uint32, name string) *PortBuilder {
	return NewPortBuilder(id, name).
		Stereo().
		Main()
}

// MainStereoOutput creates a main stereo output port
func MainStereoOutput(id uint32, name string) *PortBuilder {
	return NewPortBuilder(id, name).
		Stereo().
		Main()
}

// SurroundInput creates a surround input port
func SurroundInput(id uint32, name string, channels uint32) *PortBuilder {
	return NewPortBuilder(id, name).
		Surround(channels).
		Main()
}

// SurroundOutput creates a surround output port
func SurroundOutput(id uint32, name string, channels uint32) *PortBuilder {
	return NewPortBuilder(id, name).
		Surround(channels).
		Main()
}

// SidechainInput creates a sidechain input port
func SidechainInput(id uint32, name string) *PortBuilder {
	return NewPortBuilder(id, name).
		Stereo() // Sidechain is typically stereo
}

// Example usage:
// inputPort := MainStereoInput(0, "Main Input").
//     Supports64Bit().
//     Build()
//
// outputPort := MainStereoOutput(1, "Main Output").
//     InPlacePair(0).
//     Prefers64Bit().
//     Build()
//
// surroundPort := SurroundInput(2, "Surround Input").
//     Surround(8). // 7.1 surround
//     Build()