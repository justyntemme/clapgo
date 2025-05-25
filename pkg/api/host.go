package api

// #include "../../include/clap/include/clap/clap.h"
// #include <stdlib.h>
//
// // Helper function to call host log
// static void clap_host_log_helper(const clap_host_t* host, int32_t severity, const char* msg) {
//     if (host && host->get_extension) {
//         const clap_host_log_t* log_ext = (const clap_host_log_t*)host->get_extension(host, CLAP_EXT_LOG);
//         if (log_ext && log_ext->log) {
//             log_ext->log(host, severity, msg);
//         }
//     }
// }
//
// // Helper function to register a timer
// static bool clap_host_register_timer_helper(const clap_host_t* host, uint32_t period_ms, clap_id* timer_id) {
//     if (host && host->get_extension) {
//         const clap_host_timer_support_t* timer_ext = (const clap_host_timer_support_t*)host->get_extension(host, CLAP_EXT_TIMER_SUPPORT);
//         if (timer_ext && timer_ext->register_timer) {
//             return timer_ext->register_timer(host, period_ms, timer_id);
//         }
//     }
//     return false;
// }
//
// // Helper function to unregister a timer
// static bool clap_host_unregister_timer_helper(const clap_host_t* host, clap_id timer_id) {
//     if (host && host->get_extension) {
//         const clap_host_timer_support_t* timer_ext = (const clap_host_timer_support_t*)host->get_extension(host, CLAP_EXT_TIMER_SUPPORT);
//         if (timer_ext && timer_ext->unregister_timer) {
//             return timer_ext->unregister_timer(host, timer_id);
//         }
//     }
//     return false;
// }
//
// // Helper function to notify host about latency change
// static void clap_host_latency_changed_helper(const clap_host_t* host) {
//     if (host && host->get_extension) {
//         const clap_host_latency_t* latency_ext = (const clap_host_latency_t*)host->get_extension(host, CLAP_EXT_LATENCY);
//         if (latency_ext && latency_ext->changed) {
//             latency_ext->changed(host);
//         }
//     }
// }
//
// // Helper function to notify host about tail change
// static void clap_host_tail_changed_helper(const clap_host_t* host) {
//     if (host && host->get_extension) {
//         const clap_host_tail_t* tail_ext = (const clap_host_tail_t*)host->get_extension(host, CLAP_EXT_TAIL);
//         if (tail_ext && tail_ext->changed) {
//             tail_ext->changed(host);
//         }
//     }
// }
//
// // Helper function to get track info from host
// static bool clap_host_get_track_info_helper(const clap_host_t* host, clap_track_info_t* info) {
//     if (host && host->get_extension && info) {
//         const clap_host_track_info_t* track_info_ext = (const clap_host_track_info_t*)host->get_extension(host, CLAP_EXT_TRACK_INFO);
//         if (track_info_ext && track_info_ext->get) {
//             return track_info_ext->get(host, info);
//         }
//     }
//     return false;
// }
import "C"
import (
	"fmt"
	"strings"
	"sync"
	"unsafe"
)

// Default log buffer size - can be adjusted based on needs
const DefaultLogBufferSize = 4096 // 4KB buffer for detailed logs

// LogBuffer represents a reusable buffer for log formatting
type LogBuffer struct {
	builder strings.Builder
	cBuffer [DefaultLogBufferSize]C.char // Fixed-size C string buffer
}

// LoggerPool manages reusable buffers for zero-allocation logging
type LoggerPool struct {
	bufferPool sync.Pool
}

// Global logger pool instance
var globalLoggerPool = &LoggerPool{
	bufferPool: sync.Pool{
		New: func() interface{} {
			return &LogBuffer{}
		},
	},
}

// HostLogger provides logging functionality to the host
type HostLogger struct {
	host unsafe.Pointer
}

// NewHostLogger creates a new host logger
func NewHostLogger(host unsafe.Pointer) *HostLogger {
	return &HostLogger{host: host}
}

