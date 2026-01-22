package handlers

import (
	"context"
	"file-storage/internal/filedata"
)

type Service interface {
	Update(ctx context.Context, uc *filedata.UploadCommand) (string, error)
	Content(ctx context.Context, cc *filedata.ContentCommand) ([]byte, error)
	Info(ctx context.Context, ID string) (*filedata.FileInfo, error)
	Delete(ctx context.Context, ID string) error
}
