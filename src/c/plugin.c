#include "bridge.h"
// Note: plugin.h contains legacy declarations - we only need bridge.h
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

// Plugin factory implementation
static uint32_t clapgo_factory_get_plugin_count(const clap_plugin_factory_t *factory);
static const clap_plugin_descriptor_t *clapgo_factory_get_plugin_descriptor(
    const clap_plugin_factory_t *factory, uint32_t index);
static const clap_plugin_t *clapgo_factory_create_plugin(
    const clap_plugin_factory_t *factory, const clap_host_t *host, const char *plugin_id);

// Plugin entry point implementation
static bool clapgo_entry_init(const char *plugin_path);
static void clapgo_entry_deinit(void);
static const void *clapgo_entry_get_factory(const char *factory_id);

// Plugin factory instance
static const clap_plugin_factory_t clapgo_factory = {
    .get_plugin_count = clapgo_factory_get_plugin_count,
    .get_plugin_descriptor = clapgo_factory_get_plugin_descriptor,
    .create_plugin = clapgo_factory_create_plugin
};

// CLAP plugin entry point instance - this is the main entry point for CLAP hosts
const CLAP_EXPORT clap_plugin_entry_t clap_entry = {
    .clap_version = CLAP_VERSION_INIT,
    .init = clapgo_entry_init,
    .deinit = clapgo_entry_deinit,
    .get_factory = clapgo_entry_get_factory
};

// Plugin factory implementation
static uint32_t clapgo_factory_get_plugin_count(const clap_plugin_factory_t *factory) {
    return clapgo_get_plugin_count();
}

static const clap_plugin_descriptor_t *clapgo_factory_get_plugin_descriptor(
    const clap_plugin_factory_t *factory, uint32_t index) {
    return clapgo_get_plugin_descriptor(index);
}

static const clap_plugin_t *clapgo_factory_create_plugin(
    const clap_plugin_factory_t *factory, const clap_host_t *host, const char *plugin_id) {
    return clapgo_create_plugin(host, plugin_id);
}

// Plugin entry point implementation
static bool clapgo_entry_init(const char *plugin_path) {
    return clapgo_init(plugin_path);
}

static void clapgo_entry_deinit(void) {
    clapgo_deinit();
}

static const void *clapgo_entry_get_factory(const char *factory_id) {
    if (strcmp(factory_id, CLAP_PLUGIN_FACTORY_ID) == 0) {
        return &clapgo_factory;
    }
    return NULL;
}