package treblle

import (
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"time"
)

const (
	treblleVersion = "0.7.2"
	sdkName        = "go"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Get request info before processing
		requestInfo, errReqInfo := getRequestInfo(r, startTime)

		// intercept the response so it can be copied
		rec := httptest.NewRecorder()
		next.ServeHTTP(rec, r)

		// copy everything from response recorder to response writer
		for k, v := range rec.Header() {
			w.Header()[k] = v
		}
		w.WriteHeader(rec.Code)
		_, err := w.Write(rec.Body.Bytes())
		if err != nil {
			log.Printf("Error writing response: %v", err)
			return
		}

		// Send to Treblle if:
		// 1. The request was valid JSON (or had no body)
		// OR
		// 2. The response is JSON (regardless of status code)
		if !errors.Is(errReqInfo, ErrNotJson) || rec.Header().Get("Content-Type") == "application/json" {
			responseInfo := getResponseInfo(rec, startTime)

			// If there was an error with the request, add it to the response errors
			if errReqInfo != nil && !errors.Is(errReqInfo, ErrNotJson) {
				responseInfo.Errors = append(responseInfo.Errors, ErrorInfo{
					Source:  "request",
					Type:    "REQUEST_ERROR",
					Message: errReqInfo.Error(),
				})
			}

			ti := MetaData{
				ApiKey:    Config.APIKey,
				ProjectID: Config.ProjectID,
				Version:   treblleVersion,
				Sdk:       sdkName,
				Data: DataInfo{
					Server:   Config.serverInfo,
					Language: Config.languageInfo,
					Request:  requestInfo,
					Response: responseInfo,
				},
			}
			// don't block execution while sending data to Treblle
			go sendToTreblle(ti)
		}
	})
}

// If anything happens to go wrong inside one of treblle-go internals, recover from panic and continue
func dontPanic() {
	if err := recover(); err != nil {
		log.Printf("treblle-go panic: %s", err)
	}
}
