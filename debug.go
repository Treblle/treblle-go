package treblle

import (
	"fmt"
	"strconv"
)

// DebugCommand prints Treblle SDK configuration information
func DebugCommand() {
	fmt.Println("=== Treblle Go SDK Debug Information ===")

	// Initialize configuration from environment variables if not already set
	if Config.APIKey == "" {
		// Read configuration from environment variables
		apiKey := getEnvOrDefault("TREBLLE_API_KEY", "")
		projectID := getEnvOrDefault("TREBLLE_PROJECT_ID", "")
		endpoint := getEnvOrDefault("TREBLLE_ENDPOINT", "")
		ignoredEnvs := getEnvAsSlice("TREBLLE_IGNORED_ENVIRONMENTS", []string{"local", "development"})
		
		// Update Config with environment values
		if apiKey != "" {
			Config.APIKey = apiKey
		}
		if projectID != "" {
			Config.ProjectID = projectID
		}
		if endpoint != "" {
			Config.Endpoint = endpoint
		}
		if len(ignoredEnvs) > 0 {
			Config.IgnoredEnvironments = ignoredEnvs
		}
		
		// Set SDK info from constants
		Config.SDKName = SDKName
		Config.SDKVersion = SDKVersion
	}

	// Display basic SDK configuration
	fmt.Println("SDK Version:", strconv.FormatFloat(Config.SDKVersion, 'f', 1, 64))
	fmt.Println("Project ID:", maskString(Config.ProjectID))
	fmt.Println("API Key:", maskString(Config.APIKey))
	fmt.Println("Configured Treblle URL:", getConfiguredEndpoint())
	fmt.Println("Ignored Environments:", Config.IgnoredEnvironments)
}

// getConfiguredEndpoint returns the configured endpoint or the default if not set
func getConfiguredEndpoint() string {
	if Config.Endpoint != "" {
		return Config.Endpoint
	}
	return "Default Treblle API endpoints (load balanced)"
}

// Utility function to mask sensitive values
func maskString(value string) string {
	if value == "" {
		return "Not Set"
	}
	if len(value) <= 4 {
		return "****"
	}
	return "****" + value[len(value)-4:] // Show last 4 characters
}
