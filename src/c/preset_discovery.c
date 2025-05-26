#include "preset_discovery.h"
#include "manifest.h"
#include "bridge.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <json-c/json.h>
#include <time.h>
#include <signal.h>
#include <unistd.h>

// Debug logging to file
static FILE* debug_log = NULL;
static bool signal_handlers_installed = false;

static void crash_handler(int sig) {
    if (debug_log) {
        fprintf(debug_log, "[PRESET_DEBUG] CRASH: Signal %d received in preset discovery\n", sig);
        fflush(debug_log);
        fclose(debug_log);
    }
    _exit(1);
}

static void debug_init() {
    if (!debug_log) {
        char log_path[512];
        const char* home = getenv("HOME");
        if (!home) home = "/tmp";
        snprintf(log_path, sizeof(log_path), "%s/clapgo_preset_debug.log", home);
        debug_log = fopen(log_path, "a");
        if (debug_log) {
            time_t now = time(NULL);
            fprintf(debug_log, "\n=== ClapGo Preset Discovery Debug Log - %s", ctime(&now));
            fflush(debug_log);
        }
        
        // Install signal handlers for crash detection
        if (!signal_handlers_installed) {
            signal(SIGSEGV, crash_handler);
            signal(SIGABRT, crash_handler);
            signal(SIGFPE, crash_handler);
            signal(SIGILL, crash_handler);
            signal_handlers_installed = true;
        }
    }
}

#define DEBUG_LOG(...) do { \
    debug_init(); \
    if (debug_log) { \
        fprintf(debug_log, "[PRESET_DEBUG] "); \
        fprintf(debug_log, __VA_ARGS__); \
        fprintf(debug_log, "\n"); \
        fflush(debug_log); \
    } \
    printf("[PRESET_DEBUG] "); \
    printf(__VA_ARGS__); \
    printf("\n"); \
} while(0)

// Forward declarations
static bool provider_init(const clap_preset_discovery_provider_t* provider);
static void provider_destroy(const clap_preset_discovery_provider_t* provider);
static bool provider_get_metadata(
    const clap_preset_discovery_provider_t* provider,
    uint32_t location_kind,
    const char* location,
    const clap_preset_discovery_metadata_receiver_t* receiver);
static const void* provider_get_extension(
    const clap_preset_discovery_provider_t* provider,
    const char* extension_id);

// Factory forward declarations
static uint32_t factory_count(const clap_preset_discovery_factory_t* factory);
static const clap_preset_discovery_provider_descriptor_t* factory_get_descriptor(
    const clap_preset_discovery_factory_t* factory, uint32_t index);
static const clap_preset_discovery_provider_t* factory_create(
    const clap_preset_discovery_factory_t* factory,
    const clap_preset_discovery_indexer_t* indexer,
    const char* provider_id);

// External manifest data from bridge.c
extern manifest_plugin_entry_t manifest_plugins[MAX_PLUGIN_MANIFESTS];
extern int manifest_plugin_count;

// Helper function to check if plugin has preset directory
static bool plugin_has_presets(const char* plugin_id) {
    if (!plugin_id || !*plugin_id) {
        DEBUG_LOG("plugin_has_presets: invalid plugin_id");
        return false;
    }
    
    char preset_path[512];
    const char* home = getenv("HOME");
    if (!home) home = "/tmp";
    
    // Extract simple plugin name from ID (e.g., "com.clapgo.gain" -> "gain")
    const char* simple_name = strrchr(plugin_id, '.');
    simple_name = simple_name ? simple_name + 1 : plugin_id;
    
    snprintf(preset_path, sizeof(preset_path), "%s/.clap/%s/presets", home, simple_name);
    
    struct stat st;
    bool exists = (stat(preset_path, &st) == 0 && S_ISDIR(st.st_mode));
    DEBUG_LOG("plugin_has_presets: checking path '%s' for plugin '%s' - exists: %s", 
              preset_path, plugin_id, exists ? "true" : "false");
    return exists;
}

