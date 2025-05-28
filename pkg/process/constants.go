package process

// Process status codes that plugins return from their Process method
const (
	// ProcessError indicates a processing error occurred.
	// The plugin is expected to have silenced its audio outputs.
	// The host may deactivate and reactivate the plugin.
	ProcessError = 0

	// ProcessContinue indicates normal processing completed successfully.
	// The plugin should be called again for the next audio buffer.
	ProcessContinue = 1

	// ProcessContinueIfNotQuiet indicates processing completed but the plugin
	// may enter sleep mode if the host is not providing audio input.
	// This is useful for effects that produce output only when input is present.
	ProcessContinueIfNotQuiet = 2

	// ProcessTail indicates the plugin is in tail mode, producing a tail
	// (like reverb decay) but no longer needs regular processing.
	// The host should continue calling process until the plugin returns
	// ProcessSleep or ProcessContinue.
	ProcessTail = 3

	// ProcessSleep indicates the plugin has finished processing and can
	// enter sleep mode. The host may stop calling process until new
	// events arrive or parameters change.
	ProcessSleep = 4
)

// IsValidProcessStatus checks if a status code is valid
func IsValidProcessStatus(status int32) bool {
	return status >= ProcessError && status <= ProcessSleep
}

// ProcessStatusString returns a human-readable string for a process status
func ProcessStatusString(status int32) string {
	switch status {
	case ProcessError:
		return "ERROR"
	case ProcessContinue:
		return "CONTINUE"
	case ProcessContinueIfNotQuiet:
		return "CONTINUE_IF_NOT_QUIET"
	case ProcessTail:
		return "TAIL"
	case ProcessSleep:
		return "SLEEP"
	default:
		return "UNKNOWN"
	}
}

// Common process result helpers

// ProcessResult represents the result of a process call
type ProcessResult struct {
	Status int32
	Error  error // Go error for additional context (not part of CLAP spec)
}

// NewProcessResult creates a new process result
func NewProcessResult(status int32) ProcessResult {
	return ProcessResult{Status: status}
}

// NewProcessError creates a process result with error status and context
func NewProcessError(err error) ProcessResult {
	return ProcessResult{
		Status: ProcessError,
		Error:  err,
	}
}

// IsError returns true if the result indicates an error
func (r ProcessResult) IsError() bool {
	return r.Status == ProcessError
}

// ShouldContinue returns true if the host should continue processing
func (r ProcessResult) ShouldContinue() bool {
	return r.Status == ProcessContinue || r.Status == ProcessContinueIfNotQuiet
}

// ShouldSleep returns true if the plugin wants to sleep
func (r ProcessResult) ShouldSleep() bool {
	return r.Status == ProcessSleep
}

// IsTail returns true if the plugin is in tail mode
func (r ProcessResult) IsTail() bool {
	return r.Status == ProcessTail
}

// String returns a string representation of the result
func (r ProcessResult) String() string {
	if r.Error != nil {
		return ProcessStatusString(r.Status) + ": " + r.Error.Error()
	}
	return ProcessStatusString(r.Status)
}