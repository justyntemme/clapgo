package param

// Common parameter factory functions for frequently used parameter types

// Volume creates a volume/gain parameter with dB display
func Volume(id uint32, name string) Info {
	return Info{
		ID:           id,
		Name:         name,
		Module:       "",
		MinValue:     0.0,    // -∞ dB
		MaxValue:     2.0,    // ~6 dB
		DefaultValue: 1.0,    // 0 dB
		Flags:        FlagAutomatable | FlagModulatable | FlagBoundedBelow | FlagBoundedAbove,
	}
}

// Pan creates a pan parameter (-1 to 1)
func Pan(id uint32, name string) Info {
	return Info{
		ID:           id,
		Name:         name,
		Module:       "",
		MinValue:     -1.0,
		MaxValue:     1.0,
		DefaultValue: 0.0,
		Flags:        FlagAutomatable | FlagModulatable | FlagBoundedBelow | FlagBoundedAbove,
	}
}

// Frequency creates a frequency parameter
func Frequency(id uint32, name string, minHz, maxHz, defaultHz float64) Info {
	return Info{
		ID:           id,
		Name:         name,
		Module:       "",
		MinValue:     minHz,
		MaxValue:     maxHz,
		DefaultValue: defaultHz,
		Flags:        FlagAutomatable | FlagModulatable | FlagBoundedBelow | FlagBoundedAbove,
	}
}

// Cutoff creates a filter cutoff parameter
func Cutoff(id uint32, name string) Info {
	return Frequency(id, name, 20.0, 20000.0, 1000.0)
}

// Resonance creates a filter resonance parameter
func Resonance(id uint32, name string) Info {
	return Info{
		ID:           id,
		Name:         name,
		Module:       "",
		MinValue:     0.0,
		MaxValue:     1.0,
		DefaultValue: 0.5,
		Flags:        FlagAutomatable | FlagModulatable | FlagBoundedBelow | FlagBoundedAbove,
	}
}

// ADSR creates an envelope time parameter
func ADSR(id uint32, name string, maxSeconds float64) Info {
	return Info{
		ID:           id,
		Name:         name,
		Module:       "",
		MinValue:     0.0,
		MaxValue:     maxSeconds,
		DefaultValue: 0.1,
		Flags:        FlagAutomatable | FlagModulatable | FlagBoundedBelow | FlagBoundedAbove,
	}
}

// Switch creates a boolean switch parameter
func Switch(id uint32, name string, defaultOn bool) Info {
	defaultVal := 0.0
	if defaultOn {
		defaultVal = 1.0
	}
	return Info{
		ID:           id,
		Name:         name,
		Module:       "",
		MinValue:     0.0,
		MaxValue:     1.0,
		DefaultValue: defaultVal,
		Flags:        FlagAutomatable | FlagStepped | FlagBoundedBelow | FlagBoundedAbove,
	}
}

// Choice creates a stepped parameter for selecting between options
func Choice(id uint32, name string, numOptions int, defaultOption int) Info {
	return Info{
		ID:           id,
		Name:         name,
		Module:       "",
		MinValue:     0.0,
		MaxValue:     float64(numOptions - 1),
		DefaultValue: float64(defaultOption),
		Flags:        FlagAutomatable | FlagStepped | FlagBoundedBelow | FlagBoundedAbove,
	}
}

// Percentage creates a 0-100% parameter
func Percentage(id uint32, name string, defaultPercent float64) Info {
	return Info{
		ID:           id,
		Name:         name,
		Module:       "",
		MinValue:     0.0,
		MaxValue:     1.0,
		DefaultValue: defaultPercent / 100.0,
		Flags:        FlagAutomatable | FlagModulatable | FlagBoundedBelow | FlagBoundedAbove,
	}
}

// Bypass creates a bypass parameter
func Bypass(id uint32) Info {
	return Info{
		ID:           id,
		Name:         "Bypass",
		Module:       "",
		MinValue:     0.0,
		MaxValue:     1.0,
		DefaultValue: 0.0,
		Flags:        FlagBypass | FlagStepped | FlagBoundedBelow | FlagBoundedAbove,
	}
}