// Provider implementation
static bool provider_init(const clap_preset_discovery_provider_t* provider) {
    DEBUG_LOG("provider_init() called");
    
    if (!provider || !provider->provider_data) {
        DEBUG_LOG("provider_init() NULL provider or provider_data");
        return false;
    }
    
    provider_data_t* data = (provider_data_t*)provider->provider_data;
    
    DEBUG_LOG("Provider data: plugin_id='%s', plugin_name='%s'", 
           data->plugin_id, data->plugin_name);
    DEBUG_LOG("Indexer: %p", data->indexer);
    
    if (!data->indexer) {
        DEBUG_LOG("provider_init() NULL indexer");
        return false;
    }
    
    // Check indexer function pointers
    if (!data->indexer->declare_filetype || !data->indexer->declare_location) {
        DEBUG_LOG("ERROR: Indexer has NULL function pointers: declare_filetype=%p, declare_location=%p", 
                 data->indexer ? data->indexer->declare_filetype : NULL, 
                 data->indexer ? data->indexer->declare_location : NULL);
        return false;
    }
    DEBUG_LOG("Indexer functions: declare_filetype=%p, declare_location=%p", 
             data->indexer->declare_filetype, data->indexer->declare_location);
    
    // Step 1: Declare JSON filetype
    clap_preset_discovery_filetype_t filetype = {
        .name = "JSON Preset",
        .description = "ClapGo JSON preset format",
        .file_extension = "json"
    };
    
    DEBUG_LOG("Declaring filetype: %s", filetype.name);
    if (!data->indexer->declare_filetype) {
        DEBUG_LOG("declare_filetype function pointer is NULL");
        return false;
    }
    
    if (!data->indexer->declare_filetype(data->indexer, &filetype)) {
        DEBUG_LOG("Failed to declare filetype");
        return false;
    }
    DEBUG_LOG("Successfully declared filetype");
    
    // Step 2: Declare preset location based on plugin ID
    char preset_path[512];
    const char* home = getenv("HOME");
    if (!home) {
        home = "/tmp";  // Fallback
    }
    
    // Extract simple plugin name from ID
    const char* simple_name = strrchr(data->plugin_id, '.');
    simple_name = simple_name ? simple_name + 1 : data->plugin_id;
    
    snprintf(preset_path, sizeof(preset_path), "%s/.clap/%s/presets", home, simple_name);
    
    DEBUG_LOG("Declaring location: %s", preset_path);
    
    // Verify path exists before declaring
    struct stat st;
    if (stat(preset_path, &st) != 0 || !S_ISDIR(st.st_mode)) {
        DEBUG_LOG("Preset path does not exist or is not directory: %s", preset_path);
        return false;
    }
    
    clap_preset_discovery_location_t location = {
        .flags = CLAP_PRESET_DISCOVERY_IS_FACTORY_CONTENT,  // Changed from USER_CONTENT
        .name = "Factory Presets",
        .kind = CLAP_PRESET_DISCOVERY_LOCATION_FILE,
        .location = preset_path
    };
    
    if (!data->indexer->declare_location) {
        DEBUG_LOG("declare_location function pointer is NULL");
        return false;
    }
    
    bool result = data->indexer->declare_location(data->indexer, &location);
    DEBUG_LOG("declare_location() returned %s", result ? "true" : "false");
    DEBUG_LOG("provider_init() returning %s", result ? "true" : "false");
    
    return result;
}

static void provider_destroy(const clap_preset_discovery_provider_t* provider) {
    DEBUG_LOG("provider_destroy() called: provider=%p", provider);
    if (provider) {
        if (provider->provider_data) {
            DEBUG_LOG("Freeing provider data");
            free((void*)provider->provider_data);
        }
        DEBUG_LOG("Freeing provider");
        free((void*)provider);
    }
}

