#!/bin/bash
# install_presets.sh - Copy presets to user directory
# This script discovers all plugins with presets and installs them

set -e

echo "Installing ClapGo presets..."

# Discover all plugins with presets
for plugin_dir in examples/*/; do
    if [ -d "$plugin_dir/presets" ]; then
        plugin_name=$(basename "$plugin_dir")
        
        # Skip GUI examples per guardrails
        if [[ "$plugin_name" == *"-gui"* ]]; then
            continue
        fi
        
        # Create preset directory
        preset_dest="$HOME/.clap/$plugin_name/presets"
        mkdir -p "$preset_dest"
        
        # Copy all preset files
        if [ -d "$plugin_dir/presets/factory" ]; then
            echo "Installing presets for $plugin_name..."
            cp "$plugin_dir/presets/factory"/*.json "$preset_dest/" 2>/dev/null || true
        fi
    fi
done

# List installed presets
echo ""
echo "Installed presets:"
for plugin_dir in "$HOME/.clap"/*/; do
    if [ -d "$plugin_dir/presets" ]; then
        plugin_name=$(basename "$plugin_dir")
        preset_count=$(find "$plugin_dir/presets" -name "*.json" | wc -l)
        if [ "$preset_count" -gt 0 ]; then
            echo "  - $plugin_name: $preset_count presets"
        fi
    fi
done

echo ""
echo "Presets installed to ~/.clap/"