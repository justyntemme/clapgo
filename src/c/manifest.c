#include "manifest.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <dirent.h>
#include <ctype.h>
#include <sys/stat.h>
#include <libgen.h>
#include <json-c/json.h>

// Initialize the plugin manifest with default values
void manifest_init(plugin_manifest_t* manifest) {
    memset(manifest, 0, sizeof(plugin_manifest_t));
    strcpy(manifest->schema_version, "1.0");
    
    // Set default values for required fields
    strcpy(manifest->plugin.url, "https://github.com/justyntemme/clapgo");
    strcpy(manifest->plugin.manual_url, "https://github.com/justyntemme/clapgo");
    strcpy(manifest->plugin.support_url, "https://github.com/justyntemme/clapgo/issues");
    
    // No default entry_point - we use standardized export functions
}

bool manifest_load_from_file(const char* path, plugin_manifest_t* manifest) {
    printf("Loading manifest from file: %s\n", path);
    
    // Initialize manifest with default values
    manifest_init(manifest);
    
    // Parse JSON file using json-c
    struct json_object* root = json_object_from_file(path);
    if (!root) {
        fprintf(stderr, "Error: Failed to parse manifest file: %s\n", path);
        printf("json-c error: %s\n", json_util_get_last_err());
        return false;
    }
    
    printf("Successfully parsed JSON file: %s\n", path);
    
    // Get schema version
    struct json_object* schema_version_obj;
    if (json_object_object_get_ex(root, "schemaVersion", &schema_version_obj)) {
        const char* schema_version = json_object_get_string(schema_version_obj);
        if (schema_version) {
            strncpy(manifest->schema_version, schema_version, sizeof(manifest->schema_version) - 1);
        }
    }
    
    // Parse plugin object
    struct json_object* plugin_obj;
    if (json_object_object_get_ex(root, "plugin", &plugin_obj)) {
        // Parse plugin ID
        struct json_object* id_obj;
        if (json_object_object_get_ex(plugin_obj, "id", &id_obj)) {
            const char* id = json_object_get_string(id_obj);
            if (id) {
                strncpy(manifest->plugin.id, id, sizeof(manifest->plugin.id) - 1);
            }
        }
        
        // Parse plugin name
        struct json_object* name_obj;
        if (json_object_object_get_ex(plugin_obj, "name", &name_obj)) {
            const char* name = json_object_get_string(name_obj);
            if (name) {
                strncpy(manifest->plugin.name, name, sizeof(manifest->plugin.name) - 1);
            }
        }
        
        // Parse plugin vendor
        struct json_object* vendor_obj;
        if (json_object_object_get_ex(plugin_obj, "vendor", &vendor_obj)) {
            const char* vendor = json_object_get_string(vendor_obj);
            if (vendor) {
                strncpy(manifest->plugin.vendor, vendor, sizeof(manifest->plugin.vendor) - 1);
            }
        }
        
        // Parse plugin version
        struct json_object* version_obj;
        if (json_object_object_get_ex(plugin_obj, "version", &version_obj)) {
            const char* version = json_object_get_string(version_obj);
            if (version) {
                strncpy(manifest->plugin.version, version, sizeof(manifest->plugin.version) - 1);
            }
        }
        
        // Parse plugin description
        struct json_object* description_obj;
        if (json_object_object_get_ex(plugin_obj, "description", &description_obj)) {
            const char* description = json_object_get_string(description_obj);
            if (description) {
                strncpy(manifest->plugin.description, description, sizeof(manifest->plugin.description) - 1);
            }
        }
        
        // Parse plugin URLs
        struct json_object* url_obj;
        if (json_object_object_get_ex(plugin_obj, "url", &url_obj)) {
            const char* url = json_object_get_string(url_obj);
            if (url) {
                strncpy(manifest->plugin.url, url, sizeof(manifest->plugin.url) - 1);
            }
        }
        
        struct json_object* manual_url_obj;
        if (json_object_object_get_ex(plugin_obj, "manualUrl", &manual_url_obj)) {
            const char* manual_url = json_object_get_string(manual_url_obj);
            if (manual_url) {
                strncpy(manifest->plugin.manual_url, manual_url, sizeof(manifest->plugin.manual_url) - 1);
            }
        }
        
        struct json_object* support_url_obj;
        if (json_object_object_get_ex(plugin_obj, "supportUrl", &support_url_obj)) {
            const char* support_url = json_object_get_string(support_url_obj);
            if (support_url) {
                strncpy(manifest->plugin.support_url, support_url, sizeof(manifest->plugin.support_url) - 1);
            }
        }
        
        // Parse features array
        struct json_object* features_obj;
        if (json_object_object_get_ex(plugin_obj, "features", &features_obj) && 
            json_object_is_type(features_obj, json_type_array)) {
            
            int feature_count = json_object_array_length(features_obj);
            if (feature_count > MAX_FEATURES) {
                feature_count = MAX_FEATURES;
            }
            
            manifest->plugin.feature_count = feature_count;
            
            for (int i = 0; i < feature_count; i++) {
                struct json_object* feature = json_object_array_get_idx(features_obj, i);
                if (feature && json_object_is_type(feature, json_type_string)) {
                    const char* feature_str = json_object_get_string(feature);
                    if (feature_str) {
                        manifest->plugin.features[i] = strdup(feature_str);
                    }
                }
            }
        }
    }
    
    // Parse build object
    struct json_object* build_obj;
    if (json_object_object_get_ex(root, "build", &build_obj)) {
        // Parse Go shared library
        struct json_object* lib_obj;
        if (json_object_object_get_ex(build_obj, "goSharedLibrary", &lib_obj)) {
            const char* lib = json_object_get_string(lib_obj);
            if (lib) {
                strncpy(manifest->build.go_shared_library, lib, sizeof(manifest->build.go_shared_library) - 1);
            }
        }
        
        // We don't need the entry_point field anymore - we use standardized export functions
    }
    
    // Parse extensions array
    struct json_object* extensions_obj;
    if (json_object_object_get_ex(root, "extensions", &extensions_obj) && 
        json_object_is_type(extensions_obj, json_type_array)) {
        
        int extension_count = json_object_array_length(extensions_obj);
        if (extension_count > MAX_EXTENSIONS) {
            extension_count = MAX_EXTENSIONS;
        }
        
        manifest->extension_count = extension_count;
        
        for (int i = 0; i < extension_count; i++) {
            struct json_object* ext = json_object_array_get_idx(extensions_obj, i);
            if (ext && json_object_is_type(ext, json_type_object)) {
                struct json_object* id_obj;
                if (json_object_object_get_ex(ext, "id", &id_obj)) {
                    const char* id = json_object_get_string(id_obj);
                    if (id) {
                        strncpy(manifest->extensions[i].id, id, sizeof(manifest->extensions[i].id) - 1);
                    }
                }
                
                struct json_object* supported_obj;
                if (json_object_object_get_ex(ext, "supported", &supported_obj)) {
                    manifest->extensions[i].supported = json_object_get_boolean(supported_obj);
                }
            }
        }
    }
    
    // Parse parameters array
    struct json_object* params_obj;
    if (json_object_object_get_ex(root, "parameters", &params_obj) && 
        json_object_is_type(params_obj, json_type_array)) {
        
        int param_count = json_object_array_length(params_obj);
        if (param_count > MAX_PARAMETERS) {
            param_count = MAX_PARAMETERS;
        }
        
        manifest->parameter_count = param_count;
        
        for (int i = 0; i < param_count; i++) {
            struct json_object* param = json_object_array_get_idx(params_obj, i);
            if (param && json_object_is_type(param, json_type_object)) {
                struct json_object* id_obj;
                if (json_object_object_get_ex(param, "id", &id_obj)) {
                    manifest->parameters[i].id = json_object_get_int(id_obj);
                }
                
                struct json_object* name_obj;
                if (json_object_object_get_ex(param, "name", &name_obj)) {
                    const char* name = json_object_get_string(name_obj);
                    if (name) {
                        strncpy(manifest->parameters[i].name, name, sizeof(manifest->parameters[i].name) - 1);
                    }
                }
                
                struct json_object* min_obj;
                if (json_object_object_get_ex(param, "minValue", &min_obj)) {
                    manifest->parameters[i].min_value = json_object_get_double(min_obj);
                }
                
                struct json_object* max_obj;
                if (json_object_object_get_ex(param, "maxValue", &max_obj)) {
                    manifest->parameters[i].max_value = json_object_get_double(max_obj);
                }
                
                struct json_object* default_obj;
                if (json_object_object_get_ex(param, "defaultValue", &default_obj)) {
                    manifest->parameters[i].default_value = json_object_get_double(default_obj);
                }
                
                struct json_object* flags_obj;
                if (json_object_object_get_ex(param, "flags", &flags_obj)) {
                    manifest->parameters[i].flags = json_object_get_int(flags_obj);
                }
            }
        }
    }
    
    // Free the JSON object
    json_object_put(root);
    
    // Validate required fields
    if (manifest->plugin.id[0] == '\0' || 
        manifest->plugin.name[0] == '\0' || 
        manifest->plugin.vendor[0] == '\0' || 
        manifest->plugin.version[0] == '\0' || 
        manifest->build.go_shared_library[0] == '\0') {
        fprintf(stderr, "Error: Missing required fields in manifest file: %s\n", path);
        return false;
    }
    
    return true;
}

