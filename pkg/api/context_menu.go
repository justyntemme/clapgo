package api

// #include <stdlib.h>
// #include "../../include/clap/include/clap/ext/context-menu.h"
//
// static inline bool context_menu_builder_add_item(const clap_context_menu_builder_t* builder,
//                                                  clap_context_menu_item_kind_t item_kind,
//                                                  const void* item_data) {
//     if (builder && builder->add_item) {
//         return builder->add_item(builder, item_kind, item_data);
//     }
//     return false;
// }
//
// static inline bool context_menu_builder_supports(const clap_context_menu_builder_t* builder,
//                                                  clap_context_menu_item_kind_t item_kind) {
//     if (builder && builder->supports) {
//         return builder->supports(builder, item_kind);
//     }
//     return false;
// }
//
// static inline const void* clap_host_get_extension_helper(const clap_host_t* host, const char* id) {
//     if (host && host->get_extension) {
//         return host->get_extension(host, id);
//     }
//     return NULL;
// }
//
// static inline bool clap_host_context_menu_can_popup(const clap_host_context_menu_t* ext, const clap_host_t* host) {
//     if (ext && ext->can_popup) {
//         return ext->can_popup(host);
//     }
//     return false;
// }
//
// static inline bool clap_host_context_menu_popup(const clap_host_context_menu_t* ext, const clap_host_t* host,
//                                                 const clap_context_menu_target_t* target, int32_t screen_index,
//                                                 int32_t x, int32_t y) {
//     if (ext && ext->popup) {
//         return ext->popup(host, target, screen_index, x, y);
//     }
//     return false;
// }
import "C"
import (
	"unsafe"
)

// Context menu extension IDs
const (
	ExtContextMenuID       = "clap.context-menu/1"
	ExtContextMenuCompatID = "clap.context-menu.draft/0"
)

// Context menu target kinds
const (
	ContextMenuTargetKindGlobal = 0
	ContextMenuTargetKindParam  = 1
)

// Context menu item kinds
const (
	ContextMenuItemEntry         = 0
	ContextMenuItemCheckEntry    = 1
	ContextMenuItemSeparator     = 2
	ContextMenuItemBeginSubmenu  = 3
	ContextMenuItemEndSubmenu    = 4
	ContextMenuItemTitle         = 5
)

// ContextMenuTarget represents the target of a context menu request
type ContextMenuTarget struct {
	Kind uint32
	ID   uint64
}

// ContextMenuItem represents a menu item that can be added to a context menu
type ContextMenuItem interface {
	// GetKind returns the item kind (entry, separator, submenu, etc.)
	GetKind() uint32
}

// ContextMenuEntry represents a clickable menu entry
type ContextMenuEntry struct {
	Label     string
	IsEnabled bool
	ActionID  uint64
}

func (e *ContextMenuEntry) GetKind() uint32 { return ContextMenuItemEntry }

// ContextMenuCheckEntry represents a menu entry with a checkmark
type ContextMenuCheckEntry struct {
	Label     string
	IsEnabled bool
	IsChecked bool
	ActionID  uint64
}

func (e *ContextMenuCheckEntry) GetKind() uint32 { return ContextMenuItemCheckEntry }

// ContextMenuSeparator represents a separator line
type ContextMenuSeparator struct{}

func (s *ContextMenuSeparator) GetKind() uint32 { return ContextMenuItemSeparator }

// ContextMenuSubmenu represents the start of a submenu
type ContextMenuSubmenu struct {
	Label     string
	IsEnabled bool
}

func (s *ContextMenuSubmenu) GetKind() uint32 { return ContextMenuItemBeginSubmenu }

// ContextMenuEndSubmenu represents the end of a submenu
type ContextMenuEndSubmenu struct{}

func (e *ContextMenuEndSubmenu) GetKind() uint32 { return ContextMenuItemEndSubmenu }

// ContextMenuTitle represents a title entry
type ContextMenuTitle struct {
	Title     string
	IsEnabled bool
}

func (t *ContextMenuTitle) GetKind() uint32 { return ContextMenuItemTitle }

// ContextMenuBuilder wraps the C context menu builder
type ContextMenuBuilder struct {
	builder unsafe.Pointer
}

// NewContextMenuBuilder creates a new context menu builder wrapper
func NewContextMenuBuilder(builder unsafe.Pointer) *ContextMenuBuilder {
	return &ContextMenuBuilder{builder: builder}
}

