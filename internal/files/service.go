package files

import (
	"context"
	"file-storage/internal/config"
	"file-storage/internal/errs"
	"file-storage/internal/filedata"
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

// Update validates input data and stores file content and metadata.
// The operation is idempotent for the same file ID.
func (s *Service) Update(ctx context.Context, uc *filedata.UploadCommand) (string, error) {

	updateData := true
	createdAt := time.Now()
	fi, err := s.Info(ctx, uc.ID)
	if err == nil {
		updateData = uc.Hash != fi.Hash
		createdAt = fi.CreatedAt
	}

	var fd filedata.FileData
	if updateData {
		var imageInfo *filedata.ImageInfo
		data := uc.Data
		if uc.IsImage {
			var err error
			data, imageInfo, err = ProcessImage(data, s.cfg.Ext, s.cfg.MaxDimention, s.cfg.MaxDimention)
			if err != nil {
				return "", fmt.Errorf("image processing error: %w", err)
			}
		}

		fd = filedata.FileData{
			ID:        uc.ID,
			Data:      data,
			Hash:      uc.Hash,
			IsImage:   uc.IsImage,
			FileSize:  len(data),
			Metadata:  uc.Metadata,
			UpdatedAt: time.Now(),
			CreatedAt: createdAt,
		}

		if imageInfo != nil {
			fd.Format = imageInfo.Format
			fd.Width = imageInfo.Width
			fd.Height = imageInfo.Height
		}
	} else {
		fd = filedata.FileData{
			ID:        uc.ID,
			Data:      nil,
			Hash:      fi.Hash,
			IsImage:   fi.IsImage,
			FileSize:  fi.FileSize,
			Metadata:  uc.Metadata,
			UpdatedAt: time.Now(),
			CreatedAt: createdAt,
			Format:    fi.Format,
			Width:     fi.Width,
			Height:    fi.Height,
		}
	}

	ID, err := s.storage.Upsert(ctx, &fd)
	if err != nil {
		return "", fmt.Errorf("storage error: %w", err)
	}

	return ID, nil
}

// Content returns file content by ID with optional image transformations.
// The method returns only file bytes.
func (s *Service) Content(ctx context.Context, cc *filedata.ContentCommand) ([]byte, error) {

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

// Info returns file metadata by ID.
// The response does not contain file content.
func (s *Service) Info(ctx context.Context, ID string) (*filedata.FileInfo, error) {

	fi, err := s.storage.Info(ctx, ID)
	if err != nil {
		return nil, fmt.Errorf("storage error: %w", err)
	}

	return fi, nil
}

// Delete removes a file by ID.
// The operation is idempotent for the same file ID.
func (s *Service) Delete(ctx context.Context, ID string) error {

	err := s.storage.Delete(ctx, ID)
	if err != nil {
		return fmt.Errorf("storage error: %w", err)
	}

	return nil
}
