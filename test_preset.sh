#!/bin/bash

# Test script for loading presets

if [ $# -lt 1 ]; then
    echo "Usage: $0 <plugin-path> [preset-path]"
    exit 1
fi

PLUGIN_PATH=$1
PRESET_PATH=${2:-"presets/factory/lead.json"}

# Set environment variable to load only the synth plugin
export CLAPGO_PLUGIN_TYPE=synth

# First run the standard plugin test to make sure it loads
echo "Testing plugin basic functionality..."
./test_plugin.sh $PLUGIN_PATH

# Now test the preset load functionality
echo ""
echo "Testing preset loading from: $PRESET_PATH"

# Create a temporary test program to load the preset
cat > preset_test.c << EOF
#include <stdio.h>
#include <stdlib.h>
#include <dlfcn.h>
#include <assert.h>
#include <string.h>
#include <unistd.h>
#include "./include/clap/include/clap/clap.h"
#include "./include/clap/include/clap/ext/preset-load.h"

const char *plugin_path = "$PLUGIN_PATH";
const char *preset_path = "$PRESET_PATH";

int main() {
    void *plugin_handle = dlopen(plugin_path, RTLD_NOW | RTLD_LOCAL);
    if (!plugin_handle) {
        fprintf(stderr, "Failed to load plugin: %s\n", dlerror());
        return 1;
    }

    const clap_plugin_entry_t *entry = dlsym(plugin_handle, "clap_entry");
    if (!entry) {
        fprintf(stderr, "Failed to get clap_entry symbol\n");
        dlclose(plugin_handle);
        return 1;
    }

    const clap_plugin_factory_t *factory = entry->get_factory(CLAP_PLUGIN_FACTORY_ID);
    if (!factory) {
        fprintf(stderr, "Failed to get plugin factory\n");
        dlclose(plugin_handle);
        return 1;
    }

    // Initialize the plugin
    entry->init(plugin_path);

    // Create host
    static clap_host_t host = {
        .clap_version = CLAP_VERSION,
        .host_data = NULL,
        .name = "Preset Test",
        .vendor = "ClapGo",
        .url = "https://github.com/justyntemme/clapgo",
        .version = "1.0.0",
        .get_extension = NULL,
        .request_restart = NULL,
        .request_process = NULL,
        .request_callback = NULL
    };

    // Count available plugins
    uint32_t plugin_count = factory->get_plugin_count(factory);
    if (plugin_count == 0) {
        fprintf(stderr, "No plugins found\n");
        entry->deinit();
        dlclose(plugin_handle);
        return 1;
    }

    printf("Found %d plugins\n", plugin_count);

    // Get plugin ID for the first plugin
    const clap_plugin_descriptor_t *desc = factory->get_plugin_descriptor(factory, 0);
    printf("Plugin: %s (%s)\n", desc->name, desc->id);

    // Create the plugin instance
    const clap_plugin_t *plugin = factory->create_plugin(factory, &host, desc->id);
    if (!plugin) {
        fprintf(stderr, "Failed to create plugin instance\n");
        entry->deinit();
        dlclose(plugin_handle);
        return 1;
    }

    // Initialize plugin
    if (!plugin->init(plugin)) {
        fprintf(stderr, "Failed to initialize plugin\n");
        plugin->destroy(plugin);
        entry->deinit();
        dlclose(plugin_handle);
        return 1;
    }

    // Get preset-load extension
    const clap_plugin_preset_load_t *preset_load = 
        plugin->get_extension(plugin, CLAP_EXT_PRESET_LOAD);
    
    if (!preset_load) {
        fprintf(stderr, "Plugin does not support preset loading\n");
        plugin->destroy(plugin);
        entry->deinit();
        dlclose(plugin_handle);
        return 1;
    }

    printf("Plugin supports preset loading\n");

    // Load the preset
    printf("Loading preset from: %s\n", preset_path);
    bool success = preset_load->from_location(plugin, 0, preset_path, preset_path);
    
    if (success) {
        printf("Preset loaded successfully!\n");
    } else {
        fprintf(stderr, "Failed to load preset\n");
    }

    // Clean up
    plugin->destroy(plugin);
    entry->deinit();
    dlclose(plugin_handle);

    return success ? 0 : 1;
}
EOF

# Compile the test program
gcc -o preset_test preset_test.c -ldl

# Run the test
./preset_test

# Clean up
rm preset_test preset_test.c