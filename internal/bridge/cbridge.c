#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include "../../include/clap/include/clap/clap.h"

// Export the clap_entry symbol
__attribute__((visibility("default")))
extern const clap_plugin_entry_t clap_entry;

// Function declarations for Go functions
extern uint32_t BridgeGetPluginCount();
extern struct clap_plugin_descriptor* BridgeGetPluginInfo(uint32_t index);
extern void* BridgeCreatePlugin(struct clap_host *host, char *plugin_id);
extern bool BridgeGetVersion(uint32_t *major, uint32_t *minor, uint32_t *patch);

extern bool BridgeInit(void *plugin);
extern void BridgeDestroy(void *plugin);
extern bool BridgeActivate(void *plugin, double sample_rate, uint32_t min_frames, uint32_t max_frames);
extern void BridgeDeactivate(void *plugin);
extern bool BridgeStartProcessing(void *plugin);
extern void BridgeStopProcessing(void *plugin);
extern void BridgeReset(void *plugin);
extern int32_t BridgeProcess(void *plugin, struct clap_process *process);
extern void* BridgeGetExtension(void *plugin, char *id);
extern void BridgeOnMainThread(void *plugin);

// Plugin instance data
typedef struct {
    void *go_instance;
    const clap_plugin_descriptor_t *descriptor;
} plugin_data_t;

// Plugin factory implementation
static uint32_t factory_get_plugin_count(const clap_plugin_factory_t *factory) {
    return BridgeGetPluginCount();
}

static const clap_plugin_descriptor_t *factory_get_plugin_descriptor(
    const clap_plugin_factory_t *factory, uint32_t index) {
    return BridgeGetPluginInfo(index);
}

// Plugin instance callbacks
static bool plugin_init(const clap_plugin_t *plugin) {
    if (!plugin) return false;
    plugin_data_t *data = (plugin_data_t *)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    return BridgeInit(data->go_instance);
}

static void plugin_destroy(const clap_plugin_t *plugin) {
    if (!plugin) return;
    plugin_data_t *data = (plugin_data_t *)plugin->plugin_data;
    if (!data) return;

    if (data->go_instance) {
        BridgeDestroy(data->go_instance);
    }

    free(data);
    free((void *)plugin);
}

static bool plugin_activate(const clap_plugin_t *plugin, double sample_rate,
                           uint32_t min_frames, uint32_t max_frames) {
    if (!plugin) return false;
    plugin_data_t *data = (plugin_data_t *)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    return BridgeActivate(data->go_instance, sample_rate, min_frames, max_frames);
}

static void plugin_deactivate(const clap_plugin_t *plugin) {
    if (!plugin) return;
    plugin_data_t *data = (plugin_data_t *)plugin->plugin_data;
    if (!data || !data->go_instance) return;
    BridgeDeactivate(data->go_instance);
}

static bool plugin_start_processing(const clap_plugin_t *plugin) {
    if (!plugin) return false;
    plugin_data_t *data = (plugin_data_t *)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    return BridgeStartProcessing(data->go_instance);
}

static void plugin_stop_processing(const clap_plugin_t *plugin) {
    if (!plugin) return;
    plugin_data_t *data = (plugin_data_t *)plugin->plugin_data;
    if (!data || !data->go_instance) return;
    BridgeStopProcessing(data->go_instance);
}

static void plugin_reset(const clap_plugin_t *plugin) {
    if (!plugin) return;
    plugin_data_t *data = (plugin_data_t *)plugin->plugin_data;
    if (!data || !data->go_instance) return;
    BridgeReset(data->go_instance);
}

static clap_process_status plugin_process(const clap_plugin_t *plugin, const clap_process_t *process) {
    if (!plugin || !process) return CLAP_PROCESS_ERROR;
    plugin_data_t *data = (plugin_data_t *)plugin->plugin_data;
    if (!data || !data->go_instance) return CLAP_PROCESS_ERROR;
    return BridgeProcess(data->go_instance, (struct clap_process *)process);
}

