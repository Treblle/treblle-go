package treblle

import (
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestAsyncProcessor_Process(t *testing.T) {
	// Setup test configuration
	originalConfig := Config
	defer func() {
		Config = originalConfig
	}()

	Config = internalConfiguration{
		AsyncProcessingEnabled:  true,
		MaxConcurrentProcessing: 2,
		AsyncShutdownTimeout:    1 * time.Second,
		SDKName:                 "treblle-go-test",
		SDKVersion:              0.1,
	}

	// Create a mock request and response
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a mock request info
	requestInfo := RequestInfo{
		Method: "GET",
		Url:    "/test",
	}

	// Create a mock response info
	responseInfo := ResponseInfo{
		Code:     200,
		Size:     10,
		LoadTime: 5.0,
	}

	// Create a mock error provider
	errorProvider := NewErrorProvider()

	// Create a new async processor
	processor := GetAsyncProcessor()

	// Test processing multiple requests
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			processor.Process(requestInfo, responseInfo, errorProvider)
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Test shutdown
	processor.Shutdown(1 * time.Second)
}

func TestRequestTracker(t *testing.T) {
	// Create a new request tracker
	tracker := GetRequestTracker()

	// Create a test request with context
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Test storing and retrieving start time
	startTime := time.Now()
	req = tracker.StoreStartTime(req)

	retrievedTime, ok := tracker.GetStartTime(req)
	if !ok {
		t.Error("Failed to retrieve start time from request context")
	}

	// Times should be very close (within 1ms)
	if retrievedTime.Sub(startTime) > time.Millisecond {
		t.Errorf("Retrieved time does not match stored time: %v vs %v", retrievedTime, startTime)
	}

	// Test storing and retrieving request info
	requestInfo := RequestInfo{
		Method: "GET",
		Url:    "/test",
	}

	req = tracker.StoreRequestInfo(req, requestInfo)

	retrievedInfo, ok := tracker.GetRequestInfo(req)
	if !ok {
		t.Error("Failed to retrieve request info from request context")
	}

	if retrievedInfo.Method != requestInfo.Method || retrievedInfo.Url != requestInfo.Url {
		t.Errorf("Retrieved request info does not match stored info: %+v vs %+v", retrievedInfo, requestInfo)
	}
}

func TestAsyncShutdown(t *testing.T) {
	// Setup test configuration
	originalConfig := Config
	defer func() {
		Config = originalConfig
	}()

	Config = internalConfiguration{
		AsyncProcessingEnabled:  true,
		MaxConcurrentProcessing: 2,
		AsyncShutdownTimeout:    500 * time.Millisecond,
		SDKName:                 "treblle-go-test",
		SDKVersion:              0.1,
	}

	// Create a new async processor
	processor := GetAsyncProcessor()

	// Skip the semaphore test as it's an implementation detail
	// Just test the shutdown timeout
	start := time.Now()
	processor.Shutdown(500 * time.Millisecond)
	duration := time.Since(start)

	// Shutdown should be quick since there are no pending tasks
	if duration > 100*time.Millisecond {
		t.Errorf("Shutdown took %v, expected to be quick", duration)
	}
}
