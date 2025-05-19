# ClapGo Build Guide

This document explains how to build CLAP audio plugins using the ClapGo framework.

## Prerequisites

- Go 1.18+ (1.21+ recommended)
- CMake 3.17+
- GCC/Clang or compatible C compiler (MSVC on Windows)
- Ninja build system (recommended)
- CLAP headers (included as a submodule)

## Building Overview

ClapGo uses the CLAP project's build system to create plugins. At a high level, the build process:

1. Compiles Go code into a shared library for each plugin
2. Links this library with CLAP C wrapper code
3. Packages everything according to platform requirements

## Building Plugins

### Using the Build Script

The easiest way to build all plugins is to use our build script:

```bash
./build.sh
```

This will compile all plugins in the `examples/` directory and place the built `.clap` files in the `build/[preset]` directory.

For a debug build:

```bash
./build.sh --debug
```

### Using CMake Directly

You can also use CMake commands directly:

```bash
# Configure
cmake --preset linux  # or macos, windows

# Build
cmake --build --preset linux
```

### Building a Specific Plugin

When using CMake, all plugins are built by default. You can focus on a specific target:

```bash
cmake --build --preset linux --target gain
```

## Installing Plugins

To install the built plugins to your system's plugin directory:

```bash
./build.sh --install
```

This will copy the `.clap` files to the appropriate location for your platform:

- **Linux**: `~/.clap`
- **macOS**: `~/Library/Audio/Plug-Ins/CLAP`
- **Windows**: `%APPDATA%\CLAP`

You can also use the CMake install target directly:

```bash
cmake --build --preset linux --target install
```

## Creating a New Plugin

### Using the Template Generator

We provide a script for generating new plugins:

```bash
./create_plugin.sh myplugin
```

This will:
1. Create a new plugin directory in `examples/myplugin`
2. Copy template files from the gain example
3. Update class and function names appropriately

After creating the plugin, customize the implementation in `examples/myplugin/main.go` and build it with `./build.sh`.

### Manual Creation

Alternatively, you can manually:

1. Create a new directory in `examples/` with your plugin name
2. Create a `main.go` file that implements the `goclap.AudioProcessor` interface
3. Register your plugin with `goclap.RegisterPlugin()`

All plugins in the examples directory are automatically detected and built by CMake.

## Plugin Structure

A typical plugin consists of:

```go
package main

// #cgo CFLAGS: -I../../include/clap/include
// #include "../../include/clap/include/clap/clap.h"
// #include <stdlib.h>
import "C"
import (
    "github.com/yourusername/clapgo/src/goclap"
    "unsafe"
)

// MyPlugin implements the AudioProcessor interface
type MyPlugin struct {
    // Plugin state
}

// Implement all required methods...

func init() {
    // Register the plugin
    info := goclap.PluginInfo{
        ID:          "com.example.myplugin",
        Name:        "My Plugin",
        // ...other metadata
    }
    
    plugin := NewMyPlugin()
    goclap.RegisterPlugin(info, plugin)
}

// This export is needed to avoid name conflicts
//export MyPluginGetPluginCount
func MyPluginGetPluginCount() C.uint32_t {
    return 1
}

func main() {
    // Not used when built as a plugin
}
```

## Build Process Details

The build process consists of several steps:

1. **Go Build**: Compiles your Go code into a shared library
   ```
   go build -buildmode=c-shared -o dist/libmyplugin.so ./examples/myplugin
   ```

2. **C Compilation**: Compiles the C wrapper code
   ```
   gcc -shared -fPIC -o dist/myplugin.so src/c/plugin.c -Ldist -lmyplugin
   ```

3. **Packaging**: Creates the final `.clap` file
   ```
   cp dist/myplugin.so dist/myplugin.clap
   ```

On macOS, additional steps create a proper bundle structure.

## Debugging

### Debug Builds

To build with debug symbols:

```bash
make gain DEBUG=1
```

This will:
- Disable Go optimizations
- Include debug symbols
- Disable stripping

### Testing in a DAW

To test your plugin in a DAW:

1. Build the plugin: `make your-plugin`
2. Install it to your system: `make install` 
   - Alternatively, manually copy the `.clap` file to your DAW's plugin directory
3. Scan for new plugins in your DAW
4. Your plugin should appear under the CLAP category

### Common Issues

- **Plugin not appearing in DAW**: Ensure the plugin is properly installed and the DAW has scanned for new plugins
- **Build errors**: Check for missing dependencies or compiler issues
- **Runtime crashes**: Use debug builds and a debugger to trace the issue

## Cross-Platform Considerations

### macOS

macOS requires a specific bundle structure for plugins. The build system handles this automatically.

### Windows

On Windows, you'll need MinGW or MSVC to compile the C code. Adjust `Makefile` settings as needed.

### Linux

On Linux, ensure you have the required development packages installed:

```bash
# For Debian/Ubuntu
sudo apt-get install build-essential

# For Fedora
sudo dnf install gcc make
```