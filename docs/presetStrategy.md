# ClapGo Preset Discovery Strategy

This document outlines ClapGo's approach to CLAP preset discovery and explores alternative strategies for improved DAW compatibility.

## Current Implementation Overview

### Architecture
ClapGo currently implements CLAP preset discovery using a **manifest-driven factory** approach:

1. **Factory Registration**: Preset discovery factory registered at plugin entry level
2. **Manifest Dependency**: Factory depends on loaded manifests to discover plugins with presets
3. **External File Storage**: Presets stored as JSON files in `~/.clap/<plugin>/presets/`
4. **Dynamic Discovery**: Factory dynamically scans manifest registry for plugins with preset directories

### Current Flow
```
DAW Request → Entry Point → Preset Factory → Provider → External JSON Files
                ↓              ↓             ↓            ↓
           get_factory()   factory_count() provider_init() parse presets
```

### Strengths
- ✅ **Manifest Integration**: Consistent with overall ClapGo architecture
- ✅ **Dynamic Discovery**: Automatically detects new plugins with presets
- ✅ **External Management**: Users can easily add/edit preset files
- ✅ **CLAP Compliance**: Passes clap-validator preset discovery tests

### Weaknesses  
- ❌ **DAW Compatibility**: Some DAWs (like Reaper) may not properly detect presets
- ❌ **Timing Dependencies**: Relies on manifest loading during factory access
- ❌ **File System Dependencies**: Requires external files in specific locations
- ❌ **Discovery Complexity**: Complex initialization chain may confuse some hosts

## Alternative Strategy: Embedded Preset Factory

### Concept: Self-Contained Preset System
Instead of relying on external files and manifest discovery, implement a **self-contained embedded preset system** similar to how we load manifests.

### Proposed Architecture

#### 1. **Embedded Preset Data**
```c
// src/c/preset_data.c - Generated from build process
typedef struct embedded_preset {
    const char* plugin_id;
    const char* name;
    const char* description;
    const char* creators[4];
    const char* features[8];
    const char* preset_json_data;
    uint32_t flags;
} embedded_preset_t;

// Auto-generated from examples/*/presets/factory/*.json
static const embedded_preset_t embedded_presets[] = {
    {
        .plugin_id = "com.clapgo.gain",
        .name = "6dB Boost",
        .description = "Boost signal by 6dB",
        .creators = {"ClapGo Team", NULL},
        .features = {"utility", "gain", NULL},
        .preset_json_data = "{\"name\":\"6dB Boost\",\"preset_data\":{\"gain\":2.0}}",
        .flags = CLAP_PRESET_DISCOVERY_IS_FACTORY_CONTENT
    },
    // ... more presets
    {NULL} // Sentinel
};
```

#### 2. **Build-Time Preset Generation**
```makefile
# In Makefile - generate preset data during build
examples/gain/build/preset_data.c: examples/gain/presets/factory/*.json
	scripts/generate_preset_data.sh examples/gain/presets/factory preset_data.c

# Include generated file in build
examples/gain/build/gain.clap: ... preset_data.o
```

#### 3. **Simplified Factory Implementation**
```c
// src/c/preset_factory_embedded.c
static uint32_t embedded_factory_count(const clap_preset_discovery_factory_t* factory) {
    // Count plugins with embedded presets
    uint32_t count = 0;
    for (int i = 0; embedded_presets[i].plugin_id; i++) {
        // Check if this is a new plugin_id
        bool found = false;
        for (int j = 0; j < i; j++) {
            if (strcmp(embedded_presets[i].plugin_id, embedded_presets[j].plugin_id) == 0) {
                found = true;
                break;
            }
        }
        if (!found) count++;
    }
    return count;
}

static bool embedded_provider_init(const clap_preset_discovery_provider_t* provider) {
    // Use CLAP_PRESET_DISCOVERY_LOCATION_PLUGIN for embedded presets
    clap_preset_discovery_location_t location = {
        .flags = CLAP_PRESET_DISCOVERY_IS_FACTORY_CONTENT,
        .name = "Factory Presets",
        .kind = CLAP_PRESET_DISCOVERY_LOCATION_PLUGIN,
        .location = NULL  // Embedded in plugin
    };
    
    return indexer->declare_location(indexer, &location);
}
```

