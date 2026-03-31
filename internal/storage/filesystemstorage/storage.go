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
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

type FileSystemStorage struct {
	path   string
	gc     *GarbageCollector
	gcOnce sync.Once
}

func New(cfg *config.FileSystem, log *slog.Logger) (*FileSystemStorage, error) {

	err := os.MkdirAll(cfg.Path, 0755)
	if err != nil {
		return nil, err
	}

	err = flockSupportTest(cfg.Path)
	if err != nil {
		return nil, err
	}

	var gc *GarbageCollector
	if cfg.GarbageCollector.Enabled {
		gc = NewGarbageCollector(cfg.Path, cfg.GarbageCollector.Interval, cfg.GarbageCollector.WorkersCount, log)
	}

	fss := &FileSystemStorage{
		path: cfg.Path,
		gc:   gc,
	}

	return fss, nil
}

func flockSupportTest(path string) error {

	id := "flockSupportTest"
	f, err := lockAcquire(id, path)
	if err != nil {
		return fmt.Errorf("flock support test error: %w", err)
	}
	fClosed := false
	defer func() {
		if !fClosed {
			f.Close()
		}
		filename := filepath.Join(path, id+"."+lockExt)
		os.Remove(filename)
	}()

	fLocked, err := lockAcquireWithFlags(id, path, unix.LOCK_EX|unix.LOCK_NB)
	if err == nil {
		fLocked.Close()
		return errs.ErrFlockSupportTestError
	}
	if !errors.Is(err, unix.EWOULDBLOCK) && !errors.Is(err, unix.EAGAIN) {
		return fmt.Errorf("blocked file locking flock support test error: %w", err)
	}

	err = f.Close()
	if err != nil {
		return fmt.Errorf("flock support test error on closing lock file: %w", err)
	}
	fClosed = true

	fLocked, err = lockAcquireWithFlags(id, path, unix.LOCK_EX|unix.LOCK_NB)
	if err != nil {
		return fmt.Errorf("file locking flock support test error: %w", err)
	}
	fLocked.Close()

	return nil
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

	currentAtiveState, newAtiveState, err := slotInfo(dirPath, fd.ID)
	if err != nil {
		currentAtiveState, newAtiveState, err = slotInfoWithRecovery(dirPath, fd.ID, lockFile)
		if err != nil {
			return "", fmt.Errorf("get activeState error: %w", err)
		}
	}

	basePath := filepath.Join(dirPath, fd.ID)
	if fd.Data != nil {
		dataTempName := basePath + ".bin.tmp"
		dataName := dataFileFullName(dirPath, fd.ID, newAtiveState)
		err = writeFile(fd.Data, dataName, dataTempName)
		if err != nil {
			return "", fmt.Errorf("write file data error: %w", err)
		}
	} else {
		newAtiveState.Data = currentAtiveState.Data
	}

	fiTempName := basePath + ".meta.json.tmp"
	fiName := metadataFileFullName(dirPath, fd.ID, newAtiveState)
	err = writeFile(fiBytes, fiName, fiTempName)
	if err != nil {
		return "", fmt.Errorf("write file info error: %w", err)
	}

	err = syncDir(dirPath)
	if err != nil {
		return "", fmt.Errorf("sync dir error: %w", err)
	}

	err = commitActiveState(dirPath, fd.ID, newAtiveState)
	if err != nil {
		return "", fmt.Errorf("commit new activeState error: %w", err)
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

	v, _, err := slotInfo(dirPath, ID)
	if err != nil {
		return nil, fmt.Errorf("read activeState error: %w", err)
	}

	return readFileInfo(dirPath, ID, v)
}

func readFileInfo(dirPath, ID string, activeState activeState) (*filedata.FileInfo, error) {
	fileName := metadataFileFullName(dirPath, ID, activeState)
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

	dirPath, err := fileCatalog(f.path, ID)
	if err != nil {
		return nil, fmt.Errorf("catalog name error: %w", err)
	}

	activeState, _, err := slotInfo(dirPath, ID)
	if err != nil {
		return nil, fmt.Errorf("read activeState error: %w", err)
	}

	fi, err := readFileInfo(dirPath, ID, activeState)
	if err != nil {
		return nil, err
	}

	fileName := dataFileFullName(dirPath, ID, activeState)
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

func (f *FileSystemStorage) StartGC(ctx context.Context) {

	f.gcOnce.Do(
		func() {
			if f.gc != nil {
				go f.gc.Run(ctx)
			}
		},
	)

}
