// Package registry provides a centralized registry for plugins.
// It manages the registration and lookup of plugins by ID.
package registry

import (
	"sync"

	"github.com/justyntemme/clapgo/pkg/api"
)

// Registry is a singleton that manages the registration of plugins.
var Registry = &registry{
	plugins: make(map[string]pluginEntry),
}

// pluginEntry represents a registered plugin.
type pluginEntry struct {
	// Info is the plugin information
	Info api.PluginInfo

	// Creator is a function that creates a new instance of the plugin
	Creator func() api.Plugin
}

// registry is the internal implementation of the plugin registry.
type registry struct {
	// mu protects access to the plugins map
	mu sync.RWMutex

	// plugins maps plugin IDs to plugin entries
	plugins map[string]pluginEntry
}

// Register registers a plugin with the registry.
func (r *registry) Register(info api.PluginInfo, creator func() api.Plugin) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.plugins[info.ID] = pluginEntry{
		Info:    info,
		Creator: creator,
	}
}

// GetPluginCount returns the number of registered plugins.
func (r *registry) GetPluginCount() uint32 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return uint32(len(r.plugins))
}

// GetPluginInfo returns information about a plugin by index.
func (r *registry) GetPluginInfo(index uint32) api.PluginInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if int(index) >= len(r.plugins) {
		return api.PluginInfo{}
	}

	// Convert map to a slice
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
func (r *registry) CreatePlugin(id string) api.Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if entry, ok := r.plugins[id]; ok {
		return entry.Creator()
	}

	return nil
}

// Register registers a plugin with the registry.
func Register(info api.PluginInfo, creator func() api.Plugin) {
	Registry.Register(info, creator)
}

// GetPluginCount returns the number of registered plugins.
func GetPluginCount() uint32 {
	return Registry.GetPluginCount()
}

// GetPluginInfo returns information about a plugin by index.
func GetPluginInfo(index uint32) api.PluginInfo {
	return Registry.GetPluginInfo(index)
}

// CreatePlugin creates a new instance of a plugin by ID.
func CreatePlugin(id string) api.Plugin {
	return Registry.CreatePlugin(id)
}