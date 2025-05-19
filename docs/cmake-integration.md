# Integrating with CLAP's CMake Build System

This document outlines how to integrate our Go-based CLAP plugins with the official CLAP plugins build system.

## Overview

The CLAP plugins repository uses CMake with vcpkg for dependency management. This ensures cross-platform compatibility and streamlined building. To integrate our Go-based plugins, we need to:

1. Create a CMake configuration for our project
2. Hook our Go shared library into the build process
3. Use CLAP's build scripts for platform-specific packaging

## Required Changes

### 1. Create CMakeLists.txt Files

We'll need to create several CMake configuration files:

#### Root CMakeLists.txt

```cmake
cmake_minimum_required(VERSION 3.17)
project(CLAPGO CXX)

# Include the CLAP submodule
add_subdirectory(include/clap EXCLUDE_FROM_ALL)

# Set options
option(CLAPGO_BUILD_EXAMPLES "Build example plugins" ON)

# Include Go build helper
include(cmake/GoLibrary.cmake)

# Add our plugin wrapper
add_subdirectory(src/c)

# Add example plugins if enabled
if(CLAPGO_BUILD_EXAMPLES)
    add_subdirectory(examples)
endif()
```

#### src/c/CMakeLists.txt

```cmake
# Add the C wrapper library as a static library
add_library(clapgo-wrapper STATIC
    plugin.c
)

target_include_directories(clapgo-wrapper PUBLIC
    ${CMAKE_SOURCE_DIR}/include/clap/include
)

# Make header available
set_target_properties(clapgo-wrapper PROPERTIES
    PUBLIC_HEADER "plugin.h"
)
```

#### examples/CMakeLists.txt

```cmake
# Add each example subdirectory
file(GLOB EXAMPLE_DIRS RELATIVE ${CMAKE_CURRENT_SOURCE_DIR} "*")
foreach(EXAMPLE_DIR ${EXAMPLE_DIRS})
    if(IS_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}/${EXAMPLE_DIR})
        add_subdirectory(${EXAMPLE_DIR})
    endif()
endforeach()
```

#### examples/gain/CMakeLists.txt

```cmake
# Build the Go shared library
add_go_library(gain-go
    SOURCES
        main.go
    OUTPUT_NAME
        gain
)

# Build the CLAP plugin
add_clap_plugin(gain
    SOURCES
        ${CMAKE_SOURCE_DIR}/src/c/plugin.c
    LINK_LIBRARIES
        gain-go
)
```

### 2. Create Go Build Helper

Create a file `cmake/GoLibrary.cmake` with functions to integrate Go building:

