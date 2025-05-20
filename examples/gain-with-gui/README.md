# GUI Bridge for CLAP Go Plugins

This directory contains a GUI implementation that bridges between Go CLAP plugins and QML interfaces.

## Why C Instead of C++

While our GUI uses QML (a Qt technology), the bridge between Go and the GUI is implemented in C rather than C++. This choice was made for several reasons:

1. **CGo Compatibility**: Go's foreign function interface (CGo) is designed to work directly with C, not C++. Using C avoids name mangling issues and complex C++ exception handling that can cause problems at language boundaries.

2. **ABI Stability**: C has a more stable Application Binary Interface (ABI) compared to C++, which reduces the chances of binary incompatibility between different compiler versions.

3. **Simpler Integration**: C provides a simpler, more direct interface to the CLAP API which itself is defined in C. This makes the integration cleaner and easier to maintain.

4. **Reduced Complexity**: Using C eliminates the need to deal with complex C++ features like templates and exceptions that can cause interoperability issues with Go.

5. **Performance**: The C bridge serves as a thin layer between Go and Qt/QML, minimizing overhead in the critical path of audio processing.

## Implementation Details

The `gui_bridge.cpp` file, despite its extension, primarily uses C-style code and extern "C" declarations to ensure compatibility with Go. The QML interface is loaded and managed through this bridge, allowing the Go plugin to control UI parameters and receive user input.

## Usage

To use this GUI bridge in your own Go CLAP plugin:

1. Include the bridge code in your project
2. Create QML files for your user interface
3. Connect the GUI and your plugin using the bridge functions
4. Build with the CMake configuration provided

See the implementation in this directory for a working example.