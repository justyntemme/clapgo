package host

// #include "../../include/clap/include/clap/clap.h"
// #include "../../include/clap/include/clap/ext/draft/transport-control.h"
// #include <stdlib.h>
//
// // Timer support helpers
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
// // Latency helper
// static void clap_host_latency_changed_helper(const clap_host_t* host) {
//     if (host && host->get_extension) {
//         const clap_host_latency_t* latency_ext = (const clap_host_latency_t*)host->get_extension(host, CLAP_EXT_LATENCY);
//         if (latency_ext && latency_ext->changed) {
//             latency_ext->changed(host);
//         }
//     }
// }
//
// // Tail helper
// static void clap_host_tail_changed_helper(const clap_host_t* host) {
//     if (host && host->get_extension) {
//         const clap_host_tail_t* tail_ext = (const clap_host_tail_t*)host->get_extension(host, CLAP_EXT_TAIL);
//         if (tail_ext && tail_ext->changed) {
//             tail_ext->changed(host);
//         }
//     }
// }
//
// // Audio ports config helper
// static void clap_host_audio_ports_config_rescan_helper(const clap_host_t* host) {
//     if (host && host->get_extension) {
//         const clap_host_audio_ports_config_t* config_ext = (const clap_host_audio_ports_config_t*)host->get_extension(host, CLAP_EXT_AUDIO_PORTS_CONFIG);
//         if (config_ext && config_ext->rescan) {
//             config_ext->rescan(host);
//         }
//     }
// }
//
// // Voice info helper
// static void clap_host_voice_info_changed_helper(const clap_host_t* host) {
//     if (host && host->get_extension) {
//         const clap_host_voice_info_t* voice_info_ext = (const clap_host_voice_info_t*)host->get_extension(host, CLAP_EXT_VOICE_INFO);
//         if (voice_info_ext && voice_info_ext->changed) {
//             voice_info_ext->changed(host);
//         }
//     }
// }
//
// // Track info helper
// static bool clap_host_get_track_info_helper(const clap_host_t* host, clap_track_info_t* info) {
//     if (host && host->get_extension && info) {
//         const clap_host_track_info_t* track_info_ext = (const clap_host_track_info_t*)host->get_extension(host, CLAP_EXT_TRACK_INFO);
//         if (track_info_ext && track_info_ext->get) {
//             return track_info_ext->get(host, info);
//         }
//     }
//     return false;
// }
//
// // Transport control helpers
// static void clap_host_transport_request_start(const clap_host_t* host) {
//     if (host && host->get_extension) {
//         const clap_host_transport_control_t* transport = (const clap_host_transport_control_t*)host->get_extension(host, "clap.transport-control/1");
//         if (transport && transport->request_start) {
//             transport->request_start(host);
//         }
//     }
// }
//
// static void clap_host_transport_request_stop(const clap_host_t* host) {
//     if (host && host->get_extension) {
//         const clap_host_transport_control_t* transport = (const clap_host_transport_control_t*)host->get_extension(host, "clap.transport-control/1");
//         if (transport && transport->request_stop) {
//             transport->request_stop(host);
//         }
//     }
// }
//
// static void clap_host_transport_request_continue(const clap_host_t* host) {
//     if (host && host->get_extension) {
//         const clap_host_transport_control_t* transport = (const clap_host_transport_control_t*)host->get_extension(host, "clap.transport-control/1");
//         if (transport && transport->request_continue) {
//             transport->request_continue(host);
//         }
//     }
// }
//
// static void clap_host_transport_request_pause(const clap_host_t* host) {
//     if (host && host->get_extension) {
//         const clap_host_transport_control_t* transport = (const clap_host_transport_control_t*)host->get_extension(host, "clap.transport-control/1");
//         if (transport && transport->request_pause) {
//             transport->request_pause(host);
//         }
//     }
// }
//
// static void clap_host_transport_request_toggle_play(const clap_host_t* host) {
//     if (host && host->get_extension) {
//         const clap_host_transport_control_t* transport = (const clap_host_transport_control_t*)host->get_extension(host, "clap.transport-control/1");
//         if (transport && transport->request_toggle_play) {
//             transport->request_toggle_play(host);
//         }
//     }
// }
//
// static void clap_host_transport_request_jump(const clap_host_t* host, double position) {
//     if (host && host->get_extension) {
//         const clap_host_transport_control_t* transport = (const clap_host_transport_control_t*)host->get_extension(host, "clap.transport-control/1");
//         if (transport && transport->request_jump) {
//             transport->request_jump(host, position);
//         }
//     }
// }
//
// static void clap_host_transport_request_loop_region(const clap_host_t* host, double start, double duration) {
//     if (host && host->get_extension) {
//         const clap_host_transport_control_t* transport = (const clap_host_transport_control_t*)host->get_extension(host, "clap.transport-control/1");
//         if (transport && transport->request_loop_region) {
//             transport->request_loop_region(host, start, duration);
//         }
//     }
// }
//
// static void clap_host_transport_request_toggle_loop(const clap_host_t* host) {
//     if (host && host->get_extension) {
//         const clap_host_transport_control_t* transport = (const clap_host_transport_control_t*)host->get_extension(host, "clap.transport-control/1");
//         if (transport && transport->request_toggle_loop) {
//             transport->request_toggle_loop(host);
//         }
//     }
// }
//
// static void clap_host_transport_request_enable_loop(const clap_host_t* host, bool is_enabled) {
//     if (host && host->get_extension) {
//         const clap_host_transport_control_t* transport = (const clap_host_transport_control_t*)host->get_extension(host, "clap.transport-control/1");
//         if (transport && transport->request_enable_loop) {
//             transport->request_enable_loop(host, is_enabled);
//         }
//     }
// }
//
// static void clap_host_transport_request_record(const clap_host_t* host, bool is_recording) {
//     if (host && host->get_extension) {
//         const clap_host_transport_control_t* transport = (const clap_host_transport_control_t*)host->get_extension(host, "clap.transport-control/1");
//         if (transport && transport->request_record) {
//             transport->request_record(host, is_recording);
//         }
//     }
// }
//
// static void clap_host_transport_request_toggle_record(const clap_host_t* host) {
//     if (host && host->get_extension) {
//         const clap_host_transport_control_t* transport = (const clap_host_transport_control_t*)host->get_extension(host, "clap.transport-control/1");
//         if (transport && transport->request_toggle_record) {
//             transport->request_toggle_record(host);
//         }
//     }
// }
import "C"
import (
	"unsafe"
)

