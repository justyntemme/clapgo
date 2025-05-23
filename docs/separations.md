# Extension Separation Strategy

This document outlines the proper approach for implementing CLAP extensions in ClapGo plugins, ensuring clean separation between plugin types and preventing crashes from unsupported extensions.

## Core Principle: Explicit Extension Support

Following GUARDRAILS.md principles:
- **Complete implementation or explicit non-support** - No placeholders or partial implementations
- **Manifest-driven architecture** - Extensions declared in manifest must match implementation
- **Real CLAP compliance** - Extensions work correctly or aren't advertised

## Extension Implementation Pattern

### 1. Plugin-Side Implementation

Plugins must only export functions for extensions they actually support:

```go
// CORRECT: Audio effect only exports audio-related extensions
// gain/main.go

//export ClapGo_PluginParamsCount
//export ClapGo_PluginParamsGetInfo
// ... other param exports

// NO note ports exports - gain doesn't process MIDI

// CORRECT: Instrument exports note port extensions
// synth/main.go

//export ClapGo_PluginNotePortsCount
//export ClapGo_PluginNotePortsGet
// ... in addition to param exports
```

### 2. C Bridge Conditional Calls

The C bridge must check if Go exports exist before calling them:

```c
// In bridge.c - Check if symbol exists before calling
static uint32_t clapgo_note_ports_count(const clap_plugin_t* plugin, bool is_input) {
    go_plugin_data_t* data = (go_plugin_data_t*)plugin->plugin_data;
    if (!data || !data->go_instance) return 0;
    
    // Check if the Go plugin exports this function
    if (!data->supports_note_ports) {
        return 0;
    }
    
    return ClapGo_PluginNotePortsCount(data->go_instance, is_input);
}
```

### 3. Manifest Declaration

Extensions must be accurately declared in the manifest:

```json
// gain.json - Audio effect
{
  "extensions": [
    {
      "id": "clap.params",
      "supported": true
    },
    {
      "id": "clap.audio-ports",
      "supported": true
    }
    // NO note-ports entry
  ]
}

// synth.json - Instrument
{
  "extensions": [
    {
      "id": "clap.params",
      "supported": true
    },
    {
      "id": "clap.audio-ports", 
      "supported": true
    },
    {
      "id": "clap.note-ports",
      "supported": true
    }
  ]
}
```

## Plugin Categories and Required Extensions

### Audio Effects (gain, compressor, reverb, etc.)
- **Required**: `clap.audio-ports`, `clap.params`
- **Optional**: `clap.state`, `clap.latency`, `clap.tail`
- **Never**: `clap.note-ports` (unless MIDI-controlled parameters)

### Instruments (synth, sampler, etc.)
- **Required**: `clap.audio-ports`, `clap.note-ports`, `clap.params`
- **Optional**: `clap.state`, `clap.voice-info`, `clap.note-name`
- **Always**: Must handle note events

### MIDI Effects (arpeggiator, chord generator, etc.)
- **Required**: `clap.note-ports` (input and output)
- **Optional**: `clap.params`, `clap.state`
- **Sometimes**: `clap.audio-ports` (if audio-triggered)

## Implementation Strategy

### Step 1: Update C Bridge (Immediate Fix)

Add runtime detection of exported symbols:

```c
// In bridge.c init
typedef struct {
    // ... existing fields
    bool has_note_ports;
    bool has_voice_info;
    // ... other extension flags
} go_plugin_data_t;

// During plugin creation
data->has_note_ports = (dlsym(data->go_handle, "ClapGo_PluginNotePortsCount") != NULL);
```

### Step 2: GetExtension Pattern

Plugins return nil for unsupported extensions:

```go
func (p *GainPlugin) GetExtension(id string) unsafe.Pointer {
    switch id {
    case api.ExtParams:
        return unsafe.Pointer(p) // Supported
    case api.ExtNotePorts:
        return nil // Not supported - audio effect
    default:
        return nil
    }
}
```

### Step 3: Conditional Exports

Use build tags or separate files for extension-specific exports:

```go
// note_ports.go (only included in instrument plugins)
// +build instrument

//export ClapGo_PluginNotePortsCount
func ClapGo_PluginNotePortsCount(...) { }
```

## Safety Guarantees

1. **No Crashes**: Plugins without note ports won't crash DAWs
2. **Clear Intent**: Manifest declares what's actually supported
3. **Type Safety**: Plugin categories determine available extensions
4. **GUARDRAILS Compliance**: Complete implementation or explicit non-support

## Testing Strategy

1. Test each plugin type in isolation
2. Verify DAW doesn't call unsupported extensions
3. Ensure manifest matches actual implementation
4. Check that missing exports return appropriate defaults

This approach ensures clean separation of concerns while maintaining full CLAP compliance and preventing crashes from unsupported extensions.