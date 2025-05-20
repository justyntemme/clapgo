package goclap

// #include <stdlib.h>
// #include <stdint.h>
// #include "../../include/clap/include/clap/clap.h"
import "C"
import (
	"fmt"
	"runtime/cgo"
	"unsafe"
)

// getProcessorFromHandle retrieves the AudioProcessor from a cgo.Handle
func getProcessorFromHandle(handlePtr unsafe.Pointer) AudioProcessor {
	if handlePtr == nil {
		return nil
	}

	// Convert the opaque pointer back to a handle
	handle := cgo.Handle(uintptr(handlePtr))
	
	// Get the processor from the handle
	value := handle.Value()
	
	// Try to cast to AudioProcessor
	processor, ok := value.(AudioProcessor)
	if !ok {
		fmt.Println("Error: Failed to cast handle value to AudioProcessor")
		return nil
	}
	
	return processor
}