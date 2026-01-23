package filesystemstorage

import (
	"context"
	"encoding/json"
	"errors"
	"file-storage/internal/config"
	"file-storage/internal/errs"
	"file-storage/internal/filedata"
	"file-storage/internal/logger"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileSystemStorage struct {
	path         string
	lockLifetime time.Duration
}

func New(cfg *config.FileSystem) *FileSystemStorage {
	return &FileSystemStorage{path: cfg.Path, lockLifetime: cfg.LockLifetime}
}

func (f *FileSystemStorage) Upsert(ctx context.Context, fd *filedata.FileData) (string, error) {
	start := time.Now()

	if fd == nil {
		return "", errs.ErrInvalidFileData
	}
	if len(fd.Data) == 0 {
		return "", errs.ErrInvalidFileData
	}

	dirPatn, err := f.fileCatalog(fd.ID)
	if err != nil {
		return "", fmt.Errorf("file data creation error: %w", err)
	}
	err = os.MkdirAll(dirPatn, 0755)
	if err != nil {
		return "", fmt.Errorf("directory path creation error: %w", err)
	}

	err = lock(fd.ID, dirPatn, f.lockLifetime)
	if err != nil {
		return "", fmt.Errorf("lock creation error: %w", err)
	}
	defer func() {
		if err := unlock(fd.ID, dirPatn); err != nil {
			logger.FromContext(ctx).Warn(
				"unlock failed",
				"id", fd.ID,
				"error", err,
			)
		}
	}()

	fi := filedata.FileInfoFromFileData(fd)
	fiBytes, err := json.Marshal(fi)
	if err != nil {
		return "", fmt.Errorf("file info marshall error: %w", err)
	}

	dataTempName := filepath.Join(dirPatn, fd.ID+"."+string(fd.Format)+"_tmp")
	dataName := filepath.Join(dirPatn, fd.ID+"."+string(fd.Format))
	err = writeFile(fd.Data, dataName, dataTempName)
	if err != nil {
		return "", fmt.Errorf("write file data error: %w", err)
	}

	fiTempName := filepath.Join(dirPatn, fd.ID+".meta.json.tmp")
	fiName := filepath.Join(dirPatn, fd.ID+".meta.json")
	err = writeFile(fiBytes, fiName, fiTempName)
	if err != nil {
		return "", fmt.Errorf("write file info error: %w", err)
	}

	err = syncDir(dirPatn)
	if err != nil {
		return "", fmt.Errorf("sync dir error: %w", err)
	}

	logLongCall(ctx, fd, start, f.lockLifetime)

	return fd.ID, nil
}

func lock(id string, dirPatn string, lockLifetime time.Duration) error {

	fn := lockFileName(dirPatn, id)
	now := time.Now().UTC()

	file, err := os.OpenFile(fn, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err == nil {
		defer file.Close()
	} else {
		if errors.Is(err, fs.ErrExist) {
			b, err := os.ReadFile(fn)
			if err != nil {
				return fmt.Errorf("lock file read error: %w", err)
			}

			lockContent := strings.TrimSpace(string(b))
			t, err := time.Parse(time.RFC3339, lockContent)
			if err != nil {
				fileInfo, errStat := os.Stat(fn)
				if errStat != nil {
					return fmt.Errorf("lock file stat observing error: %w", errStat)
				}
				t = fileInfo.ModTime().Add(lockLifetime)
			}

			if now.Before(t) {
				return errs.ErrStorageFileIsLocked
			}

			err = os.Remove(fn)
			if err != nil {
				return fmt.Errorf("remove lock file error: %w", err)
			}

			file, err = os.OpenFile(fn, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("parallel lock file write error: %w", errs.ErrStorageFileIsLocked)
			}
			defer file.Close()
		} else {
			return fmt.Errorf("lock file open error: %w", err)
		}
	}

	deadline := now.Add(lockLifetime).Format(time.RFC3339)
	_, err = file.WriteString(deadline)
	if err != nil {
		return fmt.Errorf("lock file write error: %w", err)
	}

	err = file.Sync()
	if err != nil {
		return fmt.Errorf("lock file sync error: %w", err)
	}

	return nil
}

func unlock(id string, dirPatn string) error {
	fn := lockFileName(dirPatn, id)
	err := os.Remove(fn)
	return err
}

func (f *FileSystemStorage) fileCatalog(id string) (string, error) {

	r := []rune(id)

	if len(r) < 6 {
		return "", errs.ErrInvalidID
	}

	cat1 := string(r[0:2])
	cat2 := string(r[2:4])

	return filepath.Join(f.path, cat1, cat2), nil
}

func lockFileName(catalog, id string) string {
	return filepath.Join(catalog, id+".lock")
}

func writeFile(data []byte, path, tempPath string) error {
	file, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open file error: %w", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("write file error: %w", err)
	}

	err = file.Sync()
	if err != nil {
		return fmt.Errorf("file sync error: %w", err)
	}

	err = os.Rename(tempPath, path)
	if err != nil {
		return fmt.Errorf("rename file error: %w", err)
	}

	return nil
}

func syncDir(dirPatn string) error {

	cat, err := os.Open(dirPatn)
	if err != nil {
		return fmt.Errorf("open catalog error: %w", err)
	}
	defer cat.Close()

	err = cat.Sync()
	if err != nil {
		return fmt.Errorf("sync catalog error: %w", err)
	}

	return nil
}

func logLongCall(ctx context.Context, fd *filedata.FileData, start time.Time, threshold time.Duration) {
	t := time.Since(start)
	if t < threshold {
		return
	}

	log := logger.FromContext(ctx)
	log.Warn("long upsert call",
		"duration", t,
		"threshold", threshold,
		"id", fd.ID,
		"fileSize", fd.FileSize,
	)

}
