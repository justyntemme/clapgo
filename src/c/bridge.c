#include "bridge.h"
#include "manifest.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <math.h>    // For log10() and pow()
#include <libgen.h>  // For dirname()
#include <unistd.h>  // For access()

// Include CLAP extension headers for constants
#include "../../include/clap/include/clap/ext/audio-ports-activation.h"

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

// State extension Go exports
extern bool ClapGo_PluginStateSave(void* plugin, void* stream);
extern bool ClapGo_PluginStateLoad(void* plugin, void* stream);

// Note ports extension Go exports - weak symbols to allow optional implementation
__attribute__((weak)) uint32_t ClapGo_PluginNotePortsCount(void* plugin, bool is_input);
__attribute__((weak)) bool ClapGo_PluginNotePortsGet(void* plugin, uint32_t index, bool is_input, void* info);

// Phase 3 extension Go exports - weak symbols to allow optional implementation
// Latency extension
__attribute__((weak)) uint32_t ClapGo_PluginLatencyGet(void* plugin);

// Tail extension  
__attribute__((weak)) uint32_t ClapGo_PluginTailGet(void* plugin);

// Log extension - host side, no plugin exports needed

// Timer support extension
__attribute__((weak)) void ClapGo_PluginOnTimer(void* plugin, uint64_t timer_id);

// Phase 5 extension Go exports - weak symbols to allow optional implementation
// Audio ports config extension
__attribute__((weak)) uint32_t ClapGo_PluginAudioPortsConfigCount(void* plugin);
__attribute__((weak)) bool ClapGo_PluginAudioPortsConfigGet(void* plugin, uint32_t index, void* config);
__attribute__((weak)) bool ClapGo_PluginAudioPortsConfigSelect(void* plugin, uint64_t config_id);
__attribute__((weak)) uint64_t ClapGo_PluginAudioPortsConfigCurrentConfig(void* plugin);
__attribute__((weak)) bool ClapGo_PluginAudioPortsConfigGetInfo(void* plugin, uint64_t config_id, uint32_t port_index, bool is_input, void* info);

// Surround extension  
__attribute__((weak)) bool ClapGo_PluginSurroundIsChannelMaskSupported(void* plugin, uint64_t channel_mask);
__attribute__((weak)) uint32_t ClapGo_PluginSurroundGetChannelMap(void* plugin, bool is_input, uint32_t port_index, uint8_t* channel_map, uint32_t channel_map_capacity);

// Voice info extension
__attribute__((weak)) bool ClapGo_PluginVoiceInfoGet(void* plugin, void* info);

// Phase 6 extension Go exports
// State context extension
__attribute__((weak)) bool ClapGo_PluginStateSaveWithContext(void* plugin, void* stream, uint32_t context_type);
__attribute__((weak)) bool ClapGo_PluginStateLoadWithContext(void* plugin, void* stream, uint32_t context_type);

// Preset load extension
__attribute__((weak)) bool ClapGo_PluginPresetLoadFromLocation(void* plugin, uint32_t location_kind, char* location, char* load_key);

// Track info extension
__attribute__((weak)) void ClapGo_PluginTrackInfoChanged(void* plugin);

// Param indication extension
__attribute__((weak)) void ClapGo_PluginParamIndicationSetMapping(void* plugin, uint64_t param_id, bool has_mapping, void* color, char* label, char* description);
__attribute__((weak)) void ClapGo_PluginParamIndicationSetAutomation(void* plugin, uint64_t param_id, uint32_t automation_state, void* color);

// Context menu extension
__attribute__((weak)) bool ClapGo_PluginContextMenuPopulate(void* plugin, uint32_t target_kind, uint64_t target_id, void* builder);
__attribute__((weak)) bool ClapGo_PluginContextMenuPerform(void* plugin, uint32_t target_kind, uint64_t target_id, uint64_t action_id);

// Remote controls extension
__attribute__((weak)) uint32_t ClapGo_PluginRemoteControlsCount(void* plugin);
__attribute__((weak)) bool ClapGo_PluginRemoteControlsGet(void* plugin, uint32_t page_index, void* page);

// Note name extension
__attribute__((weak)) uint32_t ClapGo_PluginNoteNameCount(void* plugin);
__attribute__((weak)) bool ClapGo_PluginNoteNameGet(void* plugin, uint32_t index, void* note_name);

// Ambisonic extension
__attribute__((weak)) bool ClapGo_PluginAmbisonicIsConfigSupported(void* plugin, void* config);
__attribute__((weak)) bool ClapGo_PluginAmbisonicGetConfig(void* plugin, bool is_input, uint32_t port_index, void* config);

