package goclap

// #include <stdlib.h>
// #include <string.h>
// #include "../../include/clap/include/clap/clap.h"
// #include "../../include/clap/include/clap/ext/state.h"
import "C"
import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"unsafe"
)

// Constants for state extension
const (
	ExtState = "clap.state"
)

// State versioning
const (
	StateVersion1 = 1
)

// Define error types for state operations
type StateError string

func (e StateError) Error() string {
	return string(e)
}

const (
	ErrInvalidStream  = StateError("invalid stream")
	ErrInvalidVersion = StateError("invalid state version")
	ErrInvalidData    = StateError("invalid state data")
	ErrWriteError     = StateError("error writing to stream")
	ErrReadError      = StateError("error reading from stream")
)

// State represents serializable plugin state
type State struct {
	// Version is the state format version
	Version uint32

	// ParameterValues maps parameter IDs to their values
	ParameterValues map[uint32]float64

	// CustomData contains plugin-specific data
	CustomData map[string]interface{}
}

// NewState creates a new state object with default values
func NewState() *State {
	return &State{
		Version:         StateVersion1,
		ParameterValues: make(map[uint32]float64),
		CustomData:      make(map[string]interface{}),
	}
}

// Encode serializes the state to a byte array
func (s *State) Encode() ([]byte, error) {
	// Encode the state as JSON
	data, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("error encoding state: %w", err)
	}
	
	// Create a buffer to hold the encoded state
	buf := new(bytes.Buffer)
	
	// Write the size of the data
	err = binary.Write(buf, binary.LittleEndian, uint32(len(data)))
	if err != nil {
		return nil, fmt.Errorf("error writing data size: %w", err)
	}
	
	// Write the data
	_, err = buf.Write(data)
	if err != nil {
		return nil, fmt.Errorf("error writing data: %w", err)
	}
	
	return buf.Bytes(), nil
}

// Decode deserializes the state from a byte array
func (s *State) Decode(data []byte) error {
	// Create a buffer to read from
	buf := bytes.NewBuffer(data)
	
	// Read the size of the data
	var size uint32
	err := binary.Read(buf, binary.LittleEndian, &size)
	if err != nil {
		return fmt.Errorf("error reading data size: %w", err)
	}
	
	// Ensure we have enough data
	if buf.Len() < int(size) {
		return fmt.Errorf("data size mismatch: expected %d, got %d", size, buf.Len())
	}
	
	// Read the data
	jsonData := make([]byte, size)
	_, err = buf.Read(jsonData)
	if err != nil {
		return fmt.Errorf("error reading data: %w", err)
	}
	
	// Decode the JSON data
	err = json.Unmarshal(jsonData, s)
	if err != nil {
		return fmt.Errorf("error decoding state: %w", err)
	}
	
	// Verify the version
	if s.Version != StateVersion1 {
		return ErrInvalidVersion
	}
	
	return nil
}

// InputStreamReader wraps a CLAP input stream for Go IO operations
type InputStreamReader struct {
	stream *C.clap_istream_t
}

// Read implements the io.Reader interface
func (r *InputStreamReader) Read(p []byte) (n int, err error) {
	if r.stream == nil {
		return 0, ErrInvalidStream
	}
	
	// For now, we'll return a placeholder
	// In a real implementation, we would call through C to read from the stream
	return 0, io.EOF
}

// OutputStreamWriter wraps a CLAP output stream for Go IO operations
type OutputStreamWriter struct {
	stream *C.clap_ostream_t
}

// Write implements the io.Writer interface
func (w *OutputStreamWriter) Write(p []byte) (n int, err error) {
	if w.stream == nil {
		return 0, ErrInvalidStream
	}
	
	// For now, we'll return success
	// In a real implementation, we would call through C to write to the stream
	return len(p), nil
}

// ReadFromStream reads state data from a CLAP input stream
func ReadFromStream(stream *C.clap_istream_t) (*State, error) {
	if stream == nil {
		return nil, ErrInvalidStream
	}
	
	reader := &InputStreamReader{stream: stream}
	
	// Create a buffer to read the data into
	var buf bytes.Buffer
	
	// Copy the stream to the buffer
	_, err := io.Copy(&buf, reader)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error reading from stream: %w", err)
	}
	
	// Create a new state object
	state := NewState()
	
	// Decode the state
	err = state.Decode(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("error decoding state: %w", err)
	}
	
	return state, nil
}

