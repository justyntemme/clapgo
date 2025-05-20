#ifndef CLAPGO_MANIFEST_H
#define CLAPGO_MANIFEST_H

#include <stdbool.h>
#include <stdint.h>
#include "../../include/clap/include/clap/clap.h"

// Maximum number of features a plugin can have
#define MAX_FEATURES 32

// Maximum number of extensions a plugin can support
#define MAX_EXTENSIONS 16

// Maximum number of parameters a plugin can have
#define MAX_PARAMETERS 128

// Plugin extension information
typedef struct {
    char id[256];
    bool supported;
} plugin_extension_t;

// Plugin parameter information
typedef struct {
    uint32_t id;
    char name[256];
    double min_value;
    double max_value;
    double default_value;
    uint32_t flags;
} plugin_parameter_t;

// Plugin build information
typedef struct {
    char go_shared_library[256];
    char entry_point[256];
} plugin_build_t;

// Plugin manifest representing the JSON structure
typedef struct {
    char schema_version[32];
    
    // Plugin info
    struct {
        char id[256];
        char name[256];
        char vendor[256];
        char version[64];
        char description[1024];
        char url[512];
        char manual_url[512];
        char support_url[512];
        
        // Features
        const char* features[MAX_FEATURES + 1]; // +1 for NULL terminator
        int feature_count;
    } plugin;
    
    // Build info
    plugin_build_t build;
    
    // Extensions
    plugin_extension_t extensions[MAX_EXTENSIONS];
    int extension_count;
    
    // Parameters
    plugin_parameter_t parameters[MAX_PARAMETERS];
    int parameter_count;
} plugin_manifest_t;

// Initialize a plugin manifest with default values
void manifest_init(plugin_manifest_t* manifest);

// Load a plugin manifest from a JSON file
bool manifest_load_from_file(const char* path, plugin_manifest_t* manifest);

// Convert a manifest to a CLAP plugin descriptor
clap_plugin_descriptor_t* manifest_to_descriptor(const plugin_manifest_t* manifest);

// Free resources associated with a manifest
void manifest_free(plugin_manifest_t* manifest);

// Find manifest files in a directory
char** manifest_find_files(const char* directory, int* count);

// Free a list of manifest files
void manifest_free_file_list(char** files, int count);

#endif // CLAPGO_MANIFEST_H