# ClapGo Implementation TODO

This document outlines the remaining implementation work for ClapGo's CLAP feature support.

## Executive Summary

ClapGo has achieved **professional-grade real-time audio compliance** with 100% allocation-free processing in all critical paths. The core architecture is complete and validated.

### ✅ What's Complete
- **Core Plugin System**: Full CLAP compliance with manifest-driven architecture
- **Zero-Allocation Processing**: Event pools, fixed arrays, buffer pools - VERIFIED with benchmarks
- **Polyphonic Support**: Complete with voice management and MPE
- **All Essential Extensions**: Audio ports, params, state, latency, tail, log, timer support
- **All Advanced Audio Features**: Surround, audio configs, voice info, track info, ambisonic, audio ports activation
- **All Host Integration Extensions**: Context menu, remote controls, param indication, note name
- **Core Extensions**: Event registry, configurable audio ports, POSIX FD support, render, thread pool
- **Performance Validation**: Allocation tracking, benchmarks, profiling tools with thread safety validation

## 🚨 Current Priority: Preset Discovery Factory

The Preset Discovery Factory extension is now our highest priority implementation target. This factory-level extension enables hosts to discover and index presets across the system, providing centralized preset browsing capabilities.

### Why Preset Discovery is Important
- Allows users to browse presets from one central location consistently
- Users can browse presets without committing to a particular plugin first
- Enables fast indexing and searching within the host's browser
- Provides metadata for intelligent categorization and tagging

### Implementation Requirements

**Architecture Considerations:**
- This is a **factory-level** extension, not a plugin extension
- Lives at the plugin entry level, not individual plugin instances
- Must be FAST - the whole indexing process needs to be performant
- Must be NON-INTERACTIVE - no dialogs, windows, or user input
- Zero-allocation design for metadata scanning operations

**Key Components to Implement:**
1. **Preset Discovery Factory** (`clap_preset_discovery_factory_t`)
   - Factory interface in plugin entry
   - Creates preset discovery providers

2. **Preset Discovery Provider** (`clap_preset_discovery_provider_t`)
   - Declares supported file types
   - Declares preset locations
   - Provides metadata extraction

3. **Preset Discovery Indexer** (host-side interface)
   - Receives declarations from provider
   - Handles metadata callbacks

4. **Metadata Structures**
   - Preset metadata with plugin IDs, names, features
   - Soundpack information
   - File type associations

**Implementation Strategy:**
Following the manifest system approach for consistency:
1. **C-Native Implementation**: Build preset discovery entirely in C using json-c like the manifest system
2. Add factory support to plugin entry (not individual plugins)
3. Use json-c for preset file parsing (consistent with manifest.c)
4. Store preset metadata in C structures with fixed allocations
5. Support both file-based and bundled presets
6. Integrate with existing C bridge architecture

**Why C Implementation:**
- Consistent with manifest system architecture
- Zero allocation guarantees using json-c memory management
- Better DAW compatibility with predictable C interfaces
- Simpler architecture without Go/C boundary overhead
- Matches CLAP performance requirements for fast scanning

**Performance Requirements:**
- Pre-allocated buffers for metadata
- No string allocations during scanning
- Efficient file I/O patterns
- Caching strategies for repeated scans

## 📋 Not Planned (Out of Scope)

### GUI Extension
- **Status**: WILL NOT BE IMPLEMENTED
- **Reason**: GUI examples are explicitly forbidden per guardrails
- All GUI-related work is out of scope for ClapGo

## 🔮 Future Extensions (After Preset Discovery)

### Draft Extensions (Lower Priority)
These experimental extensions may be implemented after preset discovery is complete:

#### Tuning Extension (CLAP_EXT_TUNING)
- Microtuning table support
- Fixed-size tuning arrays
- Pre-calculated frequency tables

#### Transport Control Extension (CLAP_EXT_TRANSPORT_CONTROL)
- Transport state callbacks
- Lock-free transport info updates
- Host transport synchronization