// Log sends a log message to the host with optional formatting
func (l *HostLogger) Log(severity int32, message string, args ...interface{}) {
	if l.host == nil {
		return
	}
	
	// Get a buffer from the pool
	buffer := globalLoggerPool.bufferPool.Get().(*LogBuffer)
	defer globalLoggerPool.bufferPool.Put(buffer)
	
	// Reset the string builder
	buffer.builder.Reset()
	
	// Format message if args are provided
	if len(args) > 0 {
		// Use Fprintf with the builder to avoid allocation
		fmt.Fprintf(&buffer.builder, message, args...)
	} else {
		// Direct write for simple messages
		buffer.builder.WriteString(message)
	}
	
	// Get the formatted string
	formatted := buffer.builder.String()
	
	// Copy to C buffer if it fits
	if len(formatted) < len(buffer.cBuffer)-1 {
		// Zero-copy conversion to C string using pre-allocated buffer
		for i := 0; i < len(formatted); i++ {
			buffer.cBuffer[i] = C.char(formatted[i])
		}
		buffer.cBuffer[len(formatted)] = 0
		
		// Use the pre-allocated C buffer
		C.clap_host_log_helper((*C.clap_host_t)(l.host), C.int32_t(severity), &buffer.cBuffer[0])
	} else {
		// Fallback for messages that exceed buffer size
		cMsg := C.CString(formatted)
		defer C.free(unsafe.Pointer(cMsg))
		C.clap_host_log_helper((*C.clap_host_t)(l.host), C.int32_t(severity), cMsg)
	}
}

// Debug logs a debug message
func (l *HostLogger) Debug(message string) {
	l.Log(LogSeverityDebug, message)
}

// Info logs an info message
func (l *HostLogger) Info(message string) {
	l.Log(LogSeverityInfo, message)
}

// Warning logs a warning message
func (l *HostLogger) Warning(message string) {
	l.Log(LogSeverityWarning, message)
}

// Error logs an error message
func (l *HostLogger) Error(message string) {
	l.Log(LogSeverityError, message)
}

// Fatal logs a fatal error message
func (l *HostLogger) Fatal(message string) {
	l.Log(LogSeverityFatal, message)
}

// HostTimerSupport provides timer functionality from the host
type HostTimerSupport struct {
	host unsafe.Pointer
}

// NewHostTimerSupport creates a new host timer support
func NewHostTimerSupport(host unsafe.Pointer) *HostTimerSupport {
	return &HostTimerSupport{host: host}
}

// RegisterTimer registers a periodic timer with the host
// Returns the timer ID if successful, or InvalidID if failed
func (t *HostTimerSupport) RegisterTimer(periodMs uint32) uint64 {
	if t.host == nil {
		return InvalidID
	}
	
	var timerID C.clap_id
	success := C.clap_host_register_timer_helper((*C.clap_host_t)(t.host), C.uint32_t(periodMs), &timerID)
	if success {
		return uint64(timerID)
	}
	return InvalidID
}

// UnregisterTimer unregisters a timer
func (t *HostTimerSupport) UnregisterTimer(timerID uint64) bool {
	if t.host == nil {
		return false
	}
	
	return bool(C.clap_host_unregister_timer_helper((*C.clap_host_t)(t.host), C.clap_id(timerID)))
}

// HostLatencyNotifier notifies the host about latency changes
type HostLatencyNotifier struct {
	host unsafe.Pointer
}

// NewHostLatencyNotifier creates a new host latency notifier
func NewHostLatencyNotifier(host unsafe.Pointer) *HostLatencyNotifier {
	return &HostLatencyNotifier{host: host}
}

// NotifyLatencyChanged tells the host that the plugin's latency has changed
func (n *HostLatencyNotifier) NotifyLatencyChanged() {
	if n.host == nil {
		return
	}
	
	C.clap_host_latency_changed_helper((*C.clap_host_t)(n.host))
}

