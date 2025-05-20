add_custom_command(
    OUTPUT ${CMAKE_CURRENT_BINARY_DIR}/libsynth.so ${CMAKE_CURRENT_BINARY_DIR}/libsynth.h
    COMMENT "Building Go shared library synth"
    COMMAND ${CMAKE_COMMAND} -E env GOPATH=${CMAKE_BINARY_DIR}/go CGO_ENABLED=1 CLAPGO_PLUGIN_TYPE=synth ${GO_EXECUTABLE} build -buildmode=c-shared -o ${CMAKE_CURRENT_BINARY_DIR}/libsynth.so ${CMAKE_CURRENT_SOURCE_DIR}/main.go
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    DEPENDS ${CMAKE_CURRENT_SOURCE_DIR}/main.go
)

add_custom_target(synth-go DEPENDS ${CMAKE_CURRENT_BINARY_DIR}/libsynth.so)

add_library(synth MODULE 
    ${CMAKE_SOURCE_DIR}/src/c/plugin.c
)

set_target_properties(synth PROPERTIES 
    OUTPUT_NAME "synth"
    PREFIX ""
    SUFFIX ".clap"
)

target_include_directories(synth PRIVATE 
    ${CMAKE_SOURCE_DIR}/include
    ${CMAKE_BINARY_DIR}
    ${CMAKE_CURRENT_BINARY_DIR}
)

target_link_libraries(synth PRIVATE 
    ${CMAKE_CURRENT_BINARY_DIR}/libsynth.so
)

add_dependencies(synth synth-go)

install(TARGETS synth
    LIBRARY DESTINATION ${CMAKE_INSTALL_PREFIX}/plugins
)