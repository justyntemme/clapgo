package main

/*
#include <stdint.h>
#include <stdbool.h>
#include <stdlib.h>
*/
import "C"
import (
	"unsafe"

	"github.com/justyntemme/clapgo/internal/registry"
	"github.com/justyntemme/clapgo/pkg/api"
)

func init() {
	// Register the gain plugin directly
	// This approach solves the issue of not being able to import the main packages
	registerGainPlugin()
}

func registerGainPlugin() {
	// Create plugin info for the gain plugin
	info := api.PluginInfo{
		ID:          "com.clapgo.gain",
		Name:        "Simple Gain",
		Vendor:      "ClapGo",
		URL:         "https://github.com/justyntemme/clapgo",
		ManualURL:   "https://github.com/justyntemme/clapgo",
		SupportURL:  "https://github.com/justyntemme/clapgo/issues",
		Version:     "1.0.0",
		Description: "A simple gain plugin using ClapGo",
		Features:    []string{"audio-effect", "stereo", "mono"},
	}
	
	// Register a plugin factory
	registry.Register(info, func() api.Plugin {
		return &TestPlugin{info: info}
	})
}

// TestPlugin is a simple plugin implementation for testing
type TestPlugin struct {
	info api.PluginInfo
}

func (p *TestPlugin) Init() bool {
	return true
}

func (p *TestPlugin) Destroy() {
}

func (p *TestPlugin) Activate(sampleRate float64, minFrames, maxFrames uint32) bool {
	return true
}

func (p *TestPlugin) Deactivate() {
}

func (p *TestPlugin) StartProcessing() bool {
	return true
}

func (p *TestPlugin) StopProcessing() {
}

func (p *TestPlugin) Reset() {
}

func (p *TestPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events api.EventHandler) int {
	// Copy input to output with a gain of 0.5
	if len(audioIn) > 0 && len(audioOut) > 0 {
		numChannels := len(audioIn)
		if len(audioOut) < numChannels {
			numChannels = len(audioOut)
		}
		
		for ch := 0; ch < numChannels; ch++ {
			inChannel := audioIn[ch]
			outChannel := audioOut[ch]
			
			for i := uint32(0); i < framesCount; i++ {
				if i < uint32(len(inChannel)) && i < uint32(len(outChannel)) {
					outChannel[i] = inChannel[i] * 0.5
				}
			}
		}
	}
	
	return api.ProcessContinue
}

func (p *TestPlugin) GetExtension(id string) unsafe.Pointer {
	return nil
}

func (p *TestPlugin) OnMainThread() {
}

// GetPluginID returns the plugin ID
func (p *TestPlugin) GetPluginID() string {
	return p.info.ID
}

// GetPluginInfo returns the plugin info
func (p *TestPlugin) GetPluginInfo() api.PluginInfo {
	return p.info
}