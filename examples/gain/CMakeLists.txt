# Add Go shared library for the gain plugin
add_go_library(gain-go
    SOURCES
        ${CMAKE_CURRENT_SOURCE_DIR}/main.go
        ${CMAKE_CURRENT_SOURCE_DIR}/constants.go
    OUTPUT_NAME
        gain
)

# Find json-c
find_package(PkgConfig REQUIRED)
pkg_check_modules(JSON_C REQUIRED json-c)

# Build the CLAP plugin
add_clap_plugin(gain
    SOURCES
        ${CMAKE_SOURCE_DIR}/src/c/plugin.c
        ${CMAKE_SOURCE_DIR}/src/c/bridge.c
        ${CMAKE_SOURCE_DIR}/src/c/manifest.c
    LINK_LIBRARIES
        gain-go
        ${JSON_C_LIBRARIES}
)

# Add json-c include directories
target_include_directories(gain PRIVATE ${JSON_C_INCLUDE_DIRS})