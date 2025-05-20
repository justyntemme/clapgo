# CLAPGO Architecture

## Overview

CLAPGO is a framework that enables audio plugin development in Go using the CLAP (CLever Audio Plugin) standard. The architecture bridges the gap between the C-based CLAP API and Go programming language, allowing developers to build audio plugins in Go that can be loaded into any CLAP-compatible Digital Audio Workstation (DAW) like FL Studio, Bitwig, etc.

## Architectural Goals

The main architectural goal of CLAPGO is to facilitate the development of CLAP plugins in Go by providing:

1. **Seamless C-Go Bridge**: A robust bridge that handles all the complexity of interfacing between C and Go code.
2. **Minimal C Boilerplate**: Developers should be able to write primarily in Go with minimal C code.
3. **Runtime Safety**: Robust memory management to prevent leaks or crashes.
4. **Simple API**: An intuitive API that abstracts away the complexity of the CLAP specification.
5. **Optimized Performance**: Ensuring minimal overhead for audio processing.

## Core Components

The architecture consists of the following key components:

### 1. C Plugin Skeleton (`src/c/`)

A reusable C code skeleton that serves as the entry point for the host DAW. This component:

- Implements the CLAP entry point and plugin lifecycle functions 
- Loads the Go shared library at runtime
- Routes CLAP function calls to the appropriate Go code
- Handles memory management for C/Go data exchange

### 2. Go Plugin API (`pkg/api/`)

A set of Go interfaces and types that define the plugin API. Plugins implement these interfaces to:

- Provide plugin metadata
- Process audio
- Handle events (MIDI, parameter changes, etc.)
- Support extensions (GUI, state management, etc.)

### 3. Go-C Bridge (`pkg/bridge/`)

The Go side of the bridge that:

- Exports Go functions to be called from C
- Handles type conversions between C and Go
- Manages object lifetimes and memory
- Provides utilities for plugin developers

### 4. Plugin Registry (`pkg/registry/`)

A central registry that:

- Allows plugins to register themselves at runtime
- Creates plugin instances on demand
- Maps plugin IDs to implementation code

## Plugin Development Workflow

The intended workflow for developing a CLAP plugin with CLAPGO is:

1. **Create a Go Project**: Implement a Go plugin by implementing the required interfaces.
2. **Compile to Shared Library**: Build the Go code as a shared library (`.so`, `.dylib`, or `.dll`).
3. **Generate C Plugin Skeleton**: Use the CLAPGO tools to generate a thin C wrapper.
4. **Compile Final Plugin**: Build the C wrapper and link it with the Go shared library to create the final CLAP plugin.
5. **Deploy**: Place the resulting plugin in the appropriate directory for your DAW to discover.

## Example Project Structure

```
examples/plugin-name/
├── main.go              # Go plugin implementation
├── provider.go          # Plugin provider for registration
├── constants.go         # Plugin-specific constants
└── plugin/              # C skeleton code
    ├── plugin.c         # Generated C entry point
    └── plugin.h         # C header with plugin information
```

## Runtime Architecture

When loaded into a DAW, the following sequence occurs:

1. The DAW loads the CLAP plugin (the C skeleton).
2. The C skeleton loads the Go shared library.
3. The C skeleton initializes the Go runtime and calls into the bridge.
4. The Go bridge creates an instance of the plugin via the registry.
5. Function calls from the DAW are routed through the C skeleton to the appropriate Go code.
6. Audio processing and event handling happens in the Go implementation.
7. Results are passed back through the bridge to the DAW.

This architecture allows developers to focus on implementing the audio processing and creative aspects of their plugins in Go, while the framework handles the complexities of the CLAP API and C/Go interoperability.