package main

import (
	"context"
	"file-storage/internal/config"
	"file-storage/internal/files"
	"file-storage/internal/logger"
	"file-storage/internal/server"
	"file-storage/internal/storage/filesystemstorage"
	"file-storage/internal/storage/inmemory"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	bootstrapLogger := logger.NewBootstrap().With("service", "file-storage")

	cfg, err := getConfig()
	if err != nil {
		bootstrapLogger.Error("load configuration error", "error", err)
		os.Exit(1)
	}

	log := logger.New(&cfg.Log).With("service", "file-storage")

	logConfig(log, cfg)

	var storage files.Storage

	switch cfg.App.Storage {
	case config.StorageInmemory:
		storage = inmemory.New()
	case config.StorageFileSystem:
		storage = filesystemstorage.New(&cfg.Storage.FileSystem)
	}

	svc := files.NewService(&cfg.Image, storage)
	srv := server.NewServer(&cfg.App, svc, log)

	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Run(ctx, cfg.App.Security)
	}()

	select {
	case <-ctx.Done():
		shutdown(srv, nil)
	case err := <-errCh:
		shutdown(srv, err)
	}
}

func shutdown(srv *server.Server, runErr error) {
	if runErr == nil {
		srv.Log.Info("server shutdown")
	} else {
		srv.Log.Error("server execution error", logger.LogFieldError, runErr)
	}
	shutdownCtx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	err := srv.Shutdown(shutdownCtx)
	if err != nil {
		srv.Log.Error("server shutdown error", logger.LogFieldError, err)
	}
}
