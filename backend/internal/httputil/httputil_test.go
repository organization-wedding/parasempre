package httputil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	WriteJSON(w, http.StatusOK, map[string]string{"key": "value"})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %q", ct)
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if body["key"] != "value" {
		t.Fatalf("expected value, got %q", body["key"])
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	WriteError(w, http.StatusBadRequest, "bad input")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["error"] != "bad input" {
		t.Fatalf("expected %q, got %q", "bad input", body["error"])
	}
}

func TestHandleErrorAppError(t *testing.T) {
	w := httptest.NewRecorder()
	HandleError(w, apperror.NotFound("guest not found"))

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}

	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["error"] != "guest not found" {
		t.Fatalf("expected %q, got %q", "guest not found", body["error"])
	}
}

func TestHandleErrorGeneric(t *testing.T) {
	w := httptest.NewRecorder()
	HandleError(w, http.ErrServerClosed)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestDecodeJSON(t *testing.T) {
	type input struct {
		Name string `json:"name"`
	}

	body := bytes.NewBufferString(`{"name":"test"}`)
	r := httptest.NewRequest(http.MethodPost, "/", body)
	var dest input
	if err := DecodeJSON(r, &dest); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dest.Name != "test" {
		t.Fatalf("expected %q, got %q", "test", dest.Name)
	}
}

func TestDecodeJSONInvalid(t *testing.T) {
	body := bytes.NewBufferString("{bad")
	r := httptest.NewRequest(http.MethodPost, "/", body)
	var dest struct{}
	err := DecodeJSON(r, &dest)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	ae, ok := apperror.IsAppError(err)
	if !ok {
		t.Fatal("expected AppError")
	}
	if ae.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", ae.Code)
	}
}

func TestPathID(t *testing.T) {
	mux := http.NewServeMux()
	var gotID int64
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := PathID(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		gotID = id
	})

	req := httptest.NewRequest(http.MethodGet, "/items/42", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if gotID != 42 {
		t.Fatalf("expected 42, got %d", gotID)
	}
}

func TestPathIDInvalid(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		_, err := PathID(r)
		if err == nil {
			t.Fatal("expected error for invalid ID")
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/items/abc", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
}
