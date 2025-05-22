#include "bridge.h"
#include "manifest.h"
// Note: removed plugin.h include as it's legacy
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <libgen.h>  // For dirname()
#include <unistd.h>  // For access()

// Simplified manifest plugin entry for self-contained plugins
typedef struct {
    plugin_manifest_t manifest;
    const clap_plugin_descriptor_t* descriptor;
    bool loaded;
} manifest_plugin_entry_t;

// Manifest plugin registry
manifest_plugin_entry_t manifest_plugins[MAX_PLUGIN_MANIFESTS];
int manifest_plugin_count = 0;






// Go functions are now statically linked - declare external functions
extern void* ClapGo_CreatePlugin(void* host, char* plugin_id);
extern bool ClapGo_PluginInit(void* plugin);
extern void ClapGo_PluginDestroy(void* plugin);
extern bool ClapGo_PluginActivate(void* plugin, double sample_rate, uint32_t min_frames, uint32_t max_frames);
extern void ClapGo_PluginDeactivate(void* plugin);
extern bool ClapGo_PluginStartProcessing(void* plugin);
extern void ClapGo_PluginStopProcessing(void* plugin);
extern void ClapGo_PluginReset(void* plugin);
extern int32_t ClapGo_PluginProcess(void* plugin, void* process);
extern void* ClapGo_PluginGetExtension(void* plugin, char* id);
extern void ClapGo_PluginOnMainThread(void* plugin);



// Helper function to check if a file exists
static bool file_exists(const char* path) {
    if (!path) return false;
    
    FILE* file = fopen(path, "r");
    if (file) {
        fclose(file);
        return true;
    }
    return false;
}


// Find manifest files for the plugin
int clapgo_find_manifests(const char* plugin_path) {
    printf("Searching for manifest for plugin: %s\n", plugin_path);
    
    // Clear existing manifest entries
    memset(manifest_plugins, 0, sizeof(manifest_plugins));
    manifest_plugin_count = 0;
    
    // Extract the plugin name from the path
    const char* plugin_file = strrchr(plugin_path, '/');
    if (!plugin_file) {
        plugin_file = plugin_path;
    } else {
        plugin_file++; // Skip the '/'
    }
    
    char plugin_name[256];
    size_t plugin_name_len = strlen(plugin_file);
    
    // Remove .clap extension
    if (plugin_name_len > 5 && strcmp(plugin_file + plugin_name_len - 5, ".clap") == 0) {
        strncpy(plugin_name, plugin_file, plugin_name_len - 5);
        plugin_name[plugin_name_len - 5] = '\0';
    } else {
        strncpy(plugin_name, plugin_file, sizeof(plugin_name) - 1);
        plugin_name[sizeof(plugin_name) - 1] = '\0';
    }
    
    printf("Extracted plugin name: %s\n", plugin_name);
    
    char manifest_path[512];
    bool manifest_found = false;
    
    // First try: same directory as plugin (for development/testing)
    char* plugin_path_copy = strdup(plugin_path);
    char* plugin_dir = dirname(plugin_path_copy);
    snprintf(manifest_path, sizeof(manifest_path), "%s/%s.json", plugin_dir, plugin_name);
    printf("Looking for manifest at: %s\n", manifest_path);
    
    if (access(manifest_path, R_OK) == 0) {
        manifest_found = true;
    } else {
        // Second try: installed location ~/.clap/plugins/$PLUGIN/$PLUGIN.json
        char* home = getenv("HOME");
        if (home) {
            snprintf(manifest_path, sizeof(manifest_path), "%s/.clap/plugins/%s/%s.json", home, plugin_name, plugin_name);
            printf("Looking for manifest at: %s\n", manifest_path);
            
            if (access(manifest_path, R_OK) == 0) {
                manifest_found = true;
            } else {
                // Third try: legacy location ~/.clap/$PLUGIN.json
                snprintf(manifest_path, sizeof(manifest_path), "%s/.clap/%s.json", home, plugin_name);
                printf("Looking for manifest at: %s\n", manifest_path);
                
                if (access(manifest_path, R_OK) == 0) {
                    manifest_found = true;
                }
            }
        }
    }
    
    if (manifest_found) {
        // Load the manifest
        if (manifest_load_from_file(manifest_path, &manifest_plugins[0].manifest)) {
            printf("Loaded manifest: %s\n", manifest_path);
            manifest_plugins[0].loaded = false;
            manifest_plugins[0].library = NULL;
            manifest_plugins[0].descriptor = NULL;
            manifest_plugin_count = 1;
        } else {
            fprintf(stderr, "Error: Failed to load manifest from %s\n", manifest_path);
        }
    } else {
        fprintf(stderr, "Error: No manifest file found for plugin %s\n", plugin_name);
        fprintf(stderr, "Searched locations:\n");
        fprintf(stderr, "  - Development: %s/%s.json\n", plugin_dir, plugin_name);
        if (getenv("HOME")) {
            fprintf(stderr, "  - Installed: %s/.clap/plugins/%s/%s.json\n", getenv("HOME"), plugin_name, plugin_name);
            fprintf(stderr, "  - Legacy: %s/.clap/%s.json\n", getenv("HOME"), plugin_name);
        }
    }
    
    free(plugin_path_copy);
    
    // Return the number of manifests found (should be 0 or 1)
    return manifest_plugin_count;
}

