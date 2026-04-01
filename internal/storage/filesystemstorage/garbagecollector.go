package filesystemstorage

import (
	"context"
	"errors"
	"file-storage/internal/logger"
	"file-storage/internal/metrics"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type cleanupJob struct {
	dirEntries []os.DirEntry
	dirPath    string
	id         string
}

type GarbageCollector struct {
	path     string
	interval time.Duration
	workers  int
	log      *slog.Logger
}

func NewGarbageCollector(path string, interval time.Duration, workers int, log *slog.Logger) *GarbageCollector {
	return &GarbageCollector{
		path:     path,
		interval: interval,
		workers:  workers,
		log:      logger.WithComponent(log, logger.ComponentGC),
	}
}

func (gc *GarbageCollector) Run(ctx context.Context) {

	cleanupJobs := make(chan *cleanupJob, gc.workers*2)

	var wg sync.WaitGroup
	for range gc.workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gc.worker(ctx, cleanupJobs, gc.log)
		}()
	}

	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			close(cleanupJobs)
		}()
		for {
			func() {
				metrics.GcRunsTotal.Inc()
				metrics.GcInProgress.Set(1)
				begin := time.Now()
				defer func() {
					metrics.GcDurationSeconds.Observe(time.Since(begin).Seconds())
					metrics.GcInProgress.Set(0)
				}()

				err := gc.collectGarbage(ctx, cleanupJobs)
				if err != nil {
					if ctx.Err() != nil && errors.Is(err, ctx.Err()) {
						return
					}
					gc.log.Error("garbage collector error", slog.Any(logger.LogFieldError, err))
					metrics.GcErrorsTotal.Inc()
				}
			}()

			select {
			case <-time.After(gc.interval):
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()
}

func (gc *GarbageCollector) collectGarbage(ctx context.Context, jobs chan *cleanupJob) error {
	level1Entries, err := os.ReadDir(gc.path)
	if err != nil {
		return fmt.Errorf("gc.path %s reading error: %w", gc.path, err)
	}

	for _, level1Entry := range level1Entries {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if !level1Entry.IsDir() {
			continue
		}

		dirLevel1Path := filepath.Join(gc.path, level1Entry.Name())
		level2Entries, err := os.ReadDir(dirLevel1Path)
		if err != nil {
			err = fmt.Errorf("subdirectory %s reading error: %w", dirLevel1Path, err)
			gc.log.Warn("garbage collector error", slog.Any(logger.LogFieldError, err))
			continue
		}

		for _, level2Entry := range level2Entries {

			if ctx.Err() != nil {
				return ctx.Err()
			}
			if !level2Entry.IsDir() {
				continue
			}

			dirLevel2Path := filepath.Join(dirLevel1Path, level2Entry.Name())
			filesEntries, err := os.ReadDir(dirLevel2Path)
			if err != nil {
				err = fmt.Errorf("subdirectory %s reading error: %w", dirLevel2Path, err)
				gc.log.Warn("garbage collector error", slog.Any(logger.LogFieldError, err))
				continue
			}

			m := make(map[string][]os.DirEntry)
			for _, fileEntry := range filesEntries {
				fns := disassembleFilename(fileEntry.Name())
				m[fns.id] = append(m[fns.id], fileEntry)
			}

			for id, entries := range m {
				select {
				case jobs <- &cleanupJob{dirEntries: entries, id: id, dirPath: dirLevel2Path}:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}

	return nil
}

func (gc *GarbageCollector) worker(ctx context.Context, cleanupJobs chan *cleanupJob, log *slog.Logger) {
	for {
		select {
		case j, ok := <-cleanupJobs:
			if !ok {
				return
			}
			err := gc.removeGarbage(j, log)
			if err != nil {
				log.Error("remove garbage error", slog.Any(logger.LogFieldError, err))
				metrics.GcErrorsTotal.Inc()
			}
		case <-ctx.Done():
			return
		}
	}
}

func (gc *GarbageCollector) removeGarbage(j *cleanupJob, log *slog.Logger) error {

	if len(j.dirEntries) == 0 {
		return nil
	}

	lockFile, err := lockAcquire(j.id, j.dirPath)
	if err != nil {
		return err
	}
	defer lockFile.Close()

	keepFiles, recovered, err := activeFiles(j.id, j.dirPath, lockFile)
	if recovered {
		if err == nil {
			log.Info("active state recovered",
				"id", j.id,
			)
		} else {
			log.Warn("active state recover error",
				"id", j.id,
				"error", err,
			)
		}
	}
	if err != nil {
		return err
	}

	callSyncDir := false
	for _, e := range j.dirEntries {
		name := e.Name()
		if _, ok := keepFiles[name]; ok {
			continue
		}

		err := os.Remove(filepath.Join(j.dirPath, name))
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("remove file error: %w", err)
		}
		metrics.GcFilesDeletedTotal.Inc()
		callSyncDir = true
	}

	if callSyncDir {
		err = syncDir(j.dirPath)
		if err != nil {
			return fmt.Errorf("sync dir error: %w", err)
		}
	}

	return nil
}

func activeFiles(id, dirPath string, lockFile *os.File) (map[string]struct{}, bool, error) {

	recovered := false
	activeState, _, err := slotInfo(dirPath, id)
	if err != nil {
		recovered = true
		activeState, _, err = slotInfoWithRecovery(dirPath, id, lockFile)
		if err != nil {
			return nil, recovered, err
		}
		metrics.GcRecoveryTotal.Inc()
		err = syncDir(dirPath)
		if err != nil {
			return nil, recovered, fmt.Errorf("sync dir error: %w", err)
		}
	}

	m := make(map[string]struct{}, 4)
	m[dataFileName(id, activeState)] = struct{}{}
	m[metadataFileName(id, activeState)] = struct{}{}
	m[lockFileName(id)] = struct{}{}
	m[activeStateFileName(id)] = struct{}{}

	return m, recovered, nil
}
