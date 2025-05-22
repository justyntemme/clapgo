// The generate-manifest command creates manifest files for CLAP plugins
// based on Go source code analysis or command-line input.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/justyntemme/clapgo/pkg/manifest"
)

var (
	pluginID          = flag.String("id", "", "Plugin ID (e.g., com.clapgo.gain)")
	pluginName        = flag.String("name", "", "Plugin name (e.g., Simple Gain)")
	pluginVendor      = flag.String("vendor", "ClapGo", "Plugin vendor")
	pluginVersion     = flag.String("version", "1.0.0", "Plugin version")
	pluginDescription = flag.String("description", "", "Plugin description")
	pluginURL         = flag.String("url", "https://github.com/justyntemme/clapgo", "Plugin URL")
	pluginManualURL   = flag.String("manual-url", "https://github.com/justyntemme/clapgo", "Plugin manual URL")
	pluginSupportURL  = flag.String("support-url", "https://github.com/justyntemme/clapgo/issues", "Plugin support URL")
	pluginFeatures    = flag.String("features", "audio-effect,stereo,mono", "Comma-separated list of plugin features")
	
	outputDir         = flag.String("output-dir", "", "Output directory for manifest files")
	libName           = flag.String("lib-name", "", "Shared library name (without lib prefix or extension)")
	
	// Platform detection
	platform          = flag.String("platform", detectPlatform(), "Platform (linux, macos, or windows)")
)

func detectPlatform() string {
	goos := strings.ToLower(os.Getenv("GOOS"))
	if goos == "" {
		// Default to current OS
		if _, err := os.Stat("/proc"); err == nil {
			return "linux"
		} else if _, err := os.Stat("/Applications"); err == nil {
			return "macos"
		} else if _, err := os.Stat("C:\\Windows"); err == nil {
			return "windows"
		}
		return "linux" // Default fallback
	}
	
	if goos == "darwin" {
		return "macos"
	}
	return goos
}

func getSharedLibExtension() string {
	switch *platform {
	case "windows":
		return "dll"
	case "macos":
		return "dylib"
	default: // linux or others
		return "so"
	}
}

func main() {
	flag.Parse()
	
	if *pluginID == "" || *pluginName == "" {
		fmt.Println("Error: Plugin ID and name are required")
		flag.Usage()
		os.Exit(1)
	}
	
	if *libName == "" {
		// Default to plugin ID's last component
		parts := strings.Split(*pluginID, ".")
		*libName = parts[len(parts)-1]
	}
	
	// Create manifest
	manifest := manifest.Manifest{
		SchemaVersion: "1.0",
		Plugin: manifest.PluginInfo{
			ID:          *pluginID,
			Name:        *pluginName,
			Vendor:      *pluginVendor,
			Version:     *pluginVersion,
			Description: *pluginDescription,
			URL:         *pluginURL,
			ManualURL:   *pluginManualURL,
			SupportURL:  *pluginSupportURL,
			Features:    strings.Split(*pluginFeatures, ","),
		},
		Build: manifest.BuildInfo{
			GoSharedLibrary: fmt.Sprintf("lib%s.%s", *libName, getSharedLibExtension()),
			EntryPoint:      "CreatePlugin",
			Dependencies:    []string{},
		},
		Extensions: []manifest.Extension{
			{
				ID:        "clap.audio-ports",
				Supported: true,
			},
			{
				ID:        "clap.state",
				Supported: false,
			},
		},
	}
	
	// Marshal to JSON
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling manifest: %v\n", err)
		os.Exit(1)
	}
	
	// Determine output path
	var outputPath string
	if *outputDir == "" {
		outputPath = fmt.Sprintf("%s.json", *libName)
	} else {
		err := os.MkdirAll(*outputDir, 0755)
		if err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			os.Exit(1)
		}
		outputPath = filepath.Join(*outputDir, fmt.Sprintf("%s.json", *libName))
	}
	
	// Write manifest file
	err = os.WriteFile(outputPath, data, 0644)
	if err != nil {
		fmt.Printf("Error writing manifest file: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Manifest file written to %s\n", outputPath)
}