#### Undo Extension (CLAP_EXT_UNDO)
- State delta tracking
- Fixed-size undo buffer
- Parameter change history

#### Other Draft Extensions
- Resource Directory
- Triggers
- Scratch Memory
- Mini Curve Display
- Gain Adjustment Metering
- Extensible Audio Ports
- Project Location

## Development Guidelines

### Architecture Principles (MUST FOLLOW)
1. **Factory Extensions Are Different**
   - Live at plugin entry level, not plugin instance
   - No weak symbols on plugin methods
   - Factory creation happens before plugin instantiation

2. **Maintain Zero-Allocation Design**
   - Pre-allocate all buffers for scanning
   - Use fixed-size structures for metadata
   - Avoid string operations in hot paths

3. **Follow Established Patterns**
   - C bridge owns all CLAP interfaces
   - Go provides implementation only
   - No wrapper types or simplifications

### Plugin-Agnostic Preset Discovery Implementation

**Core Philosophy: Dynamic, Manifest-Driven Discovery**
- Presets live in `~/.clap/<plugin_id>/presets/` (plugin_id from manifest)
- C bridge dynamically discovers plugins from loaded manifests
- No hardcoded plugin names or features in C code
- Features and metadata come from preset JSON files

**Architecture Overview:**
```
1. Factory enumerates plugins from loaded manifests
2. Provider declares location based on manifest plugin ID
3. Host scans declared directories for JSON files
4. Provider parses files and extracts ALL metadata from JSON
```

