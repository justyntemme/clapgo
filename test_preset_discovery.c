#include <dlfcn.h>
#include <stdio.h>
#include <stdlib.h>
#include <clap/clap.h>

int main() {
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
    
    // Try to get preset discovery factory
    const void* factory = entry->get_factory(CLAP_PRESET_DISCOVERY_FACTORY_ID);
    if (factory) {
        printf("Got preset discovery factory!\n");
        
        const clap_preset_discovery_factory_t* pd_factory = 
            (const clap_preset_discovery_factory_t*)factory;
        
        uint32_t count = pd_factory->count(pd_factory);
        printf("Factory count: %u\n", count);
        
        for (uint32_t i = 0; i < count; i++) {
            const clap_preset_discovery_provider_descriptor_t* desc = 
                pd_factory->get_descriptor(pd_factory, i);
            if (desc) {
                printf("Provider %u: id=%s, name=%s\n", i, desc->id, desc->name);
            }
        }
    } else {
        printf("No preset discovery factory\n");
    }
    
    // Check if log files were created
    printf("\nChecking for log files...\n");
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