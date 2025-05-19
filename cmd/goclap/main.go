package main

import "C"

// We need to import the goclap package to ensure its exported functions are included
// in the shared library
import _ "github.com/justyntemme/clapgo/src/goclap"

// This main function is required for a c-shared build but is not executed
// when the shared library is loaded
func main() {}

