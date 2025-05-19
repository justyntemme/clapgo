package goclap

// #include <stdint.h>
import "C"

// API version constants
const (
    APIVersionMajor = 0
    APIVersionMinor = 1
    APIVersionPatch = 0
)

// GetVersion returns the API version of the Go plugin
//export GetVersion
func GetVersion(major, minor, patch *C.uint32_t) C.bool {
    if major != nil {
        *major = C.uint32_t(APIVersionMajor)
    }
    
    if minor != nil {
        *minor = C.uint32_t(APIVersionMinor)
    }
    
    if patch != nil {
        *patch = C.uint32_t(APIVersionPatch)
    }
    
    return C.bool(true)
}