package httputil

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

func DecodeJSON(r *http.Request, dest any) error {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		return apperror.Validation("invalid JSON")
	}
	return nil
}

func PathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return 0, apperror.Validation("invalid ID")
	}
	return id, nil
}
