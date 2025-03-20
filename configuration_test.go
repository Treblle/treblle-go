package treblle

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSDKVersioning(t *testing.T) {
	// Initialize with default configuration
	Configure(Configuration{
		APIKey:    "test-api-key",
		ProjectID: "test-project-id",
	})

	// Ensure default version is correct
	assert.Equal(t, "go", Config.SDKName)
	assert.Equal(t, "2.0.0", Config.SDKVersion)

	// Test GetSDKInfo function
	info := GetSDKInfo()
	assert.Equal(t, "go", info["SDK Name"])
	assert.Equal(t, 2.0, info["SDK Version"])

	// Set environment variables and reconfigure
	os.Setenv("TREBLLE_SDK_VERSION", "2.1.0")
	Configure(Configuration{})

	// Check if version updates
	assert.Equal(t, "2.1.0", Config.SDKVersion)

	// Clean up env
	os.Unsetenv("TREBLLE_SDK_VERSION")
}
