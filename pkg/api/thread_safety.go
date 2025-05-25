package api

// Thread Safety Documentation for ClapGo
//
// This file documents the thread safety requirements for all ClapGo API functions
// following the CLAP specification.
//
// CLAP defines two symbolic threads:
//
// 1. Main Thread [main-thread]:
//    - GUI operations
//    - Plugin lifecycle (init, destroy, activate, deactivate)
//    - Extension discovery
//    - State save/load
//    - Parameter info queries
//
// 2. Audio Thread [audio-thread]:
//    - Process() function
//    - Real-time parameter updates
//    - Event processing
//    - No blocking operations allowed
//
// Thread-Safe Functions [thread-safe]:
//    - Can be called from any thread
//    - Must use proper synchronization
//
// Thread-Safe but NOT Audio [thread-safe, !audio-thread]:
//    - Can be called from any thread EXCEPT audio
//    - May perform blocking operations

// Plugin Interface Thread Safety:
//
// Init()              [main-thread]
// Destroy()           [main-thread]
// Activate()          [main-thread]
// Deactivate()        [main-thread]
// StartProcessing()   [audio-thread]
// StopProcessing()    [audio-thread]
// Reset()             [main-thread]
// Process()           [audio-thread]
// GetExtension()      [main-thread]

// ParameterManager Thread Safety:
//
// RegisterParameter()      [main-thread]
// GetParameterInfo()       [thread-safe]
// GetParameterValue()      [thread-safe]
// SetParameterValue()      [thread-safe]
// GetParameterCount()      [thread-safe]
// AddChangeListener()      [main-thread]
// RemoveChangeListener()   [main-thread]

// EventProcessor Thread Safety:
//
// ProcessAllEvents()       [audio-thread]
// ProcessTypedEvents()     [audio-thread]
// PushBackEvent()          [audio-thread]
// GetEvent/ReturnEvent()   [audio-thread]

// State Operations Thread Safety:
//
// SaveState()              [main-thread]
// LoadState()              [main-thread]
// SaveStateWithContext()   [main-thread]
// LoadStateWithContext()   [main-thread]

// Host Communication Thread Safety:
//
// HostLogger.Log()         [thread-safe]
// RequestRestart()         [thread-safe]
// RequestProcess()         [thread-safe]
// RequestCallback()        [thread-safe, !audio-thread]

// ValidateThreadSafety performs compile-time thread safety validation
func ValidateThreadSafety() {
	// This function exists for documentation purposes
	// Actual thread validation happens at runtime with ThreadChecker
}