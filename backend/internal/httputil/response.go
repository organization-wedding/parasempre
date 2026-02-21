package httputil

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/ferjunior7/parasempre/backend/internal/apperr"
)

// WriteJSON encodes data as JSON and writes it with the given HTTP status.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// WriteError translates err into a JSON HTTP error response.
//
//   - If err is an *apperr.AppError its Code and user-facing Message are used
//     directly — no additional logging is performed here because the service
//     layer already logged the context-rich details.
//   - Any other error is treated as unexpected: a 500 is returned and the raw
//     error is logged so it can be investigated.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *apperr.AppError
	if errors.As(err, &appErr) {
		WriteJSON(w, appErr.Code, map[string]string{"error": appErr.Message})
		return
	}

	// Truly unexpected error — log it so it can be investigated.
	slog.Error("unhandled error", "path", r.URL.Path, "method", r.Method, "err", err)
	WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Erro interno do servidor"})
}
