package extension

import "unsafe"

// ContextMenuProvider is an extension for plugins that provide context menus.
// It allows plugins to add items to the host's context menu.
type ContextMenuProvider interface {
	// PopulateContextMenu populates the context menu with plugin-specific items.
	// target specifies what the menu is for (e.g., parameter, plugin).
	// Returns true if items were added to the menu.
	PopulateContextMenu(target *ContextMenuTarget, builder unsafe.Pointer) bool

	// PerformContextMenuAction performs the action for the given action ID.
	// Returns true if the action was handled.
	PerformContextMenuAction(actionID uint32) bool
}

// ContextMenuTarget describes what the context menu is for
type ContextMenuTarget struct {
	Kind     uint32
	ID       uint32
	Position [2]float64 // x, y coordinates if relevant
}

// Context menu target kinds
const (
	ContextMenuTargetPlugin    = 0
	ContextMenuTargetParameter = 1
	ContextMenuTargetNote      = 2
)

// ContextMenuBuilder helps build context menus
type ContextMenuBuilder struct {
	builder unsafe.Pointer
}

// NewContextMenuBuilder wraps a C context menu builder
func NewContextMenuBuilder(builder unsafe.Pointer) *ContextMenuBuilder {
	return &ContextMenuBuilder{builder: builder}
}

// ContextMenuItem represents a context menu item
type ContextMenuItem struct {
	ID       uint32
	Label    string
	IsActive bool
	IsChecked bool
}

// ContextMenuSeparator represents a menu separator
type ContextMenuSeparator struct{}

// ContextMenuSubmenu represents a submenu
type ContextMenuSubmenu struct {
	Label string
	Items []interface{} // Can contain ContextMenuItem, ContextMenuSeparator, or ContextMenuSubmenu
}