// Load a manifest plugin by index (simplified for self-contained plugins)
bool clapgo_load_manifest_plugin(int index) {
    if (index < 0 || index >= manifest_plugin_count) {
        fprintf(stderr, "Error: Invalid manifest index: %d\n", index);
        return false;
    }
    
    manifest_plugin_entry_t* entry = &manifest_plugins[index];
    
    // Check if already loaded
    if (entry->loaded && entry->descriptor != NULL) {
        return true;
    }
    
    printf("Loading self-contained plugin: %s\n", entry->manifest.plugin.id);
    
    // Create the descriptor from the manifest
    entry->descriptor = manifest_to_descriptor(&entry->manifest);
    if (!entry->descriptor) {
        fprintf(stderr, "Error: Failed to create descriptor from manifest\n");
        return false;
    }
    
    // Mark as loaded
    entry->loaded = true;
    
    printf("Successfully loaded manifest plugin: %s (%s)\n", 
           entry->manifest.plugin.name, entry->manifest.plugin.id);
    
    return true;
}

// Find a manifest plugin by ID
int clapgo_find_manifest_plugin_by_id(const char* plugin_id) {
    if (!plugin_id) return -1;
    
    for (int i = 0; i < manifest_plugin_count; i++) {
        if (strcmp(manifest_plugins[i].manifest.plugin.id, plugin_id) == 0) {
            return i;
        }
    }
    
    return -1;
}

