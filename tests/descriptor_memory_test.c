#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "../include/clap/include/clap/clap.h"

// Forward declarations of functions we're testing
bool clapgo_init(const char* plugin_path);
void clapgo_deinit(void);

// External function declarations that we'll override
extern uint32_t clapgo_get_plugin_count(void);
extern const clap_plugin_descriptor_t* clapgo_get_plugin_descriptor(uint32_t index);

// Create a test plugin descriptor with dynamically allocated strings
static clap_plugin_descriptor_t* create_test_descriptor(const char* id, const char* name) {
    clap_plugin_descriptor_t* desc = calloc(1, sizeof(clap_plugin_descriptor_t));
    if (!desc) return NULL;
    
    // Initialize version
    desc->clap_version.major = 1;
    desc->clap_version.minor = 0;
    desc->clap_version.revision = 0;
    
    // Allocate and set string fields
    desc->id = strdup(id);
    desc->name = strdup(name);
    desc->vendor = strdup("Test Vendor");
    desc->url = strdup("https://example.com");
    desc->manual_url = strdup("https://example.com/manual");
    desc->support_url = strdup("https://example.com/support");
    desc->version = strdup("1.0.0");
    desc->description = strdup("Test plugin description");
    
    // Create features array
    const char** features = calloc(3, sizeof(char*));
    features[0] = strdup("feature1");
    features[1] = strdup("feature2");
    features[2] = NULL;
    desc->features = features;
    
    return desc;
}

// Our test data
static clap_plugin_descriptor_t* test_desc1 = NULL;
static clap_plugin_descriptor_t* test_desc2 = NULL;

// Initialize our test data
static void init_test_data() {
    if (!test_desc1) {
        test_desc1 = create_test_descriptor("com.example.test1", "Test Plugin 1");
    }
    
    if (!test_desc2) {
        test_desc2 = create_test_descriptor("com.example.test2", "Test Plugin 2");
    }
}

// Implementation for clapgo_get_plugin_count to be used in testing
uint32_t clapgo_get_plugin_count(void) {
    return 2; // Return 2 test plugins
}

// Implementation for clapgo_get_plugin_descriptor to be used in testing
const clap_plugin_descriptor_t* clapgo_get_plugin_descriptor(uint32_t index) {
    init_test_data();
    
    if (index == 0) return test_desc1;
    if (index == 1) return test_desc2;
    return NULL;
}

int main() {
    printf("Testing descriptor memory management...\n");
    
    // Test initialization
    if (!clapgo_init("/path/to/test")) {
        printf("Failed to initialize\n");
        return 1;
    }
    
    printf("Successfully initialized\n");
    
    // Test cleanup
    clapgo_deinit();
    
    printf("Successfully cleaned up\n");
    printf("Memory management test completed. Use a memory checker like Valgrind to verify no leaks.\n");
    
    return 0;
}