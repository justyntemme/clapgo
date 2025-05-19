package goclap

// #include <stdint.h>
import "C"

// API version constants
const (
    APIVersionMajor = 0
    APIVersionMinor = 1
    APIVersionPatch = 0
)

// GetVersionImpl returns the API version of the Go plugin
func GetVersionImpl() (uint32, uint32, uint32) {
    return APIVersionMajor, APIVersionMinor, APIVersionPatch
}