// InvalidID represents an invalid ID value
const InvalidID = ^uint64(0)

// TimerSupport provides timer functionality from the host
type TimerSupport struct {
	host unsafe.Pointer
}

// NewTimerSupport creates a new host timer support
func NewTimerSupport(host unsafe.Pointer) *TimerSupport {
	return &TimerSupport{host: host}
}

// RegisterTimer registers a periodic timer with the host
// Returns the timer ID if successful, or InvalidID if failed
func (t *TimerSupport) RegisterTimer(periodMs uint32) uint64 {
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
func (t *TimerSupport) UnregisterTimer(timerID uint64) bool {
	if t.host == nil {
		return false
	}
	
	return bool(C.clap_host_unregister_timer_helper((*C.clap_host_t)(t.host), C.clap_id(timerID)))
}

// LatencyNotifier notifies the host about latency changes
type LatencyNotifier struct {
	host unsafe.Pointer
}

// NewLatencyNotifier creates a new host latency notifier
func NewLatencyNotifier(host unsafe.Pointer) *LatencyNotifier {
	return &LatencyNotifier{host: host}
}

// Changed tells the host that the plugin's latency has changed
func (n *LatencyNotifier) Changed() {
	if n.host == nil {
		return
	}
	
	C.clap_host_latency_changed_helper((*C.clap_host_t)(n.host))
}

// TailNotifier notifies the host about tail changes
type TailNotifier struct {
	host unsafe.Pointer
}

// NewTailNotifier creates a new host tail notifier
func NewTailNotifier(host unsafe.Pointer) *TailNotifier {
	return &TailNotifier{host: host}
}

// Changed tells the host that the plugin's tail has changed
func (n *TailNotifier) Changed() {
	if n.host == nil {
		return
	}
	
	C.clap_host_tail_changed_helper((*C.clap_host_t)(n.host))
}

// AudioPortsConfigNotifier notifies the host about audio ports config changes
type AudioPortsConfigNotifier struct {
	host unsafe.Pointer
}

// NewAudioPortsConfigNotifier creates a new host audio ports config notifier
func NewAudioPortsConfigNotifier(host unsafe.Pointer) *AudioPortsConfigNotifier {
	return &AudioPortsConfigNotifier{host: host}
}

// Rescan tells the host to rescan the full list of configs
func (n *AudioPortsConfigNotifier) Rescan() {
	if n.host == nil {
		return
	}
	
	C.clap_host_audio_ports_config_rescan_helper((*C.clap_host_t)(n.host))
}

// VoiceInfoNotifier notifies the host about voice info changes
type VoiceInfoNotifier struct {
	host unsafe.Pointer
}

// NewVoiceInfoNotifier creates a new host voice info notifier
func NewVoiceInfoNotifier(host unsafe.Pointer) *VoiceInfoNotifier {
	return &VoiceInfoNotifier{host: host}
}

// Changed tells the host that voice info has changed
func (n *VoiceInfoNotifier) Changed() {
	if n.host == nil {
		return
	}
	
	C.clap_host_voice_info_changed_helper((*C.clap_host_t)(n.host))
}

// Color represents an RGBA color
type Color struct {
	Alpha uint8
	Red   uint8
	Green uint8
	Blue  uint8
}

// Track info flags
const (
	TrackInfoHasTrackName      = 1 << 0
	TrackInfoHasTrackColor     = 1 << 1
	TrackInfoHasAudioChannel   = 1 << 2
	TrackInfoIsForReturnTrack  = 1 << 3
	TrackInfoIsForBus          = 1 << 4
	TrackInfoIsForMaster       = 1 << 5
)

// TrackInfo contains information about the track
type TrackInfo struct {
	Flags             uint64
	Name              string
	Color             Color
	AudioChannelCount int32
	AudioPortType     string
}

// TrackInfoProvider provides track information from the host
type TrackInfoProvider struct {
	host unsafe.Pointer
}

// NewTrackInfoProvider creates a new host track info provider
func NewTrackInfoProvider(host unsafe.Pointer) *TrackInfoProvider {
	return &TrackInfoProvider{host: host}
}

// Get retrieves current track information from the host
func (t *TrackInfoProvider) Get() (*TrackInfo, bool) {
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

// TransportControl provides transport control functionality
type TransportControl struct {
	host unsafe.Pointer
}

// NewTransportControl creates a new host transport control
func NewTransportControl(host unsafe.Pointer) *TransportControl {
	return &TransportControl{host: host}
}

// RequestStart jumps back to the start point and starts the transport
func (t *TransportControl) RequestStart() {
	if t.host == nil {
		return
	}
	C.clap_host_transport_request_start((*C.clap_host_t)(t.host))
}

// RequestStop stops the transport and jumps to the start point
func (t *TransportControl) RequestStop() {
	if t.host == nil {
		return
	}
	C.clap_host_transport_request_stop((*C.clap_host_t)(t.host))
}

// RequestContinue starts the transport from its current position
func (t *TransportControl) RequestContinue() {
	if t.host == nil {
		return
	}
	C.clap_host_transport_request_continue((*C.clap_host_t)(t.host))
}

// RequestPause stops the transport at the current position
func (t *TransportControl) RequestPause() {
	if t.host == nil {
		return
	}
	C.clap_host_transport_request_pause((*C.clap_host_t)(t.host))
}

// RequestTogglePlay toggles play/pause (like space bar in most DAWs)
func (t *TransportControl) RequestTogglePlay() {
	if t.host == nil {
		return
	}
	C.clap_host_transport_request_toggle_play((*C.clap_host_t)(t.host))
}

// RequestJump jumps the transport to the given position (in beats)
func (t *TransportControl) RequestJump(position float64) {
	if t.host == nil {
		return
	}
	C.clap_host_transport_request_jump((*C.clap_host_t)(t.host), C.double(position))
}

// RequestLoopRegion sets the loop region
func (t *TransportControl) RequestLoopRegion(start, duration float64) {
	if t.host == nil {
		return
	}
	C.clap_host_transport_request_loop_region((*C.clap_host_t)(t.host), C.double(start), C.double(duration))
}

// RequestToggleLoop toggles looping on/off
func (t *TransportControl) RequestToggleLoop() {
	if t.host == nil {
		return
	}
	C.clap_host_transport_request_toggle_loop((*C.clap_host_t)(t.host))
}

// RequestEnableLoop enables or disables looping
func (t *TransportControl) RequestEnableLoop(enable bool) {
	if t.host == nil {
		return
	}
	C.clap_host_transport_request_enable_loop((*C.clap_host_t)(t.host), C.bool(enable))
}

// RequestRecord enables or disables recording
func (t *TransportControl) RequestRecord(record bool) {
	if t.host == nil {
		return
	}
	C.clap_host_transport_request_record((*C.clap_host_t)(t.host), C.bool(record))
}

// RequestToggleRecord toggles recording on/off
func (t *TransportControl) RequestToggleRecord() {
	if t.host == nil {
		return
	}
	C.clap_host_transport_request_toggle_record((*C.clap_host_t)(t.host))
}