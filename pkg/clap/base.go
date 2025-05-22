// Package clap provides implementations and utilities for CLAP plugins.
// It contains base implementations and helper functions for creating and
// managing CLAP plugins.
package clap

import (
	"unsafe"

	"github.com/justyntemme/clapgo/pkg/api"
)

// BasePlugin provides a base implementation of the api.Plugin interface.
// It can be embedded in plugin implementations to provide common functionality.
type BasePlugin struct {
	// Info contains the plugin information
	Info api.PluginInfo

	// IsActivated indicates whether the plugin is activated
	IsActivated bool

	// IsProcessing indicates whether the plugin is processing
	IsProcessing bool

	// SampleRate stores the current sample rate
	SampleRate float64
}

// NewBasePlugin creates a new BasePlugin instance.
func NewBasePlugin(info api.PluginInfo) *BasePlugin {
	return &BasePlugin{
		Info:        info,
		IsActivated: false,
		IsProcessing: false,
		SampleRate:  44100.0,
	}
}

// Init initializes the plugin.
func (p *BasePlugin) Init() bool {
	return true
}

// Destroy releases all resources.
func (p *BasePlugin) Destroy() {
	// Base implementation has no resources to release
}

// Activate prepares the plugin for processing.
func (p *BasePlugin) Activate(sampleRate float64, minFrames, maxFrames uint32) bool {
	p.SampleRate = sampleRate
	p.IsActivated = true
	return true
}

// Deactivate stops the plugin from processing.
func (p *BasePlugin) Deactivate() {
	p.IsActivated = false
}

// StartProcessing prepares for audio processing.
func (p *BasePlugin) StartProcessing() bool {
	if !p.IsActivated {
		return false
	}
	p.IsProcessing = true
	return true
}

// StopProcessing stops audio processing.
func (p *BasePlugin) StopProcessing() {
	p.IsProcessing = false
}

// Reset resets the plugin state.
func (p *BasePlugin) Reset() {
	// Base implementation has no state to reset
}

// Process handles audio processing.
// This implementation just copies input to output.
// Plugin implementations should override this method.
func (p *BasePlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events api.EventHandler) int {
	// Check state
	if !p.IsActivated || !p.IsProcessing {
		return api.ProcessError
	}

	// Process events
	if events != nil {
		events.ProcessInputEvents()
	}

	// Copy input to output
	for ch := 0; ch < len(audioIn) && ch < len(audioOut); ch++ {
		inChannel := audioIn[ch]
		outChannel := audioOut[ch]

		// Make sure we have enough buffer space
		if len(inChannel) < int(framesCount) || len(outChannel) < int(framesCount) {
			return api.ProcessError
		}

		// Copy samples
		copy(outChannel[:framesCount], inChannel[:framesCount])
	}

	return api.ProcessContinue
}

// GetExtension retrieves a plugin extension.
// The base implementation returns nil for all extensions.
// Plugin implementations should override this for supported extensions.
func (p *BasePlugin) GetExtension(id string) unsafe.Pointer {
	// Base implementation doesn't support any extensions
	return nil
}

// OnMainThread is called on the main thread.
func (p *BasePlugin) OnMainThread() {
	// Base implementation does nothing on the main thread
}

// GetPluginID returns the plugin ID.
func (p *BasePlugin) GetPluginID() string {
	return p.Info.ID
}

// GetPluginInfo returns the plugin information.
func (p *BasePlugin) GetPluginInfo() api.PluginInfo {
	return p.Info
}