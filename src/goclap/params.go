package goclap

// #include <stdlib.h>
// #include <string.h>
// #include "../../include/clap/include/clap/clap.h"
import "C"
import (
	"unsafe"
)

// Parameter flags
const (
	ParamIsSteppable          = 1 << 0
	ParamIsPeriodic           = 1 << 1
	ParamIsHidden             = 1 << 2
	ParamIsReadOnly           = 1 << 3
	ParamIsBypass             = 1 << 4
	ParamIsAutomatable        = 1 << 5
	ParamIsAutomatablePerNoteID = 1 << 6
	ParamIsAutomatablePerKey   = 1 << 7
	ParamIsAutomatablePerChannel = 1 << 8
	ParamIsAutomatablePerPort  = 1 << 9
	ParamIsModulatable        = 1 << 10
	ParamIsModulatablePerNoteID = 1 << 11
	ParamIsModulatablePerKey   = 1 << 12
	ParamIsModulatablePerChannel = 1 << 13
	ParamIsModulatablePerPort  = 1 << 14
	ParamRequiresProcess      = 1 << 15
	
	// Additional common flags
	ParamIsBoundedBelow       = 1 << 16
	ParamIsBoundedAbove       = 1 << 17
)

// Parameter info
type ParamInfo struct {
	ID          uint32
	Name        string
	Module      string
	MinValue    float64
	MaxValue    float64
	DefaultValue float64
	Flags       uint32
}

// ParamManager provides an interface to manage plugin parameters
type ParamManager struct {
	params       map[uint32]*ParamInfo
	paramValues  map[uint32]float64
}

// NewParamManager creates a new parameter manager
func NewParamManager() *ParamManager {
	return &ParamManager{
		params:      make(map[uint32]*ParamInfo),
		paramValues: make(map[uint32]float64),
	}
}

// RegisterParam adds a parameter to the manager
func (pm *ParamManager) RegisterParam(param ParamInfo) {
	pm.params[param.ID] = &param
	pm.paramValues[param.ID] = param.DefaultValue
}

// GetParamCount returns the number of parameters
func (pm *ParamManager) GetParamCount() uint32 {
	return uint32(len(pm.params))
}

// GetParamInfo returns the parameter info
func (pm *ParamManager) GetParamInfo(paramID uint32) *ParamInfo {
	return pm.params[paramID]
}

// GetParamValue returns the current value of a parameter
func (pm *ParamManager) GetParamValue(paramID uint32) float64 {
	if value, exists := pm.paramValues[paramID]; exists {
		return value
	}
	
	// Return default value if parameter exists but value doesn't
	if param, exists := pm.params[paramID]; exists {
		return param.DefaultValue
	}
	
	return 0
}

// SetParamValue updates the value of a parameter
func (pm *ParamManager) SetParamValue(paramID uint32, value float64) bool {
	if _, exists := pm.params[paramID]; !exists {
		return false
	}
	
	pm.paramValues[paramID] = value
	return true
}

// Flush parameter changes
func (pm *ParamManager) Flush() {
	// This would trigger callbacks or update internal state based on parameter changes
}

// AddParamValueEvent adds a parameter value change event to an output event queue
func (pm *ParamManager) AddParamValueEvent(outEvents *OutputEvents, paramID uint32, value float64, time uint32) bool {
	if outEvents == nil {
		return false
	}
	
	// Create a parameter value event
	event := &Event{
		Time:  time,
		Type:  EventTypeParamValue,
		Space: 0, // Main event space
		Data: ParamEvent{
			ParamID: paramID,
			Value:   value,
		},
	}
	
	// Push the event
	return outEvents.PushEvent(event)
}

// BeginParamGesture begins a parameter gesture (e.g., user starts dragging a knob)
func (pm *ParamManager) BeginParamGesture(outEvents *OutputEvents, paramID uint32, time uint32) bool {
	if outEvents == nil {
		return false
	}
	
	// Create a gesture begin event
	event := &Event{
		Time:  time,
		Type:  EventParamGestureBegin,
		Space: 0, // Main event space
		Data: ParamEvent{
			ParamID: paramID,
		},
	}
	
	// Push the event
	return outEvents.PushEvent(event)
}

