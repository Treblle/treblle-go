package treblle

import (
	"strings"
)

// compiledRoutePatterns holds pre-compiled route exclusion patterns
// This is computed once during Configure() for optimal performance
type compiledRoutePatterns struct {
	exactMatches  map[string]bool   // O(1) lookup for exact patterns
	wildcardRules []wildcardPattern // Ordered wildcard patterns
}

// wildcardPattern represents a single wildcard pattern
type wildcardPattern struct {
	original    string   // Original pattern (for debugging)
	segments    []string // Split by '/'
	hasWildcard []bool   // Which segments are wildcards
}

// compileRoutePatterns pre-processes route patterns for efficient matching
func compileRoutePatterns(patterns []string) *compiledRoutePatterns {
	if len(patterns) == 0 {
		return nil
	}

	compiled := &compiledRoutePatterns{
		exactMatches:  make(map[string]bool),
		wildcardRules: make([]wildcardPattern, 0),
	}

	for _, pattern := range patterns {
		// Normalize pattern (lowercase, trim spaces)
		pattern = strings.ToLower(strings.TrimSpace(pattern))
		if pattern == "" {
			continue
		}

		// Ensure pattern starts with /
		if !strings.HasPrefix(pattern, "/") {
			pattern = "/" + pattern
		}

		// Remove trailing slash except for root path
		if pattern != "/" {
			pattern = strings.TrimSuffix(pattern, "/")
		}

		// Check if pattern contains wildcards
		if strings.Contains(pattern, "*") {
			// Compile wildcard pattern
			wp := compileWildcardPattern(pattern)
			compiled.wildcardRules = append(compiled.wildcardRules, wp)
		} else {
			// Exact match
			compiled.exactMatches[pattern] = true
		}
	}

	return compiled
}

// compileWildcardPattern converts a wildcard pattern into a structured format
func compileWildcardPattern(pattern string) wildcardPattern {
	// Remove trailing slashes for consistent matching (except root path)
	if pattern != "/" {
		pattern = strings.TrimSuffix(pattern, "/")
	}

	segments := strings.Split(pattern, "/")
	hasWildcard := make([]bool, len(segments))

	for i, seg := range segments {
		if seg == "*" {
			hasWildcard[i] = true
		}
	}

	return wildcardPattern{
		original:    pattern,
		segments:    segments,
		hasWildcard: hasWildcard,
	}
}

// isRouteExcluded checks if a route matches any exclusion pattern
// This is called during request processing
func isRouteExcluded(routePath string, compiled *compiledRoutePatterns) bool {
	if compiled == nil {
		return false
	}

	// Normalize route path (lowercase for case-insensitive matching)
	normalizedPath := strings.ToLower(strings.TrimSpace(routePath))
	// Remove trailing slash except for root path
	if normalizedPath != "/" {
		normalizedPath = strings.TrimSuffix(normalizedPath, "/")
	}

	// First, check exact matches (O(1) lookup)
	if compiled.exactMatches[normalizedPath] {
		return true
	}

	// Then check wildcard patterns (O(n) sequential)
	for _, wp := range compiled.wildcardRules {
		if matchWildcardPattern(normalizedPath, wp) {
			return true
		}
	}

	return false
}

// matchWildcardPattern checks if a route matches a wildcard pattern
func matchWildcardPattern(routePath string, pattern wildcardPattern) bool {
	routeSegments := strings.Split(routePath, "/")

	// If pattern has fewer segments than route (unless last is wildcard), no match
	if len(pattern.segments) > len(routeSegments) {
		return false
	}

	// Match segment by segment
	for i, patternSeg := range pattern.segments {
		if i >= len(routeSegments) {
			return false
		}

		// Wildcard matches any segment
		if pattern.hasWildcard[i] {
			continue
		}

		// Exact segment match required
		if patternSeg != routeSegments[i] {
			return false
		}
	}

	// If pattern ends with wildcard, it matches all nested paths
	if len(pattern.segments) > 0 && pattern.hasWildcard[len(pattern.segments)-1] {
		return true
	}

	// Otherwise, lengths must match exactly
	return len(pattern.segments) == len(routeSegments)
}
