#ifndef CLAPGO_BRIDGE_H
#define CLAPGO_BRIDGE_H

#include <stdbool.h>
#include <stdint.h>
#include "../../include/clap/include/clap/clap.h"
#include "manifest.h"

// Platform detection
#if defined(_WIN32) || defined(_WIN64)
    #define CLAPGO_OS_WINDOWS 1
    #include <windows.h>
    typedef HMODULE clapgo_library_t;
    typedef FARPROC clapgo_symbol_t;
#elif defined(__APPLE__)
    #define CLAPGO_OS_MACOS 1
    #include <dlfcn.h>
    typedef void* clapgo_library_t;
    typedef void* clapgo_symbol_t;
#else // Linux and others
    #define CLAPGO_OS_LINUX 1
    #include <dlfcn.h>
    typedef void* clapgo_library_t;
    typedef void* clapgo_symbol_t;
#endif

// Plugin ID handling
#ifndef CLAPGO_PLUGIN_ID
#define CLAPGO_PLUGIN_ID "com.clapgo.plugin"
#endif

// Version constants for compatibility checks
#define CLAPGO_API_VERSION_MAJOR 0
#define CLAPGO_API_VERSION_MINOR 2
#define CLAPGO_API_VERSION_PATCH 0

// Maximum number of manifests that can be tracked
#define MAX_PLUGIN_MANIFESTS 32

#ifdef __cplusplus
extern "C" {
#endif

// Go plugin state structure - holds the Go instance handle
typedef struct go_plugin_data {
    void* go_instance;
    const clap_plugin_descriptor_t* descriptor;
    // For manifest-loaded plugins, store the manifest index
    int manifest_index;
    
    // Extension support flags - determined at plugin creation
    bool supports_params;      // Has param-related exports
    bool supports_note_ports;  // Has note port exports
    bool supports_state;       // Has state save/load exports
    bool supports_latency;     // Has latency export
    bool supports_tail;        // Has tail export
    bool supports_timer;       // Has timer export
    bool supports_audio_ports_config; // Has audio ports config exports
    bool supports_surround;    // Has surround exports
    bool supports_voice_info;  // Has voice info export
    bool supports_state_context; // Has state context exports
    bool supports_preset_load;  // Has preset load export
    bool supports_track_info;   // Has track info export
    bool supports_param_indication; // Has param indication exports
    bool supports_context_menu; // Has context menu exports
    bool supports_remote_controls; // Has remote controls exports
    bool supports_note_name; // Has note name exports
    bool supports_ambisonic; // Has ambisonic exports
    bool supports_audio_ports_activation; // Has audio ports activation exports
} go_plugin_data_t;


// Forward declaration
struct manifest_plugin_entry_t;


// Manifest plugin registry - external declarations
extern int manifest_plugin_count;

// Initialize the bridge - loads the Go library and initializes the plugin
bool clapgo_init(const char* plugin_path);

// Clean up the bridge - unloads the Go library and frees resources
void clapgo_deinit(void);

// Create a plugin instance for the given ID
const clap_plugin_t* clapgo_create_plugin(const clap_host_t* host, const char* plugin_id);

// Get the number of available plugins
uint32_t clapgo_get_plugin_count(void);

// Get the plugin descriptor at the given index
const clap_plugin_descriptor_t* clapgo_get_plugin_descriptor(uint32_t index);

// Get the plugin factory (CLAP interface)
const clap_plugin_factory_t* clapgo_get_plugin_factory(void);


// Find manifest files for the plugin
int clapgo_find_manifests(const char* plugin_path);

// Load a manifest plugin by index
bool clapgo_load_manifest_plugin(int index);

// Find a manifest plugin by ID
int clapgo_find_manifest_plugin_by_id(const char* plugin_id);

// Create a plugin instance from a manifest entry
const clap_plugin_t* clapgo_create_plugin_from_manifest(const clap_host_t* host, int index);

// Check if the library can be loaded directly from the manifest
bool clapgo_check_direct_loading_supported(const plugin_manifest_t* manifest);

// Plugin callback implementations
bool clapgo_plugin_init(const clap_plugin_t* plugin);
void clapgo_plugin_destroy(const clap_plugin_t* plugin);
bool clapgo_plugin_activate(const clap_plugin_t* plugin, double sample_rate, uint32_t min_frames, uint32_t max_frames);
void clapgo_plugin_deactivate(const clap_plugin_t* plugin);
bool clapgo_plugin_start_processing(const clap_plugin_t* plugin);
void clapgo_plugin_stop_processing(const clap_plugin_t* plugin);
void clapgo_plugin_reset(const clap_plugin_t* plugin);
clap_process_status clapgo_plugin_process(const clap_plugin_t* plugin, const clap_process_t* process);
const void* clapgo_plugin_get_extension(const clap_plugin_t* plugin, const char* id);
void clapgo_plugin_on_main_thread(const clap_plugin_t* plugin);

// Audio ports extension implementation
uint32_t clapgo_audio_ports_count(const clap_plugin_t* plugin, bool is_input);
bool clapgo_audio_ports_info(const clap_plugin_t* plugin, uint32_t index, bool is_input, clap_audio_port_info_t* info);

// Function pointer types for other Go exports

// Function pointer types for standardized plugin metadata exports
typedef char* (*clapgo_export_plugin_id_func)(const char* plugin_id);
typedef char* (*clapgo_export_plugin_name_func)(const char* plugin_id);
typedef char* (*clapgo_export_plugin_vendor_func)(const char* plugin_id);
typedef char* (*clapgo_export_plugin_version_func)(const char* plugin_id);
typedef char* (*clapgo_export_plugin_description_func)(const char* plugin_id);
typedef uint32_t (*clapgo_get_registered_plugin_count_func)(void);
typedef char* (*clapgo_get_registered_plugin_id_by_index_func)(uint32_t index);

typedef bool (*clapgo_plugin_init_func)(void* plugin);
typedef void (*clapgo_plugin_destroy_func)(void* plugin);
typedef bool (*clapgo_plugin_activate_func)(void* plugin, double sample_rate, uint32_t min_frames, uint32_t max_frames);
typedef void (*clapgo_plugin_deactivate_func)(void* plugin);
typedef bool (*clapgo_plugin_start_processing_func)(void* plugin);
typedef void (*clapgo_plugin_stop_processing_func)(void* plugin);
typedef void (*clapgo_plugin_reset_func)(void* plugin);
typedef clap_process_status (*clapgo_plugin_process_func)(void* plugin, const clap_process_t* process);
typedef const void* (*clapgo_plugin_get_extension_func)(void* plugin, const char* id);
typedef void (*clapgo_plugin_on_main_thread_func)(void* plugin);

// Function pointer types for plugin creation and versioning
typedef void* (*clapgo_create_plugin_func)(void* host, const char* plugin_id);
typedef bool (*clapgo_get_version_func)(uint32_t* major, uint32_t* minor, uint32_t* patch);


#ifdef __cplusplus
}
#endif

#endif // CLAPGO_BRIDGE_H