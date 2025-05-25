//go:build debug
// +build debug

package api

import (
	"fmt"
	"runtime"
)

// DebugThreadCheck performs comprehensive thread validation in debug builds
type DebugThreadCheck struct {
	mainThreadID   uint64
	audioThreadIDs map[uint64]bool
}

// NewDebugThreadCheck creates a debug thread checker
func NewDebugThreadCheck() *DebugThreadCheck {
	return &DebugThreadCheck{
		audioThreadIDs: make(map[uint64]bool),
	}
}

// SetMainThread marks the current thread as the main thread
func (dtc *DebugThreadCheck) SetMainThread() {
	dtc.mainThreadID = getThreadID()
}

// MarkAudioThread marks the current thread as an audio thread
func (dtc *DebugThreadCheck) MarkAudioThread() {
	dtc.audioThreadIDs[getThreadID()] = true
}

// UnmarkAudioThread removes the current thread from audio threads
func (dtc *DebugThreadCheck) UnmarkAudioThread() {
	delete(dtc.audioThreadIDs, getThreadID())
}

// ValidateMainThread validates we're on the main thread
func (dtc *DebugThreadCheck) ValidateMainThread(operation string) {
	currentID := getThreadID()
	if currentID != dtc.mainThreadID {
		panic(fmt.Sprintf("Thread violation: %s called from thread %d, expected main thread %d",
			operation, currentID, dtc.mainThreadID))
	}
}

// ValidateAudioThread validates we're on an audio thread
func (dtc *DebugThreadCheck) ValidateAudioThread(operation string) {
	currentID := getThreadID()
	if !dtc.audioThreadIDs[currentID] {
		panic(fmt.Sprintf("Thread violation: %s called from non-audio thread %d",
			operation, currentID))
	}
}

// ValidateNotAudioThread validates we're NOT on an audio thread
func (dtc *DebugThreadCheck) ValidateNotAudioThread(operation string) {
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
var debugThreadCheck = NewDebugThreadCheck()

// DebugSetMainThread marks the current thread as main (debug builds only)
func DebugSetMainThread() {
	debugThreadCheck.SetMainThread()
}

// DebugMarkAudioThread marks the current thread as audio (debug builds only)
func DebugMarkAudioThread() {
	debugThreadCheck.MarkAudioThread()
}

// DebugUnmarkAudioThread unmarks the current thread as audio (debug builds only)
func DebugUnmarkAudioThread() {
	debugThreadCheck.UnmarkAudioThread()
}

// DebugAssertMainThread panics if not on main thread (debug builds only)
func DebugAssertMainThread(operation string) {
	debugThreadCheck.ValidateMainThread(operation)
}

// DebugAssertAudioThread panics if not on audio thread (debug builds only)
func DebugAssertAudioThread(operation string) {
	debugThreadCheck.ValidateAudioThread(operation)
}

// DebugAssertNotAudioThread panics if on audio thread (debug builds only)
func DebugAssertNotAudioThread(operation string) {
	debugThreadCheck.ValidateNotAudioThread(operation)
}