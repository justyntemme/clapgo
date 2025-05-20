// Package manifest provides types and utilities for working with
// plugin manifest files, which describe CLAP plugins and their metadata.
package manifest

// Manifest represents the complete structure of a plugin manifest file.
type Manifest struct {
	SchemaVersion string       `json:"schemaVersion"`
	Plugin        PluginInfo   `json:"plugin"`
	Build         BuildInfo    `json:"build"`
	Extensions    []Extension  `json:"extensions,omitempty"`
	Parameters    []Parameter  `json:"parameters,omitempty"`
}

// PluginInfo contains the core metadata about a plugin.
type PluginInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Vendor      string   `json:"vendor"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	URL         string   `json:"url,omitempty"`
	ManualURL   string   `json:"manualUrl,omitempty"`
	SupportURL  string   `json:"supportUrl,omitempty"`
	Features    []string `json:"features,omitempty"`
}

// BuildInfo contains information related to building and loading the plugin.
type BuildInfo struct {
	GoSharedLibrary string   `json:"goSharedLibrary"`
	EntryPoint      string   `json:"entryPoint,omitempty"`
	Dependencies    []string `json:"dependencies,omitempty"`
}

// Extension represents a CLAP extension supported by the plugin.
type Extension struct {
	ID        string `json:"id"`
	Supported bool   `json:"supported"`
}

// Parameter describes a plugin parameter.
type Parameter struct {
	ID           uint32    `json:"id"`
	Name         string    `json:"name"`
	MinValue     float64   `json:"minValue"`
	MaxValue     float64   `json:"maxValue"`
	DefaultValue float64   `json:"defaultValue"`
	Flags        []string  `json:"flags,omitempty"`
}