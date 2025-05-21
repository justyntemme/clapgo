#include "bridge.h"
#include "manifest.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <libgen.h>  // For dirname()
#include <unistd.h>  // For access()

// Manifest plugin entry structure (moved from the header)
typedef struct {
    plugin_manifest_t manifest;
    clapgo_library_t library;
    const clap_plugin_descriptor_t* descriptor;
    bool loaded;
    
    // Function pointers for plugin operations
    clapgo_create_plugin_func create_plugin;
    clapgo_get_version_func get_version;

    // Plugin metadata functions
    clapgo_export_plugin_id_func get_plugin_id;
    clapgo_export_plugin_name_func get_plugin_name;
    clapgo_export_plugin_vendor_func get_plugin_vendor;
    clapgo_export_plugin_version_func get_plugin_version;
    clapgo_export_plugin_description_func get_plugin_description;
    
    // Plugin lifecycle functions
    clapgo_plugin_init_func plugin_init;
    clapgo_plugin_destroy_func plugin_destroy;
    clapgo_plugin_activate_func plugin_activate;
    clapgo_plugin_deactivate_func plugin_deactivate;
    clapgo_plugin_start_processing_func plugin_start_processing;
    clapgo_plugin_stop_processing_func plugin_stop_processing;
    clapgo_plugin_reset_func plugin_reset;
    clapgo_plugin_process_func plugin_process;
    clapgo_plugin_get_extension_func plugin_get_extension;
    clapgo_plugin_on_main_thread_func plugin_on_main_thread;
} manifest_plugin_entry_t;

// Manifest plugin registry
manifest_plugin_entry_t manifest_plugins[MAX_PLUGIN_MANIFESTS];
int manifest_plugin_count = 0;

// We don't need these anymore as we're using the manifest_plugins array directly
// and manifest_plugin_count for the count


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

// Platform-specific shared library loading implementation
bool clapgo_load_library(const char* path) {
    if (clapgo_lib != NULL) {
        // Library already loaded
        return true;
    }
    
    if (path == NULL) {
        fprintf(stderr, "Error: Cannot load library, path is NULL\n");
        return false;
    }
    
#if defined(CLAPGO_OS_WINDOWS)
    // Windows implementation
    clapgo_lib = LoadLibraryA(path);
    if (clapgo_lib == NULL) {
        DWORD error = GetLastError();
        fprintf(stderr, "Error: Failed to load library: %s (error code: %lu)\n", path, error);
        return false;
    }
#elif defined(CLAPGO_OS_MACOS) || defined(CLAPGO_OS_LINUX)
    // macOS and Linux implementation (both use dlopen)
    clapgo_lib = dlopen(path, RTLD_LAZY | RTLD_LOCAL);
    if (clapgo_lib == NULL) {
        fprintf(stderr, "Error: Failed to load library: %s (%s)\n", path, dlerror());
        return false;
    }
#else
    #error "Unsupported platform"
#endif

    return true;
}

void clapgo_unload_library(void) {
    if (clapgo_lib == NULL) {
        return;
    }
    
#if defined(CLAPGO_OS_WINDOWS)
    FreeLibrary(clapgo_lib);
#elif defined(CLAPGO_OS_MACOS) || defined(CLAPGO_OS_LINUX)
    dlclose(clapgo_lib);
#endif

    clapgo_lib = NULL;
    
    // We no longer use global function pointers - they're stored in each manifest entry
}

clapgo_symbol_t clapgo_get_symbol(const char* name) {
    if (clapgo_lib == NULL || name == NULL) {
        return NULL;
    }
    
    clapgo_symbol_t symbol = NULL;
    
#if defined(CLAPGO_OS_WINDOWS)
    symbol = GetProcAddress(clapgo_lib, name);
    if (symbol == NULL) {
        DWORD error = GetLastError();
        fprintf(stderr, "Error: Failed to get symbol: %s (error code: %lu)\n", name, error);
    }
#elif defined(CLAPGO_OS_MACOS) || defined(CLAPGO_OS_LINUX)
    // Clear any existing errors
    dlerror();
    
    symbol = dlsym(clapgo_lib, name);
    const char* error = dlerror();
    if (error != NULL) {
        fprintf(stderr, "Error: Failed to get symbol: %s (%s)\n", name, error);
        symbol = NULL;
    }
#endif

    return symbol;
}

