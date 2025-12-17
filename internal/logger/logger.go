package logger

import (
	"file-storage/internal/config"
	"log/slog"
	"os"
	"strings"
)

func NewBootstrap() *slog.Logger {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil)).With("phase", "bootstrap")
	return logger
}

func New(cfg *config.Log) *slog.Logger {

	level := slog.LevelInfo
	switch strings.ToLower(cfg.Level) {
	case config.LogLevelDebug:
		level = slog.LevelDebug
	case config.LogLevelError:
		level = slog.LevelError
	case config.LogLevelWarn:
		level = slog.LevelWarn
	}
	opts := &slog.HandlerOptions{Level: level}

	var logger *slog.Logger
	if cfg.Type == config.LogTypeJSON {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	return logger
}
