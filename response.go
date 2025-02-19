package treblle

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"time"
)

type ResponseInfo struct {
	Headers  json.RawMessage `json:"headers"`
	Code     int             `json:"code"`
	Size     int             `json:"size"`
	LoadTime float64         `json:"load_time"`
	Body     json.RawMessage `json:"body"`
	Errors   []ErrorInfo     `json:"errors"`
}

type ErrorInfo struct {
	Source  string `json:"source"`
	Type    string `json:"type"`
	Message string `json:"message"`
	File    string `json:"file"`
	Line    int    `json:"line"`
}

// Extract information from the response recorder
func getResponseInfo(response *httptest.ResponseRecorder, startTime time.Time) ResponseInfo {
	defer dontPanic()

	re := ResponseInfo{
		Code:     response.Code,
		Size:     len(response.Body.Bytes()),
		LoadTime: float64(time.Since(startTime).Microseconds()),
		Errors:   []ErrorInfo{},
	}

	// Handle response body
	if bodyBytes := response.Body.Bytes(); len(bodyBytes) > 0 {
		// Try to mask if it's JSON
		sanitizedBody, err := getMaskedJSON(bodyBytes)
		if err != nil {
			// For non-JSON responses, just store the raw body as a JSON string
			if errors.Is(err, ErrNotJson) {
				// Create a JSON-encoded string without extra quotes
				jsonBytes, err := json.Marshal(string(bodyBytes))
				if err != nil {
					re.Errors = append(re.Errors, ErrorInfo{
						Source:  "response",
						Type:    "body_encoding_error",
						Message: err.Error(),
					})
				}
				re.Body = json.RawMessage(jsonBytes)
			} else {
				re.Errors = append(re.Errors, ErrorInfo{
					Source:  "response",
					Type:    "body_masking_error",
					Message: err.Error(),
				})
				jsonBytes, _ := json.Marshal(string(bodyBytes))
				re.Body = json.RawMessage(jsonBytes)
			}
		} else {
			re.Body = sanitizedBody
		}
	}

	// Handle response headers
	headers := make(map[string]interface{})
	for k, v := range response.Header() {
		if len(v) == 1 {
			if shouldMaskField(k) {
				if strings.ToLower(k) == "authorization" {
					parts := strings.SplitN(v[0], " ", 2)
					if len(parts) == 2 {
						headers[k] = parts[0] + " " + strings.Repeat("*", 9)
					} else {
						headers[k] = strings.Repeat("*", 9)
					}
				} else {
					headers[k] = strings.Repeat("*", 9)
				}
			} else {
				headers[k] = v[0]
			}
		} else {
			if shouldMaskField(k) {
				masked := make([]string, len(v))
				for i := range v {
					masked[i] = strings.Repeat("*", 9)
				}
				headers[k] = masked
			} else {
				headers[k] = v
			}
		}
	}

	headersJson, _ := json.Marshal(headers)
	re.Headers = json.RawMessage(headersJson)

	return re
}

// Helper function to check if a header should be masked
func shouldMaskHeader(headerName string) bool {
	// Convert common header variations to lowercase for consistent matching
	headerVariations := []string{
		headerName,
		"x-" + headerName,
		"x_" + headerName,
	}

	for _, h := range headerVariations {
		if _, exists := Config.FieldsMap[h]; exists {
			return true
		}
	}
	return false
}
