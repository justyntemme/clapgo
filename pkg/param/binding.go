package param

import (
	"fmt"
	"math"
)

// ParameterBinding represents a bound parameter with automatic handling
type ParameterBinding struct {
	ID        uint32
	Atomic    *AtomicFloat64
	Format    Format
	Choices   []string       // For choice parameters
	OnChange  func(float64)  // Optional callback
	Min       float64        // Minimum value
	Max       float64        // Maximum value
	Default   float64        // Default value
}

// ParameterBinder manages parameter bindings and automates common operations
type ParameterBinder struct {
	bindings map[uint32]*ParameterBinding
	manager  *Manager
}

// NewParameterBinder creates a new parameter binder
func NewParameterBinder(manager *Manager) *ParameterBinder {
	return &ParameterBinder{
		bindings: make(map[uint32]*ParameterBinding),
		manager:  manager,
	}
}

// bindParameter is the internal method for creating and registering a parameter
func (pb *ParameterBinder) bindParameter(info Info, format Format, choices []string) *AtomicFloat64 {
	// Create atomic storage
	atomic := NewAtomicFloat64(info.DefaultValue)
	
	// Create binding
	binding := &ParameterBinding{
		ID:      info.ID,
		Atomic:  atomic,
		Format:  format,
		Choices: choices,
		Min:     info.MinValue,
		Max:     info.MaxValue,
		Default: info.DefaultValue,
	}
	
	// Store binding
	pb.bindings[info.ID] = binding
	
	// Register with manager
	pb.manager.Register(info)
	
	return atomic
}

// BindPercentage binds a percentage parameter (0-100%)
func (pb *ParameterBinder) BindPercentage(id uint32, name string, defaultValue float64) *AtomicFloat64 {
	info := Percentage(id, name, defaultValue)
	return pb.bindParameter(info, FormatPercentage, nil)
}

// BindChoice binds a choice parameter with string options
func (pb *ParameterBinder) BindChoice(id uint32, name string, choices []string, defaultIndex int) *AtomicFloat64 {
	if defaultIndex >= len(choices) {
		defaultIndex = 0
	}
	info := Choice(id, name, len(choices), defaultIndex)
	return pb.bindParameter(info, FormatDefault, choices)
}

// BindCutoff binds a filter cutoff frequency parameter (20Hz-20kHz)
func (pb *ParameterBinder) BindCutoff(id uint32, name string, defaultValue float64) *AtomicFloat64 {
	info := Cutoff(id, name)
	info.DefaultValue = defaultValue
	return pb.bindParameter(info, FormatHertz, nil)
}

// BindResonance binds a filter resonance parameter (0-1)
func (pb *ParameterBinder) BindResonance(id uint32, name string, defaultValue float64) *AtomicFloat64 {
	info := Resonance(id, name)
	info.DefaultValue = defaultValue
	return pb.bindParameter(info, FormatPercentage, nil)
}

// BindADSR binds an ADSR envelope time parameter
func (pb *ParameterBinder) BindADSR(id uint32, name string, maxSeconds float64, defaultValue float64) *AtomicFloat64 {
	info := ADSR(id, name, maxSeconds)
	info.DefaultValue = defaultValue
	return pb.bindParameter(info, FormatMilliseconds, nil)
}

// BindDB binds a decibel parameter
func (pb *ParameterBinder) BindDB(id uint32, name string, minDB, maxDB, defaultDB float64) *AtomicFloat64 {
	info := Info{
		ID:           id,
		Name:         name,
		Module:       "",
		MinValue:     minDB,
		MaxValue:     maxDB,
		DefaultValue: defaultDB,
		Flags:        FlagAutomatable,
	}
	return pb.bindParameter(info, FormatDecibels, nil)
}

// BindLinear binds a linear parameter with custom range
func (pb *ParameterBinder) BindLinear(id uint32, name string, min, max, defaultValue float64) *AtomicFloat64 {
	info := Info{
		ID:           id,
		Name:         name,
		Module:       "",
		MinValue:     min,
		MaxValue:     max,
		DefaultValue: defaultValue,
		Flags:        FlagAutomatable,
	}
	return pb.bindParameter(info, FormatDefault, nil)
}

// SetCallback sets a callback function for when a parameter changes
func (pb *ParameterBinder) SetCallback(id uint32, callback func(float64)) {
	if binding, ok := pb.bindings[id]; ok {
		binding.OnChange = callback
	}
}

// HandleParamValue automatically handles parameter value changes
func (pb *ParameterBinder) HandleParamValue(paramID uint32, value float64) bool {
	binding, ok := pb.bindings[paramID]
	if !ok {
		return false
	}
	
	// Update atomic value
	if len(binding.Choices) > 0 {
		// For choice parameters, round to nearest integer
		value = math.Round(value)
	}
	
	// Clamp to valid range
	value = ClampValue(value, binding.Min, binding.Max)
	
	// Update atomic storage and manager
	binding.Atomic.UpdateWithManager(value, pb.manager, paramID)
	
	// Call callback if set
	if binding.OnChange != nil {
		binding.OnChange(value)
	}
	
	return true
}

// ValueToText converts a parameter value to text based on its binding
func (pb *ParameterBinder) ValueToText(paramID uint32, value float64) (string, bool) {
	binding, ok := pb.bindings[paramID]
	if !ok {
		return "", false
	}
	
	// Handle choice parameters specially
	if len(binding.Choices) > 0 {
		index := int(math.Round(value))
		if index >= 0 && index < len(binding.Choices) {
			return binding.Choices[index], true
		}
		return "Unknown", true
	}
	
	// Use format type for other parameters
	return FormatValue(value, binding.Format), true
}

// TextToValue converts text to a parameter value based on its binding
func (pb *ParameterBinder) TextToValue(paramID uint32, text string) (float64, error) {
	binding, ok := pb.bindings[paramID]
	if !ok {
		return 0, fmt.Errorf("unknown parameter ID: %d", paramID)
	}
	
	// Handle choice parameters specially
	if len(binding.Choices) > 0 {
		for i, choice := range binding.Choices {
			if choice == text {
				return float64(i), nil
			}
		}
		return 0, fmt.Errorf("invalid choice: %s", text)
	}
	
	// Use parser for other parameters
	parser := NewParser(binding.Format)
	value, err := parser.ParseValue(text)
	if err != nil {
		return 0, err
	}
	
	// Clamp to valid range
	return ClampValue(value, binding.Min, binding.Max), nil
}

// GetBinding returns the parameter binding for a given ID
func (pb *ParameterBinder) GetBinding(paramID uint32) (*ParameterBinding, bool) {
	binding, ok := pb.bindings[paramID]
	return binding, ok
}

// GetAllBindings returns all parameter bindings
func (pb *ParameterBinder) GetAllBindings() map[uint32]*ParameterBinding {
	// Return a copy to prevent external modification
	result := make(map[uint32]*ParameterBinding)
	for k, v := range pb.bindings {
		result[k] = v
	}
	return result
}