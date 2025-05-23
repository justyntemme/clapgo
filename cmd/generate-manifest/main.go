// The generate-manifest command creates manifest files for CLAP plugins
// based on Go source code analysis or command-line input.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

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
	pluginFeatures    = flag.String("features", "", "Comma-separated list of plugin features")
	
	outputDir         = flag.String("output-dir", "", "Output directory for manifest files")
	libName           = flag.String("lib-name", "", "Shared library name (without lib prefix or extension)")
	
	// Plugin type flags
	pluginType        = flag.String("type", "audio-effect", "Plugin type: audio-effect, instrument, note-effect, analyzer, note-detector, utility")
	pluginSubtype     = flag.String("subtype", "", "Plugin subtype: reverb, compressor, synthesizer, etc.")
	
	// Code generation flags
	generateCode      = flag.Bool("generate", false, "Generate plugin code files")
	
	// Interactive mode
	interactive       = flag.Bool("interactive", false, "Interactive plugin creation wizard")
	
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

func getPluginFeatures(pluginType, subtype string) []string {
	features := []string{}
	
	switch pluginType {
	case "audio-effect":
		features = append(features, "audio-effect")
		switch subtype {
		case "reverb", "delay":
			features = append(features, "delay")
		case "compressor", "limiter", "gate":
			features = append(features, "dynamics")
		case "eq", "filter":
			features = append(features, "filter")
		case "distortion", "saturation":
			features = append(features, "distortion")
		case "chorus", "phaser", "flanger":
			features = append(features, "modulation")
		}
	case "instrument":
		features = append(features, "instrument")
		switch subtype {
		case "synthesizer":
			features = append(features, "synthesizer")
		case "sampler":
			features = append(features, "sampler")
		case "drum", "drum-machine":
			features = append(features, "drum", "drum-machine")
		}
	case "note-effect":
		features = append(features, "note-effect")
	case "analyzer":
		features = append(features, "analyzer")
	case "note-detector":
		features = append(features, "note-detector")
	case "utility":
		features = append(features, "audio-effect", "utility")
	}
	
	// Add common features
	features = append(features, "stereo", "mono")
	
	return features
}

func runInteractiveWizard() {
	scanner := bufio.NewScanner(os.Stdin)
	
	fmt.Println("🎵 ClapGo Plugin Creation Wizard")
	fmt.Println("=================================")
	fmt.Println()
	
	// Plugin type selection
	fmt.Println("1. What type of plugin do you want to create?")
	fmt.Println("   1) Audio Effect (reverb, delay, compressor, etc.)")
	fmt.Println("   2) Instrument (synthesizer, sampler, etc.)")
	fmt.Println("   3) Note Effect (arpeggiator, chord generator, etc.)")
	fmt.Println("   4) Analyzer (spectrum analyzer, level meter, etc.)")
	fmt.Println("   5) Note Detector (audio-to-MIDI converter, etc.)")
	fmt.Println("   6) Utility (gain, test tone, routing, etc.)")
	fmt.Print("Enter choice (1-6): ")
	
	var choice int
	fmt.Scanln(&choice)
	
	typeMap := map[int]string{
		1: "audio-effect",
		2: "instrument", 
		3: "note-effect",
		4: "analyzer",
		5: "note-detector",
		6: "utility",
	}
	
	if selectedType, ok := typeMap[choice]; ok {
		*pluginType = selectedType
	} else {
		fmt.Println("Invalid choice, defaulting to audio-effect")
		*pluginType = "audio-effect"
	}
	
	// Subtype selection based on main type
	fmt.Println()
	fmt.Println("2. What specific type of", *pluginType, "is this?")
	
	subtypeOptions := getSubtypeOptions(*pluginType)
	for i, option := range subtypeOptions {
		fmt.Printf("   %d) %s\n", i+1, option)
	}
	fmt.Print("Enter choice (or press Enter to skip): ")
	
	scanner.Scan()
	subtypeChoice := strings.TrimSpace(scanner.Text())
	if subtypeChoice != "" {
		if idx := parseInt(subtypeChoice); idx > 0 && idx <= len(subtypeOptions) {
			*pluginSubtype = subtypeOptions[idx-1]
		}
	}
	
	// Plugin metadata
	fmt.Println()
	fmt.Println("3. Plugin Information")
	
	fmt.Print("Plugin name (e.g., 'My Awesome Reverb'): ")
	scanner.Scan()
	name := strings.TrimSpace(scanner.Text())
	if name != "" {
		*pluginName = name
	}
	
	fmt.Print("Vendor/Company name (default: ClapGo): ")
	scanner.Scan()
	vendor := strings.TrimSpace(scanner.Text())
	if vendor != "" {
		*pluginVendor = vendor
	}
	
	// Generate plugin ID suggestion
	if *pluginName != "" && *pluginVendor != "" {
		suggested := generatePluginID(*pluginVendor, *pluginName)
		fmt.Printf("Plugin ID (suggested: %s): ", suggested)
		scanner.Scan()
		id := strings.TrimSpace(scanner.Text())
		if id != "" {
			*pluginID = id
		} else {
			*pluginID = suggested
		}
	}
	
	fmt.Print("Short description (optional): ")
	scanner.Scan()
	desc := strings.TrimSpace(scanner.Text())
	if desc != "" {
		*pluginDescription = desc
	}
	
	// Output directory
	fmt.Println()
	fmt.Println("4. Output Configuration")
	
	if *pluginName != "" {
		pluginDirName := strings.ToLower(strings.ReplaceAll(*pluginName, " ", "-"))
		suggested := filepath.Join("plugins", pluginDirName)
		fmt.Printf("Output directory (suggested: %s): ", suggested)
		scanner.Scan()
		dir := strings.TrimSpace(scanner.Text())
		if dir != "" {
			*outputDir = dir
		} else {
			*outputDir = suggested
		}
	}
	
	// Summary
	fmt.Println()
	fmt.Println("📋 Summary")
	fmt.Println("==========")
	fmt.Printf("Type: %s", *pluginType)
	if *pluginSubtype != "" {
		fmt.Printf(" (%s)", *pluginSubtype)
	}
	fmt.Println()
	fmt.Printf("Name: %s\n", *pluginName)
	fmt.Printf("Vendor: %s\n", *pluginVendor)
	fmt.Printf("ID: %s\n", *pluginID)
	if *pluginDescription != "" {
		fmt.Printf("Description: %s\n", *pluginDescription)
	}
	fmt.Printf("Output: %s\n", *outputDir)
	fmt.Println()
	
	fmt.Print("Create this plugin? (Y/n): ")
	scanner.Scan()
	confirm := strings.ToLower(strings.TrimSpace(scanner.Text()))
	if confirm == "n" || confirm == "no" {
		fmt.Println("Cancelled.")
		os.Exit(0)
	}
	
	// Enable code generation
	*generateCode = true
	
	fmt.Println()
	fmt.Println("🚀 Creating plugin...")
}