// Create a plugin instance from a manifest entry
const clap_plugin_t* clapgo_create_plugin_from_manifest(const clap_host_t* host, int index) {
    if (index < 0 || index >= manifest_plugin_count) {
        fprintf(stderr, "Error: Invalid manifest index: %d\n", index);
        return NULL;
    }
    
    manifest_plugin_entry_t* entry = &manifest_plugins[index];
    
    // Ensure the plugin is loaded
    if (!entry->loaded) {
        if (!clapgo_load_manifest_plugin(index)) {
            fprintf(stderr, "Error: Failed to load manifest plugin\n");
            return NULL;
        }
    }
    
    // Create the plugin instance using the statically linked Go function
    void* go_instance = ClapGo_CreatePlugin((void*)host, (char*)entry->manifest.plugin.id);
    if (!go_instance) {
        fprintf(stderr, "Error: Failed to create plugin instance\n");
        return NULL;
    }
    
    printf("Successfully created plugin instance for: %s\n", entry->manifest.plugin.id);
    
    // Allocate plugin instance data
    go_plugin_data_t* data = calloc(1, sizeof(go_plugin_data_t));
    if (!data) {
        fprintf(stderr, "Error: Failed to allocate plugin data\n");
        return NULL;
    }
    
    data->descriptor = entry->descriptor;
    data->go_instance = go_instance;
    data->manifest_index = index;
    
    // We no longer need global function pointers - we use the ones in the manifest entry
    
    // Allocate a CLAP plugin structure
    clap_plugin_t* plugin = calloc(1, sizeof(clap_plugin_t));
    if (!plugin) {
        fprintf(stderr, "Error: Failed to allocate plugin structure\n");
        free(data);
        return NULL;
    }
    
    // Initialize the plugin structure
    plugin->desc = entry->descriptor;
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

// Initialize the Go runtime and plugin environment
bool clapgo_init(const char* plugin_path) {
    printf("Initializing ClapGo plugin at path: %s\n", plugin_path);
    
    // Find manifest for this plugin
    int manifest_count = clapgo_find_manifests(plugin_path);
    
    // Manifests are mandatory - fail if no manifest found
    if (manifest_count == 0) {
        fprintf(stderr, "Error: No manifest file found for plugin %s\n", plugin_path);
        fprintf(stderr, "ClapGo requires a JSON manifest file with the same name as the plugin (plugin-name.json)\n");
        return false;
    }
    
    printf("Found manifest, using manifest-based loading\n");
    
    // Initialize the descriptor directly from the manifest
    manifest_plugins[0].descriptor = manifest_to_descriptor(&manifest_plugins[0].manifest);
    
    if (manifest_plugins[0].descriptor) {
        printf("Created descriptor from manifest: %s (%s)\n", 
               manifest_plugins[0].descriptor->name, manifest_plugins[0].descriptor->id);
    } else {
        fprintf(stderr, "Error: Failed to create descriptor from manifest\n");
        return false;
    }
    
    return true;
}

// Cleanup the Go runtime and plugin environment
void clapgo_deinit(void) {
    printf("Deinitializing ClapGo plugin\n");
    
    // First clean up any manifest plugins
    for (int i = 0; i < manifest_plugin_count; i++) {
        if (manifest_plugins[i].loaded && manifest_plugins[i].library) {
            // Free the descriptor
            if (manifest_plugins[i].descriptor) {
                // Free features array if allocated
                if (manifest_plugins[i].descriptor->features) {
                    // Free each feature string
                    for (int j = 0; manifest_plugins[i].descriptor->features[j] != NULL; j++) {
                        free((void*)manifest_plugins[i].descriptor->features[j]);
                    }
                    free((void*)manifest_plugins[i].descriptor->features);
                }
                
                // Free all string fields
                if (manifest_plugins[i].descriptor->id) free((void*)manifest_plugins[i].descriptor->id);
                if (manifest_plugins[i].descriptor->name) free((void*)manifest_plugins[i].descriptor->name);
                if (manifest_plugins[i].descriptor->vendor) free((void*)manifest_plugins[i].descriptor->vendor);
                if (manifest_plugins[i].descriptor->url) free((void*)manifest_plugins[i].descriptor->url);
                if (manifest_plugins[i].descriptor->manual_url) free((void*)manifest_plugins[i].descriptor->manual_url);
                if (manifest_plugins[i].descriptor->support_url) free((void*)manifest_plugins[i].descriptor->support_url);
                if (manifest_plugins[i].descriptor->version) free((void*)manifest_plugins[i].descriptor->version);
                if (manifest_plugins[i].descriptor->description) free((void*)manifest_plugins[i].descriptor->description);
                
                // Free the descriptor itself
                free((void*)manifest_plugins[i].descriptor);
                manifest_plugins[i].descriptor = NULL;
            }
            
            // Unload the library (only if it's not self-contained)
            if (manifest_plugins[i].library != (void*)1) {
                #if defined(CLAPGO_OS_WINDOWS)
                    FreeLibrary(manifest_plugins[i].library);
                #elif defined(CLAPGO_OS_MACOS) || defined(CLAPGO_OS_LINUX)
                    dlclose(manifest_plugins[i].library);
                #endif
            }
            
            manifest_plugins[i].library = NULL;
            manifest_plugins[i].loaded = false;
            
            // Free manifest resources
            manifest_free(&manifest_plugins[i].manifest);
        }
    }
    
    manifest_plugin_count = 0;
    
    printf("ClapGo plugin deinitialized successfully\n");
}

// Create a plugin instance for the given ID
const clap_plugin_t* clapgo_create_plugin(const clap_host_t* host, const char* plugin_id) {
    printf("Creating plugin with ID: %s\n", plugin_id);
    
    // Check if we have this plugin in the manifest registry
    int manifest_index = clapgo_find_manifest_plugin_by_id(plugin_id);
    if (manifest_index < 0) {
        fprintf(stderr, "Error: Plugin ID not found in manifest registry: %s\n", plugin_id);
        return NULL;
    }
    
    printf("Found plugin in manifest registry at index %d\n", manifest_index);
    
    // Load it from the manifest
    const clap_plugin_t* plugin = clapgo_create_plugin_from_manifest(host, manifest_index);
    if (plugin) {
        printf("Successfully created plugin instance from manifest\n");
        return plugin;
    }
    
    fprintf(stderr, "Error: Failed to create plugin from manifest\n");
    return NULL;
}

// Get the number of available plugins
uint32_t clapgo_get_plugin_count(void) {
    // Return the number of manifest plugins we've loaded
    return manifest_plugin_count;
}

// Get the plugin descriptor at the given index
const clap_plugin_descriptor_t* clapgo_get_plugin_descriptor(uint32_t index) {
    if (index >= manifest_plugin_count) return NULL;
    
    // Return the descriptor from the manifest entry
    return manifest_plugins[index].descriptor;
}

// Plugin callback implementations

// Initialize a plugin instance
bool clapgo_plugin_init(const clap_plugin_t* plugin) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Get the manifest entry for this plugin
    manifest_plugin_entry_t* entry = &manifest_plugins[data->manifest_index];
    if (!entry->loaded) {
        fprintf(stderr, "Error: Plugin entry not loaded\n");
        return false;
    }
    
    // Call into Go code to initialize the plugin instance using the function pointer from the manifest
    if (entry->plugin_init != NULL) {
        return entry->plugin_init(data->go_instance);
    }
    
    fprintf(stderr, "Error: Go plugin init function not available\n");
    return false;
}

