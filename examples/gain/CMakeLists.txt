# Add Go shared library for the gain plugin
add_go_library(gain-go
    SOURCES
        ${CMAKE_CURRENT_SOURCE_DIR}/main.go
    OUTPUT_NAME
        gain
)

# Build the CLAP plugin
add_clap_plugin(gain
    SOURCES
        ${CMAKE_SOURCE_DIR}/src/c/plugin.c
    LINK_LIBRARIES
        gain-go
)