package extension

import (
	"fmt"
	"unsafe"

	hostpkg "github.com/justyntemme/clapgo/pkg/host"
	"github.com/justyntemme/clapgo/pkg/thread"
)

// ExtensionBundle consolidates all extension initialization into a single type
type ExtensionBundle struct {
	ThreadCheck      *thread.Checker
	TrackInfo        *hostpkg.TrackInfoProvider
	TransportControl *hostpkg.TransportControl
	Tuning          *HostTuning
	Logger          *hostpkg.Logger
	
	// Internal state
	host       unsafe.Pointer
	pluginName string
}

// NewExtensionBundle creates and initializes all available extensions
func NewExtensionBundle(host unsafe.Pointer, pluginName string) *ExtensionBundle {
	bundle := &ExtensionBundle{
		host:       host,
		pluginName: pluginName,
	}
	
	if host == nil {
		return bundle
	}
	
	// Initialize logger first so we can log other extensions
	bundle.Logger = hostpkg.NewLogger(host)
	
	// Initialize all extensions with automatic nil checks
	bundle.ThreadCheck = thread.NewChecker(host)
	bundle.TrackInfo = hostpkg.NewTrackInfoProvider(host)
	bundle.TransportControl = hostpkg.NewTransportControl(host)
	bundle.Tuning = NewHostTuning(host)
	
	// Log initialization status
	bundle.logInitStatus()
	
	return bundle
}

// logInitStatus logs which extensions were successfully initialized
func (b *ExtensionBundle) logInitStatus() {
	if b.Logger == nil {
		return
	}
	
	b.Logger.Info(fmt.Sprintf("%s extension bundle initialized:", b.pluginName))
	
	if b.ThreadCheck != nil && b.ThreadCheck.IsAvailable() {
		b.Logger.Info("  ✓ Thread Check - thread safety validation enabled")
	} else {
		b.Logger.Debug("  ✗ Thread Check - not available")
	}
	
	if b.TrackInfo != nil {
		b.Logger.Info("  ✓ Track Info - track context available")
	} else {
		b.Logger.Debug("  ✗ Track Info - not available")
	}
	
	if b.TransportControl != nil {
		b.Logger.Info("  ✓ Transport Control - playback control enabled")
	} else {
		b.Logger.Debug("  ✗ Transport Control - not available")
	}
	
	if b.Tuning != nil {
		b.Logger.Info("  ✓ Tuning Support - microtonal support enabled")
		
		// Log available tunings
		tunings := b.Tuning.GetAllTunings()
		if len(tunings) > 0 {
			b.Logger.Info(fmt.Sprintf("    Available tunings: %d", len(tunings)))
			for _, t := range tunings {
				b.Logger.Debug(fmt.Sprintf("      - %s (ID: %d, Dynamic: %v)",
					t.Name, t.TuningID, t.IsDynamic))
			}
		}
	} else {
		b.Logger.Debug("  ✗ Tuning Support - not available")
	}
}

// IsAvailable returns true if any extensions were successfully initialized
func (b *ExtensionBundle) IsAvailable() bool {
	return b.Logger != nil || 
		   b.ThreadCheck != nil || 
		   b.TrackInfo != nil || 
		   b.TransportControl != nil || 
		   b.Tuning != nil
}

// HasThreadCheck returns true if thread checking is available
func (b *ExtensionBundle) HasThreadCheck() bool {
	return b.ThreadCheck != nil && b.ThreadCheck.IsAvailable()
}

// HasTrackInfo returns true if track info is available
func (b *ExtensionBundle) HasTrackInfo() bool {
	return b.TrackInfo != nil
}

// HasTransportControl returns true if transport control is available
func (b *ExtensionBundle) HasTransportControl() bool {
	return b.TransportControl != nil
}

// HasTuning returns true if tuning support is available
func (b *ExtensionBundle) HasTuning() bool {
	return b.Tuning != nil
}

// HasLogger returns true if logging is available
func (b *ExtensionBundle) HasLogger() bool {
	return b.Logger != nil
}

// RequestTogglePlay requests the host to toggle play/pause state
// Returns true if the request was sent successfully
func (b *ExtensionBundle) RequestTogglePlay() bool {
	if b.TransportControl == nil {
		return false
	}
	
	b.TransportControl.RequestTogglePlay()
	if b.Logger != nil {
		b.Logger.Info("Transport toggle play requested")
	}
	return true
}

// LogInfo logs an info message if logger is available
func (b *ExtensionBundle) LogInfo(message string) {
	if b.Logger != nil {
		b.Logger.Info(message)
	}
}

// LogDebug logs a debug message if logger is available
func (b *ExtensionBundle) LogDebug(message string) {
	if b.Logger != nil {
		b.Logger.Debug(message)
	}
}

// LogWarning logs a warning message if logger is available
func (b *ExtensionBundle) LogWarning(message string) {
	if b.Logger != nil {
		b.Logger.Warning(message)
	}
}

// LogError logs an error message if logger is available
func (b *ExtensionBundle) LogError(message string) {
	if b.Logger != nil {
		b.Logger.Error(message)
	}
}

// GetTrackInfo returns track information if available
func (b *ExtensionBundle) GetTrackInfo() (*hostpkg.TrackInfo, bool) {
	if b.TrackInfo == nil {
		return nil, false
	}
	return b.TrackInfo.Get()
}

// ApplyTuning applies tuning to a frequency if tuning support is available
func (b *ExtensionBundle) ApplyTuning(baseFreq float64, tuningID int64, channel, key int32, keyboardMapping int16) float64 {
	if b.Tuning == nil || tuningID == 0 {
		return baseFreq
	}
	return b.Tuning.ApplyTuning(baseFreq, uint64(tuningID), channel, key, uint32(keyboardMapping))
}

// GetAvailableTunings returns all available tunings
func (b *ExtensionBundle) GetAvailableTunings() []TuningInfo {
	if b.Tuning == nil {
		return nil
	}
	return b.Tuning.GetAllTunings()
}