// Audio ports activation extension
__attribute__((weak)) bool ClapGo_PluginAudioPortsActivationCanActivateWhileProcessing(void* plugin);
__attribute__((weak)) bool ClapGo_PluginAudioPortsActivationSetActive(void* plugin, bool is_input, uint32_t port_index, bool is_active, uint32_t sample_size);

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
    
    // Determine extension support based on available exports
    data->supports_params = (ClapGo_PluginParamsCount != NULL);
    data->supports_note_ports = (ClapGo_PluginNotePortsCount != NULL && 
                                 ClapGo_PluginNotePortsGet != NULL);
    data->supports_state = (ClapGo_PluginStateSave != NULL && 
                           ClapGo_PluginStateLoad != NULL);
    data->supports_latency = (ClapGo_PluginLatencyGet != NULL);
    data->supports_tail = (ClapGo_PluginTailGet != NULL);
    data->supports_timer = (ClapGo_PluginOnTimer != NULL);
    data->supports_audio_ports_config = (ClapGo_PluginAudioPortsConfigCount != NULL &&
                                         ClapGo_PluginAudioPortsConfigGet != NULL &&
                                         ClapGo_PluginAudioPortsConfigSelect != NULL);
    data->supports_surround = (ClapGo_PluginSurroundIsChannelMaskSupported != NULL &&
                              ClapGo_PluginSurroundGetChannelMap != NULL);
    data->supports_voice_info = (ClapGo_PluginVoiceInfoGet != NULL);
    data->supports_state_context = (ClapGo_PluginStateSaveWithContext != NULL &&
                                   ClapGo_PluginStateLoadWithContext != NULL);
    data->supports_preset_load = (ClapGo_PluginPresetLoadFromLocation != NULL);
    data->supports_track_info = (ClapGo_PluginTrackInfoChanged != NULL);
    data->supports_param_indication = (ClapGo_PluginParamIndicationSetMapping != NULL &&
                                      ClapGo_PluginParamIndicationSetAutomation != NULL);
    data->supports_context_menu = (ClapGo_PluginContextMenuPopulate != NULL &&
                                  ClapGo_PluginContextMenuPerform != NULL);
    data->supports_remote_controls = (ClapGo_PluginRemoteControlsCount != NULL &&
                                     ClapGo_PluginRemoteControlsGet != NULL);
    data->supports_note_name = (ClapGo_PluginNoteNameCount != NULL &&
                               ClapGo_PluginNoteNameGet != NULL);
    data->supports_ambisonic = (ClapGo_PluginAmbisonicIsConfigSupported != NULL &&
                               ClapGo_PluginAmbisonicGetConfig != NULL);
    data->supports_audio_ports_activation = (ClapGo_PluginAudioPortsActivationCanActivateWhileProcessing != NULL &&
                                            ClapGo_PluginAudioPortsActivationSetActive != NULL);
    
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
    
    // Mark as not loaded yet - descriptor will be created on demand
    manifest_plugins[0].loaded = false;
    manifest_plugins[0].descriptor = NULL;
    
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

// Forward declarations for state extension
static bool clapgo_state_save(const clap_plugin_t* plugin, const clap_ostream_t* stream);
static bool clapgo_state_load(const clap_plugin_t* plugin, const clap_istream_t* stream);

// State extension structure - GUARDRAILS compliant (full implementation)
static const clap_plugin_state_t s_state_extension = {
    .save = clapgo_state_save,
    .load = clapgo_state_load
};

// Forward declarations for state context extension
static bool clapgo_state_context_save(const clap_plugin_t* plugin, const clap_ostream_t* stream, uint32_t context_type);
static bool clapgo_state_context_load(const clap_plugin_t* plugin, const clap_istream_t* stream, uint32_t context_type);

// State context extension structure - GUARDRAILS compliant (full implementation)
static const clap_plugin_state_context_t s_state_context_extension = {
    .save = clapgo_state_context_save,
    .load = clapgo_state_context_load
};

// Forward declarations for note ports extension
static uint32_t clapgo_note_ports_count(const clap_plugin_t* plugin, bool is_input);
static bool clapgo_note_ports_get(const clap_plugin_t* plugin, uint32_t index, bool is_input, clap_note_port_info_t* info);

// Note ports extension structure - GUARDRAILS compliant (full implementation)
static const clap_plugin_note_ports_t s_note_ports_extension = {
    .count = clapgo_note_ports_count,
    .get = clapgo_note_ports_get
};