clap_plugin_descriptor_t* manifest_to_descriptor(const plugin_manifest_t* manifest) {
    // Allocate descriptor
    clap_plugin_descriptor_t* desc = (clap_plugin_descriptor_t*)calloc(1, sizeof(clap_plugin_descriptor_t));
    if (!desc) return NULL;
    
    // Initialize clap_version
    desc->clap_version.major = 1;
    desc->clap_version.minor = 1;
    desc->clap_version.revision = 0;
    
    // Convert all strings to C strings
    desc->id = strdup(manifest->plugin.id);
    desc->name = strdup(manifest->plugin.name);
    desc->vendor = strdup(manifest->plugin.vendor);
    desc->url = strdup(manifest->plugin.url);
    desc->manual_url = strdup(manifest->plugin.manual_url);
    desc->support_url = strdup(manifest->plugin.support_url);
    desc->version = strdup(manifest->plugin.version);
    desc->description = strdup(manifest->plugin.description);
    
    // Handle features array if any
    if (manifest->plugin.feature_count > 0) {
        // Allocate memory for the feature array (plus 1 for NULL terminator)
        const char** features = (const char**)calloc(manifest->plugin.feature_count + 1, sizeof(const char*));
        if (features) {
            // Copy each feature string
            for (int i = 0; i < manifest->plugin.feature_count; i++) {
                features[i] = strdup(manifest->plugin.features[i]);
            }
            // Set the NULL terminator
            features[manifest->plugin.feature_count] = NULL;
            
            desc->features = features;
        }
    } else {
        // Default features if none were provided
        const char** features = (const char**)calloc(4, sizeof(const char*));
        if (features) {
            features[0] = strdup("audio-effect");
            features[1] = strdup("stereo");
            features[2] = strdup("mono");
            features[3] = NULL;
            desc->features = features;
        }
    }
    
    return desc;
}

