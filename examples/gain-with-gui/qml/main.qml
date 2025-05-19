import QtQuick 2.15
import QtQuick.Controls 2.15
import QtQuick.Layouts 1.15
import org.clap.qml 1.0 as CLAP

Rectangle {
    id: root
    width: 400
    height: 300
    color: "#2e2e2e"
    
    property var gainParam: null
    
    Component.onCompleted: {
        // Get the parameter by ID (1 is the gain parameter in our Go code)
        gainParam = plugin.getParameter(1)
    }
    
    ColumnLayout {
        anchors.fill: parent
        anchors.margins: 20
        spacing: 15
        
        Text {
            Layout.alignment: Qt.AlignHCenter
            text: "ClapGo Gain Plugin"
            color: "white"
            font.pixelSize: 24
            font.bold: true
        }
        
        Item { Layout.fillHeight: true }
        
        CLAP.Knob {
            id: gainKnob
            Layout.alignment: Qt.AlignHCenter
            width: 120
            height: 120
            from: 0.0
            to: 2.0
            value: gainParam ? gainParam.value : 1.0
            text: "Gain"
            textColor: "white"
            valueDisplayFunction: function(v) {
                // Convert linear gain to dB for display
                if (v <= 0.001) return "-∞ dB";
                return (20 * Math.log10(v)).toFixed(1) + " dB";
            }
            
            onValueChanged: {
                if (gainParam) {
                    gainParam.value = value;
                }
            }
            
            Connections {
                target: gainParam
                function onValueChanged() {
                    gainKnob.value = gainParam.value;
                }
            }
        }
        
        Item { Layout.fillHeight: true }
        
        Text {
            Layout.alignment: Qt.AlignHCenter
            text: "Made with ClapGo"
            color: "#aaaaaa"
            font.pixelSize: 12
        }
    }
}