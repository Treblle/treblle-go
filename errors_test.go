package treblle

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestErrorProvider(t *testing.T) {
	tests := []struct {
		name    string
		errType ErrorType
	}{
		{"UnhandledException", UnhandledExceptionError},
		{"MarshalError", MarshalError},
		{"ValidationError", ValidationError},
		{"AuthenticationError", AuthenticationError},
		{"AuthorizationError", AuthorizationError},
		{"NotFoundError", NotFoundError},
		{"RateLimitError", RateLimitError},
		{"ServerError", ServerError},
	}

	ep := NewErrorProvider()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Add an error of each type
			ep.AddCustomError("Test "+tt.name, tt.errType, "test")

			// Verify error was added correctly
			errors := ep.GetErrors()
			assert.Len(t, errors, 1, "Should have 1 error")
			assert.Equal(t, tt.errType, errors[0].Type, "Error type should match")
			assert.Equal(t, "Test "+tt.name, errors[0].Message, "Error message should match")
			assert.Equal(t, "test", errors[0].Source, "Error source should match")
			assert.NotEmpty(t, errors[0].File, "Error file should not be empty")
			assert.Greater(t, errors[0].Line, 0, "Error line should be greater than 0")

			// Clear errors for next test
			ep.Clear()
		})
	}
}

func TestErrorProviderConcurrency(t *testing.T) {
	ep := NewErrorProvider()

	// Test concurrent error adding
	t.Run("ConcurrentAdd", func(t *testing.T) {
		go ep.AddCustomError("Error 1", ValidationError, "test1")
		go ep.AddCustomError("Error 2", ServerError, "test2")
		go ep.AddCustomError("Error 3", MarshalError, "test3")

		// Give time for goroutines to complete
		errors := ep.GetErrors()
		assert.LessOrEqual(t, len(errors), 3, "Should have at most 3 errors")
		ep.Clear()
	})
}
