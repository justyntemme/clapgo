#include <cstring>
#include <memory>
#include "../../clap-plugins/plugins/gui/local-gui-factory.hh"
#include "../../clap-plugins/plugins/gui/plugin-proxy.hh"
#include "../../clap-plugins/plugins/gui/parameter-proxy.hh"
#include "../../include/clap/include/clap/clap.h"

// Simple replacement of the old plugin.h structure
typedef struct {
    void* go_instance;
} go_plugin_data_t;

// Go GUI function declarations
extern "C" {
    bool GoGUICreated(void* plugin);
    void GoGUIDestroyed(void* plugin);
    bool GoGUIShown(void* plugin);
    bool GoGUIHidden(void* plugin);
    bool GoGUIGetSize(void* plugin, uint32_t* width, uint32_t* height);
    bool GoGUIHasGUI(void* plugin);
    bool GoGUIGetPreferredAPI(void* plugin, const char** api, bool* is_floating);
}

namespace {
    std::shared_ptr<clap::LocalGuiFactory> guiFactory;
    std::unordered_map<const clap_plugin_t*, std::unique_ptr<clap::GuiHandle>> guiHandles;
    std::unordered_map<const clap_plugin_t*, std::unique_ptr<clap::AbstractGuiListener>> guiListeners;
}

// Simple GUI listener implementation
class ClapGoGuiListener : public clap::AbstractGuiListener {
public:
    ClapGoGuiListener(const clap_plugin_t* plugin) 
        : _plugin(plugin), _data((go_plugin_data_t*)plugin->plugin_data) {}

    void onGuiClosed() override {
        printf("GUI closed\n");
    }

    void onParamAdjust(clap_id param_id, double value) override {
        printf("Parameter %u adjusted to %f\n", param_id, value);
        
        // Create a parameter adjustment event
        // In a real implementation, we would send this to the host
    }

    void onParamBeginAdjust(clap_id param_id) override {
        printf("Begin adjusting parameter %u\n", param_id);
        
        // Create a parameter begin event
    }

    void onParamEndAdjust(clap_id param_id) override {
        printf("End adjusting parameter %u\n", param_id);
        
        // Create a parameter end event
    }

    clap_id resolveParamIdForModuleId(clap_id module_id, clap_id param_id) override {
        // We don't use module IDs in our simple implementation
        return param_id;
    }

    void onDisplayStateChanged(bool is_visible) override {
        printf("Display state changed to %s\n", is_visible ? "visible" : "hidden");
    }

    void onPluginMissingResources() override {
        printf("Plugin missing resources\n");
    }

    void onPluginResumeFromSuspend() override {
        printf("Plugin resuming from suspend\n");
    }

private:
    const clap_plugin_t* _plugin;
    go_plugin_data_t* _data;
};

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
        // Check if the plugin supports GUI
        go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
        if (data && data->go_instance && GoGUIHasGUI(data->go_instance)) {
            return &clapgo_gui_extension;
        }
    }
    
    return NULL;
}

// GUI implementation

static bool clapgo_gui_is_api_supported(const clap_plugin_t* plugin, const char* api, bool is_floating) {
    // We support these APIs
    return strcmp(api, CLAP_WINDOW_API_X11) == 0 || 
           strcmp(api, CLAP_WINDOW_API_WAYLAND) == 0 || 
           strcmp(api, CLAP_WINDOW_API_WIN32) == 0 || 
           strcmp(api, CLAP_WINDOW_API_COCOA) == 0;
}

static bool clapgo_gui_get_preferred_api(const clap_plugin_t* plugin, const char** api, bool* is_floating) {
    if (!plugin || !api || !is_floating) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;

    return GoGUIGetPreferredAPI(data->go_instance, api, is_floating);
}

static bool clapgo_gui_create(const clap_plugin_t* plugin, const char* api, bool is_floating) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    printf("Creating GUI with API: %s (floating: %d)\n", api, is_floating);
    
    // Initialize the GUI factory if needed
    if (!guiFactory) {
        guiFactory = clap::LocalGuiFactory::getInstance();
    }
    
    // Create a GUI listener for this plugin
    auto listener = std::make_unique<ClapGoGuiListener>(plugin);
    guiListeners[plugin] = std::move(listener);
    
    // Create the GUI handle
    guiHandles[plugin] = guiFactory->createGui(*guiListeners[plugin]);
    
    // Notify the Go side that GUI is created
    return GoGUICreated(data->go_instance);
}

