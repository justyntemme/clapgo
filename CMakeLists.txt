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

# Installation rules
if(CLAPGO_INSTALL_PLUGINS)
    if(WIN32)
        set(CLAP_INSTALL_DIR "$ENV{APPDATA}/CLAP")
    elseif(APPLE)
        set(CLAP_INSTALL_DIR "$ENV{HOME}/Library/Audio/Plug-Ins/CLAP")
    else()
        set(CLAP_INSTALL_DIR "$ENV{HOME}/.clap")
    endif()
    
    install(DIRECTORY ${CMAKE_BINARY_DIR}/plugins/
            DESTINATION ${CLAP_INSTALL_DIR}
            FILES_MATCHING PATTERN "*.clap"
    )
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