// Destroy a plugin instance
void clapgo_plugin_destroy(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data) return;
    
    // Get the manifest entry for this plugin
    manifest_plugin_entry_t* entry = &manifest_plugins[data->manifest_index];
    if (!entry->loaded) {
        fprintf(stderr, "Error: Plugin entry not loaded\n");
        goto cleanup;
    }
    
    // Call into Go code to clean up the plugin instance
    if (data->go_instance) {
        ClapGo_PluginDestroy(data->go_instance);
    }
    
cleanup:
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
    if (!data || !data->go_instance) return false;
    
    // Get the manifest entry for this plugin
    manifest_plugin_entry_t* entry = &manifest_plugins[data->manifest_index];
    if (!entry->loaded) {
        fprintf(stderr, "Error: Plugin entry not loaded\n");
        return false;
    }
    
    // Call into Go code to activate the plugin instance
    return ClapGo_PluginActivate(data->go_instance, sample_rate, min_frames, max_frames);
}

// Deactivate a plugin instance
void clapgo_plugin_deactivate(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return;
    
    // Get the manifest entry for this plugin
    manifest_plugin_entry_t* entry = &manifest_plugins[data->manifest_index];
    if (!entry->loaded) {
        fprintf(stderr, "Error: Plugin entry not loaded\n");
        return;
    }
    
    // Call into Go code to deactivate the plugin instance
    if (entry->plugin_deactivate != NULL) {
        entry->plugin_deactivate(data->go_instance);
    }
}

// Start processing
bool clapgo_plugin_start_processing(const clap_plugin_t* plugin) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Get the manifest entry for this plugin
    manifest_plugin_entry_t* entry = &manifest_plugins[data->manifest_index];
    if (!entry->loaded) {
        fprintf(stderr, "Error: Plugin entry not loaded\n");
        return false;
    }
    
    // Call into Go code to start processing
    if (entry->plugin_start_processing != NULL) {
        return entry->plugin_start_processing(data->go_instance);
    }
    
    fprintf(stderr, "Error: Go plugin start_processing function not available\n");
    return false;
}

// Stop processing
void clapgo_plugin_stop_processing(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return;
    
    // Get the manifest entry for this plugin
    manifest_plugin_entry_t* entry = &manifest_plugins[data->manifest_index];
    if (!entry->loaded) {
        fprintf(stderr, "Error: Plugin entry not loaded\n");
        return;
    }
    
    // Call into Go code to stop processing
    if (entry->plugin_stop_processing != NULL) {
        entry->plugin_stop_processing(data->go_instance);
    }
}

