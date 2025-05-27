package thread

/*
#include <stdlib.h>
#include "../../include/clap/include/clap/ext/thread-pool.h"

static inline const void* clap_host_get_extension_helper(const clap_host_t* host, const char* id) {
    if (host && host->get_extension) {
        return host->get_extension(host, id);
    }
    return NULL;
}

static inline bool clap_host_thread_pool_request_exec(const clap_host_thread_pool_t* ext, const clap_host_t* host, uint32_t num_tasks) {
    if (ext && ext->request_exec) {
        return ext->request_exec(host, num_tasks);
    }
    return false;
}
*/
import "C"
import (
	"runtime"
	"sync"
	"unsafe"
)

// PoolProvider is an extension for plugins that can use the host's thread pool
// for parallel processing. The plugin must be able to fall back to single-threaded
// processing if the host doesn't provide a thread pool.
type PoolProvider interface {
	// Exec is called by the thread pool to execute a specific task.
	// The taskIndex parameter identifies which task to execute (0 to numTasks-1).
	// This is called from worker threads, so it must be thread-safe.
	// [audio-thread, worker-thread]
	Exec(taskIndex uint32)
}

// PoolHost provides access to the host's thread pool extension.
// This allows plugins to leverage the host's thread pool for parallel processing.
type PoolHost struct {
	host      unsafe.Pointer
	extension *C.clap_host_thread_pool_t
}

// NewPoolHost creates a new thread pool host interface.
// Returns nil if the host doesn't support the thread pool extension.
func NewPoolHost(host unsafe.Pointer) *PoolHost {
	if host == nil {
		return nil
	}

	cHost := (*C.clap_host_t)(host)
	if cHost.get_extension == nil {
		return nil
	}
	
	cExtID := C.CString("clap.thread-pool")
	defer C.free(unsafe.Pointer(cExtID))
	
	ext := C.clap_host_get_extension_helper(cHost, cExtID)
	if ext == nil {
		return nil
	}

	return &PoolHost{
		host:      host,
		extension: (*C.clap_host_thread_pool_t)(ext),
	}
}

// RequestExec schedules numTasks jobs in the host thread pool.
// It blocks until all tasks are processed.
// This must be used exclusively for realtime processing within the process call.
// Returns true if the host executed all tasks, false if it rejected the request.
// [audio-thread]
func (h *PoolHost) RequestExec(numTasks uint32) bool {
	if h.extension == nil || h.extension.request_exec == nil {
		return false
	}

	cHost := (*C.clap_host_t)(h.host)
	return bool(C.clap_host_thread_pool_request_exec(h.extension, cHost, C.uint32_t(numTasks)))
}

// PoolHelper provides convenient methods for parallel processing
// with automatic fallback to single-threaded execution.
type PoolHelper struct {
	host     *PoolHost
	provider PoolProvider
	
	// For fallback implementation
	fallbackPool *FallbackPool
}

// NewPoolHelper creates a new thread pool helper
func NewPoolHelper(host unsafe.Pointer, provider PoolProvider) *PoolHelper {
	return &PoolHelper{
		host:         NewPoolHost(host),
		provider:     provider,
		fallbackPool: NewFallbackPool(runtime.NumCPU()),
	}
}

// Execute runs numTasks in parallel using either the host's thread pool
// or a fallback implementation.
func (h *PoolHelper) Execute(numTasks uint32) {
	// Try to use host's thread pool first
	if h.host != nil && h.host.RequestExec(numTasks) {
		// Host executed the tasks
		return
	}
	
	// Fall back to our own implementation
	h.fallbackExecute(numTasks)
}

// fallbackExecute runs tasks using a simple parallel implementation
func (h *PoolHelper) fallbackExecute(numTasks uint32) {
	if numTasks == 0 {
		return
	}
	
	// For small task counts, just run serially
	if numTasks <= 2 {
		for i := uint32(0); i < numTasks; i++ {
			h.provider.Exec(i)
		}
		return
	}
	
	// Use goroutines for parallel execution
	h.fallbackPool.Execute(numTasks, h.provider.Exec)
}

// FallbackPool provides a simple thread pool implementation
// for when the host doesn't provide one.
type FallbackPool struct {
	maxWorkers int
}

// NewFallbackPool creates a new fallback thread pool
func NewFallbackPool(maxWorkers int) *FallbackPool {
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}
	return &FallbackPool{
		maxWorkers: maxWorkers,
	}
}

// Execute runs tasks in parallel using goroutines
func (p *FallbackPool) Execute(numTasks uint32, taskFunc func(uint32)) {
	if numTasks == 0 {
		return
	}
	
	// Determine number of workers
	workers := int(numTasks)
	if workers > p.maxWorkers {
		workers = p.maxWorkers
	}
	
	var wg sync.WaitGroup
	taskChan := make(chan uint32, numTasks)
	
	// Queue all tasks
	for i := uint32(0); i < numTasks; i++ {
		taskChan <- i
	}
	close(taskChan)
	
	// Start workers
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for taskIndex := range taskChan {
				taskFunc(taskIndex)
			}
		}()
	}
	
	// Wait for all tasks to complete
	wg.Wait()
}

// ParallelProcessor provides a high-level interface for parallel audio processing
type ParallelProcessor struct {
	helper *PoolHelper
}

// NewParallelProcessor creates a new parallel processor
func NewParallelProcessor(host unsafe.Pointer, provider PoolProvider) *ParallelProcessor {
	return &ParallelProcessor{
		helper: NewPoolHelper(host, provider),
	}
}

// ProcessChannels processes multiple audio channels in parallel
func (p *ParallelProcessor) ProcessChannels(numChannels uint32, processFunc func(channelIndex uint32)) {
	// Adapter to convert channel processing to task interface
	adapter := &channelProcessAdapter{processFunc: processFunc}
	tempHelper := &PoolHelper{
		host:         p.helper.host,
		provider:     adapter,
		fallbackPool: p.helper.fallbackPool,
	}
	tempHelper.Execute(numChannels)
}

// ProcessVoices processes multiple synthesizer voices in parallel
func (p *ParallelProcessor) ProcessVoices(numVoices uint32, processFunc func(voiceIndex uint32)) {
	// Adapter to convert voice processing to task interface
	adapter := &voiceProcessAdapter{processFunc: processFunc}
	tempHelper := &PoolHelper{
		host:         p.helper.host,
		provider:     adapter,
		fallbackPool: p.helper.fallbackPool,
	}
	tempHelper.Execute(numVoices)
}

// Adapter types for specific processing patterns
type channelProcessAdapter struct {
	processFunc func(uint32)
}

func (a *channelProcessAdapter) Exec(taskIndex uint32) {
	a.processFunc(taskIndex)
}

type voiceProcessAdapter struct {
	processFunc func(uint32)
}

func (a *voiceProcessAdapter) Exec(taskIndex uint32) {
	a.processFunc(taskIndex)
}