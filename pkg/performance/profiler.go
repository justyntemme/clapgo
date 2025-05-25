//go:build debug
// +build debug

package performance

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

// Profiler manages CPU and memory profiling for performance analysis
type Profiler struct {
	cpuFile    *os.File
	memTicker  *time.Ticker
	stopChan   chan struct{}
	memFileIdx int
}

// NewProfiler creates a new profiler instance
func NewProfiler() *Profiler {
	return &Profiler{
		stopChan: make(chan struct{}),
	}
}

// StartCPUProfile starts CPU profiling
func (p *Profiler) StartCPUProfile(filename string) error {
	if p.cpuFile != nil {
		return fmt.Errorf("CPU profiling already started")
	}
	
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CPU profile file: %w", err)
	}
	
	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		return fmt.Errorf("failed to start CPU profile: %w", err)
	}
	
	p.cpuFile = f
	log.Printf("Started CPU profiling to %s", filename)
	return nil
}

// StopCPUProfile stops CPU profiling
func (p *Profiler) StopCPUProfile() error {
	if p.cpuFile == nil {
		return fmt.Errorf("CPU profiling not started")
	}
	
	pprof.StopCPUProfile()
	err := p.cpuFile.Close()
	p.cpuFile = nil
	
	if err != nil {
		return fmt.Errorf("failed to close CPU profile file: %w", err)
	}
	
	log.Println("Stopped CPU profiling")
	return nil
}

// StartMemoryProfiling starts periodic memory profiling
func (p *Profiler) StartMemoryProfiling(interval time.Duration, prefix string) {
	if p.memTicker != nil {
		log.Println("Memory profiling already started")
		return
	}
	
	p.memTicker = time.NewTicker(interval)
	
	go func() {
		for {
			select {
			case <-p.memTicker.C:
				p.captureMemoryProfile(prefix)
			case <-p.stopChan:
				return
			}
		}
	}()
	
	log.Printf("Started memory profiling every %v", interval)
}

// StopMemoryProfiling stops periodic memory profiling
func (p *Profiler) StopMemoryProfiling() {
	if p.memTicker == nil {
		return
	}
	
	p.memTicker.Stop()
	p.memTicker = nil
	close(p.stopChan)
	
	log.Println("Stopped memory profiling")
}

// CaptureMemoryProfile captures a single memory profile
func (p *Profiler) captureMemoryProfile(prefix string) {
	filename := fmt.Sprintf("%s_mem_%d_%d.prof", prefix, time.Now().Unix(), p.memFileIdx)
	p.memFileIdx++
	
	f, err := os.Create(filename)
	if err != nil {
		log.Printf("Failed to create memory profile file: %v", err)
		return
	}
	defer f.Close()
	
	// Force GC to get accurate stats
	runtime.GC()
	
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Printf("Failed to write memory profile: %v", err)
		return
	}
	
	log.Printf("Captured memory profile to %s", filename)
}

// CaptureGoroutineProfile captures goroutine profile
func (p *Profiler) CaptureGoroutineProfile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create goroutine profile file: %w", err)
	}
	defer f.Close()
	
	if err := pprof.Lookup("goroutine").WriteTo(f, 0); err != nil {
		return fmt.Errorf("failed to write goroutine profile: %w", err)
	}
	
	log.Printf("Captured goroutine profile to %s", filename)
	return nil
}

// PrintMemoryStats prints current memory statistics
func (p *Profiler) PrintMemoryStats() {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	
	log.Printf("Memory Stats:")
	log.Printf("  Alloc: %d MB", stats.Alloc/1024/1024)
	log.Printf("  TotalAlloc: %d MB", stats.TotalAlloc/1024/1024)
	log.Printf("  HeapAlloc: %d MB", stats.HeapAlloc/1024/1024)
	log.Printf("  HeapObjects: %d", stats.HeapObjects)
	log.Printf("  NumGC: %d", stats.NumGC)
	log.Printf("  LastGC: %v ago", time.Since(time.Unix(0, int64(stats.LastGC))))
}