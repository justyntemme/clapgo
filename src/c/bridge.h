#ifndef CLAPGO_BRIDGE_H
#define CLAPGO_BRIDGE_H

#include <stdbool.h>
#include <stdint.h>
#include "../../include/clap/include/clap/clap.h"

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

#ifdef __cplusplus
extern "C" {
#endif

// Go plugin state structure - holds the Go instance handle
typedef struct go_plugin_data {
    void* go_instance;
    const clap_plugin_descriptor_t* descriptor;
} go_plugin_data_t;

// Library handle for the Go shared library
extern clapgo_library_t clapgo_lib;

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

// Library loading utilities
bool clapgo_load_library(const char* path);
void clapgo_unload_library(void);
clapgo_symbol_t clapgo_get_symbol(const char* name);

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
bool clapgo_audio_ports_get(const clap_plugin_t* plugin, uint32_t index, bool is_input, clap_audio_port_info_t* info);

// Function pointer types for Go exports
typedef uint32_t (*clapgo_get_plugin_count_func)(void);
typedef void* (*clapgo_create_plugin_func)(const clap_host_t* host, const char* plugin_id);
typedef bool (*clapgo_get_version_func)(uint32_t* major, uint32_t* minor, uint32_t* patch);

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

// External function pointers
extern clapgo_get_plugin_count_func go_get_plugin_count;
extern clapgo_create_plugin_func go_create_plugin;
extern clapgo_get_version_func go_get_version;

// Standardized plugin metadata export functions
extern clapgo_export_plugin_id_func go_export_plugin_id;
extern clapgo_export_plugin_name_func go_export_plugin_name;
extern clapgo_export_plugin_vendor_func go_export_plugin_vendor;
extern clapgo_export_plugin_version_func go_export_plugin_version;
extern clapgo_export_plugin_description_func go_export_plugin_description;
extern clapgo_get_registered_plugin_count_func go_get_registered_plugin_count;
extern clapgo_get_registered_plugin_id_by_index_func go_get_registered_plugin_id_by_index;

extern clapgo_plugin_init_func go_plugin_init;
extern clapgo_plugin_destroy_func go_plugin_destroy;
extern clapgo_plugin_activate_func go_plugin_activate;
extern clapgo_plugin_deactivate_func go_plugin_deactivate;
extern clapgo_plugin_start_processing_func go_plugin_start_processing;
extern clapgo_plugin_stop_processing_func go_plugin_stop_processing;
extern clapgo_plugin_reset_func go_plugin_reset;
extern clapgo_plugin_process_func go_plugin_process;
extern clapgo_plugin_get_extension_func go_plugin_get_extension;
extern clapgo_plugin_on_main_thread_func go_plugin_on_main_thread;

#ifdef __cplusplus
}
#endif

#endif // CLAPGO_BRIDGE_H