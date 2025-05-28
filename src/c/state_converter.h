#pragma once

#include <clap/all.h>
#include <dirent.h>

// Get the plugin state converter factory
const clap_plugin_state_converter_factory_t* state_converter_get_factory(void);