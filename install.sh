#!/bin/bash
# ClapGo Plugin Installer Script

set -e  # Exit on error

# Parse command line arguments
SYSTEM_WIDE=0
BUILD_GUI=0
HELP=0

for arg in "$@"; do
  case $arg in
    --system|-s)
      SYSTEM_WIDE=1
      shift
      ;;
    --gui|-g)
      BUILD_GUI=1
      shift
      ;;
    --help|-h)
      HELP=1
      shift
      ;;
    *)
      # Unknown option
      echo "Unknown option: $arg"
      HELP=1
      shift
      ;;
  esac
done

# Show help if requested
if [ $HELP -eq 1 ]; then
  echo "ClapGo Plugin Installer"
  echo "Usage: ./install.sh [options]"
  echo ""
  echo "Options:"
  echo "  --system, -s     Install plugins system-wide to /usr/lib/clap"
  echo "                   (requires sudo privileges)"
  echo "  --gui, -g        Build with GUI support (requires Qt6)"
  echo "  --help, -h       Show this help message"
  echo ""
  echo "Without options, plugins will be installed to ~/.clap"
  exit 0
fi

# Determine installation type
if [ $SYSTEM_WIDE -eq 1 ]; then
  echo "Installing plugins system-wide to /usr/lib/clap"
  INSTALL_CMD="sudo cmake --install build"
  CMAKE_OPTIONS="-DCLAPGO_INSTALL_SYSTEM_WIDE=ON"
else
  echo "Installing plugins for current user to ~/.clap"
  INSTALL_CMD="cmake --install build"
  CMAKE_OPTIONS="-DCLAPGO_INSTALL_SYSTEM_WIDE=OFF"
fi

# Add GUI option if requested
if [ $BUILD_GUI -eq 1 ]; then
  echo "Building with GUI support"
  CMAKE_OPTIONS="$CMAKE_OPTIONS -DCLAPGO_GUI_SUPPORT=ON"
else
  echo "Building without GUI support"
  CMAKE_OPTIONS="$CMAKE_OPTIONS -DCLAPGO_GUI_SUPPORT=OFF"
fi

# Create build directory if it doesn't exist
mkdir -p build

# Build the plugins
echo "Building plugins..."
cmake -B build $CMAKE_OPTIONS
cmake --build build -j$(nproc)

# Create installation directory if it doesn't exist
if [ $SYSTEM_WIDE -eq 1 ]; then
  sudo mkdir -p /usr/lib/clap
else
  mkdir -p ~/.clap
fi

# Install the plugins directly
echo "Installing plugins..."
if [ $SYSTEM_WIDE -eq 1 ]; then
  sudo mkdir -p /usr/lib/clap
  sudo cp build/examples/gain/gain.clap /usr/lib/clap/
  sudo cp build/examples/gain-with-gui/gain-with-gui.clap /usr/lib/clap/
  sudo chmod 755 /usr/lib/clap/*.clap
else
  mkdir -p ~/.clap
  cp build/examples/gain/gain.clap ~/.clap/
  cp build/examples/gain-with-gui/gain-with-gui.clap ~/.clap/
  chmod 755 ~/.clap/*.clap
fi

echo "Installation complete!"