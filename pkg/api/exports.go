// Package api defines the core interfaces for CLAP plugins in Go.
package api

import "C"

// PluginExporter defines the interface that all plugins must implement to export
// their information to the bridge. This replaces the need for environment variables
// and stub registrations.
type PluginExporter interface {
	// GetPluginID returns the unique identifier for this plugin
	GetPluginID() string
	
	// GetPluginName returns the human-readable name of the plugin
	GetPluginName() string
	
	// GetPluginVendor returns the vendor name of the plugin
	GetPluginVendor() string
	
	// GetPluginVersion returns the version string of the plugin
	GetPluginVersion() string
	
	// GetPluginDescription returns a short description of the plugin
	GetPluginDescription() string
}

// PluginMetadata provides a simple struct to hold plugin metadata
// Plugin developers should initialize this with their plugin's constants
type PluginMetadata struct {
	ID          string
	Name        string
	Vendor      string
	Version     string
	Description string
}

// Standard plugin metadata functions that can be called directly by plugins
// to implement the PluginExporter interface
var (
	// Current plugin metadata
	CurrentPluginMetadata PluginMetadata
)

// SetPluginMetadata sets the current plugin metadata 
// This should be called by each plugin during initialization
func SetPluginMetadata(metadata PluginMetadata) {
	CurrentPluginMetadata = metadata
}

// Below are the exported functions that bridge from Go to C
// Plugin developers should NOT need to implement or call these directly

//export ExportPluginID
func ExportPluginID() *C.char {
	return C.CString(CurrentPluginMetadata.ID)
}

//export ExportPluginName
func ExportPluginName() *C.char {
	return C.CString(CurrentPluginMetadata.Name)
}

//export ExportPluginVendor
func ExportPluginVendor() *C.char {
	return C.CString(CurrentPluginMetadata.Vendor)
}

//export ExportPluginVersion
func ExportPluginVersion() *C.char {
	return C.CString(CurrentPluginMetadata.Version)
}

//export ExportPluginDescription
func ExportPluginDescription() *C.char {
	return C.CString(CurrentPluginMetadata.Description)
}

// RegisterMetadataFromConstants is a convenience function that registers plugin metadata
// using common constant names. This allows plugins to simply call this function with no arguments
// if they follow the standard naming convention of PluginID, PluginName, etc.
func RegisterMetadataFromConstants(id, name, vendor, version, description string) {
	SetPluginMetadata(PluginMetadata{
		ID:          id,
		Name:        name,
		Vendor:      vendor,
		Version:     version,
		Description: description,
	})
}