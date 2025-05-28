package controls

// Control mapping types for different hardware controllers
const (
	// Standard MIDI CC mappings
	ControllerModWheel      = 1
	ControllerBreathController = 2
	ControllerFootController = 4
	ControllerPortamento    = 5
	ControllerDataEntry     = 6
	ControllerVolume        = 7
	ControllerBalance       = 8
	ControllerPan           = 10
	ControllerExpression    = 11
	ControllerSustainPedal  = 64
	ControllerPortamentoSwitch = 65
	ControllerSostenutoPedal = 66
	ControllerSoftPedal     = 67
	
	// High-resolution controllers (14-bit)
	ControllerModWheelLSB   = 33
	ControllerDataEntryLSB  = 38
	ControllerVolumeLSB     = 39
	ControllerPanLSB        = 42
	ControllerExpressionLSB = 43
)

// Control parameter categories for organizing remote control pages
const (
	CategoryMain     = "Main"
	CategoryFilter   = "Filter"
	CategoryEQ       = "EQ"
	CategoryDynamics = "Dynamics"
	CategoryEffects  = "Effects"
	CategoryModulation = "Modulation"
	CategoryEnvelope = "Envelope"
	CategoryOscillator = "Oscillator"
	CategoryArpeggio = "Arpeggio"
	CategoryMixer    = "Mixer"
	CategoryGlobal   = "Global"
)

// Standard control ranges
const (
	ControlRangeMin = 0.0
	ControlRangeMax = 1.0
	ControlRangeMid = 0.5
	
	// Bipolar ranges (e.g., for pan, tune)
	ControlBipolarMin = -1.0
	ControlBipolarMax = 1.0
	ControlBipolarCenter = 0.0
)

// Control flags for describing parameter behavior
const (
	ControlFlagContinuous   = 1 << 0  // Continuous value (vs stepped)
	ControlFlagBipolar      = 1 << 1  // Bipolar range (-1 to +1)
	ControlFlagLogarithmic  = 1 << 2  // Logarithmic scaling
	ControlFlagToggle       = 1 << 3  // On/off toggle
	ControlFlagMomentary    = 1 << 4  // Momentary (spring-loaded)
	ControlFlagReadOnly     = 1 << 5  // Display only, not controllable
	ControlFlagHighPriority = 1 << 6  // Should be on first remote page
)

// Common page IDs for organizing controls
const (
	PageIDMain        = 1
	PageIDFilter      = 2
	PageIDEQ          = 3
	PageIDDynamics    = 4
	PageIDEffects     = 5
	PageIDModulation  = 6
	PageIDEnvelope    = 7
	PageIDOscillator  = 8
	PageIDArpeggio    = 9
	PageIDMixer       = 10
	PageIDPerformance = 100  // Performance controls start at 100
)

// Control resolution levels
const (
	Resolution7Bit  = 128   // Standard MIDI (0-127)
	Resolution14Bit = 16384 // High-res MIDI (0-16383)
	ResolutionFloat = 0     // Floating point (infinite resolution)
)

// Invalid control ID
const InvalidControlID = 0xFFFFFFFF