package treblle

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
)

// ErrorType represents different types of errors (matching Laravel's error types)
type ErrorType string

const (
	UnhandledExceptionError ErrorType = "UNHANDLED_EXCEPTION"
	MarshalError            ErrorType = "MARSHAL_ERROR"
	ValidationError         ErrorType = "VALIDATION_ERROR"
	AuthenticationError     ErrorType = "AUTHENTICATION_ERROR"
	AuthorizationError      ErrorType = "AUTHORIZATION_ERROR"
	NotFoundError           ErrorType = "NOT_FOUND_ERROR"
	RateLimitError          ErrorType = "RATE_LIMIT_ERROR"
	ServerError             ErrorType = "SERVER_ERROR"
)

// ErrorInfo represents detailed error information
type ErrorInfo struct {
	Message string    `json:"message"`
	Type    ErrorType `json:"type"`
	File    string    `json:"file"`
	Line    int       `json:"line"`
	Source  string    `json:"source"`
}

// ErrorProvider manages error collection and processing
type ErrorProvider struct {
	mu     sync.Mutex
	errors []ErrorInfo
}

// NewErrorProvider creates a new error provider
func NewErrorProvider() *ErrorProvider {
	return &ErrorProvider{
		errors: make([]ErrorInfo, 0),
	}
}

// AddError adds an error with full stack trace and file information
func (ep *ErrorProvider) AddError(err error, errType ErrorType, source string) {
	if err == nil {
		return
	}

	ep.mu.Lock()
	defer ep.mu.Unlock()

	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}

	// Clean file path (similar to Laravel's handling)
	file = cleanFilePath(file)

	ep.errors = append(ep.errors, ErrorInfo{
		Message: err.Error(),
		Type:    errType,
		File:    file,
		Line:    line,
		Source:  source,
	})
}

// AddCustomError adds an error with custom message
func (ep *ErrorProvider) AddCustomError(message string, errType ErrorType, source string) {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}

	file = cleanFilePath(file)

	ep.errors = append(ep.errors, ErrorInfo{
		Message: message,
		Type:    errType,
		File:    file,
		Line:    line,
		Source:  source,
	})
}

// GetErrors returns all collected errors
func (ep *ErrorProvider) GetErrors() []ErrorInfo {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	// Return a copy to prevent external modifications
	result := make([]ErrorInfo, len(ep.errors))
	copy(result, ep.errors)
	return result
}

// Clear removes all collected errors
func (ep *ErrorProvider) Clear() {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	ep.errors = ep.errors[:0]
}

// cleanFilePath cleans the file path similar to Laravel's handling
func cleanFilePath(path string) string {
	// Get the last two path components
	parts := strings.Split(path, "/")
	if len(parts) > 2 {
		return fmt.Sprintf(".../%s/%s", parts[len(parts)-2], parts[len(parts)-1])
	}
	return path
}
