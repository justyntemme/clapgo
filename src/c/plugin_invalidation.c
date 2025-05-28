#include "plugin_invalidation.h"
#include "bridge.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

// Forward declarations
static uint32_t invalidation_factory_count(const clap_plugin_invalidation_factory_t* factory);
static const clap_plugin_invalidation_source_t* invalidation_factory_get(
    const clap_plugin_invalidation_factory_t* factory, uint32_t index);
static bool invalidation_factory_refresh(const clap_plugin_invalidation_factory_t* factory);

// Storage for invalidation sources
#define MAX_INVALIDATION_SOURCES 16
static clap_plugin_invalidation_source_t invalidation_sources[MAX_INVALIDATION_SOURCES];
static char source_directories[MAX_INVALIDATION_SOURCES][512];
static char source_globs[MAX_INVALIDATION_SOURCES][64];
static uint32_t invalidation_source_count = 0;
static bool sources_initialized = false;

// Initialize invalidation sources based on plugin locations
static void initialize_invalidation_sources() {
    if (sources_initialized) {
        return;
    }
    
    invalidation_source_count = 0;
    
    // Add ~/.clap directory for user plugins
    const char* home = getenv("HOME");
    if (home) {
        snprintf(source_directories[invalidation_source_count], 
                sizeof(source_directories[invalidation_source_count]),
                "%s/.clap", home);
        snprintf(source_globs[invalidation_source_count],
                sizeof(source_globs[invalidation_source_count]),
                "*.json");
        
        invalidation_sources[invalidation_source_count].directory = source_directories[invalidation_source_count];
        invalidation_sources[invalidation_source_count].filename_glob = source_globs[invalidation_source_count];
        invalidation_sources[invalidation_source_count].recursive_scan = true;
        
        invalidation_source_count++;
    }
    
    // Add plugin development directory if it exists
    char dev_path[512];
    if (home) {
        snprintf(dev_path, sizeof(dev_path), "%s/Documents/code/clapgo/examples", home);
        struct stat st;
        if (stat(dev_path, &st) == 0 && S_ISDIR(st.st_mode)) {
            snprintf(source_directories[invalidation_source_count], 
                    sizeof(source_directories[invalidation_source_count]),
                    "%s", dev_path);
            snprintf(source_globs[invalidation_source_count],
                    sizeof(source_globs[invalidation_source_count]),
                    "*.json");
            
            invalidation_sources[invalidation_source_count].directory = source_directories[invalidation_source_count];
            invalidation_sources[invalidation_source_count].filename_glob = source_globs[invalidation_source_count];
            invalidation_sources[invalidation_source_count].recursive_scan = true;
            
            invalidation_source_count++;
        }
    }
    
    sources_initialized = true;
}

// Factory implementation
static uint32_t invalidation_factory_count(const clap_plugin_invalidation_factory_t* factory) {
    initialize_invalidation_sources();
    return invalidation_source_count;
}

static const clap_plugin_invalidation_source_t* invalidation_factory_get(
    const clap_plugin_invalidation_factory_t* factory, uint32_t index) {
    
    initialize_invalidation_sources();
    
    if (index >= invalidation_source_count) {
        return NULL;
    }
    
    return &invalidation_sources[index];
}

static bool invalidation_factory_refresh(const clap_plugin_invalidation_factory_t* factory) {
    // Refresh the plugin manifests
    // This would trigger a rescan of all plugin manifests
    // For now, we'll just reinitialize the bridge which will reload manifests
    
    extern void clapgo_reload_manifests(void);
    clapgo_reload_manifests();
    
    // Return true to indicate the factory can refresh without full reload
    // Return false if the plugin needs to be completely reloaded
    return true;
}

// Factory instance
static const clap_plugin_invalidation_factory_t plugin_invalidation_factory = {
    .count = invalidation_factory_count,
    .get = invalidation_factory_get,
    .refresh = invalidation_factory_refresh
};

// Public function to get the factory
const clap_plugin_invalidation_factory_t* plugin_invalidation_get_factory(void) {
    return &plugin_invalidation_factory;
}