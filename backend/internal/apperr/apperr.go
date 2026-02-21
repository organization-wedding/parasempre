package apperr

import "net/http"

// AppError represents an application error that carries a safe user-facing
// message and an HTTP status code. The underlying cause (Err) is never sent
// to the client
type AppError struct {
	Code    int    // HTTP status code to return
	Message string // safe, user-facing message
	Err     error  // underlying cause (never exposed to client)
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

// NotFound creates a 404 AppError. Pass the original sentinel error as cause
// so that errors.Is() keeps working up the call stack.
func NotFound(msg string, cause error) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: msg, Err: cause}
}

// Conflict creates a 409 AppError for uniqueness / state conflicts.
func Conflict(msg string, cause error) *AppError {
	return &AppError{Code: http.StatusConflict, Message: msg, Err: cause}
}

// Validation creates a 400 AppError for invalid input from the client.
func Validation(msg string) *AppError {
	return &AppError{Code: http.StatusBadRequest, Message: msg}
}

// Forbidden creates a 403 AppError for authorization failures.
func Forbidden(msg string) *AppError {
	return &AppError{Code: http.StatusForbidden, Message: msg}
}

// Internal creates a 500 AppError. The cause is stored for logging but the
// generic message is always returned to the client.
func Internal(cause error) *AppError {
	return &AppError{Code: http.StatusInternalServerError, Message: "Erro interno do servidor", Err: cause}
}
