package files

import (
	"context"
	"io"
)

type Storage interface {
	Upsert(ctx context.Context, fd *FileData) (string, error)
	Info(ctx context.Context, ID string) (*FileInfo, error)
	Content(ctx context.Context, ID string) (io.ReadCloser, error)
	Delete(ctx context.Context, ID string) error
}
