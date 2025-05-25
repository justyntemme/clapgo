package main

// This example demonstrates how to implement the POSIX FD Support extension
// in a ClapGo plugin. This extension allows plugins to integrate with the
// host's event loop for asynchronous I/O operations.

import "C"
import (
	"github.com/justyntemme/clapgo/pkg/api"
	"net"
	"os"
	"unsafe"
)

// Example plugin with POSIX FD support
type PluginWithPosixFDSupport struct {
	api.BasePlugin
	
	// FD manager for handling file descriptors
	fdManager *api.PosixFDManager
	
	// Example: network connection for remote control
	conn net.Conn
	
	// Example: pipe for inter-thread communication
	pipeRead, pipeWrite *os.File
}

// Initialize sets up the plugin and registers file descriptors
func (p *PluginWithPosixFDSupport) Initialize(host unsafe.Pointer) bool {
	// Create FD manager
	p.fdManager = api.NewPosixFDManager(host)
	
	if !p.fdManager.IsSupported() {
		// Host doesn't support POSIX FD extension
		// Fall back to polling or other mechanisms
		return true
	}
	
	// Example 1: Set up a pipe for inter-thread communication
	var err error
	p.pipeRead, p.pipeWrite, err = os.Pipe()
	if err == nil {
		// Register the read end of the pipe
		fd := int(p.pipeRead.Fd())
		p.fdManager.RegisterReadFD(fd)
	}
	
	// Example 2: Set up a network listener for remote control
	// (In a real plugin, you'd do this based on user configuration)
	listener, err := net.Listen("tcp", "localhost:0")
	if err == nil {
		// For a listener, we typically use a goroutine instead
		// of registering with the event loop
		go p.acceptConnections(listener)
	}
	
	return true
}

// Implement the PosixFDSupportProvider interface
func (p *PluginWithPosixFDSupport) OnFD(fd int, flags uint32) {
	// Handle events based on the file descriptor
	
	if p.pipeRead != nil && fd == int(p.pipeRead.Fd()) {
		if flags&api.PosixFDRead != 0 {
			// Data available to read from pipe
			buf := make([]byte, 1024)
			n, err := p.pipeRead.Read(buf)
			if err == nil && n > 0 {
				// Process the data
				p.handlePipeMessage(buf[:n])
			}
		}
		
		if flags&api.PosixFDError != 0 {
			// Handle error condition
			p.pipeRead.Close()
			p.fdManager.UnregisterFD(fd)
		}
	}
	
	// Handle other file descriptors...
	if p.conn != nil {
		if file, ok := p.conn.(*net.TCPConn); ok {
			connFD := int(file.Fd())
			if fd == connFD {
				p.handleNetworkEvents(fd, flags)
			}
		}
	}
}

func (p *PluginWithPosixFDSupport) handlePipeMessage(data []byte) {
	// Process messages from other threads
	// For example, parameter changes from a UI thread
}

func (p *PluginWithPosixFDSupport) handleNetworkEvents(fd int, flags uint32) {
	if flags&api.PosixFDRead != 0 {
		// Read data from network connection
		buf := make([]byte, 4096)
		n, err := p.conn.Read(buf)
		if err != nil {
			// Connection closed or error
			p.conn.Close()
			p.fdManager.UnregisterFD(fd)
			p.conn = nil
			return
		}
		
		// Process received data
		p.processNetworkCommand(buf[:n])
	}
	
	if flags&api.PosixFDWrite != 0 {
		// Socket is ready for writing
		// In a real implementation, you'd have a write buffer
		// and write pending data here
		
		// Once done writing, disable write notifications
		p.fdManager.DisableWrite(fd)
	}
}

func (p *PluginWithPosixFDSupport) processNetworkCommand(data []byte) {
	// Process remote control commands
	// This could be OSC, custom protocol, etc.
}

func (p *PluginWithPosixFDSupport) acceptConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		
		// Handle new connection
		// In this example, we only support one connection at a time
		if p.conn != nil {
			conn.Close()
			continue
		}
		
		p.conn = conn
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			fd := int(tcpConn.Fd())
			// Register for read events initially
			p.fdManager.RegisterFD(fd, api.PosixFDRead)
		}
	}
}

// Cleanup unregisters all file descriptors
func (p *PluginWithPosixFDSupport) Cleanup() {
	if p.fdManager != nil {
		p.fdManager.UnregisterAll()
	}
	
	if p.pipeRead != nil {
		p.pipeRead.Close()
	}
	if p.pipeWrite != nil {
		p.pipeWrite.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}

// Required export for the extension

//export ClapGo_PluginPosixFDSupportOnFD
func ClapGo_PluginPosixFDSupportOnFD(plugin unsafe.Pointer, fd C.int, flags C.uint32_t) {
	if plugin == nil {
		return
	}
	
	p := (*PluginWithPosixFDSupport)(plugin)
	p.OnFD(int(fd), uint32(flags))
}

// Other required plugin exports would go here...

func main() {
	// Required for c-shared build
}