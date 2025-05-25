# Note Name Extension

The Note Name Extension allows plugins to provide custom names for MIDI notes. This is particularly useful for drum plugins, samplers, or any instrument that assigns specific sounds to individual keys.

## Overview

The extension enables plugins to:
- Provide custom names for specific MIDI notes
- Specify names per port, key, and channel
- Use wildcard values (-1) to apply names broadly

## Implementation

### 1. Add Extension Support to Your Plugin

In your plugin's manifest (`.json` file), declare support for the note name extension:

```json
{
  "extensions": [
    {
      "id": "clap.note-name",
      "supported": true
    }
  ]
}
```

### 2. Export the Required Functions

Add these exports to your plugin:

```go
//export ClapGo_PluginNoteNameCount
func ClapGo_PluginNoteNameCount(plugin unsafe.Pointer) C.uint32_t {
    if plugin == nil {
        return 0
    }
    p := cgo.Handle(plugin).Value().(*YourPlugin)
    
    // Return the number of note names your plugin provides
    return C.uint32_t(len(noteNames))
}

//export ClapGo_PluginNoteNameGet
func ClapGo_PluginNoteNameGet(plugin unsafe.Pointer, index C.uint32_t, noteName unsafe.Pointer) C.bool {
    if plugin == nil || noteName == nil {
        return C.bool(false)
    }
    p := cgo.Handle(plugin).Value().(*YourPlugin)
    
    // Get the note name at the specified index
    if int(index) >= len(noteNames) {
        return C.bool(false)
    }
    
    // Convert to C structure
    api.NoteNameToC(&noteNames[index], noteName)
    
    return C.bool(true)
}
```

### 3. Define Your Note Names

ClapGo provides several helper functions and predefined note name sets:

#### Standard Note Names (C-2 to G8)
```go
noteNames := api.StandardNoteNames()
```

#### General MIDI Drum Names
```go
drumNames := api.StandardDrumNoteNames
```

#### Custom Note Names
```go
customNames := []api.NoteName{
    {Name: "Kick", Port: -1, Key: 36, Channel: -1},
    {Name: "Snare", Port: -1, Key: 38, Channel: -1},
    {Name: "Hi-Hat", Port: -1, Key: 42, Channel: -1},
    // ... more names
}
```

## NoteName Structure

The `NoteName` structure contains:
- `Name`: The display name for the note (up to 256 characters)
- `Port`: The port index, or -1 for all ports
- `Key`: The MIDI key number (0-127), or -1 for all keys
- `Channel`: The MIDI channel (0-15), or -1 for all channels

## Example: Drum Plugin

```go
type DrumPlugin struct {
    // ... other fields
    noteNames []api.NoteName
}

func NewDrumPlugin() *DrumPlugin {
    return &DrumPlugin{
        noteNames: api.StandardDrumNoteNames,
    }
}

//export ClapGo_PluginNoteNameCount
func ClapGo_PluginNoteNameCount(plugin unsafe.Pointer) C.uint32_t {
    if plugin == nil {
        return 0
    }
    p := cgo.Handle(plugin).Value().(*DrumPlugin)
    return C.uint32_t(len(p.noteNames))
}

//export ClapGo_PluginNoteNameGet
func ClapGo_PluginNoteNameGet(plugin unsafe.Pointer, index C.uint32_t, noteName unsafe.Pointer) C.bool {
    if plugin == nil || noteName == nil || index >= C.uint32_t(len(p.noteNames)) {
        return C.bool(false)
    }
    p := cgo.Handle(plugin).Value().(*DrumPlugin)
    
    api.NoteNameToC(&p.noteNames[index], noteName)
    return C.bool(true)
}
```

## Host Notification

If your plugin dynamically changes note names, you can notify the host:

```go
func (p *YourPlugin) UpdateNoteNames() {
    // Update your note names...
    
    // Notify the host
    if p.host != nil {
        hostNoteName := api.NewHostNoteName(p.host)
        if hostNoteName != nil {
            hostNoteName.NotifyChanged()
        }
    }
}
```

## Best Practices

1. **Use Standard Names When Appropriate**: For general instruments, use `api.StandardNoteNames()` to provide familiar note names.

2. **Be Specific with Channels**: For drum plugins, specify channel 9 (MIDI channel 10) if following General MIDI conventions.

3. **Consider Multiple Ports**: If your plugin has multiple input ports with different instruments, use the port field to provide port-specific names.

4. **Keep Names Concise**: While you have up to 256 characters, shorter names are more practical in most DAW interfaces.

5. **Update Dynamically**: If your plugin loads different sample sets or instruments, update the note names and notify the host.

## Testing

Use the provided test script to verify your note name implementation:

```bash
./test_note_name.sh
```

This will check that:
- Your manifest declares note name support
- The required export symbols are present
- The extension is properly registered

## Troubleshooting

1. **Names Not Showing**: Ensure your manifest declares `"supported": true` for the note name extension.

2. **Crashes**: Verify bounds checking in `ClapGo_PluginNoteNameGet` - never access beyond your array bounds.

3. **Wrong Names**: Check that you're using the correct port/channel values. Use -1 for "all".

4. **Not Updating**: After changing names dynamically, always call `NotifyChanged()` on the host extension.