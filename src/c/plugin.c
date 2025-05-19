#include "plugin.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

// Import the Go shared library functions
// These are defined in the Go code and exported through CGO

// Plugin factory implementation
static uint32_t clapgo_factory_get_plugin_count(const clap_plugin_factory_t *factory);
static const clap_plugin_descriptor_t *clapgo_factory_get_plugin_descriptor(
    const clap_plugin_factory_t *factory, uint32_t index);
static const clap_plugin_t *clapgo_factory_create_plugin(
    const clap_plugin_factory_t *factory, const clap_host_t *host, const char *plugin_id);

// Plugin entry point
static bool clapgo_entry_init(const char *plugin_path);
static void clapgo_entry_deinit(void);
static const void *clapgo_entry_get_factory(const char *factory_id);

// Plugin factory instance
static const clap_plugin_factory_t clapgo_factory = {
    .get_plugin_count = clapgo_factory_get_plugin_count,
    .get_plugin_descriptor = clapgo_factory_get_plugin_descriptor,
    .create_plugin = clapgo_factory_create_plugin
};

// Plugin entry point instance
const CLAP_EXPORT clap_plugin_entry_t clap_entry = {
    .clap_version = CLAP_VERSION_INIT,
    .init = clapgo_entry_init,
    .deinit = clapgo_entry_deinit,
    .get_factory = clapgo_entry_get_factory
};

// Internal storage for plugin descriptors
static const clap_plugin_descriptor_t **plugin_descriptors = NULL;
static uint32_t plugin_count = 0;

// Track which descriptor fields are dynamically allocated
typedef struct {
    bool id;
    bool name;
    bool vendor;
    bool url;
    bool manual_url;
    bool support_url;
    bool version;
    bool description;
    bool features_array;
    bool features_strings;
} descriptor_allocated_fields_t;

// Array to track allocated fields for each descriptor
static descriptor_allocated_fields_t *descriptor_allocations = NULL;

// Create a deep copy of a string, returning NULL if input is NULL
static char* deep_copy_string(const char* str) {
    if (!str) return NULL;
    return strdup(str);
}

// Helper function to free a descriptor's dynamically allocated fields
static void free_descriptor_fields(const clap_plugin_descriptor_t* descriptor, descriptor_allocated_fields_t* alloc_info) {
    if (!descriptor || !alloc_info) return;
    
    // Free all string fields that were allocated
    if (alloc_info->id && descriptor->id) free((void*)descriptor->id);
    if (alloc_info->name && descriptor->name) free((void*)descriptor->name);
    if (alloc_info->vendor && descriptor->vendor) free((void*)descriptor->vendor);
    if (alloc_info->url && descriptor->url) free((void*)descriptor->url);
    if (alloc_info->manual_url && descriptor->manual_url) free((void*)descriptor->manual_url);
    if (alloc_info->support_url && descriptor->support_url) free((void*)descriptor->support_url);
    if (alloc_info->version && descriptor->version) free((void*)descriptor->version);
    if (alloc_info->description && descriptor->description) free((void*)descriptor->description);
    
    // Free features array and its contents if allocated
    if (alloc_info->features_array && descriptor->features) {
        if (alloc_info->features_strings) {
            // Free each feature string
            for (int i = 0; descriptor->features[i] != NULL; i++) {
                free((void*)descriptor->features[i]);
            }
        }
        free((void*)descriptor->features);
    }
}

