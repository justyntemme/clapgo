# ClapGo Implementation Roadmap

This roadmap outlines the implementation tasks needed to complete the ClapGo project, prioritized from simplest to most complex. Tasks are organized to tackle low-hanging fruit first and focus on resolving skipped tests.

## 1. Host Callback Integration ✅

**Priority: High (Low-hanging fruit)**

Implement C wrapper functions for host callback integration in `src/goclap/hostinfo.go`.

- **Tasks:**
  - ✅ Create C wrapper functions for `RequestRestart` method
  - ✅ Create C wrapper functions for `RequestProcess` method
  - ✅ Create C wrapper functions for `RequestCallback` method
  - ✅ Create C wrapper functions for `GetExtension` method
  - ✅ Add proper error handling and logging

- **Current Status:**
  - ✅ Implemented C wrapper functions in the cgo preamble
  - ✅ Added error handling with fmt.Println for warnings
  - ✅ Updated method signatures to return bool for success/failure
  - ✅ Proper memory management for C strings

## 2. Memory Management for Descriptors ✅

**Priority: High (Low-hanging fruit)**

Implement proper memory management for plugin descriptors in `src/c/plugin.c`.

- **Tasks:**
  - ✅ Add code to free each descriptor in the `clapgo_deinit` function
  - ✅ Ensure descriptors are properly allocated and tracked
  - ✅ Implement memory cleanup to prevent leaks

- **Current Status:**
  - ✅ Implemented a tracking system for dynamically allocated descriptor fields
  - ✅ Added deep copy functionality for descriptors to ensure proper memory ownership
  - ✅ Added comprehensive cleanup in `clapgo_deinit` to free all resources
  - ✅ Added error handling for allocation failures

## 3. Plugin Implementation ✅

**Priority: High (Resolves skipped tests)**

Implement a complete version of simplified code in `src/goclap/plugin.go`.

- **Tasks:**
  - ✅ Restructure plugin implementation to avoid duplicate exported symbols
  - ✅ Set up implementation functions for plugin handling in goclap package
  - ✅ Connect wrapper functions to implementation functions
  - Complete the `GetPluginInfo` function to properly return plugin descriptors
  - Implement the `CreatePlugin` function to instantiate plugins
  - Implement the `getProcessorFromPtr` function to retrieve Go processors from C plugin pointers
  - Enhance audio buffer handling in the `GoProcess` function

- **Current Status:**
  - ✅ Basic structure has been implemented and builds successfully
  - ✅ Plugin implementation structure has been reorganized to avoid conflicts
  - ✅ Wrapper layer has been implemented to connect to Go implementation
  - Remaining implementation needed for plugin descriptors and instance creation

## 4. Shared Library Loading ✅

**Priority: Medium-High**

Implement shared library loading for Go code in `src/c/plugin.c`.

- **Tasks:**
  - ✅ Implement proper loading of the Go shared library in `clapgo_init`
  - ✅ Add error handling for library loading failures
  - ✅ Add version compatibility checks

- **Current Status:**
  - ✅ Implemented platform-specific shared library loading for Windows, macOS, and Linux
  - ✅ Added robust error handling for library loading failures
  - ✅ Implemented version compatibility checking between C and Go components
  - ✅ Added intelligent library path search to find the Go shared library in multiple locations
  - ✅ Set up function pointers to call into Go code from C
  - ✅ Updated all plugin instance functions to use the loaded function pointers

## 5. CGO Function Calls

**Priority: Medium-High**

Implement Go function calls via CGO in `src/c/plugin.c`.

- **Tasks:**
  - Implement CGO calls for `clapgo_get_plugin_count`
  - Implement CGO calls for `clapgo_get_plugin_descriptor`
  - Implement CGO calls for plugin lifecycle functions (init, destroy, activate, etc.)
  - Implement CGO calls for audio processing

- **Current Status:**
  - All functions have placeholders with printf statements
  - No actual CGO integration is implemented

## 6. Preset Discovery Support

**Priority: Medium (Resolves multiple skipped tests)**

Implement preset discovery factory support to resolve several skipped tests.

- **Tasks:**
  - Implement the 'clap.preset-discovery-factory/draft-2' factory
  - Add preset discovery crawler functionality
  - Add preset loading functionality
  - Ensure descriptor consistency

- **Current Status:**
  - Currently skipped in tests with message "The plugin does not implement the 'clap.preset-discovery-factory/draft-2' factory"
  - No implementation exists yet

## 7. Plugin ID Validation

**Priority: Medium (Resolves skipped test)**

Implement proper plugin ID validation to handle IDs with trailing garbage.

- **Tasks:**
  - Enhance the plugin ID validation logic in CreatePlugin
  - Add test for attempted creation with invalid ID

- **Current Status:**
  - Currently skipped in tests with message "The plugin library does not expose any plugins"
  - Basic plugin loading exists but no validation for malformed IDs

## 8. GUI Integration

**Priority: Low (Complex feature)**

Implement Qt/QML GUI initialization in `examples/gain-with-gui/gui_bridge.cpp`.

- **Tasks:**
  - Complete the GUI initialization code for the gain-with-gui example
  - Implement proper Qt/QML initialization
  - Connect GUI to plugin parameters
  - Handle window management across platforms

- **Current Status:**
  - Basic structure exists with stub implementations
  - Comments indicate where real Qt/QML initialization should occur

## Additional Considerations

- **Build System Enhancement:** Ensure the build system correctly handles CGO with the CLAP API
- **Documentation:** Update documentation to reflect implementation details
- **Testing:** Add unit tests for core functionality
- **Examples:** Expand the example plugins to demonstrate more features

By following this roadmap, the ClapGo project can systematically address the current gaps in implementation, starting with the simplest tasks and gradually moving to more complex features.

