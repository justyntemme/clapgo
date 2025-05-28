# ClapGo Build System Documentation

This document provides a comprehensive overview of the ClapGo build system, detailing how plugins are built, linked, and installed using the Makefile-based approach.

## Overview

ClapGo uses a Makefile-based build system that compiles Go code into shared libraries and links them with C bridge code to create CLAP plugins. The system supports cross-platform builds for Linux, macOS, and Windows.

## Build Architecture

### Component Overview

```
┌─────────────────┐     ┌──────────────┐     ┌─────────────────┐
│   Go Plugin     │────▶│  Go Shared   │────▶│   C Bridge      │
│   Source Code   │     │   Library    │     │   Wrapper       │
└─────────────────┘     └──────────────┘     └─────────────────┘
                               │                      │
                               └──────────┬───────────┘
                                          │
                                          ▼
                                   ┌─────────────┐
                                   │ CLAP Plugin │
                                   │   (.clap)   │
                                   └─────────────┘
```

### Build Process Flow

1. **Go Compilation**: Plugin Go code is compiled into a shared library (`.so`, `.dylib`, or `.dll`)
2. **C Bridge Compilation**: C bridge files are compiled into object files
3. **Linking**: C objects are linked with the Go shared library to create the final CLAP plugin

## Platform Detection

The Makefile automatically detects the platform and configures appropriate settings:

```makefile
UNAME := $(shell uname)
ifeq ($(UNAME), Darwin)
    PLATFORM := macos
    SO_EXT := dylib
    CLAP_FORMAT := bundle
else ifeq ($(UNAME), Linux)
    PLATFORM := linux
    SO_EXT := so
    CLAP_FORMAT := so
else
    PLATFORM := windows
    SO_EXT := dll
    CLAP_FORMAT := dll
endif
```

## Build Configuration

### Go Build Settings

- **Build Mode**: `-buildmode=c-shared` - Creates a shared library from Go code
- **Optimizations**: `-ldflags="-s -w"` - Strips debug info and reduces binary size
- **Debug Mode**: When `DEBUG=1`, uses `-gcflags="all=-N -l"` to disable optimizations

### C Compilation Settings

- **Include Paths**: `-I./include/clap/include` - CLAP header files
- **Position Independent Code**: `-fPIC` - Required for shared libraries
- **JSON Support**: Uses `pkg-config` to find json-c library for manifest parsing
- **Platform-specific RPATH**: On Linux, sets `RPATH` to `$ORIGIN` for relative library loading

## Module Build Process

### 1. Building the Go Bridge Library

```bash
make build-go
```

This target builds the core Go bridge library that provides common functionality:
- Creates `build/libgoclap.so` containing shared bridge code
- Used by all plugins as a common runtime

### 2. Building Individual Plugins

Each plugin follows a standardized build process:

#### Step 1: Go Shared Library
```bash
CGO_ENABLED=1 go build -buildmode=c-shared -o build/lib<plugin>.so *.go
```
- Compiles plugin Go code into a shared library
- Exports C-compatible functions for CLAP integration
- Copies the manifest JSON file to the build directory

#### Step 2: C Bridge Object Files
Four C source files are compiled into object files:

1. **bridge.o**: Main C-to-Go bridge implementation
   - Handles function calls between C and Go
   - Manages memory safety and type conversions

2. **plugin.o**: CLAP plugin interface implementation
   - Implements the core CLAP plugin structure
   - Routes CLAP callbacks to Go functions

3. **manifest.o**: Plugin metadata handling
   - Reads plugin information from JSON manifest
   - Provides plugin discovery information

4. **preset_discovery.o**: Preset discovery and loading
   - Implements CLAP preset discovery factory
   - Loads presets from filesystem

#### Step 3: Final Linking
```bash
gcc -shared -o <plugin>.clap \
    bridge.o plugin.o manifest.o preset_discovery.o \
    -L<build_dir> -l<plugin> -ljson-c
```
- Links all C objects with the Go shared library
- Creates the final `.clap` plugin file
- Uses `-rpath` on Linux to ensure library loading

## Linking Strategy

### Dynamic Library Loading

The linking strategy ensures plugins can find their dependencies:

1. **Go Shared Library**: Each plugin has its own Go shared library (`lib<plugin>.so`)
2. **Runtime Path**: On Linux, `RPATH` is set to `$ORIGIN` so the plugin finds its library in the same directory
3. **JSON-C Dependency**: Dynamically linked for manifest parsing

### Symbol Resolution

- Go functions are exported with C linkage using `//export` directives
- C bridge calls these exported functions through function pointers
- All plugin-specific functionality is contained in the Go shared library

## Installation Process

```bash
make install
```

The installation process creates a standardized directory structure:

```
~/.clap/
├── gain/
│   ├── gain.clap          # The CLAP plugin
│   ├── libgain.so         # Go shared library
│   ├── gain.json          # Plugin manifest
│   └── presets/
│       └── factory/
│           ├── boost.json
│           └── unity.json
└── synth/
    ├── synth.clap
    ├── libsynth.so
    ├── synth.json
    └── presets/
        └── factory/
            ├── lead.json
            └── pad.json
```

### Installation Steps

1. Creates plugin directory at `~/.clap/<plugin_name>/`
2. Copies the CLAP plugin file
3. Copies the Go shared library
4. Copies the JSON manifest
5. Copies preset files maintaining directory structure
6. Sets executable permissions on binary files

## Build Targets

### Primary Targets

- `make all`: Builds all components (Go bridge + all plugins)
- `make build-go`: Builds only the Go bridge library
- `make build-plugins`: Builds all example plugins
- `make build-gain`: Builds only the gain plugin
- `make build-synth`: Builds only the synth plugin

### Management Targets

- `make install`: Installs all plugins to `~/.clap/`
- `make uninstall`: Removes installed plugins
- `make clean`: Removes build artifacts
- `make clean-all`: Removes build artifacts and installed files
- `make test`: Runs plugin validation tests

### Code Generation Targets

- `make new-plugin`: Interactive plugin creation wizard
- `make generate-plugin NAME=<name> ID=<id>`: Generate new plugin from template
- `make build-plugin NAME=<name>`: Build a generated plugin
- `make validate-plugin NAME=<name>`: Validate plugin with clap-validator

## Debug Builds

Enable debug builds with:
```bash
make DEBUG=1
```

Debug builds include:
- Full debug symbols
- No optimization (`-O0`)
- No dead code elimination
- Easier debugging with GDB/LLDB

## Troubleshooting

### Common Build Issues

1. **Missing json-c**: Install with `apt-get install libjson-c-dev` (Linux) or `brew install json-c` (macOS)
2. **CGO errors**: Ensure `CGO_ENABLED=1` and a C compiler is available
3. **Library not found**: Check that `RPATH` is correctly set in the final binary

### Verification Commands

Check library dependencies:
```bash
ldd <plugin>.clap  # Linux
otool -L <plugin>.clap  # macOS
```

Verify exported symbols:
```bash
nm -D lib<plugin>.so | grep ClapGo_
```

## Performance Considerations

The build system optimizes for:
- **Small binary size**: Strips debug information in release builds
- **Fast loading**: Minimal dynamic dependencies
- **Efficient linking**: Direct function calls between C and Go

## Future Improvements

Planned enhancements to the build system:
- Parallel compilation of C objects
- Incremental builds with proper dependency tracking
- Cross-compilation support
- Automated testing integration