**Phase 1: Plugin-Agnostic C Implementation**
```c
// src/c/preset_discovery.c - Generic implementation
typedef struct {
    char plugin_id[256];        // From manifest
    char plugin_name[256];      // From manifest
    char vendor[256];           // From manifest
    const clap_preset_discovery_indexer_t* indexer;
} provider_data_t;

static bool provider_init(const clap_preset_discovery_provider_t* provider) {
    provider_data_t* data = (provider_data_t*)provider->provider_data;
    
    // Step 1: Declare JSON filetype
    clap_preset_discovery_filetype_t filetype = {
        .name = "JSON Preset",
        .description = "ClapGo JSON preset format",
        .file_extension = "json"
    };
    if (!data->indexer->declare_filetype(data->indexer, &filetype)) {
        return false;
    }
    
    // Step 2: Declare preset location based on plugin ID
    char preset_path[512];
    const char* home = getenv("HOME");
    if (!home) {
        home = "/tmp";  // Fallback
    }
    
    // Extract simple plugin name from ID (e.g., "com.clapgo.gain" -> "gain")
    const char* simple_name = strrchr(data->plugin_id, '.');
    simple_name = simple_name ? simple_name + 1 : data->plugin_id;
    
    snprintf(preset_path, sizeof(preset_path), "%s/.clap/%s/presets", home, simple_name);
    
    clap_preset_discovery_location_t location = {
        .flags = CLAP_PRESET_DISCOVERY_IS_USER_CONTENT,
        .name = "User Presets",
        .kind = CLAP_PRESET_DISCOVERY_LOCATION_FILE,
        .location = preset_path
    };
    
    return data->indexer->declare_location(data->indexer, &location);
}

static bool provider_get_metadata(
    const clap_preset_discovery_provider_t* provider,
    uint32_t location_kind,
    const char* location,
    const clap_preset_discovery_metadata_receiver_t* receiver) {
    
    provider_data_t* data = (provider_data_t*)provider->provider_data;
    
    // Parse JSON file
    struct json_object* root = json_object_from_file(location);
    if (!root) {
        return false;
    }
    
    // Extract name (required)
    struct json_object* name_obj;
    if (!json_object_object_get_ex(root, "name", &name_obj)) {
        json_object_put(root);
        return false;
    }
    
    if (!receiver->begin_preset(receiver, json_object_get_string(name_obj), NULL)) {
        json_object_put(root);
        return false;
    }
    
    // Add plugin ID (from manifest or JSON)
    struct json_object* plugin_ids_obj;
    if (json_object_object_get_ex(root, "plugin_ids", &plugin_ids_obj) && 
        json_object_is_type(plugin_ids_obj, json_type_array)) {
        // Use plugin IDs from preset
        int count = json_object_array_length(plugin_ids_obj);
        for (int i = 0; i < count; i++) {
            struct json_object* id_obj = json_object_array_get_idx(plugin_ids_obj, i);
            if (id_obj) {
                clap_universal_plugin_id_t plugin_id = {
                    .abi = "clap",
                    .id = ""
                };
                strncpy(plugin_id.id, json_object_get_string(id_obj), sizeof(plugin_id.id) - 1);
                receiver->add_plugin_id(receiver, &plugin_id);
            }
        }
    } else {
        // Fallback to manifest plugin ID
        clap_universal_plugin_id_t plugin_id = {
            .abi = "clap",
            .id = ""
        };
        strncpy(plugin_id.id, data->plugin_id, sizeof(plugin_id.id) - 1);
        receiver->add_plugin_id(receiver, &plugin_id);
    }
    
    // Extract ALL other metadata from JSON (no hardcoding)
    struct json_object* obj;
    
    // Description
    if (json_object_object_get_ex(root, "description", &obj)) {
        receiver->set_description(receiver, json_object_get_string(obj));
    }
    
    // Creators
    if (json_object_object_get_ex(root, "creators", &obj) && 
        json_object_is_type(obj, json_type_array)) {
        int count = json_object_array_length(obj);
        for (int i = 0; i < count; i++) {
            struct json_object* creator = json_object_array_get_idx(obj, i);
            if (creator) {
                receiver->add_creator(receiver, json_object_get_string(creator));
            }
        }
    }
    
    // Features (from JSON, not hardcoded)
    if (json_object_object_get_ex(root, "features", &obj) && 
        json_object_is_type(obj, json_type_array)) {
        int count = json_object_array_length(obj);
        for (int i = 0; i < count; i++) {
            struct json_object* feature = json_object_array_get_idx(obj, i);
            if (feature) {
                receiver->add_feature(receiver, json_object_get_string(feature));
            }
        }
    }
    
    // Flags
    if (json_object_object_get_ex(root, "is_favorite", &obj)) {
        if (json_object_get_boolean(obj)) {
            receiver->set_flags(receiver, CLAP_PRESET_DISCOVERY_IS_USER_CONTENT | 
                                        CLAP_PRESET_DISCOVERY_IS_FAVORITE);
        }
    } else {
        receiver->set_flags(receiver, CLAP_PRESET_DISCOVERY_IS_USER_CONTENT);
    }
    
    json_object_put(root);
    return true;
}
```