// EndParamGesture ends a parameter gesture (e.g., user releases a knob)
func (pm *ParamManager) EndParamGesture(outEvents *OutputEvents, paramID uint32, time uint32) bool {
	if outEvents == nil {
		return false
	}
	
	// Create a gesture end event
	event := &Event{
		Time:  time,
		Type:  EventParamGestureEnd,
		Space: 0, // Main event space
		Data: ParamEvent{
			ParamID: paramID,
		},
	}
	
	// Push the event
	return outEvents.PushEvent(event)
}

// NormalizeValue converts a parameter value to the normalized [0,1] range
func (pm *ParamManager) NormalizeValue(paramID uint32, value float64) float64 {
	param := pm.GetParamInfo(paramID)
	if param == nil {
		return 0
	}
	
	// If min and max are the same, return 0
	if param.MinValue == param.MaxValue {
		return 0
	}
	
	// Normalize to [0,1] range
	return (value - param.MinValue) / (param.MaxValue - param.MinValue)
}

// DenormalizeValue converts a normalized [0,1] value to the parameter range
func (pm *ParamManager) DenormalizeValue(paramID uint32, normalized float64) float64 {
	param := pm.GetParamInfo(paramID)
	if param == nil {
		return 0
	}
	
	// Denormalize to parameter range
	return param.MinValue + normalized * (param.MaxValue - param.MinValue)
}

// SetParamByName sets a parameter value by name instead of ID
func (pm *ParamManager) SetParamByName(name string, value float64) bool {
	// Find the parameter ID by name
	for id, param := range pm.params {
		if param.Name == name {
			return pm.SetParamValue(id, value)
		}
	}
	return false
}

// GetParamInfoFromC converts C parameter info to Go
func GetParamInfoFromC(cInfo *C.clap_param_info_t) ParamInfo {
	if cInfo == nil {
		return ParamInfo{}
	}
	
	return ParamInfo{
		ID:           uint32(cInfo.id),
		Name:         C.GoString(&cInfo.name[0]),
		Module:       C.GoString(&cInfo.module[0]),
		MinValue:     float64(cInfo.min_value),
		MaxValue:     float64(cInfo.max_value),
		DefaultValue: float64(cInfo.default_value),
		Flags:        uint32(cInfo.flags),
	}
}

// ToCParamInfo converts Go parameter info to C
func (pi *ParamInfo) ToCParamInfo() *C.clap_param_info_t {
	if pi == nil {
		return nil
	}
	
	cInfo := (*C.clap_param_info_t)(C.malloc(C.sizeof_clap_param_info_t))
	if cInfo == nil {
		return nil
	}
	
	cInfo.id = C.clap_id(pi.ID)
	cInfo.flags = C.uint32_t(pi.Flags)
	cInfo.min_value = C.double(pi.MinValue)
	cInfo.max_value = C.double(pi.MaxValue)
	cInfo.default_value = C.double(pi.DefaultValue)
	
	// Copy strings with proper null termination
	copyString(pi.Name, &cInfo.name[0], C.CLAP_NAME_SIZE)
	copyString(pi.Module, &cInfo.module[0], C.CLAP_PATH_SIZE)
	
	return cInfo
}

// Helper to copy Go string to fixed-size C char array
func copyString(src string, dst *C.char, maxLen C.int) {
	cStr := C.CString(src)
	defer C.free(unsafe.Pointer(cStr))
	
	C.strncpy(dst, cStr, C.size_t(maxLen-1))
	
	// Ensure null termination by setting the last byte to 0
	// Calculate the pointer to the last byte
	lastByte := unsafe.Pointer(uintptr(unsafe.Pointer(dst)) + uintptr(maxLen-1))
	*(*C.char)(lastByte) = 0
}