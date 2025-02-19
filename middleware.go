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

		// do the actual request as intended
		next.ServeHTTP(rec, r)
		// after this finishes, we have the response recorded

		// copy the original headers
		for k, v := range rec.Header() {
			w.Header()[k] = v
		}

		// copy the original code
		w.WriteHeader(rec.Code)

		// write the original body
		_, err := w.Write(rec.Body.Bytes())
		if err != nil {
			return
		}

		// Only send to Treblle if:
		// 1. The request was valid JSON (or had no body)
		// 2. The response is successful (2xx status)
		if !errors.Is(errReqInfo, ErrNotJson) && rec.Code >= 200 && rec.Code < 300 {
			ti := MetaData{
				ApiKey:    Config.APIKey,
				ProjectID: Config.ProjectID,
				Version:   treblleVersion,
				Sdk:       sdkName,
				Data: DataInfo{
					Server:   Config.serverInfo,
					Language: Config.languageInfo,
					Request:  requestInfo,
					Response: getResponseInfo(rec, startTime),
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
