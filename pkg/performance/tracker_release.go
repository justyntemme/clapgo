//go:build !debug
// +build !debug

package performance

import "runtime"

// EnableAllocationTracking is a no-op in release builds
func EnableAllocationTracking() {
	// No-op in release builds
}

// CheckGCPauses always returns false in release builds
func CheckGCPauses() bool {
	return false
}

// GetAllocationStats returns empty stats in release builds
func GetAllocationStats() runtime.MemStats {
	return runtime.MemStats{}
}

// FormatAllocInfo returns empty string in release builds
func FormatAllocInfo(info AllocInfo) string {
	return ""
}