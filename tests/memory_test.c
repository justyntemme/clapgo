#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
#include "../include/clap/include/clap/clap.h"

// Simplified descriptor and allocation structure for testing
typedef struct {
    bool id;
    bool name;
    bool features_array;
    bool features_strings;
} test_alloc_info_t;

// Create a deep copy of a string for testing
static char* test_deep_copy_string(const char* str) {
    if (!str) return NULL;
    size_t len = strlen(str);
    char* copy = malloc(len + 1);
    if (copy) {
        strcpy(copy, str);
    }
    return copy;
}

// Free descriptor fields for testing
static void test_free_descriptor_fields(const clap_plugin_descriptor_t* descriptor, test_alloc_info_t* alloc_info) {
    if (!descriptor || !alloc_info) return;
    
    // Free string fields
    if (alloc_info->id && descriptor->id) free((void*)descriptor->id);
    if (alloc_info->name && descriptor->name) free((void*)descriptor->name);
    
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

int main() {
    printf("Testing descriptor memory management...\n");
    
    // Create a test descriptor
    clap_plugin_descriptor_t* desc = calloc(1, sizeof(clap_plugin_descriptor_t));
    if (!desc) {
        printf("Failed to allocate descriptor\n");
        return 1;
    }
    
    // Initialize tracking info
    test_alloc_info_t alloc_info = {0};
    
    // Set some fields
    desc->id = test_deep_copy_string("com.example.test");
    alloc_info.id = true;
    
    desc->name = test_deep_copy_string("Test Plugin");
    alloc_info.name = true;
    
    // Create features array
    const char** features = calloc(3, sizeof(char*));
    if (features) {
        features[0] = test_deep_copy_string("feature1");
        features[1] = test_deep_copy_string("feature2");
        features[2] = NULL;
        desc->features = features;
        
        alloc_info.features_array = true;
        alloc_info.features_strings = true;
    }
    
    printf("Descriptor created and initialized\n");
    
    // Clean up the descriptor
    test_free_descriptor_fields(desc, &alloc_info);
    free(desc);
    
    printf("Descriptor cleaned up successfully\n");
    printf("Memory management test completed successfully\n");
    
    return 0;
}