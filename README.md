# ClapGo

A framework for building [CLAP](https://github.com/free-audio/clap) audio plugins using Go.

## Architecture

ClapGo uses CGo to create a bridge between:

1. A minimal C host that implements the CLAP plugin interface
2. Go code that provides the actual plugin functionality

This allows audio plugin developers to write CLAP plugins in Go while maintaining compatibility with CLAP hosts.

## Project Structure

```
.
├── dist            # Build artifacts
├── docs             # Documentation
├── examples         # Example plugins
├── include
│   └── clap         # CLAP API headers (submodule)
├── src
│   ├── c            # C plugin wrapper
│   └── goclap       # Go CLAP interface package
```

## Building a Plugin

We use the CLAP build system based on CMake to build our plugins.

### Building Plugins

```bash
# Configure and build using our build script
./build.sh

# Additional build options
./build.sh --debug   # Build in debug mode
./build.sh --clean   # Clean build before building
./build.sh --test    # Run tests after building

# Or use CMake directly
cmake --preset linux  # or macos, windows
cmake --build --preset linux
```

### Creating a New Plugin

To create a new plugin, use our helper script:

```bash
./create_plugin.sh myplugin
```

This will:
1. Create a new directory under `examples/` with your plugin name
2. Copy the structure from the gain example
3. Update classes and identifiers appropriately

New plugins in the examples directory are automatically detected and built by CMake.

## Example Plugin

See the `examples/gain` directory for a simple gain plugin example.

## Testing and Installation

After building, you can test if the plugin is valid:

```bash
./test_plugin.sh
```

This will check the dynamic dependencies and verify the CLAP entry point.

To install the plugins to your system's plugin directories:

```bash
./build.sh --install
```

This will copy the `.clap` files to the appropriate location for your platform:

- Linux: `~/.clap` or `/usr/lib/clap`
- Windows: `%COMMONPROGRAMFILES%\CLAP` or `%LOCALAPPDATA%\Programs\Common\CLAP`
- macOS: `/Library/Audio/Plug-Ins/CLAP` or `~/Library/Audio/Plug-Ins/CLAP`

## Documentation

See the [documentation](docs/README.md) for detailed usage instructions, architecture overview, API reference, and implementation details.

## License

MIT License