// Forward declarations for Phase 3 extensions
static uint32_t clapgo_latency_get(const clap_plugin_t* plugin);
static uint32_t clapgo_tail_get(const clap_plugin_t* plugin);
static void clapgo_timer_on_timer(const clap_plugin_t* plugin, clap_id timer_id);

// Latency extension structure
static const clap_plugin_latency_t s_latency_extension = {
    .get = clapgo_latency_get
};

// Tail extension structure
static const clap_plugin_tail_t s_tail_extension = {
    .get = clapgo_tail_get
};

// Timer support extension structure
static const clap_plugin_timer_support_t s_timer_support_extension = {
    .on_timer = clapgo_timer_on_timer
};

// Forward declarations for Phase 5 extensions
static uint32_t clapgo_audio_ports_config_count(const clap_plugin_t* plugin);
static bool clapgo_audio_ports_config_get(const clap_plugin_t* plugin, uint32_t index, clap_audio_ports_config_t* config);
static bool clapgo_audio_ports_config_select(const clap_plugin_t* plugin, clap_id config_id);
static clap_id clapgo_audio_ports_config_info_current_config(const clap_plugin_t* plugin);
static bool clapgo_audio_ports_config_info_get(const clap_plugin_t* plugin, clap_id config_id, uint32_t port_index, bool is_input, clap_audio_port_info_t* info);
static bool clapgo_surround_is_channel_mask_supported(const clap_plugin_t* plugin, uint64_t channel_mask);
static uint32_t clapgo_surround_get_channel_map(const clap_plugin_t* plugin, bool is_input, uint32_t port_index, uint8_t* channel_map, uint32_t channel_map_capacity);
static bool clapgo_voice_info_get(const clap_plugin_t* plugin, clap_voice_info_t* info);

// Audio ports config extension structure
static const clap_plugin_audio_ports_config_t s_audio_ports_config_extension = {
    .count = clapgo_audio_ports_config_count,
    .get = clapgo_audio_ports_config_get,
    .select = clapgo_audio_ports_config_select
};

// Audio ports config info extension structure
static const clap_plugin_audio_ports_config_info_t s_audio_ports_config_info_extension = {
    .current_config = clapgo_audio_ports_config_info_current_config,
    .get = clapgo_audio_ports_config_info_get
};

// Surround extension structure
static const clap_plugin_surround_t s_surround_extension = {
    .is_channel_mask_supported = clapgo_surround_is_channel_mask_supported,
    .get_channel_map = clapgo_surround_get_channel_map
};

// Voice info extension structure
static const clap_plugin_voice_info_t s_voice_info_extension = {
    .get = clapgo_voice_info_get
};

// Forward declaration for preset load extension
static bool clapgo_preset_load_from_location(const clap_plugin_t* plugin, uint32_t location_kind, const char* location, const char* load_key);

// Preset load extension structure
static const clap_plugin_preset_load_t s_preset_load_extension = {
    .from_location = clapgo_preset_load_from_location
};

// Forward declaration for track info extension
static void clapgo_track_info_changed(const clap_plugin_t* plugin);

// Track info extension structure
static const clap_plugin_track_info_t s_track_info_extension = {
    .changed = clapgo_track_info_changed
};

// Forward declarations for param indication extension
static void clapgo_param_indication_set_mapping(const clap_plugin_t* plugin, clap_id param_id, bool has_mapping, const clap_color_t* color, const char* label, const char* description);
static void clapgo_param_indication_set_automation(const clap_plugin_t* plugin, clap_id param_id, uint32_t automation_state, const clap_color_t* color);

// Param indication extension structure
static const clap_plugin_param_indication_t s_param_indication_extension = {
    .set_mapping = clapgo_param_indication_set_mapping,
    .set_automation = clapgo_param_indication_set_automation
};

// Forward declarations for context menu extension
static bool clapgo_context_menu_populate(const clap_plugin_t* plugin, const clap_context_menu_target_t* target, const clap_context_menu_builder_t* builder);
static bool clapgo_context_menu_perform(const clap_plugin_t* plugin, const clap_context_menu_target_t* target, clap_id action_id);

// Context menu extension structure
static const clap_plugin_context_menu_t s_context_menu_extension = {
    .populate = clapgo_context_menu_populate,
    .perform = clapgo_context_menu_perform
};

// Forward declarations for remote controls extension
static uint32_t clapgo_remote_controls_count(const clap_plugin_t* plugin);
static bool clapgo_remote_controls_get(const clap_plugin_t* plugin, uint32_t page_index, clap_remote_controls_page_t* page);