void manifest_free(plugin_manifest_t* manifest) {
    // Free features
    for (int i = 0; i < manifest->plugin.feature_count; i++) {
        free((void*)manifest->plugin.features[i]);
        manifest->plugin.features[i] = NULL;
    }
    manifest->plugin.feature_count = 0;
}

// Find manifest files in a directory
char** manifest_find_files(const char* directory, int* count) {
    printf("Searching for manifest files in directory: %s\n", directory);
    
    // First check for specific file patterns
    char manifest_path[512];
    const char* plugin_basename = basename(strdup(directory));
    snprintf(manifest_path, sizeof(manifest_path), "%s/%s.json", directory, plugin_basename);
    printf("Checking for specific manifest file: %s\n", manifest_path);
    
    // Try to find file in the directory
    FILE* test = fopen(manifest_path, "r");
    if (test) {
        printf("Found specific manifest file: %s\n", manifest_path);
        fclose(test);
        
        // Allocate and return a single file
        char** files = (char**)calloc(2, sizeof(char*));
        if (files) {
            files[0] = strdup(manifest_path);
            files[1] = NULL;
            *count = 1;
            return files;
        }
    }
    
    // Also check in central manifest repository
    char central_path[512];
    const char* plugin_name = basename(strdup(plugin_basename));
    snprintf(central_path, sizeof(central_path), "%s/.clap/manifests/%s.json", getenv("HOME"), plugin_name);
    printf("Checking for manifest in central repository: %s\n", central_path);
    
    test = fopen(central_path, "r");
    if (test) {
        printf("Found manifest file in central repository: %s\n", central_path);
        fclose(test);
        
        // Allocate and return a single file
        char** files = (char**)calloc(2, sizeof(char*));
        if (files) {
            files[0] = strdup(central_path);
            files[1] = NULL;
            *count = 1;
            return files;
        }
    }
    
    // Open directory to search for all JSON files
    DIR* dir = opendir(directory);
    if (!dir) {
        printf("Failed to open directory: %s\n", directory);
        *count = 0;
        return NULL;
    }
    
    // Count JSON files
    int file_count = 0;
    struct dirent* entry;
    while ((entry = readdir(dir)) != NULL) {
        // Check if the file has a .json extension
        const char* name = entry->d_name;
        size_t len = strlen(name);
        if (len > 5 && strcmp(name + len - 5, ".json") == 0) {
            printf("Found JSON file: %s\n", name);
            file_count++;
        }
    }
    
    // Allocate array for file paths
    char** files = NULL;
    if (file_count > 0) {
        files = (char**)calloc(file_count + 1, sizeof(char*)); // +1 for NULL terminator
        if (!files) {
            closedir(dir);
            *count = 0;
            return NULL;
        }
        
        // Rewind directory
        rewinddir(dir);
        
        // Collect JSON files
        int index = 0;
        while ((entry = readdir(dir)) != NULL && index < file_count) {
            const char* name = entry->d_name;
            size_t len = strlen(name);
            if (len > 5 && strcmp(name + len - 5, ".json") == 0) {
                size_t path_len = strlen(directory) + len + 2; // +2 for / and \0
                files[index] = (char*)malloc(path_len);
                if (files[index]) {
                    snprintf(files[index], path_len, "%s/%s", directory, name);
                    printf("Added manifest file path: %s\n", files[index]);
                    index++;
                }
            }
        }
        
        // Set the actual count
        *count = index;
        
        // Set NULL terminator
        files[index] = NULL;
    } else {
        *count = 0;
    }
    
    closedir(dir);
    return files;
}

// Free a list of manifest files
void manifest_free_file_list(char** files, int count) {
    if (!files) return;
    
    for (int i = 0; i < count; i++) {
        free(files[i]);
    }
    
    free(files);
}