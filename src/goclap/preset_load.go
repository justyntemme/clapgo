package goclap

// #include <stdlib.h>
// #include <string.h>
// #include "../../include/clap/include/clap/clap.h"
// #include "../../include/clap/include/clap/ext/preset-load.h"
//
// // Bridge function for the preset load extension
// bool preset_load_from_location(const struct clap_plugin *plugin, uint32_t location_kind,
//                              const char *location, const char *load_key) {
//     // We'll implement this bridge later - for now just return true
//     return true;
// }
//
// // Create a C preset load extension with proper function pointers
// clap_plugin_preset_load_t* create_preset_load_extension() {
//     clap_plugin_preset_load_t* ext = (clap_plugin_preset_load_t*)malloc(sizeof(clap_plugin_preset_load_t));
//     if (!ext) return NULL;
//     
//     ext->from_location = preset_load_from_location;
//     
//     return ext;
// }
import "C"
import (
	"encoding/json"
	"os"
	"unsafe"
)

// PresetLoadExtension represents the CLAP preset load extension for plugins
type PresetLoadExtension struct {
	plugin       AudioProcessor
	presetLoadExt unsafe.Pointer // Pointer to the C extension
}

// NewPresetLoadExtension creates a new preset load extension
func NewPresetLoadExtension(plugin AudioProcessor) *PresetLoadExtension {
	if plugin == nil {
		return nil
	}
	
	ext := &PresetLoadExtension{
		plugin: plugin,
	}
	
	// Create the C interface
	ext.presetLoadExt = unsafe.Pointer(C.create_preset_load_extension())
	
	return ext
}

// GetExtensionPointer returns the C extension interface pointer
func (p *PresetLoadExtension) GetExtensionPointer() unsafe.Pointer {
	return p.presetLoadExt
}

// LoadPreset loads a preset from a location
func (p *PresetLoadExtension) LoadPreset(locationKind uint32, location, loadKey string) bool {
	// Check if the plugin implements the PresetLoader interface
	if loader, ok := p.plugin.(PresetLoader); ok {
		return loader.LoadPreset(locationKind, location, loadKey)
	}
	
	// Default implementation for plugins that don't implement PresetLoader
	if loadKey == "" {
		// If no load key is provided, use the location as the file path
		loadKey = location
	}
	
	// Open the preset file
	file, err := os.Open(loadKey)
	if err != nil {
		return false
	}
	defer file.Close()
	
	// Parse the preset JSON
	var preset struct {
		// Basic preset metadata
		Name        string   `json:"name"`
		Description string   `json:"description"`
		PluginIDs   []string `json:"plugin_ids"`
		
		// Common parameter settings (can be mapped to parameter IDs)
		Volume     *float64 `json:"volume,omitempty"`
		Attack     *float64 `json:"attack,omitempty"`
		Decay      *float64 `json:"decay,omitempty"`
		Sustain    *float64 `json:"sustain,omitempty"`
		Release    *float64 `json:"release,omitempty"`
		Waveform   *int     `json:"waveform,omitempty"`
		
		// Generic parameter values by name
		Parameters map[string]float64 `json:"parameters,omitempty"`
		
		// Custom state data for the plugin
		StateData map[string]interface{} `json:"state_data,omitempty"`
	}
	
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&preset); err != nil {
		return false
	}
	
	// Check if this preset applies to our plugin
	pluginSupported := false
	if len(preset.PluginIDs) > 0 {
		pluginID := p.plugin.(interface{ GetPluginID() string }).GetPluginID()
		for _, id := range preset.PluginIDs {
			if id == pluginID {
				pluginSupported = true
				break
			}
		}
		if !pluginSupported {
			return false
		}
	}
	
	paramManager := p.plugin.GetParamManager()
	if paramManager != nil {
		// Set common parameters if they're present in the preset
		if preset.Volume != nil {
			paramManager.SetParamByName("Volume", *preset.Volume)
		}
		
		if preset.Attack != nil {
			paramManager.SetParamByName("Attack", *preset.Attack)
		}
		
		if preset.Decay != nil {
			paramManager.SetParamByName("Decay", *preset.Decay)
		}
		
		if preset.Sustain != nil {
			paramManager.SetParamByName("Sustain", *preset.Sustain)
		}
		
		if preset.Release != nil {
			paramManager.SetParamByName("Release", *preset.Release)
		}
		
		if preset.Waveform != nil {
			paramManager.SetParamByName("Waveform", float64(*preset.Waveform))
		}
		
		// Update any other parameter values
		if preset.Parameters != nil {
			for name, value := range preset.Parameters {
				paramManager.SetParamByName(name, value)
			}
		}
	}
	
	// Load state data if the plugin implements Stater interface
	if preset.StateData != nil {
		if stater, ok := p.plugin.(Stater); ok {
			stater.LoadState(preset.StateData)
		}
	}
	
	return true
}

// PresetLoader interface for plugins that support loading presets
type PresetLoader interface {
	// LoadPreset loads a preset from the specified location and load key
	// Returns true if the preset was loaded successfully
	LoadPreset(locationKind uint32, location, loadKey string) bool
}