// HostTailNotifier notifies the host about tail changes
type HostTailNotifier struct {
	host unsafe.Pointer
}

// NewHostTailNotifier creates a new host tail notifier
func NewHostTailNotifier(host unsafe.Pointer) *HostTailNotifier {
	return &HostTailNotifier{host: host}
}

// NotifyTailChanged tells the host that the plugin's tail has changed
func (n *HostTailNotifier) NotifyTailChanged() {
	if n.host == nil {
		return
	}
	
	C.clap_host_tail_changed_helper((*C.clap_host_t)(n.host))
}

// HostAudioPortsConfigNotifier notifies the host about audio ports config changes
type HostAudioPortsConfigNotifier struct {
	host unsafe.Pointer
}

// NewHostAudioPortsConfigNotifier creates a new host audio ports config notifier
func NewHostAudioPortsConfigNotifier(host unsafe.Pointer) *HostAudioPortsConfigNotifier {
	return &HostAudioPortsConfigNotifier{host: host}
}

// NotifyRescan tells the host to rescan the full list of configs
func (n *HostAudioPortsConfigNotifier) NotifyRescan() {
	if n.host == nil {
		return
	}
	
	// Host notification not yet implemented
}

// HostSurroundNotifier notifies the host about surround changes
type HostSurroundNotifier struct {
	host unsafe.Pointer
}

// NewHostSurroundNotifier creates a new host surround notifier
func NewHostSurroundNotifier(host unsafe.Pointer) *HostSurroundNotifier {
	return &HostSurroundNotifier{host: host}
}

// NotifySurroundChanged tells the host that the channel map has changed
func (n *HostSurroundNotifier) NotifySurroundChanged() {
	if n.host == nil {
		return
	}
	
	// Host notification not yet implemented
}

// HostVoiceInfoNotifier notifies the host about voice info changes
type HostVoiceInfoNotifier struct {
	host unsafe.Pointer
}

// NewHostVoiceInfoNotifier creates a new host voice info notifier
func NewHostVoiceInfoNotifier(host unsafe.Pointer) *HostVoiceInfoNotifier {
	return &HostVoiceInfoNotifier{host: host}
}

// NotifyVoiceInfoChanged tells the host that voice info has changed
func (n *HostVoiceInfoNotifier) NotifyVoiceInfoChanged() {
	if n.host == nil {
		return
	}
	
	// Host notification not yet implemented
}

// HostTrackInfo provides track information from the host
type HostTrackInfo struct {
	host unsafe.Pointer
}

// NewHostTrackInfo creates a new host track info
func NewHostTrackInfo(host unsafe.Pointer) *HostTrackInfo {
	return &HostTrackInfo{host: host}
}

// GetTrackInfo retrieves current track information from the host
func (t *HostTrackInfo) GetTrackInfo() (*TrackInfo, bool) {
	if t.host == nil {
		return nil, false
	}
	
	var cInfo C.clap_track_info_t
	success := C.clap_host_get_track_info_helper((*C.clap_host_t)(t.host), &cInfo)
	if !success {
		return nil, false
	}
	
	info := &TrackInfo{
		Flags: uint64(cInfo.flags),
	}
	
	// Extract track name if available
	if info.Flags&TrackInfoHasTrackName != 0 {
		info.Name = C.GoString(&cInfo.name[0])
	}
	
	// Extract track color if available
	if info.Flags&TrackInfoHasTrackColor != 0 {
		info.Color = Color{
			Alpha: uint8(cInfo.color.alpha),
			Red:   uint8(cInfo.color.red),
			Green: uint8(cInfo.color.green),
			Blue:  uint8(cInfo.color.blue),
		}
	}
	
	// Extract audio channel info if available
	if info.Flags&TrackInfoHasAudioChannel != 0 {
		info.AudioChannelCount = int32(cInfo.audio_channel_count)
		if cInfo.audio_port_type != nil {
			info.AudioPortType = C.GoString(cInfo.audio_port_type)
		}
	}
	
	return info, true
}