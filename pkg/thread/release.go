//go:build !debug
// +build !debug

package thread

// In release builds, thread checking functions are no-ops for performance

// SetMainThread is a no-op in release builds
func SetMainThread() {}

// MarkAudioThread is a no-op in release builds
func MarkAudioThread() {}

// UnmarkAudioThread is a no-op in release builds
func UnmarkAudioThread() {}

// AssertMainThread is a no-op in release builds
func AssertMainThread(operation string) {}

// AssertAudioThread is a no-op in release builds
func AssertAudioThread(operation string) {}

// AssertNotAudioThread is a no-op in release builds
func AssertNotAudioThread(operation string) {}