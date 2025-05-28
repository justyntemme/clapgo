#include "bridge.h"
// Note: plugin.h contains legacy declarations - we only need bridge.h
#include "preset_discovery.h"
#include "plugin_invalidation.h"
#include "state_converter.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <time.h>

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
    uint32_t count = clapgo_get_plugin_count();
    
    FILE* log = fopen("/tmp/clapgo_factory_calls.log", "a");
    if (log) {
        fprintf(log, "[%ld] clapgo_factory_get_plugin_count() called, returning %u\n", time(NULL), count);
        fflush(log);
        fclose(log);
    }
    
    return count;
}

static const clap_plugin_descriptor_t *clapgo_factory_get_plugin_descriptor(
    const clap_plugin_factory_t *factory, uint32_t index) {
    const clap_plugin_descriptor_t *desc = clapgo_get_plugin_descriptor(index);
    
    FILE* log = fopen("/tmp/clapgo_factory_calls.log", "a");
    if (log) {
        fprintf(log, "[%ld] clapgo_factory_get_plugin_descriptor() called with index %u\n", time(NULL), index);
        if (desc) {
            fprintf(log, "  Returning descriptor for: %s (%s)\n", desc->id, desc->name);
        } else {
            fprintf(log, "  Returning NULL\n");
        }
        fflush(log);
        fclose(log);
    }
    
    return desc;
}

static const clap_plugin_t *clapgo_factory_create_plugin(
    const clap_plugin_factory_t *factory, const clap_host_t *host, const char *plugin_id) {
    FILE* log = fopen("/tmp/clapgo_factory_calls.log", "a");
    if (log) {
        fprintf(log, "[%ld] clapgo_factory_create_plugin() called\n", time(NULL));
        fprintf(log, "  plugin_id: %s\n", plugin_id ? plugin_id : "NULL");
        fprintf(log, "  host: %p\n", host);
        fflush(log);
        fclose(log);
    }
    
    const clap_plugin_t *plugin = clapgo_create_plugin(host, plugin_id);
    
    log = fopen("/tmp/clapgo_factory_calls.log", "a");
    if (log) {
        fprintf(log, "  plugin created: %p\n", plugin);
        fflush(log);
        fclose(log);
    }
    
    return plugin;
}

// Plugin entry point implementation
static bool clapgo_entry_init(const char *plugin_path) {
    printf("[ClapGo] Entry init called with path: %s\n", plugin_path ? plugin_path : "NULL");
    
    FILE* log = fopen("/tmp/clapgo_factory_calls.log", "a");
    if (log) {
        fprintf(log, "[%ld] clapgo_entry_init() called with path: %s\n", 
                time(NULL), plugin_path ? plugin_path : "NULL");
        fflush(log);
        fclose(log);
    }
    
    return clapgo_init(plugin_path);
}

static void clapgo_entry_deinit(void) {
    printf("[ClapGo] Entry deinit called\n");
    
    FILE* log = fopen("/tmp/clapgo_factory_calls.log", "a");
    if (log) {
        fprintf(log, "[%ld] clapgo_entry_deinit() called\n", time(NULL));
        fflush(log);
        fclose(log);
    }
    
    clapgo_deinit();
}

static const void *clapgo_entry_get_factory(const char *factory_id) {
    // Also log to file for debugging
    FILE* log = fopen("/tmp/clapgo_factory_calls.log", "a");
    if (log) {
        fprintf(log, "[%ld] clapgo_entry_get_factory() called with factory_id: %s\n", 
                time(NULL), factory_id ? factory_id : "NULL");
        fflush(log);
    }
    
    printf("[PRESET_DEBUG] clapgo_entry_get_factory() called with factory_id: %s\n", factory_id ? factory_id : "NULL");
    
    if (!factory_id) {
        if (log) {
            fprintf(log, "  ERROR: factory_id is NULL\n");
            fclose(log);
        }
        return NULL;
    }
    
    if (strcmp(factory_id, CLAP_PLUGIN_FACTORY_ID) == 0) {
        printf("[PRESET_DEBUG] Returning plugin factory\n");
        if (log) {
            fprintf(log, "  Returning plugin factory\n");
            fclose(log);
        }
        return &clapgo_factory;
    }
    
    if (strcmp(factory_id, CLAP_PRESET_DISCOVERY_FACTORY_ID) == 0 ||
        strcmp(factory_id, CLAP_PRESET_DISCOVERY_FACTORY_ID_COMPAT) == 0) {
        printf("[PRESET_DEBUG] Returning preset discovery factory\n");
        if (log) {
            fprintf(log, "  Returning preset discovery factory for ID: %s\n", factory_id);
            fprintf(log, "  Factory address: %p\n", preset_discovery_get_factory());
            fclose(log);
        }
        return preset_discovery_get_factory();
    }
    
    if (strcmp(factory_id, CLAP_PLUGIN_INVALIDATION_FACTORY_ID) == 0) {
        printf("[INVALIDATION_DEBUG] Returning plugin invalidation factory\n");
        if (log) {
            fprintf(log, "  Returning plugin invalidation factory\n");
            fprintf(log, "  Factory address: %p\n", plugin_invalidation_get_factory());
            fclose(log);
        }
        return plugin_invalidation_get_factory();
    }
    
    if (strcmp(factory_id, CLAP_PLUGIN_STATE_CONVERTER_FACTORY_ID) == 0) {
        printf("[STATE_CONVERTER_DEBUG] Returning plugin state converter factory\n");
        if (log) {
            fprintf(log, "  Returning plugin state converter factory\n");
            fprintf(log, "  Factory address: %p\n", state_converter_get_factory());
            fclose(log);
        }
        return state_converter_get_factory();
    }
    
    printf("[PRESET_DEBUG] Unknown factory_id '%s', returning NULL\n", factory_id);
    if (log) {
        fprintf(log, "  Unknown factory_id '%s', returning NULL\n", factory_id);
        fprintf(log, "  Supported factories:\n");
        fprintf(log, "    - %s (plugin factory)\n", CLAP_PLUGIN_FACTORY_ID);
        fprintf(log, "    - %s (preset discovery)\n", CLAP_PRESET_DISCOVERY_FACTORY_ID);
        fprintf(log, "    - %s (preset discovery compat)\n", CLAP_PRESET_DISCOVERY_FACTORY_ID_COMPAT);
        fclose(log);
    }
    return NULL;
}