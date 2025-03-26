package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	treblle "github.com/Treblle/treblle-go/v2" // Updated to v2 import path
	"github.com/gorilla/mux"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	// Simulate getting a user from database
	user := User{
		ID:   userID,
		Name: "John Doe",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("POST request received to /users")

	// Parse the request body
	var newUser struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		ApiKey   string `json:"api_key"`
		CcNumber string `json:"credit_card"`
	}

	// Read the body content for debugging
	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Reset the body

	log.Printf("Request body: %s\n", string(bodyBytes))

	err := json.NewDecoder(bytes.NewBuffer(bodyBytes)).Decode(&newUser)
	if err != nil {
		log.Printf("Error decoding JSON: %v\n", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Received user data: %+v\n", newUser)

	// Create a response with masked data
	response := map[string]interface{}{
		"success": true,
		"message": "User created successfully",
		"user": map[string]string{
			"name": newUser.Name,
			"id":   "new-user-123",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

	log.Println("POST request processed successfully")
}

func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate getting users from database
	users := []User{
		{ID: "1", Name: "John Doe"},
		{ID: "2", Name: "Jane Smith"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func main() {
	// Configure Treblle
	treblle.Configure(treblle.Configuration{
		SDK_TOKEN: "Treblle SDK Token", // Set your Treblle SDK Token
		API_KEY:   "Treblle API Key",   // Set your Treblle API Key
		Debug:     true,                // Enable debug mode to see what's being sent to Treblle
	})

	// Create a new router
	r := mux.NewRouter()

	// Apply Treblle middleware to the router
	r.Use(treblle.Middleware)

	// Define routes
	r.HandleFunc("/users", getUsersHandler).Methods("GET")
	r.HandleFunc("/users/{id}", getUserHandler).Methods("GET")
	r.HandleFunc("/users", createUserHandler).Methods("POST")

	// Start server
	log.Println("Starting server on http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}
