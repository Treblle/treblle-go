package treblle

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const (
	timeoutDuration = 2 * time.Second
)

type BaseUrlOptions struct {
	Debug bool
}

func getTreblleBaseUrl() string {
	// If custom endpoint is set, use it
	if Config.Endpoint != "" {
		return Config.Endpoint
	}

	// For debug mode
	if Config.Debug {
		return "https://debug.treblle.com/"
	}

	// Default Treblle endpoints
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
	log.Printf("Sending data to Treblle at: %s", baseUrl)

	bytesRepresentation, err := json.Marshal(treblleInfo)
	if err != nil {
		log.Printf("Error marshaling Treblle data: %v", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, baseUrl, bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Printf("Error creating Treblle request: %v", err)
		return
	}
	// Set the content type from the writer, it includes necessary boundary as well
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", Config.APIKey)

	client := &http.Client{
		Timeout: timeoutDuration,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending data to Treblle: %v", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("Treblle response status: %d", resp.StatusCode)
}
