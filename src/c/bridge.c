#include "bridge.h"
#include "manifest.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <libgen.h>  // For dirname()
#include <unistd.h>  // For access()

// Global library handle
clapgo_library_t clapgo_lib = NULL;

// Function pointers to Go exported functions
clapgo_get_plugin_count_func go_get_plugin_count = NULL;
clapgo_create_plugin_func go_create_plugin = NULL;
clapgo_get_version_func go_get_version = NULL;

// Standardized plugin metadata export functions
clapgo_export_plugin_id_func go_export_plugin_id = NULL;
clapgo_export_plugin_name_func go_export_plugin_name = NULL;
clapgo_export_plugin_vendor_func go_export_plugin_vendor = NULL;
clapgo_export_plugin_version_func go_export_plugin_version = NULL;
clapgo_export_plugin_description_func go_export_plugin_description = NULL;
clapgo_get_registered_plugin_count_func go_get_registered_plugin_count = NULL;
clapgo_get_registered_plugin_id_by_index_func go_get_registered_plugin_id_by_index = NULL;

clapgo_plugin_init_func go_plugin_init = NULL;
clapgo_plugin_destroy_func go_plugin_destroy = NULL;
clapgo_plugin_activate_func go_plugin_activate = NULL;
clapgo_plugin_deactivate_func go_plugin_deactivate = NULL;
clapgo_plugin_start_processing_func go_plugin_start_processing = NULL;
clapgo_plugin_stop_processing_func go_plugin_stop_processing = NULL;
clapgo_plugin_reset_func go_plugin_reset = NULL;
clapgo_plugin_process_func go_plugin_process = NULL;
clapgo_plugin_get_extension_func go_plugin_get_extension = NULL;
clapgo_plugin_on_main_thread_func go_plugin_on_main_thread = NULL;

// Internal storage for plugin descriptors
static const clap_plugin_descriptor_t **plugin_descriptors = NULL;
static uint32_t plugin_count = 0;

// Manifest plugin registry
manifest_plugin_entry_t manifest_plugins[MAX_PLUGIN_MANIFESTS];
int manifest_plugin_count = 0;

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
    
    // Reset all function pointers
    go_get_plugin_count = NULL;
    go_create_plugin = NULL;
    go_get_version = NULL;
    go_export_plugin_id = NULL;
    go_export_plugin_name = NULL;
    go_export_plugin_vendor = NULL;
    go_export_plugin_version = NULL;
    go_export_plugin_description = NULL;
    go_get_registered_plugin_count = NULL;
    go_get_registered_plugin_id_by_index = NULL;
    
    go_plugin_init = NULL;
    go_plugin_destroy = NULL;
    go_plugin_activate = NULL;
    go_plugin_deactivate = NULL;
    go_plugin_start_processing = NULL;
    go_plugin_stop_processing = NULL;
    go_plugin_reset = NULL;
    go_plugin_process = NULL;
    go_plugin_get_extension = NULL;
    go_plugin_on_main_thread = NULL;
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

