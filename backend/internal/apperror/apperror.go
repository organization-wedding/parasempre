package apperror

import (
	"errors"
	"net/http"
)

// AppError represents an application error that carries a safe user-facing
// message and an HTTP status code. The underlying cause (Err) is never sent
// to the client.
type AppError struct {
	Code    int    // HTTP status code to return
	Message string // safe, user-facing message
	Err     error  // underlying cause (never exposed to client)
}

// Error returns only the user-safe message — never leaks Err.
func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

// NotFound creates a 404 AppError. Optionally pass the original sentinel error
// as cause so that errors.Is() keeps working up the call stack.
func NotFound(msg string, cause ...error) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: msg, Err: firstErr(cause)}
}

// Validation creates a 400 AppError for invalid input from the client.
func Validation(msg string) *AppError {
	return &AppError{Code: http.StatusBadRequest, Message: msg}
}

// Conflict creates a 409 AppError for uniqueness / state conflicts.
func Conflict(msg string, cause ...error) *AppError {
	return &AppError{Code: http.StatusConflict, Message: msg, Err: firstErr(cause)}
}

// Unauthorized creates a 401 AppError for authentication failures.
func Unauthorized(msg string) *AppError {
	return &AppError{Code: http.StatusUnauthorized, Message: msg}
}

// Forbidden creates a 403 AppError for authorization failures.
func Forbidden(msg string) *AppError {
	return &AppError{Code: http.StatusForbidden, Message: msg}
}

// Internal creates a 500 AppError. The cause is stored for logging but the
// user-facing message is always returned to the client.
func Internal(msg string, err error) *AppError {
	return &AppError{Code: http.StatusInternalServerError, Message: msg, Err: err}
}

// IsAppError extracts an *AppError from an error chain, if present.
func IsAppError(err error) (*AppError, bool) {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}

func firstErr(errs []error) error {
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}
