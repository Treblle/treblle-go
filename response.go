package treblle

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
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
	responseBytes := response.Body.Bytes()

	errInfo := ErrorInfo{}
	var body json.RawMessage
	err := json.Unmarshal(responseBytes, &body)
	if err != nil {
		errInfo.Message = err.Error()
	}

	headers := make(map[string]string)
	for k := range response.Header() {
		headers[k] = response.Header().Get(k)
	}

	re := ResponseInfo{
		Code:     response.Code,
		Size:     len(responseBytes),
		LoadTime: float64(time.Since(startTime).Microseconds()),
		Body:     body,
		Errors:   []ErrorInfo{errInfo},
	}
	bodyJson, _ := json.Marshal(body)
	sanitizedBody, _ := getMaskedJSON(bodyJson)
	re.Body = sanitizedBody

	headersJson, _ := json.Marshal(headers)
	sanitizedHeaders, _ := getMaskedJSON(headersJson)
	re.Headers = sanitizedHeaders
	var jsonData interface{}

	err = json.Unmarshal(sanitizedHeaders, &jsonData)
	if err != nil {
		fmt.Println("Error parsing raw message:", err)
	}
	return re
}
