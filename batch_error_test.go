package treblle

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBatchErrorCollector(t *testing.T) {
	// Configure Treblle with test settings
	Configure(Configuration{
		SDK_TOKEN: "test-sdk-token",
		API_KEY:   "test-api-key",
		Endpoint:  "http://localhost:8080",
	})

	// Create a batch collector with small batch size and interval for testing
	collector := NewBatchErrorCollector(2, 100*time.Millisecond)
	defer collector.Close()

	// Create test errors
	testErrors := []ErrorInfo{
		{
			Message: "Test error 1",
			Type:    ValidationError,
			Source:  "test",
			Line:    42,
			File:    "test.go",
		},
		{
			Message: "Test error 2",
			Type:    ServerError,
			Source:  "test",
			Line:    43,
			File:    "test.go",
		},
	}

	// Test batch size trigger
	t.Run("BatchSizeTrigger", func(t *testing.T) {
		// Add errors up to batch size
		for _, err := range testErrors {
			collector.Add(err)
		}

		// Allow time for batch processing
		time.Sleep(50 * time.Millisecond)

		// Verify the batch was sent (errors cleared)
		collector.mu.Lock()
		assert.Equal(t, 0, len(collector.errors), "Batch should be cleared after reaching batch size")
		collector.mu.Unlock()
	})

	// Test interval trigger
	t.Run("IntervalTrigger", func(t *testing.T) {
		// Add one error
		collector.Add(testErrors[0])

		// Wait for flush interval
		time.Sleep(150 * time.Millisecond)

		// Verify the batch was sent
		collector.mu.Lock()
		assert.Equal(t, 0, len(collector.errors), "Batch should be cleared after interval")
		collector.mu.Unlock()
	})

	// Test close functionality
	t.Run("CloseFlush", func(t *testing.T) {
		// Add an error
		collector.Add(testErrors[0])

		// Close the collector
		collector.Close()

		// Verify all errors were flushed
		collector.mu.Lock()
		assert.Equal(t, 0, len(collector.errors), "All errors should be flushed on close")
		collector.mu.Unlock()
	})
}
