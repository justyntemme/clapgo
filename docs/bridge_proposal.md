# Unified Bridge Implementation Proposal

## Overview

The ClapGo project currently has two separate C bridge implementations:
1. `internal/bridge/bridge.go`
2. `cmd/wrapper/main.go`

This duplication creates confusion and increases the risk of bugs. This proposal outlines a plan to consolidate these implementations into a single, robust bridge layer.

## Proposed Bridge Architecture

### Package Structure

```
internal/
  bridge/
    bridge.go         // Main bridge implementation
    cbridge.c         // C implementation of CLAP API
    cbridge.h         // C header for CLAP API
    handle_manager.go // Handle management utility
    event_mapper.go   // Event conversion utilities 
    audio_mapper.go   // Audio buffer conversion utilities
    extensions.go     // Extension handling
```

### Core Components

1. **Main Bridge (bridge.go)**
   - Contains CGO exports for all required CLAP functions
   - Coordinates with registry to find and create plugins
   - Performs necessary type conversions between C and Go

2. **Handle Manager (handle_manager.go)**
   - Thread-safe tracking of plugin instances
   - Proper cleanup to prevent memory leaks
   - Handle validation and error handling

3. **Audio and Event Mappers (audio_mapper.go, event_mapper.go)**
   - Efficient zero-copy conversion of audio buffers
   - Type-safe event handling
   - Performance optimization for real-time processing

4. **Extension Handling (extensions.go)**
   - Extension lookup and creation
   - Type-safe extension interfaces
   - Proper memory management for extensions

### Implementation Details

#### CGO Exports

The bridge will export the following functions via CGO:

```go
//export GetPluginCount
func GetPluginCount() C.uint32_t

//export GetPluginInfo
func GetPluginInfo(index C.uint32_t) *C.struct_clap_plugin_descriptor

//export CreatePlugin
func CreatePlugin(host *C.struct_clap_host, pluginID *C.char) unsafe.Pointer

//export GetVersion
func GetVersion(major, minor, patch *C.uint32_t) C.bool

// Plugin lifecycle functions
//export GoInit
func GoInit(plugin unsafe.Pointer) C.bool

//export GoDestroy
func GoDestroy(plugin unsafe.Pointer)

//export GoActivate
func GoActivate(plugin unsafe.Pointer, sampleRate C.double, minFrames, maxFrames C.uint32_t) C.bool

//export GoDeactivate
func GoDeactivate(plugin unsafe.Pointer)

//export GoStartProcessing
func GoStartProcessing(plugin unsafe.Pointer) C.bool

//export GoStopProcessing
func GoStopProcessing(plugin unsafe.Pointer)

//export GoReset
func GoReset(plugin unsafe.Pointer)

//export GoProcess
func GoProcess(plugin unsafe.Pointer, process *C.struct_clap_process) C.int32_t

//export GoGetExtension
func GoGetExtension(plugin unsafe.Pointer, id *C.char) unsafe.Pointer

//export GoOnMainThread
func GoOnMainThread(plugin unsafe.Pointer)
```

#### Handle Management

The handle manager will provide these key functions:

```go
// RegisterHandle adds a new plugin handle to the registry
func RegisterHandle(plugin api.Plugin) unsafe.Pointer

// GetPluginFromPtr retrieves a plugin from a pointer
func GetPluginFromPtr(ptr unsafe.Pointer) api.Plugin

// DeleteHandle removes a handle and performs cleanup
func DeleteHandle(ptr unsafe.Pointer)

// CleanupAllHandles performs cleanup of all handles
func CleanupAllHandles()
```

#### C Bridge Implementation

The C implementation will provide the entry point for CLAP hosts:

```c
// cbridge.c

#include <stdlib.h>
#include <string.h>
#include "cbridge.h"
#include "clap/clap.h"

// CLAP entry point
const struct clap_plugin_entry clap_plugin_entry = {
    .clap_version = CLAP_VERSION_INIT,
    .init = clap_entry_init,
    .deinit = clap_entry_deinit,
    .get_factory = clap_entry_get_factory
};

// Plugin factory implementation
static const struct clap_plugin_factory plugin_factory = {
    .get_plugin_count = plugin_factory_get_plugin_count,
    .get_plugin_descriptor = plugin_factory_get_plugin_descriptor,
    .create_plugin = plugin_factory_create_plugin
};

// Entry point initialization
bool clap_entry_init(const char *plugin_path) {
    // Initialize any global resources
    return true;
}

// Entry point deinitialization
void clap_entry_deinit(void) {
    // Clean up any global resources
}

// Factory retrieval
const void *clap_entry_get_factory(const char *factory_id) {
    if (strcmp(factory_id, CLAP_PLUGIN_FACTORY_ID) == 0) {
        return &plugin_factory;
    }
    return NULL;
}

// Implementation of factory methods that call into Go code
...
```

## Memory Safety Measures

1. **Consistent Handle Management**
   - All plugin instances tracked in a central registry
   - Proper cleanup in case of errors
   - Thread-safe access to handles

2. **Resource Lifetime Management**
   - Clear ownership rules for all allocated resources
   - Automatic cleanup during plugin destruction
   - Tracking of C strings and other allocated memory

3. **Error Handling**
   - Consistent error reporting
   - Graceful fallback in case of errors
   - Logging of error conditions

4. **Thread Safety**
   - Proper locking around shared resources
   - Clear documentation of thread expectations
   - Thread-safe plugin enumeration and creation

## Implementation Plan

1. **Initial Setup**
   - Create new package structure
   - Set up basic bridge implementation

2. **Core Functionality**
   - Implement handle management
   - Add CGO exports for core functions
   - Set up C bridge implementation

3. **Plugin Lifecycle**
   - Implement plugin initialization and destruction
   - Add activation/deactivation support
   - Implement processing functionality

4. **Extension Support**
   - Add extension lookup and management
   - Implement commonly used extensions
   - Ensure proper memory handling for extensions

5. **Testing and Validation**
   - Create comprehensive tests
   - Validate with different plugin types
   - Check for memory leaks and thread safety issues

6. **Documentation and Examples**
   - Document bridge usage
   - Update example code
   - Provide migration guide for existing code

## Migration Strategy

1. **Identify Current Usage**
   - Determine which bridge implementation is used where
   - Document dependencies on each implementation

2. **Gradual Transition**
   - First implement new bridge alongside existing ones
   - Move functionality one piece at a time
   - Verify each step with tests

3. **Remove Deprecated Bridges**
   - Once all functionality is in the new bridge, remove old implementations
   - Update any remaining references
   - Verify with integration tests

## Conclusion

A unified bridge implementation will significantly improve the ClapGo codebase by eliminating duplication, improving memory safety, and creating clear separation of concerns. This proposal provides a path forward that maintains compatibility while improving the overall architecture.