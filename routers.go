package treblle

import (
	"net/http"
)

// WithRoutePath is a helper function to set the route pattern for a specific handler
// This is useful for non-gorilla/mux routers or custom router implementations
func WithRoutePath(pattern string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add the route pattern to the request context
		r = SetRoutePath(r, pattern)
		handler.ServeHTTP(w, r)
	})
}

// HandleFunc is a helper that wraps an http.HandlerFunc with a route pattern
// This simplifies route pattern setting for standard http handlers
func HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) http.Handler {
	return WithRoutePath(pattern, http.HandlerFunc(handler))
}

// Example usage with standard library:
//
// For http.ServeMux:
//   mux := http.NewServeMux()
//   mux.Handle("/users", treblle.Middleware(treblle.HandleFunc("/users", listUsersHandler)))
//   mux.Handle("/users/{id}", treblle.Middleware(treblle.HandleFunc("/users/{id}", getUserHandler)))
//
// For routers that don't expose path templates:
//   router.GET("/users/:id", wrapHandler(treblle.WithRoutePath("/users/:id", 
//     treblle.Middleware(http.HandlerFunc(getUserHandler)))))
//
// For gorilla/mux (automatic support built-in):
//   r := mux.NewRouter()
//   r.Use(treblle.Middleware)  // Will automatically extract route templates
//
// For subrouters:
//   r := mux.NewRouter()
//   api := r.PathPrefix("/api").Subrouter()
//   api.Use(treblle.Middleware)  // Will automatically extract route templates