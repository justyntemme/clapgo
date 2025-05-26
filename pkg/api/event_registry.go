package api

/*
#include "../../include/clap/include/clap/ext/event-registry.h"
#include <stdlib.h>

static inline const void* clap_host_get_extension_helper(const clap_host_t* host, const char* id) {
    if (host && host->get_extension) {
        return host->get_extension(host, id);
    }
    return NULL;
}

static inline bool clap_host_event_registry_query(const clap_host_event_registry_t* ext, const clap_host_t* host, const char* space_name, uint16_t* space_id) {
    if (ext && ext->query) {
        return ext->query(host, space_name, space_id);
    }
    return false;
}
*/
import "C"
import (
	"unsafe"
)

// EventRegistryHost provides access to the host's event registry extension.
// This is a host-side extension that plugins can use to query event space IDs.
type EventRegistryHost struct {
	host      unsafe.Pointer
	extension *C.clap_host_event_registry_t
}

// NewEventRegistryHost creates a new event registry host interface.
// Returns nil if the host doesn't support the event registry extension.
func NewEventRegistryHost(host unsafe.Pointer) *EventRegistryHost {
	if host == nil {
		return nil
	}

	cHost := (*C.clap_host_t)(host)
	
	extID := C.CString("clap.event-registry")
	defer C.free(unsafe.Pointer(extID))
	
	ext := C.clap_host_get_extension_helper(cHost, extID)
	if ext == nil {
		return nil
	}

	return &EventRegistryHost{
		host:      host,
		extension: (*C.clap_host_event_registry_t)(ext),
	}
}

// Query queries an event space ID by name.
// Returns the space ID if found, or 0xFFFF if the space name is unknown to the host.
// The space ID 0 is reserved for CLAP's core events.
func (e *EventRegistryHost) Query(spaceName string) (uint16, bool) {
	if e.extension == nil || e.extension.query == nil {
		return 0xFFFF, false
	}

	cSpaceName := C.CString(spaceName)
	defer C.free(unsafe.Pointer(cSpaceName))

	var spaceID C.uint16_t
	cHost := (*C.clap_host_t)(e.host)
	
	result := C.clap_host_event_registry_query(e.extension, cHost, cSpaceName, &spaceID)
	return uint16(spaceID), bool(result)
}

// Common event space names
const (
	// CoreEventSpace is the reserved space for CLAP's core events
	CoreEventSpace = 0
	
	// Common event space names that plugins might query
	EventSpaceVST3      = "vst3"
	EventSpaceMIDI2     = "midi2" 
	EventSpaceOSC       = "osc"
	EventSpaceCustom    = "custom"
)

// EventSpaceHelper provides convenient methods for working with event spaces
type EventSpaceHelper struct {
	registry *EventRegistryHost
	cache    map[string]uint16
}

// NewEventSpaceHelper creates a new event space helper
func NewEventSpaceHelper(host unsafe.Pointer) *EventSpaceHelper {
	return &EventSpaceHelper{
		registry: NewEventRegistryHost(host),
		cache:    make(map[string]uint16),
	}
}

// GetSpaceID gets the space ID for a given space name, using a cache for efficiency
func (h *EventSpaceHelper) GetSpaceID(spaceName string) uint16 {
	if h.registry == nil {
		return 0xFFFF
	}

	// Check cache first
	if id, ok := h.cache[spaceName]; ok {
		return id
	}

	// Query the host
	id, ok := h.registry.Query(spaceName)
	if ok {
		h.cache[spaceName] = id
	}
	
	return id
}

// IsSpaceSupported checks if a space is supported by the host
func (h *EventSpaceHelper) IsSpaceSupported(spaceName string) bool {
	return h.GetSpaceID(spaceName) != 0xFFFF
}