package treblle

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// ShutdownOptions contains options for graceful shutdown
type ShutdownOptions struct {
	// Additional fields to be masked during shutdown
	AdditionalFieldsToMask []string
	// Custom error provider (optional)
	ErrorProvider *ErrorProvider
}

// Shutdown sends data to Treblle before application shutdown
// It's similar to the terminate method in the Laravel SDK
func Shutdown(r *http.Request, w http.ResponseWriter, responseBody []byte, options *ShutdownOptions) {
	// Create error provider if not provided
	errorProvider := NewErrorProvider()
	if options != nil && options.ErrorProvider != nil {
		errorProvider = options.ErrorProvider
	}
	
	var requestInfo RequestInfo
	var startTime time.Time
	
	// Try to get request info from context if async processing is enabled
	if Config.AsyncProcessingEnabled {
		tracker := GetRequestTracker()
		
		if storedRequestInfo, ok := tracker.GetRequestInfo(r); ok {
			requestInfo = storedRequestInfo
		}
		
		if storedStartTime, ok := tracker.GetStartTime(r); ok {
			startTime = storedStartTime
		} else {
			// Fallback to current time
			startTime = time.Now().Add(-time.Millisecond)
		}
	} else {
		// Get start time (using current time as we don't have the actual start time)
		startTime = time.Now().Add(-time.Millisecond) // Subtract a millisecond to ensure duration is positive
		
		// Get request info
		var errReqInfo error
		requestInfo, errReqInfo = getRequestInfo(r, startTime, errorProvider)
		if errReqInfo != nil && !errors.Is(errReqInfo, ErrNotJson) {
			errorProvider.AddError(errReqInfo, ValidationError, "shutdown_request_processing")
		}
	}
	
	// Process headers for response info
	headers := make(map[string]interface{})
	for key, values := range w.Header() {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	
	headersJson, err := json.Marshal(maskMap(headers))
	if err != nil {
		errorProvider.AddError(err, MarshalError, "header_encoding")
	}
	
	// Create response info
	responseInfo := ResponseInfo{
		Headers:  json.RawMessage(headersJson),
		Code:     http.StatusOK, // Default to 200 if not available
		Size:     len(responseBody),
		LoadTime: float64(time.Since(startTime).Milliseconds()),
		Errors:   []ErrorInfo{},
	}
	
	// If response writer is an http.ResponseWriter that allows status code retrieval
	if rw, ok := w.(interface{ Status() int }); ok {
		responseInfo.Code = rw.Status()
	}
	
	// Process response body if available
	if len(responseBody) > 0 {
		// Try to mask if it's JSON
		sanitizedBody, err := getMaskedJSON(responseBody)
		if err == nil {
			responseInfo.Body = sanitizedBody
		} else {
			// For non-JSON responses, just store the raw body as a JSON string
			jsonBytes, err := json.Marshal(string(responseBody))
			if err == nil {
				responseInfo.Body = json.RawMessage(jsonBytes)
			}
		}
	}
	
	// Add all collected errors to the response
	responseInfo.Errors = errorProvider.GetErrors()
	
	// Create metadata
	ti := MetaData{
		ApiKey:    Config.APIKey,
		ProjectID: Config.ProjectID,
		Version:   SDKVersion,
		Sdk:       SDKName,
		Data: DataInfo{
			Server:   Config.serverInfo,
			Language: Config.languageInfo,
			Request:  requestInfo,
			Response: responseInfo,
		},
	}
	
	// Flush any batch errors if batch error collector is enabled
	if Config.batchErrorCollector != nil {
		Config.batchErrorCollector.Close()
	}
	
	// Send data to Treblle synchronously (not in a goroutine since we're shutting down)
	sendToTreblle(ti)
}

// ShutdownWithCustomData sends custom request and response data to Treblle before shutdown
func ShutdownWithCustomData(requestInfo RequestInfo, responseInfo ResponseInfo, errorProvider *ErrorProvider) {
	// Add collected errors to the response if error provider is available
	if errorProvider != nil {
		responseInfo.Errors = errorProvider.GetErrors()
	}
	
	// Create metadata
	ti := MetaData{
		ApiKey:    Config.APIKey,
		ProjectID: Config.ProjectID,
		Version:   SDKVersion,
		Sdk:       SDKName,
		Data: DataInfo{
			Server:   Config.serverInfo,
			Language: Config.languageInfo,
			Request:  requestInfo,
			Response: responseInfo,
		},
	}
	
	// Flush any batch errors if batch error collector is enabled
	if Config.batchErrorCollector != nil {
		Config.batchErrorCollector.Close()
	}
	
	// Send data to Treblle synchronously
	sendToTreblle(ti)
}

// GracefulShutdown flushes any pending batch errors and ensures all data is sent to Treblle
// This can be called during application shutdown to ensure all data is properly sent
func GracefulShutdown() {
	// Wait for async processor to finish if enabled
	if Config.AsyncProcessingEnabled {
		timeout := 5 * time.Second
		if Config.AsyncShutdownTimeout > 0 {
			timeout = Config.AsyncShutdownTimeout
		}
		GetAsyncProcessor().Shutdown(timeout)
	}
	
	// Flush batch errors if enabled
	if Config.batchErrorCollector != nil {
		Config.batchErrorCollector.Flush()
	}
}
