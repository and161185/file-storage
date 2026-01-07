package files

import (
	"context"
	"file-storage/internal/imgproc"
	"time"
)

type Service struct {
	storage Storage
}

func NewService(storage Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) Update(ctx context.Context, uc *UploadCommand) (string, error) {

	mimeType, err := imgproc.MimeType(uc.Data)
	if err != nil {
		return "", err
	}

	fd := FileData{
		ID:        uc.ID,
		Data:      uc.Data,
		Hash:      uc.Hash,
		MimeType:  mimeType,
		FileSize:  len(uc.Data),
		Metadata:  uc.Metadata,
		UpdatedAt: time.Now(),
	}

	ID, err := s.storage.Upsert(ctx, &fd)
	if err != nil {
		return "", err
	}

	return ID, nil
}
