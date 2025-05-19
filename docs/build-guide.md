# ClapGo Build Guide

This document explains how to build CLAP audio plugins using the ClapGo framework.

## Prerequisites

- Go 1.18+ (1.21+ recommended)
- GCC or compatible C compiler
- CLAP headers (included as a submodule)

## Building Existing Plugins

### Building All Plugins

To build all example plugins:

```bash
make
```

This will compile all plugins in the `examples/` directory and place the built `.clap` files in the `dist/` directory.

### Building a Specific Plugin

To build a specific plugin:

```bash
make [plugin-name]
```

For example, to build the gain example:

```bash
make gain
```

### Installing Plugins

To install the built plugins to your system's plugin directory:

```bash
make install
```

This will copy the `.clap` files to the appropriate location for your platform:

- **Linux**: `~/.clap`
- **macOS**: `~/Library/Audio/Plug-Ins/CLAP`
- **Windows**: `%APPDATA%\CLAP`

## Creating a New Plugin

### Using the Template Generator

The easiest way to create a new plugin is to use the built-in template generator:

```bash
make new NAME=myplugin
```

This will:
1. Create a new plugin directory at `examples/myplugin/`
2. Set up a template based on the gain example plugin
3. Rename classes and identifiers appropriately

You can then build your new plugin with:

```bash
make myplugin
```

### Manual Plugin Creation

To create a plugin manually:

1. Create a new directory in `examples/` for your plugin
2. Create a `main.go` file that:
   - Implements the `goclap.AudioProcessor` interface
   - Registers the plugin using `goclap.RegisterPlugin()`
   - Exports any required functions with `//export`

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