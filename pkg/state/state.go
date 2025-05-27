package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Version represents the version of a plugin's state format
type Version uint32

// Common state versions
const (
	VersionUnknown Version = 0
	Version1       Version = 1
	Version2       Version = 2
	Version3       Version = 3
)

// Common errors
var (
	ErrInvalidVersion     = errors.New("invalid state version")
	ErrVersionTooNew      = errors.New("state version too new")
	ErrInvalidPluginID    = errors.New("invalid plugin ID")
	ErrEmptyPresetName    = errors.New("preset name is required")
	ErrEmptyPresetData    = errors.New("preset data is empty")
	ErrDuplicateParameter = errors.New("duplicate parameter ID")
)

// Header contains common metadata for all state formats
type Header struct {
	Version    Version `json:"version"`
	PluginID   string  `json:"plugin_id"`
	PluginName string  `json:"plugin_name,omitempty"`
	SavedAt    int64   `json:"saved_at,omitempty"` // Unix timestamp
	FormatType string  `json:"format_type,omitempty"` // "preset", "project", "duplicate"
}

// Parameter represents a single parameter's state
type Parameter struct {
	ID    uint32  `json:"id"`
	Value float64 `json:"value"`
	Name  string  `json:"name,omitempty"` // Optional, for readability
}

// State represents the complete state of a plugin
type State struct {
	Header     Header                 `json:"header"`
	Parameters []Parameter            `json:"parameters"`
	CustomData map[string]interface{} `json:"custom_data,omitempty"`
}

// PresetMetadata contains metadata about a preset
type PresetMetadata struct {
	Name        string   `json:"name"`
	Author      string   `json:"author,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Version     string   `json:"version,omitempty"`
}

// Preset represents a complete preset structure
type Preset struct {
	Metadata   PresetMetadata         `json:"metadata"`
	PresetData map[string]interface{} `json:"preset_data"`
}

// Manager provides state serialization and deserialization
type Manager struct {
	pluginID   string
	pluginName string
	version    Version
}

// NewManager creates a new state manager
func NewManager(pluginID, pluginName string, version Version) *Manager {
	return &Manager{
		pluginID:   pluginID,
		pluginName: pluginName,
		version:    version,
	}
}

// CreateState creates a new state with the given parameters
func (m *Manager) CreateState(params []Parameter, customData map[string]interface{}) *State {
	return &State{
		Header: Header{
			Version:    m.version,
			PluginID:   m.pluginID,
			PluginName: m.pluginName,
			SavedAt:    time.Now().Unix(),
		},
		Parameters: params,
		CustomData: customData,
	}
}

// ValidateState validates a plugin state
func (m *Manager) ValidateState(state *State) error {
	// Check header
	if state.Header.PluginID != m.pluginID {
		return fmt.Errorf("%w: expected %s, got %s", ErrInvalidPluginID, m.pluginID, state.Header.PluginID)
	}
	
	if state.Header.Version > m.version {
		return fmt.Errorf("%w: %d (max: %d)", ErrVersionTooNew, state.Header.Version, m.version)
	}
	
	// Check for duplicate parameter IDs
	seen := make(map[uint32]bool)
	for _, param := range state.Parameters {
		if seen[param.ID] {
			return fmt.Errorf("%w: %d", ErrDuplicateParameter, param.ID)
		}
		seen[param.ID] = true
	}
	
	return nil
}

// SaveToJSON saves state to JSON format
func (m *Manager) SaveToJSON(state *State) ([]byte, error) {
	if err := m.ValidateState(state); err != nil {
		return nil, err
	}
	return json.MarshalIndent(state, "", "  ")
}

// LoadFromJSON loads state from JSON format
func (m *Manager) LoadFromJSON(data []byte) (*State, error) {
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}
	
	if err := m.ValidateState(&state); err != nil {
		return nil, err
	}
	
	return &state, nil
}

// CreatePreset creates a new preset
func (m *Manager) CreatePreset(name, author, description string, params []Parameter, customData map[string]interface{}) *Preset {
	// Create preset data map
	presetData := make(map[string]interface{})
	
	// Add parameters
	for _, param := range params {
		// Use parameter name as key for readability
		key := param.Name
		if key == "" {
			key = fmt.Sprintf("param_%d", param.ID)
		}
		presetData[key] = param.Value
	}
	
	// Add custom data
	for k, v := range customData {
		presetData[k] = v
	}
	
	return &Preset{
		Metadata: PresetMetadata{
			Name:        name,
			Author:      author,
			Description: description,
			Version:     m.pluginName,
		},
		PresetData: presetData,
	}
}

// ValidatePreset validates a preset
func (m *Manager) ValidatePreset(preset *Preset) error {
	// Check metadata
	if preset.Metadata.Name == "" {
		return ErrEmptyPresetName
	}
	
	// Check preset data
	if len(preset.PresetData) == 0 {
		return ErrEmptyPresetData
	}
	
	return nil
}

// SavePresetToJSON saves a preset to JSON format
func (m *Manager) SavePresetToJSON(preset *Preset) ([]byte, error) {
	if err := m.ValidatePreset(preset); err != nil {
		return nil, err
	}
	return json.MarshalIndent(preset, "", "  ")
}

// LoadPresetFromJSON loads a preset from JSON format
func (m *Manager) LoadPresetFromJSON(data []byte) (*Preset, error) {
	var preset Preset
	if err := json.Unmarshal(data, &preset); err != nil {
		return nil, fmt.Errorf("failed to unmarshal preset: %w", err)
	}
	
	if err := m.ValidatePreset(&preset); err != nil {
		return nil, err
	}
	
	return &preset, nil
}