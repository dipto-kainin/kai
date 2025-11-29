package utils

import (
	"fmt"
	"net/http"
)

// HTTPError represents an HTTP error with status code and message
type HTTPError struct {
	Code    int
	Message string
	Err     error
}

func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("HTTP %d: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("HTTP %d: %s", e.Code, e.Message)
}

// NewHTTPError creates a new HTTP error
func NewHTTPError(code int, message string) *HTTPError {
	return &HTTPError{Code: code, Message: message}
}

// WrapHTTPError wraps an existing error with HTTP context
func WrapHTTPError(code int, message string, err error) *HTTPError {
	return &HTTPError{Code: code, Message: message, Err: err}
}

// Common HTTP errors
var (
	ErrBadRequest          = NewHTTPError(http.StatusBadRequest, "Bad Request")
	ErrUnauthorized        = NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	ErrForbidden           = NewHTTPError(http.StatusForbidden, "Forbidden")
	ErrNotFound            = NewHTTPError(http.StatusNotFound, "Not Found")
	ErrMethodNotAllowed    = NewHTTPError(http.StatusMethodNotAllowed, "Method Not Allowed")
	ErrConflict            = NewHTTPError(http.StatusConflict, "Conflict")
	ErrInternalServerError = NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	ErrServiceUnavailable  = NewHTTPError(http.StatusServiceUnavailable, "Service Unavailable")
)

// ValidationError represents validation errors with field-level details
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []*ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("validation failed: %s", v[0].Error())
}

// ToMap converts validation errors to a map for JSON responses
func (v ValidationErrors) ToMap() map[string]string {
	result := make(map[string]string, len(v))
	for _, err := range v {
		result[err.Field] = err.Message
	}
	return result
}
