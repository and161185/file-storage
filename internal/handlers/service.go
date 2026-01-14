package handlers

import (
	"context"
	"file-storage/internal/files"
)

type Service interface {
	Update(ctx context.Context, uc *files.UploadCommand) (string, error)
	Content(ctx context.Context, cc *files.ContentCommand) ([]byte, error)
	Info(ctx context.Context, ID string) (*files.FileInfo, error)
	Delete(ctx context.Context, ID string) error
}
