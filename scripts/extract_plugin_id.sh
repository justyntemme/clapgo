#!/bin/bash
# Extract plugin ID from a Go file by looking for the PluginID constant

if [ -z "$1" ]; then
    echo "Usage: $0 <plugin_directory>"
    exit 1
fi

PLUGIN_DIR="$1"
CONSTANTS_FILE="$PLUGIN_DIR/constants.go"

if [ ! -f "$CONSTANTS_FILE" ]; then
    echo "Error: Constants file not found: $CONSTANTS_FILE"
    exit 1
fi

# Extract the plugin ID from the constants.go file
PLUGIN_ID=$(grep -oP 'PluginID\s*=\s*"\K[^"]+' "$CONSTANTS_FILE")

if [ -z "$PLUGIN_ID" ]; then
    echo "Error: Failed to extract plugin ID from $CONSTANTS_FILE"
    exit 1
fi

echo "$PLUGIN_ID"