package treblle

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"time"
)

// Define the maximum response size (2MB in bytes)
const maxResponseSize = 2 * 1024 * 1024

type ResponseInfo struct {
	Headers  json.RawMessage `json:"headers"`
	Code     int             `json:"code"`
	Size     int             `json:"size"`
	LoadTime float64         `json:"load_time"`
	Body     json.RawMessage `json:"body"`
	Errors   []ErrorInfo     `json:"errors"`
}

// getResponseInfo extracts information from the response matching Laravel SDK structure
func getResponseInfo(response *httptest.ResponseRecorder, startTime time.Time, errorProvider *ErrorProvider) ResponseInfo {
	// Process headers (similar to Laravel's collect()->first())
	headers := make(map[string]interface{})
	for key, values := range response.Header() {
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
			"getResponseInfo",
		)
	}

	// Get response body
	body := response.Body.Bytes()
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
			contentType := response.Header().Get("Content-Type")
			if contentType == "application/json" {
				maskedBody, err := getMaskedJSON(body)
				if err != nil {
					bodyJSON = json.RawMessage("{}")
					errorProvider.AddCustomError(
						fmt.Sprintf("failed to mask response body: %v", err),
						MarshalError,
						"getResponseInfo",
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
						"getResponseInfo",
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

	// Calculate load time in milliseconds (matching Laravel's precision)
	loadTime := float64(time.Since(startTime).Microseconds()) / 1000.0

	return ResponseInfo{
		Headers:  headerJSON,
		Code:     response.Code,
		Size:     size,
		LoadTime: loadTime,
		Body:     bodyJSON,
		Errors:   errorProvider.GetErrors(),
	}
}