**Phase 2: Manifest-Driven Factory Implementation**
```c
// src/c/preset_discovery.c - Dynamic factory using manifest system
#include "manifest.h"

// Check if plugin has preset directory
static bool plugin_has_presets(const char* plugin_id) {
    char preset_path[512];
    const char* home = getenv("HOME");
    if (!home) home = "/tmp";
    
    const char* simple_name = strrchr(plugin_id, '.');
    simple_name = simple_name ? simple_name + 1 : plugin_id;
    
    snprintf(preset_path, sizeof(preset_path), "%s/.clap/%s/presets", home, simple_name);
    
    struct stat st;
    return (stat(preset_path, &st) == 0 && S_ISDIR(st.st_mode));
}

static uint32_t factory_count(const clap_preset_discovery_factory_t* factory) {
    // Count plugins with preset directories from loaded manifests
    uint32_t count = 0;
    for (int i = 0; i < manifest_plugin_count; i++) {
        if (plugin_has_presets(manifest_plugins[i].manifest.plugin.id)) {
            count++;
        }
    }
    return count;
}

static const clap_preset_discovery_provider_descriptor_t* factory_get_descriptor(
    const clap_preset_discovery_factory_t* factory, uint32_t index) {
    
    // Find Nth plugin with presets
    uint32_t current = 0;
    for (int i = 0; i < manifest_plugin_count; i++) {
        if (plugin_has_presets(manifest_plugins[i].manifest.plugin.id)) {
            if (current == index) {
                // Generate descriptor from manifest data
                static clap_preset_discovery_provider_descriptor_t desc;
                desc.clap_version = CLAP_VERSION_INIT;
                
                snprintf(desc.id, sizeof(desc.id), "%s.presets", 
                        manifest_plugins[i].manifest.plugin.id);
                snprintf(desc.name, sizeof(desc.name), "%s Presets",
                        manifest_plugins[i].manifest.plugin.name);
                strncpy(desc.vendor, manifest_plugins[i].manifest.plugin.vendor, 
                       sizeof(desc.vendor) - 1);
                
                return &desc;
            }
            current++;
        }
    }
    return NULL;
}

static const clap_preset_discovery_provider_t* factory_create(
    const clap_preset_discovery_factory_t* factory,
    const clap_preset_discovery_indexer_t* indexer,
    const char* provider_id) {
    
    // Find matching plugin from manifests
    for (int i = 0; i < manifest_plugin_count; i++) {
        char expected_id[512];
        snprintf(expected_id, sizeof(expected_id), "%s.presets", 
                manifest_plugins[i].manifest.plugin.id);
        
        if (strcmp(provider_id, expected_id) == 0) {
            // Create provider
            clap_preset_discovery_provider_t* provider = 
                calloc(1, sizeof(clap_preset_discovery_provider_t));
            provider_data_t* data = calloc(1, sizeof(provider_data_t));
            
            // Copy data from manifest
            strncpy(data->plugin_id, manifest_plugins[i].manifest.plugin.id, 
                   sizeof(data->plugin_id) - 1);
            strncpy(data->plugin_name, manifest_plugins[i].manifest.plugin.name,
                   sizeof(data->plugin_name) - 1);
            strncpy(data->vendor, manifest_plugins[i].manifest.plugin.vendor,
                   sizeof(data->vendor) - 1);
            data->indexer = indexer;
            
            provider->desc = factory_get_descriptor(factory, i);
            provider->provider_data = data;
            provider->init = provider_init;
            provider->destroy = provider_destroy;
            provider->get_metadata = provider_get_metadata;
            provider->get_extension = NULL;
            
            return provider;
        }
    }
    
    return NULL;
}
```

**Phase 3: Entry Point Integration**
```c
// src/c/plugin.c - Add factory to entry point
static const void *clapgo_entry_get_factory(const char *factory_id) {
    if (strcmp(factory_id, CLAP_PLUGIN_FACTORY_ID) == 0) {
        return &clapgo_factory;
    }
    
    if (strcmp(factory_id, CLAP_PRESET_DISCOVERY_FACTORY_ID) == 0 ||
        strcmp(factory_id, CLAP_PRESET_DISCOVERY_FACTORY_ID_COMPAT) == 0) {
        return preset_discovery_get_factory();
    }
    
    return NULL;
}
```

**Phase 4: Preset Installation Script**
```bash
#!/bin/bash
# install_presets.sh - Copy presets to user directory

# Create preset directories
mkdir -p ~/.clap/gain/presets
mkdir -p ~/.clap/synth/presets

# Copy gain presets
cp examples/gain/presets/factory/*.json ~/.clap/gain/presets/

# Copy synth presets  
cp examples/synth/presets/factory/*.json ~/.clap/synth/presets/

echo "Presets installed to ~/.clap/"
```

### Plugin-Agnostic Implementation Checklist