static bool provider_get_metadata(
    const clap_preset_discovery_provider_t* provider,
    uint32_t location_kind,
    const char* location,
    const clap_preset_discovery_metadata_receiver_t* receiver) {
    
    DEBUG_LOG("provider_get_metadata() called with location: %s", location ? location : "NULL");
    
    if (!provider || !provider->provider_data || !location || !receiver) {
        DEBUG_LOG("provider_get_metadata() NULL parameter: provider=%p, provider_data=%p, location=%p, receiver=%p",
                 provider, provider ? provider->provider_data : NULL, location, receiver);
        return false;
    }
    
    provider_data_t* data = (provider_data_t*)provider->provider_data;
    DEBUG_LOG("Processing preset file for plugin: %s", data->plugin_id);
    
    // Check if file exists and is readable
    FILE* test_file = fopen(location, "r");
    if (!test_file) {
        DEBUG_LOG("Cannot open preset file: %s", location);
        return false;
    }
    fclose(test_file);
    
    // Parse JSON file
    struct json_object* root = json_object_from_file(location);
    if (!root) {
        DEBUG_LOG("Failed to parse JSON file: %s", location);
        return false;
    }
    
    DEBUG_LOG("Successfully parsed JSON file: %s", location);
    
    // Extract name (required)
    struct json_object* name_obj;
    if (!json_object_object_get_ex(root, "name", &name_obj)) {
        DEBUG_LOG("No 'name' field found in JSON");
        json_object_put(root);
        return false;
    }
    
    const char* preset_name = json_object_get_string(name_obj);
    DEBUG_LOG("Found preset name: %s", preset_name);
    
    DEBUG_LOG("Calling receiver->begin_preset() with name: %s", preset_name);
    if (!receiver->begin_preset(receiver, preset_name, NULL)) {
        DEBUG_LOG("receiver->begin_preset() failed");
        json_object_put(root);
        return false;
    }
    DEBUG_LOG("receiver->begin_preset() succeeded");
    
    // Add plugin ID (from manifest or JSON)
    struct json_object* plugin_ids_obj;
    if (json_object_object_get_ex(root, "plugin_ids", &plugin_ids_obj) && 
        json_object_is_type(plugin_ids_obj, json_type_array)) {
        // Use plugin IDs from preset
        int count = json_object_array_length(plugin_ids_obj);
        for (int i = 0; i < count; i++) {
            struct json_object* id_obj = json_object_array_get_idx(plugin_ids_obj, i);
            if (id_obj) {
                clap_universal_plugin_id_t plugin_id = {
                    .abi = "clap",
                    .id = json_object_get_string(id_obj)
                };
                receiver->add_plugin_id(receiver, &plugin_id);
            }
        }
    } else {
        // Fallback to manifest plugin ID
        clap_universal_plugin_id_t plugin_id = {
            .abi = "clap",
            .id = data->plugin_id
        };
        receiver->add_plugin_id(receiver, &plugin_id);
    }
    
    // Extract ALL other metadata from JSON (no hardcoding)
    struct json_object* obj;
    
    // Description
    if (json_object_object_get_ex(root, "description", &obj)) {
        receiver->set_description(receiver, json_object_get_string(obj));
    }
    
    // Creators
    if (json_object_object_get_ex(root, "creators", &obj) && 
        json_object_is_type(obj, json_type_array)) {
        int count = json_object_array_length(obj);
        for (int i = 0; i < count; i++) {
            struct json_object* creator = json_object_array_get_idx(obj, i);
            if (creator) {
                receiver->add_creator(receiver, json_object_get_string(creator));
            }
        }
    }
    
    // Features (from JSON, not hardcoded)
    if (json_object_object_get_ex(root, "features", &obj) && 
        json_object_is_type(obj, json_type_array)) {
        int count = json_object_array_length(obj);
        for (int i = 0; i < count; i++) {
            struct json_object* feature = json_object_array_get_idx(obj, i);
            if (feature) {
                receiver->add_feature(receiver, json_object_get_string(feature));
            }
        }
    }
    
    // Flags
    uint32_t flags = CLAP_PRESET_DISCOVERY_IS_USER_CONTENT;
    if (json_object_object_get_ex(root, "is_favorite", &obj)) {
        if (json_object_get_boolean(obj)) {
            flags |= CLAP_PRESET_DISCOVERY_IS_FAVORITE;
        }
    }
    receiver->set_flags(receiver, flags);
    
    // Set soundpack ID only if present in JSON
    if (json_object_object_get_ex(root, "soundpack_id", &obj)) {
        const char* soundpack_id = json_object_get_string(obj);
        if (soundpack_id && *soundpack_id && receiver->set_soundpack_id) {
            receiver->set_soundpack_id(receiver, soundpack_id);
        }
    }
    
    DEBUG_LOG("provider_get_metadata() completed successfully for preset: %s", preset_name);
    
    json_object_put(root);
    return true;
}

