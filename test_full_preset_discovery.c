#include <dlfcn.h>
#include <stdio.h>
#include <stdlib.h>
#include <clap/clap.h>
#include <dirent.h>
#include <string.h>

// Mock indexer that just logs what it receives
static bool mock_declare_filetype(const clap_preset_discovery_indexer_t* indexer, 
                                   const clap_preset_discovery_filetype_t* filetype) {
    printf("MOCK: declare_filetype called: name=%s, extension=%s\n", 
           filetype->name, filetype->file_extension);
    return true;
}

static bool mock_declare_location(const clap_preset_discovery_indexer_t* indexer,
                                  const clap_preset_discovery_location_t* location) {
    printf("MOCK: declare_location called: name=%s, location=%s, flags=0x%x\n",
           location->name, location->location, location->flags);
    return true;
}

static clap_preset_discovery_indexer_t mock_indexer = {
    .declare_filetype = mock_declare_filetype,
    .declare_location = mock_declare_location,
    .declare_soundpack = NULL,
    .get_extension = NULL
};

// Mock receiver that logs what it receives
static bool mock_begin_preset(const clap_preset_discovery_metadata_receiver_t* receiver,
                              const char* name, const char* load_key) {
    printf("MOCK: begin_preset called: name=%s, load_key=%s\n",
           name ? name : "NULL", load_key ? load_key : "NULL");
    return true;
}

static void mock_add_plugin_id(const clap_preset_discovery_metadata_receiver_t* receiver,
                               const clap_universal_plugin_id_t* plugin_id) {
    printf("MOCK: add_plugin_id called: abi=%s, id=%s\n",
           plugin_id->abi, plugin_id->id);
}

static void mock_set_flags(const clap_preset_discovery_metadata_receiver_t* receiver,
                          uint32_t flags) {
    printf("MOCK: set_flags called: flags=0x%x\n", flags);
}

static void mock_set_description(const clap_preset_discovery_metadata_receiver_t* receiver,
                                 const char* description) {
    printf("MOCK: set_description called: %s\n", description ? description : "NULL");
}

static void mock_add_creator(const clap_preset_discovery_metadata_receiver_t* receiver,
                            const char* creator) {
    printf("MOCK: add_creator called: %s\n", creator ? creator : "NULL");
}

static void mock_add_feature(const clap_preset_discovery_metadata_receiver_t* receiver,
                            const char* feature) {
    printf("MOCK: add_feature called: %s\n", feature ? feature : "NULL");
}

static clap_preset_discovery_metadata_receiver_t mock_receiver = {
    .receiver_data = NULL,
    .begin_preset = mock_begin_preset,
    .add_plugin_id = mock_add_plugin_id,
    .set_flags = mock_set_flags,
    .set_description = mock_set_description,
    .add_creator = mock_add_creator,
    .add_feature = mock_add_feature,
    .add_extra_info = NULL
};

int main() {
    printf("=== Full Preset Discovery Test ===\n");
    
    // Load the plugin
    void* handle = dlopen("./examples/gain/build/gain.clap", RTLD_NOW);
    if (!handle) {
        printf("Failed to load plugin: %s\n", dlerror());
        return 1;
    }
    
    // Get the entry point
    const clap_plugin_entry_t* entry = dlsym(handle, "clap_entry");
    if (!entry) {
        printf("Failed to find clap_entry: %s\n", dlerror());
        dlclose(handle);
        return 1;
    }
    
    // Initialize
    if (!entry->init("./examples/gain/build/gain.clap")) {
        printf("Failed to initialize plugin\n");
        dlclose(handle);
        return 1;
    }
    
    printf("Plugin initialized successfully\n");
    
    // Get preset discovery factory
    const clap_preset_discovery_factory_t* factory = 
        (const clap_preset_discovery_factory_t*)entry->get_factory(CLAP_PRESET_DISCOVERY_FACTORY_ID);
    
    if (!factory) {
        printf("No preset discovery factory\n");
        entry->deinit();
        dlclose(handle);
        return 1;
    }
    
    printf("Got preset discovery factory!\n");
    
    // Get provider count and descriptors
    uint32_t count = factory->count(factory);
    printf("Factory count: %u\n", count);
    
    for (uint32_t i = 0; i < count; i++) {
        const clap_preset_discovery_provider_descriptor_t* desc = 
            factory->get_descriptor(factory, i);
        if (desc) {
            printf("Provider %u: id=%s, name=%s\n", i, desc->id, desc->name);
            
            // Create the provider
            printf("\n--- Creating provider %u ---\n", i);
            const clap_preset_discovery_provider_t* provider = 
                factory->create(factory, &mock_indexer, desc->id);
            
            if (!provider) {
                printf("Failed to create provider %u\n", i);
                continue;
            }
            
            printf("Provider created successfully!\n");
            
            // Initialize the provider
            printf("\n--- Initializing provider ---\n");
            bool init_result = provider->init(provider);
            printf("Provider init result: %s\n", init_result ? "SUCCESS" : "FAILURE");
            
            if (init_result) {
                // Test get_metadata on actual preset files
                printf("\n--- Testing get_metadata on preset files ---\n");
                
                // List preset files in the directory
                DIR* dir = opendir("/home/user/.clap/gain/presets");
                if (dir) {
                    struct dirent* entry_dir;
                    while ((entry_dir = readdir(dir)) != NULL) {
                        if (strstr(entry_dir->d_name, ".json")) {
                            char preset_path[512];
                            snprintf(preset_path, sizeof(preset_path), 
                                   "/home/user/.clap/gain/presets/%s", entry_dir->d_name);
                            
                            printf("\nTesting preset file: %s\n", preset_path);
                            bool meta_result = provider->get_metadata(provider, 
                                CLAP_PRESET_DISCOVERY_LOCATION_FILE, preset_path, &mock_receiver);
                            printf("get_metadata result: %s\n", meta_result ? "SUCCESS" : "FAILURE");
                        }
                    }
                    closedir(dir);
                } else {
                    printf("Could not open preset directory\n");
                }
            }
            
            // Destroy the provider
            provider->destroy(provider);
            printf("Provider destroyed\n");
        }
    }
    
    // Check if log files were created
    printf("\n=== Checking log files ===\n");
    FILE* f1 = fopen("/tmp/clapgo_factory_calls.log", "r");
    if (f1) {
        printf("Found /tmp/clapgo_factory_calls.log:\n");
        char line[256];
        while (fgets(line, sizeof(line), f1)) {
            printf("  %s", line);
        }
        fclose(f1);
    }
    
    char log_path[512];
    snprintf(log_path, sizeof(log_path), "%s/clapgo_preset_debug.log", getenv("HOME"));
    FILE* f2 = fopen(log_path, "r");
    if (f2) {
        printf("Found %s:\n", log_path);
        char line[256];
        while (fgets(line, sizeof(line), f2)) {
            printf("  %s", line);
        }
        fclose(f2);
    }
    
    // Cleanup
    entry->deinit();
    dlclose(handle);
    
    return 0;
}