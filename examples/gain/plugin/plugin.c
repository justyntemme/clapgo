#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include "../../../include/clap/include/clap/clap.h"

// Plugin callbacks
static bool plugin_init_cb(const clap_plugin_t *plugin);
static void plugin_destroy_cb(const clap_plugin_t *plugin);
static bool plugin_activate_cb(const clap_plugin_t *plugin, double sample_rate, 
                            uint32_t min_frames, uint32_t max_frames);
static void plugin_deactivate_cb(const clap_plugin_t *plugin);
static bool plugin_start_processing_cb(const clap_plugin_t *plugin);
static void plugin_stop_processing_cb(const clap_plugin_t *plugin);
static void plugin_reset_cb(const clap_plugin_t *plugin);
static clap_process_status plugin_process_cb(const clap_plugin_t *plugin, 
                                          const clap_process_t *process);
static const void *plugin_get_extension_cb(const clap_plugin_t *plugin, const char *id);
static void plugin_on_main_thread_cb(const clap_plugin_t *plugin);

// Forward declarations of plugin functions
static bool plugin_init(const char *plugin_path);
static void plugin_deinit(void);
static const void *plugin_get_factory(const char *factory_id);

// Forward declarations of factory functions
static uint32_t plugin_get_count(const clap_plugin_factory_t *factory);
static const clap_plugin_descriptor_t *plugin_get_descriptor(
    const clap_plugin_factory_t *factory, uint32_t index);
static const clap_plugin_t *plugin_create(
    const clap_plugin_factory_t *factory, const clap_host_t *host, const char *plugin_id);

// Plugin factory instance
static const clap_plugin_factory_t plugin_factory = {
    .get_plugin_count = plugin_get_count,
    .get_plugin_descriptor = plugin_get_descriptor,
    .create_plugin = plugin_create
};

// Static descriptor for our plugin
static const char* plugin_features[] = { "audio-effect", "stereo", "mono", NULL };
static const clap_plugin_descriptor_t plugin_descriptor = {
    .clap_version = CLAP_VERSION_INIT,
    .id = "com.clapgo.gain",
    .name = "Simple Gain",
    .vendor = "ClapGo",
    .url = "https://github.com/justyntemme/clapgo",
    .manual_url = "https://github.com/justyntemme/clapgo",
    .support_url = "https://github.com/justyntemme/clapgo/issues",
    .version = "1.0.0",
    .description = "A simple gain plugin using ClapGo",
    .features = plugin_features
};

// Export the clap_entry symbol
__attribute__((visibility("default")))
const clap_plugin_entry_t clap_entry = {
    .clap_version = CLAP_VERSION_INIT,
    .init = plugin_init,
    .deinit = plugin_deinit,
    .get_factory = plugin_get_factory
};

// Initialize the plugin library
static bool plugin_init(const char *plugin_path) {
    printf("Initializing plugin at path: %s\n", plugin_path);
    return true;
}

// Clean up the plugin library
static void plugin_deinit(void) {
    printf("Deinitializing plugin\n");
}

// Get the plugin factory
static const void *plugin_get_factory(const char *factory_id) {
    if (strcmp(factory_id, CLAP_PLUGIN_FACTORY_ID) == 0) {
        return &plugin_factory;
    }
    return NULL;
}

// Get the number of plugins
static uint32_t plugin_get_count(const clap_plugin_factory_t *factory) {
    return 1;
}

// Get the plugin descriptor
static const clap_plugin_descriptor_t *plugin_get_descriptor(
    const clap_plugin_factory_t *factory, uint32_t index) {
    
    if (index > 0) {
        return NULL;
    }
    
    return &plugin_descriptor;
}

// Create a plugin instance
static const clap_plugin_t *plugin_create(
    const clap_plugin_factory_t *factory, const clap_host_t *host, const char *plugin_id) {
    
    if (strcmp(plugin_id, "com.clapgo.gain") != 0) {
        return NULL;
    }
    
    // Allocate a plugin structure
    clap_plugin_t *plugin = calloc(1, sizeof(clap_plugin_t));
    if (!plugin) {
        return NULL;
    }
    
    // Initialize the plugin
    plugin->desc = &plugin_descriptor;
    plugin->plugin_data = NULL;  // Will be initialized by Go code
    
    // Set plugin callbacks - these will be implemented in Go
    plugin->init = plugin_init_cb;
    plugin->destroy = plugin_destroy_cb;
    plugin->activate = plugin_activate_cb;
    plugin->deactivate = plugin_deactivate_cb;
    plugin->start_processing = plugin_start_processing_cb;
    plugin->stop_processing = plugin_stop_processing_cb;
    plugin->reset = plugin_reset_cb;
    plugin->process = plugin_process_cb;
    plugin->get_extension = plugin_get_extension_cb;
    plugin->on_main_thread = plugin_on_main_thread_cb;
    
    return plugin;
}

// Plugin callbacks - these are placeholders that will be overridden
// by the Go plugin implementation
static bool plugin_init_cb(const clap_plugin_t *plugin) {
    return true;
}

static void plugin_destroy_cb(const clap_plugin_t *plugin) {
    free((void*)plugin);
}

static bool plugin_activate_cb(const clap_plugin_t *plugin, double sample_rate, 
                            uint32_t min_frames, uint32_t max_frames) {
    return true;
}

static void plugin_deactivate_cb(const clap_plugin_t *plugin) {
}

static bool plugin_start_processing_cb(const clap_plugin_t *plugin) {
    return true;
}

static void plugin_stop_processing_cb(const clap_plugin_t *plugin) {
}

static void plugin_reset_cb(const clap_plugin_t *plugin) {
}

static clap_process_status plugin_process_cb(const clap_plugin_t *plugin, 
                                          const clap_process_t *process) {
    return CLAP_PROCESS_CONTINUE;
}

static const void *plugin_get_extension_cb(const clap_plugin_t *plugin, const char *id) {
    return NULL;
}

static void plugin_on_main_thread_cb(const clap_plugin_t *plugin) {
}