static const void* provider_get_extension(
    const clap_preset_discovery_provider_t* provider,
    const char* extension_id) {
    // No extensions supported for now
    return NULL;
}

// Static storage for all descriptors - allocate for maximum plugins
#define MAX_PRESET_PROVIDERS 32
static clap_preset_discovery_provider_descriptor_t provider_descriptors[MAX_PRESET_PROVIDERS];
static char provider_ids[MAX_PRESET_PROVIDERS][256];
static char provider_names[MAX_PRESET_PROVIDERS][256];
static bool descriptors_initialized = false;

// Initialize all descriptors once
static void initialize_descriptors() {
    if (descriptors_initialized) {
        return;
    }
    
    uint32_t provider_index = 0;
    for (int i = 0; i < manifest_plugin_count && provider_index < MAX_PRESET_PROVIDERS; i++) {
        if (plugin_has_presets(manifest_plugins[i].manifest.plugin.id)) {
            clap_preset_discovery_provider_descriptor_t* desc = &provider_descriptors[provider_index];
            
            desc->clap_version = (clap_version_t)CLAP_VERSION_INIT;
            
            snprintf(provider_ids[provider_index], sizeof(provider_ids[provider_index]), 
                    "%s.presets", manifest_plugins[i].manifest.plugin.id);
            desc->id = provider_ids[provider_index];
            
            snprintf(provider_names[provider_index], sizeof(provider_names[provider_index]), 
                    "%s Presets", manifest_plugins[i].manifest.plugin.name);
            desc->name = provider_names[provider_index];
            
            desc->vendor = manifest_plugins[i].manifest.plugin.vendor;
            
            provider_index++;
        }
    }
    
    descriptors_initialized = true;
}

// Factory implementation
static uint32_t factory_count(const clap_preset_discovery_factory_t* factory) {
    DEBUG_LOG("factory_count() called");
    
    initialize_descriptors();
    
    // Count plugins with preset directories from loaded manifests
    uint32_t count = 0;
    DEBUG_LOG("manifest_plugin_count = %d", manifest_plugin_count);
    
    for (int i = 0; i < manifest_plugin_count && count < MAX_PRESET_PROVIDERS; i++) {
        DEBUG_LOG("Checking plugin %d: %s", i, manifest_plugins[i].manifest.plugin.id);
        if (plugin_has_presets(manifest_plugins[i].manifest.plugin.id)) {
            DEBUG_LOG("Plugin %s has presets", manifest_plugins[i].manifest.plugin.id);
            count++;
        } else {
            DEBUG_LOG("Plugin %s has no presets", manifest_plugins[i].manifest.plugin.id);
        }
    }
    
    DEBUG_LOG("factory_count() returning %d", count);
    return count;
}

static const clap_preset_discovery_provider_descriptor_t* factory_get_descriptor(
    const clap_preset_discovery_factory_t* factory, uint32_t index) {
    
    DEBUG_LOG("factory_get_descriptor() called with index %d", index);
    
    initialize_descriptors();
    
    if (index >= MAX_PRESET_PROVIDERS) {
        DEBUG_LOG("Index %d >= MAX_PRESET_PROVIDERS (%d)", index, MAX_PRESET_PROVIDERS);
        return NULL;
    }
    
    // Find Nth plugin with presets
    uint32_t current = 0;
    for (int i = 0; i < manifest_plugin_count; i++) {
        if (plugin_has_presets(manifest_plugins[i].manifest.plugin.id)) {
            if (current == index) {
                DEBUG_LOG("Returning descriptor for index %d: %s", index, provider_descriptors[index].id);
                return &provider_descriptors[index];
            }
            current++;
        }
    }
    
    DEBUG_LOG("factory_get_descriptor() returning NULL for index %d", index);
    return NULL;
}

