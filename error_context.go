package treblle

import (
	"fmt"
	"runtime"
	"strings"
)

// ErrorContext provides additional context about where an error occurred
type ErrorContext struct {
	Function   string `json:"function,omitempty"`
	Package    string `json:"package,omitempty"`
	Component  string `json:"component,omitempty"`
	StackTrace string `json:"stack_trace,omitempty"`
}

// getErrorContext returns contextual information about where an error occurred
func getErrorContext(skip int) ErrorContext {
	var context ErrorContext
	
	// Get caller information
	if pc, file, _, ok := runtime.Caller(skip + 1); ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			// Get full function name
			fullName := fn.Name()
			
			// Split into package and function
			if lastDot := strings.LastIndex(fullName, "."); lastDot > 0 {
				context.Package = fullName[:lastDot]
				context.Function = fullName[lastDot+1:]
			} else {
				context.Function = fullName
			}
		}
		
		// Determine component from file path
		parts := strings.Split(file, "/")
		if len(parts) > 0 {
			context.Component = strings.TrimSuffix(parts[len(parts)-1], ".go")
		}
		
		// Get stack trace
		buffer := make([]byte, 4096)
		n := runtime.Stack(buffer, false)
		context.StackTrace = string(buffer[:n])
	}
	
	return context
}

// ErrorWithContext combines an error with its context
type ErrorWithContext struct {
	Err     error
	Context ErrorContext
}

// Error implements the error interface
func (e *ErrorWithContext) Error() string {
	return fmt.Sprintf("%v [in %s.%s]", e.Err, e.Context.Package, e.Context.Function)
}
