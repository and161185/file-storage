package main

import (
	"context"
	"file-storage/internal/logger"
	"file-storage/internal/server"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	bootstrapLogger := logger.NewBootstrap()

	config, err := getConfig()
	if err != nil {
		bootstrapLogger.Error("load configuration error", "error", err)
		os.Exit(1)
	}

	srv := server.NewServer(config)

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
