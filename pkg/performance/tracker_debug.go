//go:build debug
// +build debug

package performance

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"time"
)

// EnableAllocationTracking enables runtime allocation tracking in debug builds
func EnableAllocationTracking() {
	// Set GC percentage to make allocations more visible
	debug.SetGCPercent(10)
	
	// Log when this is enabled
	fmt.Println("ClapGo: Allocation tracking enabled (debug build)")
}

// CheckGCPauses checks if a GC pause occurred recently
func CheckGCPauses() bool {
	var stats debug.GCStats
	debug.ReadGCStats(&stats)
	
	if len(stats.PauseEnd) > 0 {
		// Check if last pause was within 1ms (recent)
		lastPause := stats.PauseEnd[0]
		if time.Since(lastPause) < time.Millisecond {
			return true
		}
	}
	return false
}

// GetAllocationStats returns detailed allocation statistics
func GetAllocationStats() runtime.MemStats {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return stats
}

// FormatAllocInfo formats allocation info with stack trace
func FormatAllocInfo(info AllocInfo) string {
	frames := runtime.CallersFrames(info.Stack[:])
	
	result := fmt.Sprintf("Allocation: %d bytes at %s\n", info.Size, time.Unix(0, int64(info.Timestamp)))
	result += "Stack trace:\n"
	
	for {
		frame, more := frames.Next()
		if frame.PC == 0 {
			break
		}
		result += fmt.Sprintf("  %s:%d in %s\n", frame.File, frame.Line, frame.Function)
		if !more {
			break
		}
	}
	
	return result
}