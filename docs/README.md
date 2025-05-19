# ClapGo Documentation

Welcome to the ClapGo documentation. This guide will help you understand how to build CLAP audio plugins using Go.

## Contents

### Getting Started
- [Build Guide](build-guide.md) - How to build and create plugins
- [Usage Guide](usage.md) - Getting started with ClapGo

### Architecture
- [Architecture Overview](architecture/overview.md) - High-level architecture explanation
- [CMake Integration](cmake-integration.md) - Details on CMake build system integration

### Reference
- [API Reference](api-reference.md) - ClapGo API details
- [Implementation Guide](implementation-guide.md) - Guide for implementing remaining features
- [Roadmap](roadmap.md) - Future development plans

## What is CLAP?

CLAP (CLever Audio Plugin) is an open audio plugin standard designed to address the shortcomings of existing audio plugin formats. It offers features like:

- Sample-accurate automation
- Plugin-to-host communication
- Better performance
- Support for modularity

Learn more at the [CLAP website](https://cleveraudio.org).

## What is ClapGo?

ClapGo is a framework that enables audio plugin development using the Go programming language, targeting the CLAP standard. It provides:

1. A Go API for audio processing
2. CGo-based bridge to the C CLAP interface
3. Build tools to produce compatible plugins

## Core Concepts

1. **Plugin Implementation**: Create a Go struct that implements the `AudioProcessor` interface
2. **Registration**: Register your plugin with metadata through `RegisterPlugin()`  
3. **Processing**: Process audio samples and handle events in your `Process()` method
4. **Parameters**: Define parameters for your plugin with the `ParamManager`
5. **Building**: Compile to a `.clap` format file that hosts can load

## Examples

Check the `examples/` directory for working plugin examples:

1. **Gain**: A simple gain plugin (`examples/gain/`)
2. **Others**: More examples coming soon

## Development Status

ClapGo is currently in active development. Key components that are implemented:

- ✅ Basic plugin infrastructure
- ✅ Parameter registration
- ✅ Event handling skeleton
- ✅ Build system

Components that still need work:

- ⚠️ Complete audio buffer handling
- ⚠️ Full event system implementation
- ⚠️ Extension support
- ⚠️ GUI integration

See the [Implementation Guide](implementation-guide.md) for details on upcoming development.