// Helper function to load all required symbols from the Go shared library
static bool clapgo_load_symbols(void) {
    if (clapgo_lib == NULL) {
        fprintf(stderr, "Error: Cannot load symbols, library not loaded\n");
        return false;
    }
    
    // Load core plugin functions
    go_get_plugin_count = (clapgo_get_plugin_count_func)clapgo_get_symbol("GetPluginCount");
    go_create_plugin = (clapgo_create_plugin_func)clapgo_get_symbol("CreatePlugin");
    go_get_version = (clapgo_get_version_func)clapgo_get_symbol("GetVersion");
    
    // Load standardized plugin metadata functions
    go_export_plugin_id = (clapgo_export_plugin_id_func)clapgo_get_symbol("ExportPluginID");
    go_export_plugin_name = (clapgo_export_plugin_name_func)clapgo_get_symbol("ExportPluginName");
    go_export_plugin_vendor = (clapgo_export_plugin_vendor_func)clapgo_get_symbol("ExportPluginVendor");
    go_export_plugin_version = (clapgo_export_plugin_version_func)clapgo_get_symbol("ExportPluginVersion");
    go_export_plugin_description = (clapgo_export_plugin_description_func)clapgo_get_symbol("ExportPluginDescription");
    go_get_registered_plugin_count = (clapgo_get_registered_plugin_count_func)clapgo_get_symbol("GetRegisteredPluginCount");
    go_get_registered_plugin_id_by_index = (clapgo_get_registered_plugin_id_by_index_func)clapgo_get_symbol("GetRegisteredPluginIDByIndex");
    
    // Load plugin callback functions
    go_plugin_init = (clapgo_plugin_init_func)clapgo_get_symbol("GoInit");
    go_plugin_destroy = (clapgo_plugin_destroy_func)clapgo_get_symbol("GoDestroy");
    go_plugin_activate = (clapgo_plugin_activate_func)clapgo_get_symbol("GoActivate");
    go_plugin_deactivate = (clapgo_plugin_deactivate_func)clapgo_get_symbol("GoDeactivate");
    go_plugin_start_processing = (clapgo_plugin_start_processing_func)clapgo_get_symbol("GoStartProcessing");
    go_plugin_stop_processing = (clapgo_plugin_stop_processing_func)clapgo_get_symbol("GoStopProcessing");
    go_plugin_reset = (clapgo_plugin_reset_func)clapgo_get_symbol("GoReset");
    go_plugin_process = (clapgo_plugin_process_func)clapgo_get_symbol("GoProcess");
    go_plugin_get_extension = (clapgo_plugin_get_extension_func)clapgo_get_symbol("GoGetExtension");
    go_plugin_on_main_thread = (clapgo_plugin_on_main_thread_func)clapgo_get_symbol("GoOnMainThread");
    
    // Check if we have all required functions
    if (!go_get_plugin_count || !go_create_plugin) {
        fprintf(stderr, "Error: Required core functions not found in library\n");
        return false;
    }
    
    // Check for required plugin metadata functions
    if (!go_export_plugin_id || !go_export_plugin_name || !go_export_plugin_vendor || 
        !go_export_plugin_version || !go_export_plugin_description || 
        !go_get_registered_plugin_count || !go_get_registered_plugin_id_by_index) {
        fprintf(stderr, "Error: Required plugin metadata functions not found\n");
        return false;
    }
    
    // These functions are essential for plugin operation
    if (!go_plugin_init || !go_plugin_destroy || !go_plugin_activate || 
        !go_plugin_deactivate || !go_plugin_start_processing || 
        !go_plugin_stop_processing || !go_plugin_reset || !go_plugin_process) {
        fprintf(stderr, "Error: Required plugin callback symbols not found in library\n");
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
    if (entry->loaded && entry->library != NULL && entry->entry_point != NULL) {
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
    
    // Get the standard entry point symbols - we look for a fixed set of exported functions
    clapgo_symbol_t createPluginSymbol = NULL;
    
    // Standard entry point name for Go plugins
    const char* STANDARD_CREATE_PLUGIN_SYMBOL = "GetPluginCount";
    
#if defined(CLAPGO_OS_WINDOWS)
    createPluginSymbol = GetProcAddress(entry->library, STANDARD_CREATE_PLUGIN_SYMBOL);
    if (createPluginSymbol == NULL) {
        DWORD error = GetLastError();
        fprintf(stderr, "Error: Failed to get standard symbol: %s (error code: %lu)\n", 
                STANDARD_CREATE_PLUGIN_SYMBOL, error);
        FreeLibrary(entry->library);
        entry->library = NULL;
        return false;
    }
#elif defined(CLAPGO_OS_MACOS) || defined(CLAPGO_OS_LINUX)
    // Clear any existing errors
    dlerror();
    
    createPluginSymbol = dlsym(entry->library, STANDARD_CREATE_PLUGIN_SYMBOL);
    const char* error = dlerror();
    if (error != NULL) {
        fprintf(stderr, "Error: Failed to get standard symbol: %s (%s)\n", 
                STANDARD_CREATE_PLUGIN_SYMBOL, error);
        dlclose(entry->library);
        entry->library = NULL;
        return false;
    }
#endif
    
    // Store the standard create plugin function
    entry->entry_point = (clapgo_create_plugin_func)createPluginSymbol;
    
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
        entry->entry_point = NULL;
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
    
    // We need to load all the required Go functions for this plugin
    
    // Standard function names that all plugin libraries must export
    const char* CREATE_PLUGIN_FUNC = "CreatePlugin";
    const char* GO_INIT_FUNC = "GoInit";
    const char* GO_DESTROY_FUNC = "GoDestroy";
    const char* GO_ACTIVATE_FUNC = "GoActivate";
    const char* GO_DEACTIVATE_FUNC = "GoDeactivate";
    const char* GO_START_PROCESSING_FUNC = "GoStartProcessing";
    const char* GO_STOP_PROCESSING_FUNC = "GoStopProcessing";
    const char* GO_RESET_FUNC = "GoReset";
    const char* GO_PROCESS_FUNC = "GoProcess";
    const char* GO_GET_EXTENSION_FUNC = "GoGetExtension";
    const char* GO_ON_MAIN_THREAD_FUNC = "GoOnMainThread";
    
    // Load the create plugin function
    clapgo_create_plugin_func createFunc = NULL;
    
#if defined(CLAPGO_OS_WINDOWS)
    createFunc = (clapgo_create_plugin_func)GetProcAddress(entry->library, CREATE_PLUGIN_FUNC);
#elif defined(CLAPGO_OS_MACOS) || defined(CLAPGO_OS_LINUX)
    createFunc = (clapgo_create_plugin_func)dlsym(entry->library, CREATE_PLUGIN_FUNC);
#endif
    
    if (!createFunc) {
        fprintf(stderr, "Error: Failed to find %s function in the plugin library\n", CREATE_PLUGIN_FUNC);
        return NULL;
    }
    
    // Create the plugin instance using the standard function
    void* go_instance = createFunc(host, (char*)entry->manifest.plugin.id);
    if (!go_instance) {
        fprintf(stderr, "Error: Failed to create plugin instance\n");
        return NULL;
    }
    
    // Load all other required functions
    clapgo_plugin_init_func initFunc = NULL;
    clapgo_plugin_destroy_func destroyFunc = NULL;
    clapgo_plugin_activate_func activateFunc = NULL;
    clapgo_plugin_deactivate_func deactivateFunc = NULL;
    clapgo_plugin_start_processing_func startProcessingFunc = NULL;
    clapgo_plugin_stop_processing_func stopProcessingFunc = NULL;
    clapgo_plugin_reset_func resetFunc = NULL;
    clapgo_plugin_process_func processFunc = NULL;
    clapgo_plugin_get_extension_func getExtensionFunc = NULL;
    clapgo_plugin_on_main_thread_func onMainThreadFunc = NULL;
    
#if defined(CLAPGO_OS_WINDOWS)
    initFunc = (clapgo_plugin_init_func)GetProcAddress(entry->library, GO_INIT_FUNC);
    destroyFunc = (clapgo_plugin_destroy_func)GetProcAddress(entry->library, GO_DESTROY_FUNC);
    activateFunc = (clapgo_plugin_activate_func)GetProcAddress(entry->library, GO_ACTIVATE_FUNC);
    deactivateFunc = (clapgo_plugin_deactivate_func)GetProcAddress(entry->library, GO_DEACTIVATE_FUNC);
    startProcessingFunc = (clapgo_plugin_start_processing_func)GetProcAddress(entry->library, GO_START_PROCESSING_FUNC);
    stopProcessingFunc = (clapgo_plugin_stop_processing_func)GetProcAddress(entry->library, GO_STOP_PROCESSING_FUNC);
    resetFunc = (clapgo_plugin_reset_func)GetProcAddress(entry->library, GO_RESET_FUNC);
    processFunc = (clapgo_plugin_process_func)GetProcAddress(entry->library, GO_PROCESS_FUNC);
    getExtensionFunc = (clapgo_plugin_get_extension_func)GetProcAddress(entry->library, GO_GET_EXTENSION_FUNC);
    onMainThreadFunc = (clapgo_plugin_on_main_thread_func)GetProcAddress(entry->library, GO_ON_MAIN_THREAD_FUNC);
#elif defined(CLAPGO_OS_MACOS) || defined(CLAPGO_OS_LINUX)
    initFunc = (clapgo_plugin_init_func)dlsym(entry->library, GO_INIT_FUNC);
    destroyFunc = (clapgo_plugin_destroy_func)dlsym(entry->library, GO_DESTROY_FUNC);
    activateFunc = (clapgo_plugin_activate_func)dlsym(entry->library, GO_ACTIVATE_FUNC);
    deactivateFunc = (clapgo_plugin_deactivate_func)dlsym(entry->library, GO_DEACTIVATE_FUNC);
    startProcessingFunc = (clapgo_plugin_start_processing_func)dlsym(entry->library, GO_START_PROCESSING_FUNC);
    stopProcessingFunc = (clapgo_plugin_stop_processing_func)dlsym(entry->library, GO_STOP_PROCESSING_FUNC);
    resetFunc = (clapgo_plugin_reset_func)dlsym(entry->library, GO_RESET_FUNC);
    processFunc = (clapgo_plugin_process_func)dlsym(entry->library, GO_PROCESS_FUNC);
    getExtensionFunc = (clapgo_plugin_get_extension_func)dlsym(entry->library, GO_GET_EXTENSION_FUNC);
    onMainThreadFunc = (clapgo_plugin_on_main_thread_func)dlsym(entry->library, GO_ON_MAIN_THREAD_FUNC);
#endif
    
    // Check that we have all required functions
    if (!initFunc || !destroyFunc || !activateFunc || !deactivateFunc || 
        !startProcessingFunc || !stopProcessingFunc || !resetFunc || !processFunc) {
        fprintf(stderr, "Error: Failed to load all required plugin functions\n");
        return NULL;
    }
    
    // Allocate plugin instance data
    go_plugin_data_t* data = calloc(1, sizeof(go_plugin_data_t));
    if (!data) {
        fprintf(stderr, "Error: Failed to allocate plugin data\n");
        return NULL;
    }
    
    data->descriptor = entry->descriptor;
    data->go_instance = go_instance;
    data->manifest_index = index;
    
    // Store the function pointers in the bridge's global variables
    go_plugin_init = initFunc;
    go_plugin_destroy = destroyFunc;
    go_plugin_activate = activateFunc;
    go_plugin_deactivate = deactivateFunc;
    go_plugin_start_processing = startProcessingFunc;
    go_plugin_stop_processing = stopProcessingFunc;
    go_plugin_reset = resetFunc;
    go_plugin_process = processFunc;
    go_plugin_get_extension = getExtensionFunc;
    go_plugin_on_main_thread = onMainThreadFunc;
    
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
    
    // We'll create our plugin descriptors from the manifest
    plugin_count = manifest_count;
    
    // Allocate the descriptor array
    plugin_descriptors = calloc(plugin_count, sizeof(clap_plugin_descriptor_t*));
    if (!plugin_descriptors) {
        fprintf(stderr, "Error: Failed to allocate plugin descriptor array\n");
        return false;
    }
    
    // Initialize descriptor from the manifest
    manifest_plugins[0].descriptor = manifest_to_descriptor(&manifest_plugins[0].manifest);
    plugin_descriptors[0] = manifest_plugins[0].descriptor;
    
    if (plugin_descriptors[0]) {
        printf("Created descriptor from manifest: %s (%s)\n", 
               plugin_descriptors[0]->name, plugin_descriptors[0]->id);
    } else {
        fprintf(stderr, "Error: Failed to create descriptor from manifest\n");
        free(plugin_descriptors);
        plugin_descriptors = NULL;
        plugin_count = 0;
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
    // With the manifest approach, this is just the number of manifest plugins we loaded
    return manifest_plugin_count;
}

// Get the plugin descriptor at the given index
const clap_plugin_descriptor_t* clapgo_get_plugin_descriptor(uint32_t index) {
    if (index >= plugin_count) return NULL;
    return plugin_descriptors[index];
}

// Plugin callback implementations

// Initialize a plugin instance
bool clapgo_plugin_init(const clap_plugin_t* plugin) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Call into Go code to initialize the plugin instance
    if (go_plugin_init != NULL) {
        return go_plugin_init(data->go_instance);
    }
    
    fprintf(stderr, "Error: Go plugin init function not available\n");
    return false;
}

// Destroy a plugin instance
void clapgo_plugin_destroy(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data) return;
    
    // Call into Go code to clean up the plugin instance
    if (data->go_instance && go_plugin_destroy != NULL) {
        go_plugin_destroy(data->go_instance);
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
    if (go_plugin_activate != NULL) {
        return go_plugin_activate(data->go_instance, sample_rate, min_frames, max_frames);
    }
    
    fprintf(stderr, "Error: Go plugin activate function not available\n");
    return false;
}

// Deactivate a plugin instance
void clapgo_plugin_deactivate(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return;
    
    // Call into Go code to deactivate the plugin instance
    if (go_plugin_deactivate != NULL) {
        go_plugin_deactivate(data->go_instance);
    }
}

// Start processing
bool clapgo_plugin_start_processing(const clap_plugin_t* plugin) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Call into Go code to start processing
    if (go_plugin_start_processing != NULL) {
        return go_plugin_start_processing(data->go_instance);
    }
    
    fprintf(stderr, "Error: Go plugin start_processing function not available\n");
    return false;
}

// Stop processing
void clapgo_plugin_stop_processing(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return;
    
    // Call into Go code to stop processing
    if (go_plugin_stop_processing != NULL) {
        go_plugin_stop_processing(data->go_instance);
    }
}

// Reset a plugin instance
void clapgo_plugin_reset(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return;
    
    // Call into Go code to reset the plugin instance
    if (go_plugin_reset != NULL) {
        go_plugin_reset(data->go_instance);
    }
}

// Process audio
clap_process_status clapgo_plugin_process(const clap_plugin_t* plugin, 
                                         const clap_process_t* process) {
    if (!plugin || !process) return CLAP_PROCESS_ERROR;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return CLAP_PROCESS_ERROR;
    
    // Call into Go code to process audio
    if (go_plugin_process != NULL) {
        return go_plugin_process(data->go_instance, process);
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
    
    // Call into Go code to get the extension interface
    if (go_plugin_get_extension != NULL) {
        return go_plugin_get_extension(data->go_instance, (char*)id);
    }
    
    return NULL;
}

// Execute on main thread
void clapgo_plugin_on_main_thread(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return;
    
    // Call into Go code to execute on the main thread
    if (go_plugin_on_main_thread != NULL) {
        go_plugin_on_main_thread(data->go_instance);
    }
}