func getSubtypeOptions(pluginType string) []string {
	switch pluginType {
	case "audio-effect":
		return []string{"reverb", "delay", "compressor", "limiter", "gate", "eq", "filter", "distortion", "saturation", "chorus", "phaser", "flanger", "tremolo", "vibrato"}
	case "instrument":
		return []string{"synthesizer", "sampler", "drum", "drum-machine", "piano", "organ"}
	case "note-effect":
		return []string{"arpeggiator", "chord-generator", "sequencer", "harmony"}
	case "analyzer":
		return []string{"spectrum-analyzer", "level-meter", "oscilloscope", "phase-meter", "correlation-meter"}
	case "note-detector":
		return []string{"pitch-detector", "onset-detector", "beat-detector"}
	case "utility":
		return []string{"gain", "test-tone", "routing", "mixer", "splitter"}
	default:
		return []string{}
	}
}

func generatePluginID(vendor, name string) string {
	// Convert to lowercase and replace spaces/special chars
	vendorClean := strings.ToLower(strings.ReplaceAll(vendor, " ", ""))
	nameClean := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	nameClean = strings.ReplaceAll(nameClean, "_", "-")
	
	return fmt.Sprintf("com.%s.%s", vendorClean, nameClean)
}

func parseInt(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}

func getPluginExtensions(pluginType string) []manifest.Extension {
	extensions := []manifest.Extension{
		// Core extensions for all plugins
		{ID: "clap.params", Supported: true},
		{ID: "clap.state", Supported: true},
		{ID: "clap.log", Supported: true},
	}
	
	switch pluginType {
	case "audio-effect", "utility":
		extensions = append(extensions,
			manifest.Extension{ID: "clap.audio-ports", Supported: true},
			manifest.Extension{ID: "clap.latency", Supported: true},
			manifest.Extension{ID: "clap.tail", Supported: true},
			manifest.Extension{ID: "clap.render", Supported: true},
		)
	case "instrument":
		extensions = append(extensions,
			manifest.Extension{ID: "clap.audio-ports", Supported: true},
			manifest.Extension{ID: "clap.note-ports", Supported: true},
			manifest.Extension{ID: "clap.voice-info", Supported: true},
			manifest.Extension{ID: "clap.tuning", Supported: false},
			manifest.Extension{ID: "clap.note-name", Supported: false},
		)
	case "note-effect":
		extensions = append(extensions,
			manifest.Extension{ID: "clap.note-ports", Supported: true},
			manifest.Extension{ID: "clap.transport-control", Supported: false},
			manifest.Extension{ID: "clap.event-registry", Supported: false},
		)
	case "analyzer":
		extensions = append(extensions,
			manifest.Extension{ID: "clap.audio-ports", Supported: true},
			manifest.Extension{ID: "clap.thread-check", Supported: true},
			manifest.Extension{ID: "clap.gain-adjustment-metering", Supported: false},
			manifest.Extension{ID: "clap.mini-curve-display", Supported: false},
		)
	case "note-detector":
		extensions = append(extensions,
			manifest.Extension{ID: "clap.audio-ports", Supported: true},
			manifest.Extension{ID: "clap.note-ports", Supported: true},
		)
	}
	
	return extensions
}

func main() {
	flag.Parse()
	
	// Run interactive wizard if requested
	if *interactive {
		runInteractiveWizard()
	}
	
	if *pluginID == "" || *pluginName == "" {
		fmt.Println("Error: Plugin ID and name are required")
		fmt.Println("Use -interactive for the wizard, or provide -id and -name flags")
		flag.Usage()
		os.Exit(1)
	}
	
	if *libName == "" {
		// Default to plugin ID's last component
		parts := strings.Split(*pluginID, ".")
		*libName = parts[len(parts)-1]
	}
	
	// Determine features based on plugin type or use provided features
	var features []string
	if *pluginFeatures != "" {
		features = strings.Split(*pluginFeatures, ",")
	} else {
		features = getPluginFeatures(*pluginType, *pluginSubtype)
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
			Features:    features,
		},
		Build: manifest.BuildInfo{
			GoSharedLibrary: fmt.Sprintf("lib%s.%s", *libName, getSharedLibExtension()),
			EntryPoint:      "CreatePlugin",
			Dependencies:    []string{},
		},
		Extensions: getPluginExtensions(*pluginType),
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
		// Default to plugins directory when no output-dir specified
		pluginDirName := strings.ToLower(strings.ReplaceAll(*pluginName, " ", "-"))
		*outputDir = filepath.Join("plugins", pluginDirName)
	}
	
	// Create output directory
	err = os.MkdirAll(*outputDir, 0755)
	if err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}
	outputPath = filepath.Join(*outputDir, fmt.Sprintf("%s.json", *libName))
	
	// Write manifest file
	err = os.WriteFile(outputPath, data, 0644)
	if err != nil {
		fmt.Printf("Error writing manifest file: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Manifest file written to %s\n", outputPath)
	
	// Generate code files if requested
	if *generateCode && *outputDir != "" {
		err = generatePluginCode(*outputDir, manifest, *pluginType, *libName)
		if err != nil {
			fmt.Printf("Error generating plugin code: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Plugin code generated in %s\n", *outputDir)
	}
}

func generatePluginCode(dir string, m manifest.Manifest, pluginType, libName string) error {
	// Create template data
	structName := toCamelCase(libName)
	pluginVar := strings.ToLower(structName[:1]) + structName[1:] + "Plugin"
	data := TemplateData{
		PluginID:     m.Plugin.ID,
		PluginName:   m.Plugin.Name,
		PluginVendor: m.Plugin.Vendor,
		PluginType:   pluginType,
		LibName:      libName,
		StructName:   structName,
		PluginVarName: pluginVar,
		PluginVar:    pluginVar,
		Extensions:   m.Extensions,
		Features:     m.Plugin.Features,
		PluginDescription: m.Plugin.Description,
		PluginVersion: m.Plugin.Version,
		PluginURL:    m.Plugin.URL,
	}
	
	// Generate exports_generated.go
	err := generateFile(filepath.Join(dir, "exports_generated.go"), exportsTemplate, data)
	if err != nil {
		return fmt.Errorf("generating exports_generated.go: %w", err)
	}
	
	// Generate extensions_generated.go
	err = generateFile(filepath.Join(dir, "extensions_generated.go"), extensionsTemplate, data)
	if err != nil {
		return fmt.Errorf("generating extensions_generated.go: %w", err)
	}
	
	// Generate constants_generated.go
	err = generateFile(filepath.Join(dir, "constants_generated.go"), constantsTemplate, data)
	if err != nil {
		return fmt.Errorf("generating constants_generated.go: %w", err)
	}
	
	// Generate plugin.go if it doesn't exist
	pluginPath := filepath.Join(dir, "plugin.go")
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		err = generateFile(pluginPath, pluginTemplate, data)
		if err != nil {
			return fmt.Errorf("generating plugin.go: %w", err)
		}
	}
	
	// Generate preset scaffolding
	err = generatePresetScaffolding(dir, data)
	if err != nil {
		return fmt.Errorf("generating presets: %w", err)
	}
	
	return nil
}

