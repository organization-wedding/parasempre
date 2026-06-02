package httputil

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	var rle *apperror.RateLimitedError
	if errors.As(err, &rle) {
		WriteJSON(w, rle.Code, map[string]any{
			"error":               rle.Message,
			"retry_after_seconds": int(rle.RetryAfter.Seconds()),
		})
		return
	}

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

	slog.Error("unhandled error", "path", r.URL.Path, "method", r.Method, "err", err)
	WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}

func WriteErrorMsg(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}
