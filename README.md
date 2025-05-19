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
├── cmd
│   └── build        # Build scripts
├── docs             # Documentation
├── examples         # Example plugins
├── include
│   └── clap         # CLAP API headers (submodule)
├── src
│   ├── c            # C plugin wrapper
│   └── goclap       # Go CLAP interface package
```

## Building a Plugin

1. Create a new Go plugin in the `examples` directory
2. Implement the `goclap.AudioProcessor` interface
3. Register your plugin using `goclap.RegisterPlugin()`
4. Use the build tool to compile and package your plugin:

```bash
make examples
```

## Example Plugin

See the `examples/gain` directory for a simple gain plugin example.

## Installation

After building, copy the `.clap` files from the `dist` directory to your DAW's plugin directory:

- Linux: `~/.clap` or `/usr/lib/clap`
- Windows: `%COMMONPROGRAMFILES%\CLAP` or `%LOCALAPPDATA%\Programs\Common\CLAP`
- macOS: `/Library/Audio/Plug-Ins/CLAP` or `~/Library/Audio/Plug-Ins/CLAP`

## Usage

See the [documentation](docs/usage.md) for detailed usage instructions.

## License

MIT License