type TemplateData struct {
	PluginID     string
	PluginName   string
	PluginVendor string
	PluginType   string
	LibName      string
	StructName   string
	PluginVarName string
	PluginVar    string  // Variable name for global plugin instance
	Extensions   []manifest.Extension
	Features     []string
	PluginDescription string
	PluginVersion string
	PluginURL    string
}

func generatePresetScaffolding(dir string, data TemplateData) error {
	// Create preset directories
	presetDir := filepath.Join(dir, "presets")
	factoryDir := filepath.Join(presetDir, "factory")
	
	err := os.MkdirAll(factoryDir, 0755)
	if err != nil {
		return fmt.Errorf("creating preset directories: %w", err)
	}
	
	// Generate example presets based on plugin type
	presets := getExamplePresets(data.PluginType, data.PluginVendor, data.PluginName)
	
	for filename, content := range presets {
		presetPath := filepath.Join(factoryDir, filename)
		err = os.WriteFile(presetPath, []byte(content), 0644)
		if err != nil {
			return fmt.Errorf("writing preset %s: %w", filename, err)
		}
	}
	
	return nil
}

func getExamplePresets(pluginType, vendor, name string) map[string]string {
	presets := make(map[string]string)
	
	switch pluginType {
	case "audio-effect":
		presets["default.json"] = fmt.Sprintf(`{
  "name": "Default",
  "vendor": "%s",
  "plugin": "%s",
  "version": "1.0.0",
  "parameters": {
    "volume": 0.8,
    "mix": 1.0
  },
  "description": "Default settings for %s"
}`, vendor, name, name)
		
		presets["subtle.json"] = fmt.Sprintf(`{
  "name": "Subtle",
  "vendor": "%s",
  "plugin": "%s",
  "version": "1.0.0",
  "parameters": {
    "volume": 0.6,
    "mix": 0.3
  },
  "description": "Subtle processing"
}`, vendor, name)
		
	case "instrument":
		presets["init.json"] = fmt.Sprintf(`{
  "name": "Init",
  "vendor": "%s",
  "plugin": "%s",
  "version": "1.0.0",
  "parameters": {
    "volume": 0.7,
    "attack": 0.01,
    "decay": 0.1,
    "sustain": 0.7,
    "release": 0.3,
    "waveform": 0
  },
  "description": "Basic initialization preset"
}`, vendor, name)
		
		presets["lead.json"] = fmt.Sprintf(`{
  "name": "Lead",
  "vendor": "%s",
  "plugin": "%s", 
  "version": "1.0.0",
  "parameters": {
    "volume": 0.8,
    "attack": 0.005,
    "decay": 0.05,
    "sustain": 0.8,
    "release": 0.2,
    "waveform": 1
  },
  "description": "Lead synthesizer sound"
}`, vendor, name)
		
		presets["pad.json"] = fmt.Sprintf(`{
  "name": "Pad",
  "vendor": "%s",
  "plugin": "%s",
  "version": "1.0.0", 
  "parameters": {
    "volume": 0.6,
    "attack": 0.5,
    "decay": 0.3,
    "sustain": 0.9,
    "release": 1.0,
    "waveform": 0
  },
  "description": "Soft pad sound"
}`, vendor, name)
		
	case "utility":
		presets["bypass.json"] = fmt.Sprintf(`{
  "name": "Bypass",
  "vendor": "%s",
  "plugin": "%s",
  "version": "1.0.0",
  "parameters": {
    "volume": 1.0,
    "enabled": false
  },
  "description": "Bypass mode"
}`, vendor, name)
		
		presets["unity.json"] = fmt.Sprintf(`{
  "name": "Unity",
  "vendor": "%s",
  "plugin": "%s",
  "version": "1.0.0",
  "parameters": {
    "volume": 1.0,
    "enabled": true
  },
  "description": "Unity gain"
}`, vendor, name)
		
	default:
		// Generic preset for other types
		presets["default.json"] = fmt.Sprintf(`{
  "name": "Default",
  "vendor": "%s",
  "plugin": "%s",
  "version": "1.0.0",
  "parameters": {},
  "description": "Default preset for %s"
}`, vendor, name, name)
	}
	
	return presets
}

