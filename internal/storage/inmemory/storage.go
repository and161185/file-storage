package inmemory

import (
	"bytes"
	"context"
	"file-storage/internal/errs"
	"file-storage/internal/files"
	"io"
	"strings"
	"sync"
)

type MemoryStorage struct {
	mu      sync.RWMutex
	storage map[string]*files.FileData
}

func New() *MemoryStorage {
	return &MemoryStorage{storage: make(map[string]*files.FileData)}
}

func (s *MemoryStorage) Upsert(ctx context.Context, fd *files.FileData) (string, error) {

	if fd == nil {
		return "", errs.ErrInvalidFileData
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if strings.TrimSpace(fd.ID) == "" {
		return "", errs.ErrInvalidID
	}

	value := copyFileData(fd)

	s.storage[fd.ID] = value

	return fd.ID, nil
}

func (s *MemoryStorage) Info(ctx context.Context, ID string) (*files.FileInfo, error) {
	s.mu.RLock()
	fd := s.storage[ID]
	s.mu.RUnlock()

	if fd == nil {
		return nil, errs.ErrNotFound
	}

	return fileInfo(fd), nil
}

func (s *MemoryStorage) Content(ctx context.Context, ID string) (*files.ContentData, error) {
	s.mu.RLock()
	fd := s.storage[ID]
	s.mu.RUnlock()

	if fd == nil {
		return nil, errs.ErrNotFound
	}

	b := make([]byte, len(fd.Data))
	copy(b, fd.Data)

	cd := files.ContentData{
		Data:    io.NopCloser(bytes.NewReader(b)),
		IsImage: fd.IsImage,
	}

	return &cd, nil
}

func (s *MemoryStorage) Delete(ctx context.Context, ID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.storage, ID)

	return nil
}

func copyFileData(fd *files.FileData) *files.FileData {
	value := *fd

	if fd.Metadata != nil {
		metadata := make(map[string]any, len(fd.Metadata))
		for k, v := range fd.Metadata {
			metadata[k] = v
		}
		value.Metadata = metadata
	}

	if fd.Data != nil {
		b := make([]byte, len(fd.Data))
		copy(b, fd.Data)
		value.Data = b
	}

	return &value
}

func fileInfo(fd *files.FileData) *files.FileInfo {
	value := files.FileInfo{
		ID:        fd.ID,
		FileSize:  fd.FileSize,
		CreatedAt: fd.CreatedAt,
		UpdatedAt: fd.UpdatedAt,
	}

	if fd.Metadata != nil {
		metadata := make(map[string]any, len(fd.Metadata))
		for k, v := range fd.Metadata {
			metadata[k] = v
		}
		value.Metadata = metadata
	}

	return &value
}
