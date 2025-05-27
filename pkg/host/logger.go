package host

import (
	"fmt"
	"unsafe"
)

// #include "../../include/clap/include/clap/clap.h"
// #include <stdlib.h>
//
// static void clap_host_log_helper(const clap_host_t* host, int severity, const char* msg) {
//     if (host && host->get_extension) {
//         const clap_host_log_t* log_ext = (const clap_host_log_t*)host->get_extension(host, CLAP_EXT_LOG);
//         if (log_ext && log_ext->log) {
//             log_ext->log(host, severity, msg);
//         }
//     }
// }
//
// static const clap_host_log_t* clap_host_get_log_extension(const clap_host_t* host) {
//     if (host && host->get_extension) {
//         return (const clap_host_log_t*)host->get_extension(host, CLAP_EXT_LOG);
//     }
//     return NULL;
// }
import "C"

// Severity levels for logging
const (
	SeverityDebug   int32 = C.CLAP_LOG_DEBUG
	SeverityInfo    int32 = C.CLAP_LOG_INFO
	SeverityWarning int32 = C.CLAP_LOG_WARNING
	SeverityError   int32 = C.CLAP_LOG_ERROR
	SeverityFatal   int32 = C.CLAP_LOG_FATAL
)

// Logger provides structured logging to the host
type Logger struct {
	host unsafe.Pointer
}

// NewLogger creates a new host logger
func NewLogger(host unsafe.Pointer) *Logger {
	if host == nil {
		return nil
	}
	
	clapHost := (*C.clap_host_t)(host)
	ext := C.clap_host_get_log_extension(clapHost)
	
	if ext == nil {
		return nil
	}
	
	return &Logger{
		host: host,
	}
}

// log sends a message to the host with the specified severity
func (l *Logger) log(severity int32, message string) {
	if l == nil || l.host == nil {
		return
	}
	
	clapHost := (*C.clap_host_t)(l.host)
	cMessage := C.CString(message)
	defer C.free(unsafe.Pointer(cMessage))
	
	C.clap_host_log_helper(clapHost, C.int(severity), cMessage)
}

// Debug logs a debug message
func (l *Logger) Debug(message string) {
	l.log(SeverityDebug, message)
}

// Info logs an info message
func (l *Logger) Info(message string) {
	l.log(SeverityInfo, message)
}

// Warning logs a warning message
func (l *Logger) Warning(message string) {
	l.log(SeverityWarning, message)
}

// Error logs an error message
func (l *Logger) Error(message string) {
	l.log(SeverityError, message)
}

// Fatal logs a fatal message
func (l *Logger) Fatal(message string) {
	l.log(SeverityFatal, message)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

// Warningf logs a formatted warning message
func (l *Logger) Warningf(format string, args ...interface{}) {
	l.Warning(fmt.Sprintf(format, args...))
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

// Fatalf logs a formatted fatal message
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Fatal(fmt.Sprintf(format, args...))
}

// Log is a generic method that takes severity and message
func (l *Logger) Log(severity int32, message string, args ...interface{}) {
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}
	l.log(severity, message)
}