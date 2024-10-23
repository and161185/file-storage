package storage

import (
	"context"
	"file-storage/models"
)

type FileMetadata = models.FileMetadata

// Интерфейс для сохранения файлов
type StorageService interface {
	SaveFile(ctx context.Context, data []byte, metadata FileMetadata, fileID string) (string, error)
	GetFile(ctx context.Context, fileID string) ([]byte, FileMetadata, error)
	DeleteFile(ctx context.Context, fileID string) error
}
