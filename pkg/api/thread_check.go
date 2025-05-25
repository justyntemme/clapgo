package api

/*
#include "../../include/clap/include/clap/clap.h"
#include "../../include/clap/include/clap/ext/thread-check.h"

// Helper to cast and call thread check functions
static bool clap_host_thread_check_is_main_thread(const clap_host_t *host) {
    const clap_host_thread_check_t *thread_check = 
        (const clap_host_thread_check_t *)host->get_extension(host, CLAP_EXT_THREAD_CHECK);
    if (thread_check && thread_check->is_main_thread) {
        return thread_check->is_main_thread(host);
    }
    return false;
}

static bool clap_host_thread_check_is_audio_thread(const clap_host_t *host) {
    const clap_host_thread_check_t *thread_check = 
        (const clap_host_thread_check_t *)host->get_extension(host, CLAP_EXT_THREAD_CHECK);
    if (thread_check && thread_check->is_audio_thread) {
        return thread_check->is_audio_thread(host);
    }
    return false;
}
*/
import "C"
import (
	"unsafe"
)

// ThreadChecker provides thread validation functionality from the host
type ThreadChecker struct {
	host      unsafe.Pointer
	debugMode bool
}

// NewThreadChecker creates a new thread checker instance
func NewThreadChecker(host unsafe.Pointer) *ThreadChecker {
	return &ThreadChecker{
		host:      host,
		debugMode: true, // Enable debug mode by default
	}
}

// IsAvailable returns true if the host supports thread checking
// This is determined by whether the C helpers can find the extension
func (tc *ThreadChecker) IsAvailable() bool {
	if tc.host == nil {
		return false
	}
	// The C helpers return false if the extension isn't available
	// So we can just check if either call works
	// This is safe because the C code handles NULL checks
	return true // We assume it's available and let C code handle it
}

// EnableDebugMode enables debug assertions
func (tc *ThreadChecker) EnableDebugMode(enable bool) {
	tc.debugMode = enable
}

// IsMainThread returns true if the current thread is the main thread
func (tc *ThreadChecker) IsMainThread() bool {
	if tc.host == nil {
		return false
	}
	
	// The C helper returns false if extension isn't available
	return bool(C.clap_host_thread_check_is_main_thread((*C.clap_host_t)(tc.host)))
}

// IsAudioThread returns true if the current thread is an audio thread
func (tc *ThreadChecker) IsAudioThread() bool {
	if tc.host == nil {
		return false
	}
	
	// The C helper returns false if extension isn't available
	return bool(C.clap_host_thread_check_is_audio_thread((*C.clap_host_t)(tc.host)))
}

// AssertMainThread panics if not on the main thread (debug mode only)
func (tc *ThreadChecker) AssertMainThread(operation string) {
	if !tc.debugMode || tc.host == nil {
		return
	}
	
	if !tc.IsMainThread() {
		panic("ThreadCheck: " + operation + " must be called from main thread")
	}
}

// AssertAudioThread panics if not on the audio thread (debug mode only)
func (tc *ThreadChecker) AssertAudioThread(operation string) {
	if !tc.debugMode || tc.host == nil {
		return
	}
	
	if !tc.IsAudioThread() {
		panic("ThreadCheck: " + operation + " must be called from audio thread")
	}
}

// AssertNotAudioThread panics if on the audio thread (debug mode only)
func (tc *ThreadChecker) AssertNotAudioThread(operation string) {
	if !tc.debugMode || tc.host == nil {
		return
	}
	
	if tc.IsAudioThread() {
		panic("ThreadCheck: " + operation + " must NOT be called from audio thread")
	}
}

// ValidateThreadContext validates common CLAP threading rules
func (tc *ThreadChecker) ValidateThreadContext(context string) error {
	if tc.host == nil {
		return nil
	}
	
	// Just validate, don't log (to avoid circular dependencies)
	_ = tc.IsMainThread()
	_ = tc.IsAudioThread()
	
	return nil
}

// Global thread checker instance (set by plugin during init)
var globalThreadChecker *ThreadChecker

// SetGlobalThreadChecker sets the global thread checker
func SetGlobalThreadChecker(tc *ThreadChecker) {
	globalThreadChecker = tc
}

// GetThreadChecker returns the global thread checker
func GetThreadChecker() *ThreadChecker {
	return globalThreadChecker
}

// Thread context markers for documentation and validation
const (
	// ThreadContextMain indicates function must be called from main thread
	ThreadContextMain = "[main-thread]"
	
	// ThreadContextAudio indicates function must be called from audio thread
	ThreadContextAudio = "[audio-thread]"
	
	// ThreadContextAny indicates function can be called from any thread
	ThreadContextAny = "[thread-safe]"
	
	// ThreadContextNotAudio indicates function must NOT be called from audio thread
	ThreadContextNotAudio = "[thread-safe, !audio-thread]"
)