package files

import (
	"context"
	"file-storage/internal/config"
	"file-storage/internal/errs"
	"fmt"
	"io"
	"time"
)

const (
	minContentDimension = 10
	maxContentDimension = 10000
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
			return "", fmt.Errorf("image processing error: %w", err)
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
		return "", fmt.Errorf("storage error: %w", err)
	}

	return ID, nil
}

func (s *Service) Content(ctx context.Context, cc *ContentCommand) ([]byte, error) {

	var format string
	var width int
	var height int

	if cc.Format != nil {
		format = *cc.Format
	} else {
		format = s.cfg.Ext
	}

	if cc.Width != nil {
		width = *cc.Width
		if width < minContentDimension || width > maxContentDimension {
			return nil, fmt.Errorf("width must be between %d and %d: %w", minContentDimension, maxContentDimension, errs.ErrWrongUrlParameter)
		}
	} else {
		width = s.cfg.MaxDimention
	}

	if cc.Height != nil {
		height = *cc.Height
		if height < minContentDimension || height > maxContentDimension {
			return nil, fmt.Errorf("height must be between %d and %d: %w", minContentDimension, maxContentDimension, errs.ErrWrongUrlParameter)
		}
	} else {
		height = s.cfg.MaxDimention
	}

	cd, err := s.storage.Content(ctx, cc.ID)
	if err != nil {
		return nil, fmt.Errorf("storage error: %w", err)
	}
	defer cd.Data.Close()

	b, err := io.ReadAll(cd.Data)
	if err != nil {
		return nil, fmt.Errorf("data read error: %w: %v", errs.ErrInvalidFileData, err)
	}

	if cd.IsImage {
		b, _, err = ProcessImage(b, format, width, height)
		if err != nil {
			return nil, fmt.Errorf("processing image error: %w", err)
		}
	}

	return b, nil
}

func (s *Service) Info(ctx context.Context, ID string) (*FileInfo, error) {

	fi, err := s.storage.Info(ctx, ID)
	if err != nil {
		return nil, fmt.Errorf("storage error: %w", err)
	}

	return fi, nil
}

func (s *Service) Delete(ctx context.Context, ID string) error {

	err := s.storage.Delete(ctx, ID)
	if err != nil {
		return fmt.Errorf("storage error: %w", err)
	}

	return nil
}