#### 4. **Preset Data Management**
```bash
#!/bin/bash
# scripts/generate_preset_data.sh
PRESET_DIR="$1"
OUTPUT_FILE="$2"

cat > "$OUTPUT_FILE" << 'EOF'
#include "preset_discovery.h"

static const embedded_preset_t embedded_presets[] = {
EOF

for json_file in "$PRESET_DIR"/*.json; do
    # Parse JSON and generate C structure
    python3 scripts/json_to_c_preset.py "$json_file" >> "$OUTPUT_FILE"
done

cat >> "$OUTPUT_FILE" << 'EOF'
    {NULL} // Sentinel
};
EOF
```

### Benefits of Embedded Strategy

#### 1. **Improved DAW Compatibility**
- **Self-Contained**: No external file dependencies
- **Standard Location**: Uses `CLAP_PRESET_DISCOVERY_LOCATION_PLUGIN`
- **Simplified Discovery**: No manifest system dependencies
- **Reliable Initialization**: Embedded data always available

#### 2. **Better Performance**
- **No File I/O**: Presets compiled into binary
- **Faster Discovery**: No directory scanning required
- **Reduced Latency**: Immediate preset access
- **Memory Efficient**: Read-only embedded data

#### 3. **Deployment Simplicity**
- **Single File Distribution**: Everything in .clap file
- **No Installation Requirements**: No ~/.clap directory needed
- **Portable**: Works regardless of file system permissions
- **Version Consistency**: Presets always match plugin version

### Implementation Strategy

#### Phase 1: Dual Implementation
1. **Keep Current System**: Maintain external file support for development
2. **Add Embedded System**: Implement embedded preset factory as alternative
3. **Runtime Selection**: Choose strategy based on availability:
   ```c
   // Try embedded first, fallback to external
   if (has_embedded_presets()) {
       return &embedded_preset_factory;
   } else {
       return &external_preset_factory;
   }
   ```

#### Phase 2: Build Integration
1. **Preset Generation**: Auto-generate embedded preset data during build
2. **Makefile Integration**: Include preset generation in build process
3. **Development Workflow**: Support both development (external) and release (embedded) modes

#### Phase 3: Migration
1. **Test Compatibility**: Verify embedded presets work with various DAWs
2. **Performance Validation**: Ensure embedded approach is faster
3. **Choose Primary Strategy**: Select best approach based on testing

### Technical Implementation Details

#### Build System Changes
```makefile
# Add preset data generation
PRESET_GENERATOR := scripts/generate_preset_data.py
PRESET_DATA_C := $(BUILD_DIR)/preset_data.c
PRESET_DATA_O := $(BUILD_DIR)/preset_data.o

$(PRESET_DATA_C): $(wildcard presets/factory/*.json) $(PRESET_GENERATOR)
	python3 $(PRESET_GENERATOR) presets/factory $@

$(PRESET_DATA_O): $(PRESET_DATA_C)
	$(CC) $(CFLAGS) -c $< -o $@

# Include in plugin build
$(PLUGIN_CLAP): ... $(PRESET_DATA_O)
```

#### Factory Registration
```c
// Choose factory based on available data
const clap_preset_discovery_factory_t* preset_discovery_get_factory(void) {
    #ifdef EMBEDDED_PRESETS_AVAILABLE
        return &embedded_preset_factory;
    #else
        return &external_preset_factory;
    #endif
}
```

### Comparison: External vs Embedded

| Aspect | External Files | Embedded Data |
|--------|----------------|---------------|
| **DAW Compatibility** | Moderate (some issues) | High (standard approach) |
| **Performance** | Slower (file I/O) | Faster (memory access) |
| **Development** | Easy editing | Requires rebuild |
| **Distribution** | Multi-file | Single file |
| **User Customization** | Easy | Difficult |
| **File System Dependencies** | Yes | No |
| **Preset Versioning** | Independent | Coupled with plugin |

### Recommendation

**Implement embedded preset strategy** for the following reasons:

1. **DAW Compatibility**: More likely to work with Reaper and other DAWs
2. **Standard Practice**: Most commercial plugins use embedded presets
3. **Reliability**: No external dependencies or file system issues
4. **Performance**: Faster preset discovery and loading
5. **Distribution**: Simpler plugin distribution and installation

The external file approach can remain as a development convenience, but embedded presets should be the primary release strategy.

### Migration Path

1. **Immediate**: Add debugging to current system to understand Reaper behavior
2. **Short Term**: Implement embedded preset factory as alternative
3. **Medium Term**: Test embedded approach with multiple DAWs
4. **Long Term**: Make embedded presets the default for releases

This strategy aligns with industry best practices while maintaining ClapGo's architectural principles of simplicity and reliability.