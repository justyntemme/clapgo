package host

// #include "../../include/clap/include/clap/clap.h"
// #include <stdlib.h>
//
// // Helper function to request parameter flush from host
// static void clap_host_params_request_flush_helper(const clap_host_t* host) {
//     if (host && host->get_extension) {
//         const clap_host_params_t* params_ext = (const clap_host_params_t*)host->get_extension(host, CLAP_EXT_PARAMS);
//         if (params_ext && params_ext->request_flush) {
//             params_ext->request_flush(host);
//         }
//     }
// }
//
// // Helper function to notify host about parameter info changes
// static void clap_host_params_rescan_helper(const clap_host_t* host, uint32_t flags) {
//     if (host && host->get_extension) {
//         const clap_host_params_t* params_ext = (const clap_host_params_t*)host->get_extension(host, CLAP_EXT_PARAMS);
//         if (params_ext && params_ext->rescan) {
//             params_ext->rescan(host, flags);
//         }
//     }
// }
//
// // Helper function to clear parameter automation
// static void clap_host_params_clear_helper(const clap_host_t* host, clap_id param_id, uint32_t flags) {
//     if (host && host->get_extension) {
//         const clap_host_params_t* params_ext = (const clap_host_params_t*)host->get_extension(host, CLAP_EXT_PARAMS);
//         if (params_ext && params_ext->clear) {
//             params_ext->clear(host, param_id, flags);
//         }
//     }
// }
import "C"
import (
	"unsafe"
)

// ParamsHost provides methods to communicate parameter changes to the host
type ParamsHost struct {
	host unsafe.Pointer
}

// NewParamsHost creates a new host params interface
func NewParamsHost(host unsafe.Pointer) *ParamsHost {
	return &ParamsHost{host: host}
}

// RequestFlush requests the host to call the plugin's flush() method
// This is used when the plugin has parameter changes to report to the host
// This function is thread-safe but should not be called from audio thread
func (p *ParamsHost) RequestFlush() {
	if p.host != nil {
		C.clap_host_params_request_flush_helper((*C.clap_host_t)(p.host))
	}
}

// Rescan tells the host that parameter info has changed
// flags: combination of CLAP_PARAM_RESCAN_* flags
func (p *ParamsHost) Rescan(flags uint32) {
	if p.host != nil {
		C.clap_host_params_rescan_helper((*C.clap_host_t)(p.host), C.uint32_t(flags))
	}
}

// Clear clears parameter automation/modulation
// paramID: the parameter to clear
// flags: combination of CLAP_PARAM_CLEAR_* flags
func (p *ParamsHost) Clear(paramID uint32, flags uint32) {
	if p.host != nil {
		C.clap_host_params_clear_helper((*C.clap_host_t)(p.host), C.clap_id(paramID), C.uint32_t(flags))
	}
}

// Rescan flags
const (
	ParamRescanValues     = 1 << 0
	ParamRescanText       = 1 << 1
	ParamRescanInfo       = 1 << 2
	ParamRescanAll        = 1 << 3
)

// Clear flags
const (
	ParamClearAll         = 1 << 0
	ParamClearAutomations = 1 << 1
	ParamClearModulations = 1 << 2
)