// Remote controls extension structure
static const clap_plugin_remote_controls_t s_remote_controls_extension = {
    .count = clapgo_remote_controls_count,
    .get = clapgo_remote_controls_get
};

// Forward declarations for note name extension
static uint32_t clapgo_note_name_count(const clap_plugin_t* plugin);
static bool clapgo_note_name_get(const clap_plugin_t* plugin, uint32_t index, clap_note_name_t* note_name);

// Note name extension structure
static const clap_plugin_note_name_t s_note_name_extension = {
    .count = clapgo_note_name_count,
    .get = clapgo_note_name_get
};

// Forward declarations for ambisonic extension
static bool clapgo_ambisonic_is_config_supported(const clap_plugin_t* plugin, const clap_ambisonic_config_t* config);
static bool clapgo_ambisonic_get_config(const clap_plugin_t* plugin, bool is_input, uint32_t port_index, clap_ambisonic_config_t* config);

// Ambisonic extension structure
static const clap_plugin_ambisonic_t s_ambisonic_extension = {
    .is_config_supported = clapgo_ambisonic_is_config_supported,
    .get_config = clapgo_ambisonic_get_config
};

// Forward declarations for audio ports activation extension
static bool clapgo_audio_ports_activation_can_activate_while_processing(const clap_plugin_t* plugin);
static bool clapgo_audio_ports_activation_set_active(const clap_plugin_t* plugin, bool is_input, uint32_t port_index, bool is_active, uint32_t sample_size);

