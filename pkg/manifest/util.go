package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LoadFromFile loads a manifest from a file.
func LoadFromFile(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading manifest file: %w", err)
	}
	
	var manifest Manifest
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return nil, fmt.Errorf("error parsing manifest file: %w", err)
	}
	
	return &manifest, nil
}

// FindManifestFiles searches for manifest files in the given directory.
func FindManifestFiles(directory string) ([]string, error) {
	pattern := filepath.Join(directory, "*.json")
	return filepath.Glob(pattern)
}

// GetLibraryPath returns the full path to the Go shared library.
func (m *Manifest) GetLibraryPath(pluginDir string) string {
	return filepath.Join(pluginDir, m.Build.GoSharedLibrary)
}

// Validate checks if a manifest is valid.
func (m *Manifest) Validate() error {
	if m.SchemaVersion == "" {
		return fmt.Errorf("missing schema version")
	}
	
	if m.Plugin.ID == "" {
		return fmt.Errorf("missing plugin ID")
	}
	
	if m.Plugin.Name == "" {
		return fmt.Errorf("missing plugin name")
	}
	
	if m.Plugin.Vendor == "" {
		return fmt.Errorf("missing plugin vendor")
	}
	
	if m.Plugin.Version == "" {
		return fmt.Errorf("missing plugin version")
	}
	
	if m.Build.GoSharedLibrary == "" {
		return fmt.Errorf("missing Go shared library name")
	}
	
	return nil
}