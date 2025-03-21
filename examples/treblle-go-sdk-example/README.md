# Treblle Go SDK Example with Gorilla Mux

This repository demonstrates how to integrate the [Treblle SDK](https://github.com/treblle/treblle-go) with a Go API built using the [Gorilla Mux](https://github.com/gorilla/mux) router. Treblle provides real-time API monitoring, automatic documentation, and powerful insights to help you build better APIs.

## Features

- Simple REST API with Gorilla Mux
- Treblle integration for API monitoring and analytics
- Environment variable configuration
- Custom middleware for user tracking with Treblle
- HTTP debugging capabilities

## Prerequisites

- Go 1.16 or higher
- [ngrok](https://ngrok.com/) for exposing your local server to the internet (Only for Testing purposes)

## Installation

1. Clone this repository:

   ```bash
   git clone https://github.com/timpratim/treblle-go-sdk-example.git
   cd ngroktest
   ```

2. Install dependencies:

   ```bash
   go mod download
   ```

3. Create a `.env` file in the root directory with your Treblle credentials:

   ```
   TREBLLE_API_KEY=your_treblle_api_key
   TREBLLE_PROJECT_ID=your_treblle_project_id
   DEBUG=true
   ```

## Running the API

1. Start the server:

   ```bash
   go run main.go
   ```

2. In a separate terminal, start ngrok to expose your server:

   ```bash
   ngrok http 8085
   ```

3. Use the provided ngrok URL to access your API.

## API Endpoints

- `GET /api/v1/users` - Get all users
- `GET /api/v1/users/{id}` - Get a specific user
- `POST /api/v1/users` - Create a new user

## Treblle Integration

### Basic Integration

The basic Treblle integration is set up in the `main.go` file:

```go
treblle.Configure(treblle.Configuration{
    APIKey:                 os.Getenv("TREBLLE_API_KEY"),
    ProjectID:              os.Getenv("TREBLLE_PROJECT_ID"),
    AdditionalFieldsToMask: []string{"bank_account", "routing_number", "tax_id", "auth_token", "ssn", "api_key", "password", "credit_card"},
    Debug:                  debug == "true",
})

r := mux.NewRouter()
api := r.PathPrefix("/api/v1").Subrouter()

// Apply Treblle middleware to the subrouter
api.Use(treblle.Middleware)
```

### User Tracking with Treblle

This example includes a custom middleware that adds user tracking capabilities to Treblle:

```go
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
```

Apply the middleware before the Treblle middleware:

```go
// Apply our custom middleware first, then Treblle middleware
api.Use(addTreblleHeadersMiddleware)
api.Use(treblle.Middleware)
```

## Testing with User Tracking

You can use the included `test-request.sh` script to test the API with user tracking:

```bash
./test-request.sh
```

This script sends requests with the `X-User-ID` and `X-Trace-ID` headers, which are then transformed into Treblle headers for tracking in the dashboard.

Example curl command:

```bash
curl -X POST "https://your-ngrok-url.ngrok-free.app/api/v1/users" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 12345" \
  -H "X-Trace-ID: test-trace-123" \
  -d '{"name": "John Doe", "email": "john@example.com"}'
```

## HTTP Debugging

This example includes HTTP request/response debugging capabilities that can be enabled by setting the `DEBUG` environment variable to `true`. When enabled, all HTTP requests and responses will be logged to the console.

## License

[MIT](LICENSE)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
