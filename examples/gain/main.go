package main

// #cgo CFLAGS: -I../../include/clap/include
// #include "../../include/clap/include/clap/clap.h"
// #include "../../include/clap/include/clap/ext/remote-controls.h"
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"runtime/cgo"
	"unsafe"
	
	"github.com/justyntemme/clapgo/pkg/api"
	"github.com/justyntemme/clapgo/pkg/audio"
	"github.com/justyntemme/clapgo/pkg/param"
	"github.com/justyntemme/clapgo/pkg/plugin"
)


// Global plugin instance and shared data
var (
	gainPlugin *GainPlugin
	
	pluginInfo = plugin.Info{
		ID:          PluginID,
		Name:        PluginName,
		Vendor:      PluginVendor,
		Version:     PluginVersion,
		Description: PluginDescription,
		URL:         "https://github.com/justyntemme/clapgo",
		Manual:      "https://github.com/justyntemme/clapgo",
		Support:     "https://github.com/justyntemme/clapgo/issues",
		Features:    []string{plugin.FeatureAudioEffect, plugin.FeatureStereo, plugin.FeatureMono},
	}
)

func init() {
	gainPlugin = NewGainPlugin()
}

func getPlugin(plugin unsafe.Pointer) *GainPlugin {
	if plugin == nil {
		return &GainPlugin{}
	}
	return cgo.Handle(plugin).Value().(*GainPlugin)
}


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
func ClapGo_PluginActivate(plugin unsafe.Pointer, sampleRate C.double, minFrames, maxFrames C.uint32_t) C.bool {
	return C.bool(getPlugin(plugin).Activate(float64(sampleRate), uint32(minFrames), uint32(maxFrames)))
}

//export ClapGo_PluginDeactivate
func ClapGo_PluginDeactivate(plugin unsafe.Pointer) {
	getPlugin(plugin).Deactivate()
}

//export ClapGo_PluginStartProcessing
func ClapGo_PluginStartProcessing(plugin unsafe.Pointer) C.bool {
	api.DebugMarkAudioThread()
	defer api.DebugUnmarkAudioThread()
	return C.bool(getPlugin(plugin).StartProcessing())
}

//export ClapGo_PluginStopProcessing
func ClapGo_PluginStopProcessing(plugin unsafe.Pointer) {
	api.DebugMarkAudioThread()
	defer api.DebugUnmarkAudioThread()
	getPlugin(plugin).StopProcessing()
}

//export ClapGo_PluginReset
func ClapGo_PluginReset(plugin unsafe.Pointer) {
	getPlugin(plugin).Reset()
}