func generateFile(path, tmplStr string, data interface{}) error {
	tmpl, err := template.New("file").Parse(tmplStr)
	if err != nil {
		return err
	}
	
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	
	return tmpl.Execute(file, data)
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

const exportsTemplate = `// Code generated by clapgo-generate. DO NOT EDIT.
// This file contains all CGO exports required by the CLAP plugin API.
// All plugin lifecycle and extension callbacks are handled here.
package main

// #cgo CFLAGS: -I../../include/clap/include
// #include "../../include/clap/include/clap/clap.h"
// #include <stdlib.h>
//
// // Helper functions for CLAP event handling
// static inline uint32_t clap_input_events_size_helper(const clap_input_events_t* events) {
//     if (events && events->size) {
//         return events->size(events);
//     }
//     return 0;
// }
//
// static inline const clap_event_header_t* clap_input_events_get_helper(const clap_input_events_t* events, uint32_t index) {
//     if (events && events->get) {
//         return events->get(events, index);
//     }
//     return NULL;
// }
import "C"
import (
	"fmt"
	"runtime/cgo"
	"unsafe"
	
	"github.com/justyntemme/clapgo/pkg/api"
)

// Global plugin instance
var {{.PluginVarName}} *{{.StructName}}

func init() {
	fmt.Println("Initializing {{.PluginName}} plugin")
	{{.PluginVarName}} = New{{.StructName}}()
	fmt.Printf("{{.PluginName}} plugin initialized: %s (%s)\n", {{.PluginVarName}}.GetPluginInfo().Name, {{.PluginVarName}}.GetPluginInfo().ID)
}

// Standardized export functions for manifest system

//export ClapGo_CreatePlugin
func ClapGo_CreatePlugin(host unsafe.Pointer, pluginID *C.char) unsafe.Pointer {
	id := C.GoString(pluginID)
	fmt.Printf("{{.PluginName}} plugin - ClapGo_CreatePlugin with ID: %s\n", id)
	
	if id == PluginID {
		// Create a CGO handle to safely pass the Go object to C
		handle := cgo.NewHandle({{.PluginVarName}})
		fmt.Printf("Created plugin instance: %s\n", id)
		return unsafe.Pointer(handle)
	}
	
	fmt.Printf("Error: Unknown plugin ID: %s\n", id)
	return nil
}

//export ClapGo_GetVersion
func ClapGo_GetVersion(major, minor, patch *C.uint32_t) C.bool {
	if major != nil {
		*major = C.uint32_t(1)
	}
	if minor != nil {
		*minor = C.uint32_t(0)
	}
	if patch != nil {
		*patch = C.uint32_t(0)
	}
	return C.bool(true)
}

//export ClapGo_GetPluginID
func ClapGo_GetPluginID(pluginID *C.char) *C.char {
	return C.CString({{.PluginVarName}}.GetPluginID())
}

//export ClapGo_GetPluginName
func ClapGo_GetPluginName(pluginID *C.char) *C.char {
	return C.CString({{.PluginVarName}}.GetPluginInfo().Name)
}

//export ClapGo_GetPluginVendor
func ClapGo_GetPluginVendor(pluginID *C.char) *C.char {
	return C.CString({{.PluginVarName}}.GetPluginInfo().Vendor)
}

//export ClapGo_GetPluginVersion
func ClapGo_GetPluginVersion(pluginID *C.char) *C.char {
	return C.CString({{.PluginVarName}}.GetPluginInfo().Version)
}

//export ClapGo_GetPluginDescription
func ClapGo_GetPluginDescription(pluginID *C.char) *C.char {
	return C.CString({{.PluginVarName}}.GetPluginInfo().Description)
}

//export ClapGo_PluginInit
func ClapGo_PluginInit(plugin unsafe.Pointer) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*{{.StructName}})
	return C.bool(p.Init())
}

//export ClapGo_PluginDestroy
func ClapGo_PluginDestroy(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	handle := cgo.Handle(plugin)
	p := handle.Value().(*{{.StructName}})
	p.Destroy()
	handle.Delete()
}

//export ClapGo_PluginActivate
func ClapGo_PluginActivate(plugin unsafe.Pointer, sampleRate C.double, minFrames, maxFrames C.uint32_t) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*{{.StructName}})
	return C.bool(p.Activate(float64(sampleRate), uint32(minFrames), uint32(maxFrames)))
}

//export ClapGo_PluginDeactivate
func ClapGo_PluginDeactivate(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	p := cgo.Handle(plugin).Value().(*{{.StructName}})
	p.Deactivate()
}

//export ClapGo_PluginStartProcessing
func ClapGo_PluginStartProcessing(plugin unsafe.Pointer) C.bool {
	if plugin == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*{{.StructName}})
	return C.bool(p.StartProcessing())
}

//export ClapGo_PluginStopProcessing
func ClapGo_PluginStopProcessing(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	p := cgo.Handle(plugin).Value().(*{{.StructName}})
	p.StopProcessing()
}

//export ClapGo_PluginReset
func ClapGo_PluginReset(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	p := cgo.Handle(plugin).Value().(*{{.StructName}})
	p.Reset()
}

//export ClapGo_PluginProcess
func ClapGo_PluginProcess(plugin unsafe.Pointer, process unsafe.Pointer) C.int32_t {
	if plugin == nil || process == nil {
		return C.int32_t(api.ProcessError)
	}
	
	handle := cgo.Handle(plugin)
	p := handle.Value().(*{{.StructName}})
	
	// Convert the C clap_process_t to Go parameters
	cProcess := (*C.clap_process_t)(process)
	
	// Extract steady time and frame count
	steadyTime := int64(cProcess.steady_time)
	framesCount := uint32(cProcess.frames_count)
	
	// Convert audio buffers using our abstraction - NO MORE MANUAL CONVERSION!
	audioIn := api.ConvertFromCBuffers(unsafe.Pointer(cProcess.audio_inputs), uint32(cProcess.audio_inputs_count), framesCount)
	audioOut := api.ConvertFromCBuffers(unsafe.Pointer(cProcess.audio_outputs), uint32(cProcess.audio_outputs_count), framesCount)
	
	// Create event handler using the new abstraction - NO MORE MANUAL EVENT HANDLING!
	eventHandler := api.NewEventProcessor(
		unsafe.Pointer(cProcess.in_events),
		unsafe.Pointer(cProcess.out_events),
	)
	
	// Call the actual Go process method
	result := p.Process(steadyTime, framesCount, audioIn, audioOut, eventHandler)
	
	return C.int32_t(result)
}

//export ClapGo_PluginGetExtension
func ClapGo_PluginGetExtension(plugin unsafe.Pointer, id *C.char) unsafe.Pointer {
	if plugin == nil {
		return nil
	}
	p := cgo.Handle(plugin).Value().(*{{.StructName}})
	extID := C.GoString(id)
	return p.GetExtension(extID)
}

//export ClapGo_PluginOnMainThread
func ClapGo_PluginOnMainThread(plugin unsafe.Pointer) {
	if plugin == nil {
		return
	}
	p := cgo.Handle(plugin).Value().(*{{.StructName}})
	p.OnMainThread()
}

//export ClapGo_PluginParamsCount
func ClapGo_PluginParamsCount(plugin unsafe.Pointer) C.uint32_t {
	if plugin == nil {
		return 0
	}
	p := cgo.Handle(plugin).Value().(*{{.StructName}})
	return C.uint32_t(p.paramManager.GetParameterCount())
}

//export ClapGo_PluginParamsGetInfo
func ClapGo_PluginParamsGetInfo(plugin unsafe.Pointer, index C.uint32_t, info unsafe.Pointer) C.bool {
	if plugin == nil || info == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*{{.StructName}})
	
	// Get parameter info from manager
	paramInfo, err := p.paramManager.GetParameterInfoByIndex(uint32(index))
	if err != nil {
		return C.bool(false)
	}
	
	// Convert to C struct
	cInfo := (*C.clap_param_info_t)(info)
	cInfo.id = C.clap_id(paramInfo.ID)
	cInfo.flags = C.CLAP_PARAM_IS_AUTOMATABLE | C.CLAP_PARAM_IS_MODULATABLE
	cInfo.cookie = nil
	
	// Copy name
	nameBytes := []byte(paramInfo.Name)
	if len(nameBytes) >= C.CLAP_NAME_SIZE {
		nameBytes = nameBytes[:C.CLAP_NAME_SIZE-1]
	}
	for i, b := range nameBytes {
		cInfo.name[i] = C.char(b)
	}
	cInfo.name[len(nameBytes)] = 0
	
	// Clear module path
	cInfo.module[0] = 0
	
	// Set range
	cInfo.min_value = C.double(paramInfo.MinValue)
	cInfo.max_value = C.double(paramInfo.MaxValue)
	cInfo.default_value = C.double(paramInfo.DefaultValue)
	
	return C.bool(true)
}

//export ClapGo_PluginParamsGetValue
func ClapGo_PluginParamsGetValue(plugin unsafe.Pointer, paramID C.uint32_t, value *C.double) C.bool {
	if plugin == nil || value == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*{{.StructName}})
	
	// Get current value from parameter manager
	val := p.paramManager.GetParameterValue(uint32(paramID))
	*value = C.double(val)
	return C.bool(true)
}

//export ClapGo_PluginParamsValueToText
func ClapGo_PluginParamsValueToText(plugin unsafe.Pointer, paramID C.uint32_t, value C.double, buffer *C.char, size C.uint32_t) C.bool {
	if plugin == nil || buffer == nil || size == 0 {
		return C.bool(false)
	}
	
	// TODO: Format parameter value as text based on parameter type
	// For now, just use a simple format
	text := fmt.Sprintf("%.2f", float64(value))
	
	// Copy to C buffer
	textBytes := []byte(text)
	maxLen := int(size) - 1
	if len(textBytes) > maxLen {
		textBytes = textBytes[:maxLen]
	}
	
	cBuffer := (*[1 << 30]C.char)(unsafe.Pointer(buffer))
	for i, b := range textBytes {
		cBuffer[i] = C.char(b)
	}
	cBuffer[len(textBytes)] = 0
	
	return C.bool(true)
}

//export ClapGo_PluginParamsTextToValue
func ClapGo_PluginParamsTextToValue(plugin unsafe.Pointer, paramID C.uint32_t, text *C.char, value *C.double) C.bool {
	if plugin == nil || text == nil || value == nil {
		return C.bool(false)
	}
	
	// Convert text to Go string
	goText := C.GoString(text)
	
	// Parse the value
	var val float64
	if _, err := fmt.Sscanf(goText, "%f", &val); err == nil {
		*value = C.double(val)
		return C.bool(true)
	}
	
	return C.bool(false)
}

//export ClapGo_PluginParamsFlush
func ClapGo_PluginParamsFlush(plugin unsafe.Pointer, inEvents unsafe.Pointer, outEvents unsafe.Pointer) {
	if plugin == nil {
		return
	}
	p := cgo.Handle(plugin).Value().(*{{.StructName}})
	
	// Process events using our abstraction
	if inEvents != nil {
		eventHandler := api.NewEventProcessor(inEvents, outEvents)
		p.processEvents(eventHandler, 0)
	}
}

//export ClapGo_PluginStateSave
func ClapGo_PluginStateSave(plugin unsafe.Pointer, stream unsafe.Pointer) C.bool {
	if plugin == nil || stream == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*{{.StructName}})
	return C.bool(p.SaveState(stream))
}

//export ClapGo_PluginStateLoad
func ClapGo_PluginStateLoad(plugin unsafe.Pointer, stream unsafe.Pointer) C.bool {
	if plugin == nil || stream == nil {
		return C.bool(false)
	}
	p := cgo.Handle(plugin).Value().(*{{.StructName}})
	return C.bool(p.LoadState(stream))
}
`

const extensionsTemplate = `// Code generated by clapgo-generate. DO NOT EDIT.
// This file implements CLAP extension support based on your plugin type.
// The C bridge handles the actual extension interfaces.
package main

import (
	"unsafe"
)

// GetExtension returns the appropriate extension implementation for the given ID.
// Extensions are handled by the C bridge layer, so we always return nil here.
// The C bridge provides implementations for params, state, audio-ports, etc.
func (p *{{.StructName}}) GetExtension(id string) unsafe.Pointer {
	// All extensions are handled by the C bridge
	return nil
}
`

const constantsTemplate = `// Code generated by clapgo-generate. DO NOT EDIT.
// This file contains plugin metadata constants extracted from the manifest.
package main

import (
	"fmt"
)

// Plugin identification
const (
	PluginID      = "{{.PluginID}}"
	PluginName    = "{{.PluginName}}"
	PluginVendor  = "{{.PluginVendor}}"
	PluginVersion = "1.0.0"
	PluginURL     = "https://github.com/justyntemme/clapgo"
)

// Plugin features
var PluginFeatures = []string{
{{range .Features}}	"{{.}}",
{{end}}}

// Parameter definitions with rich metadata for DAW GUI generation
type ParameterInfo struct {
	ID          uint32
	Name        string
	ShortName   string
	Module      string
	MinValue    float64
	MaxValue    float64
	DefaultValue float64
	Flags       uint32
	Cookie      interface{}
}

// Parameter metadata for enhanced DAW integration
var ParameterMetadata = []ParameterInfo{
{{if eq .PluginType "audio-effect"}}
	{
		ID:           1,
		Name:         "Input Gain",
		ShortName:    "In Gain",
		Module:       "Input",
		MinValue:     -24.0,
		MaxValue:     24.0,
		DefaultValue: 0.0,
		Flags:        0, // CLAP_PARAM_IS_AUTOMATABLE
	},
	{
		ID:           2,
		Name:         "Dry/Wet Mix",
		ShortName:    "Mix",
		Module:       "Output",
		MinValue:     0.0,
		MaxValue:     1.0,
		DefaultValue: 1.0,
		Flags:        0,
	},
	{
		ID:           3,
		Name:         "Output Gain",
		ShortName:    "Out Gain",
		Module:       "Output",
		MinValue:     -24.0,
		MaxValue:     24.0,
		DefaultValue: 0.0,
		Flags:        0,
	},
{{end}}
{{if eq .PluginType "instrument"}}
	{
		ID:           1,
		Name:         "Master Volume",
		ShortName:    "Volume",
		Module:       "Master",
		MinValue:     0.0,
		MaxValue:     1.0,
		DefaultValue: 0.7,
		Flags:        0, // CLAP_PARAM_IS_AUTOMATABLE
	},
	{
		ID:           2,
		Name:         "Oscillator Waveform",
		ShortName:    "Wave",
		Module:       "Oscillator",
		MinValue:     0.0,
		MaxValue:     3.0,
		DefaultValue: 0.0,
		Flags:        1, // CLAP_PARAM_IS_STEPPED
	},
	{
		ID:           3,
		Name:         "Envelope Attack",
		ShortName:    "Attack",
		Module:       "Envelope",
		MinValue:     0.001,
		MaxValue:     2.0,
		DefaultValue: 0.01,
		Flags:        0,
	},
	{
		ID:           4,
		Name:         "Envelope Decay",
		ShortName:    "Decay",
		Module:       "Envelope",
		MinValue:     0.001,
		MaxValue:     2.0,
		DefaultValue: 0.1,
		Flags:        0,
	},
	{
		ID:           5,
		Name:         "Envelope Sustain",
		ShortName:    "Sustain",
		Module:       "Envelope",
		MinValue:     0.0,
		MaxValue:     1.0,
		DefaultValue: 0.7,
		Flags:        0,
	},
	{
		ID:           6,
		Name:         "Envelope Release",
		ShortName:    "Release",
		Module:       "Envelope",
		MinValue:     0.001,
		MaxValue:     5.0,
		DefaultValue: 0.3,
		Flags:        0,
	},
{{end}}
{{if eq .PluginType "utility"}}
	{
		ID:           1,
		Name:         "Gain",
		ShortName:    "Gain",
		Module:       "Main",
		MinValue:     -60.0,
		MaxValue:     24.0,
		DefaultValue: 0.0,
		Flags:        0, // CLAP_PARAM_IS_AUTOMATABLE
	},
	{
		ID:           2,
		Name:         "Enable",
		ShortName:    "On",
		Module:       "Main",
		MinValue:     0.0,
		MaxValue:     1.0,
		DefaultValue: 1.0,
		Flags:        1, // CLAP_PARAM_IS_STEPPED
	},
{{end}}
}

// Parameter display helpers for DAW integration
func formatParameterValue(paramID uint32, value float64) string {
	switch paramID {
	{{if eq .PluginType "audio-effect"}}
	case 1, 3: // Gain parameters
		return fmt.Sprintf("%.1f dB", value)
	case 2: // Mix parameter
		return fmt.Sprintf("%.0f%%", value*100)
	{{end}}
	{{if eq .PluginType "instrument"}}
	case 1: // Volume
		return fmt.Sprintf("%.0f%%", value*100)
	case 2: // Waveform
		waveforms := []string{"Sine", "Sawtooth", "Square", "Triangle"}
		if int(value) < len(waveforms) {
			return waveforms[int(value)]
		}
		return "Unknown"
	case 3, 4, 6: // Time-based parameters
		return fmt.Sprintf("%.3f s", value)
	case 5: // Sustain
		return fmt.Sprintf("%.0f%%", value*100)
	{{end}}
	{{if eq .PluginType "utility"}}
	case 1: // Gain
		return fmt.Sprintf("%.1f dB", value)
	case 2: // Enable
		if value > 0.5 {
			return "On"
		}
		return "Off"
	{{end}}
	default:
		return fmt.Sprintf("%.3f", value)
	}
}

// Audio configuration defaults
const (
	DefaultSampleRate = 48000
	DefaultBlockSize  = 512
)

{{if eq .PluginType "instrument"}}
// Voice configuration for instruments
const (
	MaxVoices     = 16
	MaxNoteLength = 120 // seconds
)
{{end}}

{{if or (eq .PluginType "audio-effect") (eq .PluginType "utility")}}
// Processing configuration for effects
const (
	MaxLatency = 512  // samples
	MaxTail    = 4800 // samples (100ms at 48kHz)
)
{{end}}
`

const pluginTemplate = `package main

import (
	"math"
	// "sync/atomic" // Uncomment when using atomic parameter storage
	"unsafe"
	
	"github.com/justyntemme/clapgo/pkg/api"
)

// {{.StructName}} implements a {{.PluginType}} plugin.
type {{.StructName}} struct {
	// Plugin state
	sampleRate   float64
	isActivated  bool
	isProcessing bool
	host         unsafe.Pointer
	
	// Parameters with atomic storage for thread safety
	// TODO: Add your parameters here with atomic storage
	// Example: gain int64 // atomic storage for gain value
	
	// Parameter management using our new abstraction
	paramManager *api.ParameterManager
{{if eq .PluginType "instrument"}}
	activeVoices int32  // atomic counter for active voices
{{end}}
}

// New{{.StructName}} creates a new instance of the plugin
func New{{.StructName}}() *{{.StructName}} {
	plugin := &{{.StructName}}{
		sampleRate:   44100.0,
		isActivated:  false,
		isProcessing: false,
		paramManager: api.NewParameterManager(),
	}
	
	// TODO: Set default parameter values atomically
	// Example: atomic.StoreInt64(&plugin.gain, int64(floatToBits(1.0)))
	
	// TODO: Register parameters using our new abstraction
	// Example: plugin.paramManager.RegisterParameter(api.CreateFloatParameter(0, "Gain", 0.0, 2.0, 1.0))
	
	return plugin
}

// Init initializes the plugin
func (p *{{.StructName}}) Init() bool {
	// TODO: Initialize plugin resources
	return true
}

// Destroy cleans up plugin resources
func (p *{{.StructName}}) Destroy() {
	// TODO: Clean up any allocated resources
}

// Activate prepares the plugin for processing
func (p *{{.StructName}}) Activate(sampleRate float64, minFrames, maxFrames uint32) bool {
	p.sampleRate = sampleRate
	p.isActivated = true
	// TODO: Prepare for processing at the given sample rate
	return true
}

// Deactivate stops the plugin from processing
func (p *{{.StructName}}) Deactivate() {
	p.isActivated = false
}

// StartProcessing begins audio processing
func (p *{{.StructName}}) StartProcessing() bool {
	if !p.isActivated {
		return false
	}
	p.isProcessing = true
	return true
}

// StopProcessing ends audio processing
func (p *{{.StructName}}) StopProcessing() {
	p.isProcessing = false
}

// Reset resets the plugin state
func (p *{{.StructName}}) Reset() {
	// TODO: Reset plugin state to defaults
}

// Process processes audio data using the new abstractions
func (p *{{.StructName}}) Process(steadyTime int64, framesCount uint32, audioIn, audioOut [][]float32, events api.EventHandler) int {
	// Check if we're in a valid state for processing
	if !p.isActivated || !p.isProcessing {
		return api.ProcessError
	}
	
	// Process events using our new abstraction
	if events != nil {
		p.processEvents(events, framesCount)
	}
	
	// TODO: Get current parameter values atomically
	// Example: gainBits := atomic.LoadInt64(&p.gain)
	// Example: gain := floatFromBits(uint64(gainBits))
	
{{if eq .PluginType "instrument"}}
	// TODO: Process MIDI events and generate audio
	// Example:
	// - Handle note on/off events
	// - Update voice states
	// - Generate audio for active voices
	// - Mix to output
{{else if or (eq .PluginType "audio-effect") (eq .PluginType "utility")}}
	// If no audio inputs or outputs, nothing to do
	if len(audioIn) == 0 || len(audioOut) == 0 {
		return api.ProcessContinue
	}
	
	// Get the number of channels (use min of input and output)
	numChannels := len(audioIn)
	if len(audioOut) < numChannels {
		numChannels = len(audioOut)
	}
	
	// TODO: Process audio - apply your effect to each sample
	// Example:
	// for ch := 0; ch < numChannels; ch++ {
	//     inChannel := audioIn[ch]
	//     outChannel := audioOut[ch]
	//     
	//     // Make sure we have enough buffer space
	//     if len(inChannel) < int(framesCount) || len(outChannel) < int(framesCount) {
	//         continue // Skip this channel if buffer is too small
	//     }
	//     
	//     // Apply effect to each sample
	//     for i := uint32(0); i < framesCount; i++ {
	//         outChannel[i] = inChannel[i] * float32(gain)
	//     }
	// }
	// }
{{else if eq .PluginType "note-effect"}}
	// TODO: Process MIDI events
	// Example:
	// events := p.eventHandler.GetNoteEvents()
	// for _, event := range events {
	//     // Transform or generate MIDI events
	//     p.eventHandler.SendNoteEvent(transformedEvent)
	// }
{{else if eq .PluginType "analyzer"}}
	// TODO: Analyze audio and update metering/display data
	// Example:
	// for ch := 0; ch < audio.ChannelCount; ch++ {
	//     level := calculateRMS(audio.Input[ch])
	//     p.updateMeter(ch, level)
	// }
{{else if eq .PluginType "note-detector"}}
	// TODO: Analyze audio and generate MIDI events
	// Example:
	// pitch := p.detectPitch(audio.Input[0])
	// if pitch.Confidence > 0.8 {
	//     note := p.frequencyToMIDI(pitch.Frequency)
	//     p.eventHandler.SendNoteOn(note, 100)
	// }
{{end}}
	
	return api.ProcessContinue
}

// processEvents handles all incoming events using our new EventHandler abstraction
func (p *{{.StructName}}) processEvents(events api.EventHandler, frameCount uint32) {
	if events == nil {
		return
	}
	
	// Process each event using our abstraction
	eventCount := events.GetInputEventCount()
	for i := uint32(0); i < eventCount; i++ {
		event := events.GetInputEvent(i)
		if event == nil {
			continue
		}
		
		// Handle parameter events using our abstraction
		switch event.Type {
		case api.EventTypeParamValue:
			if paramEvent, ok := event.Data.(api.ParamEvent); ok {
				p.handleParameterChange(paramEvent)
			}
		}
	}
}

// handleParameterChange processes a parameter change event
func (p *{{.StructName}}) handleParameterChange(paramEvent api.ParamEvent) {
	// TODO: Handle parameter changes based on ID
	// Example:
	// switch paramEvent.ParamID {
	// case 0: // Gain parameter
	//     value := paramEvent.Value
	//     // Clamp value to valid range
	//     if value < 0.0 {
	//         value = 0.0
	//     }
	//     if value > 2.0 {
	//         value = 2.0
	//     }
	//     atomic.StoreInt64(&p.gain, int64(floatToBits(value)))
	//     
	//     // Update parameter manager
	//     p.paramManager.SetParameterValue(paramEvent.ParamID, value)
	// }
}

// GetPluginInfo returns information about the plugin
func (p *{{.StructName}}) GetPluginInfo() api.PluginInfo {
	return api.PluginInfo{
		ID:          PluginID,
		Name:        PluginName,
		Vendor:      PluginVendor,
		URL:         PluginURL,
		ManualURL:   PluginURL,
		SupportURL:  PluginURL + "/issues",
		Version:     PluginVersion,
		Description: "{{.PluginName}} - A {{.PluginType}} plugin",
		Features:    PluginFeatures,
	}
}

// OnMainThread is called on the main thread
func (p *{{.StructName}}) OnMainThread() {
	// TODO: Handle main thread operations if needed
}

// GetPluginID returns the plugin ID
func (p *{{.StructName}}) GetPluginID() string {
	return PluginID
}

// SaveState saves the plugin state to a stream
func (p *{{.StructName}}) SaveState(stream unsafe.Pointer) bool {
	out := api.NewOutputStream(stream)
	
	// Write state version
	if err := out.WriteUint32(1); err != nil {
		return false
	}
	
	// Write parameter count
	paramCount := p.paramManager.GetParameterCount()
	if err := out.WriteUint32(paramCount); err != nil {
		return false
	}
	
	// Write each parameter ID and value
	for i := uint32(0); i < paramCount; i++ {
		info, err := p.paramManager.GetParameterInfoByIndex(i)
		if err != nil {
			return false
		}
		
		// Write parameter ID
		if err := out.WriteUint32(info.ID); err != nil {
			return false
		}
		
		// Write parameter value
		value := p.paramManager.GetParameterValue(info.ID)
		if err := out.WriteFloat64(value); err != nil {
			return false
		}
	}
	
	return true
}

// LoadState loads the plugin state from a stream
func (p *{{.StructName}}) LoadState(stream unsafe.Pointer) bool {
	in := api.NewInputStream(stream)
	
	// Read state version
	version, err := in.ReadUint32()
	if err != nil || version != 1 {
		return false
	}
	
	// Read parameter count
	paramCount, err := in.ReadUint32()
	if err != nil {
		return false
	}
	
	// Read each parameter ID and value
	for i := uint32(0); i < paramCount; i++ {
		// Read parameter ID
		paramID, err := in.ReadUint32()
		if err != nil {
			return false
		}
		
		// Read parameter value
		value, err := in.ReadFloat64()
		if err != nil {
			return false
		}
		
		// Set parameter value
		p.paramManager.SetParameterValue(paramID, value)
		
		// TODO: Update internal state if needed
		// Example:
		// if paramID == 0 {
		//     atomic.StoreInt64(&p.gain, int64(floatToBits(value)))
		// }
	}
	
	return true
}

// Helper functions for atomic float64 operations
// Uncomment these when using atomic parameter storage:
// func floatToBits(f float64) uint64 {
// 	return *(*uint64)(unsafe.Pointer(&f))
// }
// 
// func floatFromBits(b uint64) float64 {
// 	return *(*float64)(unsafe.Pointer(&b))
// }

{{if or (eq .PluginType "audio-effect") (eq .PluginType "utility")}}
// GetLatency returns the plugin's processing latency in samples.
func (p *{{.StructName}}) GetLatency() uint32 {
	// TODO: Return your plugin's latency
	// This is used for delay compensation in the DAW
	return 0
}

// GetTail returns the plugin's tail length in samples.
func (p *{{.StructName}}) GetTail() uint32 {
	// TODO: Return how long the plugin produces output after input stops
	// Important for reverbs, delays, etc.
	return 0
}
{{end}}

{{if eq .PluginType "instrument"}}
// Voice represents a single synthesizer voice (example)
type Voice struct {
	// TODO: Add voice state here
}

// handleNoteOn processes a note on event.
func (p *{{.StructName}}) handleNoteOn(key int, velocity float64) {
	// TODO: Implement note on handling
	// Example:
	// voice := p.allocateVoice()
	// if voice != nil {
	//     voice.start(key, velocity)
	// }
}

// handleNoteOff processes a note off event.
func (p *{{.StructName}}) handleNoteOff(key int) {
	// TODO: Implement note off handling
	// Example:
	// for _, voice := range p.voices {
	//     if voice.active && voice.note == key {
	//         voice.release()
	//     }
	// }
}

// GetVoiceInfo returns information about the instrument's voice capabilities.
func (p *{{.StructName}}) GetVoiceInfo() api.VoiceInfo {
	activeVoices := atomic.LoadInt32(&p.activeVoices)
	return api.VoiceInfo{
		VoiceCount:       uint32(activeVoices),
		VoiceCapacity:    MaxVoices,
		Flags:            api.VoiceInfoSupportsStereo,
	}
}

// allocateVoice finds a free voice for a new note.
func (p *{{.StructName}}) allocateVoice() *Voice {
	// TODO: Implement voice allocation
	// Find a free voice or steal the oldest one
	return nil
}
{{end}}

// DSP Helper Functions & Examples
// ==============================
// The following functions provide common DSP algorithms you can use as starting points:

{{if eq .PluginType "audio-effect"}}
// Example: Simple low-pass filter
func lowPassFilter(input, cutoff, sampleRate float64, state *float64) float64 {
	rc := 1.0 / (cutoff * 2.0 * math.Pi)
	dt := 1.0 / sampleRate
	alpha := dt / (rc + dt)
	*state = *state + alpha*(input-*state)
	return *state
}

// Example: dB to linear conversion
func dbToLinear(db float64) float64 {
	return math.Pow(10.0, db/20.0)
}

// Example: Simple delay line
type DelayLine struct {
	buffer []float64
	index  int
}

func newDelayLine(maxDelay int) *DelayLine {
	return &DelayLine{
		buffer: make([]float64, maxDelay),
		index:  0,
	}
}

func (d *DelayLine) process(input float64, delaySamples int) float64 {
	// Read from delay line
	readIndex := (d.index - delaySamples + len(d.buffer)) % len(d.buffer)
	output := d.buffer[readIndex]
	
	// Write to delay line
	d.buffer[d.index] = input
	d.index = (d.index + 1) % len(d.buffer)
	
	return output
}
{{end}}

{{if eq .PluginType "instrument"}}
// Example: Basic oscillator
func generateOscillator(phase float64, waveform int) float64 {
	switch waveform {
	case 0: // Sine
		return math.Sin(2 * math.Pi * phase)
	case 1: // Sawtooth
		return 2*phase - 1
	case 2: // Square
		if phase < 0.5 {
			return 1.0
		}
		return -1.0
	case 3: // Triangle
		if phase < 0.5 {
			return 4*phase - 1
		}
		return 3 - 4*phase
	default:
		return 0.0
	}
}

// Example: Note to frequency conversion
func noteToFrequency(note int) float64 {
	return 440.0 * math.Pow(2.0, float64(note-69)/12.0)
}
{{end}}

{{if eq .PluginType "utility"}}
// Example: dB to linear conversion
func dbToLinear(db float64) float64 {
	return math.Pow(10.0, db/20.0)
}

// Example: Linear to dB conversion  
func linearToDb(linear float64) float64 {
	if linear <= 0 {
		return -96.0 // Minimum dB
	}
	return 20.0 * math.Log10(linear)
}

// Example: RMS level calculation
func calculateRMS(samples []float64) float64 {
	sum := 0.0
	for _, sample := range samples {
		sum += sample * sample
	}
	return math.Sqrt(sum / float64(len(samples)))
}
{{end}}

func main() {
	// This is not called when used as a plugin,
	// but can be useful for testing
}
`
