package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	treblle "github.com/treblle/treblle-go"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var (
	users  = make(map[int]User)
	nextID = 1
	mu     sync.Mutex
)

func getUsers(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var userList []User
	for _, user := range users {
		userList = append(userList, user)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userList)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil || users[id].ID == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users[id])
}

func createUser(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	user.ID = nextID
	nextID++
	users[user.ID] = user

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

type DebugTransport struct {
	Transport http.RoundTripper
}

func (d *DebugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Log the request
	reqDump, err := httputil.DumpRequestOut(req, true)
	if err == nil {
		log.Printf("Outgoing Request:\n%s\n", string(reqDump))
	}

	// Perform the request
	resp, err := d.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Log the response
	respDump, err := httputil.DumpResponse(resp, true)
	if err == nil {
		log.Printf("Incoming Response:\n%s\n", string(respDump))
	}

	return resp, nil
}

// Custom middleware to add Treblle user ID and trace ID headers
func addTreblleHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// In a real application, you would get the user from the session, JWT token, etc.
		// For demonstration purposes, we'll check for a user ID in a custom header
		userID := r.Header.Get("X-User-ID")
		if userID != "" {
			// Add Treblle user ID header
			r.Header.Set("treblle-user-id", userID)
			
			// Add Treblle trace ID header (could be a correlation ID or any other identifier)
			traceID := r.Header.Get("X-Trace-ID")
			if traceID != "" {
				r.Header.Set("treblle-tag-id", traceID)
			} else {
				// Generate a random trace ID if none is provided
				r.Header.Set("treblle-tag-id", "trace-"+strconv.Itoa(int(time.Now().UnixNano())))
			}
		}
		
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}

	// Enable HTTP client debugging
	debug := os.Getenv("DEBUG")
	if debug == "true" {
		// Set up a custom HTTP client with debug transport
		http.DefaultTransport = &DebugTransport{
			Transport: http.DefaultTransport,
		}
	}

	treblle.Configure(treblle.Configuration{
		APIKey:                 os.Getenv("TREBLLE_API_KEY"),
		ProjectID:              os.Getenv("TREBLLE_PROJECT_ID"),
		AdditionalFieldsToMask: []string{"bank_account", "routing_number", "tax_id", "auth_token", "ssn", "api_key", "password", "credit_card"},
		Debug:                  debug == "true",
	})

	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1").Subrouter()

	// Apply our custom middleware first, then Treblle middleware
	api.Use(addTreblleHeadersMiddleware)
	api.Use(treblle.Middleware)

	// Simple route definitions
	api.HandleFunc("/users", getUsers).Methods("GET")
	api.HandleFunc("/users/{id}", getUser).Methods("GET")
	api.HandleFunc("/users", createUser).Methods("POST")

	log.Println("Server running on port 8085")
	log.Fatal(http.ListenAndServe(":8085", r))
}
