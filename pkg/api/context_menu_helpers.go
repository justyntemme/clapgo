package api

import (
	"fmt"
	"unsafe"
	
	"github.com/justyntemme/clapgo/pkg/host"
)

// DefaultContextMenuProvider provides common context menu functionality
type DefaultContextMenuProvider struct {
	paramManager *ParameterManager
	pluginName   string
	pluginVersion string
	aboutMessage string
	host         unsafe.Pointer
}

// NewDefaultContextMenuProvider creates a new default context menu provider
func NewDefaultContextMenuProvider(paramManager *ParameterManager, pluginName, pluginVersion string, host unsafe.Pointer) *DefaultContextMenuProvider {
	return &DefaultContextMenuProvider{
		paramManager:  paramManager,
		pluginName:    pluginName,
		pluginVersion: pluginVersion,
		aboutMessage:  fmt.Sprintf("%s v%s", pluginName, pluginVersion),
		host:          host,
	}
}

// SetAboutMessage sets a custom about message
func (d *DefaultContextMenuProvider) SetAboutMessage(message string) {
	d.aboutMessage = message
}

// PopulateParameterMenu adds common parameter menu items
func (d *DefaultContextMenuProvider) PopulateParameterMenu(paramID uint32, builder *ContextMenuBuilder) bool {
	info, err := d.paramManager.GetParameterInfo(paramID)
	if err != nil {
		return false
	}
	
	// Add parameter title
	builder.AddItem(&ContextMenuTitle{
		Title:     info.Name,
		IsEnabled: true,
	})
	
	builder.AddItem(&ContextMenuSeparator{})
	
	// Add reset to default
	builder.AddItem(&ContextMenuEntry{
		Label:     "Reset to Default",
		IsEnabled: true,
		ActionID:  uint64(1000 + paramID), // Base action ID 1000 for resets
	})
	
	return true
}

// PopulateGlobalMenu adds common global menu items
func (d *DefaultContextMenuProvider) PopulateGlobalMenu(builder *ContextMenuBuilder) bool {
	// Add plugin title
	builder.AddItem(&ContextMenuTitle{
		Title:     d.pluginName,
		IsEnabled: true,
	})
	
	builder.AddItem(&ContextMenuSeparator{})
	
	// Add about item
	builder.AddItem(&ContextMenuEntry{
		Label:     fmt.Sprintf("About %s", d.pluginName),
		IsEnabled: true,
		ActionID:  9999, // Standard action ID for about
	})
	
	return true
}

// HandleResetParameter handles the common reset parameter action
func (d *DefaultContextMenuProvider) HandleResetParameter(paramID uint32) bool {
	info, err := d.paramManager.GetParameterInfo(paramID)
	if err != nil {
		return false
	}
	
	// Reset to default value
	d.paramManager.SetParameterValue(paramID, info.DefaultValue)
	
	// Request parameter flush to notify host of change
	if d.host != nil {
		paramsHost := host.NewParamsHost(d.host)
		paramsHost.RequestFlush()
	}
	
	return true
}

// IsResetAction checks if an action ID is a reset action
func (d *DefaultContextMenuProvider) IsResetAction(actionID uint64) (bool, uint32) {
	if actionID >= 1000 && actionID < 2000 {
		paramID := uint32(actionID - 1000)
		_, err := d.paramManager.GetParameterInfo(paramID)
		if err == nil {
			return true, paramID
		}
	}
	return false, 0
}

// IsAboutAction checks if an action ID is the about action
func (d *DefaultContextMenuProvider) IsAboutAction(actionID uint64) bool {
	return actionID == 9999
}

// ContextMenuBuilder extension for common patterns
// AddParameterPresetSubmenu adds a submenu with parameter presets
func AddParameterPresetSubmenu(builder *ContextMenuBuilder, presetName string, presets []struct {
	Label    string
	Value    float64
	ActionID uint64
}) {
	builder.AddItem(&ContextMenuSubmenu{
		Label:     presetName,
		IsEnabled: true,
	})
	
	for _, preset := range presets {
		builder.AddItem(&ContextMenuEntry{
			Label:     preset.Label,
			IsEnabled: true,
			ActionID:  preset.ActionID,
		})
	}
	
	builder.AddItem(&ContextMenuEndSubmenu{})
}

// SimpleContextMenuPlugin is a minimal interface for plugins that want to use the default provider
type SimpleContextMenuPlugin interface {
	// GetContextMenuProvider returns the default context menu provider
	GetContextMenuProvider() *DefaultContextMenuProvider
	
	// AddCustomMenuItems allows plugins to add their custom items
	// Return true if custom items were added
	AddCustomMenuItems(target *ContextMenuTarget, builder *ContextMenuBuilder) bool
	
	// HandleCustomAction handles plugin-specific actions
	// Return true if the action was handled
	HandleCustomAction(target *ContextMenuTarget, actionID uint64) bool
}

// BuildContextMenu is a helper that uses SimpleContextMenuPlugin interface
func BuildContextMenu(plugin SimpleContextMenuPlugin, target *ContextMenuTarget, builder *ContextMenuBuilder) bool {
	provider := plugin.GetContextMenuProvider()
	
	if target != nil && target.Kind == ContextMenuTargetKindParam {
		// Build parameter menu
		provider.PopulateParameterMenu(uint32(target.ID), builder)
		
		// Add custom items
		plugin.AddCustomMenuItems(target, builder)
	} else {
		// Build global menu
		provider.PopulateGlobalMenu(builder)
		
		// Add custom items
		plugin.AddCustomMenuItems(target, builder)
	}
	
	return true
}

// PerformContextMenuAction is a helper that uses SimpleContextMenuPlugin interface
func PerformContextMenuAction(plugin SimpleContextMenuPlugin, target *ContextMenuTarget, actionID uint64, logger *HostLogger) bool {
	provider := plugin.GetContextMenuProvider()
	
	// Check for reset action
	if isReset, paramID := provider.IsResetAction(actionID); isReset {
		if provider.HandleResetParameter(paramID) {
			if logger != nil {
				info, _ := provider.paramManager.GetParameterInfo(paramID)
				logger.Info(fmt.Sprintf("%s reset to default (%.2f)", info.Name, info.DefaultValue))
			}
			return true
		}
	}
	
	// Check for about action
	if provider.IsAboutAction(actionID) {
		if logger != nil {
			logger.Info(provider.aboutMessage)
		}
		return true
	}
	
	// Let plugin handle custom actions
	return plugin.HandleCustomAction(target, actionID)
}