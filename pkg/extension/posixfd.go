package extension

/*
#include <stdlib.h>
#include "../../include/clap/include/clap/ext/posix-fd-support.h"

static inline const void* clap_host_get_extension_helper(const clap_host_t* host, const char* id) {
    if (host && host->get_extension) {
        return host->get_extension(host, id);
    }
    return NULL;
}

static inline bool clap_host_posix_fd_support_register_fd(const clap_host_posix_fd_support_t* ext, const clap_host_t* host, int fd, clap_posix_fd_flags_t flags) {
    if (ext && ext->register_fd) {
        return ext->register_fd(host, fd, flags);
    }
    return false;
}

static inline bool clap_host_posix_fd_support_modify_fd(const clap_host_posix_fd_support_t* ext, const clap_host_t* host, int fd, clap_posix_fd_flags_t flags) {
    if (ext && ext->modify_fd) {
        return ext->modify_fd(host, fd, flags);
    }
    return false;
}

static inline bool clap_host_posix_fd_support_unregister_fd(const clap_host_posix_fd_support_t* ext, const clap_host_t* host, int fd) {
    if (ext && ext->unregister_fd) {
        return ext->unregister_fd(host, fd);
    }
    return false;
}
*/
import "C"
import (
	"unsafe"
)

// POSIX FD flags for I/O events
const (
	PosixFDRead  = uint32(C.CLAP_POSIX_FD_READ)  // File descriptor is ready for reading
	PosixFDWrite = uint32(C.CLAP_POSIX_FD_WRITE) // File descriptor is ready for writing
	PosixFDError = uint32(C.CLAP_POSIX_FD_ERROR) // File descriptor has an error condition
)

// PosixFDSupportProvider is an extension for plugins that need to handle
// asynchronous I/O on the main thread using file descriptors.
type PosixFDSupportProvider interface {
	// OnFD is called when a registered file descriptor has events.
	// This callback is "level-triggered": a writable fd will continuously
	// produce OnFD events until you modify the fd to remove write notifications.
	// [main-thread]
	OnFD(fd int, flags uint32)
}

// PosixFDSupportHost provides access to the host's POSIX FD support extension.
// This allows plugins to register file descriptors with the host's event loop.
type PosixFDSupportHost struct {
	host      unsafe.Pointer
	extension *C.clap_host_posix_fd_support_t
}

// NewPosixFDSupportHost creates a new POSIX FD support host interface.
// Returns nil if the host doesn't support the POSIX FD support extension.
func NewPosixFDSupportHost(host unsafe.Pointer) *PosixFDSupportHost {
	if host == nil {
		return nil
	}

	cHost := (*C.clap_host_t)(host)
	
	cExtID := C.CString("clap.posix-fd-support")
	defer C.free(unsafe.Pointer(cExtID))
	
	ext := C.clap_host_get_extension_helper(cHost, cExtID)
	if ext == nil {
		return nil
	}

	return &PosixFDSupportHost{
		host:      host,
		extension: (*C.clap_host_posix_fd_support_t)(ext),
	}
}

// RegisterFD registers a file descriptor with the host's event loop.
// The flags parameter specifies which events to monitor (read, write, error).
// Returns true on success.
// [main-thread]
func (h *PosixFDSupportHost) RegisterFD(fd int, flags uint32) bool {
	if h.extension == nil || h.extension.register_fd == nil {
		return false
	}

	cHost := (*C.clap_host_t)(h.host)
	return bool(C.clap_host_posix_fd_support_register_fd(h.extension, cHost, C.int(fd), C.clap_posix_fd_flags_t(flags)))
}

// ModifyFD modifies the event mask for a registered file descriptor.
// Use this to change which events you're interested in, or to stop
// receiving events for a particular type (e.g., remove write notifications
// after writing is complete).
// Returns true on success.
// [main-thread]
func (h *PosixFDSupportHost) ModifyFD(fd int, flags uint32) bool {
	if h.extension == nil || h.extension.modify_fd == nil {
		return false
	}

	cHost := (*C.clap_host_t)(h.host)
	return bool(C.clap_host_posix_fd_support_modify_fd(h.extension, cHost, C.int(fd), C.clap_posix_fd_flags_t(flags)))
}

// UnregisterFD unregisters a file descriptor from the host's event loop.
// Returns true on success.
// [main-thread]
func (h *PosixFDSupportHost) UnregisterFD(fd int) bool {
	if h.extension == nil || h.extension.unregister_fd == nil {
		return false
	}

	cHost := (*C.clap_host_t)(h.host)
	return bool(C.clap_host_posix_fd_support_unregister_fd(h.extension, cHost, C.int(fd)))
}

// PosixFDManager provides a convenient interface for managing file descriptors
// with the host's event loop.
type PosixFDManager struct {
	host       *PosixFDSupportHost
	registered map[int]uint32 // fd -> flags mapping
}

// NewPosixFDManager creates a new POSIX FD manager
func NewPosixFDManager(host unsafe.Pointer) *PosixFDManager {
	return &PosixFDManager{
		host:       NewPosixFDSupportHost(host),
		registered: make(map[int]uint32),
	}
}

// IsSupported returns true if the host supports POSIX FD extension
func (m *PosixFDManager) IsSupported() bool {
	return m.host != nil
}

// RegisterReadFD registers a file descriptor for read events only
func (m *PosixFDManager) RegisterReadFD(fd int) bool {
	return m.RegisterFD(fd, PosixFDRead)
}

// RegisterWriteFD registers a file descriptor for write events only
func (m *PosixFDManager) RegisterWriteFD(fd int) bool {
	return m.RegisterFD(fd, PosixFDWrite)
}

// RegisterFD registers a file descriptor with specific flags
func (m *PosixFDManager) RegisterFD(fd int, flags uint32) bool {
	if m.host == nil {
		return false
	}

	if m.host.RegisterFD(fd, flags) {
		m.registered[fd] = flags
		return true
	}
	return false
}

// DisableWrite removes write notifications for a file descriptor
// while keeping other notifications active
func (m *PosixFDManager) DisableWrite(fd int) bool {
	if m.host == nil {
		return false
	}

	flags, ok := m.registered[fd]
	if !ok {
		return false
	}

	// Remove write flag
	newFlags := flags &^ PosixFDWrite
	if m.host.ModifyFD(fd, newFlags) {
		m.registered[fd] = newFlags
		return true
	}
	return false
}

// UnregisterFD unregisters a file descriptor
func (m *PosixFDManager) UnregisterFD(fd int) bool {
	if m.host == nil {
		return false
	}

	if m.host.UnregisterFD(fd) {
		delete(m.registered, fd)
		return true
	}
	return false
}

// UnregisterAll unregisters all file descriptors
func (m *PosixFDManager) UnregisterAll() {
	for fd := range m.registered {
		m.UnregisterFD(fd)
	}
}