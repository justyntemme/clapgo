package main

// #include "../../include/clap/include/clap/ext/remote-controls.h"
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

type GainPlugin struct {
	*plugin.PluginBase
	*audio.StereoPortProvider
	*audio.SurroundSupport
	
	gain param.AtomicFloat64
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

// GetParamInfo is provided by PluginBase

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

func main() {}