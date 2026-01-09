package files

import (
	"context"
)

type Storage interface {
	Upsert(ctx context.Context, fd *FileData) (string, error)
	Info(ctx context.Context, ID string) (*FileInfo, error)
	Content(ctx context.Context, ID string) (*ContentData, error)
	Delete(ctx context.Context, ID string) error
}