// Helper function to load all required symbols for a manifest plugin entry
static bool clapgo_load_symbols_for_entry(manifest_plugin_entry_t* entry) {
    if (entry == NULL || entry->library == NULL) {
        fprintf(stderr, "Error: Cannot load symbols, entry or library is NULL\n");
        return false;
    }
    
    // Standard function names for accessing plugins
    const char* CREATE_PLUGIN_FUNC = "ClapGo_CreatePlugin";
    const char* GET_VERSION_FUNC = "ClapGo_GetVersion";
    
    // Standard function names for plugin metadata
    const char* GET_PLUGIN_ID_FUNC = "ClapGo_GetPluginID";
    const char* GET_PLUGIN_NAME_FUNC = "ClapGo_GetPluginName";
    const char* GET_PLUGIN_VENDOR_FUNC = "ClapGo_GetPluginVendor";
    const char* GET_PLUGIN_VERSION_FUNC = "ClapGo_GetPluginVersion";
    const char* GET_PLUGIN_DESCRIPTION_FUNC = "ClapGo_GetPluginDescription";
    
    // Standard function names for plugin lifecycle
    const char* PLUGIN_INIT_FUNC = "ClapGo_PluginInit";
    const char* PLUGIN_DESTROY_FUNC = "ClapGo_PluginDestroy";
    const char* PLUGIN_ACTIVATE_FUNC = "ClapGo_PluginActivate";
    const char* PLUGIN_DEACTIVATE_FUNC = "ClapGo_PluginDeactivate";
    const char* PLUGIN_START_PROCESSING_FUNC = "ClapGo_PluginStartProcessing";
    const char* PLUGIN_STOP_PROCESSING_FUNC = "ClapGo_PluginStopProcessing";
    const char* PLUGIN_RESET_FUNC = "ClapGo_PluginReset";
    const char* PLUGIN_PROCESS_FUNC = "ClapGo_PluginProcess";
    const char* PLUGIN_GET_EXTENSION_FUNC = "ClapGo_PluginGetExtension";
    const char* PLUGIN_ON_MAIN_THREAD_FUNC = "ClapGo_PluginOnMainThread";
    
    // Load function pointers directly into the manifest entry
#if defined(CLAPGO_OS_WINDOWS)
    entry->create_plugin = (clapgo_create_plugin_func)GetProcAddress(entry->library, CREATE_PLUGIN_FUNC);
    entry->get_version = (clapgo_get_version_func)GetProcAddress(entry->library, GET_VERSION_FUNC);
    
    entry->get_plugin_id = (clapgo_export_plugin_id_func)GetProcAddress(entry->library, GET_PLUGIN_ID_FUNC);
    entry->get_plugin_name = (clapgo_export_plugin_name_func)GetProcAddress(entry->library, GET_PLUGIN_NAME_FUNC);
    entry->get_plugin_vendor = (clapgo_export_plugin_vendor_func)GetProcAddress(entry->library, GET_PLUGIN_VENDOR_FUNC);
    entry->get_plugin_version = (clapgo_export_plugin_version_func)GetProcAddress(entry->library, GET_PLUGIN_VERSION_FUNC);
    entry->get_plugin_description = (clapgo_export_plugin_description_func)GetProcAddress(entry->library, GET_PLUGIN_DESCRIPTION_FUNC);
    
    entry->plugin_init = (clapgo_plugin_init_func)GetProcAddress(entry->library, PLUGIN_INIT_FUNC);
    entry->plugin_destroy = (clapgo_plugin_destroy_func)GetProcAddress(entry->library, PLUGIN_DESTROY_FUNC);
    entry->plugin_activate = (clapgo_plugin_activate_func)GetProcAddress(entry->library, PLUGIN_ACTIVATE_FUNC);
    entry->plugin_deactivate = (clapgo_plugin_deactivate_func)GetProcAddress(entry->library, PLUGIN_DEACTIVATE_FUNC);
    entry->plugin_start_processing = (clapgo_plugin_start_processing_func)GetProcAddress(entry->library, PLUGIN_START_PROCESSING_FUNC);
    entry->plugin_stop_processing = (clapgo_plugin_stop_processing_func)GetProcAddress(entry->library, PLUGIN_STOP_PROCESSING_FUNC);
    entry->plugin_reset = (clapgo_plugin_reset_func)GetProcAddress(entry->library, PLUGIN_RESET_FUNC);
    entry->plugin_process = (clapgo_plugin_process_func)GetProcAddress(entry->library, PLUGIN_PROCESS_FUNC);
    entry->plugin_get_extension = (clapgo_plugin_get_extension_func)GetProcAddress(entry->library, PLUGIN_GET_EXTENSION_FUNC);
    entry->plugin_on_main_thread = (clapgo_plugin_on_main_thread_func)GetProcAddress(entry->library, PLUGIN_ON_MAIN_THREAD_FUNC);
#elif defined(CLAPGO_OS_MACOS) || defined(CLAPGO_OS_LINUX)
    entry->create_plugin = (clapgo_create_plugin_func)dlsym(entry->library, CREATE_PLUGIN_FUNC);
    entry->get_version = (clapgo_get_version_func)dlsym(entry->library, GET_VERSION_FUNC);
    
    entry->get_plugin_id = (clapgo_export_plugin_id_func)dlsym(entry->library, GET_PLUGIN_ID_FUNC);
    entry->get_plugin_name = (clapgo_export_plugin_name_func)dlsym(entry->library, GET_PLUGIN_NAME_FUNC);
    entry->get_plugin_vendor = (clapgo_export_plugin_vendor_func)dlsym(entry->library, GET_PLUGIN_VENDOR_FUNC);
    entry->get_plugin_version = (clapgo_export_plugin_version_func)dlsym(entry->library, GET_PLUGIN_VERSION_FUNC);
    entry->get_plugin_description = (clapgo_export_plugin_description_func)dlsym(entry->library, GET_PLUGIN_DESCRIPTION_FUNC);
    
    entry->plugin_init = (clapgo_plugin_init_func)dlsym(entry->library, PLUGIN_INIT_FUNC);
    entry->plugin_destroy = (clapgo_plugin_destroy_func)dlsym(entry->library, PLUGIN_DESTROY_FUNC);
    entry->plugin_activate = (clapgo_plugin_activate_func)dlsym(entry->library, PLUGIN_ACTIVATE_FUNC);
    entry->plugin_deactivate = (clapgo_plugin_deactivate_func)dlsym(entry->library, PLUGIN_DEACTIVATE_FUNC);
    entry->plugin_start_processing = (clapgo_plugin_start_processing_func)dlsym(entry->library, PLUGIN_START_PROCESSING_FUNC);
    entry->plugin_stop_processing = (clapgo_plugin_stop_processing_func)dlsym(entry->library, PLUGIN_STOP_PROCESSING_FUNC);
    entry->plugin_reset = (clapgo_plugin_reset_func)dlsym(entry->library, PLUGIN_RESET_FUNC);
    entry->plugin_process = (clapgo_plugin_process_func)dlsym(entry->library, PLUGIN_PROCESS_FUNC);
    entry->plugin_get_extension = (clapgo_plugin_get_extension_func)dlsym(entry->library, PLUGIN_GET_EXTENSION_FUNC);
    entry->plugin_on_main_thread = (clapgo_plugin_on_main_thread_func)dlsym(entry->library, PLUGIN_ON_MAIN_THREAD_FUNC);
#endif
    
    // Check if we have all required functions
    if (!entry->create_plugin) {
        fprintf(stderr, "Error: Required CreatePlugin function not found in library\n");
        return false;
    }
    
    // These functions are essential for plugin operation
    if (!entry->plugin_init || !entry->plugin_destroy || !entry->plugin_activate || 
        !entry->plugin_deactivate || !entry->plugin_start_processing || 
        !entry->plugin_stop_processing || !entry->plugin_reset || !entry->plugin_process) {
        fprintf(stderr, "Error: Required plugin lifecycle functions not found in library\n");
        return false;
    }
    
    return true;
}

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

