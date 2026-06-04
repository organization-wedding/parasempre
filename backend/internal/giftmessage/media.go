package giftmessage

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

const (
	maxImageBytes int64 = 5 * 1024 * 1024
	maxAudioBytes int64 = 5 * 1024 * 1024
	maxVideoBytes int64 = 50 * 1024 * 1024 // AC9: raised from 25 MB to 50 MB

	sniffSize = 512
)

var dangerousSniffTypes = map[string]bool{
	"text/html":       true,
	"image/svg+xml":   true,
	"text/xml":        true,
	"application/xml": true,
	"text/plain":      true,
}

func resolveMediaMIME(declared string, r io.Reader) (mime string, peeked io.Reader, err error) {
	candidate := normalizeMime(declared)
	buf := make([]byte, sniffSize)
	n, readErr := io.ReadFull(r, buf)
	if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
		return "", nil, readErr
	}
	buf = buf[:n]
	sniffed := normalizeMime(http.DetectContentType(buf))
	peeked = io.MultiReader(bytes.NewReader(buf), r)

	if dangerousSniffTypes[sniffed] {
		return "", nil, apperror.Validation(fmt.Sprintf("tipo de mídia não suportado: %s", candidate))
	}
	if _, ok := allowedMimes[candidate]; !ok {
		if _, ok := allowedMimes[sniffed]; !ok {
			label := candidate
			if label == "" || label == "application/octet-stream" {
				label = sniffed
			}
			return "", nil, apperror.Validation(fmt.Sprintf("tipo de mídia não suportado: %s", label))
		}
		candidate = sniffed
	}

	return candidate, peeked, nil
}

type mediaSpec struct {
	kind     string
	maxBytes int64
	ext      string
}

var allowedMimes = map[string]mediaSpec{
	"image/jpeg":      {MediaKindImage, maxImageBytes, ".jpg"},
	"image/png":       {MediaKindImage, maxImageBytes, ".png"},
	"image/webp":      {MediaKindImage, maxImageBytes, ".webp"},
	"audio/mpeg":      {MediaKindAudio, maxAudioBytes, ".mp3"},
	"audio/mp4":       {MediaKindAudio, maxAudioBytes, ".m4a"},
	"audio/x-m4a":     {MediaKindAudio, maxAudioBytes, ".m4a"},
	"audio/ogg":       {MediaKindAudio, maxAudioBytes, ".ogg"},
	"video/mp4":       {MediaKindVideo, maxVideoBytes, ".mp4"},
	"video/webm":      {MediaKindVideo, maxVideoBytes, ".webm"},
	"video/quicktime": {MediaKindVideo, maxVideoBytes, ".mov"},
}

func sniffMIME(r io.Reader) (mime string, peeked io.Reader, err error) {
	buf := make([]byte, sniffSize)
	n, err := io.ReadFull(r, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return "", nil, err
	}
	buf = buf[:n]
	mime = normalizeMime(http.DetectContentType(buf))
	return mime, io.MultiReader(bytes.NewReader(buf), r), nil
}

func ValidateMedia(mime string, size int64) (mediaSpec, error) {
	mime = normalizeMime(mime)
	spec, ok := allowedMimes[mime]
	if !ok {
		return mediaSpec{}, apperror.Validation(fmt.Sprintf("tipo de mídia não suportado: %s", mime))
	}
	if size <= 0 {
		return mediaSpec{}, apperror.Validation("arquivo de mídia vazio")
	}
	if size > spec.maxBytes {
		return mediaSpec{}, apperror.Validation(fmt.Sprintf(
			"arquivo de mídia excede o limite de %d MB para %s",
			spec.maxBytes/(1024*1024), spec.kind,
		))
	}
	return spec, nil
}

func normalizeMime(mime string) string {
	if i := strings.Index(mime, ";"); i >= 0 {
		mime = mime[:i]
	}
	return strings.ToLower(strings.TrimSpace(mime))
}
