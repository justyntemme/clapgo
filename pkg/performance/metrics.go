package performance

import (
	"sync/atomic"
	"time"
)

// PerformanceMetrics tracks real-time performance metrics for audio processing
type PerformanceMetrics struct {
	// Process timing
	processTime        int64 // Last process call duration in nanoseconds (atomic)
	maxProcessTime     int64 // Worst case duration (atomic)
	totalProcessTime   int64 // Total time spent processing (atomic)
	processCallCount   uint64 // Number of process calls (atomic)
	
	// Audio performance
	bufferUnderruns    uint64 // Audio dropouts (atomic)
	gcPausesDuringProc uint64 // GC events during processing (atomic)
	
	// Voice usage statistics
	maxVoicesUsed      int32  // Maximum voices used (atomic)
	currentVoicesUsed  int32  // Current voice count (atomic)
	voiceStealEvents   uint64 // Voice stealing events (atomic)
	
	// Event processing stats
	eventsProcessed    uint64 // Total events processed (atomic)
	maxEventsPerBuffer uint64 // Maximum events in single buffer (atomic)
	currentBufferEvents uint64 // Events in current buffer (atomic)
	
	// Configuration
	sampleRate uint32
	frameCount uint32
}

// NewPerformanceMetrics creates a new performance metrics tracker
func NewPerformanceMetrics(sampleRate, frameCount uint32) *PerformanceMetrics {
	return &PerformanceMetrics{
		sampleRate: sampleRate,
		frameCount: frameCount,
	}
}

// StartProcess marks the beginning of audio processing
func (pm *PerformanceMetrics) StartProcess() time.Time {
	return time.Now()
}

// EndProcess marks the end of audio processing and updates metrics
func (pm *PerformanceMetrics) EndProcess(startTime time.Time) {
	duration := time.Since(startTime).Nanoseconds()
	
	// Update process time
	atomic.StoreInt64(&pm.processTime, duration)
	
	// Update max process time if needed
	for {
		max := atomic.LoadInt64(&pm.maxProcessTime)
		if duration <= max {
			break
		}
		if atomic.CompareAndSwapInt64(&pm.maxProcessTime, max, duration) {
			break
		}
	}
	
	// Update totals
	atomic.AddInt64(&pm.totalProcessTime, duration)
	atomic.AddUint64(&pm.processCallCount, 1)
	
	// Check for buffer underruns (exceeded deadline)
	bufferDuration := int64(pm.frameCount) * int64(time.Second) / int64(pm.sampleRate)
	threshold := bufferDuration * 80 / 100 // 80% threshold
	
	if duration > threshold {
		atomic.AddUint64(&pm.bufferUnderruns, 1)
	}
	
	// Update max events per buffer
	current := atomic.LoadUint64(&pm.currentBufferEvents)
	for {
		max := atomic.LoadUint64(&pm.maxEventsPerBuffer)
		if current <= max {
			break
		}
		if atomic.CompareAndSwapUint64(&pm.maxEventsPerBuffer, max, current) {
			break
		}
	}
	
	// Reset current buffer events
	atomic.StoreUint64(&pm.currentBufferEvents, 0)
}

// RecordEvent increments the event counter
func (pm *PerformanceMetrics) RecordEvent() {
	atomic.AddUint64(&pm.eventsProcessed, 1)
	atomic.AddUint64(&pm.currentBufferEvents, 1)
}

// RecordGCPause increments the GC pause counter
func (pm *PerformanceMetrics) RecordGCPause() {
	atomic.AddUint64(&pm.gcPausesDuringProc, 1)
}

// UpdateVoiceCount updates the current voice count
func (pm *PerformanceMetrics) UpdateVoiceCount(count int32) {
	atomic.StoreInt32(&pm.currentVoicesUsed, count)
	
	// Update max voices if needed
	for {
		max := atomic.LoadInt32(&pm.maxVoicesUsed)
		if count <= max {
			break
		}
		if atomic.CompareAndSwapInt32(&pm.maxVoicesUsed, max, count) {
			break
		}
	}
}

// RecordVoiceSteal increments the voice steal counter
func (pm *PerformanceMetrics) RecordVoiceSteal() {
	atomic.AddUint64(&pm.voiceStealEvents, 1)
}

// GetStats returns current performance statistics
func (pm *PerformanceMetrics) GetStats() PerformanceStats {
	processCount := atomic.LoadUint64(&pm.processCallCount)
	totalTime := atomic.LoadInt64(&pm.totalProcessTime)
	
	avgProcessTime := int64(0)
	if processCount > 0 {
		avgProcessTime = totalTime / int64(processCount)
	}
	
	return PerformanceStats{
		ProcessTime:        time.Duration(atomic.LoadInt64(&pm.processTime)),
		MaxProcessTime:     time.Duration(atomic.LoadInt64(&pm.maxProcessTime)),
		AvgProcessTime:     time.Duration(avgProcessTime),
		ProcessCallCount:   processCount,
		BufferUnderruns:    atomic.LoadUint64(&pm.bufferUnderruns),
		GCPausesDuringProc: atomic.LoadUint64(&pm.gcPausesDuringProc),
		MaxVoicesUsed:      atomic.LoadInt32(&pm.maxVoicesUsed),
		CurrentVoicesUsed:  atomic.LoadInt32(&pm.currentVoicesUsed),
		VoiceStealEvents:   atomic.LoadUint64(&pm.voiceStealEvents),
		EventsProcessed:    atomic.LoadUint64(&pm.eventsProcessed),
		MaxEventsPerBuffer: atomic.LoadUint64(&pm.maxEventsPerBuffer),
	}
}

// Reset resets all performance statistics
func (pm *PerformanceMetrics) Reset() {
	atomic.StoreInt64(&pm.processTime, 0)
	atomic.StoreInt64(&pm.maxProcessTime, 0)
	atomic.StoreInt64(&pm.totalProcessTime, 0)
	atomic.StoreUint64(&pm.processCallCount, 0)
	atomic.StoreUint64(&pm.bufferUnderruns, 0)
	atomic.StoreUint64(&pm.gcPausesDuringProc, 0)
	atomic.StoreInt32(&pm.maxVoicesUsed, 0)
	atomic.StoreInt32(&pm.currentVoicesUsed, 0)
	atomic.StoreUint64(&pm.voiceStealEvents, 0)
	atomic.StoreUint64(&pm.eventsProcessed, 0)
	atomic.StoreUint64(&pm.maxEventsPerBuffer, 0)
	atomic.StoreUint64(&pm.currentBufferEvents, 0)
}

// PerformanceStats contains performance statistics
type PerformanceStats struct {
	// Timing
	ProcessTime      time.Duration
	MaxProcessTime   time.Duration
	AvgProcessTime   time.Duration
	ProcessCallCount uint64
	
	// Audio performance
	BufferUnderruns    uint64
	GCPausesDuringProc uint64
	
	// Voice usage
	MaxVoicesUsed     int32
	CurrentVoicesUsed int32
	VoiceStealEvents  uint64
	
	// Event processing
	EventsProcessed    uint64
	MaxEventsPerBuffer uint64
}