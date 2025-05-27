package param

import (
	"sync/atomic"
)

// AtomicFloat64 provides atomic operations for float64 values
type AtomicFloat64 struct {
	bits int64
}

// NewAtomicFloat64 creates a new atomic float64 with the given initial value
func NewAtomicFloat64(initial float64) *AtomicFloat64 {
	return &AtomicFloat64{
		bits: floatToBits(initial),
	}
}

// Load atomically loads and returns the stored float64 value
func (a *AtomicFloat64) Load() float64 {
	bits := atomic.LoadInt64(&a.bits)
	return bitsToFloat(bits)
}

// Store atomically stores the given float64 value
func (a *AtomicFloat64) Store(value float64) {
	atomic.StoreInt64(&a.bits, floatToBits(value))
}

// Swap atomically stores new value and returns the previous value
func (a *AtomicFloat64) Swap(new float64) float64 {
	bits := atomic.SwapInt64(&a.bits, floatToBits(new))
	return bitsToFloat(bits)
}

// CompareAndSwap executes the compare-and-swap operation for a float64 value
func (a *AtomicFloat64) CompareAndSwap(old, new float64) bool {
	return atomic.CompareAndSwapInt64(&a.bits, floatToBits(old), floatToBits(new))
}

// Add atomically adds delta to the float64 value and returns the new value
func (a *AtomicFloat64) Add(delta float64) float64 {
	for {
		old := a.Load()
		new := old + delta
		if a.CompareAndSwap(old, new) {
			return new
		}
	}
}

// UpdateWithManager atomically updates the value and notifies the parameter manager
func (a *AtomicFloat64) UpdateWithManager(value float64, manager *Manager, paramID uint32) {
	a.Store(value)
	if manager != nil {
		_ = manager.SetValue(paramID, value)
	}
}

// LoadParameterAtomic loads a parameter value from atomic storage
// This is a compatibility function for existing code
func LoadParameterAtomic(bits *int64) float64 {
	return bitsToFloat(atomic.LoadInt64(bits))
}

// UpdateParameterAtomic updates atomic storage and parameter manager
// This is a compatibility function for existing code
func UpdateParameterAtomic(bits *int64, value float64, manager *Manager, paramID uint32) {
	atomic.StoreInt64(bits, floatToBits(value))
	if manager != nil {
		_ = manager.SetValue(paramID, value)
	}
}