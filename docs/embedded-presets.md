# Using Go embed for Factory Presets

ClapGo supports embedding factory presets directly into your plugin binary using Go's `embed` package. This ensures your presets are always available without requiring separate file distribution.

## Implementation

### 1. Directory Structure

Create a preset directory in your plugin:
```
my-plugin/
├── main.go
└── presets/
    └── factory/
        ├── preset1.json
        ├── preset2.json
        └── ...
```

### 2. Embed Directive

Add the embed directive after your imports:
```go
import (
    "embed"
    // ... other imports
)

//go:embed presets/factory/*.json
var factoryPresets embed.FS
```

### 3. Loading Embedded Presets

In your `LoadPresetFromLocation` implementation:
```go
case C.uint32_t(api.PresetLocationPlugin):
    // Build the preset path
    presetPath := filepath.Join("presets", "factory", loadKeyStr+".json")
    
    // Read from embedded filesystem
    presetData, err := factoryPresets.ReadFile(presetPath)
    if err != nil {
        // Handle error
        return C.bool(false)
    }
    
    // Parse and load the preset
    var preset map[string]interface{}
    if err := json.Unmarshal(presetData, &preset); err != nil {
        // Handle error
        return C.bool(false)
    }
    
    // Apply preset to plugin state
    p.LoadState(preset)
```

### 4. Listing Available Presets

You can enumerate embedded presets:
```go
func (p *Plugin) GetAvailablePresets() []string {
    var presets []string
    
    entries, err := factoryPresets.ReadDir("presets/factory")
    if err != nil {
        return presets
    }
    
    for _, entry := range entries {
        if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
            presetName := entry.Name()[:len(entry.Name())-5]
            presets = append(presets, presetName)
        }
    }
    
    return presets
}
```

## Preset Format

Presets are stored as JSON files. The format depends on your plugin:

### Simple Parameter Preset
```json
{
  "name": "Preset Name",
  "description": "Description of the preset",
  "gain": 1.5,
  "frequency": 440.0,
  "resonance": 0.7
}
```

### Complex State Preset
```json
{
  "name": "Complex Preset",
  "description": "A more complex preset with nested data",
  "plugin_ids": ["com.vendor.plugin"],
  "parameters": {
    "gain": 1.0,
    "frequency": 440.0
  },
  "state_data": {
    "version": "1.0.0",
    "custom_data": {}
  }
}
```

## Benefits

1. **No External Files** - Presets are compiled into the binary
2. **Always Available** - No missing preset files
3. **Version Control** - Presets are part of your source code
4. **Fast Loading** - No filesystem access needed
5. **Cross-Platform** - Works identically on all platforms

## Examples

See the implemented examples:
- `examples/gain/` - Simple parameter presets
- `examples/synth/` - Complex synthesizer presets

Both examples demonstrate embedding factory presets and loading them via the CLAP preset load extension.