```cmake
# Function to build a Go shared library
function(add_go_library TARGET_NAME)
    # Parse arguments
    cmake_parse_arguments(ARG "" "OUTPUT_NAME" "SOURCES" ${ARGN})
    
    if(NOT ARG_OUTPUT_NAME)
        set(ARG_OUTPUT_NAME ${TARGET_NAME})
    endif()
    
    # Full paths to sources
    set(GO_SOURCES "")
    foreach(SRC ${ARG_SOURCES})
        list(APPEND GO_SOURCES "${CMAKE_CURRENT_SOURCE_DIR}/${SRC}")
    endforeach()
    
    # Output file name
    if(WIN32)
        set(LIB_PREFIX "")
        set(LIB_SUFFIX ".dll")
    elseif(APPLE)
        set(LIB_PREFIX "lib")
        set(LIB_SUFFIX ".dylib")
    else()
        set(LIB_PREFIX "lib")
        set(LIB_SUFFIX ".so")
    endif()
    
    set(OUTPUT_LIB "${CMAKE_CURRENT_BINARY_DIR}/${LIB_PREFIX}${ARG_OUTPUT_NAME}${LIB_SUFFIX}")
    
    # Build command
    add_custom_command(
        OUTPUT ${OUTPUT_LIB}
        COMMAND ${CMAKE_COMMAND} -E env 
                GOPATH=${CMAKE_BINARY_DIR}/go
                ${GO_EXECUTABLE} build 
                -buildmode=c-shared 
                -o ${OUTPUT_LIB}
                ${GO_BUILD_FLAGS}
                ${GO_SOURCES}
        DEPENDS ${GO_SOURCES}
        WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
        COMMENT "Building Go shared library ${ARG_OUTPUT_NAME}"
    )
    
    # Create the target
    add_custom_target(${TARGET_NAME} ALL DEPENDS ${OUTPUT_LIB})
    
    # Export library path
    set(${TARGET_NAME}_LIBRARY ${OUTPUT_LIB} PARENT_SCOPE)
endfunction()

# Function to create a CLAP plugin
function(add_clap_plugin TARGET_NAME)
    cmake_parse_arguments(ARG "" "" "SOURCES;LINK_LIBRARIES" ${ARGN})
    
    # Determine the output extension
    if(WIN32)
        set(CLAP_EXTENSION ".clap")
    elseif(APPLE)
        set(CLAP_EXTENSION ".clap")
    else()
        set(CLAP_EXTENSION ".clap")
    endif()
    
    # Build the plugin as a shared library
    add_library(${TARGET_NAME} SHARED ${ARG_SOURCES})
    
    # Include CLAP headers
    target_include_directories(${TARGET_NAME} PRIVATE
        ${CMAKE_SOURCE_DIR}/include/clap/include
    )
    
    # Special handling for Go libraries
    foreach(LIB ${ARG_LINK_LIBRARIES})
        if(TARGET ${LIB})
            add_dependencies(${TARGET_NAME} ${LIB})
            
            # Link against the Go library
            target_link_libraries(${TARGET_NAME} ${${LIB}_LIBRARY})
        else()
            target_link_libraries(${TARGET_NAME} ${LIB})
        endif()
    endforeach()
    
    # Set output name/properties
    set_target_properties(${TARGET_NAME} PROPERTIES
        PREFIX ""
        SUFFIX ${CLAP_EXTENSION}
    )
    
    # Platform-specific settings
    if(APPLE)
        # macOS requires a bundle structure
        # Add necessary implementation here
    endif()
endfunction()

# Find Go executable
find_program(GO_EXECUTABLE go REQUIRED)
```

### 3. Integrate with CLAP Build Scripts

Create a build script that uses CLAP's build infrastructure:

```bash
#!/bin/bash
set -e

# Determine platform
PLATFORM=$(uname)
if [[ "$PLATFORM" == "Darwin" ]]; then
    CMAKE_PRESET="macos"
elif [[ "$PLATFORM" == "Linux" ]]; then
    CMAKE_PRESET="linux"
else
    CMAKE_PRESET="windows"
fi

# Configure
cmake --preset $CMAKE_PRESET

# Build
cmake --build --preset $CMAKE_PRESET

# Package plugins according to platform requirements
if [[ "$PLATFORM" == "Darwin" ]]; then
    # Create macOS bundles
    ./scripts/package-macos.sh
elif [[ "$PLATFORM" == "Linux" ]]; then
    # Linux just uses .clap files directly
    mkdir -p dist
    cp build/$CMAKE_PRESET/examples/*/*.clap dist/
else
    # Windows packaging
    mkdir -p dist
    cp build/$CMAKE_PRESET/examples/*/*.clap dist/
fi
```

## Implementation Plan

1. **Create CMake Infrastructure**
   - Create root CMakeLists.txt
   - Create helper module for Go building
   - Setup plugin build configuration

2. **Adapt C Wrapper**
   - Ensure C wrapper works with CLAP plugin entry points
   - Properly export symbols for dynamic linking

3. **Update Example Plugins**
   - Add CMakeLists.txt to each example
   - Ensure correct linking against Go libraries

4. **Create Build Scripts**
   - Implement platform-specific build scripts
   - Support debug and release configurations

5. **Testing and Verification**
   - Test builds on different platforms
   - Verify plugins load correctly in hosts

## Benefits of Using CLAP's Build System

1. **Consistency**: Our plugins will be built in the same way as reference plugins
2. **Cross-Platform Support**: Leverage existing solutions for Windows/macOS/Linux
3. **GUI Integration**: Easier to leverage CLAP's GUI framework
4. **Update Compatibility**: Stay in sync with CLAP API changes

## Challenges

1. **Go Integration**: CMake doesn't natively support Go
2. **CGo Complexities**: Cross-compilation requires careful setup
3. **Platform Specifics**: Different platforms require different plugin packaging