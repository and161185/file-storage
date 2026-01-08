package files

import (
	"context"
	"file-storage/internal/config"
	"file-storage/internal/imgproc"
	"time"
)

type Service struct {
	cfg     *config.Image
	storage Storage
}

func NewService(cfg *config.Image, storage Storage) *Service {
	return &Service{cfg: cfg, storage: storage}
}

func (s *Service) Update(ctx context.Context, uc *UploadCommand) (string, error) {

	ext := ""
	width := 0
	height := 0
	data := uc.Data
	if uc.IsImage {
		var err error
		ext, width, height, err = imgproc.ImageConfig(data)
		if err != nil {
			return "", err
		}

		data, err = imgproc.Convert(data, ext, width, height, s.cfg.Ext, s.cfg.MaxDimention, s.cfg.MaxDimention)
		if err != nil {
			return "", err
		}
	}

	fd := FileData{
		ID:        uc.ID,
		Data:      data,
		Hash:      uc.Hash,
		IsImage:   uc.IsImage,
		Ext:       ext,
		Width:     width,
		Height:    height,
		FileSize:  len(data),
		Metadata:  uc.Metadata,
		UpdatedAt: time.Now(),
	}

	ID, err := s.storage.Upsert(ctx, &fd)
	if err != nil {
		return "", err
	}

	return ID, nil
}
