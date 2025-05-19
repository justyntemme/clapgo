#!/bin/bash
set -e

# Check if a plugin name is provided
if [ -z "$1" ]; then
    echo "Error: Please provide a plugin name"
    echo "Usage: $0 <plugin-name>"
    exit 1
fi

PLUGIN_NAME=$1
PLUGIN_DIR="examples/$PLUGIN_NAME"
TEMPLATE_DIR="examples/gain"

# Check if the template directory exists
if [ ! -d "$TEMPLATE_DIR" ]; then
    echo "Error: Template directory $TEMPLATE_DIR not found"
    exit 1
fi

# Check if the plugin directory already exists
if [ -d "$PLUGIN_DIR" ]; then
    echo "Error: Plugin directory $PLUGIN_DIR already exists"
    exit 1
fi

# Create the plugin directory
mkdir -p "$PLUGIN_DIR"

# Copy template files
cp "$TEMPLATE_DIR/main.go" "$PLUGIN_DIR/main.go"

# Replace plugin names in the code
PLUGIN_NAME_UPPER=$(echo $PLUGIN_NAME | sed 's/.*/\u&/')
sed -i "s/Gain/$PLUGIN_NAME_UPPER/g" "$PLUGIN_DIR/main.go"
sed -i "s/gain/$PLUGIN_NAME/g" "$PLUGIN_DIR/main.go"
sed -i "s/com\.clapgo\.gain/com.clapgo.$PLUGIN_NAME/g" "$PLUGIN_DIR/main.go"
sed -i "s/gainPlugin/${PLUGIN_NAME}Plugin/g" "$PLUGIN_DIR/main.go"
sed -i "s/GainGetPluginCount/${PLUGIN_NAME_UPPER}GetPluginCount/g" "$PLUGIN_DIR/main.go"

echo "Plugin $PLUGIN_NAME created in $PLUGIN_DIR"
echo "Build it with: ./build.sh"