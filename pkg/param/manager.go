package param

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

// ChangeListener is called when parameter values change
type ChangeListener func(paramID uint32, oldValue, newValue float64)

// Manager provides thread-safe parameter management with validation and change notification
type Manager struct {
	mutex         sync.RWMutex
	params        map[uint32]*Parameter
	paramOrder    []uint32
	listeners     [MaxListeners]ChangeListener
	listenerCount int32 // atomic
}

// NewManager creates a new thread-safe parameter manager
func NewManager() *Manager {
	return &Manager{
		params:        make(map[uint32]*Parameter),
		paramOrder:    make([]uint32, 0),
		listenerCount: 0,
	}
}

// Register registers a new parameter with the manager
func (m *Manager) Register(info Info) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if _, exists := m.params[info.ID]; exists {
		return ErrParamExists
	}
	
	param := &Parameter{
		Info: info,
	}
	
	// Set default value atomically
	atomic.StoreInt64(&param.value, int64(floatToBits(info.DefaultValue)))
	
	// Add validator if parameter has bounds
	if info.Flags&(IsBoundedBelow|IsBoundedAbove) != 0 {
		param.validator = func(value float64) error {
			if info.Flags&IsBoundedBelow != 0 && value < info.MinValue {
				return ErrValueBelowMinimum
			}
			if info.Flags&IsBoundedAbove != 0 && value > info.MaxValue {
				return ErrValueAboveMaximum
			}
			return nil
		}
	}
	
	m.params[info.ID] = param
	m.paramOrder = append(m.paramOrder, info.ID)
	
	return nil
}

// RegisterAll registers multiple parameters at once
func (m *Manager) RegisterAll(infos ...Info) error {
	for _, info := range infos {
		if err := m.Register(info); err != nil {
			return err
		}
	}
	return nil
}

// Count returns the number of registered parameters
func (m *Manager) Count() uint32 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return uint32(len(m.params))
}

// GetInfo returns information about a parameter by ID
func (m *Manager) GetInfo(paramID uint32) (Info, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	if param, exists := m.params[paramID]; exists {
		return param.Info, nil
	}
	
	return Info{}, ErrInvalidParam
}

// GetInfoByIndex returns information about a parameter by index
func (m *Manager) GetInfoByIndex(index uint32) (Info, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	if int(index) >= len(m.paramOrder) {
		return Info{}, ErrInvalidParam
	}
	
	paramID := m.paramOrder[index]
	return m.params[paramID].Info, nil
}

// Get returns the current value of a parameter (thread-safe)
func (m *Manager) Get(paramID uint32) float64 {
	m.mutex.RLock()
	param, exists := m.params[paramID]
	m.mutex.RUnlock()
	
	if !exists {
		return 0.0
	}
	
	return param.Value()
}

// Set sets the value of a parameter with validation (thread-safe)
func (m *Manager) Set(paramID uint32, value float64) error {
	m.mutex.RLock()
	param, exists := m.params[paramID]
	m.mutex.RUnlock()
	
	if !exists {
		return ErrInvalidParam
	}
	
	// Get old value for change notification
	oldValue := param.Value()
	
	// Set new value with validation
	if err := param.SetValue(value); err != nil {
		return err
	}
	
	// Notify listeners if value changed
	newValue := param.Value()
	if oldValue != newValue {
		m.notifyListeners(paramID, oldValue, newValue)
	}
	
	return nil
}

// GetParameter returns the parameter object for direct access
func (m *Manager) GetParameter(paramID uint32) (*Parameter, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	if param, exists := m.params[paramID]; exists {
		return param, nil
	}
	
	return nil, ErrInvalidParam
}

// AddListener adds a parameter change listener
func (m *Manager) AddListener(listener ChangeListener) error {
	if listener == nil {
		return ErrInvalidParam
	}
	
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	count := atomic.LoadInt32(&m.listenerCount)
	if count >= MaxListeners {
		return ErrListenerLimitReached
	}
	
	// Add listener and increment count
	m.listeners[count] = listener
	atomic.AddInt32(&m.listenerCount, 1)
	
	return nil
}

// RemoveListener removes a parameter change listener
func (m *Manager) RemoveListener(listener ChangeListener) bool {
	if listener == nil {
		return false
	}
	
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	count := atomic.LoadInt32(&m.listenerCount)
	
	// Find and remove the listener
	for i := int32(0); i < count; i++ {
		// Compare function pointers
		if m.listeners[i] == nil {
			continue
		}
		
		// Note: This comparison works for function values in Go
		listenerPtr1 := *(*uintptr)(unsafe.Pointer(&m.listeners[i]))
		listenerPtr2 := *(*uintptr)(unsafe.Pointer(&listener))
		
		if listenerPtr1 == listenerPtr2 {
			// Shift remaining listeners down
			for j := i; j < count-1; j++ {
				m.listeners[j] = m.listeners[j+1]
			}
			// Clear the last slot
			m.listeners[count-1] = nil
			// Decrement count
			atomic.AddInt32(&m.listenerCount, -1)
			return true
		}
	}
	
	return false
}

// ListenerCount returns the current number of registered listeners
func (m *Manager) ListenerCount() int32 {
	return atomic.LoadInt32(&m.listenerCount)
}

// notifyListeners notifies all registered listeners of parameter changes
func (m *Manager) notifyListeners(paramID uint32, oldValue, newValue float64) {
	// Take a snapshot of listeners under lock to avoid race conditions
	m.mutex.RLock()
	count := atomic.LoadInt32(&m.listenerCount)
	var listeners [MaxListeners]ChangeListener
	copy(listeners[:count], m.listeners[:count])
	m.mutex.RUnlock()
	
	// Call listeners without holding the lock
	for i := int32(0); i < count; i++ {
		listener := listeners[i]
		
		if listener != nil {
			listener(paramID, oldValue, newValue)
		}
	}
}

// GetAll returns a map of all current parameter values
func (m *Manager) GetAll() map[uint32]float64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	values := make(map[uint32]float64)
	for paramID, param := range m.params {
		values[paramID] = param.Value()
	}
	
	return values
}

// SetAll sets multiple parameter values efficiently
func (m *Manager) SetAll(values map[uint32]float64) {
	for paramID, value := range values {
		m.Set(paramID, value)
	}
}

// GetValue is an alias for Get (for API compatibility)
func (m *Manager) GetValue(paramID uint32) (float64, error) {
	value := m.Get(paramID)
	if value == 0.0 {
		// Check if parameter exists
		m.mutex.RLock()
		_, exists := m.params[paramID]
		m.mutex.RUnlock()
		if !exists {
			return 0, ErrInvalidParam
		}
	}
	return value, nil
}

// SetValue is an alias for Set (for API compatibility)
func (m *Manager) SetValue(paramID uint32, value float64) error {
	return m.Set(paramID, value)
}

// ResetToDefaults resets all parameters to their default values
func (m *Manager) ResetToDefaults() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	for _, param := range m.params {
		atomic.StoreInt64(&param.value, int64(floatToBits(param.Info.DefaultValue)))
	}
}

// ForEach calls the provided function for each parameter
func (m *Manager) ForEach(fn func(Info, float64)) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	for _, id := range m.paramOrder {
		param := m.params[id]
		fn(param.Info, param.Value())
	}
}