package apperror

import (
	"errors"
	"net/http"
	"time"
)

type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

func NotFound(msg string, cause ...error) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: msg, Err: firstErr(cause)}
}

func Validation(msg string) *AppError {
	return &AppError{Code: http.StatusBadRequest, Message: msg}
}

func Conflict(msg string, cause ...error) *AppError {
	return &AppError{Code: http.StatusConflict, Message: msg, Err: firstErr(cause)}
}

func Unauthorized(msg string) *AppError {
	return &AppError{Code: http.StatusUnauthorized, Message: msg}
}

func Forbidden(msg string) *AppError {
	return &AppError{Code: http.StatusForbidden, Message: msg}
}

func TooManyRequests(msg string) *AppError {
	return &AppError{Code: http.StatusTooManyRequests, Message: msg}
}

type RateLimitedError struct {
	*AppError
	RetryAfter time.Duration
}

func RateLimited(msg string, retryAfter time.Duration) *RateLimitedError {
	return &RateLimitedError{
		AppError:   &AppError{Code: http.StatusTooManyRequests, Message: msg},
		RetryAfter: retryAfter,
	}
}

func Internal(msg string, err error) *AppError {
	return &AppError{Code: http.StatusInternalServerError, Message: msg, Err: err}
}

func IsAppError(err error) (*AppError, bool) {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}

func WrapIfNotApp(msg string, err error) error {
	if err == nil {
		return nil
	}
	var ae *AppError
	if errors.As(err, &ae) {
		return err
	}
	return Internal(msg, err)
}

func firstErr(errs []error) error {
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}
