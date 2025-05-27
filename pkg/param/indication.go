package param

// Param indication extension constants
const (
	ExtIndicationID     = "clap.param-indication/4"
	ExtIndicationCompat = "clap.param-indication.draft/4"
)

// Automation state constants for parameter indication
const (
	// The host doesn't have an automation for this parameter
	IndicationAutomationNone = 0
	
	// The host has an automation for this parameter, but it isn't playing it
	IndicationAutomationPresent = 1
	
	// The host is playing an automation for this parameter
	IndicationAutomationPlaying = 2
	
	// The host is recording an automation on this parameter
	IndicationAutomationRecording = 3
	
	// The host should play an automation for this parameter, but the user has started to adjust this
	// parameter and is overriding the automation playback
	IndicationAutomationOverriding = 4
)

// Color represents an RGBA color for parameter indication
type Color struct {
	Alpha uint8
	Red   uint8
	Green uint8
	Blue  uint8
}

// IndicationProvider is the interface that plugins implement to receive parameter indications
type IndicationProvider interface {
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