**Phase 1: Remove Plugin-Specific Code from Library**
- [ ] **Audit src/c/ directory** - Ensure no "gain" or "synth" references
- [ ] **Audit pkg/ directory** - Ensure no plugin-specific imports or logic
- [ ] **Check for hardcoded plugin IDs** - Remove any "com.clapgo.gain" or "com.clapgo.synth" strings
- [ ] **Document findings** - Update pluginSpecificCodeIssues.md with any new findings

**Phase 2: Create Plugin-Agnostic C Implementation**
- [ ] **Create src/c/preset_discovery.h** - Generic header with provider_data_t structure
- [ ] **Create src/c/preset_discovery.c** - Manifest-driven implementation (~300 lines)
- [ ] **Import manifest.h** - Use existing manifest system for plugin discovery
- [ ] **Include sys/stat.h** - For directory existence checking

**Phase 3: Implement Provider Functions**
- [ ] **Implement provider_init()** - Declare filetype and location based on manifest data
- [ ] **Implement provider_get_metadata()** - Extract ALL metadata from JSON (no hardcoding)
- [ ] **Implement provider_destroy()** - Free allocated memory
- [ ] **Extract plugin name from ID** - Handle "com.clapgo.gain" → "gain" conversion

**Phase 4: Implement Manifest-Driven Factory**
- [ ] **Implement plugin_has_presets()** - Check if ~/.clap/<plugin>/presets exists
- [ ] **Implement factory_count()** - Count manifest plugins with preset directories
- [ ] **Implement factory_get_descriptor()** - Generate descriptors from manifest data
- [ ] **Implement factory_create()** - Match provider_id and create from manifest

**Phase 5: Update Preset JSON Format**
- [ ] **Add features to gain presets** - Add `"features": ["utility", "gain"]` to JSON
- [ ] **Add features to synth presets** - Add `"features": ["instrument", "synthesizer"]` to JSON
- [ ] **Add creators field** - Add `"creators": ["ClapGo Team"]` where missing
- [ ] **Document preset schema** - Create preset-schema.md with expected fields

**Phase 6: Entry Point Integration**
- [ ] **Update src/c/plugin.c** - Add preset factory check in clapgo_entry_get_factory
- [ ] **Include preset_discovery.h** - Add include to plugin.c
- [ ] **Link to manifest system** - Ensure access to manifest_plugins array

**Phase 7: Build System Integration**
- [ ] **Update Makefile** - Add preset_discovery.c to C_BRIDGE_SRCS
- [ ] **Update CMakeLists.txt** - Add preset_discovery.c to clapgo-wrapper
- [ ] **Test compilation** - Ensure no undefined references

**Phase 8: Preset Installation System**
- [ ] **Create scripts/install_presets.sh** - Script to copy presets to ~/.clap/
- [ ] **Make script plugin-agnostic** - Discover plugins from examples/*/presets/
- [ ] **Add to Makefile install** - Run preset installation during make install
- [ ] **Handle missing directories** - Create ~/.clap structure if needed

**Phase 9: Testing Plugin-Agnostic Design**
- [ ] **Add hypothetical plugin test** - Create examples/test/test.json manifest
- [ ] **Verify dynamic discovery** - Ensure test plugin appears without C changes
- [ ] **Test with clap-validator** - Verify all discovered plugins work
- [ ] **Remove plugin test** - Ensure graceful handling of plugin removal

**Phase 10: Documentation Updates**
- [ ] **Update README.md** - Document ~/.clap/ preset structure
- [ ] **Update ARCHITECTURE.md** - Note plugin-agnostic preset discovery
- [ ] **Create PRESETS.md** - Document preset format and discovery
- [ ] **Update plugin template** - Include preset support in new plugin guide

**Success Metrics:**
- Zero hardcoded plugin names in C/Go library code
- New plugins work without modifying library
- All metadata comes from JSON files, not C code
- Manifest system drives all plugin discovery

## Success Criteria
- Preset discovery factory properly integrated at entry level
- Fast, non-interactive preset scanning
- Zero allocations during metadata extraction
- Support for multiple preset locations and formats
- Clean integration with existing ClapGo architecture