package treblle

import (
	"os"
	"testing"
)

func TestIgnoredEnvironments(t *testing.T) {
	// Save original environment variables to restore later
	originalGoEnv := os.Getenv("GO_ENV")
	originalTreblleIgnoredEnv := os.Getenv("TREBLLE_IGNORED_ENV")
	
	// Clean up after test
	defer func() {
		// Restore original environment variables
		if originalGoEnv != "" {
			os.Setenv("GO_ENV", originalGoEnv)
		} else {
			os.Unsetenv("GO_ENV")
		}
		
		if originalTreblleIgnoredEnv != "" {
			os.Setenv("TREBLLE_IGNORED_ENV", originalTreblleIgnoredEnv)
		} else {
			os.Unsetenv("TREBLLE_IGNORED_ENV")
		}
	}()
	
	// Test cases
	testCases := []struct {
		name            string
		goEnv           string
		ignoredEnvs     []string
		envIgnoredEnvs  string
		expectedIgnored bool
	}{
		{
			name:            "Default ignored environment",
			goEnv:           "dev",
			ignoredEnvs:     nil,
			envIgnoredEnvs:  "",
			expectedIgnored: true,
		},
		{
			name:            "Production environment not ignored",
			goEnv:           "production",
			ignoredEnvs:     nil,
			envIgnoredEnvs:  "",
			expectedIgnored: false,
		},
		{
			name:            "Custom ignored environment from config",
			goEnv:           "staging",
			ignoredEnvs:     []string{"staging", "qa"},
			envIgnoredEnvs:  "",
			expectedIgnored: true,
		},
		{
			name:            "Custom ignored environment from env variable",
			goEnv:           "local",
			ignoredEnvs:     nil,
			envIgnoredEnvs:  "local,ci",
			expectedIgnored: true,
		},
		{
			name:            "No environment set",
			goEnv:           "",
			ignoredEnvs:     nil,
			envIgnoredEnvs:  "",
			expectedIgnored: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variables for this test case
			if tc.goEnv != "" {
				os.Setenv("GO_ENV", tc.goEnv)
			} else {
				os.Unsetenv("GO_ENV")
			}
			
			if tc.envIgnoredEnvs != "" {
				os.Setenv("TREBLLE_IGNORED_ENV", tc.envIgnoredEnvs)
			} else {
				os.Unsetenv("TREBLLE_IGNORED_ENV")
			}
			
			// Configure Treblle with test case settings
			config := Configuration{
				APIKey:              "test-api-key",
				ProjectID:           "test-project-id",
				IgnoredEnvironments: tc.ignoredEnvs,
			}
			Configure(config)
			
			// Check if environment is ignored
			result := IsEnvironmentIgnored()
			
			if result != tc.expectedIgnored {
				t.Errorf("Expected IsEnvironmentIgnored() to return %v, got %v", 
					tc.expectedIgnored, result)
			}
		})
	}
}
