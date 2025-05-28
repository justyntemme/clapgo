package extension

import "unsafe"

// ThreadCheckProvider is an extension for plugins that want thread checking.
// It allows plugins to verify they're running on the correct thread.
type ThreadCheckProvider interface {
	// IsMainThread returns true if currently on the main thread.
	IsMainThread() bool

	// IsAudioThread returns true if currently on the audio thread.
	IsAudioThread() bool
}

// ThreadPoolProvider is an extension for plugins that want to use thread pools.
// It allows plugins to request work be done on background threads.
type ThreadPoolProvider interface {
	// RequestExec requests execution of a task on a background thread.
	// Returns true if the request was accepted.
	RequestExec(taskID uint32) bool
}

// PosixFDSupportProvider is an extension for plugins that need POSIX file descriptor support.
// It allows plugins to register file descriptors for monitoring.
type PosixFDSupportProvider interface {
	// RegisterFD registers a file descriptor for monitoring.
	RegisterFD(fd int32, flags uint32) bool

	// UnregisterFD unregisters a file descriptor.
	UnregisterFD(fd int32) bool

	// ModifyFD modifies the flags for a file descriptor.
	ModifyFD(fd int32, flags uint32) bool
}

// ThreadPoolTask represents a task for thread pool execution
type ThreadPoolTask struct {
	ID       uint32
	Callback func()
}

// POSIX FD flags
const (
	PosixFDRead  = 1 << 0
	PosixFDWrite = 1 << 1
	PosixFDError = 1 << 2
)

// Thread pool task IDs - plugins can define their own starting from 1000
const (
	TaskIDUserStart = 1000
)