// Helper function to create a deep copy of a descriptor
static clap_plugin_descriptor_t* create_descriptor_copy(const clap_plugin_descriptor_t* src, 
                                                       descriptor_allocated_fields_t* alloc_info) {
    if (!src) return NULL;
    
    // Initialize the allocation tracking
    memset(alloc_info, 0, sizeof(descriptor_allocated_fields_t));
    
    // Allocate the descriptor
    clap_plugin_descriptor_t* desc = calloc(1, sizeof(clap_plugin_descriptor_t));
    if (!desc) return NULL;
    
    // Copy the version info
    desc->clap_version = src->clap_version;
    
    // Create deep copies of all string fields
    desc->id = deep_copy_string(src->id);
    alloc_info->id = (desc->id != NULL);
    
    desc->name = deep_copy_string(src->name);
    alloc_info->name = (desc->name != NULL);
    
    desc->vendor = deep_copy_string(src->vendor);
    alloc_info->vendor = (desc->vendor != NULL);
    
    desc->url = deep_copy_string(src->url);
    alloc_info->url = (desc->url != NULL);
    
    desc->manual_url = deep_copy_string(src->manual_url);
    alloc_info->manual_url = (desc->manual_url != NULL);
    
    desc->support_url = deep_copy_string(src->support_url);
    alloc_info->support_url = (desc->support_url != NULL);
    
    desc->version = deep_copy_string(src->version);
    alloc_info->version = (desc->version != NULL);
    
    desc->description = deep_copy_string(src->description);
    alloc_info->description = (desc->description != NULL);
    
    // Copy features array if present
    if (src->features) {
        // Count the number of features
        int feature_count = 0;
        while (src->features[feature_count] != NULL) {
            feature_count++;
        }
        
        // Allocate the array (+1 for NULL terminator)
        const char** features = calloc(feature_count + 1, sizeof(char*));
        if (!features) {
            // Failed to allocate, clean up and return NULL
            free_descriptor_fields(desc, alloc_info);
            free(desc);
            return NULL;
        }
        
        alloc_info->features_array = true;
        alloc_info->features_strings = true;
        
        // Copy each feature string
        for (int i = 0; i < feature_count; i++) {
            features[i] = deep_copy_string(src->features[i]);
            if (!features[i] && src->features[i]) {
                // Failed to copy a feature, clean up and return NULL
                for (int j = 0; j < i; j++) {
                    free((void*)features[j]);
                }
                free(features);
                free_descriptor_fields(desc, alloc_info);
                free(desc);
                return NULL;
            }
        }
        
        // Set the features array
        desc->features = features;
    }
    
    return desc;
}

// Initialize the Go runtime and plugin environment
bool clapgo_init(const char* plugin_path) {
    // This function would need to initialize the Go runtime
    // In a real implementation, we might load the Go shared library here
    printf("Initializing ClapGo plugin at path: %s\n", plugin_path);
    
    // Get the number of plugins and allocate the descriptor array
    plugin_count = clapgo_get_plugin_count();
    if (plugin_count == 0) {
        printf("No plugins found\n");
        return false;
    }
    
    // Allocate the descriptor array
    plugin_descriptors = calloc(plugin_count, sizeof(clap_plugin_descriptor_t*));
    if (!plugin_descriptors) {
        printf("Failed to allocate plugin descriptor array\n");
        return false;
    }
    
    // Allocate the descriptor allocations array
    descriptor_allocations = calloc(plugin_count, sizeof(descriptor_allocated_fields_t));
    if (!descriptor_allocations) {
        printf("Failed to allocate descriptor allocations array\n");
        free(plugin_descriptors);
        plugin_descriptors = NULL;
        plugin_count = 0;
        return false;
    }
    
    // Create deep copies of all plugin descriptors
    for (uint32_t i = 0; i < plugin_count; i++) {
        const clap_plugin_descriptor_t* src_desc = clapgo_get_plugin_descriptor(i);
        if (!src_desc) {
            printf("Failed to get plugin descriptor for index %u\n", i);
            continue;
        }
        
        plugin_descriptors[i] = (const clap_plugin_descriptor_t*)create_descriptor_copy(
            src_desc, &descriptor_allocations[i]);
        
        if (!plugin_descriptors[i]) {
            printf("Failed to create descriptor copy for plugin %u\n", i);
        }
    }
    
    return true;
}

