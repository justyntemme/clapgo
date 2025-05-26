#ifndef CLAPGO_PRESET_DISCOVERY_H
#define CLAPGO_PRESET_DISCOVERY_H

#include <clap/clap.h>
#include <clap/factory/preset-discovery.h>

// Provider data structure for storing plugin information
typedef struct {
    char plugin_id[256];        // From manifest
    char plugin_name[256];      // From manifest
    char vendor[256];           // From manifest
    const clap_preset_discovery_indexer_t* indexer;
} provider_data_t;

// Get the preset discovery factory
const clap_preset_discovery_factory_t* preset_discovery_get_factory(void);

#endif // CLAPGO_PRESET_DISCOVERY_H