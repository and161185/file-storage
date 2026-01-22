package filesystemstorage

import (
	"context"
	"encoding/json"
	"file-storage/internal/config"
	"file-storage/internal/errs"
	"file-storage/internal/filedata"
	"fmt"
	"os"
	"strings"
	"time"
)

type FileSystemStorage struct {
	path         string
	lockLifetime time.Duration
}

func New(cfg *config.FileSystem) *FileSystemStorage {
	p := cfg.Path
	if !strings.HasSuffix(cfg.Path, "/") {
		p = p + "/"
	}
	return &FileSystemStorage{path: p, lockLifetime: cfg.LockLifetime}
}

func (f *FileSystemStorage) Upsert(ctx context.Context, fd *filedata.FileData) (string, error) {
	if fd == nil {
		return "", errs.ErrInvalidFileData
	}

	fi := filedata.FileInfoFromFileData(fd)
	fiBytes, err := json.Marshal(fi)
	if err != nil {
		return "", fmt.Errorf("file info marshall error: %w", err)
	}

	return "", nil
}

func (f *FileSystemStorage) lock(id string) {
	fn := f.fileName(id, "lock")
	os.OpenFile()
}

func (f *FileSystemStorage) fileName(id string, ext string) string {
	r := []rune(id)
	cat1 := string(r[0:2])
	cat2 := string(r[2:4])

	return f.path + cat1 + "/" + cat2 + "/" + id + "." + ext
}
