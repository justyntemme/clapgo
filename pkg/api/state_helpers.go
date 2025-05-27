package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// StateVersion represents the version of a plugin's state format
type StateVersion uint32

// Common state versions
const (
	StateVersionUnknown StateVersion = 0
	StateVersion1       StateVersion = 1
	StateVersion2       StateVersion = 2
	StateVersion3       StateVersion = 3
)

// StateHeader contains common metadata for all state formats
type StateHeader struct {
	Version      StateVersion `json:"version"`
	PluginID     string       `json:"plugin_id"`
	PluginName   string       `json:"plugin_name,omitempty"`
	SavedAt      int64        `json:"saved_at,omitempty"` // Unix timestamp
	FormatType   string       `json:"format_type,omitempty"` // "preset", "project", "duplicate"
}

// ParameterState represents a single parameter's state
type ParameterState struct {
	ID    uint32  `json:"id"`
	Value float64 `json:"value"`
	Name  string  `json:"name,omitempty"` // Optional, for readability
}

// PluginState represents the complete state of a plugin
type PluginState struct {
	Header     StateHeader               `json:"header"`
	Parameters []ParameterState          `json:"parameters"`
	CustomData map[string]interface{}    `json:"custom_data,omitempty"`
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

// PresetBank represents a collection of presets
type PresetBank struct {
	Name        string   `json:"name"`
	Author      string   `json:"author,omitempty"`
	Description string   `json:"description,omitempty"`
	Version     string   `json:"version,omitempty"`
	Presets     []Preset `json:"presets"`
}

// StateManager provides helpers for state serialization and deserialization
type StateManager struct {
	pluginID   string
	pluginName string
	version    StateVersion
}

// NewStateManager creates a new state manager
func NewStateManager(pluginID, pluginName string, version StateVersion) *StateManager {
	return &StateManager{
		pluginID:   pluginID,
		pluginName: pluginName,
		version:    version,
	}
}

// SaveParametersToStream saves parameters to a binary stream
func (sm *StateManager) SaveParametersToStream(stream *OutputStream, params *ParameterManager) error {
	// Write version
	if err := stream.WriteUint32(uint32(sm.version)); err != nil {
		return fmt.Errorf("failed to write version: %w", err)
	}
	
	// Write parameter count
	paramCount := params.GetParameterCount()
	if err := stream.WriteUint32(paramCount); err != nil {
		return fmt.Errorf("failed to write parameter count: %w", err)
	}
	
	// Write each parameter
	for i := uint32(0); i < paramCount; i++ {
		info, err := params.GetParameterInfoByIndex(i)
		if err != nil {
			return fmt.Errorf("failed to get parameter info at index %d: %w", i, err)
		}
		
		// Write parameter ID
		if err := stream.WriteUint32(info.ID); err != nil {
			return fmt.Errorf("failed to write parameter ID: %w", err)
		}
		
		// Write parameter value
		value := params.GetParameterValue(info.ID)
		if err := stream.WriteFloat64(value); err != nil {
			return fmt.Errorf("failed to write parameter value: %w", err)
		}
	}
	
	return nil
}

// LoadParametersFromStream loads parameters from a binary stream
func (sm *StateManager) LoadParametersFromStream(stream *InputStream, params *ParameterManager) error {
	// Read version
	version, err := stream.ReadUint32()
	if err != nil {
		return fmt.Errorf("failed to read version: %w", err)
	}
	
	// Check version compatibility
	if StateVersion(version) > sm.version {
		return fmt.Errorf("unsupported state version %d (current: %d)", version, sm.version)
	}
	
	// Read parameter count
	paramCount, err := stream.ReadUint32()
	if err != nil {
		return fmt.Errorf("failed to read parameter count: %w", err)
	}
	
	// Read each parameter
	for i := uint32(0); i < paramCount; i++ {
		// Read parameter ID
		paramID, err := stream.ReadUint32()
		if err != nil {
			return fmt.Errorf("failed to read parameter ID: %w", err)
		}
		
		// Read parameter value
		value, err := stream.ReadFloat64()
		if err != nil {
			return fmt.Errorf("failed to read parameter value: %w", err)
		}
		
		// Set parameter value
		params.SetParameterValue(paramID, value)
	}
	
	return nil
}

// SaveStateToJSON saves plugin state to JSON format
func (sm *StateManager) SaveStateToJSON(params *ParameterManager, customData map[string]interface{}) ([]byte, error) {
	state := PluginState{
		Header: StateHeader{
			Version:    sm.version,
			PluginID:   sm.pluginID,
			PluginName: sm.pluginName,
			SavedAt:    timeNow(),
		},
		Parameters: make([]ParameterState, 0),
		CustomData: customData,
	}
	
	// Add all parameters
	paramCount := params.GetParameterCount()
	for i := uint32(0); i < paramCount; i++ {
		info, err := params.GetParameterInfoByIndex(i)
		if err != nil {
			continue
		}
		
		state.Parameters = append(state.Parameters, ParameterState{
			ID:    info.ID,
			Value: params.GetParameterValue(info.ID),
			Name:  info.Name,
		})
	}
	
	return json.MarshalIndent(state, "", "  ")
}

// LoadStateFromJSON loads plugin state from JSON format
func (sm *StateManager) LoadStateFromJSON(data []byte, params *ParameterManager) (map[string]interface{}, error) {
	var state PluginState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}
	
	// Check version compatibility
	if state.Header.Version > sm.version {
		return nil, fmt.Errorf("unsupported state version %d (current: %d)", state.Header.Version, sm.version)
	}
	
	// Load parameters
	for _, param := range state.Parameters {
		params.SetParameterValue(param.ID, param.Value)
	}
	
	return state.CustomData, nil
}

