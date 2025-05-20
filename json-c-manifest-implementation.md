# JSON-C Based Manifest Implementation

This document explains how we implemented the JSON-C based manifest system to solve the plugin discovery problem in ClapGo.

## Problem

The core issue was that Go plugins register information in their own Go runtime, but when the entry point C code loads the plugin, it runs in a separate Go runtime and cannot access the plugin's registered information. This prevented the clap-validator from being able to properly validate plugins.

## Solution

We implemented a manifest-based plugin discovery system:

1. Each plugin provides a JSON manifest file that describes all of its metadata (ID, name, vendor, etc.)
2. The plugin loader looks for these manifest files to discover plugins
3. We use the json-c library to parse these manifests in the C code

## Implementation Details

### 1. JSON Manifest Files

Each plugin has a JSON manifest file in its directory with the same name as the plugin, e.g., `gain.json` for the `gain` plugin. The manifest contains all the necessary metadata:

```json
{
  "schemaVersion": "1.0",
  "plugin": {
    "id": "com.clapgo.gain",
    "name": "Simple Gain",
    "vendor": "ClapGo",
    "version": "1.0.0",
    "description": "A simple gain plugin using ClapGo",
    "features": ["audio-effect", "stereo", "mono"]
  },
  "build": {
    "goSharedLibrary": "libgain.so",
    "entryPoint": "CreatePlugin"
  },
  "extensions": [
    { "id": "clap.audio-ports", "supported": true },
    { "id": "clap.state", "supported": false }
  ],
  "parameters": [
    {
      "id": 1,
      "name": "Gain",
      "minValue": 0.0,
      "maxValue": 2.0,
      "defaultValue": 1.0,
      "flags": ["automatable", "bounded-below", "bounded-above"]
    }
  ]
}
```

### 2. Manifest Loading with json-c

We upgraded the manifest loading code to use the json-c library:

1. Added the json-c library dependency in the Makefile
2. Updated manifest.c to use json-c functions instead of custom JSON parsing
3. Implemented a robust manifest file search mechanism that looks in multiple locations:
   - Plugin's build directory
   - Central manifest repository (~/.clap/manifests/)

The core implementation in `manifest_load_from_file` now uses json-c:

```c
bool manifest_load_from_file(const char* path, plugin_manifest_t* manifest) {
    // Initialize manifest with default values
    manifest_init(manifest);
    
    // Parse JSON file using json-c
    struct json_object* root = json_object_from_file(path);
    if (!root) {
        fprintf(stderr, "Error: Failed to parse manifest file: %s\n", path);
        return false;
    }
    
    // ... Parse each key from the JSON object ...
    
    // Free the JSON object
    json_object_put(root);
    
    return true;
}
```

### 3. Updated Build System

We updated the Makefile to:

1. Copy manifest files to the plugin build directory
2. Copy manifest files to a central repository (~/.clap/manifests/)
3. Link with the json-c library

```makefile
$(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/lib$(1).$(SO_EXT): $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)
	@echo "Building Go plugin library for $(1)..."
	@cd $(EXAMPLES_DIR)/$(1) && \
	CGO_ENABLED=$(CGO_ENABLED) \
	$(GO) build $(GO_FLAGS) $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/lib$(1).$(SO_EXT) *.go
	@if [ -f "$(EXAMPLES_DIR)/$(1)/$(1).json" ]; then \
		echo "Copying manifest file for $(1)..."; \
		cp "$(EXAMPLES_DIR)/$(1)/$(1).json" "$(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/"; \
		mkdir -p $(HOME)/.clap/manifests/; \
		cp "$(EXAMPLES_DIR)/$(1)/$(1).json" "$(HOME)/.clap/manifests/"; \
	fi
```

### 4. Bridge Integration

We modified the plugin bridge to check for manifests when the Go registry is empty:

```c
// Get the number of registered plugins directly from Go
plugin_count = go_get_registered_plugin_count();
printf("Found %u registered plugins in Go metadata registry\n", plugin_count);

if (plugin_count == 0) {
    fprintf(stderr, "Warning: No plugins found in Go metadata registry, checking manifest files\n");
    
    // ... Try to find and load manifest file ...
    
    if (access(manifest_path, R_OK) == 0 && manifest_load_from_file(manifest_path, &manifest)) {
        printf("Found manifest file: %s\n", manifest_path);
        
        // Override plugin count to 1
        plugin_count = 1;
        
        // Create descriptor from manifest
        plugin_descriptors[0] = manifest_to_descriptor(&manifest);
        
        if (plugin_descriptors[0]) {
            printf("Loaded plugin from manifest: %s (%s)\n", 
                   plugin_descriptors[0]->name, 
                   plugin_descriptors[0]->id);
            manifest_free(&manifest);
            return true;
        }
    }
}
```

## Results

Our implementation successfully allows the plugin validator to discover and test the plugins:

1. The bridge code now tries to find manifest files when the Go plugin registry is empty
2. It loads these manifests from either the plugin directory or central repository
3. The manifest data is used to create plugin descriptors
4. Both the gain and synth plugins now pass all applicable validator tests

## Future Improvements

1. Add proper schema validation for the manifest files
2. Improve error handling for malformed JSON
3. Add caching for manifest parsing to improve performance
4. Extend the manifest format to support more plugin features
5. Add more advanced search paths for manifests (system-wide locations, etc.)