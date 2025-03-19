# Treblle Go SDK

This Go SDK enables you to seamlessly integrate Treblle with your Go API project, providing real-time API monitoring, request/response logging, and more.

## Installation

```bash
go get github.com/pratimbhosale/treblle-go
```

## Configuration

```go
import (
    "github.com/pratimbhosale/treblle-go"
)

func main() {
    treblle.Configure(treblle.Configuration{
        APIKey:    "your-treblle-api-key",
        ProjectID: "your-treblle-project-id",
    })
    
    // Your API server setup
    // ...
}
```

## Usage with Different Routers

### With Gorilla Mux (Recommended)

The SDK automatically extracts route patterns from Gorilla Mux:

```go
import (
    "github.com/gorilla/mux"
    "github.com/pratimbhosale/treblle-go"
)

func main() {
    // Configure Treblle
    treblle.Configure(treblle.Configuration{
        APIKey:    "your-treblle-api-key",
        ProjectID: "your-treblle-project-id",
    })

    // Create a new router
    r := mux.NewRouter()
    
    // Apply the Treblle middleware to the router
    r.Use(treblle.Middleware)

    // Define your routes
    r.HandleFunc("/users", getUsersHandler).Methods("GET")
    r.HandleFunc("/users/{id}", getUserHandler).Methods("GET")
    
    http.ListenAndServe(":8080", r)
}
```

### With Standard HTTP Package

For the standard library's HTTP server, use the `HandleFunc` helper to properly set route patterns:

```go
import (
    "net/http"
    "github.com/pratimbhosale/treblle-go"
)

func main() {
    // Configure Treblle
    treblle.Configure(treblle.Configuration{
        APIKey:    "your-treblle-api-key",
        ProjectID: "your-treblle-project-id",
    })

    // Create a new serve mux
    mux := http.NewServeMux()
    
    // Define routes with route path patterns
    mux.Handle("/users", treblle.Middleware(treblle.HandleFunc("/users", getUsersHandler)))
    mux.Handle("/users/", treblle.Middleware(treblle.HandleFunc("/users/:id", getUserHandler)))
    
    http.ListenAndServe(":8080", mux)
}
```

### With Other Router Libraries

For other router libraries, use the `WithRoutePath` function to set route patterns:

```go
// Example with a hypothetical router
router.GET("/users/:id", wrapHandler(treblle.WithRoutePath("/users/:id", 
    treblle.Middleware(http.HandlerFunc(getUserHandler)))))
```

## Manual Route Path Setting

You can also set route paths programmatically in your handlers:

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    // Set the route path for this specific request
    r = treblle.SetRoutePath(r, "/api/custom/:param")
    
    // Your handler logic
    // ...
}
```

## Examples

Check the `examples` directory for complete example applications:

- `gorilla_example`: Shows integration with Gorilla Mux
- `standard_example`: Shows integration with the standard HTTP package

## License

[License information]