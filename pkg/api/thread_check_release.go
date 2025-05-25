//go:build !debug
// +build !debug

package api

// Debug thread check functions are no-ops in release builds

// DebugSetMainThread is a no-op in release builds
func DebugSetMainThread() {}

// DebugMarkAudioThread is a no-op in release builds
func DebugMarkAudioThread() {}

// DebugUnmarkAudioThread is a no-op in release builds
func DebugUnmarkAudioThread() {}

// DebugAssertMainThread is a no-op in release builds
func DebugAssertMainThread(operation string) {}

// DebugAssertAudioThread is a no-op in release builds
func DebugAssertAudioThread(operation string) {}

// DebugAssertNotAudioThread is a no-op in release builds
func DebugAssertNotAudioThread(operation string) {}