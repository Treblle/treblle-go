package treblle

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResponseSizeLimit(t *testing.T) {
	// Create a new error provider
	errorProvider := NewErrorProvider()

	// Create a response recorder
	w := httptest.NewRecorder()
	
	// Generate a response body that exceeds 2MB
	largeBody := strings.Repeat("a", maxResponseSize+1)
	w.WriteString(largeBody)
	
	// Get the response info
	startTime := time.Now().Add(-100 * time.Millisecond) // Simulate some processing time
	responseInfo := getResponseInfo(w, startTime, errorProvider)
	
	// Verify the response body was replaced with an empty JSON object
	assert.Equal(t, json.RawMessage("{}"), responseInfo.Body)
	
	// Verify the size was set to 0
	assert.Equal(t, 0, responseInfo.Size)
	
	// Verify an error was added
	errors := errorProvider.GetErrors()
	assert.Len(t, errors, 1)
	assert.Equal(t, "JSON response size is over 2MB", errors[0].Message)
	assert.Equal(t, ServerError, errors[0].Type)
	assert.Equal(t, "response_size_limit", errors[0].Source)
}

func TestResponseSizeLimitNotExceeded(t *testing.T) {
	// Create a new error provider
	errorProvider := NewErrorProvider()

	// Create a response recorder
	w := httptest.NewRecorder()
	
	// Generate a valid JSON response body that does not exceed 2MB
	smallBody := `{"test":"data"}`
	w.WriteString(smallBody)
	
	// Get the response info
	startTime := time.Now().Add(-100 * time.Millisecond) // Simulate some processing time
	responseInfo := getResponseInfo(w, startTime, errorProvider)
	
	// Verify the response body was not replaced with an empty JSON object
	assert.NotEqual(t, json.RawMessage("{}"), responseInfo.Body)
	
	// Verify the size was set correctly
	assert.Equal(t, len(smallBody), responseInfo.Size)
	
	// Verify no "response size limit" error was added
	errors := errorProvider.GetErrors()
	for _, err := range errors {
		assert.NotEqual(t, "JSON response size is over 2MB", err.Message)
	}
}
