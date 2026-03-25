package filesystemstorage

import (
	"context"
	"encoding/json"
	"errors"
	"file-storage/internal/errs"
	"file-storage/internal/filedata"
	"file-storage/internal/logger"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

const (
	versionA = "A"
	versionB = "B"
)

type versions struct {
	Data     string
	Metadata string
}

func lockAcquire(id string, dirPath string) (*os.File, error) {

	fn := lockFileName(dirPath, id)

	file, err := os.OpenFile(fn, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("lock file open error: %w", err)
	}

	err = unix.Flock(int(file.Fd()), unix.LOCK_EX)
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("lock file lock error: %w", err)
	}

	return file, nil
}

func fileCatalog(path, id string) (string, error) {

	r := []rune(id)

	if len(r) < 6 {
		return "", errs.ErrInvalidID
	}

	cat1 := string(r[0:2])
	cat2 := string(r[2:4])

	return filepath.Join(path, cat1, cat2), nil
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

func syncDir(dirPath string) error {

	cat, err := os.Open(dirPath)
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

func readVersions(basePath string) (versions, versions, error) {
	versionsPath := versionsFileName(basePath)

	v := &versions{}

	b, err := os.ReadFile(versionsPath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return versions{}, versions{}, err
		}
	} else {
		err = json.Unmarshal(b, v)
		if err != nil {
			return versions{}, versions{}, err
		}
	}

	currentDataVersion, newDataVersion := calcVersions(v.Data)
	currentMetadataVersion, newMetadataVersion := calcVersions(v.Metadata)

	return versions{Data: currentDataVersion, Metadata: currentMetadataVersion},
		versions{Data: newDataVersion, Metadata: newMetadataVersion},
		nil

}

func calcVersions(currentVersion string) (string, string) {
	newVersion := ""

	switch currentVersion {
	case versionA:
		newVersion = versionB
	case versionB:
		newVersion = versionA
	default:
		newVersion = versionA
		currentVersion = ""
	}

	return currentVersion, newVersion
}

func commitVersion(basePath string, v versions) error {
	versionsPath := versionsFileName(basePath)
	tempPath := versionsPath + ".tmp"

	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshall versions file error: %w", err)
	}
	err = writeFile(b, versionsPath, tempPath)
	if err != nil {
		return fmt.Errorf("write versions file error: %w", err)
	}
	return nil
}

func versionsFileName(basePath string) string {
	return basePath + ".versions"
}

func metadataFileName(basePath string, versions versions) string {
	return basePath + "." + versions.Metadata + ".meta.json"
}

func dataFileName(basePath string, versions versions) string {
	return basePath + "." + versions.Data + ".bin"
}

func filenamesByID(dirPath string, ID string) ([]string, error) {
	s, err := os.ReadDir(dirPath)

	if err != nil {
		return nil, err
	}

	result := make([]string, 0, 6)
	for _, file := range s {
		filename := file.Name()
		if strings.HasPrefix(filename, ID) {
			result = append(result, filename)
		}
	}
	return result, nil
}

func logLongCall(ctx context.Context, fd *filedata.FileData, start time.Time) {
	threshold := 2 * time.Second

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
