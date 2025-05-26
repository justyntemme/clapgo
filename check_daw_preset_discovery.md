# CLAP Preset Discovery - DAW Integration

## The Issue
The preset discovery implementation is working correctly (as proven by our test program), but DAWs are not using it. This is because:

1. **Preset Discovery is a separate process** - DAWs typically run preset discovery as a separate indexing/scanning process, not when loading individual plugins
2. **The plugin-factory and preset-discovery-factory are separate** - DAWs request these at different times
3. **Most DAWs have their own preset browsers** that need to be triggered to scan for presets

## How CLAP Preset Discovery Works
1. DAW runs a preset scanner/indexer (often a separate process)
2. Scanner loads each plugin and requests the preset-discovery-factory
3. Scanner uses the factory to index all presets
4. DAW stores this preset metadata in its own database
5. When user selects a preset, DAW uses the preset-load extension to load it

## To Make Presets Visible in Your DAW

### Option 1: Trigger DAW's Preset Scanner
Most DAWs have a menu option like:
- "Scan for presets"
- "Refresh preset database"
- "Rescan plugins"
- "Update preset library"

### Option 2: Check DAW's Preset Browser
Look for:
- A preset browser window/panel
- Factory presets section
- Plugin-specific preset lists

### Option 3: Some DAWs Don't Support CLAP Preset Discovery
Some DAWs may:
- Only support their own preset format
- Require presets in specific locations
- Not implement CLAP preset discovery at all

## What We've Implemented
✓ Preset discovery factory that works correctly
✓ Preset provider that finds and parses JSON presets
✓ Preset-load extension for loading presets
✓ Presets installed in correct location (~/.clap/plugin/presets/factory/)

## Next Steps
1. Check your DAW's documentation for preset discovery support
2. Look for a preset scanning/indexing feature
3. Try different DAWs that are known to support CLAP preset discovery (like Bitwig Studio, Reaper with CLAP support)
4. Consider implementing a simple preset menu in the plugin UI as a fallback