// Cleanup the Go runtime and plugin environment
void clapgo_deinit(void) {
    printf("Deinitializing ClapGo plugin\n");
    
    // Free plugin descriptors and their fields
    if (plugin_descriptors && descriptor_allocations) {
        for (uint32_t i = 0; i < plugin_count; i++) {
            if (plugin_descriptors[i]) {
                // Free all dynamically allocated fields
                free_descriptor_fields(plugin_descriptors[i], &descriptor_allocations[i]);
                
                // Free the descriptor itself
                free((void*)plugin_descriptors[i]);
                plugin_descriptors[i] = NULL;
            }
        }
        
        // Free the descriptor array
        free(plugin_descriptors);
        plugin_descriptors = NULL;
        
        // Free the allocations array
        free(descriptor_allocations);
        descriptor_allocations = NULL;
    }
    
    plugin_count = 0;
}

// Create a plugin instance for the given ID
const clap_plugin_t* clapgo_create_plugin(const clap_host_t* host, const char* plugin_id) {
    // Find the plugin descriptor
    const clap_plugin_descriptor_t* descriptor = NULL;
    for (uint32_t i = 0; i < plugin_count; i++) {
        if (strcmp(plugin_descriptors[i]->id, plugin_id) == 0) {
            descriptor = plugin_descriptors[i];
            break;
        }
    }
    
    if (!descriptor) {
        printf("Plugin ID not found: %s\n", plugin_id);
        return NULL;
    }
    
    // Allocate plugin instance data
    go_plugin_data_t* data = calloc(1, sizeof(go_plugin_data_t));
    if (!data) {
        printf("Failed to allocate plugin data\n");
        return NULL;
    }
    
    data->descriptor = descriptor;
    
    // Create the Go plugin instance
    // This would call into Go code through CGO
    // For now, we'll just set a placeholder that would be
    // populated with a real Go instance reference
    data->go_instance = NULL;  // This would be set by Go code
    
    // Allocate a CLAP plugin structure
    clap_plugin_t* plugin = calloc(1, sizeof(clap_plugin_t));
    if (!plugin) {
        printf("Failed to allocate plugin structure\n");
        free(data);
        return NULL;
    }
    
    // Initialize the plugin structure
    plugin->desc = descriptor;
    plugin->plugin_data = data;
    
    // Set plugin callbacks
    plugin->init = clapgo_plugin_init;
    plugin->destroy = clapgo_plugin_destroy;
    plugin->activate = clapgo_plugin_activate;
    plugin->deactivate = clapgo_plugin_deactivate;
    plugin->start_processing = clapgo_plugin_start_processing;
    plugin->stop_processing = clapgo_plugin_stop_processing;
    plugin->reset = clapgo_plugin_reset;
    plugin->process = clapgo_plugin_process;
    plugin->get_extension = clapgo_plugin_get_extension;
    plugin->on_main_thread = clapgo_plugin_on_main_thread;
    
    return plugin;
}

// Get the number of available plugins
uint32_t clapgo_get_plugin_count(void) {
    // This would call into Go code via CGO
    // In a real implementation, this would be provided by the Go code
    return 0;  // Placeholder
}

// Get the plugin descriptor at the given index
const clap_plugin_descriptor_t* clapgo_get_plugin_descriptor(uint32_t index) {
    // This would call into Go code via CGO
    // In a real implementation, this would be provided by the Go code
    return NULL;  // Placeholder
}

// Get the plugin factory
const clap_plugin_factory_t* clapgo_get_plugin_factory(void) {
    return &clapgo_factory;
}

// Initialize a plugin instance
bool clapgo_plugin_init(const clap_plugin_t* plugin) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data) return false;
    
    // Call into Go code to initialize the plugin instance
    // In a real implementation, this would call a Go function via CGO
    printf("Initializing plugin instance: %s\n", plugin->desc->name);
    
    return true;  // Placeholder
}

