package goclap

// #include <stdlib.h>
// #include <stdio.h>
// #include <string.h>
// #include "../../include/clap/include/clap/clap.h"
// 
// // Host callback wrapper functions
// static bool host_request_restart(const clap_host_t *host) {
//     if (!host || !host->request_restart) {
//         return false;
//     }
//     host->request_restart(host);
//     return true;
// }
// 
// static bool host_request_process(const clap_host_t *host) {
//     if (!host || !host->request_process) {
//         return false;
//     }
//     host->request_process(host);
//     return true;
// }
// 
// static bool host_request_callback(const clap_host_t *host) {
//     if (!host || !host->request_callback) {
//         return false;
//     }
//     host->request_callback(host);
//     return true;
// }
// 
// static const void* host_get_extension(const clap_host_t *host, const char *extension_id) {
//     if (!host || !host->get_extension || !extension_id) {
//         return NULL;
//     }
//     return host->get_extension(host, extension_id);
// }
import "C"
import (
	"fmt"
	"unsafe"
)

// Host represents a CLAP host
type Host struct {
	ptr unsafe.Pointer // Pointer to clap_host
}

// NewHost creates a Host object from a C host pointer
func NewHost(hostPtr unsafe.Pointer) *Host {
	if hostPtr == nil {
		return nil
	}
	
	return &Host{
		ptr: hostPtr,
	}
}

// GetName retrieves the host name
func (h *Host) GetName() string {
	if h.ptr == nil {
		return ""
	}
	
	host := (*C.clap_host_t)(h.ptr)
	return C.GoString(host.name)
}

// GetVendor retrieves the host vendor
func (h *Host) GetVendor() string {
	if h.ptr == nil {
		return ""
	}
	
	host := (*C.clap_host_t)(h.ptr)
	return C.GoString(host.vendor)
}

// GetURL retrieves the host URL
func (h *Host) GetURL() string {
	if h.ptr == nil {
		return ""
	}
	
	host := (*C.clap_host_t)(h.ptr)
	return C.GoString(host.url)
}

// GetVersion retrieves the host version
func (h *Host) GetVersion() string {
	if h.ptr == nil {
		return ""
	}
	
	host := (*C.clap_host_t)(h.ptr)
	return C.GoString(host.version)
}

// RequestRestart requests a plugin restart from the host
func (h *Host) RequestRestart() bool {
	if h.ptr == nil {
		fmt.Println("Warning: RequestRestart called with nil host pointer")
		return false
	}
	
	host := (*C.clap_host_t)(h.ptr)
	return bool(C.host_request_restart(host))
}

// RequestProcess requests the host to start processing
func (h *Host) RequestProcess() bool {
	if h.ptr == nil {
		fmt.Println("Warning: RequestProcess called with nil host pointer")
		return false
	}
	
	host := (*C.clap_host_t)(h.ptr)
	return bool(C.host_request_process(host))
}

// RequestCallback requests a callback on the main thread
func (h *Host) RequestCallback() bool {
	if h.ptr == nil {
		fmt.Println("Warning: RequestCallback called with nil host pointer")
		return false
	}
	
	host := (*C.clap_host_t)(h.ptr)
	return bool(C.host_request_callback(host))
}

// GetExtension retrieves a host extension
func (h *Host) GetExtension(id string) unsafe.Pointer {
	if h.ptr == nil {
		fmt.Println("Warning: GetExtension called with nil host pointer")
		return nil
	}
	
	if id == "" {
		fmt.Println("Warning: GetExtension called with empty extension ID")
		return nil
	}
	
	host := (*C.clap_host_t)(h.ptr)
	cID := C.CString(id)
	defer C.free(unsafe.Pointer(cID))
	
	return unsafe.Pointer(C.host_get_extension(host, cID))
}