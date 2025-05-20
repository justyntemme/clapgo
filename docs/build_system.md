# ClapGo Build System

This document outlines the new build system for ClapGo, which provides a clean and consistent way to build CLAP plugins using Go.

## Architecture Overview

The build system is designed to compile Go code into a shared library that is dynamically loaded by a C wrapper, which implements the CLAP plugin interface. This approach allows for better separation of concerns and makes it easier to maintain and extend the codebase.

### Key Components

1. **Root Makefile** (`/Makefile`) - Coordinates the build process for the entire project, including the Go shared library and individual plugins.

2. **Template Makefile** (`/examples/Makefile.template`) - Provides a consistent build configuration for all example plugins.

3. **Plugin Makefiles** (`/examples/*/Makefile`) - Contain plugin-specific configuration and inherit from the template.

4. **C Bridge** (`/src/c/bridge.c`, `/src/c/bridge.h`) - Implements dynamic loading of the Go shared library and provides CLAP plugin interface.

5. **CLAP Entry Point** (`/src/c/plugin.c`) - Implements the CLAP entry point for the plugin.

6. **Go Bridge** (`/cmd/goclap/main.go`) - Exports functions that the C bridge calls to interact with the Go implementation.

### Build Flow

The build process works as follows:

1. The root Makefile first builds the Go shared library (`libgoclap.so`), which contains the core functionality.

2. Then it builds each plugin using the plugin-specific Makefile.

3. Each plugin is compiled into a `.clap` file that contains the C wrapper code, which dynamically loads the shared library.

4. When installed, both the `.clap` files and the shared library are placed in the `~/.clap/` directory.

## Usage

### Building All Plugins

```bash
make
```

This will build the Go shared library and all plugins.

### Building a Specific Plugin

```bash
cd examples/gain
make
```

This will build the specified plugin.

### Installing Plugins

```bash
make install
```

This will install all plugins and the Go shared library to `~/.clap/`.

### Cleaning Build Artifacts

```bash
make clean
```

This will remove all build artifacts.

## Creating a New Plugin

To create a new plugin:

1. Create a new directory under `examples/` with your plugin name.
2. Create a `main.go` file that implements the `api.Plugin` interface.
3. Create a `Makefile` that includes the template and sets the plugin name and ID.
4. Build and install your plugin with `make` and `make install`.

Example plugin structure:

```
examples/my-plugin/
├── main.go       # Plugin implementation
└── Makefile      # Plugin-specific Makefile
```

Example Makefile:

```makefile
# Makefile for My Plugin

# Include the template from the examples directory
include ../Makefile.template

# Plugin-specific configuration
PLUGIN_NAME := my-plugin
PLUGIN_ID := com.clapgo.my-plugin

# Additional flags
CFLAGS += -DCLAPGO_PLUGIN_ID=\"$(PLUGIN_ID)\"

# Default target
all: $(PLUGIN_CLAP)
```

## Plugin ID Management

The plugin ID is passed through the build system in several ways:

1. In the plugin's `Makefile`, the `PLUGIN_ID` variable is set.
2. It is passed to the C compiler with the `-DCLAPGO_PLUGIN_ID` flag.
3. During Go compilation, it's passed via the `CLAPGO_PLUGIN_ID` environment variable.
4. The plugin's Go code should implement the `GetPluginID()` method to return this ID.

This ensures that the plugin ID is consistent across all components and makes it easy to create multiple plugins without conflicts.

## Dynamic Loading

The C wrapper uses dynamic loading to find and load the Go shared library at runtime. It searches for the library in several locations:

1. The same directory as the plugin
2. Parent directory of the plugin
3. `~/.clap/` directory
4. System library directories

This approach allows for flexibility in deployment and makes it easier to update the library without recompiling the plugins.

## Conclusion

The new build system provides a clean, consistent, and maintainable way to build CLAP plugins using Go. By centralizing the build process and using makefiles throughout, it's easier to understand, extend, and maintain the codebase. The clear separation between the C wrapper and Go implementation makes it easier to develop new plugins and ensures that the code remains maintainable in the long term.