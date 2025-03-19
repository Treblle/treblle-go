package treblle

import (
	"context"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

// contextKey type for request context values
type contextKey string

const (
	// treblleRequestStartedAtKey is the key for storing request start time
	treblleRequestStartedAtKey contextKey = "treblle_request_started_at"
	// treblleRequestInfoKey is the key for storing request information
	treblleRequestInfoKey contextKey = "treblle_request_info"
)

// AsyncProcessor manages asynchronous processing with controlled concurrency
type AsyncProcessor struct {
	sem    *semaphore.Weighted
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

// RequestTracker stores and retrieves request data using context
type RequestTracker struct{}

var (
	// Global async processor instance
	asyncProcessor     *AsyncProcessor
	asyncProcessorOnce sync.Once

	// Global request tracker instance
	requestTracker     *RequestTracker
	requestTrackerOnce sync.Once
)

// NewAsyncProcessor creates a new async processor with controlled concurrency
func NewAsyncProcessor(maxConcurrent int64) *AsyncProcessor {
	ctx, cancel := context.WithCancel(context.Background())
	return &AsyncProcessor{
		sem:    semaphore.NewWeighted(maxConcurrent),
		ctx:    ctx,
		cancel: cancel,
	}
}

// GetAsyncProcessor returns the singleton async processor
func GetAsyncProcessor() *AsyncProcessor {
	asyncProcessorOnce.Do(func() {
		maxConcurrent := 10 // Default value
		if Config.MaxConcurrentProcessing > 0 {
			maxConcurrent = Config.MaxConcurrentProcessing
		}
		asyncProcessor = NewAsyncProcessor(int64(maxConcurrent))
	})
	return asyncProcessor
}

// GetRequestTracker returns the singleton request tracker
func GetRequestTracker() *RequestTracker {
	requestTrackerOnce.Do(func() {
		requestTracker = &RequestTracker{}
	})
	return requestTracker
}

// Process handles the asynchronous processing of Treblle data
func (ap *AsyncProcessor) Process(requestInfo RequestInfo, responseInfo ResponseInfo, errorProvider *ErrorProvider) {
	ap.wg.Add(1)

	// Process asynchronously
	go func() {
		defer ap.wg.Done()

		// Create a context with timeout for acquiring the semaphore
		acquireCtx, cancel := context.WithTimeout(ap.ctx, 100*time.Millisecond)
		defer cancel()

		// Try to acquire the semaphore
		if err := ap.sem.Acquire(acquireCtx, 1); err != nil {
			// If we can't acquire the semaphore in time, just drop the request
			// This prevents backpressure during high load
			return
		}
		defer ap.sem.Release(1)

		// Create metadata
		ti := MetaData{
			ApiKey:    Config.APIKey,
			ProjectID: Config.ProjectID,
			Version:   Config.SDKVersion,
			Sdk:       Config.SDKName,
			//	Url:       requestInfo.RoutePath, // Use the normalized URL from requestInfo (critical for endpoint grouping)
			Data: DataInfo{
				Server:   Config.serverInfo,
				Language: Config.languageInfo,
				Request:  requestInfo,
				Response: responseInfo,
			},
		}

		// Use a context with timeout for the API call
		sendCtx, sendCancel := context.WithTimeout(ap.ctx, 2*time.Second)
		defer sendCancel()

		// Send to Treblle with context
		sendToTreblleWithContext(sendCtx, ti)
	}()
}

// Wait waits for all processing to complete with a timeout
func (ap *AsyncProcessor) Wait(timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		ap.wg.Wait()
	}()

	select {
	case <-c:
		return true // All processing completed
	case <-time.After(timeout):
		return false // Timed out
	}
}

// Shutdown gracefully shuts down the processor
func (ap *AsyncProcessor) Shutdown(timeout time.Duration) {
	// Signal cancellation to all operations
	ap.cancel()

	// Wait for ongoing operations to complete
	ap.Wait(timeout)
}

// StoreStartTime stores the request start time in context
func (rt *RequestTracker) StoreStartTime(r *http.Request) *http.Request {
	ctx := context.WithValue(r.Context(), treblleRequestStartedAtKey, time.Now())
	return r.WithContext(ctx)
}

// GetStartTime retrieves the start time from context
func (rt *RequestTracker) GetStartTime(r *http.Request) (time.Time, bool) {
	if val := r.Context().Value(treblleRequestStartedAtKey); val != nil {
		if t, ok := val.(time.Time); ok {
			return t, true
		}
	}
	return time.Time{}, false
}

// StoreRequestInfo stores request info in context
func (rt *RequestTracker) StoreRequestInfo(r *http.Request, info RequestInfo) *http.Request {
	ctx := context.WithValue(r.Context(), treblleRequestInfoKey, info)
	return r.WithContext(ctx)
}

// GetRequestInfo retrieves request info from context
func (rt *RequestTracker) GetRequestInfo(r *http.Request) (RequestInfo, bool) {
	if val := r.Context().Value(treblleRequestInfoKey); val != nil {
		if info, ok := val.(RequestInfo); ok {
			return info, true
		}
	}
	return RequestInfo{}, false
}
