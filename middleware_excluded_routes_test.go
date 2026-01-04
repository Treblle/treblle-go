package treblle

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareExclusionIntegration(t *testing.T) {
	testCases := []struct {
		name              string
		excludedRoutes    []string
		requestPath       string
		routePattern      string
		shouldCallNext    bool
		description       string
	}{
		{
			name:           "exact match excluded",
			excludedRoutes: []string{"/health"},
			requestPath:    "/health",
			routePattern:   "/health",
			shouldCallNext: true,
			description:    "Health check should be excluded and call next handler",
		},
		{
			name:           "exact match not excluded",
			excludedRoutes: []string{"/health"},
			requestPath:    "/users",
			routePattern:   "/users",
			shouldCallNext: true,
			description:    "Users endpoint should not be excluded",
		},
		{
			name:           "wildcard excluded - direct child",
			excludedRoutes: []string{"/admin/*"},
			requestPath:    "/admin/dashboard",
			routePattern:   "/admin/dashboard",
			shouldCallNext: true,
			description:    "Admin dashboard should be excluded by wildcard",
		},
		{
			name:           "wildcard excluded - nested child",
			excludedRoutes: []string{"/admin/*"},
			requestPath:    "/admin/users/123",
			routePattern:   "/admin/users/{id}",
			shouldCallNext: true,
			description:    "Nested admin route should be excluded by wildcard",
		},
		{
			name:           "wildcard not excluded",
			excludedRoutes: []string{"/admin/*"},
			requestPath:    "/api/users",
			routePattern:   "/api/users",
			shouldCallNext: true,
			description:    "API users should not be excluded",
		},
		{
			name:           "multiple wildcards excluded",
			excludedRoutes: []string{"/api/*/internal/*"},
			requestPath:    "/api/v1/internal/debug",
			routePattern:   "/api/{version}/internal/{endpoint}",
			shouldCallNext: true,
			description:    "Internal API endpoint should be excluded",
		},
		{
			name:           "case insensitive match",
			excludedRoutes: []string{"/health"},
			requestPath:    "/HEALTH",
			routePattern:   "/HEALTH",
			shouldCallNext: true,
			description:    "Case insensitive health check should be excluded",
		},
		{
			name:           "multiple patterns - one matches",
			excludedRoutes: []string{"/health", "/metrics", "/admin/*"},
			requestPath:    "/metrics",
			routePattern:   "/metrics",
			shouldCallNext: true,
			description:    "Metrics should be excluded from multiple patterns",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Track if next handler was called
			nextCalled := false

			// Configure Treblle with excluded routes (no API keys to prevent actual sending)
			Configure(Configuration{
				ExcludedRoutes: tc.excludedRoutes,
			})

			// Create a test handler that sets a flag when called
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			// Wrap with Treblle middleware
			wrappedHandler := Middleware(testHandler)

			// Create test request
			req := httptest.NewRequest("GET", tc.requestPath, nil)

			// Set route pattern in context (simulating what a router would do)
			if tc.routePattern != "" {
				req = SetRoutePath(req, tc.routePattern)
			}

			// Create response recorder
			rec := httptest.NewRecorder()

			// Execute middleware
			wrappedHandler.ServeHTTP(rec, req)

			// Verify next handler was called
			if nextCalled != tc.shouldCallNext {
				t.Errorf("Expected next handler called=%v, got %v. %s", tc.shouldCallNext, nextCalled, tc.description)
			}

			// Verify response
			if rec.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", rec.Code)
			}
		})
	}
}

func TestConfigurationExcludedRoutesLoading(t *testing.T) {
	testCases := []struct {
		name           string
		config         Configuration
		expectedRoutes []string
		description    string
	}{
		{
			name: "load from config",
			config: Configuration{
				ExcludedRoutes: []string{"/health", "/metrics"},
			},
			expectedRoutes: []string{"/health", "/metrics"},
			description:    "Should load excluded routes from configuration",
		},
		{
			name: "empty config",
			config: Configuration{
				ExcludedRoutes: []string{},
			},
			expectedRoutes: []string{},
			description:    "Should handle empty excluded routes",
		},
		{
			name: "wildcard patterns",
			config: Configuration{
				ExcludedRoutes: []string{"/admin/*", "/api/*/internal/*"},
			},
			expectedRoutes: []string{"/admin/*", "/api/*/internal/*"},
			description:    "Should handle wildcard patterns",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Configure Treblle
			Configure(tc.config)

			// Verify excluded routes were loaded
			if len(Config.ExcludedRoutes) != len(tc.expectedRoutes) {
				t.Errorf("Expected %d excluded routes, got %d. %s",
					len(tc.expectedRoutes), len(Config.ExcludedRoutes), tc.description)
			}

			// Verify compilation happened
			if len(tc.expectedRoutes) > 0 {
				if Config.compiledExclusions == nil {
					t.Errorf("Expected compiled exclusions to be non-nil. %s", tc.description)
				}
			}
		})
	}
}

func TestExcludedRoutesWithNormalization(t *testing.T) {
	testCases := []struct {
		name           string
		excludedRoutes []string
		requestPath    string
		routePattern   string
		shouldExclude  bool
		description    string
	}{
		{
			name:           "normalized route with param placeholder",
			excludedRoutes: []string{"/users/{id}"},
			requestPath:    "/users/123",
			routePattern:   "/users/{id}",
			shouldExclude:  true,
			description:    "Should exclude normalized route with param",
		},
		{
			name:           "wildcard matches normalized route",
			excludedRoutes: []string{"/users/*"},
			requestPath:    "/users/123/posts/456",
			routePattern:   "/users/{userId}/posts/{postId}",
			shouldExclude:  true,
			description:    "Wildcard should match normalized nested route",
		},
		{
			name:           "exact match on normalized route",
			excludedRoutes: []string{"/api/v1/users/{id}"},
			requestPath:    "/api/v1/users/789",
			routePattern:   "/api/v1/users/{id}",
			shouldExclude:  true,
			description:    "Exact match should work on normalized routes",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Configure Treblle
			Configure(Configuration{
				ExcludedRoutes: tc.excludedRoutes,
			})

			// Normalize the route pattern (simulating what middleware does)
			normalizedRoute := normalizeRoutePath(tc.routePattern)

			// Check if route is excluded
			isExcluded := isRouteExcluded(normalizedRoute, Config.compiledExclusions)

			if isExcluded != tc.shouldExclude {
				t.Errorf("Expected exclusion=%v, got %v for route %s (normalized: %s). %s",
					tc.shouldExclude, isExcluded, tc.routePattern, normalizedRoute, tc.description)
			}
		})
	}
}

func TestMiddlewareWithEmptyExclusionList(t *testing.T) {
	// Configure with no excluded routes
	Configure(Configuration{
		ExcludedRoutes: []string{},
	})

	nextCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := Middleware(testHandler)
	req := httptest.NewRequest("GET", "/any/path", nil)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	// Next should still be called (no exclusions)
	if !nextCalled {
		t.Error("Expected next handler to be called when no routes are excluded")
	}
}

func TestMiddlewareWithNilCompiledExclusions(t *testing.T) {
	// Explicitly set compiledExclusions to nil
	Config.compiledExclusions = nil
	Config.ExcludedRoutes = []string{}

	nextCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := Middleware(testHandler)
	req := httptest.NewRequest("GET", "/any/path", nil)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	// Next should be called (nil exclusions means no exclusions)
	if !nextCalled {
		t.Error("Expected next handler to be called when compiledExclusions is nil")
	}
}
