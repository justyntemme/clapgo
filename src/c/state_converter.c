#include "state_converter.h"
#include "bridge.h"
#include "manifest.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <json-c/json.h>

// Forward declarations
static uint32_t converter_factory_count(const clap_plugin_state_converter_factory_t* factory);
static const clap_plugin_state_converter_descriptor_t* converter_factory_get_descriptor(
    const clap_plugin_state_converter_factory_t* factory, uint32_t index);
static clap_plugin_state_converter_t* converter_factory_create(
    const clap_plugin_state_converter_factory_t* factory, const char* converter_id);

// Converter data structure
typedef struct {
    char src_plugin_id[256];
    char dst_plugin_id[256];
    // Add more fields as needed for state conversion logic
} converter_data_t;

// Storage for converter descriptors
#define MAX_STATE_CONVERTERS 16
static clap_plugin_state_converter_descriptor_t converter_descriptors[MAX_STATE_CONVERTERS];
static clap_universal_plugin_id_t src_plugin_ids[MAX_STATE_CONVERTERS];
static clap_universal_plugin_id_t dst_plugin_ids[MAX_STATE_CONVERTERS];
static char converter_ids[MAX_STATE_CONVERTERS][256];
static char converter_names[MAX_STATE_CONVERTERS][256];
static char converter_vendors[MAX_STATE_CONVERTERS][256];
static char converter_versions[MAX_STATE_CONVERTERS][32];
static char converter_descriptions[MAX_STATE_CONVERTERS][512];
static uint32_t converter_count = 0;
static bool converters_initialized = false;

// Converter implementation functions
static void converter_destroy(clap_plugin_state_converter_t* converter);
static bool converter_convert_state(clap_plugin_state_converter_t* converter,
                                  const clap_istream_t* src,
                                  const clap_ostream_t* dst,
                                  char* error_buffer,
                                  size_t error_buffer_size);
static bool converter_convert_normalized_value(clap_plugin_state_converter_t* converter,
                                             clap_id src_param_id,
                                             double src_normalized_value,
                                             clap_id* dst_param_id,
                                             double* dst_normalized_value);
static bool converter_convert_plain_value(clap_plugin_state_converter_t* converter,
                                        clap_id src_param_id,
                                        double src_plain_value,
                                        clap_id* dst_param_id,
                                        double* dst_plain_value);

// Initialize converters based on available plugins
static void initialize_converters() {
    if (converters_initialized) {
        return;
    }
    
    converter_count = 0;
    
    // Look for converter configuration files
    const char* home = getenv("HOME");
    if (!home) {
        converters_initialized = true;
        return;
    }
    
    // Check for state converter configurations in ~/.clap/converters/
    char converter_dir[512];
    snprintf(converter_dir, sizeof(converter_dir), "%s/.clap/converters", home);
    
    DIR* dir = opendir(converter_dir);
    if (!dir) {
        converters_initialized = true;
        return;
    }
    
    struct dirent* entry;
    while ((entry = readdir(dir)) != NULL && converter_count < MAX_STATE_CONVERTERS) {
        // Look for .json files
        size_t name_len = strlen(entry->d_name);
        if (name_len > 5 && strcmp(entry->d_name + name_len - 5, ".json") == 0) {
            char file_path[768];
            snprintf(file_path, sizeof(file_path), "%s/%s", converter_dir, entry->d_name);
            
            // Load converter configuration
            struct json_object* root = json_object_from_file(file_path);
            if (root) {
                struct json_object* obj;
                
                // Parse converter metadata
                if (json_object_object_get_ex(root, "id", &obj)) {
                    strncpy(converter_ids[converter_count], json_object_get_string(obj), 
                           sizeof(converter_ids[converter_count]) - 1);
                }
                
                if (json_object_object_get_ex(root, "name", &obj)) {
                    strncpy(converter_names[converter_count], json_object_get_string(obj),
                           sizeof(converter_names[converter_count]) - 1);
                }
                
                if (json_object_object_get_ex(root, "vendor", &obj)) {
                    strncpy(converter_vendors[converter_count], json_object_get_string(obj),
                           sizeof(converter_vendors[converter_count]) - 1);
                }
                
                if (json_object_object_get_ex(root, "version", &obj)) {
                    strncpy(converter_versions[converter_count], json_object_get_string(obj),
                           sizeof(converter_versions[converter_count]) - 1);
                }
                
                if (json_object_object_get_ex(root, "description", &obj)) {
                    strncpy(converter_descriptions[converter_count], json_object_get_string(obj),
                           sizeof(converter_descriptions[converter_count]) - 1);
                }
                
                // Parse source and destination plugin IDs
                if (json_object_object_get_ex(root, "src_plugin_id", &obj)) {
                    src_plugin_ids[converter_count].abi = "clap";
                    src_plugin_ids[converter_count].id = converter_ids[converter_count]; // Store in stable memory
                    strncpy(converter_ids[converter_count], json_object_get_string(obj),
                           sizeof(converter_ids[converter_count]) - 1);
                }
                
                if (json_object_object_get_ex(root, "dst_plugin_id", &obj)) {
                    dst_plugin_ids[converter_count].abi = "clap";
                    dst_plugin_ids[converter_count].id = converter_names[converter_count]; // Store in stable memory
                    strncpy(converter_names[converter_count], json_object_get_string(obj),
                           sizeof(converter_names[converter_count]) - 1);
                }
                
                // Build descriptor
                converter_descriptors[converter_count].clap_version = (clap_version_t)CLAP_VERSION_INIT;
                converter_descriptors[converter_count].src_plugin_id = src_plugin_ids[converter_count];
                converter_descriptors[converter_count].dst_plugin_id = dst_plugin_ids[converter_count];
                converter_descriptors[converter_count].id = converter_ids[converter_count];
                converter_descriptors[converter_count].name = converter_names[converter_count];
                converter_descriptors[converter_count].vendor = converter_vendors[converter_count];
                converter_descriptors[converter_count].version = converter_versions[converter_count];
                converter_descriptors[converter_count].description = converter_descriptions[converter_count];
                
                converter_count++;
                json_object_put(root);
            }
        }
    }
    
    closedir(dir);
    converters_initialized = true;
}

