package giftmessage

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

func TestNormalizeMime(t *testing.T) {
	cases := map[string]string{
		"image/jpeg":                     "image/jpeg",
		"  IMAGE/JPEG  ":                 "image/jpeg",
		"text/plain; charset=utf-8":      "text/plain",
		"audio/mp4;codecs=mp4a.40.2":     "audio/mp4",
	}
	for in, want := range cases {
		if got := normalizeMime(in); got != want {
			t.Errorf("normalizeMime(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSniffMIME_DetectsJPEG(t *testing.T) {
	// JPEG SOI marker followed by APP0 / JFIF
	jpeg := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0, 1, 1, 0, 0, 1, 0, 1, 0, 0}
	mime, peeked, err := SniffMIME(bytes.NewReader(jpeg))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mime != "image/jpeg" {
		t.Errorf("expected image/jpeg, got %q", mime)
	}
	out, _ := io.ReadAll(peeked)
	if !bytes.Equal(out, jpeg) {
		t.Errorf("peeked reader did not reinject the original bytes")
	}
}

func TestSniffMIME_TextDetectedAsPlain(t *testing.T) {
	mime, _, err := SniffMIME(strings.NewReader("hello world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mime != "text/plain" {
		t.Errorf("expected text/plain, got %q", mime)
	}
}

func TestValidateMedia_Whitelist(t *testing.T) {
	cases := []struct {
		name      string
		mime      string
		size      int64
		wantKind  string
		wantError string
	}{
		{"image jpeg ok", "image/jpeg", 1024, MediaKindImage, ""},
		{"image png ok", "image/png", 1024, MediaKindImage, ""},
		{"image webp ok", "image/webp", 1024, MediaKindImage, ""},
		{"audio mpeg ok", "audio/mpeg", 1024, MediaKindAudio, ""},
		{"video mp4 ok", "video/mp4", 1024, MediaKindVideo, ""},
		{"video quicktime ok", "video/quicktime", 1024, MediaKindVideo, ""},
		{"unsupported pdf", "application/pdf", 1024, "", "tipo de mídia não suportado"},
		{"unsupported text", "text/plain", 1024, "", "tipo de mídia não suportado"},
		{"empty file", "image/jpeg", 0, "", "vazio"},
		{"image too big", "image/jpeg", maxImageBytes + 1, "", "excede o limite"},
		{"audio too big", "audio/mpeg", maxAudioBytes + 1, "", "excede o limite"},
		{"video too big", "video/mp4", maxVideoBytes + 1, "", "excede o limite"},
		{"video at limit ok", "video/mp4", maxVideoBytes, MediaKindVideo, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			spec, err := ValidateMedia(c.mime, c.size)
			if c.wantError != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", c.wantError)
				}
				ae, ok := apperror.IsAppError(err)
				if !ok {
					t.Fatalf("expected AppError, got %T", err)
				}
				if !strings.Contains(ae.Message, c.wantError) {
					t.Errorf("expected error containing %q, got %q", c.wantError, ae.Message)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if spec.kind != c.wantKind {
				t.Errorf("expected kind %q, got %q", c.wantKind, spec.kind)
			}
		})
	}
}

func TestValidateMedia_NormalizesMime(t *testing.T) {
	spec, err := ValidateMedia("IMAGE/JPEG; charset=binary", 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.kind != MediaKindImage {
		t.Errorf("expected image kind from normalized MIME, got %q", spec.kind)
	}
}

func TestBuildObjectKey_HasTransactionAndExtension(t *testing.T) {
	key, err := buildObjectKey(42, ".jpg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(key, "messages/42/") {
		t.Errorf("key should start with messages/42/, got %q", key)
	}
	if !strings.HasSuffix(key, ".jpg") {
		t.Errorf("key should end with .jpg, got %q", key)
	}
}
