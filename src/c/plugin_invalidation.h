#pragma once

#include <clap/all.h>
#include <sys/stat.h>

// Get the plugin invalidation factory
const clap_plugin_invalidation_factory_t* plugin_invalidation_get_factory(void);