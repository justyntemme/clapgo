# Plugin Discovery Mechanism Design

## Overview

The plugin discovery mechanism is designed to solve the architectural issue where plugins register themselves at runtime, but the bridge and plugin shared libraries have separate Go runtimes and therefore separate registries. This design proposes a static approach where plugin metadata is stored in manifest files that are generated during the build process and loaded by the bridge at runtime.

## Manifest File Format

Each plugin will have a corresponding manifest file in JSON format:

```json
{
  "schemaVersion": "1.0",
  "plugin": {
    "id": "com.clapgo.gain",
    "name": "Simple Gain",
    "vendor": "ClapGo",
    "version": "1.0.0",
    "description": "A simple gain plugin using ClapGo",
    "url": "https://github.com/justyntemme/clapgo",
    "manualUrl": "https://github.com/justyntemme/clapgo",
    "supportUrl": "https://github.com/justyntemme/clapgo/issues",
    "features": ["audio-effect", "stereo", "mono"]
  },
  "build": {
    "goSharedLibrary": "libgain.so",
    "entryPoint": "CreateGainPlugin",
    "dependencies": []
  },
  "extensions": [
    {
      "id": "clap.audio-ports",
      "supported": true
    },
    {
      "id": "clap.state",
      "supported": false
    }
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

## Build Process Integration

1. **Manifest Generation**:
   - Add a step to the build process that generates a manifest file for each plugin
   - The generator can use Go reflection or direct parsing of the plugin source code to extract metadata
   - Add a command-line tool `generate-manifest` that plugin developers can use directly

2. **Directory Structure**:
   ```
   examples/plugin-name/
   ├── main.go              # Go plugin implementation
   ├── provider.go          # Plugin provider for registration
   ├── constants.go         # Plugin-specific constants
   ├── manifest.json        # Generated plugin manifest
   └── plugin/              # C skeleton code
       ├── plugin.c         # Generated C entry point
       └── plugin.h         # C header with plugin information
   ```

3. **Makefile/CMake Integration**:
   - Add a `generate-manifests` target that creates manifests for all plugins
   - Make the build process depend on manifest generation
   - Install manifest files alongside the plugins

## Bridge Implementation

The bridge will be modified to use the manifest files instead of relying on runtime plugin registration:

1. **Manifest Discovery**:
   - During initialization, scan for manifest files in predefined locations:
     - Same directory as the .clap file
     - Standard plugin directories (`~/.clap`, `/usr/lib/clap`, etc.)
   - Parse and validate each manifest file

2. **Plugin Factory**:
   - Create a new `ManifestBasedPluginFactory` that loads plugin metadata from manifests
   - Implement `GetPluginCount`, `GetPluginDescriptor`, and `CreatePlugin` based on manifest data

3. **Lazy Loading**:
   - Load plugin shared libraries on demand when a plugin instance is requested
   - Cache loaded libraries and instances for performance

4. **Fallback Mechanism**:
   - If no manifest is found, fall back to the current runtime registration approach
   - Log warnings if inconsistencies are detected

## Plugin Loading Process

When a host requests a plugin, the following sequence occurs:

1. Host loads the CLAP plugin (C skeleton)
2. C skeleton initializes the Go bridge
3. Bridge scans for manifest files
4. Bridge builds a plugin inventory from manifests
5. When a specific plugin is requested:
   - Bridge finds the corresponding manifest
   - Bridge loads the specified Go shared library if not already loaded
   - Bridge calls the plugin's entry point function to create an instance
   - Bridge returns the plugin instance to the C skeleton

## Advantages of This Approach

1. **Decoupling**: The plugin metadata is decoupled from runtime registration
2. **Reliability**: Reduces dependency on Go runtime behavior
3. **Performance**: Faster startup as metadata is loaded directly from manifests
4. **Discoverability**: Makes it easier to list available plugins without loading them
5. **Extensibility**: The manifest can be extended with additional metadata as needed

## Implementation Plan

1. **Phase 1: Basic Manifest Support**
   - Implement manifest file format
   - Create manifest generator tool
   - Modify bridge to read manifests
   - Maintain backward compatibility with runtime registration

2. **Phase 2: Enhanced Manifest Features**
   - Add support for extension metadata
   - Add parameter information
   - Implement validation for manifest data

3. **Phase 3: Full Integration**
   - Update build system to generate manifests automatically
   - Add installation logic for manifests
   - Create tooling for manifest management
   - Add versioning and compatibility checks

## Code Examples

### Manifest Generator (Pseudo-code)

```go
func GenerateManifest(plugin api.Plugin) ([]byte, error) {
    info := plugin.GetPluginInfo()
    
    manifest := ManifestData{
        SchemaVersion: "1.0",
        Plugin: PluginInfo{
            ID:          info.ID,
            Name:        info.Name,
            Vendor:      info.Vendor,
            Version:     info.Version,
            Description: info.Description,
            URL:         info.URL,
            ManualURL:   info.ManualURL,
            SupportURL:  info.SupportURL,
            Features:    info.Features,
        },
        Build: BuildInfo{
            GoSharedLibrary: fmt.Sprintf("lib%s.%s", pluginName, soExt),
            EntryPoint:      fmt.Sprintf("Create%sPlugin", pascalCase(pluginName)),
        },
    }
    
    // Add extensions and parameters data
    
    return json.MarshalIndent(manifest, "", "  ")
}
```

### Manifest Loading (Pseudo-code)

```go
func LoadManifests(pluginDirs []string) ([]ManifestData, error) {
    var manifests []ManifestData
    
    for _, dir := range pluginDirs {
        files, err := filepath.Glob(filepath.Join(dir, "*.json"))
        if err != nil {
            continue
        }
        
        for _, file := range files {
            data, err := ioutil.ReadFile(file)
            if err != nil {
                fmt.Printf("Error reading manifest %s: %v\n", file, err)
                continue
            }
            
            var manifest ManifestData
            if err := json.Unmarshal(data, &manifest); err != nil {
                fmt.Printf("Error parsing manifest %s: %v\n", file, err)
                continue
            }
            
            manifests = append(manifests, manifest)
        }
    }
    
    return manifests, nil
}
```

### Plugin Creation (Pseudo-code)

```go
func CreatePluginFromManifest(manifest ManifestData, hostPtr unsafe.Pointer) (unsafe.Pointer, error) {
    // Load the Go shared library
    libPath := filepath.Join(pluginDir, manifest.Build.GoSharedLibrary)
    lib, err := dlopen(libPath, RTLD_NOW)
    if err != nil {
        return nil, fmt.Errorf("failed to load library %s: %v", libPath, err)
    }
    
    // Get the entry point function
    createFunc, err := dlsym(lib, manifest.Build.EntryPoint)
    if err != nil {
        dlclose(lib)
        return nil, fmt.Errorf("entry point %s not found: %v", manifest.Build.EntryPoint, err)
    }
    
    // Call the entry point to create the plugin
    plugin := createFunc(hostPtr)
    if plugin == nil {
        dlclose(lib)
        return nil, fmt.Errorf("failed to create plugin instance")
    }
    
    // Create a handle and register for cleanup
    handle := cgo.NewHandle(plugin)
    registerHandle(handle, lib)
    
    return unsafe.Pointer(uintptr(handle)), nil
}
```

## Conclusion

The proposed manifest-based plugin discovery mechanism addresses the core architectural issue of separate Go runtimes having independent registries. By generating static metadata during the build process and loading it at runtime, we can ensure that the bridge has access to all plugin information without requiring complex runtime communication between separate Go programs.

This design aligns well with the target architecture described in ARCHITECTURE.md and provides a path toward a more robust and maintainable plugin system.