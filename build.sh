#!/bin/bash
set -xe

# CPU architecture detection
cpu="$(uname -m)"
case "$cpu" in
x86_64)
  cpu="x64";;
i686)
  cpu="x86";;
esac

# Platform detection
if [[ $(uname) = Linux ]] ; then
  cmake_preset="linux"
  triplet=$cpu-linux
elif [[ $(uname) = Darwin ]] ; then
  cmake_preset="macos"
  triplet=$cpu-osx
else
  cmake_preset="windows"
  triplet=$cpu-win
fi

# Parse arguments
debug=false
install=false
test=false
clean=false

while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in
        --debug)
            debug=true
            shift
            ;;
        --install)
            install=true
            shift
            ;;
        --test)
            test=true
            shift
            ;;
        --clean)
            clean=true
            shift
            ;;
        --preset)
            cmake_preset="$2"
            shift
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --debug       Build in debug mode"
            echo "  --install     Install plugins to system directory"
            echo "  --test        Run tests after building"
            echo "  --clean       Clean build directory before building"
            echo "  --preset NAME Use specific CMake preset"
            echo "  --help        Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for available options"
            exit 1
            ;;
    esac
done

# Set build directory
build_dir="build/$cmake_preset"
mkdir -p $build_dir

# Clean if requested
if $clean; then
    echo "Cleaning build directory..."
    rm -rf $build_dir/*
fi

# Set build config based on debug flag
build_config="Release"
if $debug; then
    build_config="Debug"
    cmake_options="-DCMAKE_BUILD_TYPE=Debug"
else
    cmake_options="-DCMAKE_BUILD_TYPE=Release"
fi

# Configure
echo "Configuring with preset: $cmake_preset"
cmake --preset "$cmake_preset" $cmake_options

# Build
echo "Building plugins..."
cmake --build --preset "$cmake_preset" --config $build_config -j$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)

# Build the shared Go library
echo "Building libgoclap.so shared library..."
mkdir -p $build_dir/src
CGO_ENABLED=1 go build -buildmode=c-shared -o $build_dir/src/libgoclap.so ./cmd/wrapper

# Test if requested
if $test; then
    echo "Testing plugins..."
    for plugin in $build_dir/examples/*/*.clap; do
        if [ -f "$plugin" ]; then
            ./test_plugin.sh "$plugin"
        fi
    done
fi

# Install if requested
if $install; then
    echo "Installing plugins..."
    cmake --build --preset "$cmake_preset" --config $build_config --target install
fi

echo "Build complete!"
echo "Plugins are available in: $build_dir"