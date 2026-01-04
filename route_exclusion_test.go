package treblle

import (
	"testing"
)

func TestCompileRoutePatterns(t *testing.T) {
	testCases := []struct {
		name     string
		patterns []string
		expected struct {
			exactCount    int
			wildcardCount int
		}
	}{
		{
			name:     "empty patterns",
			patterns: []string{},
			expected: struct {
				exactCount    int
				wildcardCount int
			}{0, 0},
		},
		{
			name:     "exact patterns only",
			patterns: []string{"/health", "/metrics", "/status"},
			expected: struct {
				exactCount    int
				wildcardCount int
			}{3, 0},
		},
		{
			name:     "wildcard patterns only",
			patterns: []string{"/admin/*", "/api/*/internal/*"},
			expected: struct {
				exactCount    int
				wildcardCount int
			}{0, 2},
		},
		{
			name:     "mixed patterns",
			patterns: []string{"/health", "/admin/*", "/api/v1/debug"},
			expected: struct {
				exactCount    int
				wildcardCount int
			}{2, 1},
		},
		{
			name:     "patterns with varying case",
			patterns: []string{"/Health", "/METRICS", "/Status"},
			expected: struct {
				exactCount    int
				wildcardCount int
			}{3, 0},
		},
		{
			name:     "patterns with trailing slashes",
			patterns: []string{"/health/", "/metrics/"},
			expected: struct {
				exactCount    int
				wildcardCount int
			}{2, 0},
		},
		{
			name:     "patterns without leading slash",
			patterns: []string{"health", "metrics"},
			expected: struct {
				exactCount    int
				wildcardCount int
			}{2, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			compiled := compileRoutePatterns(tc.patterns)

			if tc.expected.exactCount == 0 && tc.expected.wildcardCount == 0 {
				if compiled != nil {
					t.Errorf("Expected nil for empty patterns, got non-nil")
				}
				return
			}

			if compiled == nil {
				t.Fatalf("Expected non-nil compiled patterns")
			}

			if len(compiled.exactMatches) != tc.expected.exactCount {
				t.Errorf("Expected %d exact matches, got %d", tc.expected.exactCount, len(compiled.exactMatches))
			}

			if len(compiled.wildcardRules) != tc.expected.wildcardCount {
				t.Errorf("Expected %d wildcard rules, got %d", tc.expected.wildcardCount, len(compiled.wildcardRules))
			}
		})
	}
}

func TestIsRouteExcluded(t *testing.T) {
	testCases := []struct {
		name           string
		patterns       []string
		routePath      string
		expectedResult bool
	}{
		// Exact match tests
		{
			name:           "exact match - health",
			patterns:       []string{"/health", "/metrics"},
			routePath:      "/health",
			expectedResult: true,
		},
		{
			name:           "exact match - no match",
			patterns:       []string{"/health", "/metrics"},
			routePath:      "/users",
			expectedResult: false,
		},
		{
			name:           "case insensitive exact match",
			patterns:       []string{"/health"},
			routePath:      "/HEALTH",
			expectedResult: true,
		},
		{
			name:           "exact match - different path",
			patterns:       []string{"/api/health"},
			routePath:      "/api/health",
			expectedResult: true,
		},

		// Wildcard tests
		{
			name:           "simple wildcard - match",
			patterns:       []string{"/admin/*"},
			routePath:      "/admin/users",
			expectedResult: true,
		},
		{
			name:           "simple wildcard - nested match",
			patterns:       []string{"/admin/*"},
			routePath:      "/admin/users/123",
			expectedResult: true,
		},
		{
			name:           "simple wildcard - deep nested match",
			patterns:       []string{"/admin/*"},
			routePath:      "/admin/users/roles/permissions",
			expectedResult: true,
		},
		{
			name:           "simple wildcard - no match",
			patterns:       []string{"/admin/*"},
			routePath:      "/api/users",
			expectedResult: false,
		},
		{
			name:           "simple wildcard - partial prefix no match",
			patterns:       []string{"/admin/*"},
			routePath:      "/administrator",
			expectedResult: false,
		},
		{
			name:           "multiple wildcards - match",
			patterns:       []string{"/api/*/internal/*"},
			routePath:      "/api/v1/internal/debug",
			expectedResult: true,
		},
		{
			name:           "multiple wildcards - nested match",
			patterns:       []string{"/api/*/internal/*"},
			routePath:      "/api/v2/internal/metrics/detailed",
			expectedResult: true,
		},
		{
			name:           "multiple wildcards - no match",
			patterns:       []string{"/api/*/internal/*"},
			routePath:      "/api/v1/external/debug",
			expectedResult: false,
		},
		{
			name:           "multiple wildcards - missing segment",
			patterns:       []string{"/api/*/internal/*"},
			routePath:      "/api/v1",
			expectedResult: false,
		},
		{
			name:           "wildcard in middle",
			patterns:       []string{"/api/*/users"},
			routePath:      "/api/v1/users",
			expectedResult: true,
		},
		{
			name:           "wildcard in middle - no match with extra segments",
			patterns:       []string{"/api/*/users"},
			routePath:      "/api/v1/users/123",
			expectedResult: false,
		},

		// Edge cases
		{
			name:           "trailing slash normalization - pattern",
			patterns:       []string{"/health/"},
			routePath:      "/health",
			expectedResult: true,
		},
		{
			name:           "trailing slash normalization - route",
			patterns:       []string{"/health"},
			routePath:      "/health/",
			expectedResult: true,
		},
		{
			name:           "empty patterns",
			patterns:       []string{},
			routePath:      "/anything",
			expectedResult: false,
		},
		{
			name:           "normalized route with params",
			patterns:       []string{"/users/{id}"},
			routePath:      "/users/{id}",
			expectedResult: true,
		},
		{
			name:           "root path",
			patterns:       []string{"/"},
			routePath:      "/",
			expectedResult: true,
		},
		{
			name:           "wildcard at root",
			patterns:       []string{"/*"},
			routePath:      "/anything",
			expectedResult: true,
		},
		{
			name:           "multiple patterns - first match",
			patterns:       []string{"/health", "/metrics", "/admin/*"},
			routePath:      "/health",
			expectedResult: true,
		},
		{
			name:           "multiple patterns - last match",
			patterns:       []string{"/health", "/metrics", "/admin/*"},
			routePath:      "/admin/dashboard",
			expectedResult: true,
		},
		{
			name:           "case insensitive wildcard",
			patterns:       []string{"/admin/*"},
			routePath:      "/ADMIN/users",
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			compiled := compileRoutePatterns(tc.patterns)
			result := isRouteExcluded(tc.routePath, compiled)

			if result != tc.expectedResult {
				t.Errorf("Expected %v, got %v for route %s with patterns %v", tc.expectedResult, result, tc.routePath, tc.patterns)
			}
		})
	}
}

