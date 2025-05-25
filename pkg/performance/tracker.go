// Package performance provides real-time performance monitoring and allocation tracking
// for CLAP audio plugins. This package helps validate zero-allocation goals.
package performance

import (
	"runtime"
	"sync/atomic"
	"time"
)

// AllocInfo stores information about a single allocation
type AllocInfo struct {
	Timestamp uint64      // Unix nano timestamp
	Size      uintptr     // Allocation size in bytes
	Stack     [8]uintptr  // Stack trace (up to 8 frames)
}

// AllocationTracker monitors allocations in the audio thread
type AllocationTracker struct {
	// Counters
	audioThreadAllocs  uint64 // Total allocations in audio thread
	maxAllocsPerBuffer uint64 // Maximum allocations in a single buffer
	currentBufferAllocs uint64 // Allocations in current buffer
	
	// Ring buffer for allocation history
	allocBuffer [1000]AllocInfo
	allocIndex  int32 // Atomic index for lock-free access
	
	// Configuration
	enabled bool
}

// NewAllocationTracker creates a new allocation tracker
func NewAllocationTracker() *AllocationTracker {
	return &AllocationTracker{
		enabled: true,
	}
}

// Enable enables allocation tracking
func (at *AllocationTracker) Enable() {
	at.enabled = true
}

// Disable disables allocation tracking
func (at *AllocationTracker) Disable() {
	at.enabled = false
}

// TrackAllocation records an allocation (only in debug builds)
// This should be called from a custom allocator or runtime hook
func (at *AllocationTracker) TrackAllocation(size uintptr) {
	if !at.enabled {
		return
	}
	
	// Capture stack trace for debugging
	var stack [8]uintptr
	n := runtime.Callers(2, stack[:])
	for i := n; i < len(stack); i++ {
		stack[i] = 0
	}
	
	// Increment counters
	atomic.AddUint64(&at.audioThreadAllocs, 1)
	atomic.AddUint64(&at.currentBufferAllocs, 1)
	
	// Store allocation info in ring buffer (lock-free)
	idx := atomic.AddInt32(&at.allocIndex, 1) % 1000
	at.allocBuffer[idx] = AllocInfo{
		Timestamp: uint64(time.Now().UnixNano()),
		Size:      size,
		Stack:     stack,
	}
}

// StartBuffer marks the beginning of a new audio buffer processing
func (at *AllocationTracker) StartBuffer() {
	if !at.enabled {
		return
	}
	
	// Reset current buffer allocation count
	atomic.StoreUint64(&at.currentBufferAllocs, 0)
}

// EndBuffer marks the end of audio buffer processing
func (at *AllocationTracker) EndBuffer() {
	if !at.enabled {
		return
	}
	
	// Update max allocations per buffer if needed
	current := atomic.LoadUint64(&at.currentBufferAllocs)
	for {
		max := atomic.LoadUint64(&at.maxAllocsPerBuffer)
		if current <= max {
			break
		}
		if atomic.CompareAndSwapUint64(&at.maxAllocsPerBuffer, max, current) {
			break
		}
	}
}

// GetStats returns current allocation statistics
func (at *AllocationTracker) GetStats() AllocationStats {
	return AllocationStats{
		TotalAllocations:   atomic.LoadUint64(&at.audioThreadAllocs),
		MaxAllocsPerBuffer: atomic.LoadUint64(&at.maxAllocsPerBuffer),
		CurrentBufferAllocs: atomic.LoadUint64(&at.currentBufferAllocs),
	}
}

// Reset resets all allocation statistics
func (at *AllocationTracker) Reset() {
	atomic.StoreUint64(&at.audioThreadAllocs, 0)
	atomic.StoreUint64(&at.maxAllocsPerBuffer, 0)
	atomic.StoreUint64(&at.currentBufferAllocs, 0)
	atomic.StoreInt32(&at.allocIndex, 0)
}

// AllocationStats contains allocation statistics
type AllocationStats struct {
	TotalAllocations    uint64
	MaxAllocsPerBuffer  uint64
	CurrentBufferAllocs uint64
}