# ClapGo Runtime and Allocation Considerations

This document discusses the challenges and mitigation strategies for using Go in real-time audio processing, with particular focus on note processing, garbage collection, and memory allocation patterns.

## Table of Contents

1. [Go Runtime Challenges in Real-Time Audio](#go-runtime-challenges-in-real-time-audio)
2. [Garbage Collection Mitigation](#garbage-collection-mitigation)
3. [Note Processing Considerations](#note-processing-considerations)
4. [Memory Allocation Strategies](#memory-allocation-strategies)
5. [Thread Safety and Concurrency](#thread-safety-and-concurrency)
6. [Performance Monitoring and Debugging](#performance-monitoring-and-debugging)
7. [Best Practices and Guidelines](#best-practices-and-guidelines)

## Go Runtime Challenges in Real-Time Audio

### The Real-Time Problem

Real-time audio processing requires:
- **Deterministic timing**: Audio buffers must be processed within strict deadlines (typically 1-10ms)
- **No blocking operations**: Any pause can cause audio dropouts (pops, clicks, silence)
- **Consistent latency**: Timing jitter affects audio quality
- **High frequency processing**: Audio callbacks occur 100-1000 times per second

### Go Runtime Issues

1. **Garbage Collector Pauses**
   ```go
   // PROBLEM: This allocation can trigger GC during audio processing
   func (s *Synth) ProcessNote(note NoteEvent) {
       voice := &Voice{  // Heap allocation!
           noteID: note.NoteID,
           frequency: midiToFreq(note.Key),
       }
       s.voices = append(s.voices, voice)  // Potential slice reallocation!
   }
   ```

2. **Memory Allocator Contention**
   - Go's allocator can block during heap operations
   - Interface{} boxing causes unexpected allocations
   - String operations often allocate

3. **Goroutine Scheduler Interference**
   - Scheduler can preempt audio thread
   - System calls can block unexpectedly
   - Timer resolution limitations

## Garbage Collection Mitigation

### 1. Minimize Allocations in Audio Thread

**Strategy**: Pre-allocate everything at initialization

```go
// GOOD: Pre-allocated voice pool
type VoicePool struct {
    voices     [MaxVoices]*Voice
    freeList   [MaxVoices]int
    freeCount  int
    maxVoices  int
}

func (vp *VoicePool) AllocateVoice() *Voice {
    if vp.freeCount == 0 {
        return nil // No allocation, graceful degradation
    }
    
    vp.freeCount--
    idx := vp.freeList[vp.freeCount]
    voice := vp.voices[idx]
    voice.Reset() // Clear previous state
    return voice
}

func (vp *VoicePool) ReleaseVoice(voice *Voice) {
    // Find index and return to free list
    for i, v := range vp.voices {
        if v == voice {
            vp.freeList[vp.freeCount] = i
            vp.freeCount++
            break
        }
    }
}
```

### 2. Use Fixed-Size Data Structures

```go
// GOOD: Fixed arrays instead of slices
type Synth struct {
    voices     [MaxVoices]Voice          // Fixed array
    activeVoices [MaxVoices]int          // Active voice indices
    voiceCount int                       // Number of active voices
    
    // Pre-allocated buffers
    tempBuffer [MaxFrames]float32
    noteEvents [MaxEvents]NoteEvent
}

// BAD: Dynamic slices that can reallocate
type BadSynth struct {
    voices []Voice              // Can grow and cause allocations
    buffer []float32            // Reallocates on resize
}
```

### 3. Avoid Interface{} in Hot Paths

```go
// BAD: Interface boxing allocates
func processEvent(event interface{}) {
    switch e := event.(type) {  // Boxing allocation
    case NoteEvent:
        // process note
    case ParamEvent:
        // process param
    }
}

// GOOD: Type-specific processing
func processNoteEvent(event *NoteEvent) {
    // Direct access, no boxing
}

func processParamEvent(event *ParamEvent) {
    // Direct access, no boxing
}
```

### 4. Control GC Timing

```go
// In plugin initialization
func (p *Plugin) Init() bool {
    // Tune GC for real-time workload
    debug.SetGCPercent(800)  // Reduce GC frequency
    
    // Force GC at safe times (not in audio thread)
    go func() {
        ticker := time.NewTicker(100 * time.Millisecond)
        for range ticker.C {
            if !p.isProcessing {
                runtime.GC()  // Only when not processing audio
            }
        }
    }()
    
    return true
}
```

## Note Processing Considerations

### Voice Management Challenges

Note processing presents unique challenges:

1. **Dynamic voice allocation**: Notes can start/stop unpredictably
2. **Polyphonic complexity**: Multiple simultaneous notes
3. **Voice stealing**: Limited polyphony requires smart allocation
4. **Note expression**: MPE and per-note parameters

### Voice Pool Architecture

```go
type Voice struct {
    // Voice state
    noteID    int32
    key       int16
    velocity  float64
    
    // Synthesis state
    phase     float64
    envelope  EnvelopeState
    
    // Per-voice parameters (pre-allocated)
    volume    float64
    pitch     float64
    brightness float64
    pressure  float64
    
    // State tracking
    isActive  bool
    age       uint64  // For voice stealing algorithms
}

type EnvelopeState struct {
    stage     EnvelopeStage
    level     float64
    time      float64
    
    // ADSR parameters (cached for performance)
    attack    float64
    decay     float64
    sustain   float64
    release   float64
}

// Voice stealing with minimal allocation
func (s *Synth) StealVoice() *Voice {
    var oldestVoice *Voice
    var oldestAge uint64 = 0
    
    // Find oldest voice in release phase
    for i := 0; i < s.voiceCount; i++ {
        voice := &s.voices[s.activeVoices[i]]
        if voice.envelope.stage == StageRelease && voice.age > oldestAge {
            oldestVoice = voice
            oldestAge = voice.age
        }
    }
    
    if oldestVoice != nil {
        return oldestVoice
    }
    
    // Fall back to oldest voice
    for i := 0; i < s.voiceCount; i++ {
        voice := &s.voices[s.activeVoices[i]]
        if voice.age > oldestAge {
            oldestVoice = voice
            oldestAge = voice.age
        }
    }
    
    return oldestVoice
}
```

### Note Event Batching

```go
// Process all note events for a buffer at once
func (s *Synth) ProcessNoteEvents(events []NoteEvent, frameCount uint32) {
    // Sort events by timestamp for sample-accurate processing
    // (Use pre-allocated sort buffer to avoid allocation)
    s.sortEvents(events)
    
    currentFrame := uint32(0)
    eventIndex := 0
    
    for currentFrame < frameCount {
        // Process events at current frame
        for eventIndex < len(events) && events[eventIndex].Time == currentFrame {
            s.processNoteEvent(&events[eventIndex])
            eventIndex++
        }
        
        // Find next event time or end of buffer
        nextEventTime := frameCount
        if eventIndex < len(events) {
            nextEventTime = events[eventIndex].Time
        }
        
        // Process audio from currentFrame to nextEventTime
        framesToProcess := nextEventTime - currentFrame
        s.processAudio(currentFrame, framesToProcess)
        
        currentFrame = nextEventTime
    }
}
```

## Memory Allocation Strategies

### 1. Object Pools for Variable-Size Data

```go
type BufferPool struct {
    small  sync.Pool  // <= 64 bytes
    medium sync.Pool  // <= 1KB  
    large  sync.Pool  // <= 64KB
}

func (bp *BufferPool) Get(size int) []byte {
    switch {
    case size <= 64:
        if buf := bp.small.Get(); buf != nil {
            return buf.([]byte)[:size]
        }
        return make([]byte, size, 64)
    case size <= 1024:
        if buf := bp.medium.Get(); buf != nil {
            return buf.([]byte)[:size]
        }
        return make([]byte, size, 1024)
    case size <= 65536:
        if buf := bp.large.Get(); buf != nil {
            return buf.([]byte)[:size]
        }
        return make([]byte, size, 65536)
    default:
        // Large allocation, handle separately
        return make([]byte, size)
    }
}
```

### 2. Lock-Free Ring Buffers

```go
// Lock-free ring buffer for parameter updates
type RingBuffer struct {
    buffer    [BufferSize]ParamUpdate
    writePos  uint64
    readPos   uint64
}

func (rb *RingBuffer) Write(update ParamUpdate) bool {
    current := atomic.LoadUint64(&rb.writePos)
    next := (current + 1) % BufferSize
    
    if next == atomic.LoadUint64(&rb.readPos) {
        return false // Buffer full
    }
    
    rb.buffer[current] = update
    atomic.StoreUint64(&rb.writePos, next)
    return true
}

func (rb *RingBuffer) Read() (ParamUpdate, bool) {
    current := atomic.LoadUint64(&rb.readPos)
    
    if current == atomic.LoadUint64(&rb.writePos) {
        return ParamUpdate{}, false // Buffer empty
    }
    
    update := rb.buffer[current]
    atomic.StoreUint64(&rb.readPos, (current+1)%BufferSize)
    return update, true
}
```

### 3. Custom Allocators for Specific Use Cases

```go
// Stack allocator for temporary data in audio processing
type StackAllocator struct {
    buffer []byte
    offset int
    marks  []int  // Stack of saved positions
}

func (sa *StackAllocator) Alloc(size int) []byte {
    if sa.offset+size > len(sa.buffer) {
        panic("StackAllocator: out of memory")
    }
    
    result := sa.buffer[sa.offset : sa.offset+size]
    sa.offset += size
    return result
}

func (sa *StackAllocator) Mark() int {
    mark := sa.offset
    sa.marks = append(sa.marks, mark)
    return mark
}

func (sa *StackAllocator) Release(mark int) {
    sa.offset = mark
    // Remove mark from stack
    for i := len(sa.marks) - 1; i >= 0; i-- {
        if sa.marks[i] == mark {
            sa.marks = sa.marks[:i]
            break
        }
    }
}
```

## Thread Safety and Concurrency

### Main Thread vs Audio Thread Separation

```go
type Plugin struct {
    // Audio thread data (no locks needed)
    audioState struct {
        voices      [MaxVoices]Voice
        parameters  [NumParams]float64  // Atomic access
        sampleRate  float64
    }
    
    // Main thread data (protected by mutex)
    mainState struct {
        mu          sync.RWMutex
        presets     []Preset
        userParams  map[string]interface{}
    }
    
    // Communication channels
    paramUpdates chan ParamUpdate
    noteEvents   chan NoteEvent
}

// Safe parameter update from main thread
func (p *Plugin) SetParameter(id uint32, value float64) {
    // Atomic write for audio thread
    atomic.StoreUint64((*uint64)(&p.audioState.parameters[id]), 
                      math.Float64bits(value))
    
    // Optional: notify main thread
    select {
    case p.paramUpdates <- ParamUpdate{ID: id, Value: value}:
    default: // Don't block if channel is full
    }
}

// Audio thread parameter read
func (p *Plugin) GetParameter(id uint32) float64 {
    bits := atomic.LoadUint64((*uint64)(&p.audioState.parameters[id]))
    return math.Float64frombits(bits)
}
```

### Lock-Free Data Structures

```go
// Lock-free parameter automation
type AutomationPoint struct {
    time  uint64
    value float64
}

type ParameterAutomation struct {
    points    [MaxAutomationPoints]AutomationPoint
    count     int64  // Atomic
    readIndex int64  // Atomic
}

func (pa *ParameterAutomation) AddPoint(time uint64, value float64) bool {
    count := atomic.LoadInt64(&pa.count)
    if count >= MaxAutomationPoints {
        return false
    }
    
    // Simple append - more complex sorting would require RCU
    pa.points[count] = AutomationPoint{time: time, value: value}
    atomic.StoreInt64(&pa.count, count+1)
    return true
}

func (pa *ParameterAutomation) GetValueAtTime(time uint64) float64 {
    count := atomic.LoadInt64(&pa.count)
    if count == 0 {
        return 0.0
    }
    
    // Linear search (could be optimized with binary search)
    for i := int64(0); i < count-1; i++ {
        if pa.points[i].time <= time && time < pa.points[i+1].time {
            // Linear interpolation
            t0, v0 := pa.points[i].time, pa.points[i].value
            t1, v1 := pa.points[i+1].time, pa.points[i+1].value
            
            factor := float64(time-t0) / float64(t1-t0)
            return v0 + factor*(v1-v0)
        }
    }
    
    // Return last value if time is beyond all points
    return pa.points[count-1].value
}
```

## Performance Monitoring and Debugging

### Allocation Tracking

```go
// Add to plugin for allocation monitoring
type AllocationTracker struct {
    audioThreadAllocs uint64
    maxAllocsPerBuffer uint64
    allocBuffer       [1000]AllocInfo
    allocIndex        int
}

type AllocInfo struct {
    timestamp uint64
    size      uintptr
    stack     [8]uintptr
}

// Use with build tags for debug builds
//go:build debug
func (at *AllocationTracker) TrackAllocation(size uintptr) {
    // Capture stack trace for debugging
    var stack [8]uintptr
    runtime.Callers(2, stack[:])
    
    atomic.AddUint64(&at.audioThreadAllocs, 1)
    
    // Store allocation info (lock-free ring buffer)
    idx := atomic.AddInt32((*int32)(&at.allocIndex), 1) % 1000
    at.allocBuffer[idx] = AllocInfo{
        timestamp: uint64(time.Now().UnixNano()),
        size:      size,
        stack:     stack,
    }
}
```

### Real-Time Performance Metrics

```go
type PerformanceMetrics struct {
    processTime       time.Duration  // Last process call duration
    maxProcessTime    time.Duration  // Worst case
    avgProcessTime    time.Duration  // Moving average
    bufferUnderruns   uint64        // Audio dropouts
    gcPauses          uint64        // GC events during processing
    
    // Voice usage statistics
    maxVoicesUsed     int
    voiceStealEvents  uint64
    
    // Event processing stats
    eventsProcessed   uint64
    maxEventsPerBuffer uint64
}

func (p *Plugin) updateMetrics(startTime time.Time) {
    duration := time.Since(startTime)
    
    p.metrics.processTime = duration
    if duration > p.metrics.maxProcessTime {
        p.metrics.maxProcessTime = duration
    }
    
    // Update moving average (simple exponential smoothing)
    alpha := 0.1
    p.metrics.avgProcessTime = time.Duration(
        float64(p.metrics.avgProcessTime)*(1-alpha) + 
        float64(duration)*alpha,
    )
    
    // Check for buffer underruns (exceeded deadline)
    bufferDuration := time.Duration(p.frameCount) * time.Second / 
                     time.Duration(p.sampleRate)
    
    if duration > bufferDuration*80/100 { // 80% threshold
        atomic.AddUint64(&p.metrics.bufferUnderruns, 1)
    }
}
```

### Memory Profiling Integration

```go
//go:build debug
func (p *Plugin) StartProfiling() {
    // CPU profiling
    f, err := os.Create("cpu.prof")
    if err != nil {
        log.Fatal(err)
    }
    pprof.StartCPUProfile(f)
    
    // Memory profiling every 10 seconds
    go func() {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()
        
        for range ticker.C {
            f, err := os.Create(fmt.Sprintf("mem_%d.prof", time.Now().Unix()))
            if err != nil {
                continue
            }
            
            runtime.GC() // Force GC to get accurate stats
            pprof.WriteHeapProfile(f)
            f.Close()
        }
    }()
}
```

## Best Practices and Guidelines

### 1. Audio Thread Discipline

**DO:**
- Use atomic operations for simple data
- Pre-allocate all data structures
- Use fixed-size arrays instead of slices
- Minimize function calls in inner loops
- Cache frequently accessed data

**DON'T:**
- Allocate memory in audio thread
- Use channels or mutexes in audio processing
- Call fmt.Printf or other allocating functions
- Use interface{} types
- Access the filesystem or network

### 2. Voice Management Best Practices

```go
// GOOD: Efficient voice allocation
func (s *Synth) AllocateVoice(noteEvent NoteEvent) *Voice {
    // Try to reuse existing voice for same note
    for i := 0; i < s.voiceCount; i++ {
        voice := &s.voices[s.activeVoices[i]]
        if voice.noteID == noteEvent.NoteID {
            voice.Reset(noteEvent)
            return voice
        }
    }
    
    // Try to find free voice
    if s.voiceCount < MaxVoices {
        voice := &s.voices[s.voiceCount]
        s.activeVoices[s.voiceCount] = s.voiceCount
        s.voiceCount++
        voice.Init(noteEvent)
        return voice
    }
    
    // Voice stealing
    voice := s.StealVoice()
    if voice != nil {
        voice.Reset(noteEvent)
    }
    
    return voice
}
```

### 3. Parameter Handling

```go
// Efficient parameter updates with batching
func (p *Plugin) FlushParameters() {
    // Process all pending parameter updates
    for {
        select {
        case update := <-p.paramUpdates:
            p.applyParameterUpdate(update)
        default:
            return // No more updates
        }
    }
}

func (p *Plugin) applyParameterUpdate(update ParamUpdate) {
    // Use atomic store for audio thread access
    switch update.ID {
    case ParamVolume:
        atomic.StoreUint64((*uint64)(&p.volume), math.Float64bits(update.Value))
    case ParamAttack:
        atomic.StoreUint64((*uint64)(&p.attack), math.Float64bits(update.Value))
    // ... other parameters
    }
}
```

### 4. Error Handling in Real-Time Context

```go
// Non-allocating error handling
type ProcessResult int

const (
    ProcessOK ProcessResult = iota
    ProcessVoicePoolExhausted
    ProcessBufferOverrun
    ProcessInvalidParameters
)

func (s *Synth) ProcessNote(note NoteEvent) ProcessResult {
    if s.voiceCount >= MaxVoices {
        // Graceful degradation instead of allocation
        voice := s.StealVoice()
        if voice == nil {
            return ProcessVoicePoolExhausted
        }
        voice.Reset(note)
        return ProcessOK
    }
    
    // Normal allocation from pre-allocated pool
    voice := &s.voices[s.voiceCount]
    s.voiceCount++
    voice.Init(note)
    
    return ProcessOK
}
```

### 5. Testing for Real-Time Compliance

```go
func TestNoAllocationsInProcessing(t *testing.T) {
    plugin := NewTestPlugin()
    plugin.Init()
    plugin.Activate(44100, 64, 1024)
    
    // Prepare test data
    audioIn := make([][]float32, 2)
    audioOut := make([][]float32, 2)
    for i := range audioIn {
        audioIn[i] = make([]float32, 512)
        audioOut[i] = make([]float32, 512)
    }
    
    events := NewEventProcessor(nil, nil)
    
    // Force GC to start with clean state
    runtime.GC()
    runtime.GC()
    
    var memStatsBefore, memStatsAfter runtime.MemStats
    runtime.ReadMemStats(&memStatsBefore)
    
    // Process multiple buffers
    for i := 0; i < 100; i++ {
        plugin.Process(0, 512, audioIn, audioOut, events)
    }
    
    runtime.ReadMemStats(&memStatsAfter)
    
    // Check for allocations
    if memStatsAfter.Mallocs > memStatsBefore.Mallocs {
        t.Errorf("Memory allocations detected in audio processing: %d new allocations",
                 memStatsAfter.Mallocs-memStatsBefore.Mallocs)
    }
}
```

## Conclusion

Successfully using Go for real-time audio processing requires careful attention to:

1. **Memory allocation patterns** - Pre-allocate everything possible
2. **Garbage collection behavior** - Minimize and control GC timing
3. **Thread safety** - Use atomic operations and lock-free structures
4. **Performance monitoring** - Track allocations and timing
5. **Architectural discipline** - Separate audio and main thread concerns

The event pool system implemented in ClapGo demonstrates these principles in practice, achieving zero-allocation event processing. Similar techniques must be applied to all other audio processing components to achieve professional real-time performance.

With these considerations and the provided mitigation strategies, Go can be successfully used for professional audio plugin development while maintaining the deterministic performance requirements of real-time audio processing.