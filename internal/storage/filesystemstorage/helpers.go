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
	slotA          = "A"
	slotB          = "B"
	activeStateExt = "active"
	lockExt        = "lock"
	binExt         = "bin"
	metadataExt    = "meta.json"
)

type activeState struct {
	Data     string
	Metadata string
}

func lockAcquire(id string, dirPath string) (*os.File, error) {

	fn := lockFileFullName(dirPath, id)

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

func lockFileFullName(catalog, id string) string {
	return filepath.Join(catalog, lockFileName(id))
}

func lockFileName(id string) string {
	return id + "." + lockExt
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

func slotInfo(dirPath, id string) (activeState, activeState, error) {
	as, err := readActiveState(dirPath, id)
	if err != nil {
		return activeState{}, activeState{}, err
	}

	currentDataState, newDataState := calcActiveState(as.Data)
	currentMetadataState, newMetadataState := calcActiveState(as.Metadata)

	return activeState{Data: currentDataState, Metadata: currentMetadataState},
		activeState{Data: newDataState, Metadata: newMetadataState},
		nil
}

func readActiveState(dirPath, id string) (activeState, error) {
	activeStatePath := activeStateFileFullName(dirPath, id)

	as := &activeState{}

	b, err := os.ReadFile(activeStatePath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return activeState{}, err
		}
	} else {
		err = json.Unmarshal(b, as)
		if err != nil {
			return activeState{}, err
		}
	}

	return *as, nil
}

func calcActiveState(currentState string) (string, string) {
	newState := ""

	switch currentState {
	case slotA:
		newState = slotB
	case slotB:
		newState = slotA
	default:
		newState = slotA
		currentState = ""
	}

	return currentState, newState
}

func commitActiveState(dirPath, id string, v activeState) error {
	activeStatePath := activeStateFileFullName(dirPath, id)
	tempPath := activeStatePath + ".tmp"

	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshall activeState file error: %w", err)
	}
	err = writeFile(b, activeStatePath, tempPath)
	if err != nil {
		return fmt.Errorf("write activeState file error: %w", err)
	}
	return nil
}

func activeStateFileFullName(dirPath, id string) string {
	return filepath.Join(dirPath, activeStateFileName(id))
}

func activeStateFileName(id string) string {
	return id + "." + activeStateExt
}

func metadataFileFullName(dirPath, id string, activeState activeState) string {
	return filepath.Join(dirPath, metadataFileName(id, activeState))
}

func metadataFileName(id string, activeState activeState) string {
	if activeState.Metadata != "" {
		return id + "." + activeState.Metadata + "." + metadataExt
	}
	return id + "." + metadataExt
}

func dataFileFullName(dirPath, id string, activeState activeState) string {
	return filepath.Join(dirPath, dataFileName(id, activeState))
}

func dataFileName(id string, activeState activeState) string {
	if activeState.Data != "" {
		return id + "." + activeState.Data + "." + binExt
	}
	return id + "." + binExt
}

func filenamesByID(dirPath string, id string) ([]string, error) {
	s, err := os.ReadDir(dirPath)

	if err != nil {
		return nil, err
	}

	result := make([]string, 0, 6)
	prefix := id + "."
	for _, file := range s {
		filename := file.Name()
		if strings.HasPrefix(filename, prefix) {
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
