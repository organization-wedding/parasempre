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
	maxVideoBytes int64 = 25 * 1024 * 1024

	sniffSize = 512
)

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

func SniffMIME(r io.Reader) (mime string, peeked io.Reader, err error) {
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
