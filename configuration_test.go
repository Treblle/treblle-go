package treblle

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSDKVersioning(t *testing.T) {
	// Initialize with default configuration
	Configure(Configuration{
		SDK_TOKEN: "test-sdk-token",
		API_KEY:   "test-api-key",
	})

	// Ensure default version is correct
	assert.Equal(t, "go", Config.SDKName)
	assert.Equal(t, 2.0, Config.SDKVersion)

	// Test GetSDKInfo function
	info := GetSDKInfo()
	assert.Equal(t, "go", info["SDK Name"])
	assert.Equal(t, "2.00", info["SDK Version"])

	// Set environment variables and reconfigure
	os.Setenv("TREBLLE_SDK_VERSION", "2.1")
	Configure(Configuration{})

	// Check if version updates
	assert.Equal(t, 2.1, Config.SDKVersion)

	// Clean up env
	os.Unsetenv("TREBLLE_SDK_VERSION")
}
