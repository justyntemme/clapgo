package plugin

import (
	"context"
	"errors"
	"unsafe"
)

// Common plugin errors
var (
	ErrNotInitialized      = errors.New("plugin not initialized")
	ErrAlreadyInitialized  = errors.New("plugin already initialized")
	ErrNotActivated        = errors.New("plugin not activated")
	ErrAlreadyActivated    = errors.New("plugin already activated")
	ErrProcessingActive    = errors.New("audio processing is active")
	ErrInvalidSampleRate   = errors.New("invalid sample rate")
	ErrInvalidFrameCount   = errors.New("invalid frame count")
	ErrInvalidParameter    = errors.New("invalid parameter")
	ErrInvalidState        = errors.New("invalid plugin state")
	ErrUnsupportedExtension = errors.New("unsupported extension")
)

// PluginV2 is the idiomatic Go interface for plugins
// This interface uses errors instead of bool returns
type PluginV2 interface {
	// Lifecycle methods
	Init() error
	Destroy() error
	Activate(sampleRate float64, minFrames, maxFrames uint32) error
	Deactivate() error
	StartProcessing() error
	StopProcessing() error
	Reset() error
	
	// Extensions
	GetExtension(id string) (unsafe.Pointer, error)
	OnMainThread() error
	
	// Info
	GetInfo() Info
}

// ProcessorV2 handles audio processing with context support
type ProcessorV2 interface {
	// Process audio with context for cancellation
	Process(ctx context.Context, in, out [][]float32, steadyTime int64) error
}

// StatefulV2 handles state persistence with io interfaces
type StatefulV2 interface {
	SaveState(ctx context.Context, writer StateWriter) error
	LoadState(ctx context.Context, reader StateReader) error
}

// StateWriter abstracts state writing
type StateWriter interface {
	WriteUint32(v uint32) error
	WriteFloat64(v float64) error
	WriteString(s string) error
	WriteBytes(b []byte) error
}

// StateReader abstracts state reading
type StateReader interface {
	ReadUint32() (uint32, error)
	ReadFloat64() (float64, error)
	ReadString() (string, error)
	ReadBytes(n int) ([]byte, error)
}

// ParameterError provides detailed parameter operation errors
type ParameterError struct {
	Op      string  // operation like "get", "set", "format"
	ParamID uint32
	Value   float64
	Err     error
}

func (e *ParameterError) Error() string {
	if e.Value != 0 {
		return "param " + e.Op + " " + string(e.ParamID) + " value " + string(int(e.Value)) + ": " + e.Err.Error()
	}
	return "param " + e.Op + " " + string(e.ParamID) + ": " + e.Err.Error()
}

func (e *ParameterError) Unwrap() error {
	return e.Err
}

// ProcessError provides detailed processing errors
type ProcessError struct {
	Frame   uint32
	Channel uint32
	Err     error
}

func (e *ProcessError) Error() string {
	return "process frame " + string(e.Frame) + " channel " + string(e.Channel) + ": " + e.Err.Error()
}

func (e *ProcessError) Unwrap() error {
	return e.Err
}

// ExtensionError provides detailed extension errors
type ExtensionError struct {
	ID  string
	Err error
}

func (e *ExtensionError) Error() string {
	return "extension " + e.ID + ": " + e.Err.Error()
}

func (e *ExtensionError) Unwrap() error {
	return e.Err
}

// Adapter wraps the old Plugin interface to implement PluginV2
// This allows gradual migration
type Adapter struct {
	Plugin Plugin
}

// NewAdapter creates an adapter for old-style plugins
func NewAdapter(p Plugin) *Adapter {
	return &Adapter{Plugin: p}
}

func (a *Adapter) Init() error {
	if !a.Plugin.Init() {
		return ErrInitFailed
	}
	return nil
}

func (a *Adapter) Destroy() error {
	a.Plugin.Destroy()
	return nil
}

func (a *Adapter) Activate(sampleRate float64, minFrames, maxFrames uint32) error {
	if sampleRate <= 0 {
		return ErrInvalidSampleRate
	}
	if !a.Plugin.Activate(sampleRate, minFrames, maxFrames) {
		return errors.New("activation failed")
	}
	return nil
}

func (a *Adapter) Deactivate() error {
	a.Plugin.Deactivate()
	return nil
}

func (a *Adapter) StartProcessing() error {
	if !a.Plugin.StartProcessing() {
		return errors.New("failed to start processing")
	}
	return nil
}

func (a *Adapter) StopProcessing() error {
	a.Plugin.StopProcessing()
	return nil
}

func (a *Adapter) Reset() error {
	a.Plugin.Reset()
	return nil
}

func (a *Adapter) GetExtension(id string) (unsafe.Pointer, error) {
	ext := a.Plugin.GetExtension(id)
	if ext == nil {
		return nil, &ExtensionError{ID: id, Err: ErrUnsupportedExtension}
	}
	return ext, nil
}

func (a *Adapter) OnMainThread() error {
	a.Plugin.OnMainThread()
	return nil
}

func (a *Adapter) GetInfo() Info {
	// This would need to be implemented based on the plugin
	return Info{}
}