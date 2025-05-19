#include <cstring>
#include <memory>
#include "../../src/c/plugin.h"
#include "local-gui-factory.hh"
#include "plugin-proxy.hh"
#include "parameter-proxy.hh"

#ifdef __cplusplus
extern "C" {
#endif

// Forward declaration of GUI extension structure
static const clap_plugin_gui_t clapgo_gui_extension;

// GUI extension implementation
const void* clapgo_plugin_get_extension_with_gui(const clap_plugin_t* plugin, const char* id) {
    // First try the regular extension mechanism
    const void* ext = clapgo_plugin_get_extension(plugin, id);
    if (ext) {
        return ext;
    }
    
    // Check if this is a GUI extension request
    if (strcmp(id, CLAP_EXT_GUI) == 0) {
        return &clapgo_gui_extension;
    }
    
    return NULL;
}

// GUI implementation

static bool clapgo_gui_is_api_supported(const clap_plugin_t* plugin, const char* api, bool is_floating) {
    // We support only one GUI API for now
    return strcmp(api, CLAP_WINDOW_API_X11) == 0 || 
           strcmp(api, CLAP_WINDOW_API_WAYLAND) == 0 || 
           strcmp(api, CLAP_WINDOW_API_WIN32) == 0 || 
           strcmp(api, CLAP_WINDOW_API_COCOA) == 0;
}

static bool clapgo_gui_create(const clap_plugin_t* plugin, const char* api, bool is_floating) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data) return false;
    
    // Here we would create the GUI, but this is just a stub
    printf("Creating GUI with API: %s (floating: %d)\n", api, is_floating);
    
    // In a real implementation, this would initialize a Qt/QML GUI
    // using the clap-plugins GUI framework
    return true;
}

static void clapgo_gui_destroy(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data) return;
    
    // Here we would destroy the GUI, but this is just a stub
    printf("Destroying GUI\n");
}

static bool clapgo_gui_set_scale(const clap_plugin_t* plugin, double scale) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data) return false;
    
    // Set GUI scale factor
    printf("Setting GUI scale to: %f\n", scale);
    return true;
}

static bool clapgo_gui_get_size(const clap_plugin_t* plugin, uint32_t* width, uint32_t* height) {
    if (!plugin || !width || !height) return false;
    
    // Return default size
    *width = 800;
    *height = 600;
    return true;
}

static bool clapgo_gui_can_resize(const clap_plugin_t* plugin) {
    return true; // Allow resizing
}

static bool clapgo_gui_get_resize_hints(const clap_plugin_t* plugin, 
                                       clap_gui_resize_hints_t* hints) {
    if (!plugin || !hints) return false;
    
    // Set reasonable resize constraints
    hints->can_resize_horizontally = true;
    hints->can_resize_vertically = true;
    hints->preserve_aspect_ratio = false;
    hints->aspect_ratio_width = 1;
    hints->aspect_ratio_height = 1;
    
    return true;
}

static bool clapgo_gui_adjust_size(const clap_plugin_t* plugin, 
                                  uint32_t* width, uint32_t* height) {
    if (!plugin || !width || !height) return false;
    
    // Ensure minimum size
    if (*width < 400) *width = 400;
    if (*height < 300) *height = 300;
    
    return true;
}

static bool clapgo_gui_set_size(const clap_plugin_t* plugin, 
                               uint32_t width, uint32_t height) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data) return false;
    
    // Set GUI size
    printf("Setting GUI size to: %ux%u\n", width, height);
    return true;
}

static bool clapgo_gui_set_parent(const clap_plugin_t* plugin, 
                                 const clap_window_t* window) {
    if (!plugin || !window) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data) return false;
    
    // Set the parent window
    printf("Setting GUI parent window\n");
    return true;
}

static bool clapgo_gui_set_transient(const clap_plugin_t* plugin, 
                                    const clap_window_t* window) {
    if (!plugin || !window) return false;
    
    return true; // Not all platforms support this
}

static void clapgo_gui_suggest_title(const clap_plugin_t* plugin, const char* title) {
    if (!plugin || !title) return;
    
    printf("Suggested GUI title: %s\n", title);
}

static bool clapgo_gui_show(const clap_plugin_t* plugin) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data) return false;
    
    // Show the GUI
    printf("Showing GUI\n");
    return true;
}

static bool clapgo_gui_hide(const clap_plugin_t* plugin) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data) return false;
    
    // Hide the GUI
    printf("Hiding GUI\n");
    return true;
}

// Define the GUI extension structure
static const clap_plugin_gui_t clapgo_gui_extension = {
    .is_api_supported = clapgo_gui_is_api_supported,
    .create = clapgo_gui_create,
    .destroy = clapgo_gui_destroy,
    .set_scale = clapgo_gui_set_scale,
    .get_size = clapgo_gui_get_size,
    .can_resize = clapgo_gui_can_resize,
    .get_resize_hints = clapgo_gui_get_resize_hints,
    .adjust_size = clapgo_gui_adjust_size,
    .set_size = clapgo_gui_set_size,
    .set_parent = clapgo_gui_set_parent,
    .set_transient = clapgo_gui_set_transient,
    .suggest_title = clapgo_gui_suggest_title,
    .show = clapgo_gui_show,
    .hide = clapgo_gui_hide
};

// Override the get_extension function to provide GUI extension
extern const void* clapgo_plugin_get_extension_with_gui(const clap_plugin_t* plugin, const char* id);

#ifdef __cplusplus
}
#endif