static void clapgo_gui_destroy(const clap_plugin_t* plugin) {
    if (!plugin) return;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data) return;
    
    // Destroy the GUI handle
    if (guiHandles.count(plugin) > 0) {
        guiFactory->releaseGui(*guiHandles[plugin]);
        guiHandles.erase(plugin);
    }
    
    // Remove the listener
    if (guiListeners.count(plugin) > 0) {
        guiListeners.erase(plugin);
    }
    
    // Notify the Go side that GUI is destroyed
    if (data->go_instance) {
        GoGUIDestroyed(data->go_instance);
    }
    
    printf("GUI destroyed\n");
}

static bool clapgo_gui_set_scale(const clap_plugin_t* plugin, double scale) {
    if (!plugin) return false;
    
    // Set GUI scale factor to the GUI handle
    if (guiHandles.count(plugin) > 0) {
        guiHandles[plugin]->setScale(scale);
        return true;
    }
    
    return false;
}

static bool clapgo_gui_get_size(const clap_plugin_t* plugin, uint32_t* width, uint32_t* height) {
    if (!plugin || !width || !height) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Get size from the Go side
    return GoGUIGetSize(data->go_instance, width, height);
}

static bool clapgo_gui_can_resize(const clap_plugin_t* plugin) {
    // Allow resizing
    return true;
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
    
    // Set size to the GUI handle
    if (guiHandles.count(plugin) > 0) {
        guiHandles[plugin]->setSize(width, height);
        return true;
    }
    
    return false;
}

static bool clapgo_gui_set_parent(const clap_plugin_t* plugin, 
                                 const clap_window_t* window) {
    if (!plugin || !window) return false;
    
    // Set parent window to the GUI handle
    if (guiHandles.count(plugin) > 0) {
        if (strcmp(window->api, CLAP_WINDOW_API_X11) == 0) {
            return guiHandles[plugin]->attachX11(window->x11);
        } else if (strcmp(window->api, CLAP_WINDOW_API_WIN32) == 0) {
            return guiHandles[plugin]->attachWin32(window->win32);
        } else if (strcmp(window->api, CLAP_WINDOW_API_COCOA) == 0) {
            return guiHandles[plugin]->attachCocoa(window->cocoa);
        }
    }
    
    return false;
}

static bool clapgo_gui_set_transient(const clap_plugin_t* plugin, 
                                    const clap_window_t* window) {
    if (!plugin || !window) return false;
    
    // Set transient window to the GUI handle
    if (guiHandles.count(plugin) > 0) {
        if (strcmp(window->api, CLAP_WINDOW_API_X11) == 0) {
            return guiHandles[plugin]->setTransientX11(window->x11);
        } else if (strcmp(window->api, CLAP_WINDOW_API_WIN32) == 0) {
            return guiHandles[plugin]->setTransientWin32(window->win32);
        } else if (strcmp(window->api, CLAP_WINDOW_API_COCOA) == 0) {
            return guiHandles[plugin]->setTransientCocoa(window->cocoa);
        }
    }
    
    return false;
}

static void clapgo_gui_suggest_title(const clap_plugin_t* plugin, const char* title) {
    if (!plugin || !title) return;
    
    // Nothing to do here
    printf("Suggested GUI title: %s\n", title);
}

static bool clapgo_gui_show(const clap_plugin_t* plugin) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Show the GUI handle
    if (guiHandles.count(plugin) > 0) {
        bool success = guiHandles[plugin]->show();
        if (success) {
            // Notify the Go side that GUI is shown
            return GoGUIShown(data->go_instance);
        }
    }
    
    return false;
}

static bool clapgo_gui_hide(const clap_plugin_t* plugin) {
    if (!plugin) return false;
    
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return false;
    
    // Hide the GUI handle
    if (guiHandles.count(plugin) > 0) {
        bool success = guiHandles[plugin]->hide();
        if (success) {
            // Notify the Go side that GUI is hidden
            return GoGUIHidden(data->go_instance);
        }
    }
    
    return false;
}

// Define the GUI extension structure
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

// Override the get_extension function to provide GUI extension
extern const void* clapgo_plugin_get_extension_with_gui(const clap_plugin_t* plugin, const char* id);

// Declare the original get_extension function
extern "C" const void* clapgo_plugin_get_extension(const clap_plugin_t* plugin, const char* id);

#ifdef __cplusplus
}
#endif