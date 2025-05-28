package main

// #include "../../include/clap/include/clap/ext/remote-controls.h"
import "C"
import (
	"fmt"
	"runtime/cgo"
	"unsafe"
	
	"github.com/justyntemme/clapgo/pkg/audio"
	"github.com/justyntemme/clapgo/pkg/controls"
	"github.com/justyntemme/clapgo/pkg/event"
	"github.com/justyntemme/clapgo/pkg/param"
	"github.com/justyntemme/clapgo/pkg/plugin"
	"github.com/justyntemme/clapgo/pkg/process"
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
	event.NoOpHandler
	
	gain param.AtomicFloat64
}


func NewGainPlugin() *GainPlugin {
	p := &GainPlugin{
		PluginBase:         plugin.NewPluginBase(pluginInfo),
		StereoPortProvider: audio.NewStereoPortProvider(),
		SurroundSupport:    audio.NewStereoSurroundSupport(),
	}
	
	p.gain.Store(1.0)
	
	if err := p.ParamManager.Register(param.Volume(ParamGain, "Gain")); err != nil {
		// In a real plugin, we might want to handle this error differently
		panic("Failed to register gain parameter: " + err.Error())
	}
	p.ParamManager.SetValue(ParamGain, 1.0)
	
	return p
}

func (p *GainPlugin) CreateWithHost(host unsafe.Pointer) cgo.Handle {
	p.PluginBase.InitWithHost(host)
	handle := cgo.NewHandle(p)
	audio.RegisterPortsProvider(unsafe.Pointer(handle), p)
	return handle
}


func (p *GainPlugin) StartProcessing() error {
	return p.PluginBase.CommonStartProcessing()
}

func (p *GainPlugin) Reset() {
	p.PluginBase.CommonReset()
	p.gain.Store(1.0)
}

func (p *GainPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events *event.Processor) int {
	if !p.IsActivated || !p.IsProcessing {
		return process.ProcessError
	}
	
	if events != nil {
		events.ProcessAll(p)
	}
	
	gain := float32(p.gain.Load())
	
	// Validate buffers
	if !audio.ValidateBuffers(audioOut, audioIn) {
		if p.Logger != nil {
			p.Logger.Error("Invalid audio buffers")
		}
		return process.ProcessError
	}
	
	// Apply gain using ProcessStereo
	audio.ProcessStereo(audioIn, audioOut, func(sample float32) float32 {
		return sample * gain
	})
	
	return process.ProcessContinue
}


// HandleParamValue handles parameter value changes (implements event.Handler)
func (p *GainPlugin) HandleParamValue(paramEvent *event.ParamValueEvent, time uint32) {
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


func (p *GainPlugin) ProcessWithHandle(processPtr unsafe.Pointer) int {
	if processPtr == nil {
		return process.ProcessError
	}
	
	cProcess := (*C.clap_process_t)(processPtr)
	
	steadyTime := int64(cProcess.steady_time)
	framesCount := uint32(cProcess.frames_count)
	
	audioIn := audio.ConvertFromCBuffers(unsafe.Pointer(cProcess.audio_inputs), uint32(cProcess.audio_inputs_count), framesCount)
	audioOut := audio.ConvertFromCBuffers(unsafe.Pointer(cProcess.audio_outputs), uint32(cProcess.audio_outputs_count), framesCount)
	
	eventHandler := event.NewProcessor(
		unsafe.Pointer(cProcess.in_events),
		unsafe.Pointer(cProcess.out_events),
	)
	
	event.SetupPoolLogging(eventHandler, p.Logger)
	
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
	
	// Delegate to base for other parameters
	return p.PluginBase.GetParamValue(paramID, unsafe.Pointer(value))
}


func (p *GainPlugin) ParamsFlush(inEvents, outEvents unsafe.Pointer) {
	if inEvents != nil {
		eventHandler := event.NewProcessor(inEvents, outEvents)
		eventHandler.ProcessAll(p)
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
	controls.RemoteControlsPageToC(page, cPage)
	return true
}






func (p *GainPlugin) GetRemoteControlsPageCount() uint32 {
	return 1
}

func (p *GainPlugin) GetRemoteControlsPage(pageIndex uint32) (*controls.RemoteControlsPage, bool) {
	if pageIndex != 0 {
		return nil, false
	}
	
	return &controls.RemoteControlsPage{
		SectionName: "Main",
		PageID:      1,
		PageName:    "Gain Control",
		ParamIDs:    [controls.RemoteControlsCount]uint32{ParamGain},
		IsForPreset: false,
	}, true
}


func (p *GainPlugin) SaveState(stream unsafe.Pointer) error {
	return p.SaveStateWithParams(stream, map[uint32]float64{
		ParamGain: p.gain.Load(),
	})
}

func (p *GainPlugin) LoadState(stream unsafe.Pointer) error {
	return p.LoadStateWithCallback(stream, func(id uint32, value float64) {
		if id == ParamGain {
			p.gain.Store(value)
			p.ParamManager.SetValue(id, value)
		}
	})
}


func main() {}