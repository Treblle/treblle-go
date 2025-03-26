package treblle

import (
	"sync"
	"time"
)

// BatchErrorCollector handles batch collection and transmission of errors
type BatchErrorCollector struct {
	mu            sync.Mutex
	errors        []ErrorInfo
	batchSize     int
	flushInterval time.Duration
	done          chan struct{}
	wg            sync.WaitGroup
}

// NewBatchErrorCollector creates a new BatchErrorCollector with specified batch size and flush interval
func NewBatchErrorCollector(batchSize int, flushInterval time.Duration) *BatchErrorCollector {
	if batchSize <= 0 {
		batchSize = 100 // default batch size
	}
	if flushInterval <= 0 {
		flushInterval = 5 * time.Second // default flush interval
	}

	collector := &BatchErrorCollector{
		errors:        make([]ErrorInfo, 0, batchSize),
		batchSize:     batchSize,
		flushInterval: flushInterval,
		done:          make(chan struct{}),
	}

	go collector.periodicFlush()
	return collector
}

// Add adds an error to the batch
func (b *BatchErrorCollector) Add(err ErrorInfo) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.errors = append(b.errors, err)
	if len(b.errors) >= b.batchSize {
		b.flush()
	}
}

// flush sends the current batch of errors to Treblle
func (b *BatchErrorCollector) flush() {
	if len(b.errors) == 0 {
		return
	}

	// Create a copy of errors to send
	errorsCopy := make([]ErrorInfo, len(b.errors))
	copy(errorsCopy, b.errors)

	// Clear the current batch
	b.errors = b.errors[:0]

	// Send errors asynchronously
	b.wg.Add(1)
	go func(errors []ErrorInfo) {
		defer b.wg.Done()
		// Create metadata for batch transmission
		meta := MetaData{
			ApiKey:    Config.APIKey,
			ProjectID: Config.ProjectID,
			Version:   Config.SDKVersion,
			Sdk:       Config.SDKName,
			Data: DataInfo{
				Server:   Config.serverInfo,
				Language: Config.languageInfo,
				Request:  RequestInfo{},  // Empty request info for batch errors
				Response: ResponseInfo{}, // Empty response info for batch errors
				Errors:   errors,
			},
		}

		// Send to Treblle
		sendToTreblle(meta)
	}(errorsCopy)
}

// periodicFlush periodically flushes the error batch based on the flush interval
func (b *BatchErrorCollector) periodicFlush() {
	ticker := time.NewTicker(b.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.mu.Lock()
			b.flush()
			b.mu.Unlock()
		case <-b.done:
			return
		}
	}
}

// Close stops the periodic flushing and flushes any remaining errors
func (b *BatchErrorCollector) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	select {
	case <-b.done:
		// Channel already closed
		return
	default:
		close(b.done)
		b.flush()
		b.wg.Wait()
	}
}

// Flush sends any pending errors to Treblle immediately
func (b *BatchErrorCollector) Flush() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.flush()
}
