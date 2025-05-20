package clap

import (
	"sync"

	"github.com/justyntemme/clapgo/pkg/api"
)

// ParamManager provides parameter management functionality for plugins.
// It implements the api.ParamsProvider interface and can be used
// to add parameter support to plugins.
type ParamManager struct {
	mutex      sync.RWMutex
	params     map[uint32]api.ParamInfo
	values     map[uint32]float64
	paramOrder []uint32
}

// NewParamManager creates a new parameter manager.
func NewParamManager() *ParamManager {
	return &ParamManager{
		params:     make(map[uint32]api.ParamInfo),
		values:     make(map[uint32]float64),
		paramOrder: make([]uint32, 0),
	}
}

// RegisterParam registers a parameter with the manager.
func (m *ParamManager) RegisterParam(info api.ParamInfo) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Store the parameter info
	m.params[info.ID] = info

	// Initialize the parameter value
	m.values[info.ID] = info.DefaultValue

	// Store the parameter order
	m.paramOrder = append(m.paramOrder, info.ID)
}

// GetParamCount returns the number of parameters.
func (m *ParamManager) GetParamCount() uint32 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return uint32(len(m.params))
}

// GetParamInfo returns information about a parameter.
func (m *ParamManager) GetParamInfo(paramID uint32) api.ParamInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if info, exists := m.params[paramID]; exists {
		return info
	}

	// Return empty info if parameter doesn't exist
	return api.ParamInfo{}
}

// GetParamInfoByIndex returns information about a parameter by index.
func (m *ParamManager) GetParamInfoByIndex(index uint32) api.ParamInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if int(index) >= len(m.paramOrder) {
		return api.ParamInfo{}
	}

	paramID := m.paramOrder[index]
	return m.params[paramID]
}

// GetParamValue returns the current value of a parameter.
func (m *ParamManager) GetParamValue(paramID uint32) float64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if value, exists := m.values[paramID]; exists {
		return value
	}

	return 0.0
}

// SetParamValue sets the value of a parameter.
func (m *ParamManager) SetParamValue(paramID uint32, value float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.params[paramID]; exists {
		// Clamp value to parameter range if bounded
		info := m.params[paramID]
		if info.Flags&api.ParamIsBoundedBelow != 0 && value < info.MinValue {
			value = info.MinValue
		}
		if info.Flags&api.ParamIsBoundedAbove != 0 && value > info.MaxValue {
			value = info.MaxValue
		}

		m.values[paramID] = value
	}
}

// FlushParams writes all parameter changes to the DSP.
// This implementation does nothing as parameter changes are
// applied immediately, but can be overridden if needed.
func (m *ParamManager) FlushParams() {
	// Nothing to do in the base implementation
}