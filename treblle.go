package treblle

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"time"
)

const (
	timeoutDuration = 2 * time.Second
)

func getTreblleBaseUrl() string {
	treblleBaseUrls := []string{
		"https://rocknrolla.treblle.com",
		"https://punisher.treblle.com",
		"https://sicario.treblle.com",
	}

	rand.Seed(time.Now().Unix())
	randomUrlIndex := rand.Intn(len(treblleBaseUrls))

	return treblleBaseUrls[randomUrlIndex]
}

func sendToTreblle(treblleInfo MetaData) {
	baseUrl := getTreblleBaseUrl()

	bytesRepresentation, err := json.Marshal(treblleInfo)
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, baseUrl, bytes.NewBuffer(bytesRepresentation))
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