// AddItem adds an item to the menu
func (b *ContextMenuBuilder) AddItem(item ContextMenuItem) bool {
	if b.builder == nil {
		return false
	}

	cBuilder := (*C.clap_context_menu_builder_t)(b.builder)
	
	switch v := item.(type) {
	case *ContextMenuEntry:
		cLabel := C.CString(v.Label)
		defer C.free(unsafe.Pointer(cLabel))
		
		entry := C.clap_context_menu_entry_t{
			label:      cLabel,
			is_enabled: C.bool(v.IsEnabled),
			action_id:  C.clap_id(v.ActionID),
		}
		return bool(C.context_menu_builder_add_item(cBuilder, C.clap_context_menu_item_kind_t(ContextMenuItemEntry), unsafe.Pointer(&entry)))
		
	case *ContextMenuCheckEntry:
		cLabel := C.CString(v.Label)
		defer C.free(unsafe.Pointer(cLabel))
		
		entry := C.clap_context_menu_check_entry_t{
			label:      cLabel,
			is_enabled: C.bool(v.IsEnabled),
			is_checked: C.bool(v.IsChecked),
			action_id:  C.clap_id(v.ActionID),
		}
		return bool(C.context_menu_builder_add_item(cBuilder, C.clap_context_menu_item_kind_t(ContextMenuItemCheckEntry), unsafe.Pointer(&entry)))
		
	case *ContextMenuSeparator:
		return bool(C.context_menu_builder_add_item(cBuilder, C.clap_context_menu_item_kind_t(ContextMenuItemSeparator), nil))
		
	case *ContextMenuSubmenu:
		cLabel := C.CString(v.Label)
		defer C.free(unsafe.Pointer(cLabel))
		
		submenu := C.clap_context_menu_submenu_t{
			label:      cLabel,
			is_enabled: C.bool(v.IsEnabled),
		}
		return bool(C.context_menu_builder_add_item(cBuilder, C.clap_context_menu_item_kind_t(ContextMenuItemBeginSubmenu), unsafe.Pointer(&submenu)))
		
	case *ContextMenuEndSubmenu:
		return bool(C.context_menu_builder_add_item(cBuilder, C.clap_context_menu_item_kind_t(ContextMenuItemEndSubmenu), nil))
		
	case *ContextMenuTitle:
		cTitle := C.CString(v.Title)
		defer C.free(unsafe.Pointer(cTitle))
		
		title := C.clap_context_menu_item_title_t{
			title:      cTitle,
			is_enabled: C.bool(v.IsEnabled),
		}
		return bool(C.context_menu_builder_add_item(cBuilder, C.clap_context_menu_item_kind_t(ContextMenuItemTitle), unsafe.Pointer(&title)))
	}
	
	return false
}

// Supports checks if the builder supports a specific item kind
func (b *ContextMenuBuilder) Supports(itemKind uint32) bool {
	if b.builder == nil {
		return false
	}
	
	cBuilder := (*C.clap_context_menu_builder_t)(b.builder)
	return bool(C.context_menu_builder_supports(cBuilder, C.clap_context_menu_item_kind_t(itemKind)))
}

// ContextMenuProvider is the interface that plugins implement to provide context menu functionality
type ContextMenuProvider interface {
	// Populate is called to build the context menu
	// target contains the target information (nil for global context)
	// builder is used to add menu items
	// Returns true on success
	PopulateContextMenu(target *ContextMenuTarget, builder *ContextMenuBuilder) bool
	
	// Perform is called when a menu item is selected
	// target contains the target information (nil for global context)
	// actionID is the ID of the selected action
	// Returns true on success
	PerformContextMenuAction(target *ContextMenuTarget, actionID uint64) bool
}

// HostContextMenu provides access to host context menu functionality
type HostContextMenu struct {
	host       unsafe.Pointer
	contextExt unsafe.Pointer
}

// NewHostContextMenu creates a new host context menu interface
func NewHostContextMenu(host unsafe.Pointer) *HostContextMenu {
	if host == nil {
		return nil
	}
	
	cHost := (*C.clap_host_t)(host)
	extPtr := C.clap_host_get_extension_helper(cHost, C.CString(ExtContextMenuID))
	if extPtr == nil {
		// Try compat version
		extPtr = C.clap_host_get_extension_helper(cHost, C.CString(ExtContextMenuCompatID))
	}
	
	if extPtr == nil {
		return nil
	}
	
	return &HostContextMenu{
		host:       host,
		contextExt: extPtr,
	}
}

// CanPopup returns true if the host can display a popup menu
func (h *HostContextMenu) CanPopup() bool {
	if h.contextExt == nil {
		return false
	}
	
	ext := (*C.clap_host_context_menu_t)(h.contextExt)
	return bool(C.clap_host_context_menu_can_popup(ext, (*C.clap_host_t)(h.host)))
}

// Popup shows the host popup menu at the specified location
func (h *HostContextMenu) Popup(target *ContextMenuTarget, screenIndex, x, y int32) bool {
	if h.contextExt == nil {
		return false
	}
	
	ext := (*C.clap_host_context_menu_t)(h.contextExt)
	
	var cTarget *C.clap_context_menu_target_t
	if target != nil {
		t := C.clap_context_menu_target_t{
			kind: C.uint32_t(target.Kind),
			id:   C.clap_id(target.ID),
		}
		cTarget = &t
	}
	
	return bool(C.clap_host_context_menu_popup(ext, (*C.clap_host_t)(h.host), cTarget, C.int32_t(screenIndex), C.int32_t(x), C.int32_t(y)))
}