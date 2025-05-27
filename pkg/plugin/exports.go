package plugin

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

// Plugin is the interface that all plugins must implement
type Plugin interface {
	Init() bool
	Destroy()
	Activate(sampleRate float64, minFrames, maxFrames uint32) bool
	Deactivate()
	StartProcessing() bool
	StopProcessing()
	Reset()
	GetExtension(id string) unsafe.Pointer
	OnMainThread()
}

// ExportHelpers provides common export function implementations
// that can be used by plugins to reduce boilerplate
type ExportHelpers struct {
	GetPlugin func(unsafe.Pointer) Plugin
}

// NewExportHelpers creates export helpers for a plugin
func NewExportHelpers(getPlugin func(unsafe.Pointer) Plugin) *ExportHelpers {
	return &ExportHelpers{
		GetPlugin: getPlugin,
	}
}

// PluginInit implements ClapGo_PluginInit
func (h *ExportHelpers) PluginInit(plugin unsafe.Pointer) C.bool {
	return C.bool(h.GetPlugin(plugin).Init())
}

// PluginDestroy implements ClapGo_PluginDestroy
func (h *ExportHelpers) PluginDestroy(plugin unsafe.Pointer) {
	if p := h.GetPlugin(plugin); p != nil {
		p.Destroy()
		// Unregister from any providers
		api.UnregisterAudioPortsProvider(plugin)
		api.UnregisterVoiceInfoProvider(plugin)
		// Delete the handle to free the Go object
		if plugin != nil {
			cgo.Handle(plugin).Delete()
		}
	}
}

// PluginActivate implements ClapGo_PluginActivate
func (h *ExportHelpers) PluginActivate(plugin unsafe.Pointer, sampleRate C.double, minFrames C.uint32_t, maxFrames C.uint32_t) C.bool {
	return C.bool(h.GetPlugin(plugin).Activate(float64(sampleRate), uint32(minFrames), uint32(maxFrames)))
}

// PluginDeactivate implements ClapGo_PluginDeactivate
func (h *ExportHelpers) PluginDeactivate(plugin unsafe.Pointer) {
	h.GetPlugin(plugin).Deactivate()
}

// PluginStartProcessing implements ClapGo_PluginStartProcessing
func (h *ExportHelpers) PluginStartProcessing(plugin unsafe.Pointer) C.bool {
	thread.MarkAudioThread()
	defer thread.UnmarkAudioThread()
	return C.bool(h.GetPlugin(plugin).StartProcessing())
}

// PluginStopProcessing implements ClapGo_PluginStopProcessing
func (h *ExportHelpers) PluginStopProcessing(plugin unsafe.Pointer) {
	thread.MarkAudioThread()
	defer thread.UnmarkAudioThread()
	h.GetPlugin(plugin).StopProcessing()
}

// PluginReset implements ClapGo_PluginReset
func (h *ExportHelpers) PluginReset(plugin unsafe.Pointer) {
	h.GetPlugin(plugin).Reset()
}

// PluginProcess implements ClapGo_PluginProcess
func (h *ExportHelpers) PluginProcess(plugin unsafe.Pointer, process unsafe.Pointer) C.int32_t {
	thread.MarkAudioThread()
	defer thread.UnmarkAudioThread()
	
	if p, ok := h.GetPlugin(plugin).(ProcessorWithHandle); ok {
		return C.int32_t(p.ProcessWithHandle(process))
	}
	return C.int32_t(api.ProcessError)
}

// PluginGetExtension implements ClapGo_PluginGetExtension
func (h *ExportHelpers) PluginGetExtension(plugin unsafe.Pointer, id *C.char) unsafe.Pointer {
	return h.GetPlugin(plugin).GetExtension(C.GoString(id))
}

// PluginOnMainThread implements ClapGo_PluginOnMainThread
func (h *ExportHelpers) PluginOnMainThread(plugin unsafe.Pointer) {
	h.GetPlugin(plugin).OnMainThread()
}

// ProcessorWithHandle is an interface for plugins that handle process calls directly
type ProcessorWithHandle interface {
	ProcessWithHandle(process unsafe.Pointer) int
}