func TestWildcardPatternMatching(t *testing.T) {
	testCases := []struct {
		name           string
		pattern        string
		routePath      string
		expectedResult bool
	}{
		{
			name:           "single wildcard at end",
			pattern:        "/admin/*",
			routePath:      "/admin/dashboard",
			expectedResult: true,
		},
		{
			name:           "single wildcard - deep nesting",
			pattern:        "/admin/*",
			routePath:      "/admin/users/roles/permissions",
			expectedResult: true,
		},
		{
			name:           "wildcard in middle",
			pattern:        "/api/*/users",
			routePath:      "/api/v1/users",
			expectedResult: true,
		},
		{
			name:           "wildcard in middle - no match (extra segments)",
			pattern:        "/api/*/users",
			routePath:      "/api/v1/users/123",
			expectedResult: false,
		},
		{
			name:           "multiple wildcards",
			pattern:        "/api/*/internal/*/debug",
			routePath:      "/api/v2/internal/tools/debug",
			expectedResult: true,
		},
		{
			name:           "multiple wildcards - no match",
			pattern:        "/api/*/internal/*/debug",
			routePath:      "/api/v2/external/tools/debug",
			expectedResult: false,
		},
		{
			name:           "wildcard with single char segment",
			pattern:        "/a/*/c",
			routePath:      "/a/b/c",
			expectedResult: true,
		},
		{
			name:           "pattern longer than route",
			pattern:        "/api/v1/users/*",
			routePath:      "/api/v1",
			expectedResult: false,
		},
		{
			name:           "empty segment matching",
			pattern:        "//admin/*",
			routePath:      "//admin/users",
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			wp := compileWildcardPattern(tc.pattern)
			result := matchWildcardPattern(tc.routePath, wp)

			if result != tc.expectedResult {
				t.Errorf("Expected %v, got %v for pattern %s matching route %s",
					tc.expectedResult, result, tc.pattern, tc.routePath)
			}
		})
	}
}

func TestCompileWildcardPattern(t *testing.T) {
	testCases := []struct {
		name            string
		pattern         string
		expectedSegments int
		wildcardIndices  []int
	}{
		{
			name:             "simple wildcard",
			pattern:          "/admin/*",
			expectedSegments: 3,
			wildcardIndices:  []int{2},
		},
		{
			name:             "multiple wildcards",
			pattern:          "/api/*/internal/*",
			expectedSegments: 5,
			wildcardIndices:  []int{2, 4},
		},
		{
			name:             "no wildcards",
			pattern:          "/health/check",
			expectedSegments: 3,
			wildcardIndices:  []int{},
		},
		{
			name:             "wildcard at start",
			pattern:          "/*/admin",
			expectedSegments: 3,
			wildcardIndices:  []int{1},
		},
		{
			name:             "all wildcards",
			pattern:          "/*/*",
			expectedSegments: 3,
			wildcardIndices:  []int{1, 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			wp := compileWildcardPattern(tc.pattern)

			if len(wp.segments) != tc.expectedSegments {
				t.Errorf("Expected %d segments, got %d", tc.expectedSegments, len(wp.segments))
			}

			// Count wildcards
			wildcardCount := 0
			for i, isWildcard := range wp.hasWildcard {
				if isWildcard {
					// Check if this index is in expected wildcards
					found := false
					for _, expectedIdx := range tc.wildcardIndices {
						if i == expectedIdx {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Unexpected wildcard at index %d", i)
					}
					wildcardCount++
				}
			}

			if wildcardCount != len(tc.wildcardIndices) {
				t.Errorf("Expected %d wildcards, got %d", len(tc.wildcardIndices), wildcardCount)
			}
		})
	}
}
