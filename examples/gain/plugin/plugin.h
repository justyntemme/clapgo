#ifndef PLUGIN_H
#define PLUGIN_H

#include "../../../include/clap/include/clap/clap.h"

// Forward declarations of plugin functions
static bool plugin_init(const char *plugin_path);
static void plugin_deinit(void);
static const void *plugin_get_factory(const char *factory_id);

// Plugin factory functions
static uint32_t plugin_get_count(const clap_plugin_factory_t *factory);
static const clap_plugin_descriptor_t *plugin_get_descriptor(
    const clap_plugin_factory_t *factory, uint32_t index);
static const clap_plugin_t *plugin_create(
    const clap_plugin_factory_t *factory, const clap_host_t *host, const char *plugin_id);

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

// Plugin factory instance
static const clap_plugin_factory_t plugin_factory;

#endif // PLUGIN_H