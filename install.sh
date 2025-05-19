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

# Function to safely copy files with existence check
copy_if_exists() {
  local src="$1"
  local dest="$2"
  local sudo_mode="$3"
  
  if [ -f "$src" ]; then
    if [ "$sudo_mode" = "sudo" ]; then
      sudo cp "$src" "$dest"
    else
      cp "$src" "$dest"
    fi
    echo "Copied: $src to $dest"
  else
    echo "Warning: File not found: $src"
  fi
}

if [ $SYSTEM_WIDE -eq 1 ]; then
  sudo mkdir -p /usr/lib/clap
  
  # Copy the main Go shared library - try multiple possible locations
  for build_dir in build/linux/src build/macos/src build/windows/src build/src; do
    if [ -f "$build_dir/libgoclap.so" ]; then
      copy_if_exists "$build_dir/libgoclap.so" "/usr/lib/clap/" "sudo"
      break
    fi
  done
  
  # Scan for all built plugin files in build directory
  # Look in both the default build directory and the platform-specific directories
  for build_dir in build/ build/linux/ build/macos/ build/windows/; do
    if [ -d "${build_dir}examples/" ]; then
      for dir in ${build_dir}examples/*/; do
        if [ -d "$dir" ]; then
          plugin_name=$(basename "$dir")
          
          # Copy .clap files
          copy_if_exists "${dir}${plugin_name}.clap" "/usr/lib/clap/" "sudo"
          
          # Copy shared object libraries
          copy_if_exists "${dir}lib${plugin_name}.so" "/usr/lib/clap/" "sudo"
        fi
      done
    fi
  done
  
  # Set proper permissions
  sudo chmod 755 /usr/lib/clap/*.clap 2>/dev/null || true
  sudo chmod 755 /usr/lib/clap/*.so 2>/dev/null || true
else
  mkdir -p ~/.clap
  
  # Copy the main Go shared library - try multiple possible locations
  for build_dir in build/linux/src build/macos/src build/windows/src build/src; do
    if [ -f "$build_dir/libgoclap.so" ]; then
      copy_if_exists "$build_dir/libgoclap.so" "$HOME/.clap/" ""
      break
    fi
  done
  
  # Scan for all built plugin files in build directory
  # Look in both the default build directory and the platform-specific directories
  for build_dir in build/ build/linux/ build/macos/ build/windows/; do
    if [ -d "${build_dir}examples/" ]; then
      for dir in ${build_dir}examples/*/; do
        if [ -d "$dir" ]; then
          plugin_name=$(basename "$dir")
          
          # Copy .clap files
          copy_if_exists "${dir}${plugin_name}.clap" "$HOME/.clap/" ""
          
          # Copy shared object libraries
          copy_if_exists "${dir}lib${plugin_name}.so" "$HOME/.clap/" ""
        fi
      done
    fi
  done
  
  # Set proper permissions
  chmod 755 ~/.clap/*.clap 2>/dev/null || true
  chmod 755 ~/.clap/*.so 2>/dev/null || true
fi

echo "Installation complete!"