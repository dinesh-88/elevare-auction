package port

import (
	"context"
	"io"
)

type ObjectStorage interface {
	Put(ctx context.Context, key string, contentType string, body io.Reader, contentLength int64) (publicURL string, err error)
}
