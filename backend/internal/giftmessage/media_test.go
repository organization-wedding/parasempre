package giftmessage

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

func TestNormalizeMime(t *testing.T) {
	cases := map[string]string{
		"image/jpeg":                 "image/jpeg",
		"  IMAGE/JPEG  ":             "image/jpeg",
		"text/plain; charset=utf-8":  "text/plain",
		"audio/mp4;codecs=mp4a.40.2": "audio/mp4",
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
	mime, peeked, err := sniffMIME(bytes.NewReader(jpeg))
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
	mime, _, err := sniffMIME(strings.NewReader("hello world"))
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

func TestValidateMedia_VideoLimitIs50MB(t *testing.T) {
	spec, err := ValidateMedia("video/mp4", 50*1024*1024)
	if err != nil {
		t.Fatalf("50 MB video should pass the new 50 MB limit, got error: %v", err)
	}
	if spec.kind != MediaKindVideo {
		t.Errorf("expected video kind, got %q", spec.kind)
	}

	_, err = ValidateMedia("video/mp4", 50*1024*1024+1)
	if err == nil {
		t.Fatal("50 MB + 1 byte should be rejected")
	}
	ae, ok := apperror.IsAppError(err)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if !strings.Contains(ae.Message, "excede o limite") {
		t.Errorf("expected 'excede o limite' in error message, got %q", ae.Message)
	}
}

func mpegFrameSyncBytes() []byte {
	b := make([]byte, 128)
	b[0] = 0xFF
	b[1] = 0xFB
	b[2] = 0x90
	b[3] = 0x00
	return b
}
func oggBytes() []byte {
	b := make([]byte, 64)
	copy(b, []byte("OggS"))
	return b
}

func m4aBytes() []byte {
	b := make([]byte, 64)
	b[0], b[1], b[2], b[3] = 0, 0, 0, 32
	copy(b[4:], []byte("ftyp"))
	copy(b[8:], []byte("M4A "))
	return b
}

func movBytes() []byte {
	b := make([]byte, 64)
	b[0], b[1], b[2], b[3] = 0, 0, 0, 32
	copy(b[4:], []byte("ftyp"))
	copy(b[8:], []byte("qt  "))
	return b
}

func htmlBytes() []byte {
	return []byte("<html><body>Not a real audio file</body></html>")
}
func svgBytes() []byte {
	return []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"></svg>`)
}

func TestResolveMediaMIME(t *testing.T) {
	cases := []struct {
		name        string
		declared    string
		body        []byte
		wantMime    string
		wantKind    string
		wantErrFrag string
	}{
		{
			name:     "AC1 mp3 without ID3 declared audio/mpeg",
			declared: "audio/mpeg",
			body:     mpegFrameSyncBytes(),
			wantMime: "audio/mpeg",
			wantKind: MediaKindAudio,
		},
		{
			name:     "AC2 ogg declared audio/ogg",
			declared: "audio/ogg",
			body:     oggBytes(),
			wantMime: "audio/ogg",
			wantKind: MediaKindAudio,
		},
		{
			name:     "AC3a m4a declared audio/mp4",
			declared: "audio/mp4",
			body:     m4aBytes(),
			wantMime: "audio/mp4",
			wantKind: MediaKindAudio,
		},
		{
			name:     "AC3b m4a declared audio/x-m4a",
			declared: "audio/x-m4a",
			body:     m4aBytes(),
			wantMime: "audio/x-m4a",
			wantKind: MediaKindAudio,
		},
		{
			name:     "AC4 mov declared video/quicktime",
			declared: "video/quicktime",
			body:     movBytes(),
			wantMime: "video/quicktime",
			wantKind: MediaKindVideo,
		},
		{
			name:     "AC5a jpeg declared",
			declared: "image/jpeg",
			body: []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0, 1,
				1, 0, 0, 1, 0, 1, 0, 0},
			wantMime: "image/jpeg",
			wantKind: MediaKindImage,
		},
		{
			name:     "AC5b mp4 declared",
			declared: "video/mp4",
			body: []byte{0, 0, 0, 20, 'f', 't', 'y', 'p', 'i', 's', 'o', 'm', 0, 0, 0, 0,
				'i', 's', 'o', 'm'},
			wantMime: "video/mp4",
			wantKind: MediaKindVideo,
		},
		{
			name:        "AC6 pdf declared rejected",
			declared:    "application/pdf",
			body:        []byte("%PDF-1.4 test"),
			wantErrFrag: "tipo de mídia não suportado",
		},
		{
			name:        "AC7 declared audio/mpeg bytes are html",
			declared:    "audio/mpeg",
			body:        htmlBytes(),
			wantErrFrag: "tipo de mídia não suportado",
		},
		{
			name:        "AC7b declared audio/mpeg bytes are svg",
			declared:    "audio/mpeg",
			body:        svgBytes(),
			wantErrFrag: "tipo de mídia não suportado",
		},
		{
			name:     "AC8a empty declared fallback to sniff jpeg",
			declared: "",
			body: []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0, 1,
				1, 0, 0, 1, 0, 1, 0, 0},
			wantMime: "image/jpeg",
			wantKind: MediaKindImage,
		},
		{
			name:     "AC8b octet-stream declared fallback to sniff jpeg",
			declared: "application/octet-stream",
			body: []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0, 1,
				1, 0, 0, 1, 0, 1, 0, 0},
			wantMime: "image/jpeg",
			wantKind: MediaKindImage,
		},
		{
			name:        "AC8c empty declared sniff unknown rejected",
			declared:    "",
			body:        []byte("this is just random text data"),
			wantErrFrag: "tipo de mídia não suportado",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mime, peeked, err := resolveMediaMIME(c.declared, bytes.NewReader(c.body))

			if c.wantErrFrag != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil (mime=%q)", c.wantErrFrag, mime)
				}
				ae, ok := apperror.IsAppError(err)
				if !ok {
					t.Fatalf("expected AppError, got %T: %v", err, err)
				}
				if ae.Code != http.StatusBadRequest {
					t.Errorf("expected 400, got %d", ae.Code)
				}
				if !strings.Contains(ae.Message, c.wantErrFrag) {
					t.Errorf("expected message containing %q, got %q", c.wantErrFrag, ae.Message)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if mime != c.wantMime {
				t.Errorf("mime: want %q, got %q", c.wantMime, mime)
			}
			spec, ok := allowedMimes[mime]
			if !ok {
				t.Fatalf("resolved mime %q not in allowedMimes", mime)
			}
			if spec.kind != c.wantKind {
				t.Errorf("kind: want %q, got %q", c.wantKind, spec.kind)
			}
			got, readErr := io.ReadAll(peeked)
			if readErr != nil {
				t.Fatalf("read peeked: %v", readErr)
			}
			if !bytes.Equal(got, c.body) {
				t.Errorf("peeked reader did not replay original bytes (len %d vs %d)", len(got), len(c.body))
			}
		})
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
