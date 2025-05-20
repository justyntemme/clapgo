# ClapGo Implementation Roadmap

This roadmap outlines the remaining tasks needed to complete the ClapGo project, focusing specifically on getting both the gain and gain-with-gui plugins working properly. This is a proof-of-concept with some breaking changes allowed, but aimed at production-quality code.

## 1. Core Plugin Implementation ✅

**Priority: High (Foundation)**

- **Tasks:**
  - ✅ Fix plugin registration in `cmd/wrapper/main.go`
  - ✅ Implement proper CGO memory handling using cgo.Handle
  - ✅ Fix plugin discovery in the C plugin layer
  - ✅ Ensure basic plugin loading works

- **Current Status:**
  - ✅ Plugin registration works properly
  - ✅ The gain plugin passes basic validation tests
  - ✅ Memory safety is ensured using proper CGO patterns
  - ✅ The descriptor creation and handling is functional

## 2. Audio Processing ✅

**Priority: High (Essential functionality)**

- **Tasks:**
  - ✅ Implement proper audio buffer conversion between C and Go
  - ✅ Add proper audio channel handling for stereo and mono configurations
  - ✅ Implement gain processing with parameter control
  - ✅ Add audio port configuration to properly expose inputs/outputs

- **Current Status:**
  - ✅ Audio buffer handling is implemented with proper safety checks
  - ✅ The audio-ports extension is properly implemented
  - ✅ Basic audio processing is working in the gain plugin
  - ✅ All audio processing tests are passing successfully

## 3. Parameter Support ✅

**Priority: High (Essential functionality)**

- **Tasks:**
  - ✅ Implement the 'clap.params' extension
  - ✅ Add parameter registration for gain control
  - ✅ Implement parameter value change handling in processing
  - ✅ Add parameter value to string conversion for UI display
  - ✅ Implement parameter flush mechanism

- **Current Status:**
  - ✅ Full implementation of parameter extension with proper C/Go bridging
  - ✅ Proper parameter event handling for automation
  - ✅ Parameter value conversion for UI display
  - ✅ Parameter manager implementation with normalize/denormalize support
  - ✅ Parameter gesture begin/end support added for UI interaction

## 4. State Management ✅

**Priority: Medium (Required for persistence)**

- **Tasks:**
  - ✅ Implement the 'clap.state' extension
  - ✅ Add serialization/deserialization of plugin state
  - ✅ Ensure state reproducibility across plugin instances
  - ✅ Handle buffered stream state loading/saving

- **Current Status:**
  - ✅ State extension is fully implemented 
  - ✅ JSON-based state serialization for parameters and plugin-specific data
  - ✅ Support for the Stater interface for plugin-specific state
  - ✅ Stream reading/writing with proper error handling

## 5. GUI Integration ✅

**Priority: Medium (For gain-with-gui plugin)**

- **Tasks:**
  - ✅ Complete the GUI bridge implementation in `examples/gain-with-gui/gui_bridge.cpp`
  - ✅ Implement the CLAP GUI extension
  - ✅ Connect GUI to plugin parameters
  - ✅ Add window management across platforms (Windows, macOS, Linux)

- **Current Status:**
  - ✅ Full implementation of GUI bridge using clap-plugins GUI framework
  - ✅ GUI extension in Go with proper callbacks between C++ and Go code
  - ✅ Parameter bridge with QML for UI controls
  - ✅ Unified window management with platform-specific adapters (X11, Win32, Cocoa)
  - ✅ Added GUIProvider interface for Go plugins to implement GUI features

## 6. Event Handling ⏳

**Priority: Medium (Required for full plugin functionality)**

- **Tasks:**
  - Implement proper event handling for parameter changes
  - Add support for transport events
  - Handle note events for the synth plugin

- **Current Status:**
  - Basic event structure is defined
  - Parameter events need proper implementation
  - Note events are not yet handled

## 7. Build System Enhancements ⏳

**Priority: Medium (Required for distribution)**

- **Tasks:**
  - Ensure proper shared library loading on all platforms
  - Add packaging for plugin distribution
  - Ensure libgoclap.so is properly installed with plugins
  - Add version verification between components

- **Current Status:**
  - Basic build system is in place
  - Shared library is being copied to ~/.clap folder
  - Need to formalize installation process for various platforms

## 8. Advanced Extensions ⏳

**Priority: Low (Nice to have)**

- **Tasks:**
  - Implement note ports for synth plugin
  - Add preset system and preset discovery
  - Implement thread pool support
  - Add MIDI support

- **Current Status:**
  - Extension stubs are defined
  - Many extension tests are being skipped
  - Need to prioritize which extensions to implement first

## Next Steps

To get both plugins (gain and gain-with-gui) working properly, the highest priorities are:

1. ✅ **Complete Audio Processing**: Implement proper audio buffer handling and processing
2. ✅ **Implement Parameter Support**: Add full parameter support with automation
3. ✅ **Add State Management**: Implement basic state save/load functionality
4. ✅ **Integrate GUI for gain-with-gui**: Connect the GUI bridge to parameters

All core functionality has been implemented! The next steps could be to improve event handling and implement additional extensions to enhance the plugin functionality.

By completing these core tasks, we have achieved a functional proof-of-concept that demonstrates the viability of Go-based CLAP plugins with a path toward a production-ready implementation.