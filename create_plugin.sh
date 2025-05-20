#!/bin/bash
set -e

# Display usage information
usage() {
    echo "Usage: $0 <plugin-name> [<plugin-id>]"
    echo "  <plugin-name>  : The name of the plugin (required)"
    echo "  <plugin-id>    : Custom plugin ID (optional, defaults to com.clapgo.<plugin-name>)"
    echo "Example: $0 reverb com.mycompany.reverb"
    exit 1
}

# Check if a plugin name is provided
if [ -z "$1" ]; then
    echo "Error: Please provide a plugin name"
    usage
fi

PLUGIN_NAME=$1
PLUGIN_DIR="examples/$PLUGIN_NAME"
TEMPLATE_DIR="examples/gain"

# Set default plugin ID or use custom one if provided
if [ -z "$2" ]; then
    PLUGIN_ID="com.clapgo.$PLUGIN_NAME"
else
    PLUGIN_ID="$2"
fi

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
sed -i "s/com\.clapgo\.gain/$PLUGIN_ID/g" "$PLUGIN_DIR/main.go"
sed -i "s/gainPlugin/${PLUGIN_NAME}Plugin/g" "$PLUGIN_DIR/main.go"
sed -i "s/GainGetPluginCount/${PLUGIN_NAME_UPPER}GetPluginCount/g" "$PLUGIN_DIR/main.go"

echo "Plugin $PLUGIN_NAME created in $PLUGIN_DIR with ID $PLUGIN_ID"
echo "Build it with: ./build.sh"