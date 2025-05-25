package api

// #include <stdlib.h>
// #include <string.h>
// #include "../../include/clap/include/clap/ext/remote-controls.h"
//
// static inline const void* clap_host_get_extension_helper(const clap_host_t* host, const char* id) {
//     if (host && host->get_extension) {
//         return host->get_extension(host, id);
//     }
//     return NULL;
// }
//
// static inline void clap_host_remote_controls_changed(const clap_host_remote_controls_t* ext, const clap_host_t* host) {
//     if (ext && ext->changed) {
//         ext->changed(host);
//     }
// }
//
// static inline void clap_host_remote_controls_suggest_page(const clap_host_remote_controls_t* ext, const clap_host_t* host, clap_id page_id) {
//     if (ext && ext->suggest_page) {
//         ext->suggest_page(host, page_id);
//     }
// }
import "C"
import (
	"unsafe"
)

// Remote controls extension constants
const (
	ExtRemoteControlsID     = "clap.remote-controls/2"
	ExtRemoteControlsCompat = "clap.remote-controls.draft/2"
	RemoteControlsCount     = 8 // Number of controls per page
)

// RemoteControlsPage represents a page of remote control mappings
type RemoteControlsPage struct {
	SectionName string
	PageID      uint64
	PageName    string
	ParamIDs    [RemoteControlsCount]uint32
	IsForPreset bool // True if this page is specific to this preset
}

// RemoteControlsProvider is the interface that plugins implement to provide remote controls
type RemoteControlsProvider interface {
	// GetRemoteControlsPageCount returns the number of available pages
	GetRemoteControlsPageCount() uint32
	
	// GetRemoteControlsPage returns the page at the given index
	GetRemoteControlsPage(pageIndex uint32) (*RemoteControlsPage, bool)
}

// HostRemoteControls provides access to host remote controls functionality
type HostRemoteControls struct {
	host      unsafe.Pointer
	remoteExt unsafe.Pointer
}

// NewHostRemoteControls creates a new host remote controls interface
func NewHostRemoteControls(host unsafe.Pointer) *HostRemoteControls {
	if host == nil {
		return nil
	}
	
	cHost := (*C.clap_host_t)(host)
	
	// Try to get remote controls extension
	extPtr := C.clap_host_get_extension_helper(cHost, C.CString(ExtRemoteControlsID))
	if extPtr == nil {
		// Try compat version
		extPtr = C.clap_host_get_extension_helper(cHost, C.CString(ExtRemoteControlsCompat))
	}
	
	if extPtr == nil {
		return nil
	}
	
	return &HostRemoteControls{
		host:      host,
		remoteExt: extPtr,
	}
}

// NotifyChanged notifies the host that the remote controls have changed
func (h *HostRemoteControls) NotifyChanged() {
	if h.remoteExt == nil {
		return
	}
	
	ext := (*C.clap_host_remote_controls_t)(h.remoteExt)
	C.clap_host_remote_controls_changed(ext, (*C.clap_host_t)(h.host))
}

// SuggestPage suggests a page to the host based on what's being edited in the GUI
func (h *HostRemoteControls) SuggestPage(pageID uint64) {
	if h.remoteExt == nil {
		return
	}
	
	ext := (*C.clap_host_remote_controls_t)(h.remoteExt)
	C.clap_host_remote_controls_suggest_page(ext, (*C.clap_host_t)(h.host), C.clap_id(pageID))
}

// RemoteControlsPageToC converts a Go RemoteControlsPage to a C structure
func RemoteControlsPageToC(page *RemoteControlsPage, cPagePtr unsafe.Pointer) {
	cPage := (*C.clap_remote_controls_page_t)(cPagePtr)
	// Copy section name
	sectionBytes := []byte(page.SectionName)
	if len(sectionBytes) >= C.CLAP_NAME_SIZE {
		sectionBytes = sectionBytes[:C.CLAP_NAME_SIZE-1]
	}
	for i, b := range sectionBytes {
		cPage.section_name[i] = C.char(b)
	}
	cPage.section_name[len(sectionBytes)] = 0
	
	// Copy page name
	pageBytes := []byte(page.PageName)
	if len(pageBytes) >= C.CLAP_NAME_SIZE {
		pageBytes = pageBytes[:C.CLAP_NAME_SIZE-1]
	}
	for i, b := range pageBytes {
		cPage.page_name[i] = C.char(b)
	}
	cPage.page_name[len(pageBytes)] = 0
	
	// Copy page ID
	cPage.page_id = C.clap_id(page.PageID)
	
	// Copy parameter IDs
	for i := 0; i < RemoteControlsCount; i++ {
		cPage.param_ids[i] = C.clap_id(page.ParamIDs[i])
	}
	
	// Copy preset flag
	cPage.is_for_preset = C.bool(page.IsForPreset)
}