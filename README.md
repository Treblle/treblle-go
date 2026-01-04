# Treblle - API Intelligence Platform

[![Treblle API Intelligence](https://github.com/user-attachments/assets/b268ae9e-7c8a-4ade-95da-b4ac6fce6eea)](https://treblle.com)

[Website](http://treblle.com/) • [Documentation](https://docs.treblle.com/) • [Pricing](https://treblle.com/pricing)

Treblle is an API intelligence platfom that helps developers, teams and organizations understand their APIs from a single integration point.

---

## Treblle Go Lang SDK

### Requirements

| Go Version | Support Status |
|------------|----------------|
| 1.23+      | Fully Supported |
| 1.21 - 1.22 | Should work, not officially tested |
| < 1.21     | Not Supported |

### Router & Framework Support

| Router/Framework | Support Level | Native Integration |
|------------------|---------------|-------------------|
| **Standard Library** (`net/http`) | Fully Supported | Yes |
| **Gin** | Fully Supported | Yes |
| **Gorilla Mux** | ✅ Fully Supported | Yes |
| **Chi** | Fully Supported | Yes |
| **Echo** | Compatible | No |
| **Fiber** | Compatible | No |
| **Other Routers** | Compatible | No |

## Installation

### 1. Install the Package

```bash
go get github.com/Treblle/treblle-go/v2
```

### 2. Get Your Credentials
Get your SDK Token and API Key from the [Treblle Dashboard](https://platform.treblle.com).

### 3. Configure Treblle

```go
import (
    "github.com/treblle/treblle-go/v2"
)

func main() {
    treblle.Configure(treblle.Configuration{
        SDK_TOKEN: "your-treblle-sdk-token",
        API_KEY:   "your-treblle-api-key",
    })

    // Your API server setup
    // ...
}
```

### With Gin

The SDK provides native support for the Gin framework with automatic route pattern extraction:

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/treblle/treblle-go/v2"
)

func main() {
    // Configure Treblle
    treblle.Configure(treblle.Configuration{
        SDK_TOKEN: "your-treblle-sdk-token",
        API_KEY:   "your-treblle-api-key",
    })

    // Create Gin router
    r := gin.Default()

    // Apply Treblle middleware - route patterns are automatically extracted!
    r.Use(treblle.GinMiddleware())

    // Define your routes
    r.GET("/users", getUsersHandler)
    r.GET("/users/:id", getUserHandler)
    r.POST("/users", createUserHandler)

    r.Run(":8080")
}
```

### With Gorilla Mux

The SDK automatically extracts route patterns from Gorilla Mux:

```go
import (
    "github.com/gorilla/mux"
    "github.com/treblle/treblle-go"
)

func main() {
    // Configure Treblle
    treblle.Configure(treblle.Configuration{
        SDK_TOKEN: "your-treblle-sdk-token",
        API_KEY:   "your-treblle-api-key",
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
    "github.com/treblle/treblle-go"
)

func main() {
    // Configure Treblle
    treblle.Configure(treblle.Configuration{
        SDK_TOKEN: "your-treblle-sdk-token",
        API_KEY:   "your-treblle-api-key",
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

## Excluding Routes

You can configure Treblle to exclude specific routes from being tracked. This is useful for health checks, metrics endpoints, internal APIs, or any routes you don't want to monitor.

### Basic Usage

```go
treblle.Configure(treblle.Configuration{
    SDK_TOKEN: "your-treblle-sdk-token",
    API_KEY:   "your-treblle-api-key",
    ExcludedRoutes: []string{
        "/health",           // Exact match
        "/metrics",          // Exact match
        "/admin/*",          // Wildcard: matches all admin routes
        "/api/*/internal/*", // Multiple wildcards
    },
})
```

### Pattern Types

**Exact Match:**
```go
ExcludedRoutes: []string{"/health", "/status", "/readiness"}
```
- Matches exactly `/health`, `/status`, and `/readiness`
- Case-insensitive matching
- Trailing slashes are normalized (`/health/` matches `/health`)

**Simple Wildcard:**
```go
ExcludedRoutes: []string{"/admin/*"}
```
- Matches `/admin/dashboard`, `/admin/users`, `/admin/users/123`, etc.
- The `*` matches any segment and **all nested paths**
- Perfect for excluding entire sections of your API

**Multiple Wildcards:**
```go
ExcludedRoutes: []string{
    "/api/*/internal/*",
    "/v*/debug/*",
}
```
- Each `*` matches exactly one path segment
- `/api/*/internal/*` matches `/api/v1/internal/debug`, `/api/v2/internal/metrics/detailed`, etc.
- Provides fine-grained control over exclusions

### Environment Variable

You can also set excluded routes via environment variable:
```bash
export TREBLLE_EXCLUDED_ROUTES="/health,/metrics,/admin/*"
```

The environment variable uses comma-separated values and is loaded automatically if `ExcludedRoutes` is not set in the configuration.


## Examples

Check the `examples` directory for complete example applications:

- `gorilla_example`: Shows integration with Gorilla Mux
- `standard_example`: Shows integration with the standard HTTP package


## Getting Help

If you continue to experience issues:

1. Enable `debug: true` and check console output
2. Verify your SDK token and API key are correct in Treblle dashboard
3. Test with a simple endpoint first
4. Check [Treblle documentation](https://docs.treblle.com) for the latest updates
5. Contact support at <https://treblle.com> or email support@treblle.com

## Support

If you have problems of any kind feel free to reach out via <https://treblle.com> or email support@treblle.com and we'll do our best to help you out.

## License

Copyright 2025, Treblle Inc. Licensed under the MIT license:
http://www.opensource.org/licenses/mit-license.php
