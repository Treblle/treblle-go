package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	treblle "github.com/Treblle/treblle-go/v2" // Updated to v2 path
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	userID := pathParts[len(pathParts)-1]

	// Simulate getting a user from database
	user := User{
		ID:   userID,
		Name: "John Doe",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Check if this is a POST request
	if r.Method == "POST" {
		createUserHandler(w, r)
		return
	}

	// Otherwise, handle as GET request
	// Simulate getting users from database
	users := []User{
		{ID: "1", Name: "John Doe"},
		{ID: "2", Name: "Jane Smith"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var newUser struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		ApiKey   string `json:"api_key"`
		CcNumber string `json:"credit_card"`
	}

	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create a response with masked data
	response := map[string]interface{}{
		"success": true,
		"message": "User created successfully",
		"user": map[string]string{
			"name": newUser.Name,
			"id":   "new-user-456",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Configure Treblle
	treblle.Configure(treblle.Configuration{
		SDK_TOKEN: "Treblle SDK Token", // Set your Treblle SDK Token
		API_KEY:   "Treblle API Key",   // Set your Treblle API Key
		Debug:     true,                // Enable debug mode to see what's being sent to Treblle
	})

	// Create a new serve mux
	mux := http.NewServeMux()

	// Define routes with route path patterns
	// This is important - we use treblle.HandleFunc to set the route pattern
	mux.Handle("/users", treblle.Middleware(treblle.HandleFunc("/users", getUsersHandler)))
	mux.Handle("/users/", treblle.Middleware(treblle.HandleFunc("/users/:id", getUserHandler)))

	// Start server
	log.Println("Starting server on http://localhost:8083")
	log.Fatal(http.ListenAndServe(":8083", mux))
}
