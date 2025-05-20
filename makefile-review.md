# Makefile Review Against Target Architecture

This document reviews the current Makefile implementation against the target architecture described in ARCHITECTURE.md, identifying alignment points and gaps.

## Current Build Process

The current Makefile implements a build process that:

1. Builds a shared Go bridge library (`libgoclap.so/dll/dylib`)
2. For each plugin:
   - Creates a build directory for the plugin
   - Compiles the Go plugin code to a shared library (`libplugin.so/dll/dylib`)
   - Compiles the C bridge code to object files
   - Links the object files with the Go shared library to create a `.clap` plugin
3. Provides installation targets to copy the plugins to the appropriate directory

## Alignment with Target Architecture

### Well-Aligned Aspects

1. **Separation of Go and C Components**
   - Go code is compiled to a shared library
   - C code provides the CLAP entry point
   - Linking combines them to form a complete plugin

2. **Cross-Platform Support**
   - Platform detection and appropriate file extensions
   - Platform-specific compilation flags
   - Appropriate installation paths for different OSes

3. **Plugin Installation**
   - Installs plugins to standard locations
   - Handles permissions correctly
   - Copies both the CLAP plugin and its dependencies

### Gaps and Misalignments

1. **Plugin-Specific C Skeleton Generation**
   - The target architecture calls for generating a plugin-specific C wrapper
   - Current Makefile uses the same C code (`src/c/bridge.c` and `src/c/plugin.c`) for all plugins
   - No code generation step exists to customize the C wrapper for each plugin

2. **Plugin/Plugin Directory Structure**
   - The target architecture suggests a `plugin/` subdirectory for C code within each plugin directory
   - Current Makefile doesn't enforce or create this structure
   - All plugins share C code from the global `src/c/` directory

3. **Development Workflow Support**
   - The target architecture outlines a specific development workflow
   - Current Makefile doesn't fully support steps like "Generate C Plugin Skeleton"
   - No scaffolding tools for new plugin creation

4. **Bridge Organization**
   - The target architecture defines the bridge as a component in `pkg/bridge/`
   - The Makefile builds this as a global shared library (`libgoclap.so/dll/dylib`)
   - The relationship between this global bridge and per-plugin builds isn't clearly defined

## Recommended Makefile Improvements

To better align with the target architecture, the following improvements are recommended:

1. **Add Code Generation Target**
   - Create a `generate` target that creates plugin-specific C code
   - Should take a plugin name and generate appropriate C wrapper files in a `plugin/` subdirectory
   - Connect this to the plugin build process

2. **Restructure Plugin Building**
   - Modify the build process to use plugin-specific C code when available
   - Support the intended directory structure with `plugin/` subdirectories
   - Generate missing files if they don't exist

3. **Improve Development Workflow Support**
   - Add targets for each step in the development workflow
   - Include a `new-plugin` target to scaffold new plugins
   - Provide clear documentation within the Makefile

4. **Clarify Bridge Integration**
   - Define whether the bridge is a standalone library or integrated into each plugin
   - If standalone, ensure plugins correctly link and find it at runtime
   - If integrated, adjust build process to embed necessary code

5. **Testing Enhancements**
   - Expand test targets to include validation against the CLAP specification
   - Add targets for benchmarking plugin performance
   - Include integration tests with common DAWs if possible

## Implementation Plan

1. **Short-term Fixes**
   - Reorganize the bridge code to use proper package naming
   - Remove redundant files (like the unused main.go)
   - Document the current plugin build flow better

2. **Medium-term Improvements**
   - Develop the code generation tools
   - Restructure the build process to follow the target architecture
   - Implement plugin scaffolding tools

3. **Long-term Goals**
   - Complete alignment with the target architecture
   - Comprehensive testing framework
   - Full support for all CLAP extensions

By implementing these recommendations, the build system will better align with the target architecture and provide a more streamlined experience for plugin developers.