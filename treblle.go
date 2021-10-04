package treblle

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

const (
	treblleURL      = "https://rocknrolla.treblle.com" // TODO: don't hardcode
	timeoutDuration = 2 * time.Second
)

var ErrNotJson = errors.New("request body is not JSON")

func sendToTreblle(treblleInfo MetaData) {
	bytesRepresentation, err := json.Marshal(treblleInfo)
	if err != nil {
		//
	}
	// By now our original request body should have been populated, so let's just use it with our custom request
	req, err := http.NewRequest(http.MethodPost, treblleURL, bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)
	}
	// We need to set the content type from the writer, it includes necessary boundary as well
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", Config.APIKey)

	// Do the request
	client := &http.Client{Timeout: timeoutDuration}
	_, err = client.Do(req)
	if err != nil {
		//
	}
}
