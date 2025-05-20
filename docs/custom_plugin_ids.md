# Custom Plugin IDs in ClapGo

This document explains how to use custom plugin IDs with ClapGo.

## Overview

Every CLAP plugin requires a unique identifier (ID) to distinguish it from other plugins. By convention, CLAP plugin IDs follow a reverse domain name pattern (e.g., `com.company.plugin`).

ClapGo allows you to set custom plugin IDs for your plugins, ensuring no conflicts with other plugins and establishing your own plugin namespace.

## Setting Custom Plugin IDs

### In Example Plugins

Each plugin you create should have a unique ID. When registering your plugin, use the `PluginInfo` struct to set a custom ID:

```go
info := goclap.PluginInfo{
    ID:          "com.yourcompany.yourplugin", // Your custom plugin ID
    Name:        "Your Plugin Name",
    Vendor:      "Your Company",
    // Other info fields...
}

// Register your plugin
yourPlugin := NewYourPlugin()
goclap.RegisterPlugin(info, yourPlugin)
```

### Using the Plugin Creator

When creating a new plugin with the `create_plugin.sh` script, you can provide a custom plugin ID as the second argument:

```bash
./create_plugin.sh myplugin com.yourcompany.myplugin
```

If you don't provide a custom ID, it will default to `com.clapgo.pluginname`.

### For the Default Wrapper

The default wrapper plugin included in ClapGo can be configured with a custom ID by setting the `CLAPGO_PLUGIN_ID` environment variable:

```bash
CLAPGO_PLUGIN_ID=com.yourcompany.yourplugin ./your_program
```

## Best Practices

1. **Use Reverse Domain Notation**: Follow the standard practice of using reverse domain notation for your plugin IDs (e.g., `com.yourcompany.yourplugin`).

2. **Ensure Uniqueness**: Make sure your plugin IDs are unique to avoid conflicts with other plugins.

3. **Be Consistent**: Use the same namespace for all your plugins (e.g., all your plugins should start with `com.yourcompany`).

4. **Avoid Special Characters**: Stick to alphanumeric characters, dots, and hyphens in your plugin IDs.

5. **Add Version Information if Needed**: For significantly different versions of the same plugin, you might want to include version information in the ID (e.g., `com.yourcompany.yourplugin-v2`).

## Implementation Notes

The ClapGo library doesn't impose any restrictions on plugin IDs. The ID you provide in the `PluginInfo` struct is passed directly to the CLAP host. 

The only requirement from the CLAP specification is that plugin IDs must be unique and non-empty.