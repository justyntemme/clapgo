package goclap

// #include <stdlib.h>
// #include <string.h>
// #include "../../include/clap/include/clap/clap.h"
// #include "../../include/clap/include/clap/factory/preset-discovery.h"
import "C"
import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"
)

// Constants for preset discovery factory
const (
	PresetDiscoveryFactoryID = "clap.preset-discovery-factory/2"
)

// PresetDiscoveryProviderDescriptor describes a preset provider
type PresetDiscoveryProviderDescriptor struct {
	ID      string
	Name    string
	Vendor  string
}

// PresetDiscoveryProvider represents a preset provider
type PresetDiscoveryProvider struct {
	descriptor PresetDiscoveryProviderDescriptor
	filetypes  []PresetDiscoveryFiletype
	locations  []PresetDiscoveryLocation
	soundpacks []PresetDiscoverySoundpack
}

// PresetDiscoveryFiletype describes a preset file type
type PresetDiscoveryFiletype struct {
	Name          string
	Description   string
	FileExtension string
}

// PresetDiscoveryLocation describes a preset location
type PresetDiscoveryLocation struct {
	Flags    uint32
	Name     string
	Kind     uint32
	Location string
}

// PresetDiscoverySoundpack describes a soundpack
type PresetDiscoverySoundpack struct {
	Flags            uint32
	ID               string
	Name             string
	Description      string
	HomepageURL      string
	Vendor           string
	ImagePath        string
	ReleaseTimestamp int64
}

// PresetMetadata holds metadata about a preset
type PresetMetadata struct {
	Name        string            `json:"name"`
	LoadKey     string            `json:"load_key,omitempty"`
	PluginIDs   []string          `json:"plugin_ids"`
	SoundpackID string            `json:"soundpack_id,omitempty"`
	Flags       uint32            `json:"flags"`
	Creators    []string          `json:"creators,omitempty"`
	Description string            `json:"description,omitempty"`
	CreateTime  int64             `json:"create_time,omitempty"`
	ModifyTime  int64             `json:"modify_time,omitempty"`
	Features    []string          `json:"features,omitempty"`
	ExtraInfo   map[string]string `json:"extra_info,omitempty"`
}

// PresetProvider interface that plugins implement for preset support
type PresetProvider interface {
	// GetPresetDescriptor returns the preset provider descriptor
	GetPresetDescriptor() PresetDiscoveryProviderDescriptor
	
	// GetPresetFiletypes returns the filetypes supported by this provider
	GetPresetFiletypes() []PresetDiscoveryFiletype
	
	// GetPresetLocations returns the preset locations for this provider
	GetPresetLocations() []PresetDiscoveryLocation
	
	// GetPresetSoundpacks returns the soundpacks for this provider
	GetPresetSoundpacks() []PresetDiscoverySoundpack
	
	// GetPresetMetadata reads metadata from a preset file
	// Returns nil if the file is not a valid preset
	GetPresetMetadata(locationKind uint32, location string) ([]PresetMetadata, error)
	
	// LoadPreset loads a preset from a location
	LoadPreset(locationKind uint32, location, loadKey string) bool
}

// DefaultPresetProvider implements a basic preset provider with JSON preset files
type DefaultPresetProvider struct {
	descriptor        PresetDiscoveryProviderDescriptor
	presetExtensions  []string
	presetDirectory   string
	factoryPresetDir  string
	userPresetDir     string
	supportedPluginIDs []string
}

