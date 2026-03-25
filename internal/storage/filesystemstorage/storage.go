package filesystemstorage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"file-storage/internal/config"
	"file-storage/internal/errs"
	"file-storage/internal/filedata"
	"file-storage/internal/logger"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type FileSystemStorage struct {
	path string
	gc   *GarbageCollector
}

func New(cfg *config.FileSystem, log *slog.Logger) *FileSystemStorage {
	return &FileSystemStorage{
		path: cfg.Path,
		gc:   NewGarbageCollector(cfg.Path, 60*time.Minute, log),
	}
}

func (f *FileSystemStorage) Upsert(ctx context.Context, fd *filedata.FileData) (string, error) {
	start := time.Now()

	if fd == nil {
		return "", errs.ErrInvalidFileData
	}

	dirPath, err := fileCatalog(f.path, fd.ID)
	if err != nil {
		return "", fmt.Errorf("catalog name error: %w", err)
	}
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return "", fmt.Errorf("directory path creation error: %w", err)
	}

	lockFile, err := lockAcquire(fd.ID, dirPath)
	if err != nil {
		return "", fmt.Errorf("lock error: %w", err)
	}

	defer func() {
		if err := lockFile.Close(); err != nil {
			logger.FromContext(ctx).Warn(
				"unlock failed",
				"id", fd.ID,
				"error", err,
			)
		}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	fi := filedata.FileInfoFromFileData(fd)
	fiBytes, err := json.Marshal(fi)
	if err != nil {
		return "", fmt.Errorf("file info marshall error: %w", err)
	}

	basePath := filepath.Join(dirPath, fd.ID)
	currentVersions, newVersions, err := readVersions(basePath)
	if err != nil {
		return "", fmt.Errorf("get versions error: %w", err)
	}

	if fd.Data != nil {
		dataTempName := basePath + ".bin.tmp"
		dataName := dataFileName(basePath, newVersions)
		err = writeFile(fd.Data, dataName, dataTempName)
		if err != nil {
			return "", fmt.Errorf("write file data error: %w", err)
		}
	} else {
		newVersions.Data = currentVersions.Data
	}

	fiTempName := basePath + ".meta.json.tmp"
	fiName := metadataFileName(basePath, newVersions)
	err = writeFile(fiBytes, fiName, fiTempName)
	if err != nil {
		return "", fmt.Errorf("write file info error: %w", err)
	}

	err = commitVersion(basePath, newVersions)
	if err != nil {
		return "", fmt.Errorf("commit new version error: %w", err)
	}

	err = syncDir(dirPath)
	if err != nil {
		return "", fmt.Errorf("sync dir error: %w", err)
	}

	logLongCall(ctx, fd, start)

	return fd.ID, nil
}

func (f *FileSystemStorage) Delete(ctx context.Context, ID string) error {

	if len(ID) == 0 {
		return errs.ErrInvalidID
	}

	dirPath, err := fileCatalog(f.path, ID)
	if err != nil {
		return fmt.Errorf("catalog name error: %w", err)
	}

	if _, err := os.Stat(dirPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}

	lockFile, err := lockAcquire(ID, dirPath)
	if err != nil {
		return fmt.Errorf("lock error: %w", err)
	}
	defer func() {
		if err := lockFile.Close(); err != nil {
			logger.FromContext(ctx).Warn(
				"unlock failed",
				"id", ID,
				"error", err,
			)
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	filesToRemove, err := filenamesByID(dirPath, ID)
	if err != nil {
		return fmt.Errorf("files to remove search error: %w", err)
	}

	for _, fileName := range filesToRemove {
		err := os.Remove(filepath.Join(dirPath, fileName))
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("remove file error: %w", err)
		}
	}

	err = syncDir(dirPath)
	if err != nil {
		return fmt.Errorf("sync dir error: %w", err)
	}

	return nil
}

func (f *FileSystemStorage) Info(ctx context.Context, ID string) (*filedata.FileInfo, error) {
	if len(ID) == 0 {
		return nil, errs.ErrInvalidID
	}

	dirPath, err := fileCatalog(f.path, ID)
	if err != nil {
		return nil, fmt.Errorf("catalog name error: %w", err)
	}

	basePath := filepath.Join(dirPath, ID)
	v, _, err := readVersions(basePath)
	if err != nil {
		return nil, fmt.Errorf("read versions error: %w", err)
	}

	fileName := metadataFileName(basePath, v)
	b, err := os.ReadFile(fileName)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("read file error: %w", err)
	}

	var fi filedata.FileInfo
	err = json.Unmarshal(b, &fi)
	if err != nil {
		return nil, fmt.Errorf("unmarshal info error: %w", err)
	}

	return &fi, nil
}

func (f *FileSystemStorage) Content(ctx context.Context, ID string) (*filedata.ContentData, error) {
	fi, err := f.Info(ctx, ID)
	if err != nil {
		return nil, err
	}

	dirPath, err := fileCatalog(f.path, ID)
	if err != nil {
		return nil, fmt.Errorf("catalog name error: %w", err)
	}

	basePath := filepath.Join(dirPath, ID)
	v, _, err := readVersions(basePath)
	if err != nil {
		return nil, fmt.Errorf("read versions error: %w", err)
	}

	fileName := dataFileName(basePath, v)
	b, err := os.ReadFile(fileName)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("read file error: %w", err)
	}

	data := io.NopCloser(bytes.NewReader(b))
	return &filedata.ContentData{Data: data, IsImage: fi.IsImage}, nil
}
