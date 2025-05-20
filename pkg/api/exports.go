// Package api defines the core interfaces for CLAP plugins in Go.
package api

// This file previously contained the plugin metadata registry and export functions.
// These have been moved to the bridge package since they need to be exported from
// the main package that is built into the shared library.