package param

// #include "../../include/clap/include/clap/clap.h"
// #include <string.h>
import "C"
import (
	"unsafe"
)

// CLAP parameter flags mapping
const (
	ClapParamIsAutomatable   = C.CLAP_PARAM_IS_AUTOMATABLE
	ClapParamIsModulatable   = C.CLAP_PARAM_IS_MODULATABLE
	ClapParamIsStepped       = C.CLAP_PARAM_IS_STEPPED
	ClapParamIsReadonly      = C.CLAP_PARAM_IS_READONLY
	ClapParamIsHidden        = C.CLAP_PARAM_IS_HIDDEN
	ClapParamIsBypass        = C.CLAP_PARAM_IS_BYPASS
	ClapParamRequiresProcess = C.CLAP_PARAM_REQUIRES_PROCESS
)

// InfoToC converts a Go Info struct to a C clap_param_info_t struct
func InfoToC(info Info, cInfo unsafe.Pointer) {
	clapInfo := (*C.clap_param_info_t)(cInfo)
	
	// Set basic fields
	clapInfo.id = C.clap_id(info.ID)
	clapInfo.flags = 0
	
	// Map flags
	if info.Flags&FlagAutomatable != 0 {
		clapInfo.flags |= ClapParamIsAutomatable
	}
	if info.Flags&FlagModulatable != 0 {
		clapInfo.flags |= ClapParamIsModulatable
	}
	if info.Flags&FlagStepped != 0 {
		clapInfo.flags |= ClapParamIsStepped
	}
	if info.Flags&FlagReadonly != 0 {
		clapInfo.flags |= ClapParamIsReadonly
	}
	if info.Flags&FlagHidden != 0 {
		clapInfo.flags |= ClapParamIsHidden
	}
	if info.Flags&FlagBypass != 0 {
		clapInfo.flags |= ClapParamIsBypass
	}
	if info.Flags&FlagRequiresProcess != 0 {
		clapInfo.flags |= ClapParamRequiresProcess
	}
	
	clapInfo.cookie = nil
	
	// Copy name
	copyStringToCBuffer(info.Name, unsafe.Pointer(&clapInfo.name[0]), C.CLAP_NAME_SIZE)
	
	// Copy module path if present
	if info.Module != "" {
		copyStringToCBuffer(info.Module, unsafe.Pointer(&clapInfo.module[0]), C.CLAP_PATH_SIZE)
	} else {
		clapInfo.module[0] = 0
	}
	
	// Set range
	clapInfo.min_value = C.double(info.MinValue)
	clapInfo.max_value = C.double(info.MaxValue)
	clapInfo.default_value = C.double(info.DefaultValue)
}

// copyStringToCBuffer safely copies a Go string to a C char buffer
func copyStringToCBuffer(str string, buffer unsafe.Pointer, maxSize int) {
	bytes := []byte(str)
	if len(bytes) >= maxSize {
		bytes = bytes[:maxSize-1]
	}
	
	// Copy bytes to C buffer
	for i, b := range bytes {
		*(*C.char)(unsafe.Add(buffer, i)) = C.char(b)
	}
	
	// Null terminate
	*(*C.char)(unsafe.Add(buffer, len(bytes))) = 0
}