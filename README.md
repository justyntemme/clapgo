# ClapGo

A framework for building [CLAP](https://github.com/free-audio/clap) audio plugins using Go.

## Features

- Build CLAP-compatible audio plugins using Go
- Seamless C-Go interoperability via CGo
- Optional GUI support using Qt/QML
- Easy build and installation process

## Architecture

ClapGo uses CGo to create a bridge between:

1. A minimal C host that implements the CLAP plugin interface
2. Go code that provides the actual plugin functionality

This allows audio plugin developers to write CLAP plugins in Go while maintaining compatibility with CLAP hosts.

## Project Structure

```
.
├── build            # Build artifacts
├── cmake            # CMake helper scripts
├── examples         # Example plugins
│   ├── gain         # Simple gain plugin
│   └── gain-with-gui # Gain plugin with GUI support
├── include
│   └── clap         # CLAP API headers (submodule)
├── src
│   ├── c            # C plugin wrapper
│   └── goclap       # Go CLAP interface package
```

## Building a Plugin

We use CMake to build our plugins.

### Building Plugins

```bash
# Configure and build using our install script
./install.sh         # Install for current user
./install.sh --system # Install system-wide (requires sudo)
./install.sh --gui   # Build with GUI support (requires Qt6)

# Or use CMake directly
cmake -B build
cmake --build build
```

### Creating a New Plugin

To create a new plugin:

1. Create a new directory under `examples/` with your plugin name
2. Copy the structure from one of the existing examples
3. Update Go package name, plugin ID, and metadata in your implementation

New plugins in the examples directory are automatically detected and built by our build system.

## Example Plugins

- **Gain**: A simple gain plugin example in `examples/gain`
- **Gain with GUI**: A gain plugin with GUI support in `examples/gain-with-gui`

## Installation

The installation script provides a convenient way to build and install plugins:

```bash
# Install to user directory (~/.clap)
./install.sh

# Install system-wide (/usr/lib/clap)
./install.sh --system

# Build with GUI support
./install.sh --gui
```

This will copy the `.clap` files to the appropriate location for your platform:

- Linux: `~/.clap` (user) or `/usr/lib/clap` (system-wide)
- Windows: `%APPDATA%\CLAP` (user) or `C:\Program Files\Common Files\CLAP` (system-wide)
- macOS: `~/Library/Audio/Plug-Ins/CLAP` (user) or `/Library/Audio/Plug-Ins/CLAP` (system-wide)

## Documentation

See the [documentation](docs/README.md) for detailed usage instructions, architecture overview, API reference, and implementation details.

## License

MIT License

