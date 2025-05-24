# Add Go shared library for the synth plugin
add_go_library(synth-go
    SOURCES
        ${CMAKE_CURRENT_SOURCE_DIR}/main.go
        ${CMAKE_CURRENT_SOURCE_DIR}/constants.go
    OUTPUT_NAME
        synth
)

# Find json-c
find_package(PkgConfig REQUIRED)
pkg_check_modules(JSON_C REQUIRED json-c)

# Build the CLAP plugin
add_clap_plugin(synth
    SOURCES
        ${CMAKE_SOURCE_DIR}/src/c/plugin.c
        ${CMAKE_SOURCE_DIR}/src/c/bridge.c
        ${CMAKE_SOURCE_DIR}/src/c/manifest.c
    LINK_LIBRARIES
        synth-go
        ${JSON_C_LIBRARIES}
)

# Add json-c include directories
target_include_directories(synth PRIVATE ${JSON_C_INCLUDE_DIRS})