// WriteToStream writes state data to a CLAP output stream
func WriteToStream(state *State, stream *C.clap_ostream_t) error {
	if stream == nil {
		return ErrInvalidStream
	}
	
	if state == nil {
		state = NewState()
	}
	
	// Encode the state
	data, err := state.Encode()
	if err != nil {
		return fmt.Errorf("error encoding state: %w", err)
	}
	
	writer := &OutputStreamWriter{stream: stream}
	
	// Write the data to the stream
	_, err = writer.Write(data)
	if err != nil {
		return fmt.Errorf("error writing to stream: %w", err)
	}
	
	return nil
}

// PluginStateExtension represents the CLAP state extension for plugins
type PluginStateExtension struct {
	plugin       AudioProcessor
	stateExt     unsafe.Pointer // Pointer to the C extension
}

// NewPluginStateExtension creates a new state extension
func NewPluginStateExtension(plugin AudioProcessor) *PluginStateExtension {
	if plugin == nil {
		return nil
	}
	
	ext := &PluginStateExtension{
		plugin: plugin,
	}
	
	// Create the C interface
	ext.stateExt = createStateExtension(ext)
	
	return ext
}

// GetExtensionPointer returns the C extension interface pointer
func (s *PluginStateExtension) GetExtensionPointer() unsafe.Pointer {
	return s.stateExt
}

// Save saves the plugin state to the stream
func (s *PluginStateExtension) Save(stream *C.clap_ostream_t) bool {
	if s.plugin == nil || stream == nil {
		return false
	}
	
	// Create a new state object
	state := NewState()
	
	// Get the parameter manager from the plugin
	paramManager := s.plugin.GetParamManager()
	if paramManager != nil {
		// Save all parameter values
		for id := range paramManager.params {
			state.ParameterValues[id] = paramManager.GetParamValue(id)
		}
	}
	
	// Save custom state data from the plugin if supported
	if stater, ok := s.plugin.(Stater); ok {
		customData := stater.SaveState()
		if customData != nil {
			state.CustomData = customData
		}
	}
	
	// Write the state to the stream
	err := WriteToStream(state, stream)
	if err != nil {
		fmt.Printf("Error saving state: %v\n", err)
		return false
	}
	
	return true
}

// Load loads the plugin state from the stream
func (s *PluginStateExtension) Load(stream *C.clap_istream_t) bool {
	if s.plugin == nil || stream == nil {
		return false
	}
	
	// Read the state from the stream
	state, err := ReadFromStream(stream)
	if err != nil {
		fmt.Printf("Error loading state: %v\n", err)
		return false
	}
	
	// Get the parameter manager from the plugin
	paramManager := s.plugin.GetParamManager()
	if paramManager != nil {
		// Load all parameter values
		for id, value := range state.ParameterValues {
			paramManager.SetParamValue(id, value)
		}
	}
	
	// Load custom state data into the plugin if supported
	if stater, ok := s.plugin.(Stater); ok {
		stater.LoadState(state.CustomData)
	}
	
	return true
}

// Stater interface for plugins that support custom state data
type Stater interface {
	// SaveState returns custom state data
	SaveState() map[string]interface{}
	
	// LoadState loads custom state data
	LoadState(data map[string]interface{})
}

// External function declarations for the C bridge

//export goStateSave
func goStateSave(plugin unsafe.Pointer, stream *C.clap_ostream_t) C.bool {
	// Retrieve the extension from the plugin data
	ext := (*PluginStateExtension)(plugin)
	if ext == nil {
		return C.bool(false)
	}
	
	return C.bool(ext.Save(stream))
}

//export goStateLoad
func goStateLoad(plugin unsafe.Pointer, stream *C.clap_istream_t) C.bool {
	// Retrieve the extension from the plugin data
	ext := (*PluginStateExtension)(plugin)
	if ext == nil {
		return C.bool(false)
	}
	
	return C.bool(ext.Load(stream))
}

// The implementation of createStateExtension is in cgo.go to avoid duplication