static const void *plugin_get_extension(const clap_plugin_t *plugin, const char *id) {
    if (!plugin || !id) return NULL;
    plugin_data_t *data = (plugin_data_t *)plugin->plugin_data;
    if (!data || !data->go_instance) return NULL;
    return BridgeGetExtension(data->go_instance, (char *)id);
}

static void plugin_on_main_thread(const clap_plugin_t *plugin) {
    if (!plugin) return;
    plugin_data_t *data = (plugin_data_t *)plugin->plugin_data;
    if (!data || !data->go_instance) return;
    BridgeOnMainThread(data->go_instance);
}

static const clap_plugin_t *factory_create_plugin(
    const clap_plugin_factory_t *factory, const clap_host_t *host, const char *plugin_id) {
    
    if (!plugin_id) {
        fprintf(stderr, "Error: plugin_id is NULL\n");
        return NULL;
    }

    // Find the plugin descriptor
    uint32_t count = BridgeGetPluginCount();
    const clap_plugin_descriptor_t *descriptor = NULL;

    for (uint32_t i = 0; i < count; i++) {
        const clap_plugin_descriptor_t *desc = BridgeGetPluginInfo(i);
        if (desc && strcmp(desc->id, plugin_id) == 0) {
            descriptor = desc;
            break;
        }
    }

    if (!descriptor) {
        fprintf(stderr, "Error: Plugin ID not found: %s\n", plugin_id);
        return NULL;
    }

    // Create the Go plugin instance
    void *go_instance = BridgeCreatePlugin((struct clap_host *)host, (char *)plugin_id);
    if (!go_instance) {
        fprintf(stderr, "Error: Failed to create Go plugin instance\n");
        return NULL;
    }

    // Allocate plugin instance data
    plugin_data_t *data = calloc(1, sizeof(plugin_data_t));
    if (!data) {
        fprintf(stderr, "Error: Failed to allocate plugin data\n");
        // TODO: Should call a Go function to destroy the plugin instance
        return NULL;
    }

    data->go_instance = go_instance;
    data->descriptor = descriptor;

    // Allocate a CLAP plugin structure
    clap_plugin_t *plugin = calloc(1, sizeof(clap_plugin_t));
    if (!plugin) {
        fprintf(stderr, "Error: Failed to allocate plugin structure\n");
        free(data);
        // TODO: Should call a Go function to destroy the plugin instance
        return NULL;
    }

    // Initialize the plugin structure
    plugin->desc = descriptor;
    plugin->plugin_data = data;

    // Set plugin callbacks
    plugin->init = plugin_init;
    plugin->destroy = plugin_destroy;
    plugin->activate = plugin_activate;
    plugin->deactivate = plugin_deactivate;
    plugin->start_processing = plugin_start_processing;
    plugin->stop_processing = plugin_stop_processing;
    plugin->reset = plugin_reset;
    plugin->process = plugin_process;
    plugin->get_extension = plugin_get_extension;
    plugin->on_main_thread = plugin_on_main_thread;

    return plugin;
}

// Plugin factory instance
static const clap_plugin_factory_t plugin_factory = {
    .get_plugin_count = factory_get_plugin_count,
    .get_plugin_descriptor = factory_get_plugin_descriptor,
    .create_plugin = factory_create_plugin
};

// Entry point implementation
static bool entry_init(const char *plugin_path) {
    printf("Initializing ClapGo bridge with path: %s\n", plugin_path);
    // In the full implementation, we would load the Go shared library here if needed
    // but since Go is already initialized when this C code is called from Go,
    // we just return success
    return true;
}

static void entry_deinit(void) {
    printf("Deinitializing ClapGo bridge\n");
    // In the full implementation, we would clean up resources here
}

static const void *entry_get_factory(const char *factory_id) {
    if (strcmp(factory_id, CLAP_PLUGIN_FACTORY_ID) == 0) {
        return &plugin_factory;
    }
    return NULL;
}

// Entry point instance
const clap_plugin_entry_t clap_entry = {
    .clap_version = CLAP_VERSION_INIT,
    .init = entry_init,
    .deinit = entry_deinit,
    .get_factory = entry_get_factory
};