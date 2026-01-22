package files

import (
	"context"
	"file-storage/internal/filedata"
)

// Storage defines persistence operations required by the business layer.
type Storage interface {
	Upsert(ctx context.Context, fd *filedata.FileData) (string, error)
	Info(ctx context.Context, ID string) (*filedata.FileInfo, error)
	Content(ctx context.Context, ID string) (*filedata.ContentData, error)
	Delete(ctx context.Context, ID string) error
}
