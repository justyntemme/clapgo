package main

// #cgo CFLAGS: -I../../include/clap/include
// #include "../../include/clap/include/clap/clap.h"
// #include <stdlib.h>
import "C"
import (
	"runtime/cgo"
	"unsafe"
	
	"github.com/justyntemme/clapgo/pkg/api"
	"github.com/justyntemme/clapgo/pkg/thread"
)

//export ClapGo_CreatePlugin
func ClapGo_CreatePlugin(host unsafe.Pointer, pluginID *C.char) unsafe.Pointer {
	if C.GoString(pluginID) == PluginID {
		return unsafe.Pointer(gainPlugin.CreateWithHost(host))
	}
	return nil
}

//export ClapGo_GetVersion
func ClapGo_GetVersion(major, minor, patch *C.uint32_t) C.bool {
	if major != nil { *major = C.uint32_t(1) }
	if minor != nil { *minor = C.uint32_t(0) }
	if patch != nil { *patch = C.uint32_t(0) }
	return C.bool(true)
}

//export ClapGo_GetPluginID
func ClapGo_GetPluginID(pluginID *C.char) *C.char {
	return C.CString(PluginID)
}

//export ClapGo_GetPluginName
func ClapGo_GetPluginName(pluginID *C.char) *C.char {
	return C.CString(PluginName)
}

//export ClapGo_GetPluginVendor
func ClapGo_GetPluginVendor(pluginID *C.char) *C.char {
	return C.CString(PluginVendor)
}

//export ClapGo_GetPluginVersion
func ClapGo_GetPluginVersion(pluginID *C.char) *C.char {
	return C.CString(PluginVersion)
}

//export ClapGo_GetPluginDescription
func ClapGo_GetPluginDescription(pluginID *C.char) *C.char {
	return C.CString(PluginDescription)
}

//export ClapGo_PluginInit
func ClapGo_PluginInit(plugin unsafe.Pointer) C.bool {
	return C.bool(getPlugin(plugin).Init())
}

//export ClapGo_PluginDestroy
func ClapGo_PluginDestroy(plugin unsafe.Pointer) {
	if p := getPlugin(plugin); p != nil {
		p.Destroy()
		// Unregister from audio ports provider
		api.UnregisterAudioPortsProvider(plugin)
		// Delete the handle to free the Go object
		cgo.Handle(plugin).Delete()
	}
}

//export ClapGo_PluginActivate
func ClapGo_PluginActivate(plugin unsafe.Pointer, sampleRate C.double, minFrames C.uint32_t, maxFrames C.uint32_t) C.bool {
	return C.bool(getPlugin(plugin).Activate(float64(sampleRate), uint32(minFrames), uint32(maxFrames)))
}

//export ClapGo_PluginDeactivate
func ClapGo_PluginDeactivate(plugin unsafe.Pointer) {
	getPlugin(plugin).Deactivate()
}

//export ClapGo_PluginStartProcessing
func ClapGo_PluginStartProcessing(plugin unsafe.Pointer) C.bool {
	thread.MarkAudioThread()
	defer thread.UnmarkAudioThread()
	return C.bool(getPlugin(plugin).StartProcessing())
}

//export ClapGo_PluginStopProcessing
func ClapGo_PluginStopProcessing(plugin unsafe.Pointer) {
	thread.MarkAudioThread()
	defer thread.UnmarkAudioThread()
	getPlugin(plugin).StopProcessing()
}

//export ClapGo_PluginReset
func ClapGo_PluginReset(plugin unsafe.Pointer) {
	getPlugin(plugin).Reset()
}

//export ClapGo_PluginProcess
func ClapGo_PluginProcess(plugin unsafe.Pointer, process unsafe.Pointer) C.int32_t {
	thread.MarkAudioThread()
	defer thread.UnmarkAudioThread()
	return C.int32_t(getPlugin(plugin).ProcessWithHandle(process))
}

//export ClapGo_PluginGetExtension
func ClapGo_PluginGetExtension(plugin unsafe.Pointer, id *C.char) unsafe.Pointer {
	return getPlugin(plugin).GetExtension(C.GoString(id))
}

//export ClapGo_PluginOnMainThread
func ClapGo_PluginOnMainThread(plugin unsafe.Pointer) {
	getPlugin(plugin).OnMainThread()
}

//export ClapGo_PluginParamsCount
func ClapGo_PluginParamsCount(plugin unsafe.Pointer) C.uint32_t {
	return C.uint32_t(getPlugin(plugin).ParamManager.Count())
}

//export ClapGo_PluginParamsGetInfo
func ClapGo_PluginParamsGetInfo(plugin unsafe.Pointer, index C.uint32_t, info unsafe.Pointer) C.bool {
	return C.bool(getPlugin(plugin).GetParamInfo(uint32(index), info))
}

//export ClapGo_PluginParamsGetValue
func ClapGo_PluginParamsGetValue(plugin unsafe.Pointer, paramID C.uint32_t, value *C.double) C.bool {
	return C.bool(getPlugin(plugin).GetParamValue(uint32(paramID), value))
}