// Audio ports activation extension structure
static const clap_plugin_audio_ports_activation_t s_audio_ports_activation_extension = {
    .can_activate_while_processing = clapgo_audio_ports_activation_can_activate_while_processing,
    .set_active = clapgo_audio_ports_activation_set_active
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
        if (data->supports_params) {
            printf("DEBUG: params extension requested and supported\n");
            return &s_params_extension;
        }
        printf("DEBUG: params extension requested but not supported\n");
        return NULL;
    }
    
    // Check if this is the state extension
    if (strcmp(id, CLAP_EXT_STATE) == 0) {
        if (data->supports_state) {
            printf("DEBUG: state extension requested and supported\n");
            return &s_state_extension;
        }
        printf("DEBUG: state extension requested but not supported\n");
        return NULL;
    }
    
    // Check if this is the state context extension
    if (strcmp(id, CLAP_EXT_STATE_CONTEXT) == 0) {
        if (data->supports_state_context) {
            printf("DEBUG: state context extension requested and supported\n");
            return &s_state_context_extension;
        }
        printf("DEBUG: state context extension requested but not supported\n");
        return NULL;
    }
    
    // Check if this is the audio ports extension
    if (strcmp(id, CLAP_EXT_AUDIO_PORTS) == 0) {
        printf("DEBUG: audio ports extension requested, always supported\n");
        // Audio ports are always supported for all plugins
        return &s_audio_ports_extension;
    }
    
    // Check if this is the note ports extension
    if (strcmp(id, CLAP_EXT_NOTE_PORTS) == 0) {
        if (data->supports_note_ports) {
            printf("DEBUG: note ports extension requested and supported\n");
            return &s_note_ports_extension;
        }
        printf("DEBUG: note ports extension requested but not supported\n");
        return NULL;
    }
    
    // Check if this is the latency extension
    if (strcmp(id, CLAP_EXT_LATENCY) == 0) {
        if (data->supports_latency) {
            printf("DEBUG: latency extension requested and supported\n");
            return &s_latency_extension;
        }
        printf("DEBUG: latency extension requested but not supported\n");
        return NULL;
    }
    
    // Check if this is the tail extension
    if (strcmp(id, CLAP_EXT_TAIL) == 0) {
        if (data->supports_tail) {
            printf("DEBUG: tail extension requested and supported\n");
            return &s_tail_extension;
        }
        printf("DEBUG: tail extension requested but not supported\n");
        return NULL;
    }
    
    // Check if this is the timer support extension
    if (strcmp(id, CLAP_EXT_TIMER_SUPPORT) == 0) {
        if (data->supports_timer) {
            printf("DEBUG: timer support extension requested and supported\n");
            return &s_timer_support_extension;
        }
        printf("DEBUG: timer support extension requested but not supported\n");
        return NULL;
    }
    
    // Check if this is the audio ports config extension
    if (strcmp(id, CLAP_EXT_AUDIO_PORTS_CONFIG) == 0) {
        if (data->supports_audio_ports_config) {
            printf("DEBUG: audio ports config extension requested and supported\n");
            return &s_audio_ports_config_extension;
        }
        printf("DEBUG: audio ports config extension requested but not supported\n");
        return NULL;
    }
    
    // Check if this is the audio ports config info extension
    if (strcmp(id, CLAP_EXT_AUDIO_PORTS_CONFIG_INFO) == 0 || 
        strcmp(id, CLAP_EXT_AUDIO_PORTS_CONFIG_INFO_COMPAT) == 0) {
        if (data->supports_audio_ports_config) {
            printf("DEBUG: audio ports config info extension requested and supported\n");
            return &s_audio_ports_config_info_extension;
        }
        printf("DEBUG: audio ports config info extension requested but not supported\n");
        return NULL;
    }
    
    // Check if this is the surround extension  
    if (strcmp(id, CLAP_EXT_SURROUND) == 0 || strcmp(id, CLAP_EXT_SURROUND_COMPAT) == 0) {
        if (data->supports_surround) {
            printf("DEBUG: surround extension requested and supported\n");
            return &s_surround_extension;
        }
        printf("DEBUG: surround extension requested but not supported\n");
        return NULL;
    }
    
    // Check if this is the voice info extension
    if (strcmp(id, CLAP_EXT_VOICE_INFO) == 0) {
        if (data->supports_voice_info) {
            printf("DEBUG: voice info extension requested and supported\n");
            return &s_voice_info_extension;
        }
        printf("DEBUG: voice info extension requested but not supported\n");
        return NULL;
    }
    
    // Check if this is the preset load extension
    if (strcmp(id, CLAP_EXT_PRESET_LOAD) == 0) {
        if (data->supports_preset_load) {
            printf("DEBUG: preset load extension requested and supported\n");
            return &s_preset_load_extension;
        }
        printf("DEBUG: preset load extension requested but not supported\n");
        return NULL;
    }
    
    // Check if this is the track info extension
    if (strcmp(id, CLAP_EXT_TRACK_INFO) == 0 || strcmp(id, CLAP_EXT_TRACK_INFO_COMPAT) == 0) {
        if (data->supports_track_info) {
            printf("DEBUG: track info extension requested and supported\n");
            return &s_track_info_extension;
        }
        printf("DEBUG: track info extension requested but not supported\n");
        return NULL;
    }
    
    // Check if this is the param indication extension
    if (strcmp(id, CLAP_EXT_PARAM_INDICATION) == 0 || strcmp(id, CLAP_EXT_PARAM_INDICATION_COMPAT) == 0) {
        if (data->supports_param_indication) {
            printf("DEBUG: param indication extension requested and supported\n");
            return &s_param_indication_extension;
        }
        printf("DEBUG: param indication extension requested but not supported\n");
        return NULL;
    }
    
    // Check if this is the context menu extension
    if (strcmp(id, CLAP_EXT_CONTEXT_MENU) == 0 || strcmp(id, CLAP_EXT_CONTEXT_MENU_COMPAT) == 0) {
        if (data->supports_context_menu) {
            printf("DEBUG: context menu extension requested and supported\n");
            return &s_context_menu_extension;
        }
        printf("DEBUG: context menu extension requested but not supported\n");
        return NULL;
    }
    
    // Check if this is the remote controls extension
    if (strcmp(id, CLAP_EXT_REMOTE_CONTROLS) == 0 || strcmp(id, CLAP_EXT_REMOTE_CONTROLS_COMPAT) == 0) {
        if (data->supports_remote_controls) {
            printf("DEBUG: remote controls extension requested and supported\n");
            return &s_remote_controls_extension;
        }
        printf("DEBUG: remote controls extension requested but not supported\n");
        return NULL;
    }
    
    // Check if this is the note name extension
    if (strcmp(id, CLAP_EXT_NOTE_NAME) == 0) {
        if (data->supports_note_name) {
            printf("DEBUG: note name extension requested and supported\n");
            return &s_note_name_extension;
        }
        printf("DEBUG: note name extension requested but not supported\n");
        return NULL;
    }
    
    // Check if this is the ambisonic extension
    if (strcmp(id, CLAP_EXT_AMBISONIC) == 0 || strcmp(id, CLAP_EXT_AMBISONIC_COMPAT) == 0) {
        if (data->supports_ambisonic) {
            printf("DEBUG: ambisonic extension requested and supported\n");
            return &s_ambisonic_extension;
        }
        printf("DEBUG: ambisonic extension requested but not supported\n");
        return NULL;
    }
    
    // Check if this is the audio ports activation extension
    if (strcmp(id, CLAP_EXT_AUDIO_PORTS_ACTIVATION) == 0 || strcmp(id, CLAP_EXT_AUDIO_PORTS_ACTIVATION_COMPAT) == 0) {
        if (data->supports_audio_ports_activation) {
            printf("DEBUG: audio ports activation extension requested and supported\n");
            return &s_audio_ports_activation_extension;
        }
        printf("DEBUG: audio ports activation extension requested but not supported\n");
        return NULL;
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

// State extension implementation - GUARDRAILS compliant (full implementation)

// Save plugin state to stream
static bool clapgo_state_save(const clap_plugin_t* plugin, const clap_ostream_t* stream) {
    if (!plugin || !stream) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Call into Go to save state
    return ClapGo_PluginStateSave(data->go_instance, (void*)stream);
}

// Load plugin state from stream
static bool clapgo_state_load(const clap_plugin_t* plugin, const clap_istream_t* stream) {
    if (!plugin || !stream) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Call into Go to load state
    return ClapGo_PluginStateLoad(data->go_instance, (void*)stream);
}

// Note ports extension implementation - GUARDRAILS compliant (full implementation)

// Get the number of note ports
static uint32_t clapgo_note_ports_count(const clap_plugin_t* plugin, bool is_input) {
    if (!plugin) return 0;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return 0;
    
    // Check if the function exists (not NULL due to weak symbol)
    if (!ClapGo_PluginNotePortsCount) {
        return 0;
    }
    
    // Call into Go to get note port count
    return ClapGo_PluginNotePortsCount(data->go_instance, is_input);
}

// Get note port info
static bool clapgo_note_ports_get(const clap_plugin_t* plugin, uint32_t index, bool is_input, clap_note_port_info_t* info) {
    if (!plugin || !info) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Check if the function exists (not NULL due to weak symbol)
    if (!ClapGo_PluginNotePortsGet) {
        return false;
    }
    
    // Call into Go to get note port info
    return ClapGo_PluginNotePortsGet(data->go_instance, index, is_input, info);
}

// Phase 3 Extension Implementations

// Latency extension implementation
static uint32_t clapgo_latency_get(const clap_plugin_t* plugin) {
    if (!plugin) return 0;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return 0;
    
    // Check if the function exists (not NULL due to weak symbol)
    if (!ClapGo_PluginLatencyGet) {
        return 0;
    }
    
    // Call into Go to get latency
    return ClapGo_PluginLatencyGet(data->go_instance);
}

// Tail extension implementation
static uint32_t clapgo_tail_get(const clap_plugin_t* plugin) {
    if (!plugin) return 0;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return 0;
    
    // Check if the function exists (not NULL due to weak symbol)
    if (!ClapGo_PluginTailGet) {
        return 0;
    }
    
    // Call into Go to get tail length
    return ClapGo_PluginTailGet(data->go_instance);
}

// Timer support extension implementation
static void clapgo_timer_on_timer(const clap_plugin_t* plugin, clap_id timer_id) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return;
    
    // Check if the function exists (not NULL due to weak symbol)
    if (!ClapGo_PluginOnTimer) {
        return;
    }
    
    // Call into Go to handle timer
    ClapGo_PluginOnTimer(data->go_instance, timer_id);
}

