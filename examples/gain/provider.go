package main

import (
	"github.com/justyntemme/clapgo/pkg/api"
	// Import the registry package only when the registration code is uncommented
	// "github.com/justyntemme/clapgo/pkg/registry"
)

// GainPluginProvider demonstrates the recommended way to implement
// the api.PluginProvider interface for clean plugin registration.
type GainPluginProvider struct{}

// CreatePlugin returns a new instance of the GainPlugin.
func (p *GainPluginProvider) CreatePlugin() api.Plugin {
	return NewGainPlugin()
}

// GetPluginInfo returns the plugin information.
func (p *GainPluginProvider) GetPluginInfo() api.PluginInfo {
	return api.PluginInfo{
		ID:          PluginID,
		Name:        PluginName,
		Vendor:      PluginVendor,
		URL:         "https://github.com/justyntemme/clapgo",
		ManualURL:   "https://github.com/justyntemme/clapgo",
		SupportURL:  "https://github.com/justyntemme/clapgo/issues",
		Version:     PluginVersion,
		Description: PluginDescription,
		Features:    []string{"audio-effect", "stereo", "mono"},
	}
}

// Uncomment the code below to register using the provider approach
/*
func init() {
	// Using the provider approach
	provider := &GainPluginProvider{}
	registry.RegisterPlugin(provider)
}
*/