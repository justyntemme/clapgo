package main

/*
#include <stdint.h>
#include <stdbool.h>
#include "../../include/clap/include/clap/clap.h"

// Forward declarations of the functions we'll implement in Go
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
	"unsafe"

	"github.com/justyntemme/clapgo/src/goclap"
)

//export GetPluginCount
func GetPluginCount() C.uint32_t {
	return C.uint32_t(goclap.GetPluginCountImpl())
}

//export GetPluginInfo
func GetPluginInfo(index C.uint32_t) *C.struct_clap_plugin_descriptor {
	// Implementation is part of Plugin Implementation task, return nil for now
	return nil
}

//export CreatePlugin
func CreatePlugin(host *C.struct_clap_host, pluginID *C.char) unsafe.Pointer {
	// Implementation is part of Plugin Implementation task, return nil for now
	return nil
}

//export GetVersion
func GetVersion(major, minor, patch *C.uint32_t) C.bool {
	majorV, minorV, patchV := goclap.GetVersionImpl()
	if major != nil {
		*major = C.uint32_t(majorV)
	}
	if minor != nil {
		*minor = C.uint32_t(minorV)
	}
	if patch != nil {
		*patch = C.uint32_t(patchV)
	}
	return C.bool(true)
}

// We'll store the processor in a map using the plugin pointer as the key
var pluginProcessors = make(map[unsafe.Pointer]goclap.AudioProcessor)

//export GoInit
func GoInit(plugin unsafe.Pointer) C.bool {
	processor, exists := pluginProcessors[plugin]
	if !exists {
		return C.bool(false)
	}
	return C.bool(goclap.InitImpl(processor))
}

//export GoDestroy
func GoDestroy(plugin unsafe.Pointer) {
	processor, exists := pluginProcessors[plugin]
	if !exists {
		return
	}
	goclap.DestroyImpl(processor)
	delete(pluginProcessors, plugin)
}

//export GoActivate
func GoActivate(plugin unsafe.Pointer, sampleRate C.double, minFrames, maxFrames C.uint32_t) C.bool {
	processor, exists := pluginProcessors[plugin]
	if !exists {
		return C.bool(false)
	}
	return C.bool(goclap.ActivateImpl(processor, float64(sampleRate), uint32(minFrames), uint32(maxFrames)))
}

//export GoDeactivate
func GoDeactivate(plugin unsafe.Pointer) {
	processor, exists := pluginProcessors[plugin]
	if !exists {
		return
	}
	goclap.DeactivateImpl(processor)
}

//export GoStartProcessing
func GoStartProcessing(plugin unsafe.Pointer) C.bool {
	processor, exists := pluginProcessors[plugin]
	if !exists {
		return C.bool(false)
	}
	return C.bool(goclap.StartProcessingImpl(processor))
}

//export GoStopProcessing
func GoStopProcessing(plugin unsafe.Pointer) {
	processor, exists := pluginProcessors[plugin]
	if !exists {
		return
	}
	goclap.StopProcessingImpl(processor)
}

//export GoReset
func GoReset(plugin unsafe.Pointer) {
	processor, exists := pluginProcessors[plugin]
	if !exists {
		return
	}
	goclap.ResetImpl(processor)
}

//export GoProcess
func GoProcess(plugin unsafe.Pointer, process *C.struct_clap_process) C.int32_t {
	processor, exists := pluginProcessors[plugin]
	if !exists {
		return C.int32_t(0) // CLAP_PROCESS_ERROR
	}

	steadyTime := int64(process.steady_time)
	framesCount := uint32(process.frames_count)
	
	events := &goclap.ProcessEvents{
		InEvents:  unsafe.Pointer(process.in_events),
		OutEvents: unsafe.Pointer(process.out_events),
	}
	
	status := goclap.ProcessImpl(processor, steadyTime, framesCount, events)
	return C.int32_t(status)
}

//export GoGetExtension
func GoGetExtension(plugin unsafe.Pointer, id *C.char) unsafe.Pointer {
	processor, exists := pluginProcessors[plugin]
	if !exists {
		return nil
	}
	
	extID := C.GoString(id)
	return goclap.GetExtensionImpl(processor, extID)
}

//export GoOnMainThread
func GoOnMainThread(plugin unsafe.Pointer) {
	processor, exists := pluginProcessors[plugin]
	if !exists {
		return
	}
	goclap.OnMainThreadImpl(processor)
}

func main() {
	// This function is required for building a shared library
	// but is not used when the library is loaded
}

