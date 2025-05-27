//go:build debug
// +build debug

package thread

import (
	"fmt"
	"runtime"
)

// DebugChecker performs comprehensive thread validation in debug builds
type DebugChecker struct {
	mainThreadID   uint64
	audioThreadIDs map[uint64]bool
}

// NewDebugChecker creates a debug thread checker
func NewDebugChecker() *DebugChecker {
	return &DebugChecker{
		audioThreadIDs: make(map[uint64]bool),
	}
}

// SetMainThread marks the current thread as the main thread
func (dtc *DebugChecker) SetMainThread() {
	dtc.mainThreadID = getThreadID()
}

// MarkAudioThread marks the current thread as an audio thread
func (dtc *DebugChecker) MarkAudioThread() {
	dtc.audioThreadIDs[getThreadID()] = true
}

// UnmarkAudioThread removes the current thread from audio threads
func (dtc *DebugChecker) UnmarkAudioThread() {
	delete(dtc.audioThreadIDs, getThreadID())
}

// ValidateMainThread validates we're on the main thread
func (dtc *DebugChecker) ValidateMainThread(operation string) {
	currentID := getThreadID()
	if currentID != dtc.mainThreadID {
		panic(fmt.Sprintf("Thread violation: %s called from thread %d, expected main thread %d",
			operation, currentID, dtc.mainThreadID))
	}
}

// ValidateAudioThread validates we're on an audio thread
func (dtc *DebugChecker) ValidateAudioThread(operation string) {
	currentID := getThreadID()
	if !dtc.audioThreadIDs[currentID] {
		panic(fmt.Sprintf("Thread violation: %s called from non-audio thread %d",
			operation, currentID))
	}
}

// ValidateNotAudioThread validates we're NOT on an audio thread
func (dtc *DebugChecker) ValidateNotAudioThread(operation string) {
	currentID := getThreadID()
	if dtc.audioThreadIDs[currentID] {
		panic(fmt.Sprintf("Thread violation: %s called from audio thread %d (not allowed)",
			operation, currentID))
	}
}

// getThreadID returns a unique ID for the current OS thread
func getThreadID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	// Extract goroutine ID from stack trace
	// Format: "goroutine <id> [...]"
	for i := 10; i < n-1; i++ {
		if buf[i] == ' ' {
			id := uint64(0)
			for j := i + 1; j < n; j++ {
				if buf[j] < '0' || buf[j] > '9' {
					break
				}
				id = id*10 + uint64(buf[j]-'0')
			}
			return id
		}
	}
	return 0
}

// Global debug thread checker
var debugChecker = NewDebugChecker()

// SetMainThread marks the current thread as main (debug builds only)
func SetMainThread() {
	debugChecker.SetMainThread()
}

// MarkAudioThread marks the current thread as audio (debug builds only)
func MarkAudioThread() {
	debugChecker.MarkAudioThread()
}

// UnmarkAudioThread unmarks the current thread as audio (debug builds only)
func UnmarkAudioThread() {
	debugChecker.UnmarkAudioThread()
}

// AssertMainThread panics if not on main thread (debug builds only)
func AssertMainThread(operation string) {
	debugChecker.ValidateMainThread(operation)
}

// AssertAudioThread panics if not on audio thread (debug builds only)
func AssertAudioThread(operation string) {
	debugChecker.ValidateAudioThread(operation)
}

// AssertNotAudioThread panics if on audio thread (debug builds only)
func AssertNotAudioThread(operation string) {
	debugChecker.ValidateNotAudioThread(operation)
}