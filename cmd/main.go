package main

import (
	"context"
	"file-storage/internal/files"
	"file-storage/internal/logger"
	"file-storage/internal/server"
	"file-storage/internal/storage/inmemory"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	bootstrapLogger := logger.NewBootstrap().With("service", "file-storage")

	config, err := getConfig()
	if err != nil {
		bootstrapLogger.Error("load configuration error", "error", err)
		os.Exit(1)
	}

	log := logger.New(&config.Log).With("service", "file-storage")

	storage := inmemory.New()
	svc := files.NewService(&config.Image, storage)
	srv := server.NewServer(&config.App, svc, log)

	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Run(ctx, config.App.Security)
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
