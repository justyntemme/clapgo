# CMake module for building Go libraries and integrating with CLAP plugins

# Find Go executable
find_program(GO_EXECUTABLE go REQUIRED)

# Set build flags based on configuration
if(CMAKE_BUILD_TYPE STREQUAL "Debug")
    set(GO_BUILD_FLAGS "-gcflags=all=-N -l")
else()
    set(GO_BUILD_FLAGS "-ldflags=-s -w")
endif()

# Function to build a Go shared library
function(add_go_library TARGET_NAME)
    # Parse arguments
    cmake_parse_arguments(ARG "" "OUTPUT_NAME" "SOURCES" ${ARGN})
    
    if(NOT ARG_OUTPUT_NAME)
        set(ARG_OUTPUT_NAME ${TARGET_NAME})
    endif()
    
    # Sources are already full paths
    set(GO_SOURCES "${ARG_SOURCES}")
    
    # Output file name
    if(WIN32)
        set(LIB_PREFIX "")
        set(LIB_SUFFIX ".dll")
    elseif(APPLE)
        set(LIB_PREFIX "lib")
        set(LIB_SUFFIX ".dylib")
    else()
        set(LIB_PREFIX "lib")
        set(LIB_SUFFIX ".so")
    endif()
    
    set(OUTPUT_LIB "${CMAKE_CURRENT_BINARY_DIR}/${LIB_PREFIX}${ARG_OUTPUT_NAME}${LIB_SUFFIX}")
    set(OUTPUT_HEADER "${CMAKE_CURRENT_BINARY_DIR}/${LIB_PREFIX}${ARG_OUTPUT_NAME}.h")
    
    # Build command
    add_custom_command(
        OUTPUT ${OUTPUT_LIB} ${OUTPUT_HEADER}
        COMMAND ${CMAKE_COMMAND} -E env 
                GOPATH=${CMAKE_BINARY_DIR}/go
                CGO_ENABLED=1
                ${GO_EXECUTABLE} build 
                -buildmode=c-shared 
                -o ${OUTPUT_LIB}
                ${GO_BUILD_FLAGS}
                ${GO_SOURCES}
        DEPENDS ${GO_SOURCES}
        WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
        COMMENT "Building Go shared library ${ARG_OUTPUT_NAME}"
    )
    
    # Create interface library to facilitate CMake integration
    add_library(${TARGET_NAME}_interface INTERFACE)
    target_include_directories(${TARGET_NAME}_interface INTERFACE ${CMAKE_CURRENT_BINARY_DIR})
    
    # Create the target
    add_custom_target(${TARGET_NAME} ALL DEPENDS ${OUTPUT_LIB} ${OUTPUT_HEADER})
    
    # Create a symbolic link in the target directory if needed
    get_filename_component(OUTPUT_DIR ${OUTPUT_LIB} DIRECTORY)
    add_custom_command(TARGET ${TARGET_NAME} POST_BUILD
        COMMAND ${CMAKE_COMMAND} -E copy_if_different
            ${OUTPUT_LIB} ${CMAKE_CURRENT_BINARY_DIR}/
    )
    
    # Export library info
    set(${TARGET_NAME}_LIBRARY ${OUTPUT_LIB} PARENT_SCOPE)
    set(${TARGET_NAME}_HEADER ${OUTPUT_HEADER} PARENT_SCOPE)
    set(${TARGET_NAME}_INTERFACE ${TARGET_NAME}_interface PARENT_SCOPE)
endfunction()

# Function to create a CLAP plugin
function(add_clap_plugin TARGET_NAME)
    cmake_parse_arguments(ARG "" "" "SOURCES;LINK_LIBRARIES" ${ARGN})
    
    # Determine the output extension
    set(CLAP_EXTENSION ".clap")
    
    # Build the plugin as a shared library
    add_library(${TARGET_NAME} MODULE ${ARG_SOURCES})
    
    # Include CLAP headers
    target_include_directories(${TARGET_NAME} PRIVATE
        ${CMAKE_SOURCE_DIR}/include/clap/include
    )
    
    # Special handling for Go libraries
    foreach(LIB ${ARG_LINK_LIBRARIES})
        if(TARGET ${LIB})
            add_dependencies(${TARGET_NAME} ${LIB})
            
            # Get the plain name for the library
            string(REPLACE "-go" "" LIB_BASE_NAME ${LIB})
            
            # Link against the Go library
            if(WIN32)
                # Windows needs special handling
                target_link_libraries(${TARGET_NAME} ${${LIB}_LIBRARY})
            else()
                # Add rpath for the library directory
                set_target_properties(${TARGET_NAME} PROPERTIES
                    INSTALL_RPATH "${CMAKE_CURRENT_BINARY_DIR}")
                    
                # Link directly to the shared object (absolute path)
                target_link_libraries(${TARGET_NAME} ${${LIB}_LIBRARY})
            endif()
            
            # Add header directory
            get_filename_component(LIB_DIR ${${LIB}_HEADER} DIRECTORY)
            target_include_directories(${TARGET_NAME} PRIVATE ${LIB_DIR})
        else()
            target_link_libraries(${TARGET_NAME} ${LIB})
        endif()
    endforeach()
    
    # Set output name/properties
    set_target_properties(${TARGET_NAME} PROPERTIES
        PREFIX ""
        SUFFIX ${CLAP_EXTENSION}
    )
    
    # Platform-specific settings
    if(APPLE)
        # macOS requires a bundle structure
        set_target_properties(${TARGET_NAME} PROPERTIES
            BUNDLE TRUE
            BUNDLE_EXTENSION "clap"
            MACOSX_BUNDLE_INFO_PLIST "${CMAKE_SOURCE_DIR}/cmake/MacOSXBundleInfo.plist.in"
            MACOSX_BUNDLE_BUNDLE_NAME "${TARGET_NAME}"
            MACOSX_BUNDLE_GUI_IDENTIFIER "org.clapgo.${TARGET_NAME}"
            MACOSX_BUNDLE_BUNDLE_VERSION "1.0"
        )
    endif()
    
    # No installation rules here - handled by main CMakeLists.txt
endfunction()