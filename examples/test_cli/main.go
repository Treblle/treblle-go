package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/treblle/treblle-go"
)

// Create a test request and response for debugging

func main() {
	// Set environment variables for testing
	os.Setenv("TREBLLE_API_KEY", "test_api_key_12345")
	os.Setenv("TREBLLE_PROJECT_ID", "test_project_id_67890")
	os.Setenv("TREBLLE_ENDPOINT", "https://test-api.treblle.com")
	os.Setenv("TREBLLE_IGNORED_ENVIRONMENTS", "local,testing")

	// Configure Treblle with test values
	treblle.Configure(treblle.Configuration{
		APIKey:                  "test_api_key_12345",
		ProjectID:               "test_project_id_67890",
		AdditionalFieldsToMask:  []string{"custom_field", "secret_data"},
		MaskingEnabled:          true,
		AsyncProcessingEnabled:  true,
		MaxConcurrentProcessing: 5,
		AsyncShutdownTimeout:    time.Second * 3,
		IgnoredEnvironments:     []string{"local", "testing"},
	})

	// Create test request and response data for CLI debugging

	// Create a router
	r := chi.NewRouter()

	// Use Treblle middleware
	r.Use(treblle.Middleware)

	// Define a test endpoint that returns JSON
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Create a test response
		response := map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"message":      "Hello from Treblle!",
				"timestamp":    time.Now().Unix(),
				"custom_field": "This should be masked",
				"user": map[string]interface{}{
					"id":       123,
					"email":    "test@example.com",
					"password": "supersecret",
				},
			},
		}

		// Send the response
		json.NewEncoder(w).Encode(response)
	})

	// Define a POST endpoint for testing
	r.Post("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Create a test response
		response := map[string]interface{}{
			"status":  "success",
			"message": "POST request received",
		}

		// Send the response
		json.NewEncoder(w).Encode(response)
	})

	// Start the server in a goroutine
	go func() {
		fmt.Println("Starting test server on :3334...")
		if err := http.ListenAndServe(":3334", r); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Make a test request to our server to generate data for the CLI
	go func() {
		// Wait a moment for the server to start
		time.Sleep(500 * time.Millisecond)

		// Create a test request with some data
		testData := map[string]interface{}{
			"action": "test",
			"params": map[string]interface{}{
				"api_key": "should_be_masked",
				"data":    "test data",
			},
		}

		jsonData, _ := json.Marshal(testData)

		req, _ := http.NewRequest("POST", "http://localhost:3334/test", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test_token_12345")

		// Send the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error making test request: %v", err)
			return
		}
		defer resp.Body.Close()

		fmt.Println("Test request completed. Now run the CLI tool to see the debug information.")
		fmt.Println("Run: ./bin/treblle-go -debug")
	}()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Press Ctrl+C to exit...")
	<-sigChan
	fmt.Println("Shutting down...")
}
