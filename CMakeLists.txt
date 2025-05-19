cmake_minimum_required(VERSION 3.17)
cmake_policy(SET CMP0091 NEW) # use CMAKE_MSVC_RUNTIME_LIBRARY

project(CLAPGO VERSION 0.1.0 LANGUAGES C CXX)

# Build settings
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_MSVC_RUNTIME_LIBRARY "MultiThreaded$<$<CONFIG:Debug>:Debug>DLL")
set(CMAKE_POSITION_INDEPENDENT_CODE ON)

if(APPLE)
    set(CMAKE_OSX_DEPLOYMENT_TARGET 12.0)
endif()

# Options
option(CLAPGO_BUILD_EXAMPLES "Build example plugins" ON)
option(CLAPGO_INSTALL_PLUGINS "Install plugins to system directories" OFF)

# Add CLAP submodule
add_subdirectory(include/clap EXCLUDE_FROM_ALL)

# Add CLAP plugins for GUI support
option(CLAPGO_GUI_SUPPORT "Build with GUI support using clap-plugins" OFF)
if(CLAPGO_GUI_SUPPORT)
    # Check if clap-plugins is properly set up as a submodule
    if(EXISTS "${CMAKE_SOURCE_DIR}/clap-plugins/CMakeLists.txt" AND
       EXISTS "${CMAKE_SOURCE_DIR}/clap-plugins/clap/CMakeLists.txt" AND
       EXISTS "${CMAKE_SOURCE_DIR}/clap-plugins/clap-helpers/CMakeLists.txt")
        
        # Add clap-plugins as a dependency for GUI support
        add_subdirectory(clap-plugins EXCLUDE_FROM_ALL)
        
        # Disable tests in clap-plugins
        set(CLAP_BUILD_TESTS OFF CACHE BOOL "" FORCE)
        
        # Force GUI support
        set(CLAP_PLUGINS_HEADLESS OFF CACHE BOOL "" FORCE)
        set(CLAP_PLUGINS_EMBED_QML ON CACHE BOOL "" FORCE)
        
        message(STATUS "GUI support enabled with clap-plugins")
    else()
        message(WARNING "clap-plugins directories not found, disabling GUI support")
        set(CLAPGO_GUI_SUPPORT OFF CACHE BOOL "" FORCE)
    endif()
endif()

# Include Go build helper
include(cmake/GoLibrary.cmake)

# Add our C wrapper library
add_library(clapgo-wrapper STATIC
    src/c/plugin.c
)

target_include_directories(clapgo-wrapper PUBLIC
    ${CMAKE_SOURCE_DIR}/include/clap/include
)

# Export header for including
set_target_properties(clapgo-wrapper PROPERTIES
    PUBLIC_HEADER src/c/plugin.h
)

# Example plugins
if(CLAPGO_BUILD_EXAMPLES)
    # Add each example subdirectory
    file(GLOB EXAMPLE_DIRS ${CMAKE_SOURCE_DIR}/examples/*)
    foreach(EXAMPLE_DIR ${EXAMPLE_DIRS})
        if(IS_DIRECTORY ${EXAMPLE_DIR})
            get_filename_component(PLUGIN_NAME ${EXAMPLE_DIR} NAME)
            message(STATUS "Adding example plugin: ${PLUGIN_NAME}")
            
            if(EXISTS "${EXAMPLE_DIR}/CMakeLists.txt")
                add_subdirectory(${EXAMPLE_DIR})
            else()
                # For directories without their own CMakeLists.txt
                # Build Go shared library
                add_go_library(${PLUGIN_NAME}-go
                    SOURCES ${EXAMPLE_DIR}/main.go
                    OUTPUT_NAME ${PLUGIN_NAME}
                )
                
                # Build CLAP plugin
                add_clap_plugin(${PLUGIN_NAME}
                    SOURCES ${CMAKE_SOURCE_DIR}/src/c/plugin.c
                    LINK_LIBRARIES ${PLUGIN_NAME}-go
                )
            endif()
        endif()
    endforeach()
endif()

# Installation options
option(CLAPGO_INSTALL_PLUGINS "Install plugins to system or user directories" ON)
option(CLAPGO_INSTALL_SYSTEM_WIDE "Install plugins system-wide (requires elevated privileges)" OFF)

# Installation rules
if(CLAPGO_INSTALL_PLUGINS)
    if(WIN32)
        set(CLAP_USER_INSTALL_DIR "$ENV{APPDATA}/CLAP")
        set(CLAP_SYSTEM_INSTALL_DIR "C:/Program Files/Common Files/CLAP")
    elseif(APPLE)
        set(CLAP_USER_INSTALL_DIR "$ENV{HOME}/Library/Audio/Plug-Ins/CLAP")
        set(CLAP_SYSTEM_INSTALL_DIR "/Library/Audio/Plug-Ins/CLAP")
    else()
        set(CLAP_USER_INSTALL_DIR "$ENV{HOME}/.clap")
        set(CLAP_SYSTEM_INSTALL_DIR "/usr/lib/clap")
    endif()
    
    # Set the installation directory based on system-wide flag
    if(CLAPGO_INSTALL_SYSTEM_WIDE)
        set(CLAP_INSTALL_DIR ${CLAP_SYSTEM_INSTALL_DIR})
        message(STATUS "Plugins will be installed system-wide to: ${CLAP_INSTALL_DIR}")
        message(WARNING "System-wide installation may require elevated privileges")
    else()
        set(CLAP_INSTALL_DIR ${CLAP_USER_INSTALL_DIR})
        message(STATUS "Plugins will be installed for the current user to: ${CLAP_INSTALL_DIR}")
    endif()
    
    # Create installation rules for each example plugin
    file(GLOB EXAMPLE_DIRS ${CMAKE_SOURCE_DIR}/examples/*)
    foreach(EXAMPLE_DIR ${EXAMPLE_DIRS})
        if(IS_DIRECTORY ${EXAMPLE_DIR})
            get_filename_component(PLUGIN_NAME ${EXAMPLE_DIR} NAME)
            
            # Install the plugin
            install(FILES ${CMAKE_BINARY_DIR}/examples/${PLUGIN_NAME}/${PLUGIN_NAME}.clap
                   DESTINATION ${CLAP_INSTALL_DIR}
                   PERMISSIONS OWNER_READ OWNER_WRITE OWNER_EXECUTE
                               GROUP_READ GROUP_EXECUTE
                               WORLD_READ WORLD_EXECUTE
            )
        endif()
    endforeach()
endif()

# CPack configuration for packaging
set(CPACK_PACKAGE_NAME "clapgo-plugins")
set(CPACK_PACKAGE_VENDOR "ClapGo")
set(CPACK_PACKAGE_DESCRIPTION_SUMMARY "CLAP Plugins built with Go")
set(CPACK_PACKAGE_VERSION_MAJOR ${PROJECT_VERSION_MAJOR})
set(CPACK_PACKAGE_VERSION_MINOR ${PROJECT_VERSION_MINOR})
set(CPACK_PACKAGE_VERSION_PATCH ${PROJECT_VERSION_PATCH})
set(CPACK_PACKAGE_INSTALL_DIRECTORY "clapgo-plugins")

# Platform-specific packaging
if(WIN32)
    set(CPACK_GENERATOR "NSIS;ZIP")
    set(CPACK_NSIS_INSTALL_ROOT "$PROGRAMFILES64")
elseif(APPLE)
    set(CPACK_GENERATOR "DragNDrop;ZIP")
else()
    set(CPACK_GENERATOR "TGZ;DEB")
endif()

include(CPack)