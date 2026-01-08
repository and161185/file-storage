package files

import (
	"context"
	"file-storage/internal/config"
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

	var imageInfo *ImageInfo
	data := uc.Data
	if uc.IsImage {
		var err error
		data, imageInfo, err = ProcessImage(data, s.cfg.Ext, s.cfg.MaxDimention, s.cfg.MaxDimention)
		if err != nil {
			return "", err
		}
	}

	fd := FileData{
		ID:        uc.ID,
		Data:      data,
		Hash:      uc.Hash,
		IsImage:   uc.IsImage,
		FileSize:  len(data),
		Metadata:  uc.Metadata,
		UpdatedAt: time.Now(),
	}

	if imageInfo != nil {
		fd.Format = imageInfo.Format
		fd.Width = imageInfo.Width
		fd.Height = imageInfo.Height
	}

	ID, err := s.storage.Upsert(ctx, &fd)
	if err != nil {
		return "", err
	}

	return ID, nil
}