static const clap_preset_discovery_provider_t* factory_create(
    const clap_preset_discovery_factory_t* factory,
    const clap_preset_discovery_indexer_t* indexer,
    const char* provider_id) {
    
    DEBUG_LOG("factory_create() called with provider_id: %s", provider_id ? provider_id : "NULL");
    DEBUG_LOG("factory_create() indexer: %p", indexer);
    
    if (!provider_id || !indexer) {
        DEBUG_LOG("factory_create() NULL parameter: provider_id=%p, indexer=%p", provider_id, indexer);
        return NULL;
    }
    
    initialize_descriptors();
    
    // Find matching plugin from manifests and calculate correct provider index
    uint32_t provider_index = 0;
    for (int i = 0; i < manifest_plugin_count; i++) {
        if (plugin_has_presets(manifest_plugins[i].manifest.plugin.id)) {
            char expected_id[512];
            snprintf(expected_id, sizeof(expected_id), "%s.presets", 
                    manifest_plugins[i].manifest.plugin.id);
            
            DEBUG_LOG("Comparing provider_id '%s' with expected_id '%s'", provider_id, expected_id);
            
            if (strcmp(provider_id, expected_id) == 0) {
                DEBUG_LOG("Found matching provider, creating...");
                // Create provider
                clap_preset_discovery_provider_t* provider = 
                    calloc(1, sizeof(clap_preset_discovery_provider_t));
                if (!provider) {
                    DEBUG_LOG("Failed to allocate provider");
                    return NULL;
                }
                
                provider_data_t* data = calloc(1, sizeof(provider_data_t));
                if (!data) {
                    DEBUG_LOG("Failed to allocate provider data");
                    free(provider);
                    return NULL;
                }
                
                // Copy data from manifest
                strncpy(data->plugin_id, manifest_plugins[i].manifest.plugin.id, 
                       sizeof(data->plugin_id) - 1);
                data->plugin_id[sizeof(data->plugin_id) - 1] = '\0';  // Ensure null termination
                
                strncpy(data->plugin_name, manifest_plugins[i].manifest.plugin.name,
                       sizeof(data->plugin_name) - 1);
                data->plugin_name[sizeof(data->plugin_name) - 1] = '\0';
                
                strncpy(data->vendor, manifest_plugins[i].manifest.plugin.vendor,
                       sizeof(data->vendor) - 1);
                data->vendor[sizeof(data->vendor) - 1] = '\0';
                
                data->indexer = indexer;
                
                DEBUG_LOG("Setting provider data: plugin_id='%s', plugin_name='%s', vendor='%s'",
                         data->plugin_id, data->plugin_name, data->vendor);
                
                // Use the correct provider index, not manifest index
                provider->desc = &provider_descriptors[provider_index];
                provider->provider_data = data;
                provider->init = provider_init;
                provider->destroy = provider_destroy;
                provider->get_metadata = provider_get_metadata;
                provider->get_extension = provider_get_extension;
                
                DEBUG_LOG("Provider created successfully: %p", provider);
                return provider;
            }
            provider_index++;
        }
    }
    
    DEBUG_LOG("No matching provider found for ID: %s", provider_id);
    return NULL;
}

// Factory instance
static const clap_preset_discovery_factory_t preset_discovery_factory = {
    .count = factory_count,
    .get_descriptor = factory_get_descriptor,
    .create = factory_create
};

// Public function to get the factory
const clap_preset_discovery_factory_t* preset_discovery_get_factory(void) {
    DEBUG_LOG("preset_discovery_get_factory() called");
    return &preset_discovery_factory;
}