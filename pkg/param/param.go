package param

import (
	"errors"
	"sync/atomic"
	"unsafe"
)

// Common parameter errors
var (
	ErrInvalidParam         = errors.New("invalid parameter ID")
	ErrListenerLimitReached = errors.New("parameter listener limit reached")
	ErrValueBelowMinimum    = errors.New("value below minimum")
	ErrValueAboveMaximum    = errors.New("value above maximum")
	ErrParameterExists      = errors.New("parameter ID already exists")
	ErrParamExists          = errors.New("parameter ID already exists") // Alias for compatibility
)

// MaxListeners is the maximum number of parameter change listeners
const MaxListeners = 16

// Flags for parameter capabilities
const (
	FlagAutomatable     uint32 = 1 << 0
	FlagModulatable     uint32 = 1 << 1
	FlagStepped         uint32 = 1 << 2
	FlagReadonly        uint32 = 1 << 3
	FlagHidden          uint32 = 1 << 4
	FlagBypass          uint32 = 1 << 5
	FlagBoundedBelow    uint32 = 1 << 6
	FlagBoundedAbove    uint32 = 1 << 7
	FlagRequiresProcess uint32 = 1 << 8
	
	// Aliases for compatibility
	IsBoundedBelow = FlagBoundedBelow
	IsBoundedAbove = FlagBoundedAbove
)

// Info represents parameter metadata
type Info struct {
	ID           uint32
	Name         string
	Module       string // Path for grouping (e.g. "Filter/Cutoff")
	MinValue     float64
	MaxValue     float64
	DefaultValue float64
	Flags        uint32
}

// Parameter represents a plugin parameter with thread-safe access
type Parameter struct {
	Info      Info
	value     int64 // atomic storage for float64 bits
	validator func(float64) error
}

// Value returns the current value atomically
func (p *Parameter) Value() float64 {
	bits := atomic.LoadInt64(&p.value)
	return bitsToFloat(bits)
}

// SetValue sets the value with validation
func (p *Parameter) SetValue(value float64) error {
	// Validate if validator exists
	if p.validator != nil {
		if err := p.validator(value); err != nil {
			return err
		}
	}
	
	// Store atomically
	atomic.StoreInt64(&p.value, floatToBits(value))
	return nil
}

// Float/bits conversion helpers
func floatToBits(f float64) int64 {
	return int64(*(*uint64)(unsafe.Pointer(&f)))
}

func bitsToFloat(b int64) float64 {
	return *(*float64)(unsafe.Pointer(&b))
}