// Factory implementation
static uint32_t converter_factory_count(const clap_plugin_state_converter_factory_t* factory) {
    initialize_converters();
    return converter_count;
}

static const clap_plugin_state_converter_descriptor_t* converter_factory_get_descriptor(
    const clap_plugin_state_converter_factory_t* factory, uint32_t index) {
    
    initialize_converters();
    
    if (index >= converter_count) {
        return NULL;
    }
    
    return &converter_descriptors[index];
}

static clap_plugin_state_converter_t* converter_factory_create(
    const clap_plugin_state_converter_factory_t* factory, const char* converter_id) {
    
    if (!converter_id) {
        return NULL;
    }
    
    initialize_converters();
    
    // Find matching converter
    for (uint32_t i = 0; i < converter_count; i++) {
        if (strcmp(converter_descriptors[i].id, converter_id) == 0) {
            // Create converter instance
            clap_plugin_state_converter_t* converter = 
                calloc(1, sizeof(clap_plugin_state_converter_t));
            if (!converter) {
                return NULL;
            }
            
            converter_data_t* data = calloc(1, sizeof(converter_data_t));
            if (!data) {
                free(converter);
                return NULL;
            }
            
            // Initialize converter data
            strncpy(data->src_plugin_id, converter_descriptors[i].src_plugin_id.id,
                   sizeof(data->src_plugin_id) - 1);
            strncpy(data->dst_plugin_id, converter_descriptors[i].dst_plugin_id.id,
                   sizeof(data->dst_plugin_id) - 1);
            
            converter->desc = &converter_descriptors[i];
            converter->converter_data = data;
            converter->destroy = converter_destroy;
            converter->convert_state = converter_convert_state;
            converter->convert_normalized_value = converter_convert_normalized_value;
            converter->convert_plain_value = converter_convert_plain_value;
            
            return converter;
        }
    }
    
    return NULL;
}

// Converter implementation
static void converter_destroy(clap_plugin_state_converter_t* converter) {
    if (converter) {
        if (converter->converter_data) {
            free(converter->converter_data);
        }
        free(converter);
    }
}

static bool converter_convert_state(clap_plugin_state_converter_t* converter,
                                  const clap_istream_t* src,
                                  const clap_ostream_t* dst,
                                  char* error_buffer,
                                  size_t error_buffer_size) {
    
    if (!converter || !src || !dst) {
        if (error_buffer && error_buffer_size > 0) {
            snprintf(error_buffer, error_buffer_size, "Invalid parameters");
        }
        return false;
    }
    
    converter_data_t* data = (converter_data_t*)converter->converter_data;
    
    // TODO: Implement actual state conversion logic
    // For now, this is a placeholder that would need to:
    // 1. Read the source state from the input stream
    // 2. Parse and convert the state data
    // 3. Write the converted state to the output stream
    
    // Example: Just copy the state as-is (not a real conversion)
    uint8_t buffer[1024];
    int64_t bytes_read;
    
    while ((bytes_read = src->read(src, buffer, sizeof(buffer))) > 0) {
        if (dst->write(dst, buffer, bytes_read) != bytes_read) {
            if (error_buffer && error_buffer_size > 0) {
                snprintf(error_buffer, error_buffer_size, "Failed to write converted state");
            }
            return false;
        }
    }
    
    return true;
}

static bool converter_convert_normalized_value(clap_plugin_state_converter_t* converter,
                                             clap_id src_param_id,
                                             double src_normalized_value,
                                             clap_id* dst_param_id,
                                             double* dst_normalized_value) {
    
    if (!converter || !dst_param_id || !dst_normalized_value) {
        return false;
    }
    
    // TODO: Implement parameter mapping logic
    // For now, just pass through the values
    *dst_param_id = src_param_id;
    *dst_normalized_value = src_normalized_value;
    
    return true;
}

static bool converter_convert_plain_value(clap_plugin_state_converter_t* converter,
                                        clap_id src_param_id,
                                        double src_plain_value,
                                        clap_id* dst_param_id,
                                        double* dst_plain_value) {
    
    if (!converter || !dst_param_id || !dst_plain_value) {
        return false;
    }
    
    // TODO: Implement parameter mapping logic
    // For now, just pass through the values
    *dst_param_id = src_param_id;
    *dst_plain_value = src_plain_value;
    
    return true;
}

// Factory instance
static const clap_plugin_state_converter_factory_t state_converter_factory = {
    .count = converter_factory_count,
    .get_descriptor = converter_factory_get_descriptor,
    .create = converter_factory_create
};

// Public function to get the factory
const clap_plugin_state_converter_factory_t* state_converter_get_factory(void) {
    return &state_converter_factory;
}