// SavePreset saves a preset to JSON format
func (sm *StateManager) SavePreset(name, author, description string, params *ParameterManager, customData map[string]interface{}) ([]byte, error) {
	// Create preset data
	presetData := make(map[string]interface{})
	
	// Add parameters
	paramCount := params.GetParameterCount()
	for i := uint32(0); i < paramCount; i++ {
		info, err := params.GetParameterInfoByIndex(i)
		if err != nil {
			continue
		}
		
		// Use parameter name as key for readability
		key := info.Name
		if key == "" {
			key = fmt.Sprintf("param_%d", info.ID)
		}
		presetData[key] = params.GetParameterValue(info.ID)
	}
	
	// Add custom data
	for k, v := range customData {
		presetData[k] = v
	}
	
	preset := Preset{
		Metadata: PresetMetadata{
			Name:        name,
			Author:      author,
			Description: description,
			Version:     sm.pluginName,
		},
		PresetData: presetData,
	}
	
	return json.MarshalIndent(preset, "", "  ")
}

// LoadPreset loads a preset from JSON format
func (sm *StateManager) LoadPreset(data []byte, params *ParameterManager) (*PresetMetadata, map[string]interface{}, error) {
	var preset Preset
	if err := json.Unmarshal(data, &preset); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal preset: %w", err)
	}
	
	// Load parameters by name first, then by ID
	for key, value := range preset.PresetData {
		// Try to find parameter by name
		found := false
		paramCount := params.GetParameterCount()
		for i := uint32(0); i < paramCount; i++ {
			info, err := params.GetParameterInfoByIndex(i)
			if err != nil {
				continue
			}
			
			if info.Name == key {
				if floatVal, ok := value.(float64); ok {
					params.SetParameterValue(info.ID, floatVal)
					found = true
					break
				}
			}
		}
		
		// If not found by name, try to extract ID from key
		if !found {
			var paramID uint32
			if _, err := fmt.Sscanf(key, "param_%d", &paramID); err == nil {
				if floatVal, ok := value.(float64); ok {
					params.SetParameterValue(paramID, floatVal)
				}
			}
		}
	}
	
	return &preset.Metadata, preset.PresetData, nil
}

