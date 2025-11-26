package logger

import (
	"file-storage/internal/config"
	"log/slog"
	"os"
)

func NewBootstrap() *slog.Logger {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil)).With("phase", "bootstrap")
	return logger
}

func New(config *config.Config) *slog.Logger {

	level := slog.LevelInfo
	if config.Log.Level == "Debug" {
		level = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{Level: level}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	return logger
}
