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

## 6. Event Handling ✅

**Priority: Medium (Required for full plugin functionality)**

- **Tasks:**
  - ✅ Implement proper event handling for parameter changes
  - ✅ Add support for transport events
  - ✅ Handle note events for the synth plugin
  - ✅ Fix crashers in note event handling
  - ✅ Make event handling more robust

- **Current Status:**
  - ✅ Complete event handling structure with proper C/Go bridging
  - ✅ Parameter events fully implemented with proper type conversion 
  - ✅ Transport events implemented with host state tracking
  - ✅ Note events framework implemented for instrument plugins
  - ✅ Proper event queuing and processing in the audio thread
  - ✅ Implemented proper event handling with full C/Go type conversion
  - ✅ Fixed crashes in complex note event processing scenarios

## 7. Build System Enhancements ✅

**Priority: Medium (Required for distribution)**

- **Tasks:**
  - ✅ Ensure proper shared library loading on all platforms
  - ✅ Add packaging for plugin distribution
  - ✅ Ensure libgoclap.so is properly installed with plugins
  - ✅ Add version verification between components

- **Current Status:**
  - ✅ Robust build system is in place with CMake integration
  - ✅ Shared library is properly installed to ~/.clap folder
  - ✅ Installation process works correctly on Linux
  - ✅ Version verification between Go and C components is implemented
  - ✅ Plugin discovery and loading is working correctly

## 8. Advanced Extensions ⚡

**Priority: Low (Nice to have)**

- **Tasks:**
  - ✅ Implement note ports for synth plugin
  - ✅ Add preset system and preset discovery
  - ⚡ Complete preset discovery factory implementation
  - Implement thread pool support
  - Add MIDI support

- **Current Status:**
  - ✅ Note ports extension is fully implemented
  - ✅ Basic note event handling structure is in place
  - ✅ Synth plugin now reports proper note port configuration
  - ✅ Preset loading is implemented with JSON-based preset files
  - ⚡ Preset loading works but discovery factory needs completion
  - ⚡ Synth plugin demonstrates preset loading from filesystem
  - Extension stubs are defined for other advanced features
  - Still need to implement thread pool support and MIDI support

## Next Steps

To get both plugins (gain and gain-with-gui) working properly, the highest priorities are:

1. ✅ **Complete Audio Processing**: Implement proper audio buffer handling and processing
2. ✅ **Implement Parameter Support**: Add full parameter support with automation
3. ✅ **Add State Management**: Implement basic state save/load functionality
4. ✅ **Integrate GUI for gain-with-gui**: Connect the GUI bridge to parameters
5. ✅ **Implement Event Handling**: Add proper parameter and transport event handling
6. ✅ **Enhance Build System**: Ensure proper installation and distribution

All high-priority functionality has been implemented! The project now has:

1. ✅ **Core Plugin Framework**: Plugin registration, discovery, and lifetime management
2. ✅ **Audio Processing**: Proper buffer handling and stereo/mono support
3. ✅ **Parameter System**: Parameter registration, automation, and UI connection
4. ✅ **State Management**: Save/load functionality for plugin settings
5. ✅ **GUI Integration**: Working GUI with parameter control
6. ✅ **Event Handling**: Support for parameter, note, and transport events
7. ✅ **Build System**: Reliable build, installation, and packaging
8. ⚡ **Preset System**: Basic preset loading is working (discovery in progress)
9. ✅ **Note Port Handling**: Synth plugin with note event processing

The next steps are:

1. ✅ **Fix event handling crashes**: Complete the robust handling of event structures
   - ✅ Fix the crashes in `process-note-inconsistent` and `process-note-out-of-place-basic` tests
   - ✅ Implement proper memory safety for event processing
   - ✅ Add complete bidirectional event conversion

2. ⚡ **Complete preset discovery factory**: Finish the preset discovery factory implementation
   - Implement proper preset discovery factory creation
   - Add preset location scanning and categorization
   - Ensure proper preset metadata handling

3. **Thread Pool Support**: Implement the thread pool extension
   - Add worker thread pool for non-realtime tasks
   - Implement proper job queuing and execution
   - Add cancellation support for jobs

4. **Enhanced MIDI Support**: Add proper MIDI event handling
   - Implement MIDI mapping for parameters
   - Add MIDI output capabilities
   - Support MIDI CC automation

By completing these tasks, we will achieve a fully functional proof-of-concept that demonstrates the viability of Go-based CLAP plugins with a path toward a production-ready implementation.