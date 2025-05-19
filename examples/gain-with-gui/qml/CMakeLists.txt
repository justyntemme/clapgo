set(CLAPGO_GAIN_GUI_QML_FILES
    main.qml
)

qt6_add_resources(CLAPGO_GAIN_GUI_QRC gain-gui.qrc)

add_library(clapgo-gain-gui-qml STATIC ${CLAPGO_GAIN_GUI_QRC})
set_property(TARGET clapgo-gain-gui-qml PROPERTY POSITION_INDEPENDENT_CODE ON)

# Create QRC file
file(WRITE ${CMAKE_CURRENT_BINARY_DIR}/gain-gui.qrc "<RCC>\n  <qresource prefix=\"/org/clapgo/gain\">\n")
foreach(QML_FILE ${CLAPGO_GAIN_GUI_QML_FILES})
    file(APPEND ${CMAKE_CURRENT_BINARY_DIR}/gain-gui.qrc "    <file alias=\"${QML_FILE}\">${CMAKE_CURRENT_SOURCE_DIR}/${QML_FILE}</file>\n")
endforeach()
file(APPEND ${CMAKE_CURRENT_BINARY_DIR}/gain-gui.qrc "  </qresource>\n</RCC>\n")