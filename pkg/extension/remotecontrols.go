package extension

// RemoteControlsProvider is an extension for plugins that support remote controls.
// Remote controls map plugin parameters to physical controllers or automation.
type RemoteControlsProvider interface {
	// GetPageCount returns the number of remote control pages.
	GetPageCount() uint32

	// GetPage returns information about a remote control page.
	GetPage(pageIndex uint32) (RemoteControlsPage, bool)
}

// RemoteControlsPage represents a page of remote controls
type RemoteControlsPage struct {
	SectionName  string
	PageName     string
	ParamIDs     [8]uint32  // Parameter IDs for each control (8 controls per page)
	IsMapping    [8]bool    // Whether each control is mapped
}

// Common remote control sections
const (
	RemoteControlSectionGeneric = "Generic"
	RemoteControlSectionFilter  = "Filter"
	RemoteControlSectionLFO     = "LFO"
	RemoteControlSectionEnv     = "Envelope"
	RemoteControlSectionOsc     = "Oscillator"
	RemoteControlSectionEffect  = "Effect"
	RemoteControlSectionMod     = "Modulation"
	RemoteControlSectionMixer   = "Mixer"
)

// Maximum number of controls per page
const RemoteControlsCount = 8

// InvalidParamID represents an unmapped control
const InvalidParamID = ^uint32(0)

// CreateStandardPage creates a standard remote controls page
func CreateStandardPage(sectionName, pageName string, paramIDs []uint32) RemoteControlsPage {
	page := RemoteControlsPage{
		SectionName: sectionName,
		PageName:    pageName,
	}
	
	// Fill in parameter IDs and mapping status
	for i := 0; i < RemoteControlsCount; i++ {
		if i < len(paramIDs) && paramIDs[i] != InvalidParamID {
			page.ParamIDs[i] = paramIDs[i]
			page.IsMapping[i] = true
		} else {
			page.ParamIDs[i] = InvalidParamID
			page.IsMapping[i] = false
		}
	}
	
	return page
}

// CreateEmptyPage creates an empty remote controls page
func CreateEmptyPage(sectionName, pageName string) RemoteControlsPage {
	page := RemoteControlsPage{
		SectionName: sectionName,
		PageName:    pageName,
	}
	
	for i := 0; i < RemoteControlsCount; i++ {
		page.ParamIDs[i] = InvalidParamID
		page.IsMapping[i] = false
	}
	
	return page
}