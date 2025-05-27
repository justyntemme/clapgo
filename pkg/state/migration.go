package state

import (
	"fmt"
)

// Migrator performs state migration from one version to another
type Migrator interface {
	// Migrate performs the migration
	Migrate(oldState *State) (*State, error)
	// GetSourceVersion returns the version this migrator migrates from
	GetSourceVersion() Version
	// GetTargetVersion returns the version this migrator migrates to
	GetTargetVersion() Version
}

// MigrationChain manages a chain of state migrators
type MigrationChain struct {
	migrators []Migrator
}

// NewMigrationChain creates a new migration chain
func NewMigrationChain() *MigrationChain {
	return &MigrationChain{
		migrators: make([]Migrator, 0),
	}
}

// AddMigrator adds a migrator to the chain
func (c *MigrationChain) AddMigrator(migrator Migrator) {
	c.migrators = append(c.migrators, migrator)
}

// Migrate runs the migration chain
func (c *MigrationChain) Migrate(state *State, targetVersion Version) (*State, error) {
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

// SimpleMigrator provides a basic migrator implementation
type SimpleMigrator struct {
	sourceVersion Version
	targetVersion Version
	migrateFunc   func(*State) (*State, error)
}

// NewSimpleMigrator creates a new simple migrator
func NewSimpleMigrator(source, target Version, migrate func(*State) (*State, error)) *SimpleMigrator {
	return &SimpleMigrator{
		sourceVersion: source,
		targetVersion: target,
		migrateFunc:   migrate,
	}
}

// Migrate performs the migration
func (m *SimpleMigrator) Migrate(oldState *State) (*State, error) {
	return m.migrateFunc(oldState)
}

// GetSourceVersion returns the source version
func (m *SimpleMigrator) GetSourceVersion() Version {
	return m.sourceVersion
}

// GetTargetVersion returns the target version
func (m *SimpleMigrator) GetTargetVersion() Version {
	return m.targetVersion
}

// Example migrators

// MigrateV1ToV2 is an example migrator from version 1 to 2
func MigrateV1ToV2(oldState *State) (*State, error) {
	// Create new state with updated version
	newState := &State{
		Header: Header{
			Version:    Version2,
			PluginID:   oldState.Header.PluginID,
			PluginName: oldState.Header.PluginName,
			SavedAt:    oldState.Header.SavedAt,
			FormatType: oldState.Header.FormatType,
		},
		Parameters: make([]Parameter, len(oldState.Parameters)),
		CustomData: make(map[string]interface{}),
	}
	
	// Copy parameters
	copy(newState.Parameters, oldState.Parameters)
	
	// Copy custom data
	for k, v := range oldState.CustomData {
		newState.CustomData[k] = v
	}
	
	// Perform version-specific migrations
	// For example, you might rename parameters, change value ranges, etc.
	
	return newState, nil
}