package goclap

// #include <stdlib.h>
// #include <string.h>
// #include "../../include/clap/include/clap/clap.h"
// #include "../../include/clap/include/clap/ext/params.h"
import "C"
import (
	"unsafe"
)

// Constants for parameter extension
const (
	ExtParams = "clap.params"
)

// Parameter event types
const (
	EventParamValueType   = 4
	EventParamModType     = 5
	EventParamGestureBegin = 10
	EventParamGestureEnd   = 11
)

// Parameter clear flags
const (
	ParamClearAll          = 1 << 0
	ParamClearAutomations  = 1 << 1
	ParamClearModulations  = 1 << 2
)

// Parameter rescan flags
const (
	ParamRescanValues = 1 << 0
	ParamRescanText   = 1 << 1
	ParamRescanInfo   = 1 << 2
	ParamRescanAll    = 1 << 3
)

// PluginParamsExtension represents the CLAP parameters extension for plugins
type PluginParamsExtension struct {
	plugin           AudioProcessor
	paramManager     *ParamManager
	paramExtInstance unsafe.Pointer // Pointer to the C extension
}

// NewPluginParamsExtension creates a new parameters extension
func NewPluginParamsExtension(plugin AudioProcessor, paramManager *ParamManager) *PluginParamsExtension {
	if plugin == nil || paramManager == nil {
		return nil
	}
	
	ext := &PluginParamsExtension{
		plugin:       plugin,
		paramManager: paramManager,
	}
	
	// Create the C interface
	ext.paramExtInstance = createParamExtension(ext)
	
	return ext
}

// GetExtensionPointer returns the C extension interface pointer
func (p *PluginParamsExtension) GetExtensionPointer() unsafe.Pointer {
	return p.paramExtInstance
}

// CountParameters returns the number of parameters
func (p *PluginParamsExtension) CountParameters() uint32 {
	if p.paramManager == nil {
		return 0
	}
	
	return p.paramManager.GetParamCount()
}

// GetParameterInfo gets parameter info by index
func (p *PluginParamsExtension) GetParameterInfo(index uint32, info *C.clap_param_info_t) bool {
	if p.paramManager == nil || info == nil || int(index) >= int(p.paramManager.GetParamCount()) {
		return false
	}
	
	// Get parameter ID by index
	var paramID uint32
	found := false
	
	// Find the parameter with the given index
	i := uint32(0)
	for id := range p.paramManager.params {
		if i == index {
			paramID = id
			found = true
			break
		}
		i++
	}
	
	if !found {
		return false
	}
	
	// Get parameter info
	paramInfo := p.paramManager.GetParamInfo(paramID)
	if paramInfo == nil {
		return false
	}
	
	// Convert to C parameter info
	info.id = C.clap_id(paramInfo.ID)
	info.flags = C.uint32_t(paramInfo.Flags)
	info.min_value = C.double(paramInfo.MinValue)
	info.max_value = C.double(paramInfo.MaxValue)
	info.default_value = C.double(paramInfo.DefaultValue)
	info.cookie = nil // This could be set to a Go pointer if needed
	
	// Copy the name and module
	copyString(paramInfo.Name, &info.name[0], C.CLAP_NAME_SIZE)
	copyString(paramInfo.Module, &info.module[0], C.CLAP_PATH_SIZE)
	
	return true
}

// GetParameterValue gets the current value of a parameter
func (p *PluginParamsExtension) GetParameterValue(paramID uint32) (float64, bool) {
	if p.paramManager == nil {
		return 0, false
	}
	
	// Check if parameter exists
	if _, exists := p.paramManager.params[paramID]; !exists {
		return 0, false
	}
	
	return p.paramManager.GetParamValue(paramID), true
}

// ValueToText converts a parameter value to text
func (p *PluginParamsExtension) ValueToText(paramID uint32, value float64, buffer *C.char, capacity uint32) bool {
	if p.paramManager == nil || buffer == nil || capacity == 0 {
		return false
	}
	
	// Check if parameter exists
	paramInfo, exists := p.paramManager.params[paramID]
	if !exists {
		return false
	}
	
	// Default implementation just converts to string
	// In a real plugin, you'd have custom formatting per parameter
	var text string
	
	// Check if this is a stepped parameter
	if (paramInfo.Flags & ParamIsSteppable) != 0 {
		// Format as integer
		text = formatSteppedValue(value)
	} else {
		// Format as float with appropriate precision
		text = formatFloatValue(value)
	}
	
	// Copy to output buffer with null termination
	if len(text) >= int(capacity) {
		// Truncate if too long
		text = text[:int(capacity)-1]
	}
	
	C.strncpy(buffer, C.CString(text), C.size_t(capacity-1))
	*(*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(buffer)) + uintptr(capacity-1))) = 0
	
	return true
}

