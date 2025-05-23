// Package util provides common utility functions for the ClapGo framework.
package util

import (
	"unsafe"
)

// AtomicFloat64ToBits converts a float64 to uint64 bits for atomic operations.
// This is useful when you need to store float64 values atomically using atomic.LoadInt64/StoreInt64.
func AtomicFloat64ToBits(f float64) uint64 {
	return *(*uint64)(unsafe.Pointer(&f))
}

// AtomicFloat64FromBits converts uint64 bits back to float64 from atomic operations.
// This is the reverse of AtomicFloat64ToBits.
func AtomicFloat64FromBits(b uint64) float64 {
	return *(*float64)(unsafe.Pointer(&b))
}