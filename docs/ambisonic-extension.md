# Ambisonic Extension Guide

The ambisonic extension allows plugins to declare support for ambisonic audio processing and specify the channel ordering and normalization schemes they use.

## Overview

Ambisonic audio is a full-sphere surround sound format that uses spherical harmonics to encode spatial audio. Different ambisonic formats use different channel ordering and normalization schemes, which this extension helps declare.

## Implementation

### 1. Implement the AmbisonicProvider Interface

```go
import "github.com/your-org/clapgo/pkg/api"

type MyAmbisonicPlugin struct {
    // ... your plugin fields ...
}

// IsAmbisonicConfigSupported checks if your plugin supports a given ambisonic configuration
func (p *MyAmbisonicPlugin) IsAmbisonicConfigSupported(config *api.AmbisonicConfig) bool {
    // Example: Support only ACN ordering with SN3D normalization
    return config.Ordering == api.AmbisonicOrderingACN && 
           config.Normalization == api.AmbisonicNormalizationSN3D
}

// GetAmbisonicConfig returns the ambisonic configuration for a specific port
func (p *MyAmbisonicPlugin) GetAmbisonicConfig(isInput bool, portIndex uint32) (*api.AmbisonicConfig, bool) {
    // Example: All ports use ACN/SN3D
    if portIndex == 0 {
        return &api.AmbisonicConfig{
            Ordering:      api.AmbisonicOrderingACN,
            Normalization: api.AmbisonicNormalizationSN3D,
        }, true
    }
    
    // Port doesn't support ambisonic
    return nil, false
}
```

### 2. Export the Required C Functions

In your plugin's main.go file, add these export functions:

```go
// #include <clap/clap.h>
import "C"
import (
    "runtime/cgo"
    "unsafe"
)

//export ClapGo_PluginAmbisonicIsConfigSupported
func ClapGo_PluginAmbisonicIsConfigSupported(plugin unsafe.Pointer, configPtr unsafe.Pointer) C.bool {
    if plugin == nil || configPtr == nil {
        return C.bool(false)
    }
    p := cgo.Handle(plugin).Value().(api.Plugin)
    
    provider, ok := p.(api.AmbisonicProvider)
    if !ok {
        return C.bool(false)
    }
    
    cConfig := (*C.clap_ambisonic_config_t)(configPtr)
    config := &api.AmbisonicConfig{
        Ordering:      uint32(cConfig.ordering),
        Normalization: uint32(cConfig.normalization),
    }
    
    return C.bool(provider.IsAmbisonicConfigSupported(config))
}

//export ClapGo_PluginAmbisonicGetConfig
func ClapGo_PluginAmbisonicGetConfig(plugin unsafe.Pointer, isInput C.bool, portIndex C.uint32_t, configPtr unsafe.Pointer) C.bool {
    if plugin == nil || configPtr == nil {
        return C.bool(false)
    }
    p := cgo.Handle(plugin).Value().(api.Plugin)
    
    provider, ok := p.(api.AmbisonicProvider)
    if !ok {
        return C.bool(false)
    }
    
    config, supported := provider.GetAmbisonicConfig(bool(isInput), uint32(portIndex))
    if !supported || config == nil {
        return C.bool(false)
    }
    
    cConfig := (*C.clap_ambisonic_config_t)(configPtr)
    cConfig.ordering = C.uint32_t(config.Ordering)
    cConfig.normalization = C.uint32_t(config.Normalization)
    
    return C.bool(true)
}
```

### 3. Notify Host of Configuration Changes (Optional)

If your plugin's ambisonic configuration can change (only when deactivated), notify the host:

```go
func (p *MyAmbisonicPlugin) Deactivate() {
    // ... perform deactivation ...
    
    // If ambisonic config changed, notify host
    if p.ambisonicConfigChanged {
        hostAmbisonic := api.NewHostAmbisonic(p.host, p)
        hostAmbisonic.Changed()
    }
}
```

## Channel Ordering Schemes

- **FuMa** (`AmbisonicOrderingFuMa`): Furse-Malham channel ordering
- **ACN** (`AmbisonicOrderingACN`): Ambisonics Channel Number ordering (recommended)

## Normalization Schemes

- **maxN** (`AmbisonicNormalizationMaxN`): Maximum normalized
- **SN3D** (`AmbisonicNormalizationSN3D`): Schmidt semi-normalized (recommended)
- **N3D** (`AmbisonicNormalizationN3D`): Fully normalized
- **SN2D** (`AmbisonicNormalizationSN2D`): Schmidt semi-normalized 2D
- **N2D** (`AmbisonicNormalizationN2D`): Fully normalized 2D

## Best Practices

1. **Consistency**: Use the same configuration for all ambisonic ports in your plugin
2. **Modern Standards**: Prefer ACN ordering with SN3D normalization for new plugins
3. **Documentation**: Clearly document which ambisonic formats your plugin supports
4. **Validation**: Always validate the configuration in `IsAmbisonicConfigSupported`

## Example: Basic Ambisonic Reverb

```go
type AmbisonicReverb struct {
    // ... reverb parameters ...
    host         *api.Host
    isActivated  bool
}

func (r *AmbisonicReverb) IsAmbisonicConfigSupported(config *api.AmbisonicConfig) bool {
    // Support both common formats
    if config.Ordering == api.AmbisonicOrderingACN {
        return config.Normalization == api.AmbisonicNormalizationSN3D ||
               config.Normalization == api.AmbisonicNormalizationN3D
    }
    return false
}

func (r *AmbisonicReverb) GetAmbisonicConfig(isInput bool, portIndex uint32) (*api.AmbisonicConfig, bool) {
    // First port is ambisonic
    if portIndex == 0 {
        return &api.AmbisonicConfig{
            Ordering:      api.AmbisonicOrderingACN,
            Normalization: api.AmbisonicNormalizationSN3D,
        }, true
    }
    
    // Other ports are not ambisonic
    return nil, false
}

func (r *AmbisonicReverb) GetAudioPortInfo(index uint32, isInput bool) api.AudioPortInfo {
    if index == 0 {
        return api.AudioPortInfo{
            ID:           0,
            Name:         "Ambisonic " + map[bool]string{true: "Input", false: "Output"}[isInput],
            ChannelCount: 16, // 3rd order ambisonic = 16 channels
            Flags:        api.AudioPortIsMain,
            PortType:     api.PortAmbisonic,
            InPlacePair:  0,
        }
    }
    // ... handle other ports ...
}
```

## Testing

To test your ambisonic implementation:

1. Load your plugin in a DAW that supports ambisonic audio
2. Verify the host correctly identifies your ambisonic ports
3. Process test signals to ensure proper channel mapping
4. Test configuration changes (if supported) while deactivated

## References

- [Ambisonic Wikipedia](https://en.wikipedia.org/wiki/Ambisonics)
- [ACN Channel Ordering](https://en.wikipedia.org/wiki/Ambisonic_data_exchange_formats#ACN)
- [SN3D Normalization](https://en.wikipedia.org/wiki/Ambisonic_data_exchange_formats#SN3D)