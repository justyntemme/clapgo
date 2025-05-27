package thread

// #include "../../include/clap/include/clap/clap.h"
// #include <stdlib.h>
//
// // Helper to get thread check extension
// static const clap_host_thread_check_t* get_thread_check_ext(const clap_host_t* host) {
//     if (host && host->get_extension) {
//         return (const clap_host_thread_check_t*)host->get_extension(host, CLAP_EXT_THREAD_CHECK);
//     }
//     return NULL;
// }
//
// // Thread check helper
// static int clap_thread_check_is_main_thread(const clap_host_t* host) {
//     const clap_host_thread_check_t* thread_check = get_thread_check_ext(host);
//     if (thread_check && thread_check->is_main_thread) {
//         return thread_check->is_main_thread(host) ? 1 : 0;
//     }
//     return -1; // Unknown
// }
//
// static int clap_thread_check_is_audio_thread(const clap_host_t* host) {
//     const clap_host_thread_check_t* thread_check = get_thread_check_ext(host);
//     if (thread_check && thread_check->is_audio_thread) {
//         return thread_check->is_audio_thread(host) ? 1 : 0;
//     }
//     return -1; // Unknown
// }
import "C"
import (
	"fmt"
	"unsafe"
)

// Checker provides thread checking functionality from the host
type Checker struct {
	host      unsafe.Pointer
	available bool
}

// NewChecker creates a new thread checker
func NewChecker(host unsafe.Pointer) *Checker {
	if host == nil {
		return &Checker{host: nil, available: false}
	}
	
	// Check if the extension is available
	hostPtr := (*C.clap_host_t)(host)
	ext := C.get_thread_check_ext(hostPtr)
	if ext != nil {
		return &Checker{host: host, available: true}
	}
	
	return &Checker{host: host, available: false}
}

// IsAvailable returns true if thread checking is available
func (tc *Checker) IsAvailable() bool {
	return tc.available
}

// IsMainThread returns true if called from the main thread
func (tc *Checker) IsMainThread() bool {
	if !tc.available || tc.host == nil {
		return false
	}
	
	result := C.clap_thread_check_is_main_thread((*C.clap_host_t)(tc.host))
	return result == 1
}

// IsAudioThread returns true if called from the audio thread
func (tc *Checker) IsAudioThread() bool {
	if !tc.available || tc.host == nil {
		return false
	}
	
	result := C.clap_thread_check_is_audio_thread((*C.clap_host_t)(tc.host))
	return result == 1
}

// AssertMainThread panics if not called from the main thread
func (tc *Checker) AssertMainThread(function string) {
	if tc.available && !tc.IsMainThread() {
		panic(fmt.Sprintf("%s must be called from main thread", function))
	}
}

// AssertAudioThread panics if not called from the audio thread
func (tc *Checker) AssertAudioThread(function string) {
	if tc.available && !tc.IsAudioThread() {
		panic(fmt.Sprintf("%s must be called from audio thread", function))
	}
}