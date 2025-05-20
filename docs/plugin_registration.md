# Plugin Registration Guide

This document outlines the recommended approach for registering plugins in the ClapGo framework.

## Overview

ClapGo provides a flexible plugin registration system that allows plugin developers to register their plugins with the framework in a clean and modular way. The registration system is designed to:

1. Decouple plugins from the core bridge code
2. Allow plugins to be registered independently
3. Support easy discovery of available plugins
4. Provide a clean API for plugin registration

## Registration Options

There are two main approaches to registering a plugin:

### 1. Direct Registration

This approach is straightforward and requires minimal boilerplate:

```go
// Initialize your plugin
myPlugin := NewMyPlugin()

// Register it with the registry
registry.Register(myPlugin.GetPluginInfo(), func() api.Plugin {
    return myPlugin
})
```

### 2. Using the PluginProvider Interface

This approach is more structured and separates plugin creation from registration:

```go
// Define a provider that implements the api.PluginProvider interface
type MyPluginProvider struct{}

func (p *MyPluginProvider) CreatePlugin() api.Plugin {
    return NewMyPlugin()
}

func (p *MyPluginProvider) GetPluginInfo() api.PluginInfo {
    return api.PluginInfo{
        ID:          "com.example.my-plugin",
        Name:        "My Plugin",
        Vendor:      "Example",
        // ... other fields
    }
}

// In your init function:
func init() {
    provider := &MyPluginProvider{}
    registry.RegisterPlugin(provider)
}
```

## Recommended Pattern

The recommended pattern is to use the `PluginProvider` interface, as it:

1. Separates creation logic from plugin implementation
2. Provides a clean, consistent registration approach
3. Makes it easier to extend the registration process in the future
4. Reduces code duplication

## Implementation Steps

### 1. Create Your Plugin

Implement the `api.Plugin` interface:

```go
type MyPlugin struct {
    // Plugin state
}

// Implement all required methods from the api.Plugin interface
func (p *MyPlugin) Init() bool {
    // ...
}

// ... other required methods
```

### 2. Create a Provider

Implement the `api.PluginProvider` interface:

```go
type MyPluginProvider struct{}

func (p *MyPluginProvider) CreatePlugin() api.Plugin {
    return NewMyPlugin()
}

func (p *MyPluginProvider) GetPluginInfo() api.PluginInfo {
    return api.PluginInfo{
        ID:          "com.example.my-plugin",
        Name:        "My Plugin",
        // ... other fields
    }
}
```

### 3. Register in init()

Register your plugin during package initialization:

```go
func init() {
    provider := &MyPluginProvider{}
    registry.RegisterPlugin(provider)
}
```

## Available API Functions

The registry package provides several functions for plugin registration and discovery:

- `registry.Register(info api.PluginInfo, creator func() api.Plugin)`: Register a plugin directly
- `registry.RegisterPlugin(provider api.PluginProvider)`: Register a plugin using a provider
- `registry.GetPluginCount()`: Get the number of registered plugins
- `registry.GetPluginInfo(index uint32)`: Get information about a plugin by index
- `registry.GetPluginIDs()`: Get a list of all registered plugin IDs
- `registry.CreatePlugin(id string)`: Create a new instance of a plugin by ID

## Examples

See the example plugins in the `examples/` directory:

- `examples/gain/`: A simple gain plugin example
- `examples/synth/`: A simple synthesizer plugin example

Each example demonstrates proper plugin registration techniques.