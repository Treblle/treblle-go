package treblle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	// Use the context-aware version with a default timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	sendToTreblleWithContext(ctx, treblleInfo)
}

// sendToTreblleWithContext sends data to Treblle with context support
func sendToTreblleWithContext(ctx context.Context, treblleInfo MetaData) error {
	baseUrl := getTreblleBaseUrl()

	// Print debug info if debug mode is enabled
	if Config.Debug {
		fmt.Printf("\n==== DEBUG: TREBLLE ENDPOINT ====\n")
		fmt.Printf("Sending to: %s\n", baseUrl)
		fmt.Printf("================================\n")
	}

	bytesRepresentation, err := json.Marshal(treblleInfo)
	if err != nil {
		return err
	}

	// Print debug info if debug mode is enabled
	if Config.Debug {
		prettyJson, _ := json.MarshalIndent(treblleInfo, "", "  ")
		fmt.Println("\n==== DEBUG: TREBLLE PAYLOAD ====")
		fmt.Println(string(prettyJson))
		fmt.Println("=================================")

		// Highlight the critical fields used for endpoint grouping
		fmt.Println("\n🔍 Important fields for endpoint grouping:")
		fmt.Printf("Request 'url' field: %s\n", treblleInfo.Data.Request.Url)
		fmt.Printf("Request 'route_path' field: %s\n", treblleInfo.Data.Request.RoutePath)
		//fmt.Printf("Request 'full_url' field: %s\n", treblleInfo.Data.Request.FullUrl)
		fmt.Println("=================================")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseUrl, bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		return err
	}
	// Set the content type from the writer, it includes necessary boundary as well
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", Config.APIKey)

	client := &http.Client{
		// No need for timeout here as we're using context timeout
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if Config.Debug {
		fmt.Printf("\n==== DEBUG: TREBLLE RESPONSE ====\n")
		fmt.Printf("Status: %s\n", resp.Status)

		// Read and log response body
		respBody := make([]byte, 1024)
		n, _ := resp.Body.Read(respBody)
		if n > 0 {
			fmt.Printf("Response: %s\n", respBody[:n])
		}
		fmt.Printf("================================\n")
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("treblle api returned error status: %s", resp.Status)
	}

	return nil
}
