package giftmessage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewSupabaseStorage_StripsRestV1Suffix(t *testing.T) {
	cases := []struct{ in, want string }{
		{"https://x.supabase.co", "https://x.supabase.co"},
		{"https://x.supabase.co/", "https://x.supabase.co"},
		{"https://x.supabase.co/rest/v1", "https://x.supabase.co"},
		{"https://x.supabase.co/rest/v1/", "https://x.supabase.co"},
	}
	for _, c := range cases {
		if got := NewSupabaseStorage(c.in, "b", "k").baseURL; got != c.want {
			t.Errorf("NewSupabaseStorage(%q).baseURL = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSupabaseStorage_BucketExists_Returns_Nil_On_200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/storage/v1/bucket/") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := NewSupabaseStorage(srv.URL, "gift-messages", "test-key")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := s.BucketExists(ctx); err != nil {
		t.Fatalf("BucketExists should return nil on 200, got: %v", err)
	}
}

func TestSupabaseStorage_BucketExists_Returns_Error_On_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"Bucket not found"}`))
	}))
	defer srv.Close()

	s := NewSupabaseStorage(srv.URL, "gift-messages", "test-key")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := s.BucketExists(ctx)
	if err == nil {
		t.Fatal("BucketExists should return error on 404")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error should mention the HTTP status, got: %v", err)
	}
}
