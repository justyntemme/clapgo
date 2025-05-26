# Plugin-Specific Code Issues in ClapGo

This document identifies areas where plugin-specific logic has leaked into the generic C bridge or Go library code, violating the principle that plugin-specific code should only exist in plugin projects (e.g., examples/ directory).

## Core Principle

**Plugin-specific code should ONLY exist in:**
- `examples/gain/` - Gain plugin implementation
- `examples/synth/` - Synth plugin implementation  
- Future plugin directories

**Generic code should exist in:**
- `src/c/` - C bridge (must remain plugin-agnostic)
- `pkg/` - Go library (must remain plugin-agnostic)

## Phase 1 Audit Results (Current Status)

### ✅ src/c/ Directory - CLEAN
- No plugin-specific code found
- No references to "gain" or "synth" plugins
- No hardcoded plugin IDs
- Fully plugin-agnostic implementation

### ✅ pkg/ Directory - CLEAN  
- No plugin-specific imports or logic
- References to "gain" are generic (e.g., audio gain concept, not the plugin)
- References to "synthesizer" are generic (e.g., voice processing)
- No hardcoded plugin IDs

### ✅ Plugin ID Search - CLEAN
- All plugin IDs are confined to examples/ directory
- Found IDs:
  - `com.clapgo.gain` (examples/gain/)
  - `com.clapgo.gain-gui` (examples/gain-with-gui/)
  - `com.justyntemme.ClapGo.synth` (examples/synth/)
- cmd/generate-manifest accepts plugin ID as a parameter (not hardcoded)

### 🔍 Minor Inconsistency Found
- Synth plugin uses different ID format: `com.justyntemme.ClapGo.synth`
- Gain plugin uses: `com.clapgo.gain`
- This inconsistency exists only in examples, not in library code

## Identified Issues (From Planned Implementation)

### 1. Hardcoded Plugin Lists in Preset Discovery (TODO.md)

**Location:** TODO.md lines 266-267, 280-293
```c
static const char* known_plugins[] = {"gain", "synth"};
static const uint32_t known_plugin_count = 2;
```

**Problem:** 
- C bridge hardcodes specific plugin names
- Adding new plugins requires modifying C bridge code
- Violates plugin-agnostic principle

**Solution:**
- Use manifest system to discover plugins dynamically
- Preset factory should enumerate from loaded manifests
- No hardcoded plugin lists

### 2. Plugin-Specific Features in Preset Discovery (TODO.md)

**Location:** TODO.md lines 245-251
```c
// Add simple features based on plugin type
if (strcmp(data->plugin_id, "gain") == 0) {
    receiver->add_feature(receiver, "utility");
    receiver->add_feature(receiver, "gain");
} else if (strcmp(data->plugin_id, "synth") == 0) {
    receiver->add_feature(receiver, "instrument");
    receiver->add_feature(receiver, "synthesizer");
}
```

**Problem:**
- C code contains plugin-specific feature logic
- Features should come from preset JSON data
- Hardcoded if/else for each plugin type

**Solution:**
- Extract features from preset JSON files
- Let plugins define their own features
- No plugin-specific conditionals in C

### 3. Static Plugin Descriptors (TODO.md)

**Location:** TODO.md lines 280-293
```c
static clap_preset_discovery_provider_descriptor_t descriptors[2] = {
    {
        .clap_version = CLAP_VERSION_INIT,
        .id = "com.clapgo.gain.presets",
        .name = "ClapGo Gain Presets", 
        .vendor = "ClapGo"
    },
    {
        .clap_version = CLAP_VERSION_INIT,
        .id = "com.clapgo.synth.presets",
        .name = "ClapGo Synth Presets",
        .vendor = "ClapGo"
    }
};
```

**Problem:**
- Statically defined descriptors for specific plugins
- Cannot support dynamic plugin discovery
- New plugins require C code changes

**Solution:**
- Generate descriptors from manifest data
- Use plugin metadata from manifest.json files
- Dynamic descriptor creation

### 4. Provider ID Parsing Logic (TODO.md)

**Location:** TODO.md lines 304-310
```c
// Determine which plugin based on provider_id
const char* plugin_id = NULL;
if (strstr(provider_id, "gain")) {
    plugin_id = "gain";
} else if (strstr(provider_id, "synth")) {
    plugin_id = "synth";
} else {
    return NULL;
}
```

**Problem:**
- Hardcoded string matching for known plugins
- Cannot handle new plugins without code changes
- Plugin-specific parsing logic

**Solution:**
- Use consistent provider ID format
- Extract plugin ID systematically
- No hardcoded plugin name checks

## Proposed Solutions

### Solution 1: Manifest-Driven Preset Discovery

Instead of hardcoding plugins, use the existing manifest system:

