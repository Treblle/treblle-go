package treblle

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestCustomEndpoint(t *testing.T) {
	// Save original config
	originalEndpoint := Config.Endpoint
	defer func() {
		Config.Endpoint = originalEndpoint
	}()

	// Test custom endpoint
	Config.Endpoint = "https://custom.endpoint.com"
	url := getTreblleBaseUrl()
	assert.Equal(t, "https://custom.endpoint.com", url)
}

func TestDebugModeEndpoint(t *testing.T) {
	// Save original config
	originalDebug := Config.Debug
	originalEndpoint := Config.Endpoint
	defer func() {
		Config.Debug = originalDebug
		Config.Endpoint = originalEndpoint
	}()

	// Test that debug mode doesn't affect endpoint selection
	Config.Debug = true
	Config.Endpoint = ""
	url := getTreblleBaseUrl()
	
	validEndpoints := []string{
		"https://rocknrolla.treblle.com",
		"https://punisher.treblle.com",
		"https://sicario.treblle.com",
	}
	assert.Contains(t, validEndpoints, url)
}

func TestProductionEndpoints(t *testing.T) {
	// Save original config
	originalDebug := Config.Debug
	originalEndpoint := Config.Endpoint
	defer func() {
		Config.Debug = originalDebug
		Config.Endpoint = originalEndpoint
	}()

	// Test production endpoints
	Config.Endpoint = ""
	Config.Debug = false
	url := getTreblleBaseUrl()
	
	validEndpoints := []string{
		"https://rocknrolla.treblle.com",
		"https://punisher.treblle.com",
		"https://sicario.treblle.com",
	}
	assert.Contains(t, validEndpoints, url)
}
