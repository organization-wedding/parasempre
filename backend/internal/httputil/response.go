package httputil

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

func HandleError(w http.ResponseWriter, err error) {
	if ae, ok := apperror.IsAppError(err); ok {
		WriteError(w, ae.Code, ae.Message)
		if ae.Code >= 500 {
			slog.Error("internal error", "error", ae.Err)
		}
		return
	}
	slog.Error("unhandled error", "error", err)
	WriteError(w, http.StatusInternalServerError, "internal server error")
}
