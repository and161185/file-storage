package main

import (
	"context"
	"file-storage/internal/config"
	"file-storage/internal/files"
	"file-storage/internal/logger"
	"file-storage/internal/server"
	"file-storage/internal/storage/filesystemstorage"
	"file-storage/internal/storage/inmemory"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/pflag"
)

var version = "dev"

func main() {

	var showVersion bool
	configPathFlag := pflag.String("config", "", "config file path")
	pflag.BoolVar(&showVersion, "version", false, "print version and exit")
	pflag.Int("port", 0, "application port")
	pflag.String("host", "", "application host")
	pflag.String("loglevel", "info", "log level")
	pflag.String("logtype", "json", "log type")
	pflag.String("readtoken", "", "read token")
	pflag.String("writetoken", "", "write token")
	pflag.Duration("timeout", 5*time.Second, "request timeout")
	pflag.Int("sizelimit", 0, "sizelimit")
	pflag.String("imageext", "", "stored image format")
	pflag.Int("imageMaxDimension", 0, "max stored image dimension")
	pflag.String("storage", "", "storage")
	pflag.String("fsstoragepath", "", "file system storage path")
	pflag.Duration("fsstoragelocklifetime", 5*time.Second, "file system lock lifetime")
	pflag.Parse()

	bootstrapLogger := logger.NewBootstrap().With("service", "file-storage")

	if showVersion {
		fmt.Println(version)
		return
	}

	cfg, err := getConfig(*configPathFlag)
	if err != nil {
		bootstrapLogger.Error("load configuration error", "error", err)
		os.Exit(1)
	}

	log := logger.New(&cfg.Log).With("service", "file-storage")

	log.Info("starting file-storage", "version", version)
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
