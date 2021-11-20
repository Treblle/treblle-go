package treblle

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

const (
	timeoutDuration = 2 * time.Second
)

func sendToTreblle(treblleInfo MetaData) {
	bytesRepresentation, err := json.Marshal(treblleInfo)
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, Config.ServerURL, bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		return
	}
	// Set the content type from the writer, it includes necessary boundary as well
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", Config.APIKey)

	// Do the request
	client := &http.Client{Timeout: timeoutDuration}
	_, _ = client.Do(req)
}
