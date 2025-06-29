package response

import "fmt"

// APIError represents a custom error with status code and message
type APIError struct {
	StatusCode int
	Message    string
	Err        interface{} // can be string or error
}

// Error implements the error interface
func (e *APIError) Error() string {
	switch v := e.Err.(type) {
	case string:
		return v
	case error:
		return v.Error()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// NewAPIError creates a new APIError instance
func NewAPIError(status int, message string, err interface{}) *APIError {
	return &APIError{
		StatusCode: status,
		Message:    message,
		Err:        err,
	}
}
