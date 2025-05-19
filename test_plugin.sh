#!/bin/bash
set -e

# Default plugin location
PLUGIN_PATH="build/linux/examples/gain/gain.clap"

# Allow custom plugin path
if [ ! -z "$1" ]; then
    PLUGIN_PATH="$1"
fi

# Check if plugin exists
if [ ! -f "$PLUGIN_PATH" ]; then
    echo "Error: Plugin not found at $PLUGIN_PATH"
    exit 1
fi

# Test plugin loading with ldd
echo "Testing plugin dynamic dependencies..."
ldd "$PLUGIN_PATH"

# Test CLAP plugin entry point with nm
echo -e "\nVerifying CLAP entry point..."
nm -D "$PLUGIN_PATH" | grep clap_entry

# Run clap-validator if available
if command -v clap-validator &> /dev/null; then
    echo -e "\nRunning CLAP validator..."
    clap-validator validate "$PLUGIN_PATH"
else
    echo -e "\nNote: clap-validator not found. Consider installing it for more thorough testing."
fi

# Print plugin info
echo -e "\nPlugin successfully built at: $PLUGIN_PATH"
echo "Plugin size: $(du -h "$PLUGIN_PATH" | cut -f1)"

echo -e "\nBuild successful! To install the plugin, run:"
echo "./build.sh --install"