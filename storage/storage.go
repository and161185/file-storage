package storage

import (
	"context"
	"file-storage/models"
)

// Интерфейс для сохранения файлов
type StorageService interface {
	SaveFile(ctx context.Context, data []byte, metadata map[string]interface{}, fileID string) (string, error)
	GetFile(ctx context.Context, fileID string) ([]byte, map[string]interface{}, error)
	DeleteFile(ctx context.Context, fileID string) error
	Ready(ctx context.Context) models.ReadyResponse
}
