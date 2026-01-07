package main

import (
	"context"
	"file-storage/internal/files"
	"file-storage/internal/logger"
	"file-storage/internal/server"
	"file-storage/internal/storage/inmemory"
	"fmt"
	"os"
	"os/signal"
	"syscall"
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
	svc := files.NewService(storage)
	srv := server.NewServer(&config.App, svc, log)

	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Run(ctx)
	}()

	select {
	case <-ctx.Done():
		fmt.Print("")
	case err := <-errCh:
		fmt.Print(err.Error())
	}

}
