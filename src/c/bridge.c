#include "bridge.h"
#include "manifest.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <math.h>    // For log10() and pow()
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

// Parameter-related Go functions
extern uint32_t ClapGo_PluginParamsCount(void* plugin);
extern bool ClapGo_PluginParamsGetInfo(void* plugin, uint32_t index, void* info);
extern bool ClapGo_PluginParamsGetValue(void* plugin, uint32_t param_id, double* value);
extern bool ClapGo_PluginParamsValueToText(void* plugin, uint32_t param_id, double value, char* buffer, uint32_t size);
extern bool ClapGo_PluginParamsTextToValue(void* plugin, uint32_t param_id, char* text, double* value);
extern void ClapGo_PluginParamsFlush(void* plugin, void* in_events, void* out_events);

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
        // Second try: installed location ~/.clap/$PLUGIN/$PLUGIN.json
        char* home = getenv("HOME");
        if (home) {
            snprintf(manifest_path, sizeof(manifest_path), "%s/.clap/%s/%s.json", home, plugin_name, plugin_name);
            printf("Looking for manifest at: %s\n", manifest_path);
            
            if (access(manifest_path, R_OK) == 0) {
                manifest_found = true;
            }
        }
    }
    
    if (manifest_found) {
        // Load the manifest
        if (manifest_load_from_file(manifest_path, &manifest_plugins[0].manifest)) {
            printf("Loaded manifest: %s\n", manifest_path);
            manifest_plugins[0].loaded = false;
            manifest_plugins[0].descriptor = NULL;
            manifest_plugin_count = 1;
        } else {
            fprintf(stderr, "Error: Failed to load manifest from %s\n", manifest_path);
        }
    } else {
        fprintf(stderr, "Error: No manifest file found for plugin %s\n", plugin_name);
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
    
    // Clean up any manifest plugins
    for (int i = 0; i < manifest_plugin_count; i++) {
        if (manifest_plugins[i].loaded) {
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
            
            manifest_plugins[i].loaded = false;
            
            // Free manifest resources
            manifest_free(&manifest_plugins[i].manifest);
        }
    }
    
    manifest_plugin_count = 0;
    
    printf("ClapGo plugin deinitialized successfully\n");
}

// Get the plugin descriptor at the given index
const clap_plugin_descriptor_t* clapgo_get_plugin_descriptor(uint32_t index) {
    if (index >= manifest_plugin_count) return NULL;
    
    manifest_plugin_entry_t* entry = &manifest_plugins[index];
    
    if (!entry->loaded) {
        if (!clapgo_load_manifest_plugin((int)index)) {
            return NULL;
        }
    }
    
    return entry->descriptor;
}

// Get plugin count
uint32_t clapgo_get_plugin_count(void) {
    return (uint32_t)manifest_plugin_count;
}

// Create a plugin by ID (CLAP factory interface)
const clap_plugin_t* clapgo_create_plugin(const clap_host_t* host, const char* plugin_id) {
    if (!plugin_id) {
        fprintf(stderr, "Error: Plugin ID is NULL\n");
        return NULL;
    }
    
    // Find the plugin by ID
    int index = clapgo_find_manifest_plugin_by_id(plugin_id);
    if (index < 0) {
        fprintf(stderr, "Error: Plugin ID not found in manifest registry: %s\n", plugin_id);
        return NULL;
    }
    
    // Create the plugin using the manifest entry
    return clapgo_create_plugin_from_manifest(host, index);
}

// Initialize a plugin instance
bool clapgo_plugin_init(const clap_plugin_t* plugin) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Call into Go code to initialize the plugin instance
    return ClapGo_PluginInit(data->go_instance);
}

// Destroy a plugin instance
void clapgo_plugin_destroy(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data) return;
    
    // Call into Go code to clean up the plugin instance
    if (data->go_instance) {
        ClapGo_PluginDestroy(data->go_instance);
    }
    
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
    
    // Call into Go code to activate the plugin instance
    return ClapGo_PluginActivate(data->go_instance, sample_rate, min_frames, max_frames);
}

// Deactivate a plugin instance
void clapgo_plugin_deactivate(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return;
    
    // Call into Go code to deactivate the plugin instance
    ClapGo_PluginDeactivate(data->go_instance);
}

// Start processing
bool clapgo_plugin_start_processing(const clap_plugin_t* plugin) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Call into Go code to start processing
    return ClapGo_PluginStartProcessing(data->go_instance);
}

// Stop processing
void clapgo_plugin_stop_processing(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return;
    
    // Call into Go code to stop processing
    ClapGo_PluginStopProcessing(data->go_instance);
}

// Reset a plugin instance
void clapgo_plugin_reset(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return;
    
    // Call into Go code to reset the plugin instance
    ClapGo_PluginReset(data->go_instance);
}

// Process audio through a plugin instance
clap_process_status clapgo_plugin_process(const clap_plugin_t* plugin, const clap_process_t* process) {
    if (!plugin || !process) return CLAP_PROCESS_ERROR;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return CLAP_PROCESS_ERROR;
    
    // Call into Go code to process audio
    return ClapGo_PluginProcess(data->go_instance, (void*)process);
}

// Audio ports extension implementation - GUARDRAILS compliant (full implementation)
static const clap_plugin_audio_ports_t s_audio_ports_extension = {
    .count = clapgo_audio_ports_count,
    .get = clapgo_audio_ports_info
};

