// Package registry provides a centralized registry for CLAP plugins.
// It manages the registration, discovery, and instantiation of plugins.
package registry

import (
	"fmt"
	"runtime/cgo"
	"sync"
	"unsafe"

	"github.com/justyntemme/clapgo/pkg/api"
)

// Global registry instance
var globalRegistry = New()

// PluginEntry represents a registered plugin in the registry.
type PluginEntry struct {
	// Info is the plugin information
	Info api.PluginInfo

	// Creator is a function that creates a new instance of the plugin
	Creator func() api.Plugin
}

// Registry is the centralized registry for all plugins.
type Registry struct {
	// mu protects access to the registry maps
	mu sync.RWMutex

	// plugins maps plugin IDs to plugin entries
	plugins map[string]PluginEntry

	// handles maps plugin pointers to cgo.Handle values for proper cleanup
	handles map[uintptr]bool
}

// New creates a new Registry instance.
func New() *Registry {
	return &Registry{
		plugins: make(map[string]PluginEntry),
		handles: make(map[uintptr]bool),
	}
}

// Register registers a plugin with the registry.
func (r *Registry) Register(info api.PluginInfo, creator func() api.Plugin) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Validate plugin info
	if info.ID == "" {
		fmt.Println("Warning: Attempting to register plugin with empty ID")
		return
	}

	// Check for duplicate ID
	if _, exists := r.plugins[info.ID]; exists {
		fmt.Printf("Warning: Replacing existing plugin with ID: %s\n", info.ID)
	}

	// Store the plugin
	r.plugins[info.ID] = PluginEntry{
		Info:    info,
		Creator: creator,
	}

	fmt.Printf("Registered plugin: %s (%s)\n", info.Name, info.ID)
}

// GetPluginCount returns the number of registered plugins.
func (r *Registry) GetPluginCount() uint32 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return uint32(len(r.plugins))
}

// GetPluginInfo returns information about a plugin by index.
func (r *Registry) GetPluginInfo(index uint32) api.PluginInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if int(index) >= len(r.plugins) {
		return api.PluginInfo{}
	}

	// Convert map to a slice and get the item at the index
	i := uint32(0)
	for _, entry := range r.plugins {
		if i == index {
			return entry.Info
		}
		i++
	}

	return api.PluginInfo{}
}

// CreatePlugin creates a new instance of a plugin by ID.
func (r *Registry) CreatePlugin(id string) api.Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.plugins[id]
	if !exists {
		fmt.Printf("Error: Plugin ID not found: %s\n", id)
		return nil
	}

	plugin := entry.Creator()
	if plugin == nil {
		fmt.Printf("Error: Failed to create plugin with ID: %s\n", id)
		return nil
	}

	fmt.Printf("Created plugin instance: %s (%s)\n", entry.Info.Name, id)
	return plugin
}

// RegisterHandle registers a cgo.Handle for later cleanup.
func (r *Registry) RegisterHandle(handle cgo.Handle) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.handles[uintptr(handle)] = true
}

// UnregisterHandle removes a cgo.Handle from the registry and deletes it.
func (r *Registry) UnregisterHandle(handle cgo.Handle) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.handles[uintptr(handle)]; exists {
		handle.Delete()
		delete(r.handles, uintptr(handle))
	}
}

// CleanupAllHandles releases all registered handles.
func (r *Registry) CleanupAllHandles() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for h := range r.handles {
		handle := cgo.Handle(h)
		handle.Delete()
		delete(r.handles, h)
	}
}

// GetPluginFromPtr retrieves the Go plugin from a plugin pointer.
func (r *Registry) GetPluginFromPtr(ptr unsafe.Pointer) api.Plugin {
	if ptr == nil {
		return nil
	}

	// Convert the plugin pointer back to a handle
	handle := cgo.Handle(uintptr(ptr))

	// Get the plugin from the handle
	value := handle.Value()

	// Try to cast to Plugin
	plugin, ok := value.(api.Plugin)
	if !ok {
		fmt.Printf("Error: Failed to cast handle value to Plugin, got %T\n", value)
		return nil
	}

	return plugin
}

// Global functions that operate on the global registry

// Register registers a plugin with the global registry.
func Register(info api.PluginInfo, creator func() api.Plugin) {
	globalRegistry.Register(info, creator)
}

// GetPluginCount returns the number of registered plugins in the global registry.
func GetPluginCount() uint32 {
	return globalRegistry.GetPluginCount()
}

// GetPluginInfo returns information about a plugin by index from the global registry.
func GetPluginInfo(index uint32) api.PluginInfo {
	return globalRegistry.GetPluginInfo(index)
}

// CreatePlugin creates a new instance of a plugin by ID from the global registry.
func CreatePlugin(id string) api.Plugin {
	return globalRegistry.CreatePlugin(id)
}

// RegisterHandle registers a cgo.Handle with the global registry.
func RegisterHandle(handle cgo.Handle) {
	globalRegistry.RegisterHandle(handle)
}

// UnregisterHandle removes a cgo.Handle from the global registry.
func UnregisterHandle(handle cgo.Handle) {
	globalRegistry.UnregisterHandle(handle)
}

// CleanupAllHandles releases all handles in the global registry.
func CleanupAllHandles() {
	globalRegistry.CleanupAllHandles()
}

// GetPluginFromPtr retrieves a plugin from a pointer using the global registry.
func GetPluginFromPtr(ptr unsafe.Pointer) api.Plugin {
	return globalRegistry.GetPluginFromPtr(ptr)
}

// Global returns the global registry instance.
func Global() *Registry {
	return globalRegistry
}