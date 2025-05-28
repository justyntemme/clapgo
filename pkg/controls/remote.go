package controls

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

// Helper functions for building remote control pages

// RemoteControlsPageBuilder helps build remote control pages
type RemoteControlsPageBuilder struct {
	page RemoteControlsPage
}

// NewRemoteControlsPageBuilder creates a new page builder
func NewRemoteControlsPageBuilder() *RemoteControlsPageBuilder {
	return &RemoteControlsPageBuilder{
		page: RemoteControlsPage{
			ParamIDs: [RemoteControlsCount]uint32{},
		},
	}
}

// SetSection sets the section name for the page
func (b *RemoteControlsPageBuilder) SetSection(section string) *RemoteControlsPageBuilder {
	b.page.SectionName = section
	return b
}

// SetPageName sets the page name
func (b *RemoteControlsPageBuilder) SetPageName(name string) *RemoteControlsPageBuilder {
	b.page.PageName = name
	return b
}

// SetPageID sets the page ID
func (b *RemoteControlsPageBuilder) SetPageID(id uint64) *RemoteControlsPageBuilder {
	b.page.PageID = id
	return b
}

// SetForPreset marks the page as preset-specific
func (b *RemoteControlsPageBuilder) SetForPreset(forPreset bool) *RemoteControlsPageBuilder {
	b.page.IsForPreset = forPreset
	return b
}

// SetParamID sets a parameter ID at the given control index (0-7)
func (b *RemoteControlsPageBuilder) SetParamID(controlIndex int, paramID uint32) *RemoteControlsPageBuilder {
	if controlIndex >= 0 && controlIndex < RemoteControlsCount {
		b.page.ParamIDs[controlIndex] = paramID
	}
	return b
}

// SetParamIDs sets all parameter IDs at once
func (b *RemoteControlsPageBuilder) SetParamIDs(paramIDs []uint32) *RemoteControlsPageBuilder {
	for i := 0; i < RemoteControlsCount && i < len(paramIDs); i++ {
		b.page.ParamIDs[i] = paramIDs[i]
	}
	return b
}

// Build returns the constructed page
func (b *RemoteControlsPageBuilder) Build() RemoteControlsPage {
	return b.page
}

// RemoteControlsManager manages multiple pages of remote controls
type RemoteControlsManager struct {
	pages []RemoteControlsPage
	host  *HostRemoteControls
}

// NewRemoteControlsManager creates a new remote controls manager
func NewRemoteControlsManager(host unsafe.Pointer) *RemoteControlsManager {
	return &RemoteControlsManager{
		pages: make([]RemoteControlsPage, 0),
		host:  NewHostRemoteControls(host),
	}
}

// AddPage adds a page to the manager
func (m *RemoteControlsManager) AddPage(page RemoteControlsPage) {
	m.pages = append(m.pages, page)
}

// GetPageCount returns the number of pages
func (m *RemoteControlsManager) GetPageCount() uint32 {
	return uint32(len(m.pages))
}

// GetPage returns a page by index
func (m *RemoteControlsManager) GetPage(index uint32) (*RemoteControlsPage, bool) {
	if index >= uint32(len(m.pages)) {
		return nil, false
	}
	return &m.pages[index], true
}

// NotifyChanged notifies the host that pages have changed
func (m *RemoteControlsManager) NotifyChanged() {
	if m.host != nil {
		m.host.NotifyChanged()
	}
}

// SuggestPage suggests a page to the host
func (m *RemoteControlsManager) SuggestPage(pageID uint64) {
	if m.host != nil {
		m.host.SuggestPage(pageID)
	}
}

// Clear removes all pages
func (m *RemoteControlsManager) Clear() {
	m.pages = m.pages[:0]
}

// CreatePresetPages creates remote control pages for common preset parameters
func CreatePresetPages() []RemoteControlsPage {
	// Page 1: Main controls
	mainPage := NewRemoteControlsPageBuilder().
		SetSection("Main").
		SetPageName("Essential").
		SetPageID(1).
		SetForPreset(false).
		Build()

	// Page 2: EQ controls  
	eqPage := NewRemoteControlsPageBuilder().
		SetSection("EQ").
		SetPageName("Equalizer").
		SetPageID(2).
		SetForPreset(false).
		Build()

	// Page 3: Effects controls
	fxPage := NewRemoteControlsPageBuilder().
		SetSection("Effects").
		SetPageName("FX Send").
		SetPageID(3).
		SetForPreset(false).
		Build()

	return []RemoteControlsPage{mainPage, eqPage, fxPage}
}