// Helper function to find the shared library based on the manifest
static char* get_library_path_from_manifest(const plugin_manifest_t* manifest, const char* plugin_dir) {
    if (!manifest || !plugin_dir) {
        fprintf(stderr, "Error: Invalid manifest or plugin directory\n");
        return NULL;
    }
    
    const char* lib_name = manifest->build.go_shared_library;
    if (!lib_name || lib_name[0] == '\0') {
        fprintf(stderr, "Error: Manifest does not specify a shared library name\n");
        return NULL;
    }
    
    // Allocate buffer for the Go shared library path
    char* lib_path = NULL;
    size_t path_len = strlen(plugin_dir) + strlen(lib_name) + 2; // +2 for path separator and null terminator
    
    lib_path = malloc(path_len);
    if (!lib_path) {
        fprintf(stderr, "Error: Memory allocation failed\n");
        return NULL;
    }
    
    // Construct the full path to the library
#if defined(CLAPGO_OS_WINDOWS)
    snprintf(lib_path, path_len, "%s\\%s", plugin_dir, lib_name);
#else // macOS and Linux
    snprintf(lib_path, path_len, "%s/%s", plugin_dir, lib_name);
#endif

    // Check if the file exists
    if (!file_exists(lib_path)) {
        fprintf(stderr, "Warning: Shared library not found at %s\n", lib_path);
        
        // Try the home directory
        free(lib_path);
        
        char* home = getenv("HOME");
        if (home) {
            path_len = strlen(home) + strlen("/.clap/") + strlen(lib_name) + 1;
            lib_path = malloc(path_len);
            if (!lib_path) {
                fprintf(stderr, "Error: Memory allocation failed\n");
                return NULL;
            }
            
            snprintf(lib_path, path_len, "%s/.clap/%s", home, lib_name);
            
            if (!file_exists(lib_path)) {
                fprintf(stderr, "Error: Shared library not found at %s\n", lib_path);
                free(lib_path);
                return NULL;
            }
        } else {
            fprintf(stderr, "Error: HOME environment variable not set\n");
            return NULL;
        }
    }
    
    return lib_path;
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
    
    // Get the plugin directory
    char* plugin_path_copy = strdup(plugin_path);
    char* plugin_dir = dirname(plugin_path_copy);
    
    // Create the expected manifest file path - it should be in the same directory as the plugin
    char manifest_path[512];
    snprintf(manifest_path, sizeof(manifest_path), "%s/%s.json", plugin_dir, plugin_name);
    
    printf("Looking for manifest at: %s\n", manifest_path);
    
    // Check if the manifest file exists
    if (access(manifest_path, R_OK) == 0) {
        // Load the manifest
        if (manifest_load_from_file(manifest_path, &manifest_plugins[0].manifest)) {
            printf("Loaded manifest: %s\n", manifest_path);
            manifest_plugins[0].loaded = false;
            manifest_plugins[0].library = NULL;
            manifest_plugins[0].entry_point = NULL;
            manifest_plugins[0].descriptor = NULL;
            manifest_plugin_count = 1;
        } else {
            fprintf(stderr, "Error: Failed to load manifest from %s\n", manifest_path);
        }
    } else {
        fprintf(stderr, "Error: Manifest file %s does not exist\n", manifest_path);
    }
    
    free(plugin_path_copy);
    
    // Return the number of manifests found (should be 0 or 1)
    return manifest_plugin_count;
}

