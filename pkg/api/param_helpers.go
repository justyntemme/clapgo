package api

// #include "../../include/clap/include/clap/clap.h"
// #include <string.h>
import "C"
import (
	"fmt"
	"math"
	"sync/atomic"
	"unsafe"
)

// ParamInfoToC converts a Go ParamInfo struct to a C clap_param_info_t struct
// This helper reduces boilerplate in plugin implementations
func ParamInfoToC(paramInfo ParamInfo, cInfo unsafe.Pointer) {
	info := (*C.clap_param_info_t)(cInfo)
	
	// Set basic fields
	info.id = C.clap_id(paramInfo.ID)
	info.flags = C.CLAP_PARAM_IS_AUTOMATABLE | C.CLAP_PARAM_IS_MODULATABLE
	
	// Additional flags based on parameter properties
	if paramInfo.Flags&ParamIsSteppable != 0 {
		info.flags |= C.CLAP_PARAM_IS_STEPPED
	}
	if paramInfo.Flags&ParamIsReadonly != 0 {
		info.flags |= C.CLAP_PARAM_IS_READONLY
	}
	if paramInfo.Flags&ParamIsHidden != 0 {
		info.flags |= C.CLAP_PARAM_IS_HIDDEN
	}
	if paramInfo.Flags&ParamIsBypass != 0 {
		info.flags |= C.CLAP_PARAM_IS_BYPASS
	}
	
	info.cookie = nil
	
	// Copy name
	CopyStringToCBuffer(paramInfo.Name, unsafe.Pointer(&info.name[0]), C.CLAP_NAME_SIZE)
	
	// Copy module path if present
	if paramInfo.Module != "" {
		CopyStringToCBuffer(paramInfo.Module, unsafe.Pointer(&info.module[0]), C.CLAP_PATH_SIZE)
	} else {
		info.module[0] = 0
	}
	
	// Set range
	info.min_value = C.double(paramInfo.MinValue)
	info.max_value = C.double(paramInfo.MaxValue)
	info.default_value = C.double(paramInfo.DefaultValue)
}

// CopyStringToCBuffer safely copies a Go string to a C char buffer
func CopyStringToCBuffer(str string, buffer unsafe.Pointer, maxSize int) {
	bytes := []byte(str)
	if len(bytes) >= maxSize {
		bytes = bytes[:maxSize-1]
	}
	
	// Copy bytes to C buffer
	for i, b := range bytes {
		*(*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(buffer)) + uintptr(i))) = C.char(b)
	}
	
	// Null terminate
	*(*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(buffer)) + uintptr(len(bytes)))) = 0
}

// FormatParameterValue formats a parameter value for display
type ParameterFormat int

const (
	FormatDefault ParameterFormat = iota
	FormatDecibel        // 20 * log10(value) with "dB" suffix
	FormatPercentage     // value * 100 with "%" suffix
	FormatMilliseconds   // value * 1000 with "ms" suffix
	FormatSeconds        // value with "s" suffix
	FormatHertz          // value with "Hz" suffix
	FormatKilohertz      // value / 1000 with "kHz" suffix
)

// FormatParameterValue formats a parameter value according to the specified format
func FormatParameterValue(value float64, format ParameterFormat) string {
	switch format {
	case FormatDecibel:
		if value <= 0 {
			return "-∞ dB"
		}
		db := 20.0 * math.Log10(value)
		return fmt.Sprintf("%.1f dB", db)
		
	case FormatPercentage:
		return fmt.Sprintf("%.1f%%", value*100.0)
		
	case FormatMilliseconds:
		return fmt.Sprintf("%.0f ms", value*1000.0)
		
	case FormatSeconds:
		return fmt.Sprintf("%.2f s", value)
		
	case FormatHertz:
		return fmt.Sprintf("%.1f Hz", value)
		
	case FormatKilohertz:
		return fmt.Sprintf("%.2f kHz", value/1000.0)
		
	default:
		return fmt.Sprintf("%.3f", value)
	}
}

// AtomicFloat64 provides atomic operations for float64 values
// This is useful for thread-safe parameter storage
type AtomicFloat64 struct {
	bits int64
}

// Store atomically stores a float64 value
func (a *AtomicFloat64) Store(value float64) {
	atomic.StoreInt64(&a.bits, int64(FloatToBits(value)))
}

// Load atomically loads a float64 value
func (a *AtomicFloat64) Load() float64 {
	bits := atomic.LoadInt64(&a.bits)
	return FloatFromBits(uint64(bits))
}