//export ClapGo_PluginParamsValueToText
func ClapGo_PluginParamsValueToText(plugin unsafe.Pointer, paramID C.uint32_t, value C.double, buffer *C.char, size C.uint32_t) C.bool {
	return C.bool(getPlugin(plugin).ParamValueToText(uint32(paramID), float64(value), buffer, uint32(size)))
}

//export ClapGo_PluginParamsTextToValue
func ClapGo_PluginParamsTextToValue(plugin unsafe.Pointer, paramID C.uint32_t, text *C.char, value *C.double) C.bool {
	return C.bool(getPlugin(plugin).ParamTextToValue(uint32(paramID), C.GoString(text), value))
}

//export ClapGo_PluginParamsFlush
func ClapGo_PluginParamsFlush(plugin unsafe.Pointer, inEvents unsafe.Pointer, outEvents unsafe.Pointer) {
	getPlugin(plugin).ParamsFlush(inEvents, outEvents)
}

//export ClapGo_PluginStateSave
func ClapGo_PluginStateSave(plugin unsafe.Pointer, stream unsafe.Pointer) C.bool {
	return C.bool(getPlugin(plugin).SaveState(stream))
}

//export ClapGo_PluginStateLoad
func ClapGo_PluginStateLoad(plugin unsafe.Pointer, stream unsafe.Pointer) C.bool {
	return C.bool(getPlugin(plugin).LoadState(stream))
}

//export ClapGo_PluginStateSaveWithContext
func ClapGo_PluginStateSaveWithContext(plugin unsafe.Pointer, stream unsafe.Pointer, contextType C.uint32_t) C.bool {
	return C.bool(getPlugin(plugin).SaveStateWithContext(stream, uint32(contextType)))
}

//export ClapGo_PluginStateLoadWithContext
func ClapGo_PluginStateLoadWithContext(plugin unsafe.Pointer, stream unsafe.Pointer, contextType C.uint32_t) C.bool {
	return C.bool(getPlugin(plugin).LoadStateWithContext(stream, uint32(contextType)))
}

//export ClapGo_PluginPresetLoadFromLocation
func ClapGo_PluginPresetLoadFromLocation(plugin unsafe.Pointer, locationKind C.uint32_t, location *C.char, loadKey *C.char) C.bool {
	return C.bool(getPlugin(plugin).LoadPresetFromLocation(uint32(locationKind), C.GoString(location), C.GoString(loadKey)))
}

//export ClapGo_PluginLatencyGet
func ClapGo_PluginLatencyGet(plugin unsafe.Pointer) uint32 {
	return getPlugin(plugin).GetLatency()
}

//export ClapGo_PluginTailGet
func ClapGo_PluginTailGet(plugin unsafe.Pointer) C.uint32_t {
	return C.uint32_t(getPlugin(plugin).GetTail())
}

//export ClapGo_PluginOnTimer
func ClapGo_PluginOnTimer(plugin unsafe.Pointer, timerID C.uint64_t) {
	getPlugin(plugin).OnTimer(uint64(timerID))
}

//export ClapGo_PluginTrackInfoChanged
func ClapGo_PluginTrackInfoChanged(plugin unsafe.Pointer) {
	getPlugin(plugin).OnTrackInfoChanged()
}

//export ClapGo_PluginContextMenuPopulate
func ClapGo_PluginContextMenuPopulate(plugin unsafe.Pointer, targetKind C.uint32_t, targetID C.uint64_t, builder unsafe.Pointer) C.bool {
	// Context menu not supported
	return C.bool(false)
}

//export ClapGo_PluginContextMenuPerform
func ClapGo_PluginContextMenuPerform(plugin unsafe.Pointer, targetKind C.uint32_t, targetID C.uint64_t, actionID C.uint64_t) C.bool {
	// Context menu not supported
	return C.bool(false)
}

//export ClapGo_PluginRemoteControlsCount
func ClapGo_PluginRemoteControlsCount(plugin unsafe.Pointer) C.uint32_t {
	return C.uint32_t(getPlugin(plugin).GetRemoteControlsPageCount())
}

//export ClapGo_PluginRemoteControlsGet
func ClapGo_PluginRemoteControlsGet(plugin unsafe.Pointer, pageIndex C.uint32_t, cPage unsafe.Pointer) C.bool {
	return C.bool(getPlugin(plugin).GetRemoteControlsPageToC(uint32(pageIndex), cPage))
}

//export ClapGo_PluginParamIndicationSetMapping
func ClapGo_PluginParamIndicationSetMapping(plugin unsafe.Pointer, paramID C.uint64_t, hasMapping C.bool, color unsafe.Pointer, label *C.char, description *C.char) {
	getPlugin(plugin).OnParamMappingSet(uint32(paramID), bool(hasMapping), api.ColorFromC(color), C.GoString(label), C.GoString(description))
}

//export ClapGo_PluginParamIndicationSetAutomation
func ClapGo_PluginParamIndicationSetAutomation(plugin unsafe.Pointer, paramID C.uint64_t, automationState C.uint32_t, color unsafe.Pointer) {
	getPlugin(plugin).OnParamAutomationSet(uint32(paramID), uint32(automationState), api.ColorFromC(color))
}