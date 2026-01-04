package treblle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// ginResponseCapture wraps Gin's ResponseWriter to capture response body
type ginResponseCapture struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write captures the response body while passing it through to Gin
func (w *ginResponseCapture) Write(b []byte) (int, error) {
	w.body.Write(b) // Capture for Treblle
	return w.ResponseWriter.Write(b)
}

// WriteString captures the response body while passing it through to Gin
func (w *ginResponseCapture) WriteString(s string) (int, error) {
	w.body.WriteString(s) // Capture for Treblle
	return w.ResponseWriter.WriteString(s)
}

// GinMiddleware returns a Gin-compatible middleware for Treblle
func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if the current environment is in the ignored list
		if IsEnvironmentIgnored() {
			// Skip Treblle logging for ignored environments
			c.Next()
			return
		}

		// Create error provider for this request
		errorProvider := NewErrorProvider()
		defer errorProvider.Clear()

		// Recover from panics
		defer func() {
			if err := recover(); err != nil {
				errorProvider.AddCustomError(
					fmt.Sprintf("panic recovered: %v", err),
					UnhandledExceptionError,
					"gin_middleware",
				)
			}
		}()

		// Get the request tracker
		tracker := GetRequestTracker()

		// Store start time in request context
		startTime := time.Now()
		c.Request = tracker.StoreStartTime(c.Request)

		// Extract route pattern from Gin (e.g., "/users/:id")
		routePath := c.FullPath()
		if routePath != "" {
			c.Request = SetRoutePath(c.Request, routePath)
		}

		// Get request info before processing
		requestInfo, errReqInfo := getRequestInfo(c.Request, startTime, errorProvider)
		if errReqInfo != nil && errReqInfo != ErrNotJson {
			errorProvider.AddError(errReqInfo, ValidationError, "request_processing")
		}

		// Log the route path for debugging
		if Config.Debug {
			fmt.Printf("==== DEBUG: TREBLLE GIN ROUTE PATH ====\n")
			fmt.Printf("Original URL Path: %s\n", c.Request.URL)
			fmt.Printf("Gin Full Path: %s\n", routePath)
			fmt.Printf("Normalized Route Path: %s\n", requestInfo.RoutePath)
			fmt.Printf("Full URL: %s\n", requestInfo.Url)
			fmt.Printf("====================================\n")
		}

		// Store request info in context if async processing is enabled
		if Config.AsyncProcessingEnabled {
			c.Request = tracker.StoreRequestInfo(c.Request, requestInfo)
		}

		// Wrap Gin's ResponseWriter to capture response body
		capture := &ginResponseCapture{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = capture

		// Execute the handler chain
		c.Next()

		// Extract response info from Gin's writer
		responseInfo := getGinResponseInfo(capture, startTime, errorProvider)

		// Add all collected errors to the response
		responseInfo.Errors = errorProvider.GetErrors()

		// Create a copy of the serverInfo with the correct protocol for this request
		serverInfo := Config.serverInfo
		serverInfo.Protocol = DetectProtocol(c.Request)

		if Config.AsyncProcessingEnabled {
			// Process asynchronously with controlled concurrency
			GetAsyncProcessor().Process(requestInfo, responseInfo, errorProvider)
		} else {
			// Create metadata
			ti := MetaData{
				ApiKey:    Config.APIKey,
				ProjectID: Config.ProjectID,
				Version:   Config.SDKVersion,
				Sdk:       Config.SDKName,
				Data: DataInfo{
					Server:   serverInfo,
					Language: Config.languageInfo,
					Request:  requestInfo,
					Response: responseInfo,
				},
			}

			// Don't block execution while sending data to Treblle
			go func(ti MetaData) {
				defer func() {
					if err := recover(); err != nil {
						fmt.Printf("Panic recovered in goroutine: %v\n", err)
						// Silently recover from panic
					}
				}()
				sendToTreblle(ti)
			}(ti)
		}
	}
}

// getGinResponseInfo extracts response information from Gin's ResponseWriter
func getGinResponseInfo(capture *ginResponseCapture, startTime time.Time, errorProvider *ErrorProvider) ResponseInfo {
	// Process headers (similar to getResponseInfo)
	headers := make(map[string]interface{})
	for key, values := range capture.Header() {
		if len(values) == 0 {
			continue
		}

		// For multiple values, keep them as an array
		if len(values) > 1 {
			// If the field should be masked, mask each value
			if shouldMaskField(key) {
				maskedValues := make([]interface{}, len(values))
				for i := range values {
					maskedValues[i] = maskValue(values[i], key)
				}
				headers[key] = maskedValues
			} else {
				headers[key] = values
			}
		} else {
			// Single value
			if shouldMaskField(key) {
				headers[key] = maskValue(values[0], key)
			} else {
				headers[key] = values[0]
			}
		}
	}

	headerJSON, err := json.Marshal(headers)
	if err != nil {
		headerJSON = json.RawMessage("{}")
		errorProvider.AddCustomError(
			fmt.Sprintf("failed to marshal response headers: %v", err),
			MarshalError,
			"getGinResponseInfo",
		)
	}

	// Get response body from captured buffer
	body := capture.body.Bytes()
	var bodyJSON json.RawMessage
	var size int

	if len(body) > 0 {
		if len(body) > maxResponseSize {
			// Replace with empty JSON object
			bodyJSON = json.RawMessage("{}")
			// Set size to 0 as we're not sending the actual body
			size = 0
			errorProvider.AddCustomError(
				"JSON response size is over 2MB",
				ServerError,
				"response_size_limit",
			)
		} else {
			// Check if response is JSON
			contentType := capture.Header().Get("Content-Type")
			if contentType == "application/json" || contentType == "application/json; charset=utf-8" {
				maskedBody, err := getMaskedJSON(body)
				if err != nil {
					bodyJSON = json.RawMessage("{}")
					errorProvider.AddCustomError(
						fmt.Sprintf("failed to mask response body: %v", err),
						MarshalError,
						"getGinResponseInfo",
					)
				} else {
					bodyJSON = maskedBody
				}
			} else {
				// For non-JSON responses, wrap the raw string in JSON quotes
				bodyStr := string(body)
				bodyBytes, err := json.Marshal(bodyStr)
				if err != nil {
					bodyJSON = json.RawMessage("{}")
					errorProvider.AddCustomError(
						fmt.Sprintf("failed to marshal non-JSON response: %v", err),
						MarshalError,
						"getGinResponseInfo",
					)
				} else {
					bodyJSON = bodyBytes
				}
			}
			size = len(body)
		}
	} else {
		bodyJSON = json.RawMessage("{}")
		size = 0
	}

	// Calculate load time in milliseconds
	loadTime := float64(time.Since(startTime).Microseconds()) / 1000.0

	return ResponseInfo{
		Headers:  headerJSON,
		Code:     capture.Status(),
		Size:     size,
		LoadTime: loadTime,
		Body:     bodyJSON,
		Errors:   errorProvider.GetErrors(),
	}
}
