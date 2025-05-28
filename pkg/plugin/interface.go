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
	
	// ProcessBatch processes multiple frames with context support
	ProcessBatch(ctx context.Context, frames []ProcessFrame) error
}

// ProcessFrame represents a single frame of audio processing
type ProcessFrame struct {
	Input      [][]float32
	Output     [][]float32
	SteadyTime int64
	Events     []Event
}

// Event represents an audio event (MIDI, parameter change, etc.)
type Event struct {
	Type      EventType
	Time      uint32
	Data      interface{}
}

// EventType identifies the type of event
type EventType int

const (
	EventTypeNote EventType = iota
	EventTypeParameter
	EventTypeMIDI
	EventTypeTransport
)

// StatefulV2 handles state persistence with io interfaces
type StatefulV2 interface {
	SaveState(ctx context.Context, writer StateWriter) error
	LoadState(ctx context.Context, reader StateReader) error
	
	// SaveStateWithProgress saves state with progress reporting
	SaveStateWithProgress(ctx context.Context, writer StateWriter, progress chan<- ProgressUpdate) error
	
	// LoadStateWithProgress loads state with progress reporting
	LoadStateWithProgress(ctx context.Context, reader StateReader, progress chan<- ProgressUpdate) error
}

// ProgressUpdate represents progress information for long operations
type ProgressUpdate struct {
	Completed int64   // Bytes or items completed
	Total     int64   // Total bytes or items
	Message   string  // Human-readable progress message
	Percent   float64 // Completion percentage (0.0 to 1.0)
}

// PresetLoaderV2 handles preset loading with context support
type PresetLoaderV2 interface {
	// LoadPreset loads a preset with context for cancellation
	LoadPreset(ctx context.Context, location string) error
	
	// LoadPresetWithProgress loads preset with progress reporting
	LoadPresetWithProgress(ctx context.Context, location string, progress chan<- ProgressUpdate) error
	
	// ValidatePreset validates a preset file without loading it
	ValidatePreset(ctx context.Context, location string) error
}

// ParameterV2 provides context-aware parameter operations
type ParameterV2 interface {
	// GetParameter gets a parameter value with context
	GetParameter(ctx context.Context, id uint32) (float64, error)
	
	// SetParameter sets a parameter value with context
	SetParameter(ctx context.Context, id uint32, value float64) error
	
	// GetParameterText gets parameter display text with context
	GetParameterText(ctx context.Context, id uint32, value float64) (string, error)
	
	// BeginParameterChanges starts a batch of parameter changes
	BeginParameterChanges(ctx context.Context) (ParameterTransaction, error)
}

// ParameterTransaction allows batch parameter changes with rollback
type ParameterTransaction interface {
	// SetParameter sets a parameter in the transaction
	SetParameter(id uint32, value float64) error
	
	// Commit applies all parameter changes
	Commit() error
	
	// Rollback cancels all parameter changes
	Rollback() error
	
	// Context returns the transaction context
	Context() context.Context
}

// StateWriter abstracts state writing
type StateWriter interface {
	WriteUint32(v uint32) error
	WriteFloat64(v float64) error
	WriteString(s string) error
	WriteBytes(b []byte) error
}

// StateWriterV2 extends StateWriter with context support
type StateWriterV2 interface {
	StateWriter
	WriteUint32WithContext(ctx context.Context, v uint32) error
	WriteFloat64WithContext(ctx context.Context, v float64) error
	WriteStringWithContext(ctx context.Context, s string) error
	WriteBytesWithContext(ctx context.Context, b []byte) error
}

// StateReader abstracts state reading
type StateReader interface {
	ReadUint32() (uint32, error)
	ReadFloat64() (float64, error)
	ReadString() (string, error)
	ReadBytes(n int) ([]byte, error)
}

// StateReaderV2 extends StateReader with context support
type StateReaderV2 interface {
	StateReader
	ReadUint32WithContext(ctx context.Context) (uint32, error)
	ReadFloat64WithContext(ctx context.Context) (float64, error)
	ReadStringWithContext(ctx context.Context) (string, error)
	ReadBytesWithContext(ctx context.Context, n int) ([]byte, error)
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