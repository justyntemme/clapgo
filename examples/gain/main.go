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
	"github.com/justyntemme/clapgo/pkg/extension"
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
	
	// Parameter binding system
	params *param.ParameterBinder
	
	// Direct parameter access
	gain *param.AtomicFloat64
	
	// Extension bundle for host integration
	extensions *extension.ExtensionBundle
}


func NewGainPlugin() *GainPlugin {
	p := &GainPlugin{
		PluginBase:         plugin.NewPluginBase(pluginInfo),
		StereoPortProvider: audio.NewStereoPortProvider(),
		SurroundSupport:    audio.NewStereoSurroundSupport(),
	}
	
	// Create parameter binder for automatic registration + atomic storage
	p.params = param.NewParameterBinder(p.ParamManager)
	
	// Bind gain parameter - this creates atomic storage AND registers with manager
	p.gain = p.params.BindPercentage(ParamGain, "Gain", 100.0) // Default 100% (0dB)
	
	return p
}

func (p *GainPlugin) CreateWithHost(host unsafe.Pointer) cgo.Handle {
	p.PluginBase.InitWithHost(host)
	handle := cgo.NewHandle(p)
	audio.RegisterPortsProvider(unsafe.Pointer(handle), p)
	return handle
}

// Init initializes the plugin
func (p *GainPlugin) Init() bool {
	// Use framework's common initialization
	err := p.PluginBase.CommonInit()
	if err != nil {
		return false
	}
	
	// Initialize extension bundle for host integration
	p.extensions = extension.NewExtensionBundle(p.Host, PluginName)
	
	if p.extensions != nil {
		p.extensions.LogInfo("Gain plugin initialized")
	}
	
	return true
}

// Destroy cleans up plugin resources
func (p *GainPlugin) Destroy() {
	// Use framework's common cleanup
	p.PluginBase.CommonDestroy()
}


func (p *GainPlugin) StartProcessing() bool {
	// Use framework's common start processing with error handling and logging
	err := p.PluginBase.CommonStartProcessing()
	if err != nil {
		if p.extensions != nil {
			p.extensions.LogError(fmt.Sprintf("Failed to start processing: %v", err))
		}
		return false
	}
	
	return true
}

// StopProcessing ends audio processing
func (p *GainPlugin) StopProcessing() {
	// Use framework's common stop processing with logging
	p.PluginBase.CommonStopProcessing()
}

// Activate prepares the plugin for processing
func (p *GainPlugin) Activate(sampleRate float64, minFrames, maxFrames uint32) bool {
	// Use framework's common activation with logging
	err := p.PluginBase.CommonActivate(sampleRate, minFrames, maxFrames)
	if err != nil {
		if p.extensions != nil {
			p.extensions.LogError(fmt.Sprintf("Failed to activate: %v", err))
		}
		return false
	}
	
	return true
}

// Deactivate stops the plugin from processing
func (p *GainPlugin) Deactivate() {
	// Use framework's common deactivation with logging
	p.PluginBase.CommonDeactivate()
}

func (p *GainPlugin) Reset() {
	// Use framework's common reset (includes logging)
	p.PluginBase.CommonReset()
	
	// Reset plugin-specific state
	p.gain.Store(1.0)
}

func (p *GainPlugin) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events *event.Processor) int {
	// Check if we're in a valid state for processing
	if !p.IsActivated || !p.IsProcessing {
		return process.ProcessError
	}
	
	// Process events first
	if events != nil {
		events.ProcessAll(p)
	}
	
	// Get current gain value atomically
	gain := float32(p.gain.Load())
	
	// Validate buffers
	if !audio.ValidateBuffers(audioOut, audioIn) {
		if p.extensions != nil {
			p.extensions.LogError("Invalid audio buffers")
		}
		return process.ProcessError
	}
	
	// Apply gain using framework utility
	audio.ProcessStereo(audioIn, audioOut, func(sample float32) float32 {
		return sample * gain
	})
	
	return process.ProcessContinue
}


// HandleParamValue handles parameter value changes (implements event.Handler)
func (p *GainPlugin) HandleParamValue(paramEvent *event.ParamValueEvent, time uint32) {
	// Use parameter binder for automatic handling
	if p.params.HandleParamValue(paramEvent.ParamID, paramEvent.Value) {
		// Successfully handled by binder - log the change
		if p.extensions != nil && paramEvent.ParamID == ParamGain {
			value := p.gain.Load()
			db := audio.LinearToDb(value)
			p.extensions.LogDebug(fmt.Sprintf("Gain changed to %.1f%% (%.2f dB)", value*100, db))
		}
		return
	}
	
	// Fallback for unknown parameters (shouldn't happen in this plugin)
	if p.extensions != nil {
		p.extensions.LogWarning(fmt.Sprintf("Unknown parameter ID: %d", paramEvent.ParamID))
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

// Parameter text formatting is handled automatically by PluginBase
// via the parameter binder system - no need to implement here


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
	
	// Use modern remote controls builder pattern
	page := controls.NewRemoteControlsPageBuilder(0, "Gain Control").
		Section("Main").
		AddParameters(ParamGain).
		ClearRemaining().
		MustBuild()
	
	return &page, true
}


func (p *GainPlugin) SaveState(stream unsafe.Pointer) error {
	return p.SaveStateWithParams(stream, map[uint32]float64{
		ParamGain: p.gain.Load(),
	})
}

func (p *GainPlugin) LoadState(stream unsafe.Pointer) error {
	return p.LoadStateWithCallback(stream, func(id uint32, value float64) {
		if id == ParamGain {
			// Use parameter binder for consistent handling
			p.params.HandleParamValue(id, value)
			
			if p.extensions != nil {
				p.extensions.LogDebug(fmt.Sprintf("Loaded gain value: %.1f%%", value*100))
			}
		}
	})
}

// GetLatency and GetTail are provided by PluginBase (both return 0)
// No need to override for simple effects like gain


func main() {}