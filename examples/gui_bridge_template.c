/**
 * GUI Bridge Template for ClapGo
 * 
 * This is a template for implementing a pure C bridge between ClapGo plugins
 * and a GUI framework of your choice. Replace the TODO sections with your
 * GUI framework implementation.
 */

#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
#include "../src/c/plugin.h"
#include "../include/clap/include/clap/ext/gui.h"

// Forward declarations of Go functions
extern bool GoGUICreated(void* plugin);
extern void GoGUIDestroyed(void* plugin);
extern bool GoGUIShown(void* plugin);
extern bool GoGUIHidden(void* plugin);
extern bool GoGUIGetSize(void* plugin, uint32_t* width, uint32_t* height);
extern bool GoGUIHasGUI(void* plugin);
extern bool GoGUIGetPreferredAPI(void* plugin, const char** api, bool* is_floating);
extern bool GoSetGUIExtensionPointer(void* plugin, void* ext_ptr);

// TODO: Define your GUI framework specific data types and structures here

// Forward declaration of GUI extension structure
static const clap_plugin_gui_t clapgo_gui_extension;

/**
 * Get the GUI extension from a plugin.
 * This function is called when the host requests the GUI extension.
 */
const void* clapgo_plugin_get_extension_with_gui(const clap_plugin_t* plugin, const char* id) {
    // First try the regular extension mechanism
    const void* ext = clapgo_plugin_get_extension(plugin, id);
    if (ext) {
        return ext;
    }
    
    // Check if this is a GUI extension request
    if (strcmp(id, CLAP_EXT_GUI) == 0) {
        // Check if the plugin supports GUI
        go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
        if (data && data->go_instance && GoGUIHasGUI(data->go_instance)) {
            // Register our GUI extension with the Go plugin
            GoSetGUIExtensionPointer(data->go_instance, (void*)&clapgo_gui_extension);
            
            // Return the GUI extension
            return &clapgo_gui_extension;
        }
    }
    
    return NULL;
}

/**
 * Check if the given API is supported.
 */
static bool clapgo_gui_is_api_supported(const clap_plugin_t* plugin, const char* api, bool is_floating) {
    // We support these standard APIs by default
    return strcmp(api, CLAP_WINDOW_API_X11) == 0 || 
           strcmp(api, CLAP_WINDOW_API_WAYLAND) == 0 || 
           strcmp(api, CLAP_WINDOW_API_WIN32) == 0 || 
           strcmp(api, CLAP_WINDOW_API_COCOA) == 0;
}

/**
 * Get the preferred API for this plugin.
 */
static bool clapgo_gui_get_preferred_api(const clap_plugin_t* plugin, const char** api, bool* is_floating) {
    if (!plugin || !api || !is_floating) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;

    return GoGUIGetPreferredAPI(data->go_instance, api, is_floating);
}

/**
 * Create the GUI.
 */
static bool clapgo_gui_create(const clap_plugin_t* plugin, const char* api, bool is_floating) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // TODO: Initialize your GUI framework here
    // 1. Create a window or attach to the parent window
    // 2. Set up your GUI widgets, etc.
    
    // Notify the Go side that GUI is created
    return GoGUICreated(data->go_instance);
}

/**
 * Destroy the GUI.
 */
static void clapgo_gui_destroy(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data) return;
    
    // TODO: Clean up your GUI framework resources here
    
    // Notify the Go side that GUI is destroyed
    if (data->go_instance) {
        GoGUIDestroyed(data->go_instance);
    }
}

/**
 * Set the scale factor for the GUI.
 */
static bool clapgo_gui_set_scale(const clap_plugin_t* plugin, double scale) {
    if (!plugin) return false;
    
    // TODO: Apply the scale factor to your GUI
    
    return true;
}

/**
 * Get the current size of the GUI.
 */
static bool clapgo_gui_get_size(const clap_plugin_t* plugin, uint32_t* width, uint32_t* height) {
    if (!plugin || !width || !height) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Get size from the Go side
    return GoGUIGetSize(data->go_instance, width, height);
}

/**
 * Check if the GUI can be resized.
 */
static bool clapgo_gui_can_resize(const clap_plugin_t* plugin) {
    // Allow resizing by default
    return true;
}

/**
 * Get resize hints for the GUI.
 */
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

/**
 * Adjust the GUI size to conform to the plugin's constraints.
 */
static bool clapgo_gui_adjust_size(const clap_plugin_t* plugin, 
                                  uint32_t* width, uint32_t* height) {
    if (!plugin || !width || !height) return false;
    
    // Ensure minimum size
    if (*width < 400) *width = 400;
    if (*height < 300) *height = 300;
    
    return true;
}

/**
 * Set the GUI size.
 */
static bool clapgo_gui_set_size(const clap_plugin_t* plugin, 
                               uint32_t width, uint32_t height) {
    if (!plugin) return false;
    
    // TODO: Resize your GUI to the specified dimensions
    
    return true;
}

/**
 * Set the parent window for the GUI.
 */
static bool clapgo_gui_set_parent(const clap_plugin_t* plugin, 
                                 const clap_window_t* window) {
    if (!plugin || !window) return false;
    
    // TODO: Attach your GUI to the parent window based on the API
    if (strcmp(window->api, CLAP_WINDOW_API_X11) == 0) {
        // Attach to X11 window
        // window->x11 contains the X11 window ID
    } else if (strcmp(window->api, CLAP_WINDOW_API_WIN32) == 0) {
        // Attach to Win32 window
        // window->win32 contains the Win32 window handle (HWND)
    } else if (strcmp(window->api, CLAP_WINDOW_API_COCOA) == 0) {
        // Attach to Cocoa window
        // window->cocoa contains the NSView* pointer
    } else if (strcmp(window->api, CLAP_WINDOW_API_WAYLAND) == 0) {
        // Attach to Wayland window
        // window->wayland contains the Wayland surface pointer
    }
    
    return false; // Replace with your implementation
}

/**
 * Set the transient window for the GUI.
 */
static bool clapgo_gui_set_transient(const clap_plugin_t* plugin, 
                                    const clap_window_t* window) {
    if (!plugin || !window) return false;
    
    // TODO: Set the transient window for your GUI based on the API
    
    return false; // Replace with your implementation
}

/**
 * Suggest a title for the GUI window.
 */
static void clapgo_gui_suggest_title(const clap_plugin_t* plugin, const char* title) {
    if (!plugin || !title) return;
    
    // TODO: Set the title of your GUI window
}

/**
 * Show the GUI.
 */
static bool clapgo_gui_show(const clap_plugin_t* plugin) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // TODO: Show your GUI
    
    // Notify the Go side that GUI is shown
    return GoGUIShown(data->go_instance);
}

/**
 * Hide the GUI.
 */
static bool clapgo_gui_hide(const clap_plugin_t* plugin) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // TODO: Hide your GUI
    
    // Notify the Go side that GUI is hidden
    return GoGUIHidden(data->go_instance);
}

/**
 * Define the GUI extension structure with all the function pointers.
 */
static const clap_plugin_gui_t clapgo_gui_extension = {
    .is_api_supported = clapgo_gui_is_api_supported,
    .get_preferred_api = clapgo_gui_get_preferred_api,
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