// Check if the library can be loaded directly from the manifest
bool clapgo_check_direct_loading_supported(const plugin_manifest_t* manifest) {
    // Check required fields
    if (!manifest) return false;
    
    // Verify we have the required fields
    if (manifest->build.go_shared_library[0] == '\0') {
        fprintf(stderr, "Warning: Manifest is missing goSharedLibrary field\n");
        return false;
    }
    
    // We don't need to check for entry_point anymore - we use standardized export functions
    
    return true;
}

// Load a manifest plugin by index
bool clapgo_load_manifest_plugin(int index) {
    if (index < 0 || index >= manifest_plugin_count) {
        fprintf(stderr, "Error: Invalid manifest index: %d\n", index);
        return false;
    }
    
    manifest_plugin_entry_t* entry = &manifest_plugins[index];
    
    // Check if already loaded
    if (entry->loaded && entry->library != NULL && entry->create_plugin != NULL) {
        return true;
    }
    
    // Check if direct loading is supported
    if (!clapgo_check_direct_loading_supported(&entry->manifest)) {
        fprintf(stderr, "Error: Manifest does not support direct loading\n");
        return false;
    }
    
    // Get the plugin directory (use the home directory as fallback)
    char plugin_dir[512];
    char* home = getenv("HOME");
    if (!home) {
        fprintf(stderr, "Error: HOME environment variable not set\n");
        return false;
    }
    snprintf(plugin_dir, sizeof(plugin_dir), "%s/.clap", home);
    
    // Get the library path from the manifest
    char* lib_path = get_library_path_from_manifest(&entry->manifest, plugin_dir);
    if (!lib_path) {
        fprintf(stderr, "Error: Failed to get library path from manifest\n");
        return false;
    }
    
    printf("Loading shared library from manifest: %s\n", lib_path);
    
    // Load the shared library
    entry->library = NULL;
    
#if defined(CLAPGO_OS_WINDOWS)
    entry->library = LoadLibraryA(lib_path);
    if (entry->library == NULL) {
        DWORD error = GetLastError();
        fprintf(stderr, "Error: Failed to load library: %s (error code: %lu)\n", lib_path, error);
        free(lib_path);
        return false;
    }
#elif defined(CLAPGO_OS_MACOS) || defined(CLAPGO_OS_LINUX)
    entry->library = dlopen(lib_path, RTLD_LAZY | RTLD_LOCAL);
    if (entry->library == NULL) {
        fprintf(stderr, "Error: Failed to load library: %s (%s)\n", lib_path, dlerror());
        free(lib_path);
        return false;
    }
#else
    #error "Unsupported platform"
#endif

    free(lib_path);
    
    // Load all the required symbols for the plugin
    if (!clapgo_load_symbols_for_entry(entry)) {
        fprintf(stderr, "Error: Failed to load required symbols for plugin\n");
#if defined(CLAPGO_OS_WINDOWS)
        FreeLibrary(entry->library);
#elif defined(CLAPGO_OS_MACOS) || defined(CLAPGO_OS_LINUX)
        dlclose(entry->library);
#endif
        entry->library = NULL;
        return false;
    }
    
    printf("Successfully loaded symbols for plugin: %s\n", entry->manifest.plugin.id);
    
    // Create the descriptor from the manifest
    entry->descriptor = manifest_to_descriptor(&entry->manifest);
    if (!entry->descriptor) {
        fprintf(stderr, "Error: Failed to create descriptor from manifest\n");
#if defined(CLAPGO_OS_WINDOWS)
        FreeLibrary(entry->library);
#elif defined(CLAPGO_OS_MACOS) || defined(CLAPGO_OS_LINUX)
        dlclose(entry->library);
#endif
        entry->library = NULL;
        entry->create_plugin = NULL;
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
    
    // Create the plugin instance using the entry's create_plugin function
    void* go_instance = entry->create_plugin(host, (char*)entry->manifest.plugin.id);
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
            
            // Unload the library
            #if defined(CLAPGO_OS_WINDOWS)
                FreeLibrary(manifest_plugins[i].library);
            #elif defined(CLAPGO_OS_MACOS) || defined(CLAPGO_OS_LINUX)
                dlclose(manifest_plugins[i].library);
            #endif
            
            manifest_plugins[i].library = NULL;
            manifest_plugins[i].entry_point = NULL;
            manifest_plugins[i].loaded = false;
            
            // Free manifest resources
            manifest_free(&manifest_plugins[i].manifest);
        }
    }
    
    manifest_plugin_count = 0;
    
    // We don't need to free plugin_descriptors and plugin_count anymore
    // as they are no longer used. The descriptor cleanup is handled in
    // the manifest_plugins cleanup above.
    
    // Unload the shared library
    clapgo_unload_library();
    
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
    if (data->go_instance && entry->plugin_destroy != NULL) {
        entry->plugin_destroy(data->go_instance);
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
    if (entry->plugin_activate != NULL) {
        return entry->plugin_activate(data->go_instance, sample_rate, min_frames, max_frames);
    }
    
    fprintf(stderr, "Error: Go plugin activate function not available\n");
    return false;
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