// CompareAndSwap atomically compares and swaps float64 values
func (a *AtomicFloat64) CompareAndSwap(old, new float64) bool {
	oldBits := int64(FloatToBits(old))
	newBits := int64(FloatToBits(new))
	return atomic.CompareAndSwapInt64(&a.bits, oldBits, newBits)
}

// FloatToBits converts float64 to uint64 bit representation
func FloatToBits(f float64) uint64 {
	return *(*uint64)(unsafe.Pointer(&f))
}

// FloatFromBits converts uint64 bit representation to float64
func FloatFromBits(b uint64) float64 {
	return *(*float64)(unsafe.Pointer(&b))
}

// ClampValue ensures a value is within the specified range
func ClampValue(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Common parameter creators for typical use cases

// CreateVolumeParameter creates a standard volume/gain parameter
func CreateVolumeParameter(id uint32, name string) ParamInfo {
	return CreateFloatParameter(id, name, 0.0, 2.0, 1.0) // 0dB to +6dB
}

// CreatePercentParameter creates a 0-100% parameter
func CreatePercentParameter(id uint32, name string, defaultPercent float64) ParamInfo {
	return CreateFloatParameter(id, name, 0.0, 1.0, defaultPercent/100.0)
}

// CreateTimeParameter creates a time parameter in seconds
func CreateTimeParameter(id uint32, name string, minSeconds, maxSeconds, defaultSeconds float64) ParamInfo {
	return CreateFloatParameter(id, name, minSeconds, maxSeconds, defaultSeconds)
}

// CreateFrequencyParameter creates a frequency parameter in Hz
func CreateFrequencyParameter(id uint32, name string, minHz, maxHz, defaultHz float64) ParamInfo {
	return CreateFloatParameter(id, name, minHz, maxHz, defaultHz)
}

// CreateADSRParameter creates standard ADSR envelope parameters
type ADSRParameters struct {
	Attack  ParamInfo
	Decay   ParamInfo
	Sustain ParamInfo
	Release ParamInfo
}

// CreateADSRParameters creates a standard set of ADSR parameters
func CreateADSRParameters(baseID uint32, prefix string) ADSRParameters {
	return ADSRParameters{
		Attack:  CreateTimeParameter(baseID, prefix+" Attack", 0.001, 2.0, 0.01),
		Decay:   CreateTimeParameter(baseID+1, prefix+" Decay", 0.001, 2.0, 0.1),
		Sustain: CreatePercentParameter(baseID+2, prefix+" Sustain", 70.0),
		Release: CreateTimeParameter(baseID+3, prefix+" Release", 0.001, 5.0, 0.5),
	}
}

// ParameterValueParser helps parse parameter values from text
type ParameterValueParser struct {
	format ParameterFormat
}

// NewParameterValueParser creates a new parser for the given format
func NewParameterValueParser(format ParameterFormat) *ParameterValueParser {
	return &ParameterValueParser{format: format}
}

// ParseValue attempts to parse a formatted string back to a raw value
func (p *ParameterValueParser) ParseValue(text string) (float64, error) {
	// Remove format suffixes
	switch p.format {
	case FormatDecibel:
		var db float64
		if _, err := fmt.Sscanf(text, "%f dB", &db); err == nil {
			// Convert dB back to linear
			return math.Pow(10.0, db/20.0), nil
		}
		
	case FormatPercentage:
		var percent float64
		if _, err := fmt.Sscanf(text, "%f%%", &percent); err == nil {
			return percent / 100.0, nil
		}
		
	case FormatMilliseconds:
		var ms float64
		if _, err := fmt.Sscanf(text, "%f ms", &ms); err == nil {
			return ms / 1000.0, nil
		}
		
	case FormatSeconds:
		var s float64
		if _, err := fmt.Sscanf(text, "%f s", &s); err == nil {
			return s, nil
		}
		
	case FormatHertz:
		var hz float64
		if _, err := fmt.Sscanf(text, "%f Hz", &hz); err == nil {
			return hz, nil
		}
		
	case FormatKilohertz:
		var khz float64
		if _, err := fmt.Sscanf(text, "%f kHz", &khz); err == nil {
			return khz * 1000.0, nil
		}
	}
	
	// Try parsing as plain number
	var value float64
	_, err := fmt.Sscanf(text, "%f", &value)
	return value, err
}