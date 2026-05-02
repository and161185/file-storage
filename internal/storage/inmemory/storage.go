// Package inmemory provides an in-memory storage implementation.
//
// It is intended for testing and development and should not be used
// as a production storage backend.
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

// MemoryStorage stores files in process memory and is intended for testing or local runs.
type MemoryStorage struct {
	mu      sync.RWMutex
	storage map[string]*filedata.FileData
}

// New creates an empty in-memory storage.
func New() *MemoryStorage {
	return &MemoryStorage{storage: make(map[string]*filedata.FileData)}
}

// Upsert creates a new file or replaces an existing one in memory.
func (s *MemoryStorage) Upsert(ctx context.Context, fd *filedata.FileData) (string, error) {

	if fd == nil {
		return "", errs.ErrInvalidFileData
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if strings.TrimSpace(fd.ID) == "" {
		return "", errs.ErrInvalidID
	}

	value := copyFileData(fd, s.storage[fd.ID])

	s.storage[fd.ID] = value

	return fd.ID, nil
}

// Info returns file metadata from in-memory storage.
func (s *MemoryStorage) Info(ctx context.Context, ID string) (*filedata.FileInfo, error) {
	s.mu.RLock()
	fd := s.storage[ID]
	s.mu.RUnlock()

	if fd == nil {
		return nil, errs.ErrNotFound
	}

	return filedata.FileInfoFromFileData(fd), nil
}

// Content returns file content from in-memory storage.
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

// Delete removes a file from in-memory storage.
func (s *MemoryStorage) Delete(ctx context.Context, ID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.storage, ID)

	return nil
}

func copyFileData(fd *filedata.FileData, currentValue *filedata.FileData) *filedata.FileData {
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
	} else {
		if currentValue != nil && currentValue.Data != nil {
			b := make([]byte, len(currentValue.Data))
			copy(b, currentValue.Data)
			value.Data = b
		}
	}

	return &value
}
