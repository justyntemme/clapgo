package api

import (
	"errors"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Common parameter errors
var (
	ErrInvalidParam = errors.New("invalid parameter ID")
)

// ParameterManager provides thread-safe parameter management with validation
// and change notification. It completely abstracts C interop from plugin developers.
type ParameterManager struct {
	mutex      sync.RWMutex
	params     map[uint32]*Parameter
	paramOrder []uint32
	listeners  []ParameterChangeListener
}

// Parameter represents a plugin parameter with thread-safe access
type Parameter struct {
	Info      ParamInfo
	value     int64 // atomic storage for float64 bits
	validator func(float64) error
}

// ParameterChangeListener is called when parameter values change
type ParameterChangeListener func(paramID uint32, oldValue, newValue float64)

// NewParameterManager creates a new thread-safe parameter manager
func NewParameterManager() *ParameterManager {
	return &ParameterManager{
		params:     make(map[uint32]*Parameter),
		paramOrder: make([]uint32, 0),
		listeners:  make([]ParameterChangeListener, 0),
	}
}

// RegisterParameter registers a new parameter with the manager
func (pm *ParameterManager) RegisterParameter(info ParamInfo) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	if _, exists := pm.params[info.ID]; exists {
		return errors.New("parameter ID already exists")
	}
	
	param := &Parameter{
		Info: info,
	}
	
	// Set default value atomically
	atomic.StoreInt64(&param.value, int64(floatToBits(info.DefaultValue)))
	
	// Add validator if parameter has bounds
	if info.Flags&(ParamIsBoundedBelow|ParamIsBoundedAbove) != 0 {
		param.validator = func(value float64) error {
			if info.Flags&ParamIsBoundedBelow != 0 && value < info.MinValue {
				return errors.New("value below minimum")
			}
			if info.Flags&ParamIsBoundedAbove != 0 && value > info.MaxValue {
				return errors.New("value above maximum")
			}
			return nil
		}
	}
	
	pm.params[info.ID] = param
	pm.paramOrder = append(pm.paramOrder, info.ID)
	
	return nil
}

// GetParameterCount returns the number of registered parameters
func (pm *ParameterManager) GetParameterCount() uint32 {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	return uint32(len(pm.params))
}

// GetParameterInfo returns information about a parameter by ID
func (pm *ParameterManager) GetParameterInfo(paramID uint32) (ParamInfo, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	if param, exists := pm.params[paramID]; exists {
		return param.Info, nil
	}
	
	return ParamInfo{}, ErrInvalidParam
}

// GetParameterInfoByIndex returns information about a parameter by index
func (pm *ParameterManager) GetParameterInfoByIndex(index uint32) (ParamInfo, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	if int(index) >= len(pm.paramOrder) {
		return ParamInfo{}, ErrInvalidParam
	}
	
	paramID := pm.paramOrder[index]
	return pm.params[paramID].Info, nil
}

// GetParameterValue returns the current value of a parameter (thread-safe)
func (pm *ParameterManager) GetParameterValue(paramID uint32) float64 {
	pm.mutex.RLock()
	param, exists := pm.params[paramID]
	pm.mutex.RUnlock()
	
	if !exists {
		return 0.0
	}
	
	// Load value atomically
	bits := atomic.LoadInt64(&param.value)
	return floatFromBits(uint64(bits))
}

// SetParameterValue sets the value of a parameter with validation (thread-safe)
func (pm *ParameterManager) SetParameterValue(paramID uint32, value float64) error {
	pm.mutex.RLock()
	param, exists := pm.params[paramID]
	pm.mutex.RUnlock()
	
	if !exists {
		return ErrInvalidParam
	}
	
	// Validate value if validator exists
	if param.validator != nil {
		if err := param.validator(value); err != nil {
			// Auto-clamp if bounds checking fails
			if param.Info.Flags&ParamIsBoundedBelow != 0 && value < param.Info.MinValue {
				value = param.Info.MinValue
			}
			if param.Info.Flags&ParamIsBoundedAbove != 0 && value > param.Info.MaxValue {
				value = param.Info.MaxValue
			}
		}
	}
	
	// Get old value for change notification
	oldBits := atomic.LoadInt64(&param.value)
	oldValue := floatFromBits(uint64(oldBits))
	
	// Store new value atomically
	atomic.StoreInt64(&param.value, int64(floatToBits(value)))
	
	// Notify listeners if value changed
	if oldValue != value {
		pm.notifyListeners(paramID, oldValue, value)
	}
	
	return nil
}

// AddChangeListener adds a parameter change listener
func (pm *ParameterManager) AddChangeListener(listener ParameterChangeListener) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.listeners = append(pm.listeners, listener)
}

// notifyListeners notifies all registered listeners of parameter changes
func (pm *ParameterManager) notifyListeners(paramID uint32, oldValue, newValue float64) {
	pm.mutex.RLock()
	listeners := make([]ParameterChangeListener, len(pm.listeners))
	copy(listeners, pm.listeners)
	pm.mutex.RUnlock()
	
	// Call listeners without holding the lock
	for _, listener := range listeners {
		listener(paramID, oldValue, newValue)
	}
}

// ProcessParameterEvents processes parameter events from the event system
func (pm *ParameterManager) ProcessParameterEvents(events []Event) {
	for _, event := range events {
		if event.Type == EventTypeParamValue {
			if paramEvent, ok := event.Data.(ParamEvent); ok {
				pm.SetParameterValue(paramEvent.ParamID, paramEvent.Value)
			}
		}
	}
}

