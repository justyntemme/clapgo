package plugin

import (
	"errors"
	"unsafe"
)

// Common errors
var (
	ErrNotImplemented = errors.New("not implemented")
	ErrInvalidHost    = errors.New("invalid host")
	ErrInitFailed     = errors.New("initialization failed")
)

// Info represents plugin metadata
type Info struct {
	ID          string
	Name        string
	Vendor      string
	URL         string
	Version     string
	Description string
	Manual      string
	Support     string
	Features    []string
}

// Interface defines the core plugin interface
type Interface interface {
	// Lifecycle
	Init() error
	Destroy()
	Activate(sampleRate float64, minFrames, maxFrames uint32) error
	Deactivate()
	StartProcessing() error
	StopProcessing()
	Reset()
	
	// Processing
	Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events interface{}) ProcessResult
	
	// Extensions
	GetExtension(id string) unsafe.Pointer
	OnMainThread()
	
	// Info
	GetPluginID() string
	GetPluginInfo() Info
}

// ProcessResult represents the result of audio processing
type ProcessResult int

const (
	ProcessResultError   ProcessResult = -1
	ProcessContinue      ProcessResult = 0
	ProcessContinueIfNotQuiet ProcessResult = 1
	ProcessTail          ProcessResult = 2
	ProcessSleep         ProcessResult = 3
)

// Base provides common functionality for plugins
type Base struct {
	Host   unsafe.Pointer
	ID     string
	Info   Info
	Logger interface {
		Debug(format string, args ...interface{})
		Info(format string, args ...interface{})
		Warning(format string, args ...interface{})
		Error(format string, args ...interface{})
	}
}

// NewBase creates a new plugin base
func NewBase(id string, info Info) *Base {
	return &Base{
		ID:   id,
		Info: info,
	}
}

// SetHost sets the host pointer
func (b *Base) SetHost(host unsafe.Pointer) {
	b.Host = host
}

// GetPluginID returns the plugin ID
func (b *Base) GetPluginID() string {
	return b.ID
}

// GetPluginInfo returns plugin information
func (b *Base) GetPluginInfo() Info {
	return b.Info
}

// Init initializes the plugin (default implementation)
func (b *Base) Init() error {
	if b.Logger != nil {
		b.Logger.Info("Plugin initialized: %s", b.Info.Name)
	}
	return nil
}

// Destroy destroys the plugin (default implementation)
func (b *Base) Destroy() {
	if b.Logger != nil {
		b.Logger.Info("Plugin destroyed: %s", b.Info.Name)
	}
}

// Activate activates the plugin (default implementation)
func (b *Base) Activate(sampleRate float64, minFrames, maxFrames uint32) error {
	if b.Logger != nil {
		b.Logger.Info("Plugin activated: %.0fHz, frames: %d-%d", sampleRate, minFrames, maxFrames)
	}
	return nil
}

// Deactivate deactivates the plugin (default implementation)
func (b *Base) Deactivate() {
	if b.Logger != nil {
		b.Logger.Info("Plugin deactivated")
	}
}

// StartProcessing starts audio processing (default implementation)
func (b *Base) StartProcessing() error {
	if b.Logger != nil {
		b.Logger.Info("Processing started")
	}
	return nil
}

// StopProcessing stops audio processing (default implementation)
func (b *Base) StopProcessing() {
	if b.Logger != nil {
		b.Logger.Info("Processing stopped")
	}
}

// Reset resets the plugin state (default implementation)
func (b *Base) Reset() {
	if b.Logger != nil {
		b.Logger.Info("Plugin reset")
	}
}

// GetExtension returns an extension (default returns nil)
func (b *Base) GetExtension(id string) unsafe.Pointer {
	return nil
}

// OnMainThread is called on the main thread (default implementation)
func (b *Base) OnMainThread() {
	// Default: do nothing
}


// Common plugin features
const (
	FeatureInstrument      = "instrument"
	FeatureAudioEffect     = "audio-effect"
	FeatureNoteEffect      = "note-effect"
	FeatureNoteDetector    = "note-detector"
	FeatureAnalyzer        = "analyzer"
	FeatureSynthesizer     = "synthesizer"
	FeatureSampler         = "sampler"
	FeatureDrum            = "drum"
	FeatureFilter          = "filter"
	FeaturePhaser          = "phaser"
	FeatureEqualizer       = "equalizer"
	FeatureDeesser         = "de-esser"
	FeaturePhaseVocoder    = "phase-vocoder"
	FeatureGranular        = "granular"
	FeatureFrequencyShifter = "frequency-shifter"
	FeaturePitchShifter    = "pitch-shifter"
	FeatureDistortion      = "distortion"
	FeatureTransientShaper = "transient-shaper"
	FeatureCompressor      = "compressor"
	FeatureExpander        = "expander"
	FeatureGate            = "gate"
	FeatureLimiter         = "limiter"
	FeatureFlanger         = "flanger"
	FeatureChorus          = "chorus"
	FeatureDelay           = "delay"
	FeatureReverb          = "reverb"
	FeatureTremolo         = "tremolo"
	FeatureGlitch          = "glitch"
	FeatureUtility         = "utility"
	FeaturePitchCorrection = "pitch-correction"
	FeatureRestoration     = "restoration"
	FeatureMultiEffects    = "multi-effects"
	FeatureMixing          = "mixing"
	FeatureMastering       = "mastering"
	FeatureMono            = "mono"
	FeatureStereo          = "stereo"
	FeatureSurround        = "surround"
	FeatureAmbisonic       = "ambisonic"
)