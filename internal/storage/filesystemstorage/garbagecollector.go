package filesystemstorage

import (
	"context"
	"file-storage/internal/logger"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

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

	jobs := make(chan string, gc.workers*2)

	for {

		err := gc.CollectGarbage(ctx, jobs)
		if err != nil {
			gc.log.Error("garbage collector error", slog.Any(logger.LogFieldError, err))
		}

		select {
		case <-time.After(gc.interval):
			continue
		case <-ctx.Done():
			return
		}
	}
}

func (gc *GarbageCollector) CollectGarbage(ctx context.Context, jobs chan string) error {
	entries, err := os.ReadDir(gc.path)
	if err != nil {
		return fmt.Errorf("gc.path %s reading error: %w", gc.path, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		subDirPath := filepath.Join(gc.path, entry.Name())
		subEntries, err := os.ReadDir(subDirPath)
		if err != nil {
			err = fmt.Errorf("subdirectory %s reading error: %w", subDirPath, err)
			gc.log.Warn("garbage collector error", slog.Any(logger.LogFieldError, err))
		}

		for _, subEntry := range subEntries {

		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

	}

}
