cmake_minimum_required(VERSION 3.17)

# Check if Qt6 is available
if(CLAPGO_GUI_SUPPORT)
    find_package(Qt6 COMPONENTS Gui Qml Quick QUIET)
    
    if(NOT Qt6_FOUND)
        message(WARNING "Qt6 not found. Building without GUI support.")
        set(CLAPGO_GUI_SUPPORT OFF)
    else()
        message(STATUS "Qt6 found. Building with GUI support.")
    endif()
else()
    message(STATUS "GUI support disabled.")
endif()

# Build the Go plugin code first
add_go_library(gain-with-gui-go
    SOURCES ${CMAKE_CURRENT_SOURCE_DIR}/main.go
    OUTPUT_NAME gain-with-gui
)

# Add QML resources if GUI support is enabled
if(CLAPGO_GUI_SUPPORT)
    add_subdirectory(qml)
endif()

# Create the CLAP plugin
if(CLAPGO_GUI_SUPPORT)
    # Link with GUI support
    add_clap_plugin(gain-with-gui
        SOURCES 
            ${CMAKE_SOURCE_DIR}/src/c/plugin.c
            ${CMAKE_CURRENT_SOURCE_DIR}/gui_bridge.cpp
        LINK_LIBRARIES 
            gain-with-gui-go
            clap-plugin-gui
            clap-plugin-local-gui
            clapgo-gain-gui-qml
    )
    
    # Set C++20 standard for GUI code
    set_property(TARGET gain-with-gui PROPERTY CXX_STANDARD 20)
    
    # Add GUI-specific definitions
    target_compile_definitions(gain-with-gui PRIVATE CLAPGO_GUI_SUPPORT)
    
    # Include directories for GUI support
    target_include_directories(gain-with-gui PRIVATE
        ${CMAKE_SOURCE_DIR}/clap-plugins/plugins
        ${CMAKE_SOURCE_DIR}/clap-plugins/plugins/gui
    )
else()
    # Link without GUI support
    add_clap_plugin(gain-with-gui
        SOURCES ${CMAKE_SOURCE_DIR}/src/c/plugin.c
        LINK_LIBRARIES gain-with-gui-go
    )
endif()