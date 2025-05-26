# CLAP Preset Discovery - Current Status and Workarounds

## Summary
The preset discovery implementation is **working correctly** but DAWs are not utilizing it during normal plugin operations. This is because preset discovery is typically a separate indexing process that DAWs run independently from plugin loading.

## What We've Accomplished
✅ Fixed the DAW crash issue  
✅ Implemented full CLAP preset discovery API  
✅ Preset discovery factory works correctly  
✅ Preset provider finds and parses JSON presets  
✅ Preset-load extension is implemented  
✅ Presets are installed in the correct location  

## Why Presets Don't Show in DAWs
1. DAWs only request `clap.plugin-factory` during plugin loading
2. Preset discovery (`clap.preset-discovery-factory`) is a separate process
3. Many DAWs have their own preset management systems
4. Some DAWs may not fully support CLAP preset discovery yet

## Workarounds

### 1. Implement In-Plugin Preset Menu
Add a simple preset selection parameter or menu within the plugin itself:
```go
// Add a preset selection parameter
params = append(params, api.Parameter{
    ID: PresetSelectParamID,
    Name: "Preset",
    MinValue: 0,
    MaxValue: float64(len(presets)-1),
    DefaultValue: 0,
    Flags: api.ParamFlagIsAutomatable | api.ParamFlagIsDiscrete,
})
```

### 2. Manual Preset Loading
Implement state loading so users can manually load preset files:
- Save presets as plugin state files
- Users can load them through the DAW's preset/state system

### 3. Check DAW-Specific Documentation
Different DAWs handle CLAP presets differently:
- **Bitwig Studio**: Has good CLAP support, may require preset rescan
- **REAPER**: Check if CLAP extension is up to date
- **Other DAWs**: May have limited or no CLAP preset discovery support

### 4. Future Enhancement
Consider implementing a simple GUI with a preset dropdown when GUI support is added to ClapGo.

## Testing Preset Discovery
To verify preset discovery is working:
```bash
cd /home/user/Documents/code/clapgo
./test_full_preset_discovery
```

This shows that the implementation is correct - it's just that DAWs need to explicitly use it.