// Destroy a plugin instance
void clapgo_plugin_destroy(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data) return;
    
    // Call into Go code to clean up the plugin instance
    // In a real implementation, this would call a Go function via CGO
    printf("Destroying plugin instance: %s\n", plugin->desc->name);
    
    // Free the plugin data
    free(data);
    
    // Free the plugin structure
    free((void*)plugin);
}

// Activate a plugin instance
bool clapgo_plugin_activate(const clap_plugin_t* plugin, double sample_rate,
                           uint32_t min_frames, uint32_t max_frames) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data) return false;
    
    // Call into Go code to activate the plugin instance
    // In a real implementation, this would call a Go function via CGO
    printf("Activating plugin: %s (sample rate: %f, min frames: %u, max frames: %u)\n",
           plugin->desc->name, sample_rate, min_frames, max_frames);
    
    return true;  // Placeholder
}

// Deactivate a plugin instance
void clapgo_plugin_deactivate(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data) return;
    
    // Call into Go code to deactivate the plugin instance
    // In a real implementation, this would call a Go function via CGO
    printf("Deactivating plugin: %s\n", plugin->desc->name);
}

// Start processing
bool clapgo_plugin_start_processing(const clap_plugin_t* plugin) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data) return false;
    
    // Call into Go code to start processing
    // In a real implementation, this would call a Go function via CGO
    printf("Starting processing for plugin: %s\n", plugin->desc->name);
    
    return true;  // Placeholder
}

// Stop processing
void clapgo_plugin_stop_processing(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data) return;
    
    // Call into Go code to stop processing
    // In a real implementation, this would call a Go function via CGO
    printf("Stopping processing for plugin: %s\n", plugin->desc->name);
}

// Reset a plugin instance
void clapgo_plugin_reset(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data) return;
    
    // Call into Go code to reset the plugin instance
    // In a real implementation, this would call a Go function via CGO
    printf("Resetting plugin: %s\n", plugin->desc->name);
}

// Process audio
clap_process_status clapgo_plugin_process(const clap_plugin_t* plugin, 
                                         const clap_process_t* process) {
    if (!plugin || !process) return CLAP_PROCESS_ERROR;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data) return CLAP_PROCESS_ERROR;
    
    // Call into Go code to process audio
    // In a real implementation, this would call a Go function via CGO
    
    return CLAP_PROCESS_CONTINUE;  // Placeholder
}

// Get an extension
const void* clapgo_plugin_get_extension(const clap_plugin_t* plugin, const char* id) {
    if (!plugin || !id) return NULL;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data) return NULL;
    
    // Call into Go code to get the extension interface
    // In a real implementation, this would call a Go function via CGO
    printf("Getting extension for plugin %s: %s\n", plugin->desc->name, id);
    
    // Check if we have GUI extension support via the overridden function
    #ifdef CLAPGO_GUI_SUPPORT
    // Check if the symbol is available at runtime
    extern const void* clapgo_plugin_get_extension_with_gui(const clap_plugin_t* plugin, const char* id);
    if (clapgo_plugin_get_extension_with_gui) {
        return clapgo_plugin_get_extension_with_gui(plugin, id);
    }
    #endif
    
    return NULL;  // Placeholder
}

// Execute on main thread
void clapgo_plugin_on_main_thread(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data) return;
    
    // Call into Go code to execute on the main thread
    // In a real implementation, this would call a Go function via CGO
    printf("On main thread for plugin: %s\n", plugin->desc->name);
}

// Plugin factory implementation
static uint32_t clapgo_factory_get_plugin_count(const clap_plugin_factory_t *factory) {
    return plugin_count;
}

static const clap_plugin_descriptor_t *clapgo_factory_get_plugin_descriptor(
    const clap_plugin_factory_t *factory, uint32_t index) {
    if (index >= plugin_count) return NULL;
    return plugin_descriptors[index];
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
        return clapgo_get_plugin_factory();
    }
    return NULL;
}