// GetAllParameterValues returns a map of all current parameter values
func (pm *ParameterManager) GetAllParameterValues() map[uint32]float64 {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	values := make(map[uint32]float64)
	for paramID, param := range pm.params {
		bits := atomic.LoadInt64(&param.value)
		values[paramID] = floatFromBits(uint64(bits))
	}
	
	return values
}

// SetAllParameterValues sets multiple parameter values efficiently
func (pm *ParameterManager) SetAllParameterValues(values map[uint32]float64) {
	for paramID, value := range values {
		pm.SetParameterValue(paramID, value)
	}
}

// ResetToDefaults resets all parameters to their default values
func (pm *ParameterManager) ResetToDefaults() {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	for _, param := range pm.params {
		atomic.StoreInt64(&param.value, int64(floatToBits(param.Info.DefaultValue)))
	}
}

// Utility functions for atomic float64 operations

func floatToBits(f float64) uint64 {
	return *(*uint64)(unsafe.Pointer(&f))
}

func floatFromBits(b uint64) float64 {
	return *(*float64)(unsafe.Pointer(&b))
}

// PluginParameterWrapper wraps a plugin to provide parameter management
type PluginParameterWrapper struct {
	plugin Plugin
	params *ParameterManager
}

// NewPluginWithParameters creates a plugin wrapper with parameter management
func NewPluginWithParameters(plugin Plugin) *PluginParameterWrapper {
	return &PluginParameterWrapper{
		plugin: plugin,
		params: NewParameterManager(),
	}
}

// GetParameterManager returns the parameter manager
func (w *PluginParameterWrapper) GetParameterManager() *ParameterManager {
	return w.params
}

// Plugin interface implementation (delegates to wrapped plugin)
func (w *PluginParameterWrapper) Init() bool {
	return w.plugin.Init()
}

func (w *PluginParameterWrapper) Destroy() {
	w.plugin.Destroy()
}

func (w *PluginParameterWrapper) Activate(sampleRate float64, minFrames, maxFrames uint32) bool {
	return w.plugin.Activate(sampleRate, minFrames, maxFrames)
}

func (w *PluginParameterWrapper) Deactivate() {
	w.plugin.Deactivate()
}

func (w *PluginParameterWrapper) StartProcessing() bool {
	return w.plugin.StartProcessing()
}

func (w *PluginParameterWrapper) StopProcessing() {
	w.plugin.StopProcessing()
}

func (w *PluginParameterWrapper) Reset() {
	w.plugin.Reset()
	w.params.ResetToDefaults()
}

func (w *PluginParameterWrapper) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events EventHandler) int {
	// Process parameter events first
	if events != nil {
		eventCount := events.GetInputEventCount()
		paramEvents := make([]Event, 0, eventCount)
		
		for i := uint32(0); i < eventCount; i++ {
			event := events.GetInputEvent(i)
			if event != nil && event.Type == EventTypeParamValue {
				paramEvents = append(paramEvents, *event)
			}
		}
		
		if len(paramEvents) > 0 {
			w.params.ProcessParameterEvents(paramEvents)
		}
	}
	
	return w.plugin.Process(steadyTime, framesCount, audioIn, audioOut, events)
}

func (w *PluginParameterWrapper) GetExtension(id string) unsafe.Pointer {
	return w.plugin.GetExtension(id)
}

func (w *PluginParameterWrapper) OnMainThread() {
	w.plugin.OnMainThread()
}

func (w *PluginParameterWrapper) GetPluginID() string {
	return w.plugin.GetPluginID()
}

func (w *PluginParameterWrapper) GetPluginInfo() PluginInfo {
	return w.plugin.GetPluginInfo()
}

// Helper functions for common parameter types

// CreateFloatParameter creates a float parameter with bounds
func CreateFloatParameter(id uint32, name string, minVal, maxVal, defaultVal float64) ParamInfo {
	return ParamInfo{
		ID:           id,
		Name:         name,
		MinValue:     minVal,
		MaxValue:     maxVal,
		DefaultValue: defaultVal,
		Flags:        ParamIsAutomatable | ParamIsBoundedBelow | ParamIsBoundedAbove,
	}
}

// CreateBoolParameter creates a boolean parameter (0.0 or 1.0)
func CreateBoolParameter(id uint32, name string, defaultVal bool) ParamInfo {
	var defaultFloat float64
	if defaultVal {
		defaultFloat = 1.0
	}
	
	return ParamInfo{
		ID:           id,
		Name:         name,
		MinValue:     0.0,
		MaxValue:     1.0,
		DefaultValue: defaultFloat,
		Flags:        ParamIsAutomatable | ParamIsBoundedBelow | ParamIsBoundedAbove | ParamIsSteppable,
	}
}

// CreateIntParameter creates an integer parameter
func CreateIntParameter(id uint32, name string, minVal, maxVal, defaultVal int) ParamInfo {
	return ParamInfo{
		ID:           id,
		Name:         name,
		MinValue:     float64(minVal),
		MaxValue:     float64(maxVal),
		DefaultValue: float64(defaultVal),
		Flags:        ParamIsAutomatable | ParamIsBoundedBelow | ParamIsBoundedAbove | ParamIsSteppable,
	}
}