// Phase 5 Extension Implementations

// Audio ports config extension implementation
static uint32_t clapgo_audio_ports_config_count(const clap_plugin_t* plugin) {
    if (!plugin) return 0;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return 0;
    
    // Check if the function exists
    if (!ClapGo_PluginAudioPortsConfigCount) {
        return 0;
    }
    
    return ClapGo_PluginAudioPortsConfigCount(data->go_instance);
}

static bool clapgo_audio_ports_config_get(const clap_plugin_t* plugin, uint32_t index, clap_audio_ports_config_t* config) {
    if (!plugin || !config) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Check if the function exists
    if (!ClapGo_PluginAudioPortsConfigGet) {
        return false;
    }
    
    return ClapGo_PluginAudioPortsConfigGet(data->go_instance, index, config);
}

static bool clapgo_audio_ports_config_select(const clap_plugin_t* plugin, clap_id config_id) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Check if the function exists
    if (!ClapGo_PluginAudioPortsConfigSelect) {
        return false;
    }
    
    return ClapGo_PluginAudioPortsConfigSelect(data->go_instance, config_id);
}

// Audio ports config info extension implementation
static clap_id clapgo_audio_ports_config_info_current_config(const clap_plugin_t* plugin) {
    if (!plugin) return CLAP_INVALID_ID;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return CLAP_INVALID_ID;
    
    // Check if the function exists
    if (!ClapGo_PluginAudioPortsConfigCurrentConfig) {
        return CLAP_INVALID_ID;
    }
    
    return ClapGo_PluginAudioPortsConfigCurrentConfig(data->go_instance);
}

