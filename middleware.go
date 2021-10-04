package treblle

import (
	"net/http"
	"net/http/httptest"
	"time"
)

const (
	treblleVersion = 0.6
	sdkName        = "go"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		requestInfo, err := getRequestInfo(r, startTime)
		if err != nil {
			return // continue but don't send anything out to Treblle
		}

		// intercept the response so it can be copied
		rec := httptest.NewRecorder()

		// do the actual request as intended
		next.ServeHTTP(rec, r)
		// after this finishes, we have the response recorded

		ti := MetaData{
			ApiKey:    Config.APIKey,
			ProjectID: Config.ProjectID,
			Version:   treblleVersion,
			Sdk:       sdkName,
			Data: DataInfo{
				Server:   getServerInfo(),
				Language: getLanguageInfo(),
				Request:  requestInfo,
				Response: getResponseInfo(rec, startTime),
			},
		}
		go sendToTreblle(ti)

		w.Write(rec.Body.Bytes())
	})
}