//export ClapGo_PluginProcess
func ClapGo_PluginProcess(plugin unsafe.Pointer, process unsafe.Pointer) C.int32_t {
	api.DebugMarkAudioThread()
	defer api.DebugUnmarkAudioThread()
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

type GainPlugin struct {
	*plugin.PluginBase
	*audio.StereoPortProvider
	*audio.SurroundSupport
	
	gain param.AtomicFloat64
	
	contextMenuProvider *api.DefaultContextMenuProvider
}


func NewGainPlugin() *GainPlugin {
	p := &GainPlugin{
		PluginBase:         plugin.NewPluginBase(pluginInfo),
		StereoPortProvider: audio.NewStereoPortProvider(),
		SurroundSupport:    audio.NewStereoSurroundSupport(),
	}
	
	p.gain.Store(1.0)
	
	p.ParamManager.Register(param.Volume(ParamGain, "Gain"))
	p.ParamManager.SetValue(ParamGain, 1.0)
	
	p.contextMenuProvider = api.NewDefaultContextMenuProvider(nil, PluginName, PluginVersion, nil)
	
	return p
}

func (p *GainPlugin) CreateWithHost(host unsafe.Pointer) cgo.Handle {
	p.PluginBase.InitWithHost(host)
	handle := cgo.NewHandle(p)
	api.RegisterAudioPortsProvider(unsafe.Pointer(handle), p)
	return handle
}


func (p *GainPlugin) StartProcessing() bool {
	api.DebugAssertAudioThread("GainPlugin.StartProcessing")
	if p.ThreadCheck != nil {
		p.ThreadCheck.AssertAudioThread("GainPlugin.StartProcessing")
	}
	
	return p.PluginBase.CommonStartProcessing()
}

func (p *GainPlugin) Reset() {
	p.PluginBase.CommonReset()
	p.gain.Store(1.0)
}

func (p *GainPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events api.EventHandler) int {
	api.DebugAssertAudioThread("GainPlugin.Process")
	if p.ThreadCheck != nil {
		p.ThreadCheck.AssertAudioThread("GainPlugin.Process")
	}
	
	if !p.IsActivated || !p.IsProcessing {
		return api.ProcessError
	}
	
	if events != nil {
		p.processEvents(events, framesCount)
	}
	
	gain := float32(p.gain.Load())
	
	// Apply gain to each channel
	for ch := 0; ch < len(audioOut) && ch < len(audioIn); ch++ {
		if len(audioOut[ch]) != len(audioIn[ch]) {
			if p.Logger != nil {
				p.Logger.Error("Frame count mismatch between input and output")
			}
			return api.ProcessError
		}
		
		// Process each sample
		for i := 0; i < len(audioOut[ch]); i++ {
			audioOut[ch][i] = audioIn[ch][i] * gain
		}
	}
	
	return api.ProcessContinue
}

func (p *GainPlugin) processEvents(events api.EventHandler, frameCount uint32) {
	if events == nil {
		return
	}
	events.ProcessTypedEvents(p)
}

func (p *GainPlugin) HandleParamValue(paramEvent *api.ParamValueEvent, time uint32) {
	switch paramEvent.ParamID {
	case ParamGain:
		// Clamp value to valid range
		value := param.ClampValue(paramEvent.Value, 0.0, 2.0)
		
		p.gain.Store(value)
		if err := p.ParamManager.SetValue(paramEvent.ParamID, value); err != nil {
			if p.Logger != nil {
				p.Logger.Warning(fmt.Sprintf("Failed to set parameter %d: %v", paramEvent.ParamID, err))
			}
		}
		
		if p.Logger != nil {
			db := audio.LinearToDb(value)
			p.Logger.Debug(fmt.Sprintf("Gain changed to %.2f dB", db))
		}
	}
}


func (p *GainPlugin) ProcessWithHandle(process unsafe.Pointer) int {
	if process == nil {
		return api.ProcessError
	}
	
	cProcess := (*C.clap_process_t)(process)
	
	steadyTime := int64(cProcess.steady_time)
	framesCount := uint32(cProcess.frames_count)
	
	audioIn := api.ConvertFromCBuffers(unsafe.Pointer(cProcess.audio_inputs), uint32(cProcess.audio_inputs_count), framesCount)
	audioOut := api.ConvertFromCBuffers(unsafe.Pointer(cProcess.audio_outputs), uint32(cProcess.audio_outputs_count), framesCount)
	
	eventHandler := api.NewEventProcessor(
		unsafe.Pointer(cProcess.in_events),
		unsafe.Pointer(cProcess.out_events),
	)
	
	api.SetupPoolLogging(eventHandler, p.Logger)
	
	result := p.Process(steadyTime, framesCount, audioIn, audioOut, eventHandler)
	
	p.PoolDiagnostics.LogPoolDiagnostics(eventHandler, 1000)
	
	return result
}

func (p *GainPlugin) GetParamInfo(index uint32, info unsafe.Pointer) bool {
	if info == nil {
		return false
	}
	
	paramInfo, err := p.ParamManager.GetInfoByIndex(index)
	if err != nil {
		return false
	}
	
	param.InfoToC(paramInfo, info)
	
	return true
}

func (p *GainPlugin) GetParamValue(paramID uint32, value *C.double) bool {
	if value == nil {
		return false
	}
	
	if paramID == ParamGain {
		*value = C.double(p.gain.Load())
		return true
	}
	
	return false
}

func (p *GainPlugin) ParamValueToText(paramID uint32, value float64, buffer *C.char, size uint32) bool {
	if buffer == nil || size == 0 {
		return false
	}
	
	if paramID == ParamGain {
		text := param.FormatValue(value, param.FormatDecibel)
		
			bytes := []byte(text)
		if len(bytes) >= int(size) {
			bytes = bytes[:size-1]
		}
		for i, b := range bytes {
			*(*C.char)(unsafe.Add(unsafe.Pointer(buffer), i)) = C.char(b)
		}
		*(*C.char)(unsafe.Add(unsafe.Pointer(buffer), len(bytes))) = 0
		
		return true
	}
	
	return false
}

func (p *GainPlugin) ParamTextToValue(paramID uint32, text string, value *C.double) bool {
	if value == nil {
		return false
	}
	
	if paramID == ParamGain {
		parser := param.NewParser(param.FormatDecibel)
		if parsedValue, err := parser.ParseValue(text); err == nil {
					clamped := param.ClampValue(parsedValue, 0.0, 2.0)
			*value = C.double(clamped)
			return true
		}
	}
	
	return false
}

func (p *GainPlugin) ParamsFlush(inEvents, outEvents unsafe.Pointer) {
	if inEvents != nil {
		eventHandler := api.NewEventProcessor(inEvents, outEvents)
		p.processEvents(eventHandler, 0)
	}
}

func (p *GainPlugin) PopulateContextMenuWithTarget(targetKind uint32, targetID uint64, builder unsafe.Pointer) bool {
	if builder == nil {
		return false
	}
	
	// Create target
	var target *api.ContextMenuTarget
	if targetKind != api.ContextMenuTargetKindGlobal {
		target = &api.ContextMenuTarget{
			Kind: targetKind,
			ID:   targetID,
		}
	}
	
	// Create builder wrapper
	menuBuilder := api.NewContextMenuBuilder(builder)
	
	return p.PopulateContextMenu(target, menuBuilder)
}

func (p *GainPlugin) PerformContextMenuActionWithTarget(targetKind uint32, targetID uint64, actionID uint64) bool {
	// Create target
	var target *api.ContextMenuTarget
	if targetKind != api.ContextMenuTargetKindGlobal {
		target = &api.ContextMenuTarget{
			Kind: targetKind,
			ID:   targetID,
		}
	}
	
	return p.PerformContextMenuAction(target, actionID)
}

func (p *GainPlugin) GetRemoteControlsPageToC(pageIndex uint32, cPage unsafe.Pointer) bool {
	if cPage == nil {
		return false
	}
	
	page, ok := p.GetRemoteControlsPage(pageIndex)
	if !ok {
		return false
	}
	
	// Convert Go page to C structure
	api.RemoteControlsPageToC(page, cPage)
	return true
}

func (p *GainPlugin) GetExtension(id string) unsafe.Pointer {
	switch id {
	case api.ExtPresetLoad:
		return unsafe.Pointer(&p)
	default:
		return nil
	}
}

func (p *GainPlugin) GetPluginInfo() api.PluginInfo {
	return api.PluginInfo{
		ID:          pluginInfo.ID,
		Name:        pluginInfo.Name,
		Vendor:      pluginInfo.Vendor,
		URL:         pluginInfo.URL,
		ManualURL:   pluginInfo.Manual,
		SupportURL:  pluginInfo.Support,
		Version:     pluginInfo.Version,
		Description: pluginInfo.Description,
		Features:    pluginInfo.Features,
	}
}


func (p *GainPlugin) PopulateContextMenu(target *api.ContextMenuTarget, builder *api.ContextMenuBuilder) bool {
	api.DebugAssertMainThread("GainPlugin.PopulateContextMenu")
	
	if target != nil && target.Kind == api.ContextMenuTargetKindParam {
			p.contextMenuProvider.PopulateParameterMenu(uint32(target.ID), builder)
		
	} else {
		p.contextMenuProvider.PopulateGlobalMenu(builder)
	}
	
	return true
}

func (p *GainPlugin) PerformContextMenuAction(target *api.ContextMenuTarget, actionID uint64) bool {
	api.DebugAssertMainThread("GainPlugin.PerformContextMenuAction")
	
	if isReset, paramID := p.contextMenuProvider.IsResetAction(actionID); isReset {
		if paramID == 0 {
			p.gain.Store(1.0)
		}
		return p.contextMenuProvider.HandleResetParameter(paramID)
	}
	
	if p.contextMenuProvider.IsAboutAction(actionID) {
		if p.Logger != nil {
			p.Logger.Info("Gain Plugin v1.0.0 - A simple gain adjustment plugin")
		}
		return true
	}
	
	return false
}


func (p *GainPlugin) GetRemoteControlsPageCount() uint32 {
	return 1
}

func (p *GainPlugin) GetRemoteControlsPage(pageIndex uint32) (*api.RemoteControlsPage, bool) {
	if pageIndex != 0 {
		return nil, false
	}
	
	return &api.RemoteControlsPage{
		SectionName: "Main",
		PageID:      1,
		PageName:    "Gain Control",
		ParamIDs:    [api.RemoteControlsCount]uint32{ParamGain},
		IsForPreset: false,
	}, true
}


func (p *GainPlugin) SaveState(stream unsafe.Pointer) bool {
	return p.SaveStateWithParams(stream, map[uint32]float64{
		ParamGain: p.gain.Load(),
	})
}

func (p *GainPlugin) LoadState(stream unsafe.Pointer) bool {
	return p.LoadStateWithCallback(stream, func(id uint32, value float64) {
		if id == ParamGain {
			p.gain.Store(value)
			p.ParamManager.SetValue(id, value)
		}
	})
}

func (p *GainPlugin) SaveStateWithContext(stream unsafe.Pointer, contextType uint32) bool {
	p.PluginBase.SaveStateWithContext(stream, contextType)
	return p.SaveState(stream)
}

  
func (p *GainPlugin) LoadStateWithContext(stream unsafe.Pointer, contextType uint32) bool {
	p.PluginBase.LoadStateWithContext(stream, contextType)
	return p.LoadState(stream)
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
	return C.bool(getPlugin(plugin).PopulateContextMenuWithTarget(uint32(targetKind), uint64(targetID), builder))
}

//export ClapGo_PluginContextMenuPerform
func ClapGo_PluginContextMenuPerform(plugin unsafe.Pointer, targetKind C.uint32_t, targetID C.uint64_t, actionID C.uint64_t) C.bool {
	return C.bool(getPlugin(plugin).PerformContextMenuActionWithTarget(uint32(targetKind), uint64(targetID), uint64(actionID)))
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

func main() {}