static bool clapgo_audio_ports_config_info_get(const clap_plugin_t* plugin, clap_id config_id, 
                                                uint32_t port_index, bool is_input, 
                                                clap_audio_port_info_t* info) {
    if (!plugin || !info) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Check if the function exists
    if (!ClapGo_PluginAudioPortsConfigGetInfo) {
        return false;
    }
    
    return ClapGo_PluginAudioPortsConfigGetInfo(data->go_instance, config_id, port_index, is_input, info);
}

// Surround extension implementation
static bool clapgo_surround_is_channel_mask_supported(const clap_plugin_t* plugin, uint64_t channel_mask) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Check if the function exists
    if (!ClapGo_PluginSurroundIsChannelMaskSupported) {
        return false;
    }
    
    return ClapGo_PluginSurroundIsChannelMaskSupported(data->go_instance, channel_mask);
}

static uint32_t clapgo_surround_get_channel_map(const clap_plugin_t* plugin, bool is_input, 
                                               uint32_t port_index, uint8_t* channel_map, 
                                               uint32_t channel_map_capacity) {
    if (!plugin || !channel_map) return 0;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return 0;
    
    // Check if the function exists
    if (!ClapGo_PluginSurroundGetChannelMap) {
        return 0;
    }
    
    return ClapGo_PluginSurroundGetChannelMap(data->go_instance, is_input, port_index, 
                                              channel_map, channel_map_capacity);
}

// Voice info extension implementation
static bool clapgo_voice_info_get(const clap_plugin_t* plugin, clap_voice_info_t* info) {
    if (!plugin || !info) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Check if the function exists
    if (!ClapGo_PluginVoiceInfoGet) {
        return false;
    }
    
    return ClapGo_PluginVoiceInfoGet(data->go_instance, info);
}

// State context extension implementation
static bool clapgo_state_context_save(const clap_plugin_t* plugin, const clap_ostream_t* stream, uint32_t context_type) {
    if (!plugin || !stream) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Check if the function exists
    if (!ClapGo_PluginStateSaveWithContext) {
        return false;
    }
    
    return ClapGo_PluginStateSaveWithContext(data->go_instance, (void*)stream, context_type);
}

static bool clapgo_state_context_load(const clap_plugin_t* plugin, const clap_istream_t* stream, uint32_t context_type) {
    if (!plugin || !stream) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Check if the function exists
    if (!ClapGo_PluginStateLoadWithContext) {
        return false;
    }
    
    return ClapGo_PluginStateLoadWithContext(data->go_instance, (void*)stream, context_type);
}

// Preset load extension implementation
static bool clapgo_preset_load_from_location(const clap_plugin_t* plugin, uint32_t location_kind, const char* location, const char* load_key) {
    if (!plugin || !location) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Check if the function exists
    if (!ClapGo_PluginPresetLoadFromLocation) {
        return false;
    }
    
    return ClapGo_PluginPresetLoadFromLocation(data->go_instance, location_kind, (char*)location, (char*)load_key);
}

// Track info extension implementation
static void clapgo_track_info_changed(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return;
    
    // Check if the function exists
    if (!ClapGo_PluginTrackInfoChanged) {
        return;
    }
    
    ClapGo_PluginTrackInfoChanged(data->go_instance);
}

// Param indication extension implementation
static void clapgo_param_indication_set_mapping(const clap_plugin_t* plugin, clap_id param_id, 
                                                bool has_mapping, const clap_color_t* color, 
                                                const char* label, const char* description) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return;
    
    // Check if the function exists
    if (!ClapGo_PluginParamIndicationSetMapping) {
        return;
    }
    
    ClapGo_PluginParamIndicationSetMapping(data->go_instance, param_id, has_mapping, 
                                           (void*)color, (char*)label, (char*)description);
}

static void clapgo_param_indication_set_automation(const clap_plugin_t* plugin, clap_id param_id,
                                                   uint32_t automation_state, const clap_color_t* color) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return;
    
    // Check if the function exists
    if (!ClapGo_PluginParamIndicationSetAutomation) {
        return;
    }
    
    ClapGo_PluginParamIndicationSetAutomation(data->go_instance, param_id, automation_state, (void*)color);
}

