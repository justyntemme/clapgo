package audio

import (
	"fmt"
	"math"
)

// FilterType defines the type of filter to use
type FilterType int

const (
	FilterLowpass FilterType = iota
	FilterHighpass
	FilterBandpass
	FilterNotch
	FilterBypass
)

// String returns the string representation of the filter type
func (ft FilterType) String() string {
	switch ft {
	case FilterLowpass:
		return "Lowpass"
	case FilterHighpass:
		return "Highpass"
	case FilterBandpass:
		return "Bandpass"
	case FilterNotch:
		return "Notch"
	case FilterBypass:
		return "Bypass"
	default:
		return "Unknown"
	}
}

// SelectableFilter wraps StateVariableFilter with type selection and safety features
type SelectableFilter struct {
	filter     *StateVariableFilter
	filterType FilterType
	safeMode   bool
	sampleRate float64
	
	// Statistics for debugging
	nanCount   uint64
	infCount   uint64
	resetCount uint64
}

// NewSelectableFilter creates a new selectable filter
func NewSelectableFilter(sampleRate float64, safeMode bool) *SelectableFilter {
	return &SelectableFilter{
		filter:     NewStateVariableFilter(sampleRate),
		filterType: FilterLowpass,
		safeMode:   safeMode,
		sampleRate: sampleRate,
	}
}

// SetType sets the filter type
func (f *SelectableFilter) SetType(filterType FilterType) {
	f.filterType = filterType
}

// GetType returns the current filter type
func (f *SelectableFilter) GetType() FilterType {
	return f.filterType
}

// SetFrequency sets the filter cutoff frequency
func (f *SelectableFilter) SetFrequency(freq float64) {
	// Clamp frequency to safe range automatically
	freq = Clamp(freq, 20.0, f.sampleRate*0.45)
	f.filter.SetFrequency(freq)
}

// SetResonance sets the filter resonance (Q factor)
func (f *SelectableFilter) SetResonance(q float64) {
	// Clamp resonance to safe range automatically
	q = Clamp(q, 0.5, 20.0)
	f.filter.SetResonance(q)
}

// Process processes a single sample through the selected filter type
func (f *SelectableFilter) Process(input float64) float64 {
	var output float64
	
	switch f.filterType {
	case FilterLowpass:
		output = f.filter.ProcessLowpass(input)
	case FilterHighpass:
		output = f.filter.ProcessHighpass(input)
	case FilterBandpass:
		output = f.filter.ProcessBandpass(input)
	case FilterNotch:
		_, _, _, output = f.filter.Process(input)
	case FilterBypass:
		output = input
	default:
		output = input // Default to bypass for unknown types
	}
	
	// Built-in safety when enabled
	if f.safeMode {
		if math.IsNaN(output) {
			f.nanCount++
			f.filter.Reset()
			f.resetCount++
			return 0
		}
		if math.IsInf(output, 0) {
			f.infCount++
			f.filter.Reset()
			f.resetCount++
			return 0
		}
	}
	
	return output
}

// ProcessBuffer processes a buffer of samples in place
func (f *SelectableFilter) ProcessBuffer(buffer []float32) {
	for i := range buffer {
		buffer[i] = float32(f.Process(float64(buffer[i])))
	}
}

// ProcessBufferSeparate processes input buffer to output buffer
func (f *SelectableFilter) ProcessBufferSeparate(input, output []float32) {
	minLen := len(input)
	if len(output) < minLen {
		minLen = len(output)
	}
	
	for i := 0; i < minLen; i++ {
		output[i] = float32(f.Process(float64(input[i])))
	}
}

// Reset resets the filter state
func (f *SelectableFilter) Reset() {
	f.filter.Reset()
}

// SetSampleRate updates the sample rate and recreates the internal filter
func (f *SelectableFilter) SetSampleRate(sampleRate float64) {
	oldFreq := f.filter.frequency
	oldResonance := f.filter.resonance
	
	f.sampleRate = sampleRate
	f.filter = NewStateVariableFilter(sampleRate)
	f.filter.SetFrequency(oldFreq)
	f.filter.SetResonance(oldResonance)
}

// GetStatistics returns debugging statistics about the filter
func (f *SelectableFilter) GetStatistics() FilterStatistics {
	return FilterStatistics{
		NaNCount:   f.nanCount,
		InfCount:   f.infCount,
		ResetCount: f.resetCount,
		SafeMode:   f.safeMode,
		FilterType: f.filterType,
	}
}

// ResetStatistics clears the debugging statistics
func (f *SelectableFilter) ResetStatistics() {
	f.nanCount = 0
	f.infCount = 0
	f.resetCount = 0
}

// SetSafeMode enables or disables automatic NaN/Inf protection
func (f *SelectableFilter) SetSafeMode(enabled bool) {
	f.safeMode = enabled
}

// IsSafeModeEnabled returns true if safe mode is enabled
func (f *SelectableFilter) IsSafeModeEnabled() bool {
	return f.safeMode
}

// FilterStatistics holds debugging information about filter performance
type FilterStatistics struct {
	NaNCount   uint64     // Number of NaN values caught
	InfCount   uint64     // Number of Inf values caught
	ResetCount uint64     // Number of times filter was reset
	SafeMode   bool       // Whether safe mode is enabled
	FilterType FilterType // Current filter type
}

// HasErrors returns true if any NaN or Inf values were detected
func (fs FilterStatistics) HasErrors() bool {
	return fs.NaNCount > 0 || fs.InfCount > 0
}

// String returns a human-readable description of the statistics
func (fs FilterStatistics) String() string {
	if !fs.HasErrors() {
		return "Filter: No errors detected"
	}
	
	return fmt.Sprintf("Filter (%s): %d NaN, %d Inf, %d resets (SafeMode: %v)",
		fs.FilterType.String(), fs.NaNCount, fs.InfCount, fs.ResetCount, fs.SafeMode)
}

// MapFilterTypeFromInt converts an integer to a FilterType
func MapFilterTypeFromInt(value int) FilterType {
	switch value {
	case 0:
		return FilterLowpass
	case 1:
		return FilterHighpass
	case 2:
		return FilterBandpass
	case 3:
		return FilterNotch
	default:
		return FilterBypass
	}
}

// MapFilterTypeToInt converts a FilterType to an integer
func MapFilterTypeToInt(filterType FilterType) int {
	return int(filterType)
}