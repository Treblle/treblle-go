# Treblle

![Treblle Logo](https://github.com/user-attachments/assets/54f0c084-65bb-4431-b80d-cceab6c63dc3 "Treblle Logo")

[Integrations](https://docs.treblle.com/en/integrations) •
[Website](http://treblle.com/) •
[Docs](https://docs.treblle.com) •
[Blog](https://blog.treblle.com) •
[Twitter](https://twitter.com/treblleapi) •
[Discord](https://treblle.com/chat)

---

API Intelligence Platform.

Treblle is a lightweight SDK that helps Engineering and Product teams build, ship & maintain REST-based APIs faster.

## Features

![Treblle Features](https://github.com/user-attachments/assets/9b5f40ba-bec9-414b-af88-f1c1cc80781b "Treblle Features")

- [API Monitoring & Observability](https://www.treblle.com/features/api-monitoring-observability)
- [Auto-generated API Docs](https://www.treblle.com/features/auto-generated-api-docs)
- [API analytics](https://www.treblle.com/features/api-analytics)
- [Treblle API Score](https://www.treblle.com/features/api-quality-score)
- [API Lifecycle Collaboration](https://www.treblle.com/features/api-lifecycle)
- [Native Treblle Apps](https://www.treblle.com/features/native-apps)

## How Treblle Works

Once you've integrated a Treblle SDK in your codebase, this SDK will send requests and response data to your Treblle Dashboard.

In your Treblle Dashboard you get to see real-time requests to your API, auto-generated API docs, API analytics like how fast the response was for an endpoint, the load size of the response, etc.

Treblle also uses the requests sent to your Dashboard to calculate your API score which is a quality score that's calculated based on the performance, quality, and security best practices for your API.

> Visit [https://docs.treblle.com](http://docs.treblle.com) for the complete documentation.

## Security

### Masking fields

Masking fields ensure certain sensitive data are removed before being sent to Treblle.

To make sure masking is done before any data leaves your server [we built it into all our SDKs](https://docs.treblle.com/en/security/masked-fields#fields-masked-by-default).

This means data masking is super fast and happens on a programming level before the API request is sent to Treblle. You can [customize](https://docs.treblle.com/en/security/masked-fields#custom-masked-fields) exactly which fields are masked when you're integrating the SDK.

> Visit the [Masked fields](https://docs.treblle.com/en/security/masked-fields) section of the [docs](https://docs.sailscasts.com) for the complete documentation.

## Get Started

1. Sign in to [Treblle](https://platform.treblle.com).
2. [Create a Treblle project](https://docs.treblle.com/en/dashboard/projects#creating-a-project).
3. [Setup the SDK](#installation) for your platform.

## Installation

```bash
go get github.com/treblle/treblle-go/v2
```

## Configuration

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

## Usage with Different Routers

### With Gorilla Mux (Recommended)

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

## Available SDKs

Treblle provides [open-source SDKs](https://docs.treblle.com/en/integrations) that let you seamlessly integrate Treblle with your REST-based APIs.

- [`treblle-laravel`](https://github.com/Treblle/treblle-laravel): SDK for Laravel
- [`treblle-php`](https://github.com/Treblle/treblle-php): SDK for PHP
- [`treblle-symfony`](https://github.com/Treblle/treblle-symfony): SDK for Symfony
- [`treblle-lumen`](https://github.com/Treblle/treblle-lumen): SDK for Lumen
- [`treblle-sails`](https://github.com/Treblle/treblle-sails): SDK for Sails
- [`treblle-adonisjs`](https://github.com/Treblle/treblle-adonisjs): SDK for AdonisJS
- [`treblle-fastify`](https://github.com/Treblle/treblle-fastify): SDK for Fastify
- [`treblle-directus`](https://github.com/Treblle/treblle-directus): SDK for Directus
- [`treblle-strapi`](https://github.com/Treblle/treblle-strapi): SDK for Strapi
- [`treblle-express`](https://github.com/Treblle/treblle-express): SDK for Express
- [`treblle-koa`](https://github.com/Treblle/treblle-koa): SDK for Koa
- [`treblle-go`](https://github.com/Treblle/treblle-go): SDK for Go
- [`treblle-ruby`](https://github.com/Treblle/treblle-ruby): SDK for Ruby on Rails
- [`treblle-python`](https://github.com/Treblle/treblle-python): SDK for Python/Django

> See the [docs](https://docs.treblle.com/en/integrations) for more on SDKs and Integrations.

## Other Packages

Besides the SDKs, we also provide helpers and configuration used for SDK
development. If you're thinking about contributing to or creating a SDK, have a look at the resources
below:

- [`treblle-utils`](https://github.com/Treblle/treblle-utils): A set of helpers and
  utility functions useful for the JavaScript SDKs.
- [`php-utils`](https://github.com/Treblle/php-utils): A set of helpers and
  utility functions useful for the PHP SDKs.

## Community

First and foremost: **Star and watch this repository** to stay up-to-date.

Also, follow our [Blog](https://blog.treblle.com), and on [Twitter](https://twitter.com/treblleapi).

You can chat with the team and other members on [Discord](https://treblle.com/chat) and follow our tutorials and other video material at [YouTube](https://youtube.com/@treblle).

[![Treblle Discord](https://img.shields.io/badge/Treblle%20Discord-Join%20our%20Discord-F3F5FC?labelColor=7289DA&style=for-the-badge&logo=discord&logoColor=F3F5FC&link=https://treblle.com/chat)](https://treblle.com/chat)

[![Treblle YouTube](https://img.shields.io/badge/Treblle%20YouTube-Subscribe%20on%20YouTube-F3F5FC?labelColor=c4302b&style=for-the-badge&logo=YouTube&logoColor=F3F5FC&link=https://youtube.com/@treblle)](https://youtube.com/@treblle)

[![Treblle on Twitter](https://img.shields.io/badge/Treblle%20on%20Twitter-Follow%20Us-F3F5FC?labelColor=1DA1F2&style=for-the-badge&logo=Twitter&logoColor=F3F5FC&link=https://twitter.com/treblleapi)](https://twitter.com/treblleapi)

### How to contribute

Here are some ways of contributing to making Treblle better:

- **[Try out Treblle](https://docs.treblle.com/en/introduction#getting-started)**, and let us know ways to make Treblle better for you. Let us know here on [Discord](https://treblle.com/chat).
- Join our [Discord](https://treblle.com/chat) and connect with other members to share and learn from.
- Send a pull request to any of our [open source repositories](https://github.com/Treblle) on Github. Check the contribution guide on the repo you want to contribute to for more details about how to contribute. We're looking forward to your contribution!