// Context menu extension implementation
static bool clapgo_context_menu_populate(const clap_plugin_t* plugin, const clap_context_menu_target_t* target, 
                                        const clap_context_menu_builder_t* builder) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Check if the function exists
    if (!ClapGo_PluginContextMenuPopulate) {
        return false;
    }
    
    // Extract target info with defaults for null target
    uint32_t target_kind = target ? target->kind : CLAP_CONTEXT_MENU_TARGET_KIND_GLOBAL;
    uint64_t target_id = target ? target->id : 0;
    
    return ClapGo_PluginContextMenuPopulate(data->go_instance, target_kind, target_id, (void*)builder);
}

static bool clapgo_context_menu_perform(const clap_plugin_t* plugin, const clap_context_menu_target_t* target, 
                                       clap_id action_id) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Check if the function exists
    if (!ClapGo_PluginContextMenuPerform) {
        return false;
    }
    
    // Extract target info with defaults for null target
    uint32_t target_kind = target ? target->kind : CLAP_CONTEXT_MENU_TARGET_KIND_GLOBAL;
    uint64_t target_id = target ? target->id : 0;
    
    return ClapGo_PluginContextMenuPerform(data->go_instance, target_kind, target_id, action_id);
}

// Remote controls extension implementation
static uint32_t clapgo_remote_controls_count(const clap_plugin_t* plugin) {
    if (!plugin) return 0;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return 0;
    
    // Check if the function exists
    if (!ClapGo_PluginRemoteControlsCount) {
        return 0;
    }
    
    return ClapGo_PluginRemoteControlsCount(data->go_instance);
}

static bool clapgo_remote_controls_get(const clap_plugin_t* plugin, uint32_t page_index, 
                                      clap_remote_controls_page_t* page) {
    if (!plugin || !page) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Check if the function exists
    if (!ClapGo_PluginRemoteControlsGet) {
        return false;
    }
    
    return ClapGo_PluginRemoteControlsGet(data->go_instance, page_index, page);
}

// Note name extension implementation
static uint32_t clapgo_note_name_count(const clap_plugin_t* plugin) {
    if (!plugin) return 0;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return 0;
    
    // Check if the function exists
    if (!ClapGo_PluginNoteNameCount) {
        return 0;
    }
    
    return ClapGo_PluginNoteNameCount(data->go_instance);
}

static bool clapgo_note_name_get(const clap_plugin_t* plugin, uint32_t index, 
                                clap_note_name_t* note_name) {
    if (!plugin || !note_name) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Check if the function exists
    if (!ClapGo_PluginNoteNameGet) {
        return false;
    }
    
    return ClapGo_PluginNoteNameGet(data->go_instance, index, note_name);
}

// Ambisonic extension implementation
static bool clapgo_ambisonic_is_config_supported(const clap_plugin_t* plugin, const clap_ambisonic_config_t* config) {
    if (!plugin || !config) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Check if the function exists
    if (!ClapGo_PluginAmbisonicIsConfigSupported) {
        return false;
    }
    
    return ClapGo_PluginAmbisonicIsConfigSupported(data->go_instance, (void*)config);
}

static bool clapgo_ambisonic_get_config(const clap_plugin_t* plugin, bool is_input, uint32_t port_index, clap_ambisonic_config_t* config) {
    if (!plugin || !config) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Check if the function exists
    if (!ClapGo_PluginAmbisonicGetConfig) {
        return false;
    }
    
    return ClapGo_PluginAmbisonicGetConfig(data->go_instance, is_input, port_index, config);
}

// Audio ports activation extension implementation
static bool clapgo_audio_ports_activation_can_activate_while_processing(const clap_plugin_t* plugin) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Check if the function exists
    if (!ClapGo_PluginAudioPortsActivationCanActivateWhileProcessing) {
        return false;
    }
    
    return ClapGo_PluginAudioPortsActivationCanActivateWhileProcessing(data->go_instance);
}

static bool clapgo_audio_ports_activation_set_active(const clap_plugin_t* plugin, bool is_input, uint32_t port_index, bool is_active, uint32_t sample_size) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Check if the function exists
    if (!ClapGo_PluginAudioPortsActivationSetActive) {
        return false;
    }
    
    return ClapGo_PluginAudioPortsActivationSetActive(data->go_instance, is_input, port_index, is_active, sample_size);
}