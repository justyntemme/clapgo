#ifndef CLAPGO_PLUGIN_H
#define CLAPGO_PLUGIN_H

#include <stdbool.h>
#include <stdint.h>
#include "../../include/clap/include/clap/clap.h"

#ifdef __cplusplus
extern "C" {
#endif

// Go plugin state structure
typedef struct go_plugin_data {
    void* go_instance;
    const clap_plugin_descriptor_t* descriptor;
} go_plugin_data_t;

// Initialize the Go runtime and plugin environment
bool clapgo_init(const char* plugin_path);

// Cleanup the Go runtime and plugin environment
void clapgo_deinit(void);

// Create a plugin instance for the given ID
const clap_plugin_t* clapgo_create_plugin(const clap_host_t* host, const char* plugin_id);

// Get the number of available plugins
uint32_t clapgo_get_plugin_count(void);

// Get the plugin descriptor at the given index
const clap_plugin_descriptor_t* clapgo_get_plugin_descriptor(uint32_t index);

// Get the plugin factory
const clap_plugin_factory_t* clapgo_get_plugin_factory(void);

// Initialize a plugin instance
bool clapgo_plugin_init(const clap_plugin_t* plugin);

// Destroy a plugin instance
void clapgo_plugin_destroy(const clap_plugin_t* plugin);

// Activate a plugin instance
bool clapgo_plugin_activate(const clap_plugin_t* plugin, double sample_rate,
                           uint32_t min_frames, uint32_t max_frames);

// Deactivate a plugin instance
void clapgo_plugin_deactivate(const clap_plugin_t* plugin);

// Start processing
bool clapgo_plugin_start_processing(const clap_plugin_t* plugin);

// Stop processing
void clapgo_plugin_stop_processing(const clap_plugin_t* plugin);

// Reset a plugin instance
void clapgo_plugin_reset(const clap_plugin_t* plugin);

// Process audio
clap_process_status clapgo_plugin_process(const clap_plugin_t* plugin, 
                                         const clap_process_t* process);

// Get an extension
const void* clapgo_plugin_get_extension(const clap_plugin_t* plugin, const char* id);

// Execute on main thread
void clapgo_plugin_on_main_thread(const clap_plugin_t* plugin);

#ifdef __cplusplus
}
#endif

#endif // CLAPGO_PLUGIN_H