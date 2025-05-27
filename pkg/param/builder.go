package param

import (
	"errors"
)

// Builder provides a fluent interface for creating parameters
type Builder struct {
	info   Info
	err    error
}

// NewBuilder creates a new parameter builder
func NewBuilder(id uint32, name string) *Builder {
	return &Builder{
		info: Info{
			ID:           id,
			Name:         name,
			Module:       "",
			MinValue:     0.0,
			MaxValue:     1.0,
			DefaultValue: 0.5,
			Flags:        FlagAutomatable,
		},
	}
}

// Module sets the parameter module/group
func (b *Builder) Module(module string) *Builder {
	if b.err != nil {
		return b
	}
	b.info.Module = module
	return b
}

// Range sets the min, max, and default values
func (b *Builder) Range(min, max, defaultValue float64) *Builder {
	if b.err != nil {
		return b
	}
	if min >= max {
		b.err = errors.New("min value must be less than max value")
		return b
	}
	if defaultValue < min || defaultValue > max {
		b.err = errors.New("default value must be within min/max range")
		return b
	}
	b.info.MinValue = min
	b.info.MaxValue = max
	b.info.DefaultValue = defaultValue
	return b
}

// Min sets the minimum value
func (b *Builder) Min(min float64) *Builder {
	if b.err != nil {
		return b
	}
	if min >= b.info.MaxValue {
		b.err = errors.New("min value must be less than max value")
		return b
	}
	b.info.MinValue = min
	if b.info.DefaultValue < min {
		b.info.DefaultValue = min
	}
	return b
}

// Max sets the maximum value
func (b *Builder) Max(max float64) *Builder {
	if b.err != nil {
		return b
	}
	if max <= b.info.MinValue {
		b.err = errors.New("max value must be greater than min value")
		return b
	}
	b.info.MaxValue = max
	if b.info.DefaultValue > max {
		b.info.DefaultValue = max
	}
	return b
}

// Default sets the default value
func (b *Builder) Default(defaultValue float64) *Builder {
	if b.err != nil {
		return b
	}
	if defaultValue < b.info.MinValue || defaultValue > b.info.MaxValue {
		b.err = errors.New("default value must be within min/max range")
		return b
	}
	b.info.DefaultValue = defaultValue
	return b
}

// Flags sets the parameter flags
func (b *Builder) Flags(flags uint32) *Builder {
	if b.err != nil {
		return b
	}
	b.info.Flags = flags
	return b
}

// AddFlags adds additional flags
func (b *Builder) AddFlags(flags uint32) *Builder {
	if b.err != nil {
		return b
	}
	b.info.Flags |= flags
	return b
}

// Automatable makes the parameter automatable
func (b *Builder) Automatable() *Builder {
	return b.AddFlags(FlagAutomatable)
}

// Modulatable makes the parameter modulatable
func (b *Builder) Modulatable() *Builder {
	return b.AddFlags(FlagModulatable)
}

// Stepped makes the parameter stepped
func (b *Builder) Stepped() *Builder {
	return b.AddFlags(FlagStepped)
}

// Hidden makes the parameter hidden
func (b *Builder) Hidden() *Builder {
	return b.AddFlags(FlagHidden)
}

// ReadOnly makes the parameter read-only
func (b *Builder) ReadOnly() *Builder {
	return b.AddFlags(FlagReadonly)
}

// Bypass marks this as a bypass parameter
func (b *Builder) Bypass() *Builder {
	return b.AddFlags(FlagBypass)
}

// Bounded adds bounded flags
func (b *Builder) Bounded() *Builder {
	return b.AddFlags(FlagBoundedBelow | FlagBoundedAbove)
}

// Format sets a predefined format for the parameter
func (b *Builder) Format(format Format) *Builder {
	if b.err != nil {
		return b
	}
	
	// Apply format-specific defaults
	switch format {
	case FormatDecibel:
		// Decibel parameters typically range from -inf to some positive value
		// We'll use 0-2 for linear gain (which maps to -inf to +6dB)
		if b.info.MinValue == 0 && b.info.MaxValue == 1 {
			b.info.MinValue = 0.0
			b.info.MaxValue = 2.0
			b.info.DefaultValue = 1.0
		}
		
	case FormatPercentage:
		// Percentage is 0-1
		if b.info.MinValue == 0 && b.info.MaxValue == 1 {
			// Already correct
		}
		
	case FormatMilliseconds:
		// Time parameters often need larger ranges
		if b.info.MinValue == 0 && b.info.MaxValue == 1 {
			b.info.MaxValue = 1000.0 // 1 second default max
			b.info.DefaultValue = 100.0
		}
		
	case FormatHertz, FormatKilohertz:
		// Frequency parameters typically range from 20Hz to 20kHz
		if b.info.MinValue == 0 && b.info.MaxValue == 1 {
			b.info.MinValue = 20.0
			b.info.MaxValue = 20000.0
			b.info.DefaultValue = 1000.0
		}
	}
	
	return b
}

// Build creates the parameter info, returning an error if validation fails
func (b *Builder) Build() (Info, error) {
	if b.err != nil {
		return Info{}, b.err
	}
	
	// Final validation
	if b.info.Name == "" {
		return Info{}, errors.New("parameter name is required")
	}
	
	if b.info.MinValue >= b.info.MaxValue {
		return Info{}, errors.New("min value must be less than max value")
	}
	
	if b.info.DefaultValue < b.info.MinValue || b.info.DefaultValue > b.info.MaxValue {
		return Info{}, errors.New("default value must be within min/max range")
	}
	
	return b.info, nil
}

// MustBuild creates the parameter info, panicking on error
func (b *Builder) MustBuild() Info {
	info, err := b.Build()
	if err != nil {
		panic(err)
	}
	return info
}

// Example usage:
// param := NewBuilder(0, "Cutoff").
//     Module("Filter").
//     Range(20, 20000, 1000).
//     Format(FormatFrequency).
//     Automatable().
//     Modulatable().
//     Bounded().
//     Build()