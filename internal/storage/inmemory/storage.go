// Package inmemory provides an in-memory storage implementation.
//
// The package is intended for testing and development
// and should not be used as a production storage.
package inmemory

import (
	"bytes"
	"context"
	"file-storage/internal/errs"
	"file-storage/internal/filedata"
	"io"
	"strings"
	"sync"
)

type MemoryStorage struct {
	mu      sync.RWMutex
	storage map[string]*filedata.FileData
}

func New() *MemoryStorage {
	return &MemoryStorage{storage: make(map[string]*filedata.FileData)}
}

func (s *MemoryStorage) Upsert(ctx context.Context, fd *filedata.FileData) (string, error) {

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

func Delete(ctx context.Context, ID string) error {

	return nil
}

func (s *MemoryStorage) Info(ctx context.Context, ID string) (*filedata.FileInfo, error) {
	s.mu.RLock()
	fd := s.storage[ID]
	s.mu.RUnlock()

	if fd == nil {
		return nil, errs.ErrNotFound
	}

	return filedata.FileInfoFromFileData(fd), nil
}

func (s *MemoryStorage) Content(ctx context.Context, ID string) (*filedata.ContentData, error) {
	s.mu.RLock()
	fd := s.storage[ID]
	s.mu.RUnlock()

	if fd == nil {
		return nil, errs.ErrNotFound
	}

	b := make([]byte, len(fd.Data))
	copy(b, fd.Data)

	cd := filedata.ContentData{
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

func copyFileData(fd *filedata.FileData) *filedata.FileData {
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
