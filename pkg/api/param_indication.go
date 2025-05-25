package api

// #include <stdlib.h>
// #include <string.h>
// #include "../../include/clap/include/clap/ext/param-indication.h"
// #include "../../include/clap/include/clap/color.h"
import "C"
import (
	"unsafe"
)

// Param indication extension constants
const (
	ExtParamIndicationID     = "clap.param-indication/4"
	ExtParamIndicationCompat = "clap.param-indication.draft/4"
)

// Automation state constants
const (
	// The host doesn't have an automation for this parameter
	ParamIndicationAutomationNone = 0
	
	// The host has an automation for this parameter, but it isn't playing it
	ParamIndicationAutomationPresent = 1
	
	// The host is playing an automation for this parameter
	ParamIndicationAutomationPlaying = 2
	
	// The host is recording an automation on this parameter
	ParamIndicationAutomationRecording = 3
	
	// The host should play an automation for this parameter, but the user has started to adjust this
	// parameter and is overriding the automation playback
	ParamIndicationAutomationOverriding = 4
)


// ParamIndicationProvider is the interface that plugins implement to receive parameter indications
type ParamIndicationProvider interface {
	// OnParamMappingSet is called when the host sets or clears a mapping indication
	// paramID: the parameter ID
	// hasMapping: does the parameter currently have a mapping?
	// color: if set, the color to use to highlight the control in the plugin GUI
	// label: if set, a small string to display on top of the knob which identifies the hardware controller
	// description: if set, a string which can be used in a tooltip, which describes the current mapping
	OnParamMappingSet(paramID uint32, hasMapping bool, color *Color, label string, description string)
	
	// OnParamAutomationSet is called when the host sets or clears an automation indication
	// paramID: the parameter ID
	// automationState: current automation state for the given parameter
	// color: if set, the color to use to display the automation indication in the plugin GUI
	OnParamAutomationSet(paramID uint32, automationState uint32, color *Color)
}

// ColorFromC converts a C color to a Go color
func ColorFromC(cColor unsafe.Pointer) *Color {
	if cColor == nil {
		return nil
	}
	
	c := (*C.clap_color_t)(cColor)
	return &Color{
		Alpha: uint8(c.alpha),
		Red:   uint8(c.red),
		Green: uint8(c.green),
		Blue:  uint8(c.blue),
	}
}

// ColorToC converts a Go color to a C color
func ColorToC(color *Color) *C.clap_color_t {
	if color == nil {
		return nil
	}
	
	cColor := (*C.clap_color_t)(C.malloc(C.sizeof_clap_color_t))
	cColor.alpha = C.uint8_t(color.Alpha)
	cColor.red = C.uint8_t(color.Red)
	cColor.green = C.uint8_t(color.Green)
	cColor.blue = C.uint8_t(color.Blue)
	
	return cColor
}