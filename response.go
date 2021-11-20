package treblle

import (
	"encoding/json"
	"net/http/httptest"
	"time"
)

type ResponseInfo struct {
	Headers  map[string]string      `json:"headers"`
	Code     int                    `json:"code"`
	Size     int                    `json:"size"`
	LoadTime float64                `json:"load_time"`
	Body     map[string]interface{} `json:"body"`
	Errors   []ErrorInfo            `json:"errors"`
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
	responseBytes := response.Body.Bytes()

	errInfo := ErrorInfo{}
	var body map[string]interface{}
	err := json.Unmarshal(responseBytes, &body)
	if err != nil {
		errInfo.Message = err.Error()
	}

	headers := make(map[string]string)
	for k, _ := range response.Header() {
		headers[k] = response.Header().Get(k)
	}

	return ResponseInfo{
		Headers:  headers,
		Code:     response.Code,
		Size:     len(responseBytes),
		LoadTime: float64(time.Since(startTime).Microseconds()),
		Body:     body,
		Errors:   []ErrorInfo{errInfo},
	}
}
