package main

/*
#include <stdint.h>
#include <stdbool.h>
#include <stdlib.h>
#include "../../include/clap/include/clap/clap.h"

// Forward declarations of the functions exported by Go
uint32_t GetPluginCount();
struct clap_plugin_descriptor *GetPluginInfo(uint32_t index);
void* CreatePlugin(struct clap_host *host, char *plugin_id);
bool GetVersion(uint32_t *major, uint32_t *minor, uint32_t *patch);

// Plugin lifecycle functions
bool GoInit(void *plugin);
void GoDestroy(void *plugin);
bool GoActivate(void *plugin, double sample_rate, uint32_t min_frames, uint32_t max_frames);
void GoDeactivate(void *plugin);
bool GoStartProcessing(void *plugin);
void GoStopProcessing(void *plugin);
void GoReset(void *plugin);
int32_t GoProcess(void *plugin, struct clap_process *process);
void *GoGetExtension(void *plugin, char *id);
void GoOnMainThread(void *plugin);
*/
import "C"

import (
	"github.com/justyntemme/clapgo/pkg/bridge"
)

func main() {
	// When built as a shared library, this won't be called directly.
	// Instead, the host will load the library and call the exported functions.
	// We'll call the bridge's Main function to perform any necessary initialization.
	bridge.Main()
}