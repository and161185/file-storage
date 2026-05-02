package handlers

import (
	"context"
	"file-storage/internal/filedata"
)

// Service defines the business operations required by HTTP handlers to upload files, read content and metadata, and delete files.
type Service interface {
	Update(ctx context.Context, uc *filedata.UploadCommand) (string, error)
	Content(ctx context.Context, cc *filedata.ContentCommand) ([]byte, error)
	Info(ctx context.Context, ID string) (*filedata.FileInfo, error)
	Delete(ctx context.Context, ID string) error
}
