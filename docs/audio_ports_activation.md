# Audio Ports Activation Extension

The Audio Ports Activation extension provides a way for hosts to activate and deactivate audio ports dynamically. This is an advanced feature that allows for optimized processing paths and resource management.

## Overview

This extension enables:
- Hosts to inform plugins which audio ports are actually being used
- Plugins to optimize their processing when certain inputs/outputs are not needed
- Dynamic activation/deactivation during processing (if supported by the plugin)

## When to Use This Extension

Consider implementing this extension if your plugin:
- Has multiple audio inputs/outputs that may not all be used simultaneously
- Can optimize processing when certain ports are inactive
- Supports complex routing scenarios where port usage varies
- Implements sidechain inputs that may not always be connected
- Has auxiliary outputs that may not always be consumed

## Implementation Guide

### 1. Implement the Interface

```go
type MyPlugin struct {
    // ... other fields ...
    portStates *api.AudioPortActivationState
}

func (p *MyPlugin) CanActivateWhileProcessing() bool {
    // Return true only if your plugin can safely handle
    // port activation changes during audio processing
    return false // Most plugins should return false
}

func (p *MyPlugin) SetActive(isInput bool, portIndex uint32, isActive bool, sampleSize uint32) bool {
    // Validate the port index
    if isInput && portIndex >= p.GetAudioPortCount(true) {
        return false
    }
    if !isInput && portIndex >= p.GetAudioPortCount(false) {
        return false
    }
    
    // Update the activation state
    p.portStates.SetPortActive(isInput, portIndex, isActive)
    
    // Optionally handle sample size hint
    // sampleSize can be 32, 64, or 0 (unspecified)
    
    return true
}
```

### 2. Check Port States During Processing

```go
func (p *MyPlugin) Process(process *api.Process) api.ProcessStatus {
    // Check if sidechain input is active
    if p.portStates.IsPortActive(true, 1) { // Assuming port 1 is sidechain
        // Process with sidechain
        p.processWithSidechain(process)
    } else {
        // Process without sidechain (optimized path)
        p.processNormal(process)
    }
    
    // Check if auxiliary output is active
    if !p.portStates.IsPortActive(false, 1) { // Assuming port 1 is aux out
        // Skip computing auxiliary output
    }
    
    return api.ProcessContinue
}
```

## Important Considerations

### Thread Safety

- `CanActivateWhileProcessing()` is called on the main thread
- `SetActive()` is called on the audio thread if `CanActivateWhileProcessing()` returns true, otherwise on the main thread
- If you support activation while processing, ensure thread-safe state management

### Buffer Requirements

Even when a port is deactivated:
- The host must still provide valid audio buffers
- These buffers should be filled with zeros
- The constant_mask should be set to indicate silence

### State Management

- Audio port activation states are NOT saved with plugin state
- Ports are active by default when the plugin is created
- The host is responsible for restoring activation states after loading a project

### Compatibility

- The extension uses ID `"clap.audio-ports-activation/2"`
- For compatibility, also check for `"clap.audio-ports-activation/draft-2"`
- The activation state is invalidated when audio port configuration changes

## Example: Synthesizer with Multiple Outputs

```go
type MultiOutputSynth struct {
    api.BasePlugin
    portStates *api.AudioPortActivationState
    
    // Processing components
    mainSynth     *SynthEngine
    subOscillator *SubOscEngine
    effects       *EffectsChain
}

func (s *MultiOutputSynth) SetActive(isInput bool, portIndex uint32, isActive bool, sampleSize uint32) bool {
    if isInput {
        return false // This synth has no audio inputs
    }
    
    // Validate output port index (0: main, 1: sub, 2: effects send)
    if portIndex > 2 {
        return false
    }
    
    s.portStates.SetPortActive(false, portIndex, isActive)
    
    // Optimize processing based on active outputs
    switch portIndex {
    case 1: // Sub output
        s.subOscillator.SetEnabled(isActive)
    case 2: // Effects send
        s.effects.SetSendEnabled(isActive)
    }
    
    return true
}

func (s *MultiOutputSynth) Process(process *api.Process) api.ProcessStatus {
    // Always process main output
    s.mainSynth.Process(process.AudioOutputs[0])
    
    // Only process sub if output is active
    if s.portStates.IsPortActive(false, 1) {
        s.subOscillator.Process(process.AudioOutputs[1])
    }
    
    // Only process effects send if active
    if s.portStates.IsPortActive(false, 2) {
        s.effects.ProcessSend(process.AudioOutputs[2])
    }
    
    return api.ProcessContinue
}
```

## Best Practices

1. **Default to Safe**: Return false from `CanActivateWhileProcessing()` unless you have specific thread-safety measures
2. **Validate Indices**: Always check port indices are within valid ranges
3. **Optimize Wisely**: Only skip processing that actually saves significant CPU
4. **Handle Gracefully**: Continue to work correctly even if all ports are active
5. **Document Behavior**: Clearly document which ports can be deactivated and what the effect is

## Testing

When testing your implementation:
1. Verify the plugin works with all ports active (default state)
2. Test deactivating each port individually
3. Test various combinations of active/inactive ports
4. If supporting runtime changes, test activation during playback
5. Verify CPU usage actually decreases when ports are deactivated