// NewDefaultPresetProvider creates a new preset provider
func NewDefaultPresetProvider(descriptor PresetDiscoveryProviderDescriptor, 
	presetExtensions []string, 
	presetDirectory string,
	supportedPluginIDs []string) *DefaultPresetProvider {
	
	// Create default directories if needed
	if presetDirectory == "" {
		// Use current directory as fallback
		presetDirectory = "presets"
	}
	
	// Create provider
	provider := &DefaultPresetProvider{
		descriptor:        descriptor,
		presetExtensions:  presetExtensions,
		presetDirectory:   presetDirectory,
		factoryPresetDir:  filepath.Join(presetDirectory, "factory"),
		userPresetDir:     filepath.Join(presetDirectory, "user"),
		supportedPluginIDs: supportedPluginIDs,
	}
	
	// Create directories if they don't exist
	os.MkdirAll(provider.factoryPresetDir, 0755)
	os.MkdirAll(provider.userPresetDir, 0755)
	
	return provider
}

// GetPresetDescriptor returns the preset provider descriptor
func (p *DefaultPresetProvider) GetPresetDescriptor() PresetDiscoveryProviderDescriptor {
	return p.descriptor
}

// GetPresetFiletypes returns the filetypes supported by this provider
func (p *DefaultPresetProvider) GetPresetFiletypes() []PresetDiscoveryFiletype {
	filetypes := make([]PresetDiscoveryFiletype, 0, len(p.presetExtensions))
	
	for _, ext := range p.presetExtensions {
		filetypes = append(filetypes, PresetDiscoveryFiletype{
			Name:          ext + " preset",
			Description:   "ClapGo " + ext + " preset file",
			FileExtension: ext,
		})
	}
	
	return filetypes
}

// GetPresetLocations returns the preset locations for this provider
func (p *DefaultPresetProvider) GetPresetLocations() []PresetDiscoveryLocation {
	// Location flags
	const (
		IS_FACTORY_CONTENT = 1 << 0
		IS_USER_CONTENT    = 1 << 1
		IS_DEMO_CONTENT    = 1 << 2
		IS_FAVORITE        = 1 << 3
	)
	
	// Location kinds
	const (
		LOCATION_FILE   = 0
		LOCATION_PLUGIN = 1
	)
	
	return []PresetDiscoveryLocation{
		{
			Flags:    IS_FACTORY_CONTENT,
			Name:     "Factory Presets",
			Kind:     LOCATION_FILE,
			Location: p.factoryPresetDir,
		},
		{
			Flags:    IS_USER_CONTENT,
			Name:     "User Presets",
			Kind:     LOCATION_FILE,
			Location: p.userPresetDir,
		},
	}
}

// GetPresetSoundpacks returns the soundpacks for this provider
func (p *DefaultPresetProvider) GetPresetSoundpacks() []PresetDiscoverySoundpack {
	// For this simple implementation, we don't have any soundpacks
	return nil
}

// GetPresetMetadata reads metadata from a preset file
func (p *DefaultPresetProvider) GetPresetMetadata(locationKind uint32, location string) ([]PresetMetadata, error) {
	// Check if it's a directory
	fileInfo, err := os.Stat(location)
	if err != nil {
		return nil, err
	}
	
	// If it's a directory, scan for preset files
	if fileInfo.IsDir() {
		return p.scanDirectoryForPresets(location)
	}
	
	// Otherwise, treat as a single file
	metadata, err := p.readPresetMetadata(location)
	if err != nil {
		return nil, err
	}
	
	return []PresetMetadata{metadata}, nil
}

// scanDirectoryForPresets scans a directory for preset files
func (p *DefaultPresetProvider) scanDirectoryForPresets(directory string) ([]PresetMetadata, error) {
	presets := []PresetMetadata{}
	
	// Read directory contents
	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	
	// Process each file
	for _, file := range files {
		if file.IsDir() {
			// Skip directories
			continue
		}
		
		// Check if the file has one of our preset extensions
		filename := file.Name()
		hasValidExtension := false
		for _, ext := range p.presetExtensions {
			if strings.HasSuffix(strings.ToLower(filename), "."+ext) {
				hasValidExtension = true
				break
			}
		}
		
		if hasValidExtension {
			// Read preset metadata
			filepath := filepath.Join(directory, filename)
			metadata, err := p.readPresetMetadata(filepath)
			if err == nil {
				presets = append(presets, metadata)
			}
		}
	}
	
	return presets, nil
}