// Forward declarations for params extension
static uint32_t clapgo_params_count(const clap_plugin_t* plugin);
static bool clapgo_params_get_info(const clap_plugin_t* plugin, uint32_t param_index, clap_param_info_t* param_info);
static bool clapgo_params_get_value(const clap_plugin_t* plugin, clap_id param_id, double* out_value);
static bool clapgo_params_value_to_text(const clap_plugin_t* plugin, clap_id param_id, double value, char* out_buffer, uint32_t out_buffer_capacity);
static bool clapgo_params_text_to_value(const clap_plugin_t* plugin, clap_id param_id, const char* param_value_text, double* out_value);
static void clapgo_params_flush(const clap_plugin_t* plugin, const clap_input_events_t* in, const clap_output_events_t* out);

// Params extension implementation - GUARDRAILS compliant (full implementation)
static const clap_plugin_params_t s_params_extension = {
    .count = clapgo_params_count,
    .get_info = clapgo_params_get_info,
    .get_value = clapgo_params_get_value,
    .value_to_text = clapgo_params_value_to_text,
    .text_to_value = clapgo_params_text_to_value,
    .flush = clapgo_params_flush
};

// Get the number of audio ports
uint32_t clapgo_audio_ports_count(const clap_plugin_t* plugin, bool is_input) {
    (void)plugin; // Suppress unused parameter warning
    (void)is_input; // Suppress unused parameter warning
    
    // Return 1 audio port (stereo in and out)
    return 1;
}

// Get audio port info
bool clapgo_audio_ports_info(const clap_plugin_t* plugin, uint32_t index, bool is_input,
                            clap_audio_port_info_t* info) {
    (void)plugin; // Suppress unused parameter warning
    
    if (index != 0 || !info) {
        return false;
    }
    
    // Configure the default stereo port
    info->id = 0;
    strcpy(info->name, is_input ? "Audio Input" : "Audio Output");
    info->flags = CLAP_AUDIO_PORT_IS_MAIN;
    info->channel_count = 2;
    info->port_type = CLAP_PORT_STEREO;
    // For in-place processing: both input and output ports reference each other with ID 0
    info->in_place_pair = 0;
    
    return true;
}

// Get an extension from a plugin instance
const void* clapgo_plugin_get_extension(const clap_plugin_t* plugin, const char* id) {
    if (!plugin || !id) return NULL;
    
    printf("DEBUG: get_extension called for id: %s\n", id);
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return NULL;
    
    // Check if this is the params extension
    if (strcmp(id, CLAP_EXT_PARAMS) == 0) {
        printf("DEBUG: params extension requested, returning implementation\n");
        return &s_params_extension;
    }
    
    // Check if this is the state extension
    if (strcmp(id, CLAP_EXT_STATE) == 0) {
        printf("DEBUG: state extension requested - not yet supported\n");
        return NULL;
    }
    
    // Check if this is the audio ports extension
    if (strcmp(id, CLAP_EXT_AUDIO_PORTS) == 0) {
        printf("DEBUG: audio ports extension requested, returning implementation\n");
        // Return the fully implemented audio ports extension
        return &s_audio_ports_extension;
    }
    
    // Call into Go code to get the extension
    void* ext = ClapGo_PluginGetExtension(data->go_instance, (char*)id);
    printf("DEBUG: Go returned extension: %p\n", ext);
    return ext;
}

// Handle main thread tasks for a plugin instance
void clapgo_plugin_on_main_thread(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return;
    
    // Call into Go code to handle main thread tasks
    ClapGo_PluginOnMainThread(data->go_instance);
}

// Params extension implementation

// Get the number of parameters
static uint32_t clapgo_params_count(const clap_plugin_t* plugin) {
    if (!plugin) return 0;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return 0;
    
    // Call into Go to get parameter count
    return ClapGo_PluginParamsCount(data->go_instance);
}

// Get parameter info
static bool clapgo_params_get_info(const clap_plugin_t* plugin, uint32_t param_index, clap_param_info_t* param_info) {
    if (!plugin || !param_info) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Call into Go to get parameter info
    return ClapGo_PluginParamsGetInfo(data->go_instance, param_index, param_info);
}

// Get parameter value
static bool clapgo_params_get_value(const clap_plugin_t* plugin, clap_id param_id, double* out_value) {
    if (!plugin || !out_value) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Call into Go to get parameter value
    return ClapGo_PluginParamsGetValue(data->go_instance, param_id, out_value);
}

// Convert parameter value to text
static bool clapgo_params_value_to_text(const clap_plugin_t* plugin, clap_id param_id, double value, char* out_buffer, uint32_t out_buffer_capacity) {
    if (!plugin || !out_buffer || out_buffer_capacity == 0) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Call into Go to convert value to text
    return ClapGo_PluginParamsValueToText(data->go_instance, param_id, value, out_buffer, out_buffer_capacity);
}

// Convert text to parameter value
static bool clapgo_params_text_to_value(const clap_plugin_t* plugin, clap_id param_id, const char* param_value_text, double* out_value) {
    if (!plugin || !param_value_text || !out_value) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Call into Go to convert text to value
    return ClapGo_PluginParamsTextToValue(data->go_instance, param_id, (char*)param_value_text, out_value);
}

// Flush parameter changes
static void clapgo_params_flush(const clap_plugin_t* plugin, const clap_input_events_t* in, const clap_output_events_t* out) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return;
    
    // Call into Go to handle flush
    ClapGo_PluginParamsFlush(data->go_instance, (void*)in, (void*)out);
}