package giftmessage

import (
	"context"
	"io"
	"time"
)

type Storage interface {
	Upload(ctx context.Context, key, mime string, r io.Reader, size int64) error
	SignURLs(ctx context.Context, keys []string, ttl time.Duration) (map[string]string, error)
	Delete(ctx context.Context, key string) error
}