// readPresetMetadata reads metadata from a single preset file
func (p *DefaultPresetProvider) readPresetMetadata(filepath string) (PresetMetadata, error) {
	// Open the file
	file, err := os.Open(filepath)
	if err != nil {
		return PresetMetadata{}, err
	}
	defer file.Close()
	
	// Read the file content
	decoder := json.NewDecoder(file)
	var metadata PresetMetadata
	if err := decoder.Decode(&metadata); err != nil {
		return PresetMetadata{}, err
	}
	
	// If no plugin IDs are specified, use the provider's supported IDs
	if len(metadata.PluginIDs) == 0 {
		metadata.PluginIDs = p.supportedPluginIDs
	}
	
	// Set load key to the filename if not specified
	if metadata.LoadKey == "" {
		metadata.LoadKey = filepath
	}
	
	return metadata, nil
}

// LoadPreset loads a preset from a location
func (p *DefaultPresetProvider) LoadPreset(locationKind uint32, location, loadKey string) bool {
	// In this simple implementation, we just read the preset file
	file, err := os.Open(loadKey)
	if err != nil {
		return false
	}
	defer file.Close()
	
	// Parse the preset data
	decoder := json.NewDecoder(file)
	var preset PresetMetadata
	if err := decoder.Decode(&preset); err != nil {
		return false
	}
	
	return true
}

// PresetDiscoveryFactory represents a preset discovery factory
type PresetDiscoveryFactory struct {
	providers []PresetProvider
}

// Global factory instance
var (
	globalPresetFactory *PresetDiscoveryFactory
	factoryMutex        sync.RWMutex
)

// RegisterPresetProvider adds a preset provider to the global factory
func RegisterPresetProvider(provider PresetProvider) {
	factoryMutex.Lock()
	defer factoryMutex.Unlock()
	
	// Create global factory if it doesn't exist
	if globalPresetFactory == nil {
		globalPresetFactory = &PresetDiscoveryFactory{
			providers: make([]PresetProvider, 0),
		}
	}
	
	// Add the provider
	globalPresetFactory.providers = append(globalPresetFactory.providers, provider)
}

// GetPresetProviderCount returns the number of registered preset providers
func GetPresetProviderCount() uint32 {
	factoryMutex.RLock()
	defer factoryMutex.RUnlock()
	
	if globalPresetFactory == nil {
		return 0
	}
	
	return uint32(len(globalPresetFactory.providers))
}

// GetPresetProviderDescriptor returns the descriptor for a preset provider
func GetPresetProviderDescriptor(index uint32) *PresetDiscoveryProviderDescriptor {
	factoryMutex.RLock()
	defer factoryMutex.RUnlock()
	
	if globalPresetFactory == nil || int(index) >= len(globalPresetFactory.providers) {
		return nil
	}
	
	// Get the descriptor from the provider
	desc := globalPresetFactory.providers[index].GetPresetDescriptor()
	return &desc
}

// PresetDiscoveryFactoryProvider is an interface for providing preset
// discovery factory capabilities
type PresetDiscoveryFactoryProvider interface {
	// GetPresetProviders returns the list of preset providers
	GetPresetProviders() []PresetProvider
}

// createPresetDiscoveryFactory creates a factory for preset discovery
func createPresetDiscoveryFactory() unsafe.Pointer {
	// Create a minimal implementation to support preset discovery
	cFactory := (*C.clap_preset_discovery_factory_t)(C.malloc(C.sizeof_clap_preset_discovery_factory_t))
	if cFactory == nil {
		return nil
	}
	
	// Initialize the factory structure to zero
	C.memset(unsafe.Pointer(cFactory), 0, C.sizeof_clap_preset_discovery_factory_t)
	
	// Return the factory
	return unsafe.Pointer(cFactory)
}