// TextToValue converts text to a parameter value
func (p *PluginParamsExtension) TextToValue(paramID uint32, text string) (float64, bool) {
	if p.paramManager == nil {
		return 0, false
	}
	
	// Check if parameter exists
	paramInfo, exists := p.paramManager.params[paramID]
	if !exists {
		return 0, false
	}
	
	// Default implementation just converts from string
	// In a real plugin, you'd have custom parsing per parameter
	var value float64
	var ok bool
	
	// Check if this is a stepped parameter
	if (paramInfo.Flags & ParamIsSteppable) != 0 {
		value, ok = parseSteppedValue(text)
	} else {
		value, ok = parseFloatValue(text)
	}
	
	if !ok {
		return 0, false
	}
	
	// Clamp to min/max range
	if value < paramInfo.MinValue {
		value = paramInfo.MinValue
	}
	if value > paramInfo.MaxValue {
		value = paramInfo.MaxValue
	}
	
	return value, true
}

// Flush parameter changes
func (p *PluginParamsExtension) Flush(inEvents, outEvents unsafe.Pointer) {
	if p.paramManager == nil {
		return
	}
	
	// Process any parameter changes from input events
	if inEvents != nil {
		inputEvents := &InputEvents{Ptr: inEvents}
		eventCount := inputEvents.GetEventCount()
		
		for i := uint32(0); i < eventCount; i++ {
			event := inputEvents.GetEvent(i)
			if event == nil {
				continue
			}
			
			// Handle parameter value events
			if event.Type == EventTypeParamValue {
				if paramEvent, ok := event.Data.(ParamEvent); ok {
					p.paramManager.SetParamValue(paramEvent.ParamID, paramEvent.Value)
				}
			}
		}
	}
	
	// Allow the parameter manager to generate any output events
	p.paramManager.Flush()
}

// Helper functions for value conversion

// formatSteppedValue formats a stepped value as an integer
func formatSteppedValue(value float64) string {
	return C.GoString(C.CString(formatInt(int(value))))
}

// formatFloatValue formats a float value with appropriate precision
func formatFloatValue(value float64) string {
	return C.GoString(C.CString(formatFloat(value, 6)))
}

// parseSteppedValue parses a stepped value from text
func parseSteppedValue(text string) (float64, bool) {
	val, ok := parseInt(text)
	return float64(val), ok
}

// parseFloatValue parses a float value from text
func parseFloatValue(text string) (float64, bool) {
	return parseFloat(text)
}

// External function declarations for the C bridge

//export goParamsCount
func goParamsCount(plugin unsafe.Pointer) C.uint32_t {
	ext := (*PluginParamsExtension)(plugin)
	if ext == nil {
		return 0
	}
	
	return C.uint32_t(ext.CountParameters())
}

//export goParamsGetInfo
func goParamsGetInfo(plugin unsafe.Pointer, index C.uint32_t, info *C.clap_param_info_t) C.bool {
	ext := (*PluginParamsExtension)(plugin)
	if ext == nil || info == nil {
		return C.bool(false)
	}
	
	return C.bool(ext.GetParameterInfo(uint32(index), info))
}

//export goParamsGetValue
func goParamsGetValue(plugin unsafe.Pointer, paramID C.clap_id, outValue *C.double) C.bool {
	ext := (*PluginParamsExtension)(plugin)
	if ext == nil || outValue == nil {
		return C.bool(false)
	}
	
	value, ok := ext.GetParameterValue(uint32(paramID))
	if ok {
		*outValue = C.double(value)
	}
	
	return C.bool(ok)
}

//export goParamsValueToText
func goParamsValueToText(plugin unsafe.Pointer, paramID C.clap_id, value C.double, 
	outBuffer *C.char, outBufferCapacity C.uint32_t) C.bool {
	
	ext := (*PluginParamsExtension)(plugin)
	if ext == nil || outBuffer == nil {
		return C.bool(false)
	}
	
	return C.bool(ext.ValueToText(uint32(paramID), float64(value), outBuffer, uint32(outBufferCapacity)))
}

//export goParamsTextToValue
func goParamsTextToValue(plugin unsafe.Pointer, paramID C.clap_id, paramValueText *C.char, outValue *C.double) C.bool {
	ext := (*PluginParamsExtension)(plugin)
	if ext == nil || paramValueText == nil || outValue == nil {
		return C.bool(false)
	}
	
	value, ok := ext.TextToValue(uint32(paramID), C.GoString(paramValueText))
	if ok {
		*outValue = C.double(value)
	}
	
	return C.bool(ok)
}

//export goParamsFlush
func goParamsFlush(plugin unsafe.Pointer, inEvents, outEvents unsafe.Pointer) {
	ext := (*PluginParamsExtension)(plugin)
	if ext == nil {
		return
	}
	
	ext.Flush(inEvents, outEvents)
}

// Helper function to create the C extension interface
func createParamExtension(ext *PluginParamsExtension) unsafe.Pointer {
	// Allocate memory for the C interface
	cExt := (*C.clap_plugin_params_t)(C.malloc(C.sizeof_clap_plugin_params_t))
	if cExt == nil {
		return nil
	}
	
	// Store the extension handle in the plugin data field
	// This would be handled by the CGO implementation once we
	// have that completed - for now, leave the function pointers
	// uninitialized as they'll be set from the C code side.
	
	// Return the extension interface
	return unsafe.Pointer(cExt)
}