```c
// Dynamic plugin enumeration from manifests
static uint32_t factory_count(const clap_preset_discovery_factory_t* factory) {
    // Count plugins that have preset directories
    uint32_t count = 0;
    for (int i = 0; i < manifest_plugin_count; i++) {
        char preset_dir[512];
        snprintf(preset_dir, sizeof(preset_dir), "%s/.clap/%s/presets", 
                 getenv("HOME"), manifest_plugins[i].manifest.plugin.id);
        
        if (directory_exists(preset_dir)) {
            count++;
        }
    }
    return count;
}
```

### Solution 2: Generic Feature Extraction

Extract features from preset JSON instead of hardcoding:

```c
// Parse features from JSON if present
struct json_object* features_obj;
if (json_object_object_get_ex(root, "features", &features_obj) && 
    json_object_is_type(features_obj, json_type_array)) {
    
    int feature_count = json_object_array_length(features_obj);
    for (int i = 0; i < feature_count; i++) {
        struct json_object* feature = json_object_array_get_idx(features_obj, i);
        if (feature) {
            receiver->add_feature(receiver, json_object_get_string(feature));
        }
    }
}
```

### Solution 3: Dynamic Descriptor Generation

Generate descriptors from manifest data:

```c
static const clap_preset_discovery_provider_descriptor_t* factory_get_descriptor(
    const clap_preset_discovery_factory_t* factory, uint32_t index) {
    
    // Find Nth plugin with presets
    uint32_t current = 0;
    for (int i = 0; i < manifest_plugin_count; i++) {
        if (has_preset_directory(manifest_plugins[i].manifest.plugin.id)) {
            if (current == index) {
                // Generate descriptor from manifest
                static clap_preset_discovery_provider_descriptor_t desc;
                desc.clap_version = CLAP_VERSION_INIT;
                
                snprintf(desc.id, sizeof(desc.id), "%s.presets", 
                        manifest_plugins[i].manifest.plugin.id);
                snprintf(desc.name, sizeof(desc.name), "%s Presets",
                        manifest_plugins[i].manifest.plugin.name);
                strcpy(desc.vendor, manifest_plugins[i].manifest.plugin.vendor);
                
                return &desc;
            }
            current++;
        }
    }
    return NULL;
}
```

## Understanding CLAP Preset Discovery Architecture

### Why Both Factory and Directory Scanning?

The confusion about CLAP requiring both a factory and host directory scanning comes from misunderstanding the architecture:

1. **Factory Role**: Tells the host WHERE to look for presets
   - Provider declares locations (directories)
   - Provider declares file types (.json, .fxp, etc.)
   - Provider declares metadata about preset collections

2. **Host Role**: Actually scans the declared locations
   - Host reads directory contents
   - Host finds files matching declared types
   - Host calls provider to extract metadata from each file

3. **Provider Role**: Extracts metadata when asked
   - Host says "I found preset.json at this path"
   - Provider parses that specific file
   - Provider returns metadata to host

**The Factory doesn't scan - it declares. The Host scans based on declarations.**

### Correct Implementation Pattern

```c
// Provider declares WHERE presets are
static bool provider_init(provider) {
    // Tell host: "Look in ~/.clap/myplugin/presets for .json files"
    indexer->declare_location(indexer, &location);
    indexer->declare_filetype(indexer, &filetype);
    return true;
}

// Host scans and calls back for each file
static bool provider_get_metadata(provider, location_kind, location, receiver) {
    // Host says: "I found ~/.clap/myplugin/presets/bass.json"
    // We parse that specific file and return metadata
    parse_json_file(location);
    return true;
}
```

## Recommendations

1. **Remove ALL plugin-specific code from C bridge**
   - No hardcoded plugin names
   - No static plugin lists
   - No plugin-specific conditionals

2. **Use manifest system for plugin discovery**
   - Plugins already registered via manifests
   - Reuse this system for preset discovery
   - Keep everything dynamic

3. **Let preset files define their own metadata**
   - Features should come from JSON
   - Plugin IDs can be inferred or specified in JSON
   - No hardcoded metadata in C

4. **Maintain strict separation**
   - C bridge = generic CLAP interface
   - Go library = generic plugin framework
   - Plugin directories = all specific logic

## Testing Plugin-Agnostic Design

To verify the design is truly plugin-agnostic:

1. **Add a hypothetical new plugin**
   - Create `examples/reverb/` directory
   - Add `reverb.json` manifest
   - Add `~/.clap/reverb/presets/` directory
   - System should work without ANY C code changes

2. **Remove existing plugin references**
   - Temporarily remove "gain" and "synth" from code
   - Replace with generic plugin enumeration
   - Everything should still compile and work

3. **Dynamic testing**
   - Load plugins from different directories
   - Use different plugin IDs
   - Ensure no assumptions about specific plugins

## Conclusion

The current preset discovery design in TODO.md contains significant plugin-specific code that violates ClapGo's architecture principles. By moving to a fully dynamic, manifest-driven approach, we can ensure the C bridge remains truly plugin-agnostic and new plugins can be added without modifying any bridge code.