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

The easiest way to create a new plugin is to use the Makefile's `new` target:

```bash
make new NAME=myplugin
```

This will:

1. Create a new plugin directory in `examples/myplugin`
2. Set up a template plugin based on the gain example
3. Configure the Makefile to build your plugin

Then you can build your plugin with:

```bash
make myplugin
```

Alternatively, you can manually:

1. Create a new Go plugin in the `examples` directory
2. Implement the `goclap.AudioProcessor` interface
3. Register your plugin using `goclap.RegisterPlugin()`
4. Build with `make yourplugin`

## Example Plugin

See the `examples/gain` directory for a simple gain plugin example.

## Installation

After building, copy the `.clap` files from the `dist` directory to your DAW's plugin directory:

- Linux: `~/.clap` or `/usr/lib/clap`
- Windows: `%COMMONPROGRAMFILES%\CLAP` or `%LOCALAPPDATA%\Programs\Common\CLAP`
- macOS: `/Library/Audio/Plug-Ins/CLAP` or `~/Library/Audio/Plug-Ins/CLAP`

## Documentation

See the [documentation](docs/README.md) for detailed usage instructions, architecture overview, API reference, and implementation details.

## License

MIT License