// Reset a plugin instance
void clapgo_plugin_reset(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return;
    
    // Get the manifest entry for this plugin
    manifest_plugin_entry_t* entry = &manifest_plugins[data->manifest_index];
    if (!entry->loaded) {
        fprintf(stderr, "Error: Plugin entry not loaded\n");
        return;
    }
    
    // Call into Go code to reset the plugin instance
    if (entry->plugin_reset != NULL) {
        entry->plugin_reset(data->go_instance);
    }
}

// Process audio
clap_process_status clapgo_plugin_process(const clap_plugin_t* plugin, 
                                         const clap_process_t* process) {
    if (!plugin || !process) return CLAP_PROCESS_ERROR;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return CLAP_PROCESS_ERROR;
    
    // Get the manifest entry for this plugin
    manifest_plugin_entry_t* entry = &manifest_plugins[data->manifest_index];
    if (!entry->loaded) {
        fprintf(stderr, "Error: Plugin entry not loaded\n");
        return CLAP_PROCESS_ERROR;
    }
    
    // Call into Go code to process audio
    if (entry->plugin_process != NULL) {
        return entry->plugin_process(data->go_instance, process);
    }
    
    return CLAP_PROCESS_ERROR;
}

// Audio Ports extension implementation
static const clap_plugin_audio_ports_t clapgo_audio_ports_extension = {
    .count = clapgo_audio_ports_count,
    .get = clapgo_audio_ports_get
};

// Get the number of audio ports
uint32_t clapgo_audio_ports_count(const clap_plugin_t* plugin, bool is_input) {
    if (!plugin) return 0;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return 0;
    
    // Call into Go code to get the audio ports count
    // This would need a Go function to be exported, for now we'll hardcode stereo
    return 1; // One stereo port
}

// Get info about an audio port
bool clapgo_audio_ports_get(const clap_plugin_t* plugin, uint32_t index, bool is_input, clap_audio_port_info_t* info) {
    if (!plugin || !info || index > 0) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Fill in default stereo port info
    memset(info, 0, sizeof(clap_audio_port_info_t));
    
    // Set port ID
    info->id = 0;
    
    // Set port name
    if (is_input) {
        strcpy(info->name, "Stereo In");
    } else {
        strcpy(info->name, "Stereo Out");
    }
    
    // Set port flags - just main for now
    info->flags = CLAP_AUDIO_PORT_IS_MAIN;
    
    // Set channel count
    info->channel_count = 2; // Stereo
    
    // Set port type (stereo)
    info->port_type = CLAP_PORT_STEREO;

    // Set in-place processing capability
    info->in_place_pair = CLAP_INVALID_ID; // Not using in-place processing for simplicity
    
    return true;
}

// Get an extension
const void* clapgo_plugin_get_extension(const clap_plugin_t* plugin, const char* id) {
    if (!plugin || !id) return NULL;
    
    // Handle audio ports extension directly
    if (strcmp(id, CLAP_EXT_AUDIO_PORTS) == 0) {
        return &clapgo_audio_ports_extension;
    }
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return NULL;
    
    // Get the manifest entry for this plugin
    manifest_plugin_entry_t* entry = &manifest_plugins[data->manifest_index];
    if (!entry->loaded) {
        fprintf(stderr, "Error: Plugin entry not loaded\n");
        return NULL;
    }
    
    // Call into Go code to get the extension interface
    if (entry->plugin_get_extension != NULL) {
        return entry->plugin_get_extension(data->go_instance, (char*)id);
    }
    
    return NULL;
}

// Execute on main thread
void clapgo_plugin_on_main_thread(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return;
    
    // Get the manifest entry for this plugin
    manifest_plugin_entry_t* entry = &manifest_plugins[data->manifest_index];
    if (!entry->loaded) {
        fprintf(stderr, "Error: Plugin entry not loaded\n");
        return;
    }
    
    // Call into Go code to execute on the main thread
    if (entry->plugin_on_main_thread != NULL) {
        entry->plugin_on_main_thread(data->go_instance);
    }
}