// LoadPresetFromFile loads a preset from a JSON file
func (sm *StateManager) LoadPresetFromFile(path string, params *ParameterManager) (*PresetMetadata, map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read preset file: %w", err)
	}
	
	return sm.LoadPreset(data, params)
}

// SavePresetToFile saves a preset to a JSON file
func (sm *StateManager) SavePresetToFile(path, name, author, description string, params *ParameterManager, customData map[string]interface{}) error {
	data, err := sm.SavePreset(name, author, description, params, customData)
	if err != nil {
		return fmt.Errorf("failed to create preset: %w", err)
	}
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write preset file: %w", err)
	}
	
	return nil
}

// MigrateState migrates state from an older version to the current version
type StateMigrator interface {
	// Migrate performs the migration
	Migrate(oldState *PluginState) (*PluginState, error)
	// GetSourceVersion returns the version this migrator migrates from
	GetSourceVersion() StateVersion
	// GetTargetVersion returns the version this migrator migrates to
	GetTargetVersion() StateVersion
}

// StateMigrationChain manages a chain of state migrators
type StateMigrationChain struct {
	migrators []StateMigrator
}

// NewStateMigrationChain creates a new migration chain
func NewStateMigrationChain() *StateMigrationChain {
	return &StateMigrationChain{
		migrators: make([]StateMigrator, 0),
	}
}

// AddMigrator adds a migrator to the chain
func (c *StateMigrationChain) AddMigrator(migrator StateMigrator) {
	c.migrators = append(c.migrators, migrator)
}

// Migrate runs the migration chain
func (c *StateMigrationChain) Migrate(state *PluginState, targetVersion StateVersion) (*PluginState, error) {
	currentState := state
	
	// Keep migrating until we reach the target version
	for currentState.Header.Version < targetVersion {
		migrated := false
		
		// Find a migrator for the current version
		for _, migrator := range c.migrators {
			if migrator.GetSourceVersion() == currentState.Header.Version {
				newState, err := migrator.Migrate(currentState)
				if err != nil {
					return nil, fmt.Errorf("migration from v%d to v%d failed: %w", 
						migrator.GetSourceVersion(), migrator.GetTargetVersion(), err)
				}
				currentState = newState
				migrated = true
				break
			}
		}
		
		if !migrated {
			return nil, fmt.Errorf("no migrator found for version %d", currentState.Header.Version)
		}
	}
	
	return currentState, nil
}

// timeNow is a variable to allow mocking in tests
var timeNow = func() int64 {
	return 0 // Return epoch time for deterministic behavior
}

// ValidateState validates a plugin state
func ValidateState(state *PluginState, expectedPluginID string, maxVersion StateVersion) error {
	// Check header
	if state.Header.PluginID != expectedPluginID {
		return fmt.Errorf("plugin ID mismatch: expected %s, got %s", expectedPluginID, state.Header.PluginID)
	}
	
	if state.Header.Version > maxVersion {
		return fmt.Errorf("unsupported version %d (max: %d)", state.Header.Version, maxVersion)
	}
	
	// Check for duplicate parameter IDs
	seen := make(map[uint32]bool)
	for _, param := range state.Parameters {
		if seen[param.ID] {
			return fmt.Errorf("duplicate parameter ID: %d", param.ID)
		}
		seen[param.ID] = true
	}
	
	return nil
}

// ValidatePreset validates a preset
func ValidatePreset(preset *Preset) error {
	// Check metadata
	if preset.Metadata.Name == "" {
		return fmt.Errorf("preset name is required")
	}
	
	// Check preset data
	if len(preset.PresetData) == 0 {
		return fmt.Errorf("preset data is empty")
	}
	
	return nil
}