//export goPresetDiscoveryFactoryCount
func goPresetDiscoveryFactoryCount(factory unsafe.Pointer) C.uint32_t {
	return C.uint32_t(GetPresetProviderCount())
}

//export goPresetDiscoveryFactoryGetDescriptor
func goPresetDiscoveryFactoryGetDescriptor(factory unsafe.Pointer, index C.uint32_t) *C.clap_preset_discovery_provider_descriptor_t {
	desc := GetPresetProviderDescriptor(uint32(index))
	if desc == nil {
		return nil
	}
	
	// Allocate and fill the C descriptor
	cDesc := (*C.clap_preset_discovery_provider_descriptor_t)(C.malloc(C.sizeof_clap_preset_discovery_provider_descriptor_t))
	
	// Set the fields
	cDesc.id = C.CString(desc.ID)
	cDesc.name = C.CString(desc.Name)
	cDesc.vendor = C.CString(desc.Vendor)
	
	return cDesc
}

//export goPresetDiscoveryFactoryCreate
func goPresetDiscoveryFactoryCreate(factory unsafe.Pointer, host unsafe.Pointer, provider_id *C.char) unsafe.Pointer {
	// For now, just return nil 
	// We'll implement this properly in a future version
	return nil
}

//export goPresetDiscoveryProviderInit
func goPresetDiscoveryProviderInit(provider *C.clap_preset_discovery_provider_t) C.bool {
	// Nothing to initialize
	return C.bool(true)
}

//export goPresetDiscoveryProviderDestroy
func goPresetDiscoveryProviderDestroy(provider *C.clap_preset_discovery_provider_t) {
	// Clean up our registry
	removeProviderHandle(unsafe.Pointer(provider))
}

//export goPresetDiscoveryProviderGetMetadata
func goPresetDiscoveryProviderGetMetadata(provider unsafe.Pointer, 
	location_kind C.uint32_t, location *C.char) unsafe.Pointer {
	// For now, just return nil
	// We'll implement this properly in a future version
	return nil
}

//export goPresetDiscoveryMetadataReceiverReceive
func goPresetDiscoveryMetadataReceiverReceive(receiver unsafe.Pointer, 
	metadata unsafe.Pointer) C.bool {
	// For now, just return false
	// We'll implement this properly in a future version
	return C.bool(false)
}

// Global registry to store provider handles
var (
	providerHandlesMutex sync.RWMutex
	providerHandles      = make(map[unsafe.Pointer]unsafe.Pointer)
)

// setProviderHandle stores a Go provider handle in our registry
func setProviderHandle(cHandle, goHandle unsafe.Pointer) {
	providerHandlesMutex.Lock()
	defer providerHandlesMutex.Unlock()
	providerHandles[cHandle] = goHandle
}

// getProviderHandle retrieves a Go provider handle from our registry
func getProviderHandle(cHandle unsafe.Pointer) unsafe.Pointer {
	providerHandlesMutex.RLock()
	defer providerHandlesMutex.RUnlock()
	return providerHandles[cHandle]
}

// removeProviderHandle removes a provider handle from our registry
func removeProviderHandle(cHandle unsafe.Pointer) {
	providerHandlesMutex.Lock()
	defer providerHandlesMutex.Unlock()
	delete(providerHandles, cHandle)
}

// Global registry to store metadata handles
var (
	metadataHandlesMutex sync.RWMutex
	metadataHandles      = make(map[unsafe.Pointer]unsafe.Pointer)
)

// setMetadataHandle stores metadata in our registry
func setMetadataHandle(cHandle, goHandle unsafe.Pointer) {
	metadataHandlesMutex.Lock()
	defer metadataHandlesMutex.Unlock()
	metadataHandles[cHandle] = goHandle
}

// getMetadataHandle retrieves metadata from our registry
func getMetadataHandle(cHandle unsafe.Pointer) unsafe.Pointer {
	metadataHandlesMutex.RLock()
	defer metadataHandlesMutex.RUnlock()
	return metadataHandles[cHandle]
}