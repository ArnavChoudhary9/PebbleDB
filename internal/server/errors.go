package server

import "net/http"

// HTTPError represents an HTTP error with status code and message
type HTTPError struct {
	Code    int
	Message string
}

// Error implements the error interface
func (e HTTPError) Error() string {
	return e.Message
}

// NewHTTPError creates a new HTTP error
func NewHTTPError(code int, message string) HTTPError {
	return HTTPError{Code: code, Message: message}
}

// Common HTTP error constructors

// BadRequest creates a 400 Bad Request error
func BadRequest(message string) HTTPError {
	return NewHTTPError(http.StatusBadRequest, message)
}

// NotFound creates a 404 Not Found error
func NotFound(message string) HTTPError {
	return NewHTTPError(http.StatusNotFound, message)
}

// InternalServerError creates a 500 Internal Server Error
func InternalServerError(message string) HTTPError {
	return NewHTTPError(http.StatusInternalServerError, message)
}

// Unauthorized creates a 401 Unauthorized error
func Unauthorized(message string) HTTPError {
	return NewHTTPError(http.StatusUnauthorized, message)
}

// Forbidden creates a 403 Forbidden error
func Forbidden(message string) HTTPError {
	return NewHTTPError(http.StatusForbidden, message)
}
