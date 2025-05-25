package api

import (
	"fmt"
	"strings"
	"testing"
)

// BenchmarkLoggerPoolSimple tests logger pool with simple messages
func BenchmarkLoggerPoolSimple(b *testing.B) {
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		// Get buffer from pool
		buffer := globalLoggerPool.bufferPool.Get().(*LogBuffer)
		
		// Reset and write message
		buffer.builder.Reset()
		buffer.builder.WriteString("Simple log message")
		
		// Return to pool
		globalLoggerPool.bufferPool.Put(buffer)
	}
}

// BenchmarkLoggerPoolFormatted tests logger pool with formatted messages
func BenchmarkLoggerPoolFormatted(b *testing.B) {
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		// Get buffer from pool
		buffer := globalLoggerPool.bufferPool.Get().(*LogBuffer)
		
		// Reset and format message
		buffer.builder.Reset()
		fmt.Fprintf(&buffer.builder, "Parameter %d changed to %.3f", i, float64(i)/1000.0)
		
		// Return to pool
		globalLoggerPool.bufferPool.Put(buffer)
	}
}

// BenchmarkLoggerPoolLarge tests logger pool with large messages
func BenchmarkLoggerPoolLarge(b *testing.B) {
	// Create a large message (3KB)
	largeMsg := strings.Repeat("This is a test message. ", 128)
	
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		// Get buffer from pool
		buffer := globalLoggerPool.bufferPool.Get().(*LogBuffer)
		
		// Reset and write large message
		buffer.builder.Reset()
		buffer.builder.WriteString(largeMsg)
		
		// Return to pool
		globalLoggerPool.bufferPool.Put(buffer)
	}
}

// BenchmarkLoggerPoolConcurrent tests concurrent access to logger pool
func BenchmarkLoggerPoolConcurrent(b *testing.B) {
	b.ReportAllocs()
	
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// Get buffer from pool
			buffer := globalLoggerPool.bufferPool.Get().(*LogBuffer)
			
			// Reset and write message
			buffer.builder.Reset()
			switch i % 4 {
			case 0:
				buffer.builder.WriteString("Debug message")
			case 1:
				buffer.builder.WriteString("Info message")
			case 2:
				buffer.builder.WriteString("Warning message")
			case 3:
				buffer.builder.WriteString("Error message")
			}
			
			// Return to pool
			globalLoggerPool.bufferPool.Put(buffer)
			i++
		}
	})
}