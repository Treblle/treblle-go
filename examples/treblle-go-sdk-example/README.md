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
- Tunneling tool ( To expose your local server securely)

## Setting up Treblle

### 1. Create a Treblle Account

    If you don't have a Treblle account, go to [treblle.com](https://treblle.com) and create an account.

### 2. Create a New API Project

1. Once logged in, navigate to the dashboard and click on "Add New API".

   ![Treblle Add New API]()

2. Fill in the API details:
   - API Name (e.g., "Treblle_Go_SDK_Test_API")
   - Base URL (the URL where your API will be hosted, e.g., your ngrok URL)
   - Environment (e.g., "Development")
   - Platform (select "Go")

   ![Treblle API Setup]()

3. Click "Add New API" to create your project.

### 3. Get Your API Key and Project ID

1. After creating the API, go to "API Settings" to find your credentials.

   ![Treblle API Settings]()

2. Copy your API Key and Project ID. You'll need these to configure the SDK in your application.

## Installation

1. Clone the SDK repository and navigate to the examples folder:

   ```bash
   git clone https://github.com/Treblle/treblle-go.git
   cd treblle-go/examples/treblle-go-sdk-example
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

2. In a separate terminal, start a tunneling tool (this example uses ngrok) to expose your server:

   ```bash
   ngrok http 8085
   ```

3. Use the provided proxy URL to access your API.

4. Once your API is running, you'll be able to see your API requests in the Treblle dashboard:

   ![Treblle Dashboard]()

## API Endpoints

- `GET /api/v1/users` - Get all users
- `GET /api/v1/users/{id}` - Get a specific user
- `POST /api/v1/users` - Create a new user

## Setup Treblle Middleware

The Treblle middleware is set up in the `main.go` file:

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

### Customer Tracking with Treblle

Treblle allows you to track your API's customers with the help of Treblle headers `treblle-user-id` and `treblle-tag-id`

```go

func addTreblleHeadersMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

You can use [Aspen](https://treblle.com/product/aspen) to test the API or use a curl command:

```bash
curl -X POST "https://your-proxy-url.app/api/v1/users" \
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
