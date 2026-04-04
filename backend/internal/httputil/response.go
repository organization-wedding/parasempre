package httputil

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

// WriteJSON encodes data as JSON and writes it with the given HTTP status.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// WriteError translates err into a JSON HTTP error response.
//
//   - If err is an *apperror.AppError its Code and user-facing Message are used
//     directly. Internal errors (5xx) are logged with request context.
//   - Any other error is treated as unexpected: a 500 is returned and the raw
//     error is logged so it can be investigated.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		if appErr.Code >= 500 && appErr.Err != nil {
			slog.Error("internal app error",
				"path", r.URL.Path, "method", r.Method,
				"msg", appErr.Message, "cause", appErr.Err)
		}
		WriteJSON(w, appErr.Code, map[string]string{"error": appErr.Message})
		return
	}

	// Truly unexpected error — log it so it can be investigated.
	slog.Error("unhandled error", "path", r.URL.Path, "method", r.Method, "err", err)
	WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}

// WriteErrorMsg is a convenience for middleware that needs to write a known
// status+message without an error object (e.g., auth middleware, recovery).
func WriteErrorMsg(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}
