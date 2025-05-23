# Extension Architecture

## Problem

The current approach of returning dummy non-nil pointers from Go to indicate extension support is architecturally unsound. It creates a confusing separation of responsibilities between the C bridge and Go plugins.

## Current Issues

1. **Mixed Patterns**: Some extensions (params, audio-ports) are provided directly by C bridge, others (note-ports) check with Go first
2. **Dummy Pointers**: Using `C.get_non_nil_ptr()` is a hack that violates clean architecture
3. **Inconsistent Responsibility**: Unclear whether C bridge or Go plugin owns extension support decisions

## Proposed Solution

### Option 1: C Bridge Owns All Extensions (Recommended)

The C bridge provides all extension implementations and determines support based on:
- Manifest declarations
- Presence of required Go exports (checked at plugin load time)

```c
// In bridge.c during plugin creation
typedef struct go_plugin_data {
    void* go_instance;
    const clap_plugin_descriptor_t* descriptor;
    
    // Extension support flags determined at load time
    bool supports_params;      // Has ClapGo_PluginParamsCount, etc.
    bool supports_note_ports;  // Has ClapGo_PluginNotePortsCount, etc.
    bool supports_state;       // Has ClapGo_PluginStateSave, etc.
} go_plugin_data_t;

// During plugin creation
data->supports_note_ports = (ClapGo_PluginNotePortsCount != NULL && 
                            ClapGo_PluginNotePortsGet != NULL);

// In get_extension
if (strcmp(id, CLAP_EXT_NOTE_PORTS) == 0) {
    if (data->supports_note_ports) {
        return &s_note_ports_extension;
    }
    return NULL;
}
```

**Advantages**:
- Clean separation: C owns CLAP interface, Go owns implementation
- No dummy pointers or hacks
- Consistent pattern for all extensions
- Fast runtime checks (no Go calls needed)

### Option 2: Go Plugin Declares Support

Go plugins explicitly declare which extensions they support:

```go
// In plugin
func (p *Plugin) GetSupportedExtensions() []string {
    return []string{
        api.ExtParams,
        api.ExtNotePorts,  // Only if instrument
    }
}
```

**Disadvantages**:
- Requires additional Go export
- Can get out of sync with actual implementation
- Still needs runtime checks for function existence

### Option 3: Manifest-Driven Support

Use manifest as single source of truth:

```json
{
  "extensions": [
    {
      "id": "clap.note-ports",
      "supported": true,
      "required_exports": [
        "ClapGo_PluginNotePortsCount",
        "ClapGo_PluginNotePortsGet"
      ]
    }
  ]
}
```

**Disadvantages**:
- More complex manifest format
- Duplication of information
- Can still get out of sync

## Recommendation

**Option 1** (C Bridge Owns All Extensions) is the cleanest approach because:

1. **Single Responsibility**: C bridge handles all CLAP interface concerns
2. **No Hacks**: No dummy pointers or CGO violations
3. **Performance**: Extension support determined once at load time
4. **Consistency**: Same pattern for all extensions
5. **GUARDRAILS Compliance**: Complete implementation with no placeholders

## Implementation Steps

1. Remove `GetExtension` checks from C bridge for specific extensions
2. Add extension support flags to `go_plugin_data_t`
3. Set flags during plugin creation based on symbol presence
4. Update all `get_extension` checks to use flags
5. Remove dummy pointer hack from synth example
6. Update documentation

This approach maintains clean architecture while ensuring proper extension support.